package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// --- Public types (consumed by TUI and update command) ---

type TrackingResponse struct {
	TrackingNumber   string
	StatusCategory   string
	Status           string
	ExpectedDelivery string
	OriginCity       string
	OriginState      string
	DestCity         string
	DestState        string
	TrackingEvents   []TrackingEvent
}

type TrackingEvent struct {
	EventDate        string
	EventDescription string
	City             string
	State            string
	Zip              string
	Country          string
}

// --- Client ---

type Client struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
}

func NewClient() *Client {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	return &Client{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
	}
}

func (c *Client) Close() {
	if c.allocCancel != nil {
		c.allocCancel()
	}
}

func (c *Client) GetTracking(trackingNumber string) (*TrackingResponse, error) {
	ctx, cancel := chromedp.NewContext(c.allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	url := "https://tools.usps.com/go/TrackConfirmAction?tLabels=" + trackingNumber

	// Step 1: Inject anti-detection script before any page loads (CDP level)
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(
				`Object.defineProperty(navigator, 'webdriver', {get: () => undefined});`,
			).Do(ctx)
			return err
		}),
		chromedp.Navigate(url),
	)
	if err != nil {
		return nil, fmt.Errorf("navigating to USPS: %w", err)
	}

	// Step 2: Poll for results (1s intervals, max 30 tries = 30s)
	var found bool
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		err = chromedp.Run(ctx,
			chromedp.Evaluate(`
				!!document.querySelector('.banner-header') ||
				!!document.querySelector('.tracking_history_container') ||
				!!document.querySelector('.delivery-status-header') ||
				!!document.querySelector('.tb-step')
			`, &found),
		)
		if err != nil {
			continue
		}
		if found {
			break
		}
	}

	// Step 3: Extract all data in one JS call (handles both old and new USPS page designs)
	var resultJSON string
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				var result = {status: '', banner: '', delivery: '', events: []};

				// Status — try multiple selector patterns
				var bh = document.querySelector('.banner-header')
					|| document.querySelector('.latest-update-banner-wrapper h3')
					|| document.querySelector('.delivery-status-header h2')
					|| document.querySelector('.tb-step.current .tb-status-detail')
					|| document.querySelector('.tb-step .tb-status');
				result.status = bh ? bh.innerText.trim() : '';

				// Banner / detail text
				var bc = document.querySelector('.banner-content')
					|| document.querySelector('.latest-update-banner-wrapper p')
					|| document.querySelector('.tb-step.current .tb-date');
				result.banner = bc ? bc.innerText.trim() : '';

				// Expected delivery
				var ed = document.querySelector('.expected-delivery')
					|| document.querySelector('.expected-delivery-short-date')
					|| document.querySelector('[class*="expected"] [class*="date"]')
					|| document.querySelector('.expected-delivery-date');
				result.delivery = ed ? ed.innerText.trim() : '';

				// Events — table rows (old design)
				var rows = document.querySelectorAll('.tracking_history_container table tr');
				if (rows.length === 0) {
					rows = document.querySelectorAll('.product_tracking_details table tr');
				}
				if (rows.length > 0) {
					for (var i = 0; i < rows.length; i++) {
						var cells = rows[i].querySelectorAll('td');
						if (cells.length >= 2) {
							result.events.push({
								date: cells[0].innerText.trim(),
								description: cells[1].innerText.trim(),
								location: cells.length >= 3 ? cells[2].innerText.trim() : ''
							});
						}
					}
				} else {
					// New USPS design: .tb-step elements
					var steps = document.querySelectorAll('.tb-step');
					for (var j = 0; j < steps.length; j++) {
						var s = steps[j];
						var desc = s.querySelector('.tb-status-detail') || s.querySelector('.tb-status');
						var date = s.querySelector('.tb-date') || s.querySelector('time');
						var loc  = s.querySelector('.tb-location');
						if (desc) {
							result.events.push({
								date: date ? date.innerText.trim() : '',
								description: desc.innerText.trim(),
								location: loc ? loc.innerText.trim() : ''
							});
						}
					}
				}

				return JSON.stringify(result);
			})()
		`, &resultJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("extracting data: %w", err)
	}

	return parseResultJSON(trackingNumber, resultJSON), nil
}

// --- Internal helpers ---

type rawEvent struct {
	Date        string `json:"date"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

type scrapeResult struct {
	Status   string     `json:"status"`
	Banner   string     `json:"banner"`
	Delivery string     `json:"delivery"`
	Events   []rawEvent `json:"events"`
}

func parseResultJSON(trackingNumber, resultJSON string) *TrackingResponse {
	result := &TrackingResponse{
		TrackingNumber: trackingNumber,
		StatusCategory: "unknown",
	}

	if resultJSON == "" {
		return result
	}

	var sr scrapeResult
	if err := json.Unmarshal([]byte(resultJSON), &sr); err != nil {
		return result
	}

	statusText := sr.Status
	if statusText == "" {
		statusText = sr.Banner
	}
	result.Status = statusText
	result.StatusCategory = normalizeUSPSStatus(statusText)
	result.ExpectedDelivery = sr.Delivery

	result.TrackingEvents = make([]TrackingEvent, 0, len(sr.Events))
	for _, re := range sr.Events {
		city, state, zip := parseLocation(re.Location)
		result.TrackingEvents = append(result.TrackingEvents, TrackingEvent{
			EventDate:        re.Date,
			EventDescription: re.Description,
			City:             city,
			State:            state,
			Zip:              zip,
			Country:          "US",
		})
	}

	if len(result.TrackingEvents) > 0 {
		first := result.TrackingEvents[0]
		last := result.TrackingEvents[len(result.TrackingEvents)-1]
		result.DestCity = first.City
		result.DestState = first.State
		result.OriginCity = last.City
		result.OriginState = last.State
	}

	return result
}

func parseResults(trackingNumber, statusText, expectedDelivery, eventsJSON string) *TrackingResponse {
	result := &TrackingResponse{
		TrackingNumber:   trackingNumber,
		Status:           statusText,
		StatusCategory:   normalizeUSPSStatus(statusText),
		ExpectedDelivery: expectedDelivery,
	}

	if eventsJSON == "" || eventsJSON == "[]" {
		return result
	}

	var rawEvents []rawEvent
	if err := json.Unmarshal([]byte(eventsJSON), &rawEvents); err != nil {
		return result
	}

	result.TrackingEvents = make([]TrackingEvent, 0, len(rawEvents))
	for _, re := range rawEvents {
		city, state, zip := parseLocation(re.Location)
		result.TrackingEvents = append(result.TrackingEvents, TrackingEvent{
			EventDate:        re.Date,
			EventDescription: re.Description,
			City:             city,
			State:            state,
			Zip:              zip,
			Country:          "US",
		})
	}

	if len(result.TrackingEvents) > 0 {
		first := result.TrackingEvents[0]
		last := result.TrackingEvents[len(result.TrackingEvents)-1]
		result.DestCity = first.City
		result.DestState = first.State
		result.OriginCity = last.City
		result.OriginState = last.State
	}

	return result
}

func parseLocation(loc string) (city, state, zip string) {
	if loc == "" {
		return "", "", ""
	}
	parts := strings.SplitN(loc, ",", 2)
	city = strings.TrimSpace(parts[0])
	if len(parts) < 2 {
		return city, "", ""
	}
	remainder := strings.TrimSpace(parts[1])
	fields := strings.Fields(remainder)
	if len(fields) >= 1 {
		state = fields[0]
	}
	if len(fields) >= 2 {
		zip = fields[1]
	}
	return city, state, zip
}

func normalizeUSPSStatus(status string) string {
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "delivered"):
		return "delivered"
	case strings.Contains(lower, "out for delivery"):
		return "out_for_delivery"
	case strings.Contains(lower, "in transit"),
		strings.Contains(lower, "arrived"),
		strings.Contains(lower, "departed"),
		strings.Contains(lower, "processed"),
		strings.Contains(lower, "in-transit"):
		return "in_transit"
	case strings.Contains(lower, "accepted"),
		strings.Contains(lower, "shipping label created"),
		strings.Contains(lower, "pre-shipment"),
		strings.Contains(lower, "origin post is preparing"):
		return "pre_transit"
	case strings.Contains(lower, "alert"),
		strings.Contains(lower, "exception"),
		strings.Contains(lower, "undeliverable"),
		strings.Contains(lower, "notice left"),
		strings.Contains(lower, "not available"):
		return "alert"
	case strings.Contains(lower, "return"):
		return "returned"
	default:
		return "unknown"
	}
}

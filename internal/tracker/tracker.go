package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
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

	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	url := "https://tools.usps.com/go/TrackConfirmAction?tLabels=" + trackingNumber

	var statusText string
	var expectedDelivery string
	var eventsJSON string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// Wait for tracking results or error to appear
		chromedp.WaitVisible(`div.delivery-status-header, div.tb-step, .tracking-progress-bar-status, .banner-header, p.tracking_error, .alert-error, .delivery-status`, chromedp.ByQuery),
		// Let dynamic content finish rendering
		chromedp.Sleep(3*time.Second),
		// Extract status text
		chromedp.Evaluate(`
			(function() {
				var el = document.querySelector('.delivery-status-header h2')
					|| document.querySelector('.tb-status-detail h2')
					|| document.querySelector('.banner-header')
					|| document.querySelector('.tb-status h2')
					|| document.querySelector('.delivery-status h2')
					|| document.querySelector('.delivery_status h2');
				return el ? el.innerText.trim() : '';
			})()
		`, &statusText),
		// Extract expected delivery date
		chromedp.Evaluate(`
			(function() {
				var el = document.querySelector('.expected-delivery-short-date')
					|| document.querySelector('.expected-delivery')
					|| document.querySelector('.delivery-date');
				return el ? el.innerText.trim() : '';
			})()
		`, &expectedDelivery),
		// Extract all tracking events as JSON
		chromedp.Evaluate(`
			(function() {
				var events = [];
				var rows = document.querySelectorAll('.tb-step');
				if (rows.length === 0) {
					rows = document.querySelectorAll('.tb-steps-detail');
				}
				if (rows.length === 0) {
					rows = document.querySelectorAll('.result-col');
				}
				for (var i = 0; i < rows.length; i++) {
					var row = rows[i];
					var dateEl = row.querySelector('.tb-date') || row.querySelector('.date-col') || row.querySelector('.tb-date-month');
					var timeEl = row.querySelector('.tb-time') || row.querySelector('.time-col');
					var descEl = row.querySelector('.tb-status-detail') || row.querySelector('.event-col') || row.querySelector('.tb-status');
					var locEl = row.querySelector('.tb-location') || row.querySelector('.location-col');

					var dateText = dateEl ? dateEl.innerText.trim() : '';
					var timeText = timeEl ? timeEl.innerText.trim() : '';
					var desc = descEl ? descEl.innerText.trim() : '';
					var loc = locEl ? locEl.innerText.trim() : '';

					if (desc || dateText) {
						events.push({
							date: dateText + (timeText ? ' ' + timeText : ''),
							description: desc,
							location: loc
						});
					}
				}
				return JSON.stringify(events);
			})()
		`, &eventsJSON),
	)

	if err != nil {
		return nil, fmt.Errorf("scraping USPS: %w", err)
	}

	return parseResults(trackingNumber, statusText, expectedDelivery, eventsJSON), nil
}

// --- Internal helpers ---

type rawEvent struct {
	Date        string `json:"date"`
	Description string `json:"description"`
	Location    string `json:"location"`
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
		strings.Contains(lower, "notice left"):
		return "alert"
	case strings.Contains(lower, "return"):
		return "returned"
	default:
		return "unknown"
	}
}

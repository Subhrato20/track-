package tracker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// --- Ship24 API response types ---

type ship24Response struct {
	Data struct {
		Trackings []ship24Tracking `json:"trackings"`
		Shipments []ship24Shipment `json:"shipments"`
	} `json:"data"`
}

type ship24Tracking struct {
	TrackingNumber string `json:"trackingNumber"`
}

type ship24Shipment struct {
	ShipmentID             string         `json:"shipmentId"`
	StatusMilestone        string         `json:"statusMilestone"`
	OriginCountryCode      string         `json:"originCountryCode"`
	DestinationCountryCode string         `json:"destinationCountryCode"`
	Events                 []ship24Event  `json:"events"`
}

type ship24Event struct {
	EventID         string `json:"eventId"`
	TrackingNumber  string `json:"trackingNumber"`
	DateTime        string `json:"datetime"`
	Status          string `json:"status"`
	StatusMilestone string `json:"statusMilestone"`
	Description     string `json:"description"`
	Location        string `json:"location"`
	Country         string `json:"country"`
	State           string `json:"state"`
	City            string `json:"city"`
	Zip             string `json:"zip"`
}

type ship24CreateRequest struct {
	TrackingNumber string `json:"trackingNumber"`
}

type ship24CreateResponse struct {
	Data struct {
		Tracker struct {
			TrackerID      string `json:"trackerId"`
			TrackingNumber string `json:"trackingNumber"`
		} `json:"tracker"`
	} `json:"data"`
}

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
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    "https://api.ship24.com/public/v1",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetTracking(trackingNumber string) (*TrackingResponse, error) {
	// Step 1: Create a tracker (idempotent — safe to call if already exists)
	c.createTracker(trackingNumber)

	// Step 2: Get tracking results
	url := fmt.Sprintf("%s/trackers/search/%s/results", c.baseURL, trackingNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ship24 API returned %d: %s", resp.StatusCode, string(body))
	}

	var s24Resp ship24Response
	if err := json.Unmarshal(body, &s24Resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return c.convertResponse(trackingNumber, &s24Resp), nil
}

func (c *Client) createTracker(trackingNumber string) {
	payload, _ := json.Marshal(ship24CreateRequest{TrackingNumber: trackingNumber})

	req, err := http.NewRequest("POST", c.baseURL+"/trackers", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (c *Client) convertResponse(trackingNumber string, s24 *ship24Response) *TrackingResponse {
	result := &TrackingResponse{
		TrackingNumber: trackingNumber,
		StatusCategory: "unknown",
		Status:         "Unknown",
	}

	if len(s24.Data.Shipments) == 0 {
		return result
	}

	shipment := s24.Data.Shipments[0]
	result.StatusCategory = normalizeStatus(shipment.StatusMilestone)
	result.Status = shipment.StatusMilestone

	// Extract origin from last event, destination from first event
	if len(shipment.Events) > 0 {
		first := shipment.Events[0]
		last := shipment.Events[len(shipment.Events)-1]

		result.DestCity = cityFromEvent(first)
		result.DestState = first.State
		result.OriginCity = cityFromEvent(last)
		result.OriginState = last.State
	}

	// Convert events
	result.TrackingEvents = make([]TrackingEvent, len(shipment.Events))
	for i, e := range shipment.Events {
		result.TrackingEvents[i] = TrackingEvent{
			EventDate:        e.DateTime,
			EventDescription: e.Description,
			City:             cityFromEvent(e),
			State:            e.State,
			Zip:              e.Zip,
			Country:          e.Country,
		}
	}

	return result
}

func cityFromEvent(e ship24Event) string {
	if e.City != "" {
		return e.City
	}
	// Fallback: parse from location string like "New York, NY 10001"
	if e.Location != "" {
		parts := strings.SplitN(e.Location, ",", 2)
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return ""
}

func normalizeStatus(milestone string) string {
	switch strings.ToLower(milestone) {
	case "delivered":
		return "delivered"
	case "in_transit", "intransit", "in transit":
		return "in_transit"
	case "out_for_delivery", "outfordelivery", "out for delivery":
		return "out_for_delivery"
	case "info_received", "inforeceived", "pending", "info received":
		return "pre_transit"
	case "exception", "alert", "failed_attempt":
		return "alert"
	case "returned":
		return "returned"
	default:
		return "unknown"
	}
}

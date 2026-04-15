package usps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TrackingResponse struct {
	TrackingNumber   string          `json:"trackingNumber"`
	StatusCategory   string          `json:"statusCategory"`
	Status           string          `json:"status"`
	ExpectedDelivery string          `json:"expectedDeliveryDate"`
	OriginCity       string          `json:"originCity"`
	OriginState      string          `json:"originState"`
	DestCity         string          `json:"destinationCity"`
	DestState        string          `json:"destinationState"`
	TrackingEvents   []TrackingEvent `json:"trackingEvents"`
}

type TrackingEvent struct {
	EventDate        string `json:"eventTimestamp"`
	EventDescription string `json:"eventType"`
	City             string `json:"eventCity"`
	State            string `json:"eventState"`
	Zip              string `json:"eventZIPCode"`
	Country          string `json:"eventCountry"`
}

type Client struct {
	auth       *AuthClient
	baseURL    string
	httpClient *http.Client
}

func NewClient(clientID, clientSecret, baseURL string) *Client {
	return &Client{
		auth:       NewAuthClient(clientID, clientSecret, baseURL),
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetTracking(trackingNumber string) (*TrackingResponse, error) {
	token, err := c.auth.GetToken()
	if err != nil {
		return nil, fmt.Errorf("getting auth token: %w", err)
	}

	resp, err := c.doTrackingRequest(trackingNumber, token)
	if err != nil {
		return nil, err
	}

	// Retry once on 401
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.auth.mu.Lock()
		c.auth.token = "" // force refresh
		c.auth.mu.Unlock()

		token, err = c.auth.GetToken()
		if err != nil {
			return nil, fmt.Errorf("refreshing auth token: %w", err)
		}
		resp, err = c.doTrackingRequest(trackingNumber, token)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracking API returned %d: %s", resp.StatusCode, string(body))
	}

	var trackResp TrackingResponse
	if err := json.Unmarshal(body, &trackResp); err != nil {
		return nil, fmt.Errorf("parsing tracking response: %w", err)
	}

	return &trackResp, nil
}

func (c *Client) doTrackingRequest(trackingNumber, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/tracking/v3/tracking/%s", c.baseURL, trackingNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating tracking request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tracking request failed: %w", err)
	}
	return resp, nil
}

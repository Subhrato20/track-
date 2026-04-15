package usps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
	ExpiresIn   string `json:"expires_in"`
	Status      string `json:"status"`
	Scope       string `json:"scope"`
}

type AuthClient struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client
	token        string
	tokenExpiry  time.Time
	mu           sync.Mutex
}

func NewAuthClient(clientID, clientSecret, baseURL string) *AuthClient {
	return &AuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *AuthClient) GetToken() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token != "" && time.Now().Before(a.tokenExpiry) {
		return a.token, nil
	}

	return a.refreshToken()
}

func (a *AuthClient) refreshToken() (string, error) {
	body, _ := json.Marshal(map[string]string{
		"client_id":     a.clientID,
		"client_secret": a.clientSecret,
		"grant_type":    "client_credentials",
	})

	req, err := http.NewRequest("POST", a.baseURL+"/oauth2/v3/token", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request returned %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	a.token = tokenResp.AccessToken

	expiresIn, err := strconv.Atoi(tokenResp.ExpiresIn)
	if err != nil {
		expiresIn = 3600 // default 1 hour
	}
	a.tokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second)

	return a.token, nil
}

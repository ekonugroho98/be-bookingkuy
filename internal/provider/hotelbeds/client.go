package hotelbeds

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Client handles Hotelbeds API communication
type Client struct {
	apiKey        string
	sharedSecret string
	baseURL       string
	httpClient    *http.Client
	rateLimiter   *RateLimiter
}

// NewClient creates a new Hotelbeds client
func NewClient(apiKey, sharedSecret, baseURL string) *Client {
	return &Client{
		apiKey:        apiKey,
		sharedSecret: sharedSecret,
		baseURL:       baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: NewRateLimiter(100), // 100 requests per minute limit
	}
}

// Do performs an HTTP request to Hotelbeds API
func (c *Client) Do(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build request URL
	url := c.baseURL + endpoint

	// Marshal body if provided
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	c.setHeaders(req)

	// Log request
	logger.Infof("Hotelbeds API Request: %s %s", method, endpoint)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Hotelbeds API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// setHeaders sets required headers for Hotelbeds API
func (c *Client) setHeaders(req *http.Request) {
	// Set standard headers
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("X-Signature", c.generateSignature())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set timestamp
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")
	req.Header.Set("X-Timestamp", timestamp)
}

// generateSignature generates Hotelbeds API signature
func (c *Client) generateSignature() string {
	// Hotelbeds signature format: SHA256(sharedSecret + timestamp)
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")
	data := c.sharedSecret + timestamp

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Get performs GET request
func (c *Client) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, endpoint, nil)
}

// Post performs POST request
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, endpoint, body)
}

// Put performs PUT request
func (c *Client) Put(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	return c.Do(ctx, http.MethodPut, endpoint, body)
}

// Delete performs DELETE request
func (c *Client) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.Do(ctx, http.MethodDelete, endpoint, nil)
}

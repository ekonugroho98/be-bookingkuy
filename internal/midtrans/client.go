package midtrans

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

const (
	// Sandbox URLs
	SandboxBaseURL    = "https://api.sandbox.midtrans.com/v2"
	SandboxSnapURL    = "https://app.sandbox.midtrans.com/snap/v1"

	// Production URLs
	ProductionBaseURL = "https://api.midtrans.com/v2"
	ProductionSnapURL = "https://app.midtrans.com/snap/v1"
)

// Config represents Midtrans configuration
type Config struct {
	ServerKey    string
	ClientKey    string
	MerchantID   string
	IsProduction bool
	Timeout      time.Duration
}

// Client represents Midtrans API client
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
	snapURL    string
}

// NewClient creates a new Midtrans client
func NewClient(config Config) *Client {
	baseURL := SandboxBaseURL
	snapURL := SandboxSnapURL

	if config.IsProduction {
		baseURL = ProductionBaseURL
		snapURL = ProductionSnapURL
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		config:     config,
		httpClient: &http.Client{Timeout: config.Timeout},
		baseURL:    baseURL,
		snapURL:    snapURL,
	}
}

// Charge creates a new transaction
func (c *Client) Charge(req *ChargeRequest) (*ChargeResponse, error) {
	url := c.snapURL + "/transactions"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to charge: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		logger.Errorf("Midtrans charge failed: %s", string(respBody))
		return nil, fmt.Errorf("charge failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var chargeResp ChargeResponse
	if err := json.Unmarshal(respBody, &chargeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logger.Infof("Midtrans charge successful: OrderID=%s, TransactionID=%s",
		chargeResp.OrderID, chargeResp.TransactionID)

	return &chargeResp, nil
}

// GetTransactionStatus retrieves transaction status
func (c *Client) GetTransactionStatus(orderID string) (*GetTransactionStatusResponse, error) {
	url := fmt.Sprintf("%s/%s/status", c.baseURL, orderID)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get status failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var statusResp GetTransactionStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &statusResp, nil
}

// Cancel cancels a transaction
func (c *Client) Cancel(orderID string) (*CancelResponse, error) {
	url := fmt.Sprintf("%s/%s/cancel", c.baseURL, orderID)

	resp, err := c.doRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("cancel failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var cancelResp CancelResponse
	if err := json.Unmarshal(respBody, &cancelResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logger.Infof("Midtrans transaction cancelled: OrderID=%s", orderID)
	return &cancelResp, nil
}

// doRequest performs HTTP request with authentication
func (c *Client) doRequest(method, url string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Midtrans uses Basic Auth with server key as username and empty password
	// Format: base64(server_key:)
	auth := base64.StdEncoding.EncodeToString([]byte(c.config.ServerKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	logger.Debugf("Midtrans %s request: %s", method, url)

	return c.httpClient.Do(req)
}

// GenerateSnapToken generates SNAP token for payment frontend
func (c *Client) GenerateSnapToken(req *ChargeRequest) (string, error) {
	resp, err := c.Charge(req)
	if err != nil {
		return "", err
	}

	// For SNAP, return token or redirect URL
	if resp.TokenID != "" {
		return resp.TokenID, nil
	}

	if resp.RedirectURL != "" {
		return resp.RedirectURL, nil
	}

	if resp.PaymentURL != "" {
		return resp.PaymentURL, nil
	}

	return "", fmt.Errorf("no token or redirect URL in response")
}

// ValidateWebhookSignature validates webhook signature
func (c *Client) ValidateWebhookSignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	// Signature format: SHA512(order_id + status_code + gross_amount + server_key)
	data := orderID + statusCode + grossAmount + c.config.ServerKey
	hash := sha512.Sum512([]byte(data))
	calculatedSignature := fmt.Sprintf("%x", hash)

	return calculatedSignature == signatureKey
}

// MapTransactionStatus maps Midtrans status to our internal status
func MapTransactionStatus(midtransStatus TransactionStatus) string {
	switch midtransStatus {
	case StatusPending, StatusAuthorize:
		return "pending"
	case StatusCapture, StatusSettlement:
		return "success"
	case StatusDeny, StatusFailure, StatusExpire:
		return "failed"
	case StatusCancel, StatusPendingCancel:
		return "cancelled"
	case StatusRefund, StatusPartialRefund:
		return "refunded"
	default:
		return "unknown"
	}
}

// IsTransactionFinal checks if transaction status is final (no further updates)
func IsTransactionFinal(status TransactionStatus) bool {
	switch status {
	case StatusSettlement, StatusCapture, StatusCancel, StatusDeny,
	     StatusExpire, StatusFailure, StatusRefund, StatusPartialRefund:
		return true
	default:
		return false
	}
}

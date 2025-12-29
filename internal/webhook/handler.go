package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Handler handles webhooks from external services
type Handler struct {
	secret      string
	httpClient  *http.Client
}

// NewHandler creates a new webhook handler
func NewHandler(secret string) *Handler {
	return &Handler{
		secret: secret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VerifySignature verifies webhook signature
func (h *Handler) VerifySignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// SendWebhook sends a webhook to external URL
func (h *Handler) SendWebhook(ctx context.Context, url string, payload map[string]interface{}) error {
	logger.Infof("Sending webhook to %s", url)

	// Create webhook payload with signature
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	signature := h.signPayload(data)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	req.Header.Set("User-Agent", "Bookingkuy-Webhook/1.0")

	// Send request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	logger.Infof("Webhook sent successfully to %s (status: %d)", url, resp.StatusCode)
	return nil
}

// signPayload creates signature for payload
func (h *Handler) signPayload(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// RetryWebhook retries webhook with exponential backoff
func (h *Handler) RetryWebhook(ctx context.Context, url string, payload map[string]interface{}, maxAttempts int) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := h.SendWebhook(ctx, url, payload)
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt < maxAttempts {
			// Exponential backoff: 1s, 2s, 4s, 8s, etc.
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 60*time.Second {
				backoff = 60 * time.Second // Max 60 seconds
			}

			logger.Infof("Webhook attempt %d/%d failed, retrying in %v: %v", attempt, maxAttempts, backoff, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", maxAttempts, lastErr)
}

// HandleMidtransWebhook handles Midtrans webhook
func (h *Handler) HandleMidtransWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Verify signature if present
	if signature := r.Header.Get("X-Signature"); signature != "" {
		payloadBytes, _ := json.Marshal(payload)
		if !h.VerifySignature(payloadBytes, signature) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	logger.Infof("Midtrans webhook received: %+v", payload)

	// TODO: Process webhook payload and update payment status
	// This should trigger payment service to handle the webhook

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleHotelbedsWebhook handles Hotelbeds webhook
func (h *Handler) HandleHotelbedsWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Verify signature if present
	if signature := r.Header.Get("X-Signature"); signature != "" {
		payloadBytes, _ := json.Marshal(payload)
		if !h.VerifySignature(payloadBytes, signature) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	logger.Infof("Hotelbeds webhook received: %+v", payload)

	// TODO: Process webhook payload
	// Handle booking confirmations, cancellations, etc. from Hotelbeds

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// SendBookingWebhook sends booking notification to external webhook
func (h *Handler) SendBookingWebhook(ctx context.Context, url string, eventType string, bookingData map[string]interface{}) error {
	payload := map[string]interface{}{
		"event":      eventType,
		"timestamp":  time.Now().Format(time.RFC3339),
		"data":       bookingData,
	}

	return h.SendWebhook(ctx, url, payload)
}

// SendPaymentWebhook sends payment notification to external webhook
func (h *Handler) SendPaymentWebhook(ctx context.Context, url string, paymentData map[string]interface{}) error {
	payload := map[string]interface{}{
		"event":      "payment.status_updated",
		"timestamp":  time.Now().Format(time.RFC3339),
		"data":       paymentData,
	}

	return h.SendWebhook(ctx, url, payload)
}

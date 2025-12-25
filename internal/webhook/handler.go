package webhook

import (
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
	secret string
}

// NewHandler creates a new webhook handler
func NewHandler(secret string) *Handler {
	return &Handler{
		secret: secret,
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
		return err
	}

	signature := h.signPayload(data)

	// TODO: Implement actual HTTP POST to webhook URL
	// For now, just log
	logger.Infof("Webhook payload: %s, signature: %s", string(data), signature)

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
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = h.SendWebhook(ctx, url, payload); err == nil {
			return nil
		}

		if attempt < maxAttempts {
			backoff := time.Duration(attempt) * time.Second
			logger.Infof("Webhook attempt %d failed, retrying in %v", attempt, backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", maxAttempts, err)
}

// HandleMidtransWebhook handles Midtrans webhook
func (h *Handler) HandleMidtransWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Signature")
	if !h.VerifySignature([]byte(fmt.Sprintf("%v", payload)), signature) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	logger.Infof("Midtrans webhook received: %+v", payload)
	w.WriteHeader(http.StatusOK)
}

// HandleHotelbedsWebhook handles Hotelbeds webhook
func (h *Handler) HandleHotelbedsWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Signature")
	if !h.VerifySignature([]byte(fmt.Sprintf("%v", payload)), signature) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	logger.Infof("Hotelbeds webhook received: %+v", payload)
	w.WriteHeader(http.StatusOK)
}

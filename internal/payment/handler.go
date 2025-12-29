package payment

import (
	"encoding/json"
	"net/http"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/middleware"
)

// Handler handles HTTP requests for payment operations
type Handler struct {
	service Service
}

// NewHandler creates a new payment handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreatePayment handles POST /payments
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in request")
		return
	}

	_ = userID // Will be used to verify booking ownership

	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// TODO: Get booking amount from booking service
	// For now, use default amount
	amount := 1000000 // 1 million IDR

	payment, err := h.service.CreatePayment(r.Context(), &req, amount)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to create payment")
		if err == ErrInvalidPayment {
			respondWithError(w, http.StatusConflict, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to create payment")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, payment)
}

// HandleWebhook handles POST /payments/webhook
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.HandleWebhook(r.Context(), &payload); err != nil {
		logger.ErrorWithErr(err, "Failed to handle webhook")
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Webhook processed successfully"})
}

// GetPayment handles GET /payments/{id}
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		respondWithError(w, http.StatusBadRequest, "Payment ID is required")
		return
	}

	payment, err := h.service.GetPayment(r.Context(), paymentID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get payment")
		if err == ErrPaymentNotFound {
			respondWithError(w, http.StatusNotFound, "Payment not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, payment)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, map[string]string{"error": message})
}

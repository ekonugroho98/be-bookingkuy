package booking

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/middleware"
)

// Handler handles HTTP requests for booking operations
type Handler struct {
	service Service
}

// NewHandler creates a new booking handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateBooking handles POST /bookings
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in request")
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	booking, err := h.service.CreateBooking(r.Context(), userID, &req)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to create booking")
		respondWithError(w, http.StatusInternalServerError, "Failed to create booking")
		return
	}

	respondWithJSON(w, http.StatusCreated, booking)
}

// GetBooking handles GET /bookings/{id}
func (h *Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := r.PathValue("id")
	if bookingID == "" {
		respondWithError(w, http.StatusBadRequest, "Booking ID is required")
		return
	}

	booking, err := h.service.GetBooking(r.Context(), bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking")
		if errors.Is(err, errors.New("booking not found")) {
			respondWithError(w, http.StatusNotFound, "Booking not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, booking)
}

// GetMyBookings handles GET /bookings/my
func (h *Handler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in request")
		return
	}

	// Parse pagination
	page := 1
	perPage := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	bookings, err := h.service.GetUserBookings(r.Context(), userID, page, perPage)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get user bookings")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"bookings": bookings,
		"page":     page,
		"per_page": perPage,
	})
}

// CancelBooking handles POST /bookings/{id}/cancel
func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := r.PathValue("id")
	if bookingID == "" {
		respondWithError(w, http.StatusBadRequest, "Booking ID is required")
		return
	}

	booking, err := h.service.CancelBooking(r.Context(), bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to cancel booking")
		respondWithError(w, http.StatusInternalServerError, "Failed to cancel booking")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Booking cancelled successfully",
		"booking": booking,
	})
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, map[string]string{"error": message})
}

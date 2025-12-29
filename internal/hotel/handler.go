package hotel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Handler handles hotel HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new hotel handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetHotel handles GET /api/v1/hotels/{id}
func (h *Handler) GetHotel(w http.ResponseWriter, r *http.Request) {
	hotelID := r.PathValue("id")
	if hotelID == "" {
		respondWithError(w, http.StatusBadRequest, "Hotel ID is required")
		return
	}

	logger.Infof("GetHotel request for hotel: %s", hotelID)

	hotel, err := h.service.GetHotel(r.Context(), hotelID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get hotel")
		respondWithError(w, http.StatusNotFound, "Hotel not found")
		return
	}

	respondWithJSON(w, http.StatusOK, hotel)
}

// GetAvailableRooms handles GET /api/v1/hotels/{id}/rooms
func (h *Handler) GetAvailableRooms(w http.ResponseWriter, r *http.Request) {
	hotelID := r.PathValue("id")
	if hotelID == "" {
		respondWithError(w, http.StatusBadRequest, "Hotel ID is required")
		return
	}

	// Parse query parameters
	checkInStr := r.URL.Query().Get("checkIn")
	checkOutStr := r.URL.Query().Get("checkOut")
	guests := r.URL.Query().Get("guests")

	if checkInStr == "" || checkOutStr == "" || guests == "" {
		respondWithError(w, http.StatusBadRequest, "checkIn, checkOut, and guests query parameters are required")
		return
	}

	// Parse dates
	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid checkIn date format. Use YYYY-MM-DD")
		return
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid checkOut date format. Use YYYY-MM-DD")
		return
	}

	// Parse guests (default to 1 if parsing fails)
	var guestsNum int
	if _, err := fmt.Sscanf(guests, "%d", &guestsNum); err != nil || guestsNum < 1 {
		guestsNum = 1
	}

	logger.Infof("GetAvailableRooms request for hotel: %s, checkIn: %s, checkOut: %s, guests: %d",
		hotelID, checkInStr, checkOutStr, guestsNum)

	rooms, err := h.service.GetAvailableRooms(r.Context(), hotelID, checkIn, checkOut, guestsNum)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get available rooms")
		respondWithError(w, http.StatusInternalServerError, "Failed to check room availability")
		return
	}

	respondWithJSON(w, http.StatusOK, rooms)
}

// GetImages handles GET /api/v1/hotels/{id}/images
func (h *Handler) GetImages(w http.ResponseWriter, r *http.Request) {
	hotelID := r.PathValue("id")
	if hotelID == "" {
		respondWithError(w, http.StatusBadRequest, "Hotel ID is required")
		return
	}

	logger.Infof("GetImages request for hotel: %s", hotelID)

	images, err := h.service.GetImages(r.Context(), hotelID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get hotel images")
		respondWithError(w, http.StatusInternalServerError, "Failed to get hotel images")
		return
	}

	response := HotelImagesResponse{
		HotelID: hotelID,
		Images:  images,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// respondWithJSON writes JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.ErrorWithErr(err, "Failed to encode JSON response")
	}
}

// respondWithError writes error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}

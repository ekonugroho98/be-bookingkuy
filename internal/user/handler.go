package user

import (
	"encoding/json"
	"net/http"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/middleware"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	service Service
}

// NewHandler creates a new user handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetProfile handles GET /users/me
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in request")
		return
	}

	user, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get user profile")
		if err == ErrUserNotFound {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// UpdateProfile handles PUT /users/me
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in request")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to update user profile")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// Helper functions for HTTP responses
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, map[string]string{"error": message})
}

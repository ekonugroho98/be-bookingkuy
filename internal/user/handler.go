package user

import (
	"encoding/json"
	"net/http"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
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
	// TODO: Extract user ID from JWT context
	userID := "placeholder-user-id"

	user, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user profile: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateProfile handles PUT /users/me
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Extract user ID from JWT context
	user.ID = "placeholder-user-id"

	if err := h.service.UpdateProfile(r.Context(), &user); err != nil {
		logger.Error("Failed to update user profile: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

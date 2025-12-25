package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Handler handles HTTP requests for auth operations
type Handler struct {
	service Service
}

// NewHandler creates a new auth handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register handles POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), &req)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to register user")
		if errors.Is(err, errors.New("email already registered")) {
			respondWithError(w, http.StatusConflict, "Email already registered")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login handles POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.service.Login(r.Context(), &req)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to login user")
		if errors.Is(err, errors.New("invalid email or password")) {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

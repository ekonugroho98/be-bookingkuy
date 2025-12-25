package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Handler handles HTTP requests for review service
type Handler struct {
	service    Service
	jwtSecret  string
	jwtManager *jwt.Manager
}

// NewHandler creates a new review handler
func NewHandler(service Service, jwtSecret string) *Handler {
	return &Handler{
		service:    service,
		jwtSecret:  jwtSecret,
		jwtManager: jwt.NewManager(jwtSecret),
	}
}

// extractUserID extracts user ID from JWT token
func (h *Handler) extractUserID(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	token := parts[1]

	// Parse and validate token
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}

// sendJSON sends a JSON response
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError sends an error response
func sendError(w http.ResponseWriter, status int, message string) {
	logger.Error(message)
	sendJSON(w, status, map[string]string{"error": message})
}

// GetReviewByID handles GET /api/v1/reviews/:id
func (h *Handler) GetReviewByID(w http.ResponseWriter, r *http.Request) {
	// Extract review ID from path
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-1]

	userID, _ := h.extractUserID(r)

	// Get review
	review, err := h.service.GetReviewByID(r.Context(), reviewID, userID)
	if err != nil {
		if err.Error() == "review not found" {
			sendError(w, http.StatusNotFound, "Review not found")
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to get review")
		return
	}

	sendJSON(w, http.StatusOK, review)
}

// GetHotelReviews handles GET /api/v1/reviews/hotel/:hotelId
func (h *Handler) GetHotelReviews(w http.ResponseWriter, r *http.Request) {
	// Extract hotel ID from path
	parts := strings.Split(r.URL.Path, "/")
	hotelID := parts[len(parts)-1]

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	status := ReviewStatus(query.Get("status"))
	sortBy := query.Get("sort_by")

	// Get reviews
	response, err := h.service.GetReviewsByHotelID(r.Context(), hotelID, status, page, pageSize, sortBy)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to get reviews")
		return
	}

	sendJSON(w, http.StatusOK, response)
}

// CreateReview handles POST /api/v1/reviews
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create review
	response, err := h.service.CreateReview(r.Context(), userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "must be at least") ||
		   strings.Contains(err.Error(), "must not exceed") ||
		   strings.Contains(err.Error(), "between 1 and 5") {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to create review")
		return
	}

	sendJSON(w, http.StatusCreated, response)
}

// UpdateReview handles PUT /api/v1/reviews/:id
func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract review ID
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-1]

	var req UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update review
	review, err := h.service.UpdateReview(r.Context(), reviewID, userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "can only edit") || strings.Contains(err.Error(), "must be") {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to update review")
		return
	}

	sendJSON(w, http.StatusOK, review)
}

// DeleteReview handles DELETE /api/v1/reviews/:id
func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract review ID
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-1]

	// Delete review
	err = h.service.DeleteReview(r.Context(), reviewID, userID)
	if err != nil {
		if err.Error() == "you can only delete your own reviews" {
			sendError(w, http.StatusForbidden, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to delete review")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMyReviews handles GET /api/v1/reviews/my-reviews
func (h *Handler) GetMyReviews(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))

	// Get reviews
	response, err := h.service.GetMyReviews(r.Context(), userID, page, pageSize)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to get reviews")
		return
	}

	sendJSON(w, http.StatusOK, response)
}

// ToggleHelpful handles POST /api/v1/reviews/:id/helpful
func (h *Handler) ToggleHelpful(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract review ID
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-2]

	// Toggle helpful vote
	voted, err := h.service.ToggleHelpful(r.Context(), reviewID, userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to toggle helpful vote")
		return
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"voted": voted,
	})
}

// FlagReview handles POST /api/v1/reviews/:id/flag
func (h *Handler) FlagReview(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract review ID
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-2]

	var req FlagReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Flag review
	err = h.service.FlagReview(r.Context(), reviewID, userID, &req)
	if err != nil {
		if err.Error() == "you cannot flag your own review" {
			sendError(w, http.StatusForbidden, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to flag review")
		return
	}

	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Review flagged successfully",
	})
}

// GetHotelStats handles GET /api/v1/reviews/hotel/:hotelId/stats
func (h *Handler) GetHotelStats(w http.ResponseWriter, r *http.Request) {
	// Extract hotel ID from path
	parts := strings.Split(r.URL.Path, "/")
	hotelID := parts[len(parts)-2]

	// Get stats
	stats, err := h.service.GetHotelStats(r.Context(), hotelID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	sendJSON(w, http.StatusOK, stats)
}

// === ADMIN ENDPOINTS ===

// GetPendingReviews handles GET /api/v1/admin/reviews/pending
func (h *Handler) GetPendingReviews(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))

	// Get pending reviews
	response, err := h.service.GetPendingReviews(r.Context(), page, pageSize)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to get pending reviews")
		return
	}

	sendJSON(w, http.StatusOK, response)
}

// GetFlaggedReviews handles GET /api/v1/admin/reviews/flagged
func (h *Handler) GetFlaggedReviews(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))

	// Get flagged reviews
	response, err := h.service.GetFlaggedReviews(r.Context(), page, pageSize)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to get flagged reviews")
		return
	}

	sendJSON(w, http.StatusOK, response)
}

// ModerateReview handles PUT /api/v1/admin/reviews/:id/moderate
func (h *Handler) ModerateReview(w http.ResponseWriter, r *http.Request) {
	// Extract admin ID from JWT (simplified - should use admin middleware)
	userID, err := h.extractUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract review ID
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-2]

	var req ModerateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Moderate review
	err = h.service.ModerateReview(r.Context(), userID, reviewID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "already been") {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to moderate review")
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Review %s successfully", req.Action),
	})
}

// GetModerationStats handles GET /api/v1/admin/reviews/stats
func (h *Handler) GetModerationStats(w http.ResponseWriter, r *http.Request) {
	// Get stats
	stats, err := h.service.GetModerationStats(r.Context())
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to get moderation stats")
		return
	}

	sendJSON(w, http.StatusOK, stats)
}

// GetAnalytics handles GET /api/v1/admin/reviews/analytics
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// Get analytics
	analytics, err := h.service.GetAnalytics(r.Context())
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to get analytics")
		return
	}

	sendJSON(w, http.StatusOK, analytics)
}

// === HOTEL PARTNER ENDPOINTS ===

// AddHotelResponse handles POST /api/v1/hotels/:hotelId/reviews/:reviewId/response
func (h *Handler) AddHotelResponse(w http.ResponseWriter, r *http.Request) {
	// Extract review ID from path
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-2]

	var req AddHotelResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Add response
	err := h.service.AddHotelResponse(r.Context(), reviewID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "must be at least") || strings.Contains(err.Error(), "must not exceed") {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to add response")
		return
	}

	sendJSON(w, http.StatusCreated, map[string]string{
		"message": "Hotel response added successfully",
	})
}

// UpdateHotelResponse handles PUT /api/v1/hotels/:hotelId/reviews/:reviewId/response
func (h *Handler) UpdateHotelResponse(w http.ResponseWriter, r *http.Request) {
	// Extract review ID from path
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-2]

	var req AddHotelResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update response
	err := h.service.UpdateHotelResponse(r.Context(), reviewID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "must be at least") || strings.Contains(err.Error(), "must not exceed") {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, "Failed to update response")
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{
		"message": "Hotel response updated successfully",
	})
}

// DeleteHotelResponse handles DELETE /api/v1/hotels/:hotelId/reviews/:reviewId/response
func (h *Handler) DeleteHotelResponse(w http.ResponseWriter, r *http.Request) {
	// Extract review ID from path
	parts := strings.Split(r.URL.Path, "/")
	reviewID := parts[len(parts)-2]

	// Delete response
	err := h.service.DeleteHotelResponse(r.Context(), reviewID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to delete response")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

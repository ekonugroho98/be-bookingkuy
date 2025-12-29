package search

import (
	"encoding/json"
	"net/http"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Handler handles HTTP requests for search operations
type Handler struct {
	service Service
}

// NewHandler creates a new search handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SearchHotels handles POST /search/hotels
func (h *Handler) SearchHotels(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse search options from query params
	opts := &SearchOptions{
		Page:    1,
		PerPage: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := parseQueryParamInt(page); err == nil {
			opts.Page = p
		}
	}

	if perPage := r.URL.Query().Get("per_page"); perPage != "" {
		if p, err := parseQueryParamInt(perPage); err == nil {
			opts.PerPage = p
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		opts.SortBy = SortBy(sortBy)
	}

	result, err := h.service.SearchHotels(r.Context(), &req, opts)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to search hotels")
		if err == ErrInvalidDates {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// Autocomplete handles GET /search/autocomplete
func (h *Handler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	// Parse limit from query params (default 10, max 20)
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := parseQueryParamInt(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	result, err := h.service.Autocomplete(r.Context(), query, limit)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get autocomplete results")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// GetPopularDestinations handles GET /search/destinations
func (h *Handler) GetPopularDestinations(w http.ResponseWriter, r *http.Request) {
	// Parse limit from query params (default 10, max 20)
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := parseQueryParamInt(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	results, err := h.service.GetPopularDestinations(r.Context(), limit)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get popular destinations")
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"destinations": results,
	})
}

func parseQueryParamInt(s string) (int, error) {
	var i int
	err := json.Unmarshal([]byte(s), &i)
	return i, err
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, map[string]string{"error": message})
}

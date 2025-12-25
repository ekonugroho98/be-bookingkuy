package search

import (
	"encoding/json"
	"errors"
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
		if errors.Is(err, errors.New("check-out date must be after check-in date")) {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, result)
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

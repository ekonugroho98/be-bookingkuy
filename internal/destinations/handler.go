package destinations

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Handler handles destination-related HTTP requests
type Handler struct {
	db *pgxpool.Pool
}

// NewHandler creates a new destinations handler
func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		db: db,
	}
}

// AutocompleteResponse represents the autocomplete response
type AutocompleteResponse struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	CountryCode string   `json:"country_code"`
	CountryName string   `json:"country_name"`
	Type        string   `json:"type"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

// AutocompleteRequest represents the autocomplete request parameters
type AutocompleteRequest struct {
	Query string
	Limit int
}

// Autocomplete handles destination autocomplete requests
// GET /api/v1/destinations/autocomplete?q={query}&limit={limit}
func (h *Handler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	// Get limit from query params, default to 10
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	// Use PostgreSQL full-text search for autocomplete with priority to exact prefix matches
	searchQuery := `
		SELECT
			code,
			name,
			country_code,
			country_name,
			type,
			latitude,
			longitude
		FROM destinations
		WHERE name ILIKE $1
		ORDER BY
			CASE WHEN name ILIKE $2 THEN 0 ELSE 1 END,
			name
		LIMIT $3
	`

	rows, err := h.db.Query(r.Context(), searchQuery, "%"+query+"%", query+"%", limit)
	if err != nil {
		logger.Errorf("Failed to search destinations: %v", err)
		http.Error(w, `{"error":"failed to search destinations"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []AutocompleteResponse
	for rows.Next() {
		var dest AutocompleteResponse
		err := rows.Scan(
			&dest.Code,
			&dest.Name,
			&dest.CountryCode,
			&dest.CountryName,
			&dest.Type,
			&dest.Latitude,
			&dest.Longitude,
		)
		if err != nil {
			logger.Errorf("Failed to scan destination row: %v", err)
			continue
		}
		results = append(results, dest)
	}

	if err = rows.Err(); err != nil {
		logger.Errorf("Error iterating destination rows: %v", err)
		http.Error(w, `{"error":"failed to process results"}`, http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"data": results,
		"meta": map[string]interface{}{
			"query": query,
			"count": len(results),
			"limit": limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

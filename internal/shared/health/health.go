package health

import (
	"encoding/json"
	"net/http"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
)

// Handler handles health check requests
type Handler struct {
	db *db.DB
}

// NewHandler creates a new health check handler
func NewHandler(database *db.DB) *Handler {
	return &Handler{
		db: database,
	}
}

// CheckResponse represents health check response
type CheckResponse struct {
	Status   string                 `json:"status"`
	Version string                 `json:"version,omitempty"`
	Services map[string]interface{} `json:"services"`
}

// Check handles GET /health
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	response := CheckResponse{
		Status:   "healthy",
		Services: make(map[string]interface{}),
	}

	// Check database
	if h.db != nil {
		dbStatus := h.db.Health(r.Context())
		response.Services["database"] = dbStatus

		if dbStatus["status"] != "healthy" {
			response.Status = "unhealthy"
		}
	}

	// TODO: Add checks for Redis, external services, etc

	// Set status code
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Ready handles GET /health/ready (readiness probe)
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	// For readiness, check if critical services are ready
	if h.db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not_ready","error":"database not initialized"}`))
		return
	}

	if err := h.db.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"not_ready","error":"` + err.Error() + `"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// Live handles GET /health/live (liveness probe)
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

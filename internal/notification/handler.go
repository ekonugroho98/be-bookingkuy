package notification

import (
	"net/http"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Handler handles HTTP requests for notification operations
type Handler struct {
	service *Service
}

// NewHandler creates a new notification handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SendTestEmail sends a test email (for debugging)
func (h *Handler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement test email endpoint
	logger.Info("Test email endpoint called")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Test email endpoint not yet implemented"}`))
}

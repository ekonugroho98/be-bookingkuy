package notification

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// EmailService handles email sending
type EmailService struct {
	apiKey  string
	baseURL string
	from    string
}

// NewEmailService creates a new email service
func NewEmailService(apiKey, baseURL, from string) *EmailService {
	return &EmailService{
		apiKey:  apiKey,
		baseURL: baseURL,
		from:    from,
	}
}

// SendEmail sends an email (mock implementation)
func (e *EmailService) SendEmail(ctx context.Context, to, subject string, data map[string]interface{}) error {
	logger.Infof("Sending email to %s: %s", to, subject)
	// TODO: Implement actual SendGrid/Mailgun API call
	// For now, just log
	logger.Infof("Email data: %+v", data)
	return fmt.Errorf("email service not yet implemented")
}

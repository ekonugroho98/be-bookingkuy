package notification

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// SMSService handles SMS sending
type SMSService struct {
	apiKey  string
	baseURL string
	from    string
}

// NewSMSService creates a new SMS service
func NewSMSService(apiKey, baseURL, from string) *SMSService {
	return &SMSService{
		apiKey:  apiKey,
		baseURL: baseURL,
		from:    from,
	}
}

// SendSMS sends an SMS (mock implementation)
func (s *SMSService) SendSMS(ctx context.Context, to, message string) error {
	logger.Infof("Sending SMS to %s: %s", to, message)
	// TODO: Implement actual Twilio/Nexmo API call
	// For now, just log
	return fmt.Errorf("SMS service not yet implemented")
}

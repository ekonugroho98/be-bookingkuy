package notification

import (
	"context"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service handles notifications
type Service struct {
	emailService *EmailService
	smsService  *SMSService
}

// NewService creates a new notification service
func NewService(emailService *EmailService, smsService *SMSService) *Service {
	return &Service{
		emailService: emailService,
		smsService:  smsService,
	}
}

// SendBookingConfirmation sends booking confirmation email
func (s *Service) SendBookingConfirmation(ctx context.Context, email, name string, bookingDetails map[string]interface{}) error {
	logger.Infof("Sending booking confirmation to %s", email)
	return s.emailService.SendEmail(ctx, email, "Booking Confirmation", bookingDetails)
}

// SendPaymentConfirmation sends payment confirmation email
func (s *Service) SendPaymentConfirmation(ctx context.Context, email, name string, paymentDetails map[string]interface{}) error {
	logger.Infof("Sending payment confirmation to %s", email)
	return s.emailService.SendEmail(ctx, email, "Payment Confirmation", paymentDetails)
}

// SendBookingCancelled sends booking cancellation email
func (s *Service) SendBookingCancelled(ctx context.Context, email, name string, cancellationDetails map[string]interface{}) error {
	logger.Infof("Sending cancellation notice to %s", email)
	return s.emailService.SendEmail(ctx, email, "Booking Cancelled", cancellationDetails)
}

// SendOTPSMS sends OTP via SMS
func (s *Service) SendOTPSMS(ctx context.Context, phone, otp string) error {
	logger.Infof("Sending OTP to %s", phone)
	return s.smsService.SendSMS(ctx, phone, "Your OTP is: "+otp)
}

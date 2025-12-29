package notification

import (
	"context"

	"github.com/ekonugroho98/be-bookingkuy/internal/queue"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service handles notifications
type Service struct {
	emailService *EmailService
	smsService  *SMSService
	queueClient *queue.RabbitMQClient
}

// NewService creates a new notification service
func NewService(emailService *EmailService, smsService *SMSService) *Service {
	return &Service{
		emailService: emailService,
		smsService:  smsService,
	}
}

// SetQueueClient sets the queue client for async notifications
func (s *Service) SetQueueClient(qc *queue.RabbitMQClient) {
	s.queueClient = qc
}

// SendBookingConfirmation sends booking confirmation email
func (s *Service) SendBookingConfirmation(ctx context.Context, email, name string, bookingDetails map[string]interface{}) error {
	logger.Infof("Sending booking confirmation to %s", email)

	// If queue is available, publish to queue for async processing
	if s.queueClient != nil && s.queueClient.IsConnected() {
		message := queue.Message{
			Type:    "booking_confirmation",
			Payload: map[string]interface{}{
				"email":          email,
				"name":           name,
				"booking_details": bookingDetails,
			},
		}

		if err := s.queueClient.Publish(ctx, queue.QueueEmail, message); err != nil {
			logger.ErrorWithErr(err, "Failed to publish email to queue, sending synchronously")
			// Fallback to synchronous sending
			return s.emailService.SendBookingConfirmationEmail(ctx, email, name, bookingDetails)
		}

		logger.Infof("Booking confirmation queued for %s", email)
		return nil
	}

	// Synchronous fallback
	return s.emailService.SendBookingConfirmationEmail(ctx, email, name, bookingDetails)
}

// SendPaymentConfirmation sends payment confirmation email
func (s *Service) SendPaymentConfirmation(ctx context.Context, email, name string, paymentDetails map[string]interface{}) error {
	logger.Infof("Sending payment confirmation to %s", email)

	// If queue is available, publish to queue
	if s.queueClient != nil && s.queueClient.IsConnected() {
		message := queue.Message{
			Type:    "payment_confirmation",
			Payload: map[string]interface{}{
				"email":           email,
				"name":            name,
				"payment_details": paymentDetails,
			},
		}

		if err := s.queueClient.Publish(ctx, queue.QueueEmail, message); err != nil {
			logger.ErrorWithErr(err, "Failed to publish email to queue, sending synchronously")
			return s.emailService.SendPaymentConfirmationEmail(ctx, email, name, paymentDetails)
		}

		logger.Infof("Payment confirmation queued for %s", email)
		return nil
	}

	return s.emailService.SendPaymentConfirmationEmail(ctx, email, name, paymentDetails)
}

// SendBookingCancelled sends booking cancellation email
func (s *Service) SendBookingCancelled(ctx context.Context, email, name string, cancellationDetails map[string]interface{}) error {
	logger.Infof("Sending cancellation notice to %s", email)

	if s.queueClient != nil && s.queueClient.IsConnected() {
		message := queue.Message{
			Type:    "booking_cancelled",
			Payload: map[string]interface{}{
				"email":                email,
				"name":                 name,
				"cancellation_details": cancellationDetails,
			},
		}

		if err := s.queueClient.Publish(ctx, queue.QueueEmail, message); err != nil {
			logger.ErrorWithErr(err, "Failed to publish email to queue, sending synchronously")
			return s.emailService.SendEmail(ctx, email, "Booking Cancelled", cancellationDetails)
		}

		logger.Infof("Cancellation notice queued for %s", email)
		return nil
	}

	return s.emailService.SendEmail(ctx, email, "Booking Cancelled", cancellationDetails)
}

// SendOTPSMS sends OTP via SMS
func (s *Service) SendOTPSMS(ctx context.Context, phone, otp string) error {
	logger.Infof("Sending OTP to %s", phone)

	if s.queueClient != nil && s.queueClient.IsConnected() {
		message := queue.Message{
			Type: "otp_sms",
			Payload: map[string]interface{}{
				"phone": phone,
				"otp":   otp,
			},
		}

		if err := s.queueClient.Publish(ctx, queue.QueueSMS, message); err != nil {
			logger.ErrorWithErr(err, "Failed to publish SMS to queue, sending synchronously")
			return s.smsService.SendSMS(ctx, phone, "Your OTP is: "+otp)
		}

		logger.Infof("OTP SMS queued for %s", phone)
		return nil
	}

	return s.smsService.SendSMS(ctx, phone, "Your OTP is: "+otp)
}

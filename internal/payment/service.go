package payment

import (
	"context"
	"errors"

	"github.com/ekonugroho98/be-bookingkuy/internal/midtrans"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines interface for payment business logic
type Service interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest, amount int) (*Payment, error)
	HandleWebhook(ctx context.Context, payload *WebhookPayload) error
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
}

type service struct {
	repo           Repository
	eventBus       eventbus.EventBus
	midtransClient *midtrans.Client
	midtransMapper *midtrans.Mapper
}

// NewService creates a new payment service
func NewService(repo Repository, eb eventbus.EventBus) Service {
	return &service{
		repo:           repo,
		eventBus:       eb,
		midtransMapper: midtrans.NewMapper(),
	}
}

// NewServiceWithMidtrans creates a new payment service with Midtrans client
func NewServiceWithMidtrans(repo Repository, eb eventbus.EventBus, midtransClient *midtrans.Client) Service {
	return &service{
		repo:           repo,
		eventBus:       eb,
		midtransClient: midtransClient,
		midtransMapper: midtrans.NewMapper(),
	}
}

func (s *service) CreatePayment(ctx context.Context, req *CreatePaymentRequest, amount int) (*Payment, error) {
	// Check if payment already exists for this booking
	existingPayment, err := s.repo.GetByBookingID(ctx, req.BookingID)
	if err == nil && existingPayment != nil {
		// If payment exists and is still pending, return it
		if existingPayment.Status == StatusPending {
			logger.Infof("Pending payment already exists for booking %s", req.BookingID)
			return existingPayment, nil
		}
		// If payment is completed, return error
		return nil, errors.New("payment already completed for this booking")
	}

	// Create new payment
	payment := NewPayment(req.BookingID, req, amount)

	// Call payment provider based on provider type
	if req.Provider == ProviderMidtrans && s.midtransClient != nil {
		// Create Midtrans charge request
		chargeReq := s.midtransMapper.ToChargeRequestWithPaymentType(
			&midtrans.PaymentInput{
				OrderID:   payment.ID, // Use payment ID as order ID
				BookingID: payment.BookingID,
				Amount:    payment.Amount,
			},
			&midtrans.CustomerDetails{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "customer@example.com",
				Phone:     "+628123456789",
			},
			midtrans.PaymentType(req.Method),
		)

		// Call Midtrans API
		chargeResp, err := s.midtransClient.Charge(chargeReq)
		if err != nil {
			logger.ErrorWithErr(err, "Failed to charge Midtrans")
			return nil, errors.New("failed to create payment with provider")
		}

		// Update payment with Midtrans response
		if chargeResp.TransactionID != "" {
			payment.ProviderRef = chargeResp.TransactionID
		}

		// Update payment URL
		if chargeResp.RedirectURL != "" {
			payment.PaymentURL = chargeResp.RedirectURL
		} else if chargeResp.PaymentURL != "" {
			payment.PaymentURL = chargeResp.PaymentURL
		}

		// Update payment method
		payment.Method = chargeResp.PaymentType
	} else {
		// Fallback to mock implementation
		payment.PaymentURL = s.generatePaymentURL(payment)
		logger.Warnf("Payment provider %s not implemented, using mock", req.Provider)
	}

	// Save payment
	if err := s.repo.Create(ctx, payment); err != nil {
		logger.ErrorWithErr(err, "Failed to create payment")
		return nil, errors.New("failed to create payment")
	}

	logger.Infof("Payment created: %s for booking %s", payment.ID, payment.BookingID)
	return payment, nil
}

func (s *service) HandleWebhook(ctx context.Context, payload *WebhookPayload) error {
	// For Midtrans webhooks, we need to find payment by OrderID first
	var payment *Payment
	var err error

	if payload.OrderID != "" && s.midtransClient != nil {
		// Midtrans webhook - find payment by provider reference
		payment, err = s.repo.GetByProviderRef(ctx, payload.OrderID)
		if err != nil {
			logger.ErrorWithErr(err, "Failed to get payment by provider ref")
			return err
		}

		// Validate Midtrans signature
		if !s.midtransClient.ValidateWebhookSignature(
			payload.OrderID,
			payload.StatusCode,
			payload.GrossAmount,
			payload.SignatureKey,
		) {
			logger.Error("Invalid Midtrans webhook signature")
			return errors.New("invalid webhook signature")
		}

		// Map Midtrans status to internal status
		newStatus := PaymentStatus(midtrans.MapTransactionStatus(
			midtrans.TransactionStatus(payload.TransactionStatus),
		))

		if newStatus == "" {
			logger.Error("Invalid payment status in webhook")
			return errors.New("invalid payment status")
		}

		// Update payment status
		if err := s.repo.UpdateStatus(ctx, payment.ID, newStatus, payload.ProviderRef); err != nil {
			logger.ErrorWithErr(err, "Failed to update payment status")
			return err
		}

		// Publish payment event
		if err := s.publishPaymentEvent(ctx, payment, newStatus); err != nil {
			logger.ErrorWithErr(err, "Failed to publish payment event")
		}

		logger.Infof("Payment %s updated to status: %s (from Midtrans)", payment.ID, newStatus)
		return nil
	}

	// Fallback to generic webhook handling
	payment, err = s.repo.GetByID(ctx, payload.PaymentID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get payment")
		return err
	}

	// Validate webhook signature (mock implementation)
	if !s.validateSignature(payload) {
		logger.Error("Invalid webhook signature")
		return errors.New("invalid webhook signature")
	}

	// Update payment status
	newStatus := map[string]PaymentStatus{
		"success": StatusSuccess,
		"failed":  StatusFailed,
		"refund":  StatusRefunded,
	}[payload.Status]

	if newStatus == "" {
		logger.Error("Invalid payment status in webhook")
		return errors.New("invalid payment status")
	}

	if err := s.repo.UpdateStatus(ctx, payment.ID, newStatus, payload.ProviderRef); err != nil {
		logger.ErrorWithErr(err, "Failed to update payment status")
		return err
	}

	// Publish payment event
	if err := s.publishPaymentEvent(ctx, payment, newStatus); err != nil {
		logger.ErrorWithErr(err, "Failed to publish payment event")
	}

	logger.Infof("Payment %s updated to status: %s", payment.ID, newStatus)
	return nil
}

func (s *service) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	return s.repo.GetByID(ctx, paymentID)
}

func (s *service) generatePaymentURL(payment *Payment) string {
	// Mock implementation - in real scenario, this would call payment gateway API
	return "https://payment-gateway.example.com/pay/" + payment.ID
}

func (s *service) validateSignature(payload *WebhookPayload) bool {
	// Mock implementation - in real scenario, validate webhook signature
	return payload.Signature != ""
}

func (s *service) publishPaymentEvent(ctx context.Context, payment *Payment, status PaymentStatus) error {
	var eventType string
	switch status {
	case StatusSuccess:
		eventType = eventbus.EventPaymentSuccess
	case StatusFailed:
		eventType = eventbus.EventPaymentFailed
	case StatusRefunded:
		eventType = eventbus.EventPaymentRefunded
	default:
		return nil
	}

	return s.eventBus.Publish(ctx, eventType, map[string]interface{}{
		"payment_id": payment.ID,
		"booking_id": payment.BookingID,
		"amount":     payment.Amount,
		"currency":   payment.Currency,
		"status":     string(status),
		"provider":   string(payment.Provider),
	})
}

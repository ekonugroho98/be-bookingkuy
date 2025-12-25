package payment

import (
	"context"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// HandlePaymentSuccess handles payment success event
func HandlePaymentSuccess(ctx context.Context, event eventbus.Event) error {
	paymentID, ok := event.Payload["payment_id"].(string)
	if !ok {
		logger.Error("Invalid payment ID in event payload")
		return nil
	}

	bookingID, _ := event.Payload["booking_id"].(string)
	amount, _ := event.Payload["amount"].(int)

	logger.Infof("Payment success event received: %s for booking %s, amount: %d", paymentID, bookingID, amount)

	// TODO: Update booking status to PAID
	// TODO: Trigger booking confirmation flow

	return nil
}

// HandlePaymentFailed handles payment failed event
func HandlePaymentFailed(ctx context.Context, event eventbus.Event) error {
	paymentID, _ := event.Payload["payment_id"].(string)
	bookingID, _ := event.Payload["booking_id"].(string)

	logger.Infof("Payment failed event received: %s for booking %s", paymentID, bookingID)

	// TODO: Update booking status if needed
	// TODO: Send payment failed notification

	return nil
}

// HandlePaymentRefunded handles payment refunded event
func HandlePaymentRefunded(ctx context.Context, event eventbus.Event) error {
	paymentID, _ := event.Payload["payment_id"].(string)
	bookingID, _ := event.Payload["booking_id"].(string)

	logger.Infof("Payment refunded event received: %s for booking %s", paymentID, bookingID)

	// TODO: Process cancellation
	// TODO: Send refund confirmation email

	return nil
}

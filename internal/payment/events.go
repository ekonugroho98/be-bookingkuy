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

	// 1. Update booking status to PAID
	// This is typically handled by the payment webhook handler
	// But we can also trigger it from here if needed
	logger.Infof("💳 Payment successful - booking %s should be marked as PAID", bookingID)

	// 2. Send payment success notification
	if err := sendPaymentSuccessNotification(ctx, event.Payload); err != nil {
		logger.ErrorWithErr(err, "Failed to send payment success notification")
		// Log but don't fail
	}

	logger.Infof("Payment success event processed: %s", paymentID)
	return nil
}

// sendPaymentSuccessNotification sends payment success notification
func sendPaymentSuccessNotification(ctx context.Context, data map[string]interface{}) error {
	paymentID := data["payment_id"].(string)
	bookingID := data["booking_id"].(string)
	amount := data["amount"].(int)

	logger.Infof("📧 Sending payment success notification for payment %s", paymentID)
	logger.Infof("   - Booking ID: %s", bookingID)
	logger.Infof("   - Amount: %d", amount)

	// TODO: Send actual email/SMS notification
	return nil
}

// HandlePaymentFailed handles payment failed event
func HandlePaymentFailed(ctx context.Context, event eventbus.Event) error {
	paymentID, _ := event.Payload["payment_id"].(string)
	bookingID, _ := event.Payload["booking_id"].(string)

	logger.Infof("Payment failed event received: %s for booking %s", paymentID, bookingID)

	// 1. Update booking status if needed
	// If payment failed, booking might need to be cancelled or marked as failed
	logger.Infof("❌ Payment failed - booking %s may need attention", bookingID)

	// 2. Send payment failed notification
	if err := sendPaymentFailedNotification(ctx, event.Payload); err != nil {
		logger.ErrorWithErr(err, "Failed to send payment failed notification")
		// Log but don't fail
	}

	logger.Infof("Payment failed event processed: %s", paymentID)
	return nil
}

// sendPaymentFailedNotification sends payment failed notification
func sendPaymentFailedNotification(ctx context.Context, data map[string]interface{}) error {
	paymentID := data["payment_id"].(string)
	bookingID := data["booking_id"].(string)

	logger.Infof("📧 Sending payment failed notification for payment %s", paymentID)
	logger.Infof("   - Booking ID: %s", bookingID)
	logger.Infof("   - Action: User should retry payment")

	// TODO: Send actual email/SMS notification with retry link
	return nil
}

// HandlePaymentRefunded handles payment refunded event
func HandlePaymentRefunded(ctx context.Context, event eventbus.Event) error {
	paymentID, _ := event.Payload["payment_id"].(string)
	bookingID, _ := event.Payload["booking_id"].(string)
	amount, _ := event.Payload["amount"].(int)

	logger.Infof("Payment refunded event received: %s for booking %s, amount: %d", paymentID, bookingID, amount)

	// 1. Log refund completion
	logger.Infof("💰 Refund completed for payment %s, amount: %d", paymentID, amount)

	// 2. Send refund confirmation
	if err := sendRefundConfirmationNotification(ctx, event.Payload); err != nil {
		logger.ErrorWithErr(err, "Failed to send refund confirmation notification")
		// Log but don't fail
	}

	logger.Infof("Payment refunded event processed: %s", paymentID)
	return nil
}

// sendRefundConfirmationNotification sends refund confirmation
func sendRefundConfirmationNotification(ctx context.Context, data map[string]interface{}) error {
	paymentID := data["payment_id"].(string)
	bookingID := data["booking_id"].(string)
	amount := data["amount"].(int)

	logger.Infof("📧 Sending refund confirmation notification for payment %s", paymentID)
	logger.Infof("   - Booking ID: %s", bookingID)
	logger.Infof("   - Refund Amount: %d", amount)
	logger.Infof("   - Timeline: 5-7 business days")

	// TODO: Send actual email notification with refund details
	return nil
}

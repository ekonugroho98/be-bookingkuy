package booking

import (
	"context"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// HandleBookingCreated handles booking created event
func HandleBookingCreated(ctx context.Context, event eventbus.Event) error {
	bookingID, ok := event.Payload["booking_id"].(string)
	if !ok {
		logger.Error("Invalid booking ID in event payload")
		return nil
	}

	userID, _ := event.Payload["user_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)
	amount, _ := event.Payload["total_amount"].(int)
	currency, _ := event.Payload["currency"].(string)

	logger.Infof("Booking created event received: %s (%s) - User: %s, Amount: %d %s",
		bookingID, bookingRef, userID, amount, currency)

	// TODO: Trigger payment flow for PAY_NOW bookings
	// TODO: Send confirmation email

	return nil
}

// HandleBookingPaid handles booking paid event
func HandleBookingPaid(ctx context.Context, event eventbus.Event) error {
	bookingID, _ := event.Payload["booking_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)

	logger.Infof("Booking paid event received: %s (%s)", bookingID, bookingRef)

	// TODO: Confirm booking with supplier
	// TODO: Send payment confirmation email

	return nil
}

// HandleBookingConfirmed handles booking confirmed event
func HandleBookingConfirmed(ctx context.Context, event eventbus.Event) error {
	bookingID, _ := event.Payload["booking_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)

	logger.Infof("Booking confirmed event received: %s (%s)", bookingID, bookingRef)

	// TODO: Send booking confirmation email with details

	return nil
}

// HandleBookingCancelled handles booking cancelled event
func HandleBookingCancelled(ctx context.Context, event eventbus.Event) error {
	bookingID, _ := event.Payload["booking_id"].(string)
	bookingRef, _ := event.Payload["booking_reference"].(string)

	logger.Infof("Booking cancelled event received: %s (%s)", bookingID, bookingRef)

	// TODO: Process refund if applicable
	// TODO: Send cancellation email

	return nil
}

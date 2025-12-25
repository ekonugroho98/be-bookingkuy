package notification

import (
	"context"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// HandleBookingCreated handles booking created event
func HandleBookingCreated(ctx context.Context, event eventbus.Event) error {
	logger.Info("Handling booking created event")
	// Send booking confirmation email
	// Extract user email and booking details from event.Payload
	logger.Infof("Booking created: %+v", event.Payload)
	return nil
}

// HandleBookingPaid handles booking paid event
func HandleBookingPaid(ctx context.Context, event eventbus.Event) error {
	logger.Info("Handling booking paid event")
	// Send payment confirmation email
	logger.Infof("Booking paid: %+v", event.Payload)
	return nil
}

// HandleBookingConfirmed handles booking confirmed event
func HandleBookingConfirmed(ctx context.Context, event eventbus.Event) error {
	logger.Info("Handling booking confirmed event")
	// Send booking confirmation with voucher
	logger.Infof("Booking confirmed: %+v", event.Payload)
	return nil
}

// HandleBookingCancelled handles booking cancelled event
func HandleBookingCancelled(ctx context.Context, event eventbus.Event) error {
	logger.Info("Handling booking cancelled event")
	// Send cancellation notification
	logger.Infof("Booking cancelled: %+v", event.Payload)
	return nil
}

// RegisterEventHandlers registers notification event handlers
func RegisterEventHandlers(eb eventbus.EventBus, service *Service) {
	eb.Subscribe(context.Background(), eventbus.EventBookingCreated, HandleBookingCreated)
	eb.Subscribe(context.Background(), eventbus.EventBookingPaid, HandleBookingPaid)
	eb.Subscribe(context.Background(), eventbus.EventBookingConfirmed, HandleBookingConfirmed)
	eb.Subscribe(context.Background(), eventbus.EventBookingCancelled, HandleBookingCancelled)
}

package booking

import (
	"context"
	"errors"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/pricing"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines interface for booking business logic
type Service interface {
	CreateBooking(ctx context.Context, userID string, req *CreateBookingRequest) (*Booking, error)
	GetBooking(ctx context.Context, bookingID string) (*Booking, error)
	GetUserBookings(ctx context.Context, userID string, page, perPage int) ([]*Booking, error)
	UpdateStatus(ctx context.Context, bookingID string, status BookingStatus) (*Booking, error)
	CancelBooking(ctx context.Context, bookingID string) (*Booking, error)
}

type service struct {
	repo          Repository
	eventBus      eventbus.EventBus
	pricingService pricing.Service
}

// NewService creates a new booking service
func NewService(repo Repository, eb eventbus.EventBus, ps pricing.Service) Service {
	return &service{
		repo:          repo,
		eventBus:      eb,
		pricingService: ps,
	}
}

func (s *service) CreateBooking(ctx context.Context, userID string, req *CreateBookingRequest) (*Booking, error) {
	// Create booking
	booking := NewBooking(userID, req)

	// TODO: Get hotel and room details to calculate price
	// For now, set a default price
	booking.TotalAmount = 1000000 // 1 million IDR default

	// Save booking
	if err := s.repo.Create(ctx, booking); err != nil {
		logger.ErrorWithErr(err, "Failed to create booking")
		return nil, errors.New("failed to create booking")
	}

	// Transition to AWAITING_PAYMENT
	sm := NewStateMachine(booking)
	if err := sm.Transition(StatusAwaitingPayment); err != nil {
		logger.ErrorWithErr(err, "Failed to transition booking state")
		return nil, err
	}

	// Update status in database
	if err := s.repo.UpdateStatus(ctx, booking.ID, booking.Status); err != nil {
		logger.ErrorWithErr(err, "Failed to update booking status")
		return nil, errors.New("failed to update booking status")
	}

	// Publish booking.created event
	if err := s.eventBus.Publish(ctx, eventbus.EventBookingCreated, map[string]interface{}{
		"booking_id":        booking.ID,
		"user_id":           booking.UserID,
		"hotel_id":          booking.HotelID,
		"room_id":           booking.RoomID,
		"booking_reference": booking.BookingReference,
		"total_amount":      booking.TotalAmount,
		"currency":          booking.Currency,
		"status":            string(booking.Status),
	}); err != nil {
		logger.ErrorWithErr(err, "Failed to publish booking.created event")
	}

	logger.Infof("Booking created: %s (%s)", booking.ID, booking.BookingReference)
	return booking, nil
}

func (s *service) GetBooking(ctx context.Context, bookingID string) (*Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking")
		return nil, err
	}
	return booking, nil
}

func (s *service) GetUserBookings(ctx context.Context, userID string, page, perPage int) ([]*Booking, error) {
	offset := (page - 1) * perPage
	bookings, err := s.repo.GetByUserID(ctx, userID, perPage, offset)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get user bookings")
		return nil, err
	}
	return bookings, nil
}

func (s *service) UpdateStatus(ctx context.Context, bookingID string, status BookingStatus) (*Booking, error) {
	// Get booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking")
		return nil, err
	}

	// Transition state
	sm := NewStateMachine(booking)
	if err := sm.Transition(status); err != nil {
		logger.ErrorWithErr(err, "Invalid state transition")
		return nil, err
	}

	// Update status in database
	if err := s.repo.UpdateStatus(ctx, bookingID, status); err != nil {
		logger.ErrorWithErr(err, "Failed to update booking status")
		return nil, errors.New("failed to update booking status")
	}

	// Publish event based on new status
	if err := s.publishStatusEvent(ctx, booking, status); err != nil {
		logger.ErrorWithErr(err, fmt.Sprintf("Failed to publish booking.%s event", status))
	}

	logger.Infof("Booking %s updated to status: %s", bookingID, status)
	return booking, nil
}

func (s *service) CancelBooking(ctx context.Context, bookingID string) (*Booking, error) {
	return s.UpdateStatus(ctx, bookingID, StatusCancelled)
}

func (s *service) publishStatusEvent(ctx context.Context, booking *Booking, status BookingStatus) error {
	var eventType string
	switch status {
	case StatusPaid:
		eventType = eventbus.EventBookingPaid
	case StatusConfirmed:
		eventType = eventbus.EventBookingConfirmed
	case StatusCancelled:
		eventType = eventbus.EventBookingCancelled
	default:
		return nil
	}

	return s.eventBus.Publish(ctx, eventType, map[string]interface{}{
		"booking_id":        booking.ID,
		"user_id":           booking.UserID,
		"booking_reference": booking.BookingReference,
		"status":            string(status),
	})
}

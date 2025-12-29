package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/hotelbeds"
	"github.com/ekonugroho98/be-bookingkuy/internal/pricing"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines interface for booking business logic
type Service interface {
	CreateBooking(ctx context.Context, userID string, req *CreateBookingRequest) (*Booking, error)
	GetBooking(ctx context.Context, bookingID string) (*Booking, error)
	GetBookingWithDetails(ctx context.Context, bookingID string) (*BookingResponse, error)
	GetUserBookings(ctx context.Context, userID string, page, perPage int) ([]*Booking, error)
	GetUserBookingsWithDetails(ctx context.Context, userID string, page, perPage int) ([]*BookingResponse, error)
	UpdateBooking(ctx context.Context, bookingID string, req *UpdateBookingRequest) (*Booking, error)
	UpdateStatus(ctx context.Context, bookingID string, status BookingStatus) (*Booking, error)
	CancelBooking(ctx context.Context, bookingID string) (*Booking, error)
	ConfirmBookingWithSupplier(ctx context.Context, bookingID string) (*Booking, error)
}

type service struct {
	repo            Repository
	eventBus        eventbus.EventBus
	pricingService  pricing.Service
	hotelbedsClient hotelbeds.ClientInterface
}

// NewService creates a new booking service
func NewService(repo Repository, eb eventbus.EventBus, ps pricing.Service, hbClient hotelbeds.ClientInterface) Service {
	return &service{
		repo:            repo,
		eventBus:        eb,
		pricingService:  ps,
		hotelbedsClient: hbClient,
	}
}

func (s *service) CreateBooking(ctx context.Context, userID string, req *CreateBookingRequest) (*Booking, error) {
	// 1. Validate dates
	if req.CheckOut.Before(req.CheckIn) {
		return nil, ErrInvalidCheckOut
	}

	if req.CheckIn.Before(time.Now().AddDate(0, 0, -1)) {
		return nil, ErrInvalidCheckIn
	}

	// 2. Check availability with HotelBeds
	availability, err := s.hotelbedsClient.GetHotelAvailability(ctx, &hotelbeds.AvailabilityRequest{
		HotelCode: req.HotelID,
		RoomCode:  req.RoomID,
		CheckIn:   req.CheckIn,
		CheckOut:  req.CheckOut,
		Guests:    req.Guests,
	})
	if err != nil {
		logger.ErrorWithErr(err, "Failed to check availability with HotelBeds")
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	if !availability.IsAvailable || len(availability.Rooms) == 0 {
		return nil, ErrRoomNotAvailable
	}

	// 3. Get room pricing from HotelBeds
	roomRate, err := s.hotelbedsClient.GetRoomRates(ctx, &hotelbeds.RoomRateRequest{
		HotelCode: req.HotelID,
		RoomCode:  req.RoomID,
		CheckIn:   req.CheckIn,
		CheckOut:  req.CheckOut,
		Guests:    req.Guests,
	})
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get room rates from HotelBeds")
		return nil, fmt.Errorf("failed to get room rates: %w", err)
	}

	// 4. Create booking with REAL price from HotelBeds
	booking := NewBooking(userID, req)
	booking.TotalAmount = roomRate.TotalPrice
	booking.Currency = roomRate.Currency

	// 5. Save booking
	if err := s.repo.Create(ctx, booking); err != nil {
		logger.ErrorWithErr(err, "Failed to create booking")
		return nil, ErrFailedToCreate
	}

	// 6. Transition to AWAITING_PAYMENT
	sm := NewStateMachine(booking)
	if err := sm.Transition(StatusAwaitingPayment); err != nil {
		logger.ErrorWithErr(err, "Failed to transition booking state")
		return nil, err
	}

	// 7. Update status in database
	if err := s.repo.UpdateStatus(ctx, booking.ID, booking.Status); err != nil {
		logger.ErrorWithErr(err, "Failed to update booking status")
		return nil, ErrFailedToUpdateStatus
	}

	// 8. Publish booking.created event
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

	logger.Infof("Booking created: %s (%s) - Amount: %d %s",
		booking.ID, booking.BookingReference, booking.TotalAmount, booking.Currency)
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

// GetBookingWithDetails returns booking with hotel and room details for FE
func (s *service) GetBookingWithDetails(ctx context.Context, bookingID string) (*BookingResponse, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking")
		return nil, err
	}

	// Fetch hotel details from HotelBeds
	hotel := &HotelDetails{}
	if s.hotelbedsClient != nil {
		hotelData, err := s.hotelbedsClient.GetHotelDetails(ctx, booking.HotelID)
		if err == nil && hotelData != nil {
			hotel.ID = hotelData.HotelCode
			hotel.Name = hotelData.HotelName
			hotel.City = hotelData.CityCode
			hotel.Country = hotelData.CountryCode
			// Get first image if available
			if len(hotelData.Images) > 0 {
				hotel.Image = hotelData.Images[0].URL
			}
			hotel.Rating = hotelData.Rating
			hotel.Description = hotelData.Description
		}
	}

	// For now, use basic room info from RoomID
	// In real implementation, we would fetch room details from HotelBeds
	room := &RoomDetails{
		ID:   booking.RoomID,
		Name: "Standard Room", // Default fallback
	}

	return ToBookingResponse(booking, hotel, room), nil
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

// GetUserBookingsWithDetails returns user bookings with hotel and room details for FE
func (s *service) GetUserBookingsWithDetails(ctx context.Context, userID string, page, perPage int) ([]*BookingResponse, error) {
	offset := (page - 1) * perPage
	bookings, err := s.repo.GetByUserID(ctx, userID, perPage, offset)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get user bookings")
		return nil, err
	}

	// Fetch unique hotel IDs
	hotelIDs := make(map[string]bool)
	for _, b := range bookings {
		hotelIDs[b.HotelID] = true
	}

	// Fetch all hotels
	hotels := make(map[string]*HotelDetails)
	if s.hotelbedsClient != nil {
		for hotelID := range hotelIDs {
			hotelData, err := s.hotelbedsClient.GetHotelDetails(ctx, hotelID)
			if err == nil && hotelData != nil {
				// Get first image if available
				imageURL := ""
				if len(hotelData.Images) > 0 {
					imageURL = hotelData.Images[0].URL
				}
				hotels[hotelID] = &HotelDetails{
					ID:          hotelData.HotelCode,
					Name:        hotelData.HotelName,
					City:        hotelData.CityCode,
					Country:     hotelData.CountryCode,
					Image:       imageURL,
					Rating:      hotelData.Rating,
					Description: hotelData.Description,
				}
			}
		}
	}

	// For rooms, use basic info for now
	rooms := make(map[string]*RoomDetails)
	for _, b := range bookings {
		if _, exists := rooms[b.RoomID]; !exists {
			rooms[b.RoomID] = &RoomDetails{
				ID:   b.RoomID,
				Name: "Standard Room", // Default fallback
			}
		}
	}

	return ToBookingResponseList(bookings, hotels, rooms), nil
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
	// 1. Get booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking")
		return nil, err
	}

	// 2. Cancel with HotelBeds supplier if already confirmed
	if booking.SupplierReference != "" {
		err := s.hotelbedsClient.CancelBooking(ctx, booking.SupplierReference)
		if err != nil {
			logger.ErrorWithErr(err, "Failed to cancel booking with HotelBeds supplier")
			// Log error but continue - we still want to mark as cancelled locally
			// This allows for manual reconciliation later
		} else {
			logger.Infof("Booking cancelled with HotelBeds supplier: %s", booking.SupplierReference)
		}
	}

	// 3. Update status to CANCELLED
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

// ConfirmBookingWithSupplier confirms booking with HotelBeds supplier
func (s *service) ConfirmBookingWithSupplier(ctx context.Context, bookingID string) (*Booking, error) {
	// 1. Get booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking")
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// 2. Validate booking status
	if booking.Status != StatusPaid {
		return nil, fmt.Errorf("booking must be PAID before confirming with supplier, current status: %s", booking.Status)
	}

	// 3. Get user information for holder details
	// TODO: Call user service to get user name and email
	// For now, use placeholder values
	holderName := "Guest" // Should come from user service
	holderEmail := "guest@example.com" // Should come from user service

	// 4. Create booking with HotelBeds
	hbBooking, err := s.hotelbedsClient.CreateBooking(ctx, &hotelbeds.BookingRequest{
		HotelCode: booking.HotelID,
		RoomCode:  booking.RoomID,
		CheckIn:   booking.CheckIn,
		CheckOut:  booking.CheckOut,
		Guests:    booking.Guests,
		Holder: hotelbeds.HolderInfo{
			Name:  holderName,
			Email: holderEmail,
		},
		Payment: hotelbeds.PaymentInfo{
			PaymentMethodType: "CREDITCARD", // Default payment method
		},
	})
	if err != nil {
		logger.ErrorWithErr(err, "Failed to create booking with HotelBeds")
		return nil, fmt.Errorf("failed to confirm with supplier: %w", err)
	}

	// 5. Update booking with supplier reference
	booking.SupplierReference = hbBooking.BookingReference

	// 6. Transition to CONFIRMED
	sm := NewStateMachine(booking)
	if err := sm.Transition(StatusConfirmed); err != nil {
		logger.ErrorWithErr(err, "Failed to transition booking to CONFIRMED")
		return nil, err
	}

	// 7. Update status in database
	if err := s.repo.UpdateStatus(ctx, booking.ID, booking.Status); err != nil {
		logger.ErrorWithErr(err, "Failed to update booking status")
		return nil, ErrFailedToUpdateStatus
	}

	// 8. Publish booking.confirmed event
	if err := s.eventBus.Publish(ctx, eventbus.EventBookingConfirmed, map[string]interface{}{
		"booking_id":         booking.ID,
		"user_id":            booking.UserID,
		"booking_reference":  booking.BookingReference,
		"supplier_reference": hbBooking.BookingReference,
		"status":             string(booking.Status),
	}); err != nil {
		logger.ErrorWithErr(err, "Failed to publish booking.confirmed event")
	}

	logger.Infof("Booking confirmed with supplier: %s - Supplier Ref: %s",
		booking.BookingReference, hbBooking.BookingReference)
	return booking, nil
}

// UpdateBooking updates booking details
func (s *service) UpdateBooking(ctx context.Context, bookingID string, req *UpdateBookingRequest) (*Booking, error) {
	logger.Infof("Updating booking: %s", bookingID)

	// 1. Get existing booking
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get booking")
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// 2. Validate booking status - can only update certain statuses
	if booking.Status != StatusInit && booking.Status != StatusAwaitingPayment {
		return nil, fmt.Errorf("cannot update booking with status: %s", booking.Status)
	}

	// 3. Update fields if provided
	if req.GuestName != "" {
		booking.GuestName = req.GuestName
	}
	if req.GuestEmail != "" {
		booking.GuestEmail = req.GuestEmail
	}
	if req.GuestPhone != "" {
		booking.GuestPhone = req.GuestPhone
	}
	if req.SpecialRequests != "" {
		booking.SpecialRequests = req.SpecialRequests
	}
	if req.PaymentType != "" {
		// Validate payment type
		if req.PaymentType != PaymentTypePayNow && req.PaymentType != PaymentTypePayAtHotel && req.PaymentType != PaymentTypePayLater {
			return nil, ErrInvalidPaymentType
		}
		booking.PaymentType = req.PaymentType
	}

	// 4. Update in database
	if err := s.repo.Update(ctx, booking); err != nil {
		logger.ErrorWithErr(err, "Failed to update booking in database")
		return nil, ErrFailedToUpdate
	}

	logger.Infof("Booking updated successfully: %s", bookingID)
	return booking, nil
}

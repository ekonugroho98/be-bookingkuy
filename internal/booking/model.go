package booking

import (
	"time"

	"github.com/google/uuid"
)

// BookingStatus represents booking status
type BookingStatus string

const (
	StatusInit          BookingStatus = "INIT"
	StatusAwaitingPayment BookingStatus = "AWAITING_PAYMENT"
	StatusPaid          BookingStatus = "PAID"
	StatusConfirmed     BookingStatus = "CONFIRMED"
	StatusCompleted     BookingStatus = "COMPLETED"
	StatusCancelled     BookingStatus = "CANCELLED"
)

// PaymentType represents payment type
type PaymentType string

const (
	PaymentTypePayNow   PaymentType = "PAY_NOW"
	PaymentTypePayAtHotel PaymentType = "PAY_AT_HOTEL"
)

// Booking represents a booking
type Booking struct {
	ID               string        `json:"id" db:"id"`
	UserID           string        `json:"user_id" db:"user_id"`
	HotelID          string        `json:"hotel_id" db:"hotel_id"`
	RoomID           string        `json:"room_id" db:"room_id"`
	BookingReference string        `json:"booking_reference" db:"booking_reference"`
	SupplierReference string       `json:"supplier_reference,omitempty" db:"supplier_reference"`
	CheckIn          time.Time     `json:"check_in" db:"check_in"`
	CheckOut         time.Time     `json:"check_out" db:"check_out"`
	Guests           int           `json:"guests" db:"guests"`
	Status           BookingStatus `json:"status" db:"status"`
	TotalAmount      int           `json:"total_amount" db:"total_amount"`
	Currency         string        `json:"currency" db:"currency"`
	PaymentType      PaymentType   `json:"payment_type" db:"payment_type"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
}

// CreateBookingRequest represents request to create booking
type CreateBookingRequest struct {
	HotelID     string       `json:"hotel_id" validate:"required"`
	RoomID      string       `json:"room_id" validate:"required"`
	CheckIn     time.Time    `json:"check_in" validate:"required"`
	CheckOut    time.Time    `json:"check_out" validate:"required,gtfield=CheckIn"`
	Guests      int          `json:"guests" validate:"required,min=1,max=10"`
	PaymentType PaymentType  `json:"payment_type" validate:"required,oneof=PAY_NOW PAY_AT_HOTEL"`
}

// NewBooking creates a new booking
func NewBooking(userID string, req *CreateBookingRequest) *Booking {
	now := time.Now()
	return &Booking{
		ID:               uuid.New().String(),
		UserID:           userID,
		HotelID:          req.HotelID,
		RoomID:           req.RoomID,
		BookingReference: generateBookingReference(),
		CheckIn:          req.CheckIn,
		CheckOut:         req.CheckOut,
		Guests:           req.Guests,
		Status:           StatusInit,
		Currency:         "IDR",
		PaymentType:      req.PaymentType,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// generateBookingReference generates a unique booking reference
func generateBookingReference() string {
	return "BKG-" + uuid.New().String()[:8]
}

// CanTransitionTo checks if state transition is valid
func (b *Booking) CanTransitionTo(newStatus BookingStatus) bool {
	validTransitions := map[BookingStatus][]BookingStatus{
		StatusInit:            {StatusAwaitingPayment, StatusCancelled},
		StatusAwaitingPayment: {StatusPaid, StatusCancelled},
		StatusPaid:            {StatusConfirmed, StatusCancelled},
		StatusConfirmed:       {StatusCompleted, StatusCancelled},
		StatusCompleted:       {},
		StatusCancelled:       {},
	}

	allowedStates, ok := validTransitions[b.Status]
	if !ok {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

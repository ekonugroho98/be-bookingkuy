package booking

import "errors"

// Package-level errors for booking operations
var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrInvalidBooking       = errors.New("invalid booking")
	ErrBookingCancelled     = errors.New("booking cancelled")
	ErrInvalidStatus        = errors.New("invalid status transition")
	ErrInvalidDateRange     = errors.New("invalid date range")
	ErrInvalidGuests        = errors.New("invalid number of guests")
	ErrInvalidCheckOut      = errors.New("check-out date must be after check-in date")
	ErrInvalidCheckIn       = errors.New("check-in date cannot be in the past")
	ErrRoomNotAvailable     = errors.New("room is not available for the selected dates")
	ErrFailedToCreate       = errors.New("failed to create booking")
	ErrFailedToUpdateStatus = errors.New("failed to update booking status")
	ErrInvalidPaymentType   = errors.New("invalid payment type")
	ErrFailedToUpdate       = errors.New("failed to update booking")
)

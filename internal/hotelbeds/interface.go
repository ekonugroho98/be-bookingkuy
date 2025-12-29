package hotelbeds

import "context"

// ClientInterface defines the interface for HotelBeds API client
// This allows mocking for tests
type ClientInterface interface {
	GetHotelAvailability(ctx context.Context, req *AvailabilityRequest) (*AvailabilityResponse, error)
	GetRoomRates(ctx context.Context, req *RoomRateRequest) (*RoomRateResponse, error)
	GetHotelDetails(ctx context.Context, hotelCode string) (*HotelDetailsResponse, error)
	CreateBooking(ctx context.Context, req *BookingRequest) (*BookingResponse, error)
	CancelBooking(ctx context.Context, bookingReference string) error
}

// Ensure Client implements the interface
var _ ClientInterface = (*Client)(nil)

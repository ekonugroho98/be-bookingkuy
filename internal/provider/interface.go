package provider

import (
	"context"
	"time"
)

// Provider defines the interface that all providers must implement
// Ini adalah kunci untuk mudah menambah provider baru!
type Provider interface {
	// Name returns the provider name (e.g., "hotelbeds", "hotelplanner")
	Name() string

	// SearchAvailability searches for available hotels
	SearchAvailability(ctx context.Context, req *AvailabilityRequest) (*AvailabilityResponse, error)

	// GetHotelDetails retrieves detailed hotel information
	GetHotelDetails(ctx context.Context, hotelID string) (*Hotel, error)

	// CreateBooking creates a booking with the provider
	CreateBooking(ctx context.Context, req *BookingRequest) (*BookingConfirmation, error)

	// CancelBooking cancels a booking
	CancelBooking(ctx context.Context, bookingID string) error

	// GetBookingStatus retrieves booking status
	GetBookingStatus(ctx context.Context, bookingID string) (string, error)

	// HealthCheck checks if provider is healthy
	HealthCheck(ctx context.Context) error
}

// Config represents provider configuration
type Config struct {
	APIKey      string
	APISecret   string
	BaseURL     string
	Timeout     time.Duration
	Enabled     bool
	Priority    int // Lower = higher priority (for cheap-first strategy)
}

// Metrics represents provider metrics
type Metrics struct {
	TotalRequests    int64
	SuccessRequests  int64
	FailedRequests   int64
	AverageResponseTime time.Duration
}

package provider

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/hotelbeds"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// HotelbedsProvider implements Provider interface for Hotelbeds
type HotelbedsProvider struct {
	client *hotelbeds.Client
	mapper *hotelbeds.Mapper
}

// NewHotelbedsProvider creates a new Hotelbeds provider
func NewHotelbedsProvider(apiKey, sharedSecret, baseURL string) *HotelbedsProvider {
	return &HotelbedsProvider{
		client: hotelbeds.NewClient(apiKey, sharedSecret, baseURL),
		mapper: hotelbeds.NewMapper(),
	}
}

// Name returns the provider name
func (h *HotelbedsProvider) Name() string {
	return "hotelbeds"
}

// SearchAvailability searches for available hotels on Hotelbeds
func (h *HotelbedsProvider) SearchAvailability(ctx context.Context, req *AvailabilityRequest) (*AvailabilityResponse, error) {
	logger.Infof("Searching Hotelbeds availability for %s from %s to %s",
		req.City, req.CheckIn.Format("2006-01-02"), req.CheckOut.Format("2006-01-02"))

	// Build Hotelbeds API request
	// TODO: Implement actual Hotelbeds API request structure
	// For now, use mock endpoint
	endpoint := fmt.Sprintf("/hotel-api/1.0/hotels?destination=%s&from=%s&to=%s&occupancy=%d",
		req.City,
		req.CheckIn.Format("2006-01-02"),
		req.CheckOut.Format("2006-01-02"),
		req.Guests,
	)

	// Call Hotelbeds API
	resp, err := h.client.Get(ctx, endpoint)
	if err != nil {
		logger.Errorf("Hotelbeds API error: %v", err)
		return nil, fmt.Errorf("failed to search availability: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Infof("Hotelbeds response: %s", string(body))

	// Parse and convert to canonical model
	// TODO: Implement actual response parsing
	// For now, return mock data converted through mapper
	mockResponse := &AvailabilityResponse{
		Hotels: []HotelAvailability{
			{
				Hotel: Hotel{
					ID:          "HB-12345",
					Name:        "Grand Hotel (Hotelbeds)",
					CountryCode: "ID",
					City:        req.City,
					Rating:      4.5,
				},
				Rooms: []RoomRate{
					{
						Room: Room{
							ID:       "HB-ROOM-1",
							HotelID:  "HB-12345",
							Name:     "Deluxe Room",
							Capacity: 2,
						},
						Rates: []Rate{
							{
								RoomID:    "HB-ROOM-1",
								NetPrice:  1500000,
								Currency:  "IDR",
								Allotment: 10,
							},
						},
					},
				},
			},
		},
	}

	return mockResponse, nil
}

// GetHotelDetails retrieves hotel details from Hotelbeds
func (h *HotelbedsProvider) GetHotelDetails(ctx context.Context, hotelID string) (*Hotel, error) {
	logger.Infof("Getting Hotelbeds hotel details: %s", hotelID)

	// Remove HB- prefix if present
	code := hotelID
	if len(hotelID) > 3 && hotelID[:3] == "HB-" {
		code = hotelID[3:]
	}

	endpoint := fmt.Sprintf("/hotel-api/1.0/hotels/%s", code)
	resp, err := h.client.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get hotel details: %w", err)
	}
	defer resp.Body.Close()

	// TODO: Parse actual response
	return &Hotel{
		ID:     hotelID,
		Name:   "Sample Hotel from Hotelbeds",
		City:   "Jakarta",
		Rating: 4.5,
	}, nil
}

// CreateBooking creates a booking on Hotelbeds
func (h *HotelbedsProvider) CreateBooking(ctx context.Context, req *BookingRequest) (*BookingConfirmation, error) {
	logger.Infof("Creating booking on Hotelbeds for hotel %s", req.HotelID)

	// Convert to Hotelbeds request format
	hbReq := h.mapper.ToHotelbedsBookingRequest(req)

	// Call Hotelbeds booking API
	endpoint := "/hotel-api/1.0/bookings"
	resp, err := h.client.Post(ctx, endpoint, hbReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	defer resp.Body.Close()

	// TODO: Parse actual response
	confirmation := &BookingConfirmation{
		BookingID:         fmt.Sprintf("HB-%d", time.Now().Unix()),
		ProviderReference: fmt.Sprintf("HBS-%d", time.Now().Unix()),
		Status:            "CONFIRMED",
		TotalPrice:        req.Rate.NetPrice,
		Currency:          req.Rate.Currency,
	}

	logger.Infof("Booking created on Hotelbeds: %s", confirmation.BookingID)
	return confirmation, nil
}

// CancelBooking cancels a booking on Hotelbeds
func (h *HotelbedsProvider) CancelBooking(ctx context.Context, bookingID string) error {
	logger.Infof("Cancelling Hotelbeds booking: %s", bookingID)

	// Remove HB- prefix if present
	ref := bookingID
	if len(bookingID) > 3 && bookingID[:3] == "HB-" {
		ref = bookingID[3:]
	}

	endpoint := fmt.Sprintf("/hotel-api/1.0/bookings/%s/cancellations", ref)
	resp, err := h.client.Post(ctx, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}
	defer resp.Body.Close()

	logger.Infof("Booking %s cancelled on Hotelbeds", bookingID)
	return nil
}

// GetBookingStatus retrieves booking status from Hotelbeds
func (h *HotelbedsProvider) GetBookingStatus(ctx context.Context, bookingID string) (string, error) {
	logger.Infof("Getting Hotelbeds booking status: %s", bookingID)

	// Remove HB- prefix if present
	ref := bookingID
	if len(bookingID) > 3 && bookingID[:3] == "HB-" {
		ref = bookingID[3:]
	}

	endpoint := fmt.Sprintf("/hotel-api/1.0/bookings/%s", ref)
	resp, err := h.client.Get(ctx, endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to get booking status: %w", err)
	}
	defer resp.Body.Close()

	// TODO: Parse actual status from response
	return "CONFIRMED", nil
}

// HealthCheck checks if Hotelbeds is healthy
func (h *HotelbedsProvider) HealthCheck(ctx context.Context) error {
	logger.Infof("Health check for Hotelbeds")

	// Ping Hotelbeds status endpoint
	endpoint := "/hotel-api/1.0/status"
	resp, err := h.client.Get(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("Hotelbeds health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Hotelbeds unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

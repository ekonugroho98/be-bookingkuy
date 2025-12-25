package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/provider/types"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// HotelPlannerProvider implements Provider interface for HotelPlanner
// CONTOH: Betapa mudahnya menambah provider baru!
// Cukup implement interface Provider, dan otomatis terintegrasi ke sistem.
type HotelPlannerProvider struct {
	apiKey    string
	apiSecret string
	baseURL   string
	timeout   time.Duration
}

// NewHotelPlannerProvider creates a new HotelPlanner provider
func NewHotelPlannerProvider(apiKey, apiSecret, baseURL string) *HotelPlannerProvider {
	return &HotelPlannerProvider{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		timeout:   30 * time.Second,
	}
}

// Name returns the provider name
func (h *HotelPlannerProvider) Name() string {
	return "hotelplanner"
}

// SearchAvailability searches for available hotels on HotelPlanner
func (h *HotelPlannerProvider) SearchAvailability(ctx context.Context, req *types.AvailabilityRequest) (*types.AvailabilityResponse, error) {
	logger.Infof("Searching HotelPlanner availability for %s", req.City)

	// TODO: Implement actual HotelPlanner API call
	// Format response ke canonical models
	mockResponse := &types.AvailabilityResponse{
		Hotels: []types.HotelAvailability{
			{
				Hotel: types.Hotel{
					ID:          "HP-98765",
					Name:        "Planner Hotel (HotelPlanner)",
					CountryCode: "ID",
					City:        req.City,
					Rating:      4.0,
				},
				Rooms: []types.RoomRate{
					{
						Room: types.Room{
							ID:       "HP-ROOM-1",
							HotelID:  "HP-98765",
							Name:     "Standard Room",
							Capacity: 2,
						},
						Rates: []types.Rate{
							{
								RoomID:    "HP-ROOM-1",
								NetPrice:  1200000,
								Currency:  "IDR",
								Allotment: 5,
							},
						},
					},
				},
			},
		},
	}

	return mockResponse, nil
}

// GetHotelDetails retrieves hotel details from HotelPlanner
func (h *HotelPlannerProvider) GetHotelDetails(ctx context.Context, hotelID string) (*types.Hotel, error) {
	logger.Infof("Getting HotelPlanner hotel details: %s", hotelID)
	// TODO: Implement actual HotelPlanner API call
	return &types.Hotel{
		ID:     hotelID,
		Name:   "Sample Hotel from HotelPlanner",
		City:   "Bali",
		Rating: 4.0,
	}, nil
}

// CreateBooking creates a booking on HotelPlanner
func (h *HotelPlannerProvider) CreateBooking(ctx context.Context, req *types.BookingRequest) (*types.BookingConfirmation, error) {
	logger.Infof("Creating booking on HotelPlanner for hotel %s", req.HotelID)
	// TODO: Implement actual HotelPlanner booking API call
	confirmation := &types.BookingConfirmation{
		BookingID:         fmt.Sprintf("HP-%d", time.Now().Unix()),
		ProviderReference: fmt.Sprintf("HPL-%d", time.Now().Unix()),
		Status:            "CONFIRMED",
		TotalPrice:        req.Rate.NetPrice,
		Currency:          req.Rate.Currency,
	}
	return confirmation, nil
}

// CancelBooking cancels a booking on HotelPlanner
func (h *HotelPlannerProvider) CancelBooking(ctx context.Context, bookingID string) error {
	logger.Infof("Cancelling HotelPlanner booking: %s", bookingID)
	// TODO: Implement actual HotelPlanner cancellation API call
	return nil
}

// GetBookingStatus retrieves booking status from HotelPlanner
func (h *HotelPlannerProvider) GetBookingStatus(ctx context.Context, bookingID string) (string, error) {
	logger.Infof("Getting HotelPlanner booking status: %s", bookingID)
	// TODO: Implement actual HotelPlanner status API call
	return "CONFIRMED", nil
}

// HealthCheck checks if HotelPlanner is healthy
func (h *HotelPlannerProvider) HealthCheck(ctx context.Context) error {
	// TODO: Implement actual health check (ping HotelPlanner API)
	logger.Infof("Health check for HotelPlanner")
	return nil
}

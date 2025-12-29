package hotel

import (
	"context"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/hotelbeds"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines hotel service interface
type Service interface {
	GetHotel(ctx context.Context, hotelID string) (*HotelDetailsResponse, error)
	GetAvailableRooms(ctx context.Context, hotelID string, checkIn, checkOut time.Time, guests int) (*RoomAvailabilityResponse, error)
	GetImages(ctx context.Context, hotelID string) ([]Image, error)
}

type service struct {
	repo            Repository
	hotelbedsClient *hotelbeds.Client
}

// NewService creates a new hotel service
func NewService(repo Repository, hbClient *hotelbeds.Client) Service {
	return &service{
		repo:            repo,
		hotelbedsClient: hbClient,
	}
}

func (s *service) GetHotel(ctx context.Context, hotelID string) (*HotelDetailsResponse, error) {
	logger.Infof("Fetching hotel details: %s", hotelID)

	// 1. Get hotel from database
	hotel, err := s.repo.GetByID(ctx, hotelID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get hotel from database")
		return nil, err
	}

	// 2. Get images from database
	images, err := s.repo.GetImages(ctx, hotelID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get hotel images")
		// Don't fail, just continue with empty images
		images = []Image{}
	}

	// 3. Get rooms from database
	rooms, err := s.repo.GetRooms(ctx, hotelID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get rooms")
		// Don't fail, just continue with empty rooms
		rooms = []RoomInfo{}
	}

	// 4. Build response
	response := &HotelDetailsResponse{
		ID:          hotel.ID,
		Name:        hotel.Name,
		Description: hotel.Description,
		CountryCode: hotel.CountryCode,
		City:        hotel.City,
		Address:     hotel.Address,
		Rating:      hotel.Rating,
		Category:    hotel.Category,
		Images:      images,
		Amenities:   hotel.Amenities,
		Location:    hotel.Location,
		Policies:    hotel.Policies,
		Rooms:       rooms,
	}

	logger.Infof("Hotel details fetched successfully: %s", hotelID)
	return response, nil
}

func (s *service) GetAvailableRooms(ctx context.Context, hotelID string, checkIn, checkOut time.Time, guests int) (*RoomAvailabilityResponse, error) {
	logger.Infof("Checking room availability: hotel=%s, checkIn=%s, checkOut=%s, guests=%d",
		hotelID, checkIn.Format("2006-01-02"), checkOut.Format("2006-01-02"), guests)

	// 1. Check availability with HotelBeds
	availability, err := s.hotelbedsClient.GetHotelAvailability(ctx, &hotelbeds.AvailabilityRequest{
		HotelCode: hotelID,
		CheckIn:   checkIn,
		CheckOut:  checkOut,
		Guests:    guests,
	})

	if err != nil {
		logger.ErrorWithErr(err, "Failed to check availability with HotelBeds")
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	// 2. Transform to response format
	var rooms []AvailableRoom
	for _, room := range availability.Rooms {
		if room.Available {
			rooms = append(rooms, AvailableRoom{
				RoomID:    room.RoomCode,
				RoomName:  room.RoomName,
				Available: room.Available,
				Price:     room.Price,
				Currency:  room.Currency,
				MaxGuests: guests, // Default to requested guests
				Beds:      "1 King Bed", // Could be parsed from room details
			})
		}
	}

	response := &RoomAvailabilityResponse{
		HotelID:   hotelID,
		HotelName: availability.HotelName,
		CheckIn:   checkIn.Format("2006-01-02"),
		CheckOut:  checkOut.Format("2006-01-02"),
		Guests:    guests,
		Rooms:     rooms,
	}

	logger.Infof("Room availability check complete: %d available rooms", len(rooms))
	return response, nil
}

func (s *service) GetImages(ctx context.Context, hotelID string) ([]Image, error) {
	logger.Infof("Fetching hotel images: %s", hotelID)

	images, err := s.repo.GetImages(ctx, hotelID)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get hotel images")
		return nil, err
	}

	logger.Infof("Fetched %d images for hotel: %s", len(images), hotelID)
	return images, nil
}

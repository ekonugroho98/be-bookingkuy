package hotelbeds

import (
	"strconv"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/provider/types"
)

// Mapper handles conversion between Hotelbeds and canonical models
type Mapper struct{}

// NewMapper creates a new mapper
func NewMapper() *Mapper {
	return &Mapper{}
}

// ToCanonicalHotel converts Hotelbeds hotel to canonical model
func (m *Mapper) ToCanonicalHotel(hbHotel HotelbedsHotel) *types.Hotel {
	rating := m.extractRating(hbHotel.Category)

	return &types.Hotel{
		ID:          "HB-" + hbHotel.Code,
		Name:        hbHotel.Name,
		CountryCode: hbHotel.CountryCode,
		City:        hbHotel.Destination.Name,
		Rating:      rating,
		Images:      m.extractImages(hbHotel.Images),
	}
}

// ToCanonicalAvailability converts Hotelbeds availability response to canonical model
func (m *Mapper) ToCanonicalAvailability(hbResponse HotelbedsAvailabilityResponse) *types.AvailabilityResponse {
	hotels := make([]types.HotelAvailability, 0, len(hbResponse.Hotels))

	for _, hbHotel := range hbResponse.Hotels {
		hotel := types.Hotel{
			ID:          "HB-" + hbHotel.Code,
			Name:        hbHotel.Name,
			CountryCode: "ID",
			City:        "Unknown",
			Rating:      m.extractRating(hbHotel.Category),
			Images:      m.extractImages(hbHotel.Images),
		}

		rooms := m.toCanonicalRooms(hbHotel.Rooms)
		hotels = append(hotels, types.HotelAvailability{
			Hotel: hotel,
			Rooms: rooms,
		})
	}

	return &types.AvailabilityResponse{
		Hotels: hotels,
	}
}

// toCanonicalRooms converts Hotelbeds rooms to canonical rooms
func (m *Mapper) toCanonicalRooms(hbRooms []HotelbedsRoom) []types.RoomRate {
	rooms := make([]types.RoomRate, 0, len(hbRooms))

	for _, hbRoom := range hbRooms {
		room := types.Room{
			ID:      "HB-" + hbRoom.Code,
			HotelID: "", // Will be filled by caller
			Name:    hbRoom.Name,
		}

		rates := m.toCanonicalRates(hbRoom.Rates)
		rooms = append(rooms, types.RoomRate{
			Room:  room,
			Rates: rates,
		})
	}

	return rooms
}

// toCanonicalRates converts Hotelbeds rates to canonical rates
func (m *Mapper) toCanonicalRates(hbRates []HotelbedsRate) []types.Rate {
	rates := make([]types.Rate, 0, len(hbRates))

	for _, hbRate := range hbRates {
		freeCancel := m.parseFreeCancellation(hbRate.Cancellation.FreeCancellationDate)

		rate := types.Rate{
			RoomID:    "HB-" + hbRate.RateKey, // Use rateKey as temporary ID
			NetPrice:  int(hbRate.NetPrice),
			Currency:  hbRate.Currency,
			Allotment: hbRate.Allotment,
			MealPlan:  hbRate.MealPlan,
			Cancellation: provider.CancellationPolicy{
				FreeCancellationBefore: freeCancel,
				NonRefundable:           freeCancel.IsZero(),
			},
		}
		rates = append(rates, rate)
	}

	return rates
}

// ToHotelbedsBookingRequest converts canonical booking request to Hotelbeds format
func (m *Mapper) ToHotelbedsBookingRequest(req *types.BookingRequest) HotelbedsBookingRequest {
	// Create guest information
	holder := HotelbedsHolder{
		Name:    req.GuestInfo.FirstName,
		Surname: req.GuestInfo.LastName,
	}

	// Create room booking
	roomBooking := HotelbedsRoomBooking{
		RateKey: req.Rate.RoomID, // Assuming rate.RoomID contains rateKey
		PaxRooms: []HotelbedsPaxRoom{
			{
				PaxType: "AD",
				Guests: []HotelbedsGuest{
					{
						Name:    req.GuestInfo.FirstName,
						Surname: req.GuestInfo.LastName,
					},
				},
			},
		},
	}

	return HotelbedsBookingRequest{
		Holder:          holder,
		Rooms:           []HotelbedsRoomBooking{roomBooking},
		ClientReference: "BKG-" + strconv.FormatInt(time.Now().Unix(), 10),
	}
}

// ToCanonicalBookingConfirmation converts Hotelbeds booking response to canonical model
func (m *Mapper) ToCanonicalBookingConfirmation(hbResponse HotelbedsBookingResponse, hotelID, roomID string, checkIn, checkOut time.Time, totalPrice int, currency string) *types.BookingConfirmation {
	return &types.BookingConfirmation{
		BookingID:         "HB-" + hbResponse.BookingReference,
		ProviderReference: hbResponse.BookingReference,
		Hotel: types.Hotel{
			ID: hotelID,
		},
		Room: types.Room{
			ID: roomID,
		},
		CheckIn:    checkIn,
		CheckOut:   checkOut,
		Status:     hbResponse.Status,
		TotalPrice: totalPrice,
		Currency:   currency,
	}
}

// Helper methods

func (m *Mapper) extractRating(category map[string]interface{}) float64 {
	if categoryCode, ok := category["code"].(string); ok {
		// Extract rating from code like "4STAR" -> 4.0
		if len(categoryCode) > 0 {
			ratingStr := categoryCode[:1]
			if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
				return rating
			}
		}
	}
	return 0.0
}

func (m *Mapper) extractImages(images []HotelbedsImage) []string {
	result := make([]string, 0, len(images))
	for _, img := range images {
		// Extract actual URL from image structure
		// This is a simplified version
		result = append(result, img.ImageURL)
	}
	return result
}

func (m *Mapper) parseFreeCancellation(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	// Parse Hotelbeds date format
	// Format: "2024-12-25 00:00:00"
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

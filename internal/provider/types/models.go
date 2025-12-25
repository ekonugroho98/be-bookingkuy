package types

import "time"

// Hotel represents canonical hotel model
type Hotel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CountryCode string    `json:"country_code"`
	City        string    `json:"city"`
	Address     string    `json:"address,omitempty"`
	Latitude    float64   `json:"latitude,omitempty"`
	Longitude   float64   `json:"longitude,omitempty"`
	Rating      float64   `json:"rating,omitempty"`
	Description string    `json:"description,omitempty"`
	Amenities   []string  `json:"amenities,omitempty"`
	Images      []string  `json:"images,omitempty"`
}

// Room represents canonical room model
type Room struct {
	ID          string           `json:"id"`
	HotelID     string           `json:"hotel_id"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Capacity    int              `json:"capacity"`
	BedType     string           `json:"bed_type,omitempty"`
	Amenities   []string         `json:"amenities,omitempty"`
	Images      []string         `json:"images,omitempty"`
}

// Rate represents canonical rate/pricing model
type Rate struct {
	RoomID       string    `json:"room_id"`
	NetPrice     int       `json:"net_price"`
	Currency     string    `json:"currency"`
	Allotment    int       `json:"allotment"`    // Number of rooms available
	MealPlan     string    `json:"meal_plan,omitempty"` // BB, HB, FB, AI
	Cancellation CancellationPolicy `json:"cancellation"`
}

// CancellationPolicy represents cancellation policy
type CancellationPolicy struct {
	FreeCancellationBefore time.Time `json:"free_cancellation_before"`
	NonRefundable           bool      `json:"non_refundable"`
	PenaltyType             string    `json:"penalty_type,omitempty"` // NIGHTS, PERCENTAGE, FIXED
	PenaltyAmount           int       `json:"penalty_amount,omitempty"`
}

// Availability represents availability search request
type AvailabilityRequest struct {
	CheckIn  time.Time `json:"check_in"`
	CheckOut time.Time `json:"check_out"`
	Guests   int      `json:"guestes"`
	City     string   `json:"city"`
	Country  string   `json:"country,omitempty"`
}

// AvailabilityResponse represents availability search response
type AvailabilityResponse struct {
	Hotels []HotelAvailability `json:"hotels"`
}

// HotelAvailability represents hotel with available rooms
type HotelAvailability struct {
	Hotel Hotel        `json:"hotel"`
	Rooms []RoomRate   `json:"rooms"`
}

// RoomRate represents room with rate
type RoomRate struct {
	Room  Room  `json:"room"`
	Rates []Rate `json:"rates"`
}

// BookingRequest represents booking request
type BookingRequest struct {
	HotelID    string    `json:"hotel_id"`
	RoomID     string    `json:"room_id"`
	CheckIn    time.Time `json:"check_in"`
	CheckOut   time.Time `json:"check_out"`
	Guests     int       `json:"guests"`
	GuestInfo  GuestInfo `json:"guest_info"`
	Rate       Rate      `json:"rate"`
}

// GuestInfo represents guest information
type GuestInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// BookingConfirmation represents booking confirmation
type BookingConfirmation struct {
	BookingID         string    `json:"booking_id"`
	ProviderReference string    `json:"provider_reference"`
	Hotel             Hotel     `json:"hotel"`
	Room              Room      `json:"room"`
	CheckIn           time.Time `json:"check_in"`
	CheckOut          time.Time `json:"check_out"`
	Status            string    `json:"status"`
	TotalPrice        int       `json:"total_price"`
	Currency          string    `json:"currency"`
}

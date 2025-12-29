package hotel

import "time"

// Hotel represents a hotel with all details
type Hotel struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description,omitempty" db:"description"`
	CountryCode string     `json:"country_code" db:"country_code"`
	City        string     `json:"city" db:"city"`
	Address     string     `json:"address,omitempty" db:"address"`
	Rating      float64    `json:"rating" db:"overall_rating"`
	Category    string     `json:"category" db:"category"`
	Images      []Image    `json:"images,omitempty" db:"-"`
	Amenities   []string   `json:"amenities,omitempty" db:"-"`
	Location    Location   `json:"location" db:"-"`
	Policies    Policies   `json:"policies,omitempty" db:"-"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Image represents a hotel image
type Image struct {
	ID        string `json:"id" db:"id"`
	HotelID   string `json:"hotel_id" db:"hotel_id"`
	URL       string `json:"url" db:"url"`
	Type      string `json:"type" db:"type"` // exterior, interior, pool, room, restaurant, etc
	Caption   string `json:"caption,omitempty" db:"caption"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
}

// RoomInfo represents room information
type RoomInfo struct {
	ID        string   `json:"id" db:"id"`
	HotelID   string   `json:"hotel_id" db:"hotel_id"`
	Name      string   `json:"name" db:"name"`
	MaxGuests int      `json:"max_guests" db:"max_guests"`
	Beds      string   `json:"beds,omitempty" db:"beds"`
	Size      string   `json:"size,omitempty" db:"size"`
	Amenities []string `json:"amenities,omitempty" db:"-"`
}

// Location represents hotel location
type Location struct {
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
}

// Policies represents hotel policies
type Policies struct {
	CheckInTime      string `json:"check_in_time,omitempty" db:"check_in_time"`
	CheckOutTime     string `json:"check_out_time,omitempty" db:"check_out_time"`
	Cancellation     string `json:"cancellation,omitempty" db:"cancellation"`
}

// HotelDetailsResponse represents detailed hotel information for API
type HotelDetailsResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CountryCode string       `json:"country_code"`
	City        string       `json:"city"`
	Address     string       `json:"address"`
	Rating      float64      `json:"rating"`
	Category    string       `json:"category"`
	Images      []Image      `json:"images"`
	Amenities   []string     `json:"amenities"`
	Location    Location     `json:"location"`
	Policies    Policies     `json:"policies"`
	Rooms       []RoomInfo   `json:"rooms,omitempty"`
}

// RoomAvailabilityRequest represents request to check room availability
type RoomAvailabilityRequest struct {
	CheckIn  string `json:"check_in"`  // YYYY-MM-DD format
	CheckOut string `json:"check_out"` // YYYY-MM-DD format
	Guests   int    `json:"guests"`
}

// RoomAvailabilityResponse represents available rooms
type RoomAvailabilityResponse struct {
	HotelID    string       `json:"hotel_id"`
	HotelName  string       `json:"hotel_name"`
	CheckIn    string       `json:"check_in"`
	CheckOut   string       `json:"check_out"`
	Guests     int          `json:"guests"`
	Rooms      []AvailableRoom `json:"rooms"`
}

// AvailableRoom represents an available room with pricing
type AvailableRoom struct {
	RoomID    string `json:"room_id"`
	RoomName  string `json:"room_name"`
	Available bool   `json:"available"`
	Price     int    `json:"price"`
	Currency  string `json:"currency"`
	MaxGuests int    `json:"max_guests"`
	Beds      string `json:"beds"`
}

// HotelImagesResponse represents hotel images
type HotelImagesResponse struct {
	HotelID string  `json:"hotel_id"`
	Images  []Image `json:"images"`
}

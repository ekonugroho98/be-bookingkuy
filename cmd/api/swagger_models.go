package main

// Swagger Models for API Documentation

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid email or password"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Name     string `json:"name" example:"John Doe" validate:"required"`
	Email    string `json:"email" example:"john@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=6"`
	Phone    string `json:"phone" example:"+628123456789" validate:"required"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" example:"john@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User   UserDetail `json:"user"`
}

// UserDetail represents user details
type UserDetail struct {
	ID        string `json:"id" example:"550b8b66-746f-4c65-8e0b-7fcb7c0c20d5"`
	Name      string `json:"name" example:"John Doe"`
	Email     string `json:"email" example:"john@example.com"`
	Phone     string `json:"phone" example:"+628123456789"`
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// SearchHotelsRequest represents hotel search request
type SearchHotelsRequest struct {
	CheckIn  string `json:"check_in" example:"2025-01-15T00:00:00Z" validate:"required"`
	CheckOut string `json:"check_out" example:"2025-01-17T00:00:00Z" validate:"required"`
	City     string `json:"city" example:"Bali" validate:"required"`
	Guests   int    `json:"guests" example:"2" validate:"required,min=1,max=10"`
	MinPrice *int   `json:"min_price,omitempty" example:"50"`
	MaxPrice *int   `json:"max_price,omitempty" example:"500"`
}

// SearchHotelResponse represents a hotel in search results
type SearchHotelResponse struct {
	ID          string   `json:"id" example:"60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe"`
	Name        string   `json:"name" example:"Bali Paradise Hotel"`
	CountryCode string   `json:"country_code" example:"ID"`
	City        string   `json:"city" example:"Bali"`
	Rating      *float64 `json:"rating" example:"4.5"`
}

// SearchHotelsResponse represents search results
type SearchHotelsResponse struct {
	Hotels     []SearchHotelResponse `json:"hotels"`
	Total      int                  `json:"total" example:"25"`
	Page       int                  `json:"page" example:"1"`
	PerPage    int                  `json:"per_page" example:"20"`
	TotalPages int                  `json:"total_pages" example:"2"`
}

// HotelDetailsResponse represents detailed hotel information
type HotelDetailsResponse struct {
	ID          string     `json:"id" example:"60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe"`
	Name        string     `json:"name" example:"Bali Paradise Hotel"`
	Description string     `json:"description" example:"Luxurious beachfront resort"`
	CountryCode string     `json:"country_code" example:"ID"`
	City        string     `json:"city" example:"Bali"`
	Address     string     `json:"address" example:"Jalan Beach 123"`
	Rating      float64    `json:"rating" example:"4.5"`
	Category    string     `json:"category" example:"4 Star"`
	Images      []Image    `json:"images"`
	Amenities   []string   `json:"amenities" example:"WiFi,Pool,Spa"`
	Location    Location   `json:"location"`
	Policies    Policies   `json:"policies"`
	Rooms       []RoomInfo `json:"rooms"`
}

// Image represents hotel image
type Image struct {
	ID        string `json:"id" example:"img-001"`
	HotelID   string `json:"hotel_id" example:"60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe"`
	URL       string `json:"url" example:"https://example.com/image.jpg"`
	Type      string `json:"type" example:"exterior"`
	Caption   string `json:"caption,omitempty" example:"Hotel exterior view"`
	SortOrder int    `json:"sort_order" example:"1"`
}

// Location represents hotel location
type Location struct {
	Latitude  float64 `json:"latitude" example:"-8.409518"`
	Longitude float64 `json:"longitude" example:"115.188919"`
}

// Policies represents hotel policies
type Policies struct {
	CheckInTime  string `json:"check_in_time,omitempty" example:"14:00"`
	CheckOutTime string `json:"check_out_time,omitempty" example:"12:00"`
	Cancellation string `json:"cancellation,omitempty" example:"Free cancellation until 24h before check-in"`
}

// RoomInfo represents room information
type RoomInfo struct {
	ID        string   `json:"id" example:"room-001"`
	HotelID   string   `json:"hotel_id" example:"60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe"`
	Name      string   `json:"name" example:"Deluxe Ocean View"`
	MaxGuests int      `json:"max_guests" example:"2"`
	Beds      string   `json:"beds,omitempty" example:"1 King Bed"`
	Size      string   `json:"size,omitempty" example:"45m²"`
	Amenities []string `json:"amenities" example:"WiFi,AC,Minibar"`
}

// RoomAvailabilityRequest represents room availability request
type RoomAvailabilityRequest struct {
	CheckIn  string `json:"check_in" example:"2025-01-15"`
	CheckOut string `json:"check_out" example:"2025-01-17"`
	Guests   int    `json:"guests" example:"2"`
}

// RoomAvailabilityResponse represents available rooms
type RoomAvailabilityResponse struct {
	HotelID    string         `json:"hotel_id" example:"60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe"`
	HotelName  string         `json:"hotel_name" example:"Bali Paradise Hotel"`
	CheckIn    string         `json:"check_in" example:"2025-01-15"`
	CheckOut   string         `json:"check_out" example:"2025-01-17"`
	Guests     int            `json:"guests" example:"2"`
	Rooms      []AvailableRoom `json:"rooms"`
}

// AvailableRoom represents an available room with pricing
type AvailableRoom struct {
	RoomID    string `json:"room_id" example:"room-001"`
	RoomName  string `json:"room_name" example:"Deluxe Ocean View"`
	Available bool   `json:"available" example:"true"`
	Price     int    `json:"price" example:"150000"`
	Currency  string `json:"currency" example:"IDR"`
	MaxGuests int    `json:"max_guests" example:"2"`
	Beds      string `json:"beds" example:"1 King Bed"`
}

// CreateBookingRequest represents booking creation request
type CreateBookingRequest struct {
	HotelID    string `json:"hotel_id" example:"60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe" validate:"required"`
	RoomID     string `json:"room_id" example:"room-001" validate:"required"`
	CheckIn    string `json:"check_in" example:"2025-01-15T00:00:00Z" validate:"required"`
	CheckOut   string `json:"check_out" example:"2025-01-17T00:00:00Z" validate:"required"`
	Guests     int    `json:"guests" example:"2" validate:"required,min=1,max=10"`
	PaymentType string `json:"payment_type" example:"PAY_NOW" validate:"required,oneof=PAY_NOW PAY_AT_HOTEL"`
}

// BookingResponse represents booking response
type BookingResponse struct {
	ID               string `json:"id" example:"bk-001"`
	UserID           string `json:"user_id" example:"user-001"`
	HotelID          string `json:"hotel_id" example:"60cb194d-4cf3-4429-a4a3-9dbdd09cbcfe"`
	HotelName        string `json:"hotelName" example:"Bali Paradise Hotel"`
	HotelImage       string `json:"hotelImage" example:"https://example.com/hotel.jpg"`
	City             string `json:"city" example:"Bali"`
	RoomID           string `json:"room_id" example:"room-001"`
	RoomName         string `json:"roomName" example:"Deluxe Ocean View"`
	BookingReference string `json:"booking_reference" example:"BK20250115001"`
	CheckIn          string `json:"checkIn" example:"Jan 15, 2025"`
	CheckOut         string `json:"checkOut" example:"Jan 17, 2025"`
	Guests           int    `json:"guests" example:"2"`
	GuestsFormatted  string `json:"guestsFormatted" example:"2 Adults"`
	TotalPrice       int    `json:"totalPrice" example:"300000"`
	TotalAmount      int    `json:"total_amount" example:"300000"`
	Currency         string `json:"currency" example:"IDR"`
	Status           string `json:"status" example:"Confirmed" enum:"Confirmed,Pending,Cancelled"`
	PaymentType      string `json:"payment_type" example:"PAY_NOW"`
	CreatedAt        string `json:"created_at" example:"2025-01-10T10:30:00Z"`
}

// MyBookingsResponse represents user's bookings
type MyBookingsResponse struct {
	Bookings  []BookingResponse `json:"bookings"`
	Page      int               `json:"page" example:"1"`
	PerPage   int               `json:"per_page" example:"20"`
}

// CreatePaymentRequest represents payment creation request
type CreatePaymentRequest struct {
	BookingID string `json:"booking_id" example:"bk-001" validate:"required"`
	Amount    int    `json:"amount" example:"300000" validate:"required,gt=0"`
}

// PaymentResponse represents payment response
type PaymentResponse struct {
	ID              string `json:"id" example:"pay-001"`
	BookingID       string `json:"booking_id" example:"bk-001"`
	Amount          int    `json:"amount" example:"300000"`
	Currency        string `json:"currency" example:"IDR"`
	Status          string `json:"status" example:"Pending" enum:"Pending,Completed,Failed,Refunded"`
	PaymentType     string `json:"payment_type" example:"MIDTRANS"`
	PaymentURL      string `json:"payment_url,omitempty" example:"https://app.midtrans.com/payment-link/xxx"`
	MidtransTxID    string `json:"midtrans_tx_id,omitempty" example:"midtrans-001"`
	CreatedAt       string `json:"created_at" example:"2025-01-10T10:30:00Z"`
	UpdatedAt       string `json:"updated_at" example:"2025-01-10T10:35:00Z"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string `json:"status" example:"healthy"`
	Timestamp string `json:"timestamp" example:"2025-01-10T10:30:00Z"`
}

// ReadyResponse represents readiness check response
type ReadyResponse struct {
	Status   string `json:"status" example:"ready"`
	Database string `json:"database" example:"up"`
	Redis    string `json:"redis" example:"up"`
}

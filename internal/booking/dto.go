package booking

import (
	"fmt"
	"time"
)

// BookingResponse represents the booking response format for frontend
type BookingResponse struct {
	// Core fields (matching FE expectations)
	ID            string `json:"id"`
	HotelID       string `json:"hotel_id"`
	HotelName     string `json:"hotelName"`      // ✅ FE: camelCase
	HotelImage    string `json:"hotelImage"`     // ✅ FE: camelCase
	City          string `json:"city"`           // ✅ FE expects this
	RoomID        string `json:"room_id"`
	RoomName      string `json:"roomName"`       // ✅ FE: camelCase
	CheckIn       string `json:"checkIn"`        // ✅ FE: camelCase, formatted string
	CheckOut      string `json:"checkOut"`       // ✅ FE: camelCase, formatted string
	Guests        int    `json:"guests"`         // ✅ Keep original for API
	GuestsFormatted string `json:"guestsFormatted,omitempty"` // ✅ FE display string
	TotalPrice    int    `json:"totalPrice"`     // ✅ FE: camelCase (not total_amount)
	TotalAmount   int    `json:"total_amount"`   // Also keep for API consistency
	Currency      string `json:"currency"`
	Status        string `json:"status"`         // ✅ Will be formatted
	BookingReference string `json:"booking_reference"`

	// Additional fields
	UserID          string `json:"user_id"`
	SupplierReference string `json:"supplier_reference,omitempty"`
	GuestName        string `json:"guest_name,omitempty"`
	GuestEmail       string `json:"guest_email,omitempty"`
	GuestPhone       string `json:"guest_phone,omitempty"`
	SpecialRequests  string `json:"special_requests,omitempty"`
	PaymentType      string `json:"payment_type"`
	CreatedAt        string `json:"created_at"`   // ✅ Formatted string
	UpdatedAt        string `json:"updated_at"`   // ✅ Formatted string
}

// HotelDetails represents hotel information from HotelBeds
type HotelDetails struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Image       string `json:"image"`
	Rating      float64 `json:"rating"`
	Description string `json:"description,omitempty"`
}

// RoomDetails represents room information from HotelBeds
type RoomDetails struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image,omitempty"`
}

// ToBookingResponse converts Booking model to BookingResponse for frontend
func ToBookingResponse(b *Booking, hotel *HotelDetails, room *RoomDetails) *BookingResponse {
	if b == nil {
		return nil
	}

	// Format dates as readable strings (Oct 24, 2024)
	checkIn := b.CheckIn.Format("Jan 2, 2006")
	checkOut := b.CheckOut.Format("Jan 2, 2006")

	// Format guests as string (2 Adults)
	guestsFormatted := fmt.Sprintf("%d Adults", b.Guests)

	// Format status to match frontend expectations
	status := formatStatusForFE(b.Status)

	response := &BookingResponse{
		ID:               b.ID,
		UserID:           b.UserID,
		HotelID:          b.HotelID,
		RoomID:           b.RoomID,
		BookingReference: b.BookingReference,
		CheckIn:          checkIn,
		CheckOut:         checkOut,
		Guests:           b.Guests,
		GuestsFormatted:  guestsFormatted,
		TotalPrice:       b.TotalAmount,
		TotalAmount:      b.TotalAmount,
		Currency:         b.Currency,
		Status:           status,
		PaymentType:      string(b.PaymentType),
		SupplierReference: b.SupplierReference,
		GuestName:        b.GuestName,
		GuestEmail:       b.GuestEmail,
		GuestPhone:       b.GuestPhone,
		SpecialRequests:  b.SpecialRequests,
		CreatedAt:        b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        b.UpdatedAt.Format(time.RFC3339),
	}

	// Add hotel details if available
	if hotel != nil {
		response.HotelName = hotel.Name
		response.HotelImage = hotel.Image
		response.City = hotel.City
	}

	// Add room details if available
	if room != nil {
		response.RoomName = room.Name
	}

	return response
}

// formatStatusForFE converts backend status to frontend-friendly format
func formatStatusForFE(status BookingStatus) string {
	// Frontend expects: 'Confirmed' | 'Pending' | 'Cancelled'
	// Backend has: INIT, AWAITING_PAYMENT, PAID, CONFIRMED, COMPLETED, CANCELLED

	switch status {
	case StatusInit, StatusAwaitingPayment, StatusPaid:
		return "Pending"
	case StatusConfirmed, StatusCompleted:
		return "Confirmed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return string(status)
	}
}

// ToBookingResponseList converts multiple bookings to response format
func ToBookingResponseList(bookings []*Booking, hotels map[string]*HotelDetails, rooms map[string]*RoomDetails) []*BookingResponse {
	responses := make([]*BookingResponse, 0, len(bookings))

	for _, b := range bookings {
		hotel := hotels[b.HotelID]
		room := rooms[b.RoomID]
		response := ToBookingResponse(b, hotel, room)
		responses = append(responses, response)
	}

	return responses
}

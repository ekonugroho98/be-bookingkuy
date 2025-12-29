package main

import (
	"net/http"
)

// Swagger annotations untuk API endpoints
// File ini berisi komentar untuk generate dokumentasi Swagger

// @Summary Register new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} map[string]interface{} "message":"User registered successfully","user":UserDetail
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "Email already registered"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func swaggerRegister(w http.ResponseWriter, r *http.Request) {}

// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func swaggerLogin(w http.ResponseWriter, r *http.Request) {}

// @Summary Search hotels
// @Description Search for available hotels by city and dates
// @Tags search
// @Accept json
// @Produce json
// @Param request body SearchHotelsRequest true "Search criteria"
// @Success 200 {object} SearchHotelsResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /search/hotels [post]
func swaggerSearchHotels(w http.ResponseWriter, r *http.Request) {}

// @Summary Get hotel details
// @Description Get detailed information about a specific hotel
// @Tags hotels
// @Accept json
// @Produce json
// @Param id path string true "Hotel ID"
// @Success 200 {object} HotelDetailsResponse
// @Failure 404 {object} ErrorResponse "Hotel not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /hotels/{id} [get]
func swaggerGetHotel(w http.ResponseWriter, r *http.Request) {}

// @Summary Get available rooms
// @Description Get available rooms for a hotel with pricing
// @Tags hotels
// @Accept json
// @Produce json
// @Param id path string true "Hotel ID"
// @Param check_in query string true "Check-in date (YYYY-MM-DD)"
// @Param check_out query string true "Check-out date (YYYY-MM-DD)"
// @Success 200 {object} RoomAvailabilityResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Hotel not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /hotels/{id}/rooms [get]
func swaggerGetAvailableRooms(w http.ResponseWriter, r *http.Request) {}

// @Summary Get hotel images
// @Description Get all images for a specific hotel
// @Tags hotels
// @Accept json
// @Produce json
// @Param id path string true "Hotel ID"
// @Success 200 {object} map[string]interface{} "images":[]Image
// @Failure 404 {object} ErrorResponse "Hotel not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /hotels/{id}/images [get]
func swaggerGetImages(w http.ResponseWriter, r *http.Request) {}

// @Summary Create booking
// @Description Create a new hotel booking (requires authentication)
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBookingRequest true "Booking details"
// @Success 201 {object} BookingResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /bookings [post]
func swaggerCreateBooking(w http.ResponseWriter, r *http.Request) {}

// @Summary Get booking
// @Description Get details of a specific booking (requires authentication)
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200 {object} BookingResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /bookings/{id} [get]
func swaggerGetBooking(w http.ResponseWriter, r *http.Request) {}

// @Summary Get my bookings
// @Description Get all bookings for the authenticated user (requires authentication)
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} MyBookingsResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /bookings/my [get]
func swaggerGetMyBookings(w http.ResponseWriter, r *http.Request) {}

// @Summary Cancel booking
// @Description Cancel a specific booking (requires authentication)
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]interface{} "message":"Booking cancelled successfully","booking":BookingResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Booking not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /bookings/{id}/cancel [post]
func swaggerCancelBooking(w http.ResponseWriter, r *http.Request) {}

// @Summary Create payment
// @Description Create a payment for a booking (requires authentication)
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePaymentRequest true "Payment details"
// @Success 201 {object} PaymentResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /payments [post]
func swaggerCreatePayment(w http.ResponseWriter, r *http.Request) {}

// @Summary Get payment
// @Description Get payment details (requires authentication)
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} PaymentResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Payment not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /payments/{id} [get]
func swaggerGetPayment(w http.ResponseWriter, r *http.Request) {}

// @Summary Get user profile
// @Description Get authenticated user's profile (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserDetail
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/me [get]
func swaggerGetProfile(w http.ResponseWriter, r *http.Request) {}

// @Summary Health check
// @Description Check API health status
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func swaggerHealthCheck(w http.ResponseWriter, r *http.Request) {}

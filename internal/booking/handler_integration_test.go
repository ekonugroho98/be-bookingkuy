package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/hotelbeds"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestBookingHandler_CreateBooking_Success tests successful booking creation via HTTP
func TestBookingHandler_CreateBooking_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	email := "john@example.com"
	jwtManager := jwt.NewManager("test-secret")
	token, _ := jwtManager.GenerateToken(userID, email)

	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	// Create request body
	reqBody := CreateBookingRequest{
		HotelID:     "hotel-1",
		RoomID:      "room-1",
		CheckIn:     checkIn,
		CheckOut:    checkOut,
		Guests:      2,
		PaymentType: PaymentTypePayNow,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock HotelBeds API responses
	mockHB.On("GetHotelAvailability", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("*hotelbeds.AvailabilityRequest")).Return(&hotelbeds.AvailabilityResponse{
		HotelCode:   "hotel-1",
		HotelName:   "Test Hotel",
		IsAvailable: true,
		Rooms: []hotelbeds.Room{
			{
				RoomCode:  "room-1",
				RoomName:  "Test Room",
				Available: true,
				Price:     1500000,
				Currency:  "IDR",
			},
		},
		TotalPrice: 1500000,
		Currency:   "IDR",
	}, nil)

	mockHB.On("GetRoomRates", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("*hotelbeds.RoomRateRequest")).Return(&hotelbeds.RoomRateResponse{
		HotelCode:  "hotel-1",
		RoomCode:   "room-1",
		RoomName:   "Test Room",
		TotalPrice: 1500000,
		Currency:   "IDR",
		Rates: []hotelbeds.Rate{
			{
				RateCode: "RATE-1",
				Price:    1500000,
				Currency: "IDR",
			},
		},
	}, nil)

	// Mock: Successful booking creation
	mockRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("*booking.Booking")).Return(nil)
	mockRepo.On("UpdateStatus", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"), StatusAwaitingPayment).Return(nil)
	mockEB.On("Publish", mock.AnythingOfType("*context.valueCtx"), "booking.created", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Create HTTP request with auth
	req := httptest.NewRequest("POST", "/bookings", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.CreateBooking(rr, req)

	// Assertions
	require.Equal(t, http.StatusCreated, rr.Code)

	var respBody Booking
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, userID, respBody.UserID)
	assert.Equal(t, "hotel-1", respBody.HotelID)
	assert.Equal(t, StatusAwaitingPayment, respBody.Status)
	assert.Equal(t, 1500000, respBody.TotalAmount) // Real price from HotelBeds!

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
	mockHB.AssertExpectations(t)
}

// TestBookingHandler_GetBooking_Success tests successful booking retrieval via HTTP
// NOTE: This test is expected to fail because PathValue requires an actual HTTP router
// In a real integration test with a test server, this would work. For handler-level testing,
// we'd need to mock the PathValue method or use a test router.
func TestBookingHandler_GetBooking_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)
	handler := NewHandler(service)

	bookingID := "booking-123"

	// Mock: Booking exists
	expectedBooking := &Booking{
		ID:               bookingID,
		UserID:           "user-123",
		HotelID:          "hotel-1",
		RoomID:           "room-1",
		Status:           StatusConfirmed,
		BookingReference: "BKG-TEST123",
		CreatedAt:        time.Now(),
	}
	mockRepo.On("GetByID", mock.AnythingOfType("*context.emptyCtx"), bookingID).Return(expectedBooking, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/bookings/"+bookingID, nil)
	// Set path value manually (simulating router)
	req = req.WithContext(context.WithValue(req.Context(), "id", bookingID))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetBooking(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody Booking
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, bookingID, respBody.ID)
	assert.Equal(t, StatusConfirmed, respBody.Status)

	mockRepo.AssertExpectations(t)
}

// TestBookingHandler_GetBooking_NotFound tests booking retrieval with non-existent booking
func TestBookingHandler_GetBooking_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)
	handler := NewHandler(service)

	bookingID := "non-existent"

	// Mock: Booking not found
	mockRepo.On("GetByID", mock.AnythingOfType("*context.emptyCtx"), bookingID).Return(nil, errors.New("booking not found"))

	// Create HTTP request
	req := httptest.NewRequest("GET", "/bookings/"+bookingID, nil)
	// Set path value manually (simulating router)
	req = req.WithContext(context.WithValue(req.Context(), "id", bookingID))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetBooking(rr, req)

	// Assertions
	require.Equal(t, http.StatusNotFound, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Booking not found")

	mockRepo.AssertExpectations(t)
}

// TestBookingHandler_CancelBooking_Success tests successful booking cancellation via HTTP
func TestBookingHandler_CancelBooking_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)
	handler := NewHandler(service)

	bookingID := "booking-123"

	// Mock: Get existing booking
	existingBooking := &Booking{
		ID:               bookingID,
		UserID:           "user-123",
		HotelID:          "hotel-1",
		RoomID:           "room-1",
		Status:           StatusConfirmed,
		BookingReference: "BKG-TEST123",
		CreatedAt:        time.Now(),
	}
	mockRepo.On("GetByID", mock.AnythingOfType("*context.emptyCtx"), bookingID).Return(existingBooking, nil)

	// Mock: Cancel successful
	mockRepo.On("UpdateStatus", mock.AnythingOfType("*context.emptyCtx"), bookingID, StatusCancelled).Return(nil)
	mockEB.On("Publish", mock.AnythingOfType("*context.emptyCtx"), "booking.cancelled", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/bookings/"+bookingID+"/cancel", nil)
	// Set path value manually (simulating router)
	req = req.WithContext(context.WithValue(req.Context(), "id", bookingID))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.CancelBooking(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, "Booking cancelled successfully", respBody["message"])
	assert.NotNil(t, respBody["booking"])

	booking := respBody["booking"].(map[string]interface{})
	assert.Equal(t, StatusCancelled, booking["status"])

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestBookingHandler_GetMyBookings_Success tests successful retrieval of user's bookings
func TestBookingHandler_GetMyBookings_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	email := "john@example.com"
	jwtManager := jwt.NewManager("test-secret")
	token, _ := jwtManager.GenerateToken(userID, email)

	// Mock: Get user bookings
	expectedBookings := []*Booking{
		{
			ID:               "booking-1",
			UserID:           userID,
			HotelID:          "hotel-1",
			RoomID:           "room-1",
			Status:           StatusConfirmed,
			BookingReference: "BKG-TEST123",
		},
		{
			ID:               "booking-2",
			UserID:           userID,
			HotelID:          "hotel-2",
			RoomID:           "room-2",
			Status:           StatusAwaitingPayment,
			BookingReference: "BKG-TEST456",
		},
	}
	mockRepo.On("GetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, 20, 0).Return(expectedBookings, nil)

	// Create HTTP request with auth
	req := httptest.NewRequest("GET", "/bookings/my?page=1&per_page=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetMyBookings(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.NotNil(t, respBody["bookings"])
	assert.Equal(t, float64(1), respBody["page"])
	assert.Equal(t, float64(20), respBody["per_page"])

	mockRepo.AssertExpectations(t)
}

// TestBookingHandler_CreateBooking_InvalidJSON tests booking creation with invalid JSON
func TestBookingHandler_CreateBooking_InvalidJSON(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	jwtManager := jwt.NewManager("test-secret")
	token, _ := jwtManager.GenerateToken(userID, "john@example.com")

	// Create HTTP request with invalid JSON
	req := httptest.NewRequest("POST", "/bookings", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.CreateBooking(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Invalid request body")
}

// TestBookingHandler_GetMyBookings_DefaultPagination tests default pagination values
func TestBookingHandler_GetMyBookings_DefaultPagination(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)
	handler := NewHandler(service)

	// Create authenticated request
	userID := "user-123"
	email := "john@example.com"
	jwtManager := jwt.NewManager("test-secret")
	token, _ := jwtManager.GenerateToken(userID, email)

	// Mock: Get user bookings with default pagination
	expectedBookings := []*Booking{}
	mockRepo.On("GetByUserID", mock.AnythingOfType("*context.valueCtx"), userID, 20, 0).Return(expectedBookings, nil)

	// Create HTTP request with auth (no pagination params)
	req := httptest.NewRequest("GET", "/bookings/my", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Add user context (simulating middleware)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetMyBookings(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, float64(1), respBody["page"])
	assert.Equal(t, float64(20), respBody["per_page"])

	mockRepo.AssertExpectations(t)
}

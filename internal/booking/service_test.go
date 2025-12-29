package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/hotelbeds"
	"github.com/ekonugroho98/be-bookingkuy/internal/pricing"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of booking.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, booking *Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Booking), args.Error(1)
}

func (m *MockRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Booking, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Booking), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, booking *Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id string, status BookingStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockEventBus is a mock implementation of eventbus.EventBus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, eventType string, data map[string]interface{}) error {
	args := m.Called(ctx, eventType, data)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(ctx context.Context, eventType string, handler eventbus.Handler) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) SubscribeAsync(ctx context.Context, eventType string, handler eventbus.Handler) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

// MockPricingService is a mock implementation of pricing.Service
type MockPricingService struct {
	mock.Mock
}

func (m *MockPricingService) CalculateSellPrice(netPrice int, category pricing.HotelCategory) (*pricing.PriceCalculation, error) {
	args := m.Called(netPrice, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pricing.PriceCalculation), args.Error(1)
}

func (m *MockPricingService) CalculateMargin(netPrice, sellPrice int) (int, float64) {
	args := m.Called(netPrice, sellPrice)
	return args.Int(0), args.Get(1).(float64)
}

// MockHotelbedsClient is a mock implementation of hotelbeds.ClientInterface
type MockHotelbedsClient struct {
	mock.Mock
}

func (m *MockHotelbedsClient) GetHotelAvailability(ctx context.Context, req *hotelbeds.AvailabilityRequest) (*hotelbeds.AvailabilityResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*hotelbeds.AvailabilityResponse), args.Error(1)
}

func (m *MockHotelbedsClient) GetRoomRates(ctx context.Context, req *hotelbeds.RoomRateRequest) (*hotelbeds.RoomRateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*hotelbeds.RoomRateResponse), args.Error(1)
}

func (m *MockHotelbedsClient) GetHotelDetails(ctx context.Context, hotelCode string) (*hotelbeds.HotelDetailsResponse, error) {
	args := m.Called(ctx, hotelCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*hotelbeds.HotelDetailsResponse), args.Error(1)
}

func (m *MockHotelbedsClient) CreateBooking(ctx context.Context, req *hotelbeds.BookingRequest) (*hotelbeds.BookingResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*hotelbeds.BookingResponse), args.Error(1)
}

func (m *MockHotelbedsClient) CancelBooking(ctx context.Context, bookingReference string) error {
	args := m.Called(ctx, bookingReference)
	return args.Error(0)
}

// TestNewService tests creating a new booking service
func TestNewService(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	require.NotNil(t, service)
}

// TestService_CreateBooking_Success tests successful booking creation
func TestService_CreateBooking_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	userID := "user-123"
	req := &CreateBookingRequest{
		HotelID:     "hotel-123",
		RoomID:      "room-123",
		CheckIn:     time.Now().Add(24 * time.Hour),
		CheckOut:    time.Now().Add(48 * time.Hour),
		Guests:      2,
		PaymentType: PaymentTypePayNow,
	}

	// Setup HotelBeds mock expectations
	mockHB.On("GetHotelAvailability", ctx, mock.AnythingOfType("*hotelbeds.AvailabilityRequest")).Return(&hotelbeds.AvailabilityResponse{
		HotelCode:   req.HotelID,
		HotelName:   "Test Hotel",
		IsAvailable: true,
		Rooms: []hotelbeds.Room{
			{
				RoomCode:  req.RoomID,
				RoomName:  "Test Room",
				Available: true,
				Price:     1500000,
				Currency:  "IDR",
			},
		},
		TotalPrice: 1500000,
		Currency:   "IDR",
	}, nil)

	mockHB.On("GetRoomRates", ctx, mock.AnythingOfType("*hotelbeds.RoomRateRequest")).Return(&hotelbeds.RoomRateResponse{
		HotelCode:  req.HotelID,
		RoomCode:   req.RoomID,
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

	// Setup repository and event bus expectations
	mockRepo.On("Create", ctx, mock.AnythingOfType("*booking.Booking")).Return(nil)
	mockRepo.On("UpdateStatus", ctx, mock.AnythingOfType("string"), StatusAwaitingPayment).Return(nil)
	mockEB.On("Publish", ctx, "booking.created", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	booking, err := service.CreateBooking(ctx, userID, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, booking)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, req.HotelID, booking.HotelID)
	assert.Equal(t, req.RoomID, booking.RoomID)
	assert.Equal(t, req.Guests, booking.Guests)
	assert.Equal(t, req.CheckIn, booking.CheckIn)
	assert.Equal(t, req.CheckOut, booking.CheckOut)
	assert.Equal(t, req.PaymentType, booking.PaymentType)
	assert.Equal(t, StatusAwaitingPayment, booking.Status)
	assert.NotEmpty(t, booking.ID)
	assert.NotEmpty(t, booking.BookingReference)
	assert.Contains(t, booking.BookingReference, "BKG-")
	assert.Equal(t, "IDR", booking.Currency)
	assert.Equal(t, 1500000, booking.TotalAmount) // Real price from HotelBeds!

	// Verify all mocks were called
	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
	mockHB.AssertExpectations(t)
}

// TestService_CreateBooking_RepositoryError tests booking creation with repository error
func TestService_CreateBooking_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	userID := "user-123"
	req := &CreateBookingRequest{
		HotelID:     "hotel-123",
		RoomID:      "room-123",
		CheckIn:     time.Now().Add(24 * time.Hour),
		CheckOut:    time.Now().Add(48 * time.Hour),
		Guests:      2,
		PaymentType: PaymentTypePayNow,
	}

	// Setup HotelBeds mock expectations (success, but repo fails)
	mockHB.On("GetHotelAvailability", ctx, mock.AnythingOfType("*hotelbeds.AvailabilityRequest")).Return(&hotelbeds.AvailabilityResponse{
		HotelCode:   req.HotelID,
		HotelName:   "Test Hotel",
		IsAvailable: true,
		Rooms: []hotelbeds.Room{
			{
				RoomCode:  req.RoomID,
				RoomName:  "Test Room",
				Available: true,
				Price:     1500000,
				Currency:  "IDR",
			},
		},
		TotalPrice: 1500000,
		Currency:   "IDR",
	}, nil)

	mockHB.On("GetRoomRates", ctx, mock.AnythingOfType("*hotelbeds.RoomRateRequest")).Return(&hotelbeds.RoomRateResponse{
		HotelCode:  req.HotelID,
		RoomCode:   req.RoomID,
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

	// Setup expectations - Create fails
	mockRepo.On("Create", ctx, mock.AnythingOfType("*booking.Booking")).Return(errors.New("database error"))

	// Execute
	booking, err := service.CreateBooking(ctx, userID, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "failed to create booking")

	mockRepo.AssertExpectations(t)
	mockHB.AssertExpectations(t)
}

// TestService_CreateBooking_UpdateStatusError tests booking creation with update status error
func TestService_CreateBooking_UpdateStatusError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	userID := "user-123"
	req := &CreateBookingRequest{
		HotelID:     "hotel-123",
		RoomID:      "room-123",
		CheckIn:     time.Now().Add(24 * time.Hour),
		CheckOut:    time.Now().Add(48 * time.Hour),
		Guests:      2,
		PaymentType: PaymentTypePayNow,
	}

	// Setup HotelBeds mock expectations
	mockHB.On("GetHotelAvailability", ctx, mock.AnythingOfType("*hotelbeds.AvailabilityRequest")).Return(&hotelbeds.AvailabilityResponse{
		HotelCode:   req.HotelID,
		HotelName:   "Test Hotel",
		IsAvailable: true,
		Rooms: []hotelbeds.Room{
			{
				RoomCode:  req.RoomID,
				RoomName:  "Test Room",
				Available: true,
				Price:     1500000,
				Currency:  "IDR",
			},
		},
		TotalPrice: 1500000,
		Currency:   "IDR",
	}, nil)

	mockHB.On("GetRoomRates", ctx, mock.AnythingOfType("*hotelbeds.RoomRateRequest")).Return(&hotelbeds.RoomRateResponse{
		HotelCode:  req.HotelID,
		RoomCode:   req.RoomID,
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

	// Setup expectations
	mockRepo.On("Create", ctx, mock.AnythingOfType("*booking.Booking")).Return(nil)
	mockRepo.On("UpdateStatus", ctx, mock.AnythingOfType("string"), StatusAwaitingPayment).Return(errors.New("database error"))

	// Execute
	booking, err := service.CreateBooking(ctx, userID, req)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "failed to update booking status")

	mockRepo.AssertExpectations(t)
	mockHB.AssertExpectations(t)
}

// TestService_GetBooking_Success tests successful booking retrieval
func TestService_GetBooking_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	bookingID := "booking-123"
	expectedBooking := &Booking{
		ID:               bookingID,
		UserID:           "user-123",
		HotelID:          "hotel-123",
		RoomID:           "room-123",
		BookingReference: "BKG-ABC123",
		Status:           StatusAwaitingPayment,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, bookingID).Return(expectedBooking, nil)

	// Execute
	booking, err := service.GetBooking(ctx, bookingID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, booking)
	assert.Equal(t, expectedBooking.ID, booking.ID)
	assert.Equal(t, expectedBooking.UserID, booking.UserID)
	assert.Equal(t, expectedBooking.HotelID, booking.HotelID)
	assert.Equal(t, expectedBooking.RoomID, booking.RoomID)
	assert.Equal(t, expectedBooking.BookingReference, booking.BookingReference)
	assert.Equal(t, expectedBooking.Status, booking.Status)

	mockRepo.AssertExpectations(t)
}

// TestService_GetBooking_NotFound tests booking retrieval with not found error
func TestService_GetBooking_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	bookingID := "non-existent-booking"

	// Setup expectations
	mockRepo.On("GetByID", ctx, bookingID).Return(nil, errors.New("booking not found"))

	// Execute
	booking, err := service.GetBooking(ctx, bookingID)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, booking)

	mockRepo.AssertExpectations(t)
}

// TestService_GetUserBookings_Success tests successful user bookings retrieval
func TestService_GetUserBookings_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	userID := "user-123"
	page := 1
	perPage := 10

	expectedBookings := []*Booking{
		{
			ID:               "booking-1",
			UserID:           userID,
			HotelID:          "hotel-1",
			RoomID:           "room-1",
			BookingReference: "BKG-ABC123",
			Status:           StatusConfirmed,
		},
		{
			ID:               "booking-2",
			UserID:           userID,
			HotelID:          "hotel-2",
			RoomID:           "room-2",
			BookingReference: "BKG-DEF456",
			Status:           StatusCompleted,
		},
	}

	// Setup expectations (offset = (page-1) * perPage = 0)
	mockRepo.On("GetByUserID", ctx, userID, perPage, 0).Return(expectedBookings, nil)

	// Execute
	bookings, err := service.GetUserBookings(ctx, userID, page, perPage)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, bookings)
	assert.Len(t, bookings, 2)
	assert.Equal(t, expectedBookings[0].ID, bookings[0].ID)
	assert.Equal(t, expectedBookings[1].ID, bookings[1].ID)

	mockRepo.AssertExpectations(t)
}

// TestService_GetUserBookings_Empty tests user bookings retrieval with no bookings
func TestService_GetUserBookings_Empty(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	userID := "user-123"
	page := 1
	perPage := 10

	// Setup expectations
	mockRepo.On("GetByUserID", ctx, userID, perPage, 0).Return([]*Booking{}, nil)

	// Execute
	bookings, err := service.GetUserBookings(ctx, userID, page, perPage)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, bookings)
	assert.Empty(t, bookings)

	mockRepo.AssertExpectations(t)
}

// TestService_GetUserBookings_Pagination tests pagination calculation
func TestService_GetUserBookings_Pagination(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"

	tests := []struct {
		name     string
		page     int
		perPage  int
		expected int // expected offset
	}{
		{name: "Page 1", page: 1, perPage: 10, expected: 0},
		{name: "Page 2", page: 2, perPage: 10, expected: 10},
		{name: "Page 3", page: 3, perPage: 5, expected: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockEB := new(MockEventBus)
			mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)
			service := NewService(mockRepo, mockEB, mockPS, mockHB)

			// Setup expectations
			mockRepo.On("GetByUserID", ctx, userID, tt.perPage, tt.expected).Return([]*Booking{}, nil)

			// Execute
			_, err := service.GetUserBookings(ctx, userID, tt.page, tt.perPage)

			// Assertions
			require.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestService_UpdateStatus_Success tests successful status update
func TestService_UpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	bookingID := "booking-123"
	existingBooking := &Booking{
		ID:               bookingID,
		UserID:           "user-123",
		HotelID:          "hotel-123",
		RoomID:           "room-123",
		BookingReference: "BKG-ABC123",
		Status:           StatusAwaitingPayment,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)
	mockRepo.On("UpdateStatus", ctx, bookingID, StatusPaid).Return(nil)
	mockEB.On("Publish", ctx, "booking.paid", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	booking, err := service.UpdateStatus(ctx, bookingID, StatusPaid)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, booking)
	assert.Equal(t, StatusPaid, booking.Status)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_UpdateStatus_InvalidTransition tests invalid state transition
func TestService_UpdateStatus_InvalidTransition(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	bookingID := "booking-123"
	existingBooking := &Booking{
		ID:               bookingID,
		UserID:           "user-123",
		HotelID:          "hotel-123",
		RoomID:           "room-123",
		BookingReference: "BKG-ABC123",
		Status:           StatusPaid, // Already paid
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)

	// Execute - trying to go back to AWAITING_PAYMENT (invalid)
	booking, err := service.UpdateStatus(ctx, bookingID, StatusAwaitingPayment)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "invalid state transition")

	mockRepo.AssertExpectations(t)
}

// TestService_UpdateStatus_BookingNotFound tests status update with non-existent booking
func TestService_UpdateStatus_BookingNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	bookingID := "non-existent-booking"

	// Setup expectations
	mockRepo.On("GetByID", ctx, bookingID).Return(nil, errors.New("booking not found"))

	// Execute
	booking, err := service.UpdateStatus(ctx, bookingID, StatusPaid)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, booking)

	mockRepo.AssertExpectations(t)
}

// TestService_UpdateStatus_UpdateError tests status update with database error
func TestService_UpdateStatus_UpdateError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	bookingID := "booking-123"
	existingBooking := &Booking{
		ID:               bookingID,
		UserID:           "user-123",
		HotelID:          "hotel-123",
		RoomID:           "room-123",
		BookingReference: "BKG-ABC123",
		Status:           StatusAwaitingPayment,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)
	mockRepo.On("UpdateStatus", ctx, bookingID, StatusPaid).Return(errors.New("database error"))

	// Execute
	booking, err := service.UpdateStatus(ctx, bookingID, StatusPaid)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, booking)
	assert.Contains(t, err.Error(), "failed to update booking status")

	mockRepo.AssertExpectations(t)
}

// TestService_CancelBooking_Success tests successful booking cancellation
func TestService_CancelBooking_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)
	mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

	service := NewService(mockRepo, mockEB, mockPS, mockHB)

	ctx := context.Background()
	bookingID := "booking-123"
	existingBooking := &Booking{
		ID:               bookingID,
		UserID:           "user-123",
		HotelID:          "hotel-123",
		RoomID:           "room-123",
		BookingReference: "BKG-ABC123",
		Status:           StatusAwaitingPayment,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)
	mockRepo.On("UpdateStatus", ctx, bookingID, StatusCancelled).Return(nil)
	mockEB.On("Publish", ctx, "booking.cancelled", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	booking, err := service.CancelBooking(ctx, bookingID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, booking)
	assert.Equal(t, StatusCancelled, booking.Status)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_UpdateStatus_AllStatuses tests all status transitions
func TestService_UpdateStatus_AllStatuses(t *testing.T) {
	statusEvents := map[BookingStatus]string{
		StatusPaid:      "booking.paid",
		StatusConfirmed: "booking.confirmed",
		StatusCancelled: "booking.cancelled",
	}

	for status, eventType := range statusEvents {
		t.Run(string(status), func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockEB := new(MockEventBus)
			mockPS := new(MockPricingService)
	mockHB := new(MockHotelbedsClient)

			service := NewService(mockRepo, mockEB, mockPS, mockHB)

			ctx := context.Background()
			bookingID := "booking-123"
			existingBooking := &Booking{
				ID:               bookingID,
				UserID:           "user-123",
				HotelID:          "hotel-123",
				RoomID:           "room-123",
				BookingReference: "BKG-ABC123",
				Status:           StatusAwaitingPayment,
			}

			// For PAID, we need to be in AWAITING_PAYMENT
			// For CONFIRMED, we need to be in PAID
			// For CANCELLED, we can be in AWAITING_PAYMENT
			if status == StatusConfirmed {
				existingBooking.Status = StatusPaid
			}

			// Setup expectations
			mockRepo.On("GetByID", ctx, bookingID).Return(existingBooking, nil)
			mockRepo.On("UpdateStatus", ctx, bookingID, status).Return(nil)
			mockEB.On("Publish", ctx, eventType, mock.AnythingOfType("map[string]interface {}")).Return(nil)

			// Execute
			booking, err := service.UpdateStatus(ctx, bookingID, status)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, booking)
			assert.Equal(t, status, booking.Status)

			mockRepo.AssertExpectations(t)
			mockEB.AssertExpectations(t)
		})
	}
}

// TestNewBooking tests creating a new booking
func TestNewBooking(t *testing.T) {
	userID := "user-123"
	req := &CreateBookingRequest{
		HotelID:     "hotel-123",
		RoomID:      "room-123",
		CheckIn:     time.Now().Add(24 * time.Hour),
		CheckOut:    time.Now().Add(48 * time.Hour),
		Guests:      2,
		PaymentType: PaymentTypePayNow,
	}

	booking := NewBooking(userID, req)

	require.NotNil(t, booking)
	assert.NotEmpty(t, booking.ID)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, req.HotelID, booking.HotelID)
	assert.Equal(t, req.RoomID, booking.RoomID)
	assert.Equal(t, req.CheckIn, booking.CheckIn)
	assert.Equal(t, req.CheckOut, booking.CheckOut)
	assert.Equal(t, req.Guests, booking.Guests)
	assert.Equal(t, req.PaymentType, booking.PaymentType)
	assert.Equal(t, StatusInit, booking.Status)
	assert.Equal(t, "IDR", booking.Currency)
	assert.NotEmpty(t, booking.BookingReference)
	assert.Contains(t, booking.BookingReference, "BKG-")
	assert.False(t, booking.CreatedAt.IsZero())
	assert.False(t, booking.UpdatedAt.IsZero())
}

// TestGenerateBookingReference tests booking reference generation
func TestGenerateBookingReference(t *testing.T) {
	ref1 := generateBookingReference()
	ref2 := generateBookingReference()

	assert.NotEmpty(t, ref1)
	assert.NotEmpty(t, ref2)
	assert.Contains(t, ref1, "BKG-")
	assert.Contains(t, ref2, "BKG-")
	assert.NotEqual(t, ref1, ref2, "Each reference should be unique")
	assert.Len(t, ref1, 12) // "BKG-" + 8 characters
}

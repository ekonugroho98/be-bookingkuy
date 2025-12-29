package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of payment.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, payment *Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Payment), args.Error(1)
}

func (m *MockRepository) GetByBookingID(ctx context.Context, bookingID string) (*Payment, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Payment), args.Error(1)
}

func (m *MockRepository) GetByProviderRef(ctx context.Context, providerRef string) (*Payment, error) {
	args := m.Called(ctx, providerRef)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Payment), args.Error(1)
}

func (m *MockRepository) UpdateStatus(ctx context.Context, id string, status PaymentStatus, providerRef string) error {
	args := m.Called(ctx, id, status, providerRef)
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

// TestNewService tests creating a new payment service
func TestNewService(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	require.NotNil(t, service)
}

// TestService_CreatePayment_NewPayment_Success tests successful payment creation for new booking
func TestService_CreatePayment_NewPayment_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	req := &CreatePaymentRequest{
		BookingID: "booking-123",
		Provider:  ProviderMidtrans,
		Method:    "gopay",
	}
	amount := 1000000

	// Setup expectations - no existing payment
	mockRepo.On("GetByBookingID", ctx, req.BookingID).Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, mock.AnythingOfType("*payment.Payment")).Return(nil)

	// Execute
	payment, err := service.CreatePayment(ctx, req, amount)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, payment)
	assert.Equal(t, req.BookingID, payment.BookingID)
	assert.Equal(t, req.Provider, payment.Provider)
	assert.Equal(t, req.Method, payment.Method)
	assert.Equal(t, amount, payment.Amount)
	assert.Equal(t, "IDR", payment.Currency)
	assert.Equal(t, StatusPending, payment.Status)
	assert.NotEmpty(t, payment.ID)
	assert.NotEmpty(t, payment.PaymentURL)
	assert.Contains(t, payment.PaymentURL, "payment-gateway.example.com")

	mockRepo.AssertExpectations(t)
}

// TestService_CreatePayment_ExistingPending_ReturnsExisting tests returning existing pending payment
func TestService_CreatePayment_ExistingPending_ReturnsExisting(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	req := &CreatePaymentRequest{
		BookingID: "booking-123",
		Provider:  ProviderMidtrans,
		Method:    "gopay",
	}
	amount := 1000000

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: req.BookingID,
		Provider:  req.Provider,
		Method:    req.Method,
		Amount:    amount,
		Currency:  "IDR",
		Status:    StatusPending,
	}

	// Setup expectations - existing pending payment
	mockRepo.On("GetByBookingID", ctx, req.BookingID).Return(existingPayment, nil)

	// Execute
	payment, err := service.CreatePayment(ctx, req, amount)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, payment)
	assert.Equal(t, existingPayment.ID, payment.ID)
	assert.Equal(t, existingPayment.Status, StatusPending)

	mockRepo.AssertExpectations(t)
}

// TestService_CreatePayment_ExistingCompleted_ReturnsError tests error when payment already completed
func TestService_CreatePayment_ExistingCompleted_ReturnsError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	req := &CreatePaymentRequest{
		BookingID: "booking-123",
		Provider:  ProviderMidtrans,
		Method:    "gopay",
	}
	amount := 1000000

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: req.BookingID,
		Status:    StatusSuccess,
	}

	// Setup expectations - existing completed payment
	mockRepo.On("GetByBookingID", ctx, req.BookingID).Return(existingPayment, nil)

	// Execute
	payment, err := service.CreatePayment(ctx, req, amount)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "payment already completed")

	mockRepo.AssertExpectations(t)
}

// TestService_CreatePayment_CreateError tests error when creating payment fails
func TestService_CreatePayment_CreateError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	req := &CreatePaymentRequest{
		BookingID: "booking-123",
		Provider:  ProviderMidtrans,
		Method:    "gopay",
	}
	amount := 1000000

	// Setup expectations
	mockRepo.On("GetByBookingID", ctx, req.BookingID).Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, mock.AnythingOfType("*payment.Payment")).Return(errors.New("database error"))

	// Execute
	payment, err := service.CreatePayment(ctx, req, amount)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "failed to create payment")

	mockRepo.AssertExpectations(t)
}

// TestService_GetPayment_Success tests successful payment retrieval
func TestService_GetPayment_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	paymentID := "payment-123"
	expectedPayment := &Payment{
		ID:        paymentID,
		BookingID: "booking-123",
		Status:    StatusPending,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, paymentID).Return(expectedPayment, nil)

	// Execute
	payment, err := service.GetPayment(ctx, paymentID)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, payment)
	assert.Equal(t, expectedPayment.ID, payment.ID)
	assert.Equal(t, expectedPayment.BookingID, payment.BookingID)

	mockRepo.AssertExpectations(t)
}

// TestService_GetPayment_NotFound tests payment retrieval with not found error
func TestService_GetPayment_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	paymentID := "non-existent-payment"

	// Setup expectations
	mockRepo.On("GetByID", ctx, paymentID).Return(nil, errors.New("payment not found"))

	// Execute
	payment, err := service.GetPayment(ctx, paymentID)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, payment)

	mockRepo.AssertExpectations(t)
}

// TestService_HandleWebhook_Generic_Success tests successful generic webhook handling
func TestService_HandleWebhook_Generic_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	payload := &WebhookPayload{
		PaymentID:   "payment-123",
		Status:      "success",
		Signature:   "valid-signature",
		ProviderRef: "provider-ref-123",
	}

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: "booking-123",
		Status:    StatusPending,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, payload.PaymentID).Return(existingPayment, nil)
	mockRepo.On("UpdateStatus", ctx, existingPayment.ID, StatusSuccess, payload.ProviderRef).Return(nil)
	mockEB.On("Publish", ctx, "payment.success", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	err := service.HandleWebhook(ctx, payload)

	// Assertions
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_HandleWebhook_Generic_Failed tests failed payment webhook
func TestService_HandleWebhook_Generic_Failed(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	payload := &WebhookPayload{
		PaymentID:   "payment-123",
		Status:      "failed",
		Signature:   "valid-signature",
		ProviderRef: "provider-ref-123",
	}

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: "booking-123",
		Status:    StatusPending,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, payload.PaymentID).Return(existingPayment, nil)
	mockRepo.On("UpdateStatus", ctx, existingPayment.ID, StatusFailed, payload.ProviderRef).Return(nil)
	mockEB.On("Publish", ctx, "payment.failed", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	err := service.HandleWebhook(ctx, payload)

	// Assertions
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_HandleWebhook_Generic_Refunded tests refund webhook
func TestService_HandleWebhook_Generic_Refunded(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	payload := &WebhookPayload{
		PaymentID:   "payment-123",
		Status:      "refund",
		Signature:   "valid-signature",
		ProviderRef: "provider-ref-123",
	}

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: "booking-123",
		Status:    StatusSuccess,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, payload.PaymentID).Return(existingPayment, nil)
	mockRepo.On("UpdateStatus", ctx, existingPayment.ID, StatusRefunded, payload.ProviderRef).Return(nil)
	mockEB.On("Publish", ctx, "payment.refunded", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Execute
	err := service.HandleWebhook(ctx, payload)

	// Assertions
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEB.AssertExpectations(t)
}

// TestService_HandleWebhook_PaymentNotFound tests webhook with non-existent payment
func TestService_HandleWebhook_PaymentNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	payload := &WebhookPayload{
		PaymentID: "non-existent-payment",
		Status:    "success",
		Signature: "valid-signature",
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, payload.PaymentID).Return(nil, errors.New("payment not found"))

	// Execute
	err := service.HandleWebhook(ctx, payload)

	// Assertions
	require.Error(t, err)

	mockRepo.AssertExpectations(t)
}

// TestService_HandleWebhook_InvalidSignature tests webhook with invalid signature
func TestService_HandleWebhook_InvalidSignature(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	payload := &WebhookPayload{
		PaymentID: "payment-123",
		Status:    "success",
		Signature: "", // Empty signature
	}

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: "booking-123",
		Status:    StatusPending,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, payload.PaymentID).Return(existingPayment, nil)

	// Execute
	err := service.HandleWebhook(ctx, payload)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid webhook signature")

	mockRepo.AssertExpectations(t)
}

// TestService_HandleWebhook_UpdateStatusError tests webhook with update status error
func TestService_HandleWebhook_UpdateStatusError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	payload := &WebhookPayload{
		PaymentID:   "payment-123",
		Status:      "success",
		Signature:   "valid-signature",
		ProviderRef: "provider-ref-123",
	}

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: "booking-123",
		Status:    StatusPending,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, payload.PaymentID).Return(existingPayment, nil)
	mockRepo.On("UpdateStatus", ctx, existingPayment.ID, StatusSuccess, payload.ProviderRef).Return(errors.New("database error"))

	// Execute
	err := service.HandleWebhook(ctx, payload)

	// Assertions
	require.Error(t, err)

	mockRepo.AssertExpectations(t)
}

// TestService_HandleWebhook_InvalidStatus tests webhook with invalid status
func TestService_HandleWebhook_InvalidStatus(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEB := new(MockEventBus)

	service := NewService(mockRepo, mockEB)

	ctx := context.Background()
	payload := &WebhookPayload{
		PaymentID: "payment-123",
		Status:    "invalid-status",
		Signature: "valid-signature",
	}

	existingPayment := &Payment{
		ID:        "payment-123",
		BookingID: "booking-123",
		Status:    StatusPending,
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, payload.PaymentID).Return(existingPayment, nil)

	// Execute
	err := service.HandleWebhook(ctx, payload)

	// Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payment status")

	mockRepo.AssertExpectations(t)
}

// TestNewPayment tests creating a new payment
func TestNewPayment(t *testing.T) {
	bookingID := "booking-123"
	req := &CreatePaymentRequest{
		BookingID: bookingID,
		Provider:  ProviderMidtrans,
		Method:    "gopay",
	}
	amount := 1000000

	payment := NewPayment(bookingID, req, amount)

	require.NotNil(t, payment)
	assert.NotEmpty(t, payment.ID)
	assert.Equal(t, bookingID, payment.BookingID)
	assert.Equal(t, req.Provider, payment.Provider)
	assert.Equal(t, req.Method, payment.Method)
	assert.Equal(t, amount, payment.Amount)
	assert.Equal(t, "IDR", payment.Currency)
	assert.Equal(t, StatusPending, payment.Status)
	assert.False(t, payment.ExpiresAt.IsZero())
	assert.False(t, payment.CreatedAt.IsZero())
}

// TestPaymentStatus_Constants tests payment status constants
func TestPaymentStatus_Constants(t *testing.T) {
	assert.Equal(t, PaymentStatus("PENDING"), StatusPending)
	assert.Equal(t, PaymentStatus("SUCCESS"), StatusSuccess)
	assert.Equal(t, PaymentStatus("FAILED"), StatusFailed)
	assert.Equal(t, PaymentStatus("REFUNDED"), StatusRefunded)
}

// TestPaymentProvider_Constants tests payment provider constants
func TestPaymentProvider_Constants(t *testing.T) {
	assert.Equal(t, PaymentProvider("midtrans"), ProviderMidtrans)
	assert.Equal(t, PaymentProvider("stripe"), ProviderStripe)
	assert.Equal(t, PaymentProvider("xendit"), ProviderXendit)
}

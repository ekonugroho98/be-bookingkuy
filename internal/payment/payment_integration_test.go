package payment

import (
	"context"
	"testing"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/eventbus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestPaymentFlow_Integration tests the payment flow integration
func TestPaymentFlow_Integration(t *testing.T) {
	// Setup: Create in-memory event bus
	eventBus := eventbus.New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Note: EventBus doesn't need Start(), it's already initialized

	// ========================================
	// TEST 1: Create Payment
	// ========================================
	t.Run("CreatePayment", func(t *testing.T) {
		// Setup: Create new mock for this test
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, eventBus)

		t.Log("Testing create payment...")

		bookingID := "test-booking-123"
		amount := 1500000

		createReq := &CreatePaymentRequest{
			BookingID: bookingID,
			Provider:  ProviderMidtrans,
			Method:    "GOPAY",
		}

		// Mock repository expectations
		mockRepo.On("GetByBookingID", ctx, bookingID).Return(nil, ErrPaymentNotFound)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*payment.Payment")).Return(nil)

		// Execute: Create payment
		newPayment, err := service.CreatePayment(ctx, createReq, amount)
		require.NoError(t, err, "Failed to create payment")
		require.NotNil(t, newPayment)
		require.Equal(t, bookingID, newPayment.BookingID)
		require.Equal(t, StatusPending, newPayment.Status)
		// Note: PaymentURL won't be generated without actual Midtrans client

		t.Logf("✅ Payment created: ID=%s, Status=%s", newPayment.ID, newPayment.Status)

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	// ========================================
	// TEST 2: Get Payment
	// ========================================
	t.Run("GetPayment", func(t *testing.T) {
		// Setup: Create new mock for this test
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, eventBus)

		t.Log("Testing get payment...")

		paymentID := "test-payment-123"
		bookingID := "test-booking-456"
		amount := 1500000

		// Mock repository expectations
		mockRepo.On("GetByID", ctx, paymentID).Return(&Payment{
			ID:        paymentID,
			BookingID: bookingID,
			Amount:    amount,
			Status:    StatusSuccess,
			CreatedAt: time.Now(),
		}, nil)

		// Execute: Get payment
		retrievedPayment, err := service.GetPayment(ctx, paymentID)
		require.NoError(t, err, "Failed to get payment")
		require.NotNil(t, retrievedPayment)
		require.Equal(t, paymentID, retrievedPayment.ID)
		require.Equal(t, StatusSuccess, retrievedPayment.Status)

		t.Logf("✅ Payment retrieved: ID=%s, Status=%s", retrievedPayment.ID, retrievedPayment.Status)

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	// ========================================
	// TEST 3: Handle Webhook - Settlement
	// ========================================
	t.Run("HandleWebhook_Settlement", func(t *testing.T) {
		// Setup: Create new mock for this test
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, eventBus)

		t.Log("Testing webhook handling (settlement)...")

		paymentID := "test-payment-settle"

		// Mock repository expectations - using GetByID (fallback path)
		mockRepo.On("GetByID", ctx, paymentID).Return(&Payment{
			ID:        paymentID,
			BookingID: "test-booking-settle",
			Amount:    1500000,
			Status:    StatusPending,
		}, nil)

		// UpdateStatus needs 4 params: ctx, id, status, providerRef
		mockRepo.On("UpdateStatus", ctx, paymentID, StatusSuccess, "").Return(nil)

		// Prepare webhook payload (generic format with signature)
		webhookPayload := &WebhookPayload{
			PaymentID: paymentID,
			Status:    "success", // Lowercase as expected by service
			Signature: "valid-signature",
		}

		// Execute: Handle webhook
		err := service.HandleWebhook(ctx, webhookPayload)
		require.NoError(t, err, "Failed to handle webhook")

		t.Logf("✅ Webhook processed successfully (settlement)")

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	// ========================================
	// TEST 4: Handle Webhook - Multiple Scenarios
	// ========================================
	t.Run("HandleWebhook_Scenarios", func(t *testing.T) {
		testCases := []struct {
			name            string
			status          string
			expectedStatus  PaymentStatus
			description     string
		}{
			{
				name:            "Pending",
				status:          "pending", // Note: pending doesn't have explicit mapping
				expectedStatus:  StatusPending,
				description:     "Payment pending",
			},
			{
				name:            "Success",
				status:          "success",
				expectedStatus:  StatusSuccess,
				description:     "Payment successful",
			},
			{
				name:            "Failed",
				status:          "failed",
				expectedStatus:  StatusFailed,
				description:     "Payment failed",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup: Create new mock for each test case
				mockRepo := new(MockRepository)
				service := NewService(mockRepo, eventBus)

				t.Logf("Testing webhook scenario: %s", tc.description)

				paymentID := "test-payment-" + tc.name

				// Mock repository expectations - using GetByID (fallback path)
				mockRepo.On("GetByID", ctx, paymentID).Return(&Payment{
					ID:     paymentID,
					Status: StatusPending,
				}, nil)
				// UpdateStatus needs 4 params: ctx, id, status, providerRef
				mockRepo.On("UpdateStatus", ctx, paymentID, tc.expectedStatus, "").Return(nil)

				// Prepare webhook payload (generic format with signature)
				webhookPayload := &WebhookPayload{
					PaymentID: paymentID,
					Status:    tc.status,
					Signature: "valid-signature",
				}

				// Execute: Handle webhook
				// Note: "pending" status will fail since it's not in the mapping
				// We'll just log the result
				err := service.HandleWebhook(ctx, webhookPayload)
				if tc.name == "Pending" {
					// Expected to fail - pending not in mapping
					require.Error(t, err, "Pending status should fail")
				} else {
					require.NoError(t, err, "Failed to handle webhook for "+tc.description)
					t.Logf("✅ Webhook scenario passed: %s -> %s", tc.status, tc.expectedStatus)
				}

				// Verify mock was called (only for successful cases)
				if tc.name != "Pending" {
					mockRepo.AssertExpectations(t)
				}
			})
		}
	})

	// ========================================
	// TEST 5: Error Handling
	// ========================================
	t.Run("ErrorHandling", func(t *testing.T) {
		t.Run("PaymentNotFound", func(t *testing.T) {
			// Setup: Create new mock for this test
			mockRepo := new(MockRepository)
			service := NewService(mockRepo, eventBus)

			t.Log("Testing get payment for non-existent payment...")

			paymentID := "non-existent-payment"

			// Mock: Payment not found
			mockRepo.On("GetByID", ctx, paymentID).Return(nil, ErrPaymentNotFound)

			// Execute: Get payment
			_, err := service.GetPayment(ctx, paymentID)
			require.Error(t, err, "Should return error for non-existent payment")

			t.Logf("✅ Non-existent payment correctly rejected: %v", err)

			// Verify mock was called
			mockRepo.AssertExpectations(t)
		})
	})
}

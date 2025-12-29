package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// MockHotelBedsServer creates a mock HotelBeds API server
type MockHotelBedsServer struct {
	Server *httptest.Server
	T      *testing.T
}

// NewMockHotelBedsServer creates a new mock HotelBeds server
func NewMockHotelBedsServer(t *testing.T) *MockHotelBedsServer {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		require.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Return mock response based on endpoint
		switch r.URL.Path {
		case "/hotel-api/1.0/hotels":
			// Availability check
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"hotels": []map[string]interface{}{
					{
						"code":       "test-hotel-123",
						"name":       "Test Hotel Bali",
						"category":   "4STAR",
						"country":    "Indonesia",
						"city":       "Bali",
						"rating":     4.5,
						"isAvailable": true,
						"rooms": []map[string]interface{}{
							{
								"roomCode":  "test-room-123",
								"roomName":  "Deluxe Room",
								"available": true,
								"price":     1500000,
								"currency":  "IDR",
							},
						},
						"totalPrice": 1500000,
						"currency":   "IDR",
					},
				},
			})

		case "/hotel-api/1.0/bookings":
			// Create booking
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"bookingReference": "HB-TEST-12345",
				"status":          "CONFIRMED",
				"hotel": map[string]interface{}{
					"code": "test-hotel-123",
					"name": "Test Hotel Bali",
				},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "endpoint not found",
			})
		}
	}))

	return &MockHotelBedsServer{
		Server: server,
		T:      t,
	}
}

// Close closes the mock server
func (m *MockHotelBedsServer) Close() {
	m.Server.Close()
}

// URL returns the mock server URL
func (m *MockHotelBedsServer) URL() string {
	return m.Server.URL
}

// MockMidtransServer creates a mock Midtrans API server
type MockMidtransServer struct {
	Server *httptest.Server
	T      *testing.T
}

// NewMockMidtransServer creates a new mock Midtrans server
func NewMockMidtransServer(t *testing.T) *MockMidtransServer {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization
		require.Equal(t, "SB-Mid-server-test-key", r.Header.Get("Authorization"))

		// Return mock response based on endpoint
		switch r.URL.Path {
		case "/v2/charge":
			// Create payment
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status_code":     "201",
				"status_message":  "Success, Transaction is found",
				"transaction_id":  "test-trx-123",
				"order_id":        "test-order-123",
				"payment_type":    "GOPAY",
				"redirect_url":    "https://mock-midtrans.com/payment",
				"transaction_status": "pending",
				"gross_amount":    "1500000.00",
			})

		case "/v2/test-order-123/status":
			// Check status
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status_code":     "200",
				"status_message":  "Success, Transaction is found",
				"transaction_id":  "test-trx-123",
				"order_id":        "test-order-123",
				"payment_type":    "GOPAY",
				"transaction_status": "settlement",
				"gross_amount":    "1500000.00",
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "endpoint not found",
			})
		}
	}))

	return &MockMidtransServer{
		Server: server,
		T:      t,
	}
}

// Close closes the mock server
func (m *MockMidtransServer) Close() {
	m.Server.Close()
}

// URL returns the mock server URL
func (m *MockMidtransServer) URL() string {
	return m.Server.URL
}

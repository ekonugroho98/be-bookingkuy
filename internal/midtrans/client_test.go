package midtrans

import (
	"crypto/sha512"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient tests creating a new Midtrans client
func TestNewClient(t *testing.T) {
	config := Config{
		ServerKey:  "test-server-key",
		ClientKey:  "test-client-key",
		MerchantID: "test-merchant",
	}

	client := NewClient(config)

	require.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, SandboxBaseURL, client.baseURL)
	assert.Equal(t, SandboxSnapURL, client.snapURL)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
	// Note: client.config.Timeout gets set to default (30s) if 0
	assert.Equal(t, 30*time.Second, client.config.Timeout)
}

// TestNewClient_Production tests creating client for production environment
func TestNewClient_Production(t *testing.T) {
	config := Config{
		ServerKey:    "prod-server-key",
		ClientKey:    "prod-client-key",
		MerchantID:   "prod-merchant",
		IsProduction: true,
	}

	client := NewClient(config)

	require.NotNil(t, client)
	assert.Equal(t, ProductionBaseURL, client.baseURL)
	assert.Equal(t, ProductionSnapURL, client.snapURL)
}

// TestNewClient_CustomTimeout tests creating client with custom timeout
func TestNewClient_CustomTimeout(t *testing.T) {
	config := Config{
		ServerKey:  "test-server-key",
		ClientKey:  "test-client-key",
		MerchantID: "test-merchant",
		Timeout:    60 * time.Second,
	}

	client := NewClient(config)

	require.NotNil(t, client)
	assert.Equal(t, 60*time.Second, client.httpClient.Timeout)
}

// TestValidateWebhookSignature_ValidSignature tests valid webhook signature
func TestValidateWebhookSignature_ValidSignature(t *testing.T) {
	config := Config{
		ServerKey: "SB-Mid-server-TEST123",
	}

	client := NewClient(config)

	orderID := "ORDER-123"
	statusCode := "200"
	grossAmount := "100000.00"

	// Calculate expected signature
	data := orderID + statusCode + grossAmount + config.ServerKey
	hash := sha512.Sum512([]byte(data))
	expectedSignature := fmt.Sprintf("%x", hash)

	// Test with valid signature
	isValid := client.ValidateWebhookSignature(orderID, statusCode, grossAmount, expectedSignature)

	assert.True(t, isValid)
}

// TestValidateWebhookSignature_InvalidSignature tests invalid webhook signature
func TestValidateWebhookSignature_InvalidSignature(t *testing.T) {
	config := Config{
		ServerKey: "SB-Mid-server-TEST123",
	}

	client := NewClient(config)

	// Test with invalid signature
	isValid := client.ValidateWebhookSignature("ORDER-123", "200", "100000.00", "invalid-signature")

	assert.False(t, isValid)
}

// TestValidateWebhookSignature_WrongServerKey tests signature with different server key
func TestValidateWebhookSignature_WrongServerKey(t *testing.T) {
	config1 := Config{
		ServerKey: "SB-Mid-server-KEY1",
	}
	config2 := Config{
		ServerKey: "SB-Mid-server-KEY2",
	}

	client1 := NewClient(config1)
	client2 := NewClient(config2)

	orderID := "ORDER-123"
	statusCode := "200"
	grossAmount := "100000.00"

	// Calculate signature with KEY1
	data := orderID + statusCode + grossAmount + config1.ServerKey
	hash := sha512.Sum512([]byte(data))
	signature := fmt.Sprintf("%x", hash)

	// Validate with KEY1 should be valid
	isValid1 := client1.ValidateWebhookSignature(orderID, statusCode, grossAmount, signature)
	assert.True(t, isValid1)

	// Validate with KEY2 should be invalid
	isValid2 := client2.ValidateWebhookSignature(orderID, statusCode, grossAmount, signature)
	assert.False(t, isValid2)
}

// TestMapTransactionStatus tests mapping Midtrans transaction statuses
func TestMapTransactionStatus(t *testing.T) {
	tests := []struct {
		name           string
		midtransStatus TransactionStatus
		expected       string
	}{
		{name: "Pending", midtransStatus: StatusPending, expected: "pending"},
		{name: "Authorize", midtransStatus: StatusAuthorize, expected: "pending"},
		{name: "Capture", midtransStatus: StatusCapture, expected: "success"},
		{name: "Settlement", midtransStatus: StatusSettlement, expected: "success"},
		{name: "Deny", midtransStatus: StatusDeny, expected: "failed"},
		{name: "Failure", midtransStatus: StatusFailure, expected: "failed"},
		{name: "Expire", midtransStatus: StatusExpire, expected: "failed"},
		{name: "Cancel", midtransStatus: StatusCancel, expected: "cancelled"},
		{name: "Pending Cancel", midtransStatus: StatusPendingCancel, expected: "cancelled"},
		{name: "Refund", midtransStatus: StatusRefund, expected: "refunded"},
		{name: "Partial Refund", midtransStatus: StatusPartialRefund, expected: "refunded"},
		{name: "Unknown", midtransStatus: TransactionStatus("unknown"), expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapTransactionStatus(tt.midtransStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsTransactionFinal tests checking if transaction is final
func TestIsTransactionFinal(t *testing.T) {
	tests := []struct {
		name     string
		status   TransactionStatus
		expected bool
	}{
		{name: "Pending is not final", status: StatusPending, expected: false},
		{name: "Authorize is not final", status: StatusAuthorize, expected: false},
		{name: "Capture is final", status: StatusCapture, expected: true},
		{name: "Settlement is final", status: StatusSettlement, expected: true},
		{name: "Deny is final", status: StatusDeny, expected: true},
		{name: "Cancel is final", status: StatusCancel, expected: true},
		{name: "Expire is final", status: StatusExpire, expected: true},
		{name: "Failure is final", status: StatusFailure, expected: true},
		{name: "Refund is final", status: StatusRefund, expected: true},
		{name: "Partial Refund is final", status: StatusPartialRefund, expected: true},
		{name: "Pending Cancel is not final", status: StatusPendingCancel, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTransactionFinal(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPaymentType_Constants tests payment type constants
func TestPaymentType_Constants(t *testing.T) {
	assert.Equal(t, PaymentType("credit_card"), PaymentTypeCreditCard)
	assert.Equal(t, PaymentType("bank_transfer"), PaymentTypeBankTransfer)
	assert.Equal(t, PaymentType("gopay"), PaymentTypeGopay)
	assert.Equal(t, PaymentType("qris"), PaymentTypeQRIS)
	assert.Equal(t, PaymentType("shopeepay"), PaymentTypeShopeePay)
}

// TestTransactionStatus_Constants tests transaction status constants
func TestTransactionStatus_Constants(t *testing.T) {
	assert.Equal(t, TransactionStatus("pending"), StatusPending)
	assert.Equal(t, TransactionStatus("authorize"), StatusAuthorize)
	assert.Equal(t, TransactionStatus("capture"), StatusCapture)
	assert.Equal(t, TransactionStatus("settlement"), StatusSettlement)
	assert.Equal(t, TransactionStatus("deny"), StatusDeny)
	assert.Equal(t, TransactionStatus("pending_cancel"), StatusPendingCancel)
	assert.Equal(t, TransactionStatus("cancel"), StatusCancel)
	assert.Equal(t, TransactionStatus("expire"), StatusExpire)
	assert.Equal(t, TransactionStatus("failure"), StatusFailure)
	assert.Equal(t, TransactionStatus("refund"), StatusRefund)
	assert.Equal(t, TransactionStatus("partial_refund"), StatusPartialRefund)
}

// TestConfig_DefaultValues tests config default values
func TestConfig_DefaultValues(t *testing.T) {
	config := Config{
		ServerKey: "test-key",
	}

	client := NewClient(config)

	// Should default to 30 seconds timeout
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)

	// Should default to sandbox URLs
	assert.Equal(t, SandboxBaseURL, client.baseURL)
	assert.Equal(t, SandboxSnapURL, client.snapURL)
}

// TestValidateWebhookSignature_EmptyInputs tests signature validation with empty inputs
func TestValidateWebhookSignature_EmptyInputs(t *testing.T) {
	config := Config{
		ServerKey: "SB-Mid-server-TEST123",
	}

	client := NewClient(config)

	// Calculate signature for empty values
	data := "" + "" + "" + config.ServerKey
	hash := sha512.Sum512([]byte(data))
	expectedSignature := fmt.Sprintf("%x", hash)

	// Should still work with empty values
	isValid := client.ValidateWebhookSignature("", "", "", expectedSignature)
	assert.True(t, isValid)
}

// TestValidateWebhookSignature_SpecialCharacters tests signature with special characters
func TestValidateWebhookSignature_SpecialCharacters(t *testing.T) {
	config := Config{
		ServerKey: "SB-Mid-server-TEST!@#$%",
	}

	client := NewClient(config)

	orderID := "ORDER-123"
	statusCode := "200"
	grossAmount := "100,000.00" // With comma

	// Calculate signature
	data := orderID + statusCode + grossAmount + config.ServerKey
	hash := sha512.Sum512([]byte(data))
	expectedSignature := fmt.Sprintf("%x", hash)

	isValid := client.ValidateWebhookSignature(orderID, statusCode, grossAmount, expectedSignature)
	assert.True(t, isValid)
}

// TestMapTransactionStatus_AllStatuses tests all status mappings cover all constants
func TestMapTransactionStatus_AllStatuses(t *testing.T) {
	// All defined status constants should return a valid mapping
	statuses := []TransactionStatus{
		StatusPending,
		StatusAuthorize,
		StatusCapture,
		StatusSettlement,
		StatusDeny,
		StatusPendingCancel,
		StatusCancel,
		StatusExpire,
		StatusFailure,
		StatusRefund,
		StatusPartialRefund,
	}

	validMappings := map[string]bool{
		"pending":    true,
		"success":    true,
		"failed":     true,
		"cancelled":  true,
		"refunded":   true,
		"unknown":    true,
	}

	for _, status := range statuses {
		result := MapTransactionStatus(status)
		assert.True(t, validMappings[result], "Status %s mapped to invalid value: %s", status, result)
	}
}

// TestIsTransactionFinal_Consistency tests consistency with MapTransactionStatus
func TestIsTransactionStatus_Consistency(t *testing.T) {
	// Final statuses should map to terminal states (success, failed, cancelled, refunded)
	finalStatuses := []TransactionStatus{
		StatusSettlement,
		StatusCapture,
		StatusCancel,
		StatusDeny,
		StatusExpire,
		StatusFailure,
		StatusRefund,
		StatusPartialRefund,
	}

	for _, status := range finalStatuses {
		mapped := MapTransactionStatus(status)
		isFinal := IsTransactionFinal(status)
		assert.True(t, isFinal, "Status %s should be final", status)
		assert.Contains(t, []string{"success", "failed", "cancelled", "refunded"}, mapped,
			"Final status %s should map to terminal state, got: %s", status, mapped)
	}
}

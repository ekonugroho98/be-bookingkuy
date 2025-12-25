package payment

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "PENDING"
	StatusSuccess  PaymentStatus = "SUCCESS"
	StatusFailed   PaymentStatus = "FAILED"
	StatusRefunded PaymentStatus = "REFUNDED"
)

// PaymentProvider represents payment provider
type PaymentProvider string

const (
	ProviderMidtrans PaymentProvider = "midtrans"
	ProviderStripe   PaymentProvider = "stripe"
	ProviderXendit   PaymentProvider = "xendit"
)

// Payment represents a payment
type Payment struct {
	ID               string          `json:"id" db:"id"`
	BookingID        string          `json:"booking_id" db:"booking_id"`
	Provider         PaymentProvider `json:"provider" db:"provider"`
	Method           string          `json:"method" db:"method"`
	Amount           int             `json:"amount" db:"amount"`
	Currency         string          `json:"currency" db:"currency"`
	Status           PaymentStatus   `json:"status" db:"status"`
	ProviderRef      string          `json:"provider_reference,omitempty" db:"provider_reference"`
	PaymentURL       string          `json:"payment_url,omitempty"`
	ExpiresAt        time.Time       `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}

// CreatePaymentRequest represents request to create payment
type CreatePaymentRequest struct {
	BookingID string       `json:"booking_id" validate:"required"`
	Provider  PaymentProvider `json:"provider" validate:"required,oneof=midtrans stripe xendit"`
	Method    string       `json:"method" validate:"required"`
}

// NewPayment creates a new payment
func NewPayment(bookingID string, req *CreatePaymentRequest, amount int) *Payment {
	now := time.Now()
	return &Payment{
		ID:        uuid.New().String(),
		BookingID: bookingID,
		Provider:  req.Provider,
		Method:    req.Method,
		Amount:    amount,
		Currency:  "IDR",
		Status:    StatusPending,
		ExpiresAt: now.Add(24 * time.Hour), // Payment expires in 24 hours
		CreatedAt: now,
	}
}

// WebhookPayload represents payment webhook payload
type WebhookPayload struct {
	PaymentID       string `json:"payment_id"`
	ProviderRef     string `json:"provider_reference"`
	Status          string `json:"status"`
	Signature       string `json:"signature"`

	// Midtrans specific fields
	OrderID         string `json:"order_id,omitempty"`
	StatusCode      string `json:"status_code,omitempty"`
	GrossAmount     string `json:"gross_amount,omitempty"`
	SignatureKey    string `json:"signature_key,omitempty"`
	PaymentType     string `json:"payment_type,omitempty"`
	TransactionStatus string `json:"transaction_status,omitempty"`
}

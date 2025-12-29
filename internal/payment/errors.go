package payment

import "errors"

// Package-level errors for payment operations
var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrInvalidPayment     = errors.New("invalid payment")
	ErrPaymentFailed      = errors.New("payment failed")
	ErrInvalidAmount      = errors.New("invalid amount")
)

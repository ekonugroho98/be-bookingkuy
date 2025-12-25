package payment

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/jackc/pgx/v5"
)

// Repository defines interface for payment data operations
type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	GetByBookingID(ctx context.Context, bookingID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id string, status PaymentStatus, providerRef string) error
}

type repository struct {
	db *db.DB
}

// NewRepository creates a new payment repository
func NewRepository(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

func (r *repository) Create(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (id, booking_id, provider, method, amount, currency, status, provider_reference, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		payment.ID, payment.BookingID, payment.Provider, payment.Method,
		payment.Amount, payment.Currency, payment.Status, payment.ProviderRef, payment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Payment, error) {
	query := `
		SELECT id, booking_id, provider, method, amount, currency, status, provider_reference, created_at
		FROM payments
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)

	var payment Payment
	err := row.Scan(
		&payment.ID, &payment.BookingID, &payment.Provider, &payment.Method,
		&payment.Amount, &payment.Currency, &payment.Status, &payment.ProviderRef, &payment.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

func (r *repository) GetByBookingID(ctx context.Context, bookingID string) (*Payment, error) {
	query := `
		SELECT id, booking_id, provider, method, amount, currency, status, provider_reference, created_at
		FROM payments
		WHERE booking_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.Pool.QueryRow(ctx, query, bookingID)

	var payment Payment
	err := row.Scan(
		&payment.ID, &payment.BookingID, &payment.Provider, &payment.Method,
		&payment.Amount, &payment.Currency, &payment.Status, &payment.ProviderRef, &payment.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment by booking ID: %w", err)
	}

	return &payment, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status PaymentStatus, providerRef string) error {
	query := `
		UPDATE payments
		SET status = $2, provider_reference = $3
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, id, status, providerRef)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

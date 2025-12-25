package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/jackc/pgx/v5"
)

// Repository defines interface for booking data operations
type Repository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id string) (*Booking, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Booking, error)
	Update(ctx context.Context, booking *Booking) error
	UpdateStatus(ctx context.Context, id string, status BookingStatus) error
}

type repository struct {
	db *db.DB
}

// NewRepository creates a new booking repository
func NewRepository(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

func (r *repository) Create(ctx context.Context, booking *Booking) error {
	query := `
		INSERT INTO bookings (id, user_id, hotel_id, room_id, booking_reference, check_in, check_out, guests, status, total_amount, currency, payment_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		booking.ID, booking.UserID, booking.HotelID, booking.RoomID,
		booking.BookingReference, booking.CheckIn, booking.CheckOut,
		booking.Guests, booking.Status, booking.TotalAmount,
		booking.Currency, booking.PaymentType, booking.CreatedAt, booking.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Booking, error) {
	query := `
		SELECT id, user_id, hotel_id, room_id, booking_reference, supplier_reference,
		       check_in, check_out, guests, status, total_amount, currency, payment_type,
		       created_at, updated_at
		FROM bookings
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)

	var booking Booking
	err := row.Scan(
		&booking.ID, &booking.UserID, &booking.HotelID, &booking.RoomID,
		&booking.BookingReference, &booking.SupplierReference,
		&booking.CheckIn, &booking.CheckOut, &booking.Guests, &booking.Status,
		&booking.TotalAmount, &booking.Currency, &booking.PaymentType,
		&booking.CreatedAt, &booking.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Booking, error) {
	query := `
		SELECT id, user_id, hotel_id, room_id, booking_reference, supplier_reference,
		       check_in, check_out, guests, status, total_amount, currency, payment_type,
		       created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		var booking Booking
		err := rows.Scan(
			&booking.ID, &booking.UserID, &booking.HotelID, &booking.RoomID,
			&booking.BookingReference, &booking.SupplierReference,
			&booking.CheckIn, &booking.CheckOut, &booking.Guests, &booking.Status,
			&booking.TotalAmount, &booking.Currency, &booking.PaymentType,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	return bookings, nil
}

func (r *repository) Update(ctx context.Context, booking *Booking) error {
	query := `
		UPDATE bookings
		SET status = $2, supplier_reference = $3, total_amount = $4, updated_at = $5
		WHERE id = $1
	`

	booking.UpdatedAt = time.Now()

	result, err := r.db.Pool.Exec(ctx, query,
		booking.ID, booking.Status, booking.SupplierReference,
		booking.TotalAmount, booking.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("booking not found")
	}

	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status BookingStatus) error {
	query := `
		UPDATE bookings
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	return nil
}

package notification

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
)

// Repository defines interface for notification data operations
type Repository interface {
	Create(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, id string) (*Notification, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type repository struct {
	db *db.DB
}

// NewRepository creates a new notification repository
func NewRepository(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

func (r *repository) Create(ctx context.Context, notification *Notification) error {
	query := `
		INSERT INTO notifications (id, type, to, subject, message, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		notification.ID, notification.Type, notification.To,
		notification.Subject, notification.Message, notification.Status,
		notification.Metadata,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Notification, error) {
	query := `
		SELECT id, type, to, subject, message, status, metadata
		FROM notifications
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)

	var notification Notification
	err := row.Scan(
		&notification.ID, &notification.Type, &notification.To,
		&notification.Subject, &notification.Message, &notification.Status,
		&notification.Metadata,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE notifications
		SET status = $2
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

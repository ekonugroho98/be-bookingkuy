package auth

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/jackc/pgx/v5"
)

// Repository defines interface for auth data operations
type Repository interface {
	StorePassword(ctx context.Context, userID string, hashedPassword string) error
	GetPassword(ctx context.Context, email string) (string, string, error)
}

type repository struct {
	db *db.DB
}

// NewRepository creates a new auth repository
func NewRepository(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

func (r *repository) StorePassword(ctx context.Context, userID string, hashedPassword string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, userID, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to store password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *repository) GetPassword(ctx context.Context, email string) (string, string, error) {
	query := `
		SELECT id, password_hash
		FROM users
		WHERE email = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, email)

	var userID, passwordHash string
	err := row.Scan(&userID, &passwordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", fmt.Errorf("user not found")
		}
		return "", "", fmt.Errorf("failed to get password: %w", err)
	}

	return userID, passwordHash, nil
}

package user

import (
	"context"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/jackc/pgx/v5"
)

// Repository defines the interface for user data operations
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}

type repository struct {
	db *db.DB
}

// NewRepository creates a new user repository
func NewRepository(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, name, email, email_verified, phone, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID, user.Name, user.Email, user.EmailVerified,
		user.Phone, user.Role, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, name, email, email_verified, phone, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)

	var user User
	err := row.Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified,
		&user.Phone, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, name, email, email_verified, phone, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, email)

	var user User
	err := row.Scan(
		&user.ID, &user.Name, &user.Email, &user.EmailVerified,
		&user.Phone, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *repository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET name = $2, email_verified = $3, phone = $4, updated_at = $5
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.Pool.Exec(ctx, query,
		user.ID, user.Name, user.EmailVerified, user.Phone, user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

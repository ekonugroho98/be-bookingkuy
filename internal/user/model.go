package user

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents user role
type UserRole string

const (
	UserRoleAdmin UserRole = "ADMIN"
	UserRoleUser  UserRole = "USER"
)

// User represents a user in the system
type User struct {
	ID            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"email" db:"email"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	Phone         string    `json:"phone,omitempty" db:"phone"`
	Role          UserRole   `json:"role" db:"role"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser creates a new user instance
func NewUser(name, email string) *User {
	now := time.Now()
	return &User{
		ID:            uuid.New().String(),
		Name:          name,
		Email:         email,
		EmailVerified: false,
		Role:          UserRoleUser,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// CreateUserRequest represents request to create/update user
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone,omitempty" validate:"omitempty,max=20"`
}

// UpdateUserRequest represents request to update user profile
type UpdateUserRequest struct {
	Name  *string `json:"name" validate:"omitempty,min=2,max=100"`
	Phone *string `json:"phone" validate:"omitempty,max=20"`
}

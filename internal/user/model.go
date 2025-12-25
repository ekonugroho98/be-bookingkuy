package user

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Phone        string    `json:"phone,omitempty"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

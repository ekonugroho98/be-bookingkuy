package user

import "errors"

// Package-level errors for user operations
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

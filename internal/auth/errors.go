package auth

import "errors"

// Package-level errors for authentication operations
var (
	ErrUserExists        = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token expired")
)

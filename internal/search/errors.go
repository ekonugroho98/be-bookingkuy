package search

import "errors"

// Package-level errors for search operations
var (
	ErrInvalidDates    = errors.New("invalid date range")
	ErrInvalidGuests   = errors.New("invalid number of guests")
	ErrNoResults       = errors.New("no search results found")
	ErrMissingLocation = errors.New("location is required")
)

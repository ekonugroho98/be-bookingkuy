package search

import "time"

// SearchRequest represents a hotel search request
type SearchRequest struct {
	CheckIn  time.Time `json:"check_in" validate:"required"`
	CheckOut time.Time `json:"check_out" validate:"required,gtfield=CheckIn"`
	City     string    `json:"city" validate:"required"`
	Guests   int       `json:"guests" validate:"required,min=1,max=10"`
	MinPrice *int      `json:"min_price,omitempty" validate:"omitempty,min=0"`
	MaxPrice *int      `json:"max_price,omitempty" validate:"omitempty,min=0"`
}

// SortBy represents sort options
type SortBy string

const (
	SortByPrice  SortBy = "price"
	SortByRating SortBy = "rating"
	SortByName   SortBy = "name"
)

// SearchOptions represents search options
type SearchOptions struct {
	Page    int    `json:"page" validate:"min=1"`
	PerPage int    `json:"per_page" validate:"min=1,max=100"`
	SortBy  SortBy `json:"sort_by" validate:"omitempty,oneof=price rating name"`
}

// Hotel represents a hotel in search results
type Hotel struct {
	ID          string   `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Country     string   `json:"country" db:"country"`
	City        string   `json:"city" db:"city"`
	Rating      *float64 `json:"rating" db:"rating"`
	MinPrice    *int     `json:"min_price,omitempty"`
	MaxPrice    *int     `json:"max_price,omitempty"`
	Description string   `json:"description,omitempty"`
}

// SearchResult represents search results with pagination
type SearchResult struct {
	Hotels     []Hotel `json:"hotels"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PerPage    int     `json:"per_page"`
	TotalPages int     `json:"total_pages"`
}

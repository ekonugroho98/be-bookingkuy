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
	CountryCode string   `json:"country_code" db:"country_code"`
	City        string   `json:"city" db:"city"`
	Rating      *float64 `json:"rating" db:"overall_rating"`
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

// AutocompleteRequest represents autocomplete search request
type AutocompleteRequest struct {
	Query string `json:"query" validate:"required,min=2"`
	Limit int    `json:"limit" validate:"omitempty,min=1,max=20"`
}

// AutocompleteResultType represents the type of autocomplete result
type AutocompleteResultType string

const (
	AutocompleteTypeRegion AutocompleteResultType = "region" // Country/State level
	AutocompleteTypeCity   AutocompleteResultType = "city"   // City level
	AutocompleteTypeHotel  AutocompleteResultType = "hotel"  // Specific hotel
)

// AutocompleteResult represents autocomplete result
type AutocompleteResult struct {
	Type        AutocompleteResultType `json:"type"`         // region, city, hotel
	ID          string                 `json:"id"`           // For region/city: code, for hotel: hotel_id
	Name        string                 `json:"name"`         // Display name
	FullName    string                 `json:"full_name"`    // Full name with location
	City        string                 `json:"city,omitempty"`      // City name (for hotels)
	CountryCode string                 `json:"country_code,omitempty"` // Country code (e.g., "ID")
	CountryName string                 `json:"country_name,omitempty"` // Country name (e.g., "Indonesia")
	Region      string                 `json:"region,omitempty"`      // State/Province (optional)
}

// AutocompleteResponse represents autocomplete response
type AutocompleteResponse struct {
	Query   string                `json:"query"`
	Results []AutocompleteResult  `json:"results"`
}

// AutocompleteOptions represents autocomplete options
type AutocompleteOptions struct {
	Limit int
}

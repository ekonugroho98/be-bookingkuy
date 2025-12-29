package sync

import "time"

// Destination represents a synced destination from HotelBeds
type Destination struct {
	ID           string     `db:"id" json:"id"`
	Code         string     `db:"code" json:"code"`             // HotelBeds code
	Name         string     `db:"name" json:"name"`
	CountryCode  string     `db:"country_code" json:"country_code"`
	CountryName  string     `db:"country_name" json:"country_name"`
	Type         string     `db:"type" json:"type"`              // CITY, REGION, COUNTRY
	ParentCode   *string    `db:"parent_code" json:"parent_code,omitempty"`
	Latitude     *float64   `db:"latitude" json:"latitude,omitempty"`
	Longitude    *float64   `db:"longitude" json:"longitude,omitempty"`
	HotelbedsData []byte    `db:"hotelbeds_data" json:"-"`      // Raw JSON
	SyncedAt     time.Time  `db:"synced_at" json:"synced_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// SyncOptions represents options for sync operation
type SyncOptions struct {
	CountryCode string // Filter by country code (empty = all countries)
	Limit       int    // Max records to sync (0 = unlimited)
	BatchSize   int    // Records per batch
	DryRun      bool   // If true, don't save to database
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Total     int           `json:"total"`
	Inserted  int           `json:"inserted"`
	Updated   int           `json:"updated"`
	Failed    int           `json:"failed"`
	Skipped   int           `json:"skipped"`
	Duration  time.Duration `json:"duration"`
	Errors    []SyncError   `json:"errors,omitempty"`
}

// SyncError represents an error during sync
type SyncError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Record      string `json:"record,omitempty"` // Record identifier
	Recoverable bool   `json:"recoverable"`
}

// Hotel represents a synced hotel from HotelBeds
type Hotel struct {
	ID               string    `db:"id" json:"id"`
	ProviderCode     string    `db:"provider_code" json:"provider_code"`         // "hotelbeds"
	ProviderHotelID  string    `db:"provider_hotel_id" json:"provider_hotel_id"` // HotelBeds code
	Name             string    `db:"name" json:"name"`
	Description      string    `db:"description" json:"description"`
	StarRating       float64   `db:"star_rating" json:"star_rating"`
	CountryCode      string    `db:"country_code" json:"country_code"`
	City             string    `db:"city" json:"city"`
	Address          string    `db:"address" json:"address"`
	PostalCode       string    `db:"postal_code" json:"postal_code,omitempty"`
	Latitude         *float64  `db:"latitude" json:"latitude,omitempty"`
	Longitude        *float64  `db:"longitude" json:"longitude,omitempty"`
	Images           []byte    `db:"images" json:"-"`          // JSONB as byte array
	Amenities        []byte    `db:"amenities" json:"-"`       // JSONB as byte array
	DestinationCode  int       `db:"-" json:"-"`              // HotelBeds destination code (not in DB)
	HotelbedsData    []byte    `db:"-" json:"-"`              // Raw HotelBeds JSON
	SyncedAt         time.Time `db:"-" json:"-"`              // Not in DB table
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

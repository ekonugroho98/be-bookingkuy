package search

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
)

// Repository defines interface for search data operations
type Repository interface {
	SearchHotels(ctx context.Context, req *SearchRequest, opts *SearchOptions) (*SearchResult, error)
	Autocomplete(ctx context.Context, query string, opts *AutocompleteOptions) (*AutocompleteResponse, error)
	GetPopularDestinations(ctx context.Context, limit int) ([]AutocompleteResult, error)
}

type repository struct {
	db *db.DB
}

// NewRepository creates a new search repository
func NewRepository(database *db.DB) Repository {
	return &repository{
		db: database,
	}
}

func (r *repository) SearchHotels(ctx context.Context, req *SearchRequest, opts *SearchOptions) (*SearchResult, error) {
	// Build base query
	baseQuery := `
		SELECT id, name, country_code, city, overall_rating
		FROM hotels
		WHERE city ILIKE $1
			AND deleted_at IS NULL
	`

	countQuery := `
		SELECT COUNT(*)
		FROM hotels
		WHERE city ILIKE $1
			AND deleted_at IS NULL
	`

	args := []interface{}{req.City}
	argPos := 2

	// Add price filters if provided
	if req.MinPrice != nil {
		baseQuery += fmt.Sprintf(" AND $%d >= $%d", argPos, argPos+1)
		countQuery += fmt.Sprintf(" AND $%d >= $%d", argPos, argPos+1)
		args = append(args, *req.MinPrice)
		argPos += 2
	}
	if req.MaxPrice != nil {
		baseQuery += fmt.Sprintf(" AND $%d <= $%d", argPos, argPos+1)
		countQuery += fmt.Sprintf(" AND $%d <= $%d", argPos, argPos+1)
		args = append(args, *req.MaxPrice)
		argPos += 2
	}

	// Add sorting
	sortClause := " ORDER BY name ASC"
	if opts.SortBy != "" {
		switch opts.SortBy {
		case SortByPrice:
			sortClause = " ORDER BY name ASC" // TODO: Add price sorting when pricing is implemented
		case SortByRating:
			sortClause = " ORDER BY overall_rating DESC NULLS LAST, name ASC"
		case SortByName:
			sortClause = " ORDER BY name ASC"
		}
	}
	baseQuery += sortClause

	// Add pagination
	limit := opts.PerPage
	offset := (opts.Page - 1) * opts.PerPage
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	// Get total count
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, args[0]).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count hotels: %w", err)
	}

	// Execute search query
	rows, err := r.db.Pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search hotels: %w", err)
	}
	defer rows.Close()

	var hotels []Hotel
	for rows.Next() {
		var hotel Hotel
		err := rows.Scan(
			&hotel.ID,
			&hotel.Name,
			&hotel.CountryCode,
			&hotel.City,
			&hotel.Rating,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan hotel: %w", err)
		}
		hotels = append(hotels, hotel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating hotels: %w", err)
	}

	// Calculate total pages
	totalPages := (total + opts.PerPage - 1) / opts.PerPage
	if totalPages == 0 {
		totalPages = 1
	}

	return &SearchResult{
		Hotels:     hotels,
		Total:      total,
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: totalPages,
	}, nil
}

// Autocomplete searches for regions, cities, and hotels by query
func (r *repository) Autocomplete(ctx context.Context, query string, opts *AutocompleteOptions) (*AutocompleteResponse, error) {
	if opts == nil {
		opts = &AutocompleteOptions{Limit: 10}
	}

	limit := opts.Limit
	if limit > 20 {
		limit = 20
	}
	if limit < 1 {
		limit = 10
	}

	results := []AutocompleteResult{}

	// 1. Search for cities from destinations table - priority 1
	cityQuery := `
		SELECT name, country_code
		FROM destinations
		WHERE name ILIKE $1
		ORDER BY
			CASE WHEN name ILIKE $2 THEN 0 ELSE 1 END,
			name
		LIMIT $3
	`

	cityRows, err := r.db.Pool.Query(ctx, cityQuery, "%"+query+"%", query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search cities: %w", err)
	}
	defer cityRows.Close()

	for cityRows.Next() {
		var cityName, countryCode string
		if err := cityRows.Scan(&cityName, &countryCode); err != nil {
			continue
		}

		results = append(results, AutocompleteResult{
			Type:        AutocompleteTypeCity,
			ID:          cityName,
			Name:        cityName,
			FullName:    fmt.Sprintf("%s, %s", cityName, countryCode),
			CountryCode: countryCode,
			CountryName: countryCode, // Will be looked up if needed
		})
	}

	// 2. Search for hotels by name - priority 2
	hotelQuery := `
		SELECT id, name, city, country_code
		FROM hotels
		WHERE name ILIKE $1
			AND deleted_at IS NULL
		ORDER BY overall_rating DESC NULLS LAST, name ASC
		LIMIT $2
	`

	hotelRows, err := r.db.Pool.Query(ctx, hotelQuery, "%"+query+"%", limit/2)
	if err != nil {
		return nil, fmt.Errorf("failed to search hotels: %w", err)
	}
	defer hotelRows.Close()

	for hotelRows.Next() {
		var id, name, city, countryCode string
		if err := hotelRows.Scan(&id, &name, &city, &countryCode); err != nil {
			continue
		}

		results = append(results, AutocompleteResult{
			Type:        AutocompleteTypeHotel,
			ID:          id,
			Name:        name,
			FullName:    fmt.Sprintf("%s, %s", name, city),
			City:        city,
			CountryCode: countryCode,
			CountryName: "Indonesia",
		})
	}

	if len(results) == 0 {
		return &AutocompleteResponse{
			Query:   query,
			Results: []AutocompleteResult{},
		}, nil
	}

	return &AutocompleteResponse{
		Query:   query,
		Results: results,
	}, nil
}

// GetPopularDestinations retrieves popular destinations (most searched cities)
func (r *repository) GetPopularDestinations(ctx context.Context, limit int) ([]AutocompleteResult, error) {
	if limit > 20 {
		limit = 20
	}
	if limit < 1 {
		limit = 10
	}

	query := `
		SELECT DISTINCT city, country_code
		FROM hotels
		WHERE deleted_at IS NULL
		ORDER BY city ASC
		LIMIT $1
	`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular destinations: %w", err)
	}
	defer rows.Close()

	results := []AutocompleteResult{}
	for rows.Next() {
		var city, countryCode string
		if err := rows.Scan(&city, &countryCode); err != nil {
			continue
		}

		results = append(results, AutocompleteResult{
			Type:        AutocompleteTypeCity,
			ID:          city,
			Name:        city,
			FullName:    fmt.Sprintf("%s, Indonesia", city),
			CountryCode: countryCode,
			CountryName: "Indonesia",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating destinations: %w", err)
	}

	if results == nil {
		results = []AutocompleteResult{}
	}

	return results, nil
}

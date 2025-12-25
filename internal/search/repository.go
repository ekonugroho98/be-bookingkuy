package search

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
)

// Repository defines interface for search data operations
type Repository interface {
	SearchHotels(ctx context.Context, req *SearchRequest, opts *SearchOptions) (*SearchResult, error)
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
		SELECT id, name, country, city, rating
		FROM hotels
		WHERE city = $1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM hotels
		WHERE city = $1
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
			sortClause = " ORDER BY rating DESC NULLS LAST, name ASC"
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
			&hotel.Country,
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

package search

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines interface for search business logic
type Service interface {
	SearchHotels(ctx context.Context, req *SearchRequest, opts *SearchOptions) (*SearchResult, error)
}

type service struct {
	repo Repository
}

// NewService creates a new search service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) SearchHotels(ctx context.Context, req *SearchRequest, opts *SearchOptions) (*SearchResult, error) {
	// Validate dates
	if req.CheckOut.Before(req.CheckIn) || req.CheckOut.Equal(req.CheckIn) {
		return nil, fmt.Errorf("check-out date must be after check-in date")
	}

	// Validate guests
	if req.Guests < 1 || req.Guests > 10 {
		return nil, fmt.Errorf("number of guests must be between 1 and 10")
	}

	// Set default pagination options
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}

	logger.Infof("Searching hotels in %s for %d guests", req.City, req.Guests)

	// Execute search
	result, err := s.repo.SearchHotels(ctx, req, opts)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to search hotels")
		return nil, fmt.Errorf("failed to search hotels: %w", err)
	}

	logger.Infof("Found %d hotels (page %d of %d)", result.Total, result.Page, result.TotalPages)
	return result, nil
}

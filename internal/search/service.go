package search

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines interface for search business logic
type Service interface {
	SearchHotels(ctx context.Context, req *SearchRequest, opts *SearchOptions) (*SearchResult, error)
	Autocomplete(ctx context.Context, query string, limit int) (*AutocompleteResponse, error)
	GetPopularDestinations(ctx context.Context, limit int) ([]AutocompleteResult, error)
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

// Autocomplete provides search suggestions for cities and hotels
func (s *service) Autocomplete(ctx context.Context, query string, limit int) (*AutocompleteResponse, error) {
	// Validate query length
	if len(query) < 2 {
		return &AutocompleteResponse{
			Query:   query,
			Results: []AutocompleteResult{},
		}, nil
	}

	logger.Infof("Autocomplete search for: %s", query)

	opts := &AutocompleteOptions{Limit: limit}
	result, err := s.repo.Autocomplete(ctx, query, opts)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get autocomplete results")
		return nil, fmt.Errorf("failed to get autocomplete: %w", err)
	}

	return result, nil
}

// GetPopularDestinations returns popular destination cities
func (s *service) GetPopularDestinations(ctx context.Context, limit int) ([]AutocompleteResult, error) {
	logger.Infof("Fetching popular destinations (limit: %d)", limit)

	results, err := s.repo.GetPopularDestinations(ctx, limit)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to get popular destinations")
		return nil, fmt.Errorf("failed to get popular destinations: %w", err)
	}

	return results, nil
}

package search

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of search.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SearchHotels(ctx context.Context, req *SearchRequest, opts *SearchOptions) (*SearchResult, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SearchResult), args.Error(1)
}

// TestNewService tests creating a new search service
func TestNewService(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	require.NotNil(t, service)
}

// TestService_SearchHotels_Success tests successful hotel search
func TestService_SearchHotels_Success(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}

	opts := &SearchOptions{
		Page:    1,
		PerPage: 10,
		SortBy:  SortByRating,
	}

	expectedResult := &SearchResult{
		Hotels: []Hotel{
			{
				ID:      "hotel-1",
				Name:    "Grand Hotel Bali",
				Country: "Indonesia",
				City:    "Bali",
				Rating:  pointerToFloat64(4.5),
			},
			{
				ID:      "hotel-2",
				Name:    "Bali Beach Resort",
				Country: "Indonesia",
				City:    "Bali",
				Rating:  pointerToFloat64(4.0),
			},
		},
		Total:      25,
		Page:       1,
		PerPage:    10,
		TotalPages: 3,
	}

	// Setup expectations
	mockRepo.On("SearchHotels", ctx, mock.MatchedBy(func(r *SearchRequest) bool {
		return r.City == "Bali" && r.Guests == 2
	}), mock.MatchedBy(func(o *SearchOptions) bool {
		return o.Page == 1 && o.PerPage == 10
	})).Return(expectedResult, nil)

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Hotels, 2)
	assert.Equal(t, 25, result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.Equal(t, 3, result.TotalPages)
	assert.Equal(t, "Grand Hotel Bali", result.Hotels[0].Name)

	mockRepo.AssertExpectations(t)
}

// TestService_SearchHotels_InvalidDates_CheckOutBeforeCheckIn tests error when check-out is before check-in
func TestService_SearchHotels_InvalidDates_CheckOutBeforeCheckIn(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(48 * time.Hour)
	checkOut := time.Now().Add(24 * time.Hour) // Before check-in

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}

	opts := &SearchOptions{}

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "check-out date must be after check-in date")
}

// TestService_SearchHotels_InvalidDates_EqualDates tests error when check-out equals check-in
func TestService_SearchHotels_InvalidDates_EqualDates(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := checkIn // Same as check-in

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}

	opts := &SearchOptions{}

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "check-out date must be after check-in date")
}

// TestService_SearchHotels_InvalidGuests_TooFew tests error when guests < 1
func TestService_SearchHotels_InvalidGuests_TooFew(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   0, // Invalid
	}

	opts := &SearchOptions{}

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "number of guests must be between 1 and 10")
}

// TestService_SearchHotels_InvalidGuests_TooMany tests error when guests > 10
func TestService_SearchHotels_InvalidGuests_TooMany(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   11, // Invalid
	}

	opts := &SearchOptions{}

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "number of guests must be between 1 and 10")
}

// TestService_SearchHotels_DefaultPagination tests default pagination values
func TestService_SearchHotels_DefaultPagination(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}

	opts := &SearchOptions{
		Page:    0, // Should default to 1
		PerPage: 0, // Should default to 20
	}

	expectedResult := &SearchResult{
		Hotels:     []Hotel{},
		Total:      0,
		Page:       1,
		PerPage:    20,
		TotalPages: 1,
	}

	// Setup expectations - should receive defaults
	mockRepo.On("SearchHotels", ctx, mock.AnythingOfType("*search.SearchRequest"), mock.MatchedBy(func(o *SearchOptions) bool {
		return o.Page == 1 && o.PerPage == 20
	})).Return(expectedResult, nil)

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PerPage)

	mockRepo.AssertExpectations(t)
}

// TestService_SearchHotels_WithPriceFilters tests search with price filters
func TestService_SearchHotels_WithPriceFilters(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	minPrice := 500000
	maxPrice := 2000000

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	}

	opts := &SearchOptions{
		Page:    1,
		PerPage: 10,
	}

	expectedResult := &SearchResult{
		Hotels: []Hotel{
			{
				ID:      "hotel-1",
				Name:    "Bali Resort",
				Country: "Indonesia",
				City:    "Bali",
			},
		},
		Total:      1,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}

	// Setup expectations
	mockRepo.On("SearchHotels", ctx, mock.MatchedBy(func(r *SearchRequest) bool {
		return r.MinPrice != nil && *r.MinPrice == 500000 &&
			   r.MaxPrice != nil && *r.MaxPrice == 2000000
	}), mock.AnythingOfType("*search.SearchOptions")).Return(expectedResult, nil)

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Hotels, 1)

	mockRepo.AssertExpectations(t)
}

// TestService_SearchHotels_RepositoryError tests error when repository fails
func TestService_SearchHotels_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}

	opts := &SearchOptions{
		Page:    1,
		PerPage: 10,
	}

	// Setup expectations
	mockRepo.On("SearchHotels", ctx, mock.AnythingOfType("*search.SearchRequest"), mock.AnythingOfType("*search.SearchOptions")).Return(nil, errors.New("database connection error"))

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to search hotels")

	mockRepo.AssertExpectations(t)
}

// TestService_SearchHotels_EmptyResults tests search with no results
func TestService_SearchHotels_EmptyResults(t *testing.T) {
	mockRepo := new(MockRepository)

	service := NewService(mockRepo)

	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "UnknownCity",
		Guests:   2,
	}

	opts := &SearchOptions{
		Page:    1,
		PerPage: 10,
	}

	expectedResult := &SearchResult{
		Hotels:     []Hotel{},
		Total:      0,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}

	// Setup expectations
	mockRepo.On("SearchHotels", ctx, mock.AnythingOfType("*search.SearchRequest"), mock.AnythingOfType("*search.SearchOptions")).Return(expectedResult, nil)

	// Execute
	result, err := service.SearchHotels(ctx, req, opts)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Hotels)
	assert.Equal(t, 0, result.Total)
	assert.Equal(t, 1, result.TotalPages)

	mockRepo.AssertExpectations(t)
}

// TestService_SearchHotels_DifferentSortOptions tests different sort options
func TestService_SearchHotels_DifferentSortOptions(t *testing.T) {
	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	testCases := []struct {
		name      string
		sortBy    SortBy
	}{
		{name: "Sort by price", sortBy: SortByPrice},
		{name: "Sort by rating", sortBy: SortByRating},
		{name: "Sort by name", sortBy: SortByName},
		{name: "No sort specified", sortBy: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			req := &SearchRequest{
				CheckIn:  checkIn,
				CheckOut: checkOut,
				City:     "Bali",
				Guests:   2,
			}

			opts := &SearchOptions{
				Page:    1,
				PerPage: 10,
				SortBy:  tc.sortBy,
			}

			expectedResult := &SearchResult{
				Hotels:     []Hotel{},
				Total:      0,
				Page:       1,
				PerPage:    10,
				TotalPages: 1,
			}

			// Setup expectations
			mockRepo.On("SearchHotels", ctx, mock.AnythingOfType("*search.SearchRequest"), mock.MatchedBy(func(o *SearchOptions) bool {
				return o.SortBy == tc.sortBy
			})).Return(expectedResult, nil)

			// Execute
			result, err := service.SearchHotels(ctx, req, opts)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, result)

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestService_SearchHotels_MultiplePages tests pagination across multiple pages
func TestService_SearchHotels_MultiplePages(t *testing.T) {
	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	pages := []int{1, 2, 3}

	for _, page := range pages {
		t.Run("Page "+string(rune(page+'0')), func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			req := &SearchRequest{
				CheckIn:  checkIn,
				CheckOut: checkOut,
				City:     "Bali",
				Guests:   2,
			}

			opts := &SearchOptions{
				Page:    page,
				PerPage: 10,
			}

			expectedResult := &SearchResult{
				Hotels:     []Hotel{},
				Total:      25,
				Page:       page,
				PerPage:    10,
				TotalPages: 3,
			}

			// Setup expectations
			mockRepo.On("SearchHotels", ctx, mock.AnythingOfType("*search.SearchRequest"), mock.MatchedBy(func(o *SearchOptions) bool {
				return o.Page == page && o.PerPage == 10
			})).Return(expectedResult, nil)

			// Execute
			result, err := service.SearchHotels(ctx, req, opts)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, page, result.Page)
			assert.Equal(t, 3, result.TotalPages)

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestService_SearchHotels_ValidGuestRange tests all valid guest counts
func TestService_SearchHotels_ValidGuestRange(t *testing.T) {
	ctx := context.Background()
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	for guests := 1; guests <= 10; guests++ {
		t.Run("Guests "+string(rune(guests+'0')), func(t *testing.T) {
			mockRepo := new(MockRepository)
			service := NewService(mockRepo)

			req := &SearchRequest{
				CheckIn:  checkIn,
				CheckOut: checkOut,
				City:     "Bali",
				Guests:   guests,
			}

			opts := &SearchOptions{
				Page:    1,
				PerPage: 10,
			}

			expectedResult := &SearchResult{
				Hotels:     []Hotel{},
				Total:      0,
				Page:       1,
				PerPage:    10,
				TotalPages: 1,
			}

			// Setup expectations
			mockRepo.On("SearchHotels", ctx, mock.MatchedBy(func(r *SearchRequest) bool {
				return r.Guests == guests
			}), mock.AnythingOfType("*search.SearchOptions")).Return(expectedResult, nil)

			// Execute
			result, err := service.SearchHotels(ctx, req, opts)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, result)

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestSortBy_Constants tests sort by constants
func TestSortBy_Constants(t *testing.T) {
	assert.Equal(t, SortBy("price"), SortByPrice)
	assert.Equal(t, SortBy("rating"), SortByRating)
	assert.Equal(t, SortBy("name"), SortByName)
}

// TestSearchRequest_Structure tests search request structure
func TestSearchRequest_Structure(t *testing.T) {
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)
	minPrice := 500000
	maxPrice := 2000000

	req := &SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	}

	assert.Equal(t, "Bali", req.City)
	assert.Equal(t, 2, req.Guests)
	assert.NotNil(t, req.MinPrice)
	assert.NotNil(t, req.MaxPrice)
	assert.Equal(t, 500000, *req.MinPrice)
	assert.Equal(t, 2000000, *req.MaxPrice)
}

// TestSearchOptions_Structure tests search options structure
func TestSearchOptions_Structure(t *testing.T) {
	opts := &SearchOptions{
		Page:    2,
		PerPage: 20,
		SortBy:  SortByRating,
	}

	assert.Equal(t, 2, opts.Page)
	assert.Equal(t, 20, opts.PerPage)
	assert.Equal(t, SortByRating, opts.SortBy)
}

// TestSearchResult_Structure tests search result structure
func TestSearchResult_Structure(t *testing.T) {
	result := &SearchResult{
		Hotels: []Hotel{
			{
				ID:      "hotel-1",
				Name:    "Test Hotel",
				Country: "Indonesia",
				City:    "Bali",
			},
		},
		Total:      1,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}

	assert.Len(t, result.Hotels, 1)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.Equal(t, 1, result.TotalPages)
}

// TestHotel_Structure tests hotel structure
func TestHotel_Structure(t *testing.T) {
	rating := 4.5
	minPrice := 500000
	maxPrice := 2000000

	hotel := &Hotel{
		ID:          "hotel-1",
		Name:        "Grand Hotel",
		Country:     "Indonesia",
		City:        "Bali",
		Rating:      &rating,
		MinPrice:    &minPrice,
		MaxPrice:    &maxPrice,
		Description: "Luxury hotel in Bali",
	}

	assert.Equal(t, "hotel-1", hotel.ID)
	assert.Equal(t, "Grand Hotel", hotel.Name)
	assert.Equal(t, "Indonesia", hotel.Country)
	assert.Equal(t, "Bali", hotel.City)
	assert.Equal(t, 4.5, *hotel.Rating)
	assert.Equal(t, 500000, *hotel.MinPrice)
	assert.Equal(t, 2000000, *hotel.MaxPrice)
	assert.Equal(t, "Luxury hotel in Bali", hotel.Description)
}

// Helper function to create pointer to float64
func pointerToFloat64(f float64) *float64 {
	return &f
}

package search

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockContextType is the actual context type used in handlers
type mockContextType struct {
	context.Context
}

// TestSearchHandler_SearchHotels_Success tests successful hotel search via HTTP
func TestSearchHandler_SearchHotels_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create request body
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	reqBody := SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: Successful search
	expectedResult := &SearchResult{
		Hotels: []Hotel{
			{
				ID:      "hotel-1",
				Name:    "Grand Hotel Bali",
				Country: "Indonesia",
				City:    "Bali",
			},
		},
		Total:      1,
		Page:       1,
		PerPage:    10,
		TotalPages: 1,
	}
	mockRepo.On("SearchHotels", mock.Anything, mock.AnythingOfType("*search.SearchRequest"), mock.AnythingOfType("*search.SearchOptions")).Return(expectedResult, nil)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody SearchResult
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Len(t, respBody.Hotels, 1)
	assert.Equal(t, "Grand Hotel Bali", respBody.Hotels[0].Name)

	mockRepo.AssertExpectations(t)
}

// TestSearchHandler_SearchHotels_InvalidDates_CheckOutBeforeCheckIn tests search with invalid dates
func TestSearchHandler_SearchHotels_InvalidDates_CheckOutBeforeCheckIn(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create request body with invalid dates
	checkIn := time.Now().Add(48 * time.Hour)
	checkOut := time.Now().Add(24 * time.Hour) // Before check-in

	reqBody := SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "check-out date must be after check-in date")
}

// TestSearchHandler_SearchHotels_InvalidGuests_TooFew tests search with too few guests
func TestSearchHandler_SearchHotels_InvalidGuests_TooFew(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create request body with invalid guests
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	reqBody := SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   0, // Invalid
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "number of guests must be between 1 and 10")
}

// TestSearchHandler_SearchHotels_InvalidGuests_TooMany tests search with too many guests
func TestSearchHandler_SearchHotels_InvalidGuests_TooMany(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create request body with invalid guests
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	reqBody := SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   11, // Invalid
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "number of guests must be between 1 and 10")
}

// TestSearchHandler_SearchHotels_WithPagination tests search with pagination parameters
func TestSearchHandler_SearchHotels_WithPagination(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create request body with pagination
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	reqBody := SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: Successful search with pagination
	expectedResult := &SearchResult{
		Hotels:     []Hotel{},
		Total:      50,
		Page:       2,
		PerPage:    20,
		TotalPages: 3,
	}
	mockRepo.On("SearchHotels", mock.Anything, mock.AnythingOfType("*search.SearchRequest"), mock.MatchedBy(func(opts *SearchOptions) bool {
		return opts.Page == 2 && opts.PerPage == 20
	})).Return(expectedResult, nil)

	// Create HTTP request with query params
	req := httptest.NewRequest("POST", "/search?page=2&per_page=20", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody SearchResult
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Equal(t, 2, respBody.Page)
	assert.Equal(t, 20, respBody.PerPage)
	assert.Equal(t, 50, respBody.Total)
	assert.Equal(t, 3, respBody.TotalPages)

	mockRepo.AssertExpectations(t)
}

// TestSearchHandler_SearchHotels_WithPriceFilters tests search with price filters
func TestSearchHandler_SearchHotels_WithPriceFilters(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create request body with price filters
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)
	minPrice := 500000
	maxPrice := 2000000

	reqBody := SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "Bali",
		Guests:   2,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: Successful search
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
		PerPage:    20,
		TotalPages: 1,
	}
	mockRepo.On("SearchHotels", mock.Anything, mock.MatchedBy(func(req *SearchRequest) bool {
		return req.MinPrice != nil && *req.MinPrice == 500000 &&
			req.MaxPrice != nil && *req.MaxPrice == 2000000
	}), mock.AnythingOfType("*search.SearchOptions")).Return(expectedResult, nil)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody SearchResult
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Len(t, respBody.Hotels, 1)
	assert.Equal(t, "Bali Resort", respBody.Hotels[0].Name)

	mockRepo.AssertExpectations(t)
}

// TestSearchHandler_SearchHotels_EmptyResults tests search with no results
func TestSearchHandler_SearchHotels_EmptyResults(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create request body
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(48 * time.Hour)

	reqBody := SearchRequest{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		City:     "UnknownCity",
		Guests:   2,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Mock: No results
	expectedResult := &SearchResult{
		Hotels:     []Hotel{},
		Total:      0,
		Page:       1,
		PerPage:    20,
		TotalPages: 1,
	}
	mockRepo.On("SearchHotels", mock.Anything, mock.AnythingOfType("*search.SearchRequest"), mock.AnythingOfType("*search.SearchOptions")).Return(expectedResult, nil)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/search", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusOK, rr.Code)

	var respBody SearchResult
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Empty(t, respBody.Hotels)
	assert.Equal(t, 0, respBody.Total)

	mockRepo.AssertExpectations(t)
}

// TestSearchHandler_SearchHotels_InvalidJSON tests search with invalid JSON
func TestSearchHandler_SearchHotels_InvalidJSON(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	handler := NewHandler(service)

	// Create HTTP request with invalid JSON
	req := httptest.NewRequest("POST", "/search", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.SearchHotels(rr, req)

	// Assertions
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var respBody map[string]string
	err := json.NewDecoder(rr.Body).Decode(&respBody)
	require.NoError(t, err)

	assert.Contains(t, respBody["error"], "Invalid request body")
}

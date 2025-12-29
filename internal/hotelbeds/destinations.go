package hotelbeds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// DestinationName represents HotelBeds destination name with content and language
type DestinationName struct {
	Content string `json:"content"`
}

// Destination represents a HotelBeds destination
type Destination struct {
	Code        string           `json:"code"`
	Name        DestinationName  `json:"name"`
	CountryCode string           `json:"countryCode"`
	CountryName string           `json:"countryName,omitempty"` // May not be in Content API
	Type        string           `json:"type,omitempty"`        // May not be in Content API
	ParentCode  *string          `json:"parentCode,omitempty"`
	Latitude    *float64         `json:"latitude,omitempty"`
	Longitude   *float64         `json:"longitude,omitempty"`
	Group       string           `json:"group,omitempty"` // ISLAND, ZONE, etc.
	IsoCode     string           `json:"isoCode,omitempty"`
}

// DestinationsResponse represents HotelBeds destinations API response
type DestinationsResponse struct {
	From         *int          `json:"from,omitempty"`
	To           *int          `json:"to,omitempty"`
	Total        *int          `json:"total,omitempty"`
	Destinations []Destination `json:"destinations,omitempty"`
}

// GetDestinations fetches destinations from HotelBeds API
func (c *Client) GetDestinations(ctx context.Context, countryCode string, offset, limit int) ([]Destination, error) {
	// Build endpoint with query parameters
	endpoint := "/hotel-content-api/1.0/locations/destinations"
	queryParams := make(map[string]string)

	if countryCode != "" {
		queryParams["countryCode"] = countryCode
	}
	if offset > 0 {
		queryParams["from"] = strconv.Itoa(offset)
	}
	if limit > 0 {
		queryParams["to"] = strconv.Itoa(offset + limit - 1)
	}

	// Add query string to endpoint
	if len(queryParams) > 0 {
		endpoint += "?"
		first := true
		for k, v := range queryParams {
			if !first {
				endpoint += "&"
			}
			endpoint += k + "=" + v
			first = false
		}
	}

	logger.Infof("Fetching destinations from HotelBeds: country=%s, offset=%d, limit=%d",
		countryCode, offset, limit)

	// Make API request
	resp, err := c.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch destinations: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HotelBeds API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log raw response for debugging
	logger.Infof("Raw Hotelbeds response: %s", string(bodyBytes))

	// Parse response
	var response DestinationsResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode destinations response: %w\nRaw response: %s", err, string(bodyBytes))
	}

	logger.Infof("Successfully fetched %d destinations from HotelBeds", len(response.Destinations))

	return response.Destinations, nil
}

// GetAllDestinations fetches all destinations with pagination
func (c *Client) GetAllDestinations(ctx context.Context, countryCode string) ([]Destination, error) {
	var allDestinations []Destination
	offset := 0
	limit := 1000 // HotelBeds max pagination

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		destinations, err := c.GetDestinations(ctx, countryCode, offset, limit)
		if err != nil {
			return nil, err
		}

		if len(destinations) == 0 {
			break // No more data
		}

		allDestinations = append(allDestinations, destinations...)
		offset += limit

		logger.Infof("Fetched %d destinations (total: %d)", len(destinations), len(allDestinations))
	}

	return allDestinations, nil
}

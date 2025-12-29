package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/hotelbeds"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/db"
	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service handles data synchronization from HotelBeds
type Service struct {
	db        *db.DB
	hotelbeds *hotelbeds.Client
}

// NewService creates a new sync service
func NewService(database *db.DB, hbClient *hotelbeds.Client) *Service {
	return &Service{
		db:        database,
		hotelbeds: hbClient,
	}
}

// SyncDestinations syncs destinations from HotelBeds to local database
func (s *Service) SyncDestinations(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		Errors: []SyncError{},
	}

	// Set defaults
	if opts.BatchSize == 0 {
		opts.BatchSize = 100
	}

	logger.Infof("Starting destination sync: country=%s, limit=%d, dry_run=%v",
		opts.CountryCode, opts.Limit, opts.DryRun)

	// Fetch destinations from HotelBeds
	var destinations []hotelbeds.Destination
	var err error

	if opts.Limit > 0 {
		// Fetch with limit
		destinations, err = s.hotelbeds.GetDestinations(ctx, "", 0, opts.Limit)
	} else {
		// Fetch all destinations
		destinations, err = s.hotelbeds.GetAllDestinations(ctx, "")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch destinations from HotelBeds: %w", err)
	}

	// Filter by country code if specified (API doesn't support country filter)
	if opts.CountryCode != "" {
		filtered := make([]hotelbeds.Destination, 0)
		for _, dest := range destinations {
			if dest.CountryCode == opts.CountryCode {
				filtered = append(filtered, dest)
			}
		}
		destinations = filtered
	}

	logger.Infof("Fetched %d destinations from HotelBeds (filtered to country=%s)", len(destinations), opts.CountryCode)

	// Process each destination
	for _, dest := range destinations {
		select {
		case <-ctx.Done():
			logger.Warnf("Sync cancelled by context")
			return result, ctx.Err()
		default:
		}

		// Convert to our model
		destination := s.convertToDestination(dest)

		// Dry run - skip database operations
		if opts.DryRun {
			result.Skipped++
			result.Total++
			continue
		}

		// Upsert to database
		inserted, err := s.upsertDestination(ctx, &destination)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, SyncError{
				Code:        dest.Code,
				Message:     err.Error(),
				Record:      dest.Name.Content,
				Recoverable: true,
			})
			logger.Errorf("Failed to sync destination %s (%s): %v", dest.Code, dest.Name.Content, err)
			continue
		}

		if inserted {
			result.Inserted++
		} else {
			result.Updated++
		}

		result.Total++

		// Log progress every 100 records
		if result.Total%100 == 0 {
			logger.Infof("Sync progress: %d/%d destinations processed", result.Total, len(destinations))
		}
	}

	result.Duration = time.Since(startTime)

	logger.Infof("Destination sync completed: total=%d, inserted=%d, updated=%d, failed=%d, duration=%v",
		result.Total, result.Inserted, result.Updated, result.Failed, result.Duration)

	return result, nil
}

// upsertDestination inserts or updates a destination in the database
func (s *Service) upsertDestination(ctx context.Context, dest *Destination) (bool, error) {
	query := `
		INSERT INTO destinations (code, name, country_code, country_name, type, parent_code, latitude, longitude, hotelbeds_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			country_code = EXCLUDED.country_code,
			country_name = EXCLUDED.country_name,
			type = EXCLUDED.type,
			parent_code = EXCLUDED.parent_code,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			hotelbeds_data = EXCLUDED.hotelbeds_data,
			updated_at = NOW()
		RETURNING (xmax = 0) AS inserted
	`

	var inserted bool
	err := s.db.Pool.QueryRow(ctx, query,
		dest.Code,
		dest.Name,
		dest.CountryCode,
		dest.CountryName,
		dest.Type,
		dest.ParentCode,
		dest.Latitude,
		dest.Longitude,
		dest.HotelbedsData,
	).Scan(&inserted)

	if err != nil {
		return false, fmt.Errorf("failed to upsert destination: %w", err)
	}

	return inserted, nil
}

// convertToDestination converts HotelBeds destination to our model
func (s *Service) convertToDestination(hbDest hotelbeds.Destination) Destination {
	// Extract name content from DestinationName object
	name := hbDest.Name.Content

	// Use country code as country name if not provided
	countryName := hbDest.CountryName
	if countryName == "" {
		countryName = hbDest.CountryCode // Fallback to country code
	}

	// Default type to CITY if not provided
	destType := hbDest.Type
	if destType == "" {
		destType = "CITY"
	}

	// Marshal to JSON for storage
	hotelbedsData, _ := json.Marshal(hbDest)

	return Destination{
		Code:         hbDest.Code,
		Name:         name,
		CountryCode:  hbDest.CountryCode,
		CountryName:  countryName,
		Type:         destType,
		ParentCode:   hbDest.ParentCode,
		Latitude:     hbDest.Latitude,
		Longitude:    hbDest.Longitude,
		HotelbedsData: hotelbedsData,
		SyncedAt:     time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// SyncHotels syncs hotels from HotelBeds to local database
func (s *Service) SyncHotels(ctx context.Context, destinationCode int, opts SyncOptions) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		Errors: []SyncError{},
	}

	// Set defaults
	if opts.BatchSize == 0 {
		opts.BatchSize = 100
	}

	logger.Infof("Starting hotel sync: destination=%d, limit=%d, dry_run=%v",
		destinationCode, opts.Limit, opts.DryRun)

	// Fetch hotels from HotelBeds
	var hotels []hotelbeds.ContentHotelContent
	var err error

	destCodeStr := fmt.Sprintf("%d", destinationCode)

	if opts.Limit > 0 {
		// Fetch with limit
		hotels, err = s.hotelbeds.GetHotels(ctx, destCodeStr, 0, opts.Limit)
	} else {
		// Fetch all hotels
		hotels, err = s.hotelbeds.GetAllHotels(ctx, destCodeStr)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch hotels from HotelBeds: %w", err)
	}

	logger.Infof("Fetched %d hotels from HotelBeds (destination=%s)", len(hotels), destCodeStr)

	// Process each hotel
	for _, hotel := range hotels {
		select {
		case <-ctx.Done():
			logger.Warnf("Sync cancelled by context")
			return result, ctx.Err()
		default:
		}

		// Convert to our model
		h := s.convertToHotel(hotel)

		// Dry run - skip database operations
		if opts.DryRun {
			result.Skipped++
			result.Total++
			continue
		}

		// Upsert to database
		inserted, err := s.upsertHotel(ctx, &h)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, SyncError{
				Code:        hotel.Code,
				Message:     err.Error(),
				Record:      hotel.Name,
				Recoverable: true,
			})
			logger.Errorf("Failed to sync hotel %s (%s): %v", hotel.Code, hotel.Name, err)
			continue
		}

		if inserted {
			result.Inserted++
		} else {
			result.Updated++
		}

		result.Total++

		// Log progress every 50 records (hotels sync is slower)
		if result.Total%50 == 0 {
			logger.Infof("Sync progress: %d/%d hotels processed", result.Total, len(hotels))
		}
	}

	result.Duration = time.Since(startTime)

	logger.Infof("Hotel sync completed: total=%d, inserted=%d, updated=%d, failed=%d, duration=%v",
		result.Total, result.Inserted, result.Updated, result.Failed, result.Duration)

	return result, nil
}

// upsertHotel inserts or updates a hotel in the database
func (s *Service) upsertHotel(ctx context.Context, hotel *Hotel) (bool, error) {
	query := `
		INSERT INTO hotels (
			provider_code, provider_hotel_id, name, description, star_rating,
			country_code, city, address, postal_code, latitude, longitude,
			images, amenities
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (provider_code, provider_hotel_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			star_rating = EXCLUDED.star_rating,
			country_code = EXCLUDED.country_code,
			city = EXCLUDED.city,
			address = EXCLUDED.address,
			postal_code = EXCLUDED.postal_code,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			images = EXCLUDED.images,
			amenities = EXCLUDED.amenities,
			updated_at = NOW()
		RETURNING (xmax = 0) AS inserted
	`

	var inserted bool
	err := s.db.Pool.QueryRow(ctx, query,
		hotel.ProviderCode,
		hotel.ProviderHotelID,
		hotel.Name,
		hotel.Description,
		hotel.StarRating,
		hotel.CountryCode,
		hotel.City,
		hotel.Address,
		hotel.PostalCode,
		hotel.Latitude,
		hotel.Longitude,
		hotel.Images,
		hotel.Amenities,
	).Scan(&inserted)

	if err != nil {
		return false, fmt.Errorf("failed to upsert hotel: %w", err)
	}

	return inserted, nil
}

// convertToHotel converts HotelBeds hotel to our model
func (s *Service) convertToHotel(hbHotel hotelbeds.ContentHotelContent) Hotel {
	// Extract star rating from category code
	starRating := 0.0
	if len(hbHotel.Category) > 0 {
		categoryCode := hbHotel.Category[0].Code
		// Parse category code (e.g., "4EST" = 4 stars)
		if len(categoryCode) > 0 {
			rating := float64(categoryCode[0] - '0')
			if rating >= 1 && rating <= 5 {
				starRating = rating
			}
		}
	}

	// Extract description
	description := ""
	if len(hbHotel.Description.ContentText) > 0 {
		description = hbHotel.Description.ContentText
	}

	// Build images array for JSON storage
	images := make([]map[string]interface{}, 0)
	for _, img := range hbHotel.Images {
		images = append(images, map[string]interface{}{
			"type":          img.ImageType,
			"path":          img.Path,
			"visual_order":  img.VisualOrder,
			"url":           fmt.Sprintf("https://photos.hotelbeds.com/giata/%s", img.Path),
		})
	}
	imagesJSON, _ := json.Marshal(images)

	// Build facilities/amenities array for JSON storage
	amenities := make([]map[string]interface{}, 0)
	for _, fg := range hbHotel.Facilities {
		for _, f := range fg.Facilities {
			amenities = append(amenities, map[string]interface{}{
				"code":           f.FacilityCode,
				"number":         f.FacilityNumber,
				"group_code":     fg.FacilityGroupCode,
				"group_name":     fg.FacilityGroupName,
				"description":    f.Content.ContentText,
				"ind":            f.Ind,
			})
		}
	}
	amenitiesJSON, _ := json.Marshal(amenities)

	// Marshal full HotelBeds data for reference
	hotelbedsData, _ := json.Marshal(hbHotel)

	// Extract city from destination name (fallback to hotel city field)
	city := hbHotel.City
	if city == "" && hbHotel.Destination.Name != "" {
		city = hbHotel.Destination.Name
	}

	return Hotel{
		ProviderCode:    "hotelbeds",
		ProviderHotelID: hbHotel.Code,
		Name:            hbHotel.Name,
		Description:     description,
		StarRating:      starRating,
		CountryCode:     "", // Will need to be looked up or passed in
		City:            city,
		Address:         hbHotel.Address,
		PostalCode:      hbHotel.PostalCode,
		Latitude:        nil, // Not provided in Content API list endpoint
		Longitude:       nil, // Not provided in Content API list endpoint
		Images:          imagesJSON,
		Amenities:       amenitiesJSON,
		DestinationCode: hbHotel.Destination.Code,
		HotelbedsData:   hotelbedsData,
		SyncedAt:        time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

package hotelbeds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// AvailabilityRequest represents a request to check hotel room availability
type AvailabilityRequest struct {
	HotelCode string    `json:"hotelCode"`
	RoomCode  string    `json:"roomCode,omitempty"`
	CheckIn   time.Time `json:"checkIn"`
	CheckOut  time.Time `json:"checkOut"`
	Guests    int       `json:"guests"`
}

// AvailabilityResponse represents availability check response from HotelBeds
type AvailabilityResponse struct {
	HotelCode    string  `json:"hotelCode"`
	HotelName    string  `json:"hotelName"`
	IsAvailable  bool    `json:"available"`
	Rooms        []Room  `json:"rooms"`
	TotalPrice   int     `json:"totalPrice,omitempty"`
	Currency     string  `json:"currency,omitempty"`
}

// Room represents a room with availability info
type Room struct {
	RoomCode    string  `json:"roomCode"`
	RoomName    string  `json:"roomName"`
	Available   bool    `json:"available"`
	Price       int     `json:"price"`
	Currency    string  `json:"currency"`
	MaxGuests   int     `json:"maxGuests"`
	Beds        string  `json:"beds"`
}

// HotelDetailsRequest represents a request to get hotel information
type HotelDetailsRequest struct {
	HotelCode string  `json:"hotelCode"`
	Language  string  `json:"language,omitempty"` // Default: "ENG"
}

// HotelDetailsResponse represents detailed hotel information
type HotelDetailsResponse struct {
	HotelCode       string             `json:"hotelCode"`
	HotelName       string             `json:"hotelName"`
	Description     string             `json:"description"`
	CountryCode     string             `json:"countryCode"`
	CityCode        string             `json:"cityCode"`
	Address         string             `json:"address"`
	PostalCode      string             `json:"postalCode"`
	Rating          float64            `json:"categoryCode"` // HotelBeds uses categoryCode
	CategoryName    string             `json:"categoryName"`
	Images          []HotelImage      `json:"images"`
	Amenities       []Amenity          `json:"amenities"`
	Location        HotelLocation      `json:"location"`
	Policies        HotelPolicies      `json:"policies"`
}

// HotelImage represents a hotel image
type HotelImage struct {
	ImageType string `json:"imageType"` // exterior, interior, room, facilities, etc.
	URL       string `json:"url"`
	Order     int    `json:"visualOrder"`
}

// Amenity represents hotel amenity
type Amenity struct {
	Code        int     `json:"amenityCode"`
	Description string  `json:"description"`
	Number       int     `json:"number,omitempty"` // For things like pools, restaurants count
}

// HotelLocation represents hotel location data
type HotelLocation struct {
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	MapZoomLevel     int     `json:"mapZoomLevel,omitempty"`
	MapCoordinates   string  `json:"mapCoordinates,omitempty"`
}

// HotelPolicies represents hotel policies
type HotelPolicies struct {
	CheckInTime  string `json:"checkInTime"`
	CheckOutTime string `json:"checkOutTime"`
	Cancellation string `json:"cancellationDescription,omitempty"`
}

// RoomRateRequest represents a request to get room pricing
type RoomRateRequest struct {
	HotelCode string    `json:"hotelCode"`
	RoomCode  string    `json:"roomCode,omitempty"`
	CheckIn   time.Time `json:"checkIn"`
	CheckOut  time.Time `json:"checkOut"`
	Guests    int       `json:"guests"`
}

// RoomRateResponse represents room pricing information
type RoomRateResponse struct {
	HotelCode  string `json:"hotelCode"`
	RoomCode   string `json:"roomCode"`
	RoomName   string `json:"roomName"`
	Rates      []Rate `json:"rates"`
	TotalPrice int    `json:"totalPrice"`
	Currency   string `json:"currency"`
}

// Rate represents a rate for a specific date or package
type Rate struct {
	RateCode     string  `json:"rateCode"`
	RateName     string  `json:"rateName"`
	Price        int     `json:"price"`
	Currency     string  `json:"currency"`
	Description  string  `json:"description,omitempty"`
}

// BookingRequest represents a request to book a room with HotelBeds
type BookingRequest struct {
	HotelCode   string           `json:"hotelCode"`
	RoomCode    string           `json:"roomCode"`
	CheckIn     time.Time        `json:"checkIn"`
	CheckOut    time.Time        `json:"checkOut"`
	Guests      int              `json:"guests"`
	Holder      HolderInfo       `json:"holder"`
	Payment     PaymentInfo      `json:"payment"`
}

// HolderInfo represents guest/booking holder information
type HolderInfo struct {
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

// PaymentInfo represents payment information
type PaymentInfo struct {
	PaymentMethodType string `json:"paymentMethodType"` // "CREDITCARD", "ATM", "BANKTRANSFER", etc.
	CardNumber        string `json:"cardNumber,omitempty"`
	CardHolderName    string `json:"cardHolderName,omitempty"`
	CardExpiry        string `json:"cardExpiry,omitempty"`
	CardCVC           string `json:"cardCVC,omitempty"`
}

// BookingResponse represents HotelBeds booking confirmation
type BookingResponse struct {
	BookingReference  string    `json:"reference"`
	HotelCode         string    `json:"hotelCode"`
	RoomCode          string    `json:"roomCode"`
	CheckIn           time.Time `json:"checkIn"`
	CheckOut          time.Time `json:"checkOut"`
	Status            string    `json:"status"`
	Cancellation      CancellationInfo `json:"cancellation"`
}

// CancellationInfo represents cancellation information
type CancellationInfo struct {
	Cancellable      bool   `json:"cancellable"`
	Deadline         string `json:"deadline,omitempty"`
	Description      string `json:"description,omitempty"`
	Amount           int    `json:"amount,omitempty"`
	Currency          string `json:"currency,omitempty"`
}

// GetHotelAvailability checks room availability for specific hotel and dates
func (c *Client) GetHotelAvailability(ctx context.Context, req *AvailabilityRequest) (*AvailabilityResponse, error) {
	logger.Infof("Checking HotelBeds availability: hotel=%s, checkIn=%s, checkOut=%s, guests=%d",
		req.HotelCode, req.CheckIn.Format("2006-01-02"), req.CheckOut.Format("2006-01-02"), req.Guests)

	// Build HotelBeds API request body
	apiReq := map[string]interface{}{
		"stay": map[string]interface{}{
			"checkIn":  req.CheckIn.Format("2006-01-02"),
			"checkOut": req.CheckOut.Format("2006-01-02"),
			"shift":   "STANDARD", // or "FIRST_NIGHT" for check-in/out flexibility
		},
		"occupancies": []map[string]interface{}{
			{
				"rooms": 1,
				"adults": req.Guests,
				"children": 0,
				"paxes": []map[string]interface{}{}, // Empty for now
			},
		},
		"availability": true, // Only show available hotels
	}

	// Make API call
	endpoint := fmt.Sprintf("/hotel-api/1.0/hotels/%s/availability", req.HotelCode)
	resp, err := c.Post(ctx, endpoint, apiReq)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to check HotelBeds availability")
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log raw response for debugging
	logger.Debugf("HotelBeds availability response: %s", string(body))

	var apiResp struct {
		HotelCode  string `json:"hotelCode"`
		HotelName  string `json:"hotelName"`
		Rooms      []struct {
			RoomCode   string  `json:"roomCode"`
			RoomName   string  `json:"roomName"`
			Available  bool    `json:"available"`
			Rates      []struct {
				RateCode   string `json:"rateCode"`
				RateName   string `json:"rateName"`
				NetPrice   int    `json:"net"`
				GrossPrice int    `json:"gross"`
				Currency   string `json:"currency"`
			} `json:"rates"`
		} `json:"rooms"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse availability response: %w", err)
	}

	// Transform to our response format
	response := &AvailabilityResponse{
		HotelCode:   apiResp.HotelCode,
		HotelName:   apiResp.HotelName,
		IsAvailable: len(apiResp.Rooms) > 0 && apiResp.Rooms[0].Available,
	}

	// Convert rooms
	for _, apiRoom := range apiResp.Rooms {
		if apiRoom.Available && len(apiRoom.Rates) > 0 {
			// Use the first available rate
			rate := apiRoom.Rates[0]
			totalPrice := rate.GrossPrice // Use gross price (what customer pays)

			response.Rooms = append(response.Rooms, Room{
				RoomCode:  apiRoom.RoomCode,
				RoomName:  apiRoom.RoomName,
				Available: apiRoom.Available,
				Price:     totalPrice,
				Currency:   rate.Currency,
				MaxGuests: req.Guests, // Default to requested guests
				Beds:       "1 King Bed", // Could be parsed from room details
			})

			// Set total price from first room
			response.TotalPrice = totalPrice
			response.Currency = rate.Currency
		}
	}

	logger.Infof("HotelBeds availability check complete: available=%t, rooms=%d, price=%d",
		response.IsAvailable, len(response.Rooms), response.TotalPrice)

	return response, nil
}

// GetHotelDetails fetches complete hotel information from HotelBeds
func (c *Client) GetHotelDetails(ctx context.Context, hotelCode string) (*HotelDetailsResponse, error) {
	logger.Infof("Fetching HotelBeds hotel details: %s", hotelCode)

	// Determine language (default to English)
	language := "ENG"

	endpoint := fmt.Sprintf("/hotel-api/1.0/hotels/%s?language=%s", hotelCode, language)
	resp, err := c.Get(ctx, endpoint)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to fetch HotelBeds hotel details")
		return nil, fmt.Errorf("failed to get hotel details: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debugf("HotelBeds hotel details response: %s", string(body))

	var apiResp struct {
		HotelCode       string             `json:"code"`
		HotelName       string             `json:"name"`
		Description     string             `json:"description,omitempty"`
		CountryCode     string             `json:"countryCode"`
		CityCode        string             `json:"cityCode"`
	Content         HotelContent       `json:"content,omitempty"`
	Images          []HotelImage      `json:"images,omitempty"`
		Amenities       []Amenity          `json:"amenities,omitempty"`
		InterestPoints  []HotelInterestPoint `json:"interestPoints,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse hotel details response: %w", err)
	}

	response := &HotelDetailsResponse{
		HotelCode:    apiResp.HotelCode,
		HotelName:    apiResp.HotelName,
		Description:  apiResp.Description,
		CountryCode:  apiResp.CountryCode,
		CityCode:     apiResp.CityCode,
		Images:       apiResp.Images,
		Amenities:    apiResp.Amenities,
	}

	// Extract location if available
	if apiResp.Content.Location != nil {
		response.Location.Latitude = apiResp.Content.Location.Latitude
		response.Location.Longitude = apiResp.Content.Location.Longitude
	}

	logger.Infof("HotelBeds hotel details fetched successfully: %s", hotelCode)

	return response, nil
}

// GetRoomRates fetches pricing for specific room and dates
func (c *Client) GetRoomRates(ctx context.Context, req *RoomRateRequest) (*RoomRateResponse, error) {
	logger.Infof("Fetching HotelBeds room rates: hotel=%s, room=%s, checkIn=%s, checkOut=%s",
		req.HotelCode, req.RoomCode, req.CheckIn.Format("2006-01-02"), req.CheckOut.Format("2006-01-02"))

	// Build request
	apiReq := map[string]interface{}{
		"stay": map[string]interface{}{
			"checkIn":  req.CheckIn.Format("2006-01-02"),
			"checkOut": req.CheckOut.Format("2006-01-02"),
		},
		"occupancies": []map[string]interface{}{
			{
				"rooms": 1,
				"adults": req.Guests,
				"children": 0,
			},
		},
	}

	// If room code specified, get specific room rates
	endpoint := fmt.Sprintf("/hotel-api/1.0/hotels/%s/availability", req.HotelCode)
	if req.RoomCode != "" {
		// Note: HotelBeds might have different endpoint for specific room
		// For now, we'll get availability and filter by room code
	}

	resp, err := c.Post(ctx, endpoint, apiReq)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to fetch HotelBeds room rates")
		return nil, fmt.Errorf("failed to get room rates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debugf("HotelBeds room rates response: %s", string(body))

	var apiResp struct {
		HotelCode  string `json:"hotelCode"`
		HotelName  string `json:"hotelName"`
		Rooms      []struct {
			RoomCode  string  `json:"roomCode"`
			RoomName  string  `json:"roomName"`
			Rates     []struct {
				RateCode    string  `json:"rateCode"`
				RateName    string  `json:"rateName"`
				NetPrice    int     `json:"net"`
				GrossPrice  int     `json:"gross"`
				Currency    string  `json:"currency"`
			} `json:"rates"`
		} `json:"rooms"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse room rates response: %w", err)
	}

	// Find requested room
	var targetRoom *struct {
		RoomCode  string  `json:"roomCode"`
		RoomName  string  `json:"roomName"`
		Rates     []struct {
			RateCode    string  `json:"rateCode"`
			RateName    string  `json:"rateName"`
			NetPrice    int     `json:"net"`
			GrossPrice  int     `json:"gross"`
			Currency    string  `json:"currency"`
		} `json:"rates"`
	}

	if req.RoomCode != "" {
		// Find specific room
		for i := range apiResp.Rooms {
			if apiResp.Rooms[i].RoomCode == req.RoomCode {
				targetRoom = &apiResp.Rooms[i]
				break
			}
		}
	} else {
		// Use first available room
		if len(apiResp.Rooms) > 0 {
			targetRoom = &apiResp.Rooms[0]
		}
	}

	if targetRoom == nil || len(targetRoom.Rates) == 0 {
		return nil, fmt.Errorf("no available rooms found")
	}

	// Calculate total price (sum of all rates for the stay)
	// HotelBeds returns rates per room per night, so we need to calculate total
	rate := targetRoom.Rates[0]
	nights := int(req.CheckOut.Sub(req.CheckIn).Hours() / 24)
	totalPrice := rate.GrossPrice * nights

	response := &RoomRateResponse{
		HotelCode:  apiResp.HotelCode,
		RoomCode:   targetRoom.RoomCode,
		RoomName:   targetRoom.RoomName,
		Rates: []Rate{
			{
				RateCode:    rate.RateCode,
				RateName:    rate.RateName,
				Price:       rate.GrossPrice,
				Currency:    rate.Currency,
				Description: fmt.Sprintf("Rate per night (%d nights)", nights),
			},
		},
		TotalPrice: totalPrice,
		Currency:   rate.Currency,
	}

	logger.Infof("HotelBeds room rates fetched: price=%d %s for %d nights",
		totalPrice, response.Currency, nights)

	return response, nil
}

// CreateBooking books a room with HotelBeds
func (c *Client) CreateBooking(ctx context.Context, req *BookingRequest) (*BookingResponse, error) {
	logger.Infof("Creating HotelBeds booking: hotel=%s, room=%s, checkIn=%s, checkOut=%s",
		req.HotelCode, req.RoomCode, req.CheckIn.Format("2006-01-02"), req.CheckOut.Format("2006-01-02"))

	// Build HotelBeds booking request
	apiReq := map[string]interface{}{
		"stay": map[string]interface{}{
			"checkIn":  req.CheckIn.Format("2006-01-02"),
			"checkOut": req.CheckOut.Format("2006-01-02"),
			"shift":    "STANDARD",
		},
		"occupancies": []map[string]interface{}{
			{
				"rooms":   1,
				"adults":   req.Guests,
				"children": 0,
				"paxes":    []map[string]interface{}{},
			},
		},
		"holder": map[string]interface{}{
			"name":      req.Holder.Name,
			"surname":   req.Holder.Surname,
			"email":     req.Holder.Email,
		},
		"rooms": []map[string]interface{}{
			{
				"roomCode": req.RoomCode,
				"rateCode": "STANDARD", // Could be made configurable
			},
		},
		"payment": map[string]interface{}{
			"paymentMethodType": req.Payment.PaymentMethodType,
			// Add card details if credit card
		},
		"remark": "Booking from Bookingkuy",
	}

	endpoint := "/hotel-api/1.0/bookings"
	resp, err := c.Post(ctx, endpoint, apiReq)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to create HotelBeds booking")
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debugf("HotelBeds booking response: %s", string(body))

	var apiResp struct {
		BookingReference string    `json:"reference"`
		HotelCode        string    `json:"hotelCode"`
		RoomCode         string    `json:"roomCode"`
		CheckIn          string    `json:"checkIn"`
		CheckOut         string    `json:"checkOut"`
		Status           string    `json:"status"`
		Cancellation     CancellationInfo `json:"cancellation"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse booking response: %w", err)
	}

	response := &BookingResponse{
		BookingReference: apiResp.BookingReference,
		HotelCode:        apiResp.HotelCode,
		RoomCode:         apiResp.RoomCode,
		Status:           apiResp.Status,
		Cancellation:     apiResp.Cancellation,
	}

	// Parse dates
	if apiResp.CheckIn != "" {
		response.CheckIn, _ = time.Parse("2006-01-02", apiResp.CheckIn)
	}
	if apiResp.CheckOut != "" {
		response.CheckOut, _ = time.Parse("2006-01-02", apiResp.CheckOut)
	}

	logger.Infof("HotelBeds booking created successfully: reference=%s", response.BookingReference)

	return response, nil
}

// CancelBooking cancels a booking with HotelBeds
func (c *Client) CancelBooking(ctx context.Context, bookingReference string) error {
	logger.Infof("Cancelling HotelBeds booking: %s", bookingReference)

	endpoint := fmt.Sprintf("/hotel-api/1.0/bookings/%s/cancellation", bookingReference)
	apiReq := map[string]interface{}{
		"cancellation": map[string]interface{}{
			"reason": "Customer cancellation",
		},
	}

	resp, err := c.Put(ctx, endpoint, apiReq)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to cancel HotelBeds booking")
		return fmt.Errorf("failed to cancel booking: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		logger.Infof("HotelBeds booking cancelled successfully: %s", bookingReference)
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("failed to cancel booking, status: %d, body: %s", resp.StatusCode, string(body))
}

// HotelContent represents additional hotel content information
type HotelContent struct {
	Description string         `json:"description,omitempty"`
	Location    *HotelLocation  `json:"location,omitempty"`
}

// HotelInterestPoint represents points of interest near the hotel
type HotelInterestPoint struct {
	Name        string  `json:"name"`
	Distance    int     `json:"distance"`
	DistanceUnit string  `json:"distanceUnit"`
}

// =====================================================
// HotelBeds Content API Types
// =====================================================

// ContentHotelContent represents hotel from HotelBeds Content API
type ContentHotelContent struct {
	Code         string                    `json:"code"`
	Name         string                    `json:"name"`
	Destination  ContentDestinationCode    `json:"destination"`
	Category     []ContentCategoryCode     `json:"category"`
	Address      string                    `json:"address,omitempty"`
	PostalCode   string                    `json:"postalCode,omitempty"`
	City         string                    `json:"city,omitempty"`
	Email        string                    `json:"email,omitempty"`
	Web          string                    `json:"web,omitempty"`
	PhoneNumber string                    `json:"phoneNumber,omitempty"`
	Fax          string                    `json:"fax,omitempty"`
	S2C          string                    `json:"s2c,omitempty"`
	Images       []ContentHotelImage       `json:"images,omitempty"`
	Terms        ContentTerms              `json:"terms,omitempty"`
	Description  ContentDescription        `json:"description,omitempty"`
	RoomDetails  []ContentRoomDetail       `json:"rooms,omitempty"`
	Facilities   []ContentFacilityGroup    `json:"facilities,omitempty"`
	InterestPoints []ContentInterestPoint `json:"interestPoints,omitempty"`
	Issues       []ContentIssue            `json:"issues,omitempty"`
	Ancillaries  ContentAncillaries        `json:"ancillaries,omitempty"`
}

// ContentDestinationCode represents destination reference
type ContentDestinationCode struct {
	Code int    `json:"code"`
	Name string `json:"name,omitempty"`
}

// ContentCategoryCode represents hotel category (stars)
type ContentCategoryCode struct {
	Code        string `json:"code"`
	SimpleCode  string `json:"simpleCode,omitempty"`
	Description string `json:"description,omitempty"`
}

// ContentHotelImage represents hotel image from Content API
type ContentHotelImage struct {
	ImageType   string `json:"imageType"`
	Path        string `json:"path"`
	VisualOrder int    `json:"visualOrder,omitempty"`
}

// ContentDescription represents hotel description
type ContentDescription struct {
	ContentText  string `json:"contentText,omitempty"`
	LanguageCode string `json:"languageCode,omitempty"`
}

// ContentRoomDetail represents room type details
type ContentRoomDetail struct {
	RoomCode        string   `json:"roomCode"`
	RoomType        string   `json:"roomType,omitempty"`
	Characteristics []string `json:"characteristics,omitempty"`
	MinPax          int      `json:"minPax,omitempty"`
	MaxPax          int      `json:"maxPax,omitempty"`
	MaxAdults       int      `json:"maxAdults,omitempty"`
	MaxChildren     int      `json:"maxChildren,omitempty"`
	MinAdults       int      `json:"minAdults,omitempty"`
}

// ContentFacilityGroup represents hotel facilities grouped by type
type ContentFacilityGroup struct {
	FacilityGroupCode  int               `json:"facilityGroupCode"`
	FacilityGroupName  string            `json:"facilityGroupName"`
	Facilities         []ContentFacility `json:"facilities"`
	Number             int               `json:"number,omitempty"`
	Ind                bool              `json:"ind,omitempty"`
}

// ContentFacility represents a single facility
type ContentFacility struct {
	FacilityCode   int             `json:"facilityCode"`
	FacilityNumber int             `json:"facilityNumber,omitempty"`
	Ind            bool            `json:"ind,omitempty"`
	Content        ContentContent  `json:"content,omitempty"`
}

// ContentContent represents multilingual content
type ContentContent struct {
	ContentText  string `json:"contentText,omitempty"`
	LanguageCode string `json:"languageCode,omitempty"`
}

// ContentInterestPoint represents points of interest near hotel
type ContentInterestPoint struct {
	FacilityCode      int             `json:"facilityCode"`
	FacilityGroupCode int             `json:"facilityGroupCode"`
	Distance          int             `json:"distance"`
	DistanceUnit      string          `json:"distanceUnit"`
	POIName           string          `json:"poiName,omitempty"`
	POICode           int             `json:"poiCode,omitempty"`
	Content           ContentContent  `json:"content,omitempty"`
}

// ContentIssue represents hotel issues
type ContentIssue struct {
	IssueCode  int             `json:"issueCode"`
	IssueType  string          `json:"issueType"`
	Content    ContentContent  `json:"content,omitempty"`
}

// ContentTerms represents hotel terms
type ContentTerms struct {
	InDate   string `json:"inDate,omitempty"`
	OutDate  string `json:"outDate,omitempty"`
}

// ContentAncillaries represents ancillary services
type ContentAncillaries struct {
	CreditCard ContentCreditCard `json:"creditCard,omitempty"`
}

// ContentCreditCard represents credit card info
type ContentCreditCard struct {
	Content []string `json:"content,omitempty"`
}

// ContentHotelsResponse represents HotelBeds hotels Content API response
type ContentHotelsResponse struct {
	From    *int                     `json:"from,omitempty"`
	To      *int                     `json:"to,omitempty"`
	Total   *int                     `json:"total,omitempty"`
	Hotels  []ContentHotelContent     `json:"hotels,omitempty"`
}

// GetHotels fetches hotels from HotelBeds Content API
func (c *Client) GetHotels(ctx context.Context, destinationCode string, offset, limit int) ([]ContentHotelContent, error) {
	// Build endpoint with query parameters
	endpoint := "/hotel-content-api/1.0/hotels"
	queryParams := make(map[string]string)

	if destinationCode != "" {
		queryParams["fields"] = "all" // Get all hotel details
		queryParams["destinationCode"] = destinationCode
	}

	if offset > 0 {
		queryParams["from"] = fmt.Sprintf("%d", offset)
	}
	if limit > 0 {
		queryParams["to"] = fmt.Sprintf("%d", offset + limit - 1)
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

	logger.Infof("Fetching hotels from HotelBeds: destination=%s, offset=%d, limit=%d",
		destinationCode, offset, limit)

	// Make API request
	resp, err := c.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch hotels: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HotelBeds API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log raw response for debugging (first 500 chars)
	logPreview := string(bodyBytes)
	if len(logPreview) > 500 {
		logPreview = logPreview[:500] + "..."
	}
	logger.Infof("Raw Hotelbeds hotels response (preview): %s", logPreview)

	// Parse response
	var response ContentHotelsResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode hotels response: %w\nRaw response: %s", err, string(bodyBytes))
	}

	logger.Infof("Successfully fetched %d hotels from HotelBeds", len(response.Hotels))

	return response.Hotels, nil
}

// GetAllHotels fetches all hotels with pagination
func (c *Client) GetAllHotels(ctx context.Context, destinationCode string) ([]ContentHotelContent, error) {
	var allHotels []ContentHotelContent
	offset := 0
	limit := 1000 // HotelBeds max pagination

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		hotels, err := c.GetHotels(ctx, destinationCode, offset, limit)
		if err != nil {
			return nil, err
		}

		if len(hotels) == 0 {
			break // No more data
		}

		allHotels = append(allHotels, hotels...)
		offset += limit

		logger.Infof("Fetched %d hotels (total: %d)", len(hotels), len(allHotels))
	}

	return allHotels, nil
}

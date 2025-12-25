package hotelbeds

// HotelbedsTypes represents Hotelbeds API response structures

// HotelbedsSearchResponse represents Hotelbeds search API response
type HotelbedsSearchResponse struct {
	Hotels []HotelbedsHotel `json:"hotels"`
}

// HotelbedsHotel represents hotel from Hotelbeds
type HotelbedsHotel struct {
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	CountryCode string            `json:"countryCode"`
	StateCode   string            `json:"stateCode"`
	Destination Code              `json:"destination"`
	Category    map[string]interface{} `json:"category"`
	Images      []HotelbedsImage  `json:"images"`
}

// Code represents destination code
type Code struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// HotelbedsImage represents image from Hotelbeds
type HotelbedsImage struct {
	 imageURL `json:"imageTypeCode"`
	RoomType string  `json:"roomType"`
}

// HotelbedsAvailabilityResponse represents availability response
type HotelbedsAvailabilityResponse struct {
	AuditData    map[string]interface{} `json:"auditData"`
	Hotels       []HotelbedsAvailableHotel `json:"hotels"`
}

// HotelbedsAvailableHotel represents available hotel
type HotelbedsAvailableHotel struct {
	Code       string                      `json:"code"`
	Name       string                      `json:"name"`
	Category   map[string]interface{}       `json:"category"`
	Images     []HotelbedsImage             `json:"images"`
	Rooms      []HotelbedsRoom             `json:"rooms"`
}

// HotelbedsRoom represents room from Hotelbeds
type HotelbedsRoom struct {
	Code        string              `json:"code"`
	Name        string              `json:"name"`
	Rates       []HotelbedsRate     `json:"rates"`
}

// HotelbedsRate represents rate from Hotelbeds
type HotelbedsRate struct {
	RateKey       string  `json:"rateKey"`
	RateType      string  `json:"rateType"`
	NetPrice      float64 `json:"net"`
	Allotment     int     `json:"allotment"`
	Currency      string  `json:"currency"`
	MealPlan      string  `json:"mealPlan"`
	Cancellation  HotelbedsCancellation `json:"cancellation"`
}

// HotelbedsCancellation represents cancellation policy
type HotelbedsCancellation struct {
	FreeCancellationDate string `json:"freeCancellationDate"`
}

// HotelbedsBookingRequest represents booking request to Hotelbeds
type HotelbedsBookingRequest struct {
	Holder     HotelbedsHolder       `json:"holder"`
	Rooms      []HotelbedsRoomBooking `json:"rooms"`
	ClientReference string            `json:"clientReference"`
}

// HotelbedsHolder represents holder information
type HotelbedsHolder struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

// HotelbedsRoomBooking represents room booking
type HotelbedsRoomBooking struct {
	RateKey  string `json:"rateKey"`
	PaxRooms []HotelbedsPaxRoom `json:"paxRooms"`
}

// HotelbedsPaxRoom represents passenger room
type HotelbedsPaxRoom struct {
	PaxType string `json:"paxType"`
	Guests  []HotelbedsGuest `json:"guests"`
}

// HotelbedsGuest represents guest
type HotelbedsGuest struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

// HotelbedsBookingResponse represents booking confirmation
type HotelbedsBookingResponse struct {
	BookingReference string `json:"bookingReference"`
	Hotel            HotelbedsHotelBooking `json:"hotel"`
	Status           string `json:"status"`
}

// HotelbedsHotelBooking represents hotel in booking response
type HotelbedsHotelBooking struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

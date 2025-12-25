package pricing

// HotelCategory represents hotel star rating/category
type HotelCategory int

const (
	CategoryOneStar     HotelCategory = 1
	CategoryTwoStar     HotelCategory = 2
	CategoryThreeStar   HotelCategory = 3
	CategoryFourStar    HotelCategory = 4
	CategoryFiveStar    HotelCategory = 5
)

// PriceCalculation represents the result of price calculation
type PriceCalculation struct {
	NetPrice   int `json:"net_price"`
	SellPrice  int `json:"sell_price"`
	Margin     int `json:"margin"`
	MarginPercent float64 `json:"margin_percent"`
	MarkupPercent float64 `json:"markup_percent"`
}

// PricingConfig represents pricing configuration
type PricingConfig struct {
	BaseMarkupPercent     float64            `json:"base_markup_percent"`
	CategoryMarkupPercent map[HotelCategory]float64 `json:"category_markup_percent"`
}

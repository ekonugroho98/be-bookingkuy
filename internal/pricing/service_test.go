package pricing

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewService tests creating a new pricing service
func TestNewService(t *testing.T) {
	service := NewService()

	require.NotNil(t, service)
}

// TestService_CalculateSellPrice_OneStar tests pricing for 1-star hotel
func TestService_CalculateSellPrice_OneStar(t *testing.T) {
	service := NewService()

	netPrice := 1000000 // 1,000,000 IDR
	category := CategoryOneStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// 1-star hotel: 10% markup
	// margin = 1,000,000 * 0.10 = 100,000
	// sellPrice = 1,000,000 + 100,000 = 1,100,000
	// marginPercent = 100,000 / 1,000,000 * 100 = 10%
	assert.Equal(t, netPrice, result.NetPrice)
	assert.Equal(t, 1100000, result.SellPrice)
	assert.Equal(t, 100000, result.Margin)
	assert.Equal(t, 10.0, result.MarginPercent)
	assert.Equal(t, 10.0, result.MarkupPercent)
}

// TestService_CalculateSellPrice_TwoStar tests pricing for 2-star hotel
func TestService_CalculateSellPrice_TwoStar(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	category := CategoryTwoStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// 2-star hotel: 12% markup
	assert.Equal(t, netPrice, result.NetPrice)
	assert.Equal(t, 1120000, result.SellPrice)
	assert.Equal(t, 120000, result.Margin)
	assert.Equal(t, 12.0, result.MarginPercent)
	assert.Equal(t, 12.0, result.MarkupPercent)
}

// TestService_CalculateSellPrice_ThreeStar tests pricing for 3-star hotel
func TestService_CalculateSellPrice_ThreeStar(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	category := CategoryThreeStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// 3-star hotel: 15% markup
	assert.Equal(t, netPrice, result.NetPrice)
	assert.Equal(t, 1150000, result.SellPrice)
	assert.Equal(t, 150000, result.Margin)
	assert.Equal(t, 15.0, result.MarginPercent)
	assert.Equal(t, 15.0, result.MarkupPercent)
}

// TestService_CalculateSellPrice_FourStar tests pricing for 4-star hotel
func TestService_CalculateSellPrice_FourStar(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	category := CategoryFourStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// 4-star hotel: 18% markup
	assert.Equal(t, netPrice, result.NetPrice)
	assert.Equal(t, 1180000, result.SellPrice)
	assert.Equal(t, 180000, result.Margin)
	assert.Equal(t, 18.0, result.MarginPercent)
	assert.Equal(t, 18.0, result.MarkupPercent)
}

// TestService_CalculateSellPrice_FiveStar tests pricing for 5-star hotel
func TestService_CalculateSellPrice_FiveStar(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	category := CategoryFiveStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// 5-star hotel: 20% markup
	assert.Equal(t, netPrice, result.NetPrice)
	assert.Equal(t, 1200000, result.SellPrice)
	assert.Equal(t, 200000, result.Margin)
	assert.Equal(t, 20.0, result.MarginPercent)
	assert.Equal(t, 20.0, result.MarkupPercent)
}

// TestService_CalculateSellPrice_UnknownCategory tests pricing for unknown category
func TestService_CalculateSellPrice_UnknownCategory(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	category := HotelCategory(99) // Unknown category

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Unknown category: should use base markup (15%)
	assert.Equal(t, netPrice, result.NetPrice)
	assert.Equal(t, 1150000, result.SellPrice)
	assert.Equal(t, 150000, result.Margin)
	assert.Equal(t, 15.0, result.MarginPercent)
	assert.Equal(t, 15.0, result.MarkupPercent)
}

// TestService_CalculateSellPrice_NegativePrice tests error when net price is negative
func TestService_CalculateSellPrice_NegativePrice(t *testing.T) {
	service := NewService()

	netPrice := -1000000 // Negative
	category := CategoryThreeStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "net price cannot be negative")
}

// TestService_CalculateSellPrice_ZeroPrice tests zero price
func TestService_CalculateSellPrice_ZeroPrice(t *testing.T) {
	service := NewService()

	netPrice := 0
	category := CategoryThreeStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Zero price: margin = 0, sellPrice = 0
	// Note: marginPercent will be NaN (0/0) because CalculateSellPrice doesn't handle this case
	assert.Equal(t, 0, result.NetPrice)
	assert.Equal(t, 0, result.SellPrice)
	assert.Equal(t, 0, result.Margin)
	assert.True(t, math.IsNaN(result.MarginPercent), "MarginPercent should be NaN for zero net price")
	assert.Equal(t, 15.0, result.MarkupPercent)
}

// TestService_CalculateSellPrice_LargeAmount tests large price amounts
func TestService_CalculateSellPrice_LargeAmount(t *testing.T) {
	service := NewService()

	netPrice := 10000000 // 10 million IDR
	category := CategoryFiveStar

	result, err := service.CalculateSellPrice(netPrice, category)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// 5-star hotel: 20% markup
	// margin = 10,000,000 * 0.20 = 2,000,000
	// sellPrice = 10,000,000 + 2,000,000 = 12,000,000
	assert.Equal(t, netPrice, result.NetPrice)
	assert.Equal(t, 12000000, result.SellPrice)
	assert.Equal(t, 2000000, result.Margin)
	assert.Equal(t, 20.0, result.MarginPercent)
}

// TestService_CalculateSellPrice_DifferentPrices tests various price points
func TestService_CalculateSellPrice_DifferentPrices(t *testing.T) {
	service := NewService()

	category := CategoryThreeStar
	testCases := []struct {
		name     string
		netPrice int
		expected int
	}{
		{name: "100k", netPrice: 100000, expected: 115000},
		{name: "500k", netPrice: 500000, expected: 575000},
		{name: "1M", netPrice: 1000000, expected: 1150000},
		{name: "2.5M", netPrice: 2500000, expected: 2875000},
		{name: "5M", netPrice: 5000000, expected: 5750000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.CalculateSellPrice(tc.netPrice, category)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expected, result.SellPrice)
			assert.Equal(t, 15.0, result.MarkupPercent)
		})
	}
}

// TestService_CalculateMargin_NormalCase tests margin calculation
func TestService_CalculateMargin_NormalCase(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	sellPrice := 1150000

	margin, marginPercent := service.CalculateMargin(netPrice, sellPrice)

	// Assertions
	assert.Equal(t, 150000, margin)
	assert.Equal(t, 15.0, marginPercent)
}

// TestService_CalculateMargin_ZeroNetPrice tests margin when net price is zero
func TestService_CalculateMargin_ZeroNetPrice(t *testing.T) {
	service := NewService()

	netPrice := 0
	sellPrice := 100000

	margin, marginPercent := service.CalculateMargin(netPrice, sellPrice)

	// Assertions
	assert.Equal(t, 100000, margin)
	assert.Equal(t, 0.0, marginPercent) // Should be 0 to avoid division by zero
}

// TestService_CalculateMargin_NoMargin tests when sell price equals net price
func TestService_CalculateMargin_NoMargin(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	sellPrice := 1000000 // Same as net price

	margin, marginPercent := service.CalculateMargin(netPrice, sellPrice)

	// Assertions
	assert.Equal(t, 0, margin)
	assert.Equal(t, 0.0, marginPercent)
}

// TestService_CalculateMargin_Loss tests when sell price is less than net price
func TestService_CalculateMargin_Loss(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	sellPrice := 900000 // Selling at a loss

	margin, marginPercent := service.CalculateMargin(netPrice, sellPrice)

	// Assertions
	assert.Equal(t, -100000, margin)
	assert.Equal(t, -10.0, marginPercent)
}

// TestService_CalculateMargin_HighMargin tests high margin scenario
func TestService_CalculateMargin_HighMargin(t *testing.T) {
	service := NewService()

	netPrice := 1000000
	sellPrice := 2000000 // 100% markup

	margin, marginPercent := service.CalculateMargin(netPrice, sellPrice)

	// Assertions
	assert.Equal(t, 1000000, margin)
	assert.Equal(t, 100.0, marginPercent)
}

// TestHotelCategory_Constants tests hotel category constants
func TestHotelCategory_Constants(t *testing.T) {
	assert.Equal(t, HotelCategory(1), CategoryOneStar)
	assert.Equal(t, HotelCategory(2), CategoryTwoStar)
	assert.Equal(t, HotelCategory(3), CategoryThreeStar)
	assert.Equal(t, HotelCategory(4), CategoryFourStar)
	assert.Equal(t, HotelCategory(5), CategoryFiveStar)
}

// TestPriceCalculation_Structure tests price calculation structure
func TestPriceCalculation_Structure(t *testing.T) {
	calc := &PriceCalculation{
		NetPrice:      1000000,
		SellPrice:     1150000,
		Margin:        150000,
		MarginPercent: 15.0,
		MarkupPercent: 15.0,
	}

	assert.Equal(t, 1000000, calc.NetPrice)
	assert.Equal(t, 1150000, calc.SellPrice)
	assert.Equal(t, 150000, calc.Margin)
	assert.Equal(t, 15.0, calc.MarginPercent)
	assert.Equal(t, 15.0, calc.MarkupPercent)
}

// TestPricingConfig_Structure tests pricing config structure
func TestPricingConfig_Structure(t *testing.T) {
	config := &PricingConfig{
		BaseMarkupPercent: 15.0,
		CategoryMarkupPercent: map[HotelCategory]float64{
			CategoryOneStar:   10.0,
			CategoryTwoStar:   12.0,
			CategoryThreeStar: 15.0,
			CategoryFourStar:  18.0,
			CategoryFiveStar:  20.0,
		},
	}

	assert.Equal(t, 15.0, config.BaseMarkupPercent)
	assert.Len(t, config.CategoryMarkupPercent, 5)
	assert.Equal(t, 10.0, config.CategoryMarkupPercent[CategoryOneStar])
	assert.Equal(t, 20.0, config.CategoryMarkupPercent[CategoryFiveStar])
}

// TestService_CalculateSellPrice_AllCategories tests all hotel categories
func TestService_CalculateSellPrice_AllCategories(t *testing.T) {
	service := NewService()

	netPrice := 1000000

	categories := []struct {
		name             string
		category         HotelCategory
		expectedSell     int
		expectedMargin   int
		expectedMarkup   float64
	}{
		{name: "One Star", category: CategoryOneStar, expectedSell: 1100000, expectedMargin: 100000, expectedMarkup: 10.0},
		{name: "Two Star", category: CategoryTwoStar, expectedSell: 1120000, expectedMargin: 120000, expectedMarkup: 12.0},
		{name: "Three Star", category: CategoryThreeStar, expectedSell: 1150000, expectedMargin: 150000, expectedMarkup: 15.0},
		{name: "Four Star", category: CategoryFourStar, expectedSell: 1180000, expectedMargin: 180000, expectedMarkup: 18.0},
		{name: "Five Star", category: CategoryFiveStar, expectedSell: 1200000, expectedMargin: 200000, expectedMarkup: 20.0},
	}

	for _, tc := range categories {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.CalculateSellPrice(netPrice, tc.category)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedSell, result.SellPrice)
			assert.Equal(t, tc.expectedMargin, result.Margin)
			assert.Equal(t, tc.expectedMarkup, result.MarkupPercent)
		})
	}
}

// TestService_CalculateMargin_VariousMargins tests various margin scenarios
func TestService_CalculateMargin_VariousMargins(t *testing.T) {
	service := NewService()

	testCases := []struct {
		name           string
		netPrice       int
		sellPrice      int
		expectedMargin int
		expectedPercent float64
	}{
		{name: "10% margin", netPrice: 1000000, sellPrice: 1100000, expectedMargin: 100000, expectedPercent: 10.0},
		{name: "20% margin", netPrice: 1000000, sellPrice: 1200000, expectedMargin: 200000, expectedPercent: 20.0},
		{name: "25% margin", netPrice: 1000000, sellPrice: 1250000, expectedMargin: 250000, expectedPercent: 25.0},
		{name: "50% margin", netPrice: 1000000, sellPrice: 1500000, expectedMargin: 500000, expectedPercent: 50.0},
		{name: "Small margin", netPrice: 500000, sellPrice: 510000, expectedMargin: 10000, expectedPercent: 2.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			margin, marginPercent := service.CalculateMargin(tc.netPrice, tc.sellPrice)

			assert.Equal(t, tc.expectedMargin, margin)
			assert.Equal(t, tc.expectedPercent, marginPercent)
		})
	}
}

// TestService_Integration_EndToEnd tests complete pricing workflow
func TestService_Integration_EndToEnd(t *testing.T) {
	service := NewService()

	// Scenario: Calculate sell price for a 4-star hotel with net price of 2M IDR
	netPrice := 2000000
	category := CategoryFourStar

	// Step 1: Calculate sell price
	priceCalc, err := service.CalculateSellPrice(netPrice, category)
	require.NoError(t, err)
	require.NotNil(t, priceCalc)

	// Verify pricing
	assert.Equal(t, netPrice, priceCalc.NetPrice)
	assert.Equal(t, 2360000, priceCalc.SellPrice) // 2M + 18% = 2.36M
	assert.Equal(t, 360000, priceCalc.Margin)
	assert.Equal(t, 18.0, priceCalc.MarginPercent)
	assert.Equal(t, 18.0, priceCalc.MarkupPercent)

	// Step 2: Verify margin calculation independently
	margin, marginPercent := service.CalculateMargin(netPrice, priceCalc.SellPrice)

	assert.Equal(t, priceCalc.Margin, margin)
	assert.Equal(t, priceCalc.MarginPercent, marginPercent)
}

// TestService_CalculateSellPrice_Rounding tests rounding behavior
func TestService_CalculateSellPrice_Rounding(t *testing.T) {
	service := NewService()

	// Use a price that might not divide evenly
	netPrice := 333333
	category := CategoryThreeStar // 15% markup

	result, err := service.CalculateSellPrice(netPrice, category)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify calculation
	// margin = 333,333 * 0.15 = 49,999.95 → 49,999 (int truncates)
	expectedMargin := 333333 * 15 / 100
	expectedSellPrice := netPrice + expectedMargin

	assert.Equal(t, expectedSellPrice, result.SellPrice)
	assert.Equal(t, expectedMargin, result.Margin)
}

// TestService_DefaultConfig tests default configuration values
func TestService_DefaultConfig(t *testing.T) {
	service := NewService().(*service)

	require.NotNil(t, service.config)
	assert.Equal(t, 15.0, service.config.BaseMarkupPercent)
	assert.Len(t, service.config.CategoryMarkupPercent, 5)

	// Verify all category markups
	assert.Equal(t, 10.0, service.config.CategoryMarkupPercent[CategoryOneStar])
	assert.Equal(t, 12.0, service.config.CategoryMarkupPercent[CategoryTwoStar])
	assert.Equal(t, 15.0, service.config.CategoryMarkupPercent[CategoryThreeStar])
	assert.Equal(t, 18.0, service.config.CategoryMarkupPercent[CategoryFourStar])
	assert.Equal(t, 20.0, service.config.CategoryMarkupPercent[CategoryFiveStar])
}

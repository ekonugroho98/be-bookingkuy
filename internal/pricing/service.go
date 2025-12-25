package pricing

import (
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Service defines interface for pricing operations
type Service interface {
	CalculateSellPrice(netPrice int, category HotelCategory) (*PriceCalculation, error)
	CalculateMargin(netPrice, sellPrice int) (int, float64)
}

type service struct {
	config *PricingConfig
}

// NewService creates a new pricing service with default configuration
func NewService() Service {
	return &service{
		config: &PricingConfig{
			BaseMarkupPercent: 15.0, // 15% base markup
			CategoryMarkupPercent: map[HotelCategory]float64{
				CategoryOneStar:   10.0,
				CategoryTwoStar:   12.0,
				CategoryThreeStar: 15.0,
				CategoryFourStar:  18.0,
				CategoryFiveStar:  20.0,
			},
		},
	}
}

// CalculateSellPrice calculates the sell price from net price
func (s *service) CalculateSellPrice(netPrice int, category HotelCategory) (*PriceCalculation, error) {
	if netPrice < 0 {
		return nil, fmt.Errorf("net price cannot be negative")
	}

	// Get markup percent based on category
	markupPercent := s.getMarkupPercent(category)

	// Calculate sell price: netPrice + (netPrice * markupPercent / 100)
	margin := int(float64(netPrice) * markupPercent / 100)
	sellPrice := netPrice + margin

	// Calculate margin percentage
	marginPercent := float64(margin) / float64(netPrice) * 100

	calc := &PriceCalculation{
		NetPrice:      netPrice,
		SellPrice:     sellPrice,
		Margin:        margin,
		MarginPercent: marginPercent,
		MarkupPercent: markupPercent,
	}

	logger.Infof("Price calculation: net=%d, sell=%d, margin=%d (%.2f%%), markup=%.2f%%",
		netPrice, sellPrice, margin, marginPercent, markupPercent)

	return calc, nil
}

// CalculateMargin calculates the margin and margin percent
func (s *service) CalculateMargin(netPrice, sellPrice int) (int, float64) {
	margin := sellPrice - netPrice
	if netPrice == 0 {
		return margin, 0.0
	}
	marginPercent := float64(margin) / float64(netPrice) * 100
	return margin, marginPercent
}

// getMarkupPercent returns the markup percentage based on hotel category
func (s *service) getMarkupPercent(category HotelCategory) float64 {
	if markup, ok := s.config.CategoryMarkupPercent[category]; ok {
		return markup
	}
	return s.config.BaseMarkupPercent
}

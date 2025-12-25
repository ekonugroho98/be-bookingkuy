package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Registry manages provider instances
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
// Ini yang membuat menambah provider baru jadi sangat mudah!
func (r *Registry) Register(provider Provider) {
	r.providers[provider.Name()] = provider
	logger.Infof("Provider registered: %s", provider.Name())
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// GetAll returns all registered providers
func (r *Registry) GetAll() []Provider {
	providers := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// GetHealthy returns all healthy providers
func (r *Registry) GetHealthy(ctx context.Context) []Provider {
	healthy := make([]Provider, 0)
	for _, provider := range r.providers {
		if err := provider.HealthCheck(ctx); err == nil {
			healthy = append(healthy, provider)
		} else {
			logger.Warnf("Provider %s is unhealthy: %v", provider.Name(), err)
		}
	}
	return healthy
}

// GetByPriority returns providers sorted by priority
// Ini untuk implementasi cheap-first strategy!
func (r *Registry) GetByPriority(ctx context.Context) []Provider {
	healthy := r.GetHealthy(ctx)

	// Sort by priority (lower = higher priority)
	sort.Slice(healthy, func(i, j int) bool {
		// TODO: Add priority field to Provider interface
		// For now, just return in insertion order
		return healthy[i].Name() < healthy[j].Name()
	})

	return healthy
}

// SearchAll searches all providers and aggregates results
func (r *Registry) SearchAll(ctx context.Context, req *AvailabilityRequest) (*AvailabilityResponse, error) {
	healthyProviders := r.GetHealthy(ctx)
	if len(healthyProviders) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}

	// Search all providers in parallel
	// TODO: Implement parallel search with goroutines
	// For now, just use the first healthy provider
	provider := healthyProviders[0]
	return provider.SearchAvailability(ctx, req)
}

// CreateBookingWithFallback creates booking with automatic failover
func (r *Registry) CreateBookingWithFallback(ctx context.Context, req *BookingRequest) (*BookingConfirmation, error) {
	providers := r.GetByPriority(ctx)

	for _, provider := range providers {
		logger.Infof("Attempting to create booking with provider: %s", provider.Name())

		confirmation, err := provider.CreateBooking(ctx, req)
		if err != nil {
			logger.Warnf("Failed to create booking with %s: %v, trying next provider", provider.Name(), err)
			continue
		}

		logger.Infof("Booking created successfully with provider: %s", provider.Name())
		return confirmation, nil
	}

	return nil, fmt.Errorf("failed to create booking with any provider")
}

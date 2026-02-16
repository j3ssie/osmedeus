package cloud

import (
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/config"
)

// ProviderRegistry manages provider instances
type ProviderRegistry struct {
	providers map[ProviderType]Provider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[ProviderType]Provider),
	}
}

// Register registers a provider
func (r *ProviderRegistry) Register(provider Provider) {
	r.providers[provider.Type()] = provider
}

// Get retrieves a provider by type
func (r *ProviderRegistry) Get(providerType ProviderType) (Provider, error) {
	provider, ok := r.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("provider %s not registered", providerType)
	}
	return provider, nil
}

// List returns all registered provider types
func (r *ProviderRegistry) List() []ProviderType {
	types := make([]ProviderType, 0, len(r.providers))
	for t := range r.providers {
		types = append(types, t)
	}
	return types
}

// CreateProvider creates a provider instance based on cloud configuration
func CreateProvider(cfg *config.CloudConfigs, providerType ProviderType) (Provider, error) {
	switch providerType {
	case ProviderDigitalOcean:
		return NewDigitalOceanProvider(
			cfg.Providers.DigitalOcean.Token,
			cfg.Providers.DigitalOcean.Region,
			cfg.Providers.DigitalOcean.Size,
			cfg.Providers.DigitalOcean.SnapshotID,
			cfg.Providers.DigitalOcean.SSHKeyFingerprint,
		)
	case ProviderAWS:
		return nil, fmt.Errorf("AWS provider not yet implemented")
	case ProviderGCP:
		return nil, fmt.Errorf("GCP provider not yet implemented")
	case ProviderLinode:
		return nil, fmt.Errorf("linode provider not yet implemented")
	case ProviderAzure:
		return nil, fmt.Errorf("azure provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

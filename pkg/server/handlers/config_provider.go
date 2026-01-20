package handlers

import (
	"github.com/j3ssie/osmedeus/v5/internal/config"
)

// ConfigProvider provides access to the current configuration.
// This interface allows handlers to retrieve fresh configuration on each request,
// supporting hot reload when enabled.
type ConfigProvider interface {
	// Get returns the current configuration.
	Get() *config.Config
}

// StaticConfigProvider wraps a static config for backward compatibility.
// It always returns the same configuration instance.
type StaticConfigProvider struct {
	cfg *config.Config
}

// NewStaticConfigProvider creates a new StaticConfigProvider.
func NewStaticConfigProvider(cfg *config.Config) *StaticConfigProvider {
	return &StaticConfigProvider{cfg: cfg}
}

// Get returns the static configuration.
func (p *StaticConfigProvider) Get() *config.Config {
	return p.cfg
}

// HotReloadConfigProvider wraps a HotReloadableConfig.
// It returns the latest configuration on each call.
type HotReloadConfigProvider struct {
	hotConfig *config.HotReloadableConfig
}

// NewHotReloadConfigProvider creates a new HotReloadConfigProvider.
func NewHotReloadConfigProvider(hotConfig *config.HotReloadableConfig) *HotReloadConfigProvider {
	return &HotReloadConfigProvider{hotConfig: hotConfig}
}

// Get returns the current configuration from the hot reload config.
func (p *HotReloadConfigProvider) Get() *config.Config {
	return p.hotConfig.Get()
}

// GetHotConfig returns the underlying HotReloadableConfig.
// This is useful for accessing hot reload specific functionality like version and reload.
func (p *HotReloadConfigProvider) GetHotConfig() *config.HotReloadableConfig {
	return p.hotConfig
}

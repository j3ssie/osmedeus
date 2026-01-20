package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// HotReloadableConfig provides thread-safe configuration hot reload functionality.
// It watches for changes to the configuration file and automatically reloads
// when changes are detected.
type HotReloadableConfig struct {
	current atomic.Value  // *Config - lock-free reads
	version atomic.Uint64 // version counter for tracking config updates

	watcher    *fsnotify.Watcher
	configPath string
	baseFolder string

	// Thread-safe callback management
	callbacksMu sync.RWMutex
	callbacks   map[int]func(old, new *Config)
	nextID      int

	// Debouncing (300ms default)
	debounceMu       sync.Mutex
	debounceTimer    *time.Timer
	debounceDuration time.Duration

	running atomic.Bool
	stopCh  chan struct{}
	log     *zap.Logger
}

// Option configures a HotReloadableConfig.
type Option func(*HotReloadableConfig)

// WithDebounceDuration sets the debounce duration for file change events.
// Default is 300ms.
func WithDebounceDuration(d time.Duration) Option {
	return func(h *HotReloadableConfig) {
		h.debounceDuration = d
	}
}

// WithLogger sets a custom logger for the hot reload config.
func WithLogger(log *zap.Logger) Option {
	return func(h *HotReloadableConfig) {
		h.log = log
	}
}

// NewHotReloadableConfig creates a new HotReloadableConfig for the given base folder.
// It loads the initial configuration from osm-settings.yaml in the base folder.
func NewHotReloadableConfig(baseFolder string, opts ...Option) (*HotReloadableConfig, error) {
	configPath := filepath.Join(baseFolder, "osm-settings.yaml")

	// Check that the config file exists
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	h := &HotReloadableConfig{
		configPath:       configPath,
		baseFolder:       baseFolder,
		callbacks:        make(map[int]func(old, new *Config)),
		debounceDuration: 300 * time.Millisecond,
		stopCh:           make(chan struct{}),
		log:              logger.Get(),
	}

	// Apply options
	for _, opt := range opts {
		opt(h)
	}

	// Load initial config
	cfg, err := h.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}
	h.current.Store(cfg)
	h.version.Store(1)

	// Also set the global config for backward compatibility
	Set(cfg)

	return h, nil
}

// Get returns the current configuration. This is a lock-free read.
func (h *HotReloadableConfig) Get() *Config {
	return h.current.Load().(*Config)
}

// GetVersion returns the current configuration version number.
// This increments each time the configuration is reloaded.
func (h *HotReloadableConfig) GetVersion() uint64 {
	return h.version.Load()
}

// Watch starts watching the configuration file for changes.
// File changes are debounced to handle editors that perform multiple write operations.
func (h *HotReloadableConfig) Watch() error {
	if h.running.Load() {
		return fmt.Errorf("already watching")
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	h.watcher = watcher

	// Watch the directory (not the file) for vim/emacs compatibility
	// These editors delete and recreate files on save
	configDir := filepath.Dir(h.configPath)
	if err := watcher.Add(configDir); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	h.running.Store(true)

	go h.watchLoop()

	h.log.Info("Started config hot reload watcher",
		zap.String("config_path", h.configPath),
		zap.Duration("debounce", h.debounceDuration),
	)

	return nil
}

// watchLoop is the main event loop for file watching.
func (h *HotReloadableConfig) watchLoop() {
	configName := filepath.Base(h.configPath)

	for {
		select {
		case <-h.stopCh:
			return
		case event, ok := <-h.watcher.Events:
			if !ok {
				return
			}

			// Only care about our config file
			if filepath.Base(event.Name) != configName {
				continue
			}

			// Check for write or create events
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				h.scheduleReload()
			}

		case err, ok := <-h.watcher.Errors:
			if !ok {
				return
			}
			h.log.Error("Config watcher error", zap.Error(err))
		}
	}
}

// scheduleReload schedules a config reload with debouncing.
func (h *HotReloadableConfig) scheduleReload() {
	h.debounceMu.Lock()
	defer h.debounceMu.Unlock()

	// Cancel existing timer if any
	if h.debounceTimer != nil {
		h.debounceTimer.Stop()
	}

	// Schedule new reload
	h.debounceTimer = time.AfterFunc(h.debounceDuration, func() {
		if err := h.Reload(); err != nil {
			h.log.Error("Failed to reload config",
				zap.Error(err),
				zap.String("config_path", h.configPath),
			)
		}
	})
}

// Reload manually reloads the configuration from disk.
// Returns an error if the new config is invalid (keeps old config).
func (h *HotReloadableConfig) Reload() error {
	newCfg, err := h.loadConfig()
	if err != nil {
		h.log.Warn("Config reload failed - keeping old config",
			zap.Error(err),
		)
		return err
	}

	oldCfg := h.Get()
	h.current.Store(newCfg)
	newVersion := h.version.Add(1)

	// Also update the global config for backward compatibility
	Set(newCfg)

	h.log.Info("Configuration reloaded",
		zap.Uint64("version", newVersion),
		zap.String("config_path", h.configPath),
	)

	// Notify callbacks (in goroutine with panic recovery)
	h.notifyCallbacks(oldCfg, newCfg)

	return nil
}

// loadConfig loads and validates the configuration from disk.
func (h *HotReloadableConfig) loadConfig() (*Config, error) {
	data, err := os.ReadFile(h.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse with strict validation
	cfg, err := ParseConfigStrict(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Override base folder if needed
	if h.baseFolder != "" {
		cfg.BaseFolder = h.baseFolder
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Resolve paths
	cfg.ResolvePaths()

	return cfg, nil
}

// notifyCallbacks notifies all registered callbacks about a config change.
func (h *HotReloadableConfig) notifyCallbacks(old, new *Config) {
	h.callbacksMu.RLock()
	callbacks := make([]func(old, new *Config), 0, len(h.callbacks))
	for _, cb := range h.callbacks {
		callbacks = append(callbacks, cb)
	}
	h.callbacksMu.RUnlock()

	for _, cb := range callbacks {
		go func(callback func(old, new *Config)) {
			defer func() {
				if r := recover(); r != nil {
					h.log.Error("Panic in config change callback",
						zap.Any("panic", r),
					)
				}
			}()
			callback(old, new)
		}(cb)
	}
}

// OnChange registers a callback to be called when the configuration changes.
// Returns an unsubscribe function that can be called to remove the callback.
func (h *HotReloadableConfig) OnChange(fn func(old, new *Config)) func() {
	h.callbacksMu.Lock()
	id := h.nextID
	h.nextID++
	h.callbacks[id] = fn
	h.callbacksMu.Unlock()

	// Return unsubscribe function
	return func() {
		h.callbacksMu.Lock()
		delete(h.callbacks, id)
		h.callbacksMu.Unlock()
	}
}

// Stop stops watching for configuration changes and releases resources.
func (h *HotReloadableConfig) Stop() error {
	if !h.running.Load() {
		return nil
	}

	h.running.Store(false)
	close(h.stopCh)

	// Cancel any pending debounce timer
	h.debounceMu.Lock()
	if h.debounceTimer != nil {
		h.debounceTimer.Stop()
		h.debounceTimer = nil
	}
	h.debounceMu.Unlock()

	// Close the watcher
	if h.watcher != nil {
		if err := h.watcher.Close(); err != nil {
			return fmt.Errorf("failed to close watcher: %w", err)
		}
	}

	h.log.Info("Stopped config hot reload watcher")

	return nil
}

// IsRunning returns true if the watcher is currently running.
func (h *HotReloadableConfig) IsRunning() bool {
	return h.running.Load()
}

// GetConfigPath returns the path to the configuration file being watched.
func (h *HotReloadableConfig) GetConfigPath() string {
	return h.configPath
}

// CallbackCount returns the number of registered callbacks.
func (h *HotReloadableConfig) CallbackCount() int {
	h.callbacksMu.RLock()
	defer h.callbacksMu.RUnlock()
	return len(h.callbacks)
}

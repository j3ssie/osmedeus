package config

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewHotReloadableConfig(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	// Write initial config
	initialConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create hot reload config
	hotCfg, err := NewHotReloadableConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	// Verify initial config is loaded
	cfg := hotCfg.Get()
	if cfg == nil {
		t.Fatal("Expected config to be loaded")
	}
	if cfg.Server.Port != 8002 {
		t.Errorf("Expected port 8002, got %d", cfg.Server.Port)
	}

	// Verify version is 1
	if v := hotCfg.GetVersion(); v != 1 {
		t.Errorf("Expected version 1, got %d", v)
	}
}

func TestNewHotReloadableConfig_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create config file

	_, err := NewHotReloadableConfig(tmpDir)
	if err == nil {
		t.Error("Expected error when config file not found")
	}
}

func TestHotReloadableConfig_Reload(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	// Write initial config
	initialConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create hot reload config
	hotCfg, err := NewHotReloadableConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	// Verify initial port
	if hotCfg.Get().Server.Port != 8002 {
		t.Errorf("Expected initial port 8002, got %d", hotCfg.Get().Server.Port)
	}

	// Update config file
	updatedConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 9000
`
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}

	// Manual reload
	if err := hotCfg.Reload(); err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	// Verify updated port
	if hotCfg.Get().Server.Port != 9000 {
		t.Errorf("Expected updated port 9000, got %d", hotCfg.Get().Server.Port)
	}

	// Verify version incremented
	if v := hotCfg.GetVersion(); v != 2 {
		t.Errorf("Expected version 2, got %d", v)
	}
}

func TestHotReloadableConfig_ReloadInvalidConfig(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	// Write initial config
	initialConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create hot reload config
	hotCfg, err := NewHotReloadableConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	// Write invalid config
	invalidConfig := `
this is not: valid yaml
  because: indentation is wrong
    and: missing base_folder
`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Reload should fail
	if err := hotCfg.Reload(); err == nil {
		t.Error("Expected error when reloading invalid config")
	}

	// Old config should be preserved
	if hotCfg.Get().Server.Port != 8002 {
		t.Errorf("Expected old port 8002 to be preserved, got %d", hotCfg.Get().Server.Port)
	}

	// Version should not have changed
	if v := hotCfg.GetVersion(); v != 1 {
		t.Errorf("Expected version 1 (unchanged), got %d", v)
	}
}

func TestHotReloadableConfig_OnChange(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	// Write initial config
	initialConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create hot reload config
	hotCfg, err := NewHotReloadableConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	// Register callback
	var callbackCalled atomic.Bool
	var oldPort, newPort int
	done := make(chan struct{})

	unsubscribe := hotCfg.OnChange(func(old, new *Config) {
		callbackCalled.Store(true)
		oldPort = old.Server.Port
		newPort = new.Server.Port
		close(done)
	})

	// Update config file
	updatedConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 9000
`
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}

	// Reload
	if err := hotCfg.Reload(); err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	// Wait for callback with timeout
	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Fatal("Callback was not called within timeout")
	}

	if !callbackCalled.Load() {
		t.Error("Expected callback to be called")
	}
	if oldPort != 8002 {
		t.Errorf("Expected old port 8002, got %d", oldPort)
	}
	if newPort != 9000 {
		t.Errorf("Expected new port 9000, got %d", newPort)
	}

	// Test unsubscribe
	unsubscribe()
	if hotCfg.CallbackCount() != 0 {
		t.Errorf("Expected 0 callbacks after unsubscribe, got %d", hotCfg.CallbackCount())
	}
}

func TestHotReloadableConfig_Watch(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	// Write initial config
	initialConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create hot reload config with short debounce for testing
	hotCfg, err := NewHotReloadableConfig(tmpDir, WithDebounceDuration(50*time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	// Start watching
	if err := hotCfg.Watch(); err != nil {
		t.Fatalf("Failed to start watching: %v", err)
	}
	defer func() { _ = hotCfg.Stop() }()

	// Verify running
	if !hotCfg.IsRunning() {
		t.Error("Expected watcher to be running")
	}

	// Register callback to detect changes
	reloaded := make(chan struct{})
	hotCfg.OnChange(func(old, new *Config) {
		close(reloaded)
	})

	// Wait a bit for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Update config file
	updatedConfig := `
base_folder: /tmp/test
server:
  host: "0.0.0.0"
  port: 9000
`
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}

	// Wait for reload with timeout
	select {
	case <-reloaded:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("Config was not reloaded within timeout")
	}

	// Verify config was updated
	if hotCfg.Get().Server.Port != 9000 {
		t.Errorf("Expected port 9000 after file change, got %d", hotCfg.Get().Server.Port)
	}
}

func TestHotReloadableConfig_Stop(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	// Write initial config
	initialConfig := `
base_folder: /tmp/test
server:
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create hot reload config
	hotCfg, err := NewHotReloadableConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	// Start watching
	if err := hotCfg.Watch(); err != nil {
		t.Fatalf("Failed to start watching: %v", err)
	}

	// Stop
	if err := hotCfg.Stop(); err != nil {
		t.Fatalf("Failed to stop: %v", err)
	}

	// Verify not running
	if hotCfg.IsRunning() {
		t.Error("Expected watcher to be stopped")
	}

	// Stop again should be no-op
	if err := hotCfg.Stop(); err != nil {
		t.Errorf("Second stop should not error: %v", err)
	}
}

func TestHotReloadableConfig_WatchTwice(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	// Write initial config
	initialConfig := `
base_folder: /tmp/test
server:
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create hot reload config
	hotCfg, err := NewHotReloadableConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}
	defer func() { _ = hotCfg.Stop() }()

	// Start watching
	if err := hotCfg.Watch(); err != nil {
		t.Fatalf("Failed to start watching: %v", err)
	}

	// Watch again should error
	if err := hotCfg.Watch(); err == nil {
		t.Error("Expected error when watching twice")
	}
}

func TestWithDebounceDuration(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	initialConfig := `
base_folder: /tmp/test
server:
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	customDuration := 500 * time.Millisecond
	hotCfg, err := NewHotReloadableConfig(tmpDir, WithDebounceDuration(customDuration))
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	if hotCfg.debounceDuration != customDuration {
		t.Errorf("Expected debounce duration %v, got %v", customDuration, hotCfg.debounceDuration)
	}
}

func TestHotReloadableConfig_GetConfigPath(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "osm-settings.yaml")

	initialConfig := `
base_folder: /tmp/test
server:
  port: 8002
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	hotCfg, err := NewHotReloadableConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hot reload config: %v", err)
	}

	if hotCfg.GetConfigPath() != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, hotCfg.GetConfigPath())
	}
}

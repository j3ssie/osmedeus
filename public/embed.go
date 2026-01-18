package public

import (
	"embed"
	"io/fs"
)

//go:embed favicon.ico
//go:embed presets/*
//go:embed all:examples/osmedeus-base.example
//go:embed all:ui
var EmbedFS embed.FS

// GetFavicon returns the favicon.ico file contents
func GetFavicon() ([]byte, error) {
	return EmbedFS.ReadFile("favicon.ico")
}

// GetPresetsFS returns a sub-filesystem for presets directory
func GetPresetsFS() (fs.FS, error) {
	return fs.Sub(EmbedFS, "presets")
}

// GetUIFS returns a sub-filesystem for UI files
func GetUIFS() (fs.FS, error) {
	return fs.Sub(EmbedFS, "ui")
}

// GetRegistryMetadata returns the embedded binary registry JSON
func GetRegistryMetadata() ([]byte, error) {
	return EmbedFS.ReadFile("presets/registry-metadata-direct-fetch.json")
}

// GetFlakeNix returns the embedded flake.nix content
func GetFlakeNix() ([]byte, error) {
	return EmbedFS.ReadFile("presets/flake.nix")
}

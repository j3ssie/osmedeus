package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeRegistryFile writes a registry JSON file and returns its path
func writeRegistryFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "registry.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func callRegistryInfo(t *testing.T, query string) (*http.Response, map[string]interface{}) {
	t.Helper()
	app := fiber.New()
	app.Get("/osm/api/registry-info", GetRegistryInfo(&config.Config{}))

	resp, err := app.Test(httptest.NewRequest("GET", "/osm/api/registry-info"+query, nil), 10000)
	require.NoError(t, err)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &body), "response: %s", raw)
	return resp, body
}

// TestGetRegistryInfo_CustomRegistryDoesNotExecuteValidateCommand guards the endpoint
// against command execution: valide-command is run as shell, so a registry supplied
// through the registry_url query param must never have it executed. "amass" is used as
// the entry name because nix-build mode only reads metadata for tools listed in flake.nix.
func TestGetRegistryInfo_CustomRegistryDoesNotExecuteValidateCommand(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "executed.txt")
	registryPath := writeRegistryFile(t, `{
		"amass": {"desc": "hi", "valide-command": "touch `+marker+`"}
	}`)

	for _, mode := range []string{"direct-fetch", "nix-build"} {
		t.Run(mode, func(t *testing.T) {
			resp, _ := callRegistryInfo(t, "?registry_mode="+mode+"&registry_url="+registryPath)
			assert.Equal(t, 200, resp.StatusCode)

			_, statErr := os.Stat(marker)
			assert.True(t, os.IsNotExist(statErr),
				"valide-command from a caller-supplied registry must not be executed")
		})
	}
}

// TestGetRegistryInfo_EmbeddedRegistry checks the default response shape is unchanged.
func TestGetRegistryInfo_EmbeddedRegistry(t *testing.T) {
	resp, body := callRegistryInfo(t, "")
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "direct-fetch", body["registry_mode"])
	assert.NotEmpty(t, body["registry_url"], "embedded registry should report a concrete source")
	assert.NotEmpty(t, body["binaries"])
}

// TestGetRegistryInfo_CustomRegistryReturnsItsBinaries checks the feature itself still
// works: a caller-supplied registry replaces the binary list in direct-fetch mode.
func TestGetRegistryInfo_CustomRegistryReturnsItsBinaries(t *testing.T) {
	registryPath := writeRegistryFile(t, `{"only-tool": {"desc": "the only one"}}`)

	resp, body := callRegistryInfo(t, "?registry_url="+registryPath)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, registryPath, body["registry_url"])

	binaries, ok := body["binaries"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, binaries, "only-tool")
	assert.Len(t, binaries, 1)
}

// TestGetRegistryInfo_NixBuildReportsRegistrySource checks both modes name a concrete
// source, so callers can tell which registry produced the metadata.
func TestGetRegistryInfo_NixBuildReportsRegistrySource(t *testing.T) {
	registryPath := writeRegistryFile(t, `{"amass": {"desc": "custom desc"}}`)

	resp, body := callRegistryInfo(t, "?registry_mode=nix-build&registry_url="+registryPath)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "nix-build", body["registry_mode"])
	assert.Equal(t, registryPath, body["registry_url"])

	_, defaultBody := callRegistryInfo(t, "?registry_mode=nix-build")
	assert.NotEmpty(t, defaultBody["registry_url"], "embedded metadata source should be reported too")
}

// TestGetRegistryInfo_BadCustomRegistryIsReported ensures a registry_url that cannot be
// loaded surfaces an error rather than a success response with the metadata silently gone.
func TestGetRegistryInfo_BadCustomRegistryIsReported(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.json")

	for _, mode := range []string{"direct-fetch", "nix-build"} {
		t.Run(mode, func(t *testing.T) {
			resp, body := callRegistryInfo(t, "?registry_mode="+mode+"&registry_url="+missing)
			assert.Equal(t, 500, resp.StatusCode)
			assert.Equal(t, true, body["error"])
		})
	}
}

package updater

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAssetURL(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		repo     string
		version  string
		expected string
	}{
		{
			name:     "version with v prefix",
			owner:    "j3ssie",
			repo:     "osmedeus",
			version:  "v5.1.0",
			expected: buildExpectedURL("j3ssie", "osmedeus", "5.1.0"),
		},
		{
			name:     "version without v prefix",
			owner:    "j3ssie",
			repo:     "osmedeus",
			version:  "5.2.0",
			expected: buildExpectedURL("j3ssie", "osmedeus", "5.2.0"),
		},
		{
			name:     "different owner/repo",
			owner:    "testowner",
			repo:     "testrepo",
			version:  "v1.0.0",
			expected: buildExpectedURL("testowner", "testrepo", "1.0.0"),
		},
		{
			name:     "prerelease version",
			owner:    "j3ssie",
			repo:     "osmedeus",
			version:  "v5.0.0-beta.1",
			expected: buildExpectedURL("j3ssie", "osmedeus", "5.0.0-beta.1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDirectDownloader(tt.owner, tt.repo, false)
			url := d.BuildAssetURL(tt.version)
			assert.Equal(t, tt.expected, url)
		})
	}
}

// buildExpectedURL constructs the expected URL based on current runtime
func buildExpectedURL(owner, repo, version string) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	ext := "tar.gz"
	if osName == "windows" {
		ext = "zip"
	}
	return "https://github.com/" + owner + "/" + repo + "/releases/download/v" + version + "/osmedeus_" + version + "_" + osName + "_" + arch + "." + ext
}

func TestBuildAssetURL_PlatformSpecific(t *testing.T) {
	d := NewDirectDownloader("j3ssie", "osmedeus", false)
	url := d.BuildAssetURL("v5.1.0")

	// Verify it contains the correct OS
	assert.Contains(t, url, runtime.GOOS)
	// Verify it contains the correct arch
	assert.Contains(t, url, runtime.GOARCH)
	// Verify it contains the version without v prefix in filename
	assert.Contains(t, url, "osmedeus_5.1.0_")
	// Verify it points to the correct repo
	assert.Contains(t, url, "github.com/j3ssie/osmedeus/releases/download/v5.1.0/")
}

func TestNewDirectDownloader(t *testing.T) {
	d := NewDirectDownloader("owner", "repo", true)

	assert.Equal(t, "owner", d.owner)
	assert.Equal(t, "repo", d.repo)
	assert.True(t, d.verbose)
}

func TestNewDirectDownloader_Verbose(t *testing.T) {
	d1 := NewDirectDownloader("owner", "repo", false)
	assert.False(t, d1.verbose)

	d2 := NewDirectDownloader("owner", "repo", true)
	assert.True(t, d2.verbose)
}

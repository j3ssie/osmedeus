package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSubPath(t *testing.T) {
	tests := []struct {
		name     string
		parent   string
		child    string
		expected bool
	}{
		{
			name:     "child inside parent",
			parent:   "/home/user/osmedeus-base",
			child:    "/home/user/osmedeus-base/workflows",
			expected: true,
		},
		{
			name:     "child is parent",
			parent:   "/home/user/osmedeus-base",
			child:    "/home/user/osmedeus-base",
			expected: true,
		},
		{
			name:     "child outside parent",
			parent:   "/home/user/osmedeus-base",
			child:    "/opt/workflows",
			expected: false,
		},
		{
			name:     "child is sibling",
			parent:   "/home/user/osmedeus-base",
			child:    "/home/user/other-folder",
			expected: false,
		},
		{
			name:     "empty parent",
			parent:   "",
			child:    "/home/user/workflows",
			expected: false,
		},
		{
			name:     "empty child",
			parent:   "/home/user/osmedeus-base",
			child:    "",
			expected: false,
		},
		{
			name:     "relative paths - child inside",
			parent:   "osmedeus-base",
			child:    "osmedeus-base/workflows",
			expected: true,
		},
		{
			name:     "relative paths - child outside",
			parent:   "osmedeus-base",
			child:    "other-folder",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSubPath(tt.parent, tt.child)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaybePrependSudo(t *testing.T) {
	// On darwin/windows, maybePrependSudo is a no-op
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		assert.Equal(t, "apt install coreutils", maybePrependSudo("apt install coreutils"),
			"should be no-op on darwin/windows")
		return
	}

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"apt install", "apt install coreutils", "sudo apt install coreutils"},
		{"apt-get install", "apt-get install -y curl", "sudo apt-get install -y curl"},
		{"dnf install", "dnf install nmap", "sudo dnf install nmap"},
		{"yum install", "yum install git", "sudo yum install git"},
		{"pacman install", "pacman -S nmap", "sudo pacman -S nmap"},
		{"already has sudo", "sudo apt install coreutils", "sudo apt install coreutils"},
		{"go install unchanged", "go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest", "go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest"},
		{"pip unchanged", "pip install semgrep", "pip install semgrep"},
		{"git clone unchanged", "git clone https://github.com/example/repo", "git clone https://github.com/example/repo"},
		{"curl unchanged", "curl -fsSL https://example.com | bash", "curl -fsSL https://example.com | bash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maybePrependSudo(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		expect bool
	}{
		{"github repo", "https://github.com/owner/repo", true},
		{"raw content", "https://raw.githubusercontent.com/owner/repo/main/f.json", true},
		{"api subdomain", "https://api.github.com/repos/owner/repo/releases", true},
		{"release objects", "https://objects.githubusercontent.com/foo", true},
		{"uppercase host", "https://GitHub.com/owner/repo", true},
		{"with port", "https://github.com:443/owner/repo", true},
		{"lookalike suffix host", "https://github.com.evil.tld/owner/repo", false},
		{"host as query param", "https://evil.tld/?x=github.com", false},
		{"host in path", "https://evil.tld/github.com/owner/repo", false},
		{"userinfo trick", "https://github.com@evil.tld/repo", false},
		{"unrelated host", "https://gitlab.com/owner/repo", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, IsGitHubURL(tt.url),
				"the GitHub token is attached based on this check")
		})
	}
}

func TestIsBinaryInstalledNoExec(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "executed.txt")
	entry := &BinaryEntry{ValidateCommand: "touch " + marker}

	assert.False(t, IsBinaryInstalledNoExec("definitely-not-a-real-binary-xyz", entry))
	_, err := os.Stat(marker)
	assert.True(t, os.IsNotExist(err), "no-exec variant must not run the validate command")

	// The trusted variant still executes it
	assert.True(t, IsBinaryInstalled("definitely-not-a-real-binary-xyz", entry))
	_, err = os.Stat(marker)
	assert.NoError(t, err, "trusted variant should still run the validate command")
}

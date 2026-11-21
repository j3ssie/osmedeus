package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// stubSudo sets the sudo seams for one test and restores them afterwards.
func stubSudo(t *testing.T, applies, root, available bool) {
	t.Helper()
	origApplies, origRoot, origAvail := sudoApplies, isRoot, sudoAvailable
	t.Cleanup(func() { sudoApplies, isRoot, sudoAvailable = origApplies, origRoot, origAvail })
	sudoApplies = applies
	isRoot = func() bool { return root }
	sudoAvailable = func() bool { return available }
}

func TestMaybePrependSudo(t *testing.T) {
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

	t.Run("non-root with sudo", func(t *testing.T) {
		stubSudo(t, true, false, true)
		for _, tt := range tests {
			got, err := maybePrependSudo(tt.input)
			require.NoError(t, err, tt.name)
			assert.Equal(t, tt.expect, got, tt.name)
		}
	})

	t.Run("root never needs sudo", func(t *testing.T) {
		stubSudo(t, true, true, false)
		got, err := maybePrependSudo("apt install coreutils")
		require.NoError(t, err)
		assert.Equal(t, "apt install coreutils", got)
	})

	t.Run("darwin and windows are left alone", func(t *testing.T) {
		stubSudo(t, false, false, false)
		got, err := maybePrependSudo("apt install coreutils")
		require.NoError(t, err)
		assert.Equal(t, "apt install coreutils", got)
	})

	// Issue #330: a rootless container without sudo ran "sudo apt install ..."
	// and failed with "sh: sudo: not found" / exit 127.
	t.Run("non-root without sudo says so instead of running", func(t *testing.T) {
		stubSudo(t, true, false, false)

		_, err := maybePrependSudo("apt install coreutils")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"apt install coreutils" needs root`)
		assert.Contains(t, err.Error(), "sudo is not installed")

		_, err = maybePrependSudo("sudo apt install coreutils")
		require.Error(t, err, "an explicit sudo cannot run either")

		got, err := maybePrependSudo("go install example.com/tool@latest")
		require.NoError(t, err, "commands that need no root still run")
		assert.Equal(t, "go install example.com/tool@latest", got)
	})
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

func TestInjectGitHubToken(t *testing.T) {
	t.Setenv("GITHUB_API_KEY", "tok")
	assert.Equal(t, "https://tok@github.com/owner/repo.git", injectGitHubToken("https://github.com/owner/repo.git"))
	assert.Equal(t, "https://github.com.example.org/owner/repo.git", injectGitHubToken("https://github.com.example.org/owner/repo.git"))
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

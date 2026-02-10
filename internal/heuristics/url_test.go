package heuristics

import "testing"

func TestExtractRepoSlug(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		urlPath string
		want    string
	}{
		// GitHub
		{"github basic", "github.com", "/juice-shop/juice-shop", "juice-shop__juice-shop"},
		{"github archive", "github.com", "/juice-shop/juice-shop/archive/refs/heads/master.zip", "juice-shop__juice-shop"},
		{"github releases", "github.com", "/owner/repo/releases/tag/v1.0", "owner__repo"},
		{"github blob", "github.com", "/owner/repo/blob/main/README.md", "owner__repo"},
		{"github tree", "github.com", "/owner/repo/tree/main/src", "owner__repo"},
		{"github pull", "github.com", "/owner/repo/pull/123", "owner__repo"},
		{"github issues", "github.com", "/owner/repo/issues/456", "owner__repo"},
		{"github trailing slash", "github.com", "/owner/repo/", "owner__repo"},

		// GitLab
		{"gitlab basic", "gitlab.com", "/org/repo", "org__repo"},
		{"gitlab nested group", "gitlab.com", "/org/subgroup/repo", "org__subgroup__repo"},
		{"gitlab archive path", "gitlab.com", "/org/repo/-/archive/main/repo.tar.gz", "org__repo"},
		{"gitlab pipeline", "gitlab.com", "/org/repo/-/pipelines", "org__repo"},

		// Bitbucket
		{"bitbucket basic", "bitbucket.org", "/owner/repo", "owner__repo"},

		// Codeberg
		{"codeberg basic", "codeberg.org", "/owner/repo", "owner__repo"},

		// Non-code-hosting
		{"not code host", "example.com", "/foo/bar", ""},
		{"custom domain", "git.company.com", "/owner/repo", ""},

		// Edge cases
		{"empty path", "github.com", "", ""},
		{"root path", "github.com", "/", ""},
		{"single segment", "github.com", "/org", ""},
		{"single segment trailing", "github.com", "/org/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractRepoSlug(tt.host, tt.urlPath)
			if got != tt.want {
				t.Errorf("ExtractRepoSlug(%q, %q) = %q, want %q", tt.host, tt.urlPath, got, tt.want)
			}
		})
	}
}

func TestParseURL_RepoSlug(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		wantSlug string
		wantRoot string
	}{
		{
			"github repo URL",
			"https://github.com/juice-shop/juice-shop",
			"juice-shop__juice-shop",
			"github.com",
		},
		{
			"github archive URL",
			"https://github.com/juice-shop/juice-shop/archive/refs/heads/master.zip",
			"juice-shop__juice-shop",
			"github.com",
		},
		{
			"gitlab nested URL",
			"https://gitlab.com/org/subgroup/repo",
			"org__subgroup__repo",
			"gitlab.com",
		},
		{
			"non-code-hosting URL",
			"https://example.com/foo/bar",
			"",
			"example.com",
		},
		{
			"github without scheme",
			"github.com/owner/repo",
			"owner__repo",
			"github.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseURL(tt.rawURL)
			if err != nil {
				t.Fatalf("ParseURL(%q) error: %v", tt.rawURL, err)
			}
			if info.RepoSlug != tt.wantSlug {
				t.Errorf("ParseURL(%q).RepoSlug = %q, want %q", tt.rawURL, info.RepoSlug, tt.wantSlug)
			}
			if info.RootDomain != tt.wantRoot {
				t.Errorf("ParseURL(%q).RootDomain = %q, want %q", tt.rawURL, info.RootDomain, tt.wantRoot)
			}
		})
	}
}

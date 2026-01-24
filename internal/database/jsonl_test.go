package database

import "testing"

func TestClassifyAssetType(t *testing.T) {
	tests := []struct {
		name       string
		assetValue string
		expected   string
	}{
		// URL tests
		{"https URL", "https://example.com/path", "url"},
		{"http URL", "http://example.com/api", "url"},
		{"https URL with port", "https://example.com:8080/path", "url"},
		{"https URL with query", "https://example.com/path?foo=bar", "url"},

		// IP tests
		{"IPv4 address", "192.168.1.1", "ip"},
		{"IPv4 loopback", "127.0.0.1", "ip"},
		{"IPv6 loopback", "::1", "ip"},
		{"IPv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "ip"},
		{"IPv6 abbreviated", "2001:db8::1", "ip"},

		// repo_name tests
		{"GitHub repo", "anthropics/claude", "repo_name"},
		{"GitHub repo with numbers", "j3ssie/osmedeus", "repo_name"},
		{"Single word org and repo", "owner/repo", "repo_name"},

		// domain tests
		{"Simple domain", "example.com", "domain"},
		{"Subdomain", "api.example.com", "domain"},
		{"Multi-level subdomain", "api.staging.example.com", "domain"},
		{"Domain with hyphen", "my-site.example.com", "domain"},
		{"TLD only dot", "localhost.localdomain", "domain"},

		// unknown tests
		{"Single word", "localhost", "unknown"},
		{"Empty string", "", "unknown"},
		{"Word with spaces", "hello world", "unknown"},
		{"Just a slash", "/", "unknown"},
		{"Path only", "/path/to/file", "unknown"},

		// Edge cases
		{"IP with port should be domain (contains colon but not valid IP)", "192.168.1.1:8080", "domain"},
		{"Repo-like with dot", "user/repo.git", "domain"},
		{"Multiple slashes no dot", "a/b/c", "unknown"},
		{"Path with dot", "path/to/file.txt", "domain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyAssetType(tt.assetValue)
			if result != tt.expected {
				t.Errorf("ClassifyAssetType(%q) = %q, expected %q", tt.assetValue, result, tt.expected)
			}
		})
	}
}

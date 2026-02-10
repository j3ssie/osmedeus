package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterestingUrls_BasicDeduplication(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "urls.txt")
	destFile := filepath.Join(tmpDir, "interesting.txt")

	// URLs with same hostname, path, and param names should dedupe
	content := `http://sample.example.com/product.aspx?productID=123&type=customer
http://sample.example.com/product.aspx?productID=456&type=admin
http://other.example.com/api/data?id=1
http://other.example.com/api/data?id=2
`
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`interesting_urls("`+srcFile+`", "`+destFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Should have 2 unique URLs (deduplicated by hostname+path+params)
	output, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, "http://sample.example.com/product.aspx?productID=123&type=customer\nhttp://other.example.com/api/data?id=1\n", string(output))
}

func TestInterestingUrls_StaticFileFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "urls.txt")
	destFile := filepath.Join(tmpDir, "interesting.txt")

	// Static files should be filtered, but .js should be kept
	content := `http://example.com/style.css
http://example.com/font.woff2
http://example.com/image.png
http://example.com/icon.ico
http://example.com/app.js
http://example.com/api/endpoint
`
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`interesting_urls("`+srcFile+`", "`+destFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	output, err := os.ReadFile(destFile)
	require.NoError(t, err)
	// Only .js and non-static URLs should remain
	assert.Contains(t, string(output), "http://example.com/app.js")
	assert.Contains(t, string(output), "http://example.com/api/endpoint")
	assert.NotContains(t, string(output), "style.css")
	assert.NotContains(t, string(output), "font.woff2")
	assert.NotContains(t, string(output), "image.png")
	assert.NotContains(t, string(output), "icon.ico")
}

func TestInterestingUrls_NoisePatternFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "urls.txt")
	destFile := filepath.Join(tmpDir, "interesting.txt")

	// Noise patterns like blog, news, calendar dates should be filtered
	content := `https://www.example.com/cn/news/all-news/public-1.html
https://www.example.com/de/blog/2022/01/02/blog-title.htm
https://www.example.com/data/0001.html
https://www.example.com/data/0002.html
https://www.example.com/api/v1/users
`
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`interesting_urls("`+srcFile+`", "`+destFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	output, err := os.ReadFile(destFile)
	require.NoError(t, err)
	// Non-noise URL should be kept
	assert.Contains(t, string(output), "https://www.example.com/api/v1/users")
}

func TestInterestingUrls_JSONFieldExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "urls.jsonl")
	destFile := filepath.Join(tmpDir, "interesting.jsonl")

	// JSONL input with URL field
	content := `{"url":"http://example.com/api/v1?id=1","status":200}
{"url":"http://example.com/api/v1?id=2","status":200}
{"url":"http://example.com/api/v2?id=1","status":200}
`
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`interesting_urls("`+srcFile+`", "`+destFile+`", "url")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	output, err := os.ReadFile(destFile)
	require.NoError(t, err)
	// Should have 2 unique URLs and preserve original JSON
	assert.Contains(t, string(output), `"url":"http://example.com/api/v1?id=1"`)
	assert.Contains(t, string(output), `"url":"http://example.com/api/v2?id=1"`)
}

func TestInterestingUrls_LongPathFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "urls.txt")
	destFile := filepath.Join(tmpDir, "interesting.txt")

	// Long path segments (>100 chars) should be filtered
	longPath := "https://example.com/" + string(make([]byte, 101)) + "/page"
	content := longPath + "\nhttps://example.com/short/page\n"
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`interesting_urls("`+srcFile+`", "`+destFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	output, err := os.ReadFile(destFile)
	require.NoError(t, err)
	// Only short path should remain
	assert.Contains(t, string(output), "https://example.com/short/page")
}

func TestInterestingUrls_EmptyArguments(t *testing.T) {
	registry := NewRegistry()

	// Empty src
	result, err := registry.Execute(
		`interesting_urls("", "/tmp/dest.txt")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, false, result)

	// Empty dest
	result, err = registry.Execute(
		`interesting_urls("/tmp/src.txt", "")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestInterestingUrls_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	destFile := filepath.Join(tmpDir, "interesting.txt")

	registry := NewRegistry()
	result, err := registry.Execute(
		`interesting_urls("/nonexistent/file.txt", "`+destFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestInterestingUrls_DashyPathFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "urls.txt")
	destFile := filepath.Join(tmpDir, "interesting.txt")

	// Paths with >3 dashes should be filtered
	content := `https://example.com/this-is-a-very-long-slug-path/page
https://example.com/short-path/page
`
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`interesting_urls("`+srcFile+`", "`+destFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	output, err := os.ReadFile(destFile)
	require.NoError(t, err)
	// Only short-path should remain
	assert.Contains(t, string(output), "https://example.com/short-path/page")
	assert.NotContains(t, string(output), "this-is-a-very-long-slug-path")
}

func TestIsStaticURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://example.com/style.css", true},
		{"http://example.com/font.woff2", true},
		{"http://example.com/image.png", true},
		{"http://example.com/icon.ico", true},
		{"http://example.com/video.mp4", true},
		{"http://example.com/app.js", false},   // .js should NOT be filtered
		{"http://example.com/api/data", false}, // API endpoints should not be filtered
		{"http://example.com/page.html", false},
		{"http://example.com/style.css?v=123", true}, // Query params don't matter
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isStaticURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNoiseURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/blog/2022/post", true},     // blog path
		{"https://example.com/news/article", true},       // news path
		{"https://example.com/2022/01/02/page", true},    // calendar date
		{"https://example.com/data/12345.html", true},    // numeric-only path
		{"https://example.com/data/12345", true},         // numeric-only path no ext
		{"https://example.com/api/v1/users", false},      // API endpoint
		{"https://example.com/product/details", true},    // product is noise path
		{"https://example.com/articles/tech-news", true}, // articles path
		{"https://example.com/careers/engineer", true},   // careers path
		{"https://example.com/search?q=test", false},     // search endpoint
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isNoiseURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractHostname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple domain", "example.com", "example.com"},
		{"domain with www", "www.example.com", "www.example.com"},
		{"http URL", "http://example.com", "example.com"},
		{"https URL", "https://example.com", "example.com"},
		{"URL with path", "https://example.com/path/to/page", "example.com"},
		{"URL with port", "https://example.com:8080/path", "example.com"},
		{"URL with query", "https://example.com/page?foo=bar", "example.com"},
		{"domain with port no scheme", "example.com:8080", "example.com"},
		{"domain with path no scheme", "example.com/path", "example.com"},
		{"subdomain URL", "https://api.sub.example.com/v1", "api.sub.example.com"},
		{"IP address", "192.168.1.1", "192.168.1.1"},
		{"IP with port", "192.168.1.1:8080", "192.168.1.1"},
		{"IP URL", "http://192.168.1.1/path", "192.168.1.1"},
		{"empty string", "", ""},
		{"whitespace", "  example.com  ", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHostname(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIP_WithRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test with a well-known domain (google.com should always resolve)
	result, err := registry.Execute(`get_ip("google.com")`, map[string]interface{}{})
	require.NoError(t, err)
	// Should return a non-empty IP string
	ip, ok := result.(string)
	assert.True(t, ok, "result should be a string")
	assert.NotEmpty(t, ip, "should resolve to an IP")

	// Test with URL format
	result, err = registry.Execute(`get_ip("https://google.com/path")`, map[string]interface{}{})
	require.NoError(t, err)
	ip, ok = result.(string)
	assert.True(t, ok, "result should be a string")
	assert.NotEmpty(t, ip, "should resolve URL to an IP")

	// Test with empty input
	result, err = registry.Execute(`get_ip("")`, map[string]interface{}{})
	require.NoError(t, err)
	ip, ok = result.(string)
	assert.True(t, ok, "result should be a string")
	assert.Empty(t, ip, "empty input should return empty string")

	// Test with invalid domain
	result, err = registry.Execute(`get_ip("this-domain-does-not-exist-12345.invalid")`, map[string]interface{}{})
	require.NoError(t, err)
	ip, ok = result.(string)
	assert.True(t, ok, "result should be a string")
	assert.Empty(t, ip, "invalid domain should return empty string")
}

func TestResolveToIP(t *testing.T) {
	// Test with empty hostname
	result := resolveToIP("")
	assert.Empty(t, result)

	// Test with invalid hostname
	result = resolveToIP("this-domain-does-not-exist-12345.invalid")
	assert.Empty(t, result)

	// Test with localhost (should resolve)
	result = resolveToIP("localhost")
	// localhost typically resolves to 127.0.0.1 or ::1
	// We just check it returns something
	assert.NotEmpty(t, result, "localhost should resolve")
}

func TestGetParentURLImpl(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with file and query",
			input:    "https://example.com/j3ssie/sample.php?query=123",
			expected: "https://example.com/j3ssie/",
		},
		{
			name:     "URL with trailing slash",
			input:    "https://example.com/a/b/c/",
			expected: "https://example.com/a/b/",
		},
		{
			name:     "URL with single path",
			input:    "https://example.com/file.txt",
			expected: "https://example.com/",
		},
		{
			name:     "URL with root path",
			input:    "https://example.com/",
			expected: "https://example.com/",
		},
		{
			name:     "URL without path",
			input:    "https://example.com",
			expected: "https://example.com/",
		},
		{
			name:     "URL with port and path",
			input:    "http://example.com:8080/path/to/file",
			expected: "http://example.com:8080/path/to/",
		},
		{
			name:     "URL with fragment",
			input:    "https://example.com/path/file.html#section",
			expected: "https://example.com/path/",
		},
		{
			name:     "URL with query and fragment",
			input:    "https://example.com/api/data?id=1#results",
			expected: "https://example.com/api/",
		},
		{
			name:     "deep nested path",
			input:    "https://example.com/a/b/c/d/e/f.js",
			expected: "https://example.com/a/b/c/d/e/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getParentURLImpl(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetParentURL_WithRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test basic URL
	result, err := registry.Execute(`get_parent_url("https://example.com/path/file.txt")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/path/", result)

	// Test with query parameters
	result, err = registry.Execute(`get_parent_url("https://example.com/api/endpoint?foo=bar")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/api/", result)

	// Test with empty input
	result, err = registry.Execute(`get_parent_url("")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestParseURLImpl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		format   string
		expected string
	}{
		// Basic directives
		{
			name:     "scheme",
			url:      "https://example.com",
			format:   "%s",
			expected: "https",
		},
		{
			name:     "full domain",
			url:      "https://sub.example.com",
			format:   "%d",
			expected: "sub.example.com",
		},
		{
			name:     "subdomain",
			url:      "https://sub.example.com",
			format:   "%S",
			expected: "sub",
		},
		{
			name:     "root domain",
			url:      "https://sub.example.com",
			format:   "%r",
			expected: "example",
		},
		{
			name:     "tld",
			url:      "https://sub.example.com",
			format:   "%t",
			expected: "com",
		},
		{
			name:     "port",
			url:      "https://example.com:8080/path",
			format:   "%P",
			expected: "8080",
		},
		{
			name:     "path",
			url:      "https://example.com/path/file.jpg",
			format:   "%p",
			expected: "/path/file.jpg",
		},
		{
			name:     "extension",
			url:      "https://example.com/path/file.jpg",
			format:   "%e",
			expected: "jpg",
		},
		{
			name:     "query string",
			url:      "https://example.com?a=1&b=2",
			format:   "%q",
			expected: "a=1&b=2",
		},
		{
			name:     "fragment",
			url:      "https://example.com#section",
			format:   "%f",
			expected: "section",
		},
		{
			name:     "user info",
			url:      "https://user:pass@example.com",
			format:   "%u",
			expected: "user:pass",
		},

		// Conditional directives
		{
			name:     "at sign with user info",
			url:      "https://user:pass@example.com",
			format:   "%u%@%d",
			expected: "user:pass@example.com",
		},
		{
			name:     "at sign without user info",
			url:      "https://example.com",
			format:   "%u%@%d",
			expected: "example.com",
		},
		{
			name:     "colon with port",
			url:      "https://example.com:8080",
			format:   "%d%:%P",
			expected: "example.com:8080",
		},
		{
			name:     "colon without port",
			url:      "https://example.com",
			format:   "%d%:%P",
			expected: "example.com",
		},
		{
			name:     "question mark with query",
			url:      "https://example.com?q=1",
			format:   "%d%?%q",
			expected: "example.com?q=1",
		},
		{
			name:     "question mark without query",
			url:      "https://example.com",
			format:   "%d%?%q",
			expected: "example.com",
		},
		{
			name:     "hash with fragment",
			url:      "https://example.com#section",
			format:   "%d%#%f",
			expected: "example.com#section",
		},
		{
			name:     "hash without fragment",
			url:      "https://example.com",
			format:   "%d%#%f",
			expected: "example.com",
		},

		// Authority directive
		{
			name:     "authority full",
			url:      "https://user:pass@example.com:8080",
			format:   "%a",
			expected: "user:pass@example.com:8080",
		},
		{
			name:     "authority no port",
			url:      "https://user:pass@example.com",
			format:   "%a",
			expected: "user:pass@example.com",
		},
		{
			name:     "authority no user",
			url:      "https://example.com:8080",
			format:   "%a",
			expected: "example.com:8080",
		},
		{
			name:     "authority simple",
			url:      "https://example.com",
			format:   "%a",
			expected: "example.com",
		},

		// Literal percent
		{
			name:     "literal percent",
			url:      "https://example.com",
			format:   "100%%",
			expected: "100%",
		},

		// Combined formats
		{
			name:     "subdomain and root",
			url:      "https://api.sub.example.com",
			format:   "%S.%r.%t",
			expected: "api.sub.example.com",
		},
		{
			name:     "scheme and domain",
			url:      "https://example.com",
			format:   "%s://%d",
			expected: "https://example.com",
		},
		{
			name:     "full URL reconstruction",
			url:      "https://example.com:8080/path?q=1#sec",
			format:   "%s://%d%:%P%p%?%q%#%f",
			expected: "https://example.com:8080/path?q=1#sec",
		},

		// Multi-part TLDs
		{
			name:     "co.uk domain",
			url:      "https://api.example.co.uk",
			format:   "%S|%r|%t",
			expected: "api|example|co.uk",
		},
		{
			name:     "com.au domain",
			url:      "https://sub.site.com.au",
			format:   "%S|%r|%t",
			expected: "sub|site|com.au",
		},

		// Edge cases
		{
			name:     "no extension",
			url:      "https://example.com/path/file",
			format:   "%e",
			expected: "",
		},
		{
			name:     "no subdomain",
			url:      "https://example.com",
			format:   "%S",
			expected: "",
		},
		{
			name:     "empty path",
			url:      "https://example.com",
			format:   "%p",
			expected: "",
		},
		{
			name:     "unknown directive",
			url:      "https://example.com",
			format:   "%z",
			expected: "%z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseURLImpl(tt.url, tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseURL_WithRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test scheme extraction
	result, err := registry.Execute(`parse_url("https://example.com", "%s")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "https", result)

	// Test domain extraction
	result, err = registry.Execute(`parse_url("https://sub.example.com", "%d")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "sub.example.com", result)

	// Test extension extraction
	result, err = registry.Execute(`parse_url("https://example.com/path/image.jpg", "%e")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "jpg", result)

	// Test with empty URL
	result, err = registry.Execute(`parse_url("", "%s")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "", result)

	// Test with empty format
	result, err = registry.Execute(`parse_url("https://example.com", "")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestExtractDomainParts(t *testing.T) {
	tests := []struct {
		name              string
		domain            string
		expectedSubdomain string
		expectedRoot      string
		expectedTLD       string
	}{
		{
			name:              "simple domain",
			domain:            "example.com",
			expectedSubdomain: "",
			expectedRoot:      "example",
			expectedTLD:       "com",
		},
		{
			name:              "subdomain",
			domain:            "sub.example.com",
			expectedSubdomain: "sub",
			expectedRoot:      "example",
			expectedTLD:       "com",
		},
		{
			name:              "multiple subdomains",
			domain:            "api.sub.example.com",
			expectedSubdomain: "api.sub",
			expectedRoot:      "example",
			expectedTLD:       "com",
		},
		{
			name:              "co.uk domain",
			domain:            "example.co.uk",
			expectedSubdomain: "",
			expectedRoot:      "example",
			expectedTLD:       "co.uk",
		},
		{
			name:              "co.uk with subdomain",
			domain:            "api.example.co.uk",
			expectedSubdomain: "api",
			expectedRoot:      "example",
			expectedTLD:       "co.uk",
		},
		{
			name:              "com.au domain",
			domain:            "site.com.au",
			expectedSubdomain: "",
			expectedRoot:      "site",
			expectedTLD:       "com.au",
		},
		{
			name:              "single part",
			domain:            "localhost",
			expectedSubdomain: "",
			expectedRoot:      "localhost",
			expectedTLD:       "",
		},
		{
			name:              "empty domain",
			domain:            "",
			expectedSubdomain: "",
			expectedRoot:      "",
			expectedTLD:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subdomain, root, tld := extractDomainParts(tt.domain)
			assert.Equal(t, tt.expectedSubdomain, subdomain, "subdomain mismatch")
			assert.Equal(t, tt.expectedRoot, root, "root mismatch")
			assert.Equal(t, tt.expectedTLD, tld, "tld mismatch")
		})
	}
}

func TestQueryReplaceImpl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		value    string
		mode     string
		expected string
	}{
		{"single param replace", "https://example.com?id=123", "new", "replace", "https://example.com?id=new"},
		{"multiple params replace", "https://example.com/path?one=1&two=2", "newval", "replace", "https://example.com/path?one=newval&two=newval"},
		{"append mode", "https://example.com?a=1&b=2", "FUZZ", "append", "https://example.com?a=1FUZZ&b=2FUZZ"},
		{"no query params", "https://example.com/path", "new", "replace", "https://example.com/path"},
		{"empty value", "https://example.com?a=1", "", "replace", "https://example.com?a="},
		{"with fragment", "https://example.com?a=1#section", "new", "replace", "https://example.com?a=new#section"},
		{"default mode", "https://example.com?x=old", "test", "", "https://example.com?x=test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := queryReplaceImpl(tt.url, tt.value, tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPathReplaceImpl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		value    string
		position int
		expected string
	}{
		{"replace first", "https://example.com/first/second/third", "new", 1, "https://example.com/new/second/third"},
		{"replace second", "https://example.com/first/second/third", "new", 2, "https://example.com/first/new/third"},
		{"replace third", "https://example.com/first/second/third", "new", 3, "https://example.com/first/second/new"},
		{"replace all", "https://example.com/a/b/c", "x", 0, "https://example.com/x/x/x"},
		{"replace all negative", "https://example.com/a/b/c", "x", -1, "https://example.com/x/x/x"},
		{"position out of range", "https://example.com/a/b", "new", 5, "https://example.com/a/b"},
		{"with query", "https://example.com/a/b?q=1", "new", 1, "https://example.com/new/b?q=1"},
		{"single segment", "https://example.com/only", "new", 1, "https://example.com/new"},
		{"no path", "https://example.com", "new", 1, "https://example.com"},
		{"root path only", "https://example.com/", "new", 1, "https://example.com/"},
		{"with fragment", "https://example.com/a/b#sec", "new", 2, "https://example.com/a/new#sec"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathReplaceImpl(tt.url, tt.value, tt.position)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryReplace_WithRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test replace mode (default)
	result, err := registry.Execute(`query_replace("https://example.com?a=1&b=2", "test")`, nil)
	require.NoError(t, err)
	resultStr := result.(string)
	assert.Contains(t, resultStr, "a=test")
	assert.Contains(t, resultStr, "b=test")

	// Test append mode
	result, err = registry.Execute(`query_replace("https://example.com?x=old", "FUZZ", "append")`, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com?x=oldFUZZ", result)

	// Test with empty URL
	result, err = registry.Execute(`query_replace("", "test")`, nil)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestPathReplace_WithRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test default position (1)
	result, err := registry.Execute(`path_replace("https://example.com/a/b/c", "new")`, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/new/b/c", result)

	// Test specific position
	result, err = registry.Execute(`path_replace("https://example.com/a/b/c", "new", 2)`, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/a/new/c", result)

	// Test replace all (position 0)
	result, err = registry.Execute(`path_replace("https://example.com/a/b/c", "x", 0)`, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/x/x/x", result)

	// Test with empty URL
	result, err = registry.Execute(`path_replace("", "test")`, nil)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

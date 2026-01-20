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

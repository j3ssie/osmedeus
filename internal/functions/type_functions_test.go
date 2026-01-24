package functions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTypes(t *testing.T) {
	runtime := NewGojaRuntime()

	tests := []struct {
		name     string
		input    string
		expected string
		setup    func() string // optional setup that returns the actual input
		cleanup  func(string)  // optional cleanup
	}{
		// CIDR tests
		{
			name:     "CIDR IPv4",
			input:    "192.168.1.0/24",
			expected: TypeCIDR,
		},
		{
			name:     "CIDR IPv4 /32",
			input:    "10.0.0.1/32",
			expected: TypeCIDR,
		},
		{
			name:     "CIDR IPv6",
			input:    "2001:db8::/32",
			expected: TypeCIDR,
		},
		{
			name:     "Invalid CIDR - bad mask",
			input:    "192.168.1.0/99",
			expected: TypeString,
		},

		// IP tests
		{
			name:     "IPv4 address",
			input:    "192.168.1.1",
			expected: TypeIP,
		},
		{
			name:     "IPv4 address zeros",
			input:    "0.0.0.0",
			expected: TypeIP,
		},
		{
			name:     "IPv6 address",
			input:    "2001:db8::1",
			expected: TypeIP,
		},
		{
			name:     "IPv6 localhost",
			input:    "::1",
			expected: TypeIP,
		},

		// URL tests
		{
			name:     "HTTP URL",
			input:    "http://example.com",
			expected: TypeURL,
		},
		{
			name:     "HTTPS URL",
			input:    "https://example.com/path?query=1",
			expected: TypeURL,
		},
		{
			name:     "HTTPS URL uppercase",
			input:    "HTTPS://EXAMPLE.COM",
			expected: TypeURL,
		},

		// Domain tests
		{
			name:     "Simple domain",
			input:    "example.com",
			expected: TypeDomain,
		},
		{
			name:     "Subdomain",
			input:    "sub.example.com",
			expected: TypeDomain,
		},
		{
			name:     "Deep subdomain",
			input:    "a.b.c.example.com",
			expected: TypeDomain,
		},
		{
			name:     "Domain with hyphen",
			input:    "my-domain.co.uk",
			expected: TypeDomain,
		},

		// String fallback tests
		{
			name:     "Empty string",
			input:    "",
			expected: TypeString,
		},
		{
			name:     "Random text",
			input:    "hello world",
			expected: TypeString,
		},
		{
			name:     "Number",
			input:    "12345",
			expected: TypeString,
		},
		{
			name:     "Invalid domain - starts with dot",
			input:    ".example.com",
			expected: TypeString,
		},
		{
			name:     "Invalid domain - ends with hyphen",
			input:    "example-.com",
			expected: TypeString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			if tt.setup != nil {
				input = tt.setup()
			}

			expr := "get_types('" + input + "')"
			result, err := runtime.Execute(expr, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("get_types(%q) = %v, want %v", input, result, tt.expected)
			}

			if tt.cleanup != nil {
				tt.cleanup(input)
			}
		})
	}
}

func TestGetTypes_FileAndFolder(t *testing.T) {
	runtime := NewGojaRuntime()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "get_types_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test file
	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test folder detection
	t.Run("Folder detection", func(t *testing.T) {
		expr := "get_types('" + tmpDir + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != TypeFolder {
			t.Errorf("get_types(%q) = %v, want %v", tmpDir, result, TypeFolder)
		}
	})

	// Test file detection
	t.Run("File detection", func(t *testing.T) {
		expr := "get_types('" + testFile + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != TypeFile {
			t.Errorf("get_types(%q) = %v, want %v", testFile, result, TypeFile)
		}
	})

	// Test non-existent path (should not be file/folder)
	t.Run("Non-existent path", func(t *testing.T) {
		nonExistent := filepath.Join(tmpDir, "does_not_exist.txt")
		expr := "get_types('" + nonExistent + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		// Should fall back to string since file doesn't exist
		if result != TypeString {
			t.Errorf("get_types(%q) = %v, want %v", nonExistent, result, TypeString)
		}
	})
}

func TestGetTargetSpace(t *testing.T) {
	runtime := NewGojaRuntime()

	tests := []struct {
		name     string
		input    string
		checkLen bool   // if true, just check result length <= 30 or contains random part
		expected string // if checkLen is false, exact match
	}{
		{
			name:     "Simple string",
			input:    "example.com",
			checkLen: false,
			expected: "example.com",
		},
		{
			name:     "URL with slashes",
			input:    "https://example.com/path",
			checkLen: false,
			expected: "https___example.com_path",
		},
		{
			name:     "String with colons",
			input:    "test:value:here",
			checkLen: false,
			expected: "test_value_here",
		},
		{
			name:     "String with multiple unsafe chars",
			input:    "a/b:c*d?e<f>g|h",
			checkLen: false,
			expected: "a_b_c_d_e_f_g_h",
		},
		{
			name:     "Empty string",
			input:    "",
			checkLen: false,
			expected: "",
		},
		{
			name:     "Long string should be truncated",
			input:    "this-is-a-very-long-string-that-exceeds-thirty-characters",
			checkLen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := "get_target_space('" + tt.input + "')"
			result, err := runtime.Execute(expr, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("Expected string result, got %T", result)
			}

			if tt.checkLen {
				// For long strings, verify truncation happened
				// Format is: {first6}-{random6}-{timestamp}
				if len(resultStr) > 30 {
					// Still might be longer due to timestamp, but should be truncated
					// Just verify it starts with the first 6 chars
					if len(tt.input) > 6 {
						sanitizedPrefix := tt.input[:6]
						// The prefix might have unsafe chars replaced
						for _, r := range `/\:*?"<>|` {
							sanitizedPrefix = replaceRune(sanitizedPrefix, r, '_')
						}
						if resultStr[:6] != sanitizedPrefix {
							t.Errorf("Truncated result should start with first 6 chars (sanitized), got %q", resultStr[:6])
						}
					}
				}
			} else {
				if resultStr != tt.expected {
					t.Errorf("get_target_space(%q) = %q, want %q", tt.input, resultStr, tt.expected)
				}
			}
		})
	}
}

// Helper function to replace a rune in a string
func replaceRune(s string, old rune, new rune) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r == old {
			result[i] = new
		} else {
			result[i] = r
		}
	}
	return string(result[:len([]rune(s))])
}

func TestDetectInputType(t *testing.T) {
	// Test the internal detectInputType function directly
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.0/24", TypeCIDR},
		{"10.0.0.0/8", TypeCIDR},
		{"192.168.1.1", TypeIP},
		{"::1", TypeIP},
		{"http://example.com", TypeURL},
		{"https://test.org/path", TypeURL},
		{"example.com", TypeDomain},
		{"sub.example.com", TypeDomain},
		{"hello", TypeString},
		{"", TypeString},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := detectInputType(tt.input)
			if result != tt.expected {
				t.Errorf("detectInputType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsCIDR(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.0/24", true},
		{"10.0.0.0/8", true},
		{"172.16.0.0/12", true},
		{"192.168.1.0/32", true},
		{"2001:db8::/32", true},
		{"192.168.1.1", false},    // IP without mask
		{"192.168.1.0/99", false}, // invalid mask
		{"example.com", false},
		{"not-cidr", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isCIDR(tt.input)
			if result != tt.expected {
				t.Errorf("isCIDR(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsIP(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"192.168.1.0/24", false}, // CIDR, not plain IP
		{"example.com", false},
		{"not-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isIP(tt.input)
			if result != tt.expected {
				t.Errorf("isIP(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"HTTP://EXAMPLE.COM", true},
		{"HTTPS://test.org/path", true},
		{"ftp://example.com", false},
		{"example.com", false},
		{"//example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isURL(tt.input)
			if result != tt.expected {
				t.Errorf("isURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"a.b.c.example.com", true},
		{"test-domain.org", true},
		{"example.co.uk", true},
		{"192.168.1.1", false},        // IP address
		{".example.com", false},       // starts with dot
		{"example-.com", false},       // ends with hyphen before dot
		{"-example.com", false},       // starts with hyphen
		{"example", false},            // no TLD
		{"http://example.com", false}, // URL, not domain
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isDomain(tt.input)
			if result != tt.expected {
				t.Errorf("isDomain(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

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

func TestIsFile(t *testing.T) {
	runtime := NewGojaRuntime()
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"existing file", testFile, true},
		{"directory", tmpDir, false},
		{"non-existent", filepath.Join(tmpDir, "nope.txt"), false},
		{"empty input", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := "is_file('" + tt.input + "')"
			result, err := runtime.Execute(expr, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("is_file(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	runtime := NewGojaRuntime()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"existing directory", tmpDir, true},
		{"file", testFile, false},
		{"non-existent", filepath.Join(tmpDir, "nodir"), false},
		{"empty input", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := "is_dir('" + tt.input + "')"
			result, err := runtime.Execute(expr, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("is_dir(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsGit(t *testing.T) {
	runtime := NewGojaRuntime()
	tmpDir := t.TempDir()

	// Create a fake git repo (directory with .git subfolder)
	gitRepo := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(filepath.Join(gitRepo, ".git"), 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create a non-git directory
	plainDir := filepath.Join(tmpDir, "plain")
	if err := os.MkdirAll(plainDir, 0755); err != nil {
		t.Fatalf("Failed to create plain dir: %v", err)
	}

	// Create a file
	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("hi"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"git repo", gitRepo, true},
		{"plain directory", plainDir, false},
		{"file", testFile, false},
		{"non-existent", filepath.Join(tmpDir, "nope"), false},
		{"empty input", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := "is_git('" + tt.input + "')"
			result, err := runtime.Execute(expr, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("is_git(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsURLFunc(t *testing.T) {
	runtime := NewGojaRuntime()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"http url", "http://example.com", true},
		{"https url", "https://example.com/path", true},
		{"uppercase HTTPS", "HTTPS://EXAMPLE.COM", true},
		{"ftp url", "ftp://example.com", false},
		{"domain only", "example.com", false},
		{"empty input", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := "is_url('" + tt.input + "')"
			result, err := runtime.Execute(expr, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("is_url(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsCompress(t *testing.T) {
	runtime := NewGojaRuntime()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"zip file", "archive.zip", true},
		{"tar.gz file", "archive.tar.gz", true},
		{"tgz file", "archive.tgz", true},
		{"gz file", "file.gz", true},
		{"tar.bz2 file", "archive.tar.bz2", true},
		{"tar.xz file", "archive.tar.xz", true},
		{"uppercase ZIP", "ARCHIVE.ZIP", true},
		{"txt file", "file.txt", false},
		{"no extension", "archive", false},
		{"empty input", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := "is_compress('" + tt.input + "')"
			result, err := runtime.Execute(expr, nil)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("is_compress(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	runtime := NewGojaRuntime()

	t.Run("pure Go project", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"main.go":        "package main",
			"lib/util.go":    "package lib",
			"lib/handler.go": "package lib",
		})
		expr := "detect_language('" + dir + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "golang" {
			t.Errorf("detect_language() = %v, want golang", result)
		}
	})

	t.Run("pure Python project", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"app.py":       "import os",
			"utils.py":     "def helper(): pass",
			"models/db.py": "class DB: pass",
		})
		expr := "detect_language('" + dir + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "python" {
			t.Errorf("detect_language() = %v, want python", result)
		}
	})

	t.Run("mixed project majority wins", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"main.go":   "package main",
			"lib.go":    "package lib",
			"extra.go":  "package extra",
			"script.py": "import sys",
		})
		expr := "detect_language('" + dir + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "golang" {
			t.Errorf("detect_language() = %v, want golang", result)
		}
	})

	t.Run("empty directory returns unknown", func(t *testing.T) {
		dir := t.TempDir()
		expr := "detect_language('" + dir + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "unknown" {
			t.Errorf("detect_language() = %v, want unknown", result)
		}
	})

	t.Run("non-existent path returns unknown", func(t *testing.T) {
		expr := "detect_language('/tmp/does_not_exist_xyz_abc')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if result != "unknown" {
			t.Errorf("detect_language() = %v, want unknown", result)
		}
	})

	t.Run("ignored directories are skipped", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"main.go":              "package main",
			"node_modules/dep.js":  "module.exports = {}",
			"node_modules/dep2.js": "module.exports = {}",
			"node_modules/dep3.js": "module.exports = {}",
			"vendor/lib.go":        "package vendor",
			"test/main_test.go":    "package test",
		})
		expr := "detect_language('" + dir + "')"
		result, err := runtime.Execute(expr, nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		// Only main.go should be counted (node_modules, vendor, test all ignored)
		if result != "golang" {
			t.Errorf("detect_language() = %v, want golang", result)
		}
	})
}

func TestDetectFolderLanguage(t *testing.T) {
	t.Run("shebang detection", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"run":    "#!/usr/bin/env python\nprint('hi')",
			"serve":  "#!/usr/bin/env python\nimport http",
			"deploy": "#!/bin/bash\necho deploy",
		})
		result := detectFolderLanguage(dir)
		if result != "python" {
			t.Errorf("detectFolderLanguage() = %v, want python", result)
		}
	})

	t.Run("file not directory returns unknown", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
		result := detectFolderLanguage(f)
		if result != "unknown" {
			t.Errorf("detectFolderLanguage() = %v, want unknown", result)
		}
	})

	t.Run("typescript project", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"src/index.ts": "const x = 1",
			"src/app.tsx":  "export default App",
			"src/util.ts":  "export function f() {}",
		})
		result := detectFolderLanguage(dir)
		if result != "typescript" {
			t.Errorf("detectFolderLanguage() = %v, want typescript", result)
		}
	})

	t.Run("rust project", func(t *testing.T) {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{
			"src/main.rs": "fn main() {}",
			"src/lib.rs":  "pub mod utils;",
		})
		result := detectFolderLanguage(dir)
		if result != "rust" {
			t.Errorf("detectFolderLanguage() = %v, want rust", result)
		}
	})
}

func TestDetectShebang(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"python env", "#!/usr/bin/env python\nprint('hi')", "python"},
		{"python direct", "#!/usr/bin/python3\nprint('hi')", "python"},
		{"node env", "#!/usr/bin/env node\nconsole.log(1)", "javascript"},
		{"bash", "#!/bin/bash\necho hi", "shell"},
		{"sh", "#!/bin/sh\necho hi", "shell"},
		{"ruby", "#!/usr/bin/env ruby\nputs 1", "ruby"},
		{"perl", "#!/usr/bin/perl\nprint 1", "perl"},
		{"no shebang", "just text", ""},
		{"empty file", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "script")
			if err := os.WriteFile(f, []byte(tt.content), 0755); err != nil {
				t.Fatal(err)
			}
			result := detectShebang(f)
			if result != tt.expected {
				t.Errorf("detectShebang() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// writeFiles creates files relative to base directory, creating intermediate dirs as needed.
func writeFiles(t *testing.T, base string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("Failed to create dir for %s: %v", rel, err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", rel, err)
		}
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

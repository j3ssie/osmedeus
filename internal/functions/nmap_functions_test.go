package functions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/json"
)

func TestParseNmapXML(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "test", "testdata", "sample-jsonl-output", "sample-nmap-result-1.xml")

	// Check if test file exists
	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testdataPath)
		return
	}

	results, err := parseNmapXML(testdataPath)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 host, got 0")
		return
	}

	// Verify first host
	host := results[0]
	if host.HostIP == "" {
		t.Error("Expected non-empty host IP")
	}

	if host.AssetType != "ip" {
		t.Errorf("Expected asset type 'ip', got '%s'", host.AssetType)
	}

	if len(host.Ports) == 0 {
		t.Error("Expected ports, got none")
	}

	// Verify JSONL serialization
	data, err := json.Marshal(host)
	if err != nil {
		t.Errorf("Failed to marshal to JSON: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	// Verify it's valid JSON
	var unmarshal AssetOutput
	if err := json.Unmarshal(data, &unmarshal); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}
}

func TestParseNmapGrepable(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "test", "testdata", "sample-jsonl-output", "sample-nmap-result-1.gnmap")

	// Check if test file exists
	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testdataPath)
		return
	}

	results, err := parseNmapGrepable(testdataPath)
	if err != nil {
		t.Fatalf("Failed to parse grepable: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 host, got 0")
		return
	}

	// Verify first host
	host := results[0]
	if host.HostIP == "" {
		t.Error("Expected non-empty host IP")
	}

	if host.AssetType != "ip" {
		t.Errorf("Expected asset type 'ip', got '%s'", host.AssetType)
	}

	if len(host.Ports) == 0 {
		t.Error("Expected ports, got none")
	}
}

func TestParseNmapXMLMultipleHosts(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "test", "testdata", "sample-jsonl-output", "sample-nmap-result-2.xml")

	// Check if test file exists
	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testdataPath)
		return
	}

	results, err := parseNmapXML(testdataPath)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 host, got 0")
	}

	// Verify all hosts have required fields
	for i, host := range results {
		if host.HostIP == "" {
			t.Errorf("Host %d: Expected non-empty host IP", i)
		}
		if host.AssetValue == "" {
			t.Errorf("Host %d: Expected non-empty asset value", i)
		}
		if host.AssetType != "ip" {
			t.Errorf("Host %d: Expected asset type 'ip', got '%s'", i, host.AssetType)
		}
	}
}

func TestNmapToJSONLIntegration(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "test", "testdata", "sample-jsonl-output", "sample-nmap-result-1.xml")

	// Check if test file exists
	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testdataPath)
		return
	}

	// Create temp output file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.jsonl")

	// Parse and write
	results, err := parseNmapXML(testdataPath)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Write JSONL
	outFile, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}

	for _, asset := range results {
		data, err := json.Marshal(asset)
		if err != nil {
			t.Fatalf("Failed to marshal asset: %v", err)
		}
		if _, err := outFile.Write(data); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}
		if _, err := outFile.WriteString("\n"); err != nil {
			t.Fatalf("Failed to write newline: %v", err)
		}
	}
	if err := outFile.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	// Verify output file exists and is not empty
	stat, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}

	if stat.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Read and verify JSONL format
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		t.Error("Expected at least one line in JSONL output")
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		if line == "" {
			continue
		}
		var asset AssetOutput
		if err := json.Unmarshal([]byte(line), &asset); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}

		// Verify required fields
		if asset.HostIP == "" {
			t.Errorf("Line %d: Missing host_ip", i+1)
		}
		if asset.AssetValue == "" {
			t.Errorf("Line %d: Missing asset_value", i+1)
		}
		if asset.AssetType == "" {
			t.Errorf("Line %d: Missing asset_type", i+1)
		}
	}
}

func TestParseNmapXMLEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.xml")

	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	_, err := parseNmapXML(emptyFile)
	if err == nil {
		t.Error("Expected error for empty file, got nil")
	}
}

func TestParseNmapXMLInvalidXML(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.xml")

	if err := os.WriteFile(invalidFile, []byte("not valid xml"), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	_, err := parseNmapXML(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid XML, got nil")
	}
}

func TestParseNmapGrepableInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.gnmap")

	if err := os.WriteFile(invalidFile, []byte("not a grepable format"), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	results, err := parseNmapGrepable(invalidFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should return empty results for invalid format
	if len(results) != 0 {
		t.Errorf("Expected 0 results for invalid format, got %d", len(results))
	}
}

func TestSanitizeTargetForPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.0/24", "192-168-1-0-24"},
		{"10.0.0.1", "10-0-0-1"},
		{"example.com", "example-com"},
		{"192.168.1.1:8080", "192-168-1-1-8080"},
		{"targets.txt", "targets"},
		{"scan.xml", "scan"},
		{"http://example.com", "http---example-com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeTargetForPath(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeTargetForPath(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

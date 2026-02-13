package functions

import (
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestMergeAssetFields_NewFillsGaps(t *testing.T) {
	existing := &database.Asset{
		ID:         42,
		Workspace:  "example.com",
		AssetValue: "sub.example.com",
		Title:      "Existing Title",
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	incoming := &database.Asset{
		Workspace:  "example.com",
		AssetValue: "sub.example.com",
		StatusCode: 200,
	}

	mergeAssetFields(existing, incoming)

	assert.Equal(t, int64(42), incoming.ID)
	assert.Equal(t, "Existing Title", incoming.Title)
	assert.Equal(t, 200, incoming.StatusCode)
	assert.Equal(t, "example.com", incoming.Workspace)
	assert.Equal(t, "sub.example.com", incoming.AssetValue)
}

func TestMergeAssetFields_NewWinsWhenNonEmpty(t *testing.T) {
	existing := &database.Asset{
		ID:         1,
		Workspace:  "example.com",
		AssetValue: "sub.example.com",
		Title:      "Old Title",
		StatusCode: 200,
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	incoming := &database.Asset{
		Workspace:  "example.com",
		AssetValue: "sub.example.com",
		Title:      "New Title",
		StatusCode: 301,
	}

	mergeAssetFields(existing, incoming)

	assert.Equal(t, "New Title", incoming.Title)
	assert.Equal(t, 301, incoming.StatusCode)
}

func TestMergeAssetFields_PreservesWhenNewEmpty(t *testing.T) {
	existing := &database.Asset{
		ID:            1,
		Workspace:     "example.com",
		AssetValue:    "sub.example.com",
		StatusCode:    200,
		ContentLength: 5000,
		HostIP:        "1.2.3.4",
		CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	incoming := &database.Asset{
		Workspace:  "example.com",
		AssetValue: "sub.example.com",
		// StatusCode, ContentLength, HostIP are zero/empty
	}

	mergeAssetFields(existing, incoming)

	assert.Equal(t, 200, incoming.StatusCode)
	assert.Equal(t, int64(5000), incoming.ContentLength)
	assert.Equal(t, "1.2.3.4", incoming.HostIP)
}

func TestMergeAssetFields_SlicePreserved(t *testing.T) {
	existing := &database.Asset{
		ID:           1,
		Workspace:    "example.com",
		AssetValue:   "sub.example.com",
		Technologies: []string{"nginx", "php"},
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	incoming := &database.Asset{
		Workspace:    "example.com",
		AssetValue:   "sub.example.com",
		Technologies: nil,
	}

	mergeAssetFields(existing, incoming)

	assert.Equal(t, []string{"nginx", "php"}, incoming.Technologies)
}

func TestMergeAssetFields_SliceOverwritten(t *testing.T) {
	existing := &database.Asset{
		ID:           1,
		Workspace:    "example.com",
		AssetValue:   "sub.example.com",
		Technologies: []string{"nginx"},
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	incoming := &database.Asset{
		Workspace:    "example.com",
		AssetValue:   "sub.example.com",
		Technologies: []string{"apache", "java"},
	}

	mergeAssetFields(existing, incoming)

	assert.Equal(t, []string{"apache", "java"}, incoming.Technologies)
}

func TestMergeAssetFields_Identity(t *testing.T) {
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := &database.Asset{
		ID:         99,
		Workspace:  "ws1",
		AssetValue: "host.example.com",
		CreatedAt:  created,
	}
	incoming := &database.Asset{
		Workspace:  "ws2",        // should be overwritten
		AssetValue: "other.host", // should be overwritten
	}

	mergeAssetFields(existing, incoming)

	assert.Equal(t, int64(99), incoming.ID)
	assert.Equal(t, created, incoming.CreatedAt)
	assert.Equal(t, "ws1", incoming.Workspace)
	assert.Equal(t, "host.example.com", incoming.AssetValue)
}

func TestMergeVulnFields_Basic(t *testing.T) {
	existing := &database.Vulnerability{
		ID:         10,
		Workspace:  "example.com",
		VulnInfo:   "CVE-2024-0001",
		VulnTitle:  "Existing Title",
		Severity:   "high",
		Confidence: "Firm",
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	incoming := &database.Vulnerability{
		Workspace:  "example.com",
		VulnInfo:   "CVE-2024-0001",
		VulnTitle:  "Updated Title",
		VulnDesc:   "New description",
		Severity:   "", // empty - should preserve existing
		Confidence: "", // empty - should preserve existing
	}

	mergeVulnFields(existing, incoming)

	assert.Equal(t, int64(10), incoming.ID)
	assert.Equal(t, "example.com", incoming.Workspace)
	assert.Equal(t, "Updated Title", incoming.VulnTitle)
	assert.Equal(t, "New description", incoming.VulnDesc)
	assert.Equal(t, "high", incoming.Severity)
	assert.Equal(t, "Firm", incoming.Confidence)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), incoming.CreatedAt)
}

func TestMergeVulnFields_TagsPreserved(t *testing.T) {
	existing := &database.Vulnerability{
		ID:        1,
		Workspace: "example.com",
		Tags:      []string{"cve", "rce"},
	}
	incoming := &database.Vulnerability{
		Workspace: "example.com",
		Tags:      nil,
	}

	mergeVulnFields(existing, incoming)

	assert.Equal(t, []string{"cve", "rce"}, incoming.Tags)
}

func TestMergePrimitives(t *testing.T) {
	// mergeString
	assert.Equal(t, "new", mergeString("old", "new"))
	assert.Equal(t, "old", mergeString("old", ""))
	assert.Equal(t, "new", mergeString("", "new"))
	assert.Equal(t, "", mergeString("", ""))

	// mergeInt
	assert.Equal(t, 42, mergeInt(10, 42))
	assert.Equal(t, 10, mergeInt(10, 0))
	assert.Equal(t, 42, mergeInt(0, 42))
	assert.Equal(t, 0, mergeInt(0, 0))

	// mergeInt64
	assert.Equal(t, int64(42), mergeInt64(10, 42))
	assert.Equal(t, int64(10), mergeInt64(10, 0))
	assert.Equal(t, int64(42), mergeInt64(0, 42))
	assert.Equal(t, int64(0), mergeInt64(0, 0))

	// mergeStringSlice
	assert.Equal(t, []string{"b"}, mergeStringSlice([]string{"a"}, []string{"b"}))
	assert.Equal(t, []string{"a"}, mergeStringSlice([]string{"a"}, nil))
	assert.Equal(t, []string{"a"}, mergeStringSlice([]string{"a"}, []string{}))
	assert.Equal(t, []string{"b"}, mergeStringSlice(nil, []string{"b"}))
	assert.Nil(t, mergeStringSlice(nil, nil))
}

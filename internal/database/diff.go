package database

import (
	"context"
	"fmt"
	"time"
)

// AssetDiff represents the difference between two scans
type AssetDiff struct {
	WorkspaceName string        `json:"workspace_name"`
	FromTime      time.Time     `json:"from_time"`
	ToTime        time.Time     `json:"to_time"`
	Added         []Asset       `json:"added"`
	Removed       []Asset       `json:"removed"`
	Changed       []AssetChange `json:"changed"`
	Summary       DiffSummary   `json:"summary"`
}

// AssetChange represents a changed asset with field-level diffs
type AssetChange struct {
	AssetID    int64         `json:"asset_id"`
	AssetValue string        `json:"asset_value"`
	URL        string        `json:"url"`
	Changes    []FieldChange `json:"changes"`
}

// FieldChange represents a single field that changed
type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

// DiffSummary provides counts
type DiffSummary struct {
	TotalAdded   int `json:"total_added"`
	TotalRemoved int `json:"total_removed"`
	TotalChanged int `json:"total_changed"`
}

// VulnerabilityDiff represents the difference between two scans for vulnerabilities
type VulnerabilityDiff struct {
	WorkspaceName string              `json:"workspace_name"`
	FromTime      time.Time           `json:"from_time"`
	ToTime        time.Time           `json:"to_time"`
	Added         []Vulnerability     `json:"added"`
	Removed       []Vulnerability     `json:"removed"`
	Changed       []VulnerabilityChange `json:"changed"`
	Summary       DiffSummary         `json:"summary"`
}

// VulnerabilityChange represents a changed vulnerability
type VulnerabilityChange struct {
	VulnID     int64         `json:"vuln_id"`
	VulnInfo   string        `json:"vuln_info"`
	AssetValue string        `json:"asset_value"`
	Changes    []FieldChange `json:"changes"`
}

// GetAssetDiff calculates the difference between two time points
func GetAssetDiff(ctx context.Context, workspace string, fromTime, toTime time.Time) (*AssetDiff, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	diff := &AssetDiff{
		WorkspaceName: workspace,
		FromTime:      fromTime,
		ToTime:        toTime,
		Added:         []Asset{},
		Removed:       []Asset{},
		Changed:       []AssetChange{},
	}

	// Added: created_at >= fromTime AND created_at <= toTime
	err := db.NewSelect().Model(&diff.Added).
		Where("workspace = ?", workspace).
		Where("created_at >= ?", fromTime).
		Where("created_at <= ?", toTime).
		Scan(ctx)
	if err != nil {
		// Not a fatal error, continue with empty slice
		diff.Added = []Asset{}
	}

	// Removed: last_seen_at < fromTime AND last_seen_at is not zero
	// These are assets that haven't been seen since before the time window
	err = db.NewSelect().Model(&diff.Removed).
		Where("workspace = ?", workspace).
		Where("last_seen_at < ?", fromTime).
		Where("last_seen_at != ?", time.Time{}).
		Where("created_at < ?", fromTime).
		Scan(ctx)
	if err != nil {
		diff.Removed = []Asset{}
	}

	// Changed: updated_at within range AND created_at before range
	var changedAssets []Asset
	err = db.NewSelect().Model(&changedAssets).
		Where("workspace = ?", workspace).
		Where("updated_at >= ?", fromTime).
		Where("updated_at <= ?", toTime).
		Where("created_at < ?", fromTime).
		Scan(ctx)
	if err != nil {
		changedAssets = []Asset{}
	}

	// Convert to AssetChange (field-level diffs would require storing previous values)
	for _, a := range changedAssets {
		diff.Changed = append(diff.Changed, AssetChange{
			AssetID:    a.ID,
			AssetValue: a.AssetValue,
			URL:        a.URL,
			// Note: Without historical storage, we can only indicate it changed
			Changes: []FieldChange{},
		})
	}

	diff.Summary = DiffSummary{
		TotalAdded:   len(diff.Added),
		TotalRemoved: len(diff.Removed),
		TotalChanged: len(diff.Changed),
	}

	return diff, nil
}

// GetVulnerabilityDiff calculates the difference between two time points for vulnerabilities
func GetVulnerabilityDiff(ctx context.Context, workspace string, fromTime, toTime time.Time) (*VulnerabilityDiff, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	diff := &VulnerabilityDiff{
		WorkspaceName: workspace,
		FromTime:      fromTime,
		ToTime:        toTime,
		Added:         []Vulnerability{},
		Removed:       []Vulnerability{},
		Changed:       []VulnerabilityChange{},
	}

	// Added: created_at >= fromTime AND created_at <= toTime
	err := db.NewSelect().Model(&diff.Added).
		Where("workspace = ?", workspace).
		Where("created_at >= ?", fromTime).
		Where("created_at <= ?", toTime).
		Scan(ctx)
	if err != nil {
		diff.Added = []Vulnerability{}
	}

	// Removed: last_seen_at < fromTime AND last_seen_at is not zero
	err = db.NewSelect().Model(&diff.Removed).
		Where("workspace = ?", workspace).
		Where("last_seen_at < ?", fromTime).
		Where("last_seen_at != ?", time.Time{}).
		Where("created_at < ?", fromTime).
		Scan(ctx)
	if err != nil {
		diff.Removed = []Vulnerability{}
	}

	// Changed: updated_at within range AND created_at before range
	var changedVulns []Vulnerability
	err = db.NewSelect().Model(&changedVulns).
		Where("workspace = ?", workspace).
		Where("updated_at >= ?", fromTime).
		Where("updated_at <= ?", toTime).
		Where("created_at < ?", fromTime).
		Scan(ctx)
	if err != nil {
		changedVulns = []Vulnerability{}
	}

	// Convert to VulnerabilityChange
	for _, v := range changedVulns {
		diff.Changed = append(diff.Changed, VulnerabilityChange{
			VulnID:     v.ID,
			VulnInfo:   v.VulnInfo,
			AssetValue: v.AssetValue,
			Changes:    []FieldChange{},
		})
	}

	diff.Summary = DiffSummary{
		TotalAdded:   len(diff.Added),
		TotalRemoved: len(diff.Removed),
		TotalChanged: len(diff.Changed),
	}

	return diff, nil
}

// ImportStats tracks statistics from import operations
type ImportStats struct {
	New       int `json:"new"`
	Updated   int `json:"updated"`
	Unchanged int `json:"unchanged"`
	Errors    int `json:"errors"`
}

package database

import (
	"context"
	"fmt"
	"sort"

	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/uptrace/bun"
)

// SystemStats contains aggregated system statistics
type SystemStats struct {
	Workflows       WorkflowStats      `json:"workflows"`
	Runs            RunStats           `json:"runs"`
	Workspaces      WorkspaceStats     `json:"workspaces"`
	Assets          AssetStats         `json:"assets"`
	Vulnerabilities VulnerabilityStats `json:"vulnerabilities"`
	Schedules       ScheduleStats      `json:"schedules"`
}

// WorkflowStats contains workflow counts
type WorkflowStats struct {
	Total   int `json:"total"`
	Flows   int `json:"flows"`
	Modules int `json:"modules"`
}

// RunStats contains run counts by status
type RunStats struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Running   int `json:"running"`
	Failed    int `json:"failed"`
	Pending   int `json:"pending"`
}

// WorkspaceStats contains workspace counts
type WorkspaceStats struct {
	Total int `json:"total"`
}

// AssetStats contains asset counts
type AssetStats struct {
	Total int `json:"total"`
}

// VulnerabilityStats contains vulnerability counts by severity
type VulnerabilityStats struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// ScheduleStats contains schedule counts
type ScheduleStats struct {
	Total   int `json:"total"`
	Enabled int `json:"enabled"`
}

// GetSystemStats retrieves aggregated system statistics from the database and workflows
func GetSystemStats(ctx context.Context, workflowsPath string) (*SystemStats, error) {
	db := GetDB()
	stats := &SystemStats{}

	// Get workflow stats from loader
	if workflowsPath != "" {
		loader := parser.NewLoader(workflowsPath)
		flows, modules, err := loader.ListAllWorkflows()
		if err == nil {
			stats.Workflows = WorkflowStats{
				Total:   len(flows) + len(modules),
				Flows:   len(flows),
				Modules: len(modules),
			}
		}
	}

	// Get run stats
	runStats, err := getRunStats(ctx, db)
	if err == nil {
		stats.Runs = runStats
	}

	// Get workspace stats
	workspaceCount, err := db.NewSelect().Model((*Workspace)(nil)).Count(ctx)
	if err == nil {
		stats.Workspaces = WorkspaceStats{Total: workspaceCount}
	}

	// Get asset stats
	assetCount, err := db.NewSelect().Model((*Asset)(nil)).Count(ctx)
	if err == nil {
		stats.Assets = AssetStats{Total: assetCount}
	}

	// Get vulnerability stats (aggregated from workspaces)
	vulnStats, err := getVulnerabilityStats(ctx, db)
	if err == nil {
		stats.Vulnerabilities = vulnStats
	}

	// Get schedule stats
	scheduleStats, err := getScheduleStats(ctx, db)
	if err == nil {
		stats.Schedules = scheduleStats
	}

	return stats, nil
}

// getRunStats retrieves run counts grouped by status in a single query
func getRunStats(ctx context.Context, db *bun.DB) (RunStats, error) {
	var result struct {
		Total     int `bun:"total"`
		Completed int `bun:"completed"`
		Running   int `bun:"running"`
		Failed    int `bun:"failed"`
		Pending   int `bun:"pending"`
	}

	err := db.NewSelect().
		Model((*Run)(nil)).
		ColumnExpr("COUNT(*) AS total").
		ColumnExpr("SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed").
		ColumnExpr("SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) AS running").
		ColumnExpr("SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failed").
		ColumnExpr("SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending").
		Scan(ctx, &result)

	if err != nil {
		return RunStats{}, err
	}

	return RunStats{
		Total:     result.Total,
		Completed: result.Completed,
		Running:   result.Running,
		Failed:    result.Failed,
		Pending:   result.Pending,
	}, nil
}

// getVulnerabilityStats retrieves aggregated vulnerability counts from workspaces
func getVulnerabilityStats(ctx context.Context, db *bun.DB) (VulnerabilityStats, error) {
	stats := VulnerabilityStats{}

	var result struct {
		Critical int `bun:"critical"`
		High     int `bun:"high"`
		Medium   int `bun:"medium"`
		Low      int `bun:"low"`
	}

	err := db.NewSelect().
		Model((*Workspace)(nil)).
		ColumnExpr("COALESCE(SUM(vuln_critical), 0) AS critical").
		ColumnExpr("COALESCE(SUM(vuln_high), 0) AS high").
		ColumnExpr("COALESCE(SUM(vuln_medium), 0) AS medium").
		ColumnExpr("COALESCE(SUM(vuln_low), 0) AS low").
		Scan(ctx, &result)

	if err != nil {
		return stats, err
	}

	stats.Critical = result.Critical
	stats.High = result.High
	stats.Medium = result.Medium
	stats.Low = result.Low
	stats.Total = result.Critical + result.High + result.Medium + result.Low

	return stats, nil
}

// getScheduleStats retrieves schedule counts in a single query
func getScheduleStats(ctx context.Context, db *bun.DB) (ScheduleStats, error) {
	var result struct {
		Total   int `bun:"total"`
		Enabled int `bun:"enabled"`
	}

	err := db.NewSelect().
		Model((*Schedule)(nil)).
		ColumnExpr("COUNT(*) AS total").
		ColumnExpr("SUM(CASE WHEN is_enabled = true THEN 1 ELSE 0 END) AS enabled").
		Scan(ctx, &result)

	if err != nil {
		return ScheduleStats{}, err
	}

	return ScheduleStats{
		Total:   result.Total,
		Enabled: result.Enabled,
	}, nil
}

// AssetStatsData contains unique lists of asset metadata
type AssetStatsData struct {
	Technologies []string `json:"technologies"`
	Sources      []string `json:"sources"`
	Remarks      []string `json:"remarks"`
	AssetTypes   []string `json:"asset_types"`
}

// GetAssetStats retrieves unique values for technologies, sources, remarks, and asset_types
// with optional workspace filtering
func GetAssetStats(ctx context.Context, workspace string) (*AssetStatsData, error) {
	db := GetDB()

	// Fetch all assets (optionally filtered by workspace)
	var assets []Asset
	query := db.NewSelect().
		Model(&assets).
		Column("tech", "source", "remarks", "asset_type")

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assets: %w", err)
	}

	// Deduplicate in Go (simpler than DB-specific JSON functions)
	techSet := make(map[string]bool)
	sourceSet := make(map[string]bool)
	remarkSet := make(map[string]bool)
	assetTypeSet := make(map[string]bool)

	for _, asset := range assets {
		// Technologies (JSON array)
		for _, tech := range asset.Technologies {
			if tech != "" {
				techSet[tech] = true
			}
		}

		// Source (string)
		if asset.Source != "" {
			sourceSet[asset.Source] = true
		}

		// Remarks (JSON array)
		for _, remark := range asset.Remarks {
			if remark != "" {
				remarkSet[remark] = true
			}
		}

		// AssetType (string)
		if asset.AssetType != "" {
			assetTypeSet[asset.AssetType] = true
		}
	}

	// Convert maps to sorted slices
	result := &AssetStatsData{
		Technologies: mapKeysToSortedSlice(techSet),
		Sources:      mapKeysToSortedSlice(sourceSet),
		Remarks:      mapKeysToSortedSlice(remarkSet),
		AssetTypes:   mapKeysToSortedSlice(assetTypeSet),
	}

	return result, nil
}

// mapKeysToSortedSlice converts map keys to sorted string slice
func mapKeysToSortedSlice(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

package repository

import (
	"context"
	"fmt"
	"io"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/uptrace/bun"
)

// AssetRepository handles asset database operations
type AssetRepository struct {
	db *bun.DB
}

// NewAssetRepository creates a new asset repository
func NewAssetRepository(db *bun.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

// AssetQuery represents query parameters for asset search
type AssetQuery struct {
	Workspace   string
	AssetValue  string
	HostIP      string
	StatusCode  int
	ContentType string
	Tech        string
	Source      string
	Page        int
	PerPage     int
}

// Create creates a new asset
func (r *AssetRepository) Create(ctx context.Context, asset *database.Asset) error {
	_, err := r.db.NewInsert().Model(asset).Exec(ctx)
	return err
}

// GetByID retrieves an asset by ID
func (r *AssetRepository) GetByID(ctx context.Context, id int64) (*database.Asset, error) {
	asset := new(database.Asset)
	err := r.db.NewSelect().
		Model(asset).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return asset, nil
}

// Update updates an existing asset
func (r *AssetRepository) Update(ctx context.Context, asset *database.Asset) error {
	_, err := r.db.NewUpdate().
		Model(asset).
		WherePK().
		Exec(ctx)
	return err
}

// Delete deletes an asset by ID
func (r *AssetRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*database.Asset)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// DeleteByWorkspace deletes all assets in a workspace
func (r *AssetRepository) DeleteByWorkspace(ctx context.Context, workspace string) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*database.Asset)(nil)).
		Where("workspace = ?", workspace).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ListByWorkspace lists assets for a workspace with pagination
func (r *AssetRepository) ListByWorkspace(ctx context.Context, workspace string, page, perPage int) ([]*database.Asset, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	var assets []*database.Asset
	count, err := r.db.NewSelect().
		Model(&assets).
		Where("workspace = ?", workspace).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		ScanAndCount(ctx)

	return assets, count, err
}

// ListByAssetValue lists assets for a specific asset value
func (r *AssetRepository) ListByAssetValue(ctx context.Context, assetValue string) ([]*database.Asset, error) {
	var assets []*database.Asset
	err := r.db.NewSelect().
		Model(&assets).
		Where("asset_value = ?", assetValue).
		Order("created_at DESC").
		Scan(ctx)
	return assets, err
}

// ListByHostIP lists assets for a specific IP
func (r *AssetRepository) ListByHostIP(ctx context.Context, hostIP string) ([]*database.Asset, error) {
	var assets []*database.Asset
	err := r.db.NewSelect().
		Model(&assets).
		Where("host_ip = ?", hostIP).
		Order("created_at DESC").
		Scan(ctx)
	return assets, err
}

// ListByStatus lists assets with a specific status code
func (r *AssetRepository) ListByStatus(ctx context.Context, workspace string, statusCode int) ([]*database.Asset, error) {
	var assets []*database.Asset
	query := r.db.NewSelect().Model(&assets)

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}

	err := query.
		Where("status_code = ?", statusCode).
		Order("created_at DESC").
		Scan(ctx)
	return assets, err
}

// Search searches assets with multiple criteria
func (r *AssetRepository) Search(ctx context.Context, query AssetQuery) ([]*database.Asset, int, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 50
	}
	offset := (query.Page - 1) * query.PerPage

	var assets []*database.Asset
	q := r.db.NewSelect().Model(&assets)

	if query.Workspace != "" {
		q = q.Where("workspace = ?", query.Workspace)
	}
	if query.AssetValue != "" {
		q = q.Where("asset_value LIKE ?", "%"+query.AssetValue+"%")
	}
	if query.HostIP != "" {
		q = q.Where("host_ip = ?", query.HostIP)
	}
	if query.StatusCode > 0 {
		q = q.Where("status_code = ?", query.StatusCode)
	}
	if query.ContentType != "" {
		q = q.Where("content_type LIKE ?", "%"+query.ContentType+"%")
	}
	if query.Source != "" {
		q = q.Where("source = ?", query.Source)
	}
	// Note: Tech search would need JSON-specific query depending on database

	count, err := q.
		Order("created_at DESC").
		Limit(query.PerPage).
		Offset(offset).
		ScanAndCount(ctx)

	return assets, count, err
}

// ImportFromJSONL imports assets from a JSONL file
func (r *AssetRepository) ImportFromJSONL(ctx context.Context, filePath, workspace, source string) (*database.ImportResult, error) {
	importer := database.NewJSONLImporter(r.db)
	return importer.ImportAssets(ctx, filePath, workspace, source)
}

// ImportFromReader imports assets from an io.Reader
func (r *AssetRepository) ImportFromReader(ctx context.Context, reader io.Reader, workspace, source string) (*database.ImportResult, error) {
	importer := database.NewJSONLImporter(r.db)
	return importer.ImportAssetsFromReader(ctx, reader, workspace, source)
}

// CountByWorkspace returns the count of assets in a workspace
func (r *AssetRepository) CountByWorkspace(ctx context.Context, workspace string) (int, error) {
	return r.db.NewSelect().
		Model((*database.Asset)(nil)).
		Where("workspace = ?", workspace).
		Count(ctx)
}

// GetTechSummary returns a summary of technologies found in a workspace
func (r *AssetRepository) GetTechSummary(ctx context.Context, workspace string) (map[string]int, error) {
	// This implementation varies by database
	// For SQLite/PostgreSQL with JSON support, we need to unnest the array
	var results []struct {
		Tech  string `bun:"tech"`
		Count int    `bun:"count"`
	}

	// Simple approach: fetch all and count in Go
	// For production, use database-specific JSON functions
	var assets []*database.Asset
	err := r.db.NewSelect().
		Model(&assets).
		Column("tech").
		Where("workspace = ?", workspace).
		Where("tech IS NOT NULL").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	techCount := make(map[string]int)
	for _, asset := range assets {
		for _, tech := range asset.Technologies {
			techCount[tech]++
		}
	}

	_ = results // unused in simple implementation
	return techCount, nil
}

// GetStatusSummary returns a summary of status codes in a workspace
func (r *AssetRepository) GetStatusSummary(ctx context.Context, workspace string) (map[int]int, error) {
	var results []struct {
		StatusCode int `bun:"status_code"`
		Count      int `bun:"count"`
	}

	err := r.db.NewSelect().
		Model((*database.Asset)(nil)).
		ColumnExpr("status_code, COUNT(*) AS count").
		Where("workspace = ?", workspace).
		Where("status_code > 0").
		Group("status_code").
		Order("count DESC").
		Scan(ctx, &results)

	if err != nil {
		return nil, err
	}

	summary := make(map[int]int)
	for _, r := range results {
		summary[r.StatusCode] = r.Count
	}

	return summary, nil
}

// GetAssetValueSummary returns a summary of unique asset values in a workspace
func (r *AssetRepository) GetAssetValueSummary(ctx context.Context, workspace string) (int, error) {
	var count int
	err := r.db.NewSelect().
		Model((*database.Asset)(nil)).
		ColumnExpr("COUNT(DISTINCT asset_value)").
		Where("workspace = ?", workspace).
		Scan(ctx, &count)

	return count, err
}

// Upsert creates or updates an asset based on workspace, asset_value, url
func (r *AssetRepository) Upsert(ctx context.Context, asset *database.Asset) error {
	_, err := r.db.NewInsert().
		Model(asset).
		On("CONFLICT (workspace, asset_value, url) DO UPDATE").
		Set("status_code = EXCLUDED.status_code").
		Set("title = EXCLUDED.title").
		Set("tech = EXCLUDED.tech").
		Set("content_type = EXCLUDED.content_type").
		Set("content_length = EXCLUDED.content_length").
		Set("host_ip = EXCLUDED.host_ip").
		Set("a_records = EXCLUDED.a_records").
		Set("tls = EXCLUDED.tls").
		Set("response_time = EXCLUDED.response_time").
		Set("words = EXCLUDED.words").
		Set("lines = EXCLUDED.lines").
		Set("remarks = EXCLUDED.remarks").
		Set("raw_data = EXCLUDED.raw_data").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

// BulkUpsert performs bulk upsert of assets
func (r *AssetRepository) BulkUpsert(ctx context.Context, assets []*database.Asset) error {
	if len(assets) == 0 {
		return nil
	}

	_, err := r.db.NewInsert().
		Model(&assets).
		On("CONFLICT (workspace, asset_value, url) DO UPDATE").
		Set("status_code = EXCLUDED.status_code").
		Set("title = EXCLUDED.title").
		Set("tech = EXCLUDED.tech").
		Set("content_type = EXCLUDED.content_type").
		Set("content_length = EXCLUDED.content_length").
		Set("host_ip = EXCLUDED.host_ip").
		Set("a_records = EXCLUDED.a_records").
		Set("tls = EXCLUDED.tls").
		Set("response_time = EXCLUDED.response_time").
		Set("words = EXCLUDED.words").
		Set("lines = EXCLUDED.lines").
		Set("remarks = EXCLUDED.remarks").
		Set("raw_data = EXCLUDED.raw_data").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

// ExportToJSONL exports assets to a JSONL writer
func (r *AssetRepository) ExportToJSONL(ctx context.Context, workspace string, w io.Writer) (int, error) {
	var assets []*database.Asset
	err := r.db.NewSelect().
		Model(&assets).
		Where("workspace = ?", workspace).
		Order("created_at ASC").
		Scan(ctx)

	if err != nil {
		return 0, err
	}

	count := 0
	for _, asset := range assets {
		if asset.RawJsonData != "" {
			_, err := fmt.Fprintf(w, "%s\n", asset.RawJsonData)
			if err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}

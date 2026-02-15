package database

import (
	"bufio"
	"context"
	"fmt"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

// JSONLImporter handles batch import from JSONL files
type JSONLImporter struct {
	db        *bun.DB
	batchSize int
}

// NewJSONLImporter creates a new JSONL importer
func NewJSONLImporter(db *bun.DB) *JSONLImporter {
	return &JSONLImporter{
		db:        db,
		batchSize: 100,
	}
}

// WithBatchSize sets the batch size for imports
func (i *JSONLImporter) WithBatchSize(size int) *JSONLImporter {
	if size > 0 {
		i.batchSize = size
	}
	return i
}

// ImportResult holds import statistics
type ImportResult struct {
	Total    int           `json:"total"`
	Imported int           `json:"imported"`
	Updated  int           `json:"updated"`
	Failed   int           `json:"failed"`
	Errors   []ImportError `json:"errors,omitempty"`
	Duration time.Duration `json:"duration"`
}

// ImportError represents a single import error
type ImportError struct {
	Line  int    `json:"line"`
	Error string `json:"error"`
	Data  string `json:"data,omitempty"`
}

// ImportAssets imports assets from a JSONL file
func (i *JSONLImporter) ImportAssets(ctx context.Context, filePath, workspace, source string) (*ImportResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return i.ImportAssetsFromReader(ctx, file, workspace, source)
}

// ImportAssetsFromReader imports assets from an io.Reader
func (i *JSONLImporter) ImportAssetsFromReader(ctx context.Context, r io.Reader, workspace, source string) (*ImportResult, error) {
	startTime := time.Now()
	scanner := bufio.NewScanner(r)
	// Allow large lines (up to 10MB)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	result := &ImportResult{}
	batch := make([]*Asset, 0, i.batchSize)

	for scanner.Scan() {
		result.Total++
		line := scanner.Bytes()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		asset, err := ParseAssetLine(line, workspace, source)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Line:  result.Total,
				Error: err.Error(),
				Data:  truncateString(string(line), 200),
			})
			continue
		}

		batch = append(batch, asset)

		if len(batch) >= i.batchSize {
			imported, err := i.insertAssetBatch(ctx, batch)
			if err != nil {
				return result, fmt.Errorf("batch insert failed at line %d: %w", result.Total, err)
			}
			result.Imported += imported
			batch = batch[:0]
		}
	}

	// Insert remaining batch
	if len(batch) > 0 {
		imported, err := i.insertAssetBatch(ctx, batch)
		if err != nil {
			return result, fmt.Errorf("final batch insert failed: %w", err)
		}
		result.Imported += imported
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("scanner error: %w", err)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// insertAssetBatch inserts a batch of assets with upsert
func (i *JSONLImporter) insertAssetBatch(ctx context.Context, assets []*Asset) (int, error) {
	if len(assets) == 0 {
		return 0, nil
	}

	// Use ON CONFLICT for upsert with merge semantics (non-empty incoming wins)
	res, err := i.db.NewInsert().
		Model(&assets).
		On("CONFLICT (workspace, asset_value, url) DO UPDATE").
		Set("status_code = CASE WHEN EXCLUDED.status_code != 0 THEN EXCLUDED.status_code ELSE assets.status_code END").
		Set("title = CASE WHEN EXCLUDED.title != '' THEN EXCLUDED.title ELSE assets.title END").
		Set("tech = CASE WHEN EXCLUDED.tech IS NOT NULL AND EXCLUDED.tech != '[]' AND EXCLUDED.tech != 'null' THEN EXCLUDED.tech ELSE assets.tech END").
		Set("content_type = CASE WHEN EXCLUDED.content_type != '' THEN EXCLUDED.content_type ELSE assets.content_type END").
		Set("content_length = CASE WHEN EXCLUDED.content_length != 0 THEN EXCLUDED.content_length ELSE assets.content_length END").
		Set("host_ip = CASE WHEN EXCLUDED.host_ip != '' THEN EXCLUDED.host_ip ELSE assets.host_ip END").
		Set("a_records = CASE WHEN EXCLUDED.a_records IS NOT NULL AND EXCLUDED.a_records != '[]' AND EXCLUDED.a_records != 'null' THEN EXCLUDED.a_records ELSE assets.a_records END").
		Set("tls = CASE WHEN EXCLUDED.tls != '' THEN EXCLUDED.tls ELSE assets.tls END").
		Set("response_time = CASE WHEN EXCLUDED.response_time != '' THEN EXCLUDED.response_time ELSE assets.response_time END").
		Set("words = CASE WHEN EXCLUDED.words != 0 THEN EXCLUDED.words ELSE assets.words END").
		Set("lines = CASE WHEN EXCLUDED.lines != 0 THEN EXCLUDED.lines ELSE assets.lines END").
		Set("remarks = CASE WHEN EXCLUDED.remarks IS NOT NULL AND EXCLUDED.remarks != '[]' AND EXCLUDED.remarks != 'null' THEN EXCLUDED.remarks ELSE assets.remarks END").
		Set("language = CASE WHEN EXCLUDED.language != '' THEN EXCLUDED.language ELSE assets.language END").
		Set("size = CASE WHEN EXCLUDED.size != 0 THEN EXCLUDED.size ELSE assets.size END").
		Set("loc = CASE WHEN EXCLUDED.loc != 0 THEN EXCLUDED.loc ELSE assets.loc END").
		Set("blob_content = CASE WHEN EXCLUDED.blob_content != '' THEN EXCLUDED.blob_content ELSE assets.blob_content END").
		Set("raw_data = CASE WHEN EXCLUDED.raw_data != '' THEN EXCLUDED.raw_data ELSE assets.raw_data END").
		Set("asset_type = CASE WHEN EXCLUDED.asset_type != '' THEN EXCLUDED.asset_type ELSE assets.asset_type END").
		Set("is_cdn = CASE WHEN EXCLUDED.is_cdn OR assets.is_cdn THEN TRUE ELSE FALSE END").
		Set("is_cloud = CASE WHEN EXCLUDED.is_cloud OR assets.is_cloud THEN TRUE ELSE FALSE END").
		Set("is_waf = CASE WHEN EXCLUDED.is_waf OR assets.is_waf THEN TRUE ELSE FALSE END").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)

	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return int(rowsAffected), nil
}

// cloudProviderPatterns are substrings used to detect cloud-hosted CDNs.
var cloudProviderPatterns = []string{
	"aws", "amazon", "cloudfront",
	"google", "gcp", "cloud cdn",
	"azure", "microsoft",
	"akamai", "fastly",
	"oracle",
	"alibaba", "aliyun",
	"tencent",
	"digitalocean",
}

// DetectCDNFlags derives is_cdn, is_cloud, and is_waf booleans from raw httpx JSON data.
//
// Rules:
//   - is_cdn  = "cdn" is true OR "cdn_name" is non-empty
//   - is_cloud = cdn_name matches a cloud provider pattern
//   - is_waf  = "cdn_type" equals "waf"
func DetectCDNFlags(raw map[string]interface{}) (isCDN, isCloud, isWAF bool) {
	// Check "cdn" boolean field
	if v, ok := raw["cdn"].(bool); ok && v {
		isCDN = true
	}

	// Check "cdn_name" string field
	cdnName := ""
	if v, ok := raw["cdn_name"].(string); ok && v != "" {
		cdnName = v
		isCDN = true
	}

	// Cloud detection from cdn_name
	if cdnName != "" {
		lower := strings.ToLower(cdnName)
		for _, pattern := range cloudProviderPatterns {
			if strings.Contains(lower, pattern) {
				isCloud = true
				break
			}
		}
	}

	// WAF detection from cdn_type
	if v, ok := raw["cdn_type"].(string); ok && strings.EqualFold(v, "waf") {
		isWAF = true
	}

	return
}

// ParseAssetLine parses a single JSONL line into an Asset
func ParseAssetLine(line []byte, defaultWorkspace, source string) (*Asset, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	now := time.Now()
	asset := &Asset{
		Workspace:   defaultWorkspace,
		Source:      source,
		RawJsonData: string(line),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Map JSON fields to Asset struct
	// Required fields
	if v, ok := raw["workspace"].(string); ok && v != "" {
		asset.Workspace = v
	}
	if v, ok := raw["asset_value"].(string); ok {
		asset.AssetValue = v
	}

	// HTTP data
	if v, ok := raw["url"].(string); ok {
		asset.URL = v
	}
	if v, ok := raw["input"].(string); ok {
		asset.Input = v
	}
	if v, ok := raw["scheme"].(string); ok {
		asset.Scheme = v
	}
	if v, ok := raw["method"].(string); ok {
		asset.Method = v
	}
	if v, ok := raw["path"].(string); ok {
		asset.Path = v
	}

	// Response data
	if v, ok := raw["status_code"].(float64); ok {
		asset.StatusCode = int(v)
	}
	if v, ok := raw["content_type"].(string); ok {
		asset.ContentType = v
	}
	if v, ok := raw["content_length"].(float64); ok {
		asset.ContentLength = int64(v)
	}
	if v, ok := raw["title"].(string); ok {
		asset.Title = v
	}
	if v, ok := raw["words"].(float64); ok {
		asset.Words = int(v)
	}
	if v, ok := raw["lines"].(float64); ok {
		asset.Lines = int(v)
	}

	// Network data
	if v, ok := raw["host_ip"].(string); ok {
		asset.HostIP = v
	}
	if v, ok := raw["dns_records"].([]interface{}); ok {
		asset.DnsRecords = interfaceSliceToStringSlice(v)
	} else if v, ok := raw["a"].([]interface{}); ok {
		asset.DnsRecords = interfaceSliceToStringSlice(v)
	}
	if v, ok := raw["tls"].(string); ok {
		asset.TLS = v
	}

	// Metadata
	if v, ok := raw["tech"].([]interface{}); ok {
		asset.Technologies = interfaceSliceToStringSlice(v)
	}
	if v, ok := raw["time"].(string); ok {
		asset.ResponseTime = v
	}
	if v, ok := raw["remarks"].(string); ok && v != "" {
		asset.Remarks = []string{v}
	} else if v, ok := raw["remarks"].([]interface{}); ok {
		asset.Remarks = interfaceSliceToStringSlice(v)
	}
	// Merge legacy "tags" field into Remarks
	if v, ok := raw["tags"].([]interface{}); ok {
		asset.Remarks = append(asset.Remarks, interfaceSliceToStringSlice(v)...)
	}
	// Language field
	if v, ok := raw["language"].(string); ok {
		asset.Language = v
	}
	if v, ok := raw["size"].(float64); ok {
		asset.Size = int64(v)
	}
	if v, ok := raw["loc"].(float64); ok {
		asset.LOC = int64(v)
	}
	if v, ok := raw["asset_type"].(string); ok {
		asset.AssetType = v
	}

	// CDN/WAF classification
	asset.IsCDN, asset.IsCloud, asset.IsWAF = DetectCDNFlags(raw)

	// Validate required fields
	if asset.AssetValue == "" {
		return nil, fmt.Errorf("asset_value is required")
	}
	if asset.Workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}

	// Auto-classify asset type if not provided
	if asset.AssetType == "" {
		asset.AssetType = ClassifyAssetType(asset.AssetValue)
	}
	// Refine: if we have HTTP response data, it's an "http" asset
	if asset.AssetType == "url" && (asset.StatusCode > 0 || asset.ContentLength > 0) {
		asset.AssetType = "http"
	}

	return asset, nil
}

// ImportEventLogs imports event logs from a JSONL file
func (i *JSONLImporter) ImportEventLogs(ctx context.Context, filePath string) (*ImportResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return i.ImportEventLogsFromReader(ctx, file)
}

// ImportEventLogsFromReader imports event logs from an io.Reader
func (i *JSONLImporter) ImportEventLogsFromReader(ctx context.Context, r io.Reader) (*ImportResult, error) {
	startTime := time.Now()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	result := &ImportResult{}
	batch := make([]*EventLog, 0, i.batchSize)

	for scanner.Scan() {
		result.Total++
		line := scanner.Bytes()

		if len(line) == 0 {
			continue
		}

		event, err := ParseEventLogLine(line)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Line:  result.Total,
				Error: err.Error(),
			})
			continue
		}

		batch = append(batch, event)

		if len(batch) >= i.batchSize {
			if _, err := i.db.NewInsert().Model(&batch).Exec(ctx); err != nil {
				return result, fmt.Errorf("batch insert failed: %w", err)
			}
			result.Imported += len(batch)
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if _, err := i.db.NewInsert().Model(&batch).Exec(ctx); err != nil {
			return result, fmt.Errorf("final batch insert failed: %w", err)
		}
		result.Imported += len(batch)
	}

	result.Duration = time.Since(startTime)
	return result, scanner.Err()
}

// ParseEventLogLine parses a single JSONL line into an EventLog
func ParseEventLogLine(line []byte) (*EventLog, error) {
	var event EventLog
	if err := json.Unmarshal(line, &event); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if event.Topic == "" {
		return nil, fmt.Errorf("topic is required")
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	return &event, nil
}

// Helper functions

func interfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ClassifyAssetType determines the asset type from the asset value
// Returns: "domain", "url", "ip", "repo_name", or "unknown"
func ClassifyAssetType(assetValue string) string {
	// Check for URL (starts with http:// or https://)
	if strings.HasPrefix(assetValue, "http://") || strings.HasPrefix(assetValue, "https://") {
		return "url"
	}

	// Check for IP address (IPv4 or IPv6)
	if net.ParseIP(assetValue) != nil {
		return "ip"
	}

	// Check for repo_name pattern (org/repo or user/repo)
	if strings.Count(assetValue, "/") == 1 && !strings.Contains(assetValue, ".") {
		parts := strings.Split(assetValue, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return "repo_name"
		}
	}

	// Check for domain (contains dot, no spaces, alphanumeric with hyphens)
	if strings.Contains(assetValue, ".") && !strings.Contains(assetValue, " ") {
		return "domain"
	}

	return "unknown"
}

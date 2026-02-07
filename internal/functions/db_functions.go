package functions

import (
	"bufio"
	"context"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/state"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// dbUpdate updates a specific field in a database table
// Usage: db_update('workspaces', 'example.com', 'total_urls', 100)
func (vf *vmFunc) dbUpdate(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbUpdate"))

	if len(call.Arguments) < 4 {
		logger.Get().Warn("dbUpdate: requires 4 arguments")
		return vf.errorValue("db_update requires 4 arguments: table, key, field, value")
	}

	table := call.Argument(0).String()
	key := call.Argument(1).String()
	field := call.Argument(2).String()
	value := call.Argument(3)
	logger.Get().Debug(terminal.HiGreen("dbUpdate")+" params", zap.String("table", table), zap.String("key", key), zap.String("field", field))

	db := database.GetDB()
	if db == nil {
		logger.Get().Warn("dbUpdate: database not connected")
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()
	var err error

	switch table {
	case "workspaces":
		err = vf.updateWorkspaceField(ctx, key, field, value)
	case "assets":
		err = vf.updateAssetField(ctx, key, field, value)
	case "runs":
		err = vf.updateRunField(ctx, key, field, value)
	default:
		return vf.errorValue(fmt.Sprintf("unsupported table: %s", table))
	}

	if err != nil {
		return vf.errorValue(fmt.Sprintf("update failed: %v", err))
	}

	return vf.vm.ToValue(true)
}

// updateWorkspaceField updates a field in the workspaces table
func (vf *vmFunc) updateWorkspaceField(ctx context.Context, name, field string, value goja.Value) error {
	db := database.GetDB()

	// Get current workspace
	var ws database.Workspace
	err := db.NewSelect().Model(&ws).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return fmt.Errorf("workspace not found: %s", name)
	}

	// Update the specific field
	switch field {
	case "total_assets":
		ws.TotalAssets, _ = strconv.Atoi(value.String())
	case "total_subdomains":
		ws.TotalSubdomains, _ = strconv.Atoi(value.String())
	case "total_urls":
		ws.TotalURLs, _ = strconv.Atoi(value.String())
	case "total_vulns":
		ws.TotalVulns, _ = strconv.Atoi(value.String())
	case "vuln_critical":
		ws.VulnCritical, _ = strconv.Atoi(value.String())
	case "vuln_high":
		ws.VulnHigh, _ = strconv.Atoi(value.String())
	case "vuln_medium":
		ws.VulnMedium, _ = strconv.Atoi(value.String())
	case "vuln_low":
		ws.VulnLow, _ = strconv.Atoi(value.String())
	case "vuln_potential":
		ws.VulnPotential, _ = strconv.Atoi(value.String())
	case "risk_score":
		ws.RiskScore, _ = strconv.ParseFloat(value.String(), 64)
	case "run_workflow":
		ws.RunWorkflow = value.String()
	case "local_path":
		ws.LocalPath = value.String()
	default:
		return fmt.Errorf("unsupported field: %s", field)
	}

	ws.UpdatedAt = time.Now()

	_, err = db.NewUpdate().Model(&ws).WherePK().Exec(ctx)
	return err
}

// updateAssetField updates a field in the assets table
func (vf *vmFunc) updateAssetField(ctx context.Context, idStr, field string, value goja.Value) error {
	db := database.GetDB()
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var asset database.Asset
	err := db.NewSelect().Model(&asset).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return fmt.Errorf("asset not found: %s", idStr)
	}

	switch field {
	case "labels":
		asset.Labels = value.String()
	case "source":
		asset.Source = value.String()
	case "asset_type":
		asset.AssetType = value.String()
	default:
		return fmt.Errorf("unsupported field: %s", field)
	}

	asset.UpdatedAt = time.Now()
	_, err = db.NewUpdate().Model(&asset).WherePK().Exec(ctx)
	return err
}

// updateRunField updates a field in the runs table
func (vf *vmFunc) updateRunField(ctx context.Context, id, field string, value goja.Value) error {
	db := database.GetDB()

	var run database.Run
	err := db.NewSelect().Model(&run).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return fmt.Errorf("run not found: %s", id)
	}

	switch field {
	case "status":
		run.Status = value.String()
	case "error_message":
		run.ErrorMessage = value.String()
	default:
		return fmt.Errorf("unsupported field: %s", field)
	}

	run.UpdatedAt = time.Now()
	_, err = db.NewUpdate().Model(&run).WherePK().Exec(ctx)
	return err
}

// dbImportAsset imports an asset from JSON data
// Usage: db_import_asset('example.com', '{"host":"sub.example.com","url":"https://..."}')
func (vf *vmFunc) dbImportAsset(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		return vf.errorValue("db_import_asset requires 2 arguments: workspace, json_data")
	}

	workspace := call.Argument(0).String()
	jsonData := call.Argument(1).String()

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	var asset database.Asset
	if err := json.Unmarshal([]byte(jsonData), &asset); err != nil {
		return vf.errorValue(fmt.Sprintf("invalid JSON: %v", err))
	}

	asset.Workspace = workspace
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	ctx := context.Background()
	_, err := db.NewInsert().Model(&asset).
		On("CONFLICT (workspace, asset_value, url) DO UPDATE").
		Set("updated_at = EXCLUDED.updated_at").
		Set("status_code = EXCLUDED.status_code").
		Set("title = EXCLUDED.title").
		Set("technologies = EXCLUDED.technologies").
		Exec(ctx)

	if err != nil {
		return vf.errorValue(fmt.Sprintf("import failed: %v", err))
	}

	return vf.vm.ToValue(true)
}

// dbQuickImportAsset imports an asset with just workspace and asset_value
// Usage: db_quick_import_asset('example.com', 'sub.example.com')
// Usage: db_quick_import_asset('example.com', 'sub.example.com', 'domain')
// Creates EventLog with topic "db.new.asset" when a new asset is created
func (vf *vmFunc) dbQuickImportAsset(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		return vf.errorValue("db_quick_import_asset requires at least 2 arguments: workspace, asset_value")
	}

	workspace := call.Argument(0).String()
	assetValue := call.Argument(1).String()

	// Optional asset_type (3rd argument)
	var assetType string
	if len(call.Arguments) >= 3 {
		assetType = call.Argument(2).String()
	}
	if assetType == "" {
		assetType = database.ClassifyAssetType(assetValue)
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Check if asset already exists
	exists, err := db.NewSelect().
		Model((*database.Asset)(nil)).
		Where("workspace = ? AND asset_value = ?", workspace, assetValue).
		Exists(ctx)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("check failed: %v", err))
	}

	now := time.Now()
	asset := &database.Asset{
		Workspace:  workspace,
		AssetValue: assetValue,
		AssetType:  assetType,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Insert or update
	_, err = db.NewInsert().Model(asset).
		On("CONFLICT (workspace, asset_value, url) DO UPDATE").
		Set("updated_at = EXCLUDED.updated_at").
		Set("asset_type = EXCLUDED.asset_type").
		Exec(ctx)

	if err != nil {
		return vf.errorValue(fmt.Sprintf("import failed: %v", err))
	}

	// Create event log for new assets only
	if !exists {
		eventLog := &database.EventLog{
			Topic:      "db.new.asset",
			EventID:    uuid.New().String(),
			Name:       "New asset imported",
			Source:     "db_quick_import_asset",
			SourceType: "function",
			DataType:   "asset",
			Data:       fmt.Sprintf(`{"workspace":"%s","asset_value":"%s","asset_type":"%s"}`, workspace, assetValue, assetType),
			Workspace:  workspace,
			Processed:  false,
			CreatedAt:  now,
		}
		_, _ = db.NewInsert().Model(eventLog).Exec(ctx)
	}

	return vf.vm.ToValue(true)
}

// dbPartialImportAsset imports an asset with only workspace, asset_type, and asset_value
// Usage: db_partial_import_asset('example.com', 'domain', 'sub.example.com')
func (vf *vmFunc) dbPartialImportAsset(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 3 {
		return vf.errorValue("db_partial_import_asset requires 3 arguments: workspace, asset_type, asset_value")
	}

	workspace := call.Argument(0).String()
	assetType := call.Argument(1).String()
	assetValue := call.Argument(2).String()

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()
	now := time.Now()
	asset := &database.Asset{
		Workspace:  workspace,
		AssetValue: assetValue,
		AssetType:  assetType,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := db.NewInsert().Model(asset).
		On("CONFLICT (workspace, asset_value, url) DO UPDATE").
		Set("updated_at = EXCLUDED.updated_at").
		Set("asset_type = EXCLUDED.asset_type").
		Exec(ctx)

	if err != nil {
		return vf.errorValue(fmt.Sprintf("import failed: %v", err))
	}

	return vf.vm.ToValue(true)
}

// dbPartialImportAssetFile reads a file line by line and imports each line as an asset
// Usage: db_partial_import_asset_file('example.com', 'domain', '/path/to/subdomains.txt')
func (vf *vmFunc) dbPartialImportAssetFile(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 3 {
		return vf.errorValue("db_partial_import_asset_file requires 3 arguments: workspace, asset_type, file_path")
	}

	workspace := call.Argument(0).String()
	assetType := call.Argument(1).String()
	filePath := call.Argument(2).String()

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to open file: %v", err))
	}
	defer f.Close()

	ctx := context.Background()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		now := time.Now()
		asset := &database.Asset{
			Workspace:  workspace,
			AssetValue: line,
			AssetType:  assetType,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		_, err := db.NewInsert().Model(asset).
			On("CONFLICT (workspace, asset_value, url) DO UPDATE").
			Set("updated_at = EXCLUDED.updated_at").
			Set("asset_type = EXCLUDED.asset_type").
			Exec(ctx)

		if err != nil {
			logger.Get().Warn("dbPartialImportAssetFile: failed to import line",
				zap.String("line", line), zap.Error(err))
			continue
		}
		count++
	}

	if err := scanner.Err(); err != nil {
		return vf.errorValue(fmt.Sprintf("error reading file: %v", err))
	}

	return vf.vm.ToValue(count)
}

// dbRawInsertAsset inserts an asset from JSON data (pure insert, fails if duplicate exists)
// Usage: db_raw_insert_asset('example.com', '{"asset_value":"sub.example.com","asset_type":"subdomain","url":"https://..."}')
// Returns: asset ID (int) on success, error string on failure
func (vf *vmFunc) dbRawInsertAsset(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		return vf.errorValue("db_raw_insert_asset requires 2 arguments: workspace, json_data")
	}

	workspace := call.Argument(0).String()
	jsonData := call.Argument(1).String()

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	var asset database.Asset
	if err := json.Unmarshal([]byte(jsonData), &asset); err != nil {
		return vf.errorValue(fmt.Sprintf("invalid JSON: %v", err))
	}

	asset.Workspace = workspace
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	ctx := context.Background()
	_, err := db.NewInsert().Model(&asset).Exec(ctx)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("insert failed: %v", err))
	}

	return vf.vm.ToValue(asset.ID)
}

// dbTotalURLs counts lines in a file and adds to workspace total_urls (+=)
// Usage: db_total_urls('/path/to/urls.txt') or db_total_urls(0) to reset
func (vf *vmFunc) dbTotalURLs(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_urls")
}

// dbTotalSubdomains counts lines in a file and adds to workspace total_subdomains (+=)
// Usage: db_total_subdomains('/path/to/subdomains.txt') or db_total_subdomains(0) to reset
func (vf *vmFunc) dbTotalSubdomains(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_subdomains")
}

// dbTotalAssets counts lines in a file and adds to workspace total_assets (+=)
// Usage: db_total_assets('/path/to/assets.txt') or db_total_assets(0) to reset
func (vf *vmFunc) dbTotalAssets(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_assets")
}

// dbTotalVulns counts lines in a file and adds to workspace total_vulns (+=)
// Usage: db_total_vulns('/path/to/vulns.txt') or db_total_vulns(0) to reset
func (vf *vmFunc) dbTotalVulns(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_vulns")
}

// dbVulnCritical counts lines in a file and adds to workspace vuln_critical (+=)
// Usage: db_vuln_critical('/path/to/critical.txt') or db_vuln_critical(0) to reset
func (vf *vmFunc) dbVulnCritical(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "vuln_critical")
}

// dbVulnHigh counts lines in a file and adds to workspace vuln_high (+=)
// Usage: db_vuln_high('/path/to/high.txt') or db_vuln_high(0) to reset
func (vf *vmFunc) dbVulnHigh(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "vuln_high")
}

// dbVulnMedium counts lines in a file and adds to workspace vuln_medium (+=)
// Usage: db_vuln_medium('/path/to/medium.txt') or db_vuln_medium(0) to reset
func (vf *vmFunc) dbVulnMedium(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "vuln_medium")
}

// dbVulnLow counts lines in a file and adds to workspace vuln_low (+=)
// Usage: db_vuln_low('/path/to/low.txt') or db_vuln_low(0) to reset
func (vf *vmFunc) dbVulnLow(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "vuln_low")
}

// dbTotalIPs counts lines in a file and adds to workspace total_ips (+=)
// Usage: db_total_ips('/path/to/ips.txt') or db_total_ips(0) to reset
func (vf *vmFunc) dbTotalIPs(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_ips")
}

// dbTotalLinks counts lines in a file and adds to workspace total_links (+=)
// Usage: db_total_links('/path/to/links.txt') or db_total_links(0) to reset
func (vf *vmFunc) dbTotalLinks(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_links")
}

// dbTotalContent counts lines in a file and adds to workspace total_content (+=)
// Usage: db_total_content('/path/to/content.txt') or db_total_content(0) to reset
func (vf *vmFunc) dbTotalContent(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_content")
}

// dbTotalArchive counts lines in a file and adds to workspace total_archive (+=)
// Usage: db_total_archive('/path/to/archive.txt') or db_total_archive(0) to reset
func (vf *vmFunc) dbTotalArchive(call goja.FunctionCall) goja.Value {
	return vf.updateWorkspaceStatFromFile(call, "total_archive")
}

// updateWorkspaceStatFromFile is a helper that counts non-empty lines and updates workspace
// Supports += accumulation behavior and reset to 0 when called with 0 argument
func (vf *vmFunc) updateWorkspaceStatFromFile(call goja.FunctionCall, field string) goja.Value {
	if len(call.Arguments) < 1 {
		return vf.errorValue(fmt.Sprintf("%s requires 1 argument: file_path or 0 to reset", field))
	}

	arg := call.Argument(0)

	// Check for reset: if arg is 0 (number) or "0" (string), reset to 0
	isReset := false
	argExport := arg.Export()
	switch v := argExport.(type) {
	case int64:
		if v == 0 {
			isReset = true
		}
	case float64:
		if v == 0 {
			isReset = true
		}
	case string:
		if v == "0" {
			isReset = true
		}
	}

	var count int
	if isReset {
		count = 0
	} else {
		filePath := arg.String()
		// Count non-empty lines in file
		var err error
		count, err = countNonEmptyLines(filePath)
		if err != nil {
			// If file doesn't exist, count is 0
			count = 0
		}
	}

	// Get workspace name from context
	workspace := vf.getContext().workspaceName
	if workspace == "" {
		return vf.errorValue("workspace not set in context")
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Get existing workspace or create new
	var ws database.Workspace
	err := db.NewSelect().Model(&ws).Where("name = ?", workspace).Scan(ctx)
	if err != nil {
		// Create workspace if it doesn't exist
		ws = database.Workspace{
			Name:      workspace,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Set the appropriate field: reset to 0 or += count
	switch field {
	case "total_urls":
		if isReset {
			ws.TotalURLs = 0
		} else {
			ws.TotalURLs += count
		}
	case "total_subdomains":
		if isReset {
			ws.TotalSubdomains = 0
		} else {
			ws.TotalSubdomains += count
		}
	case "total_assets":
		if isReset {
			ws.TotalAssets = 0
		} else {
			ws.TotalAssets += count
		}
	case "total_vulns":
		if isReset {
			ws.TotalVulns = 0
		} else {
			ws.TotalVulns += count
		}
	case "total_ips":
		if isReset {
			ws.TotalIPs = 0
		} else {
			ws.TotalIPs += count
		}
	case "total_links":
		if isReset {
			ws.TotalLinks = 0
		} else {
			ws.TotalLinks += count
		}
	case "total_content":
		if isReset {
			ws.TotalContent = 0
		} else {
			ws.TotalContent += count
		}
	case "total_archive":
		if isReset {
			ws.TotalArchive = 0
		} else {
			ws.TotalArchive += count
		}
	case "vuln_critical":
		if isReset {
			ws.VulnCritical = 0
		} else {
			ws.VulnCritical += count
		}
	case "vuln_high":
		if isReset {
			ws.VulnHigh = 0
		} else {
			ws.VulnHigh += count
		}
	case "vuln_medium":
		if isReset {
			ws.VulnMedium = 0
		} else {
			ws.VulnMedium += count
		}
	case "vuln_low":
		if isReset {
			ws.VulnLow = 0
		} else {
			ws.VulnLow += count
		}
	}

	ws.UpdatedAt = time.Now()

	// Upsert workspace (include new fields)
	_, err = db.NewInsert().Model(&ws).
		On("CONFLICT (name) DO UPDATE").
		Set("total_urls = EXCLUDED.total_urls").
		Set("total_subdomains = EXCLUDED.total_subdomains").
		Set("total_assets = EXCLUDED.total_assets").
		Set("total_vulns = EXCLUDED.total_vulns").
		Set("total_ips = EXCLUDED.total_ips").
		Set("total_links = EXCLUDED.total_links").
		Set("total_content = EXCLUDED.total_content").
		Set("total_archive = EXCLUDED.total_archive").
		Set("vuln_critical = EXCLUDED.vuln_critical").
		Set("vuln_high = EXCLUDED.vuln_high").
		Set("vuln_medium = EXCLUDED.vuln_medium").
		Set("vuln_low = EXCLUDED.vuln_low").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)

	if err != nil {
		return vf.errorValue(fmt.Sprintf("update failed: %v", err))
	}

	// Return the new value of the field
	var resultValue int
	switch field {
	case "total_urls":
		resultValue = ws.TotalURLs
	case "total_subdomains":
		resultValue = ws.TotalSubdomains
	case "total_assets":
		resultValue = ws.TotalAssets
	case "total_vulns":
		resultValue = ws.TotalVulns
	case "total_ips":
		resultValue = ws.TotalIPs
	case "total_links":
		resultValue = ws.TotalLinks
	case "total_content":
		resultValue = ws.TotalContent
	case "total_archive":
		resultValue = ws.TotalArchive
	case "vuln_critical":
		resultValue = ws.VulnCritical
	case "vuln_high":
		resultValue = ws.VulnHigh
	case "vuln_medium":
		resultValue = ws.VulnMedium
	case "vuln_low":
		resultValue = ws.VulnLow
	default:
		resultValue = count
	}

	return vf.vm.ToValue(resultValue)
}

// countNonEmptyLines counts non-empty lines in a file
func countNonEmptyLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			count++
		}
	}

	return count, scanner.Err()
}

// runtimeExport exports the current run state to StateFile
// Usage: runtime_export()
func (vf *vmFunc) runtimeExport(call goja.FunctionCall) goja.Value {
	if vf.getContext().stateFile == "" {
		return vf.errorValue("StateFile not set in context")
	}

	ctx := &state.ExportContext{
		RunUUID:       vf.getContext().scanID,
		WorkspaceName: vf.getContext().workspaceName,
		WorkspacePath: vf.getContext().workspacePath,
		WorkflowName:  vf.getContext().workflowName,
		WorkflowKind:  vf.getContext().workflowKind,
		Target:        vf.getContext().target,
		Status:        "running",
	}

	err := state.Export(vf.getContext().stateFile, ctx)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("state export failed: %v", err))
	}

	return vf.vm.ToValue(true)
}

// errorValue creates an error return value
func (vf *vmFunc) errorValue(msg string) goja.Value {
	return vf.vm.ToValue(fmt.Sprintf("error: %s", msg))
}

// Supported artifact types
var validArtifactTypes = map[string]bool{
	"text":     true,
	"jsonl":    true,
	"markdown": true,
	"html":     true,
}

// validOutputFormats defines the allowed output formats for db_select functions
var validOutputFormats = map[string]bool{
	"markdown": true,
	"jsonl":    true,
}

// validateOutputFormat checks if the format is valid
func validateOutputFormat(format string) error {
	if !validOutputFormats[format] {
		return fmt.Errorf("invalid format '%s'. Supported: markdown, jsonl", format)
	}
	return nil
}

// assetColumns defines the columns for asset markdown table
var assetColumns = []string{"id", "workspace", "asset_value", "url", "status_code", "title"}

// vulnerabilityColumns defines the columns for vulnerability markdown table
var vulnerabilityColumns = []string{"id", "workspace", "severity", "vuln_title", "asset_value"}

// runColumns defines the columns for run markdown table
var runColumns = []string{"run_uuid", "workflow_name", "target", "workspace", "status", "total_steps", "completed_steps", "started_at"}

// formatAsMarkdownTable converts a slice of maps to a markdown table
func formatAsMarkdownTable(rows []map[string]interface{}, columns []string) string {
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder

	// Header row
	sb.WriteString("| ")
	sb.WriteString(strings.Join(columns, " | "))
	sb.WriteString(" |\n")

	// Separator row
	sb.WriteString("|")
	for range columns {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows {
		sb.WriteString("| ")
		for i, col := range columns {
			if i > 0 {
				sb.WriteString(" | ")
			}
			if val, ok := row[col]; ok {
				sb.WriteString(formatMarkdownValue(val))
			}
		}
		sb.WriteString(" |\n")
	}

	return sb.String()
}

// formatMarkdownValue converts a value to a markdown-safe string
func formatMarkdownValue(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		// Escape pipe characters and truncate long strings
		escaped := strings.ReplaceAll(v, "|", "\\|")
		escaped = strings.ReplaceAll(escaped, "\n", " ")
		if len(escaped) > 100 {
			escaped = escaped[:97] + "..."
		}
		return escaped
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case []string:
		return strings.Join(v, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatAsJSONL converts rows to JSONL format
func formatAsJSONL(rows []map[string]interface{}) string {
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, row := range rows {
		jsonBytes, err := json.Marshal(row)
		if err != nil {
			continue
		}
		sb.Write(jsonBytes)
		sb.WriteString("\n")
	}
	return sb.String()
}

// assetToMap converts an Asset struct to a map for formatting
func assetToMap(asset *database.Asset) map[string]interface{} {
	return map[string]interface{}{
		"id":             asset.ID,
		"workspace":      asset.Workspace,
		"asset_value":    asset.AssetValue,
		"url":            asset.URL,
		"input":          asset.Input,
		"scheme":         asset.Scheme,
		"method":         asset.Method,
		"path":           asset.Path,
		"status_code":    asset.StatusCode,
		"content_type":   asset.ContentType,
		"content_length": asset.ContentLength,
		"title":          asset.Title,
		"words":          asset.Words,
		"lines":          asset.Lines,
		"host_ip":        asset.HostIP,
		"dns_records":    asset.DnsRecords,
		"tls":            asset.TLS,
		"asset_type":     asset.AssetType,
		"technologies":   asset.Technologies,
		"response_time":  asset.ResponseTime,
		"labels":         asset.Labels,
		"source":         asset.Source,
		"created_at":     asset.CreatedAt,
		"updated_at":     asset.UpdatedAt,
	}
}

// vulnerabilityToMap converts a Vulnerability struct to a map for formatting
func vulnerabilityToMap(vuln *database.Vulnerability) map[string]interface{} {
	return map[string]interface{}{
		"id":          vuln.ID,
		"workspace":   vuln.Workspace,
		"vuln_info":   vuln.VulnInfo,
		"vuln_title":  vuln.VulnTitle,
		"vuln_desc":   vuln.VulnDesc,
		"vuln_poc":    vuln.VulnPOC,
		"severity":    vuln.Severity,
		"asset_type":  vuln.AssetType,
		"asset_value": vuln.AssetValue,
		"tags":        vuln.Tags,
		"created_at":  vuln.CreatedAt,
		"updated_at":  vuln.UpdatedAt,
	}
}

// runToMap converts a Run struct to a map for formatting
func runToMap(run *database.Run) map[string]interface{} {
	result := map[string]interface{}{
		"run_uuid":        run.RunUUID,
		"workflow_name":   run.WorkflowName,
		"workflow_kind":   run.WorkflowKind,
		"target":          run.Target,
		"workspace":       run.Workspace,
		"status":          run.Status,
		"total_steps":     run.TotalSteps,
		"completed_steps": run.CompletedSteps,
		"trigger_type":    run.TriggerType,
	}
	if run.StartedAt != nil {
		result["started_at"] = run.StartedAt.Format(time.RFC3339)
	}
	if run.CompletedAt != nil {
		result["completed_at"] = run.CompletedAt.Format(time.RFC3339)
	}
	if run.ErrorMessage != "" {
		result["error_message"] = run.ErrorMessage
	}
	return result
}

// Forbidden SQL keywords for security
var forbiddenSQLKeywords = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
	"TRUNCATE", "GRANT", "REVOKE", "EXECUTE", "EXEC",
	"INTO OUTFILE", "INTO DUMPFILE", "LOAD_FILE",
}

// validateSelectQuery ensures the query is a safe SELECT statement
func validateSelectQuery(query string) error {
	// Normalize query for checking
	normalized := strings.ToUpper(strings.TrimSpace(query))

	// Must start with SELECT
	if !strings.HasPrefix(normalized, "SELECT") {
		return fmt.Errorf("only SELECT queries are allowed")
	}

	// Check for forbidden keywords
	for _, keyword := range forbiddenSQLKeywords {
		if strings.Contains(normalized, keyword) {
			return fmt.Errorf("query contains forbidden keyword: %s", keyword)
		}
	}

	// Check for common SQL injection patterns
	trimmedQuery := strings.TrimSpace(query)
	if strings.Contains(normalized, ";") && !strings.HasSuffix(trimmedQuery, ";") {
		return fmt.Errorf("multiple statements not allowed")
	}

	// Disallow comments that could hide malicious code
	if strings.Contains(query, "--") || strings.Contains(query, "/*") {
		return fmt.Errorf("SQL comments not allowed")
	}

	return nil
}

// dbSelectAssets selects assets from the database and returns formatted output
// Usage: db_select_assets(workspace, format) -> string
// workspace: optional filter by workspace (empty = all)
// format: "markdown" or "jsonl"
func (vf *vmFunc) dbSelectAssets(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbSelectAssets"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_select_assets requires 2 arguments: workspace, format")
	}

	workspace := call.Argument(0).String()
	format := call.Argument(1).String()

	// Handle "undefined" from JavaScript
	if workspace == "undefined" {
		workspace = ""
	}

	if err := validateOutputFormat(format); err != nil {
		return vf.errorValue(err.Error())
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Build query
	var assets []*database.Asset
	query := db.NewSelect().Model(&assets).Order("created_at DESC")

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}

	// Limit results to prevent memory issues
	query = query.Limit(1000)

	if err := query.Scan(ctx); err != nil {
		return vf.errorValue(fmt.Sprintf("query failed: %v", err))
	}

	// Convert to maps
	rows := make([]map[string]interface{}, len(assets))
	for i, asset := range assets {
		rows[i] = assetToMap(asset)
	}

	// Format output
	var output string
	switch format {
	case "markdown":
		output = formatAsMarkdownTable(rows, assetColumns)
	case "jsonl":
		output = formatAsJSONL(rows)
	}

	return vf.vm.ToValue(output)
}

// dbSelectAssetsFiltered selects assets from the database with multiple filters
// Usage: db_select_assets_filtered(workspace, status_code, asset_type, format) -> string
// workspace: filter by workspace (empty = all)
// status_code: filter by HTTP status code (0 = any)
// asset_type: filter by asset type (empty = any)
// format: "markdown" or "jsonl"
func (vf *vmFunc) dbSelectAssetsFiltered(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbSelectAssetsFiltered"))

	if len(call.Arguments) < 4 {
		return vf.errorValue("db_select_assets_filtered requires 4 arguments: workspace, status_code, asset_type, format")
	}

	workspace := call.Argument(0).String()
	statusCodeStr := call.Argument(1).String()
	assetType := call.Argument(2).String()
	format := call.Argument(3).String()

	// Handle "undefined" from JavaScript
	if workspace == "undefined" {
		workspace = ""
	}
	if assetType == "undefined" {
		assetType = ""
	}

	statusCode := 0
	if statusCodeStr != "" && statusCodeStr != "undefined" && statusCodeStr != "0" {
		statusCode, _ = strconv.Atoi(statusCodeStr)
	}

	if err := validateOutputFormat(format); err != nil {
		return vf.errorValue(err.Error())
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Build query with filters
	var assets []*database.Asset
	query := db.NewSelect().Model(&assets).Order("created_at DESC")

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}
	if statusCode != 0 {
		query = query.Where("status_code = ?", statusCode)
	}
	if assetType != "" {
		query = query.Where("asset_type = ?", assetType)
	}

	// Limit results to prevent memory issues
	query = query.Limit(1000)

	if err := query.Scan(ctx); err != nil {
		return vf.errorValue(fmt.Sprintf("query failed: %v", err))
	}

	// Convert to maps
	rows := make([]map[string]interface{}, len(assets))
	for i, asset := range assets {
		rows[i] = assetToMap(asset)
	}

	// Format output
	var output string
	switch format {
	case "markdown":
		output = formatAsMarkdownTable(rows, assetColumns)
	case "jsonl":
		output = formatAsJSONL(rows)
	}

	return vf.vm.ToValue(output)
}

// dbSelectVulnerabilities selects vulnerabilities from the database and returns formatted output
// Usage: db_select_vulnerabilities(workspace, format) -> string
// workspace: optional filter by workspace (empty = all)
// format: "markdown" or "jsonl"
func (vf *vmFunc) dbSelectVulnerabilities(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbSelectVulnerabilities"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_select_vulnerabilities requires 2 arguments: workspace, format")
	}

	workspace := call.Argument(0).String()
	format := call.Argument(1).String()

	// Handle "undefined" from JavaScript
	if workspace == "undefined" {
		workspace = ""
	}

	if err := validateOutputFormat(format); err != nil {
		return vf.errorValue(err.Error())
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Build query
	var vulns []*database.Vulnerability
	query := db.NewSelect().Model(&vulns).Order("created_at DESC")

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}

	// Limit results to prevent memory issues
	query = query.Limit(1000)

	if err := query.Scan(ctx); err != nil {
		return vf.errorValue(fmt.Sprintf("query failed: %v", err))
	}

	// Convert to maps
	rows := make([]map[string]interface{}, len(vulns))
	for i, vuln := range vulns {
		rows[i] = vulnerabilityToMap(vuln)
	}

	// Format output
	var output string
	switch format {
	case "markdown":
		output = formatAsMarkdownTable(rows, vulnerabilityColumns)
	case "jsonl":
		output = formatAsJSONL(rows)
	}

	return vf.vm.ToValue(output)
}

// dbSelectVulnerabilitiesFiltered selects vulnerabilities from the database with multiple filters
// Usage: db_select_vulnerabilities_filtered(workspace, severity, asset_value, format) -> string
// workspace: filter by workspace (empty = all)
// severity: filter by severity level (empty = any) - critical, high, medium, low
// asset_value: filter by asset value (empty = any)
// format: "markdown" or "jsonl"
func (vf *vmFunc) dbSelectVulnerabilitiesFiltered(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbSelectVulnerabilitiesFiltered"))

	if len(call.Arguments) < 4 {
		return vf.errorValue("db_select_vulnerabilities_filtered requires 4 arguments: workspace, severity, asset_value, format")
	}

	workspace := call.Argument(0).String()
	severity := call.Argument(1).String()
	assetValue := call.Argument(2).String()
	format := call.Argument(3).String()

	// Handle "undefined" from JavaScript
	if workspace == "undefined" {
		workspace = ""
	}
	if severity == "undefined" {
		severity = ""
	}
	if assetValue == "undefined" {
		assetValue = ""
	}

	if err := validateOutputFormat(format); err != nil {
		return vf.errorValue(err.Error())
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Build query with filters
	var vulns []*database.Vulnerability
	query := db.NewSelect().Model(&vulns).Order("created_at DESC")

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if assetValue != "" {
		query = query.Where("asset_value = ?", assetValue)
	}

	// Limit results to prevent memory issues
	query = query.Limit(1000)

	if err := query.Scan(ctx); err != nil {
		return vf.errorValue(fmt.Sprintf("query failed: %v", err))
	}

	// Convert to maps
	rows := make([]map[string]interface{}, len(vulns))
	for i, vuln := range vulns {
		rows[i] = vulnerabilityToMap(vuln)
	}

	// Format output
	var output string
	switch format {
	case "markdown":
		output = formatAsMarkdownTable(rows, vulnerabilityColumns)
	case "jsonl":
		output = formatAsJSONL(rows)
	}

	return vf.vm.ToValue(output)
}

// dbSelect executes an arbitrary SELECT query and returns formatted output
// Usage: db_select(sql_query, format) -> string
// sql_query: SELECT SQL query (only SELECT allowed)
// format: "markdown" or "jsonl"
func (vf *vmFunc) dbSelect(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbSelect"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_select requires 2 arguments: sql_query, format")
	}

	sqlQuery := call.Argument(0).String()
	format := call.Argument(1).String()

	if sqlQuery == "" || sqlQuery == "undefined" {
		return vf.errorValue("sql_query cannot be empty")
	}

	if err := validateOutputFormat(format); err != nil {
		return vf.errorValue(err.Error())
	}

	// Security validation
	if err := validateSelectQuery(sqlQuery); err != nil {
		return vf.errorValue(fmt.Sprintf("security check failed: %v", err))
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Execute raw query
	sqlDB := db.DB

	rawRows, err := sqlDB.QueryContext(ctx, sqlQuery)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("query execution failed: %v", err))
	}
	defer func() { _ = rawRows.Close() }()

	// Get column names
	columns, err := rawRows.Columns()
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to get columns: %v", err))
	}

	// Scan all rows
	var rows []map[string]interface{}
	for rawRows.Next() {
		// Create slice to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rawRows.Scan(valuePtrs...); err != nil {
			return vf.errorValue(fmt.Sprintf("failed to scan row: %v", err))
		}

		// Convert to map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle byte slices (common for TEXT columns)
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		rows = append(rows, row)

		// Limit results
		if len(rows) >= 1000 {
			break
		}
	}

	if err := rawRows.Err(); err != nil {
		return vf.errorValue(fmt.Sprintf("row iteration error: %v", err))
	}

	// Format output
	var output string
	switch format {
	case "markdown":
		output = formatAsMarkdownTable(rows, columns)
	case "jsonl":
		output = formatAsJSONL(rows)
	}

	return vf.vm.ToValue(output)
}

// dbRegisterArtifact registers a file as an artifact in the database
// Usage: register_artifact('/path/to/file.txt', 'text')
// type is optional, defaults to 'text'. Supported: text, jsonl, markdown, html
func (vf *vmFunc) dbRegisterArtifact(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return vf.errorValue("register_artifact requires at least 1 argument: file_path")
	}

	filePath := call.Argument(0).String()
	if filePath == "" || filePath == "undefined" {
		return vf.errorValue("file_path cannot be empty")
	}

	// Get optional type, default to "text"
	artifactType := "text"
	if len(call.Arguments) >= 2 {
		typeArg := call.Argument(1).String()
		if typeArg != "" && typeArg != "undefined" {
			artifactType = typeArg
		}
	}

	// Validate artifact type
	if !validArtifactTypes[artifactType] {
		return vf.errorValue(fmt.Sprintf("invalid artifact type '%s'. Supported: text, jsonl, markdown, html", artifactType))
	}

	// Check if file exists and get stats
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("file not found: %s", filePath))
	}

	// Get run ID from context (integer for database foreign key)
	runID := vf.getContext().runID
	if runID == 0 {
		return vf.errorValue("run ID not set in context")
	}

	// Count lines in file
	lineCount, _ := countNonEmptyLines(filePath)

	// Generate unique ID
	artifactID := fmt.Sprintf("artifact-%d", time.Now().UnixNano())

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Create artifact record
	artifact := database.Artifact{
		ID:           artifactID,
		RunID:        runID,
		Name:         filepath.Base(filePath),
		ArtifactPath: filePath,
		ArtifactType: artifactType,
		ContentType:  database.ContentTypeUnknown,
		SizeBytes:    fileInfo.Size(),
		LineCount:    lineCount,
		CreatedAt:    time.Now(),
	}

	// Insert or update artifact
	_, err = db.NewInsert().Model(&artifact).
		On("CONFLICT (id) DO UPDATE").
		Set("artifact_path = EXCLUDED.artifact_path").
		Set("artifact_type = EXCLUDED.artifact_type").
		Set("content_type = EXCLUDED.content_type").
		Set("size_bytes = EXCLUDED.size_bytes").
		Set("line_count = EXCLUDED.line_count").
		Exec(ctx)

	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to register artifact: %v", err))
	}

	return vf.vm.ToValue(true)
}

func (vf *vmFunc) storeArtifact(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return vf.errorValue("store_artifact requires at least 1 argument: file_path")
	}

	filePath := call.Argument(0).String()
	if filePath == "" || filePath == "undefined" {
		return vf.errorValue("file_path cannot be empty")
	}

	workspaceName := vf.getContext().workspaceName
	if workspaceName == "" {
		return vf.errorValue("workspace not set in context")
	}

	runID := vf.getContext().runID
	if runID == 0 {
		return vf.errorValue("run ID not set in context")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("file not found: %s", filePath))
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	lineCount := 0
	sizeBytes := int64(0)
	var contentType string
	if info.IsDir() {
		contentType = database.ContentTypeFolder
	} else {
		sizeBytes = info.Size()
		lineCount, _ = countNonEmptyLines(filePath)
		contentType = inferContentTypeFromPath(filePath)
	}

	ctx := context.Background()

	artifact := database.Artifact{
		ID:           uuid.New().String(),
		RunID:        runID,
		Workspace:    workspaceName,
		Name:         filepath.Base(filePath),
		ArtifactPath: filePath,
		ArtifactType: database.ArtifactTypeOutput,
		ContentType:  contentType,
		SizeBytes:    sizeBytes,
		LineCount:    lineCount,
		CreatedAt:    time.Now(),
	}

	_, err = db.NewInsert().Model(&artifact).
		On("CONFLICT (id) DO UPDATE").
		Set("run_id = EXCLUDED.run_id").
		Set("workspace = EXCLUDED.workspace").
		Set("name = EXCLUDED.name").
		Set("artifact_path = EXCLUDED.artifact_path").
		Set("artifact_type = EXCLUDED.artifact_type").
		Set("content_type = EXCLUDED.content_type").
		Set("size_bytes = EXCLUDED.size_bytes").
		Set("line_count = EXCLUDED.line_count").
		Set("description = EXCLUDED.description").
		Exec(ctx)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to store artifact: %v", err))
	}

	return vf.vm.ToValue(true)
}

func inferContentTypeFromPath(filePath string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	switch ext {
	case "json":
		return database.ContentTypeJSON
	case "jsonl":
		return database.ContentTypeJSONL
	case "yaml", "yml":
		return database.ContentTypeYAML
	case "html", "htm":
		return database.ContentTypeHTML
	case "md", "markdown":
		return database.ContentTypeMarkdown
	case "log":
		return database.ContentTypeLog
	case "pdf":
		return database.ContentTypePDF
	case "png":
		return database.ContentTypePNG
	case "zip":
		return database.ContentTypeZip
	case "txt", "text":
		return database.ContentTypeText
	default:
		return database.ContentTypeUnknown
	}
}

// getWorkspaceStats retrieves workspace statistics from database
func (vf *vmFunc) getWorkspaceStats(workspaceName string) (*database.Workspace, error) {
	if workspaceName == "" {
		return nil, fmt.Errorf("workspace name not set")
	}
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	var ws database.Workspace
	err := db.NewSelect().Model(&ws).Where("name = ?", workspaceName).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

// dbSelectTotalSubdomains returns the current total_subdomains from workspace
// Usage: db_select_total_subdomains() -> int
func (vf *vmFunc) dbSelectTotalSubdomains(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_total_subdomains"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_total_subdomains")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.TotalSubdomains)
}

// dbSelectTotalURLs returns the current total_urls from workspace
// Usage: db_select_total_urls() -> int
func (vf *vmFunc) dbSelectTotalURLs(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_total_urls"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_total_urls")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.TotalURLs)
}

// dbSelectTotalAssets returns the current total_assets from workspace
// Usage: db_select_total_assets() -> int
func (vf *vmFunc) dbSelectTotalAssets(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_total_assets"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_total_assets")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.TotalAssets)
}

// dbSelectTotalVulns returns the current total_vulns from workspace
// Usage: db_select_total_vulns() -> int
func (vf *vmFunc) dbSelectTotalVulns(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_total_vulns"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_total_vulns")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.TotalVulns)
}

// dbSelectVulnCritical returns the current vuln_critical count from workspace
// Usage: db_select_vuln_critical() -> int
func (vf *vmFunc) dbSelectVulnCritical(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_vuln_critical"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_vuln_critical")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.VulnCritical)
}

// dbSelectVulnHigh returns the current vuln_high count from workspace
// Usage: db_select_vuln_high() -> int
func (vf *vmFunc) dbSelectVulnHigh(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_vuln_high"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_vuln_high")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.VulnHigh)
}

// dbSelectVulnMedium returns the current vuln_medium count from workspace
// Usage: db_select_vuln_medium() -> int
func (vf *vmFunc) dbSelectVulnMedium(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_vuln_medium"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_vuln_medium")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.VulnMedium)
}

// dbSelectVulnLow returns the current vuln_low count from workspace
// Usage: db_select_vuln_low() -> int
func (vf *vmFunc) dbSelectVulnLow(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_select_vuln_low"))
	ws, err := vf.getWorkspaceStats(vf.getContext().workspaceName)
	if err != nil {
		logger.Get().Debug(terminal.HiGreen("db_select_vuln_low")+" error", zap.Error(err))
		return vf.vm.ToValue(0)
	}
	return vf.vm.ToValue(ws.VulnLow)
}

// dbSelectToFile executes a SELECT query and writes markdown output to a file
// Usage: db_select_to_file(sql_query, dest_file) -> bool
func (vf *vmFunc) dbSelectToFile(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbSelectToFile"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_select_to_file requires 2 arguments: sql_query, dest_file")
	}

	sqlQuery := call.Argument(0).String()
	destFile := call.Argument(1).String()

	if sqlQuery == "" || sqlQuery == "undefined" {
		return vf.errorValue("sql_query cannot be empty")
	}

	if destFile == "" || destFile == "undefined" {
		return vf.errorValue("dest_file cannot be empty")
	}

	// Security validation
	if err := validateSelectQuery(sqlQuery); err != nil {
		return vf.errorValue(fmt.Sprintf("security check failed: %v", err))
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Execute raw query
	sqlDB := db.DB

	rawRows, err := sqlDB.QueryContext(ctx, sqlQuery)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("query execution failed: %v", err))
	}
	defer func() { _ = rawRows.Close() }()

	// Get column names
	columns, err := rawRows.Columns()
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to get columns: %v", err))
	}

	// Scan all rows
	var rows []map[string]interface{}
	for rawRows.Next() {
		// Create slice to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rawRows.Scan(valuePtrs...); err != nil {
			return vf.errorValue(fmt.Sprintf("failed to scan row: %v", err))
		}

		// Convert to map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle byte slices (common for TEXT columns)
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		rows = append(rows, row)

		// Limit results
		if len(rows) >= 1000 {
			break
		}
	}

	if err := rawRows.Err(); err != nil {
		return vf.errorValue(fmt.Sprintf("row iteration error: %v", err))
	}

	// Format as markdown
	output := formatAsMarkdownTable(rows, columns)

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to create dest directory: %v", err))
	}

	// Write to file
	if err := os.WriteFile(destFile, []byte(output), 0644); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to write file: %v", err))
	}

	logger.Get().Debug("dbSelectToFile completed",
		zap.String("dest", destFile),
		zap.Int("rows", len(rows)))

	return vf.vm.ToValue(true)
}

// dbSelectToJSONL executes a SELECT query and writes JSONL output with specified fields to a file
// Usage: db_select_to_jsonl(sql_query, fields, dest_file) -> bool
// fields: comma-separated field names to include in output (e.g., "host,port,status")
func (vf *vmFunc) dbSelectToJSONL(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbSelectToJSONL"))

	if len(call.Arguments) < 3 {
		return vf.errorValue("db_select_to_jsonl requires 3 arguments: sql_query, fields, dest_file")
	}

	sqlQuery := call.Argument(0).String()
	fieldsArg := call.Argument(1).String()
	destFile := call.Argument(2).String()

	if sqlQuery == "" || sqlQuery == "undefined" {
		return vf.errorValue("sql_query cannot be empty")
	}

	if destFile == "" || destFile == "undefined" {
		return vf.errorValue("dest_file cannot be empty")
	}

	// Parse fields (comma-separated)
	var fields []string
	if fieldsArg != "" && fieldsArg != "undefined" {
		for _, f := range strings.Split(fieldsArg, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				fields = append(fields, f)
			}
		}
	}

	// Security validation
	if err := validateSelectQuery(sqlQuery); err != nil {
		return vf.errorValue(fmt.Sprintf("security check failed: %v", err))
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()

	// Execute raw query
	sqlDB := db.DB

	rawRows, err := sqlDB.QueryContext(ctx, sqlQuery)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("query execution failed: %v", err))
	}
	defer func() { _ = rawRows.Close() }()

	// Get column names
	columns, err := rawRows.Columns()
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to get columns: %v", err))
	}

	// If no fields specified, use all columns
	if len(fields) == 0 {
		fields = columns
	}

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to create dest directory: %v", err))
	}

	// Create output file
	outFile, err := os.Create(destFile)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to create file: %v", err))
	}
	defer func() { _ = outFile.Close() }()

	writer := bufio.NewWriter(outFile)
	rowCount := 0

	for rawRows.Next() {
		// Create slice to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rawRows.Scan(valuePtrs...); err != nil {
			return vf.errorValue(fmt.Sprintf("failed to scan row: %v", err))
		}

		// Convert to map with only requested fields
		row := make(map[string]interface{})
		for i, col := range columns {
			// Check if this column is in the requested fields
			for _, f := range fields {
				if f == col {
					val := values[i]
					// Handle byte slices (common for TEXT columns)
					if b, ok := val.([]byte); ok {
						row[col] = string(b)
					} else {
						row[col] = val
					}
					break
				}
			}
		}

		// Write as JSON line
		jsonBytes, err := json.Marshal(row)
		if err != nil {
			continue
		}
		_, _ = writer.Write(jsonBytes)
		_, _ = writer.WriteString("\n")
		rowCount++

		// Limit results
		if rowCount >= 1000 {
			break
		}
	}

	if err := rawRows.Err(); err != nil {
		return vf.errorValue(fmt.Sprintf("row iteration error: %v", err))
	}

	if err := writer.Flush(); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to flush file: %v", err))
	}

	logger.Get().Debug("dbSelectToJSONL completed",
		zap.String("dest", destFile),
		zap.Int("rows", rowCount),
		zap.Strings("fields", fields))

	return vf.vm.ToValue(true)
}

// dbImportAssetFromFile imports assets from a JSONL file (httpx format)
// Usage: db_import_asset_from_file(workspace, file_path) -> map with stats {new, updated, unchanged, errors, total}
func (vf *vmFunc) dbImportAssetFromFile(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbImportAssetFromFile"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_import_asset_from_file requires 2 arguments: workspace, file_path")
	}

	workspace := call.Argument(0).String()
	filePath := call.Argument(1).String()

	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}

	if filePath == "" || filePath == "undefined" {
		return vf.errorValue("file_path cannot be empty")
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to open file: %v", err))
	}
	defer func() { _ = file.Close() }()

	ctx := context.Background()
	stats := database.ImportStats{}

	// Use scanner with larger buffer for large JSONL files
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 10*1024*1024) // 10MB buffer
	scanner.Buffer(buf, 10*1024*1024)

	now := time.Now()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse JSON into map
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			logger.Get().Debug("skipping invalid JSON line", zap.Error(err))
			stats.Errors++
			continue
		}

		// Map httpx fields to Asset model
		asset := mapJSONToAsset(data, workspace, line)
		asset.LastSeenAt = now

		// Check if asset already exists
		var existing database.Asset
		selectErr := db.NewSelect().Model(&existing).
			Where("workspace = ?", workspace).
			Where("asset_value = ?", asset.AssetValue).
			Where("url = ?", asset.URL).
			Scan(ctx)

		if selectErr != nil {
			// New asset - insert
			asset.CreatedAt = now
			asset.UpdatedAt = now
			_, insertErr := db.NewInsert().Model(&asset).Exec(ctx)
			if insertErr != nil {
				logger.Get().Debug("failed to insert asset", zap.Error(insertErr))
				stats.Errors++
				continue
			}
			stats.New++
		} else if hasAssetChanged(&existing, &asset) {
			// Changed - full update
			asset.ID = existing.ID
			asset.CreatedAt = existing.CreatedAt
			asset.UpdatedAt = now
			_, updateErr := db.NewUpdate().Model(&asset).WherePK().Exec(ctx)
			if updateErr != nil {
				logger.Get().Debug("failed to update asset", zap.Error(updateErr))
				stats.Errors++
				continue
			}
			stats.Updated++
		} else {
			// Unchanged - only update last_seen_at
			_, updateErr := db.NewUpdate().Model((*database.Asset)(nil)).
				Set("last_seen_at = ?", now).
				Where("id = ?", existing.ID).
				Exec(ctx)
			if updateErr != nil {
				logger.Get().Debug("failed to update last_seen_at", zap.Error(updateErr))
				stats.Errors++
				continue
			}
			stats.Unchanged++
		}
	}

	if err := scanner.Err(); err != nil {
		return vf.errorValue(fmt.Sprintf("error reading file: %v", err))
	}

	total := stats.New + stats.Updated + stats.Unchanged
	logger.Get().Debug("dbImportAssetFromFile completed",
		zap.String("workspace", workspace),
		zap.String("file", filePath),
		zap.Int("new", stats.New),
		zap.Int("updated", stats.Updated),
		zap.Int("unchanged", stats.Unchanged),
		zap.Int("errors", stats.Errors),
		zap.Int("total", total))

	// Return stats as a map
	return vf.vm.ToValue(map[string]interface{}{
		"new":       stats.New,
		"updated":   stats.Updated,
		"unchanged": stats.Unchanged,
		"errors":    stats.Errors,
		"total":     total,
	})
}

// hasAssetChanged compares two assets for meaningful changes
func hasAssetChanged(existing, new *database.Asset) bool {
	return existing.StatusCode != new.StatusCode ||
		existing.Title != new.Title ||
		existing.ContentType != new.ContentType ||
		existing.ContentLength != new.ContentLength ||
		existing.HostIP != new.HostIP ||
		existing.TLS != new.TLS ||
		existing.Words != new.Words ||
		existing.Lines != new.Lines ||
		!slicesEqual(existing.Technologies, new.Technologies)
}

// slicesEqual compares two string slices for equality
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// mapJSONToAsset maps httpx JSON fields to Asset model
func mapJSONToAsset(data map[string]interface{}, workspace, rawLine string) database.Asset {
	asset := database.Asset{
		Workspace:   workspace,
		RawJsonData: rawLine,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Map httpx fields to Asset fields
	if v, ok := data["host"].(string); ok {
		asset.AssetValue = v
	}
	if v, ok := data["url"].(string); ok {
		asset.URL = v
	}
	if v, ok := data["input"].(string); ok {
		asset.Input = v
	}
	if v, ok := data["scheme"].(string); ok {
		asset.Scheme = v
	}
	if v, ok := data["method"].(string); ok {
		asset.Method = v
	}
	if v, ok := data["path"].(string); ok {
		asset.Path = v
	}
	if v, ok := data["status_code"].(float64); ok {
		asset.StatusCode = int(v)
	}
	if v, ok := data["content_type"].(string); ok {
		asset.ContentType = v
	}
	if v, ok := data["content_length"].(float64); ok {
		asset.ContentLength = int64(v)
	}
	if v, ok := data["title"].(string); ok {
		asset.Title = v
	}
	if v, ok := data["words"].(float64); ok {
		asset.Words = int(v)
	}
	if v, ok := data["lines"].(float64); ok {
		asset.Lines = int(v)
	}
	if v, ok := data["host_ip"].(string); ok {
		asset.HostIP = v
	}

	// DNS A records
	if aRecords, ok := data["a"].([]interface{}); ok {
		var records []string
		for _, r := range aRecords {
			if s, ok := r.(string); ok {
				records = append(records, s)
			}
		}
		asset.DnsRecords = records
	}

	// Technologies
	if tech, ok := data["tech"].([]interface{}); ok {
		var technologies []string
		for _, t := range tech {
			if s, ok := t.(string); ok {
				technologies = append(technologies, s)
			}
		}
		asset.Technologies = technologies
	}

	// Response time
	if v, ok := data["time"].(string); ok {
		asset.ResponseTime = v
	}

	// TLS info - serialize if object
	if tls, ok := data["tls"]; ok {
		switch t := tls.(type) {
		case string:
			asset.TLS = t
		case map[string]interface{}:
			if tlsBytes, err := json.Marshal(t); err == nil {
				asset.TLS = string(tlsBytes)
			}
		}
	}

	// Webserver as source
	if v, ok := data["webserver"].(string); ok {
		asset.Source = v
	}

	return asset
}

// dbImportVuln imports a single vulnerability from JSON data (nuclei format)
// Usage: db_import_vuln(workspace, json_data) -> bool
func (vf *vmFunc) dbImportVuln(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbImportVuln"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_import_vuln requires 2 arguments: workspace, json_data")
	}

	workspace := call.Argument(0).String()
	jsonData := call.Argument(1).String()

	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}

	if jsonData == "" || jsonData == "undefined" {
		return vf.errorValue("json_data cannot be empty")
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	// Parse JSON into map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return vf.errorValue(fmt.Sprintf("invalid JSON: %v", err))
	}

	// Map nuclei fields to Vulnerability model
	vuln := mapJSONToVuln(data, workspace, jsonData)

	ctx := context.Background()

	// Check if vulnerability already exists
	var existing database.Vulnerability
	err := db.NewSelect().Model(&existing).
		Where("workspace = ?", workspace).
		Where("vuln_info = ?", vuln.VulnInfo).
		Where("asset_value = ?", vuln.AssetValue).
		Scan(ctx)

	if err == nil {
		// Vulnerability exists, update it
		vuln.ID = existing.ID
		vuln.CreatedAt = existing.CreatedAt
		_, err = db.NewUpdate().Model(&vuln).WherePK().Exec(ctx)
	} else {
		// Insert new vulnerability
		_, err = db.NewInsert().Model(&vuln).Exec(ctx)
	}

	if err != nil {
		return vf.errorValue(fmt.Sprintf("import failed: %v", err))
	}

	return vf.vm.ToValue(true)
}

// dbImportVulnFromFile imports vulnerabilities from a JSONL file (nuclei format)
// Usage: db_import_vuln_from_file(workspace, file_path) -> map with stats {new, updated, unchanged, errors, total}
func (vf *vmFunc) dbImportVulnFromFile(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbImportVulnFromFile"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_import_vuln_from_file requires 2 arguments: workspace, file_path")
	}

	workspace := call.Argument(0).String()
	filePath := call.Argument(1).String()

	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}

	if filePath == "" || filePath == "undefined" {
		return vf.errorValue("file_path cannot be empty")
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to open file: %v", err))
	}
	defer func() { _ = file.Close() }()

	ctx := context.Background()
	stats := database.ImportStats{}

	// Use scanner with larger buffer for large JSONL files
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 10*1024*1024) // 10MB buffer
	scanner.Buffer(buf, 10*1024*1024)

	now := time.Now()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse JSON into map
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			logger.Get().Debug("skipping invalid JSON line", zap.Error(err))
			stats.Errors++
			continue
		}

		// Map nuclei fields to Vulnerability model
		vuln := mapJSONToVuln(data, workspace, line)
		vuln.LastSeenAt = now

		// Check if vulnerability already exists
		var existing database.Vulnerability
		selectErr := db.NewSelect().Model(&existing).
			Where("workspace = ?", workspace).
			Where("vuln_info = ?", vuln.VulnInfo).
			Where("asset_value = ?", vuln.AssetValue).
			Scan(ctx)

		if selectErr != nil {
			// New vulnerability - insert
			vuln.CreatedAt = now
			vuln.UpdatedAt = now
			_, insertErr := db.NewInsert().Model(&vuln).Exec(ctx)
			if insertErr != nil {
				logger.Get().Debug("failed to insert vulnerability", zap.Error(insertErr))
				stats.Errors++
				continue
			}
			stats.New++
		} else if hasVulnChanged(&existing, &vuln) {
			// Changed - full update
			vuln.ID = existing.ID
			vuln.CreatedAt = existing.CreatedAt
			vuln.UpdatedAt = now
			_, updateErr := db.NewUpdate().Model(&vuln).WherePK().Exec(ctx)
			if updateErr != nil {
				logger.Get().Debug("failed to update vulnerability", zap.Error(updateErr))
				stats.Errors++
				continue
			}
			stats.Updated++
		} else {
			// Unchanged - only update last_seen_at
			_, updateErr := db.NewUpdate().Model((*database.Vulnerability)(nil)).
				Set("last_seen_at = ?", now).
				Where("id = ?", existing.ID).
				Exec(ctx)
			if updateErr != nil {
				logger.Get().Debug("failed to update last_seen_at", zap.Error(updateErr))
				stats.Errors++
				continue
			}
			stats.Unchanged++
		}
	}

	if err := scanner.Err(); err != nil {
		return vf.errorValue(fmt.Sprintf("error reading file: %v", err))
	}

	total := stats.New + stats.Updated + stats.Unchanged
	logger.Get().Debug("dbImportVulnFromFile completed",
		zap.String("workspace", workspace),
		zap.String("file", filePath),
		zap.Int("new", stats.New),
		zap.Int("updated", stats.Updated),
		zap.Int("unchanged", stats.Unchanged),
		zap.Int("errors", stats.Errors),
		zap.Int("total", total))

	// Return stats as a map
	return vf.vm.ToValue(map[string]interface{}{
		"new":       stats.New,
		"updated":   stats.Updated,
		"unchanged": stats.Unchanged,
		"errors":    stats.Errors,
		"total":     total,
	})
}

// hasVulnChanged compares two vulnerabilities for meaningful changes
func hasVulnChanged(existing, new *database.Vulnerability) bool {
	return existing.Severity != new.Severity ||
		existing.VulnTitle != new.VulnTitle ||
		existing.VulnDesc != new.VulnDesc ||
		existing.Confidence != new.Confidence ||
		!slicesEqual(existing.Tags, new.Tags)
}

// mapJSONToVuln maps nuclei JSON fields to Vulnerability model
func mapJSONToVuln(data map[string]interface{}, workspace, rawLine string) database.Vulnerability {
	vuln := database.Vulnerability{
		Workspace:   workspace,
		RawVulnJSON: rawLine,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// template-id -> VulnInfo
	if v, ok := data["template-id"].(string); ok {
		vuln.VulnInfo = v
	}

	// Nested info fields
	if info, ok := data["info"].(map[string]interface{}); ok {
		if name, ok := info["name"].(string); ok {
			vuln.VulnTitle = name
		}
		if desc, ok := info["description"].(string); ok {
			vuln.VulnDesc = desc
		}
		if severity, ok := info["severity"].(string); ok {
			vuln.Severity = severity
		}
		// Tags can be array or comma-separated string
		if tags, ok := info["tags"].([]interface{}); ok {
			var tagList []string
			for _, t := range tags {
				if s, ok := t.(string); ok {
					tagList = append(tagList, s)
				}
			}
			vuln.Tags = tagList
		} else if tagsStr, ok := info["tags"].(string); ok {
			vuln.Tags = strings.Split(tagsStr, ",")
		}
	}

	// host -> AssetValue
	if v, ok := data["host"].(string); ok {
		vuln.AssetValue = v
	}

	// matched-at as fallback for AssetValue
	if vuln.AssetValue == "" {
		if v, ok := data["matched-at"].(string); ok {
			vuln.AssetValue = v
		}
	}

	// type -> AssetType
	if v, ok := data["type"].(string); ok {
		vuln.AssetType = v
	}

	// request -> DetailHTTPRequest
	if v, ok := data["request"].(string); ok {
		vuln.DetailHTTPRequest = v
	}

	// response -> DetailHTTPResponse
	if v, ok := data["response"].(string); ok {
		vuln.DetailHTTPResponse = v
	}

	return vuln
}

// getAssetDiffInternal is a helper that retrieves asset diff data
func (vf *vmFunc) getAssetDiffInternal(workspace string) (*database.AssetDiff, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	ctx := context.Background()

	// Get workspace to determine fromTime
	var fromTime time.Time
	ws, err := database.GetWorkspaceByName(ctx, workspace)
	if err == nil && ws != nil {
		fromTime = ws.CreatedAt
	} else {
		// Fallback: use oldest asset's created_at or 24 hours ago
		var oldestAsset database.Asset
		err := db.NewSelect().Model(&oldestAsset).
			Where("workspace = ?", workspace).
			Order("created_at ASC").
			Limit(1).
			Scan(ctx)
		if err == nil {
			fromTime = oldestAsset.CreatedAt
		} else {
			// Default to 24 hours ago if no data
			fromTime = time.Now().Add(-24 * time.Hour)
		}
	}

	toTime := time.Now()

	logger.Get().Debug(terminal.HiGreen("db_asset_diff")+" params",
		zap.String("workspace", workspace),
		zap.Time("from_time", fromTime),
		zap.Time("to_time", toTime))

	return database.GetAssetDiff(ctx, workspace, fromTime, toTime)
}

// assetDiffToJSONL converts asset diff to JSONL format string
func assetDiffToJSONL(diff *database.AssetDiff) string {
	var sb strings.Builder

	// Write added assets
	for _, asset := range diff.Added {
		row := assetToMap(&asset)
		row["diff_type"] = "added"
		row["workspace_name"] = diff.WorkspaceName
		if jsonBytes, err := json.Marshal(row); err == nil {
			sb.Write(jsonBytes)
			sb.WriteString("\n")
		}
	}

	// Write removed assets
	for _, asset := range diff.Removed {
		row := assetToMap(&asset)
		row["diff_type"] = "removed"
		row["workspace_name"] = diff.WorkspaceName
		if jsonBytes, err := json.Marshal(row); err == nil {
			sb.Write(jsonBytes)
			sb.WriteString("\n")
		}
	}

	// Write changed assets
	for _, change := range diff.Changed {
		row := map[string]interface{}{
			"diff_type":      "changed",
			"workspace_name": diff.WorkspaceName,
			"asset_id":       change.AssetID,
			"asset_value":    change.AssetValue,
			"url":            change.URL,
			"changes":        fieldChangeSliceToMaps(change.Changes),
		}
		if jsonBytes, err := json.Marshal(row); err == nil {
			sb.Write(jsonBytes)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// dbAssetDiff gets asset diff with auto-populated time range
// Usage: db_asset_diff(workspace) -> string (JSONL format)
// workspace: workspace name
// Returns: JSONL string with each line containing diff_type (added/removed/changed)
func (vf *vmFunc) dbAssetDiff(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_asset_diff"))

	if len(call.Arguments) < 1 {
		return vf.errorValue("db_asset_diff requires 1 argument: workspace")
	}

	workspace := call.Argument(0).String()
	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}

	diff, err := vf.getAssetDiffInternal(workspace)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to get asset diff: %v", err))
	}

	return vf.vm.ToValue(assetDiffToJSONL(diff))
}

// dbAssetDiffToFile gets asset diff and writes to file
// Usage: db_asset_diff_to_file(workspace, dest) -> bool
// workspace: workspace name
// dest: output file path
func (vf *vmFunc) dbAssetDiffToFile(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_asset_diff_to_file"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_asset_diff_to_file requires 2 arguments: workspace, dest")
	}

	workspace := call.Argument(0).String()
	dest := call.Argument(1).String()

	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}
	if dest == "" || dest == "undefined" {
		return vf.errorValue("dest cannot be empty")
	}

	diff, err := vf.getAssetDiffInternal(workspace)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to get asset diff: %v", err))
	}

	jsonl := assetDiffToJSONL(diff)

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to create dest directory: %v", err))
	}

	// Write to file
	if err := os.WriteFile(dest, []byte(jsonl), 0644); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to write file: %v", err))
	}

	logger.Get().Debug("db_asset_diff_to_file completed",
		zap.String("dest", dest),
		zap.Int("added", diff.Summary.TotalAdded),
		zap.Int("removed", diff.Summary.TotalRemoved),
		zap.Int("changed", diff.Summary.TotalChanged))

	return vf.vm.ToValue(true)
}

// getVulnDiffInternal is a helper that retrieves vulnerability diff data
func (vf *vmFunc) getVulnDiffInternal(workspace string) (*database.VulnerabilityDiff, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	ctx := context.Background()

	// Get workspace to determine fromTime
	var fromTime time.Time
	ws, err := database.GetWorkspaceByName(ctx, workspace)
	if err == nil && ws != nil {
		fromTime = ws.CreatedAt
	} else {
		// Fallback: use oldest vulnerability's created_at or 24 hours ago
		var oldestVuln database.Vulnerability
		err := db.NewSelect().Model(&oldestVuln).
			Where("workspace = ?", workspace).
			Order("created_at ASC").
			Limit(1).
			Scan(ctx)
		if err == nil {
			fromTime = oldestVuln.CreatedAt
		} else {
			// Default to 24 hours ago if no data
			fromTime = time.Now().Add(-24 * time.Hour)
		}
	}

	toTime := time.Now()

	logger.Get().Debug(terminal.HiGreen("db_vuln_diff")+" params",
		zap.String("workspace", workspace),
		zap.Time("from_time", fromTime),
		zap.Time("to_time", toTime))

	return database.GetVulnerabilityDiff(ctx, workspace, fromTime, toTime)
}

// vulnDiffToJSONL converts vulnerability diff to JSONL format string
func vulnDiffToJSONL(diff *database.VulnerabilityDiff) string {
	var sb strings.Builder

	// Write added vulnerabilities
	for _, vuln := range diff.Added {
		row := vulnerabilityToMap(&vuln)
		row["diff_type"] = "added"
		row["workspace_name"] = diff.WorkspaceName
		if jsonBytes, err := json.Marshal(row); err == nil {
			sb.Write(jsonBytes)
			sb.WriteString("\n")
		}
	}

	// Write removed vulnerabilities
	for _, vuln := range diff.Removed {
		row := vulnerabilityToMap(&vuln)
		row["diff_type"] = "removed"
		row["workspace_name"] = diff.WorkspaceName
		if jsonBytes, err := json.Marshal(row); err == nil {
			sb.Write(jsonBytes)
			sb.WriteString("\n")
		}
	}

	// Write changed vulnerabilities
	for _, change := range diff.Changed {
		row := map[string]interface{}{
			"diff_type":      "changed",
			"workspace_name": diff.WorkspaceName,
			"vuln_id":        change.VulnID,
			"vuln_info":      change.VulnInfo,
			"asset_value":    change.AssetValue,
			"changes":        fieldChangeSliceToMaps(change.Changes),
		}
		if jsonBytes, err := json.Marshal(row); err == nil {
			sb.Write(jsonBytes)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// dbVulnDiff gets vulnerability diff with auto-populated time range
// Usage: db_vuln_diff(workspace) -> string (JSONL format)
// workspace: workspace name
// Returns: JSONL string with each line containing diff_type (added/removed/changed)
func (vf *vmFunc) dbVulnDiff(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_vuln_diff"))

	if len(call.Arguments) < 1 {
		return vf.errorValue("db_vuln_diff requires 1 argument: workspace")
	}

	workspace := call.Argument(0).String()
	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}

	diff, err := vf.getVulnDiffInternal(workspace)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to get vulnerability diff: %v", err))
	}

	return vf.vm.ToValue(vulnDiffToJSONL(diff))
}

// dbVulnDiffToFile gets vulnerability diff and writes to file
// Usage: db_vuln_diff_to_file(workspace, dest) -> bool
// workspace: workspace name
// dest: output file path
func (vf *vmFunc) dbVulnDiffToFile(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("db_vuln_diff_to_file"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_vuln_diff_to_file requires 2 arguments: workspace, dest")
	}

	workspace := call.Argument(0).String()
	dest := call.Argument(1).String()

	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}
	if dest == "" || dest == "undefined" {
		return vf.errorValue("dest cannot be empty")
	}

	diff, err := vf.getVulnDiffInternal(workspace)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to get vulnerability diff: %v", err))
	}

	jsonl := vulnDiffToJSONL(diff)

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to create dest directory: %v", err))
	}

	// Write to file
	if err := os.WriteFile(dest, []byte(jsonl), 0644); err != nil {
		return vf.errorValue(fmt.Sprintf("failed to write file: %v", err))
	}

	logger.Get().Debug("db_vuln_diff_to_file completed",
		zap.String("dest", dest),
		zap.Int("added", diff.Summary.TotalAdded),
		zap.Int("removed", diff.Summary.TotalRemoved),
		zap.Int("changed", diff.Summary.TotalChanged))

	return vf.vm.ToValue(true)
}

// fieldChangeSliceToMaps converts a slice of FieldChanges to slice of maps
func fieldChangeSliceToMaps(changes []database.FieldChange) []map[string]interface{} {
	result := make([]map[string]interface{}, len(changes))
	for i, change := range changes {
		result[i] = map[string]interface{}{
			"field":     change.Field,
			"old_value": change.OldValue,
			"new_value": change.NewValue,
		}
	}
	return result
}

// dbSelectRuns queries run records by workspace
// Usage: run_status(workspace, format) -> string
func (vf *vmFunc) dbSelectRuns(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen(FnDBSelectRuns))

	if len(call.Arguments) < 2 {
		return vf.errorValue(FnDBSelectRuns + " requires 2 arguments: workspace, format")
	}

	workspace := call.Argument(0).String()
	format := call.Argument(1).String()

	if workspace == "undefined" {
		workspace = ""
	}

	if err := validateOutputFormat(format); err != nil {
		return vf.errorValue(err.Error())
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()
	var runs []*database.Run
	query := db.NewSelect().Model(&runs).Order("started_at DESC")

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}

	query = query.Limit(1000)

	if err := query.Scan(ctx); err != nil {
		return vf.errorValue(fmt.Sprintf("query failed: %v", err))
	}

	rows := make([]map[string]interface{}, len(runs))
	for i, run := range runs {
		rows[i] = runToMap(run)
	}

	var output string
	switch format {
	case "markdown":
		output = formatAsMarkdownTable(rows, runColumns)
	case "jsonl":
		output = formatAsJSONL(rows)
	}

	return vf.vm.ToValue(output)
}

// dbSelectRunByUUID queries a run record by UUID
// Usage: run_status_by_uuid(uuid, format) -> string
func (vf *vmFunc) dbSelectRunByUUID(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen(FnDBSelectRunByUUID))

	if len(call.Arguments) < 2 {
		return vf.errorValue(FnDBSelectRunByUUID + " requires 2 arguments: uuid, format")
	}

	runUUID := call.Argument(0).String()
	format := call.Argument(1).String()

	if runUUID == "" || runUUID == "undefined" {
		return vf.errorValue("uuid argument is required")
	}

	if err := validateOutputFormat(format); err != nil {
		return vf.errorValue(err.Error())
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	ctx := context.Background()
	var run database.Run
	err := db.NewSelect().Model(&run).Where("run_uuid = ?", runUUID).Scan(ctx)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("run not found: %v", err))
	}

	rows := []map[string]interface{}{runToMap(&run)}

	var output string
	switch format {
	case "markdown":
		output = formatAsMarkdownTable(rows, runColumns)
	case "jsonl":
		output = formatAsJSONL(rows)
	}

	return vf.vm.ToValue(output)
}

// dbResetEventLogs resets the processed status of event logs from true to false
// Usage: db_reset_event_logs(workspace?, topic_pattern?) -> {reset: int, total: int}
func (vf *vmFunc) dbResetEventLogs(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen(FnDBResetEventLogs))

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	// Extract optional parameters
	workspace := ""
	topicPattern := ""
	if len(call.Arguments) >= 1 {
		workspace = call.Argument(0).String()
		if workspace == "undefined" {
			workspace = ""
		}
	}
	if len(call.Arguments) >= 2 {
		topicPattern = call.Argument(1).String()
		if topicPattern == "undefined" {
			topicPattern = ""
		}
	}

	ctx := context.Background()

	// Build query to count matching processed=true events
	countQuery := db.NewSelect().
		Model((*database.EventLog)(nil)).
		Where("processed = ?", true)

	if workspace != "" {
		countQuery = countQuery.Where("workspace = ?", workspace)
	}
	if topicPattern != "" {
		likePattern := globToSQLLike(topicPattern)
		countQuery = countQuery.Where("topic LIKE ?", likePattern)
	}

	total, err := countQuery.Count(ctx)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to count events: %v", err))
	}

	// Build update query
	updateQuery := db.NewUpdate().
		Model((*database.EventLog)(nil)).
		Set("processed = ?", false).
		Set("processed_at = NULL").
		Where("processed = ?", true)

	if workspace != "" {
		updateQuery = updateQuery.Where("workspace = ?", workspace)
	}
	if topicPattern != "" {
		likePattern := globToSQLLike(topicPattern)
		updateQuery = updateQuery.Where("topic LIKE ?", likePattern)
	}

	result, err := updateQuery.Exec(ctx)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to reset events: %v", err))
	}

	resetCount, _ := result.RowsAffected()

	return vf.vm.ToValue(map[string]interface{}{
		"reset": resetCount,
		"total": total,
	})
}

// globToSQLLike converts a glob pattern to SQL LIKE pattern
// * -> % (match any characters)
// ? -> _ (match single character)
func globToSQLLike(pattern string) string {
	var sb strings.Builder
	for _, c := range pattern {
		switch c {
		case '*':
			sb.WriteRune('%')
		case '?':
			sb.WriteRune('_')
		case '%', '_':
			// Escape SQL LIKE special chars
			sb.WriteRune('\\')
			sb.WriteRune(c)
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

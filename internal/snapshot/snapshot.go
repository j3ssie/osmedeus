package snapshot

import (
	"archive/zip"
	"compress/flate"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/state"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
)

// ExportResult contains the result of a workspace export
type ExportResult struct {
	WorkspaceName string
	SourcePath    string
	OutputPath    string
	FileSize      int64
}

// ImportResult contains the result of a workspace import
type ImportResult struct {
	WorkspaceName string
	LocalPath     string
	DataSource    string
	FilesCount    int
}

// ExportWorkspace creates a compressed zip of the workspace folder
func ExportWorkspace(workspacePath, outputPath string) (*ExportResult, error) {
	// Validate source exists
	info, err := os.Stat(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace path is not a directory: %s", workspacePath)
	}

	// Extract workspace name from path
	workspaceName := filepath.Base(workspacePath)

	// Generate output path if not specified
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s_%s.zip", workspaceName, time.Now().UTC().Format("2006-01-02T15-04-05Z"))
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the zip archive with highest compression
	if err := createHighCompressionZip(workspacePath, outputPath); err != nil {
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}

	// Get file size
	zipInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get archive info: %w", err)
	}

	return &ExportResult{
		WorkspaceName: workspaceName,
		SourcePath:    workspacePath,
		OutputPath:    outputPath,
		FileSize:      zipInfo.Size(),
	}, nil
}

// ImportWorkspace extracts a zip file and optionally imports to database
func ImportWorkspace(source, workspacesPath string, skipDB bool, cfg *config.Config) (*ImportResult, error) {
	var zipPath string
	var cleanup func()

	// Check if source is URL or local file
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Download the file
		tempPath, err := downloadFile(source)
		if err != nil {
			return nil, fmt.Errorf("failed to download: %w", err)
		}
		zipPath = tempPath
		cleanup = func() { _ = os.Remove(tempPath) }
	} else {
		// Use local file
		if _, err := os.Stat(source); err != nil {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		zipPath = source
		cleanup = func() {} // No cleanup needed for local files
	}
	defer cleanup()

	// Extract workspace name from zip filename
	zipName := filepath.Base(zipPath)
	workspaceName := strings.TrimSuffix(zipName, ".zip")
	// Remove timestamp suffix if present (e.g., "example.com_2026-02-13T18-20-34Z" or "example.com_1234567890" -> "example.com")
	if idx := strings.LastIndex(workspaceName, "_"); idx > 0 {
		potentialTimestamp := workspaceName[idx+1:]
		if isTimestampSuffix(potentialTimestamp) {
			workspaceName = workspaceName[:idx]
		}
	}

	// Destination path
	destPath := filepath.Join(workspacesPath, workspaceName)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return nil, fmt.Errorf("workspace already exists: %s (use --force to overwrite)", destPath)
	}

	// Extract the zip into the workspaces root — zip entries already contain
	// the workspace directory prefix (e.g. "shopee.vn/file.txt"), so extracting
	// into workspacesPath produces workspacesPath/shopee.vn/file.txt.
	filesCount, err := extractZip(zipPath, workspacesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	result := &ImportResult{
		WorkspaceName: workspaceName,
		LocalPath:     destPath,
		DataSource:    "imported",
		FilesCount:    filesCount,
	}

	// Import to database if not skipped
	if !skipDB && cfg != nil {
		if err := ImportWorkspaceToDB(context.Background(), workspaceName, destPath, cfg); err != nil {
			// Log warning but don't fail the import
			fmt.Printf("%s Failed to import to database: %v\n", terminal.Yellow(terminal.SymbolWarning), err)
		}
		if err := replayDBOperations(destPath, workspaceName, cfg); err != nil {
			fmt.Printf("%s Failed to replay database operations: %v\n", terminal.Yellow(terminal.SymbolWarning), err)
		}
	}

	return result, nil
}

// ImportWorkspaceToDB creates or updates a workspace record with data_source="imported"
func ImportWorkspaceToDB(ctx context.Context, workspaceName, localPath string, cfg *config.Config) error {
	// Connect to database
	_, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ensure tables exist
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create or update workspace record
	workspace := &database.Workspace{
		Name:       workspaceName,
		LocalPath:  localPath,
		DataSource: "imported",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return database.UpsertWorkspace(ctx, workspace)
}

// ForceImportWorkspace imports a workspace, overwriting if it exists
func ForceImportWorkspace(source, workspacesPath string, skipDB bool, cfg *config.Config) (*ImportResult, error) {
	var zipPath string
	var cleanup func()

	// Check if source is URL or local file
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		tempPath, err := downloadFile(source)
		if err != nil {
			return nil, fmt.Errorf("failed to download: %w", err)
		}
		zipPath = tempPath
		cleanup = func() { _ = os.Remove(tempPath) }
	} else {
		if _, err := os.Stat(source); err != nil {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		zipPath = source
		cleanup = func() {}
	}
	defer cleanup()

	// Extract workspace name
	zipName := filepath.Base(zipPath)
	workspaceName := strings.TrimSuffix(zipName, ".zip")
	if idx := strings.LastIndex(workspaceName, "_"); idx > 0 {
		potentialTimestamp := workspaceName[idx+1:]
		if isTimestampSuffix(potentialTimestamp) {
			workspaceName = workspaceName[:idx]
		}
	}

	destPath := filepath.Join(workspacesPath, workspaceName)

	// Remove existing if present
	if _, err := os.Stat(destPath); err == nil {
		if err := os.RemoveAll(destPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing workspace: %w", err)
		}
	}

	// Extract the zip into the workspaces root — zip entries already contain
	// the workspace directory prefix (e.g. "shopee.vn/file.txt"), so extracting
	// into workspacesPath produces workspacesPath/shopee.vn/file.txt.
	filesCount, err := extractZip(zipPath, workspacesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	result := &ImportResult{
		WorkspaceName: workspaceName,
		LocalPath:     destPath,
		DataSource:    "imported",
		FilesCount:    filesCount,
	}

	// Import to database if not skipped
	if !skipDB && cfg != nil {
		if err := ImportWorkspaceToDB(context.Background(), workspaceName, destPath, cfg); err != nil {
			fmt.Printf("%s Failed to import to database: %v\n", terminal.Yellow(terminal.SymbolWarning), err)
		}
		if err := replayDBOperations(destPath, workspaceName, cfg); err != nil {
			fmt.Printf("%s Failed to replay database operations: %v\n", terminal.Yellow(terminal.SymbolWarning), err)
		}
	}

	return result, nil
}

// dbFunctionEntry holds a db_ function expression collected from a workflow step
type dbFunctionEntry struct {
	expr         string
	preCondition string
	stepName     string
}

// collectDBFunctionsFromSteps walks steps and collects all db_ function expressions
func collectDBFunctionsFromSteps(steps []core.Step) []dbFunctionEntry {
	var entries []dbFunctionEntry

	for i := range steps {
		step := &steps[i]

		// Collect from function steps
		if step.Type == core.StepTypeFunction {
			for _, expr := range step.GetFunctions() {
				if strings.Contains(expr, "db_") {
					entries = append(entries, dbFunctionEntry{
						expr:         expr,
						preCondition: step.PreCondition,
						stepName:     step.Name,
					})
				}
			}
			for _, expr := range step.ParallelFunctions {
				if strings.Contains(expr, "db_") {
					entries = append(entries, dbFunctionEntry{
						expr:         expr,
						preCondition: step.PreCondition,
						stepName:     step.Name,
					})
				}
			}
		}

		// Recurse into parallel-steps
		if step.Type == core.StepTypeParallel && len(step.ParallelSteps) > 0 {
			entries = append(entries, collectDBFunctionsFromSteps(step.ParallelSteps)...)
		}

		// Recurse into foreach inner step
		if step.Type == core.StepTypeForeach && step.Step != nil {
			inner := collectDBFunctionsFromSteps([]core.Step{*step.Step})
			entries = append(entries, inner...)
		}
	}

	return entries
}

// collectDBFunctionsFromFlow collects db_ functions from a flow's modules
func collectDBFunctionsFromFlow(flow *core.Workflow, preloadedModules map[string]*core.Workflow, loader *parser.Loader) []dbFunctionEntry {
	var entries []dbFunctionEntry

	for i := range flow.Modules {
		modRef := &flow.Modules[i]

		if modRef.IsInline() {
			// Inline module — use its steps directly
			inlineWf := modRef.ToWorkflow()
			if inlineWf != nil {
				entries = append(entries, collectDBFunctionsFromSteps(inlineWf.Steps)...)
			}
			continue
		}

		// External module reference
		var mod *core.Workflow
		if preloadedModules != nil {
			mod = preloadedModules[modRef.Name]
		}
		if mod == nil && loader != nil {
			mod, _ = loader.LoadWorkflow(modRef.Name)
		}
		if mod != nil {
			entries = append(entries, collectDBFunctionsFromSteps(mod.Steps)...)
		}
	}

	return entries
}

// buildReplayContext builds the template variable map for replaying db_ functions.
// It delegates to executor.BuildBuiltinVariables to produce the same ~46+ variables
// that normal workflow execution computes (heuristic target vars, platform vars,
// thread defaults, config paths, etc.), then overrides Output/TargetSpace/Workspace
// with the actual imported workspace path.
func buildReplayContext(destPath, workspaceName string, runInfo *state.RunInfo, cfg *config.Config) map[string]interface{} {
	// Convert runInfo.Params (map[string]any) to map[string]string for BuildBuiltinVariables
	params := make(map[string]string)
	params["target"] = runInfo.Target
	if runInfo.Params != nil {
		for k, v := range runInfo.Params {
			params[k] = fmt.Sprintf("%v", v)
		}
	}

	// Build full built-in variables (same as normal workflow execution)
	vars := executor.BuildBuiltinVariables(cfg, params)

	// Merge any remaining user params that BuildBuiltinVariables doesn't handle
	// (custom workflow params like enableDnsBruteForcing, etc.)
	if runInfo.Params != nil {
		for k, v := range runInfo.Params {
			if _, exists := vars[k]; !exists {
				vars[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Override Output with actual imported workspace path (may differ from computed path)
	vars["Output"] = destPath
	vars["TargetSpace"] = workspaceName
	vars["Workspace"] = workspaceName

	// Clean up temp directory created by BuildBuiltinVariables
	if tempDir, ok := vars["TempDir"].(string); ok && tempDir != "" {
		_ = os.RemoveAll(tempDir)
		delete(vars, "TempDir")
		delete(vars, "TempFile")
	}

	return vars
}

// resolveParamDefaults iterates params and resolves default values into vars.
// Already-set vars take priority (are not overwritten). Each resolved value is
// stored immediately so subsequent params can reference earlier ones (chained defaults).
func resolveParamDefaults(params []core.Param, vars map[string]any, engine *template.Engine) {
	for _, param := range params {
		if _, exists := vars[param.Name]; exists {
			continue // provided params take priority
		}
		if !param.HasDefault() {
			continue
		}
		rendered, err := engine.Render(param.DefaultString(), vars)
		if err != nil {
			continue
		}
		vars[param.Name] = rendered
	}
}

// replayDBOperations reads run-state.json from an imported workspace, loads the
// original workflow, and re-executes all db_ function calls to populate the database.
func replayDBOperations(destPath, workspaceName string, cfg *config.Config) error {
	// 1. Read run-state.json
	fmt.Printf("  %s Reading run-state.json...\n", terminal.Gray(terminal.SymbolBullet))
	stateFile := filepath.Join(destPath, "run-state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("  %s No run-state.json found, skipping DB replay\n", terminal.Gray(terminal.SymbolBullet))
			return nil
		}
		fmt.Printf("  %s Failed to read run-state.json: %v\n", terminal.Red(terminal.SymbolError), err)
		return nil
	}

	var stateExport state.StateExport
	if err := json.Unmarshal(data, &stateExport); err != nil {
		fmt.Printf("  %s Failed to parse run-state.json: %v\n", terminal.Red(terminal.SymbolError), err)
		return nil
	}

	// 2. Validate run info
	if stateExport.Run == nil || stateExport.Run.WorkflowName == "" {
		fmt.Printf("  %s No workflow info in run-state.json, skipping DB replay\n", terminal.Gray(terminal.SymbolBullet))
		return nil
	}

	runInfo := stateExport.Run
	fmt.Printf("  %s Found workflow: %s (kind: %s)\n", terminal.Cyan(terminal.SymbolBullet), terminal.Cyan(runInfo.WorkflowName), terminal.Gray(runInfo.WorkflowKind))

	// 3. Load workflow
	fmt.Printf("  %s Loading workflow...\n", terminal.Gray(terminal.SymbolBullet))
	loader := parser.NewLoader(cfg.WorkflowsPath)
	var dbFuncs []dbFunctionEntry
	var flow *core.Workflow
	var preloaded map[string]*core.Workflow
	var mod *core.Workflow

	switch core.WorkflowKind(runInfo.WorkflowKind) {
	case core.KindFlow:
		flow, preloaded, err = loader.LoadFlowWithModules(runInfo.WorkflowName)
		if err != nil {
			fmt.Printf("  %s Failed to load flow workflow %q: %v\n", terminal.Red(terminal.SymbolError), runInfo.WorkflowName, err)
			return nil
		}
		dbFuncs = collectDBFunctionsFromFlow(flow, preloaded, loader)
	default:
		mod, err = loader.LoadWorkflow(runInfo.WorkflowName)
		if err != nil {
			fmt.Printf("  %s Failed to load module workflow %q: %v\n", terminal.Red(terminal.SymbolError), runInfo.WorkflowName, err)
			return nil
		}
		dbFuncs = collectDBFunctionsFromSteps(mod.Steps)
	}

	if len(dbFuncs) == 0 {
		fmt.Printf("  %s No db_ functions found in workflow %q\n", terminal.Gray(terminal.SymbolBullet), runInfo.WorkflowName)
		return nil
	}

	fmt.Printf("  %s Collected %s db_ functions to replay\n", terminal.Cyan(terminal.SymbolBullet), terminal.Cyan(fmt.Sprintf("%d", len(dbFuncs))))

	// 4. Build template context
	vars := buildReplayContext(destPath, workspaceName, runInfo, cfg)
	engine := template.NewEngine()

	// 4b. Resolve workflow/module parameter defaults into vars.
	// Precedence (highest to lowest, matching normal execution):
	//   1. System vars (Target, Output, etc.) — from buildReplayContext()
	//   2. runInfo.Params — user-provided at original runtime
	//   3. ModuleRef.Params — flow-level overrides for modules
	//   4. Module param defaults — lowest priority
	switch core.WorkflowKind(runInfo.WorkflowKind) {
	case core.KindFlow:
		// Resolve flow-level param defaults first
		resolveParamDefaults(flow.Params, vars, engine)

		// Then process each module reference
		for i := range flow.Modules {
			modRef := &flow.Modules[i]

			// Apply ModuleRef.Params overrides (only if not already set)
			for k, v := range modRef.Params {
				if _, exists := vars[k]; !exists {
					rendered, renderErr := engine.Render(v, vars)
					if renderErr == nil {
						vars[k] = rendered
					}
				}
			}

			// Load module and resolve its param defaults
			if modRef.IsInline() {
				inlineWf := modRef.ToWorkflow()
				if inlineWf != nil {
					resolveParamDefaults(inlineWf.Params, vars, engine)
				}
			} else {
				var modWorkflow *core.Workflow
				if preloaded != nil {
					modWorkflow = preloaded[modRef.Name]
				}
				if modWorkflow == nil {
					modWorkflow, _ = loader.LoadWorkflow(modRef.Name)
				}
				if modWorkflow != nil {
					resolveParamDefaults(modWorkflow.Params, vars, engine)
				}
			}
		}
	default:
		resolveParamDefaults(mod.Params, vars, engine)
	}

	// 5. Render and execute each db_ function (skip pre_conditions — db_ functions
	// handle missing files gracefully, and pre_conditions may incorrectly gate
	// replay when used as "did this step run?" checks in the original workflow)
	fmt.Printf("  %s Replaying database operations...\n", terminal.Blue(terminal.SymbolRunning))
	registry := functions.GetDefaultRegistry()
	executed := 0
	failed := 0

	for i, entry := range dbFuncs {
		progress := terminal.Gray(fmt.Sprintf("[%d/%d]", i+1, len(dbFuncs)))

		// Render the function expression
		rendered, err := engine.Render(entry.expr, vars)
		if err != nil {
			fmt.Printf("    %s %s %s render error: %v\n", progress, terminal.Red(terminal.SymbolError), terminal.Gray(entry.expr), err)
			failed++
			continue
		}

		// Execute the function
		if _, err := registry.Execute(rendered, vars); err != nil {
			fmt.Printf("    %s %s %s %v\n", progress, terminal.Red(terminal.SymbolError), terminal.Gray(rendered), err)
			failed++
			continue
		}

		fmt.Printf("    %s %s %s\n", progress, terminal.Green(terminal.SymbolSuccess), terminal.Gray(rendered))
		executed++
	}

	if failed > 0 {
		fmt.Printf("  %s Replayed %s/%s database operations from workflow %s (%s failed)\n",
			terminal.Yellow(terminal.SymbolWarning),
			terminal.Green(fmt.Sprintf("%d", executed)),
			fmt.Sprintf("%d", len(dbFuncs)),
			terminal.Cyan(runInfo.WorkflowName),
			terminal.Red(fmt.Sprintf("%d", failed)))
	} else {
		fmt.Printf("  %s Replayed %s/%s database operations from workflow %s\n",
			terminal.Green(terminal.SymbolSuccess),
			terminal.Green(fmt.Sprintf("%d", executed)),
			fmt.Sprintf("%d", len(dbFuncs)),
			terminal.Cyan(runInfo.WorkflowName))
	}

	return nil
}

// createHighCompressionZip creates a zip archive with highest compression level
func createHighCompressionZip(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()

	archive := zip.NewWriter(zipFile)
	defer func() { _ = archive.Close() }()

	// Register highest compression level
	archive.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	baseDir := filepath.Base(source)

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Include base directory name in archive path
		archivePath := filepath.Join(baseDir, relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = archivePath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(writer, file)
		return err
	})
}

// extractZip extracts a zip archive to destination, returns count of files extracted
func extractZip(src, dest string) (int, error) {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return 0, err
	}
	defer func() { _ = reader.Close() }()

	filesCount := 0

	for _, file := range reader.File {
		// Prevent zip slip attack
		destPath := filepath.Join(dest, file.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filesCount, fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return filesCount, err
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return filesCount, err
		}

		// Extract file
		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return filesCount, err
		}

		rc, err := file.Open()
		if err != nil {
			_ = outFile.Close()
			return filesCount, err
		}

		_, err = io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()

		if err != nil {
			return filesCount, err
		}

		filesCount++
	}

	return filesCount, nil
}

// downloadFile downloads a file from URL to a temp file
func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "snapshot-*.zip")
	if err != nil {
		return "", err
	}
	defer func() { _ = tempFile.Close() }()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// isTimestampSuffix returns true if s looks like a timestamp suffix:
// either ISO 8601 format (e.g. "2026-02-13T18-20-34Z") or Unix epoch digits.
func isTimestampSuffix(s string) bool {
	return isISO8601Timestamp(s) || (len(s) >= 10 && isNumeric(s))
}

// isISO8601Timestamp checks if s matches either:
//   - compact: YYYYMMDDTHHMMSSZ           (16 chars)
//   - dashed:  YYYY-MM-DDTHH-MM-SSZ       (20 chars)
func isISO8601Timestamp(s string) bool {
	switch len(s) {
	case 16: // 20260213T182034Z
		if s[8] != 'T' || s[15] != 'Z' {
			return false
		}
		for _, i := range []int{0, 1, 2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 13, 14} {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
		return true
	case 20: // 2026-02-13T18-20-34Z
		if s[4] != '-' || s[7] != '-' || s[10] != 'T' || s[13] != '-' || s[16] != '-' || s[19] != 'Z' {
			return false
		}
		for _, i := range []int{0, 1, 2, 3, 5, 6, 8, 9, 11, 12, 14, 15, 17, 18} {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// ListSnapshots returns a list of snapshot files in the snapshot directory
func ListSnapshots(snapshotPath string) ([]SnapshotInfo, error) {
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return []SnapshotInfo{}, nil
	}

	entries, err := os.ReadDir(snapshotPath)
	if err != nil {
		return nil, err
	}

	var snapshots []SnapshotInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		snapshots = append(snapshots, SnapshotInfo{
			Name:      entry.Name(),
			Path:      filepath.Join(snapshotPath, entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	return snapshots, nil
}

// SnapshotInfo contains information about a snapshot file
type SnapshotInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

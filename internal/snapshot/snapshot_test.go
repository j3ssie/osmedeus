package snapshot

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/state"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNumeric(t *testing.T) {
	t.Run("valid numeric string", func(t *testing.T) {
		assert.True(t, isNumeric("1234567890"))
	})

	t.Run("empty string returns false", func(t *testing.T) {
		assert.False(t, isNumeric(""))
	})

	t.Run("string with letters returns false", func(t *testing.T) {
		assert.False(t, isNumeric("123abc"))
	})

	t.Run("string with spaces returns false", func(t *testing.T) {
		assert.False(t, isNumeric("123 456"))
	})

	t.Run("timestamp-like string is numeric", func(t *testing.T) {
		assert.True(t, isNumeric("1704067200"))
	})
}

func TestIsISO8601Timestamp(t *testing.T) {
	t.Run("compact format", func(t *testing.T) {
		assert.True(t, isISO8601Timestamp("20260213T182034Z"))
	})

	t.Run("compact format another value", func(t *testing.T) {
		assert.True(t, isISO8601Timestamp("20240101T000000Z"))
	})

	t.Run("dashed format", func(t *testing.T) {
		assert.True(t, isISO8601Timestamp("2026-02-13T18-20-34Z"))
	})

	t.Run("dashed format another value", func(t *testing.T) {
		assert.True(t, isISO8601Timestamp("2024-01-01T00-00-00Z"))
	})

	t.Run("too short", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("20260213T1820Z"))
	})

	t.Run("missing T separator", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("2026021318203400"))
	})

	t.Run("missing Z suffix compact", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("20260213T182034X"))
	})

	t.Run("missing Z suffix dashed", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("2026-02-13T18-20-34X"))
	})

	t.Run("unix timestamp is not ISO 8601", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("1704067200"))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp(""))
	})

	t.Run("letters in date portion compact", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("2026ab13T182034Z"))
	})

	t.Run("letters in date portion dashed", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("2026-ab-13T18-20-34Z"))
	})

	t.Run("wrong length 17 chars", func(t *testing.T) {
		assert.False(t, isISO8601Timestamp("20260213T1820345Z"))
	})
}

func TestIsTimestampSuffix(t *testing.T) {
	t.Run("ISO 8601 compact format", func(t *testing.T) {
		assert.True(t, isTimestampSuffix("20260213T182034Z"))
	})

	t.Run("ISO 8601 dashed format", func(t *testing.T) {
		assert.True(t, isTimestampSuffix("2026-02-13T18-20-34Z"))
	})

	t.Run("Unix timestamp", func(t *testing.T) {
		assert.True(t, isTimestampSuffix("1704067200"))
	})

	t.Run("short numeric string is not a timestamp", func(t *testing.T) {
		assert.False(t, isTimestampSuffix("123"))
	})

	t.Run("random text", func(t *testing.T) {
		assert.False(t, isTimestampSuffix("notadate"))
	})
}

func TestCreateHighCompressionZip(t *testing.T) {
	t.Run("creates valid zip archive", func(t *testing.T) {
		// Create temp source directory
		sourceDir, err := os.MkdirTemp("", "snapshot-test-source-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(sourceDir) }()

		// Create test files
		err = os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test content"), 0644)
		require.NoError(t, err)

		subDir := filepath.Join(sourceDir, "subdir")
		err = os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644)
		require.NoError(t, err)

		// Create zip
		targetFile := filepath.Join(os.TempDir(), "test-snapshot.zip")
		defer func() { _ = os.Remove(targetFile) }()

		err = createHighCompressionZip(sourceDir, targetFile)
		require.NoError(t, err)

		// Verify zip exists and is valid
		reader, err := zip.OpenReader(targetFile)
		require.NoError(t, err)
		defer func() { _ = reader.Close() }()

		assert.GreaterOrEqual(t, len(reader.File), 2, "zip should contain at least 2 files")
	})

	t.Run("fails with non-existent source", func(t *testing.T) {
		err := createHighCompressionZip("/nonexistent/path", "/tmp/test.zip")
		assert.Error(t, err)
	})
}

func TestExtractZip(t *testing.T) {
	t.Run("extracts files correctly", func(t *testing.T) {
		// Create a test zip file
		zipPath := filepath.Join(os.TempDir(), "test-extract.zip")
		defer func() { _ = os.Remove(zipPath) }()

		// Create zip with test content
		zipFile, err := os.Create(zipPath)
		require.NoError(t, err)

		writer := zip.NewWriter(zipFile)
		f, err := writer.Create("test.txt")
		require.NoError(t, err)
		_, err = f.Write([]byte("test content"))
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)
		err = zipFile.Close()
		require.NoError(t, err)

		// Extract to temp directory
		destDir, err := os.MkdirTemp("", "snapshot-test-extract-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(destDir) }()

		filesCount, err := extractZip(zipPath, destDir)
		require.NoError(t, err)
		assert.Equal(t, 1, filesCount)

		// Verify file exists
		content, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})

	t.Run("prevents zip slip attack", func(t *testing.T) {
		// Create a malicious zip with path traversal
		zipPath := filepath.Join(os.TempDir(), "test-zipslip.zip")
		defer func() { _ = os.Remove(zipPath) }()

		zipFile, err := os.Create(zipPath)
		require.NoError(t, err)

		writer := zip.NewWriter(zipFile)
		// Try to create a file with path traversal
		f, err := writer.Create("../../../etc/malicious.txt")
		require.NoError(t, err)
		_, err = f.Write([]byte("malicious content"))
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)
		err = zipFile.Close()
		require.NoError(t, err)

		// Attempt extraction should fail
		destDir, err := os.MkdirTemp("", "snapshot-test-zipslip-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(destDir) }()

		_, err = extractZip(zipPath, destDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "illegal file path")
	})

	t.Run("extracts zip with directory prefix without double nesting", func(t *testing.T) {
		// Simulate what createHighCompressionZip produces: entries prefixed with workspace name
		// e.g. "shopee.vn/output.txt", "shopee.vn/subdir/nested.txt"
		zipPath := filepath.Join(os.TempDir(), "test-extract-prefix.zip")
		defer func() { _ = os.Remove(zipPath) }()

		zipFile, err := os.Create(zipPath)
		require.NoError(t, err)

		writer := zip.NewWriter(zipFile)

		// Directory entry with proper permissions
		dirHeader := &zip.FileHeader{Name: "shopee.vn/"}
		dirHeader.SetMode(0755)
		_, err = writer.CreateHeader(dirHeader)
		require.NoError(t, err)

		// File inside the directory
		f, err := writer.Create("shopee.vn/output.txt")
		require.NoError(t, err)
		_, err = f.Write([]byte("scan results"))
		require.NoError(t, err)

		// Nested subdirectory with proper permissions
		subDirHeader := &zip.FileHeader{Name: "shopee.vn/subdir/"}
		subDirHeader.SetMode(0755)
		_, err = writer.CreateHeader(subDirHeader)
		require.NoError(t, err)

		f, err = writer.Create("shopee.vn/subdir/nested.txt")
		require.NoError(t, err)
		_, err = f.Write([]byte("nested content"))
		require.NoError(t, err)

		require.NoError(t, writer.Close())
		require.NoError(t, zipFile.Close())

		// Extract into a "workspaces" parent dir (simulating workspacesPath)
		workspacesDir, err := os.MkdirTemp("", "snapshot-test-workspaces-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workspacesDir) }()

		filesCount, err := extractZip(zipPath, workspacesDir)
		require.NoError(t, err)
		assert.Equal(t, 2, filesCount) // 2 files (directories are not counted)

		// Verify correct structure: workspacesDir/shopee.vn/output.txt (NOT workspacesDir/shopee.vn/shopee.vn/output.txt)
		content, err := os.ReadFile(filepath.Join(workspacesDir, "shopee.vn", "output.txt"))
		require.NoError(t, err)
		assert.Equal(t, "scan results", string(content))

		content, err = os.ReadFile(filepath.Join(workspacesDir, "shopee.vn", "subdir", "nested.txt"))
		require.NoError(t, err)
		assert.Equal(t, "nested content", string(content))

		// Verify NO double nesting
		_, err = os.Stat(filepath.Join(workspacesDir, "shopee.vn", "shopee.vn"))
		assert.True(t, os.IsNotExist(err), "should not have double-nested shopee.vn/shopee.vn directory")
	})

	t.Run("fails with non-existent zip", func(t *testing.T) {
		destDir, err := os.MkdirTemp("", "snapshot-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(destDir) }()

		_, err = extractZip("/nonexistent/file.zip", destDir)
		assert.Error(t, err)
	})
}

func TestExportWorkspace(t *testing.T) {
	t.Run("exports workspace successfully", func(t *testing.T) {
		// Create temp workspace directory
		workspaceDir, err := os.MkdirTemp("", "workspace-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workspaceDir) }()

		// Create test file
		err = os.WriteFile(filepath.Join(workspaceDir, "output.txt"), []byte("scan results"), 0644)
		require.NoError(t, err)

		// Export
		outputPath := filepath.Join(os.TempDir(), "workspace-export.zip")
		defer func() { _ = os.Remove(outputPath) }()

		result, err := ExportWorkspace(workspaceDir, outputPath)
		require.NoError(t, err)

		assert.Equal(t, filepath.Base(workspaceDir), result.WorkspaceName)
		assert.Equal(t, workspaceDir, result.SourcePath)
		assert.Equal(t, outputPath, result.OutputPath)
		assert.Greater(t, result.FileSize, int64(0))
	})

	t.Run("fails with non-existent workspace", func(t *testing.T) {
		_, err := ExportWorkspace("/nonexistent/workspace", "/tmp/test.zip")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workspace not found")
	})

	t.Run("fails if workspace is a file not directory", func(t *testing.T) {
		// Create a file instead of directory
		tmpFile, err := os.CreateTemp("", "not-a-dir-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		require.NoError(t, tmpFile.Close())

		_, err = ExportWorkspace(tmpFile.Name(), "/tmp/test.zip")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("generates output path if empty", func(t *testing.T) {
		// Create temp workspace directory
		workspaceDir, err := os.MkdirTemp("", "workspace-test-autopath-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workspaceDir) }()

		err = os.WriteFile(filepath.Join(workspaceDir, "file.txt"), []byte("content"), 0644)
		require.NoError(t, err)

		result, err := ExportWorkspace(workspaceDir, "")
		require.NoError(t, err)
		defer func() { _ = os.Remove(result.OutputPath) }()

		assert.Contains(t, result.OutputPath, filepath.Base(workspaceDir))
		assert.Contains(t, result.OutputPath, ".zip")
	})
}

func TestListSnapshots(t *testing.T) {
	t.Run("lists zip files in directory", func(t *testing.T) {
		// Create temp snapshot directory
		snapshotDir, err := os.MkdirTemp("", "snapshot-list-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(snapshotDir) }()

		// Create test zip files
		err = os.WriteFile(filepath.Join(snapshotDir, "workspace1_1234567890.zip"), []byte("zip1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(snapshotDir, "workspace2_1234567891.zip"), []byte("zip2"), 0644)
		require.NoError(t, err)

		// Create a non-zip file (should be ignored)
		err = os.WriteFile(filepath.Join(snapshotDir, "readme.txt"), []byte("readme"), 0644)
		require.NoError(t, err)

		// Create a directory (should be ignored)
		err = os.MkdirAll(filepath.Join(snapshotDir, "subdir"), 0755)
		require.NoError(t, err)

		snapshots, err := ListSnapshots(snapshotDir)
		require.NoError(t, err)

		assert.Len(t, snapshots, 2)
	})

	t.Run("returns empty list for non-existent directory", func(t *testing.T) {
		snapshots, err := ListSnapshots("/nonexistent/path")
		require.NoError(t, err)
		assert.Empty(t, snapshots)
	})

	t.Run("returns empty list for empty directory", func(t *testing.T) {
		snapshotDir, err := os.MkdirTemp("", "snapshot-list-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(snapshotDir) }()

		snapshots, err := ListSnapshots(snapshotDir)
		require.NoError(t, err)
		assert.Empty(t, snapshots)
	})
}

func TestExportResult(t *testing.T) {
	t.Run("struct fields are populated", func(t *testing.T) {
		result := ExportResult{
			WorkspaceName: "example.com",
			SourcePath:    "/home/user/workspaces/example.com",
			OutputPath:    "/home/user/snapshot/example.com_123.zip",
			FileSize:      1024,
		}

		assert.Equal(t, "example.com", result.WorkspaceName)
		assert.Equal(t, "/home/user/workspaces/example.com", result.SourcePath)
		assert.Equal(t, "/home/user/snapshot/example.com_123.zip", result.OutputPath)
		assert.Equal(t, int64(1024), result.FileSize)
	})
}

func TestImportResult(t *testing.T) {
	t.Run("struct fields are populated", func(t *testing.T) {
		result := ImportResult{
			WorkspaceName: "example.com",
			LocalPath:     "/home/user/workspaces/example.com",
			DataSource:    "imported",
			FilesCount:    100,
		}

		assert.Equal(t, "example.com", result.WorkspaceName)
		assert.Equal(t, "/home/user/workspaces/example.com", result.LocalPath)
		assert.Equal(t, "imported", result.DataSource)
		assert.Equal(t, 100, result.FilesCount)
	})
}

func TestSnapshotInfo(t *testing.T) {
	t.Run("struct fields are accessible", func(t *testing.T) {
		info := SnapshotInfo{
			Name: "example.com_123.zip",
			Path: "/home/user/snapshot/example.com_123.zip",
			Size: 2048,
		}

		assert.Equal(t, "example.com_123.zip", info.Name)
		assert.Equal(t, "/home/user/snapshot/example.com_123.zip", info.Path)
		assert.Equal(t, int64(2048), info.Size)
	})
}

func TestCollectDBFunctionsFromSteps(t *testing.T) {
	t.Run("collects db_ functions from function steps", func(t *testing.T) {
		steps := []core.Step{
			{
				Name: "import-assets",
				Type: core.StepTypeFunction,
				Functions: []string{
					`db_import("{{Workspace}}", "{{Output}}/assets.txt")`,
					`log_info("done importing")`,
				},
				PreCondition: `file_exists("{{Output}}/assets.txt")`,
			},
			{
				Name:    "run-scan",
				Type:    core.StepTypeBash,
				Command: "nmap {{Target}}",
			},
			{
				Name:     "import-vulns",
				Type:     core.StepTypeFunction,
				Function: `db_import_sarif("{{Workspace}}", "{{Output}}/results.sarif")`,
			},
		}

		entries := collectDBFunctionsFromSteps(steps)

		assert.Len(t, entries, 2)

		assert.Contains(t, entries[0].expr, "db_import")
		assert.Equal(t, "import-assets", entries[0].stepName)
		assert.Contains(t, entries[0].preCondition, "file_exists")

		assert.Contains(t, entries[1].expr, "db_import_sarif")
		assert.Equal(t, "import-vulns", entries[1].stepName)
	})

	t.Run("collects from parallel_functions", func(t *testing.T) {
		steps := []core.Step{
			{
				Name: "parallel-import",
				Type: core.StepTypeFunction,
				ParallelFunctions: []string{
					`db_import("{{Workspace}}", "{{Output}}/subs.txt")`,
					`db_import("{{Workspace}}", "{{Output}}/urls.txt")`,
					`log_info("not a db function")`,
				},
			},
		}

		entries := collectDBFunctionsFromSteps(steps)

		assert.Len(t, entries, 2)
	})

	t.Run("ignores non-function steps", func(t *testing.T) {
		steps := []core.Step{
			{
				Name:    "run-scan",
				Type:    core.StepTypeBash,
				Command: "echo db_import is in the command but not a function step",
			},
		}

		entries := collectDBFunctionsFromSteps(steps)
		assert.Empty(t, entries)
	})

	t.Run("returns empty for no steps", func(t *testing.T) {
		entries := collectDBFunctionsFromSteps(nil)
		assert.Empty(t, entries)
	})
}

func TestCollectDBFunctionsFromSteps_ParallelAndForeach(t *testing.T) {
	t.Run("recurses into parallel-steps", func(t *testing.T) {
		steps := []core.Step{
			{
				Name: "parallel-group",
				Type: core.StepTypeParallel,
				ParallelSteps: []core.Step{
					{
						Name:     "inner-import",
						Type:     core.StepTypeFunction,
						Function: `db_import("{{Workspace}}", "{{Output}}/inner.txt")`,
					},
					{
						Name:    "inner-bash",
						Type:    core.StepTypeBash,
						Command: "echo hello",
					},
				},
			},
		}

		entries := collectDBFunctionsFromSteps(steps)

		assert.Len(t, entries, 1)
		assert.Contains(t, entries[0].expr, "db_import")
		assert.Equal(t, "inner-import", entries[0].stepName)
	})

	t.Run("recurses into foreach inner step", func(t *testing.T) {
		innerStep := core.Step{
			Name:     "foreach-import",
			Type:     core.StepTypeFunction,
			Function: `db_import("{{Workspace}}", "[[item]]")`,
		}
		steps := []core.Step{
			{
				Name:     "foreach-loop",
				Type:     core.StepTypeForeach,
				Input:    "{{Output}}/files.txt",
				Variable: "item",
				Step:     &innerStep,
			},
		}

		entries := collectDBFunctionsFromSteps(steps)

		assert.Len(t, entries, 1)
		assert.Equal(t, "foreach-import", entries[0].stepName)
	})
}

func TestCollectDBFunctionsFromFlow(t *testing.T) {
	t.Run("collects from inline modules", func(t *testing.T) {
		flow := &core.Workflow{
			Kind: core.KindFlow,
			Modules: []core.ModuleRef{
				{
					Name: "inline-mod",
					Steps: []core.Step{
						{
							Name:     "db-step",
							Type:     core.StepTypeFunction,
							Function: `db_import("{{Workspace}}", "{{Output}}/data.txt")`,
						},
					},
				},
			},
		}

		entries := collectDBFunctionsFromFlow(flow, nil, nil)

		assert.Len(t, entries, 1)
		assert.Contains(t, entries[0].expr, "db_import")
	})

	t.Run("collects from preloaded modules", func(t *testing.T) {
		flow := &core.Workflow{
			Kind: core.KindFlow,
			Modules: []core.ModuleRef{
				{Name: "recon-module"},
			},
		}

		preloaded := map[string]*core.Workflow{
			"recon-module": {
				Kind: core.KindModule,
				Steps: []core.Step{
					{
						Name:     "import-step",
						Type:     core.StepTypeFunction,
						Function: `db_import("{{Workspace}}", "{{Output}}/recon.txt")`,
					},
				},
			},
		}

		entries := collectDBFunctionsFromFlow(flow, preloaded, nil)

		assert.Len(t, entries, 1)
		assert.Contains(t, entries[0].expr, "db_import")
	})

	t.Run("returns empty for no modules", func(t *testing.T) {
		flow := &core.Workflow{
			Kind: core.KindFlow,
		}

		entries := collectDBFunctionsFromFlow(flow, nil, nil)
		assert.Empty(t, entries)
	})
}

func TestBuildReplayContext(t *testing.T) {
	t.Run("populates all expected variables", func(t *testing.T) {
		runInfo := &state.RunInfo{
			Target:       "example.com",
			WorkflowName: "test-flow",
			Params: map[string]any{
				"enableDnsBruteForcing": true,
			},
		}

		cfg := &config.Config{
			BaseFolder:          "/home/user/osmedeus-base",
			BinariesPath:        "/home/user/osmedeus-base/external-binaries",
			DataPath:            "/home/user/osmedeus-base/external-data",
			WorkspacesPath:      "/home/user/workspaces-osmedeus",
			WorkflowsPath:       "/home/user/osmedeus-base/workflows",
			ConfigsPath:         "/home/user/osmedeus-base/external-configs",
			ExternalScriptsPath: "/home/user/osmedeus-base/external-scripts",
		}

		vars := buildReplayContext("/home/user/workspaces-osmedeus/example.com", "example.com", runInfo, cfg)

		// Core variables
		assert.Equal(t, "example.com", vars["Target"])
		assert.Equal(t, "example.com", vars["Workspace"], "Workspace should be overridden to workspaceName")
		assert.Equal(t, "example.com", vars["TargetSpace"], "TargetSpace should be overridden to workspaceName")
		assert.Equal(t, "/home/user/workspaces-osmedeus/example.com", vars["Output"], "Output should be overridden to destPath")

		// Config path variables (from BuildBuiltinVariables via cfg)
		assert.Equal(t, "/home/user/osmedeus-base", vars["BaseFolder"])
		assert.Equal(t, "/home/user/osmedeus-base/external-binaries", vars["Binaries"])
		assert.Equal(t, "/home/user/osmedeus-base/external-data", vars["Data"])
		assert.Equal(t, "/home/user/workspaces-osmedeus", vars["Workspaces"])
		assert.Equal(t, "/home/user/osmedeus-base/workflows", vars["Workflows"])

		// User params should be merged
		assert.Equal(t, "true", vars["enableDnsBruteForcing"])

		// Heuristic target-derived variables (computed by BuildBuiltinVariables)
		assert.NotEmpty(t, vars["TargetRootDomain"], "TargetRootDomain should be computed from target")
		assert.NotEmpty(t, vars["TargetType"], "TargetType should be computed from target")

		// Platform variables
		assert.NotEmpty(t, vars["PlatformOS"], "PlatformOS should be set")
		assert.NotEmpty(t, vars["PlatformArch"], "PlatformArch should be set")

		// Thread variables (defaults from BuildBuiltinVariables)
		assert.NotNil(t, vars["threads"], "threads should have a default value")
		assert.NotNil(t, vars["baseThreads"], "baseThreads should have a default value")

		// TempDir/TempFile should be cleaned up
		_, hasTempDir := vars["TempDir"]
		_, hasTempFile := vars["TempFile"]
		assert.False(t, hasTempDir, "TempDir should be cleaned up")
		assert.False(t, hasTempFile, "TempFile should be cleaned up")
	})

	t.Run("overrides Output and Workspace with imported values", func(t *testing.T) {
		runInfo := &state.RunInfo{
			Target:       "example.com",
			WorkflowName: "test-flow",
		}
		cfg := &config.Config{
			WorkspacesPath: "/home/user/workspaces-osmedeus",
		}

		// Use a different destPath than what BuildBuiltinVariables would compute
		destPath := "/custom/imported/workspace/example.com"
		vars := buildReplayContext(destPath, "my-workspace", runInfo, cfg)

		assert.Equal(t, destPath, vars["Output"], "Output should be destPath, not computed")
		assert.Equal(t, "my-workspace", vars["TargetSpace"], "TargetSpace should be workspaceName, not computed")
		assert.Equal(t, "my-workspace", vars["Workspace"], "Workspace should be workspaceName, not computed")
	})

	t.Run("handles nil params", func(t *testing.T) {
		runInfo := &state.RunInfo{
			Target: "test.com",
		}
		cfg := &config.Config{}

		vars := buildReplayContext("/tmp/ws/test.com", "test.com", runInfo, cfg)
		assert.Equal(t, "test.com", vars["Target"])
		assert.NotEmpty(t, vars["PlatformOS"], "platform vars should still be set")
	})
}

func TestReplayDBOperations_MissingRunState(t *testing.T) {
	t.Run("returns nil when run-state.json is missing", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "replay-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		cfg := &config.Config{
			WorkflowsPath: "/nonexistent/workflows",
		}

		err = replayDBOperations(tmpDir, "example.com", cfg)
		assert.NoError(t, err)
	})
}

func TestReplayDBOperations_EmptyWorkflowName(t *testing.T) {
	t.Run("returns nil when workflow name is empty", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "replay-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Write run-state.json with empty workflow name
		stateData := `{"run":{"workflow_name":"","target":"example.com"}}`
		err = os.WriteFile(filepath.Join(tmpDir, "run-state.json"), []byte(stateData), 0644)
		require.NoError(t, err)

		cfg := &config.Config{
			WorkflowsPath: "/nonexistent/workflows",
		}

		err = replayDBOperations(tmpDir, "example.com", cfg)
		assert.NoError(t, err)
	})

	t.Run("returns nil when run info is nil", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "replay-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Write run-state.json with no run field
		stateData := `{"workspace":{"name":"example.com"}}`
		err = os.WriteFile(filepath.Join(tmpDir, "run-state.json"), []byte(stateData), 0644)
		require.NoError(t, err)

		cfg := &config.Config{
			WorkflowsPath: "/nonexistent/workflows",
		}

		err = replayDBOperations(tmpDir, "example.com", cfg)
		assert.NoError(t, err)
	})

	t.Run("returns nil when run-state.json is invalid JSON", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "replay-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tmpDir) }()

		err = os.WriteFile(filepath.Join(tmpDir, "run-state.json"), []byte("not json"), 0644)
		require.NoError(t, err)

		cfg := &config.Config{
			WorkflowsPath: "/nonexistent/workflows",
		}

		err = replayDBOperations(tmpDir, "example.com", cfg)
		assert.NoError(t, err)
	})
}

func TestResolveParamDefaults(t *testing.T) {
	t.Run("resolves simple defaults", func(t *testing.T) {
		params := []core.Param{
			{Name: "dnsFile", Default: "{{Output}}/dns.txt"},
		}
		vars := map[string]any{
			"Output": "/workspace/example.com",
		}
		engine := template.NewEngine()

		resolveParamDefaults(params, vars, engine)

		assert.Equal(t, "/workspace/example.com/dns.txt", vars["dnsFile"])
	})

	t.Run("does not overwrite existing vars", func(t *testing.T) {
		params := []core.Param{
			{Name: "threads", Default: "10"},
		}
		vars := map[string]any{
			"threads": "20",
		}
		engine := template.NewEngine()

		resolveParamDefaults(params, vars, engine)

		assert.Equal(t, "20", vars["threads"], "existing var should not be overwritten")
	})

	t.Run("skips params without defaults", func(t *testing.T) {
		params := []core.Param{
			{Name: "noDefault"},
			{Name: "emptyDefault", Default: ""},
		}
		vars := map[string]any{}
		engine := template.NewEngine()

		resolveParamDefaults(params, vars, engine)

		_, hasNoDefault := vars["noDefault"]
		_, hasEmptyDefault := vars["emptyDefault"]
		assert.False(t, hasNoDefault, "param without default should not be added")
		assert.False(t, hasEmptyDefault, "param with empty default should not be added")
	})

	t.Run("handles chained defaults", func(t *testing.T) {
		params := []core.Param{
			{Name: "baseDir", Default: "{{Output}}/probing"},
			{Name: "dnsFile", Default: "{{baseDir}}/dns.txt"},
		}
		vars := map[string]any{
			"Output": "/workspace/example.com",
		}
		engine := template.NewEngine()

		resolveParamDefaults(params, vars, engine)

		assert.Equal(t, "/workspace/example.com/probing", vars["baseDir"])
		assert.Equal(t, "/workspace/example.com/probing/dns.txt", vars["dnsFile"])
	})

	t.Run("handles nil params slice", func(t *testing.T) {
		vars := map[string]any{"Output": "/workspace"}
		engine := template.NewEngine()

		// Should not panic
		resolveParamDefaults(nil, vars, engine)

		assert.Equal(t, "/workspace", vars["Output"], "existing vars should be unchanged")
	})

	t.Run("handles empty params slice", func(t *testing.T) {
		vars := map[string]any{"Output": "/workspace"}
		engine := template.NewEngine()

		resolveParamDefaults([]core.Param{}, vars, engine)

		assert.Equal(t, "/workspace", vars["Output"], "existing vars should be unchanged")
	})

	t.Run("resolves multiple params with template references", func(t *testing.T) {
		params := []core.Param{
			{Name: "dnsFile", Default: "{{Output}}/probing/dns-{{TargetSpace}}.txt"},
			{Name: "httpFile", Default: "{{Output}}/probing/http-{{TargetSpace}}.txt"},
		}
		vars := map[string]any{
			"Output":      "/workspace/example.com",
			"TargetSpace": "example.com",
		}
		engine := template.NewEngine()

		resolveParamDefaults(params, vars, engine)

		assert.Equal(t, "/workspace/example.com/probing/dns-example.com.txt", vars["dnsFile"])
		assert.Equal(t, "/workspace/example.com/probing/http-example.com.txt", vars["httpFile"])
	})
}

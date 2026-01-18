package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/public"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check and fix environment health (alias for 'osmedeus install validate')",
	Long:  UsageHealth(),
	RunE:  runHealthWithSample,
}

func runHealthWithSample(cmd *cobra.Command, args []string) error {
	return runHealth(cmd, args)
}

func runHealth(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()
	cfg := config.Get()

	// Check if first-time setup is needed and run it
	if isFirstTimeSetupNeeded(cfg.BaseFolder) {
		if err := runFirstTimeSetup(cfg.BaseFolder, cfg); err != nil {
			printer.Warning("First-time setup had issues: %s", err)
		}
		// Reload config after setup
		if reloaded, err := config.Load(cfg.BaseFolder); err == nil {
			cfg = reloaded
			config.Set(cfg)
		}
	}

	printer.Newline()
	printer.Println("%s Osmedeus Environment Health Check %s",
		terminal.Yellow(terminal.SymbolMenu),
		terminal.Cyan(core.VERSION))

	hasErrors := false

	// 1. Check/create folders
	hasErrors = checkFolders(printer, cfg) || hasErrors

	// 2. Check config files
	hasErrors = checkConfigFiles(printer, cfg) || hasErrors

	// 3. Check workflows
	hasErrors = checkWorkflows(printer, cfg) || hasErrors

	// Summary
	fmt.Println()
	if hasErrors {
		printer.Warning("Some issues were found. Review the output above.")
	} else {
		printer.Success("All checks passed!")
	}

	printer.Newline()
	printer.Println("%s %s %s", terminal.Yellow(terminal.SymbolLightning), terminal.BoldCyan("Tip:"), terminal.Gray("See the full CLI documentation below for more details"))
	printer.Println(" %s", terminal.Green("https://docs.osmedeus.org/getting-started/cli"))

	return nil
}

func checkFolders(printer *terminal.Printer, cfg *config.Config) bool {
	printer.Section("Environments Folders")

	if _, err := os.Stat(cfg.BaseFolder); os.IsNotExist(err) {
		err := copyEmbeddedAssets(cfg.BaseFolder)
		if err != nil {
			printer.Error("  Base folder: failed to create %s - %v", terminal.White(cfg.BaseFolder), err)
			return true
		}
		printer.Success("  Base folder: created %s", terminal.White(cfg.BaseFolder))
	} else {
		printer.Success("  Base folder: %s", terminal.White(cfg.BaseFolder))
	}

	folders := []struct {
		name string
		path string
	}{
		{"Workspaces", cfg.WorkspacesPath},
		{"Workflows", cfg.WorkflowsPath},
		{"Binaries", cfg.BinariesPath},
		{"Data", cfg.DataPath},
		{"Markdown Report Templates", cfg.MarkdownReportTemplatesPath},
		{"External Agent Configs", cfg.ExternalAgentConfigsPath},
	}

	hasErrors := false
	for _, f := range folders {
		if f.path == "" {
			if f.name == "Binaries" {
				printer.Error("  %s: path not configured", f.name)
				hasErrors = true
			} else {
				printer.Info("  %s: not configured", f.name)
			}
			continue
		}

		if _, err := os.Stat(f.path); os.IsNotExist(err) {
			// Create folder
			if err := os.MkdirAll(f.path, 0755); err != nil {
				printer.Error("  %s: failed to create %s - %v", f.name, terminal.White(f.path), err)
				hasErrors = true
			} else {
				printer.Success("  %s: created %s", f.name, terminal.White(f.path))
			}
		} else {
			printer.Success("  %s: %s", f.name, terminal.White(f.path))
			// Check if binaries folder is empty (ignoring hidden files like .gitkeep)
			if f.name == "Binaries" {
				entries, err := os.ReadDir(f.path)
				if err == nil {
					binaryCount := 0
					for _, entry := range entries {
						if !strings.HasPrefix(entry.Name(), ".") {
							binaryCount++
						}
					}
					if binaryCount == 0 {
						printer.Error("  %s No binaries detected in %s",
							terminal.Red("✗"),
							terminal.White(f.path))
						printer.Println("    %s Run %s to fetch required binaries",
							terminal.Yellow("→"),
							terminal.Cyan("osmedeus install binary --all"))
						hasErrors = true
					}
				}
			}
		}
	}

	binariesFolder := cfg.BinariesPath
	if binariesFolder == "" {
		binariesFolder = filepath.Join(cfg.BaseFolder, "binaries")
	}

	// Use the shared helper to setup PATH (updates shell config + current process)
	ensureBinariesPathInEnv(printer, binariesFolder, true)

	return hasErrors
}

func checkConfigFiles(printer *terminal.Printer, cfg *config.Config) bool {
	printer.Section("Configuration Files")

	hasErrors := false

	// Check osm-settings.yaml
	settingsPath := filepath.Join(cfg.BaseFolder, "osm-settings.yaml")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		if err := config.EnsureConfigExists(cfg.BaseFolder); err != nil {
			printer.Error("  osm-settings.yaml: failed to create - %v", err)
			hasErrors = true
		} else {
			printer.Success("  osm-settings.yaml: created default config")
		}
	} else {
		// Validate existing config
		if err := cfg.Validate(); err != nil {
			printer.Error("  osm-settings.yaml: validation failed - %v", err)
			hasErrors = true
		} else {
			printer.Success("  osm-settings.yaml: valid")
		}
	}

	// Check global_vars configuration
	if len(cfg.GlobalVars) > 0 {
		printer.Success("  global_vars: %s variable(s) configured", terminal.White(fmt.Sprintf("%d", len(cfg.GlobalVars))))
	} else {
		printer.Info("  global_vars: no variables configured")
	}

	// Check notification configuration
	if cfg.IsNotificationConfigured() {
		printer.Success("  notification: %s enabled", terminal.White(cfg.Notification.Provider))
	} else {
		printer.Info("  notification: not configured")
	}

	return hasErrors
}

func checkWorkflows(printer *terminal.Printer, cfg *config.Config) bool {
	printer.Section("Workflows")

	if cfg.WorkflowsPath == "" {
		printer.Error("  Workflows path not configured")
		return true
	}

	// Check if workflows directory exists
	if _, err := os.Stat(cfg.WorkflowsPath); os.IsNotExist(err) {
		printer.Warning("  No workflows found in %s", terminal.White(cfg.WorkflowsPath))
		return false
	}

	workflowFiles, err := findWorkflowYAMLFiles(cfg.WorkflowsPath)
	if err != nil {
		printer.Error("  Failed to scan workflows folder: %v", err)
		return true
	}

	if len(workflowFiles) == 0 {
		printer.Warning("  No workflows found in %s", terminal.White(cfg.WorkflowsPath))
		return false
	}

	p := parser.NewParser()
	validCount := 0
	invalidCount := 0

	for _, filePath := range workflowFiles {
		relPath, _ := filepath.Rel(cfg.WorkflowsPath, filePath)
		relPath = filepath.Clean(relPath)

		wf, err := p.Parse(filePath)
		if err != nil {
			printer.Error("  [INVALID] %s: %v", terminal.White(relPath), err)
			invalidCount++
			continue
		}

		if err := p.Validate(wf); err != nil {
			printer.Error("  [INVALID] %s (%s): %v", terminal.White(relPath), terminal.White(wf.Name), err)
			invalidCount++
			continue
		}

		printer.Success("  [VALID] %s (%s) - %s", terminal.White(wf.Name), wf.Kind, terminal.Gray(relPath))
		validCount++
	}

	fmt.Println()
	printer.Info("  Total: %s valid, %s invalid", terminal.Green(fmt.Sprintf("%d", validCount)), terminal.Red(fmt.Sprintf("%d", invalidCount)))

	return invalidCount > 0
}

func findWorkflowYAMLFiles(root string) ([]string, error) {
	root = filepath.Clean(root)

	var out []string
	queue := []string{root}
	seen := make(map[string]struct{})

	for len(queue) > 0 {
		dir := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		realDir := dir
		if eval, err := filepath.EvalSymlinks(dir); err == nil {
			realDir = eval
		}
		if _, ok := seen[realDir]; ok {
			continue
		}
		seen[realDir] = struct{}{}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			name := entry.Name()
			if name == "." || name == ".." {
				continue
			}

			// Skip hidden directories and files
			if strings.HasPrefix(name, ".") {
				continue
			}

			fullPath := filepath.Join(dir, name)

			if entry.IsDir() {
				queue = append(queue, fullPath)
				continue
			}

			if entry.Type()&os.ModeSymlink != 0 {
				info, err := os.Stat(fullPath)
				if err == nil && info.IsDir() {
					queue = append(queue, fullPath)
					continue
				}
			}

			lower := strings.ToLower(name)
			if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") {
				// Only include files that are actually workflow YAML files
				if isWorkflowYAML(fullPath) {
					out = append(out, fullPath)
				}
			}
		}
	}

	sort.Strings(out)
	return out, nil
}

func copyEmbeddedAssets(dest string) error {
	srcFS := public.EmbedFS
	srcRoot := "examples/osmedeus-base.example"

	return fs.WalkDir(srcFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return os.MkdirAll(dest, 0755)
		}

		destPath := filepath.Join(dest, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		srcFile, err := srcFS.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = srcFile.Close() }()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer func() { _ = destFile.Close() }()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}

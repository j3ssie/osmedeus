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
	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/internal/utils"
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
	workflows, workflowErrors := checkWorkflows(printer, cfg)
	hasErrors = workflowErrors || hasErrors

	// 4. Cross-check what the workflows declare against what is installed
	hasErrors = checkWorkflowDependencies(printer, cfg, workflows) || hasErrors

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
						if strings.EqualFold(os.Getenv("OSM_IGNORE_REGISTRY"), "true") {
							printer.Info("  %s Binaries folder empty (OSM_IGNORE_REGISTRY=true, skipping check)",
								terminal.Yellow("⚠"))
						} else {
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

// checkWorkflows validates every workflow and returns the ones that parsed, so a
// later pass can inspect them without re-reading and re-unmarshalling each file.
func checkWorkflows(printer *terminal.Printer, cfg *config.Config) ([]*core.Workflow, bool) {
	printer.Section("Workflows")

	if cfg.WorkflowsPath == "" {
		printer.Error("  Workflows path not configured")
		return nil, true
	}

	// Check if workflows directory exists
	if _, err := os.Stat(cfg.WorkflowsPath); os.IsNotExist(err) {
		printer.Warning("  No workflows found in %s", terminal.White(cfg.WorkflowsPath))
		return nil, false
	}

	workflowFiles, err := findWorkflowYAMLFiles(cfg.WorkflowsPath)
	if err != nil {
		printer.Error("  Failed to scan workflows folder: %v", err)
		return nil, true
	}

	if len(workflowFiles) == 0 {
		printer.Error("  Workflows folder is empty: %s", terminal.White(cfg.WorkflowsPath))
		return nil, true
	}

	p := parser.NewParser()
	validCount := 0
	invalidCount := 0
	parsed := make([]*core.Workflow, 0, len(workflowFiles))

	for _, filePath := range workflowFiles {
		relPath, _ := filepath.Rel(cfg.WorkflowsPath, filePath)
		relPath = filepath.Clean(relPath)

		wf, err := p.Parse(filePath)
		if err != nil {
			printer.Error("  [INVALID] %s: %v", terminal.White(relPath), err)
			invalidCount++
			continue
		}
		parsed = append(parsed, wf)

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

	return parsed, invalidCount > 0
}

// checkWorkflowDependencies cross-checks every dependencies.commands entry declared
// by a workflow against what is actually installed, and against the binary registry.
//
// A missing declared command is a hard failure at run time ("required command not
// found: X"), so finding it here is the difference between a clear install-time
// message and a scan that dies mid-flow. Resolution deliberately mirrors the run-time
// gate in the executor (utils.LookPathWithBinaries): the binaries folder first, then
// PATH, so a tool installed where osmedeus puts it is never reported missing.
//
// Splitting the report by whether the registry knows the name matters: a registry
// entry just needs installing, while a name the registry has never heard of is
// usually a typo in the workflow.
func checkWorkflowDependencies(printer *terminal.Printer, cfg *config.Config, workflows []*core.Workflow) bool {
	printer.Section("Workflow Dependencies")

	if len(workflows) == 0 {
		printer.Info("  No workflows to check")
		return false
	}

	// command -> set of workflows that require it
	required := make(map[string]map[string]bool)
	for _, wf := range workflows {
		if wf.Dependencies == nil {
			continue
		}
		for _, cmdName := range wf.Dependencies.Commands {
			cmdName = strings.TrimSpace(cmdName)
			if cmdName == "" {
				continue
			}
			if required[cmdName] == nil {
				required[cmdName] = make(map[string]bool)
			}
			required[cmdName][wf.Name] = true
		}
	}

	if len(required) == 0 {
		printer.Info("  No workflow declares a command dependency")
		return false
	}

	// The registry is best-effort: without it we can still report what is missing,
	// just not whether it is installable.
	registry, regErr := installer.LoadRegistry("", nil)
	if regErr != nil {
		printer.Warning("  Could not load the binary registry (%v); reporting availability only", regErr)
	}

	names := make([]string, 0, len(required))
	for name := range required {
		names = append(names, name)
	}
	sort.Strings(names)

	var missingInstallable, missingUnknown []string
	for _, name := range names {
		// Same resolution order the executor uses, so this predicts the run-time gate.
		if _, err := utils.LookPathWithBinaries(name, cfg.BinariesPath); err == nil {
			continue
		}
		// Indexing a nil map is safe, so no registry means every name is unknown.
		if _, known := registry[name]; known {
			missingInstallable = append(missingInstallable, name)
		} else {
			missingUnknown = append(missingUnknown, name)
		}
	}

	available := len(names) - len(missingInstallable) - len(missingUnknown)
	printer.Success("  %d of %d declared commands available", available, len(names))

	requiredBy := func(name string) string {
		owners := make([]string, 0, len(required[name]))
		for owner := range required[name] {
			owners = append(owners, owner)
		}
		sort.Strings(owners)
		return terminal.Gray(strings.Join(owners, ", "))
	}

	if len(missingInstallable) > 0 {
		fmt.Println()
		printer.Warning("  Missing, but in the registry:")
		for _, name := range missingInstallable {
			printer.Warning("    %s - required by %s", terminal.White(name), requiredBy(name))
		}
		printer.Info("    Install them: %s", terminal.Yellow("osmedeus install binary --name "+strings.Join(missingInstallable, " --name ")))
	}

	if len(missingUnknown) > 0 {
		fmt.Println()
		printer.Error("  Missing and NOT in the registry (likely a typo, or a tool you must install yourself):")
		for _, name := range missingUnknown {
			printer.Error("    %s - required by %s", terminal.White(name), requiredBy(name))
		}
	}

	// Only an unknown name is an error: it cannot be resolved by installing, and a
	// workflow declaring it fails its dependency check every time it runs.
	return len(missingUnknown) > 0
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
	return copyEmbeddedTree("examples/osmedeus-base.example", dest)
}

// copyEmbeddedTree materializes an embedded subtree onto disk, creating parent
// directories as needed. Shared by the base-folder installer and the skills
// installer so both get the same directory/permission semantics.
func copyEmbeddedTree(srcRoot, dest string) error {
	srcFS := public.EmbedFS

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

package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/public"
	"github.com/spf13/cobra"
)

var (
	registryPath            string
	installAll              bool
	binaryNames             []string
	checkOnly               bool
	customHeaders           []string
	nixBuildInstall         bool
	nixInstallation         bool
	nixPkgs                 []string
	installOptional         bool
	hideBinaryTags          bool
	baseSample              bool
	basePreset              bool
	workflowPreset          bool
	validateSample          bool
	validatePreset          bool
	installEnvAll           bool
	goGetterSources         []string
	goGetterDest            string
	listRegistryNixBuild    bool
	listRegistryDirectFetch bool
)

// installCmd represents the install parent command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install workflows, base folder, or binaries",
	Long:  UsageInstall(),
}

// installWorkflowCmd installs workflows from a source
var installWorkflowCmd = &cobra.Command{
	Use:   "workflow [source]",
	Short: "Install workflows from git URL, zip URL, local zip file, or local folder",
	Args: func(cmd *cobra.Command, args []string) error {
		if workflowPreset {
			if len(args) != 0 {
				return fmt.Errorf("no source argument is allowed with --preset")
			}
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: RunInstallWorkflow,
}

// installBaseCmd installs the base folder from a source
var installBaseCmd = &cobra.Command{
	Use:   "base [source]",
	Short: "Install base folder from git URL, zip URL, local zip file, or local folder",
	Args: func(cmd *cobra.Command, args []string) error {
		if baseSample || basePreset {
			if len(args) != 0 {
				return fmt.Errorf("no source argument is allowed with --sample or --preset")
			}
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: runInstallBase,
}

// installBinaryCmd installs binaries from a registry
var installBinaryCmd = &cobra.Command{
	Use:     "binary",
	Aliases: []string{"bin", "binray", "binr", "birary"},
	Short:   "Install binary tools from registry",
	Long:    `Install one or more binary tools from the registry. Skips binaries already available in PATH.`,
	Example: `  # List binaries in registry
  osmedeus install binary --list-registry-direct-fetch
  osmedeus install binary --list-registry-nix-build

  # Install specific binaries
  osmedeus install binary --name nuclei
  osmedeus install binary --name nuclei --name ffuf --name httpx
  osmedeus install binary -n nuclei -n ffuf

  # Install all required binaries
  osmedeus install binary --all

  # Install all binaries including optional
  osmedeus install binary --all --install-optional

  # Check binary installation status
  osmedeus install binary --name amass --check
  osmedeus install binary --all --check

  # Install via go install
  osmedeus install binary --name nuclei --go-install
  osmedeus install binary --all --go-install
  osmedeus install binary --go-install-pkg github.com/tomnomnom/waybackurls@latest

  # Install via Nix
  osmedeus install binary --nix-installation
  osmedeus install binary --name nuclei --nix-build-install
  osmedeus install binary --all --nix-build-install`,
	RunE: runInstallBinary,
}

// installEnvCmd adds binaries path to shell configuration
var installEnvCmd = &cobra.Command{
	Use:     "env",
	Short:   "Add binaries path to shell configuration",
	Long:    `Add the osmedeus binaries folder to your PATH in shell configuration files (~/.bashrc, ~/.zshrc, ~/.profile).`,
	Example: `  osmedeus install env`,
	RunE:    runInstallEnv,
}

// installValidateCmd checks environment health
var installValidateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"val"},
	Short:   "Check and fix environment health (folders, config, workflows)",
	Long: `Validate the osmedeus installation by checking:
  - Required folders exist (base, workspaces, workflows, binaries, data)
  - Configuration files are valid
  - Workflows can be loaded and parsed correctly

This is the primary command for health checks. 'osmedeus health' is an alias for this command.`,
	Example: `  osmedeus install validate`,
	RunE:    runInstallValidate,
}

// RunInstallWorkflow installs workflows from a source (exported for use by workflow install alias)
func RunInstallWorkflow(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	headers := parseCustomHeaders(customHeaders)
	inst := installer.NewInstaller(
		cfg.BaseFolder,
		cfg.WorkflowsPath,
		cfg.BinariesPath,
		headers,
	)

	printer := terminal.NewPrinter()

	// Check if workflow folder exists and prompt for confirmation
	if _, err := os.Stat(inst.WorkflowFolder); err == nil {
		if !globalForce {
			printer.Warning("Existing workflow folder detected!")
			printer.Warning("Path: %s", inst.WorkflowFolder)
			printer.Warning("This operation will REMOVE the existing workflow folder and all its contents.")
			fmt.Println()

			if !confirmPrompt("Do you want to continue?") {
				printer.Info("Operation cancelled. Use --force to skip this confirmation.")
				return nil
			}
		}
	}

	if workflowPreset {
		workflowURL := os.Getenv("OSM_WORKFLOW_URL")
		if workflowURL != "" {
			printer.Info("Using workflow URL from OSM_WORKFLOW_URL environment variable")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(workflowURL))
		} else {
			workflowURL = core.DEFAULT_WORKFLOW_REPO
			printer.Info("Using default workflow URL")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(workflowURL))
		}

		if err := inst.InstallWorkflow(workflowURL); err != nil {
			return err
		}
		printWorkflowSummary(printer, cfg.WorkflowsPath)
		return nil
	}

	source := args[0]
	if err := inst.InstallWorkflow(source); err != nil {
		return err
	}
	printWorkflowSummary(printer, cfg.WorkflowsPath)
	return nil
}

// printWorkflowSummary counts and prints the number of workflows loaded
func printWorkflowSummary(printer *terminal.Printer, workflowsPath string) {
	loader := parser.NewLoader(workflowsPath)

	flows, flowErr := loader.ListFlows()
	modules, modErr := loader.ListModules()

	if flowErr == nil && modErr == nil {
		total := len(flows) + len(modules)
		if total > 0 {
			printer.Info("Loaded %s workflows (%s flows, %s modules)",
				terminal.Green(fmt.Sprintf("%d", total)),
				terminal.Cyan(fmt.Sprintf("%d", len(flows))),
				terminal.Yellow(fmt.Sprintf("%d", len(modules))))
			printer.Println("  %s Run %s to see workflow details",
				terminal.Gray(terminal.SymbolLightning),
				terminal.Cyan("osmedeus workflow ls"))
		}
	}
}

func runInstallBase(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	printer := terminal.NewPrinter()

	// Determine binaries folder for PATH setup
	binariesFolder := cfg.BinariesPath
	if binariesFolder == "" {
		binariesFolder = filepath.Join(cfg.BaseFolder, "external-binaries")
	}

	if baseSample {
		if err := replaceBaseFolderWithEmbeddedSample(cfg.BaseFolder); err != nil {
			return err
		}
		reloaded, err := config.Load(cfg.BaseFolder)
		if err == nil {
			config.Set(reloaded)
			// Update binariesFolder from reloaded config
			if reloaded.BinariesPath != "" {
				binariesFolder = reloaded.BinariesPath
			}
		}
		ensureBinariesPathInEnv(printer, binariesFolder, true)
		return nil
	}

	if basePreset {
		// Get preset URL from environment or use default
		presetURL := os.Getenv("OSM_PRESET_URL")
		if presetURL != "" {
			printer.Info("Using preset URL from OSM_PRESET_URL environment variable")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(presetURL))
		} else {
			presetURL = core.DEFAULT_BASE_REPO
			printer.Info("Using default preset URL")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(presetURL))
		}

		headers := parseCustomHeaders(customHeaders)
		inst := installer.NewInstaller(
			cfg.BaseFolder,
			cfg.WorkflowsPath,
			cfg.BinariesPath,
			headers,
		)
		if err := inst.InstallBase(presetURL); err != nil {
			return err
		}
		printer.Success("Base installed from: %s", terminal.Cyan(presetURL))
		printer.Newline()

		// Install workflows from OSM_WORKFLOW_URL or DEFAULT_WORKFLOW_REPO
		workflowURL := os.Getenv("OSM_WORKFLOW_URL")
		if workflowURL != "" {
			printer.Info("Using workflow URL from OSM_WORKFLOW_URL environment variable")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(workflowURL))
		} else {
			workflowURL = core.DEFAULT_WORKFLOW_REPO
			printer.Info("Using default workflow URL")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(workflowURL))
		}

		if err := inst.InstallWorkflow(workflowURL); err != nil {
			printer.Warning("Failed to install workflows: %s", err)
			// Continue - workflow installation failure shouldn't block base setup
		} else {
			printer.Success("Workflows installed from: %s", terminal.Cyan(workflowURL))
			printWorkflowSummary(printer, cfg.WorkflowsPath)
		}
		printer.Newline()

		// Reload config after base and workflow installation
		reloaded, err := config.Load(cfg.BaseFolder)
		if err == nil {
			config.Set(reloaded)
			cfg = reloaded
			if reloaded.BinariesPath != "" {
				binariesFolder = reloaded.BinariesPath
			}
		}

		// Check if first-time setup is needed (install binaries if so)
		if isFirstTimeSetupNeeded(cfg.BaseFolder) {
			printer.Println("%s %s", terminal.BoldBlue(terminal.SymbolLightning), terminal.HiBlue("First-time setup detected. Installing binaries..."))
			printer.Newline()

			// Install required binaries
			installRequiredBinaries(cfg, printer)

			// Create initialization marker
			_ = createInitializationMarker(printer)

			printer.Newline()
			printer.Println("%s %s", terminal.Green(terminal.SymbolSuccess), terminal.BoldGreen("First-time setup complete!"))
		}

		ensureBinariesPathInEnv(printer, binariesFolder, true)
		return nil
	}

	source := args[0]
	headers := parseCustomHeaders(customHeaders)
	inst := installer.NewInstaller(
		cfg.BaseFolder,
		cfg.WorkflowsPath,
		cfg.BinariesPath,
		headers,
	)

	if err := inst.InstallBase(source); err != nil {
		return err
	}

	// Reload config to get updated paths after base installation
	reloaded, err := config.Load(cfg.BaseFolder)
	if err == nil {
		config.Set(reloaded)
		cfg = reloaded
		if reloaded.BinariesPath != "" {
			binariesFolder = reloaded.BinariesPath
		}
	}

	// Check if workflows folder exists under the base folder and print stats
	printWorkflowSummary(printer, cfg.WorkflowsPath)

	ensureBinariesPathInEnv(printer, binariesFolder, true)
	return nil
}

func runInstallBinary(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	printer := terminal.NewPrinter()

	// Auto-detect Nix installation and add to PATH if not already
	// Check if /nix/var/nix/profiles/default/bin is not empty
	nixBinDir := "/nix/var/nix/profiles/default/bin"
	if entries, err := os.ReadDir(nixBinDir); err == nil && len(entries) > 0 {
		ensureNixProfileBinInProcess(nil) // Silent - no printer output for auto-detection
	}

	// Handle --nix-installation flag first
	if nixInstallation {
		silent = true // Nix installation produces verbose output, enable silent by default
		hasFollowupAction := nixBuildInstall || len(nixPkgs) > 0 || installAll || len(binaryNames) > 0 || checkOnly || listRegistryNixBuild || listRegistryDirectFetch

		if installer.IsNixInstalled() {
			printer.Info("Nix already available in PATH")
		} else {
			printer.Println("%s Running Nix install command:", terminal.BoldCyan("◆"))
			printer.GrayOutput(installer.NixInstallCommandPrettyForHost())
			printer.Newline()
		}

		printer.Info("Installing Nix package manager...")
		if err := installer.InstallNix(); err != nil {
			printer.Error("Failed to install Nix: %s", err)
			return err
		}
		printer.Success("Nix installed successfully!")
		ensureNixProfileBinInProcess(printer)
		ensureNixProfileBinInShell(printer)
		if !hasFollowupAction {
			return nil
		}
	}

	if nixBuildInstall {
		ensureNixProfileBinInProcess(printer)
	}

	// Check if Nix build install is requested but Nix isn't installed
	if (nixBuildInstall || len(nixPkgs) > 0) && !installer.IsNixInstalled() {
		printer.Error("Nix is not installed")
		printer.Println("%s Install Nix with: %s", terminal.BoldCyan("◆"), terminal.Gray("osmedeus install binary --nix-installation"))
		printer.Info("Or manually run:")
		printer.GrayOutput(installer.NixInstallCommandPrettyForHost())
		printer.Newline()
		printer.Println("%s Ensure Nix is in PATH:", terminal.BoldCyan("◆"))
		printer.Println("  %s", terminal.Gray("export PATH=\"$PATH:/nix/var/nix/profiles/default/bin\""))
		return fmt.Errorf("nix not installed")
	}

	// Handle --list-registry-nix-build flag
	if listRegistryNixBuild {
		return printNixBinaries(printer, globalWidth, !hideBinaryTags)
	}

	// Handle --list-registry-direct-fetch flag
	if listRegistryDirectFetch {
		headers := parseCustomHeaders(customHeaders)
		return printRegistryBinaries(registryPath, headers, printer, globalWidth, !hideBinaryTags)
	}

	// Handle --go-getter flag (download/clone via go-getter)
	if len(goGetterSources) > 0 {
		if checkOnly {
			return fmt.Errorf("cannot use --check with --go-getter")
		}
		if installAll || len(binaryNames) > 0 {
			return fmt.Errorf("cannot combine --go-getter with --all or --name")
		}
		if nixBuildInstall {
			return fmt.Errorf("cannot combine --go-getter with --nix-build-install")
		}

		var lastErr error
		var failed []string
		for _, srcArg := range goGetterSources {
			srcArg = strings.TrimSpace(srcArg)
			if srcArg == "" {
				continue
			}

			// Parse inline destination if present (format: "source destination")
			src := srcArg
			dest := goGetterDest
			if strings.Contains(srcArg, " ") {
				parts := strings.SplitN(srcArg, " ", 2)
				src = strings.TrimSpace(parts[0])
				dest = strings.TrimSpace(parts[1])
			}

			// Use default destination if not specified
			if dest == "" {
				homeDir, _ := os.UserHomeDir()
				// Extract repo name from source URL for default destination
				repoName := extractRepoNameFromURL(src)
				if repoName != "" {
					dest = filepath.Join(homeDir, repoName)
				} else {
					dest = homeDir
				}
			} else {
				// Expand ~ and $HOME in destination
				dest = installer.ExpandPath(dest)
			}

			printer.Println("%s Destination: %s", terminal.BoldCyan("◆"), terminal.Gray(dest))
			printer.Installing(src)
			if err := installer.GetViaGoGetter(src, dest); err != nil {
				printer.Error("Failed to download %s: %s", src, err)
				lastErr = err
				failed = append(failed, src)
			} else {
				printer.Success("Downloaded '%s' to %s", src, dest)
			}
		}
		printFailedBinariesSummary(printer, failed)
		return lastErr
	}

	if len(nixPkgs) > 0 {
		if checkOnly {
			return fmt.Errorf("cannot use --check with --nix-pkgs")
		}
		if installAll || len(binaryNames) > 0 {
			return fmt.Errorf("cannot combine --nix-pkgs with --all or --name")
		}
	}

	if shouldInitializeBaseForBinary(cfg.BaseFolder) {
		printer.Info("Base folder not initialized, running: %s", terminal.Gray("osmedeus install base --sample"))
		if err := replaceBaseFolderWithEmbeddedSample(cfg.BaseFolder); err != nil {
			return err
		}
		reloaded, err := config.Load(cfg.BaseFolder)
		if err == nil {
			config.Set(reloaded)
			cfg = reloaded
		}
	}

	// Validate: either --all, --name, or --nix-pkgs must be specified
	if !installAll && len(binaryNames) == 0 && len(nixPkgs) == 0 {
		printer.Error("Please specify binary name(s) with --name, Nix packages with --nix-pkgs, or use --all")
		return fmt.Errorf("either --all, --name, or --nix-pkgs flag is required")
	}

	// Determine binaries folder
	binariesFolder := cfg.BinariesPath
	if binariesFolder == "" {
		binariesFolder = filepath.Join(cfg.BaseFolder, "binaries")
	}

	headers := parseCustomHeaders(customHeaders)

	// Load registry once for all operations
	registry, err := installer.LoadRegistry(registryPath, headers)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	registrySource := ""
	if registryPath == "" {
		registrySource = "embedded preset: public/presets/registry-metadata-direct-fetch.json"
	} else if installer.IsURL(registryPath) {
		registrySource = registryPath
	} else {
		registrySource = registryPath
	}
	printer.Println("%s Registry metadata: %s", terminal.BoldCyan("◆"), terminal.Gray(registrySource))

	// Show security warning before installation (skip for check-only mode)
	if !checkOnly {
		printer.SecurityWarning("This command will download and execute external binaries")
		printer.BulletColored("github-release: Downloads binaries from URLs in the registry", terminal.Yellow)
		printer.BulletColored("command-based: Executes commands like 'pip install' or 'brew install'", terminal.Yellow)
		printer.Println("%s %s", terminal.BoldCyan("◆"), terminal.Gray("Only run with trusted registry sources. Review registry metadata before proceeding."))
		printer.Newline()
	}

	// Handle --all flag
	if installAll {
		var allNames []string
		var skippedOptional int

		if nixBuildInstall {
			flakeContent, err := public.GetFlakeNix()
			if err != nil {
				return fmt.Errorf("failed to read embedded flake.nix: %w", err)
			}
			categories, err := installer.ParseFlakeNixBinariesFromString(string(flakeContent))
			if err != nil {
				return fmt.Errorf("failed to parse flake.nix: %w", err)
			}
			for _, name := range installer.GetAllFlakeBinaries(categories) {
				if entry, ok := registry[name]; ok {
					if !installOptional && containsTag(entry.Tags, "optional") {
						skippedOptional++
						continue
					}
				}
				allNames = append(allNames, name)
			}
		} else {
			for name, entry := range registry {
				if !installOptional && containsTag(entry.Tags, "optional") {
					skippedOptional++
					continue
				}
				allNames = append(allNames, name)
			}
			sort.Strings(allNames)
		}
		if skippedOptional > 0 {
			printOptionalBinarySkipMessage(printer, skippedOptional, nixBuildInstall)
		}
		if installOptional {
			printer.Println("%s %s", terminal.BoldCyan("◆"), terminal.Gray("Optional binaries may require extra tools in your $PATH (e.g., go, pip)."))
		}

		if checkOnly {
			// Check-only mode: just verify binary availability for all binaries
			return checkBinaries(allNames, registry, printer)
		}
		if nixBuildInstall {
			failed, err := installBinariesViaNix(allNames, registry, binariesFolder, printer)
			printFailedBinariesSummary(printer, failed)
			ensureBinariesPathInEnv(printer, binariesFolder, true)
			return err
		}

		// Install binaries in parallel (suppress spinner display)
		silent = true
		printer.Println("  %s Installing %s binaries to: %s",
			terminal.SymbolBullet,
			terminal.Green(fmt.Sprintf("%d", len(allNames))),
			terminal.Cyan(binariesFolder))
		printer.Newline()

		failed := installBinariesParallel(allNames, registry, binariesFolder, headers, printer, silent)
		printer.Newline()
		printFailedBinariesSummary(printer, failed)
		ensureBinariesPathInEnv(printer, binariesFolder, true)
		if len(failed) > 0 {
			return fmt.Errorf("%d binaries failed to install", len(failed))
		}
		return nil
	}

	if len(nixPkgs) > 0 {
		silent = true // Nix profile add produces verbose output, enable silent by default
		ensureNixProfileBinInProcess(printer)
		ensureNixProfileBinInShell(printer)

		var lastErr error
		var failed []string
		for _, pkg := range nixPkgs {
			pkg = strings.TrimSpace(pkg)
			if pkg == "" {
				continue
			}
			installable := pkg
			if !strings.Contains(installable, "#") {
				printer.Warning("Package '%s' doesn't contain '#'. Using '%s' format.", pkg, "nixpkgs#"+pkg)
				printer.Info("Tip: Use format 'nixpkgs#%s' or 'flake#package' directly", pkg)
				installable = "nixpkgs#" + pkg
			}

			printer.Installing(installable)
			if !silent {
				printer.GrayOutput("nix profile add " + installable)
			}
			if err := installer.InstallBinaryViaNix(pkg, pkg, ""); err != nil {
				printer.Error("Failed to add %s: %s", pkg, err)
				lastErr = err
				failed = append(failed, pkg)
			} else {
				if !silent {
					if installer.NixInstallOutput != "" {
						printer.GrayOutput(installer.NixInstallOutput)
					}
				}
				printer.Success("Nix package '%s' added", pkg)
			}
		}
		printFailedBinariesSummary(printer, failed)
		return lastErr
	}

	// Check-only mode: just verify binary availability
	if checkOnly {
		return checkBinaries(binaryNames, registry, printer)
	}

	// Install via Nix if --nix-build-install is set
	if nixBuildInstall {
		failed, err := installBinariesViaNix(binaryNames, registry, binariesFolder, printer)
		printFailedBinariesSummary(printer, failed)
		ensureBinariesPathInEnv(printer, binariesFolder, true)
		return err
	}

	// Install each specified binary
	var lastErr error
	var failed []string
	for _, name := range binaryNames {
		printer.Info("Processing binary: %s", name)
		if err := installer.InstallBinary(name, registry, binariesFolder, headers); err != nil {
			printer.Error("Failed to install %s: %s", name, err)
			lastErr = err
			failed = append(failed, name)
		} else {
			printer.Success("Binary '%s' ready", terminal.HiBlue(name))
		}
	}
	printFailedBinariesSummary(printer, failed)
	ensureBinariesPathInEnv(printer, binariesFolder, true)

	return lastErr
}

func shouldInitializeBaseForBinary(baseFolder string) bool {
	if baseFolder == "" {
		return false
	}
	info, err := os.Stat(baseFolder)
	if err != nil {
		return os.IsNotExist(err)
	}
	if !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir(baseFolder)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == "osm-settings.yaml" || name == ".DS_Store" {
			continue
		}
		return false
	}
	return true
}

// installBinariesViaNix installs binaries using Nix and copies them to binariesFolder
func installBinariesViaNix(names []string, registry installer.BinaryRegistry, binariesFolder string, printer *terminal.Printer) ([]string, error) {
	ensureNixProfileBinInProcess(printer)
	ensureNixProfileBinInShell(printer)

	printer.Println("%s %s", terminal.BoldCyan(terminal.SymbolLightning), terminal.HiWhite("Nix builds packages from source, ensuring reproducibility. It might take longer, but the results are well worth the wait"))
	printer.Newline()

	var lastErr error
	var failed []string
	for _, name := range names {
		// Always show the binary name being processed
		printer.Info("Processing binary: %s", terminal.HiBlue(name))

		// Check if already in PATH
		if installer.IsBinaryInPath(name) {
			printer.Info("Binary '%s' already available in PATH, skipping", terminal.HiBlue(name))
			continue
		}

		entry, ok := registry[name]
		if !ok {
			printer.Warning("Binary '%s' not found in registry, trying direct Nix install", name)
		}

		nixPkg := installer.GetNixPackageName(entry, name)

		// Show installing message with package name
		printer.Installing(fmt.Sprintf("nixpkgs#%s", nixPkg))
		if !silent {
			printer.GrayOutput(fmt.Sprintf("nix profile add nixpkgs#%s", nixPkg))
		}

		if err := installer.InstallBinaryViaNix(name, nixPkg, binariesFolder); err != nil {
			printer.Error("Failed to install %s: %s", name, err)
			lastErr = err
			failed = append(failed, name)
		} else {
			if !silent {
				if installer.NixInstallOutput != "" {
					printer.GrayOutput(installer.NixInstallOutput)
				}
			}
			printer.Success("Binary '%s' installed via Nix and copied to %s", terminal.HiBlue(name), binariesFolder)
		}
	}
	return failed, lastErr
}

func printOptionalBinarySkipMessage(printer *terminal.Printer, skippedOptional int, nixMode bool) {
	flag := terminal.Gray("--install-optional")
	printer.Println("%s Skipping %d optional binaries (use %s to include them)", terminal.BoldCyan("◆"), skippedOptional, flag)
	printer.Println("  %s", terminal.Gray("Optional binaries may require extra tools in your $PATH (e.g., go, pip)."))

	cmd := "osmedeus install binary --all --install-optional"
	if nixMode {
		cmd = "osmedeus install binary --all --nix-build-install --install-optional"
	}
	printer.Println("  %s", terminal.Gray("Example: "+cmd))
}

func printFailedBinariesSummary(printer *terminal.Printer, failed []string) {
	if len(failed) == 0 {
		return
	}

	unique := make(map[string]struct{}, len(failed))
	for _, name := range failed {
		unique[name] = struct{}{}
	}

	list := make([]string, 0, len(unique))
	for name := range unique {
		list = append(list, name)
	}
	sort.Strings(list)

	printer.Newline()
	printer.Section("Failed Binaries")
	for _, name := range list {
		printer.Bullet(name)
	}
	printer.Newline()

	printer.Println("%s %s", terminal.BoldCyan(terminal.SymbolLightning), terminal.Gray("Please note that your workflow may not require all binaries listed in the registry, so the scan can "+terminal.Green("still function properly")+terminal.Gray(" even if some tools are absent.")))
	printer.Println("%s %s", terminal.BoldCyan(terminal.SymbolLightning), terminal.Gray("The registry is intended to provide all external tools that are worthwhile to include in your workflow and is not mandatory."))
}

// printNixBinaries prints the list of binaries available in Nix flake
func printNixBinaries(printer *terminal.Printer, descWidth int, showTags bool) error {
	printer.Println(terminal.BoldCyan("Available binaries in Nix flake"))
	// Get embedded flake.nix content
	flakeContent, err := public.GetFlakeNix()
	if err != nil {
		return fmt.Errorf("failed to read embedded flake.nix: %w", err)
	}

	// Parse the flake to extract tool categories
	categories, err := installer.ParseFlakeNixBinariesFromString(string(flakeContent))
	if err != nil {
		return fmt.Errorf("failed to parse flake.nix: %w", err)
	}

	// Load registry for descriptions and tags
	registry, _ := installer.LoadRegistry("", nil) // Ignore error, descriptions are optional

	// Column widths
	nameWidth := 18
	statusWidth := 10
	tagsWidth := 30

	// Print header
	printer.Println(terminal.BoldCyan("## Nix Flake Binaries (nix-build mode)"))
	printer.Println(terminal.Gray("Source: public/presets/flake.nix"))
	printer.Newline()

	totalTools := 0
	for _, cat := range categories {
		printer.Println(terminal.BoldGreen("### %s (%d)"), cat.Name, len(cat.Tools))
		printer.Newline()

		// Print table header
		if showTags {
			printer.Println("| %s | %s | %s | %s |",
				padRight("Name", nameWidth),
				padRight("Status", statusWidth),
				padRight("Description", descWidth),
				padRight("Tags", tagsWidth))
			printer.Println("|%s|%s|%s|%s|",
				strings.Repeat("-", nameWidth+2),
				strings.Repeat("-", statusWidth+2),
				strings.Repeat("-", descWidth+2),
				strings.Repeat("-", tagsWidth+2))
		} else {
			printer.Println("| %s | %s | %s |",
				padRight("Name", nameWidth),
				padRight("Status", statusWidth),
				padRight("Description", descWidth))
			printer.Println("|%s|%s|%s|",
				strings.Repeat("-", nameWidth+2),
				strings.Repeat("-", statusWidth+2),
				strings.Repeat("-", descWidth+2))
		}

		for _, tool := range cat.Tools {
			statusText := "missing"
			statusColor := terminal.Red

			// Get registry entry for validation command and metadata
			var entryPtr *installer.BinaryEntry
			desc := ""
			tags := ""
			if registry != nil {
				if entry, ok := registry[tool]; ok {
					entryPtr = &entry
					desc = entry.Desc
					tags = strings.Join(entry.Tags, ", ")
				}
			}

			if installer.IsBinaryInstalled(tool, entryPtr) {
				statusText = "installed"
				statusColor = terminal.Green
			}
			desc = truncatePad(desc, descWidth)
			tags = truncatePad(tags, tagsWidth)

			if showTags {
				printer.Println("| %s | %s | %s | %s |",
					colorPadRight(tool, nameWidth, terminal.Cyan),
					colorPadRight(statusText, statusWidth, statusColor),
					desc,
					tags)
			} else {
				printer.Println("| %s | %s | %s |",
					colorPadRight(tool, nameWidth, terminal.Cyan),
					colorPadRight(statusText, statusWidth, statusColor),
					desc)
			}
		}
		totalTools += len(cat.Tools)
		printer.Newline()
	}

	// Print usage examples
	printer.Println(terminal.BoldCyan("### Usage Examples"))
	printer.Println("  " + terminal.Green("# Enter Nix development shell with all tools"))
	printer.Println("  nix develop")
	printer.Newline()
	printer.Println("  " + terminal.Green("# Install specific binary via Nix"))
	printer.Println("  osmedeus install binary --name nuclei --nix-build-install")
	printer.Newline()
	printer.Println("  " + terminal.Green("# Install all binaries via Nix"))
	printer.Println("  osmedeus install binary --all --nix-build-install")

	printer.Newline()
	printer.Println(terminal.Gray("Total: %d binaries in %d categories"), totalTools, len(categories))
	return nil
}

// printRegistryBinaries prints the list of binaries available in registry JSON
func printRegistryBinaries(registryPath string, headers map[string]string, printer *terminal.Printer, descWidth int, showTags bool) error {
	printer.Println(terminal.BoldCyan("Available binaries in registry"))
	registry, err := installer.LoadRegistry(registryPath, headers)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	// Separate required and optional binaries
	var required, optional []string
	for name, entry := range registry {
		if containsTag(entry.Tags, "optional") {
			optional = append(optional, name)
		} else {
			required = append(required, name)
		}
	}
	sort.Strings(required)
	sort.Strings(optional)

	// Column widths
	nameWidth := 18
	statusWidth := 10
	tagsWidth := 30

	// Print header
	printer.Println(terminal.BoldCyan("## Registry Binaries (direct-fetch mode)"))
	printer.Newline()

	// Print required binaries as markdown table
	printer.Println(terminal.BoldGreen("### Required Binaries (%d)"), len(required))
	printer.Newline()

	// Print table header
	if showTags {
		printer.Println("| %s | %s | %s | %s |",
			padRight("Name", nameWidth),
			padRight("Status", statusWidth),
			padRight("Description", descWidth),
			padRight("Tags", tagsWidth))
		printer.Println("|%s|%s|%s|%s|",
			strings.Repeat("-", nameWidth+2),
			strings.Repeat("-", statusWidth+2),
			strings.Repeat("-", descWidth+2),
			strings.Repeat("-", tagsWidth+2))
	} else {
		printer.Println("| %s | %s | %s |",
			padRight("Name", nameWidth),
			padRight("Status", statusWidth),
			padRight("Description", descWidth))
		printer.Println("|%s|%s|%s|",
			strings.Repeat("-", nameWidth+2),
			strings.Repeat("-", statusWidth+2),
			strings.Repeat("-", descWidth+2))
	}

	// Print required binaries
	for _, name := range required {
		entry := registry[name]
		statusText := "missing"
		statusColor := terminal.Red
		if installer.IsBinaryInstalled(name, &entry) {
			statusText = "installed"
			statusColor = terminal.Green
		}
		desc := truncatePad(entry.Desc, descWidth)

		if showTags {
			tags := truncatePad(strings.Join(entry.Tags, ", "), tagsWidth)
			printer.Println("| %s | %s | %s | %s |",
				colorPadRight(name, nameWidth, terminal.Cyan),
				colorPadRight(statusText, statusWidth, statusColor),
				desc,
				tags)
		} else {
			printer.Println("| %s | %s | %s |",
				colorPadRight(name, nameWidth, terminal.Cyan),
				colorPadRight(statusText, statusWidth, statusColor),
				desc)
		}
	}

	// Print optional binaries section
	printer.Newline()
	printer.Println(terminal.BoldYellow("### Optional Binaries (%d)"), len(optional))
	printer.Println(terminal.Gray("(use --install-optional to include these)"))
	printer.Newline()

	// Print table header for optional
	if showTags {
		printer.Println("| %s | %s | %s | %s |",
			padRight("Name", nameWidth),
			padRight("Status", statusWidth),
			padRight("Description", descWidth),
			padRight("Tags", tagsWidth))
		printer.Println("|%s|%s|%s|%s|",
			strings.Repeat("-", nameWidth+2),
			strings.Repeat("-", statusWidth+2),
			strings.Repeat("-", descWidth+2),
			strings.Repeat("-", tagsWidth+2))
	} else {
		printer.Println("| %s | %s | %s |",
			padRight("Name", nameWidth),
			padRight("Status", statusWidth),
			padRight("Description", descWidth))
		printer.Println("|%s|%s|%s|",
			strings.Repeat("-", nameWidth+2),
			strings.Repeat("-", statusWidth+2),
			strings.Repeat("-", descWidth+2))
	}

	// Print optional binaries
	for _, name := range optional {
		entry := registry[name]
		statusText := "missing"
		statusColor := terminal.Red
		if installer.IsBinaryInstalled(name, &entry) {
			statusText = "installed"
			statusColor = terminal.Green
		}
		desc := truncatePad(entry.Desc, descWidth)

		if showTags {
			tags := truncatePad(strings.Join(entry.Tags, ", "), tagsWidth)
			printer.Println("| %s | %s | %s | %s |",
				colorPadRight(name, nameWidth, terminal.Yellow),
				colorPadRight(statusText, statusWidth, statusColor),
				desc,
				tags)
		} else {
			printer.Println("| %s | %s | %s |",
				colorPadRight(name, nameWidth, terminal.Yellow),
				colorPadRight(statusText, statusWidth, statusColor),
				desc)
		}
	}

	// Print usage examples
	printer.Newline()
	printer.Println(terminal.BoldCyan("### Usage Examples"))
	printer.Println("  " + terminal.Green("# Install specific binaries"))
	printer.Println("  osmedeus install binary --name nuclei --name httpx")
	printer.Newline()
	printer.Println("  " + terminal.Green("# Install all required binaries"))
	printer.Println("  osmedeus install binary --all")
	printer.Newline()
	printer.Println("  " + terminal.Green("# Install all binaries including optional"))
	printer.Println("  osmedeus install binary --all --install-optional")

	printer.Newline()
	printer.Println(terminal.Gray("Total: %d required, %d optional"), len(required), len(optional))
	return nil
}

// checkBinaries checks if binaries are installed without downloading
func checkBinaries(names []string, registry installer.BinaryRegistry, printer *terminal.Printer) error {
	ready := 0
	missing := 0

	for _, name := range names {
		entry, exists := registry[name]
		if !exists {
			printer.Warning("%s - not found in registry", name)
			continue
		}

		if installer.IsBinaryInstalled(name, &entry) {
			printer.Success("%s - ready", name)
			ready++
		} else {
			printer.Error("%s - not installed", name)
			missing++
		}
	}

	printer.Info("")
	printer.Info("Summary: %d/%d binaries ready", ready, ready+missing)

	if missing > 0 {
		return fmt.Errorf("%d binaries not installed", missing)
	}
	return nil
}

func runInstallValidate(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if validateSample && validatePreset {
		return fmt.Errorf("cannot use --sample and --preset together")
	}

	if validateSample {
		if err := replaceBaseFolderWithEmbeddedSample(cfg.BaseFolder); err != nil {
			return err
		}
		reloaded, err := config.Load(cfg.BaseFolder)
		if err == nil {
			config.Set(reloaded)
		}
	}

	if validatePreset {
		printer := terminal.NewPrinter()

		// Get preset URL from environment or use default
		presetURL := os.Getenv("OSM_PRESET_URL")
		if presetURL != "" {
			printer.Info("Using preset URL from OSM_PRESET_URL environment variable")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(presetURL))
		} else {
			presetURL = core.DEFAULT_BASE_REPO
			printer.Info("Using default preset URL")
			printer.Println("  %s %s", terminal.SymbolBullet, terminal.Gray(presetURL))
		}

		headers := parseCustomHeaders(customHeaders)
		inst := installer.NewInstaller(
			cfg.BaseFolder,
			cfg.WorkflowsPath,
			cfg.BinariesPath,
			headers,
		)
		if err := inst.InstallBase(presetURL); err != nil {
			return err
		}
		reloaded, err := config.Load(cfg.BaseFolder)
		if err == nil {
			config.Set(reloaded)
		}
	}

	return runHealth(cmd, args)
}

func replaceBaseFolderWithEmbeddedSample(baseFolder string) error {
	if baseFolder == "" {
		return fmt.Errorf("base folder is not configured")
	}
	if _, err := os.Stat(baseFolder); err == nil {
		if err := os.RemoveAll(baseFolder); err != nil {
			return fmt.Errorf("failed to remove existing base folder: %w", err)
		}
	}
	return copyEmbeddedAssets(baseFolder)
}

func runInstallEnv(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	printer := terminal.NewPrinter()

	// Determine binaries folder
	binariesFolder := cfg.BinariesPath
	if binariesFolder == "" {
		binariesFolder = filepath.Join(cfg.BaseFolder, "external-binaries")
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	shell := detectShell()
	primaryConfig := shellConfigForShell(homeDir, shell)

	// The export line to add
	exportLine := fmt.Sprintf(`export PATH="%s:$PATH"`, binariesFolder)
	comment := "# Added by osmedeus"

	printer.Println("%s Binaries folder: %s", terminal.BoldCyan("◆"), terminal.White(binariesFolder))
	printer.Println("%s Detected shell: %s", terminal.BoldCyan("◆"), terminal.White(shellDisplayName(shell)))

	configs := []string{primaryConfig}
	if installEnvAll {
		configs = allShellConfigs(homeDir, primaryConfig)
	}

	for _, configFile := range configs {
		updated, err := addOrUpdateShellConfigBlock(configFile, exportLine, comment)
		if err != nil {
			printer.Error("Failed to update %s: %s", terminal.White(configFile), err)
			continue
		}
		if updated {
			printer.Success("Added PATH to %s", terminal.White(configFile))
		} else {
			printer.Info("PATH already configured in %s", terminal.White(configFile))
		}
	}

	// Update current process PATH immediately
	currentPath := os.Getenv("PATH")
	if !pathContainsDir(currentPath, binariesFolder) {
		newPath := binariesFolder + string(os.PathListSeparator) + currentPath
		_ = os.Setenv("PATH", newPath)
		printer.Success("Updated PATH in current process")
	}

	// Always show the export command for the user's terminal session
	// (we can't modify the parent shell's environment, only our own process)
	printer.Println("%s For this terminal session, run: %s", terminal.BoldCyan("◆"),
		terminal.Gray(fmt.Sprintf(`export PATH="%s:$PATH"`, binariesFolder)))

	return nil
}

func detectShell() string {
	return strings.TrimSpace(os.Getenv("SHELL"))
}

func shellDisplayName(shell string) string {
	base := filepath.Base(shell)
	if base == "." || base == "/" {
		base = shell
	}
	base = strings.TrimSpace(base)
	if base == "" {
		return "unknown"
	}
	return base
}

func shellConfigForShell(homeDir, shell string) string {
	s := strings.ToLower(shell)
	zshrc := filepath.Join(homeDir, ".zshrc")
	bashrc := filepath.Join(homeDir, ".bashrc")
	bashProfile := filepath.Join(homeDir, ".bash_profile")
	profile := filepath.Join(homeDir, ".profile")

	if strings.Contains(s, "zsh") {
		return zshrc
	}
	if strings.Contains(s, "bash") {
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		if _, err := os.Stat(bashProfile); err == nil {
			return bashProfile
		}
		return bashrc
	}

	// Fallback when $SHELL is not set: prefer bashrc over zshrc
	// (bash is more common default, especially in Docker containers)
	if _, err := os.Stat(bashrc); err == nil {
		return bashrc
	}
	if _, err := os.Stat(bashProfile); err == nil {
		return bashProfile
	}
	if _, err := os.Stat(zshrc); err == nil {
		return zshrc
	}
	return profile
}

func allShellConfigs(homeDir, primary string) []string {
	paths := []string{}
	seen := map[string]struct{}{}

	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}

	add(primary)
	add(filepath.Join(homeDir, ".bashrc"))
	add(filepath.Join(homeDir, ".bash_profile"))
	add(filepath.Join(homeDir, ".zshrc"))
	add(filepath.Join(homeDir, ".profile"))

	return paths
}

func pathContainsDir(pathEnv, dir string) bool {
	if dir == "" {
		return false
	}
	dir = filepath.Clean(dir)
	for _, p := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		if filepath.Clean(strings.TrimSpace(p)) == dir {
			return true
		}
	}
	return false
}

// extractRepoNameFromURL extracts the repository name from a go-getter URL
// Examples:
//   - "github.com/user/repo.git?ref=main" -> "repo"
//   - "https://github.com/user/repo.git//subfolder" -> "repo"
//   - "github.com/projectdiscovery/nuclei-templates.git?ref=main&depth=1" -> "nuclei-templates"
func extractRepoNameFromURL(url string) string {
	// Remove query parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}
	// Remove subfolder reference
	if idx := strings.Index(url, "//"); idx != -1 {
		url = url[:idx]
	}
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")
	// Get the last path segment
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// ensureBinariesPathInEnv ensures the binaries folder is in PATH for both
// the current process and shell config. Returns true if PATH was already
// configured or successfully updated.
func ensureBinariesPathInEnv(printer *terminal.Printer, binariesFolder string, showExportHint bool) bool {
	if binariesFolder == "" {
		return true
	}

	// Skip shell config modification in test environments
	if os.Getenv("OSM_SKIP_PATH_SETUP") == "1" {
		// Still add to current process PATH
		if !pathContainsDir(os.Getenv("PATH"), binariesFolder) {
			currentPath := os.Getenv("PATH")
			newPath := binariesFolder + string(os.PathListSeparator) + currentPath
			_ = os.Setenv("PATH", newPath)
		}
		return true
	}

	// Check if already in PATH
	if pathContainsDir(os.Getenv("PATH"), binariesFolder) {
		return true
	}

	// Add to current process PATH immediately
	currentPath := os.Getenv("PATH")
	newPath := binariesFolder + string(os.PathListSeparator) + currentPath
	_ = os.Setenv("PATH", newPath)

	// Add to shell config for persistence
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if showExportHint {
			printer.Println("%s For this session: %s", terminal.BoldCyan("◆"),
				terminal.Gray(fmt.Sprintf(`export PATH="%s:$PATH"`, binariesFolder)))
		}
		return false
	}

	shell := detectShell()
	configFile := shellConfigForShell(homeDir, shell)
	exportLine := fmt.Sprintf(`export PATH="%s:$PATH"`, binariesFolder)
	comment := "# Added by osmedeus"

	updated, _ := addOrUpdateShellConfigBlock(configFile, exportLine, comment)
	if updated {
		printer.Success("Added PATH to %s", terminal.White(configFile))
	}

	if showExportHint {
		printer.Println("%s For this session: %s", terminal.BoldCyan("◆"),
			terminal.Gray(fmt.Sprintf(`export PATH="%s:$PATH"`, binariesFolder)))
	}

	return true
}

func ensureNixProfileBinInProcess(printer *terminal.Printer) {
	nixBinDir := "/nix/var/nix/profiles/default/bin"
	if pathContainsDir(os.Getenv("PATH"), nixBinDir) {
		return
	}
	if _, err := os.Stat(filepath.Join(nixBinDir, "nix")); err != nil {
		return
	}
	_ = os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+nixBinDir)
	if printer != nil {
		printer.Success("Added Nix binary path to current PATH: %s", terminal.White(nixBinDir))
	}
}

func ensureNixProfileBinInShell(printer *terminal.Printer) {
	// Skip in test environments
	if os.Getenv("OSM_SKIP_PATH_SETUP") == "1" {
		return
	}

	nixBinDir := "/nix/var/nix/profiles/default/bin"
	if _, err := os.Stat(filepath.Join(nixBinDir, "nix")); err != nil {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	shell := detectShell()
	primaryConfig := shellConfigForShell(homeDir, shell)
	exportLine := fmt.Sprintf(`export PATH="$PATH:%s"`, nixBinDir)
	comment := "# Added by osmedeus (nix)"

	updated, err := addToShellConfig(primaryConfig, exportLine, comment)
	if err != nil {
		if printer != nil {
			printer.Warning("Failed to update %s: %s", terminal.White(primaryConfig), err)
		}
		return
	}
	if updated {
		if printer != nil {
			printer.Success("Added Nix PATH to %s", terminal.White(primaryConfig))
		}
	}
}

// addToShellConfig adds the export line to a shell config file if not already present
func addToShellConfig(configFile, exportLine, comment string) (bool, error) {
	// Check if file exists
	content := ""
	if data, err := os.ReadFile(configFile); err == nil {
		content = string(data)
	}

	// Check if the export line already exists
	if strings.Contains(content, exportLine) {
		return false, nil
	}

	// Open file for appending (create if doesn't exist)
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}
	defer func() { _ = file.Close() }()

	// Write the export line with a comment
	writer := bufio.NewWriter(file)
	_, err = fmt.Fprintf(writer, "\n%s\n%s\n", comment, exportLine)
	if err != nil {
		return false, err
	}

	return true, writer.Flush()
}

func addOrUpdateShellConfigBlock(configFile, exportLine, comment string) (bool, error) {
	data, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	content := string(data)
	if strings.Contains(content, exportLine) {
		return false, nil
	}

	legacyComments := []string{"# Added by osmedeus", "# Added by osmedeus install env"}
	lines := strings.Split(content, "\n")

	for i := 0; i < len(lines); i++ {
		lineTrim := strings.TrimSpace(lines[i])
		matched := false
		for _, legacy := range legacyComments {
			if lineTrim == legacy {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		lines[i] = comment
		if i+1 >= len(lines) {
			lines = append(lines, exportLine)
		} else {
			nextTrim := strings.TrimSpace(lines[i+1])
			if strings.HasPrefix(nextTrim, "export PATH=") {
				lines[i+1] = exportLine
			} else {
				lines = append(lines[:i+1], append([]string{exportLine}, lines[i+1:]...)...)
			}
		}

		mode := os.FileMode(0644)
		if st, statErr := os.Stat(configFile); statErr == nil {
			mode = st.Mode()
		}
		return true, os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), mode)
	}

	return addToShellConfig(configFile, exportLine, comment)
}

func printInstallBinaryHelp(cmd *cobra.Command) {
	fmt.Print(terminal.Banner())

	fmt.Println(terminal.BoldCyan("◆ Usage"))
	fmt.Printf("  %s %s\n\n", terminal.Yellow("osmedeus install binary"), terminal.Gray("[flags]"))

	fmt.Println(terminal.BoldCyan("◆ Modes"))
	fmt.Printf("  %s %s\n", terminal.Yellow("direct-fetch"), terminal.Gray("download from registry metadata (github-release)"))
	fmt.Printf("  %s %s\n", terminal.Yellow("go-getter"), terminal.Gray("clone repos or download archives via go-getter"))
	fmt.Printf("  %s %s\n\n", terminal.Yellow("nix"), terminal.Gray("install with 'nix profile add'"))

	fmt.Println(terminal.BoldCyan("◆ Examples"))
	fmt.Printf("  %s\n", terminal.Green("# List available binaries"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --list-registry-direct-fetch"))
	fmt.Printf("  %s\n\n", terminal.Gray("osmedeus install binary --list-registry-nix-build"))

	fmt.Printf("  %s\n", terminal.Green("# Install specific binaries (auto-detects method from registry)"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --name nuclei"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --name nuclei --name httpx --name ffuf"))
	fmt.Printf("  %s\n\n", terminal.Gray("osmedeus install binary -n dalfox -n gau -n interactsh"))

	fmt.Printf("  %s\n", terminal.Green("# Install all required binaries"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --all"))
	fmt.Printf("  %s\n\n", terminal.Gray("osmedeus install binary --all --install-optional"))

	fmt.Printf("  %s\n", terminal.Green("# Check binary installation status"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --name nuclei --check"))
	fmt.Printf("  %s\n\n", terminal.Gray("osmedeus install binary --all --check"))

	fmt.Printf("  %s\n", terminal.Green("# Download via go-getter (git clone, archives, etc.)"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --go-getter github.com/user/repo.git"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --go-getter github.com/user/repo.git?ref=main&depth=1"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --go-getter https://github.com/user/repo.git//subfolder"))
	fmt.Printf("  %s\n\n", terminal.Gray("osmedeus install binary --go-getter github.com/user/repo.git --go-getter-dest /opt/tools"))

	fmt.Printf("  %s\n", terminal.Green("# Install via Nix (requires Nix)"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --nix-installation"))
	fmt.Printf("  %s\n", terminal.Gray("osmedeus install binary --name nuclei --nix-build-install"))
	fmt.Printf("  %s\n\n", terminal.Gray("osmedeus install binary --nix-pkgs nixpkgs#redis --nix-pkgs nixpkgs#jq"))

	if cmd == nil {
		return
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s\n\n", cmd.UseLine())

	// Only show local flags (not inherited from parent commands)
	fmt.Println("Flags:")
	fmt.Print(cmd.LocalFlags().FlagUsages())

	if cmd.HasAvailableInheritedFlags() {
		fmt.Println("\nGlobal Flags:")
		fmt.Print(cmd.InheritedFlags().FlagUsages())
	}
}

func init() {
	// Persistent flag for custom headers (inherited by all subcommands)
	installCmd.PersistentFlags().StringArrayVar(&customHeaders, "custom-header", []string{}, "custom HTTP header(s) for downloads (format: 'Key: Value', can be repeated)")

	installBaseCmd.Flags().BoolVar(&baseSample, "sample", false, "initialize base folder from embedded sample (replaces existing base folder)")
	installBaseCmd.Flags().BoolVar(&basePreset, "preset", false, "install from OSM_PRESET_URL environment variable (default: DEFAULT_BASE_REPO)")
	installWorkflowCmd.Flags().BoolVar(&workflowPreset, "preset", false, "install from OSM_WORKFLOW_URL environment variable (default: DEFAULT_WORKFLOW_REPO)")
	// Note: --force flag is now global (defined in root.go)

	// Binary command flags
	installBinaryCmd.Flags().StringSliceVarP(&binaryNames, "name", "n", []string{}, "binary name(s) to install (can be repeated)")
	installBinaryCmd.Flags().StringSliceVar(&nixPkgs, "nix-pkgs", []string{}, "Nix package(s) to add to profile (repeatable)")
	installBinaryCmd.Flags().StringVarP(&registryPath, "registry", "r", "", "path or URL to binary registry JSON (default: embedded registry)")
	installBinaryCmd.Flags().BoolVar(&installAll, "all", false, "install all binaries from the registry")
	installBinaryCmd.Flags().BoolVar(&checkOnly, "check", false, "check if binaries are installed without downloading")
	installBinaryCmd.Flags().BoolVar(&nixBuildInstall, "nix-build-install", false, "use Nix to install binaries instead of direct downloads")
	installBinaryCmd.Flags().BoolVar(&nixInstallation, "nix-installation", false, "install Nix package manager (Determinate Systems installer)")
	installBinaryCmd.Flags().BoolVar(&installOptional, "install-optional", false, "include optional binaries in installation")
	// Note: --width flag is now global (defined in root.go), --max-width removed in favor of --width
	installBinaryCmd.Flags().BoolVar(&hideBinaryTags, "disable-tags", false, "hide tags column in list output")
	installBinaryCmd.Flags().StringSliceVar(&goGetterSources, "go-getter", []string{}, "source URL(s) to download via go-getter (supports git repos, archives, etc.)")
	installBinaryCmd.Flags().StringVar(&goGetterDest, "go-getter-dest", "", "destination directory for go-getter downloads (default: $HOME)")
	installBinaryCmd.Flags().BoolVar(&listRegistryNixBuild, "list-registry-nix-build", false, "list binaries available in Nix flake")
	installBinaryCmd.Flags().BoolVar(&listRegistryNixBuild, "list-binary-nix", false, "list binaries available in Nix flake (alias)")
	installBinaryCmd.Flags().BoolVar(&listRegistryDirectFetch, "list-registry-direct-fetch", false, "list binaries available in registry JSON")
	installBinaryCmd.Flags().BoolVar(&listRegistryDirectFetch, "list-binary-registry", false, "list binaries available in registry JSON (alias)")
	installEnvCmd.Flags().BoolVar(&installEnvAll, "all", false, "add binaries path to all supported shell config files")

	installBinaryCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printInstallBinaryHelp(cmd)
	})

	// Add subcommands
	installCmd.AddCommand(installWorkflowCmd)
	installCmd.AddCommand(installBaseCmd)
	installCmd.AddCommand(installBinaryCmd)
	installCmd.AddCommand(installEnvCmd)
	installCmd.AddCommand(installValidateCmd)

	installValidateCmd.Flags().BoolVar(&validateSample, "sample", false, "initialize base folder from embedded sample (replaces existing base folder)")
	installValidateCmd.Flags().BoolVar(&validatePreset, "preset", false, "install ready-to-use base from default repository")
}

// parseCustomHeaders converts the string slice of "Key: Value" pairs to a map
func parseCustomHeaders(headers []string) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		// Split on first ": " to handle values that contain colons
		idx := strings.Index(h, ": ")
		if idx == -1 {
			// Try just ":" without space
			idx = strings.Index(h, ":")
			if idx == -1 {
				continue
			}
			key := strings.TrimSpace(h[:idx])
			value := strings.TrimSpace(h[idx+1:])
			if key != "" {
				result[key] = value
			}
		} else {
			key := strings.TrimSpace(h[:idx])
			value := strings.TrimSpace(h[idx+2:])
			if key != "" {
				result[key] = value
			}
		}
	}
	return result
}

// containsTag checks if a slice of tags contains a specific tag
func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

// padRight pads a string to the specified width with spaces on the right
// This ensures ANSI codes don't affect alignment when applied after padding
func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// colorPadRight pads string to width, then applies color function
func colorPadRight(s string, width int, colorFn func(string) string) string {
	padded := padRight(s, width)
	return colorFn(padded)
}

// truncatePad truncates then pads to exact width
func truncatePad(s string, width int) string {
	if len(s) > width {
		if width > 3 {
			return s[:width-3] + "..."
		}
		return s[:width]
	}
	return padRight(s, width)
}

// defaultParallelWorkers is the number of concurrent binary installation workers
const defaultParallelWorkers = 3

// binariesPerRow is the number of binary names displayed per row
const binariesPerRow = 6

// splitIntoRows splits binary names into rows of specified size
func splitIntoRows(names []string, perRow int) [][]string {
	var rows [][]string
	for i := 0; i < len(names); i += perRow {
		end := i + perRow
		if end > len(names) {
			end = len(names)
		}
		rows = append(rows, names[i:end])
	}
	return rows
}

// renderMultiRowDisplay renders rows of binaries with spinners
// Each row gets its own spinner if any binary in that row is pending
func renderMultiRowDisplay(rows [][]string, status map[string]string, spinnerFrames []string,
	spinnerIdx int, mu *sync.Mutex, showSpinner bool, isFirstRender *bool) {
	mu.Lock()
	defer mu.Unlock()

	// Only move cursor up if not the first render (avoids race condition with initial output)
	lineCount := len(rows)
	if !*isFirstRender {
		fmt.Printf("\033[%dA\r", lineCount)
	}
	*isFirstRender = false

	for _, row := range rows {
		fmt.Print("\033[K") // Clear line

		// Check if any binary in this row is pending
		hasPending := false
		for _, name := range row {
			if status[name] == "pending" {
				hasPending = true
				break
			}
		}

		if showSpinner && hasPending {
			fmt.Printf("  %s %s ", terminal.Cyan(spinnerFrames[spinnerIdx]), terminal.SymbolBullet)
		} else {
			fmt.Printf("  %s ", terminal.SymbolBullet)
		}

		for i, name := range row {
			if i > 0 {
				fmt.Print(" ")
			}
			switch status[name] {
			case "installed":
				fmt.Print(terminal.HiGreen(name))
			case "failed":
				fmt.Print(terminal.Red(name))
			default:
				fmt.Print(terminal.Gray(name))
			}
		}
		fmt.Println()
	}
}

// installBinariesParallel installs binaries using parallel workers with multi-row spinner
func installBinariesParallel(names []string, registry installer.BinaryRegistry,
	binariesFolder string, headers map[string]string, printer *terminal.Printer, silent bool) (failed []string) {

	if len(names) == 0 {
		return nil
	}

	rows := splitIntoRows(names, binariesPerRow)

	// Initialize status tracking
	status := make(map[string]string)
	var statusMu sync.Mutex
	for _, name := range names {
		status[name] = "pending"
	}

	// Spinner configuration
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	// Pre-check which binaries are already installed
	var toInstall []string
	for _, name := range names {
		if installer.IsBinaryInPath(name) {
			status[name] = "installed"
		} else {
			toInstall = append(toInstall, name)
		}
	}

	// If all binaries are already installed, just show final state and return
	if len(toInstall) == 0 {
		if !silent {
			isFirst := true
			renderMultiRowDisplay(rows, status, spinnerFrames, 0, &statusMu, false, &isFirst)
		}
		return nil
	}
	spinnerIdx := 0
	spinnerDone := make(chan struct{})

	// Start spinner display (skip in silent mode)
	if !silent {
		isFirstRender := true

		// Start spinner goroutine (only in non-silent mode)
		go func() {
			// Do immediate first render to avoid race condition
			renderMultiRowDisplay(rows, status, spinnerFrames, spinnerIdx, &statusMu, true, &isFirstRender)

			ticker := time.NewTicker(80 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-spinnerDone:
					return
				case <-ticker.C:
					spinnerIdx = (spinnerIdx + 1) % len(spinnerFrames)
					renderMultiRowDisplay(rows, status, spinnerFrames, spinnerIdx, &statusMu, true, &isFirstRender)
				}
			}
		}()
	}

	// Create work channel and results
	workCh := make(chan string, len(toInstall))
	var wg sync.WaitGroup
	var failedMu sync.Mutex

	// Start workers
	for i := 0; i < defaultParallelWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range workCh {
				if silent {
					fmt.Printf("  %s Installing: %s\n", terminal.SymbolBullet, terminal.Cyan(name))
				}
				err := installer.InstallBinary(name, registry, binariesFolder, headers)
				statusMu.Lock()
				if err != nil {
					status[name] = "failed"
					failedMu.Lock()
					failed = append(failed, name)
					failedMu.Unlock()
				} else {
					status[name] = "installed"
				}
				statusMu.Unlock()
			}
		}()
	}

	// Send work (only binaries that need installation)
	for _, name := range toInstall {
		workCh <- name
	}
	close(workCh)

	// Wait for completion
	wg.Wait()
	close(spinnerDone)

	if !silent {
		time.Sleep(100 * time.Millisecond)
		// Final render without spinners
		isFirstFinal := false
		renderMultiRowDisplay(rows, status, spinnerFrames, 0, &statusMu, false, &isFirstFinal)
	}

	return failed
}

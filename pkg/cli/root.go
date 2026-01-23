package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	settingsFile        string
	baseFolder          string
	workflowFolder      string
	verbose             bool
	debug               bool
	silent              bool
	logFile             string
	logFileTmp          bool
	showUsageExamples   bool
	showSpinner         bool
	disableLogging      bool
	disableColor        bool
	disableNotification bool
	fullUsageExample    bool
	disableDB           bool
	ciOutputFormat      bool
	skipAutoSetup       bool

	// Global flags available to all subcommands
	globalForce bool
	globalJSON  bool
	globalWidth int

	// Build info - set via SetBuildInfo from main.go
	buildTime  = "unknown"
	commitHash = "unknown"
)

// SetBuildInfo sets the build time and commit hash for version display
// This must be called before Execute() to ensure version output is correct
func SetBuildInfo(bt, ch string) {
	buildTime = bt
	commitHash = ch

	// Set version template here (after build info is set) instead of in init()
	// This ensures the template uses actual values instead of "unknown"
	rootCmd.SetVersionTemplate(fmt.Sprintf(`%s - %s
Version: {{.Version}}
Build: %s
Commit: %s
Author: %s
Docs: %s
`, core.BINARY, core.DESC, buildTime, commitHash, core.AUTHOR, core.DOCS))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "osmedeus",
	Short:   "Osmedeus - Workflow Engine for Automated Reconnaissance",
	Version: core.VERSION,
	Long:    UsageRoot(),
	Run: func(cmd *cobra.Command, args []string) {
		if showUsageExamples {
			fmt.Print(terminal.Banner())
			fmt.Println(UsageAllExamples())
			return
		}
		if fullUsageExample {
			fmt.Print(terminal.Banner())
			showInPager(UsageFullExample())
			return
		}
		// Default: show help
		_ = cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Handle CI output format - disable colors and decorations early
		if ciOutputFormat {
			disableColor = true
			silent = true // Auto-enable silent mode in CI
			terminal.SetColorEnabled(false)
			terminal.SetCIMode(true)
		}

		// Debug mode enables verbose automatically
		if debug {
			verbose = true
		}

		// Determine log file path
		actualLogFile := logFile
		if logFileTmp && actualLogFile == "" {
			// Generate temporary log file with timestamp
			timestamp := time.Now().Format("20060102-150405")
			actualLogFile = fmt.Sprintf("osmedeus-log-%s.log", timestamp)
		}

		// Initialize logger
		logCfg := logger.DefaultConfig()
		if silent || disableLogging {
			logCfg.Level = "error" // Only show errors in silent mode
			logCfg.Silent = true
		} else if debug {
			// Only debug mode sets DEBUG level, verbose mode shows step output instead
			logCfg.Level = "debug"
			logCfg.Development = true
			logCfg.Verbose = true // Show caller/source file in debug mode
		}
		// Note: verbose mode now shows actual step output instead of DEBUG logs
		if actualLogFile != "" {
			logCfg.LogFile = actualLogFile
		}
		if err := logger.Init(logCfg); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		if silent {
			_ = os.Setenv("OSMEDEUS_SILENT", "1")
		} else {
			_ = os.Unsetenv("OSMEDEUS_SILENT")
		}

		// Disable colors if requested
		if disableColor {
			terminal.SetColorEnabled(false)
		}

		// Log the log file path if set
		if actualLogFile != "" {
			logger.Get().Debug("Logging to file", zap.String("log_file", actualLogFile))
		}

		// Load configuration
		if baseFolder == "" {
			homeDir, _ := os.UserHomeDir()
			baseFolder = homeDir + "/osmedeus-base"
		}
		logger.Get().Debug("Using base folder", zap.String("base_folder", baseFolder))

		// Auto-generate config files if they don't exist
		if err := config.EnsureConfigExists(baseFolder); err != nil {
			logger.Get().Warn("Failed to create default config", zap.Error(err))
		}

		// Load configuration from custom file or base folder
		var cfg *config.Config
		var err error

		if settingsFile != "" {
			// Load from custom settings file
			logger.Get().Debug("Loading configuration from custom file",
				zap.String("settings_file", settingsFile),
			)
			cfg, err = config.LoadFromFile(settingsFile)
			if err != nil {
				return fmt.Errorf("failed to load config from %s: %w", settingsFile, err)
			}
			// Set base folder if not specified in config
			if cfg.BaseFolder == "" {
				cfg.BaseFolder = baseFolder
			}
		} else {
			logger.Get().Debug("Loading configuration",
				zap.String("base_folder", baseFolder),
				zap.String("config_file", baseFolder+"/osm-settings.yaml"),
			)

			cfg, err = config.Load(baseFolder)
			if err != nil {
				// Use default config if loading fails
				logger.Get().Debug("Using default config", zap.Error(err))
				cfg = config.DefaultConfig()
				cfg.BaseFolder = baseFolder
				cfg.ResolvePaths() // Recalculate all paths with new base folder
			}
		}

		logger.Get().Debug("Configuration loaded",
			zap.String("base_folder", cfg.BaseFolder),
			zap.String("workflows_path", cfg.WorkflowsPath),
			zap.String("workspaces_path", cfg.WorkspacesPath),
			zap.String("db_path", cfg.GetDBPath()),
		)

		// Override workflow folder if specified
		if workflowFolder != "" {
			cfg.WorkflowsPath = workflowFolder
			logger.Get().Info("Using custom workflow folder", zap.String("path", workflowFolder))
		}

		// Export global vars to environment
		cfg.ExportGlobalVarsToEnv()
		logger.Get().Debug("Exported global vars to environment",
			zap.Int("var_count", len(cfg.GlobalVars)),
		)

		// Disable notifications if flag is set
		if disableNotification {
			cfg.Notification.Enabled = false
			logger.Get().Debug("Notifications disabled via CLI flag")
		}

		config.Set(cfg)

		// Check for first-time setup (after config is loaded)
		if !shouldSkipAutoSetup(cmd) && isFirstTimeSetupNeeded(baseFolder) {
			if err := runFirstTimeSetup(baseFolder, cfg); err != nil {
				logger.Get().Warn("First-time setup had issues", zap.Error(err))
			}
			// Reload config after setup
			if reloaded, err := config.Load(baseFolder); err == nil {
				cfg = reloaded
				config.Set(cfg)
			}
		}

		// Show warning if database is disabled (skip warning in CI mode)
		if disableDB && !ciOutputFormat {
			printer := terminal.NewPrinter()
			printer.Warning("Database disabled via --disable-db flag. The following features are unavailable:")
			fmt.Println("    - Database queries (db_select, db_select_*, etc.)")
			fmt.Println("    - Asset/vulnerability tracking")
			fmt.Println("    - Workspace statistics")
			fmt.Println("    - The 'db' subcommand")
			fmt.Println("    Use this mode for lightweight scanning without persistence.")
			fmt.Println()
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Skip printing for TargetTypeMismatchError (already printed with formatting)
		var ttmErr *executor.TargetTypeMismatchError
		if !errors.As(err, &ttmErr) {
			fmt.Fprintf(os.Stderr, "%s %s\n", terminal.Red("Error:"), err)
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&settingsFile, "settings-file", "", "settings file path (default is $HOME/osmedeus-base/osm-settings.yaml)")
	rootCmd.PersistentFlags().StringVarP(&baseFolder, "base-folder", "b", "", "base folder containing workflows and settings (default is $HOME/osmedeus-base/)")
	rootCmd.PersistentFlags().StringVarP(&workflowFolder, "workflow-folder", "F", "", "custom workflow folder (default is $HOME/osmedeus-base/workflows/)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode (verbose + debug logging)")
	rootCmd.PersistentFlags().BoolVarP(&silent, "silent", "q", false, "silent mode - suppress all output except errors")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "path to log file (logs to both console and file)")
	rootCmd.PersistentFlags().BoolVar(&logFileTmp, "log-file-tmp", false, "create temporary log file osmedeus-log-<timestamp>.log")
	rootCmd.PersistentFlags().BoolVarP(&showUsageExamples, "usage-example", "H", false, "show comprehensive usage examples for all commands")
	rootCmd.PersistentFlags().BoolVar(&showSpinner, "spinner", false, "show spinner animations during execution")
	rootCmd.PersistentFlags().BoolVar(&disableLogging, "disable-logging", false, "disable all logging output")
	rootCmd.PersistentFlags().BoolVar(&disableColor, "disable-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&disableNotification, "disable-notification", false, "disable all notifications")
	rootCmd.PersistentFlags().BoolVar(&fullUsageExample, "full-usage-example", false, "show full usage with all flags in pager mode")
	rootCmd.PersistentFlags().BoolVar(&disableDB, "disable-db", false, "disable database connection (warning: some features unavailable)")
	rootCmd.PersistentFlags().BoolVar(&ciOutputFormat, "ci-output-format", false, "output results in JSON format for CI pipelines")
	rootCmd.PersistentFlags().BoolVar(&skipAutoSetup, "skip-auto-setup", false, "skip automatic first-time setup")

	// Global flags available to all subcommands
	rootCmd.PersistentFlags().BoolVar(&globalForce, "force", false, "skip confirmation prompts and force operations")
	rootCmd.PersistentFlags().BoolVar(&globalJSON, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().IntVar(&globalWidth, "width", 80, "max column width for table display (0 = no limit)")

	// Suppress usage display and default error output (we handle errors in Execute())
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	// Note: Version template is set in SetBuildInfo() to use actual build values

	// Set custom help function to show banner before help
	defaultHelpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Print(terminal.Banner())
		defaultHelpFunc(cmd, args)
	})

	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(scanCmd) // Alias for runCmd (backward compatibility)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(functionCmd)
	rootCmd.AddCommand(workerCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(evalCmd)
	rootCmd.AddCommand(clientCmd)
}

// installRequiredBinaries installs all required binaries from the registry.
// Returns counts of installed, skipped, and failed binaries.
// This is a shared helper used by both runFirstTimeSetup and runInstallBase --preset.
func installRequiredBinaries(cfg *config.Config, printer *terminal.Printer) (installed, skipped, failed int) {
	// Check for registry URL override
	registryURL := os.Getenv("OSM_REGISTRY_URL")
	registryDisplay := "the default registry"
	if registryURL != "" {
		registryDisplay = registryURL
	}

	printer.Println("%s %s", terminal.BoldBlue(terminal.SymbolLightning),
		terminal.HiBlue(fmt.Sprintf("Installing security binaries from %s", terminal.Cyan(registryDisplay))))
	printer.Println("  %s This may take a few minutes depending on your network speed", terminal.SymbolBullet)
	printer.Newline()

	// Show spinner while loading registry
	loadingSpinner := terminal.LoadingSpinner("Loading binary registry")
	loadingSpinner.Start()
	registry, err := installer.LoadRegistry(registryURL, nil)
	loadingSpinner.Stop()
	if err != nil {
		printer.Warning("Failed to load binary registry: %s", err)
		printer.Println("  %s", terminal.Gray("You can manually run: osmedeus install binary --all"))
		return 0, 0, 0
	}

	binariesFolder := cfg.BinariesPath
	if binariesFolder == "" {
		binariesFolder = filepath.Join(cfg.BaseFolder, "external-binaries")
	}

	// Create binaries folder if it doesn't exist
	if err := os.MkdirAll(binariesFolder, 0755); err != nil {
		printer.Warning("Failed to create binaries folder: %s", err)
	}

	// Collect all non-optional binaries (InstallBinary handles skip logic and shows names)
	var toInstall []string
	for name, entry := range registry {
		isOptional := false
		for _, tag := range entry.Tags {
			if tag == "optional" {
				isOptional = true
				break
			}
		}
		if !isOptional {
			toInstall = append(toInstall, name)
		}
	}

	if len(toInstall) > 0 {
		printer.Println("  %s Installing %s binaries to: %s", terminal.SymbolBullet, terminal.Green(fmt.Sprintf("%d", len(toInstall))), terminal.Cyan(binariesFolder))
		printer.Newline()
	}

	// Sort binary names for consistent display
	sort.Strings(toInstall)

	// Set silent mode for binary installation to reduce noise
	_ = os.Setenv("OSMEDEUS_SILENT", "1")
	defer func() { _ = os.Unsetenv("OSMEDEUS_SILENT") }()

	// Suppress logger output during binary installation
	logCfg := logger.DefaultConfig()
	logCfg.Silent = true
	_ = logger.Init(logCfg)
	defer func() {
		// Restore normal logging after installation
		logCfg.Silent = false
		_ = logger.Init(logCfg)
	}()

	// Install binaries in parallel with multi-row spinner display
	var failedNames []string
	if len(toInstall) > 0 {
		failedNames = installBinariesParallel(toInstall, registry, binariesFolder, nil, printer, false)
	}

	// Count results
	installedCount := len(toInstall) - len(failedNames)
	failedCount := len(failedNames)

	// Count skipped (already in PATH) - these weren't in toInstall
	skippedCount := 0
	for name, entry := range registry {
		isOptional := false
		for _, tag := range entry.Tags {
			if tag == "optional" {
				isOptional = true
				break
			}
		}
		if !isOptional && installer.IsBinaryInPath(name) {
			skippedCount++
		}
	}

	printer.Newline()

	// Show summary
	printer.Success("Installed %s binaries (%s skipped, %s failed)",
		terminal.Green(fmt.Sprintf("%d", installedCount)),
		terminal.Yellow(fmt.Sprintf("%d", skippedCount)),
		terminal.Red(fmt.Sprintf("%d", failedCount)))

	// Ensure binaries path is in environment
	ensureBinariesPathInEnv(printer, binariesFolder, false)

	return installedCount, skippedCount, failedCount
}

// createInitializationMarker creates the $HOME/.osmedeus/initialized marker file.
// This marker indicates that first-time setup has been completed.
func createInitializationMarker(printer *terminal.Printer) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		printer.Warning("Failed to get home directory: %s", err)
		return err
	}

	osmDir := filepath.Join(homeDir, ".osmedeus")
	if err := os.MkdirAll(osmDir, 0755); err != nil {
		printer.Warning("Failed to create osmedeus config directory: %s", err)
		return err
	}

	markerFile := filepath.Join(osmDir, "initialized")
	if err := os.WriteFile(markerFile, []byte("initialized\n"), 0644); err != nil {
		printer.Warning("Failed to create initialization marker: %s", err)
		return err
	}

	return nil
}

// versionCmd shows version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s - %s\n", core.BINARY, core.DESC)
		fmt.Printf("Version: %s\n", core.VERSION)
		fmt.Printf("Build: %s\n", buildTime)
		fmt.Printf("Commit: %s\n", commitHash)
		fmt.Printf("Author: %s\n", core.AUTHOR)
		fmt.Printf("Docs: %s\n", core.DOCS)
	},
}

// showInPager displays content using a pager (less/more) if available
func showInPager(content string) {
	// Try less first, fall back to more, fall back to direct output
	pagers := []string{"less", "more"}
	for _, pager := range pagers {
		if path, err := exec.LookPath(pager); err == nil {
			cmd := exec.Command(path, "-R") // -R for color support
			cmd.Stdin = strings.NewReader(content)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err == nil {
				return
			}
		}
	}
	// Fallback: direct output
	fmt.Print(content)
}

// shouldSkipAutoSetup returns true if auto-setup should be skipped for this command
func shouldSkipAutoSetup(cmd *cobra.Command) bool {
	if skipAutoSetup {
		return true
	}
	// Skip for commands that handle their own setup or don't need it
	skipCommands := map[string]bool{
		"install":    true,
		"health":     true,
		"version":    true,
		"help":       true,
		"update":     true,
		"completion": true,
		"client":     true,
	}
	// Check command and all parent commands
	for c := cmd; c != nil; c = c.Parent() {
		if skipCommands[c.Name()] {
			return true
		}
	}
	return false
}

// isFirstTimeSetupNeeded checks if base folder needs initialization
// Returns false if $HOME/.osmedeus/initialized marker or database file exists
func isFirstTimeSetupNeeded(baseFolder string) bool {
	if baseFolder == "" {
		return false
	}

	// Check for initialization marker file in $HOME/.osmedeus/
	homeDir, _ := os.UserHomeDir()
	markerFile := filepath.Join(homeDir, ".osmedeus", "initialized")
	if _, err := os.Stat(markerFile); err == nil {
		return false // Marker exists, setup already done
	}

	// Check for database file (also indicates setup was done)
	dbFile := filepath.Join(baseFolder, "database-osm.sqlite")
	if _, err := os.Stat(dbFile); err == nil {
		return false // Database exists, setup already done
	}

	return true // Neither exists, first-time setup needed
}

// runFirstTimeSetup performs automatic first-time setup
func runFirstTimeSetup(baseFolder string, cfg *config.Config) error {
	printer := terminal.NewPrinter()

	// Show welcome banner
	printer.Newline()
	printer.Println("%s %s", terminal.Yellow(terminal.SymbolStar), terminal.BoldCyan("Welcome to Osmedeus!"))
	printer.Println("  %s", terminal.HiMagenta("First-time setup detected."))
	printer.Newline()

	// Step 1: Initialize the environment
	printer.Println("%s %s", terminal.BoldBlue(terminal.SymbolLightning), terminal.HiBlue("Initializing environment..."))
	printer.Println("  %s Base folder: %s", terminal.SymbolBullet, terminal.Cyan(baseFolder))

	// Check for custom preset URL from environment
	presetURL := os.Getenv("OSM_PRESET_URL")
	if presetURL == "" {
		presetURL = core.DEFAULT_BASE_REPO
	}
	printer.Println("  %s Preset URL: %s", terminal.SymbolBullet, terminal.Cyan(presetURL))

	// Check for custom workflow URL from environment (default to DEFAULT_WORKFLOW_REPO)
	workflowURL := os.Getenv("OSM_WORKFLOW_URL")
	if workflowURL == "" {
		workflowURL = core.DEFAULT_WORKFLOW_REPO
	}
	printer.Println("  %s Workflow URL: %s", terminal.SymbolBullet, terminal.Cyan(workflowURL))
	printer.Newline()

	// Step 2: Install preset base folder
	printer.Println("%s %s", terminal.BoldMagenta(terminal.SymbolLightning), terminal.HiMagenta("Installing preset base folder..."))
	printer.Println("  %s Downloading from: %s", terminal.SymbolBullet, terminal.Cyan(presetURL))
	inst := installer.NewInstaller(
		baseFolder,
		filepath.Join(baseFolder, "workflows"),
		filepath.Join(baseFolder, "external-binaries"),
		nil,
	)
	if err := inst.InstallBase(presetURL); err != nil {
		printer.Warning("Failed to install preset base: %s", err)
		printer.Println("  %s", terminal.Gray("You can manually run: osmedeus install validate --preset"))
		return err
	}
	printer.Success("Base folder installed to: %s", terminal.Cyan(baseFolder))
	printer.Newline()

	// Step 3: Install workflows from workflow URL
	printer.Println("%s %s", terminal.BoldBlue(terminal.SymbolLightning), terminal.HiBlue("Installing workflows..."))
	printer.Println("  %s Downloading from: %s", terminal.SymbolBullet, terminal.Cyan(workflowURL))
	if err := inst.InstallWorkflow(workflowURL); err != nil {
		printer.Warning("Failed to install workflows: %s", err)
		printer.Println("  %s", terminal.Gray("You can manually run: osmedeus install workflow --preset"))
	} else {
		printer.Success("Workflows installed to: %s", terminal.Cyan(filepath.Join(baseFolder, "workflows")))
	}
	printer.Newline()

	// Step 4: Reload config and load workflows
	printer.Println("%s %s", terminal.BoldMagenta(terminal.SymbolLightning), terminal.HiMagenta("Loading workflows..."))

	// Ensure config exists after InstallBase (which may have removed the base folder)
	if err := config.EnsureConfigExists(baseFolder); err != nil {
		printer.Warning("Failed to create config: %s", err)
	}

	reloaded, err := config.Load(baseFolder)
	if err == nil {
		config.Set(reloaded)
		cfg = reloaded
		printer.Success("Configuration loaded from: %s", terminal.Cyan(filepath.Join(baseFolder, "osm-settings.yaml")))
	} else {
		printer.Warning("Failed to reload config: %s", err)
	}
	printer.Newline()

	// Step 5: Install all binaries using shared helper
	installRequiredBinaries(cfg, printer)

	// Create initialization marker file
	_ = createInitializationMarker(printer)

	// Print completion message
	printer.Newline()
	printer.Println("%s %s", terminal.Green(terminal.SymbolSuccess), terminal.BoldGreen("First-time setup complete!"))
	printer.Newline()

	// Print next steps hint
	printer.Println("%s %s", terminal.BoldMagenta(terminal.SymbolLightning), terminal.HiMagenta("Next Steps:"))
	printer.Println("  %s Run a scan: %s", terminal.SymbolBullet, terminal.Cyan("osmedeus run -f basic-recon -t example.com"))
	printer.Println("  %s Check health: %s", terminal.SymbolBullet, terminal.Cyan("osmedeus health"))
	printer.Newline()

	return nil
}

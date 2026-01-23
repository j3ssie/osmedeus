package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
	oslogger "github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/metrics"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/pkg/server/handlers"
	"github.com/j3ssie/osmedeus/v5/pkg/server/middleware"
	"github.com/j3ssie/osmedeus/v5/public"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	_ "github.com/j3ssie/osmedeus/v5/docs/api-swagger" // swagger docs
)

// Options contains server configuration options
type Options struct {
	NoAuth              bool                // Disable authentication when true
	Master              *distributed.Master // Master node for distributed mode (nil if not in master mode)
	Debug               bool                // Enable debug mode (log request bodies, detailed errors)
	HotReload           bool                // Enable config hot reload (watches osm-settings.yaml for changes)
	EnableEventReceiver bool                // Enable event receiver for event-triggered workflows (default: true)
}

// cachedServerInfo holds server info read once at startup to avoid reading config on every request
var cachedServerInfo struct {
	License string
	Version string
	Binary  string
	Repo    string
	Author  string
	Docs    string
}

// Server represents the web server
type Server struct {
	app            *fiber.App
	config         *config.Config
	configProvider handlers.ConfigProvider
	hotConfig      *config.HotReloadableConfig // nil if hot reload is disabled
	options        *Options
	eventReceiver  *EventReceiver // nil if event receiver is disabled
}

// New creates a new server instance
func New(cfg *config.Config, opts *Options) (*Server, error) {
	if opts == nil {
		opts = &Options{}
	}

	// Initialize database connection once at startup
	_, err := database.Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations once at startup
	ctx := context.Background()
	if err := database.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}
	oslogger.Get().Debug("Database initialized", zap.String("engine", cfg.Database.DBEngine))

	// Index workflows from filesystem to database at startup
	if cfg.WorkflowsPath != "" {
		result, err := database.IndexWorkflowsFromFilesystem(ctx, cfg.WorkflowsPath, false)
		if err != nil {
			oslogger.Get().Warn("Failed to index workflows at startup", zap.Error(err))
		} else {
			oslogger.Get().Info("Workflows indexed at startup",
				zap.Int("added", result.Added),
				zap.Int("updated", result.Updated),
				zap.Int("removed", result.Removed),
			)
		}
	}

	// Cache server info at startup (read once, not on every request)
	license := cfg.Server.License
	if license == "" {
		license = core.LICENSE // fallback to constant
	}
	cachedServerInfo.License = license
	cachedServerInfo.Version = core.VERSION
	cachedServerInfo.Binary = core.BINARY
	cachedServerInfo.Repo = core.REPO_URL
	cachedServerInfo.Author = core.AUTHOR
	cachedServerInfo.Docs = core.DOCS

	// Set cached info for handlers to use
	handlers.SetServerInfo(&handlers.ServerInfoData{
		License: cachedServerInfo.License,
		Version: cachedServerInfo.Version,
		Binary:  cachedServerInfo.Binary,
		Repo:    cachedServerInfo.Repo,
		Author:  cachedServerInfo.Author,
		Docs:    cachedServerInfo.Docs,
	})

	// Select error handler based on debug mode
	var errHandler fiber.ErrorHandler
	if opts.Debug {
		errHandler = middleware.DebugErrorHandler
	} else {
		errHandler = errorHandler
	}

	app := fiber.New(fiber.Config{
		AppName:               "Osmedeus API Server",
		ServerHeader:          fmt.Sprintf("%s %s (%s)", core.BINARY, core.VERSION, license),
		ErrorHandler:          errHandler,
		DisableStartupMessage: true,
	})

	// Apply middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// Reflect all origins - returns true to allow and echo back the origin
			return true
		},
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,HEAD",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Apply Prometheus metrics middleware (conditional)
	if cfg.Server.IsMetricsEnabled() {
		app.Use(middleware.PrometheusMetrics())
	}

	// Apply debug middleware if debug mode is enabled
	if opts.Debug {
		app.Use(middleware.BodyReusable())
		app.Use(middleware.DebugRequestBody())
		app.Use(middleware.DebugResponseLogger())
	}

	s := &Server{
		app:     app,
		config:  cfg,
		options: opts,
	}

	// Initialize config provider (hot reload or static)
	if opts.HotReload && cfg.BaseFolder != "" {
		hotCfg, err := config.NewHotReloadableConfig(cfg.BaseFolder)
		if err != nil {
			oslogger.Get().Warn("Failed to initialize hot reload config, falling back to static",
				zap.Error(err),
			)
			s.configProvider = handlers.NewStaticConfigProvider(cfg)
		} else {
			s.hotConfig = hotCfg
			s.configProvider = handlers.NewHotReloadConfigProvider(hotCfg)

			// Start watching for config changes
			if err := hotCfg.Watch(); err != nil {
				oslogger.Get().Warn("Failed to start config watcher",
					zap.Error(err),
				)
			} else {
				oslogger.Get().Debug("Config hot reload enabled",
					zap.String("config_path", hotCfg.GetConfigPath()),
				)
			}
		}
	} else {
		s.configProvider = handlers.NewStaticConfigProvider(cfg)
	}

	// Initialize event receiver if enabled (default: true when not explicitly disabled)
	if opts.EnableEventReceiver {
		er, err := NewEventReceiver(cfg)
		if err != nil {
			oslogger.Get().Warn("Failed to initialize event receiver", zap.Error(err))
		} else {
			s.eventReceiver = er
			oslogger.Get().Debug("Event receiver initialized")
		}
	}

	// Setup routes
	s.setupRoutes()

	// Start memory metrics collector if metrics are enabled
	if cfg.Server.IsMetricsEnabled() {
		startMemoryMetricsCollector(context.Background(), 15*time.Second)
	}

	return s, nil
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Favicon from embedded filesystem
	s.app.Use(favicon.New(favicon.Config{
		File:       "favicon.ico",
		FileSystem: http.FS(public.EmbedFS),
	}))

	// Health checks
	s.app.Get("/health", handlers.HealthCheck)
	s.app.Get("/health/ready", handlers.ReadinessCheck)

	// Prometheus metrics endpoint (conditional)
	if s.config.Server.IsMetricsEnabled() {
		s.app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	}

	// Server info (JSON version info)
	s.app.Get("/server-info", handlers.ServerInfo(s.config))
	s.app.Get("/status", handlers.ServerInfo(s.config))   // Alias for /server-info
	s.app.Get("/api/info", handlers.ServerInfo(s.config)) // Alternative endpoint for server info

	// Swagger documentation
	s.app.Get("/swagger/*", swagger.HandlerDefault)

	// API routes
	api := s.app.Group("/osm/api")

	// Login - always accessible
	api.Post("/login", handlers.Login(s.config, s.options.NoAuth))

	// Logout - always accessible (clears session cookie)
	api.Post("/logout", handlers.Logout())

	// Apply auth middleware conditionally
	if !s.options.NoAuth {
		if s.config.Server.EnabledAuthAPI {
			// Support both API key and JWT auth
			api.Use(middleware.CombinedAuth(s.config))
		} else {
			api.Use(middleware.JWTAuth(s.config))
		}
	}

	// Workflows
	api.Get("/workflows", handlers.ListWorkflowsVerbose(s.config))
	api.Get("/workflows/tags", handlers.GetAllWorkflowTags(s.config))
	api.Post("/workflows/refresh", handlers.RefreshWorkflowIndex(s.config))
	api.Get("/workflows/:name", handlers.GetWorkflowVerbose(s.config))

	// Runs
	api.Post("/runs", handlers.CreateRun(s.config))
	api.Get("/runs", handlers.ListRuns(s.config))
	api.Get("/runs/:id", handlers.GetRun(s.config))
	api.Delete("/runs/:id", handlers.CancelRun(s.config))
	api.Get("/runs/:id/steps", handlers.GetRunSteps)
	api.Get("/runs/:id/artifacts", handlers.GetRunArtifacts)

	// Jobs (group of runs from same request)
	api.Get("/jobs/:id", handlers.GetJobStatus(s.config))

	// File uploads
	api.Post("/upload-file", handlers.UploadFile(s.config))
	api.Post("/workflow-upload", handlers.UploadWorkflow(s.config))

	// Snapshots (legacy endpoint for backward compatibility)
	api.Get("/snapshot-download/:workspace_name", handlers.SnapshotDownload(s.config))

	// Snapshots (new endpoints)
	api.Get("/snapshots", handlers.ListSnapshots(s.config))
	api.Post("/snapshots/export", handlers.SnapshotExport(s.config))
	api.Post("/snapshots/import", handlers.SnapshotImport(s.config))
	api.Delete("/snapshots/:name", handlers.DeleteSnapshot(s.config))

	// Workspaces
	api.Get("/workspaces", handlers.ListWorkspaces(s.config))
	api.Get("/workspace-names", handlers.ListWorkspaceNames(s.config))

	// Artifacts
	api.Get("/artifacts", handlers.ListArtifacts(s.config))
	api.Get("/artifacts/:workspace_name", handlers.DownloadWorkspaceArtifact(s.config))

	// Step Results (global listing)
	api.Get("/step-results", handlers.ListStepResults(s.config))

	// Assets
	api.Get("/assets", handlers.ListAssets(s.config))
	api.Get("/assets/diff", handlers.GetAssetDiff(s.config))
	api.Get("/assets/diffs", handlers.ListAssetDiffSnapshots(s.config))

	// Vulnerabilities
	api.Get("/vulnerabilities", handlers.ListVulnerabilities(s.config))
	api.Get("/vulnerabilities/diff", handlers.GetVulnerabilityDiff(s.config))
	api.Get("/vulnerabilities/diffs", handlers.ListVulnDiffSnapshots(s.config))
	api.Get("/vulnerabilities/summary", handlers.GetVulnerabilitySummary(s.config))
	api.Get("/vulnerabilities/:id", handlers.GetVulnerability(s.config))
	api.Post("/vulnerabilities", handlers.CreateVulnerability(s.config))
	api.Delete("/vulnerabilities/:id", handlers.DeleteVulnerability(s.config))

	// Stats
	api.Get("/stats", handlers.GetSystemStats(s.config))

	// Install - registry info and installation endpoints
	api.Get("/registry-info", handlers.GetRegistryInfo(s.config))
	api.Post("/registry-install", handlers.RegistryInstall(s.config))

	// Schedules
	api.Get("/schedules", handlers.ListSchedules(s.config))
	api.Post("/schedules", handlers.CreateSchedule(s.config, s.eventReceiver))
	api.Get("/schedules/:id", handlers.GetSchedule(s.config))
	api.Put("/schedules/:id", handlers.UpdateSchedule(s.config))
	api.Delete("/schedules/:id", handlers.DeleteSchedule(s.config))
	api.Post("/schedules/:id/enable", handlers.EnableSchedule(s.config))
	api.Post("/schedules/:id/disable", handlers.DisableSchedule(s.config))
	api.Post("/schedules/:id/trigger", handlers.TriggerSchedule(s.config))

	// Event logs
	api.Get("/event-logs", handlers.ListEventLogs(s.config))

	// Functions
	api.Post("/functions/eval", handlers.FunctionEval(s.config))
	api.Get("/functions/list", handlers.FunctionList(s.config))

	// Settings API
	api.Get("/settings/yaml", handlers.GetSettingsYAML(s.config))
	api.Get("/settings/yaml/", handlers.GetSettingsYAML(s.config))
	api.Post("/settings/reload", handlers.ReloadConfig(s.hotConfig))
	api.Get("/settings/status", handlers.GetConfigStatus(s.hotConfig))

	// LLM endpoints (OpenAI-compatible)
	api.Post("/llm/v1/chat/completions", handlers.LLMChat(s.config))
	api.Post("/llm/v1/embeddings", handlers.LLMEmbedding(s.config))

	// Distributed endpoints (only available when running in master mode)
	if s.options.Master != nil {
		api.Get("/workers", handlers.ListWorkers(s.options.Master))
		api.Get("/workers/:id", handlers.GetWorker(s.options.Master))
		api.Get("/tasks", handlers.ListTasks(s.options.Master))
		api.Get("/tasks/:id", handlers.GetTask(s.options.Master))
		api.Post("/tasks", handlers.SubmitTask(s.options.Master))
	}

	// Event receiver endpoints (only available when event receiver is enabled)
	if s.eventReceiver != nil {
		api.Get("/event-receiver/status", handlers.GetEventReceiverStatus(s.eventReceiver))
		api.Get("/event-receiver/workflows", handlers.ListEventReceiverWorkflows(s.eventReceiver))
		api.Post("/events/emit", handlers.EmitEvent(s.eventReceiver))
	}

	// Serve workspace files under /ws/{workspace_prefix_key}/
	// Allows direct access to run outputs in workspaces directory (no auth required)
	if s.config.Server.WorkspacePrefixKey != "" && s.config.WorkspacesPath != "" {
		wsPath := fmt.Sprintf("/ws/%s", s.config.Server.WorkspacePrefixKey)
		s.app.Static(wsPath, s.config.WorkspacesPath, fiber.Static{
			Browse: true, // Enable directory listing
		})
	}

	// Handle HEAD requests for UI routes (filesystem middleware only handles GET)
	// This middleware intercepts HEAD requests to non-API paths and returns 200 OK
	s.app.Use(func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodHead {
			// For HEAD requests to non-API paths, return 200 OK
			// This handles UI routes that the filesystem middleware would skip
			if !strings.HasPrefix(c.Path(), "/osm/api") &&
				!strings.HasPrefix(c.Path(), "/health") &&
				!strings.HasPrefix(c.Path(), "/metrics") &&
				!strings.HasPrefix(c.Path(), "/swagger") {
				c.Set("Content-Type", "text/html; charset=utf-8")
				return c.SendStatus(fiber.StatusOK)
			}
		}
		return c.Next()
	})

	// Serve UI at root /
	// Priority: external UI path > embedded UI
	if s.config.UIPath != "" {
		if _, err := os.Stat(s.config.UIPath); err == nil {
			s.app.Use("/", filesystem.New(filesystem.Config{
				Root:         http.Dir(s.config.UIPath),
				Index:        "index.html",
				Browse:       false,
				NotFoundFile: "index.html", // SPA fallback: serve index.html for client-side routes
			}))
		} else {
			// Fallback to embedded UI if external path doesn't exist
			s.serveEmbeddedUI()
		}
	} else {
		// Use embedded UI when no external path configured
		s.serveEmbeddedUI()
	}
}

// serveEmbeddedUI serves the embedded UI files at root /
func (s *Server) serveEmbeddedUI() {
	uiFS, err := public.GetUIFS()
	if err != nil {
		return
	}
	s.app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(uiFS),
		Index:        "index.html",
		Browse:       false,
		PathPrefix:   "",
		NotFoundFile: "index.html", // SPA fallback: serve index.html for client-side routes
	}))
}

// StartEventReceiver starts the event receiver and scheduler.
// This should be called before PrintStartupInfo to ensure triggers are registered.
func (s *Server) StartEventReceiver() {
	if s.eventReceiver != nil {
		if err := s.eventReceiver.Start(context.Background()); err != nil {
			oslogger.Get().Warn("Failed to start event receiver", zap.Error(err))
		}
	}
}

// StartListener starts only the HTTP listener (event receiver already started).
func (s *Server) StartListener(addr string) error {
	return s.app.Listen(addr)
}

// Start starts the server (event receiver + HTTP listener).
// For finer control, use StartEventReceiver() + StartListener() instead.
func (s *Server) Start(addr string) error {
	s.StartEventReceiver()
	return s.StartListener(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	// Stop event receiver if running
	if s.eventReceiver != nil {
		_ = s.eventReceiver.Stop()
	}
	// Stop hot config watcher if enabled
	if s.hotConfig != nil {
		_ = s.hotConfig.Stop()
	}
	return s.app.Shutdown()
}

// ShutdownWithContext gracefully shuts down the server with a context for timeout
func (s *Server) ShutdownWithContext(ctx context.Context) error {
	// Stop event receiver if running
	if s.eventReceiver != nil {
		_ = s.eventReceiver.Stop()
	}
	// Stop hot config watcher if enabled
	if s.hotConfig != nil {
		_ = s.hotConfig.Stop()
	}
	return s.app.ShutdownWithContext(ctx)
}

// GetConfigProvider returns the config provider used by this server.
// This can be used by handlers that need access to fresh configuration.
func (s *Server) GetConfigProvider() handlers.ConfigProvider {
	return s.configProvider
}

// GetHotConfig returns the hot reloadable config, or nil if hot reload is disabled.
func (s *Server) GetHotConfig() *config.HotReloadableConfig {
	return s.hotConfig
}

// IsHotReloadEnabled returns true if config hot reload is enabled.
func (s *Server) IsHotReloadEnabled() bool {
	return s.hotConfig != nil
}

// errorHandler handles errors globally
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Custom handling for 403 Forbidden
	if code == fiber.StatusForbidden {
		path := c.Path()
		// Redirect to root for login and schedules routes
		if path == "/login" || strings.HasPrefix(path, "/schedules") || strings.HasPrefix(path, "/inventory") {
			return c.Redirect("/", fiber.StatusFound)
		}
		return c.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": "Oh dear! It seems you've wandered off the path. If you'd like to see the UI page, please pop back root route at /",
		})
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}

// startMemoryMetricsCollector starts a background goroutine that periodically
// collects memory statistics and updates the Prometheus metrics.
func startMemoryMetricsCollector(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var m runtime.MemStats
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runtime.ReadMemStats(&m)
				metrics.UpdateMemoryMetrics(&m)
			}
		}
	}()
}

// PrintStartupInfo prints colorized server startup information to stdout.
// This replaces Fiber's default banner with custom messaging.
func (s *Server) PrintStartupInfo(addr string) {
	p := terminal.NewPrinter()

	// Banner line
	fmt.Printf("%s Initiating Osmedeus %s - Crafted with %s by %s\n",
		terminal.Yellow(terminal.SymbolLightning),
		terminal.Cyan(core.VERSION),
		terminal.Red("<3"),
		terminal.Yellow(core.AUTHOR))

	// Starting server
	p.Info("Starting Osmedeus server %s", terminal.Cyan("http://"+addr))

	// Database info
	p.Info("Database initialized %s", terminal.Cyan(s.config.Database.DBEngine))

	// Hot reload status
	if s.hotConfig != nil {
		p.Info("Started config hot reload watcher %s", terminal.Yellow(s.hotConfig.GetConfigPath()))
	}

	// Workflows path (always show, in yellow)
	p.Info("Workflows loaded from %s", terminal.Yellow(s.config.GetWorkflowsDir()))

	// Event receiver info with detailed triggers
	if s.eventReceiver != nil {
		triggers := s.eventReceiver.GetRegisteredTriggersInfo()

		if len(triggers) > 0 {
			p.Info("Event receiver initialized")

			// Show each registered trigger with tree formatting
			for i, t := range triggers {
				prefix := "├─"
				if i == len(triggers)-1 {
					prefix = "└─"
				}

				// Format trigger details based on type
				var detail string
				if t.Topic != "" {
					detail = fmt.Sprintf("%s: %s", terminal.Blue("event"), terminal.Blue(t.Topic))
				} else {
					detail = terminal.Blue(t.Type)
				}

				fmt.Printf("  %s Registered trigger: %s (%s)\n",
					prefix,
					terminal.Yellow(t.WorkflowName),
					detail)
			}

			// Scheduler started message
			p.Info("Scheduler started with %s triggers", terminal.Cyan(fmt.Sprintf("%d", len(triggers))))
		} else {
			p.Info("Event receiver initialized %s", terminal.Gray("(no event triggers)"))
			p.Info("Scheduler started")
		}
	} else {
		// No event receiver, but still show scheduler status
		p.Info("Scheduler started")
	}

	// Print credentials tip
	fmt.Println()
	fmt.Printf("💡 %s\n", terminal.Gray("Tip: View your login credentials with:"))
	fmt.Printf("   %s\n", terminal.Cyan("osmedeus config view server.username"))
	fmt.Printf("   %s\n", terminal.Cyan("osmedeus config view server.password"))
}

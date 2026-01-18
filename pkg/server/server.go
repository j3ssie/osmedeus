package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
	oslogger "github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/pkg/server/handlers"
	"github.com/j3ssie/osmedeus/v5/pkg/server/middleware"
	"github.com/j3ssie/osmedeus/v5/public"
	"go.uber.org/zap"

	_ "github.com/j3ssie/osmedeus/v5/docs/api-swagger" // swagger docs
)

// Options contains server configuration options
type Options struct {
	NoAuth bool                // Disable authentication when true
	Master *distributed.Master // Master node for distributed mode (nil if not in master mode)
	Debug  bool                // Enable debug mode (log request bodies, detailed errors)
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
	app     *fiber.App
	config  *config.Config
	options *Options
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
	oslogger.Get().Info("Database initialized", zap.String("engine", cfg.Database.DBEngine))

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
		AppName:      "Osmedeus API Server",
		ServerHeader: fmt.Sprintf("%s %s (%s)", core.BINARY, core.VERSION, license),
		ErrorHandler: errHandler,
	})

	// Apply middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,HEAD",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Apply Prometheus metrics middleware
	app.Use(middleware.PrometheusMetrics())

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

	// Setup routes
	s.setupRoutes()

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

	// Prometheus metrics endpoint
	s.app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

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

	// Apply auth middleware conditionally
	if s.config.Server.EnabledAuthAPI {
		api.Use(middleware.APIKeyAuth(s.config))
	} else if !s.options.NoAuth {
		api.Use(middleware.JWTAuth(s.config))
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

	// Assets
	api.Get("/assets", handlers.ListAssets(s.config))

	// Vulnerabilities
	api.Get("/vulnerabilities", handlers.ListVulnerabilities(s.config))
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
	api.Post("/schedules", handlers.CreateSchedule(s.config))
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

// Start starts the server
func (s *Server) Start(addr string) error {
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

// ShutdownWithContext gracefully shuts down the server with a context for timeout
func (s *Server) ShutdownWithContext(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// errorHandler handles errors globally
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}

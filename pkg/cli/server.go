package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/pkg/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	serverHost           string
	serverPort           int
	noAuth               bool
	masterMode           bool
	redisURLServe        string
	disableHotReload     bool
	disableEventReceiver bool
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"server"},
	Short:   "Start the Osmedeus web server",
	Long:    UsageServe(),
	RunE:    runServer,
}

func init() {
	serveCmd.Flags().StringVar(&serverHost, "host", "", "host to bind (default from config)")
	serveCmd.Flags().IntVar(&serverPort, "port", 0, "port to listen on (default from config)")
	serveCmd.Flags().BoolVarP(&noAuth, "no-auth", "A", false, "disable all authentication")
	serveCmd.Flags().BoolVar(&masterMode, "master", false, "run as distributed master node (requires Redis)")
	serveCmd.Flags().StringVar(&redisURLServe, "redis-url", "", "Redis connection URL for master mode (overrides settings)")
	serveCmd.Flags().BoolVar(&disableHotReload, "no-hot-reload", false, "disable config hot reload (default: hot reload is enabled)")
	serveCmd.Flags().BoolVar(&disableEventReceiver, "no-event-receiver", false, "disable automatic event receiver (event-triggered workflows)")
}

func runServer(cmd *cobra.Command, args []string) error {
	log := logger.Get()
	cfg := config.Get()

	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Override with flags if provided
	if serverHost != "" {
		cfg.Server.Host = serverHost
	}
	if serverPort != 0 {
		cfg.Server.Port = serverPort
	}

	// Show critical warning if running without authentication
	if noAuth {
		fmt.Println("\033[31m\033[1m")
		fmt.Println("╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║  ⚠️  CRITICAL WARNING: Server running WITHOUT authentication   ║")
		fmt.Println("║  Anyone can access all API endpoints without credentials!      ║")
		fmt.Println("║  Only use -A flag for development/testing purposes.            ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		fmt.Println("\033[0m")
	}

	// Handle shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutting down...")
		cancel()
	}()

	// Start master node if --master flag is set
	var master *distributed.Master
	if masterMode {
		// Override Redis config if --redis-url provided
		if redisURLServe != "" {
			redisCfg, err := distributed.ParseRedisURL(redisURLServe)
			if err != nil {
				return fmt.Errorf("invalid redis URL: %w", err)
			}
			cfg.Redis = *redisCfg
		}

		if !cfg.IsRedisConfigured() {
			return fmt.Errorf("redis not configured. Add redis section to osm-settings.yaml or use --redis-url")
		}

		var err error
		master, err = distributed.NewMaster(cfg)
		if err != nil {
			return fmt.Errorf("failed to create master: %w", err)
		}

		// Start master in background
		go func() {
			if err := master.Start(ctx); err != nil {
				log.Error("Master error", zap.Error(err))
			}
		}()

		log.Info("Started distributed master node")
	}

	// Create server with master reference for distributed endpoints
	opts := &server.Options{
		NoAuth:              noAuth,
		Master:              master,
		Debug:               debug,
		HotReload:           !disableHotReload,
		EnableEventReceiver: !disableEventReceiver,
	}
	srv, err := server.New(cfg, opts)
	if err != nil {
		log.Error("Failed to create server", zap.Error(err))
		return err
	}

	// Log debug mode status
	if debug {
		log.Info("Debug mode enabled - request bodies and detailed errors will be logged")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// Start event receiver first (registers triggers synchronously)
	srv.StartEventReceiver()

	// Print startup info (triggers are now registered)
	srv.PrintStartupInfo(addr)

	// Start HTTP listener in goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.StartListener(addr)
	}()

	// Wait for shutdown or server error
	select {
	case <-ctx.Done():
		log.Info("Shutting down server...")

		// Give server 5 seconds to shutdown gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.ShutdownWithContext(shutdownCtx); err != nil {
			log.Error("Error during shutdown", zap.Error(err))
		}
		log.Info("Server stopped")
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	return nil
}

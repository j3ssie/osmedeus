package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml/parser"
	"github.com/j3ssie/osmedeus/v5/internal/cloud"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/public"
	"github.com/spf13/cobra"
)

var (
	// Cloud command flags
	cloudProvider  string
	cloudMode      string
	cloudInstances int
	cloudForce     bool
)

// cloudCmd represents the cloud command
var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Cloud infrastructure management commands",
	Long:  `Provision and manage cloud infrastructure for distributed scanning`,
}

// cloudConfigCmd manages cloud configuration
var cloudConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage cloud configuration",
	Long:  `View and update cloud configuration settings`,
}

// cloudConfigSetCmd sets a cloud config value
var cloudConfigSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a cloud configuration value",
	Long:  `Set a cloud configuration value using dot notation (e.g., defaults.provider digitalocean)`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		configPath := cfg.Cloud.CloudSettings
		if configPath == "" {
			configPath = filepath.Join(cfg.BaseFolder, "cloud", "cloud-settings.yaml")
		}

		// Expand template variables
		configPath = strings.ReplaceAll(configPath, "{{base_folder}}", cfg.BaseFolder)

		// Auto-create from preset if file doesn't exist
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := ensureCloudConfig(configPath); err != nil {
				return err
			}
		}

		// Load existing config
		cloudCfg, err := cloud.LoadCloudConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load cloud config: %w", err)
		}

		// Set the value using dot notation
		key := args[0]
		value := args[1]

		if err := setCloudConfigValue(cloudCfg, key, value); err != nil {
			return err
		}

		// Save config
		if err := cloud.SaveCloudConfig(cloudCfg, configPath); err != nil {
			return fmt.Errorf("failed to save cloud config: %w", err)
		}

		printer.Success("Cloud config updated: %s = %s", key, value)
		return nil
	},
}

var cloudConfigListShowSecrets bool

// cloudConfigListCmd lists cloud configuration as flattened key=value pairs
var cloudConfigListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List cloud configuration values",
	Long:    `Display cloud configuration as flattened key=value pairs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		configPath := cfg.Cloud.CloudSettings
		if configPath == "" {
			configPath = filepath.Join(cfg.BaseFolder, "cloud", "cloud-settings.yaml")
		}

		// Expand template variables
		configPath = strings.ReplaceAll(configPath, "{{base_folder}}", cfg.BaseFolder)

		// Auto-create from preset if file doesn't exist
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := ensureCloudConfig(configPath); err != nil {
				return err
			}
		}

		// Read and parse YAML via AST to flatten
		content, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read cloud config: %w", err)
		}

		file, err := parser.ParseBytes(content, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse cloud config: %w", err)
		}
		if len(file.Docs) == 0 {
			return fmt.Errorf("empty cloud config file")
		}

		out := map[string]string{}
		flattenASTScalars(file.Docs[0].Body, "", out)

		keys := make([]string, 0, len(out))
		for k := range out {
			keys = append(keys, k)
		}
		sortStrings(keys)

		for _, k := range keys {
			v := out[k]
			if !cloudConfigListShowSecrets {
				v = redactValueForDisplay(k, v, false)
			}
			fmt.Printf("%s = %s\n", getCategoryColor(k)(k), v)
		}
		return nil
	},
}

// ensureCloudConfig creates the cloud config file from the embedded preset
func ensureCloudConfig(configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create cloud config directory: %w", err)
	}

	data, err := public.GetCloudConfigExample()
	if err != nil {
		return fmt.Errorf("failed to read embedded cloud config preset: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cloud config: %w", err)
	}

	printer.Success("Created cloud config from preset at %s", configPath)
	return nil
}

// cloudCreateCmd provisions cloud infrastructure
var cloudCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create cloud infrastructure",
	Long:  `Provision cloud infrastructure (VMs or serverless functions)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		if !cfg.Cloud.Enabled {
			return fmt.Errorf("cloud features are disabled. Enable in osm-settings.yaml: cloud.enabled = true")
		}

		// Load cloud config
		configPath := strings.ReplaceAll(cfg.Cloud.CloudSettings, "{{base_folder}}", cfg.BaseFolder)
		cloudCfg, err := cloud.LoadCloudConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load cloud config: %w", err)
		}

		// Override provider if specified
		providerType := cloud.ProviderType(cloudCfg.Defaults.Provider)
		if cloudProvider != "" {
			providerType = cloud.ProviderType(cloudProvider)
		}

		// Override mode if specified
		mode := cloud.ExecutionMode(cloudCfg.Defaults.Mode)
		if cloudMode != "" {
			mode = cloud.ExecutionMode(cloudMode)
		}

		// Override instance count if specified
		instanceCount := cloudCfg.Defaults.MaxInstances
		if cloudInstances > 0 {
			instanceCount = cloudInstances
		}

		// Validate against limits
		if instanceCount > cloudCfg.Limits.MaxInstances {
			return fmt.Errorf("instance count (%d) exceeds limit (%d)", instanceCount, cloudCfg.Limits.MaxInstances)
		}

		printer.Section("Creating Cloud Infrastructure")
		printer.Info("Provider: %s", providerType)
		printer.Info("Mode: %s", mode)
		printer.Info("Instances: %d", instanceCount)

		// Create provider
		provider, err := cloud.CreateProvider(cloudCfg, providerType)
		if err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}

		// Validate provider credentials
		ctx := context.Background()
		if err := provider.Validate(ctx); err != nil {
			return fmt.Errorf("provider validation failed: %w", err)
		}

		// Estimate cost
		estimate, err := provider.EstimateCost(mode, instanceCount)
		if err != nil {
			return fmt.Errorf("failed to estimate cost: %w", err)
		}

		printer.Info("Estimated cost: $%.2f/hour ($%.2f/day)", estimate.HourlyCost, estimate.DailyCost)
		for note := range estimate.Notes {
			printer.Info("  - %s", estimate.Notes[note])
		}

		// TODO: Add confirmation prompt

		// Create infrastructure (placeholder - will be implemented)
		printer.Warning("Infrastructure creation not yet fully implemented")
		return fmt.Errorf("not yet implemented")
	},
}

// cloudListCmd lists cloud infrastructure
var cloudListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloud infrastructure",
	Long:  `List all active cloud infrastructure`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		statePath := strings.ReplaceAll(cfg.Cloud.CloudPath, "{{base_folder}}", cfg.BaseFolder)
		infrastructures, err := cloud.ListInfrastructures(statePath)
		if err != nil {
			return fmt.Errorf("failed to list infrastructures: %w", err)
		}

		if len(infrastructures) == 0 {
			printer.Info("No active cloud infrastructure found")
			return nil
		}

		printer.Section("Cloud Infrastructure")
		for _, infra := range infrastructures {
			printer.Info("ID: %s", infra.ID)
			printer.Info("  Provider: %s", infra.Provider)
			printer.Info("  Mode: %s", infra.Mode)
			printer.Info("  Created: %s", infra.CreatedAt.Format("2006-01-02 15:04:05"))
			printer.Info("  Resources: %d", len(infra.Resources))
			for _, res := range infra.Resources {
				printer.Info("    - %s (%s) - %s", res.Name, res.Type, res.Status)
				if res.PublicIP != "" {
					printer.Info("      IP: %s", res.PublicIP)
				}
				if res.WorkerID != "" {
					printer.Info("      Worker: %s", res.WorkerID)
				}
			}
			fmt.Println()
		}

		return nil
	},
}

// cloudDestroyCmd destroys cloud infrastructure
var cloudDestroyCmd = &cobra.Command{
	Use:   "destroy [infrastructure-id]",
	Short: "Destroy cloud infrastructure",
	Long:  `Tear down cloud infrastructure and clean up resources`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		// TODO: Implement destroy logic
		printer.Warning("Infrastructure destruction not yet fully implemented")
		return fmt.Errorf("not yet implemented")
	},
}

// cloudRunCmd runs a workflow on cloud infrastructure
var cloudRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run workflow on cloud infrastructure",
	Long:  `Provision cloud infrastructure, run workflow, and collect results`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		// TODO: Implement cloud run logic
		printer.Warning("Cloud run not yet fully implemented")
		return fmt.Errorf("not yet implemented")
	},
}

// setCloudConfigValue sets a nested config value using dot notation
func setCloudConfigValue(cfg *config.CloudConfigs, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key format. Use dot notation (e.g., defaults.provider)")
	}

	// Simple implementation for common keys
	switch parts[0] {
	case "defaults":
		switch parts[1] {
		case "provider":
			cfg.Defaults.Provider = value
		case "mode":
			cfg.Defaults.Mode = value
		case "max_instances":
			var val int
			if _, err := fmt.Sscanf(value, "%d", &val); err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			cfg.Defaults.MaxInstances = val
		case "cleanup_on_failure":
			cfg.Defaults.CleanupOnFailure = (value == "true")
		default:
			return fmt.Errorf("unknown key: %s", key)
		}

	case "providers":
		if len(parts) < 3 {
			return fmt.Errorf("provider key requires 3 parts (e.g., providers.digitalocean.token)")
		}
		switch parts[1] {
		case "digitalocean":
			switch parts[2] {
			case "token":
				cfg.Providers.DigitalOcean.Token = value
			case "region":
				cfg.Providers.DigitalOcean.Region = value
			case "size":
				cfg.Providers.DigitalOcean.Size = value
			default:
				return fmt.Errorf("unknown DigitalOcean key: %s", parts[2])
			}
		case "aws":
			switch parts[2] {
			case "access_key_id":
				cfg.Providers.AWS.AccessKeyID = value
			case "secret_access_key":
				cfg.Providers.AWS.SecretAccessKey = value
			case "region":
				cfg.Providers.AWS.Region = value
			default:
				return fmt.Errorf("unknown AWS key: %s", parts[2])
			}
		default:
			return fmt.Errorf("unknown provider: %s", parts[1])
		}

	case "limits":
		switch parts[1] {
		case "max_hourly_spend":
			var val float64
			if _, err := fmt.Sscanf(value, "%f", &val); err != nil {
				return fmt.Errorf("invalid float value: %s", value)
			}
			cfg.Limits.MaxHourlySpend = val
		case "max_total_spend":
			var val float64
			if _, err := fmt.Sscanf(value, "%f", &val); err != nil {
				return fmt.Errorf("invalid float value: %s", value)
			}
			cfg.Limits.MaxTotalSpend = val
		case "max_instances":
			var val int
			if _, err := fmt.Sscanf(value, "%d", &val); err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			cfg.Limits.MaxInstances = val
		default:
			return fmt.Errorf("unknown limit key: %s", parts[1])
		}

	default:
		return fmt.Errorf("unknown config section: %s", parts[0])
	}

	return nil
}

func init() {
	// Add subcommands
	cloudCmd.AddCommand(cloudConfigCmd)
	cloudCmd.AddCommand(cloudCreateCmd)
	cloudCmd.AddCommand(cloudListCmd)
	cloudCmd.AddCommand(cloudDestroyCmd)
	cloudCmd.AddCommand(cloudRunCmd)

	cloudConfigCmd.AddCommand(cloudConfigSetCmd)
	cloudConfigCmd.AddCommand(cloudConfigListCmd)
	cloudConfigListCmd.Flags().BoolVar(&cloudConfigListShowSecrets, "show-secrets", false, "show sensitive values")

	// Flags for create command
	cloudCreateCmd.Flags().StringVarP(&cloudProvider, "provider", "p", "", "Cloud provider (aws, gcp, digitalocean, linode, azure)")
	cloudCreateCmd.Flags().StringVarP(&cloudMode, "mode", "m", "", "Execution mode (vm, serverless)")
	cloudCreateCmd.Flags().IntVarP(&cloudInstances, "instances", "n", 0, "Number of instances to create")
	cloudCreateCmd.Flags().BoolVarP(&cloudForce, "force", "f", false, "Force recreation of existing infrastructure")

	// Flags for run command (inherit from run command)
	cloudRunCmd.Flags().StringVarP(&cloudProvider, "provider", "p", "", "Cloud provider")
	cloudRunCmd.Flags().StringVarP(&cloudMode, "mode", "m", "", "Execution mode")
	cloudRunCmd.Flags().IntVarP(&cloudInstances, "instances", "n", 0, "Number of instances")
}

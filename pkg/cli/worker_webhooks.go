package cli

import (
	"context"
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/database/repository"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

// workerWebhooksCmd lists registered webhook triggers
var workerWebhooksCmd = &cobra.Command{
	Use:     "webhooks",
	Aliases: []string{"webhook"},
	Short:   "List webhook trigger URLs",
	RunE:    runWorkerWebhooks,
}

func init() {
	workerCmd.AddCommand(workerWebhooksCmd)
}

func runWorkerWebhooks(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return errConfigNotLoaded
	}

	ctx := context.Background()

	// Connect to database
	dbConn, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	runRepo := repository.NewRunRepository(dbConn)
	runs, err := runRepo.ListWebhookRuns(ctx)
	if err != nil {
		return fmt.Errorf("failed to query webhook runs: %w", err)
	}

	// JSON output
	if globalJSON {
		if len(runs) == 0 {
			fmt.Println("[]")
			return nil
		}
		jsonBytes, err := json.MarshalIndent(runs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal runs: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	p := terminal.NewPrinter()

	if len(runs) == 0 {
		p.Info("No webhook triggers registered")
		p.Info("Use '%s' to register a webhook", terminal.Cyan("osmedeus run --as-webhook -m <module> -t <target>"))
		return nil
	}

	p.Section("Webhook Triggers")

	headers := []string{"ID", "Workflow", "Target", "Status", "Webhook URL", "Auth Key", "Created"}
	var rows [][]string
	for _, r := range runs {
		webhookURL := fmt.Sprintf("/osm/api/webhook-runs/%s/trigger", r.WebhookUUID)
		authKey := ""
		if r.WebhookAuthKey != "" {
			authKey = r.WebhookAuthKey
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", r.ID),
			r.WorkflowName,
			truncateQueueStr(r.Target, 30),
			colorizeQueueStatus(r.Status),
			webhookURL,
			authKey,
			r.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	printMarkdownTable(headers, rows)
	fmt.Println()
	p.Info("Total: %d webhook(s)", len(runs))

	if !cfg.Server.EnableTriggerViaWebhook {
		p.Warning("Webhook triggering is currently disabled. Set 'enable_trigger_via_webhook: true' in osm-settings.yaml")
	}

	return nil
}

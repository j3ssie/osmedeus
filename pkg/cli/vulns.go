package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	vulnsWorkspace      string
	vulnsSeverity       string
	vulnsMinSeverity    string
	vulnsConfidence     string
	vulnsTemplate       string
	vulnsType           string
	vulnsStats          bool
	vulnsLimit          int
	vulnsOffset         int
	vulnsColumns        string
	vulnsExcludeColumns string
	vulnsAll            bool
	vulnsWhere          []string
	vulnsSearch         string
	vulnsValue          string
	vulnsDetail         int64
)

// severityOrder is the single source of severity ordering, most severe first. It is
// the --severity whitelist, the --min-severity expansion and the --stats display
// order, so a band cannot be added to one and forgotten in the others.
var severityOrder = []string{"critical", "high", "medium", "low", "info"}

// severityRank returns a sortable rank, highest for the most severe. An
// unrecognised or empty severity ranks lowest so it never outranks a real finding.
func severityRank(severity string) int {
	for i, s := range severityOrder {
		if s == severity {
			return len(severityOrder) - i
		}
	}
	return 0
}

var vulnsCmd = &cobra.Command{
	Use:     "vulns [search]",
	Aliases: []string{"vuln", "vulnerabilities", "findings"},
	Short:   "Query and list discovered vulnerabilities",
	Long:    UsageVulns(),
	RunE:    runVulns,
}

func init() {
	vulnsCmd.Flags().StringVarP(&vulnsWorkspace, "workspace", "w", "", "filter by workspace name")
	vulnsCmd.Flags().StringVar(&vulnsSeverity, "severity", "", "filter by exact severity (comma-separated, e.g., critical,high)")
	vulnsCmd.Flags().StringVar(&vulnsMinSeverity, "min-severity", "", "only show findings at or above this severity (info|low|medium|high|critical)")
	vulnsCmd.Flags().StringVar(&vulnsConfidence, "confidence", "", "filter by confidence (fuzzy match, e.g., certain)")
	vulnsCmd.Flags().StringVar(&vulnsTemplate, "template", "", "filter by detection template/module id (fuzzy match on vuln_info)")
	vulnsCmd.Flags().StringVar(&vulnsType, "type", "", "filter by asset_type (fuzzy match, e.g., http)")
	vulnsCmd.Flags().BoolVar(&vulnsStats, "stats", false, "show vulnerability statistics (severity breakdown, templates, confidences)")
	vulnsCmd.Flags().IntVar(&vulnsLimit, "limit", 50, "maximum number of records to return")
	vulnsCmd.Flags().IntVar(&vulnsOffset, "offset", 0, "number of records to skip (for pagination)")
	vulnsCmd.Flags().StringVar(&vulnsColumns, "columns", "", "comma-separated columns to display")
	vulnsCmd.Flags().StringVar(&vulnsExcludeColumns, "exclude-columns", "", "comma-separated columns to exclude from output")
	vulnsCmd.Flags().BoolVar(&vulnsAll, "all", false, "show all columns including hidden ones (id, timestamps)")
	vulnsCmd.Flags().StringArrayVar(&vulnsWhere, "where", nil, "filter by column (key=value, fuzzy match, repeatable)")
	vulnsCmd.Flags().StringVar(&vulnsSearch, "search", "", "search all columns for substring (case-insensitive)")
	vulnsCmd.Flags().StringVarP(&vulnsValue, "value", "V", "", "filter by asset_value (fuzzy match)")
	vulnsCmd.Flags().Int64Var(&vulnsDetail, "id", 0, "show one finding in full, including description, PoC and both HTTP messages")
}

func runVulns(cmd *cobra.Command, args []string) error {
	if err := connectDB(); err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	ctx := context.Background()

	if vulnsDetail > 0 {
		return runVulnDetail(ctx)
	}
	if vulnsStats {
		return runVulnsStats(ctx)
	}
	return runVulnsList(ctx, args)
}

func runVulnsList(ctx context.Context, args []string) error {
	validatePagination(&vulnsLimit, &vulnsOffset)

	// Exact-match filters
	filters := make(map[string]string)
	if vulnsWorkspace != "" {
		filters["workspace"] = vulnsWorkspace
	}

	// Org filter. An empty resolution means no org was selected, so no clause is
	// added and the query spans every org exactly as the assets listing does.
	orgUUID, err := resolveOrgUUID(ctx)
	if err != nil {
		return err
	}
	if orgUUID != "" {
		filters["org_uuid"] = orgUUID
	}

	// Fuzzy-match filters (LIKE)
	fuzzyFilters := parseWhereFilters(vulnsWhere)
	if vulnsConfidence != "" {
		fuzzyFilters["confidence"] = vulnsConfidence
	}
	if vulnsTemplate != "" {
		fuzzyFilters["vuln_info"] = vulnsTemplate
	}
	if vulnsType != "" {
		fuzzyFilters["asset_type"] = vulnsType
	}
	if vulnsValue != "" {
		fuzzyFilters["asset_value"] = vulnsValue
	}

	// A single --severity is an exact filter the database can apply; a list or a
	// --min-severity threshold is applied after the fetch, since GetTableRecords
	// takes one value per column.
	wanted, err := severitySelection()
	if err != nil {
		return err
	}
	if len(wanted) == 1 {
		filters["severity"] = wanted[0]
	}

	search := vulnsSearch
	if len(args) > 0 {
		search = args[0]
	}

	// Fetch a wider page when post-filtering, so the requested limit can still be met.
	fetchLimit := vulnsLimit
	postFilter := len(wanted) > 1
	if postFilter && fetchLimit < 10000 {
		fetchLimit = 10000
	}

	records, err := database.GetTableRecords(ctx, "vulnerabilities", vulnsOffset, fetchLimit, filters, fuzzyFilters, search, getDefaultExcludeColumns("vulnerabilities"))
	if err != nil {
		return fmt.Errorf("failed to get vulnerabilities: %w", err)
	}

	rows, ok := records.Records.([]database.Vulnerability)
	if !ok {
		return fmt.Errorf("unexpected record type %T from vulnerabilities query", records.Records)
	}
	total := records.TotalCount
	if postFilter {
		rows = filterBySeverity(rows, wanted)
		total = len(rows)
		if len(rows) > vulnsLimit {
			rows = rows[:vulnsLimit]
		}
	}

	sortBySeverity(rows)

	if globalJSON {
		jsonBytes, err := json.Marshal(rows)
		if err != nil {
			return fmt.Errorf("failed to format records: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	printer := terminal.NewPrinter()

	requestedColumns := parseColumns(vulnsColumns)
	columns := getEffectiveColumns("vulnerabilities", requestedColumns, vulnsAll)
	excludeColumns := parseExcludeColumns(vulnsExcludeColumns)

	startRecord := vulnsOffset + 1
	endRecord := vulnsOffset + len(rows)
	if total == 0 {
		startRecord = 0
	}

	printer.Info("Vulnerabilities")
	fmt.Printf("Showing records %d-%d of %d\n\n", startRecord, endRecord, total)

	// tableDefaultColumns has an entry for this table, so default columns are never hidden.
	renderTableWithTablewriter("vulnerabilities", rows, columns, globalWidth, false, excludeColumns)

	if total > endRecord {
		printer.Info("Next page: osmedeus vulns --offset %d --limit %d", vulnsOffset+len(rows), vulnsLimit)
	}
	if len(rows) > 0 {
		printer.Info("Full detail for one finding: osmedeus vulns --id <id>")
	}

	return nil
}

// severitySelection resolves --severity and --min-severity into the set of
// severities to show. An empty result means no severity filtering.
func severitySelection() ([]string, error) {
	if vulnsSeverity != "" && vulnsMinSeverity != "" {
		return nil, fmt.Errorf("only one of --severity or --min-severity can be specified")
	}

	if vulnsSeverity != "" {
		var out []string
		for _, s := range strings.Split(vulnsSeverity, ",") {
			s = strings.ToLower(strings.TrimSpace(s))
			if s == "" {
				continue
			}
			if severityRank(s) == 0 {
				return nil, fmt.Errorf("invalid severity %q. Must be one of: %s", s, strings.Join(severityOrder, ", "))
			}
			out = append(out, s)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("--severity was empty")
		}
		return out, nil
	}

	if vulnsMinSeverity != "" {
		min := strings.ToLower(strings.TrimSpace(vulnsMinSeverity))
		if severityRank(min) == 0 {
			return nil, fmt.Errorf("invalid --min-severity %q. Must be one of: %s", min, strings.Join(severityOrder, ", "))
		}
		// severityOrder is already most-severe-first, so everything at or above the
		// floor is just the prefix ending at it.
		for i, s := range severityOrder {
			if s == min {
				return severityOrder[:i+1], nil
			}
		}
	}

	return nil, nil
}

func filterBySeverity(rows []database.Vulnerability, wanted []string) []database.Vulnerability {
	keep := make(map[string]bool, len(wanted))
	for _, w := range wanted {
		keep[w] = true
	}
	out := make([]database.Vulnerability, 0, len(rows))
	for _, r := range rows {
		if keep[normalizeSeverity(r.Severity)] {
			out = append(out, r)
		}
	}
	return out
}

// sortBySeverity puts the most severe findings first, which is the order a human
// reads a finding list in.
func sortBySeverity(rows []database.Vulnerability) {
	sort.SliceStable(rows, func(i, j int) bool {
		return severityRank(normalizeSeverity(rows[i].Severity)) > severityRank(normalizeSeverity(rows[j].Severity))
	})
}

func normalizeSeverity(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "unknown"
	}
	return s
}

func runVulnsStats(ctx context.Context) error {
	orgUUID, err := resolveOrgUUID(ctx)
	if err != nil {
		return err
	}

	stats, err := database.GetVulnStats(ctx, vulnsWorkspace, orgUUID)
	if err != nil {
		return fmt.Errorf("failed to get vulnerability stats: %w", err)
	}

	if globalJSON {
		jsonBytes, err := json.Marshal(stats)
		if err != nil {
			return fmt.Errorf("failed to format stats: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	printer := terminal.NewPrinter()
	if vulnsWorkspace != "" {
		printer.Info("Vulnerability Statistics (workspace: %s)", vulnsWorkspace)
	} else {
		printer.Info("Vulnerability Statistics")
	}
	fmt.Println()

	fmt.Printf("%s: %d\n\n", terminal.Bold("Total"), stats.Total)

	fmt.Printf("%s:\n", terminal.Bold("Severity"))
	if stats.Total == 0 {
		fmt.Println("  (none)")
	} else {
		// highest severity first, and only severities actually present
		for _, sev := range append(append([]string{}, severityOrder...), "unknown") {
			if n, ok := stats.Severities[sev]; ok && n > 0 {
				fmt.Printf("  %s %-9s %d\n", terminal.SymbolBullet, sev, n)
			}
		}
	}
	fmt.Println()

	printStatCategory("Confidences", stats.Confidences)
	printStatCategory("Asset Types", stats.AssetTypes)
	printStatCategory("Templates", stats.Templates)
	if vulnsWorkspace == "" {
		printStatCategory("Workspaces", stats.Workspaces)
	}

	return nil
}

// runVulnDetail prints one finding in full. This is the only path that selects the
// heavy evidence columns, so a listing never pays for them.
func runVulnDetail(ctx context.Context) error {
	vuln, err := database.GetVulnerabilityByID(ctx, vulnsDetail)
	if err != nil {
		return fmt.Errorf("no vulnerability found with id %d (list them with 'osmedeus vulns')", vulnsDetail)
	}
	v := *vuln

	if globalJSON {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to format record: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	printer := terminal.NewPrinter()
	printer.Info("Vulnerability #%d", v.ID)
	fmt.Println()
	for _, f := range []struct{ label, value string }{
		{"Title", v.VulnTitle},
		{"Template", v.VulnInfo},
		{"Severity", v.Severity},
		{"Confidence", v.Confidence},
		{"Asset", v.AssetValue},
		{"Asset Type", v.AssetType},
		{"Workspace", v.Workspace},
		{"Tags", strings.Join(v.Tags, ", ")},
		{"Finding Hash", v.FindingHash},
	} {
		if val := strings.TrimSpace(f.value); val != "" {
			fmt.Printf("  %-13s %s\n", terminal.Bold(f.label)+":", val)
		}
	}

	for _, sec := range []struct{ label, value string }{
		{"Description", v.VulnDesc},
		{"Proof of Concept", v.VulnPOC},
		{"HTTP Request", v.DetailHTTPRequest},
		{"HTTP Response", v.DetailHTTPResponse},
	} {
		if val := strings.TrimSpace(sec.value); val != "" {
			fmt.Printf("\n%s\n%s\n", terminal.BoldCyan("▷ "+sec.label), val)
		}
	}

	return nil
}

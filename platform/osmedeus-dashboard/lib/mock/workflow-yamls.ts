export const MOCK_WORKFLOW_YAMLS: Record<string, string> = {
  "test-complex-docker-workflow": `name: test-complex-docker-workflow
kind: module
description: Complex workflow demonstrating bash, function steps with docker step_runner

params:
  - name: target
    required: true
  - name: output_dir
    default: /tmp/osm-complex-test
  - name: threads
    default: "5"

steps:
  # Step 1: Setup - Create directories using function
  - name: setup-workspace
    type: function
    log: "Setting up workspace for {{target}}"
    function: createDir("{{output_dir}}")
    exports:
      workspace_created: "output"

  # Step 2: Create input file with bash
  - name: create-target-list
    type: bash
    log: "Creating target list for {{target}}"
    commands:
      - mkdir -p {{output_dir}}/targets
      - |
        cat > {{output_dir}}/targets/hosts.txt << 'EOF'
        sub1.{{target}}
        sub2.{{target}}
        api.{{target}}
        www.{{target}}
        admin.{{target}}
        EOF
    exports:
      target_file: "{{output_dir}}/targets/hosts.txt"

  # Step 3: Docker-based DNS resolution simulation
  - name: dns-resolve
    type: remote-bash
    log: "Resolving DNS for targets in Docker"
    timeout: 60
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      env:
        TARGET_DOMAIN: "{{target}}"
      volumes:
        - "{{output_dir}}:/workspace"
      workdir: /workspace
    command: |
      echo "Resolving DNS for $TARGET_DOMAIN"
      cat /workspace/targets/hosts.txt | while read host; do
        echo "$host -> 127.0.0.1" >> /workspace/dns-resolved.txt
      done
      echo "DNS resolution complete"
    exports:
      dns_output: "{{output_dir}}/dns-resolved.txt"

  # Step 4: Parallel docker commands - simulating port scanning
  - name: parallel-port-scan
    type: remote-bash
    log: "Running parallel port scans in Docker"
    timeout: 120
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    parallel_commands:
      - 'echo "Scanning ports 1-1000 on {{target}}" && sleep 1 && echo "Port 80 open" > /workspace/ports-1.txt'
      - 'echo "Scanning ports 1001-2000 on {{target}}" && sleep 1 && echo "Port 443 open" > /workspace/ports-2.txt'
      - 'echo "Scanning ports 2001-3000 on {{target}}" && sleep 1 && echo "Port 8080 open" > /workspace/ports-3.txt'
      - 'echo "Scanning ports 3001-4000 on {{target}}" && sleep 1 && echo "Port 3306 open" > /workspace/ports-4.txt'

  # Step 5: Merge port scan results
  - name: merge-port-results
    type: bash
    log: "Merging port scan results"
    command: cat {{output_dir}}/ports-*.txt > {{output_dir}}/all-ports.txt
    exports:
      ports_file: "{{output_dir}}/all-ports.txt"

  # Step 6: Function to check file existence
  - name: verify-ports-file
    type: function
    log: "Verifying ports file exists"
    function: fileExists("{{ports_file}}")
    exports:
      ports_verified: "output"

  # Step 7: Docker-based HTTP probing with parallel steps
  - name: http-probe-parallel
    type: parallel-steps
    log: "Running parallel HTTP probes"
    parallel_steps:
      - name: probe-http
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTP on port 80"
          echo "http://{{target}}:80 [200]" > /workspace/http-80.txt
      - name: probe-https
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing HTTPS on port 443"
          echo "https://{{target}}:443 [200]" > /workspace/https-443.txt
      - name: probe-alt
        type: remote-bash
        step_runner: docker
        step_runner_config:
          image: alpine:latest
          volumes:
            - "{{output_dir}}:/workspace"
        command: |
          echo "Probing alternate port 8080"
          echo "http://{{target}}:8080 [404]" > /workspace/http-8080.txt

  # Step 8: Foreach loop with docker - process each subdomain
  - name: process-subdomains
    type: foreach
    log: "Processing each subdomain"
    input: "{{output_dir}}/targets/hosts.txt"
    variable: subdomain
    threads: 3
    step:
      name: scan-subdomain
      type: remote-bash
      step_runner: docker
      step_runner_config:
        image: alpine:latest
        volumes:
          - "{{output_dir}}:/workspace"
      command: |
        echo "Scanning [[subdomain]]..."
        echo "[[subdomain]]: status=200, title=Example" >> /workspace/subdomain-results.txt

  # Step 9: Read results with function
  - name: read-subdomain-results
    type: function
    log: "Reading subdomain scan results"
    function: readFile("{{output_dir}}/subdomain-results.txt")
    exports:
      scan_results: "output"

  # Step 10: Decision based routing
  - name: check-results
    type: bash
    log: "Checking scan results"
    command: wc -l < {{output_dir}}/subdomain-results.txt
    exports:
      result_count: "output"
    decision:
      - condition: result_count == "0"
        next: "_end"
      - condition: result_count != "0"
        next: "generate-report"

  # Step 11: Generate final report in docker
  - name: generate-report
    type: remote-bash
    log: "Generating final report"
    timeout: 30
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - "{{output_dir}}:/workspace"
    commands:
      - echo "=== Scan Report for {{target}} ===" > /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- DNS Results ---" >> /workspace/report.txt
      - cat /workspace/dns-resolved.txt >> /workspace/report.txt 2>/dev/null || echo "No DNS results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Open Ports ---" >> /workspace/report.txt
      - cat /workspace/all-ports.txt >> /workspace/report.txt 2>/dev/null || echo "No ports found" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "--- Subdomain Results ---" >> /workspace/report.txt
      - cat /workspace/subdomain-results.txt >> /workspace/report.txt 2>/dev/null || echo "No subdomain results" >> /workspace/report.txt
      - echo "" >> /workspace/report.txt
      - echo "Report generated at $(date)" >> /workspace/report.txt
    exports:
      report_file: "{{output_dir}}/report.txt"

  # Step 12: Parallel functions to get file stats
  - name: get-file-stats
    type: function
    log: "Getting file statistics"
    parallel_functions:
      - fileLength("{{output_dir}}/report.txt")
      - fileExists("{{output_dir}}/all-ports.txt")
      - trim("  {{target}}  ")
    exports:
      file_stats: "output"

  # Step 13: Cleanup (optional - controlled by pre_condition)
  - name: cleanup-temp-files
    type: bash
    log: "Cleaning up temporary files"
    pre_condition: "false"
    command: rm -rf {{output_dir}}/ports-*.txt
    on_error:
      - action: log
        message: "Cleanup failed but continuing"
      - action: continue
`,
  "test-decision": `name: test-decision
kind: module
description: Test conditional step routing with decision

params:
  - name: target
    required: true

steps:
  - name: check-condition
    type: bash
    command: echo "{{target}}"
    exports:
      target_value: "output"
    decision:
      - condition: target_value == "skip"
        next: "_end"
      - condition: target_value == "jump"
        next: "final-step"

  - name: middle-step
    type: bash
    command: echo "middle executed"
    exports:
      middle_output: "output"

  - name: final-step
    type: bash
    command: echo "final executed"
    exports:
      final_output: "output"
`,
  "test-docker-flow": `name: test-docker-flow
kind: flow
description: Flow orchestrating multiple Docker-based security scanning modules

params:
  - name: target
    required: true
  - name: Output
    default: /tmp/osm-docker-flow
  - name: mode
    default: "full"
  - name: threads
    default: "10"
  - name: skip_vuln_scan
    default: "false"

modules:
  # Module 1: Initial reconnaissance
  - name: recon-module
    path: modules/test-docker-recon
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/recon"
      threads: "{{threads}}"
    on_success:
      - action: log
        message: "Reconnaissance completed for {{target}}"
      - action: export
        key: recon_complete
        value: "true"
    on_error:
      - action: log
        message: "Reconnaissance failed for {{target}}"
      - action: abort

  # Module 2: Subdomain enumeration (depends on recon)
  - name: subdomain-module
    path: modules/test-docker-subdomain
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/subdomains"
      wordlist: "/usr/share/wordlists/subdomains.txt"
    condition: "mode == 'full' || mode == 'subdomain'"
    on_success:
      - action: export
        key: subdomains_file
        value: "{{Output}}/subdomains/all.txt"

  # Module 3: Port scanning (parallel with subdomain)
  - name: portscan-module
    path: modules/test-docker-portscan
    depends_on:
      - recon-module
    params:
      target: "{{target}}"
      output_dir: "{{Output}}/ports"
      port_range: "1-10000"
      rate: "1000"
    condition: "mode == 'full' || mode == 'portscan'"

  # Module 4: HTTP probing (depends on subdomain results)
  - name: httpx-module
    path: modules/test-docker-httpx
    depends_on:
      - subdomain-module
    params:
      input: "{{subdomains_file}}"
      output_dir: "{{Output}}/http"
      threads: "{{threads}}"
    on_success:
      - action: export
        key: alive_hosts
        value: "{{Output}}/http/alive.txt"
      - action: export
        key: httpx_json
        value: "{{Output}}/http/httpx.json"
    decision:
      - condition: "fileLength('{{Output}}/http/alive.txt') == 0"
        next: "report-module"

  # Module 5: Technology detection (depends on HTTP probe)
  - name: tech-detect-module
    path: modules/test-docker-techdetect
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/tech"

  # Module 6: Screenshot capture (parallel with tech detection)
  - name: screenshot-module
    path: modules/test-docker-screenshot
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/screenshots"
      threads: "5"

  # Module 7: Vulnerability scanning (conditional)
  - name: vulnscan-module
    path: modules/test-docker-scanning
    depends_on:
      - httpx-module
      - tech-detect-module
    params:
      target: "{{target}}"
      Output: "{{Output}}/vulns"
      severity: "critical,high,medium"
      threads: "{{threads}}"
    condition: "skip_vuln_scan != 'true'"
    on_error:
      - action: log
        message: "Vulnerability scan encountered errors but continuing"
      - action: continue

  # Module 8: Directory bruteforcing (optional - depends on mode)
  - name: dirbrute-module
    path: modules/test-docker-dirbrute
    depends_on:
      - httpx-module
    params:
      input: "{{alive_hosts}}"
      output_dir: "{{Output}}/dirs"
      wordlist: "/usr/share/wordlists/common.txt"
      threads: "20"
    condition: "mode == 'full'"

  # Module 9: JavaScript analysis (depends on dir results)
  - name: js-analysis-module
    path: modules/test-docker-jsanalysis
    depends_on:
      - dirbrute-module
    params:
      input: "{{Output}}/dirs/js-files.txt"
      output_dir: "{{Output}}/js"
    condition: "mode == 'full'"

  # Module 10: Final report generation
  - name: report-module
    path: modules/test-docker-report
    depends_on:
      - screenshot-module
      - vulnscan-module
      - tech-detect-module
    params:
      target: "{{target}}"
      input_dir: "{{Output}}"
      output_dir: "{{Output}}/reports"
      format: "html,json,markdown"
    on_success:
      - action: log
        message: "Flow completed successfully for {{target}}"
      - action: notify
        message: "Security assessment complete: {{target}}"
`,
  "test-loop": `name: test-loop
kind: module
description: Test foreach loop with threading

params:
  - name: target
    required: true

steps:
  - name: create-input
    type: bash
    commands:
      - mkdir -p {{Output}}
      - printf 'one\\ntwo\\nthree\\nfour\\nfive\\n' > {{Output}}/items.txt

  - name: process-items
    type: foreach
    input: "{{Output}}/items.txt"
    variable: item
    threads: 2
    step:
      name: process-item
      type: bash
      command: echo "Processing [[item]] for {{target}}"
`,
  "comprehensive-flow-example": `# =============================================================================
# Flow Workflow: Comprehensive Example
# =============================================================================
# This file demonstrates ALL fields available in a flow-kind workflow.
# Flows orchestrate multiple modules with dependencies, conditions, and routing.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# Same as module workflows (kind, name, description, tags, params, etc.)
# -----------------------------------------------------------------------------

# kind: Workflow type - "flow" orchestrates multiple modules
kind: flow

# name: Unique identifier for this workflow (required)
name: comprehensive-flow-example

# description: Human-readable description
description: Demonstrates all flow-specific fields including modules, dependencies, conditions, and decisions

# tags: Comma-separated tags for filtering
tags: flow, comprehensive, example

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Parameters available to all modules in this flow
# -----------------------------------------------------------------------------
params:
  - name: threads
    default: "10"

  - name: timeout
    default: "3600"

  - name: scan_depth
    default: "normal"

  - name: output_format
    default: "json"

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Flow-level dependencies checked before any module executes
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - nmap
    - nuclei
    - httpx

  files:
    - /tmp

  variables:
    - name: Target
      type: domain
      required: true

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Reports aggregated from all modules in this flow
# -----------------------------------------------------------------------------
reports:
  - name: flow-summary
    path: "{{Output}}/flow-summary.json"
    type: json
    description: Aggregated results from all modules

  - name: vulnerabilities
    path: "{{Output}}/vulnerabilities.txt"
    type: text
    description: All discovered vulnerabilities

# -----------------------------------------------------------------------------
# PREFERENCES SECTION
# Flow-level preferences apply to all module executions
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: false
  heuristics_check: 'basic'

# -----------------------------------------------------------------------------
# MODULES SECTION (Flow-specific)
# Ordered list of module references to execute
# =============================================================================
modules:
  # ===========================================================================
  # Module Reference: Basic Configuration
  # ===========================================================================
  - # name: Display name for this module execution (required)
    name: reconnaissance

    # path: Path to the module YAML file (required)
    # Can be relative to workflows directory or absolute
    path: modules/recon.yaml

    # params: Parameters to pass to this module
    # Overrides module defaults and flow-level params
    params:
      threads: "20"  # Override flow-level threads
      output_dir: "{{Output}}/recon"

  # ===========================================================================
  # Module Reference: With Dependencies (depends_on)
  # ===========================================================================
  - name: port-scanning
    path: modules/portscan.yaml

    # depends_on: List of module names that must complete before this module runs
    # Creates a DAG (Directed Acyclic Graph) for execution order
    depends_on:
      - reconnaissance

    params:
      target_list: "{{Output}}/recon/subdomains.txt"
      threads: "{{threads}}"

  # ===========================================================================
  # Module Reference: With Condition
  # ===========================================================================
  - name: web-scanning
    path: modules/webscan.yaml

    depends_on:
      - port-scanning

    # condition: JavaScript expression - module only runs if evaluates to true
    # Can reference exported variables from previous modules
    condition: 'fileLength("{{Output}}/portscan/http-services.txt") > 0'

    params:
      input: "{{Output}}/portscan/http-services.txt"

  # ===========================================================================
  # Module Reference: With on_success Handler
  # ===========================================================================
  - name: vulnerability-scanning
    path: modules/vuln-scan.yaml

    depends_on:
      - web-scanning

    condition: 'fileExists("{{Output}}/webscan/endpoints.txt")'

    params:
      endpoints: "{{Output}}/webscan/endpoints.txt"
      timeout: "{{timeout}}"

    # on_success: Actions to execute when this module completes successfully
    on_success:
      # action: log - Log a message
      - action: log
        message: "Vulnerability scanning completed for {{Target}}"

      # action: export - Export a variable for subsequent modules
      - action: export
        name: vuln_scan_complete
        value: "true"

      # action: notify - Send a notification
      - action: notify
        notify: "Vulnerability scan finished for {{Target}}"

      # action: run - Execute a follow-up step
      - action: run
        type: bash
        command: 'echo "Vuln scan done" >> {{Output}}/flow-log.txt'

      # action: run with functions
      - action: run
        type: function
        functions:
          - 'log_info("Module completed successfully")'

  # ===========================================================================
  # Module Reference: With on_error Handler
  # ===========================================================================
  - name: exploit-verification
    path: modules/exploit-verify.yaml

    depends_on:
      - vulnerability-scanning

    condition: '{{vuln_scan_complete}} == "true"'

    params:
      vulns_file: "{{Output}}/vuln-scan/vulnerabilities.json"

    # on_error: Actions to execute when this module fails
    on_error:
      # action: log - Log error message
      - action: log
        message: "Exploit verification failed for {{Target}}"
        # condition: Only execute if this condition is true
        condition: 'true'

      # action: continue - Allow flow to continue despite error
      - action: continue
        message: "Continuing flow despite exploit verification failure"

      # action: abort - Stop the entire flow
      # (Usually with a condition so it doesn't always abort)
      - action: abort
        message: "Critical failure - aborting flow"
        condition: 'false'  # Only abort under specific conditions

      # action: notify - Alert on failure
      - action: notify
        notify: "Module failed: exploit-verification for {{Target}}"

      # action: export - Export error state
      - action: export
        name: exploit_verify_failed
        value: "true"

  # ===========================================================================
  # Module Reference: With Decision Routing
  # ===========================================================================
  - name: deep-scan
    path: modules/deep-scan.yaml

    depends_on:
      - vulnerability-scanning

    # decision: Conditional routing based on results
    # Determines which module to execute next based on conditions
    decision:
      # condition: JavaScript expression to evaluate
      # next: Module name to jump to, or "_end" to finish flow
      - condition: 'fileLength("{{Output}}/vuln-scan/critical.txt") > 0'
        next: notification-critical

      - condition: 'fileLength("{{Output}}/vuln-scan/high.txt") > 0'
        next: notification-high

      # Default case - continue to next module in list
      - condition: 'true'
        next: cleanup

    params:
      scan_depth: "{{scan_depth}}"

  # ===========================================================================
  # Module Reference: Notification branches (targets of decision routing)
  # ===========================================================================
  - name: notification-critical
    path: modules/notify.yaml

    # Note: This module can be jumped to via decision routing
    # It won't run in normal sequential flow unless explicitly in depends_on

    params:
      severity: critical
      message: "Critical vulnerabilities found for {{Target}}"
      channel: security-alerts

    on_success:
      - action: export
        name: notification_sent
        value: "critical"

  - name: notification-high
    path: modules/notify.yaml

    params:
      severity: high
      message: "High severity vulnerabilities found for {{Target}}"
      channel: security-team

    on_success:
      - action: export
        name: notification_sent
        value: "high"

  # ===========================================================================
  # Module Reference: Parallel Module Execution
  # Modules with same depends_on and no inter-dependencies run in parallel
  # ===========================================================================
  - name: ssl-analysis
    path: modules/ssl-check.yaml

    depends_on:
      - port-scanning  # Same dependency as web-scanning

    params:
      input: "{{Output}}/portscan/ssl-services.txt"

  - name: dns-analysis
    path: modules/dns-check.yaml

    depends_on:
      - reconnaissance  # Can run in parallel with port-scanning

    params:
      domains: "{{Output}}/recon/subdomains.txt"

  # ===========================================================================
  # Module Reference: Cleanup/Final Module
  # ===========================================================================
  - name: cleanup
    path: modules/cleanup.yaml

    # depends_on multiple modules - waits for all to complete
    depends_on:
      - vulnerability-scanning
      - exploit-verification
      - ssl-analysis
      - dns-analysis

    # condition with multiple checks
    condition: 'true'  # Always run cleanup

    params:
      output_dir: "{{Output}}"
      format: "{{output_format}}"

    on_success:
      - action: log
        message: "Flow completed successfully for {{Target}}"

      - action: notify
        notify: "Security scan flow completed for {{Target}}"

      - action: export
        name: flow_status
        value: "completed"

    on_error:
      - action: log
        message: "Cleanup failed but flow results are preserved"

      - action: continue
        message: "Flow complete despite cleanup issues"
`,
  "triggers-example": `# =============================================================================
# Flow Workflow: All Trigger Types Example
# =============================================================================
# This file demonstrates ALL trigger types available in osmedeus workflows.
# Triggers define when/how a workflow should automatically execute.
# Trigger types: cron, event, watch, manual
# =============================================================================

kind: flow
name: triggers-example
description: Demonstrates all trigger types with comprehensive field documentation
tags: triggers, automation, scheduled

# -----------------------------------------------------------------------------
# TRIGGERS SECTION
# Define automatic execution triggers for this workflow
# Multiple triggers can be defined; any triggered condition will start execution
# =============================================================================
trigger:
  # ===========================================================================
  # TRIGGER TYPE: cron
  # Schedule-based execution using cron expressions
  # ===========================================================================
  - # name: Identifier for this trigger (for logging and management)
    name: daily-scan

    # on: Trigger type - cron, event, watch, or manual
    on: cron

    # schedule: Cron expression defining when to run
    # Format: minute hour day-of-month month day-of-week
    # Examples:
    #   "0 0 * * *"     - Every day at midnight
    #   "0 */6 * * *"   - Every 6 hours
    #   "0 9 * * 1-5"   - 9 AM on weekdays
    #   "0 0 1 * *"     - First day of every month at midnight
    schedule: "0 2 * * *"  # Every day at 2 AM

    # input: Defines where the target input comes from for scheduled runs
    input:
      # type: Input source type - file, event_data, function, or param
      type: file

      # path: For "file" type - path to file containing targets (one per line)
      path: "/data/targets/active-targets.txt"

    # enabled: Whether this trigger is active
    # true = trigger is active and will fire
    # false = trigger is defined but disabled
    enabled: true

  # ---------------------------------------------------------------------------
  # Cron trigger with function-based input
  # ---------------------------------------------------------------------------
  - name: weekly-full-scan
    on: cron
    schedule: "0 0 * * 0"  # Every Sunday at midnight

    input:
      # type: function - Generate input dynamically using a function
      type: function

      # function: JavaScript function to generate/retrieve targets
      # Can use built-in functions like db queries, API calls, etc.
      function: 'get_targets_from_db("scope:production")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: event
  # Event-driven execution based on system events
  # Events follow topic format: <component>.<event_type>
  # ===========================================================================
  - name: webhook-trigger
    on: event

    # event: Event configuration for event triggers
    event:
      # topic: Event topic to subscribe to
      # Common topics:
      #   webhook.received    - External webhook received
      #   assets.new          - New asset discovered
      #   assets.changed      - Asset data changed
      #   db.change           - Database record changed
      #   watch.files         - File system change detected
      topic: webhook.received

      # filters: JavaScript expressions to filter events
      # Event data available as 'event' object with fields:
      #   event.name      - Event name
      #   event.source    - Event source
      #   event.data      - JSON payload (string)
      #   event.data_type - Type of data
      # All filters must evaluate to true for trigger to fire
      filters:
        - 'event.source == "github"'
        - 'event.name == "push"'

    # input: How to extract target from event data
    input:
      # type: event_data - Extract from event payload
      type: event_data

      # field: JSON path to extract from event.data
      # Uses dot notation for nested fields
      field: "repository.html_url"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger for new asset discovery
  # ---------------------------------------------------------------------------
  - name: new-asset-scan
    on: event

    event:
      topic: assets.new

      filters:
        # Filter for specific asset types
        - 'event.data_type == "subdomain"'
        # Filter by source tool
        - 'event.source == "subfinder" || event.source == "amass"'

    input:
      type: event_data
      field: "hostname"

    enabled: true

  # ---------------------------------------------------------------------------
  # Event trigger with function-based input extraction
  # ---------------------------------------------------------------------------
  - name: vuln-alert-trigger
    on: event

    event:
      topic: webhook.received

      filters:
        - 'event.name == "vulnerability_alert"'
        - 'JSON.parse(event.data).severity == "critical"'

    input:
      # type: function - Use function to parse/transform event data
      type: function

      # function: Transform event data to target format
      function: 'jq("{{event.data}}", ".affected_host")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: watch
  # File system watch - triggers when files change
  # ===========================================================================
  - name: targets-file-watch
    on: watch

    # path: File or directory path to watch for changes
    # Supports glob patterns in some implementations
    path: "/data/targets/new-targets.txt"

    # input: How to get targets when file changes
    input:
      type: file
      path: "/data/targets/new-targets.txt"

    enabled: true

  # ---------------------------------------------------------------------------
  # Watch trigger on directory
  # ---------------------------------------------------------------------------
  - name: input-directory-watch
    on: watch

    path: "/data/incoming/"

    input:
      # type: function - Process newly added files
      type: function
      function: 'get_new_files("/data/incoming/", "*.txt")'

    enabled: true

  # ===========================================================================
  # TRIGGER TYPE: manual
  # Explicit manual trigger control
  # Used to enable/disable CLI execution for this workflow
  # ===========================================================================
  - name: manual-execution
    on: manual

    # For manual triggers, enabled controls whether CLI can run this workflow
    # enabled: true  - Allow: osmedeus run -f triggers-example -t target
    # enabled: false - Block CLI execution (only scheduled/event triggers work)
    enabled: true

    # input: Default input for manual execution
    # This is optional; CLI -t flag overrides this
    input:
      # type: param - Use a parameter as input
      type: param

      # name: Parameter name to use as target
      name: Target

  # ---------------------------------------------------------------------------
  # Disabled manual trigger example
  # This workflow can ONLY be triggered via cron/events, not CLI
  # ---------------------------------------------------------------------------
  # Uncomment to see the effect:
  # - name: block-manual
  #   on: manual
  #   enabled: false

# -----------------------------------------------------------------------------
# PARAMS SECTION
# -----------------------------------------------------------------------------
params:
  - name: scan_type
    default: "standard"

  - name: threads
    default: "10"

# -----------------------------------------------------------------------------
# MODULES SECTION
# The actual workflow steps to execute when any trigger fires
# -----------------------------------------------------------------------------
modules:
  - name: initial-recon
    path: modules/recon.yaml
    params:
      threads: "{{threads}}"

  - name: scanning
    path: modules/scan.yaml
    depends_on:
      - initial-recon
    params:
      scan_type: "{{scan_type}}"

  - name: reporting
    path: modules/report.yaml
    depends_on:
      - scanning

    on_success:
      - action: notify
        notify: "Triggered scan completed for {{Target}}"
        # condition: Only notify for certain triggers
        condition: 'true'

      - action: export
        name: completed_at
        value: "{{currentDate()}}"
`,
  "docker-runner-example": `# =============================================================================
# Module Workflow: Docker Runner Configuration Example
# =============================================================================
# This file demonstrates all Docker runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: docker-runner-example
description: Demonstrates Docker runner configuration with all available fields
tags: docker, runner, container

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: docker

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # DOCKER-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # image: Docker image to use (required for docker runner)
  # Format: registry/image:tag or just image:tag
  image: ubuntu:22.04

  # env: Environment variables to set inside the container
  # Map of VAR_NAME: value
  env:
    MY_VAR: my-value
    API_KEY: "{{api_key}}"  # Can use template variables
    THREADS: "{{threads}}"

  # volumes: Volume mounts in docker format
  # Format: host_path:container_path[:options]
  # Options: ro (read-only), rw (read-write)
  volumes:
    - "/tmp/osmedeus:/data"
    - "{{Output}}:/output"
    - "/etc/hosts:/etc/hosts:ro"

  # network: Docker network mode
  # Options: bridge (default), host, none, container:<name>, or network name
  network: host

  # persistent: Container lifecycle mode
  # true = reuse the same container across steps (faster, state preserved)
  # false = ephemeral, create new container per step (isolated, clean state)
  persistent: true

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory inside the container/remote
  # Commands will execute in this directory
  workdir: /app

params:
  - name: api_key
    default: "demo-key"

  - name: threads
    default: "5"

steps:
  # ===========================================================================
  # Step using workflow-level runner (docker with ubuntu:22.04)
  # ===========================================================================
  - name: use-workflow-runner
    type: bash
    log: "Running in workflow-level Docker container"
    command: 'echo "Running inside ubuntu:22.04 container"'

  # ===========================================================================
  # Step with per-step Docker runner override
  # Uses different image than workflow-level config
  # ===========================================================================
  - name: step-with-runner-override
    type: bash
    log: "Running in step-specific Docker container"

    # step_runner: Override runner type for this step only
    # Options: host, docker, ssh
    step_runner: docker

    # step_runner_config: Override runner configuration for this step
    # Same structure as runner_config but applies only to this step
    step_runner_config:
      # Use a different image for this specific step
      image: python:3.11-slim

      env:
        PYTHONPATH: /app

      volumes:
        - "{{Output}}:/output:rw"

      network: bridge

      persistent: false

      workdir: /app

    command: 'python3 -c "print(\\"Running in Python container\\")"'

  # ===========================================================================
  # Remote-bash step type with Docker (explicit remote-bash type)
  # remote-bash is specifically for executing commands in remote environments
  # ===========================================================================
  - name: remote-bash-docker
    # type: remote-bash is specifically for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution in Docker"

    # step_runner: Required for remote-bash type - specifies execution environment
    # Must be "docker" or "ssh"
    step_runner: docker

    step_runner_config:
      image: alpine:latest
      workdir: /tmp

    # command/commands/parallel_commands: Same as bash step
    command: 'echo "Hello from Alpine container" > /tmp/output.txt'

    # step_remote_file: File path on remote (inside container) to copy after execution
    # This file will be copied from the container to the host
    step_remote_file: /tmp/output.txt

    # host_output_file: Local path where the remote file will be copied
    # Template variables are supported
    host_output_file: "{{Output}}/docker-output.txt"

  # ===========================================================================
  # Parallel commands in Docker container
  # ===========================================================================
  - name: docker-parallel-commands
    type: bash
    log: "Running parallel commands in Docker"
    step_runner: docker
    step_runner_config:
      image: ubuntu:22.04
      persistent: true

    parallel_commands:
      - 'sleep 2 && echo "Parallel job A completed"'
      - 'sleep 1 && echo "Parallel job B completed"'
      - 'sleep 3 && echo "Parallel job C completed"'

  # ===========================================================================
  # Foreach loop executing in Docker
  # ===========================================================================
  - name: docker-foreach
    type: foreach
    log: "Processing items in Docker containers"
    input: "{{Output}}/targets.txt"
    variable: target
    threads: 3

    step:
      name: process-in-docker
      type: bash
      step_runner: docker
      step_runner_config:
        image: curlimages/curl:latest
        network: host
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[target]]"'
      exports:
        http_status: "{{stdout}}"

  # ===========================================================================
  # Step running on host (override workflow's docker runner)
  # ===========================================================================
  - name: run-on-host
    type: bash
    log: "Running on host machine (overriding workflow runner)"

    # Override to run locally instead of in container
    step_runner: host

    command: 'echo "This runs directly on the host machine"'

  # ===========================================================================
  # Docker step with all structured arguments
  # ===========================================================================
  - name: docker-with-args
    type: bash
    log: "Docker step with structured arguments"
    step_runner: docker
    step_runner_config:
      image: nuclei:latest
      volumes:
        - "{{Output}}:/output"
        - "/root/nuclei-templates:/templates:ro"
      workdir: /output

    command: nuclei
    speed_args: '-rate-limit 100 -c {{threads}}'
    config_args: '-t /templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /output/nuclei-results.txt'

    step_remote_file: /output/nuclei-results.txt
    host_output_file: "{{Output}}/nuclei-results.txt"

    exports:
      nuclei_output: "{{Output}}/nuclei-results.txt"
`,
  "ssh-runner-example": `# =============================================================================
# Module Workflow: SSH Runner Configuration Example
# =============================================================================
# This file demonstrates all SSH runner configuration fields at both
# the workflow level (for all steps) and step level (per-step override).
# =============================================================================

kind: module
name: ssh-runner-example
description: Demonstrates SSH runner configuration with all available fields
tags: ssh, runner, remote

# -----------------------------------------------------------------------------
# RUNNER CONFIGURATION (Workflow-Level)
# Applies to all steps unless overridden at step level
# -----------------------------------------------------------------------------

# runner: Execution environment for this workflow
# Options: host (default - local machine), docker, ssh
runner: ssh

# runner_config: Configuration for the selected runner type
runner_config:
  # -------------------------------------------------------------------------
  # SSH-SPECIFIC CONFIGURATION
  # -------------------------------------------------------------------------

  # host: SSH hostname or IP address (required for ssh runner)
  # Can use template variables for dynamic targeting
  host: "{{ssh_host}}"

  # port: SSH port number
  # Default: 22
  port: 22

  # user: SSH username for authentication
  user: "{{ssh_user}}"

  # key_file: Path to SSH private key file for key-based authentication
  # Preferred over password authentication for security
  key_file: "{{ssh_key_path}}"

  # password: SSH password for password-based authentication
  # WARNING: Not recommended - use key_file instead when possible
  # Can use template variables or environment references
  # password: "{{ssh_password}}"

  # -------------------------------------------------------------------------
  # COMMON CONFIGURATION (applies to docker and ssh)
  # -------------------------------------------------------------------------

  # workdir: Working directory on the remote machine
  # Commands will execute in this directory
  workdir: /home/scanner/workspace

params:
  - name: ssh_host
    default: "192.168.1.100"
    required: true

  - name: ssh_user
    default: "scanner"
    required: true

  - name: ssh_key_path
    default: "~/.ssh/id_rsa"

  - name: threads
    default: "10"

steps:
  # ===========================================================================
  # Step using workflow-level SSH runner
  # ===========================================================================
  - name: setup-remote-workspace
    type: bash
    log: "Setting up workspace on remote SSH server"
    command: 'mkdir -p /home/scanner/workspace/results && echo "Workspace ready"'

  # ===========================================================================
  # Remote-bash step type with SSH (explicit remote-bash type)
  # remote-bash is specifically designed for remote execution scenarios
  # ===========================================================================
  - name: remote-bash-ssh
    # type: remote-bash is explicitly for remote execution (docker/ssh)
    type: remote-bash
    log: "Remote bash execution via SSH"

    # step_runner: Required for remote-bash type - must be "docker" or "ssh"
    step_runner: ssh

    # step_runner_config: SSH configuration (inherits from workflow if not set)
    # Omitting this uses workflow-level runner_config
    step_runner_config:
      host: "{{ssh_host}}"
      port: 22
      user: "{{ssh_user}}"
      key_file: "{{ssh_key_path}}"
      workdir: /tmp

    # command: Command to execute on remote server
    command: 'hostname && whoami && pwd > /tmp/remote-info.txt'

    # step_remote_file: File on remote server to copy back to local host
    # This is useful for retrieving results from remote execution
    step_remote_file: /tmp/remote-info.txt

    # host_output_file: Local path where remote file will be copied
    host_output_file: "{{Output}}/remote-info.txt"

    exports:
      remote_file: "{{Output}}/remote-info.txt"

  # ===========================================================================
  # Step overriding SSH connection to different server
  # ===========================================================================
  - name: connect-to-secondary-server
    type: bash
    log: "Connecting to secondary server"

    # Override workflow runner with different SSH target
    step_runner: ssh

    step_runner_config:
      host: "192.168.1.101"  # Different server
      port: 2222             # Non-standard port
      user: admin
      key_file: "~/.ssh/secondary_key"
      workdir: /opt/scanner

    command: 'echo "Connected to secondary server" && uptime'

  # ===========================================================================
  # Multiple sequential commands via SSH
  # ===========================================================================
  - name: ssh-multiple-commands
    type: bash
    log: "Running multiple commands on remote"

    # commands: List of commands executed sequentially on remote
    commands:
      - 'echo "Step 1: Checking system"'
      - 'df -h'
      - 'echo "Step 2: Checking memory"'
      - 'free -m'
      - 'echo "Step 3: Checking processes"'
      - 'ps aux | head -10'

    std_file: "{{Output}}/system-check.txt"

  # ===========================================================================
  # Parallel commands on SSH (run concurrently on remote)
  # ===========================================================================
  - name: ssh-parallel-commands
    type: bash
    log: "Running parallel commands on remote SSH server"

    parallel_commands:
      - 'nmap -sS -p 80 {{Target}} > /tmp/port80.txt'
      - 'nmap -sS -p 443 {{Target}} > /tmp/port443.txt'
      - 'nmap -sS -p 22 {{Target}} > /tmp/port22.txt'

  # ===========================================================================
  # Run tool with structured arguments via SSH
  # ===========================================================================
  - name: ssh-nuclei-scan
    type: bash
    log: "Running nuclei scan via SSH"
    timeout: 3600

    command: nuclei
    speed_args: '-rate-limit 50 -c {{threads}}'
    config_args: '-t ~/nuclei-templates/cves/'
    input_args: '-u {{Target}}'
    output_args: '-o /home/scanner/workspace/nuclei-results.json -json'

    step_remote_file: /home/scanner/workspace/nuclei-results.json
    host_output_file: "{{Output}}/nuclei-results.json"

    exports:
      scan_results: "{{Output}}/nuclei-results.json"

  # ===========================================================================
  # Foreach loop with SSH execution
  # Processes multiple targets on remote server
  # ===========================================================================
  - name: ssh-foreach-targets
    type: foreach
    log: "Processing targets via SSH"

    # input: File containing targets (one per line)
    input: "{{Output}}/targets.txt"

    # variable: Loop variable accessed as [[variable]] in inner step
    variable: current_target

    # threads: Number of concurrent SSH executions
    threads: 5

    step:
      name: probe-target
      type: bash
      # Inner step inherits workflow-level SSH runner
      command: 'curl -s -o /dev/null -w "%{http_code}" "[[current_target]]" 2>/dev/null || echo "failed"'
      exports:
        probe_result: "{{stdout}}"

  # ===========================================================================
  # Step running on local host (override workflow's SSH runner)
  # Useful for local processing of results retrieved from remote
  # ===========================================================================
  - name: process-results-locally
    type: bash
    log: "Processing results on local host"

    # Override to run locally instead of via SSH
    step_runner: host

    command: 'cat "{{Output}}/nuclei-results.json" | jq -r ".info.severity" | sort | uniq -c'

    exports:
      severity_summary: "{{stdout}}"

  # ===========================================================================
  # Function step (always runs locally, regardless of workflow runner)
  # Note: Function steps execute on the host running osmedeus, not remote
  # ===========================================================================
  - name: log-completion
    type: function
    log: "Logging scan completion"
    function: 'log_info("SSH scan completed for {{Target}}")'

  # ===========================================================================
  # Cleanup step on remote server
  # ===========================================================================
  - name: cleanup-remote
    type: bash
    log: "Cleaning up remote workspace"
    command: 'rm -rf /home/scanner/workspace/temp/* 2>/dev/null; echo "Cleanup complete"'

    on_success:
      - action: log
        message: "Remote cleanup completed successfully"

    on_error:
      - action: continue
        message: "Cleanup failed but continuing workflow"
`,
  "all-step-types-example": `# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: all-step-types-example

# description: Human-readable description of what this workflow does
description: Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  # name: Parameter identifier used in templates as {{param_name}}
  # default: Default value if not provided via CLI
  # required: If true, workflow fails without this value
  # generator: Function to generate value, e.g., uuid(), currentDate(), getEnvVar("KEY")
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"  # Can reference built-in variables
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()  # Generates a unique ID automatically

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  # commands: List of binaries/commands that must exist in PATH
  commands:
    - echo
    - curl

  # files: List of files/directories that must exist
  files:
    - /tmp

  # variables: Define variable requirements with type validation
  # Types: domain, path, number, file, string
  variables:
    - name: Target
      type: string
      required: true

  # functions_conditions: JavaScript expressions that must evaluate to true
  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  # name: Display name for the report
  # path: File path (can use templates like {{Output}})
  # type: Format type - text, csv, json, markdown, etc.
  # description: Human-readable description
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  # skip_workspace: Equivalent to --disable-workspace-creation
  skip_workspace: false

  # disable_notifications: Equivalent to --disable-notification
  disable_notifications: true

  # disable_logging: Equivalent to --disable-logging
  disable_logging: false

  # heuristics_check: Equivalent to --heuristics-check (none, basic, advanced)
  heuristics_check: 'basic'

  # ci_output_format: Equivalent to --ci-output-format
  ci_output_format: false

  # silent: Equivalent to --silent
  silent: false

  # repeat: Equivalent to --repeat
  repeat: false

  # repeat_wait_time: Equivalent to --repeat-wait-time (e.g., 30s, 1h, 2h30m)
  repeat_wait_time: '60s'

  # clean_up_workspace: Equivalent to --clean-up-workspace
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  # ===========================================================================
  # STEP TYPE: bash
  # Execute shell commands on the host (or configured runner)
  # ===========================================================================
  - name: bash-single-command
    # type: Step type - bash, function, parallel-steps, foreach, remote-bash, http, llm
    type: bash

    # pre_condition: JavaScript expression - step only runs if this evaluates to true
    pre_condition: 'true'

    # log: Custom log message displayed when step starts (supports templates)
    log: "Executing single bash command for {{Target}}"

    # timeout: Maximum execution time in seconds (0 = no timeout)
    timeout: 60

    # command: Single command to execute
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'

    # std_file: File path to save stdout/stderr output
    std_file: "{{Output}}/step1-output.txt"

    # exports: Variables to export for subsequent steps
    # Key = variable name, Value = extraction pattern or literal value
    exports:
      step1_result: "completed"

  # ---------------------------------------------------------------------------
  # Bash step with multiple sequential commands
  # ---------------------------------------------------------------------------
  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"

    # commands: List of commands executed sequentially
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  # ---------------------------------------------------------------------------
  # Bash step with parallel commands
  # ---------------------------------------------------------------------------
  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"

    # parallel_commands: List of commands executed concurrently
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  # ---------------------------------------------------------------------------
  # Bash step with structured arguments
  # Arguments are joined in order: command + speed + config + input + output
  # ---------------------------------------------------------------------------
  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"

    command: 'echo'

    # speed_args: Performance-related arguments (e.g., thread count, rate limits)
    speed_args: '-n'

    # config_args: Configuration arguments (e.g., config file paths)
    config_args: ''

    # input_args: Input-related arguments (e.g., input file, target)
    input_args: '"Structured arguments test"'

    # output_args: Output-related arguments (e.g., output file, format)
    output_args: ''

  # ===========================================================================
  # STEP TYPE: function
  # Execute built-in utility functions via Otto JavaScript runtime
  # ===========================================================================
  - name: function-single
    type: function
    log: "Executing single function"

    # function: Single function call (JavaScript expression)
    function: 'log_info("Processing {{Target}} in function step")'

  # ---------------------------------------------------------------------------
  # Function step with multiple sequential functions
  # ---------------------------------------------------------------------------
  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"

    # functions: List of functions executed sequentially
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  # ---------------------------------------------------------------------------
  # Function step with parallel functions
  # ---------------------------------------------------------------------------
  - name: function-parallel
    type: function
    log: "Executing functions in parallel"

    # parallel_functions: List of functions executed concurrently
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  # ===========================================================================
  # STEP TYPE: parallel-steps
  # Execute multiple complete steps in parallel
  # ===========================================================================
  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"

    # parallel_steps: List of Step objects executed concurrently
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'

      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'

      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  # ===========================================================================
  # STEP TYPE: foreach
  # Iterate over input lines, executing inner step for each
  # ===========================================================================
  - name: foreach-example
    type: foreach
    log: "Iterating over items"

    # input: File path or direct content to iterate over (one item per line)
    input: "{{Output}}/items.txt"

    # variable: Name for the loop variable, accessed as [[variable]] in inner step
    variable: item

    # threads: Number of concurrent iterations (default: 1 = sequential)
    threads: 5

    # step: The inner step to execute for each item (single Step object)
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  # ===========================================================================
  # STEP TYPE: http
  # Make HTTP requests to external APIs
  # ===========================================================================
  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30

    # url: Target URL for the request (required for http type)
    url: "https://httpbin.org/post"

    # method: HTTP method - GET, POST, PUT, DELETE, PATCH, etc.
    method: POST

    # headers: Map of HTTP headers to send
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value

    # request_body: Request body content (typically JSON for POST/PUT)
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }

    exports:
      http_response: "{{response.body}}"

  # ===========================================================================
  # STEP TYPE: llm
  # Make LLM API calls for AI-powered processing
  # ===========================================================================
  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120

    # messages: Conversation messages for chat completion
    # role: system, user, assistant, or tool
    # content: Message text (can be string or multimodal array)
    messages:
      - role: system
        content: "You are a security analysis assistant."

      - role: user
        # content can be a simple string or complex multimodal content
        content: "Analyze this target: {{Target}}"

    # tools: Function tools available to the LLM
    tools:
      - type: function  # Currently only "function" type supported
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          # parameters: JSON Schema defining function parameters
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target

    # tool_choice: How the model should choose tools
    # Can be: "auto", "none", "required", or {"type": "function", "function": {"name": "fn_name"}}
    tool_choice: auto

    # llm_config: Step-level LLM configuration overrides
    llm_config:
      # provider: Specific provider to use (overrides rotation)
      provider: openai

      # model: Model override for this step
      model: gpt-4

      # Generation parameters
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0

      # Request settings
      timeout: "60s"
      max_retries: 3
      stream: false

      # response_format: Control output format
      # type: "text", "json_object", or "json_schema"
      response_format:
        type: json_object

    # extra_llm_parameters: Additional provider-specific parameters
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0

    exports:
      llm_analysis: "{{response.content}}"

  # ---------------------------------------------------------------------------
  # LLM step for embeddings
  # ---------------------------------------------------------------------------
  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"

    # is_embedding: Flag to indicate this is an embedding request
    is_embedding: true

    # embedding_input: List of texts to generate embeddings for
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"

    llm_config:
      model: text-embedding-3-small

    exports:
      embeddings: "{{response.embeddings}}"

  # ===========================================================================
  # COMMON STEP FIELDS: on_success, on_error, decision
  # These fields are available on ALL step types
  # ===========================================================================
  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'

    # on_success: Actions to execute when step succeeds
    on_success:
      # action: Handler type - log, abort, continue, export, run, notify
      - action: log
        message: "Step completed successfully for {{Target}}"

      - action: export
        # name: Variable name to export
        name: success_flag
        # value: Value to export (can be string, number, or template)
        value: "true"

      - action: notify
        # notify: Notification message
        notify: "Step succeeded for {{Target}}"

      - action: run
        # type: Step type to run (bash or function)
        type: bash
        command: 'echo "Running follow-up command"'

      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'

    # on_error: Actions to execute when step fails
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        # condition: Only execute this action if condition evaluates to true
        condition: 'true'

      - action: notify
        notify: "Error in workflow for {{Target}}"

      # abort: Stops workflow execution immediately
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'  # Only abort under specific conditions

      # continue: Allows workflow to continue despite error
      - action: continue
        message: "Continuing despite error"

    # decision: Conditional routing to other steps or workflow end
    decision:
      # condition: JavaScript expression to evaluate
      # next: Step name to jump to, or "_end" to finish workflow
      - condition: '{{success_flag}} == "true"'
        next: final-step

      - condition: '{{success_flag}} != "true"'
        next: _end  # Special value to end workflow

  # ---------------------------------------------------------------------------
  # Final step
  # ---------------------------------------------------------------------------
  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`,
  "mock-all-step-types-example": `# =============================================================================
# Module Workflow: All Step Types Example
# =============================================================================
# This file demonstrates ALL fields available in a module-kind workflow,
# showcasing every step type with comprehensive comments.
# =============================================================================

# -----------------------------------------------------------------------------
# WORKFLOW-LEVEL FIELDS
# -----------------------------------------------------------------------------

# kind: Workflow type - either "module" (single unit with steps) or "flow" (orchestrates modules)
kind: module

# name: Unique identifier for this workflow (required)
name: mock-all-step-types-example

# description: Human-readable description of what this workflow does
description: Mock Demonstrates all step types and their fields with detailed comments

# tags: Comma-separated tags for filtering and categorization (parsed as []string)
tags: example, comprehensive, demo

# -----------------------------------------------------------------------------
# PARAMS SECTION
# Define workflow parameters that can be passed via CLI or referenced in templates
# -----------------------------------------------------------------------------
params:
  - name: message
    default: "Hello World"
    required: false

  - name: output_dir
    default: "{{Output}}/results"
    required: false

  - name: threads
    default: "10"
    required: false

  - name: run_id
    generator: uuid()

# -----------------------------------------------------------------------------
# DEPENDENCIES SECTION
# Validate requirements before workflow execution
# -----------------------------------------------------------------------------
dependencies:
  commands:
    - echo
    - curl

  files:
    - /tmp

  variables:
    - name: Target
      type: string
      required: true

  functions_conditions:
    - '1 + 1 == 2'

# -----------------------------------------------------------------------------
# REPORTS SECTION
# Define output files produced by this workflow
# -----------------------------------------------------------------------------
reports:
  - name: main-output
    path: "{{Output}}/main-results.txt"
    type: text
    description: Main output file from the workflow

  - name: json-results
    path: "{{Output}}/results.json"
    type: json
    description: Structured JSON output

# -----------------------------------------------------------------------------
# PREFERENCES SECTION (Optional)
# Set CLI-like flags directly in the workflow. CLI flags always take precedence.
# -----------------------------------------------------------------------------
preferences:
  skip_workspace: false
  disable_notifications: true
  disable_logging: false
  heuristics_check: 'basic'
  ci_output_format: false
  silent: false
  repeat: false
  repeat_wait_time: '60s'
  clean_up_workspace: false

# -----------------------------------------------------------------------------
# STEPS SECTION
# The ordered list of execution steps for this module
# -----------------------------------------------------------------------------
steps:
  - name: bash-single-command
    type: bash
    pre_condition: 'true'
    log: "Executing single bash command for {{Target}}"
    timeout: 60
    command: 'echo "Processing target: {{Target}} with message: {{message}}"'
    std_file: "{{Output}}/step1-output.txt"
    exports:
      step1_result: "completed"

  - name: bash-multiple-commands
    type: bash
    log: "Running multiple sequential commands"
    commands:
      - 'echo "First command"'
      - 'echo "Second command"'
      - 'echo "Third command"'

  - name: bash-parallel-commands
    type: bash
    log: "Running commands in parallel"
    parallel_commands:
      - 'echo "Parallel A" && sleep 1'
      - 'echo "Parallel B" && sleep 1'
      - 'echo "Parallel C" && sleep 1'

  - name: bash-structured-args
    type: bash
    log: "Using structured argument fields"
    command: 'echo'
    speed_args: '-n'
    config_args: ''
    input_args: '"Structured arguments test"'
    output_args: ''

  - name: function-single
    type: function
    log: "Executing single function"
    function: 'log_info("Processing {{Target}} in function step")'

  - name: function-multiple
    type: function
    log: "Executing multiple functions sequentially"
    functions:
      - 'log_info("Function 1")'
      - 'log_info("Function 2")'
      - 'log_info("Function 3")'

  - name: function-parallel
    type: function
    log: "Executing functions in parallel"
    parallel_functions:
      - 'log_info("Parallel Function A")'
      - 'log_info("Parallel Function B")'
      - 'log_info("Parallel Function C")'

  - name: parallel-step-container
    type: parallel-steps
    log: "Running multiple steps in parallel"
    parallel_steps:
      - name: parallel-inner-1
        type: bash
        command: 'echo "Inner parallel step 1"'
      - name: parallel-inner-2
        type: function
        function: 'log_info("Inner parallel step 2")'
      - name: parallel-inner-3
        type: bash
        command: 'echo "Inner parallel step 3"'

  - name: foreach-example
    type: foreach
    log: "Iterating over items"
    input: "{{Output}}/items.txt"
    variable: item
    threads: 5
    step:
      name: process-item
      type: bash
      command: 'echo "Processing [[item]]"'
      exports:
        processed_item: "[[item]]"

  - name: http-request
    type: http
    log: "Making HTTP request"
    timeout: 30
    url: "https://httpbin.org/post"
    method: POST
    headers:
      Content-Type: application/json
      Authorization: "Bearer {{api_token}}"
      X-Custom-Header: custom-value
    request_body: |
      {
        "target": "{{Target}}",
        "message": "{{message}}"
      }
    exports:
      http_response: "{{response.body}}"

  - name: llm-chat-completion
    type: llm
    log: "Calling LLM for analysis"
    timeout: 120
    messages:
      - role: system
        content: "You are a security analysis assistant."
      - role: user
        content: "Analyze this target: {{Target}}"
    tools:
      - type: function
        function:
          name: analyze_target
          description: Analyzes a target for security vulnerabilities
          parameters:
            type: object
            properties:
              target:
                type: string
                description: The target to analyze
              depth:
                type: string
                enum: [shallow, deep]
            required:
              - target
    tool_choice: auto
    llm_config:
      provider: openai
      model: gpt-4
      max_tokens: 1000
      temperature: 0.7
      top_p: 1.0
      timeout: "60s"
      max_retries: 3
      stream: false
      response_format:
        type: json_object
    extra_llm_parameters:
      seed: 42
      presence_penalty: 0.0
    exports:
      llm_analysis: "{{response.content}}"

  - name: llm-embedding
    type: llm
    log: "Generating text embeddings"
    is_embedding: true
    embedding_input:
      - "Security vulnerability in {{Target}}"
      - "Network reconnaissance results"
      - "Port scan findings"
    llm_config:
      model: text-embedding-3-small
    exports:
      embeddings: "{{response.embeddings}}"

  - name: step-with-handlers
    type: bash
    log: "Step demonstrating success/error handlers and decision routing"
    command: 'echo "Running step with all handler types"'
    on_success:
      - action: log
        message: "Step completed successfully for {{Target}}"
      - action: export
        name: success_flag
        value: "true"
      - action: notify
        notify: "Step succeeded for {{Target}}"
      - action: run
        type: bash
        command: 'echo "Running follow-up command"'
      - action: run
        type: function
        functions:
          - 'log_info("Running follow-up function")'
    on_error:
      - action: log
        message: "Step failed for {{Target}}"
        condition: 'true'
      - action: notify
        notify: "Error in workflow for {{Target}}"
      - action: abort
        message: "Aborting due to critical failure"
        condition: 'false'
      - action: continue
        message: "Continuing despite error"
    decision:
      - condition: '{{success_flag}} == "true"'
        next: final-step
      - condition: '{{success_flag}} != "true"'
        next: _end

  - name: final-step
    type: function
    log: "Final step - workflow complete"
    function: 'log_info("All step types demonstrated for {{Target}}")'
`,
};

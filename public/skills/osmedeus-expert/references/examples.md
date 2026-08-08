# Complete Workflow Examples

## Example 1: Subdomain Enumeration Module

A module that discovers subdomains using multiple tools in parallel, merges results, and resolves live hosts.

```yaml
name: subdomain-enumeration
kind: module
desc: Comprehensive subdomain enumeration

params:
  - name: threads
    value: "10"
  - name: resolvers
    value: "{{Data}}/resolvers.txt"
  - name: wordlist
    value: "{{Data}}/subdomains-top1million-5000.txt"

steps:
  - name: setup-directories
    type: bash
    commands:
      - mkdir -p {{Output}}/subdomains
      - mkdir -p {{Output}}/resolved
    exports:
      subdomain_dir: "{{Output}}/subdomains"
      resolved_dir: "{{Output}}/resolved"

  - name: validate-target
    type: function
    script: |
      log_info("Starting enumeration for: {{Target}}");
      if (is_empty("{{Target}}")) {
        log_error("Target is empty");
        return false;
      }
      return true;

  - name: passive-enumeration
    type: parallel-steps
    parallel_steps:
      - name: subfinder-scan
        type: bash
        command: "{{Binaries}}/subfinder -d {{Target}} -silent -o {{subdomain_dir}}/subfinder.txt"
        timeout: 600
      - name: amass-passive
        type: bash
        command: "{{Binaries}}/amass enum -passive -d {{Target}} -o {{subdomain_dir}}/amass.txt"
        timeout: 900
      - name: assetfinder-scan
        type: bash
        command: "{{Binaries}}/assetfinder --subs-only {{Target}} > {{subdomain_dir}}/assetfinder.txt"
        timeout: 300

  - name: merge-results
    type: bash
    command: "cat {{subdomain_dir}}/*.txt | sort -u > {{subdomain_dir}}/all-subdomains.txt"
    exports:
      all_subdomains: "{{subdomain_dir}}/all-subdomains.txt"

  - name: check-results
    type: function
    script: |
      var count = file_length("{{all_subdomains}}");
      log_info("Found " + count + " unique subdomains");
      return count;
    exports:
      subdomain_count: "{{Result}}"

  - name: active-bruteforce
    type: bash
    pre_condition: "file_length('{{all_subdomains}}') < 50"
    command: "{{Binaries}}/puredns bruteforce {{wordlist}} {{Target}} -r {{resolvers}} -w {{subdomain_dir}}/bruteforce.txt"
    timeout: 1800
    on_error: continue

  - name: resolve-subdomains
    type: foreach
    input: "{{all_subdomains}}"
    variable: subdomain
    threads: "{{threads}}"
    step:
      name: resolve-single
      type: bash
      command: "echo [[subdomain]] | {{Binaries}}/dnsx -silent -a -resp -o {{resolved_dir}}/[[subdomain]].txt"
      timeout: 30
      on_error: continue

  - name: final-aggregation
    type: bash
    parallel_commands:
      - "cat {{resolved_dir}}/*.txt 2>/dev/null | sort -u > {{Output}}/resolved-subdomains.txt"
      - "wc -l {{subdomain_dir}}/all-subdomains.txt > {{Output}}/stats.txt"
    exports:
      final_subdomains: "{{Output}}/resolved-subdomains.txt"
```

**Run it:** `osmedeus run -m subdomain-enumeration -t example.com`

---

## Example 2: Reconnaissance Flow

A flow orchestrating multiple modules with dependencies.

```yaml
name: basic-recon-flow
kind: flow
desc: Recon flow with subdomain enum, port scan, and screenshots

params:
  - name: threads
    value: "20"

modules:
  - name: subdomain-enum
    steps:
      - name: passive-enum
        type: bash
        parallel_commands:
          - "{{Binaries}}/subfinder -d {{Target}} -silent -o {{Output}}/subfinder.txt"
          - "{{Binaries}}/assetfinder --subs-only {{Target}} > {{Output}}/assetfinder.txt"
        timeout: 600
      - name: merge
        type: bash
        command: "cat {{Output}}/*.txt | sort -u > {{Output}}/all-subdomains.txt"
        exports:
          all_subdomains: "{{Output}}/all-subdomains.txt"

  - name: port-scan
    depends_on: [subdomain-enum]
    condition: "file_length('{{all_subdomains}}') > 0"
    steps:
      - name: scan-ports
        type: foreach
        input: "{{all_subdomains}}"
        variable: subdomain
        threads: "{{threads}}"
        step:
          name: naabu-scan
          type: bash
          command: "{{Binaries}}/naabu -host [[subdomain]] -top-ports 100 -silent >> {{Output}}/open-ports.txt"
          timeout: 120
          on_error: continue
      - name: aggregate
        type: bash
        command: "sort -u {{Output}}/open-ports.txt -o {{Output}}/all-ports.txt"
        exports:
          all_ports: "{{Output}}/all-ports.txt"

  - name: screenshot
    depends_on: [port-scan]
    condition: "file_length('{{all_ports}}') > 0"
    steps:
      - name: probe-http
        type: bash
        command: "{{Binaries}}/httpx -l {{all_subdomains}} -silent -o {{Output}}/http-hosts.txt"
        exports:
          http_hosts: "{{Output}}/http-hosts.txt"
      - name: capture
        type: foreach
        input: "{{http_hosts}}"
        variable: url
        threads: 10
        step:
          name: gowitness
          type: bash
          command: "{{Binaries}}/gowitness single --url=[[url]] --screenshot-path={{Output}}/screenshots"
          timeout: 60
          on_error: continue
```

**Run it:** `osmedeus run -f basic-recon-flow -t example.com`

---

## Example 3: Vulnerability Assessment Flow

Flow combining discovery, scanning, and reporting.

```yaml
name: vulnerability-flow
kind: flow

params:
  - name: threads
    value: "25"
  - name: severity
    value: "medium,high,critical"
  - name: templates
    value: "{{Data}}/nuclei-templates"

modules:
  - name: discovery
    steps:
      - name: find-endpoints
        type: bash
        parallel_commands:
          - "{{Binaries}}/waybackurls {{Target}} > {{Output}}/wayback.txt"
          - "{{Binaries}}/gau {{Target}} > {{Output}}/gau.txt"
          - "{{Binaries}}/katana -u {{Target}} -silent -o {{Output}}/katana.txt"
        timeout: 900
        on_error: continue
      - name: merge
        type: bash
        command: "cat {{Output}}/*.txt | sort -u > {{Output}}/all-endpoints.txt"
        exports:
          all_endpoints: "{{Output}}/all-endpoints.txt"

  - name: scanning
    depends_on: [discovery]
    condition: "file_length('{{all_endpoints}}') > 0"
    steps:
      - name: nuclei-scan
        type: bash
        command: "{{Binaries}}/nuclei -l {{all_endpoints}} -t {{templates}} -severity {{severity}} -c {{threads}} -o {{Output}}/nuclei.json -jsonl"
        timeout: 7200
        on_error: continue
        exports:
          nuclei_results: "{{Output}}/nuclei.json"
      - name: xss-sqli
        type: parallel-steps
        parallel_steps:
          - name: xss
            type: bash
            command: "cat {{all_endpoints}} | {{Binaries}}/dalfox pipe -o {{Output}}/xss.txt"
            timeout: 3600
            on_error: continue
          - name: sqli
            type: bash
            command: "{{Binaries}}/sqlmap -m {{all_endpoints}} --batch --output-dir={{Output}}/sqli"
            timeout: 3600
            on_error: continue

  - name: reporting
    depends_on: [scanning]
    steps:
      - name: aggregate
        type: function
        script: |
          var count = 0;
          if (file_exists("{{nuclei_results}}")) {
            count = file_length("{{nuclei_results}}");
          }
          log_info("Total findings: " + count);
          return count;
        exports:
          finding_count: "{{Result}}"
      - name: report
        type: bash
        command: |
          cat > {{Output}}/report.md << EOF
          # Vulnerability Report
          **Target:** {{Target}}
          **Findings:** {{finding_count}}
          **Severity Filter:** {{severity}}
          EOF
```

**Run it:** `osmedeus run -f vulnerability-flow -t example.com -p severity=high,critical`

---

## Example 4: Flow with Decision Routing

```yaml
name: adaptive-scan
kind: flow

params:
  - name: scan_depth
    default: "standard"

modules:
  - name: recon
    steps:
      - name: quick-recon
        type: bash
        command: "subfinder -d {{Target}} -silent | wc -l"
        exports:
          asset_count: "100"
    decision:
      switch: "{{scan_depth}}"
      cases:
        "quick": {goto: quick-scan}
        "deep": {goto: deep-scan}
      default: {goto: standard-scan}

  - name: quick-scan
    depends_on: [recon]
    steps:
      - name: fast
        type: bash
        command: "naabu -host {{Target}} -top-ports 100"
    decision:
      switch: "true"
      cases:
        "true": {goto: _end}

  - name: standard-scan
    depends_on: [recon]
    steps:
      - name: standard
        type: bash
        command: "naabu -host {{Target}} -top-ports 1000"

  - name: deep-scan
    depends_on: [recon]
    steps:
      - name: full
        type: bash
        command: "nmap -p- {{Target}}"
```

**Run it:** `osmedeus run -f adaptive-scan -t example.com -p scan_depth=deep`

---

## Example 5: Agent-Powered Security Analysis

```yaml
name: agent-analysis
kind: module

steps:
  - name: collect-data
    type: bash
    commands:
      - "subfinder -d {{Target}} -silent -o {{Output}}/subs.txt"
      - "httpx -l {{Output}}/subs.txt -silent -o {{Output}}/live.txt"
    exports:
      live_hosts: "{{Output}}/live.txt"

  - name: analyze
    type: agent
    query: |
      Analyze the live hosts discovered for {{Target}}.
      Read {{live_hosts}} and for each host:
      1. Check HTTP headers for security issues
      2. Identify interesting technologies
      3. Summarize findings with severity ratings
    system_prompt: "You are an expert security analyst. Be thorough but concise."
    max_iterations: 15
    agent_tools:
      - preset: bash
      - preset: read_file
      - preset: save_content
      - preset: http_get
      - preset: grep_regex
    memory:
      max_messages: 40
      persist_path: "{{Output}}/agent/conversation.json"
    exports:
      analysis: "{{agent_content}}"

  - name: save-report
    type: bash
    command: echo "{{analysis}}" > {{Output}}/analysis-report.md
```

**Run it:** `osmedeus run -m agent-analysis -t example.com`

---

## Example 6: Workflow with Inheritance

### Base Module

```yaml
kind: module
name: base-scan
params:
  - name: threads
    default: "10"
  - name: timeout
    default: "3600"
steps:
  - name: setup
    type: bash
    command: mkdir -p {{Output}}
  - name: scan
    type: bash
    command: "nmap -sV -T4 --top-ports 1000 {{Target}}"
    timeout: "{{timeout}}"
  - name: report
    type: bash
    command: echo "Scan complete"
```

### Fast Child

```yaml
kind: module
name: fast-scan
extends: base-scan
override:
  params:
    threads: "50"
    timeout: "600"
```

### Aggressive Child (Replace Steps)

```yaml
kind: module
name: aggressive-scan
extends: base-scan
override:
  params:
    threads: "100"
  steps:
    mode: append
    add:
      - name: vuln-scan
        type: bash
        command: "nuclei -u {{Target}} -severity critical,high"
    replace:
      scan:
        name: scan
        type: bash
        command: "nmap -sV -T5 -p- {{Target}}"
        timeout: 7200
```

---

## Example 7: Event-Triggered Workflow

```yaml
name: auto-scan-new-assets
kind: flow

triggers:
  - name: on-new-subdomain
    on: event
    event:
      topic: assets.new
      filters:
        - 'event.data_type == "subdomain"'
    input:
      target: event_data.hostname
    enabled: true

  - name: nightly-rescan
    on: cron
    schedule: "0 2 * * *"
    input:
      type: file
      path: "{{Data}}/targets.txt"
    enabled: true

  - name: manual
    on: manual
    enabled: true

modules:
  - name: scan
    steps:
      - name: probe
        type: bash
        command: "httpx -u {{Target}} -silent -status-code -o {{Output}}/probe.txt"
      - name: scan
        type: bash
        command: "nuclei -u {{Target}} -severity high,critical -o {{Output}}/vulns.txt"
        on_error: continue
```

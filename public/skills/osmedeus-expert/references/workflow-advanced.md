# Advanced Workflow Features

## Triggers

Triggers define when a workflow executes automatically. Multiple triggers per workflow are supported.

### Cron Trigger

```yaml
triggers:
  - name: daily-scan
    on: cron
    schedule: "0 2 * * *"    # minute hour day month weekday
    enabled: true
    input:
      type: file
      path: "/data/targets.txt"
```

### Event Trigger

```yaml
triggers:
  - name: on-new-asset
    on: event
    event:
      topic: assets.new       # webhook.received, assets.changed, db.change, etc.
      filters:                # JS expressions, all must be true
        - 'event.source == "subfinder"'
        - 'event.data_type == "subdomain"'
      dedupe_key: "{{event.data.hostname}}"
      dedupe_window: "5m"
    input:
      # New exports-style (multiple variables):
      target: event_data.hostname
      source: event.source
      # OR legacy syntax (single variable):
      # type: event_data
      # field: hostname
    enabled: true
```

### Watch Trigger

```yaml
triggers:
  - name: file-watch
    on: watch
    path: "/data/targets/new.txt"
    debounce: "500ms"
    input:
      type: file
      path: "/data/targets/new.txt"
    enabled: true
```

### Manual Trigger

```yaml
triggers:
  - name: manual
    on: manual
    enabled: true    # false = block CLI execution
    input:
      type: param
      name: Target
```

## Workflow Inheritance

Child workflows extend parent workflows:

### Parent

```yaml
kind: module
name: base-scan
params:
  - name: threads
    default: "10"
  - name: timeout
    default: "3600"
steps:
  - name: step-one
    type: bash
    command: echo "Scanning {{Target}}"
  - name: step-two
    type: bash
    command: echo "Threads: {{threads}}"
```

### Child (Override Params Only)

```yaml
kind: module
name: fast-scan
extends: base-scan
override:
  params:
    threads: "5"
    timeout: "1800"
```

### Child (Modify Steps)

```yaml
extends: base-scan
override:
  steps:
    mode: append       # append (default), prepend, merge
    add:
      - name: extra-step
        type: bash
        command: echo "Added step"
    replace:
      step-one:
        name: step-one
        type: bash
        command: echo "Replaced step one"
    remove:
      - step-two
```

### Flow Child (Override Modules)

```yaml
extends: base-flow
override:
  modules:
    mode: merge
    add:
      - name: extra-module
        path: extra.yaml
    remove:
      - unused-module
  triggers:             # Replaces parent triggers entirely
    - name: new-trigger
      on: cron
      schedule: "0 0 * * *"
  dependencies:         # Merged with parent
    commands: [extra-tool]
  runner: docker        # Override parent runner
  runner_config:
    image: "custom:latest"
```

## Runners

### Host Runner (Default)

No configuration needed. Commands run on the local machine.

### Docker Runner

```yaml
# Module-level default runner
runner: docker
runner_config:
  image: "ubuntu:22.04"
  env:
    API_KEY: "{{api_key}}"
  volumes:
    - "{{Output}}:/output"
    - "{{Data}}:/data:ro"
  network: host
  persistent: true       # Reuse container across steps
  workdir: "/workspace"
```

### SSH Runner

```yaml
runner: ssh
runner_config:
  host: "scanner.example.com"
  port: 22
  user: "scanner"
  key_file: "~/.ssh/scanner_key"
  workdir: "/tmp/osmedeus"
```

### Per-Step Runner Override

```yaml
steps:
  - name: local-step
    type: bash
    command: echo "Runs locally"

  - name: docker-step
    type: remote-bash
    step_runner: docker
    step_runner_config:
      image: "nmap:latest"
    command: "nmap {{Target}}"
    step_remote_file: "/tmp/results.txt"
    host_output_file: "{{Output}}/nmap.txt"
```

## Dependencies Section

Validate requirements before execution:

```yaml
dependencies:
  commands: [nmap, subfinder, nuclei]  # Must be in PATH
  files: [/usr/local/bin/tool]         # Must exist
  variables:
    - name: Target
      type: domain       # domain, subdomain, url, ip, cidr, repo, path, file, string, number
      required: true
  target_types: [domain, subdomain]    # Allowed target types
  functions_conditions:
    - 'file_exists("/data/config.yaml")'
```

## Reports Section

Declare output files:

```yaml
reports:
  - name: scan-results
    path: "{{Output}}/results.json"
    type: json
    description: "Scan results in JSON format"
    optional: false
  - name: summary
    path: "{{Output}}/summary.md"
    type: markdown
    description: "Human-readable summary"
    optional: true
```

## Preferences Section

Set CLI-equivalent defaults (CLI flags override these):

```yaml
preferences:
  disable_notifications: true
  disable_logging: false
  heuristics_check: "basic"     # none, basic, advanced
  ci_output_format: false
  silent: false
  repeat: false
  repeat_wait_time: "60s"
  empty_target: false
```

## Hooks

Pre/post scan steps:

```yaml
hooks:
  pre_scan_steps:
    - name: pre-check
      type: bash
      command: echo "Starting scan at $(date)"
  post_scan_steps:
    - name: cleanup
      type: bash
      command: echo "Scan finished at $(date)"
```

## Decision Routing

Conditional branching using switch/case:

```yaml
steps:
  - name: classify
    type: bash
    command: |
      if [ $(wc -l < {{Output}}/vulns.txt) -gt 100 ]; then
        echo "high"
      else
        echo "low"
      fi
    exports:
      risk_level: "high"

  - name: route
    type: bash
    command: echo "Routing based on risk"
    decision:
      switch: "{{risk_level}}"
      cases:
        "high": {goto: deep-analysis}
        "low": {goto: summary-only}
      default: {goto: summary-only}

  - name: deep-analysis
    type: bash
    command: echo "Running deep analysis"
    decision:
      switch: "true"
      cases:
        "true": {goto: _end}    # _end terminates workflow

  - name: summary-only
    type: bash
    command: echo "Quick summary"
```

Decisions also work on flow modules:

```yaml
modules:
  - name: recon
    steps: [...]
    decision:
      switch: "{{scan_depth}}"
      cases:
        "quick": {goto: fast-scan}
        "deep": {goto: full-scan}
      default: {goto: standard-scan}
```

## Condition-Based Decision Routing

Instead of switch/case (exact string matching), use `conditions` for JS boolean expressions. All matching conditions execute (no short-circuit).

```yaml
steps:
  - name: analyze-results
    type: bash
    command: echo "Analyzing"
    decision:
      conditions:
        # Goto a specific step
        - if: "file_length('{{Output}}/vulns.txt') > 100"
          goto: deep-analysis

        # Execute inline commands
        - if: "{{enableNmap}} && contains('{{Port}}', '-')"
          commands:
            - "nmap -sV -p {{Port}} {{Target}}"

        # Execute inline functions
        - if: "file_length('{{inputFile}}') > 0"
          function: "log_info('file has content')"

        # Multiple functions
        - if: "is_empty('{{Output}}/results.txt')"
          functions:
            - "log_warn('no results found')"
            - "skip('empty results')"
```

Each condition supports:
- `if` - JS boolean expression (evaluated via Goja)
- `goto` - Jump to named step (or `_end` to terminate)
- `command` / `commands` - Execute shell command(s) inline
- `function` / `functions` - Execute JS function(s) inline

**Key differences from switch/case:**
- Switch/case does exact string matching; conditions support boolean logic
- All matching conditions execute (not just the first match)
- Conditions can execute commands/functions inline without needing a goto

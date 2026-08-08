---
name: osmedeus-expert
description: "Expert guide for the Osmedeus security automation workflow engine. Use when: (1) writing or editing YAML workflows (modules and flows), (2) running osmedeus CLI commands (scan, workflow management, installation, server), (3) configuring steps, runners, triggers, or template variables, (4) debugging workflow execution issues, (5) building security scanning pipelines, (6) working with agent/LLM step types, or (7) any question about osmedeus features, architecture, or best practices."
---

# Osmedeus Expert

Expert knowledge for writing YAML workflows and operating the Osmedeus security automation engine.

## Quick Orientation

Osmedeus executes YAML-defined workflows with two kinds:
- **module** - Single execution unit containing steps (the building block)
- **flow** - Orchestrates multiple modules with dependency ordering

Template variables use `{{Variable}}` syntax. Foreach loop variables use `[[variable]]` to avoid conflicts.

## Running Osmedeus

### Essential Commands

```bash
# Run a flow against a target
osmedeus run -f <flow-name> -t <target>

# Run a module
osmedeus run -m <module-name> -t <target>

# Run multiple modules in sequence
osmedeus run -m mod1 -m mod2 -t <target>

# Multiple targets from file with concurrency
osmedeus run -m <module> -T targets.txt -c 5

# With parameters
osmedeus run -m <module> -t <target> -p threads=20 -p depth=2
osmedeus run -m <module> -t <target> -P params.yaml

# With timeout and repeat
osmedeus run -m <module> -t <target> --timeout 2h
osmedeus run -m <module> -t <target> --repeat --repeat-wait-time 30m

# Dry run (show what would execute)
osmedeus run -m <module> -t <target> --dry-run

# Chunked processing for large target lists
osmedeus run -m <module> -T targets.txt --chunk-size 100 --chunk-part 0

# Distributed execution
osmedeus run -m <module> -t <target> --distributed-run
```

### Workflow Management

```bash
osmedeus workflow list                # List available workflows
osmedeus workflow show <name>         # Show workflow details
osmedeus workflow lint <workflow-path> # Validate workflow YAML
```

### Installation & Setup

```bash
osmedeus install base --preset              # Install base from preset repo
osmedeus install base --preset --keep-setting  # Install base, keep settings
osmedeus install workflow --preset          # Install workflows from preset
osmedeus install binary --all               # Install all tool binaries
osmedeus install binary --name <name>       # Install specific binary
osmedeus install binary --all --check       # Check binary status
osmedeus install env                        # Add binaries to PATH
osmedeus install validate --preset          # Validate installation
```

### Server & Workers

```bash
osmedeus server                       # Start REST API server
osmedeus server --master              # Start as distributed master
osmedeus worker join                  # Join as distributed worker
osmedeus worker join --get-public-ip  # Join with public IP detection
osmedeus worker status                # Show registered workers
osmedeus worker eval -e '<expr>'     # Evaluate function with distributed hooks
osmedeus worker set <id> <field> <value>  # Update worker metadata
osmedeus worker queue list            # List queued tasks
osmedeus worker queue new -f <flow> -t <target>  # Queue task
osmedeus worker queue run --concurrency 5        # Process queued tasks
```

### Cloud

```bash
osmedeus cloud config set <key> <value>   # Configure cloud provider
osmedeus cloud config list                # List cloud config
osmedeus cloud create --instances N       # Provision infrastructure
osmedeus cloud list                       # List active infrastructure
osmedeus cloud run -f <flow> -t <target> --instances N  # Run distributed
osmedeus cloud destroy <id>               # Destroy infrastructure
osmedeus cloud setup 1.2.3.4 5.6.7.8     # Setup existing machines
osmedeus cloud setup 1.2.3.4 --ansible   # Use Ansible playbook
```

### Other Commands

```bash
osmedeus func list                    # List utility functions
osmedeus func e 'log_info("test")'   # Evaluate a function
osmedeus snapshot export <workspace>  # Export workspace as ZIP
osmedeus snapshot import <source>     # Import workspace
osmedeus snapshot list                # List snapshots
osmedeus update                       # Self-update
osmedeus update --check               # Check for updates
osmedeus assets                       # List discovered assets
osmedeus assets -w <workspace>        # Filter by workspace
osmedeus assets --source httpx --type web  # Filter by source/type
osmedeus assets --stats               # Show asset statistics
osmedeus assets --columns url,title,status_code  # Custom columns
osmedeus assets --json                # JSON output
osmedeus query vulns                  # Query vulnerabilities
osmedeus query vulns --severity high -w example.com  # Filter vulns
osmedeus query runs --status running  # Query workflow runs
osmedeus query steps --run <run-uuid> # Query steps for a run
osmedeus uninstall                    # Uninstall osmedeus
osmedeus uninstall --clean            # Also remove workspaces data
```

For complete CLI flags, see [references/cli-flags.md](references/cli-flags.md).

## Writing Workflows

### Module Structure (Minimal)

```yaml
name: my-module
kind: module

params:
  - name: threads
    default: "10"

steps:
  - name: scan-target
    type: bash
    command: echo "Scanning {{Target}}"
    exports:
      result: "output.txt"
```

### Flow Structure (Minimal)

```yaml
name: my-flow
kind: flow

modules:
  - name: enumeration
    steps:
      - name: find-subdomains
        type: bash
        command: subfinder -d {{Target}} -o {{Output}}/subs.txt
        exports:
          subdomains: "{{Output}}/subs.txt"

  - name: scanning
    depends_on: [enumeration]
    condition: "file_length('{{subdomains}}') > 0"
    steps:
      - name: port-scan
        type: bash
        command: naabu -l {{subdomains}} -o {{Output}}/ports.txt
```

### Step Types

| Type | Purpose | Key Fields |
|------|---------|------------|
| `bash` | Shell commands | `command`, `commands`, `parallel_commands` |
| `function` | JS utility functions | `function`, `functions`, `parallel_functions` |
| `parallel-steps` | Run steps concurrently | `parallel_steps: [Step list]` |
| `foreach` | Iterate over items | `input`, `variable`, `threads`, `step` |
| `remote-bash` | Execute on docker/ssh runner | Same as bash + `step_runner_config` |
| `http` | HTTP requests | `url`, `method`, `headers`, `request_body` |
| `llm` | LLM API calls | `messages`, `tools`, `llm_config` |
| `agent` | Agentic LLM with tool loop | `query`, `agent_tools`, `max_iterations` |
| `agent-acp` | Delegate to external ACP agent | `agent`, `messages`, `acp_config` |

For complete field reference per step type, see [references/step-types.md](references/step-types.md).

### Common Step Fields (All Types)

```yaml
- name: step-name            # Required, unique identifier
  type: bash                 # Required
  pre_condition: "expr"      # JS expression, skip if false
  log: "Custom message"      # Log message (supports templates)
  timeout: 60                # Max seconds (or "1h", "30m")
  exports:                   # Variables for subsequent steps
    var_name: "value"
  on_success: [{action: log, message: "done"}]
  on_error: [{action: continue}]
  decision:                  # Conditional routing
    switch: "{{var}}"
    cases:
      "val1": {goto: step-a}
    default: {goto: _end}    # _end terminates workflow
  depends_on: [other-step]   # DAG dependencies
```

### Template Variables

**Built-in:** `{{Target}}`, `{{Output}}`, `{{Workspaces}}`, `{{RunUUID}}`, `{{WorkflowName}}`

**Platform:** `{{PlatformOS}}`, `{{PlatformArch}}`, `{{PlatformInDocker}}`, `{{PlatformInKubernetes}}`, `{{PlatformCloudProvider}}`

**Custom params** defined in `params:` are accessed as `{{param_name}}`.

**Foreach variables** use double brackets: `[[variable]]`.

For parameter generators and all variables, see [references/template-variables.md](references/template-variables.md).

### Workflow Inheritance

```yaml
extends: parent-workflow-name
override:
  params:
    threads: "5"
  steps:
    mode: append    # append | prepend | merge
    add: [{name: extra, type: bash, command: "..."}]
    remove: [step-to-remove]
```

For the complete inheritance system, see [references/workflow-advanced.md](references/workflow-advanced.md).

## Workflow Patterns

### Pattern: Parallel Tool Execution

```yaml
- name: parallel-enum
  type: parallel-steps
  parallel_steps:
    - name: subfinder
      type: bash
      command: subfinder -d {{Target}} -o {{Output}}/subfinder.txt
      timeout: 600
    - name: amass
      type: bash
      command: amass enum -passive -d {{Target}} -o {{Output}}/amass.txt
      timeout: 900
```

### Pattern: Foreach with Concurrency

```yaml
- name: scan-each-host
  type: foreach
  input: "{{hosts_file}}"
  variable: host
  threads: "{{threads}}"
  step:
    name: scan-host
    type: bash
    command: nmap -sV [[host]] -oX {{Output}}/nmap/[[host]].xml
    timeout: 120
    on_error: continue
```

### Pattern: Conditional Branching (Switch/Case)

```yaml
- name: check-depth
  type: bash
  command: echo "{{scan_depth}}"
  decision:
    switch: "{{scan_depth}}"
    cases:
      "quick": {goto: fast-scan}
      "deep": {goto: full-scan}
    default: {goto: standard-scan}
```

### Pattern: Conditional Branching (Conditions)

```yaml
- name: route-by-conditions
  type: bash
  command: echo "Evaluating conditions"
  decision:
    conditions:
      - if: "file_length('{{inputFile}}') > 100"
        goto: deep-analysis
      - if: "file_length('{{inputFile}}') > 0"
        function: "log_info('file has content')"
      - if: "{{enableNmap}}"
        commands:
          - "nmap -sV {{Target}}"
```

### Pattern: Agent-Powered Analysis

```yaml
- name: analyze-findings
  type: agent
  query: "Analyze vulnerabilities in {{Output}}/vulns.json and prioritize by severity"
  system_prompt: "You are a security analyst."
  max_iterations: 10
  agent_tools:
    - preset: bash
    - preset: read_file
    - preset: grep_regex
    - preset: save_content
  memory:
    max_messages: 30
    persist_path: "{{Output}}/agent/conversation.json"
  exports:
    analysis: "{{agent_content}}"
```

### Pattern: Flow with Module Dependencies

```yaml
modules:
  - name: recon
    steps: [...]

  - name: scanning
    depends_on: [recon]
    condition: "file_length('{{subdomains}}') > 0"
    steps: [...]

  - name: reporting
    depends_on: [scanning]
    steps: [...]
```

## Reference Files

- **[references/cli-flags.md](references/cli-flags.md)** - All CLI flags for every command
- **[references/step-types.md](references/step-types.md)** - Complete field reference for each step type
- **[references/template-variables.md](references/template-variables.md)** - All template variables, generators, and utility functions
- **[references/workflow-advanced.md](references/workflow-advanced.md)** - Triggers, inheritance, runners, action handlers, reports
- **[references/examples.md](references/examples.md)** - Full annotated workflow examples (module, flow, agent, triggers)

## Debugging Tips

1. **Validate YAML** before running: `osmedeus workflow lint <workflow-path>`
2. **Dry run** to see execution plan: `osmedeus run -m <module> -t test --dry-run`
3. **Verbose output**: `osmedeus run -m <module> -t <target> -v`
4. **Check exports**: each step's exports propagate to subsequent steps only
5. **Foreach uses `[[var]]`** not `{{var}}` - this is the most common mistake
6. **pre_condition** uses JS expressions: `file_length('path') > 0`, `is_empty('{{var}}')`
7. **on_error: continue** prevents a failing step from stopping the workflow

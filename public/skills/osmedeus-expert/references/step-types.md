# Step Types - Complete Field Reference

## Bash Step

Execute shell commands.

```yaml
- name: my-step
  type: bash

  # Single command
  command: "echo hello"

  # OR multiple sequential commands
  commands:
    - "echo first"
    - "echo second"

  # OR parallel commands
  parallel_commands:
    - "tool1 -t {{Target}}"
    - "tool2 -t {{Target}}"

  # Structured arguments (joined: command + speed + config + input + output)
  speed_args: "-t {{threads}}"
  config_args: "-c config.yaml"
  input_args: "-i {{Target}}"
  output_args: "-o {{Output}}/result.txt"

  # Save stdout/stderr to file
  std_file: "{{Output}}/step-output.txt"

  # Common fields (available on all step types)
  pre_condition: "file_exists('{{input_file}}')"
  log: "Scanning {{Target}}"
  timeout: 600
  exports:
    result_file: "{{Output}}/result.txt"
  on_success: [{action: log, message: "done"}]
  on_error: [{action: continue}]
```

## Function Step

Execute Goja JS utility functions.

```yaml
- name: my-func
  type: function

  # Single function
  function: 'log_info("Processing {{Target}}")'

  # OR multiple sequential functions
  functions:
    - 'log_info("Step 1")'
    - 'log_info("Step 2")'

  # OR parallel functions
  parallel_functions:
    - 'log_info("Parallel A")'
    - 'log_info("Parallel B")'

  # Script block (multiline JS)
  script: |
    var count = file_length("{{Output}}/results.txt");
    log_info("Found " + count + " items");
    if (count == 0) {
      log_warn("No results");
    }
    return count;

  exports:
    item_count: "{{Result}}"
```

## Parallel-Steps

Run multiple complete steps concurrently.

```yaml
- name: parallel-scans
  type: parallel-steps
  parallel_steps:
    - name: scan-a
      type: bash
      command: "tool-a {{Target}}"
      timeout: 300
    - name: scan-b
      type: bash
      command: "tool-b {{Target}}"
      timeout: 300
    - name: scan-c
      type: function
      function: 'log_info("Running C")'
```

## Foreach

Iterate over items from a file or content (one per line).

**Important:** Use `[[variable]]` (double brackets) for the loop variable, NOT `{{variable}}`.

```yaml
- name: scan-each
  type: foreach

  # File path or direct content (one item per line)
  input: "{{Output}}/subdomains.txt"

  # Loop variable name (accessed as [[variable]])
  variable: subdomain

  # JS to transform each line before storing
  variable_pre_process: "trim(line)"

  # Concurrent iterations (default: 1 = sequential)
  threads: "{{threads}}"

  # Single step to execute for each item
  step:
    name: scan-subdomain
    type: bash
    command: "nmap [[subdomain]] -oX {{Output}}/nmap/[[subdomain]].xml"
    timeout: 120
    on_error: continue
```

## Remote-Bash

Execute commands via Docker or SSH runner.

```yaml
- name: remote-scan
  type: remote-bash

  command: "nmap -sV {{Target}}"

  # Runner-specific configuration
  step_runner_config:
    # Docker config
    image: "instrumentisto/nmap"
    env:
      SCAN_TARGET: "{{Target}}"
    volumes:
      - "{{Output}}:/output"
    network: host

    # OR SSH config
    host: "scanner.example.com"
    port: 22
    user: "scanner"
    key_file: "~/.ssh/scanner_key"
    workdir: "/tmp/scan"

  # Copy file from remote after execution
  step_remote_file: "/tmp/scan/results.txt"
  host_output_file: "{{Output}}/remote-results.txt"
```

## HTTP

Make HTTP requests.

```yaml
- name: api-call
  type: http
  url: "https://api.example.com/scan"
  method: POST
  headers:
    Content-Type: application/json
    Authorization: "Bearer {{api_token}}"
  request_body: |
    {"target": "{{Target}}", "depth": "{{scan_depth}}"}
  timeout: 30
  exports:
    api_response: "{{response.body}}"
```

## LLM

Call LLM API for AI-powered processing.

```yaml
- name: analyze
  type: llm
  messages:
    - role: system
      content: "You are a security analyst."
    - role: user
      content: "Analyze: {{Target}}"
  tools:
    - type: function
      function:
        name: classify_vuln
        description: "Classify a vulnerability"
        parameters:
          type: object
          properties:
            severity: {type: string, enum: [low, medium, high, critical]}
          required: [severity]
  tool_choice: auto
  llm_config:
    provider: openai
    model: gpt-4
    max_tokens: 1000
    temperature: 0.7
    timeout: "60s"
    max_retries: 3
    response_format:
      type: json_object
  exports:
    analysis: "{{response.content}}"
```

### LLM Embedding Request

```yaml
- name: embed
  type: llm
  is_embedding: true
  embedding_input:
    - "Security finding for {{Target}}"
    - "Port scan results"
  llm_config:
    model: text-embedding-3-small
  exports:
    vectors: "{{response.embeddings}}"
```

## Agent

Agentic LLM execution with tool-calling loop.

```yaml
- name: security-agent
  type: agent

  # Task prompt (single or multi-goal)
  query: "Enumerate and analyze {{Target}}"
  # OR multiple sequential goals:
  # queries:
  #   - "Enumerate subdomains of {{Target}}"
  #   - "Scan discovered hosts for vulnerabilities"

  system_prompt: "You are a security reconnaissance agent."
  max_iterations: 10   # Required, > 0

  # Available tools
  agent_tools:
    # Preset tools (auto-schema)
    - preset: bash
    - preset: read_file
    - preset: save_content
    - preset: grep_regex
    - preset: http_get
    # Custom tool
    - name: classify
      description: "Classify a finding"
      parameters:
        type: object
        properties:
          severity: {type: string}
        required: [severity]
      handler: 'log_info("Classified: " + args.severity)'

  # Stop early if condition met
  stop_condition: 'iterations > 5 && has_results'

  # Parallel tool execution (default: true)
  parallel_tool_calls: true

  # Conversation memory
  memory:
    max_messages: 30
    summarize_on_truncate: true
    persist_path: "{{Output}}/agent/conversation.json"
    resume_path: "{{Output}}/agent/prior.json"

  # Model preferences
  models: ["gpt-4", "claude-3-opus"]
  llm_config:
    max_tokens: 4000
    temperature: 0.3

  # Structured output on final iteration
  output_schema: '{"type":"object","properties":{"findings":{"type":"array"}}}'

  # Planning stage
  plan_prompt: "Plan your approach for analyzing {{Target}}"
  plan_max_tokens: 2000

  # Tool call hooks (JS expressions)
  on_tool_start: 'log_info("Calling: " + tool_name)'
  on_tool_end: 'log_info("Result: " + tool_result.substring(0, 100))'

  # Sub-agents for delegation
  sub_agents:
    - name: port-scanner
      description: "Specialized port scanning agent"
      system_prompt: "You scan ports efficiently."
      max_iterations: 5
      agent_tools:
        - preset: bash
      memory:
        max_messages: 20

  max_agent_depth: 3  # Max nesting depth (default: 3)

  exports:
    findings: "{{agent_content}}"
    history: "{{agent_history}}"
    iterations: "{{agent_iterations}}"
    tokens_used: "{{agent_total_tokens}}"
```

### Agent Preset Tools

| Preset | Purpose |
|--------|---------|
| `bash` | Execute shell command |
| `read_file` | Read file contents |
| `read_lines` | Read file as array of lines |
| `file_exists` | Check if file exists |
| `file_length` | Count non-empty lines |
| `append_file` | Append to file |
| `save_content` | Write/overwrite file |
| `glob` | Find files by pattern |
| `grep_string` | Search file for literal string |
| `grep_regex` | Search file with regex |
| `http_get` | HTTP GET request |
| `http_request` | HTTP with method/headers/body |
| `jq` | Query JSON with jq syntax |
| `exec_python` | Run inline Python |
| `exec_python_file` | Run Python file |
| `exec_ts` | Run inline TypeScript (bun) |
| `exec_ts_file` | Run TypeScript file |
| `run_module` | Run osmedeus module |
| `run_flow` | Run osmedeus flow |

### Agent Export Variables

```
{{agent_content}}             Final agent response
{{agent_history}}             Conversation history
{{agent_iterations}}          Number of iterations
{{agent_total_tokens}}        Total tokens used
{{agent_prompt_tokens}}       Prompt tokens
{{agent_completion_tokens}}   Completion tokens
{{agent_tool_results}}        Tool call results
{{agent_plan}}                Planning stage output
{{agent_goal_results}}        Per-goal results (multi-goal)
```

## Agent-ACP

Spawn an external AI coding agent as a subprocess via the Agent Communication Protocol (ACP). Unlike `agent` (internal LLM loop), `agent-acp` delegates to real agent binaries.

Built-in agents: `claude-code`, `codex`, `opencode`, `gemini`

```yaml
- name: code-review
  type: agent-acp

  # Built-in agent name (or use acp_config.command for custom)
  agent: claude-code

  # Working directory for the ACP session
  cwd: "{{Output}}"

  # Restrict file access
  allowed_paths:
    - "{{Output}}"

  # Agent configuration
  acp_config:
    env:
      CUSTOM_VAR: "value"
    write_enabled: true   # Allow file writes (default: false)
    # Custom agent (overrides built-in):
    # command: "my-agent"
    # args: ["--flag"]

  # Conversation messages (prompt)
  messages:
    - role: system
      content: "You are a security analyst."
    - role: user
      content: "Analyze the scan results in {{Output}} and create a summary."

  exports:
    analysis: "{{acp_output}}"
    errors: "{{acp_stderr}}"
    agent_used: "{{acp_agent}}"
```

### Agent-ACP Export Variables

```
{{acp_output}}    Collected agent text output
{{acp_stderr}}    Agent process stderr
{{acp_agent}}     Agent name used
```

## Decision Fields

The `decision` field supports two routing styles:

### Switch/Case (Exact String Matching)

```yaml
decision:
  switch: "{{variable}}"
  cases:
    "value1": {goto: step-a}
    "value2": {goto: step-b, command: "echo 'also run this'"}
  default: {goto: fallback}
```

Each case (`DecisionCase`) supports inline execution alongside `goto`:
- `goto` - Jump to named step (`_end` terminates workflow)
- `command` / `commands` - Execute shell command(s)
- `function` / `functions` - Execute JS function(s)

### Conditions (Boolean Expressions)

```yaml
decision:
  conditions:
    - if: "file_length('{{file}}') > 0"
      goto: process-results
    - if: "{{enableScan}}"
      command: "nmap {{Target}}"
    - if: "is_empty('{{output}}')"
      functions:
        - "log_warn('empty output')"
        - "skip('nothing to process')"
```

Each condition (`DecisionCondition`) supports:
- `if` - JS boolean expression (required)
- `goto` - Jump to named step
- `command` / `commands` - Execute shell command(s)
- `function` / `functions` - Execute JS function(s)

All matching conditions execute (no short-circuit). See [workflow-advanced.md](workflow-advanced.md#condition-based-decision-routing) for details.

## Action Handlers (on_success / on_error)

Available on all step types:

```yaml
on_success:
  - action: log
    message: "Step completed"
    condition: "true"  # Optional JS condition

  - action: export
    name: var_name
    value: "exported_value"

  - action: notify
    notify: "Scan done for {{Target}}"

  - action: run
    type: bash
    command: "echo cleanup"

  - action: run
    type: function
    functions: ['log_info("follow-up")']

  - action: abort
    message: "Critical failure"
    condition: "false"  # Guard with condition

  - action: continue
    message: "Non-critical, continuing"
```

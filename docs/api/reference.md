# API Reference

## Error Responses

All endpoints return errors in a consistent format:

```json
{
  "error": true,
  "message": "Error description"
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `201` - Created
- `202` - Accepted (async operation started)
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid token)
- `404` - Not Found
- `500` - Internal Server Error

---

## Pagination

Endpoints that return lists support pagination via query parameters:

| Parameter | Default | Max | Description |
|-----------|---------|-----|-------------|
| `offset` | 0 | - | Number of records to skip |
| `limit` | 20 | 10000 | Maximum records to return |

**Example:**
```bash
curl "http://localhost:8002/osm/api/assets?offset=100&limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Cron Expression Reference

Schedules use standard cron expressions:

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sunday=0)
│ │ │ │ │
* * * * *
```

**Examples:**
- `0 2 * * *` - Every day at 2:00 AM
- `0 0 * * 0` - Every Sunday at midnight
- `*/30 * * * *` - Every 30 minutes
- `0 9-17 * * 1-5` - Every hour from 9 AM to 5 PM, Monday to Friday

---

## Workflow Step Types

Reference documentation for workflow step types used in YAML workflow definitions.

### bash

Execute shell commands on the local system or configured runner.

```yaml
- name: run-nuclei
  type: bash
  log: "Running nuclei scan"
  command: nuclei -u {{Target}} -o {{Output}}/nuclei.txt
  timeout: 3600
  exports:
    nuclei_results: "{{Output}}/nuclei.txt"
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique step name |
| `type` | string | Yes | Must be `bash` |
| `command` | string | No* | Single command to execute |
| `commands` | array | No* | Sequential commands |
| `parallel_commands` | array | No* | Commands to run in parallel |
| `timeout` | int | No | Timeout in seconds |
| `log` | string | No | Log message displayed during execution |
| `pre_condition` | string | No | Condition that must be true to run |
| `exports` | map | No | Variables to export after execution |

*One of `command`, `commands`, or `parallel_commands` is required.

---

### function

Execute utility functions written in JavaScript via Otto VM.

```yaml
- name: check-file
  type: function
  log: "Checking if results exist"
  function: fileExists("{{Output}}/results.txt")
  exports:
    has_results: "{{check_file_output}}"
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique step name |
| `type` | string | Yes | Must be `function` |
| `function` | string | No* | Single function to execute |
| `functions` | array | No* | Sequential functions |
| `parallel_functions` | array | No* | Functions to run in parallel |

*One of `function`, `functions`, or `parallel_functions` is required.

---

### parallel-steps

Run multiple steps concurrently.

```yaml
- name: parallel-scans
  type: parallel-steps
  log: "Running scans in parallel"
  parallel_steps:
    - name: nuclei-scan
      type: bash
      command: nuclei -u {{Target}}
    - name: httpx-scan
      type: bash
      command: httpx -u {{Target}}
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique step name |
| `type` | string | Yes | Must be `parallel-steps` |
| `parallel_steps` | array | Yes | Array of steps to run concurrently |

---

### foreach

Iterate over items from a file or array.

```yaml
- name: scan-subdomains
  type: foreach
  log: "Scanning each subdomain"
  input: "{{Output}}/subdomains.txt"
  variable: subdomain
  step:
    name: scan-subdomain
    type: bash
    command: httpx -u [[subdomain]]
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique step name |
| `type` | string | Yes | Must be `foreach` |
| `input` | string | Yes | File path or array to iterate over |
| `variable` | string | Yes | Loop variable name (use `[[variable]]`) |
| `step` | object | Yes | Step to execute for each item |
| `parallel` | int | No | Number of parallel iterations |

---

### remote-bash

Execute commands in Docker containers or via SSH.

```yaml
- name: docker-scan
  type: remote-bash
  log: "Running scan in Docker"
  step_runner: docker
  step_runner_config:
    image: "projectdiscovery/nuclei:latest"
    volumes:
      - "{{Output}}:/output"
  command: nuclei -u {{Target}} -o /output/nuclei.txt
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique step name |
| `type` | string | Yes | Must be `remote-bash` |
| `step_runner` | string | Yes | `docker` or `ssh` |
| `step_runner_config` | object | No | Runner-specific configuration |
| `command` | string | No* | Command to execute |
| `commands` | array | No* | Sequential commands |

**Docker Configuration:**
```yaml
step_runner_config:
  image: "image:tag"
  volumes: ["host:container"]
  env:
    KEY: value
  network: "host"
  workdir: "/app"
```

**SSH Configuration:**
```yaml
step_runner_config:
  host: "worker.example.com"
  user: "ubuntu"
  key_file: "~/.ssh/id_rsa"
  port: 22
```

---

### http

Make HTTP requests and capture responses.

```yaml
- name: api-request
  type: http
  log: "Calling API"
  url: "https://api.example.com/endpoint"
  method: POST
  headers:
    Content-Type: "application/json"
    Authorization: "Bearer {{api_token}}"
  request_body: '{"domain": "{{Target}}"}'
  timeout: 30
  exports:
    api_response: "{{api_request_body}}"
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique step name |
| `type` | string | Yes | Must be `http` |
| `url` | string | Yes | Request URL |
| `method` | string | No | HTTP method (default: GET) |
| `headers` | map | No | Request headers |
| `request_body` | string | No | Request body for POST/PUT |
| `timeout` | int | No | Timeout in seconds |

**Auto-Exports:**
- `<step_name>_status_code` - HTTP status code
- `<step_name>_body` - Response body
- `<step_name>_headers` - Response headers

---

### llm

Execute LLM (Large Language Model) API calls for AI-powered analysis.

```yaml
- name: analyze-target
  type: llm
  log: "Analyzing target with LLM"
  messages:
    - role: system
      content: "You are a security analyst."
    - role: user
      content: "Analyze the security of {{Target}}"
  llm_config:
    max_tokens: 1000
    temperature: 0.7
  timeout: 60
  exports:
    analysis: "{{analyze_target_content}}"
```

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique step name |
| `type` | string | Yes | Must be `llm` |
| `messages` | array | No* | Chat messages (role + content) |
| `is_embedding` | bool | No | Set to true for embeddings |
| `embedding_input` | array | No* | Strings to embed |
| `llm_config` | object | No | Step-level LLM configuration |
| `tools` | array | No | Tool definitions for function calling |
| `tool_choice` | string | No | Tool selection (`auto`, `none`, etc.) |
| `extra_llm_parameters` | map | No | Additional provider-specific params |
| `timeout` | int | No | Timeout in seconds |

*Either `messages` or `embedding_input` (with `is_embedding: true`) is required.

**Message Format:**
```yaml
messages:
  - role: system
    content: "System prompt"
  - role: user
    content: "User message with {{variables}}"
```

**Multimodal Messages (with images):**
```yaml
messages:
  - role: user
    content:
      - type: text
        text: "What do you see in this screenshot?"
      - type: image_url
        image_url:
          url: "data:image/png;base64,{{screenshot_base64}}"
```

**LLM Configuration Override:**
```yaml
llm_config:
  model: "gpt-4"
  max_tokens: 2000
  temperature: 0.3
  response_format:
    type: json_object
```

**Embeddings:**
```yaml
- name: generate-embeddings
  type: llm
  is_embedding: true
  embedding_input:
    - "{{Target}} security analysis"
    - "vulnerability assessment"
  exports:
    embeddings: "{{generate_embeddings_llm_resp}}"
```

**Tool Calling:**
```yaml
- name: with-tools
  type: llm
  messages:
    - role: user
      content: "What DNS records exist for {{Target}}?"
  tools:
    - type: function
      function:
        name: dns_lookup
        description: "Look up DNS records"
        parameters:
          type: object
          properties:
            domain:
              type: string
          required: [domain]
  tool_choice: auto
```

**Auto-Exports:**
- `<step_name>_llm_resp` - Full response object (id, model, usage, content, tool_calls)
- `<step_name>_content` - Just the content string for easy access

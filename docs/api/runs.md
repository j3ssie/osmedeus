# Runs (Scans)

## Create a New Scan

Execute a workflow against a target.

**Basic scan with flow workflow:**
```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "flow": "subdomain-enum",
    "target": "example.com"
  }'
```

**Basic scan with module workflow:**
```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "module": "port-scan",
    "target": "example.com"
  }'
```

**Scan with custom parameters:**
```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "flow": "subdomain-enum",
    "target": "example.com",
    "params": {
      "threads": "50",
      "timeout": "30"
    }
  }'
```

**Scan with priority and timeout:**
```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "flow": "subdomain-enum",
    "target": "example.com",
    "priority": "high",
    "timeout": 60
  }'
```

**Scan with Docker runner:**
```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "flow": "subdomain-enum",
    "target": "example.com",
    "runner_type": "docker",
    "docker_image": "osmedeus/osmedeus:latest"
  }'
```

**Scan with SSH runner:**
```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "flow": "subdomain-enum",
    "target": "example.com",
    "runner_type": "ssh",
    "ssh_host": "worker1.example.com"
  }'
```

**Response:**
```json
{
  "message": "Run started",
  "workflow": "subdomain-enum",
  "kind": "flow",
  "target": "example.com",
  "target_count": 1,
  "priority": "high",
  "runner_type": "docker",
  "timeout": 60
}
```

---

## Multi-Target Scanning

Scan multiple targets with concurrency control:

```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "flow": "subdomain-enum",
    "targets": ["example.com", "test.com", "demo.com"],
    "concurrency": 3
  }'
```

**Response:**
```json
{
  "message": "Scan started",
  "workflow": "subdomain-enum",
  "kind": "flow",
  "target_count": 3,
  "targets": ["example.com", "test.com", "demo.com"],
  "concurrency": 3,
  "priority": "medium"
}
```

---

## Scan from Uploaded Target File

Use an uploaded target file (from `/osm/api/upload-file`) for running:

```bash
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "module": "port-scan",
    "target_file": "/home/user/osmedeus-base/data/uploads/targets.txt",
    "concurrency": 5
  }'
```

This is similar to CLI's `-T` flag: `osmedeus run -m port-scan -T targets.txt`

---

## List Runs

Get a paginated list of all runs.

```bash
curl http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | - | Filter by status: `pending`, `running`, `completed`, `failed` |
| `workflow_name` | string | - | Filter by workflow name |
| `target` | string | - | Filter by target |
| `offset` | int | 0 | Pagination offset |
| `limit` | int | 20 | Maximum records to return |

**Response:**
```json
{
  "data": [
    {
      "id": "run-abc123",
      "run_id": "run-2025-01-15-subdomain-enum-example.com",
      "workflow_name": "subdomain-enum",
      "workflow_kind": "flow",
      "target": "example.com",
      "params": {"threads": "50"},
      "status": "running",
      "workspace_path": "/home/user/osmedeus-base/workspaces/example.com",
      "started_at": "2025-01-15T10:00:00Z",
      "completed_at": null,
      "total_steps": 10,
      "completed_steps": 3,
      "trigger_type": "manual",
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-01-15T10:03:00Z"
    }
  ],
  "pagination": {
    "total": 50,
    "offset": 0,
    "limit": 20
  }
}
```

---

## Get Run Details

Get details of a specific run by ID.

```bash
curl http://localhost:8002/osm/api/runs/run-abc123 \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "id": "run-abc123",
  "run_id": "run-2025-01-15-subdomain-enum-example.com",
  "workflow_name": "subdomain-enum",
  "workflow_kind": "flow",
  "target": "example.com",
  "params": {"threads": "50"},
  "status": "completed",
  "workspace_path": "/home/user/osmedeus-base/workspaces/example.com",
  "started_at": "2025-01-15T10:00:00Z",
  "completed_at": "2025-01-15T10:30:00Z",
  "error_message": "",
  "schedule_id": "",
  "trigger_type": "manual",
  "trigger_name": "",
  "total_steps": 10,
  "completed_steps": 10,
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

---

## Cancel Run

Cancel a running workflow execution.

```bash
curl -X DELETE http://localhost:8002/osm/api/runs/run-abc123 \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "message": "Run cancellation requested",
  "id": "run-abc123"
}
```

---

## Get Run Steps

Get all step results for a specific run.

```bash
curl http://localhost:8002/osm/api/runs/run-abc123/steps \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "step-xyz789",
      "run_id": "run-abc123",
      "step_name": "run-subfinder",
      "step_type": "bash",
      "status": "completed",
      "command": "subfinder -d example.com -o subdomains.txt",
      "output": "Found 150 subdomains",
      "error_message": "",
      "exports": {"subdomains_file": "subdomains.txt"},
      "duration_ms": 45000,
      "log_file": "/workspaces/example.com/logs/run-subfinder.log",
      "started_at": "2025-01-15T10:01:00Z",
      "completed_at": "2025-01-15T10:01:45Z",
      "created_at": "2025-01-15T10:01:00Z"
    },
    {
      "id": "step-def456",
      "run_id": "run-abc123",
      "step_name": "run-httpx",
      "step_type": "bash",
      "status": "completed",
      "command": "httpx -l subdomains.txt -o alive.txt",
      "output": "Probed 150 hosts, 89 alive",
      "error_message": "",
      "exports": {"alive_file": "alive.txt"},
      "duration_ms": 120000,
      "log_file": "/workspaces/example.com/logs/run-httpx.log",
      "started_at": "2025-01-15T10:01:45Z",
      "completed_at": "2025-01-15T10:03:45Z",
      "created_at": "2025-01-15T10:01:45Z"
    }
  ]
}
```

---

## Get Run Artifacts

Get all output artifacts for a specific run.

```bash
curl http://localhost:8002/osm/api/runs/run-abc123/artifacts \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "artifact-001",
      "run_id": "run-abc123",
      "name": "subdomains.txt",
      "path": "/workspaces/example.com/subdomains.txt",
      "type": "text",
      "size_bytes": 4523,
      "line_count": 150,
      "description": "Discovered subdomains",
      "created_at": "2025-01-15T10:01:45Z"
    },
    {
      "id": "artifact-002",
      "run_id": "run-abc123",
      "name": "alive.txt",
      "path": "/workspaces/example.com/alive.txt",
      "type": "text",
      "size_bytes": 2890,
      "line_count": 89,
      "description": "Alive HTTP endpoints",
      "created_at": "2025-01-15T10:03:45Z"
    },
    {
      "id": "artifact-003",
      "run_id": "run-abc123",
      "name": "nuclei-results.json",
      "path": "/workspaces/example.com/nuclei-results.json",
      "type": "json",
      "size_bytes": 15234,
      "line_count": 45,
      "description": "Nuclei vulnerability scan results",
      "created_at": "2025-01-15T10:15:00Z"
    }
  ]
}
```

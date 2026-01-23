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
  "job_id": "a1b2c3d4",
  "run_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "poll_url": "/osm/api/jobs/a1b2c3d4",
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
  "message": "Run started",
  "workflow": "subdomain-enum",
  "kind": "flow",
  "target_count": 3,
  "targets": ["example.com", "test.com", "demo.com"],
  "concurrency": 3,
  "priority": "medium",
  "job_id": "b2c3d4e5",
  "status": "queued",
  "poll_url": "/osm/api/jobs/b2c3d4e5"
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
      "id": 1,
      "run_uuid": "550e8400-e29b-41d4-a716-446655440000",
      "workflow_name": "subdomain-enum",
      "workflow_kind": "flow",
      "target": "example.com",
      "params": {"threads": "50"},
      "status": "running",
      "workspace": "example.com",
      "started_at": "2025-01-15T10:00:00Z",
      "completed_at": null,
      "total_steps": 10,
      "completed_steps": 3,
      "current_pid": 12345,
      "trigger_type": "manual",
      "run_group_id": "a1b2c3d4",
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

**Note:** The `current_pid` field shows the process ID of the currently running command. This can be used to identify and cancel the running process. When the run completes, this field is cleared (set to 0 or omitted).

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
  "data": {
    "id": 1,
    "run_uuid": "550e8400-e29b-41d4-a716-446655440000",
    "workflow_name": "subdomain-enum",
    "workflow_kind": "flow",
    "target": "example.com",
    "params": {"threads": "50"},
    "status": "completed",
    "workspace": "example.com",
    "started_at": "2025-01-15T10:00:00Z",
    "completed_at": "2025-01-15T10:30:00Z",
    "error_message": "",
    "schedule_id": "",
    "trigger_type": "manual",
    "trigger_name": "",
    "run_group_id": "a1b2c3d4",
    "total_steps": 10,
    "completed_steps": 10,
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

**Note:** You can use either the numeric `id` or the `run_uuid` to fetch run details.

---

## Cancel Run

Cancel a running workflow execution. This will terminate all running processes associated with the run.

```bash
# Cancel by run_uuid
curl -X DELETE http://localhost:8002/osm/api/runs/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $TOKEN"

# Or cancel by numeric id
curl -X DELETE http://localhost:8002/osm/api/runs/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response (processes killed successfully):**
```json
{
  "message": "Run cancelled successfully",
  "id": 1,
  "run_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "killed_pids": [12345, 12346],
  "processes_terminated": 2,
  "kill_method": "registry"
}
```

**Response (using database PID fallback):**
```json
{
  "message": "Run cancelled successfully",
  "id": 1,
  "run_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "killed_pids": [12345],
  "processes_terminated": 1,
  "kill_method": "database_pid"
}
```

**Response (no active processes found):**
```json
{
  "message": "Run cancelled successfully",
  "id": 1,
  "run_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "note": "No active processes found to terminate; database status updated"
}
```

**Kill Methods:**
- `registry` - Processes were tracked in memory and killed via the run registry (API-initiated runs)
- `database_pid` - Process was killed using the PID stored in the database (fallback method)

---

## Get Run Steps

Get all step results for a specific run.

```bash
# Using run_uuid
curl http://localhost:8002/osm/api/runs/550e8400-e29b-41d4-a716-446655440000/steps \
  -H "Authorization: Bearer $TOKEN"

# Or using numeric id
curl http://localhost:8002/osm/api/runs/1/steps \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "run_id": 1,
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
      "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
      "run_id": 1,
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
# Using run_uuid
curl http://localhost:8002/osm/api/runs/550e8400-e29b-41d4-a716-446655440000/artifacts \
  -H "Authorization: Bearer $TOKEN"

# Or using numeric id
curl http://localhost:8002/osm/api/runs/1/artifacts \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
      "run_id": 1,
      "workspace": "example.com",
      "name": "subdomains.txt",
      "artifact_path": "/workspaces/example.com/subdomains.txt",
      "artifact_type": "output",
      "content_type": "txt",
      "size_bytes": 4523,
      "line_count": 150,
      "description": "Discovered subdomains",
      "created_at": "2025-01-15T10:01:45Z"
    },
    {
      "id": "d4e5f6a7-b8c9-0123-def0-234567890123",
      "run_id": 1,
      "workspace": "example.com",
      "name": "alive.txt",
      "artifact_path": "/workspaces/example.com/alive.txt",
      "artifact_type": "output",
      "content_type": "txt",
      "size_bytes": 2890,
      "line_count": 89,
      "description": "Alive HTTP endpoints",
      "created_at": "2025-01-15T10:03:45Z"
    },
    {
      "id": "e5f6a7b8-c9d0-1234-ef01-345678901234",
      "run_id": 1,
      "workspace": "example.com",
      "name": "nuclei-results.json",
      "artifact_path": "/workspaces/example.com/nuclei-results.json",
      "artifact_type": "output",
      "content_type": "json",
      "size_bytes": 15234,
      "line_count": 45,
      "description": "Nuclei vulnerability scan results",
      "created_at": "2025-01-15T10:15:00Z"
    }
  ]
}
```

**Artifact Types:**
- `report` - Generated reports from the workflow's reports section
- `state_file` - State files like run-state.json, run-execution.log
- `output` - General output files from steps
- `screenshot` - Screenshots captured during the scan

**Content Types:**
- `json`, `jsonl`, `yaml`, `html`, `md`, `log`, `pdf`, `png`, `txt`, `zip`, `folder`, `unknown`

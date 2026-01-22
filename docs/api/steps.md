# Step Results

## List Step Results

Get a paginated list of step execution results with optional filtering.

**List all step results:**
```bash
curl http://localhost:8002/osm/api/step-results \
  -H "Authorization: Bearer $TOKEN"
```

**With pagination:**
```bash
curl "http://localhost:8002/osm/api/step-results?offset=0&limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by workspace:**
```bash
curl "http://localhost:8002/osm/api/step-results?workspace=example_com" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by status:**
```bash
curl "http://localhost:8002/osm/api/step-results?status=completed" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by step type:**
```bash
curl "http://localhost:8002/osm/api/step-results?step_type=bash" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by run ID:**
```bash
curl "http://localhost:8002/osm/api/step-results?run_id=123" \
  -H "Authorization: Bearer $TOKEN"
```

**Multiple filters:**
```bash
curl "http://localhost:8002/osm/api/step-results?workspace=example_com&status=completed&limit=100" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `workspace` | string | Filter by workspace name |
| `status` | string | Filter by status (pending, running, completed, failed) |
| `step_type` | string | Filter by step type (bash, function, foreach, parallel-steps, remote-bash, http, llm) |
| `run_id` | int | Filter by run ID |
| `offset` | int | Pagination offset (default: 0) |
| `limit` | int | Maximum records to return (default: 20, max: 10000) |

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "run_id": 123,
      "step_name": "run-subfinder",
      "step_type": "bash",
      "status": "completed",
      "command": "subfinder -d example.com -o {{Output}}/subdomains.txt",
      "output": "Found 150 subdomains",
      "error": "",
      "started_at": "2025-01-15T10:00:00Z",
      "completed_at": "2025-01-15T10:01:30Z",
      "duration_ms": 90000,
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-01-15T10:01:30Z"
    },
    {
      "id": 2,
      "run_id": 123,
      "step_name": "run-httpx",
      "step_type": "bash",
      "status": "completed",
      "command": "httpx -l {{Output}}/subdomains.txt -o {{Output}}/httpx.txt",
      "output": "Probed 150 hosts, 120 alive",
      "error": "",
      "started_at": "2025-01-15T10:01:30Z",
      "completed_at": "2025-01-15T10:03:00Z",
      "duration_ms": 90000,
      "created_at": "2025-01-15T10:01:30Z",
      "updated_at": "2025-01-15T10:03:00Z"
    },
    {
      "id": 3,
      "run_id": 124,
      "step_name": "process-results",
      "step_type": "function",
      "status": "completed",
      "command": "",
      "output": "true",
      "error": "",
      "started_at": "2025-01-15T11:00:00Z",
      "completed_at": "2025-01-15T11:00:01Z",
      "duration_ms": 1000,
      "created_at": "2025-01-15T11:00:00Z",
      "updated_at": "2025-01-15T11:00:01Z"
    }
  ],
  "pagination": {
    "total": 250,
    "offset": 0,
    "limit": 20
  }
}
```

**Step Status Values:**

| Status | Description |
|--------|-------------|
| `pending` | Step is queued but not yet started |
| `running` | Step is currently executing |
| `completed` | Step finished successfully |
| `failed` | Step failed with an error |
| `skipped` | Step was skipped (pre_condition not met) |

**Step Types:**

| Type | Description |
|------|-------------|
| `bash` | Shell command execution |
| `function` | Utility function execution |
| `foreach` | Loop over input items |
| `parallel-steps` | Execute steps in parallel |
| `remote-bash` | Remote command execution (Docker/SSH) |
| `http` | HTTP request step |
| `llm` | LLM/AI step |

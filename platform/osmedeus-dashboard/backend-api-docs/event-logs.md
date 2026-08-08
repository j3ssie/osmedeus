# Event Logs

## List Event Logs

Get a paginated list of event logs with optional filtering.

**List all event logs:**
```bash
curl http://localhost:8002/osm/api/event-logs \
  -H "Authorization: Bearer $TOKEN"
```

**With pagination:**
```bash
curl "http://localhost:8002/osm/api/event-logs?offset=0&limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by workspace:**
```bash
curl "http://localhost:8002/osm/api/event-logs?workspace=example.com" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by topic:**
```bash
curl "http://localhost:8002/osm/api/event-logs?topic=run.completed" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by run ID:**
```bash
curl "http://localhost:8002/osm/api/event-logs?run_id=abc12345" \
  -H "Authorization: Bearer $TOKEN"
```

**Multiple filters:**
```bash
curl "http://localhost:8002/osm/api/event-logs?workspace=example.com&processed=false&limit=100" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `topic` | string | Filter by event topic (e.g., "run.started", "run.completed") |
| `name` | string | Filter by event name |
| `source` | string | Filter by event source (e.g., "executor", "scheduler", "api") |
| `workspace` | string | Filter by workspace name |
| `run_id` | string | Filter by run ID |
| `workflow_name` | string | Filter by workflow name |
| `processed` | bool | Filter by processed status ("true" or "false") |
| `offset` | int | Pagination offset (default: 0) |
| `limit` | int | Maximum records to return (default: 20, max: 10000) |

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "topic": "run.completed",
      "event_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "subdomain-enum-completed",
      "source": "executor",
      "data_type": "scan",
      "data": "{\"scan_id\":\"abc12345\",\"target\":\"example.com\",\"duration_ms\":3600000,\"assets_found\":150,\"steps_completed\":10}",
      "workspace": "example.com",
      "run_id": "abc12345",
      "workflow_name": "subdomain-enum",
      "processed": true,
      "processed_at": "2025-01-15T10:30:00Z",
      "error": "",
      "created_at": "2025-01-15T09:30:00Z"
    },
    {
      "id": 2,
      "topic": "run.started",
      "event_id": "660e8400-e29b-41d4-a716-446655440001",
      "name": "port-scan-started",
      "source": "api",
      "data_type": "scan",
      "data": "{\"scan_id\":\"def67890\",\"target\":\"test.com\",\"params\":{\"ports\":\"top-1000\"}}",
      "workspace": "test.com",
      "run_id": "def67890",
      "workflow_name": "port-scan",
      "processed": true,
      "processed_at": "2025-01-15T11:00:00Z",
      "error": "",
      "created_at": "2025-01-15T11:00:00Z"
    },
    {
      "id": 3,
      "topic": "asset.discovered",
      "event_id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "httpx-asset-found",
      "source": "executor",
      "data_type": "asset",
      "data": "{\"url\":\"https://api.example.com\",\"status_code\":200,\"title\":\"API Documentation\",\"tech\":[\"nginx\",\"nodejs\"]}",
      "workspace": "example.com",
      "run_id": "abc12345",
      "workflow_name": "subdomain-enum",
      "processed": true,
      "processed_at": "2025-01-15T10:15:00Z",
      "error": "",
      "created_at": "2025-01-15T10:15:00Z"
    },
    {
      "id": 4,
      "topic": "schedule.triggered",
      "event_id": "880e8400-e29b-41d4-a716-446655440003",
      "name": "daily-scan-triggered",
      "source": "scheduler",
      "data_type": "schedule",
      "data": "{\"schedule_id\":\"sch_1234567890\",\"trigger_type\":\"cron\",\"schedule\":\"0 2 * * *\"}",
      "workspace": "example.com",
      "run_id": "ghi11111",
      "workflow_name": "subdomain-enum",
      "processed": true,
      "processed_at": "2025-01-16T02:00:00Z",
      "error": "",
      "created_at": "2025-01-16T02:00:00Z"
    },
    {
      "id": 5,
      "topic": "run.failed",
      "event_id": "990e8400-e29b-41d4-a716-446655440004",
      "name": "nuclei-scan-failed",
      "source": "executor",
      "data_type": "scan",
      "data": "{\"scan_id\":\"jkl22222\",\"target\":\"unreachable.com\",\"error\":\"connection timeout\"}",
      "workspace": "unreachable.com",
      "run_id": "jkl22222",
      "workflow_name": "nuclei-scan",
      "processed": false,
      "processed_at": null,
      "error": "connection timeout after 5 retries",
      "created_at": "2025-01-15T14:00:00Z"
    },
    {
      "id": 6,
      "topic": "step.completed",
      "event_id": "aae8400-e29b-41d4-a716-446655440005",
      "name": "run-subfinder-completed",
      "source": "executor",
      "data_type": "step",
      "data": "{\"step_name\":\"run-subfinder\",\"duration_ms\":45000,\"output_lines\":150}",
      "workspace": "example.com",
      "run_id": "abc12345",
      "workflow_name": "subdomain-enum",
      "processed": true,
      "processed_at": "2025-01-15T10:01:45Z",
      "error": "",
      "created_at": "2025-01-15T10:01:45Z"
    }
  ],
  "pagination": {
    "total": 150,
    "offset": 0,
    "limit": 20
  }
}
```

**Available Event Topics:**

| Topic | Description |
|-------|-------------|
| `run.started` | Workflow execution started |
| `run.completed` | Workflow execution completed successfully |
| `run.failed` | Workflow execution failed |
| `asset.discovered` | New asset discovered during scan |
| `asset.updated` | Existing asset information updated |
| `webhook.received` | External webhook received |
| `schedule.triggered` | Scheduled workflow triggered |
| `step.completed` | Individual step completed |
| `step.failed` | Individual step failed |

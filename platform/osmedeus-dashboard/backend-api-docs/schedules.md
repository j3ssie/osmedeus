# Schedules

## List Schedules

Get a paginated list of all scheduled workflows.

```bash
curl http://localhost:8002/osm/api/schedules \
  -H "Authorization: Bearer $TOKEN"
```

**With pagination:**
```bash
curl "http://localhost:8002/osm/api/schedules?offset=0&limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "sch_1234567890",
      "name": "daily-scan",
      "workflow_name": "subdomain-enum",
      "workflow_path": "/home/user/osmedeus-base/workflows/flows/subdomain-enum.yaml",
      "trigger_name": "daily-scan-trigger",
      "trigger_type": "cron",
      "schedule": "0 2 * * *",
      "event_topic": "",
      "watch_path": "",
      "input_config": {
        "target": "example.com",
        "threads": "50"
      },
      "is_enabled": true,
      "last_run": "2025-01-15T02:00:00Z",
      "next_run": "2025-01-16T02:00:00Z",
      "run_count": 30,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-15T02:00:00Z"
    },
    {
      "id": "sch_0987654321",
      "name": "weekly-full-recon",
      "workflow_name": "full-recon",
      "workflow_path": "/home/user/osmedeus-base/workflows/flows/full-recon.yaml",
      "trigger_name": "weekly-trigger",
      "trigger_type": "cron",
      "schedule": "0 0 * * 0",
      "event_topic": "",
      "watch_path": "",
      "input_config": {
        "target": "example.com",
        "threads": "100",
        "runner_type": "docker"
      },
      "is_enabled": true,
      "last_run": "2025-01-12T00:00:00Z",
      "next_run": "2025-01-19T00:00:00Z",
      "run_count": 5,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-12T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 5,
    "offset": 0,
    "limit": 20
  }
}
```

---

## Create Schedule

Create a new scheduled workflow execution.

```bash
curl -X POST http://localhost:8002/osm/api/schedules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "daily-scan",
    "workflow_name": "subdomain-enum",
    "workflow_kind": "flow",
    "target": "example.com",
    "schedule": "0 2 * * *",
    "enabled": true
  }'
```

**With additional parameters:**
```bash
curl -X POST http://localhost:8002/osm/api/schedules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "weekly-full-scan",
    "workflow_name": "full-recon",
    "workflow_kind": "flow",
    "target": "example.com",
    "schedule": "0 0 * * 0",
    "enabled": true,
    "params": {
      "threads": "100"
    },
    "runner_type": "docker"
  }'
```

**Response:**
```json
{
  "message": "Schedule created",
  "data": {
    "id": "sch_1234567890",
    "name": "daily-scan",
    "workflow_name": "subdomain-enum",
    "schedule": "0 2 * * *",
    "is_enabled": true
  }
}
```

---

## Get Schedule

Get details of a specific schedule.

```bash
curl http://localhost:8002/osm/api/schedules/sch_1234567890 \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "id": "sch_1234567890",
  "name": "daily-scan",
  "workflow_name": "subdomain-enum",
  "workflow_path": "/home/user/osmedeus-base/workflows/flows/subdomain-enum.yaml",
  "trigger_name": "daily-scan-trigger",
  "trigger_type": "cron",
  "schedule": "0 2 * * *",
  "event_topic": "",
  "watch_path": "",
  "input_config": {
    "target": "example.com",
    "threads": "50"
  },
  "is_enabled": true,
  "last_run": "2025-01-15T02:00:00Z",
  "next_run": "2025-01-16T02:00:00Z",
  "run_count": 30,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-15T02:00:00Z"
}
```

---

## Update Schedule

Update an existing schedule.

```bash
curl -X PUT http://localhost:8002/osm/api/schedules/sch_1234567890 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "updated-daily-scan",
    "schedule": "0 3 * * *"
  }'
```

**Update only the schedule:**
```bash
curl -X PUT http://localhost:8002/osm/api/schedules/sch_1234567890 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "schedule": "0 4 * * *"
  }'
```

**Response:**
```json
{
  "message": "Schedule updated",
  "data": {
    "id": "sch_1234567890",
    "name": "updated-daily-scan",
    "schedule": "0 3 * * *"
  }
}
```

---

## Delete Schedule

Delete a schedule.

```bash
curl -X DELETE http://localhost:8002/osm/api/schedules/sch_1234567890 \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "message": "Schedule deleted"
}
```

---

## Enable Schedule

Enable a disabled schedule.

```bash
curl -X POST http://localhost:8002/osm/api/schedules/sch_1234567890/enable \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "message": "Schedule enabled"
}
```

---

## Disable Schedule

Disable an enabled schedule.

```bash
curl -X POST http://localhost:8002/osm/api/schedules/sch_1234567890/disable \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "message": "Schedule disabled"
}
```

---

## Trigger Schedule

Manually trigger a scheduled workflow execution.

```bash
curl -X POST http://localhost:8002/osm/api/schedules/sch_1234567890/trigger \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "message": "Schedule triggered",
  "schedule": "daily-scan",
  "workflow": "subdomain-enum"
}
```

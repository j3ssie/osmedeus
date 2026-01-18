# System Statistics

## Get System Stats

Get aggregated system statistics including workflows, runs, workspaces, assets, vulnerabilities, and schedules.

```bash
curl http://localhost:8002/osm/api/stats \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "workflows": {
    "total": 25,
    "flows": 10,
    "modules": 15
  },
  "runs": {
    "total": 150,
    "completed": 120,
    "running": 5,
    "failed": 10,
    "pending": 15
  },
  "workspaces": {
    "total": 50
  },
  "assets": {
    "total": 5000
  },
  "vulnerabilities": {
    "total": 150,
    "critical": 10,
    "high": 25,
    "medium": 50,
    "low": 65
  },
  "schedules": {
    "total": 8,
    "enabled": 5
  }
}
```

**Statistics Fields:**

| Category | Field | Description |
|----------|-------|-------------|
| workflows.total | int | Total number of workflows (flows + modules) |
| workflows.flows | int | Number of flow-type workflows |
| workflows.modules | int | Number of module-type workflows |
| runs.total | int | Total number of runs |
| runs.completed | int | Successfully completed runs |
| runs.running | int | Currently running workflows |
| runs.failed | int | Failed runs |
| runs.pending | int | Pending runs waiting to start |
| workspaces.total | int | Total number of scan workspaces |
| assets.total | int | Total discovered assets across all workspaces |
| vulnerabilities.total | int | Total vulnerabilities (sum of all severities) |
| vulnerabilities.critical | int | Critical severity vulnerabilities |
| vulnerabilities.high | int | High severity vulnerabilities |
| vulnerabilities.medium | int | Medium severity vulnerabilities |
| vulnerabilities.low | int | Low severity vulnerabilities |
| schedules.total | int | Total configured schedules |
| schedules.enabled | int | Currently enabled schedules |

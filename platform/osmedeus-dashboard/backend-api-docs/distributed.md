# Distributed Mode

These endpoints are only available when running the server in master mode.

## List Workers

Get a list of all registered workers in the distributed pool.

```bash
curl http://localhost:8002/osm/api/workers \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": "worker-001",
      "hostname": "worker1.example.com",
      "ip_address": "192.168.1.10",
      "status": "idle",
      "current_task": null,
      "joined_at": "2025-01-15T08:00:00Z",
      "last_heartbeat": "2025-01-15T10:30:00Z",
      "tasks_complete": 150,
      "tasks_failed": 2,
      "capabilities": ["docker", "nmap", "nuclei"],
      "cpu_cores": 8,
      "memory_gb": 16,
      "version": "1.0.0"
    },
    {
      "id": "worker-002",
      "hostname": "worker2.example.com",
      "ip_address": "192.168.1.11",
      "status": "busy",
      "current_task": "task-12345",
      "joined_at": "2025-01-15T08:05:00Z",
      "last_heartbeat": "2025-01-15T10:30:05Z",
      "tasks_complete": 120,
      "tasks_failed": 1,
      "capabilities": ["docker", "nmap", "nuclei", "masscan"],
      "cpu_cores": 16,
      "memory_gb": 32,
      "version": "1.0.0"
    }
  ],
  "count": 2
}
```

---

## Get Worker

Get details of a specific worker.

```bash
curl http://localhost:8002/osm/api/workers/worker-001 \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "id": "worker-001",
  "hostname": "worker1.example.com",
  "ip_address": "192.168.1.10",
  "status": "busy",
  "current_task": "task-12345",
  "joined_at": "2025-01-15T08:00:00Z",
  "last_heartbeat": "2025-01-15T10:30:00Z",
  "tasks_complete": 150,
  "tasks_failed": 2,
  "capabilities": ["docker", "nmap", "nuclei"],
  "cpu_cores": 8,
  "memory_gb": 16,
  "version": "1.0.0"
}
```

---

## List Tasks

Get a list of all running and completed tasks.

```bash
curl http://localhost:8002/osm/api/tasks \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "running": [
    {
      "id": "task-12345",
      "scan_id": "scan-abc123",
      "workflow_name": "subdomain-enum",
      "workflow_kind": "flow",
      "target": "example.com",
      "params": {"threads": "50", "timeout": "60"},
      "status": "running",
      "worker_id": "worker-001",
      "progress": 45,
      "current_step": "run-httpx",
      "created_at": "2025-01-15T10:00:00Z",
      "started_at": "2025-01-15T10:01:00Z"
    },
    {
      "id": "task-12346",
      "scan_id": "scan-def456",
      "workflow_name": "port-scan",
      "workflow_kind": "module",
      "target": "test.com",
      "params": {"ports": "top-1000"},
      "status": "running",
      "worker_id": "worker-002",
      "progress": 80,
      "current_step": "nmap-scan",
      "created_at": "2025-01-15T10:05:00Z",
      "started_at": "2025-01-15T10:06:00Z"
    }
  ],
  "completed": [
    {
      "task_id": "task-12340",
      "scan_id": "scan-xyz789",
      "status": "completed",
      "output": "Scan completed: 150 subdomains found, 89 alive hosts",
      "error": "",
      "exports": {
        "subdomains": "/workspaces/example.com/subdomains.txt",
        "alive_hosts": "/workspaces/example.com/alive.txt"
      },
      "completed_at": "2025-01-15T09:30:00Z",
      "duration_seconds": 1800
    },
    {
      "task_id": "task-12339",
      "scan_id": "scan-uvw456",
      "status": "failed",
      "output": "",
      "error": "Connection timeout to target",
      "exports": {},
      "completed_at": "2025-01-15T09:15:00Z",
      "duration_seconds": 300
    }
  ]
}
```

---

## Get Task

Get details of a specific task.

```bash
curl http://localhost:8002/osm/api/tasks/task-12345 \
  -H "Authorization: Bearer $TOKEN"
```

**Response (running task):**
```json
{
  "id": "task-12345",
  "scan_id": "scan-abc123",
  "workflow_name": "subdomain-enum",
  "workflow_kind": "flow",
  "target": "example.com",
  "params": {"threads": "50", "timeout": "60"},
  "status": "running",
  "worker_id": "worker-001",
  "progress": 45,
  "current_step": "run-httpx",
  "created_at": "2025-01-15T10:00:00Z",
  "started_at": "2025-01-15T10:01:00Z"
}
```

**Response (completed task):**
```json
{
  "task_id": "task-12345",
  "scan_id": "scan-abc123",
  "status": "completed",
  "output": "Scan completed: 150 subdomains found, 89 alive hosts",
  "error": "",
  "exports": {
    "subdomains": "/workspaces/example.com/subdomains.txt",
    "alive_hosts": "/workspaces/example.com/alive.txt",
    "httpx_json": "/workspaces/example.com/httpx.json"
  },
  "completed_at": "2025-01-15T10:30:00Z",
  "duration_seconds": 1740
}
```

---

## Submit Task

Submit a new task to the distributed worker queue.

```bash
curl -X POST http://localhost:8002/osm/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_name": "subdomain-enum",
    "workflow_kind": "flow",
    "target": "example.com"
  }'
```

**With parameters:**
```bash
curl -X POST http://localhost:8002/osm/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_name": "subdomain-enum",
    "workflow_kind": "flow",
    "target": "example.com",
    "params": {
      "threads": 50,
      "timeout": 30
    }
  }'
```

**Response:**
```json
{
  "message": "Task submitted",
  "task_id": "task-12346"
}
```

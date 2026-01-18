# Osmedeus API Documentation

## Overview

The Osmedeus API provides a RESTful interface for managing security automation workflows, runs, and distributed task execution.

**Base URL:** `http://localhost:8002`

**Default Port:** `8002`

## Authentication

Most API endpoints require JWT authentication. First, obtain a token via the login endpoint, then include it in subsequent requests using the `Authorization: Bearer <token>` header.

See [Authentication](authentication.md) for details.

## API Reference

| Category | Description |
|----------|-------------|
| [Public Endpoints](public.md) | Server info, health checks, Swagger docs |
| [Authentication](authentication.md) | Login and JWT token management |
| [Workflows](workflows.md) | List, view, and refresh workflows |
| [Runs](runs.md) | Create and manage workflow executions |
| [File Uploads](uploads.md) | Upload target files and workflows |
| [Snapshots](snapshots.md) | Download workspace snapshots |
| [Workspaces](workspaces.md) | List and manage workspaces |
| [Assets](assets.md) | View discovered assets |
| [Vulnerabilities](vulnerabilities.md) | View and manage vulnerabilities |
| [Event Logs](event-logs.md) | View execution event logs |
| [Functions](functions.md) | Execute and list utility functions |
| [System Statistics](system.md) | Get aggregated system stats |
| [Settings](settings.md) | Manage server configuration |
| [Installation](install.md) | Install binaries and workflows |
| [Schedules](schedules.md) | Manage scheduled workflows |
| [Distributed Mode](distributed.md) | Worker and task management |
| [LLM API](llm.md) | Large Language Model API |
| [Reference](reference.md) | Error codes, pagination, cron expressions, step types |

## Quick Start

```bash
# Get server info (no auth required)
curl http://localhost:8002/server-info

# Login and get token
export TOKEN=$(curl -s -X POST http://localhost:8002/osm/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "osmedeus", "password": "admin"}' | jq -r '.token')

# List workflows
curl http://localhost:8002/osm/api/workflows \
  -H "Authorization: Bearer $TOKEN"

# Start a scan
curl -X POST http://localhost:8002/osm/api/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"flow": "subdomain-enum", "target": "example.com"}'
```

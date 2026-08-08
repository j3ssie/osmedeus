# Public Endpoints

These endpoints do not require authentication.

## Server Info

Get server version and information.

```bash
curl http://localhost:8002/server-info
```

**Response:**
```json
{
  "message": "Oh dear me, how delightful to notice you're taking a look at this! I'm ever so pleased to let you know that osmedeus is ticking along quite nicely, thank you.",
  "version": "v5.0.0",
  "repo": "https://github.com/j3ssie/osmedeus",
  "author": "j3ssie",
  "docs": "https://docs.osmedeus.org"
}
```

---

## Health Check

Check if the server is running.

```bash
curl http://localhost:8002/health
```

**Response:**
```json
{
  "status": "ok"
}
```

---

## Readiness Check

Check if the server is ready to accept requests.

```bash
curl http://localhost:8002/health/ready
```

**Response:**
```json
{
  "status": "ready"
}
```

---

## Swagger Documentation

Access the interactive Swagger UI documentation.

```bash
# Open in browser
open http://localhost:8002/swagger/index.html
```

---

## Web UI

The web UI is served at the root path. It uses embedded UI files by default, with an option to serve from an external path.

```bash
# Access the web UI in browser
open http://localhost:8002/
```

**UI Serving Priority:**
1. If `ui_path` is configured and exists, serves from that directory
2. Otherwise, serves embedded UI files from `public/ui/`

---

## Workspace Files

Scan output files can be accessed directly via the workspace path. This endpoint is only available when `workspace_prefix` is configured in server settings.

```bash
# Access run outputs (no authentication required)
curl http://localhost:8002/ws/{workspace_prefix}/example.com/subdomain/final.txt
```

The workspace path serves files from the configured workspaces directory with directory listing enabled.

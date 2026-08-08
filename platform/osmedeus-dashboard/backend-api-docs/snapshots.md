# Snapshots

## List Snapshots

Get a list of available snapshot files in the snapshot directory.

```bash
curl http://localhost:8002/osm/api/snapshots \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "name": "example.com_1704067200.zip",
      "path": "/home/user/osmedeus-base/snapshot/example.com_1704067200.zip",
      "size": 15728640,
      "created_at": "2025-01-01T12:00:00Z"
    },
    {
      "name": "test.com_1704153600.zip",
      "path": "/home/user/osmedeus-base/snapshot/test.com_1704153600.zip",
      "size": 8388608,
      "created_at": "2025-01-02T12:00:00Z"
    }
  ],
  "count": 2,
  "path": "/home/user/osmedeus-base/snapshot"
}
```

---

## Export Workspace Snapshot

Export a workspace to a compressed zip archive and download it.

```bash
curl -X POST http://localhost:8002/osm/api/snapshots/export \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"workspace": "example.com"}' \
  --output example.com_snapshot.zip
```

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `workspace` | string | Yes | Name of the workspace to export |

**Response:**
- On success: Returns the zip file as a binary download
- Response headers include:
  - `Content-Disposition: attachment; filename=<workspace>_<timestamp>.zip`
  - `Content-Type: application/zip`
  - `X-Snapshot-Size: <size_in_bytes>`

**Error Response (404):**
```json
{
  "error": true,
  "message": "Workspace not found: example.com"
}
```

---

## Import Workspace Snapshot

Import a workspace from an uploaded zip file or URL.

**Import from file upload:**
```bash
curl -X POST http://localhost:8002/osm/api/snapshots/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@example.com_1704067200.zip"
```

**Import from URL:**
```bash
curl -X POST http://localhost:8002/osm/api/snapshots/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "url=https://example.com/snapshots/workspace.zip"
```

**Import with force overwrite:**
```bash
curl -X POST http://localhost:8002/osm/api/snapshots/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@example.com_snapshot.zip" \
  -F "force=true"
```

**Import files only (skip database):**
```bash
curl -X POST http://localhost:8002/osm/api/snapshots/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@example.com_snapshot.zip" \
  -F "skip_db=true"
```

**Form Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file` | file | No* | Snapshot zip file to import |
| `url` | string | No* | URL of snapshot to download and import |
| `force` | bool | No | Overwrite existing workspace if present (default: false) |
| `skip_db` | bool | No | Skip database import, extract files only (default: false) |

*Either `file` or `url` is required.

**Response (200):**
```json
{
  "message": "Workspace imported successfully",
  "workspace": "example.com",
  "local_path": "/home/user/workspaces-osmedeus/example.com",
  "data_source": "imported",
  "files_count": 1523,
  "warning": "Imported workspace database state may be unstable. Only import from trusted sources."
}
```

**Error Response (400):**
```json
{
  "error": true,
  "message": "Either file or url is required"
}
```

**Error Response (500 - workspace exists):**
```json
{
  "error": true,
  "message": "Failed to import snapshot: workspace already exists: /home/user/workspaces-osmedeus/example.com (use --force to overwrite)"
}
```

---

## Delete Snapshot

Delete a snapshot file by name.

```bash
curl -X DELETE http://localhost:8002/osm/api/snapshots/example.com_1704067200.zip \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "message": "Snapshot deleted successfully",
  "name": "example.com_1704067200.zip"
}
```

**Error Response (404):**
```json
{
  "error": true,
  "message": "Snapshot not found: example.com_1704067200.zip"
}
```

---

## Legacy Endpoint

The legacy snapshot download endpoint is still available for backward compatibility:

```bash
curl http://localhost:8002/osm/api/snapshot-download/example.com \
  -H "Authorization: Bearer $TOKEN" \
  --output snapshot.zip
```

---

## Data Source Values

When a workspace is imported, its `data_source` field is set to indicate how it was created:

| Value | Description |
|-------|-------------|
| `local` | Created locally via scan (default) |
| `cloud` | Synced from cloud storage |
| `imported` | Imported from snapshot file |

---

## Security Considerations

**Warning:** Only import snapshots from trusted sources!

Imported workspace data may contain:
- Database records that could conflict with existing data
- File paths that reference external resources
- Configuration that may not be compatible

The imported workspace database state may be unstable. Use the `skip_db=true` parameter if you only need the files without database import.

# Assets

## List Assets

Get a paginated list of assets with optional workspace filtering.

**List all assets:**
```bash
curl http://localhost:8002/osm/api/assets \
  -H "Authorization: Bearer $TOKEN"
```

**List assets with pagination:**
```bash
curl "http://localhost:8002/osm/api/assets?offset=0&limit=100" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by workspace:**
```bash
curl "http://localhost:8002/osm/api/assets?workspace=example.com" \
  -H "Authorization: Bearer $TOKEN"
```

**Combine workspace filter with pagination:**
```bash
curl "http://localhost:8002/osm/api/assets?workspace=example.com&offset=50&limit=25" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "workspace": "example.com",
      "asset_value": "api.example.com",
      "url": "https://api.example.com",
      "input": "api.example.com",
      "scheme": "https",
      "method": "GET",
      "path": "/",
      "status_code": 200,
      "content_type": "application/json",
      "content_length": 4523,
      "title": "API Documentation",
      "words": 523,
      "lines": 89,
      "host_ip": "93.184.216.34",
      "a": ["93.184.216.34", "93.184.216.35"],
      "tls": "TLS 1.3",
      "asset_type": "web",
      "tech": ["nginx/1.21.0", "nodejs", "express"],
      "time": "245ms",
      "remarks": "production",
      "source": "httpx",
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "workspace": "example.com",
      "asset_value": "admin.example.com",
      "url": "https://admin.example.com",
      "input": "admin.example.com",
      "scheme": "https",
      "method": "GET",
      "path": "/login",
      "status_code": 401,
      "content_type": "text/html",
      "content_length": 2156,
      "title": "Admin Login - Example Corp",
      "words": 156,
      "lines": 45,
      "host_ip": "93.184.216.36",
      "a": ["93.184.216.36"],
      "tls": "TLS 1.2",
      "asset_type": "web",
      "tech": ["nginx/1.20.0", "php/8.1", "wordpress"],
      "time": "312ms",
      "remarks": "admin-panel",
      "source": "httpx",
      "created_at": "2025-01-15T10:31:00Z",
      "updated_at": "2025-01-15T10:31:00Z"
    }
  ],
  "pagination": {
    "total": 500,
    "offset": 0,
    "limit": 20
  }
}
```

**Asset Fields Reference:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Unique asset identifier |
| `workspace` | string | Workspace/scan target name |
| `asset_value` | string | Primary asset identifier (hostname/subdomain) |
| `url` | string | Full URL of the asset |
| `input` | string | Original input value |
| `scheme` | string | Protocol scheme (http, https) |
| `method` | string | HTTP method used |
| `path` | string | URL path |
| `status_code` | int | HTTP response status code |
| `content_type` | string | Response content type |
| `content_length` | int | Response body size in bytes |
| `title` | string | HTML page title |
| `words` | int | Word count in response |
| `lines` | int | Line count in response |
| `host_ip` | string | Resolved IP address |
| `a` | array | DNS A records |
| `tls` | string | TLS version information |
| `asset_type` | string | Asset type classification |
| `tech` | array | Detected technologies |
| `time` | string | Response time |
| `remarks` | string | Custom labels/remarks |
| `source` | string | Discovery source (httpx, nuclei, etc.) |
| `created_at` | timestamp | Creation timestamp |
| `updated_at` | timestamp | Last update timestamp |

---

## List Asset Diff Snapshots

Get a paginated list of stored asset diff snapshots. These snapshots capture changes in assets over time.

**List all asset diff snapshots:**
```bash
curl http://localhost:8002/osm/api/assets/diffs \
  -H "Authorization: Bearer $TOKEN"
```

**List with pagination:**
```bash
curl "http://localhost:8002/osm/api/assets/diffs?offset=0&limit=50" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by workspace:**
```bash
curl "http://localhost:8002/osm/api/assets/diffs?workspace=example.com" \
  -H "Authorization: Bearer $TOKEN"
```

**Combine workspace filter with pagination:**
```bash
curl "http://localhost:8002/osm/api/assets/diffs?workspace=example.com&offset=0&limit=25" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "workspace_name": "example.com",
      "from_time": "2025-01-14T00:00:00Z",
      "to_time": "2025-01-15T00:00:00Z",
      "total_added": 15,
      "total_removed": 3,
      "total_changed": 7,
      "diff_data": "{\"added\":[...],\"removed\":[...],\"changed\":[...]}",
      "created_at": "2025-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "workspace_name": "example.com",
      "from_time": "2025-01-15T00:00:00Z",
      "to_time": "2025-01-16T00:00:00Z",
      "total_added": 8,
      "total_removed": 1,
      "total_changed": 12,
      "diff_data": "{\"added\":[...],\"removed\":[...],\"changed\":[...]}",
      "created_at": "2025-01-16T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 30,
    "offset": 0,
    "limit": 20
  }
}
```

**Asset Diff Snapshot Fields Reference:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Unique snapshot identifier |
| `workspace_name` | string | Workspace name for this diff |
| `from_time` | timestamp | Start time of the diff period |
| `to_time` | timestamp | End time of the diff period |
| `total_added` | int | Number of new assets added |
| `total_removed` | int | Number of assets removed |
| `total_changed` | int | Number of assets that changed |
| `diff_data` | string | JSON serialized diff data containing added, removed, and changed assets |
| `created_at` | timestamp | When the snapshot was created |

# Settings

Manage server configuration settings.

## Get YAML Configuration

Get the entire YAML configuration file with sensitive fields redacted. Fields containing `_key`, `secret`, `password`, `username`, or `_token` are replaced with `[REDACTED]`.

```bash
curl http://localhost:8002/osm/api/settings/yaml \
  -H "Authorization: Bearer $TOKEN"
```

**Response:** (text/yaml)
```yaml
# =============================================================================
# Osmedeus Configuration File
# =============================================================================
base_folder: ~/osmedeus-base

environments:
  binaries_path: "{{base_folder}}/binaries"
  # ... more config

server:
  host: "0.0.0.0"
  port: 8002
  workspace_prefix_key: "[REDACTED]"
  simple_user_map_key: "[REDACTED]"
  jwt:
    secret_signing_key: "[REDACTED]"
    expiration_minutes: 180

database:
  host: ""
  port: 5432
  username: "[REDACTED]"
  password: "[REDACTED]"
  # ... more config
```

---

## Update YAML Configuration

Replace the entire YAML configuration file with new content. A backup of the existing configuration is created before overwriting.

```bash
curl -X PUT http://localhost:8002/osm/api/settings/yaml \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/yaml" \
  --data-binary @new-config.yaml
```

**Request Body:** Raw YAML configuration content

**Response:**
```json
{
  "message": "Configuration updated successfully",
  "path": "/home/user/osmedeus-base/osm-settings.yaml",
  "backup": "/home/user/osmedeus-base/osm-settings.yaml.backup"
}
```

**Error Response (Invalid YAML):**
```json
{
  "error": true,
  "message": "Invalid YAML configuration: yaml: unmarshal errors: ..."
}
```

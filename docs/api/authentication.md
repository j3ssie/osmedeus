# Authentication

Most API endpoints require JWT authentication. First, obtain a token via the login endpoint, then include it in subsequent requests.

## Login

**POST** `/osm/api/login`

Authenticate and obtain a JWT token.

### Request

```bash
curl -X POST http://localhost:8002/osm/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "osmedeus",
    "password": "your-password"
  }'
```

### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | Yes | Username configured in server settings |
| `password` | string | Yes | Password for the user |

### Response (200 OK)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6Im9zbWVkZXVzIiwiZXhwIjoxNzA0MDY3MjAwLCJpYXQiOjE3MDQwNjM2MDB9.abc123..."
}
```

### Error Responses

**400 Bad Request** - Invalid request body:
```json
{
  "error": true,
  "message": "Invalid request body"
}
```

**401 Unauthorized** - Invalid credentials:
```json
{
  "error": true,
  "message": "Invalid credentials"
}
```

## Token Details

- **Algorithm**: HS256 (HMAC-SHA256)
- **Expiration**: Configurable via `server.jwt.expiration_minutes` in settings (default: 1440 minutes / 1 day)
- **Claims**: Contains `username`, `exp` (expiration), and `iat` (issued at)

## Using the Token

Include the token in subsequent requests using the `Authorization: Bearer <token>` header:

```bash
# Store token in environment variable
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Use in API requests
curl http://localhost:8002/osm/api/workflows \
  -H "Authorization: Bearer $TOKEN"
```

### Authentication Errors

**401 Unauthorized** - Missing header:
```json
{
  "error": true,
  "message": "Missing authorization header"
}
```

**401 Unauthorized** - Invalid format:
```json
{
  "error": true,
  "message": "Invalid authorization header format"
}
```

**401 Unauthorized** - Expired or invalid token:
```json
{
  "error": true,
  "message": "Invalid or expired token"
}
```

## API Key Authentication

As an alternative to JWT tokens, you can authenticate using a static API key via the `x-osm-api-key` header. This is useful for scripts, CI/CD pipelines, or integrations where managing JWT token refresh is impractical.

### Configuration

API key authentication is configured in `~/osmedeus-base/osm-settings.yaml`:

```yaml
server:
  # Enable API key authentication (default: true)
  enabled_auth_api: true
  # API key for x-osm-api-key header authentication
  # A random 32-character key is generated on first run
  auth_api_key: "your-api-key-here"
```

### Using the API Key

Include the API key in requests using the `x-osm-api-key` header:

```bash
# Store API key in environment variable
export OSM_API_KEY="your-api-key-here"

# Use in API requests
curl http://localhost:8002/osm/api/workflows \
  -H "x-osm-api-key: $OSM_API_KEY"
```

### Error Response

**401 Unauthorized** - Invalid or missing API key:
```json
{
  "error": true,
  "message": "Invalid or missing API key"
}
```

### Notes

- API key authentication takes priority over JWT when enabled
- A random 32-character API key is automatically generated on first server start
- The API key is stored in plain text in the settings file; ensure appropriate file permissions
- Empty, whitespace-only, or placeholder values (`null`, `undefined`, `nil`) are rejected

## Disabling Authentication

Authentication can be disabled by starting the server with the `--no-auth` flag:

```bash
osmedeus server --no-auth
```

When disabled, all API endpoints are accessible without a token.

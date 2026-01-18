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
- **Expiration**: Configurable via `server.jwt.expiration_minutes` in settings (default: 60 minutes)
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

## Disabling Authentication

Authentication can be disabled by starting the server with the `--no-auth` flag:

```bash
osmedeus server --no-auth
```

When disabled, all API endpoints are accessible without a token.

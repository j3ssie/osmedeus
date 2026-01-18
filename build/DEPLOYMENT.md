# Deployment Guide

This guide covers building, deploying, and running Osmedeus in various environments.

## Prerequisites

- Go 1.21+ (for local builds)
- Docker 20.10+ (for containerized deployment)
- Docker Compose 2.0+ (for distributed mode)

## Quick Start

```bash
# Local build and run
make build
./build/bin/osmedeus serve

# Docker single container
docker build -t osmedeus:latest -f build/docker/Dockerfile .
docker run -p 8001:8001 osmedeus:latest

# Distributed mode with Docker Compose
docker-compose -f build/docker/docker-compose.yml up -d
```

## Building

### Local Build

```bash
# Build for current platform
make build

# Cross-platform builds
make build-all       # All platforms
make build-linux     # Linux amd64
make build-darwin    # macOS amd64 + arm64
make build-windows   # Windows amd64

# Output location
./build/bin/osmedeus
```

### Docker Build

```bash
docker build -t osmedeus:latest -f build/docker/Dockerfile .

# Development image (with hot-reload)
docker build -t osmedeus:dev -f build/docker/Dockerfile.dev .

# With custom version
docker build --build-arg VERSION=5.1.0 -t osmedeus:5.1.0 -f build/docker/Dockerfile .
```

## Deployment Modes

### Single Host

#### Direct Binary

```bash
# Run server
./build/bin/osmedeus serve --port 8001

# Run with authentication disabled (development only)
./build/bin/osmedeus serve -A

# Run a scan
./build/bin/osmedeus scan -f general -t example.com
```

#### Docker Container

```bash
# Basic server
docker run -d \
  --name osmedeus \
  -p 8001:8001 \
  -v osmedeus-data:/root/osmedeus-base \
  -v workspaces:/root/workspaces-osmedeus \
  osmedeus:latest

# With custom workflows
docker run -d \
  --name osmedeus \
  -p 8001:8001 \
  -v /path/to/workflows:/root/osmedeus-base/workflows \
  -v /path/to/workspaces:/root/workspaces-osmedeus \
  osmedeus:latest
```

### Distributed Mode (Master/Worker)

Distributed mode allows scaling scan workloads across multiple worker nodes using Redis as a message queue.

#### Architecture

```
                    ┌─────────────┐
                    │   Client    │
                    └──────┬──────┘
                           │ REST API
                    ┌──────▼──────┐
                    │   Master    │
                    │  (Server)   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │    Redis    │
                    │   (Queue)   │
                    └──────┬──────┘
              ┌────────────┼────────────┐
              │            │            │
        ┌─────▼────┐ ┌─────▼────┐ ┌─────▼────┐
        │ Worker 1 │ │ Worker 2 │ │ Worker N │
        └──────────┘ └──────────┘ └──────────┘
```

#### Docker Compose Setup

```bash
# Start with 2 workers (default)
docker-compose -f build/docker/docker-compose.yml up -d

# Scale to 5 workers
docker-compose -f build/docker/docker-compose.yml up -d --scale worker=5

# View logs
docker-compose -f build/docker/docker-compose.yml logs -f

# Stop all services
docker-compose -f build/docker/docker-compose.yml down

# Stop and remove volumes
docker-compose -f build/docker/docker-compose.yml down -v
```

#### Manual Distributed Setup

If not using Docker Compose:

```bash
# 1. Start Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# 2. Start Master
./build/bin/osmedeus serve --master --port 8001

# 3. Start Workers (on same or different machines)
./build/bin/osmedeus worker join --redis-url redis://localhost:6379
```

#### Submitting Distributed Scans

```bash
# Submit scan to distributed queue
./build/bin/osmedeus scan -f general -t example.com -D

# With custom Redis URL
./build/bin/osmedeus scan -f general -t example.com -D --redis-url redis://redis-host:6379

# Check worker status
./build/bin/osmedeus worker status
```

## Configuration

### Configuration File

Default location: `~/osmedeus-base/osm-settings.yaml`

```yaml
base_folder: ~/osmedeus-base

environments:
  binaries_path: "{{base_folder}}/binaries"
  data: "{{base_folder}}/data"
  workspaces: ~/workspaces-osmedeus
  workflows: "{{base_folder}}/workflows"

server:
  host: 0.0.0.0
  port: 8001

# Required for distributed mode
redis:
  host: localhost
  port: 6379
  password: ""    # Optional
  db: 0

database:
  db_engine: sqlite    # or postgresql
  db_path: "{{base_folder}}/osm-data.db"

client:
  username: admin
  password: admin
  jwt:
    secret: "change-this-in-production"
    expiration_minutes: 60

scan_tactic:
  aggressive: 40
  default: 10
  gently: 5
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis hostname | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `OSM_BASE_FOLDER` | Base folder path | ~/osmedeus-base |

### Command Line Overrides

```bash
# Override base folder
osmedeus -b /custom/path scan -f general -t example.com

# Override workflow folder
osmedeus -F /custom/workflows workflow list

# Override Redis URL (distributed mode)
osmedeus scan -f general -t example.com -D --redis-url redis://user:pass@host:6379/0
```

## Docker Compose Reference

The included `build/docker/docker-compose.yml` provides a complete distributed setup:

### Services

| Service | Purpose | Ports |
|---------|---------|-------|
| `redis` | Task queue and coordination | 6379 |
| `master` | API server and task distributor | 8001 |
| `worker` | Task executor (scalable) | - |

### Volumes

| Volume | Purpose |
|--------|---------|
| `redis-data` | Redis persistence |
| `osmedeus-data` | Workflows and configuration |
| `workspaces` | Scan output data |

### Scaling

```bash
# Scale workers dynamically
docker-compose -f build/docker/docker-compose.yml up -d --scale worker=10

# View running containers
docker-compose -f build/docker/docker-compose.yml ps
```

## Production Considerations

### Security

1. **Authentication**: Never use `-A` (no-auth) in production
2. **JWT Secret**: Change the default JWT secret in config
3. **TLS**: Use a reverse proxy (nginx, traefik) for HTTPS
4. **Network**: Restrict Redis access to internal network only

```yaml
# Example: Secure JWT configuration
client:
  jwt:
    secret: "your-256-bit-secret-key-here"
    expiration_minutes: 30
```

### Resource Limits

Worker resource limits in docker-compose.yml:

```yaml
deploy:
  resources:
    limits:
      cpus: '1'
      memory: 1G
    reservations:
      cpus: '0.5'
      memory: 512M
```

Adjust based on workflow requirements.

### Health Checks

The Docker image includes built-in health checks:

```bash
# Check master health
curl http://localhost:8001/health

# Check readiness
curl http://localhost:8001/health/ready
```

### Logging

```bash
# View master logs
docker logs osmedeus-master -f

# View all worker logs
docker-compose -f build/docker/docker-compose.yml logs -f worker

# Log levels are controlled by --verbose/-v flag
./build/bin/osmedeus -v serve
```

### Database Options

For production, consider PostgreSQL instead of SQLite:

```yaml
database:
  db_engine: postgresql
  db_host: postgres-host
  db_port: 5432
  db_name: osmedeus
  db_user: osmedeus
  db_password: secure-password
```

### Backup

```bash
# Backup volumes
docker run --rm \
  -v osmedeus-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/osmedeus-backup.tar.gz /data

# Backup workspaces
docker run --rm \
  -v workspaces:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/workspaces-backup.tar.gz /data
```

## Troubleshooting

### Common Issues

**Workers not connecting:**
```bash
# Check Redis connectivity
docker exec osmedeus-redis redis-cli ping

# Check worker logs
docker-compose logs worker
```

**Scans not executing:**
```bash
# Verify workflow exists
./build/bin/osmedeus workflow list

# Check master logs
docker logs osmedeus-master
```

**Port conflicts:**
```bash
# Use different ports
docker run -p 8080:8001 osmedeus:latest
```

### Useful Commands

```bash
# Environment health check
./build/bin/osmedeus health

# Validate workflows
./build/bin/osmedeus workflow validate <workflow-name>

# Test workflow (dry-run)
./build/bin/osmedeus scan -f general -t example.com --dry-run
```

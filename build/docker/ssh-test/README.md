# SSH Test Container

Ubuntu 22.04 container with SSH server for testing osmedeus SSH runner.

## Quick Start

```bash
# Build and start
docker-compose up -d --build

# Test SSH connection
ssh -i id_ed25519 -p 2222 testuser@localhost

# Stop
docker-compose down
```

## Credentials

- **User**: `testuser`
- **Password**: `testpass`
- **SSH Key**: `id_ed25519` (in this directory)
- **Port**: `2222`

## SSH Command Examples

```bash
# Using key authentication
ssh -i id_ed25519 -p 2222 -o StrictHostKeyChecking=no testuser@localhost

# Run command
ssh -i id_ed25519 -p 2222 -o StrictHostKeyChecking=no testuser@localhost "whoami"
```

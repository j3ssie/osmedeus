#!/usr/bin/env bash
# Canary test entrypoint: starts API server in background and keeps container alive.
# Tests interact via:
#   docker exec osm-canary osmedeus run ...
#   curl http://localhost:8002/osm/api/...

set -euo pipefail

echo "[canary] Starting osmedeus API server on :8002 ..."
osmedeus serve --debug --port 8002 --host 0.0.0.0 -A &
SERVER_PID=$!

# Give the server a moment to bind
sleep 2

echo "[canary] API server started (PID ${SERVER_PID})"
echo "[canary] Container is ready — waiting for test commands."

# Keep the container alive; forward SIGTERM to the server
trap "kill ${SERVER_PID} 2>/dev/null; exit 0" SIGTERM SIGINT

wait ${SERVER_PID}

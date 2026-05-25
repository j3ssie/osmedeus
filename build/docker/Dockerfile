# Osmedeus Production Dockerfile
# Multi-stage build: compile from source, then create minimal runtime image

# ── Stage 1: build from source ──────────────────────────────────────────────
# Pin the builder to the native build platform so the Go compiler runs without
# emulation, then cross-compile to the requested target arch (amd64/arm64).
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder

ARG BUILD_TIME
ARG COMMIT_HASH
# Provided automatically by buildx for multi-platform builds.
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}" \
    -o /osmedeus ./cmd/osmedeus

# ── Stage 2: runtime ────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# Install essential tools (retry on transient network failures for large packages like chromium)
RUN apt-get update && \
    for i in 1 2 3; do \
        apt-get install -y --no-install-recommends \
            build-essential \
            git \
            curl \
            wget \
            ca-certificates \
            python3 \
            python3-pip \
            chromium \
        && break || { echo "Attempt $i failed, retrying..."; apt-get update; }; \
    done && \
    rm -rf /var/lib/apt/lists/* && \
    ln -sf /usr/bin/python3 /usr/bin/python

# Create base directories
RUN mkdir -p /root/osmedeus-base /root/workspaces-osmedeus

WORKDIR /root

# Copy binary from builder
COPY --from=builder /osmedeus /usr/local/bin/osmedeus

# Initialize osmedeus base folder with preset workflows
RUN osmedeus install base --preset

# Install SAST tools (used for testing). Retry on transient download/hash
# failures — large wheel/binary downloads are prone to truncation when this
# stage runs under QEMU emulation during multi-arch builds (a corrupted wheel
# trips pip's hash verification).
RUN for i in 1 2 3; do \
        osmedeus install binary --name trivy --name semgrep --name kingfisher --name bearer && break; \
        echo "Attempt $i failed, clearing pip cache and retrying..."; \
        pip cache purge 2>/dev/null || true; \
        sleep 5; \
        [ "$i" = "3" ] && exit 1; \
    done

# Set up PATH for external binaries
ENV PATH="/root/osmedeus-base/external-binaries:${PATH}"

# Expose default server port
EXPOSE 8002

# Default entrypoint - exposes osmedeus CLI only
ENTRYPOINT ["osmedeus"]

# Default command shows help (user can override with run/server/etc.)
CMD ["--help"]

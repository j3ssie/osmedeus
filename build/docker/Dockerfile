# Osmedeus Production Dockerfile
# Ubuntu/Debian-based image with essential tools for security scanning

# Stage 1: Build osmedeus binary
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=5.0.0
ARG BUILD_TIME
ARG COMMIT_HASH
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}" \
    -o /app/bin/osmedeus ./cmd/osmedeus

# Stage 2: Runtime with essential tools
FROM golang:1.25-bookworm

# Install essential tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    curl \
    wget \
    ca-certificates \
    python3 \
    python3-pip \
    chromium \
    && rm -rf /var/lib/apt/lists/* \
    && ln -sf /usr/bin/python3 /usr/bin/python

# Copy osmedeus binary from builder
COPY --from=builder /app/bin/osmedeus /usr/local/bin/osmedeus

# Create base directories
RUN mkdir -p /root/osmedeus-base /root/workspaces-osmedeus

WORKDIR /root

# Initialize osmedeus base folder with preset workflows
RUN osmedeus install base --preset

# Set up PATH for external binaries
ENV PATH="/root/osmedeus-base/external-binaries:${PATH}"

# Expose default server port
EXPOSE 8002

# Default entrypoint - exposes osmedeus CLI only
ENTRYPOINT ["osmedeus"]

# Default command shows help (user can override with run/server/etc.)
CMD ["--help"]

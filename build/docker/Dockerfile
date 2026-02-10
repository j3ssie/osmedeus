# Osmedeus Production Dockerfile
# Ubuntu/Debian-based image with essential tools for security scanning

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

# Install osmedeus via official install script
RUN curl -fsSL https://www.osmedeus.org/install.sh | bash

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

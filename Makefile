.PHONY: build run test test-unit test-integration test-workflow-integration test-e2e test-e2e-verbose test-e2e-ssh test-e2e-api test-e2e-nix test-e2e-install test-e2e-cloud test-sudo test-cloud test-docker test-ssh test-distributed distributed-e2e-up distributed-e2e-run distributed-e2e-down test-canary-all test-canary-repo test-canary-domain test-canary-ip test-canary-general canary-up canary-down test-all test-summary test-ci clean install install-gotestsum lint fmt db-seed db-clean db-migrate run-server-debug swagger update-ui sync-skills sync-platform snapshot-release github-release bump-version npm-binaries npm-build npm-pack npm-publish run-github-action docker-toolbox docker-toolbox-run docker-toolbox-shell docker-publish docker-buildx-setup

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod
BINARY_NAME=osmedeus
BINARY_DIR=build/bin

# Console output prefix (cyan color)
PREFIX=\033[36m[*]\033[0m

# Gotestsum configuration - check GOPATH/bin first, then use go test fallback
GOPATH_BIN=$(shell go env GOPATH)/bin
GOTESTSUM_PATH=$(shell command -v gotestsum 2>/dev/null || echo $(GOPATH_BIN)/gotestsum)
GOTESTSUM_EXISTS=$(shell test -x $(GOTESTSUM_PATH) && echo yes || echo no)

# GOBIN for install target (falls back to GOPATH/bin if GOBIN is not set)
GOBIN_PATH=$(shell go env GOBIN)
ifeq ($(GOBIN_PATH),)
    GOBIN_PATH=$(GOPATH_BIN)
endif

ifeq ($(GOTESTSUM_EXISTS),yes)
    TESTCMD=@$(GOTESTSUM_PATH)
    TESTFLAGS=--format testdox --format-hide-empty-pkg --hide-summary=skipped,output --
    CANARY_TESTFLAGS=--format standard-verbose -- -v
else
    TESTCMD=$(GOTEST)
    TESTFLAGS=-v
    CANARY_TESTFLAGS=-v
endif

# Build flags
VERSION=$(shell cat internal/core/constants.go | grep 'VERSION =' | cut -d '"' -f 2)
AUTHOR=$(shell cat internal/core/constants.go | grep 'AUTHOR =' | cut -d '"' -f 2)
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"

# Default target
all: build

# Build the application and install to GOBIN
build:
	@echo "$(PREFIX) Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/osmedeus
	@echo "$(PREFIX) Installing $(BINARY_NAME) to $(GOBIN_PATH)..."
	@mkdir -p $(GOBIN_PATH)
	@rm -f $(GOBIN_PATH)/$(BINARY_NAME)
	@cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/$(BINARY_NAME)

# Install to GOBIN (or GOPATH/bin) - requires prior build
install:
	@echo "$(PREFIX) Installing $(BINARY_NAME) to $(GOBIN_PATH)..."
	@if [ ! -f "$(BINARY_DIR)/$(BINARY_NAME)" ]; then \
		echo "$(PREFIX) Binary not found, building first..."; \
		$(MAKE) build; \
	else \
		mkdir -p $(GOBIN_PATH) && rm -f $(GOBIN_PATH)/$(BINARY_NAME) && cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/$(BINARY_NAME); \
	fi

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "$(PREFIX) Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/osmedeus

build-darwin:
	@echo "$(PREFIX) Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/osmedeus
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/osmedeus

build-windows:
	@echo "$(PREFIX) Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/osmedeus

# Run the application
run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/osmedeus
	./$(BINARY_DIR)/$(BINARY_NAME)

# Run with specific command
run-server: build
	@echo "$(PREFIX) Starting server..."
	./$(BINARY_DIR)/$(BINARY_NAME) serve

# Run server in debug mode without authentication
run-server-debug: build
	@echo "$(PREFIX) Starting debug server (no auth)..."
	./$(BINARY_DIR)/$(BINARY_NAME) serve -A --debug

# Install gotestsum (idempotent - silent if already installed)
install-gotestsum:
	@if [ ! -x "$(GOPATH_BIN)/gotestsum" ]; then \
		echo "Installing gotestsum..."; \
		go install gotest.tools/gotestsum@latest; \
	fi

# Run tests (install gotestsum first)
test: install-gotestsum
	$(TESTCMD) $(TESTFLAGS) -race ./...

# Run tests with coverage
test-coverage: install-gotestsum
	$(TESTCMD) $(TESTFLAGS) -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Unit tests (fast, no external dependencies)
test-unit: install-gotestsum
	$(TESTCMD) $(TESTFLAGS) -short ./...

# Integration tests (requires Docker for some tests)
test-integration: install-gotestsum
	$(TESTCMD) $(TESTFLAGS) -run Integration ./...

# Workflow integration tests (test/integration/)
test-workflow-integration: install-gotestsum
	$(TESTCMD) $(TESTFLAGS) ./test/integration/...

# E2E CLI tests (requires binary to be built first)
test-e2e: build install-gotestsum
	$(TESTCMD) $(TESTFLAGS) ./test/e2e/...

# E2E CLI tests with verbose output (for debugging)
test-e2e-verbose: build install-gotestsum
	@$(GOPATH_BIN)/gotestsum --format standard-verbose -- -v ./test/e2e/...

# Docker runner tests
test-docker: install-gotestsum
	docker-compose -f docker-compose.test.yaml up -d
	$(TESTCMD) $(TESTFLAGS) -run Docker ./internal/runner/...
	docker-compose -f docker-compose.test.yaml down

# SSH runner tests (using linuxserver/openssh-server)
test-ssh: install-gotestsum
	docker-compose -f build/docker/docker-compose.test.yaml up -d ssh-server
	sleep 5
	$(TESTCMD) $(TESTFLAGS) -run SSH ./internal/runner/...
	docker-compose -f build/docker/docker-compose.test.yaml down

# SSH E2E tests (full workflow tests with SSH runner)
test-e2e-ssh: build install-gotestsum
	@echo "$(PREFIX) Starting SSH server for E2E tests..."
	docker-compose -f build/docker/docker-compose.test.yaml up -d ssh-server
	@echo "$(PREFIX) Waiting for SSH server to be ready..."
	@sleep 5
	@echo "$(PREFIX) Running SSH E2E tests..."
	$(TESTCMD) $(TESTFLAGS) -run SSH ./test/e2e/...
	@echo "$(PREFIX) Cleaning up..."
	docker-compose -f build/docker/docker-compose.test.yaml down -v

# Cloud setup E2E tests (SSH password, key, ansible, post-commands)
test-e2e-cloud-setup: build install-gotestsum
	@echo "$(PREFIX) Starting SSH server for cloud setup tests..."
	docker-compose -f build/docker/docker-compose.test.yaml up -d ssh-server
	@sleep 5
	@echo "$(PREFIX) Running cloud setup E2E tests..."
	$(TESTCMD) $(TESTFLAGS) -run TestCloudSetup ./test/e2e/...
	@echo "$(PREFIX) Cleaning up..."
	docker-compose -f build/docker/docker-compose.test.yaml down -v

# Distributed scan e2e tests (requires Docker for Redis)
test-distributed: build install-gotestsum
	@echo "$(PREFIX) Starting Redis for distributed tests..."
	docker-compose -f build/docker/docker-compose.distributed-test.yaml up -d
	@echo "$(PREFIX) Waiting for Redis to be ready..."
	@sleep 3
	@echo "$(PREFIX) Running distributed tests..."
	$(TESTCMD) $(TESTFLAGS) -run Distributed ./test/e2e/...
	@echo "$(PREFIX) Cleaning up..."
	docker-compose -f build/docker/docker-compose.distributed-test.yaml down -v

# Distributed E2E stack: redis + master + worker in Docker, then submit a real scan
distributed-e2e-up:
	@echo "$(PREFIX) Building distributed E2E stack..."
	docker-compose -f build/docker/docker-compose.distributed-e2e.yaml build
	@echo "$(PREFIX) Starting distributed E2E stack (redis + master + worker)..."
	docker-compose -f build/docker/docker-compose.distributed-e2e.yaml up -d
	@echo "$(PREFIX) Waiting for master to be healthy..."
	@for i in $$(seq 1 30); do \
		curl -sf http://localhost:8002/health > /dev/null 2>&1 && break; \
		sleep 2; \
	done
	@echo "$(PREFIX) Stack is ready. Master: http://localhost:8002"

distributed-e2e-run:
	@echo "$(PREFIX) Submitting distributed scan from master container..."
	docker exec osm-e2e-master osmedeus run -f repo -D -t https://github.com/juice-shop/juice-shop
	@echo "$(PREFIX) Scan submitted. Tailing worker logs (Ctrl+C to stop)..."
	docker-compose -f build/docker/docker-compose.distributed-e2e.yaml logs -f worker

distributed-e2e-down:
	@echo "$(PREFIX) Stopping distributed E2E stack..."
	docker-compose -f build/docker/docker-compose.distributed-e2e.yaml down -v

# API E2E tests (requires Docker for Redis, builds binary first)
test-e2e-api: build install-gotestsum
	@echo "$(PREFIX) Starting Redis for API tests..."
	docker-compose -f build/docker/docker-compose.distributed-test.yaml up -d
	@echo "$(PREFIX) Waiting for Redis to be ready..."
	@sleep 3
	@echo "$(PREFIX) Running API E2E tests..."
	$(TESTCMD) $(TESTFLAGS) -run API ./test/e2e/...
	@echo "$(PREFIX) Cleaning up..."
	docker-compose -f build/docker/docker-compose.distributed-test.yaml down -v

# Nix E2E tests (requires Docker for Nix container)
test-e2e-nix: build install-gotestsum
	@echo "$(PREFIX) Building Nix test container..."
	docker-compose -f build/docker/docker-compose.nix-test.yaml build
	@echo "$(PREFIX) Starting Nix test container..."
	docker-compose -f build/docker/docker-compose.nix-test.yaml up -d
	@echo "$(PREFIX) Waiting for Nix container to be ready..."
	@sleep 3
	@echo "$(PREFIX) Running Nix E2E tests..."
	$(TESTCMD) $(TESTFLAGS) -run TestNix ./test/e2e/...
	@echo "$(PREFIX) Cleaning up..."
	docker-compose -f build/docker/docker-compose.nix-test.yaml down -v

# Install E2E tests (workflow and base installation from zip/URL/git)
test-e2e-install: build install-gotestsum
	@echo "$(PREFIX) Running install E2E tests..."
	$(TESTCMD) $(TESTFLAGS) -run TestInstall ./test/e2e/...

# Cloud E2E tests (cloud CLI commands and workflow)
test-e2e-cloud: build install-gotestsum
	@echo "$(PREFIX) Running cloud E2E tests..."
	$(TESTCMD) $(TESTFLAGS) -run TestCloud ./test/e2e/...

# Sudo-aware tests (requires interactive sudo prompt)
test-sudo: export OSM_TEST_SUDO=1
test-sudo: build install-gotestsum
	@echo "$(PREFIX) Running sudo-aware tests (may prompt for password)..."
	$(TESTCMD) $(TESTFLAGS) -run TestSudo ./test/e2e/...

# Cloud integration tests (internal cloud package tests)
test-cloud: install-gotestsum
	@echo "$(PREFIX) Running cloud integration tests..."
	$(TESTCMD) $(TESTFLAGS) ./test/integration/cloud_integration_test.go

# ── Canary tests (real scans inside Docker toolbox, requires Docker) ──────────

# Build and start the canary container (shared setup for individual targets)
canary-up: install-gotestsum
	@echo "$(PREFIX) Cleaning up any existing canary container..."
	-docker-compose -f build/docker/docker-compose.canary.yaml down -v 2>/dev/null
	@echo "$(PREFIX) Building canary Docker image..."
	docker-compose -f build/docker/docker-compose.canary.yaml build
	@echo "$(PREFIX) Starting canary container..."
	docker-compose -f build/docker/docker-compose.canary.yaml up -d
	@echo "$(PREFIX) Waiting for API server..."
	@for i in $$(seq 1 60); do curl -sf http://localhost:8002/health > /dev/null 2>&1 && break || sleep 2; done
	@echo "$(PREFIX) Canary container ready."

# Tear down the canary container
canary-down:
	@echo "$(PREFIX) Cleaning up canary container..."
	docker-compose -f build/docker/docker-compose.canary.yaml down -v

# Run ALL canary scans (builds container, runs all 4, cleans up — 60-120min)
test-canary-all: canary-up
	@echo "$(PREFIX) Running all canary tests (60-120 minutes)..."
	$(TESTCMD) $(CANARY_TESTFLAGS) -run TestCanary_FullSuite -timeout 120m ./test/e2e/... || ($(MAKE) canary-down && exit 1)
	@$(MAKE) canary-down

# Repo scan canary (juice-shop SAST, ~25min)
test-canary-repo: canary-up
	@echo "$(PREFIX) Running repo scan canary test..."
	$(TESTCMD) $(CANARY_TESTFLAGS) -run TestCanary_Repo -timeout 30m ./test/e2e/... || ($(MAKE) canary-down && exit 1)
	@$(MAKE) canary-down

# Domain-lite scan canary (hackerone.com, ~20min)
test-canary-domain: canary-up
	@echo "$(PREFIX) Running domain-lite scan canary test..."
	$(TESTCMD) $(CANARY_TESTFLAGS) -run TestCanary_Domain -timeout 25m ./test/e2e/... || ($(MAKE) canary-down && exit 1)
	@$(MAKE) canary-down

# CIDR scan canary (IP list, ~25min)
test-canary-ip: canary-up
	@echo "$(PREFIX) Running CIDR scan canary test..."
	$(TESTCMD) $(CANARY_TESTFLAGS) -run TestCanary_CIDR -timeout 30m ./test/e2e/... || ($(MAKE) canary-down && exit 1)
	@$(MAKE) canary-down

# Domain-list-recon scan canary (hackerone.com subdomains, ~40min)
test-canary-general: canary-up
	@echo "$(PREFIX) Running general scan canary test..."
	$(TESTCMD) $(CANARY_TESTFLAGS) -run TestCanary_General -timeout 45m ./test/e2e/... || ($(MAKE) canary-down && exit 1)
	@$(MAKE) canary-down

# All tests
test-all: test-unit test-integration

# Quick test summary (pass/fail only)
test-summary: install-gotestsum
	@$(GOPATH_BIN)/gotestsum --format dots-v2 -- -v ./...

# Test with JUnit XML output (for CI)
test-ci: install-gotestsum
	@$(GOPATH_BIN)/gotestsum --junitfile test-results.xml --format testdox --format-hide-empty-pkg --hide-summary=skipped,output -- -v -race ./...

# Clean build artifacts
clean:
	@echo "$(PREFIX) Cleaning..."
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html test-results.xml

# Format code
fmt:
	$(GOFMT) ./...

# Lint code
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Download dependencies
deps:
	$(GOMOD) download

# Update dependencies
update-deps:
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Generate code (if needed)
generate:
	$(GOCMD) generate ./...

# Generate swagger documentation
swagger:
	@echo "$(PREFIX) Generating swagger documentation..."
	swag init -g pkg/server/server.go -o docs/api-swagger/ --packageName apiswagger

# Build the dashboard from platform/ and refresh the embedded UI.
#
# platform/osmedeus-dashboard/ is the source of truth; its build output is
# gitignored, so this builds before copying. Set DASHBOARD_SKIP_BUILD=1 to reuse
# an existing build (e.g. when the tree has not changed).
DASHBOARD_DIR=platform/osmedeus-dashboard
update-ui:
	@if [ ! -d "$(DASHBOARD_DIR)" ]; then \
		echo "$(PREFIX) $(DASHBOARD_DIR) not found"; exit 1; \
	fi
	@if [ -z "$(DASHBOARD_SKIP_BUILD)" ]; then \
		echo "$(PREFIX) Building dashboard in $(DASHBOARD_DIR)..."; \
		cd $(DASHBOARD_DIR) && bun install --frozen-lockfile && bun run build; \
	fi
	@if [ ! -d "$(DASHBOARD_DIR)/build" ]; then \
		echo "$(PREFIX) $(DASHBOARD_DIR)/build missing - run without DASHBOARD_SKIP_BUILD"; exit 1; \
	fi
	@echo "$(PREFIX) Updating embedded UI..."
	rm -rf public/ui/*
	cp -R $(DASHBOARD_DIR)/build/* public/ui/
	@echo "$(PREFIX) UI updated successfully!"

# Publish the vendored platform/ sub-projects OUT to their standalone repos.
#
# platform/<name>/ is the source of truth: the dashboard, registry and workflow
# projects live here alongside the Go code they talk to. This writes files only;
# review and commit in the destination repos yourself unless PLATFORM_COMMIT=1.
#
#   make sync-platform
#   make sync-platform PLATFORM=osmedeus-registry
#   make sync-platform PLATFORM_COMMIT=1
sync-platform:
	@bash build/scripts/sync-platform.sh

# Publish the embedded coding-agent skills OUT to the standalone skills repo.
#
# public/skills/ is the source of truth: bundles are authored here and embedded
# in the binary by //go:embed, so an installed skill always matches the running
# version. This target pushes them to a local checkout of the public
# osmedeus-skills repo (SKILLS_DEST) — it only writes files; review and commit
# in that repo yourself.
#
#   make sync-skills
#   make sync-skills SKILLS_DEST=/path/to/osmedeus-skills
#
# Every top-level directory containing a SKILL.md is copied, so adding a bundle
# needs no change here. --delete is scoped to each bundle directory, so file
# removals propagate without ever touching the destination's .git, README.md
# (upstream keeps its own landing page) or anything else that is not a bundle.
SKILLS_REPO ?= https://github.com/osmedeus/osmedeus-skills.git
SKILLS_SRC = public/skills
SKILLS_DEST ?= ../osmedeus-skills

sync-skills:
	@set -e; \
	src="$(SKILLS_SRC)"; \
	dest="$(SKILLS_DEST)"; \
	if [ ! -d "$$dest" ]; then \
		echo "$(PREFIX) Destination checkout not found: $$dest" >&2; \
		echo "$(PREFIX)   clone it first: git clone $(SKILLS_REPO) $$dest" >&2; \
		exit 1; \
	fi; \
	if [ "$$(cd "$$src" && pwd -P)" = "$$(cd "$$dest" && pwd -P)" ]; then \
		echo "$(PREFIX) Source and destination are the same directory ($$dest) — refusing to sync." >&2; \
		echo "$(PREFIX)   This target pushes $(SKILLS_SRC)/ -> the skills repo, not the other way around." >&2; \
		exit 1; \
	fi; \
	found=0; \
	for skill in "$$src"/*/; do \
		[ -f "$$skill/SKILL.md" ] || continue; \
		name=$$(basename "$$skill"); \
		echo "$(PREFIX) Syncing $$name"; \
		rsync -a --delete --exclude '.git' "$$skill" "$$dest/$$name/"; \
		found=$$((found + 1)); \
	done; \
	if [ "$$found" -eq 0 ]; then echo "$(PREFIX) No SKILL.md bundles found in $$src" >&2; exit 1; fi; \
	echo "$(PREFIX) Pushed $$found skill bundle(s) to $$dest/"; \
	git -C "$$dest" --no-pager status --short 2>/dev/null || true

# Development setup
dev-setup: install-gotestsum
	@echo "$(PREFIX) Setting up development environment..."
	$(GOMOD) download
	@echo "$(PREFIX) Done!"

# Docker build
docker-build:
	docker build -t osmedeus:$(VERSION) -f build/docker/Dockerfile .

# Docker run
docker-run:
	docker run -p 8002:8002 osmedeus:$(VERSION)

# Docker toolbox build (with all tools pre-installed)
docker-toolbox:
	@echo "$(PREFIX) Building osmedeus-toolbox Docker image..."
	docker-compose -f build/docker/docker-compose.toolbox.yaml build \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH)
	@echo "$(PREFIX) osmedeus-toolbox image built successfully!"
	@echo "$(PREFIX) Run with: docker-compose -f build/docker/docker-compose.toolbox.yaml up -d"

# Docker toolbox run
docker-toolbox-run:
	@echo "$(PREFIX) Starting osmedeus-toolbox container..."
	docker-compose -f build/docker/docker-compose.toolbox.yaml up -d
	@echo "$(PREFIX) Container started! Enter with: docker exec -it osmedeus-toolbox bash"

# Docker toolbox shell (interactive)
docker-toolbox-shell:
	docker exec -it osmedeus-toolbox bash

DOCKER_IMAGE ?= j3ssie/osmedeus:latest

# Docker publish (build and push a multi-arch image to Docker Hub)
#
# The two arches are built SEQUENTIALLY, not via a single `--platform a,b`
# build. Compiling the large generated SDKs (pulumi-azure-native, aws-sdk-go-v2)
# is memory-heavy, and building both arches at once OOM-kills the compiler on a
# typical (~16GB) Docker VM. Each arch is pushed by digest (no stray tags), then
# combined into one multi-arch manifest with `imagetools create`.
#
# Requires QEMU/binfmt for the foreign-arch runtime steps (run
# `make docker-buildx-setup` once if builds fail with exec-format errors).
docker-publish:
	@echo "$(PREFIX) Ensuring buildx builder 'osmedeus-builder' exists..."
	@docker buildx inspect osmedeus-builder >/dev/null 2>&1 || \
		docker buildx create --name osmedeus-builder --driver docker-container --bootstrap
	@repo=$$(echo "$(DOCKER_IMAGE)" | cut -d: -f1); \
	for arch in amd64 arm64; do \
		echo "$(PREFIX) Building linux/$$arch..."; \
		docker buildx build \
			--builder osmedeus-builder \
			--platform linux/$$arch \
			-f build/docker/Dockerfile \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			--build-arg COMMIT_HASH=$(COMMIT_HASH) \
			--output "type=image,name=$$repo,push-by-digest=true,name-canonical=true,push=true" \
			--metadata-file /tmp/osmedeus-$$arch.json \
			. || exit 1; \
	done; \
	amd64_digest=$$(jq -r '."containerimage.digest"' /tmp/osmedeus-amd64.json); \
	arm64_digest=$$(jq -r '."containerimage.digest"' /tmp/osmedeus-arm64.json); \
	echo "$(PREFIX) Combining into multi-arch manifest $(DOCKER_IMAGE)..."; \
	docker buildx imagetools create -t $(DOCKER_IMAGE) \
		$$repo@$$amd64_digest $$repo@$$arm64_digest || exit 1; \
	echo "$(PREFIX) Published $(DOCKER_IMAGE) (linux/amd64, linux/arm64) successfully!"; \
	docker buildx imagetools inspect $(DOCKER_IMAGE) | grep -iE "platform|name:"

# One-time setup: register QEMU emulators for cross-arch buildx builds
docker-buildx-setup:
	@echo "$(PREFIX) Registering QEMU emulators for cross-platform builds..."
	docker run --privileged --rm tonistiigi/binfmt --install all
	@echo "$(PREFIX) QEMU emulators registered."

# Release commands (GoReleaser)
# Registry assets are vendored under platform/osmedeus-registry/, so these copies
# no longer depend on a sibling checkout existing next to this repo.
REGISTRY_DIR=platform/osmedeus-registry

snapshot-release:
	@echo "$(PREFIX) Update registry-metadata-direct-fetch.json..."
	cp $(REGISTRY_DIR)/registry-metadata-direct-fetch.json public/presets/registry-metadata-direct-fetch.json
	@echo "$(PREFIX) Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/osmedeus
	@echo "$(PREFIX) Installing $(BINARY_NAME) to $(GOBIN_PATH)..."
	@mkdir -p $(GOBIN_PATH)
	@rm -f $(GOBIN_PATH)/$(BINARY_NAME)
	@cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/$(BINARY_NAME)
	@echo "$(PREFIX) Building snapshot release"
	export GORELEASER_CURRENT_TAG="$(VERSION)" && goreleaser release --clean --skip=announce,publish,validate
	@echo "$(PREFIX) Install script copied to dist/install.sh"
	cp $(REGISTRY_DIR)/install.sh dist/install.sh
	@echo "$(PREFIX) Prepare registry-metadata-direct-fetch.json"

local-release:
	@echo "$(PREFIX) Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/osmedeus
	@mkdir -p $(GOBIN_PATH)
	@rm -f $(GOBIN_PATH)/$(BINARY_NAME)
	@cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/$(BINARY_NAME)
	@echo "$(PREFIX) Building local snapshot for mac and linux arm only for testing..."
	export GORELEASER_CURRENT_TAG="$(VERSION)" && goreleaser release --config test/goreleaser-debug.yaml --clean --skip=announce,publish,validate

github-release:
	@echo "$(PREFIX) Building and publishing GitHub release..."
	export GORELEASER_CURRENT_TAG="$(VERSION)" && goreleaser release --clean

run-github-action:
	unset GH_TOKEN && gh workflow run manual-release.yaml && gh workflow run nightly-release.yaml

run-homebrew-action:
	unset GH_TOKEN && (cd ../homebrew-tap/ && gh workflow run 226998251)

# Bump the single source-of-truth version in internal/core/constants.go. npm
# versions are immutable, so every release needs a new number here before
# npm-publish. Default bumps the patch (v5.0.3 -> v5.0.4). Override with
# PART=minor|major|pre|release, LABEL=<label>, or SET=<explicit version>.
# DRY_RUN=1 previews only.
bump-version:
	@PART="$(PART)" LABEL="$(LABEL)" SET="$(SET)" DRY_RUN="$(DRY_RUN)" bash build/scripts/bump-version.sh

# --- npm distribution -----------------------------------------------------
# Publish the osmedeus binary to npm as @j3ssie/osmedeus. The binary ships
# gzipped inside per-platform optional-dependency packages (one npm name,
# version-suffixed platform builds). See build/npm/build.mjs.
NPM_OUT_DIR=build/dist-npm
NPM_BIN_DIR=build/dist-npm-bin
# npm versions are semver — strip the leading "v" from internal/core/constants.go
NPM_VERSION=$(patsubst v%,%,$(VERSION))
NPM_TARGETS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
# -s -w strips the symbol table and DWARF: the binary embeds the UI, presets and
# skills, so it is ~350MB unstripped and every megabyte lands in the npm tarball.
NPM_LDFLAGS=-ldflags "-s -w -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"

# "yes" when the npm source binaries are missing, incomplete, or were built for a
# different version than internal/core/constants.go — i.e. stale after a version
# bump. npm versions are immutable, so shipping a stale binary under a fresh
# version number can never be corrected, only superseded.
#
# The check is the .build-version stamp npm-binaries writes after all four builds
# succeed, NOT a version-string grep of the binary: osmedeus embeds docs, presets
# and UI assets that mention other versions, so a substring match false-matches.
# Recursive (=) so it only runs when referenced by the npm-build/npm-pack guards.
NPM_NEEDS_BUILD=$(shell stamp=$$(cat $(NPM_BIN_DIR)/.build-version 2>/dev/null); bins=$$(ls $(NPM_BIN_DIR)/$(BINARY_NAME)_*/$(BINARY_NAME) 2>/dev/null | wc -l | tr -d ' '); if [ "$$stamp" != "$(NPM_VERSION)" ] || [ "$$bins" -lt 4 ]; then echo yes; else echo no; fi)

# Cross-compile the four npm target platforms into $(NPM_BIN_DIR), in the
# goreleaser-compatible layout build.mjs expects (<binary>_<goos>_<goarch>/).
npm-binaries:
	@echo "$(PREFIX) Cross-compiling $(BINARY_NAME) $(VERSION) for: $(NPM_TARGETS)"
	@rm -rf $(NPM_BIN_DIR)
	@mkdir -p $(NPM_BIN_DIR)
	@set -e; for target in $(NPM_TARGETS); do \
		GOOS=$${target%/*}; GOARCH=$${target#*/}; \
		stage_dir="$(NPM_BIN_DIR)/$(BINARY_NAME)_$${GOOS}_$${GOARCH}"; \
		echo "$(PREFIX)   -> $${GOOS}/$${GOARCH}"; \
		mkdir -p $${stage_dir}; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} CGO_ENABLED=0 \
			$(GOBUILD) -trimpath $(NPM_LDFLAGS) \
			-o $${stage_dir}/$(BINARY_NAME) ./cmd/osmedeus; \
	done
	@echo "$(NPM_VERSION)" > $(NPM_BIN_DIR)/.build-version
	@echo "$(PREFIX) Binaries staged in $(NPM_BIN_DIR)/ (version $(NPM_VERSION))"

# Stage the npm packages. Rebuilds the binaries first if they are missing or
# stale (built at a different version), so a version bump never ships a
# mismatched binary.
npm-build:
	@if [ "$(NPM_NEEDS_BUILD)" = "yes" ]; then \
		echo "$(PREFIX) Binaries missing or stale for $(VERSION) — running 'make npm-binaries'..."; \
		$(MAKE) npm-binaries; \
	fi
	@echo "$(PREFIX) Staging npm packages (version $(NPM_VERSION))..."
	OSMEDEUS_VERSION=$(NPM_VERSION) node build/npm/build.mjs

# Stage + produce inspectable .tgz tarballs (npm pack) for each package.
npm-pack:
	@if [ "$(NPM_NEEDS_BUILD)" = "yes" ]; then \
		echo "$(PREFIX) Binaries missing or stale for $(VERSION) — running 'make npm-binaries'..."; \
		$(MAKE) npm-binaries; \
	fi
	@echo "$(PREFIX) Staging + packing npm tarballs (version $(NPM_VERSION))..."
	OSMEDEUS_VERSION=$(NPM_VERSION) node build/npm/build.mjs --pack
	@echo "$(PREFIX) Tarballs written to $(NPM_OUT_DIR)/"

# Publish to npm: platform packages FIRST (so the main package's
# optionalDependencies resolve), then the main package. Every run pins the
# `latest` dist-tag to the version being published (the main package is
# published as `latest` and `latest` is re-asserted + verified afterward), so
# `npm i -g @j3ssie/osmedeus` always installs this version.
#
# Auth comes from ~/.npmrc:
#   //registry.npmjs.org/:_authToken=${NPM_TOKEN}
# npm reads ~/.npmrc automatically (default userconfig, cwd-independent) and
# interpolates ${NPM_TOKEN} from the environment — just keep NPM_TOKEN
# exported. Set DRY_RUN=1 to preview (no token needed; --dry-run only packs).
npm-publish: npm-build
	@echo "$(PREFIX) Publishing @j3ssie/osmedeus ($(NPM_VERSION)) to npm [latest -> $(NPM_VERSION)]..."
	@set -e; \
		if [ "$(DRY_RUN)" = "1" ]; then \
			DRY="--dry-run"; \
			echo "$(PREFIX) DRY RUN — nothing will be published"; \
		else \
			DRY=""; \
			if [ -z "$${NPM_TOKEN:-}" ]; then \
				echo "\033[31m[!] NPM_TOKEN not set; ~/.npmrc auth will fail. Export it or use DRY_RUN=1.\033[0m"; \
				exit 1; \
			fi; \
		fi; \
		for d in $(NPM_OUT_DIR)/$(BINARY_NAME)-*/; do \
			ptag=$$(basename "$$d" | sed 's/^$(BINARY_NAME)-//'); \
			echo "$(PREFIX)   publishing platform package [$$ptag]"; \
			( cd "$$d" && npm publish --access public --tag "$$ptag" $$DRY ) \
				|| { echo "\033[31m[!] publish failed: $$ptag\033[0m"; exit 11; }; \
		done; \
		echo "$(PREFIX)   publishing $(NPM_OUT_DIR)/$(BINARY_NAME)/ (main) [tag=latest]"; \
		( cd $(NPM_OUT_DIR)/$(BINARY_NAME) && npm publish --access public --tag latest $$DRY ) \
			|| { echo "\033[31m[!] publish failed: main\033[0m"; exit 12; }; \
		if [ "$(DRY_RUN)" != "1" ]; then \
			echo "$(PREFIX)   pointing 'latest' dist-tag at $(NPM_VERSION)"; \
			npm dist-tag add @j3ssie/osmedeus@$(NPM_VERSION) latest \
				|| { echo "\033[31m[!] dist-tag add failed\033[0m"; exit 13; }; \
			resolved=""; \
			for i in 1 2 3 4 5 6; do \
				resolved=$$(npm dist-tag ls @j3ssie/osmedeus --prefer-online 2>/dev/null \
					| sed -n 's/^latest: //p'); \
				[ "$$resolved" = "$(NPM_VERSION)" ] && break; \
				echo "$(PREFIX)   latest still '$$resolved' (npm registry cache lag) — retry $$i/6 in 10s"; \
				sleep 10; \
			done; \
			if [ "$$resolved" != "$(NPM_VERSION)" ]; then \
				echo "\033[31m[!] latest resolved to '$$resolved', expected '$(NPM_VERSION)' after retries\033[0m"; \
				echo "\033[31m    Publish likely succeeded — verify: npm dist-tag ls @j3ssie/osmedeus\033[0m"; \
				exit 14; \
			fi; \
			echo "$(PREFIX)   verified: latest -> $$resolved"; \
		fi
	@echo "$(PREFIX) npm publish complete"

# Database commands
db-seed: build
	@echo "$(PREFIX) Seeding database..."
	./$(BINARY_DIR)/$(BINARY_NAME) db seed

db-clean: build
	@echo "$(PREFIX) Cleaning database..."
	./$(BINARY_DIR)/$(BINARY_NAME) db clean --force

db-migrate: build
	@echo "$(PREFIX) Running database migrations..."
	./$(BINARY_DIR)/$(BINARY_NAME) db migrate

# Help
help:
	@echo ""
	@echo "\033[32m Osmedeus $(VERSION) - A Modern Orchestration Engine for Security\033[0m"
	@echo "\033[36m                 Crafted with \033[31m<3\033[35m by $(AUTHOR)                      \033[0m"
	@echo "\033[34m     ──────────────────────────────────────────────────\033[0m"
	@echo ""
	@echo "\033[33m  BUILD & INSTALL\033[0m"
	@echo "    make build            Build and install binary to \$$GOBIN (or \$$GOPATH/bin)"
	@echo "    make build-all        Build for all platforms (linux, darwin, windows)"
	@echo "    make install          Install binary to \$$GOBIN (builds first if needed)"
	@echo "    make clean            Clean build artifacts"
	@echo ""
	@echo "\033[33m  RUN\033[0m"
	@echo "    make run              Build and run the application"
	@echo "    make run-server       Build and start the server"
	@echo "    make run-server-debug Build and start server in debug mode (no auth)"
	@echo ""
	@echo "\033[33m  TEST\033[0m"
	@echo "    make test             Run all tests with race detection"
	@echo "    make test-unit        Run unit tests (fast, no external deps)"
	@echo "    make test-integration Run integration tests (pattern match)"
	@echo "    make test-workflow-integration  Run workflow integration tests (test/integration/)"
	@echo "    make test-all         Run unit + integration tests"
	@echo "    make test-e2e         Run E2E CLI tests"
	@echo "    make test-e2e-verbose Run E2E tests with verbose output"
	@echo "    make test-e2e-ssh     Run SSH E2E tests (full workflows)"
	@echo "    make test-e2e-api     Run API E2E tests (all endpoints, requires Redis)"
	@echo "    make test-e2e-nix     Run Nix mode E2E tests (requires Docker)"
	@echo "    make test-e2e-install Run install E2E tests (workflow/base from zip/URL/git)"
	@echo "    make test-sudo        Run sudo-aware E2E tests (requires sudo prompt)"
	@echo "    make test-docker      Run Docker runner tests"
	@echo "    make test-ssh         Run SSH runner unit tests"
	@echo "    make test-distributed Run distributed scan E2E tests (requires Redis)"
	@echo "    make distributed-e2e-up    Build and start distributed E2E stack (Docker)"
	@echo "    make distributed-e2e-run   Submit a real scan to the distributed stack"
	@echo "    make distributed-e2e-down  Tear down distributed E2E stack"
	@echo "    make test-canary-all   Run all canary tests (real scans in Docker, 30-60min)"
	@echo "    make test-canary-repo  Run repo scan canary (juice-shop SAST, ~25min)"
	@echo "    make test-canary-domain Run domain-lite canary (hackerone.com, ~20min)"
	@echo "    make test-canary-ip    Run CIDR scan canary (IP list, ~25min)"
	@echo "    make test-canary-general Run general scan canary (hackerone.com, ~40min)"
	@echo "    make test-coverage    Run tests with coverage report"
	@echo "    make test-summary     Quick pass/fail summary (dots format)"
	@echo "    make test-ci          Run tests with JUnit XML output"
	@echo ""
	@echo "\033[33m  DEVELOPMENT\033[0m"
	@echo "    make dev-setup        Set up development environment"
	@echo "    make fmt              Format code"
	@echo "    make lint             Run golangci-lint"
	@echo "    make tidy             Tidy go.mod dependencies"
	@echo "    make deps             Download dependencies"
	@echo "    make update-deps      Update all dependencies"
	@echo "    make generate         Run go generate"
	@echo "    make swagger          Generate swagger documentation"
	@echo "    make update-ui        Update embedded UI from dashboard build"
	@echo "    make sync-skills      Push public/skills/ out to $(SKILLS_DEST)"
	@echo ""
	@echo "\033[33m  DOCKER\033[0m"
	@echo "    make docker-build     Build Docker image"
	@echo "    make docker-run       Run Docker container"
	@echo "    make docker-publish   Build & push multi-arch (amd64+arm64) j3ssie/osmedeus:latest (sequential)"
	@echo "    make docker-buildx-setup  Register QEMU for cross-arch builds (one-time)"
	@echo "    make docker-toolbox       Build toolbox image (all tools pre-installed)"
	@echo "    make docker-toolbox-run   Start toolbox container"
	@echo "    make docker-toolbox-shell Enter toolbox container shell"
	@echo ""
	@echo "\033[33m  RELEASE\033[0m"
	@echo "    make snapshot-release Build local snapshot release (no publish)"
	@echo "    make local-release    Build local snapshot for mac/linux arm (testing)"
	@echo "    make github-release   Build and publish GitHub release"
	@echo "    make run-github-action    Trigger manual and nightly GitHub workflows"
	@echo "    make bump-version     Bump version in constants.go ($(VERSION) -> next patch)"
	@echo "                          PART=minor|major|pre|release SET=<ver> DRY_RUN=1"
	@echo "    make npm-binaries     Cross-compile the 4 npm target platforms"
	@echo "    make npm-build        Stage @j3ssie/osmedeus npm packages into build/dist-npm/"
	@echo "    make npm-pack         npm-build + produce inspectable .tgz tarballs"
	@echo "    make npm-publish      Publish to npm; auth via ~/.npmrc (NPM_TOKEN env var)"
	@echo ""
	@echo "\033[33m  DATABASE\033[0m"
	@echo "    make db-seed          Seed database with sample data"
	@echo "    make db-clean         Clean all data from database"
	@echo "    make db-migrate       Run database migrations"
	@echo ""

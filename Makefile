.PHONY: build run test test-unit test-integration test-workflow-integration test-e2e test-e2e-verbose test-e2e-ssh test-e2e-api test-e2e-nix test-e2e-install test-docker test-ssh test-distributed test-all test-summary test-ci clean install install-gotestsum lint fmt db-seed db-clean db-migrate run-server-debug swagger update-ui snapshot-release github-release github-action docker-toolbox docker-toolbox-run docker-toolbox-shell docker-publish

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
else
    TESTCMD=$(GOTEST)
    TESTFLAGS=-v
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
	@cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/

# Install to GOBIN (or GOPATH/bin) - requires prior build
install:
	@echo "$(PREFIX) Installing $(BINARY_NAME) to $(GOBIN_PATH)..."
	@if [ ! -f "$(BINARY_DIR)/$(BINARY_NAME)" ]; then \
		echo "$(PREFIX) Binary not found, building first..."; \
		$(MAKE) build; \
	else \
		cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/; \
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

# Update embedded UI from dashboard build
update-ui:
	@echo "$(PREFIX) Updating embedded UI..."
	rm -rf public/ui/*
	cp -R ../osmedeus-dashboard/build/* public/ui/
	@echo "$(PREFIX) UI updated successfully!"

# Development setup
dev-setup: install-gotestsum
	@echo "$(PREFIX) Setting up development environment..."
	$(GOMOD) download
	@echo "$(PREFIX) Done!"

# Docker build
docker-build:
	docker build -t osmedeus:$(VERSION) .

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

# Docker publish (build and push to Docker Hub)
docker-publish:
	@echo "$(PREFIX) Building Docker image j3ssie/osmedeus:latest..."
	docker build -t j3ssie/osmedeus:latest \
		-f build/docker/Dockerfile \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		.
	@echo "$(PREFIX) Pushing to Docker Hub..."
	docker push j3ssie/osmedeus:latest
	@echo "$(PREFIX) Published j3ssie/osmedeus:latest successfully!"

# Release commands (GoReleaser)
snapshot-release:
	@echo "$(PREFIX) Update registry-metadata-direct-fetch.json..."
	cp ../osmedeus-registry/registry-metadata-direct-fetch.json public/presets/registry-metadata-direct-fetch.json
	@echo "$(PREFIX) Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/osmedeus
	@echo "$(PREFIX) Installing $(BINARY_NAME) to $(GOBIN_PATH)..."
	@cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/
	@echo "$(PREFIX) Building snapshot release"
	export GORELEASER_CURRENT_TAG="$(VERSION)" && goreleaser release --clean --skip=announce,publish,validate
	@echo "$(PREFIX) Install script copied to dist/install.sh"
	cp ../osmedeus-registry/install.sh dist/install.sh
	@echo "$(PREFIX) Prepare registry-metadata-direct-fetch.json"

local-release:
	@echo "$(PREFIX) Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/osmedeus
	@cp $(BINARY_DIR)/$(BINARY_NAME) $(GOBIN_PATH)/
	@echo "$(PREFIX) Building local snapshot for mac and linux arm only for testing..."
	export GORELEASER_CURRENT_TAG="$(VERSION)" && goreleaser release --config test/goreleaser-debug.yaml --clean --skip=announce,publish,validate

github-release:
	@echo "$(PREFIX) Building and publishing GitHub release..."
	export GORELEASER_CURRENT_TAG="$(VERSION)" && goreleaser release --clean

github-action:
	unset GH_TOKEN &&gh workflow run manual-release.yaml && gh workflow run nightly-release.yaml

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
	@echo "    make test-docker      Run Docker runner tests"
	@echo "    make test-ssh         Run SSH runner unit tests"
	@echo "    make test-distributed Run distributed scan E2E tests (requires Redis)"
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
	@echo ""
	@echo "\033[33m  DOCKER\033[0m"
	@echo "    make docker-build     Build Docker image"
	@echo "    make docker-run       Run Docker container"
	@echo "    make docker-publish   Build and push j3ssie/osmedeus:latest to Docker Hub"
	@echo "    make docker-toolbox       Build toolbox image (all tools pre-installed)"
	@echo "    make docker-toolbox-run   Start toolbox container"
	@echo "    make docker-toolbox-shell Enter toolbox container shell"
	@echo ""
	@echo "\033[33m  RELEASE\033[0m"
	@echo "    make snapshot-release Build local snapshot release (no publish)"
	@echo "    make local-release    Build local snapshot for mac/linux arm (testing)"
	@echo "    make github-release   Build and publish GitHub release"
	@echo "    make github-action    Trigger manual and nightly GitHub workflows"
	@echo ""
	@echo "\033[33m  DATABASE\033[0m"
	@echo "    make db-seed          Seed database with sample data"
	@echo "    make db-clean         Clean all data from database"
	@echo "    make db-migrate       Run database migrations"
	@echo ""

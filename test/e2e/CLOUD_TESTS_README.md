# Cloud Tests for Osmedeus

This document describes the E2E and integration tests for the osmedeus cloud functionality.

## Overview

The cloud feature enables distributed security scanning across multiple cloud providers (DigitalOcean, AWS, GCP, Linode, Azure). Tests are organized into two categories:

1. **E2E Tests** (`test/e2e/cloud_test.go`) - CLI command testing
2. **Integration Tests** (`test/integration/cloud_integration_test.go`) - Internal package testing

## Running Cloud Tests

### Quick Start

```bash
# Run all cloud E2E tests
make test-e2e-cloud

# Run cloud integration tests
make test-cloud

# Run both
make test-e2e-cloud test-cloud

# Run specific test
go test -v -run TestCloud_ConfigSet ./test/e2e/
go test -v -run TestCloudConfig_LoadAndSave ./test/integration/
```

### With Go Test Directly

```bash
# E2E tests
go test -v ./test/e2e/cloud_test.go ./test/e2e/e2e_test.go

# Integration tests
go test -v ./test/integration/cloud_integration_test.go

# Run with short mode (skips long-running tests)
go test -v -short ./test/e2e/cloud_test.go ./test/e2e/e2e_test.go
```

## Test Coverage

### E2E CLI Tests (`cloud_test.go`)

These tests verify CLI commands work correctly:

#### Configuration Management
- ✅ `TestCloud_ConfigSet` - Setting configuration values
- ✅ `TestCloud_ConfigShow` - Displaying configuration
- ✅ `TestCloud_ConfigSetInvalidKey` - Error handling for invalid keys
- ✅ `TestCloud_ConfigEnvironmentVariables` - Environment variable resolution
- ✅ `TestCloud_ConfigFileCreation` - Config file creation
- ✅ `TestCloud_MultipleProviderConfigs` - Multi-provider configuration
- ✅ `TestCloud_CostLimitConfiguration` - Cost limit settings
- ✅ `TestCloud_SSHKeyConfiguration` - SSH key configuration
- ✅ `TestCloud_ProviderRegions` - Region configuration for all providers

#### Infrastructure Management
- ✅ `TestCloud_CreateHelp` - Create command help output
- ✅ `TestCloud_CreateWithoutConfig` - Error handling without config
- ✅ `TestCloud_CreateDryRun` - Cost estimation before provisioning
- ✅ `TestCloud_List` - Listing infrastructure
- ✅ `TestCloud_DestroyHelp` - Destroy command help output
- ✅ `TestCloud_DestroyNonExistent` - Handling non-existent infrastructure

#### Workflow Execution
- ✅ `TestCloud_RunHelp` - Run command help output
- ✅ `TestCloud_RunWithoutTarget` - Error handling without target
- ✅ `TestCloud_RunWithDistributed` - Cloud run with distributed workflow

#### Validation & Error Handling
- ✅ `TestCloud_ProviderValidation` - Invalid provider name rejection
- ✅ `TestCloud_InstanceCountValidation` - Instance count validation
- ✅ `TestCloud_StateDirectory` - State directory creation
- ✅ `TestCloud_WithTimeout` - Timeout configuration
- ✅ `TestCloud_Cleanup_OnFailure` - Cleanup behavior on failure

#### Advanced Tests
- ✅ `TestCloud_Integration_FullLifecycle` - Complete workflow sequence
- ✅ `TestCloud_Distributed_WorkerRegistration` - Worker registration (placeholder)
- ✅ `TestCloud_CostTracking` - Cost tracking (placeholder)
- ✅ `TestCloud_ParallelOperations` - Concurrent operation isolation
- ✅ `TestCloud_SpotInstanceConfiguration` - Spot/preemptible instance settings
- ✅ `TestCloud_CustomSetupCommands` - Custom worker setup commands

### Integration Tests (`cloud_integration_test.go`)

These tests verify internal cloud package functionality:

#### Configuration Management
- ✅ `TestCloudConfig_LoadAndSave` - Config persistence
- ✅ `TestCloudConfig_EnvironmentVariableResolution` - Env var substitution
- ✅ `TestCloudConfig_Validation` - Config validation rules
- ✅ `TestCloudConfigMultipleProviders` - Multi-provider configuration
- ✅ `TestDefaultCloudConfig` - Default configuration values
- ✅ `TestCloudConfigValidation_AWS` - AWS-specific validation
- ✅ `TestCloudConfigValidation_GCP` - GCP-specific validation

#### Cost Tracking
- ✅ `TestCostTracking` - Cost tracking functionality
- ✅ `TestCostLimitCheck` - Cost limit enforcement
- ✅ `TestCostEstimation` - Cost estimation for durations

#### State Management
- ✅ `TestInfrastructureState` - Infrastructure state CRUD operations
- ✅ `TestMultipleInfrastructureStates` - Managing multiple states
- ✅ `TestInfrastructureStateEmpty` - Empty state directory handling
- ✅ `TestInfrastructureMetadata` - Metadata persistence

## Test Architecture

### Test Utilities

E2E tests use shared utilities from `e2e_test.go`:

```go
// Custom logger with colored output
log := NewTestLogger(t)
log.Step("Testing cloud config set command")
log.Info("Setting provider token...")
log.Success("Configuration updated successfully")

// CLI execution helpers
stdout, stderr, err := runCLIWithLog(t, log, "cloud", "config", "set", "key", "value")
baseDir, stdout, stderr, err := runCLIWithLogAndBase(t, log, "cloud", "create", ...)
```

### Test Patterns

#### Pattern 1: Simple CLI Test
```go
func TestCloud_ConfigShow(t *testing.T) {
    log := NewTestLogger(t)
    log.Step("Testing cloud config show command")

    stdout, _, err := runCLIWithLog(t, log, "cloud", "config", "show")
    require.NoError(t, err)
    assert.Contains(t, stdout, "provider:")

    log.Success("Cloud config show works correctly")
}
```

#### Pattern 2: Integration Test with State
```go
func TestInfrastructureState(t *testing.T) {
    tempDir := t.TempDir()
    stateDir := filepath.Join(tempDir, "cloud-state")

    // Create, save, load, verify, cleanup
    infra := &cloud.Infrastructure{...}
    err := cloud.SaveInfrastructureState(infra, stateDir)
    require.NoError(t, err)

    loaded, err := cloud.LoadInfrastructureState(infra.ID, stateDir)
    require.NoError(t, err)
    assert.Equal(t, infra.ID, loaded.ID)
}
```

#### Pattern 3: Skipping Long Tests
```go
func TestCloud_Integration_FullLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping full lifecycle integration test in short mode")
    }

    // Long-running test...
}
```

## Cloud Providers Tested

| Provider | Configuration | Validation | Cost Estimation |
|----------|---------------|------------|-----------------|
| **DigitalOcean** | ✅ | ✅ | ✅ |
| **AWS** | ✅ | ✅ | ⏳ |
| **GCP** | ✅ | ✅ | ⏳ |
| **Linode** | ✅ | ⏳ | ⏳ |
| **Azure** | ✅ | ⏳ | ⏳ |

**Legend:**
- ✅ Fully tested
- ⏳ Partial/placeholder (implementation pending)
- ❌ Not tested

## Test Data

### Configuration Files
- Temporary directories created via `t.TempDir()`
- Config files: `cloud-settings.yaml`
- State files: `cloud-state/infrastructure/<id>.json`

### Mock Data
```go
// Test tokens/credentials
"test-token-12345"
"env-resolved-token-67890"
"do-token-12345"

// Test infrastructure IDs
"test-infra-12345"
"infra-1", "infra-2", "infra-3"

// Test regions
DigitalOcean: "nyc3"
AWS: "us-east-1"
GCP: "us-central1"
```

## Environment Variables

Tests use environment variables for:
- `TEST_DO_TOKEN` - DigitalOcean token for env resolution tests
- Other `${VAR}` placeholders in default configs

## Short Mode

Run with `-short` to skip long-running tests:

```bash
go test -short ./test/e2e/cloud_test.go ./test/e2e/e2e_test.go
go test -short ./test/integration/cloud_integration_test.go
```

Skipped tests:
- Cloud create/destroy operations
- Full lifecycle integration tests
- Distributed worker registration
- Cost tracking over time

## Future Test Additions

### When Cloud Functionality is Fully Implemented:

1. **Real Provider Tests** (with credentials)
   - Actual infrastructure provisioning
   - Worker registration and monitoring
   - Distributed workflow execution
   - Result collection via SSH

2. **Cost Tracking**
   - Real-time cost accumulation
   - Max spend limit enforcement
   - Cost summary reporting

3. **Worker Management**
   - Cloud-init script execution
   - Worker auto-registration
   - Worker health monitoring

4. **Snapshot Support**
   - Custom VM snapshots with pre-installed tools
   - Boot time optimization (~30s vs 5min)

5. **Error Recovery**
   - Network failure handling
   - Partial infrastructure cleanup
   - State corruption recovery

## Troubleshooting

### Tests Failing

```bash
# Check if cloud config is interfering
rm -rf ~/.osmedeus-base/cloud-settings.yaml

# Run with verbose output
go test -v -run TestCloud_ConfigSet ./test/e2e/

# Check for leftover temp files
ls /tmp | grep -i osmedeus
```

### Debugging

```bash
# Add -v for verbose test output
go test -v ./test/e2e/cloud_test.go ./test/e2e/e2e_test.go

# Run single test
go test -v -run TestCloud_ConfigShow ./test/e2e/

# Show test coverage
go test -coverprofile=coverage.out ./test/integration/cloud_integration_test.go
go tool cover -html=coverage.out
```

## CI/CD Integration

These tests are designed to run in CI pipelines:

```yaml
# GitHub Actions example
- name: Run cloud tests
  run: |
    make test-e2e-cloud
    make test-cloud
```

All tests use temporary directories and cleanup after themselves, making them safe for parallel execution.

## Contributing

When adding new cloud features, please:

1. Add corresponding E2E tests in `cloud_test.go`
2. Add integration tests in `cloud_integration_test.go`
3. Update this README with new test coverage
4. Follow existing test patterns (TestLogger, assertions, cleanup)
5. Support `-short` mode for quick validation

## Related Documentation

- [Cloud Usage Guide](../../docs/cloud-usage-guide.md) - User-facing documentation
- [Cloud Config Example](../../public/presets/cloud-settings.example.yaml) - Configuration template
- [CLAUDE.md](../../CLAUDE.md) - Architecture overview
- [API Tests](./api_test.go) - API endpoint tests
- [Distributed Tests](./distributed_test.go) - Distributed workflow tests

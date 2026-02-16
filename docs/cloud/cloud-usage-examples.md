# Osmedeus Cloud - Usage Examples

Complete guide with practical examples for using osmedeus cloud functionality to run distributed security scans across multiple cloud providers.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Infrastructure Management](#infrastructure-management)
- [Running Workflows](#running-workflows)
- [Cost Management](#cost-management)
- [Provider-Specific Examples](#provider-specific-examples)
- [Advanced Scenarios](#advanced-scenarios)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### 1. Initial Setup

```bash
# Install osmedeus cloud feature (if not already installed)
osmedeus install base --preset

# Initialize cloud configuration with defaults
osmedeus cloud config show
```

### 2. Configure Provider (DigitalOcean Example)

```bash
# Set DigitalOcean token
osmedeus cloud config set providers.digitalocean.token "dop_v1_abc123..."

# Set preferred region
osmedeus cloud config set providers.digitalocean.region "nyc3"

# Set default provider
osmedeus cloud config set defaults.provider "digitalocean"
```

### 3. Run Your First Cloud Scan

```bash
# Quick domain reconnaissance on 3 cloud workers
osmedeus cloud run -f general -t example.com --instances 3

# This will:
# 1. Provision 3 VMs on DigitalOcean
# 2. Install osmedeus on each worker
# 3. Run the general flow on example.com
# 4. Collect results
# 5. Destroy infrastructure
```

---

## Configuration

### View Current Configuration

```bash
# Show all cloud configuration
osmedeus cloud config show

# Show in YAML format
cat ~/osmedeus-base/cloud-settings.yaml
```

### Setting Configuration Values

```bash
# Set provider credentials
osmedeus cloud config set providers.digitalocean.token "YOUR_TOKEN"
osmedeus cloud config set providers.aws.access_key_id "YOUR_KEY"
osmedeus cloud config set providers.gcp.project_id "your-project"

# Set default provider and mode
osmedeus cloud config set defaults.provider "digitalocean"
osmedeus cloud config set defaults.mode "vm"

# Set resource limits
osmedeus cloud config set defaults.max_instances 10
osmedeus cloud config set limits.max_hourly_spend 5.00
osmedeus cloud config set limits.max_total_spend 50.00

# Set timeout for operations
osmedeus cloud config set defaults.timeout "30m"

# Enable cleanup on failure
osmedeus cloud config set defaults.cleanup_on_failure true
```

### Using Environment Variables

```bash
# Set credentials via environment variables
export DIGITALOCEAN_TOKEN="dop_v1_abc123..."
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export GCP_PROJECT_ID="my-gcp-project"

# Reference in config (stored as ${VAR_NAME})
osmedeus cloud config set providers.digitalocean.token '${DIGITALOCEAN_TOKEN}'
osmedeus cloud config set providers.aws.access_key_id '${AWS_ACCESS_KEY_ID}'
```

### Configure SSH Keys

```bash
# Set SSH key for accessing cloud workers
osmedeus cloud config set ssh.private_key_path "~/.ssh/cloud_rsa"
osmedeus cloud config set ssh.public_key_path "~/.ssh/cloud_rsa.pub"
osmedeus cloud config set ssh.user "root"

# Or provide key content directly
osmedeus cloud config set ssh.private_key_content "$(cat ~/.ssh/cloud_rsa)"
```

### Configure Instance Types

```bash
# DigitalOcean droplet size
osmedeus cloud config set providers.digitalocean.size "s-4vcpu-8gb"

# AWS instance type
osmedeus cloud config set providers.aws.instance_type "t3.xlarge"

# GCP machine type
osmedeus cloud config set providers.gcp.machine_type "n1-standard-4"

# Enable spot/preemptible instances for cost savings
osmedeus cloud config set providers.aws.use_spot true
osmedeus cloud config set providers.gcp.use_preemptible true
```

---

## Infrastructure Management

### Create Infrastructure

```bash
# Create 5 DigitalOcean droplets
osmedeus cloud create --instances 5

# Create with specific provider
osmedeus cloud create --provider aws --instances 10

# Create with custom mode (vm or serverless)
osmedeus cloud create --provider gcp --mode vm --instances 3

# Force creation (skip confirmation prompts)
osmedeus cloud create --instances 5 --force
```

**Example Output:**
```
[*] Estimating costs...
    Hourly Rate: $0.11/hr (5 × s-2vcpu-4gb)
    Daily Cost: $2.68

[*] Creating infrastructure...
    ✓ Provisioning VM 1/5 (droplet-123456)
    ✓ Provisioning VM 2/5 (droplet-123457)
    ...

[*] Waiting for workers to register...
    ✓ Worker wosm-203.0.113.1 registered
    ✓ Worker wosm-203.0.113.2 registered
    ...

[*] Infrastructure created: cloud-do-1708234567
```

### List Infrastructure

```bash
# List all active cloud infrastructure
osmedeus cloud list
```

**Example Output:**
```
┌─────────────────────┬──────────────┬──────┬───────────┬─────────────────────┐
│ ID                  │ Provider     │ Mode │ Resources │ Created             │
├─────────────────────┼──────────────┼──────┼───────────┼─────────────────────┤
│ cloud-do-1708234567 │ digitalocean │ vm   │ 5 VMs     │ 2024-02-17 10:30:45 │
│ cloud-aws-170823891 │ aws          │ vm   │ 10 VMs    │ 2024-02-17 09:15:22 │
└─────────────────────┴──────────────┴──────┴───────────┴─────────────────────┘
```

### Destroy Infrastructure

```bash
# Destroy specific infrastructure by ID
osmedeus cloud destroy cloud-do-1708234567

# Destroy all infrastructure
osmedeus cloud destroy --all

# Force destroy without confirmation
osmedeus cloud destroy cloud-do-1708234567 --force
```

**Example Output:**
```
[*] Destroying infrastructure: cloud-do-1708234567
    ✓ Terminating droplet-123456
    ✓ Terminating droplet-123457
    ...
    ✓ Removing state file

[*] Infrastructure destroyed
    Total runtime: 1h 23m
    Total cost: $0.15
```

---

## Running Workflows

### Basic Cloud Run

```bash
# Run general flow on single target
osmedeus cloud run -f general -t example.com --instances 3

# Run module on target
osmedeus cloud run -m subdomain-enumeration -t example.com --instances 5

# Run on multiple targets from file
osmedeus cloud run -f general -T targets.txt --instances 10
```

### With Custom Provider

```bash
# Run on AWS
osmedeus cloud run -f general -t example.com --provider aws --instances 5

# Run on GCP with preemptible instances
osmedeus cloud config set providers.gcp.use_preemptible true
osmedeus cloud run -f general -t example.com --provider gcp --instances 8
```

### With Concurrent Targets

```bash
# Scan 10 targets concurrently on 5 workers
osmedeus cloud run -f general -T targets.txt -c 10 --instances 5

# This distributes targets across workers automatically
```

### With Timeout

```bash
# Run with 2-hour timeout
osmedeus cloud run -f general -t example.com --instances 3 --timeout 2h

# Timeout applies to entire cloud run (provision + scan + cleanup)
```

### With Custom Workflows

```bash
# Run custom workflow from file
osmedeus cloud run -m /path/to/custom-workflow.yaml -t example.com --instances 5

# Run with parameters file
osmedeus cloud run -f general -t example.com --instances 3 -P params.yaml
```

---

## Cost Management

### View Cost Estimates

```bash
# Configuration shows cost limits
osmedeus cloud config show

# Example output includes:
# limits:
#   max_hourly_spend: 5.00
#   max_total_spend: 50.00
```

### Set Cost Limits

```bash
# Set maximum hourly spend (blocks provisioning if exceeded)
osmedeus cloud config set limits.max_hourly_spend 10.00

# Set maximum total spend (terminates if exceeded during run)
osmedeus cloud config set limits.max_total_spend 100.00

# Set maximum number of instances
osmedeus cloud config set limits.max_instances 20
```

### Cost Estimation Examples

**DigitalOcean Pricing:**
```bash
# 3 × s-2vcpu-4gb droplets
# Cost: 3 × $0.02232/hr = $0.06696/hr ($1.61/day)
osmedeus cloud create --instances 3

# 10 × s-4vcpu-8gb droplets
# Cost: 10 × $0.04464/hr = $0.44640/hr ($10.71/day)
osmedeus cloud create --instances 10
osmedeus cloud config set providers.digitalocean.size "s-4vcpu-8gb"
```

**AWS Spot Instances (70% savings):**
```bash
# Enable spot instances
osmedeus cloud config set providers.aws.use_spot true

# 5 × t3.medium spot instances
# Regular: ~$0.0416/hr × 5 = $0.208/hr
# Spot: ~$0.0125/hr × 5 = $0.0625/hr (saves ~$0.14/hr)
osmedeus cloud run -f general -t example.com --provider aws --instances 5
```

---

## Provider-Specific Examples

### DigitalOcean

```bash
# Configure DigitalOcean
osmedeus cloud config set providers.digitalocean.token "dop_v1_..."
osmedeus cloud config set providers.digitalocean.region "nyc3"
osmedeus cloud config set providers.digitalocean.size "s-2vcpu-4gb"
osmedeus cloud config set providers.digitalocean.image "ubuntu-22-04-x64"

# Run with DigitalOcean
osmedeus cloud run -f general -t example.com --provider digitalocean --instances 5

# Available regions: nyc1, nyc2, nyc3, sfo1, sfo2, sfo3, ams2, ams3, sgp1, lon1, fra1, tor1, blr1
# Available sizes: s-1vcpu-1gb, s-2vcpu-4gb, s-4vcpu-8gb, s-8vcpu-16gb, etc.
```

### AWS

```bash
# Configure AWS
osmedeus cloud config set providers.aws.access_key_id "AKIAIOSFODNN7EXAMPLE"
osmedeus cloud config set providers.aws.secret_access_key "wJalrXUtnFEMI..."
osmedeus cloud config set providers.aws.region "us-east-1"
osmedeus cloud config set providers.aws.instance_type "t3.medium"
osmedeus cloud config set providers.aws.use_spot true

# Run with AWS
osmedeus cloud run -f general -t example.com --provider aws --instances 10

# Available regions: us-east-1, us-west-2, eu-west-1, ap-southeast-1, etc.
# Instance types: t3.micro, t3.small, t3.medium, t3.large, t3.xlarge, c5.large, etc.
```

### Google Cloud Platform (GCP)

```bash
# Configure GCP
osmedeus cloud config set providers.gcp.project_id "my-project-12345"
osmedeus cloud config set providers.gcp.credentials_file "/path/to/credentials.json"
osmedeus cloud config set providers.gcp.region "us-central1"
osmedeus cloud config set providers.gcp.zone "us-central1-a"
osmedeus cloud config set providers.gcp.machine_type "n1-standard-2"
osmedeus cloud config set providers.gcp.use_preemptible true

# Run with GCP
osmedeus cloud run -f general -t example.com --provider gcp --instances 8

# Machine types: f1-micro, g1-small, n1-standard-1, n1-standard-2, n1-standard-4, etc.
```

### Linode

```bash
# Configure Linode
osmedeus cloud config set providers.linode.token "linod_..."
osmedeus cloud config set providers.linode.region "us-east"
osmedeus cloud config set providers.linode.type "g6-standard-2"
osmedeus cloud config set providers.linode.image "linode/ubuntu22.04"

# Run with Linode
osmedeus cloud run -f general -t example.com --provider linode --instances 5

# Regions: us-east, us-west, eu-west, ap-south, etc.
# Types: g6-nanode-1, g6-standard-1, g6-standard-2, g6-standard-4, etc.
```

### Azure

```bash
# Configure Azure
osmedeus cloud config set providers.azure.subscription_id "..."
osmedeus cloud config set providers.azure.tenant_id "..."
osmedeus cloud config set providers.azure.client_id "..."
osmedeus cloud config set providers.azure.client_secret "..."
osmedeus cloud config set providers.azure.location "eastus"
osmedeus cloud config set providers.azure.vm_size "Standard_B2s"

# Run with Azure
osmedeus cloud run -f general -t example.com --provider azure --instances 5

# Locations: eastus, westus, westeurope, southeastasia, etc.
# VM sizes: Standard_B1s, Standard_B2s, Standard_D2s_v3, etc.
```

---

## Advanced Scenarios

### Large-Scale Campaign

```bash
# Step 1: Create infrastructure once
osmedeus cloud create --provider digitalocean --instances 20

# Step 2: Run multiple scans without recreating infrastructure
osmedeus run -f general -T targets-batch-1.txt -c 10
osmedeus run -f repo -T repo-list.txt -c 5
osmedeus run -m subdomain-enumeration -T domains.txt -c 15

# Step 3: Destroy when done
osmedeus cloud destroy --all
```

### Custom Snapshots (Fast Boot)

```bash
# 1. Create a custom snapshot with pre-installed tools
#    (Manual: Boot VM, install tools, create snapshot)

# 2. Configure snapshot ID
osmedeus cloud config set providers.digitalocean.snapshot_id "123456789"

# 3. Workers now boot in ~30s instead of 5 minutes
osmedeus cloud run -f general -t example.com --instances 10
```

### Custom Worker Setup

```bash
# Add custom setup commands (runs on worker boot via cloud-init)
osmedeus cloud config set setup.commands[0] "apt-get update && apt-get install -y custom-tool"
osmedeus cloud config set setup.commands[1] "pip3 install custom-python-package"
osmedeus cloud config set setup.commands[2] "wget https://example.com/custom-binary -O /usr/local/bin/tool"

# Workers will execute these commands after osmedeus installation
osmedeus cloud run -f general -t example.com --instances 5
```

### Multi-Provider Strategy

```bash
# Use different providers for different tasks

# Fast enumeration on cheap DigitalOcean droplets
osmedeus cloud config set defaults.provider "digitalocean"
osmedeus cloud run -m subdomain-enumeration -T domains.txt --instances 10

# Heavy scanning on AWS spot instances
osmedeus cloud config set defaults.provider "aws"
osmedeus cloud config set providers.aws.use_spot true
osmedeus cloud run -m port-scanning -T targets.txt --instances 20

# SAST scans on GCP preemptible instances
osmedeus cloud config set defaults.provider "gcp"
osmedeus cloud config set providers.gcp.use_preemptible true
osmedeus cloud run -f repo -T repos.txt --instances 15
```

### Distributed Workflow Example

```bash
# Create master server (long-running)
osmedeus cloud create --instances 1 --provider digitalocean
# Note: Get master IP from output

# Configure master as Redis endpoint
export MASTER_IP="203.0.113.1"

# Create additional workers that connect to master
osmedeus cloud create --instances 10 --redis-url "redis://${MASTER_IP}:6379"

# Submit jobs to master
osmedeus run -f general -T targets.txt -c 20 --distributed

# Destroy all when done
osmedeus cloud destroy --all
```

### Budget-Constrained Scanning

```bash
# Set strict budget limits
osmedeus cloud config set limits.max_hourly_spend 1.00
osmedeus cloud config set limits.max_total_spend 10.00

# Use smallest instances
osmedeus cloud config set providers.digitalocean.size "s-1vcpu-1gb"

# Run with budget constraints
osmedeus cloud run -f general -t example.com --instances 5 --timeout 10h

# This will:
# - Block if 5 × s-1vcpu-1gb exceeds $1.00/hr
# - Terminate if accumulated cost exceeds $10.00
```

### Emergency Cleanup

```bash
# If osmedeus crashes or is interrupted, infrastructure may remain

# List all active infrastructure
osmedeus cloud list

# Destroy specific infrastructure
osmedeus cloud destroy cloud-do-1708234567

# Or destroy all
osmedeus cloud destroy --all --force

# Manual cleanup via provider (if needed)
doctl compute droplet list | grep osmedeus
doctl compute droplet delete <droplet-id>
```

---

## Real-World Examples

### Example 1: Bug Bounty Recon

```bash
# Quick subdomain enumeration for multiple programs
cat targets.txt
# hackerone.com
# bugcrowd.com
# intigriti.com

# Run distributed enumeration
osmedeus cloud run -m subdomain-enumeration -T targets.txt --instances 10 --timeout 2h

# Results stored in: ~/workspaces-osmedeus/*/
```

### Example 2: SAST Audit (100 Repositories)

```bash
# Create repo list
cat repos.txt
# https://github.com/org/repo1
# https://github.com/org/repo2
# ...

# Run SAST on all repos with 20 workers
osmedeus cloud run -f repo -T repos.txt -c 20 --instances 20 --timeout 6h

# Expected cost (DigitalOcean s-2vcpu-4gb):
# 20 × $0.02232/hr × 6 hours = $2.68
```

### Example 3: IP Range Scanning

```bash
# Create IP ranges file
cat cidr.txt
# 203.0.113.0/24
# 198.51.100.0/24

# Scan with nmap-based flow
osmedeus cloud run -f ip-scanning -T cidr.txt --instances 15 --timeout 4h

# Use spot instances for cost savings
osmedeus cloud config set providers.aws.use_spot true
osmedeus cloud run -f ip-scanning -T cidr.txt --provider aws --instances 15
```

### Example 4: Continuous Monitoring

```bash
# Schedule: Daily scan of critical assets

# Create infrastructure (keep running)
osmedeus cloud create --instances 5

# Run daily scans via cron
0 2 * * * osmedeus run -f general -T critical-assets.txt -c 5

# Weekly cost: 5 × $0.02232/hr × 168 hours = $18.75/week
# Monthly: ~$75

# Destroy when monitoring period ends
osmedeus cloud destroy --all
```

---

## Workflow Integration

### With CI/CD Pipeline

```yaml
# .github/workflows/security-scan.yml
name: Security Scan

on:
  push:
    branches: [main]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install osmedeus
        run: |
          wget https://github.com/j3ssie/osmedeus/releases/latest/download/osmedeus-linux-amd64
          chmod +x osmedeus-linux-amd64
          sudo mv osmedeus-linux-amd64 /usr/local/bin/osmedeus

      - name: Configure cloud
        env:
          DO_TOKEN: ${{ secrets.DIGITALOCEAN_TOKEN }}
        run: |
          osmedeus cloud config set providers.digitalocean.token "$DO_TOKEN"
          osmedeus cloud config set defaults.provider "digitalocean"

      - name: Run cloud scan
        run: |
          osmedeus cloud run -f general -t ${{ github.repository }} --instances 3 --timeout 1h

      - name: Upload results
        uses: actions/upload-artifact@v2
        with:
          name: scan-results
          path: ~/workspaces-osmedeus/
```

### With Slack Notifications

```bash
# Add webhook to config
export SLACK_WEBHOOK="https://hooks.slack.com/services/..."

# Run with notification wrapper
osmedeus cloud run -f general -t example.com --instances 5 && \
  curl -X POST $SLACK_WEBHOOK -d '{"text":"✅ Cloud scan completed for example.com"}' || \
  curl -X POST $SLACK_WEBHOOK -d '{"text":"❌ Cloud scan failed for example.com"}'
```

---

## Troubleshooting

### Common Issues

**Issue 1: "Provider credentials not configured"**
```bash
# Solution: Set provider credentials
osmedeus cloud config set providers.digitalocean.token "YOUR_TOKEN"
osmedeus cloud config show  # Verify
```

**Issue 2: "Max hourly spend exceeded"**
```bash
# Solution: Increase limit or reduce instances
osmedeus cloud config set limits.max_hourly_spend 10.00

# Or use smaller instances
osmedeus cloud config set providers.digitalocean.size "s-1vcpu-1gb"
```

**Issue 3: "Workers not registering"**
```bash
# Solution: Check cloud-init logs
ssh root@<worker-ip>
cat /var/log/cloud-init-output.log

# Manually join worker
osmedeus worker join --redis-url "redis://<master-ip>:6379" --get-public-ip
```

**Issue 4: "Infrastructure state not found"**
```bash
# List all states
ls -la ~/osmedeus-base/cloud-state/infrastructure/

# Manually load state
cat ~/osmedeus-base/cloud-state/infrastructure/cloud-do-*.json
```

### Debug Mode

```bash
# Run with debug output
osmedeus --debug cloud run -f general -t example.com --instances 3

# View detailed logs
tail -f ~/osmedeus-base/logs/osmedeus-*.log
```

### Cost Tracking

```bash
# Monitor real-time cost during execution
# (Output shows cost updates every 30 seconds)
osmedeus cloud run -f general -t example.com --instances 5

# Example output:
# [*] Cost: Elapsed 15m | Current Cost: $0.05 | Hourly Rate: $0.20/hr
```

---

## Best Practices

### 1. Start Small
```bash
# Test with 1-2 instances first
osmedeus cloud run -f general -t example.com --instances 2

# Scale up after verification
osmedeus cloud run -f general -T targets.txt --instances 20
```

### 2. Use Spot Instances
```bash
# Save 60-80% on costs
osmedeus cloud config set providers.aws.use_spot true
osmedeus cloud config set providers.gcp.use_preemptible true
```

### 3. Set Cost Limits
```bash
# Always set limits to prevent surprise bills
osmedeus cloud config set limits.max_hourly_spend 5.00
osmedeus cloud config set limits.max_total_spend 50.00
```

### 4. Clean Up Infrastructure
```bash
# Always destroy when done
osmedeus cloud run -f general -t example.com --instances 5
# (Auto-destroys after completion)

# Or manually destroy
osmedeus cloud destroy --all
```

### 5. Use Snapshots for Repeated Scans
```bash
# Create custom snapshot with tools pre-installed
# Reduces boot time from 5 minutes to 30 seconds
osmedeus cloud config set providers.digitalocean.snapshot_id "123456"
```

---

## Quick Reference

```bash
# Configuration
osmedeus cloud config set <key> <value>     # Set config value
osmedeus cloud config show                   # Show configuration

# Infrastructure
osmedeus cloud create --instances N          # Create infrastructure
osmedeus cloud list                          # List infrastructure
osmedeus cloud destroy <id>                  # Destroy infrastructure

# Workflow Execution
osmedeus cloud run -f <flow> -t <target> --instances N     # Run flow
osmedeus cloud run -m <module> -T <file> --instances N     # Run module on multiple targets

# Common Flags
--provider <name>        # Cloud provider (digitalocean, aws, gcp, linode, azure)
--mode <mode>            # Execution mode (vm, serverless)
--instances <n>          # Number of instances to provision
--timeout <duration>     # Timeout for cloud run (e.g., 2h, 30m)
--force                  # Skip confirmation prompts
-c <n>                   # Concurrent target scanning
```

---

## Next Steps

- 📖 Read [Cloud Architecture Guide](./cloud-architecture.md)
- 🧪 Review [Cloud Tests](../test/e2e/CLOUD_TESTS_README.md)
- 📝 Check [Cloud Config Template](../public/presets/cloud-settings.example.yaml)
- 🚀 Explore [Advanced Workflows](./workflows.md)

---

**Cost Calculator:**
- DigitalOcean s-2vcpu-4gb: $0.02232/hr × instances × hours
- AWS t3.medium: $0.0416/hr × instances × hours (spot: ~$0.0125/hr)
- GCP n1-standard-2: $0.095/hr × instances × hours (preemptible: ~$0.020/hr)

**Support:** https://github.com/j3ssie/osmedeus/issues

# Cloud Infrastructure Usage Guide

> 📚 **For detailed examples and advanced usage, see [Cloud Usage Examples](./cloud-usage-examples.md)**

This guide provides an overview of the cloud infrastructure feature. For comprehensive examples with copy-paste commands, detailed provider configurations, cost calculations, and troubleshooting, refer to the [Cloud Usage Examples](./cloud-usage-examples.md) documentation.

## Quick Start

### 1. Enable Cloud Features

Edit `~/osmedeus-base/osm-settings.yaml`:
```yaml
cloud:
  cloud_path: "{{base_folder}}/cloud"
  cloud_settings: "{{base_folder}}/cloud/cloud-settings.yaml"
  enabled: true  # Set to true
```

### 2. Configure Cloud Provider

```bash
# Set default provider
osmedeus cloud config set defaults.provider digitalocean

# Set DigitalOcean credentials
osmedeus cloud config set providers.digitalocean.token ${DIGITALOCEAN_TOKEN}
osmedeus cloud config set providers.digitalocean.region nyc1
osmedeus cloud config set providers.digitalocean.size s-2vcpu-4gb

# Set cost limits
osmedeus cloud config set limits.max_hourly_spend 10.0
osmedeus cloud config set limits.max_total_spend 100.0
```

Alternatively, manually create `~/osmedeus-base/cloud/cloud-settings.yaml` using the example in `docs/cloud-settings.example.yaml`.

### 3. Verify Configuration

```bash
# Show current config
osmedeus cloud config show

# Check estimated cost for 5 instances
# (will be implemented in cloud create command)
```

## Usage Scenarios

### Scenario 1: One-Off Distributed Scan

Provision infrastructure, run scan, collect results, and destroy in one command:

```bash
# Run general reconnaissance on example.com using 5 cloud workers
osmedeus cloud run -f general -t example.com --instances 5

# Or use multiple targets
osmedeus cloud run -f general -T targets.txt --instances 10
```

**What happens:**
1. Validates cost limits
2. Provisions 5 VMs on DigitalOcean
3. Waits for workers to register (auto-join via cloud-init)
4. Distributes workflow tasks across workers
5. Monitors progress
6. Collects results via SSH to master workspace
7. Destroys infrastructure
8. Shows final cost summary

### Scenario 2: Manual Infrastructure Management

For more control, manage infrastructure lifecycle manually:

```bash
# 1. Create infrastructure
osmedeus cloud create --instances 5

# 2. Verify workers joined
osmedeus worker status
# Should show 5 workers with wosm-<ip> IDs

# 3. Run workflows (uses existing workers)
osmedeus run -f general -t example.com
osmedeus run -m recon/httprobe -T targets.txt

# 4. List cloud infrastructure
osmedeus cloud list

# 5. When done, destroy infrastructure
osmedeus cloud destroy
```

### Scenario 3: Multi-Target Campaign

Run large-scale reconnaissance across many targets:

```bash
# Prepare target list
echo "hackerone.com" > targets.txt
echo "bugcrowd.com" >> targets.txt
echo "synack.com" >> targets.txt

# Create infrastructure with max instances
osmedeus cloud create --instances 20

# Run parallel scans (each target gets distributed to workers)
osmedeus run -f general -T targets.txt -c 5

# Monitor progress
osmedeus worker status

# Results are in ~/workspaces-osmedeus/
ls -lh ~/workspaces-osmedeus/

# Cleanup
osmedeus cloud destroy
```

### Scenario 4: Provider Override

Use a different provider for specific tasks:

```bash
# Create AWS infrastructure instead of default
osmedeus cloud create --provider aws --instances 3

# Or override in run command
osmedeus cloud run -f general -t example.com --provider gcp --instances 10
```

## Cost Management

### Understanding Costs

```bash
# DigitalOcean pricing examples:
# s-1vcpu-1gb:   $5/month  = $0.00744/hour
# s-2vcpu-4gb:   $15/month = $0.02232/hour
# s-4vcpu-8gb:   $30/month = $0.04464/hour

# For 5 x s-2vcpu-4gb instances:
# Hourly: 5 × $0.02232 = $0.1116/hour
# Daily:  $0.1116 × 24 = $2.6784/day
```

### Cost Limits

Limits are enforced at two stages:

1. **Pre-provisioning**: Checks `max_hourly_spend` before creating infrastructure
2. **During execution**: Checks `max_total_spend` every 30 seconds

```bash
# Set conservative limits for testing
osmedeus cloud config set limits.max_hourly_spend 0.5
osmedeus cloud config set limits.max_total_spend 5.0

# This will fail if it would exceed limits
osmedeus cloud create --instances 50
# Error: estimated hourly cost ($1.116) exceeds limit ($0.50)
```

### Cost Monitoring

During execution, you'll see cost updates:
```
[INFO] Creating Cloud Infrastructure
[INFO] Provider: digitalocean, Mode: vm, Instances: 5
[INFO] Estimated cost: $0.11/hour ($2.68/day)
[INFO] Provisioning 5 droplets...
[INFO] Waiting for workers... (3/5 registered)
[INFO] All workers ready!
[INFO] Running workflow...
[INFO] Cost: Elapsed: 0h 15m | Current: $0.03 | Rate: $0.11/hr
[SUCCESS] Workflow complete!
[INFO] Final cost: $0.05 (18 minutes)
```

## Worker Management

### Worker Auto-Registration

Cloud VMs automatically register as workers via cloud-init script:

```bash
#!/bin/bash
# Installed on boot by cloud provider

# Install osmedeus
curl -fsSL https://www.osmedeus.org/install.sh | bash

# Join master as worker
osmedeus worker join --redis-url redis://master:6379 --get-public-ip
# Worker ID: wosm-203.0.113.42
```

### Monitoring Workers

```bash
# List all workers
osmedeus worker status

# Example output:
# ID                 STATUS  TASKS  IP             JOINED
# wosm-203.0.113.1   idle    5/0    203.0.113.1    2m ago
# wosm-203.0.113.2   busy    3/0    203.0.113.2    2m ago
# wosm-203.0.113.3   idle    7/0    203.0.113.3    2m ago
```

### Debugging Worker Issues

```bash
# If workers don't register:
# 1. Check cloud infrastructure status
osmedeus cloud list

# 2. SSH into a VM manually
ssh root@<vm-ip>

# 3. Check osmedeus worker logs
journalctl -u osmedeus-worker -f

# 4. Check Redis connectivity
redis-cli -h <master-redis-ip> ping
```

## Advanced Configuration

### Custom Worker Setup

Add custom commands to run on worker boot:

```yaml
# In cloud-settings.yaml
setup:
  commands:
    - "apt-get update && apt-get install -y custom-tool"
    - "echo 'export CUSTOM_VAR=value' >> ~/.bashrc"
    - "cp /path/to/config /etc/config"
```

### Using Custom Snapshots

Pre-bake VMs with all tools installed for faster boot:

```bash
# 1. Create a VM manually
# 2. Install osmedeus and all tools
# 3. Create snapshot via provider console
# 4. Get snapshot ID

# Configure to use snapshot
osmedeus cloud config set providers.digitalocean.snapshot_id <snapshot-id>

# Now VMs boot with everything pre-installed
osmedeus cloud create --instances 5
# Boot time: ~30s instead of ~5min
```

### SSH Key Management

```bash
# Option 1: Use existing SSH key
osmedeus cloud config set ssh.private_key_path ~/.ssh/id_rsa

# Option 2: Generate new key for cloud workers
ssh-keygen -t rsa -b 4096 -f ~/.ssh/osmedeus-cloud -N ""
osmedeus cloud config set ssh.private_key_path ~/.ssh/osmedeus-cloud
osmedeus cloud config set ssh.public_key_path ~/.ssh/osmedeus-cloud.pub
```

## Troubleshooting

### Issue: Workers don't register

**Possible causes:**
1. Redis URL not reachable from worker VMs
2. Cloud-init script failed
3. Network/firewall blocking connection

**Solution:**
```bash
# Check infrastructure status
osmedeus cloud list

# SSH into a VM
ssh root@<vm-ip>

# Check cloud-init logs
tail -f /var/log/cloud-init-output.log

# Check worker process
ps aux | grep osmedeus
```

### Issue: Cost exceeded during execution

**What happens:**
- Infrastructure is immediately destroyed
- Partial results are collected
- Error message shows final cost

**Prevention:**
```bash
# Set realistic limits
osmedeus cloud config set limits.max_total_spend 50.0

# Estimate before running
# (5 instances × $0.02/hr × 2 hours = $0.20)
```

### Issue: Infrastructure not destroyed

**Recovery:**
```bash
# List all infrastructure
osmedeus cloud list

# Destroy by ID
osmedeus cloud destroy <infrastructure-id>

# Or destroy directly via provider console
# (check cloud-state/*.json for resource IDs)
```

## Best Practices

1. **Start small**: Test with 1-2 instances before scaling up
2. **Set cost limits**: Always configure max_hourly_spend and max_total_spend
3. **Use snapshots**: Pre-bake images for faster provisioning
4. **Clean up**: Always destroy infrastructure when done
5. **Monitor costs**: Check cloud provider billing dashboard
6. **Secure Redis**: Use password authentication for Redis in production
7. **SSH keys**: Use dedicated keys for cloud workers, not personal keys

## Examples

### Example 1: Quick Domain Recon
```bash
osmedeus cloud run -m recon/httprobe -t example.com --instances 3
```

### Example 2: Large-Scale Asset Discovery
```bash
# Prepare target list
cat targets.txt
# example1.com
# example2.com
# ...
# example100.com

# Run distributed scan
osmedeus cloud run -f general -T targets.txt --instances 20
```

### Example 3: Custom Workflow with Specific Provider
```bash
osmedeus cloud run \
  -f custom-workflow \
  -t target.com \
  --provider aws \
  --mode vm \
  --instances 10
```

## Current Limitations

⚠️ **Note**: As of this implementation, the following features are **foundational** and require completion:

1. **DigitalOcean droplet creation**: Pulumi program needs completion
2. **Cloud run workflow**: Task distribution and monitoring needs implementation
3. **Result collection**: SSH sync needs integration
4. **Other providers**: AWS, GCP, Linode, Azure need implementation

The CLI commands and infrastructure are in place, but will show "not yet fully implemented" errors until the above are completed.

## Getting Help

```bash
# Show cloud command help
osmedeus cloud --help

# Show specific subcommand help
osmedeus cloud config --help
osmedeus cloud create --help
osmedeus cloud run --help

# Check configuration
osmedeus cloud config show

# List available providers
# (Currently: digitalocean, aws, gcp, linode, azure)
```

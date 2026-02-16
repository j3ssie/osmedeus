# Osmedeus Cloud - Quick Reference Card

Quick reference for common `osmedeus cloud` commands. For detailed examples, see [cloud-usage-examples.md](./cloud-usage-examples.md).

---

## 🚀 Quick Start (30 seconds)

```bash
# 1. Set credentials
osmedeus cloud config set providers.digitalocean.token "YOUR_TOKEN"

# 2. Set default provider
osmedeus cloud config set defaults.provider "digitalocean"

# 3. Run first cloud scan
osmedeus cloud run -f general -t example.com --instances 3
```

---

## 📝 Configuration Commands

```bash
# View configuration
osmedeus cloud config show

# Set provider credentials
osmedeus cloud config set providers.digitalocean.token "dop_v1_..."
osmedeus cloud config set providers.aws.access_key_id "AKIA..."
osmedeus cloud config set providers.gcp.project_id "my-project"

# Set defaults
osmedeus cloud config set defaults.provider "digitalocean"
osmedeus cloud config set defaults.max_instances 10

# Set cost limits
osmedeus cloud config set limits.max_hourly_spend 5.00
osmedeus cloud config set limits.max_total_spend 50.00

# Set instance type
osmedeus cloud config set providers.digitalocean.size "s-2vcpu-4gb"
osmedeus cloud config set providers.aws.instance_type "t3.medium"

# Enable spot/preemptible instances (cost savings)
osmedeus cloud config set providers.aws.use_spot true
osmedeus cloud config set providers.gcp.use_preemptible true
```

---

## 🏗️ Infrastructure Management

```bash
# Create infrastructure
osmedeus cloud create --instances 5
osmedeus cloud create --provider aws --instances 10
osmedeus cloud create --instances 3 --force

# List active infrastructure
osmedeus cloud list

# Destroy infrastructure
osmedeus cloud destroy <infrastructure-id>
osmedeus cloud destroy --all
osmedeus cloud destroy <id> --force
```

---

## ▶️ Running Workflows

```bash
# Run flow on single target
osmedeus cloud run -f general -t example.com --instances 3

# Run module on single target
osmedeus cloud run -m subdomain-enumeration -t example.com --instances 5

# Run on multiple targets from file
osmedeus cloud run -f general -T targets.txt --instances 10

# Run with specific provider
osmedeus cloud run -f general -t example.com --provider aws --instances 5

# Run with concurrent targets
osmedeus cloud run -f general -T targets.txt -c 10 --instances 5

# Run with timeout
osmedeus cloud run -f general -t example.com --instances 3 --timeout 2h
```

---

## 💰 Cost Estimates (per hour)

### DigitalOcean
```
s-1vcpu-1gb:    $0.00744/hr  ($0.18/day)   [1 vCPU, 1GB RAM]
s-2vcpu-4gb:    $0.02232/hr  ($0.54/day)   [2 vCPU, 4GB RAM] ⭐ Default
s-4vcpu-8gb:    $0.04464/hr  ($1.07/day)   [4 vCPU, 8GB RAM]
s-8vcpu-16gb:   $0.08928/hr  ($2.14/day)   [8 vCPU, 16GB RAM]
```

### AWS (On-Demand)
```
t3.micro:       $0.0104/hr   ($0.25/day)   [2 vCPU, 1GB RAM]
t3.medium:      $0.0416/hr   ($1.00/day)   [2 vCPU, 4GB RAM] ⭐ Default
t3.large:       $0.0832/hr   ($2.00/day)   [2 vCPU, 8GB RAM]
t3.xlarge:      $0.1664/hr   ($4.00/day)   [4 vCPU, 16GB RAM]

Spot instances: ~70% savings
```

### GCP
```
f1-micro:       $0.0076/hr   ($0.18/day)   [0.6 vCPU, 0.6GB RAM]
n1-standard-1:  $0.0475/hr   ($1.14/day)   [1 vCPU, 3.75GB RAM]
n1-standard-2:  $0.0950/hr   ($2.28/day)   [2 vCPU, 7.5GB RAM] ⭐ Default
n1-standard-4:  $0.1900/hr   ($4.56/day)   [4 vCPU, 15GB RAM]

Preemptible: ~80% savings
```

**Cost Calculator:**
```
Total Cost = (Hourly Rate × Number of Instances × Runtime Hours)

Example:
5 × s-2vcpu-4gb × 2 hours = 5 × $0.02232 × 2 = $0.22
```

---

## 🎯 Common Workflows

### Bug Bounty Recon
```bash
osmedeus cloud run -m subdomain-enumeration -T targets.txt --instances 10 --timeout 2h
```

### Repository Audit
```bash
osmedeus cloud run -f repo -T repos.txt -c 20 --instances 20 --timeout 6h
```

### IP Range Scanning
```bash
osmedeus cloud run -f ip-scanning -T cidr.txt --instances 15 --timeout 4h
```

### Large Campaign (Persistent Infrastructure)
```bash
# Step 1: Create once
osmedeus cloud create --instances 20

# Step 2: Run multiple scans
osmedeus run -f general -T batch-1.txt -c 10
osmedeus run -f general -T batch-2.txt -c 10

# Step 3: Destroy when done
osmedeus cloud destroy --all
```

---

## 🌍 Provider Regions

### DigitalOcean
`nyc1, nyc2, nyc3, sfo1, sfo2, sfo3, ams2, ams3, sgp1, lon1, fra1, tor1, blr1`

### AWS
`us-east-1, us-west-2, eu-west-1, eu-central-1, ap-southeast-1, ap-northeast-1`

### GCP
`us-central1, us-east1, europe-west1, asia-southeast1`

---

## ⚙️ Environment Variables

```bash
# Set credentials via environment
export DIGITALOCEAN_TOKEN="dop_v1_..."
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="..."
export GCP_PROJECT_ID="my-project"

# Reference in config
osmedeus cloud config set providers.digitalocean.token '${DIGITALOCEAN_TOKEN}'
osmedeus cloud config set providers.aws.access_key_id '${AWS_ACCESS_KEY_ID}'
```

---

## 🔧 Troubleshooting

```bash
# Debug mode
osmedeus --debug cloud run -f general -t example.com --instances 3

# View logs
tail -f ~/osmedeus-base/logs/osmedeus-*.log

# List infrastructure states
ls -la ~/osmedeus-base/cloud-state/infrastructure/

# Emergency cleanup
osmedeus cloud destroy --all --force

# Manual provider cleanup
doctl compute droplet list | grep osmedeus
aws ec2 describe-instances --filters "Name=tag:osmedeus,Values=*"
gcloud compute instances list | grep osmedeus
```

---

## 🎨 Advanced Features

### Custom Snapshots
```bash
# Use pre-baked VM image (boot in 30s vs 5min)
osmedeus cloud config set providers.digitalocean.snapshot_id "123456789"
```

### Custom Worker Setup
```bash
# Run custom commands on worker boot
osmedeus cloud config set setup.commands[0] "apt-get update"
osmedeus cloud config set setup.commands[1] "pip3 install custom-tool"
```

### SSH Configuration
```bash
osmedeus cloud config set ssh.private_key_path "~/.ssh/cloud_rsa"
osmedeus cloud config set ssh.user "root"
```

---

## 📊 Cost Management Best Practices

1. **Always set limits:**
   ```bash
   osmedeus cloud config set limits.max_hourly_spend 5.00
   osmedeus cloud config set limits.max_total_spend 50.00
   ```

2. **Use spot/preemptible instances:**
   ```bash
   osmedeus cloud config set providers.aws.use_spot true        # 70% savings
   osmedeus cloud config set providers.gcp.use_preemptible true # 80% savings
   ```

3. **Start small, scale up:**
   ```bash
   # Test with 1-2 instances first
   osmedeus cloud run -f general -t example.com --instances 2
   ```

4. **Clean up after use:**
   ```bash
   osmedeus cloud destroy --all
   ```

5. **Monitor costs:**
   - Check cost estimates before creating infrastructure
   - Cloud run shows real-time cost tracking during execution

---

## 📚 More Information

- **Detailed Examples:** [cloud-usage-examples.md](./cloud-usage-examples.md)
- **Architecture:** [cloud-usage-guide.md](./cloud-usage-guide.md)
- **Tests:** [../test/e2e/CLOUD_TESTS_README.md](../test/e2e/CLOUD_TESTS_README.md)
- **Config Template:** [../public/presets/cloud-settings.example.yaml](../public/presets/cloud-settings.example.yaml)

---

## 🆘 Getting Help

```bash
# Command help
osmedeus cloud --help
osmedeus cloud config --help
osmedeus cloud create --help
osmedeus cloud run --help

# Full usage examples
osmedeus --usage-example

# GitHub Issues
https://github.com/j3ssie/osmedeus/issues
```

---

**Version:** v5.0+
**License:** MIT
**Author:** @j3ssie

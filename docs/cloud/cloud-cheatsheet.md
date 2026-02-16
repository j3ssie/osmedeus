# Osmedeus Cloud Cheatsheet

```
╔════════════════════════════════════════════════════════════════════════════╗
║                     OSMEDEUS CLOUD CHEATSHEET                              ║
║                    Distributed Security Scanning                           ║
╚════════════════════════════════════════════════════════════════════════════╝

┌─ QUICK START ──────────────────────────────────────────────────────────────┐
│                                                                             │
│  # Configure                                                                │
│  osmedeus cloud config set providers.digitalocean.token "YOUR_TOKEN"        │
│  osmedeus cloud config set defaults.provider "digitalocean"                 │
│                                                                             │
│  # Run                                                                      │
│  osmedeus cloud run -f general -t example.com --instances 3                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ CONFIGURATION ────────────────────────────────────────────────────────────┐
│                                                                             │
│  osmedeus cloud config show                      # View configuration       │
│  osmedeus cloud config set <key> <value>         # Set value               │
│                                                                             │
│  # Provider Credentials                                                     │
│  providers.digitalocean.token "dop_v1_..."                                  │
│  providers.aws.access_key_id "AKIA..."                                      │
│  providers.aws.secret_access_key "..."                                      │
│  providers.gcp.project_id "my-project"                                      │
│                                                                             │
│  # Defaults                                                                 │
│  defaults.provider "digitalocean|aws|gcp|linode|azure"                      │
│  defaults.max_instances 10                                                  │
│                                                                             │
│  # Cost Limits                                                              │
│  limits.max_hourly_spend 5.00                                               │
│  limits.max_total_spend 50.00                                               │
│                                                                             │
│  # Instance Types                                                           │
│  providers.digitalocean.size "s-2vcpu-4gb"                                  │
│  providers.aws.instance_type "t3.medium"                                    │
│  providers.gcp.machine_type "n1-standard-2"                                 │
│                                                                             │
│  # Cost Savings (70-80% off)                                                │
│  providers.aws.use_spot true                                                │
│  providers.gcp.use_preemptible true                                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ INFRASTRUCTURE MANAGEMENT ────────────────────────────────────────────────┐
│                                                                             │
│  osmedeus cloud create --instances 5              # Create                 │
│  osmedeus cloud create --provider aws --instances 10                        │
│  osmedeus cloud create --instances 3 --force                                │
│                                                                             │
│  osmedeus cloud list                              # List                   │
│                                                                             │
│  osmedeus cloud destroy <id>                      # Destroy                │
│  osmedeus cloud destroy --all                                               │
│  osmedeus cloud destroy <id> --force                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ RUNNING WORKFLOWS ────────────────────────────────────────────────────────┐
│                                                                             │
│  # Single target                                                            │
│  osmedeus cloud run -f general -t example.com --instances 3                 │
│  osmedeus cloud run -m subdomain-enum -t example.com --instances 5          │
│                                                                             │
│  # Multiple targets                                                         │
│  osmedeus cloud run -f general -T targets.txt --instances 10                │
│  osmedeus cloud run -f general -T targets.txt -c 10 --instances 5           │
│                                                                             │
│  # With provider                                                            │
│  osmedeus cloud run -f general -t example.com --provider aws --instances 5  │
│                                                                             │
│  # With timeout                                                             │
│  osmedeus cloud run -f general -t example.com --instances 3 --timeout 2h    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ COST ESTIMATES ───────────────────────────────────────────────────────────┐
│                                                                             │
│  ╔══════════════════╦═════════════╦══════════════╦═══════════════════════╗ │
│  ║ Provider         ║ Type        ║ Hourly       ║ Daily                 ║ │
│  ╠══════════════════╬═════════════╬══════════════╬═══════════════════════╣ │
│  ║ DigitalOcean     ║ s-1vcpu-1gb ║ $0.00744/hr  ║ $0.18/day             ║ │
│  ║ DigitalOcean ⭐   ║ s-2vcpu-4gb ║ $0.02232/hr  ║ $0.54/day             ║ │
│  ║ DigitalOcean     ║ s-4vcpu-8gb ║ $0.04464/hr  ║ $1.07/day             ║ │
│  ║ AWS              ║ t3.micro    ║ $0.0104/hr   ║ $0.25/day             ║ │
│  ║ AWS ⭐            ║ t3.medium   ║ $0.0416/hr   ║ $1.00/day             ║ │
│  ║ AWS (spot)       ║ t3.medium   ║ ~$0.0125/hr  ║ ~$0.30/day (-70%)     ║ │
│  ║ GCP              ║ n1-std-1    ║ $0.0475/hr   ║ $1.14/day             ║ │
│  ║ GCP ⭐            ║ n1-std-2    ║ $0.0950/hr   ║ $2.28/day             ║ │
│  ║ GCP (preempt)    ║ n1-std-2    ║ ~$0.0190/hr  ║ ~$0.46/day (-80%)     ║ │
│  ╚══════════════════╩═════════════╩══════════════╩═══════════════════════╝ │
│                                                                             │
│  Formula: Total = (Hourly Rate × Instances × Hours)                        │
│  Example: 5 × s-2vcpu-4gb × 2h = 5 × $0.02232 × 2 = $0.22                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ COMMON WORKFLOWS ─────────────────────────────────────────────────────────┐
│                                                                             │
│  # Bug Bounty Recon                                                         │
│  osmedeus cloud run -m subdomain-enum -T targets.txt --instances 10 \       │
│    --timeout 2h                                                             │
│                                                                             │
│  # Repository Audit (100 repos)                                             │
│  osmedeus cloud run -f repo -T repos.txt -c 20 --instances 20 --timeout 6h  │
│                                                                             │
│  # IP Range Scanning                                                        │
│  osmedeus cloud run -f ip-scanning -T cidr.txt --instances 15 --timeout 4h  │
│                                                                             │
│  # Large Campaign (Persistent)                                              │
│  osmedeus cloud create --instances 20                                       │
│  osmedeus run -f general -T batch-1.txt -c 10                               │
│  osmedeus run -f general -T batch-2.txt -c 10                               │
│  osmedeus cloud destroy --all                                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ PROVIDER REGIONS ─────────────────────────────────────────────────────────┐
│                                                                             │
│  DigitalOcean: nyc1, nyc3, sfo1, sfo3, ams3, sgp1, lon1, fra1, tor1, blr1  │
│  AWS:          us-east-1, us-west-2, eu-west-1, ap-southeast-1             │
│  GCP:          us-central1, us-east1, europe-west1, asia-southeast1        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ ENVIRONMENT VARIABLES ────────────────────────────────────────────────────┐
│                                                                             │
│  export DIGITALOCEAN_TOKEN="dop_v1_..."                                     │
│  export AWS_ACCESS_KEY_ID="AKIA..."                                         │
│  export AWS_SECRET_ACCESS_KEY="..."                                         │
│  export GCP_PROJECT_ID="my-project"                                         │
│                                                                             │
│  # Reference in config                                                      │
│  osmedeus cloud config set providers.digitalocean.token \                   │
│    '${DIGITALOCEAN_TOKEN}'                                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ TROUBLESHOOTING ──────────────────────────────────────────────────────────┐
│                                                                             │
│  osmedeus --debug cloud run ...              # Debug mode                  │
│  tail -f ~/osmedeus-base/logs/osmedeus-*.log # View logs                   │
│  osmedeus cloud list                         # List infrastructure         │
│  osmedeus cloud destroy --all --force        # Emergency cleanup           │
│                                                                             │
│  # Check states                                                             │
│  ls -la ~/osmedeus-base/cloud-state/infrastructure/                         │
│                                                                             │
│  # Manual cleanup                                                           │
│  doctl compute droplet list | grep osmedeus                                 │
│  aws ec2 describe-instances --filters "Name=tag:osmedeus,Values=*"          │
│  gcloud compute instances list | grep osmedeus                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ BEST PRACTICES ───────────────────────────────────────────────────────────┐
│                                                                             │
│  ✓ Always set cost limits (max_hourly_spend, max_total_spend)              │
│  ✓ Use spot/preemptible instances for 70-80% cost savings                  │
│  ✓ Start small (1-2 instances) then scale up                               │
│  ✓ Clean up infrastructure after use (cloud destroy --all)                 │
│  ✓ Use custom snapshots for faster boot (30s vs 5min)                      │
│  ✓ Monitor real-time cost during execution                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ ADVANCED FEATURES ────────────────────────────────────────────────────────┐
│                                                                             │
│  # Custom Snapshots (Fast Boot)                                             │
│  osmedeus cloud config set providers.digitalocean.snapshot_id "123456789"   │
│                                                                             │
│  # Custom Worker Setup                                                      │
│  osmedeus cloud config set setup.commands[0] "apt-get update"               │
│  osmedeus cloud config set setup.commands[1] "pip3 install custom-tool"     │
│                                                                             │
│  # SSH Configuration                                                        │
│  osmedeus cloud config set ssh.private_key_path "~/.ssh/cloud_rsa"          │
│  osmedeus cloud config set ssh.user "root"                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ FLAGS REFERENCE ──────────────────────────────────────────────────────────┐
│                                                                             │
│  --provider <name>        Cloud provider (digitalocean|aws|gcp|linode...)  │
│  --mode <mode>            Execution mode (vm|serverless)                   │
│  --instances <n>          Number of instances to provision                 │
│  --timeout <duration>     Timeout for cloud run (e.g., 2h, 30m)            │
│  --force                  Skip confirmation prompts                        │
│  -c <n>                   Concurrent target scanning                       │
│  -f <flow>                Flow name (general, repo, etc.)                  │
│  -m <module>              Module name (subdomain-enum, etc.)               │
│  -t <target>              Single target                                    │
│  -T <file>                Multiple targets from file                       │
│  --debug                  Enable debug output                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─ HELP & DOCUMENTATION ─────────────────────────────────────────────────────┐
│                                                                             │
│  osmedeus cloud --help                   # Cloud help                      │
│  osmedeus --usage-example                # Full examples                   │
│                                                                             │
│  Documentation:                                                             │
│  • docs/cloud-usage-examples.md          # Detailed examples               │
│  • docs/cloud-quick-reference.md         # Quick reference                 │
│  • docs/cloud-usage-guide.md             # Architecture guide              │
│  • test/e2e/CLOUD_TESTS_README.md        # Test documentation              │
│                                                                             │
│  GitHub: https://github.com/j3ssie/osmedeus/issues                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

╔════════════════════════════════════════════════════════════════════════════╗
║  Version: v5.0+ | License: MIT | Author: @j3ssie                           ║
╚════════════════════════════════════════════════════════════════════════════╝
```

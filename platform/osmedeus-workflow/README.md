# Community Workflow for Osmedeus

<p align="center">
  <a href="https://github.com/j3ssie/osmedeus"><img alt="Osmedeus" src="https://raw.githubusercontent.com/osmedeus/assets/main/osm-logo-with-white-border.png" height="120" /></a>
  <br />
  <strong>A basic reconnaissance methodology workflow for the <a href="https://github.com/j3ssie/osmedeus">Osmedeus Engine</a></strong>
</p>

This repository provides a reference workflow implementation demonstrating basic reconnaissance methodology. Use it as a starting point to understand Osmedeus workflows and build your own custom automation pipelines.

## Installation

```bash
osmedeus install workflow https://github.com/osmedeus/osmedeus-workflow.git
```

See [Osmedeus documentation](https://docs.osmedeus.org/workflows/overview) for more details.

## More Examples

For additional workflow examples and patterns, see the [test workflows](https://github.com/j3ssie/osmedeus/tree/main/test/testdata/workflows) in the main Osmedeus repository.

## Folder Structure

```
.
├── common/              # Reusable module workflows
├── events/              # Event-driven workflows
├── fragments/           # Fragments used by workflows
├── cidr.yaml            # CIDR/IP range workflow
├── cidr-extensive.yaml  # Extended CIDR workflow
├── domain-lite.yaml     # Lightweight domain recon
├── domain-standard.yaml # Standard domain recon
├── domain-extensive.yaml# Extended domain recon
├── general.yaml         # Full reconnaissance flow
├── repo.yaml            # Repository scanning flow
├── sast.yaml            # SAST scanning flow
├── url.yaml             # URL-based recon flow
└── web-analysis.yaml    # Web analysis flow
```

## Reconnaissance Methodology

The workflow follows a phased approach to reconnaissance:

```
┌─────────────────┐
│   Subdomain     │  Phase 1: Discover subdomains using multiple sources
│   Enumeration   │  (subfinder, findomain, assetfinder, amass)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Probing      │  Phase 2: DNS resolution and HTTP probing
│  (DNS + HTTP)   │  (puredns, massdns, pd-httpx, dnsx)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Fingerprint    │  Phase 3: Technology detection and fingerprinting
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┐
    ▼         ▼          ▼          ▼
┌───────┐ ┌───────┐ ┌─────────┐ ┌─────────┐
│Screen │ │Archive│ │IP Space │ │Portscan │  Phase 4+: Parallel analysis
│ shot  │ │       │ │  Enum   │ │         │
└───┬───┘ └───┬───┘ └────┬────┘ └────┬────┘
    │         │          │           │
    └─────────┴──────────┴───────────┘
              │
    ┌─────────┴───────────────┐
    ▼                         ▼
┌─────────────────┐      ┌───────────┐
│Vulnerability    │      │ Content   │  Final: Vulnerability and content discovery
│ Scanning        │      │ Discovery │
└─────────────────┘      └───────────┘
```

## Available Workflows

### Flow Workflows

| Workflow | Description |
|----------|-------------|
| `general.yaml` | Full reconnaissance pipeline with all phases |
| `domain-lite.yaml` | Lightweight domain reconnaissance |
| `domain-standard.yaml` | Standard domain reconnaissance |
| `domain-extensive.yaml` | Extended domain reconnaissance |
| `cidr.yaml` | CIDR/IP range reconnaissance |
| `cidr-extensive.yaml` | Extended CIDR reconnaissance with additional phases |
| `url.yaml` | URL-based reconnaissance workflow |
| `web-analysis.yaml` | Web application analysis workflow |
| `repo.yaml` | Source repository scanning workflow |
| `sast.yaml` | Static application security testing workflow |

### Module Workflows (common/)

| Module | Description |
|--------|-------------|
| `enum-subdomain.yaml` | Subdomain enumeration (subfinder, findomain, assetfinder) |
| `probe-dns.yaml` | DNS resolution and probing |
| `recon-http-fp.yaml` | HTTP fingerprinting and technology detection |
| `recon-screenshot.yaml` | Visual screenshots of discovered assets |
| `util-archive.yaml` | Archive/wayback machine data collection |
| `enum-ipspace.yaml` | IP space enumeration |
| `probe-port.yaml` | Port scanning |
| `scan-vuln.yaml` | Vulnerability scanning |
| `scan-vuln-thorough.yaml` | Thorough Vigolium vulnerability scanning |
| `scan-content.yaml` | Directory and content bruteforcing |
| `recon-spider.yaml` | Web spidering/crawling |

### Event Workflows (events/)

| Event | Description |
|-------|-------------|
| `simple-emitter.yaml` | Simple event emitter example |
| `simple-receiver.yaml` | Simple event receiver example |
| `vuln-scan-receiver.yaml` | Vulnerability scan event receiver |

### Fragments (fragments/)

| Fragment | Description |
|----------|-------------|
| `do-enum-subdomain.yaml` | Subdomain enumeration flow fragment |
| `do-recon-http-fp.yaml` | HTTP fingerprinting fragment |
| `do-recon-spider.yaml` | Web spidering fragment |
| `do-probe-port.yaml` | Port scan fragment |
| `do-scan-content.yaml` | Content discovery fragment |
| `do-scan-vuln.yaml` | Vulnerability scan fragment |
| `do-scan-vuln-thorough.yaml` | Thorough Vigolium vulnerability scan fragment |
| `do-deep-vuln-scan.yaml` | Deep vulnerability scan fragment |
| `do-scan-repo.yaml` | Repository scanning fragment |
| `do-util-normalize.yaml` | Normalization utility fragment |
| `do-util-prepare-repo.yaml` | Repository preparation utility fragment |

## Usage

```bash
# Run the general reconnaissance flow
osmedeus run -f general -t example.com

# Run the fast reconnaissance flow
osmedeus run -f fast -t example.com

# Run a specific module
osmedeus run -m subdomain-enum -t example.com

# Dry-run to preview execution
osmedeus run -f general -t example.com --dry-run
```

## Building Your Own Workflow

1. **Study the common modules** - Each module in `common/` demonstrates a specific recon phase
2. **Understand the flow structure** - See `general.yaml` for how modules are orchestrated with dependencies
3. **Customize parameters** - Modules accept params for threads, wordlists, and toggles
4. **Chain modules** - Use `depends_on` to create execution dependencies

Example module structure:

```yaml
kind: module
name: my-module
description: Description of what this module does

params:
  - name: customParam
    default: "value"

dependencies:
  commands:
    - tool1
    - tool2

steps:
  - name: step-one
    type: bash
    command: 'tool1 -t {{Target}} -o {{Output}}/results.txt'
```


## Documentation

- [Osmedeus Documentation](https://docs.osmedeus.org/)
- [Workflow Overview](https://docs.osmedeus.org/workflows/overview)
- [CLI Reference](https://docs.osmedeus.org/getting-started/cli)

## License

Osmedeus is made with ♥ by [@j3ssie](https://twitter.com/j3ssie) and it is released under the MIT license.

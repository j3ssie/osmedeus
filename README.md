# Osmedeus

<p align="center">
  <a href="https://www.osmedeus.org"><img alt="Osmedeus" src="https://raw.githubusercontent.com/osmedeus/assets/main/osm-logo-with-white-border.png" height="140" /></a>
  <br />
  <strong>A modern orchestration engine for security</strong>

  <p align="center">
  <a href="https://docs.osmedeus.org/"><img src="https://img.shields.io/badge/Documentation-0078D4?style=for-the-badge&logo=GitBook&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://docs.osmedeus.org/donation/"><img src="https://img.shields.io/badge/Sponsors-0078D4?style=for-the-badge&logo=GitHub-Sponsors&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://twitter.com/OsmedeusEngine"><img src="https://img.shields.io/badge/%40OsmedeusEngine-0078D4?style=for-the-badge&logo=Twitter&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://discord.gg/mtQG2FQsYA"><img src="https://img.shields.io/badge/Discord%20Server-0078D4?style=for-the-badge&logo=Discord&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://github.com/j3ssie/osmedeus/releases"><img src="https://img.shields.io/github/release/j3ssie/osmedeus?style=for-the-badge&labelColor=black&color=2fc414&logo=Github"></a>
  </p>
</p>

## What is Osmedeus?

[Osmedeus](https://www.osmedeus.org) is a security-focused declarative orchestration engine. It defines workflow automation in YAML you can read and audit, with encrypted data handling, credential management, and sandboxed execution.

A workflow composes from modules, so the same YAML runs on your host, in Docker, or over SSH.

## Key features

- **Declarative YAML workflows.** Define pipelines with hooks, decision routing, module exclusion, and conditional branching across the host, Docker, and SSH runners.
- **Distributed execution.** A Redis-backed master and worker setup with a task queue, webhook triggers, and file sync between workers.
- **Function library.** 80+ utility functions, including nmap integration, tmux sessions, SSH execution, TypeScript and Python scripting, SARIF parsing, and CDN and WAF classification.
- **Event-driven scheduling.** Cron, file-watch, and event triggers with filtering, deduplication, and delayed task queues.
- **Agentic LLM steps.** Tool-calling agent loops with sub-agent orchestration, memory management, and structured output, plus ACP subprocess agents (Claude Code, Codex, OpenCode, Gemini).
- **Cloud infrastructure.** Provision machines and run scans on DigitalOcean, AWS, GCP, Linode, and Azure, with cost controls and automatic cleanup.
- **CLI.** Interactive database queries, bulk function evaluation, workflow linting, progress bars, and usage examples for every command.
- **REST API and web UI.** An API server with webhook triggers and database queries, and an embedded dashboard that renders the results.

See [Documentation Page](https://docs.osmedeus.org/) for more details.

## Installation

```bash
curl -sSL http://www.osmedeus.org/install.sh | bash
```

### [npm](https://www.npmjs.com/package/@j3ssie/osmedeus)

```bash
npm install -g @j3ssie/osmedeus
```

The npm package ships prebuilt binaries for Linux and macOS on x64 and arm64.

See [Quickstart](https://docs.osmedeus.org/quickstart/) for setup and [Installation](https://docs.osmedeus.org/installation/) for advanced configuration.

| CLI Usage | Web UI Assets |
|-----------|---------------|
| ![CLI Usage](https://github.com/osmedeus/docs/blob/main/images/cli/cli-run-with-verbose-output.png?raw=true) | ![Web UI Assets](https://github.com/osmedeus/docs/blob/main/images/web-ui/web-ui-assets.png?raw=true) |
| **Workflow Visualization** | **Vulnerabilities** |
| ![Workflow Visualization](https://github.com/osmedeus/docs/blob/main/images/web-ui/web-ui-workflow.png?raw=true) | ![Vulnerabilities](https://github.com/osmedeus/docs/blob/main/images/web-ui/web-ui-vuln.png?raw=true) |


## Sponsor: Infistar.cc

[![Osmedeus and Infistar. Claude, GPT, Gemini, Grok, DeepSeek and more on one key at a fraction of official prices](assets/infistar-banner.png)](https://infistar.cc/register?aff=OSMEDEUS&ref_source=link)

[Infistar.cc](https://infistar.cc/register?aff=OSMEDEUS&ref_source=link) serves Claude, GPT, Gemini, Grok, DeepSeek and more on one API key, at a fraction of official prices.

- **Up to 99% off.** Pricing starts at 0.1x the official rate.
- **Official upstreams and 100% genuine models.** No downgrades, no dilution, and every model is verifiable.
- **1,000+ redundant routes with automatic failover.** When one route fails, the next takes over.
- **200+ models from 16 brands.** Text, image, and video on one key.
- **Recommended by 100+ open source projects**, the largest with 100k+ stars.
- **Built for Osmedeus.** Run the agentic LLM steps and multi-model security automation against one API key.
- **¥5 free for Osmedeus users** who sign up through the [exclusive link](https://infistar.cc/register?aff=OSMEDEUS&ref_source=link).

[Try Infistar.cc](https://infistar.cc/register?aff=OSMEDEUS&ref_source=link)


## Quick start

```bash
# Run a module workflow
osmedeus run -m recon -t example.com

# Run a flow workflow
osmedeus run -f general -t example.com

# Multiple targets with concurrency
osmedeus run -m recon -T targets.txt -c 5

# Dry-run mode (preview)
osmedeus run -f general -t example.com --dry-run

# Start API server
osmedeus serve

# List available workflows
osmedeus workflow list

# Query discovered assets
osmedeus assets -w example.com                          # List assets for workspace
osmedeus assets --stats                                 # Show unique technologies, sources, types
osmedeus assets --source httpx --type web --json        # Filter and output as JSON

# Query vulnerabilities, runs, and steps
osmedeus query vulns --severity high --workspace example.com
osmedeus query runs --status running
osmedeus query steps --run <run-uuid>

# Query database tables
osmedeus db list --table runs
osmedeus db list --table event_logs --search "nuclei"

# Evaluate utility functions
osmedeus func eval 'log_info("hello")'
osmedeus func eval -e 'http_get("https://example.com")' -T targets.txt -c 10

# Platform variables available in eval
osmedeus func eval 'log_info("OS: " + PlatformOS + ", Arch: " + PlatformArch)'

# Install from preset repositories
osmedeus install base --preset
osmedeus install base --preset --keep-setting   # preserve existing osm-settings.yaml
osmedeus install workflow --preset

# Exclude modules from flow execution
osmedeus run -f general -t example.com -x portscan
osmedeus run -f general -t example.com -X vuln    # Fuzzy exclude by substring

# Worker queue system
osmedeus worker queue new -f general -t example.com   # Queue for later
osmedeus worker queue run --concurrency 5              # Process queue

# Worker management
osmedeus worker status                          # Show workers
osmedeus worker eval -e 'ssh_exec("host", "whoami")'  # Eval with distributed hooks

# Run an ACP agent interactively
osmedeus agent "analyze this codebase"
osmedeus agent --agent codex "explain main.go"
osmedeus agent --list

# Cloud infrastructure management
osmedeus cloud create --instances 3                    # Provision cloud machines
osmedeus cloud setup 1.2.3.4 5.6.7.8                  # Setup existing machines
osmedeus cloud list                                    # List active infrastructure
osmedeus cloud run -f general -t example.com --instances 3

# Show all usage examples
osmedeus --usage-example
```

## Docker

```bash
# Show help
docker run --rm j3ssie/osmedeus:latest --help

# Run a scan
docker run --rm -v $(pwd)/output:/root/workspaces-osmedeus \
    j3ssie/osmedeus:latest run -f general -t example.com
```

For more CLI usage and example commands, refer to the [CLI Reference](https://docs.osmedeus.org/getting-started/cli).

## High-level architecture

```plaintext
┌───────────────────────────────────────────────────────────────────────────┐
│                   Osmedeus Orchestration Engine                           │
├───────────────────────────────────────────────────────────────────────────┤
│  ENTRY POINTS                                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐                │
│  │   CLI    │  │ REST API │  │Scheduler │  │ Distributed │                │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └─────┬───────┘                │
│       └─────────────┴─────────────┴──────────────┘                        │
│                              │                                            │
│                              ▼                                            │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │ CONFIG ──▶ PARSER ──▶ EXECUTOR ──▶ STEP DISPATCHER ──▶ RUNNER       │  │
│  │                          │                                          │  │
│  │  Step Executors: bash | function | parallel | foreach | remote-bash │  │
│  │                  http | llm | agent | agent-acp | SARIF/SAST       │  │
│  │  Hooks: pre_scan_steps → [main steps] → post_scan_steps             │  │
│  │                          │                                          │  │
│  │  Runners: HostRunner | DockerRunner | SSHRunner                     │  │
│  │  Queue: DB + Redis polling → dedup → concurrent execution           │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────────────┘
```

For more information about the architecture, refer to the [Architecture Documentation](https://docs.osmedeus.org/architecture).

## Roadmap and status

The plan for the project, in order:

|  #  | Step                                                                        |  Status |
| :-: | --------------------------------------------------------------------------- |  :----: |
|  1  | Rebuild the Osmedeus engine on a new architecture                           |    ✅   |
|  2  | Flexible workflows and step types                                           |    ✅   |
|  3  | Event-driven model and the trigger event categories                         |    ✅   |
|  4  | Web UI for results and workflow diagrams                                    |    ✅   |
|  5  | Rewrite the workflows for the new architecture and syntax                   |    ✅   |
|  6  | Test more utility functions, such as notifications                          |    ✅   |
|  7  | SAST integration with SARIF parsing (Semgrep, Trivy, and others)            |    ✅   |
|  8  | Cloud integration that runs a scan on a cloud provider                      |    ✅   |
|  9  | Diff reports listing new, removed, and unchanged assets between runs        |    ❌   |
|  10 | A step type that runs on a cloud provider's serverless platform             |    ❌   |
|  N  | Further features, to be decided                                             |    ❌   |

## Documentation

| Topic                | Link                                                                                                     |
|----------------------|----------------------------------------------------------------------------------------------------------|
| Getting Started      | [docs.osmedeus.org/getting-started](https://docs.osmedeus.org/getting-started) |
| CLI Usage & Examples | [docs.osmedeus.org/getting-started/cli](https://docs.osmedeus.org/getting-started/cli) |
| Writing Workflows    | [docs.osmedeus.org/workflows/overview](https://docs.osmedeus.org/workflows/overview) |
| Event-Driven Triggers| [docs.osmedeus.org/advanced/event-driven](https://docs.osmedeus.org/advanced/event-driven) |
| Deployment           | [docs.osmedeus.org/deployment](https://docs.osmedeus.org/deployment) |
| Architecture         | [docs.osmedeus.org/concepts/architecture](https://docs.osmedeus.org/concepts/architecture) |
| Development          | [docs.osmedeus.org/development](https://docs.osmedeus.org/development) and [HACKING.md](HACKING.md) |
| Extending Osmedeus   | [docs.osmedeus.org/development/extending-osmedeus](https://docs.osmedeus.org/development/extending-osmedeus)   |
| Security Model       | [SECURITY.md](SECURITY.md) and [docs.osmedeus.org/others/security-warning](https://docs.osmedeus.org/others/security-warning) |
| Full Documentation   | [docs.osmedeus.org](https://docs.osmedeus.org) |

## Disclaimer

Osmedeus executes arbitrary code and commands from user-supplied input through the CLI, the API, and workflow definitions. That is intentional and central to how the engine operates.

Read the [Security Warning](https://docs.osmedeus.org/others/security-warning) page for how to stay safe. [SECURITY.md](SECURITY.md) draws the line between the code execution that is intentional and what counts as a vulnerability, and is where to start if you want to report one.

**Think twice before you:**
- Run workflows downloaded from untrusted sources
- Execute commands or scans against targets you don't own or have permission to test
- Use workflows without reviewing their contents first

You are responsible for what you run. Always review workflow YAML files before execution, especially those obtained from third parties.

## License

Osmedeus is made with ♥ by [@j3ssie](https://x.com/j3ssie) and it is released under the MIT license.

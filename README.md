# Osmedeus

<p align="center">
  <a href="https://www.osmedeus.org"><img alt="Osmedeus" src="https://raw.githubusercontent.com/osmedeus/assets/main/osm-logo-with-white-border.png" height="140" /></a>
  <br />
  <strong>Osmedeus - A Modern Orchestration Engine for Security</strong>

  <p align="center">
  <a href="https://docs.osmedeus.org/"><img src="https://img.shields.io/badge/Documentation-0078D4?style=for-the-badge&logo=GitBook&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://docs.osmedeus.org/donation/"><img src="https://img.shields.io/badge/Sponsors-0078D4?style=for-the-badge&logo=GitHub-Sponsors&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://twitter.com/OsmedeusEngine"><img src="https://img.shields.io/badge/%40OsmedeusEngine-0078D4?style=for-the-badge&logo=Twitter&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://discord.gg/gy4SWhpaPU"><img src="https://img.shields.io/badge/Discord%20Server-0078D4?style=for-the-badge&logo=Discord&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://github.com/j3ssie/osmedeus/releases"><img src="https://img.shields.io/github/release/j3ssie/osmedeus?style=for-the-badge&labelColor=black&color=2fc414&logo=Github"></a>
  </p>
</p>

## What is Osmedeus?

[Osmedeus](https://www.osmedeus.org) is a security focused declarative orchestration engine that simplifies complex workflow automation into auditable YAML definitions, complete with encrypted data handling, secure credential management, and sandboxed execution.

Built for both beginners and experts, it delivers powerful, composable automation without sacrificing the integrity and safety of your infrastructure.

## Features

- **Declarative YAML Workflows** - Define reconnaissance pipelines using simple, readable YAML syntax
- **Two Workflow Types** - Modules for single execution units, Flows for multi-module orchestration
- **Multiple Runners** - Execute on local host, Docker containers, or remote machines via SSH
- **Distributed Execution** - Scale with Redis-based master-worker pattern for parallel scanning
- **Event-Driven Triggers** - Cron scheduling, file watching, and event-based workflow triggers
- **Decision Routing** - Conditional workflow branching with switch/case syntax
- **Template Engine** - Powerful variable interpolation with built-in and custom variables
- **Utility Functions** - Rich function library for file operations, string manipulation, and JSON processing
- **REST API Server** - Manage and trigger workflows programmatically
- **Database Support** - SQLite (default) and PostgreSQL for asset tracking
- **Notifications** - Telegram bot and webhook integrations
- **Cloud Storage** - S3-compatible storage for artifact management
- **LLM Integration** - AI-powered workflow steps with chat completions and embeddings

See [Documentation Page](https://docs.osmedeus.org/) for more details.

## Installation

```bash
curl -sSL http://www.osmedeus.org/install.sh | bash
```

See [Quickstart](https://docs.osmedeus.org/quickstart/) for quick setup and [Installation](https://docs.osmedeus.org/installation/) for advanced configurations.

## Quick Start

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

# Show all usage examples
osmedeus --usage-example
```

## Docker

```bash
# Show help
docker run --rm osmedeus:latest --help

# Run a scan
docker run --rm -v $(pwd)/output:/root/workspaces-osmedeus \
    osmedeus:latest run -f general -t example.com
```

For more CLI usage and example commands, refer to the [CLI Reference](https://docs.osmedeus.org/getting-started/cli).


| CLI Usage | Web UI Assets | Web UI Workflow |
|-----------|--------------|-----------------|
| ![CLI Usage](https://raw.githubusercontent.com/osmedeus/assets/refs/heads/main/demo-images/cli-run-with-verbose-output.png) | ![Web UI Assets](https://raw.githubusercontent.com/osmedeus/assets/refs/heads/main/demo-images/web-ui-assets.png) | ![Web UI Workflow](https://raw.githubusercontent.com/osmedeus/assets/refs/heads/main/demo-images/web-ui-workflow.png) |

## Core Components

### Trigger

| Type | Description | Use Case |
|------|-------------|----------|
| **Cron** | Schedule workflows at specific times | Regular scans |
| **File Watch** | Trigger workflows when files change | Continuous monitoring |
| **Event** | Trigger workflows based on external events | Integration with other tools |
| **Webhook** | Trigger workflows based on HTTP requests | External system integration |
| **Manual** | Trigger workflows manually via CLI or API | One-time tasks |

### Workflows

| Type | Description | Use Case |
|------|-------------|----------|
| **Module** | Single execution unit with sequential/parallel steps | Individual scanning tasks |
| **Flow** | Orchestrates multiple modules with dependencies | Complete reconnaissance pipelines |

### Runners

| Runner | Description |
|--------|-------------|
| **Host** | Local machine execution (default) |
| **Docker** | Container-based execution |
| **SSH** | Remote machine execution |

### Step Types

| Type | Description |
|------|-------------|
| `bash` | Execute shell commands |
| `function` | Call utility functions |
| `foreach` | Iterate over file contents |
| `parallel-steps` | Run multiple steps concurrently |
| `remote-bash` | Per-step Docker/SSH execution |
| `http` | Make HTTP requests |
| `llm` | AI-powered processing |

### Workflow Example

```yaml
kind: module
name: demo-bash
description: Demo bash steps with functions and exports

params:
  - name: target
    required: true

steps:
  - name: setup
    type: bash
    command: mkdir -p {{Output}}/demo && echo "{{Target}}" > {{Output}}/demo/target.txt
    exports:
      target_file: "{{Output}}/demo/target.txt"

  - name: run-parallel
    type: bash
    parallel_commands:
      - 'echo "Thread 1: {{Target}}" >> {{Output}}/demo/results.txt'
      - 'echo "Thread 2: {{Target}}" >> {{Output}}/demo/results.txt'

  - name: check-result
    type: function
    function: 'fileLength("{{Output}}/demo/results.txt")'
    exports:
      line_count: "output"

  - name: summary
    type: bash
    command: 'echo "Processed {{Target}} with {{line_count}} lines"'

```

For writing your first workflow, refer to the [Workflow Overview](https://docs.osmedeus.org/workflows/overview).

## Roadmap and Status

The high-level ambitious plan for the project, in order:

|  #  | Step                                                                          | Status |
| :-: | ----------------------------------------------------------------------------- | :----: |
|  1  | Osmedeus Engine reforged with a next-generation architecture                  |   ✅   |
|  2  | Flexible workflows and step types                                             |   ✅   |
|  3  | Beautiful UI for visualize results and workflow diagram                       |   ✅   |
|  4  | Rewriting the workflow to adapt to new architecture and syntax                |   ⚠️   |
|  5  | Testing more utility functions like notifications                             |   ⚠️   |
|  6  | Generate diff reports showing new/removed/unchanged assets between runs.      |   ❌   |
|  7  | Adding step type from cloud provider that can be run via serverless           |   ❌   |
|  N  | Fancy features (to be expanded upon later)                                    |   ❌   |

## Documentation

| Topic                | Link                                                                                                     |
|----------------------|----------------------------------------------------------------------------------------------------------|
| Getting Started      | [docs.osmedeus.org/getting-started](https://docs.osmedeus.org/getting-started) |
| CLI Usage & Examples | [docs.osmedeus.org/getting-started/cli](https://docs.osmedeus.org/getting-started/cli) |
| Writing Workflows    | [docs.osmedeus.org/workflows/overview](https://docs.osmedeus.org/workflows/overview) |
| Deployment           | [docs.osmedeus.org/deployment](https://docs.osmedeus.org/deployment) |
| Architecture         | [docs.osmedeus.org/concepts/architecture](https://docs.osmedeus.org/concepts/architecture) |
| Development          | [docs.osmedeus.org/development](https://docs.osmedeus.org/development) and [HACKING.md](HACKING.md) |
| Extending Osmedeus   | [docs.osmedeus.org/development/extending-osmedeus](https://docs.osmedeus.org/development/extending-osmedeus)   |
| Full Documentation   | [docs.osmedeus.org](https://docs.osmedeus.org) |

## License

Osmedeus is made with ♥ by [@j3ssie](https://twitter.com/j3ssie) and it is released under the MIT license.

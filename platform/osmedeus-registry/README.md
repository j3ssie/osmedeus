# Osmedeus Registry Metadata

[![Osmedeus](https://img.shields.io/badge/Osmedeus-Registry-39ff14?style=for-the-badge&logo=github&logoColor=white&labelColor=black)](https://github.com/j3ssie/osmedeus)

This repository generates and maintains **registry metadata** for [Osmedeus](https://github.com/j3ssie/osmedeus) third-party tool installation. It powers the `osmedeus install binary` command by providing download URLs, version information, and installation instructions for security tools.

## Overview

When you run:

```bash
osmedeus install binary
```

Osmedeus fetches metadata from this registry to:

1. **Discover available tools** — List installable security tools with descriptions
2. **Download binaries** — Fetch pre-built binaries for your OS/architecture
3. **Execute installation** — Run package manager commands or custom install scripts
4. **Track versions** — Show current tool versions from upstream releases

## Quick Start

### Update All Tool Versions

```bash
# Set GitHub token for higher rate limits (optional but recommended)
export GH_TOKEN="your_github_token"

# Update all tools
python3 update_registry_metadata.py

# Update a specific tool
python3 update_registry_metadata.py --tool subfinder
```

### Requirements

- Python 3.6+
- `requests` library (`pip install requests`)
- GitHub API access (token recommended for rate limits)

## Registry Format

### registry-metadata-direct-fetch.json

The registry defines tools with three installation methods:

#### 1. GitHub Releases (Direct Download)

```json
{
  "subfinder": {
    "desc": "Fast passive subdomain enumeration tool.",
    "version": "2.12.0",
    "repo_link": "https://github.com/projectdiscovery/subfinder",
    "package-manager": "github-release",
    "tags": ["recon", "subdomain", "passive"],
    "valide-command": "",
    "linux": {
      "amd64": "https://github.com/.../subfinder_2.12.0_linux_amd64.zip",
      "arm64": "https://github.com/.../subfinder_2.12.0_linux_arm64.zip"
    },
    "darwin": {
      "amd64": "https://github.com/.../subfinder_2.12.0_macOS_amd64.zip",
      "arm64": "https://github.com/.../subfinder_2.12.0_macOS_arm64.zip"
    }
  }
}
```

#### 2. Go Getter (go install)

```json
{
  "dalfox": {
    "desc": "Powerful open-source XSS scanner",
    "package-manager": "go-getter",
    "command-dual": {
      "dual": "github.com/hahwul/dalfox/v2@latest"
    }
  }
}
```

#### 3. Package Manager / Custom Commands

```json
{
  "curl": {
    "desc": "Command line tool for transferring data with URL syntax",
    "package-manager": "<auto_detect_package_manager>",
    "multi-commands-linux": [
      "<auto_detect_package_manager> install curl",
      "curl -h"
    ],
    "multi-commands-darwin": [
      "<auto_detect_package_manager> install curl",
      "curl -h"
    ]
  }
}
```

### Field Reference

| Field | Description |
|-------|-------------|
| `desc` | Tool description (pulled from GitHub repo) |
| `version` | Current version number |
| `repo_link` | GitHub repository URL |
| `package-manager` | Installation method: `github-release`, `go-getter`, `pip`, or `<auto_detect_package_manager>` |
| `tags` | Categories for filtering (e.g., `recon`, `scanning`, `optional`) |
| `valide-command` | Command to verify installation |
| `linux` / `darwin` | OS-specific download URLs by architecture |
| `command-dual` | Cross-platform installation command |
| `multi-commands-*` | Sequence of installation commands |

## Supported Tools

### Core Security Tools

| Tool | Category | Description |
|------|----------|-------------|
| **amass** | Recon | In-depth attack surface mapping and asset discovery |
| **subfinder** | Recon | Fast passive subdomain enumeration |
| **nuclei** | Scanning | Customizable vulnerability scanner with YAML DSL |
| **httpx** | Probing | Multi-purpose HTTP toolkit |
| **naabu** | Scanning | Fast port scanner |
| **ffuf** | Fuzzing | Fast web fuzzer |
| **dnsx** | DNS | Multi-purpose DNS toolkit |
| **katana** | Crawling | Web crawler for endpoint discovery |
| **trufflehog** | Secrets | Secret detection and scanning |

### Optional Tools

Tools tagged with `optional` are available but not installed by default:

- **trivy** — Container/code vulnerability scanner
- **semgrep** — Static analysis for code scanning
- **dalfox** — XSS scanner and utility
- **cloudfox** — Cloud penetration testing automation
- **legba** — Multi-protocol credential bruteforcer

### Utilities

Common utilities that workflows may depend on:

- **curl**, **wget**, **rsync**, **jq**, **rg** (ripgrep), **coreutils**

## Contributing

### Add a New Tool

1. **Add to registry JSON** — Create entry in `registry-metadata-direct-fetch.json`

2. **For GitHub release tools** — Add repo mapping in `update_registry_metadata.py`:
   ```python
   REPO_WITH_ARTIFACTS = {
       "your-tool": "owner/repo",
       # ...
   }
   ```

3. **Run update script** to populate versions and URLs:
   ```bash
   python3 update_registry_metadata.py --tool your-tool
   ```

4. **Test with Osmedeus**:
   ```bash
   osmedeus install binary your-tool
   ```

### Tool Entry Template

```json
{
  "tool-name": {
    "desc": "Brief description",
    "repo_link": "https://github.com/owner/repo",
    "version": "1.0.0",
    "package-manager": "github-release",
    "tags": ["category1", "category2"],
    "valide-command": "",
    "linux": {
      "amd64": "https://github.com/.../download/v1.0.0/tool_linux_amd64.tar.gz",
      "arm64": "https://github.com/.../download/v1.0.0/tool_linux_arm64.tar.gz"
    },
    "darwin": {
      "amd64": "https://github.com/.../download/v1.0.0/tool_darwin_amd64.tar.gz",
      "arm64": "https://github.com/.../download/v1.0.0/tool_darwin_arm64.tar.gz"
    }
  }
}
```

## Automation

The `update_registry_metadata.py` script:

- Fetches latest release versions from GitHub API
- Updates download URLs with new version numbers
- Pulls repository descriptions
- Handles various filename patterns (e.g., `tool_1.0.0_linux_amd64.zip`, `tool-linux-amd64-1.0.0.tgz`)

### GitHub Actions (Recommended)

Set up automated updates with a scheduled workflow:

```yaml
name: Update Registry
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.x'
      - run: pip install requests
      - run: python3 update_registry_metadata.py
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      - uses: stefanzweifel/git-auto-commit-action@v5
        with:
          commit_message: "chore: update registry metadata"
```

## Related

- [Osmedeus](https://github.com/j3ssie/osmedeus) — Security orchestration engine
- [Osmedeus Documentation](https://docs.osmedeus.org) — Full documentation
- [Osmedeus Workflows](https://docs.osmedeus.org/workflows/overview) — Writing custom workflows

## License

MIT License — Part of the [Osmedeus](https://github.com/j3ssie/osmedeus) project.

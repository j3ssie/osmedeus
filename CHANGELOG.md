# Changelog

## v5.1.4

v5.1.3 was tagged but never published; this is the first release to ship its
changes, listed under v5.1.3 below.

### Changed
- All Go dependencies updated to their latest versions, including bun v1.3 and acp-go-sdk v0.20.0. Go 1.26.10 or newer is required to build.
- All dashboard dependencies updated to their latest versions, including Next.js 16.4, @xyflow/react 13, motion 14, ESLint 10 and TypeScript 7.
- ACP agents that ask for user input (elicitation) are declined: the ACP client runs unattended.

## v5.1.3

### Added
- `suspect` severity (a lead the scanner could not confirm) across the dashboard, API filters and `osmedeus vulns`, ranked between low and info.
- Vulnerabilities store `cwe_id`, `cvss_score` and `remediation`; the finding drawer shows CVSS, a linked CWE and a Remediation section.
- Install API responses report `skipped` / `skipped_count` for binaries already on PATH.

### Changed
- Vigolium import brings in technologies, response time, CWE/CVSS/remediation, vigolium's own tags and `vuln_class` (v0.5.7 exports).
- Workflows use vigolium's canonical `--max-time`, `--discovery-wordlist` and `--only discovery`.

### Fixed
- `agent-acp` steps can run custom agents from `acp_config.command` (#331).
- Assets page pagination and filter changes now fetch immediately (#332).
- Multi-select severity and confidence filters on the vulnerabilities API, and the page now applies the confidence filter its total did.
- First-time setup no longer reports `coreutils` as failed when it is installed, and counts present binaries as skipped rather than installed (#330).
- Installing a package that needs root, without root or sudo, explains why instead of failing with exit 127.
- The GitHub token is only added to clone URLs on GitHub hosts (#333).
- Workspace URL and asset totals count long JSONL lines (#329).

## v5.1.2

### Added
- `osmedeus vulns` command to query findings with severity, confidence, template and workspace filters, plus `--stats` and `--id` detail view.
- `osmedeus health` now checks that every workflow-declared dependency command is installed or available in the registry.
- Linter flags reports that are missing the required `path` field.
- `jq` added to the binary registry.

### Changed
- `osmedeus db` skips heavy vulnerability columns (description, PoC, HTTP request/response) by default.
- `assetfinder` registry entry now uses a single cross-platform `go install`.
- `subfinder` bumped from 2.14.0 to 2.16.0.
- Docker images install native Chromium from the xtradeb PPA instead of the snap-only stub.

### Fixed
- ACP agent no longer hangs when stderr output exceeds the 64KiB scanner limit.
- ACP agent errors now return the full stderr instead of racing its reader.
- Reading targets from stdin or an API target file no longer fails on long lines.
- Line counting now handles long lines in buffered files.

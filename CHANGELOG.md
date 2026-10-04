# Changelog

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

# Event Workflow Examples

This folder contains sample workflows demonstrating the event trigger and generation system.

## Quick Start

```bash
# Start the server with event receiver
osmedeus server -F events/vuln-scan-receiver.yaml

# Generate a sample event
osmedeus eval 'generate_event("hackerone.com", "test.asset.vars", "cli", "subdomain", "docs.hackerone.com")'
```

## Workflows

### Emitters (Event Generators)

| File | Description |
|------|-------------|
| `simple-emitter.yaml` | Basic event emission with `generate_event` and `generate_event_from_file` |

### Receivers (Event Triggers)

| File | Description |
|------|-------------|
| `simple-receiver.yaml` | Basic event trigger with filters and deduplication |
| `vuln-scan-receiver.yaml` | HTTP fingerprinting and Nuclei vulnerability scan on received events |

## Event Functions

### generate_event(workspace, topic, source, data_type, data)

Emit a single event with optional structured data.

```yaml
- type: function
  function: |
    generate_event("{{TargetSpace}}", "discovery.asset", "my-scanner", "subdomain", "api.example.com")
```

### generate_event_from_file(workspace, topic, source, data_type, file_path)

Emit one event per line from a file. Returns the count of events emitted.

```yaml
- type: function
  functions:
    - 'generate_event_from_file("{{TargetSpace}}", "discovery.asset", "my-scanner", "subdomain", "{{Output}}/subdomains.txt")'
```

## Trigger Configuration

### Basic Event Trigger

```yaml
triggers:
  - name: on-new-asset
    on: event
    event:
      topic: "discovery.asset"
    input:
      Target: event_data.value
    enabled: true
```

### Filtered Event Trigger with Deduplication

```yaml
triggers:
  - name: on-new-asset
    on: event
    event:
      topic: "discovery.asset"
      filters:
        - "event.source == 'simple-emitter'"
        - "event.data_type == 'subdomain'"
      dedupe_key: "{{event.source}}-{{event.data}}"
      dedupe_window: 10s
    input:
      target: event_data.value
    enabled: true
```

### Advanced Input Extraction

```yaml
triggers:
  - name: on-asset-vars
    on: event
    event:
      topic: "*"
    input:
      Target: pick_valid(event_data.sample, event_data.value)
      RawData: event.data
      asset_type: event.type
      source: event.source
      description: trim_left(event.topic, 'test.')
    enabled: true
```

### Disabling Manual CLI Execution

```yaml
triggers:
  - name: manual-trigger
    on: manual
    enabled: false
```

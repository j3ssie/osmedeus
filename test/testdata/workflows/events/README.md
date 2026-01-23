# Event Workflow Examples

This folder contains sample workflows demonstrating the event trigger and generation system.

## Workflows

### Emitters (Event Generators)

| File | Description |
|------|-------------|
| `simple-emitter.yaml` | Basic event emission with `generate_event` and `generate_event_from_file` |
| `vuln-emitter.yaml` | Simulates a vulnerability scanner emitting structured finding events |

### Receivers (Event Triggers)

| File | Description |
|------|-------------|
| `simple-receiver.yaml` | Basic event trigger that listens for discovery events |
| `filtered-receiver.yaml` | Advanced filtering with severity checks and jq extraction |
| `dedupe-receiver.yaml` | Event deduplication to prevent duplicate processing |

## Event Functions

### generate_event(workspace, topic, source, data_type, data)
Emit a single event with optional structured data. The workspace parameter identifies the target space for the event.

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
trigger:
  - name: on-new-asset
    on: event
    event:
      topic: "discovery.asset"
    input:
      type: event_data
      field: "value"
      name: target
    enabled: true
```

### Filtered Event Trigger
```yaml
trigger:
  - name: on-high-severity
    on: event
    event:
      topic: "scan.finding"
      filters:
        - "event.data.severity == 'high'"
        - "event.data.confirmed == true"
    input:
      type: function
      function: 'jq("{{event.data}}", ".url")'
      name: target
    enabled: true
```

### Deduplicated Event Trigger
```yaml
trigger:
  - name: on-url-dedupe
    on: event
    event:
      topic: "crawler.url"
      dedupe_key: "{{event.data.url}}"
      dedupe_window: "5m"
    input:
      type: event_data
      field: "value"
      name: target
    enabled: true
```

## Running Examples

```bash
# Run the emitter to generate events
osmedeus run -m simple-emitter -t example.com

# The receiver will automatically process events if registered with the scheduler
osmedeus run -m simple-receiver -t example.com
```

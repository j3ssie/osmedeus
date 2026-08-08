# File Uploads

## Upload Input File

Upload a file containing a list of inputs (targets, URLs, etc.) for later use in runs.

```bash
curl -X POST http://localhost:8002/osm/api/upload-file \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@targets.txt"
```

**Response:**
```json
{
  "message": "File uploaded",
  "filename": "1704326400000000000_targets.txt",
  "path": "/home/user/osmedeus-base/data/uploads/1704326400000000000_targets.txt",
  "size": 1024,
  "lines": 50
}
```

The returned `path` can be used as a target in subsequent run requests.

---

## Upload Workflow

Upload a raw YAML workflow file and save it to the workflows directory.

```bash
curl -X POST http://localhost:8002/osm/api/workflow-upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@my-custom-workflow.yaml"
```

**Response:**
```json
{
  "message": "Workflow uploaded",
  "name": "my-custom-workflow",
  "kind": "module",
  "description": "A custom security workflow",
  "path": "/home/user/osmedeus-base/workflows/modules/my-custom-workflow.yaml"
}
```

The workflow file must be a valid YAML with `.yaml` or `.yml` extension. It will be saved to either the `flows/` or `modules/` subdirectory based on the workflow kind.

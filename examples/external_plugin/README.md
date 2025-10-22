# External Plugin Example - Notes Plugin

This example demonstrates how to create an external DarwinFlow plugin that runs as a separate process and communicates via JSON-RPC.

## Overview

The **notes-plugin** is an external plugin written in Go that provides note entities. It demonstrates:

- ✅ JSON-RPC protocol implementation (stdin/stdout communication)
- ✅ IEntityProvider capability (query and get entities)
- ✅ IEntityUpdater capability (update entities)
- ✅ IEventEmitter capability (emit events on changes)
- ✅ Proper error handling and protocol compliance
- ✅ Running as a separate subprocess

## Features

### Entity Management
- **Entity Type**: `note`
- **Operations**: Query, Get, Update
- **Fields**: ID, Title, Content, CreatedAt, UpdatedAt

### Event Streaming
- Emits `stream.started` when event streaming begins
- Emits `note.updated` when a note is modified

## Building

From the DarwinFlow root directory:

```bash
# Build the plugin executable
go build -o bin/notes-plugin ./examples/external_plugin/cmd/notes-plugin

# Or build from this directory
cd examples/external_plugin
go build -o ../../bin/notes-plugin ./cmd/notes-plugin
```

## Usage

### Testing with DarwinFlow

```go
// In your DarwinFlow application code:
import "github.com/kgatilin/darwinflow-pub/internal/infra"

// Create subprocess plugin
plugin := infra.NewSubprocessPlugin("./bin/notes-plugin")

// Initialize
err := plugin.Initialize(ctx, "/path/to/working/dir", nil)
if err != nil {
    log.Fatal(err)
}
defer plugin.Shutdown()

// Query notes
query := pluginsdk.EntityQuery{
    EntityType: "note",
    Limit:      10,
}
notes, err := plugin.Query(ctx, query)

// Get specific note
note, err := plugin.GetEntity(ctx, "note-1")

// Update note
updated, err := plugin.UpdateEntity(ctx, "note-1", map[string]interface{}{
    "title": "Updated Title",
})
```

### Manual Testing

You can test the plugin manually by sending JSON-RPC requests via stdin:

```bash
# Start the plugin
./bin/notes-plugin

# Send a request (in JSON-RPC 2.0 format)
{"jsonrpc":"2.0","id":1,"method":"get_info"}

# Response:
{"jsonrpc":"2.0","id":1,"result":{"name":"notes-external","version":"1.0.0","description":"External notes plugin (Go subprocess example)","is_core":false}}
```

## Protocol Implementation

The plugin implements the DarwinFlow JSON-RPC protocol as defined in `pkg/pluginsdk/rpc.go`:

### Supported Methods

| Method | Description | Parameters | Result |
|--------|-------------|------------|--------|
| `init` | Initialize plugin | `InitParams` | `null` |
| `get_info` | Get plugin metadata | none | `PluginInfo` |
| `get_capabilities` | Get capabilities | none | `[]string` |
| `get_entity_types` | Get entity types | none | `[]EntityTypeInfo` |
| `query_entities` | Query notes | `EntityQuery` | `[]map[string]interface{}` |
| `get_entity` | Get note by ID | `GetEntityParams` | `map[string]interface{}` |
| `update_entity` | Update note | `UpdateEntityParams` | `map[string]interface{}` |
| `start_event_stream` | Start events | none | `null` |
| `stop_event_stream` | Stop events | none | `null` |

### Protocol Format

**Request** (stdin):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "query_entities",
  "params": {
    "entity_type": "note",
    "limit": 10
  }
}
```

**Response** (stdout):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    {
      "id": "note-1",
      "type": "note",
      "title": "Example Note",
      "content": "This is an example note.",
      "created_at": "2025-10-21T10:00:00Z",
      "updated_at": "2025-10-21T10:00:00Z",
      "capabilities": []
    }
  ]
}
```

**Event** (stdout):
```json
{
  "event": "event",
  "type": "note.updated",
  "source": "notes-external",
  "timestamp": "2025-10-22T15:30:00Z",
  "payload": {
    "note_id": "note-1",
    "title": "Updated Title"
  }
}
```

## Architecture

```
┌─────────────────────────────────────────┐
│          DarwinFlow Main Process        │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │   SubprocessPlugin (Adapter)     │  │
│  │  - Implements SDK interfaces     │  │
│  │  - Delegates to RPC client       │  │
│  └──────────────┬────────────────────┘  │
│                 │                       │
│  ┌──────────────▼────────────────────┐  │
│  │      RPCClient                    │  │
│  │  - Manages subprocess             │  │
│  │  - stdin/stdout pipes             │  │
│  │  - Request/response correlation   │  │
│  └──────────────┬────────────────────┘  │
└─────────────────┼────────────────────────┘
                  │ JSON-RPC over pipes
┌─────────────────▼────────────────────────┐
│      notes-plugin (Subprocess)          │
│                                         │
│  - Reads requests from stdin            │
│  - Sends responses to stdout            │
│  - Manages note entities in memory      │
│  - Emits events when notes change       │
└─────────────────────────────────────────┘
```

## Writing Your Own External Plugin

To create your own external plugin in any language:

1. **Implement JSON-RPC 2.0 server** that reads from stdin and writes to stdout
2. **Implement required methods**: `init`, `get_info`, `get_capabilities`
3. **Implement capability methods** based on your plugin's features
4. **Follow newline-delimited JSON format**: One JSON object per line
5. **Handle errors properly**: Use standard JSON-RPC error codes
6. **Test thoroughly**: Ensure proper request/response correlation

### Language Support

While this example is in Go, you can implement external plugins in **any language**:

- **Python**: Use `json` module and `sys.stdin`/`sys.stdout`
- **TypeScript/Node.js**: Use `readline` and `console.log`
- **Rust**: Use `serde_json` and `std::io`
- **Java**: Use `jackson` and `System.in`/`System.out`

See the Phase 7 deliverables for official Python and TypeScript SDKs.

## Testing

The plugin includes comprehensive tests in the main DarwinFlow test suite:

```bash
# Run subprocess plugin tests
go test ./internal/infra -run SubprocessPlugin -v
```

## License

Same as DarwinFlow main project.

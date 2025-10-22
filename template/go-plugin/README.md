# DarwinFlow Plugin Template (Go)

This template provides a complete starting point for building external DarwinFlow plugins in Go. External plugins run as separate processes and communicate with DarwinFlow via JSON-RPC over stdin/stdout.

## Quick Start

1. **Copy this template**:
   ```bash
   cp -r template/go-plugin /path/to/your-plugin
   cd /path/to/your-plugin
   ```

2. **Customize the module name**:
   ```bash
   # Edit go.mod and replace:
   # module example.com/myplugin
   # with:
   # module github.com/yourname/yourplugin
   ```

3. **Update the replace directive** in `go.mod`:
   ```go
   // Point to your local DarwinFlow installation
   replace github.com/kgatilin/darwinflow-pub => /path/to/darwinflow
   ```

4. **Build the plugin**:
   ```bash
   make build
   ```

5. **Test it**:
   ```bash
   ./bin/myplugin
   # In another terminal, send a test request:
   echo '{"jsonrpc":"2.0","id":1,"method":"get_info"}' | ./bin/myplugin
   ```

## Prerequisites

- **Go 1.23+**: Required for building the plugin
- **DarwinFlow**: The main DarwinFlow application must be installed
- **Make**: Optional, but recommended for building

## Building

The template includes a Makefile with common tasks:

```bash
# Build the plugin binary
make build

# Clean build artifacts
make clean

# Build and run
make run

# Install to /usr/local/bin (requires sudo)
sudo make install
```

To build manually:

```bash
mkdir -p bin
go build -o bin/myplugin ./cmd/myplugin
```

## Integration with DarwinFlow

### Using plugins.yaml

Create or edit `~/.darwinflow/plugins.yaml`:

```yaml
plugins:
  myplugin:
    # Absolute path to the plugin binary
    command: /path/to/your-plugin/bin/myplugin
    enabled: true

    # Optional: Plugin-specific configuration
    config:
      custom_setting: "value"
```

### Using CLI

```bash
# List all plugins (including external ones)
dw plugins list

# Query items from your plugin
dw entities query --type item

# Get a specific item
dw entities get item-1

# Update an item
dw entities update item-1 --field name="Updated Name"
```

### Programmatic Usage

```go
import (
    "github.com/kgatilin/darwinflow-pub/internal/infra"
    "github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// Create subprocess plugin
plugin := infra.NewSubprocessPlugin("/path/to/bin/myplugin")

// Initialize
ctx := context.Background()
err := plugin.Initialize(ctx, "/working/dir", nil)
if err != nil {
    log.Fatal(err)
}
defer plugin.Shutdown()

// Query entities
query := pluginsdk.EntityQuery{
    EntityType: "item",
    Limit:      10,
}
items, err := plugin.Query(ctx, query)

// Get specific entity
item, err := plugin.GetEntity(ctx, "item-1")

// Update entity
updated, err := plugin.UpdateEntity(ctx, "item-1", map[string]interface{}{
    "name": "New Name",
    "tags": []string{"updated"},
})
```

## Customization

### Changing the Entity Type

The template uses a generic "item" entity. To customize:

1. **Rename the entity type** in `internal/entity.go`:
   ```go
   type Task struct {  // Changed from Item
       ID          string
       Title       string  // Changed from Name
       Status      string  // Add custom fields
       // ...
   }
   ```

2. **Update the ToMap() method** to reflect new fields:
   ```go
   func (t *Task) ToMap() map[string]interface{} {
       return map[string]interface{}{
           "id":     t.ID,
           "type":   "task",  // Change type name
           "title":  t.Title,
           "status": t.Status,
           // ...
       }
   }
   ```

3. **Update handlers** in `internal/handlers.go`:
   ```go
   // In handleGetEntityTypes:
   Type:              "task",  // Change from "item"
   DisplayName:       "Task",
   DisplayNamePlural: "Tasks",

   // In handleQueryEntities:
   if query.EntityType != "task" { // Change from "item"
   ```

4. **Update the main.go** sample data to use your new entity type.

### Adding Custom Fields

To add fields to your entity:

1. Add fields to the struct in `internal/entity.go`
2. Update the `ToMap()` serialization method
3. Handle updates in `handleUpdateEntity` in `internal/handlers.go`

Example:
```go
// entity.go
type Item struct {
    ID          string
    Name        string
    Priority    int      // NEW FIELD
    Assignee    string   // NEW FIELD
    // ...
}

// ToMap()
func (i *Item) ToMap() map[string]interface{} {
    return map[string]interface{}{
        // ...existing fields...
        "priority": i.Priority,
        "assignee": i.Assignee,
    }
}

// handlers.go - handleUpdateEntity
if priority, ok := params.Fields["priority"].(float64); ok {
    item.Priority = int(priority)
}
if assignee, ok := params.Fields["assignee"].(string); ok {
    item.Assignee = assignee
}
```

### Adding Persistence

The template stores entities in memory. To add persistence:

1. **Add database initialization** in `internal/plugin.go`:
   ```go
   import "database/sql"

   type ItemPlugin struct {
       db *sql.DB
       // ...
   }
   ```

2. **Initialize in NewItemPlugin()**:
   ```go
   func NewItemPlugin() *ItemPlugin {
       db, err := sql.Open("sqlite3", "items.db")
       if err != nil {
           log.Fatal(err)
       }
       // Create tables...
       return &ItemPlugin{db: db}
   }
   ```

3. **Update handlers** to query/update the database instead of the in-memory map.

### Adding New Capabilities

The template implements three capabilities. To add more:

1. **Update GetCapabilities** in `internal/handlers.go`:
   ```go
   capabilities := []string{
       "IEntityProvider",
       "IEntityUpdater",
       "IEventEmitter",
       "ICommandProvider",  // NEW
   }
   ```

2. **Implement the required RPC methods**. See `pkg/pluginsdk/rpc.go` for method signatures.

3. **Add handlers** in `internal/handlers.go` and register in `handleRequest()`.

## Project Structure

```
template/go-plugin/
├── cmd/
│   └── myplugin/
│       └── main.go          # Entry point - initializes plugin and starts server
├── internal/
│   ├── entity.go            # Item entity definition and serialization
│   ├── plugin.go            # ItemPlugin struct and RPC server loop
│   └── handlers.go          # RPC method handlers
├── .gitignore
├── go.mod                   # Module definition and dependencies
├── Makefile                 # Build automation
└── README.md                # This file
```

### File Responsibilities

**cmd/myplugin/main.go**:
- Creates plugin instance
- Initializes sample data
- Starts JSON-RPC server

**internal/entity.go**:
- Defines entity structure (Item)
- Implements serialization (ToMap)

**internal/plugin.go**:
- ItemPlugin struct and state
- RPC server loop (Serve)
- Request routing (handleRequest)
- Response helpers (sendResult, sendError)
- Event emission (emitEvent)

**internal/handlers.go**:
- Individual RPC method handlers
- Business logic for each operation
- Parameter validation

## Protocol Reference

### Supported RPC Methods

| Method | Description | Parameters | Response |
|--------|-------------|------------|----------|
| `init` | Initialize plugin | `InitParams` | `null` |
| `get_info` | Plugin metadata | none | `PluginInfo` |
| `get_capabilities` | List capabilities | none | `[]string` |
| `get_entity_types` | Entity type metadata | none | `[]EntityTypeInfo` |
| `query_entities` | Query entities | `EntityQuery` | `[]map[string]interface{}` |
| `get_entity` | Get by ID | `GetEntityParams` | `map[string]interface{}` |
| `update_entity` | Update entity | `UpdateEntityParams` | `map[string]interface{}` |
| `start_event_stream` | Start events | none | `null` |
| `stop_event_stream` | Stop events | none | `null` |

### Request Format

All requests follow JSON-RPC 2.0 format:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "query_entities",
  "params": {
    "entity_type": "item",
    "limit": 10
  }
}
```

### Response Format

Success:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [...]
}
```

Error:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params"
  }
}
```

### Event Format

Events are emitted to stdout when streaming is active:

```json
{
  "event": "event",
  "type": "item.updated",
  "source": "myplugin",
  "timestamp": "2025-10-22T15:30:00Z",
  "payload": {
    "item_id": "item-1",
    "name": "Updated Name"
  }
}
```

### Standard Error Codes

- `-32700`: Parse error (invalid JSON)
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error
- `-32000 to -32099`: Server-defined errors

For complete protocol details, see `pkg/pluginsdk/rpc.go` in the DarwinFlow repository.

## Testing

### Manual Testing

Start the plugin and send JSON requests:

```bash
# Terminal 1: Start plugin
./bin/myplugin

# Terminal 2: Send requests
echo '{"jsonrpc":"2.0","id":1,"method":"get_info"}' | nc localhost <stdin>
```

### Integration Testing

The DarwinFlow test suite includes subprocess plugin tests:

```bash
cd /path/to/darwinflow
go test ./internal/infra -run SubprocessPlugin -v
```

### Unit Testing

Add tests for your custom logic in `internal/*_test.go`:

```go
package internal_test

import "testing"

func TestItemToMap(t *testing.T) {
    item := &Item{
        ID:   "test-1",
        Name: "Test Item",
    }

    m := item.ToMap()
    if m["id"] != "test-1" {
        t.Errorf("Expected id=test-1, got %v", m["id"])
    }
}
```

## Debugging

### Enable Verbose Logging

Add logging to stderr (stdout is reserved for RPC):

```go
import "log"

func (p *ItemPlugin) handleRequest(req *pluginsdk.RPCRequest) {
    log.Printf("[DEBUG] Received method: %s\n", req.Method)
    // ...
}
```

Run with stderr redirected:

```bash
./bin/myplugin 2> debug.log
```

### Common Issues

**Plugin doesn't start**:
- Check if binary is executable: `chmod +x bin/myplugin`
- Verify go.mod replace directive points to correct path
- Run `go mod tidy` to ensure dependencies are correct

**RPC errors**:
- Ensure JSON is valid and newline-delimited
- Check that method names match `pluginsdk.RPCMethod*` constants
- Verify parameter structures match `pluginsdk.*Params` types

**Events not appearing**:
- Ensure `start_event_stream` was called
- Check that `p.eventStreaming` is true before emitting
- Verify events are written to stdout, not stderr

## Advanced Topics

### Multi-Language Plugins

While this template is in Go, you can implement plugins in any language:

- **Python**: Use `json`, `sys.stdin`, `sys.stdout`
- **TypeScript/Node.js**: Use `readline`, `console.log`
- **Rust**: Use `serde_json`, `std::io`
- **Java**: Use Jackson, `System.in`/`System.out`

The protocol is language-agnostic JSON-RPC.

### Performance Optimization

For high-volume plugins:

1. **Buffer management**: Increase scanner buffer sizes
2. **Connection pooling**: Reuse database connections
3. **Caching**: Cache frequently accessed entities
4. **Batch operations**: Process multiple entities at once

### Security Considerations

- **Input validation**: Always validate parameters before processing
- **Resource limits**: Prevent unbounded memory/CPU usage
- **File access**: Respect the working directory constraint
- **Secrets**: Never log sensitive data (use stderr, not stdout for logs)

## Examples

### Example 1: Simple Query

Request:
```json
{"jsonrpc":"2.0","id":1,"method":"query_entities","params":{"entity_type":"item","limit":2}}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    {
      "id": "item-1",
      "type": "item",
      "name": "Example Item",
      "description": "This is an example item from the external plugin.",
      "tags": ["example", "demo"],
      "created_at": "2025-10-21T15:30:00Z",
      "updated_at": "2025-10-21T15:30:00Z",
      "capabilities": []
    },
    {
      "id": "item-2",
      "type": "item",
      "name": "Another Item",
      "description": "External plugins can run in any language!",
      "tags": ["external", "plugin"],
      "created_at": "2025-10-22T13:30:00Z",
      "updated_at": "2025-10-22T14:30:00Z",
      "capabilities": []
    }
  ]
}
```

### Example 2: Update with Event

Request:
```json
{"jsonrpc":"2.0","id":2,"method":"update_entity","params":{"entity_id":"item-1","fields":{"name":"Updated Item","tags":["updated"]}}}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "id": "item-1",
    "type": "item",
    "name": "Updated Item",
    "description": "This is an example item from the external plugin.",
    "tags": ["updated"],
    "created_at": "2025-10-21T15:30:00Z",
    "updated_at": "2025-10-22T15:45:00Z",
    "capabilities": []
  }
}
```

Event (if streaming enabled):
```json
{
  "event": "event",
  "type": "item.updated",
  "source": "myplugin",
  "timestamp": "2025-10-22T15:45:00Z",
  "payload": {
    "item_id": "item-1",
    "name": "Updated Item"
  }
}
```

## Resources

- **DarwinFlow Documentation**: `/path/to/darwinflow/README.md`
- **Plugin SDK Reference**: `pkg/pluginsdk/` (interface definitions)
- **RPC Protocol**: `pkg/pluginsdk/rpc.go` (protocol types and constants)
- **Example Plugin**: `examples/external_plugin/` (notes-plugin reference)
- **Plugin Development Guide**: `docs/plugin-development-guide.md`

## License

Same as DarwinFlow main project.

## Support

For questions or issues:
1. Check the DarwinFlow documentation
2. Review the notes-plugin example in `examples/external_plugin/`
3. Consult `pkg/pluginsdk/rpc.go` for protocol details
4. Open an issue in the DarwinFlow repository

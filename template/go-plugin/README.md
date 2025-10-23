# DarwinFlow Plugin Template (Go)

External plugin template for building DarwinFlow plugins in Go. Plugins run as separate processes and communicate via JSON-RPC over stdin/stdout.

---

## Quick Start

```bash
# 1. Copy template
cp -r template/go-plugin /path/to/your-plugin && cd /path/to/your-plugin

# 2. Edit go.mod - update module name and replace directive
# module github.com/yourname/yourplugin
# replace github.com/kgatilin/darwinflow-pub => /path/to/darwinflow

# 3. Build
make build

# 4. Test locally
./bin/myplugin-cli list

# 5. Register in ~/.darwinflow/plugins.yaml
# See "Integration" section below
```

**Prerequisites**: Go 1.23+, DarwinFlow installed

---

## What the Framework Provides

**Key Principle**: Framework handles infrastructure and cross-plugin concerns. Plugins handle domain logic.

### 1. Event Storage (`EventRepository`)

**Interface**: `pluginsdk.EventRepository`

- Persistent SQLite storage for all plugin events
- Query by type, time, session, metadata, full-text search
- Raw SQL for advanced analytics
- Versioning and migration support

**Plugin receives**: `PluginContext.EventRepository` in `Init()`

**Plugin responsibility**: Define event types/payloads, emit events

**Framework responsibility**: Storage, indexing, queries, schema

### 2. Centralized Analysis

**Why Centralized**: Framework sees ALL events from ALL plugins → enables cross-plugin pattern detection

- AI-powered session analysis with configurable prompts
- Token budget management, parallel execution
- LLM integration (Claude CLI)
- Analysis persistence and retrieval
- Multiple prompt support

**Plugin responsibility**: Emit rich, meaningful events with context

**Framework responsibility**: Collect events, run LLM, store results, provide UI/CLI

**User configures**: `~/.darwinflow/config.yaml` (token limits, models, prompts)

### 3. Cross-Plugin Communication (`EventBus`)

**Interface**: `pluginsdk.EventBus`

- Publish/subscribe real-time communication
- Filter by type patterns (glob), labels, source
- Async, thread-safe, non-blocking delivery
- Optional persistence and replay

**Plugin receives**: EventBus in `Init(ctx, pluginCtx, eventBus)`

**Example**: Gmail publishes `gmail.email_received` → Calendar subscribes, creates event → Telegram subscribes, sends notification

**Note**: EventBus ≠ EventRepository. EventBus = real-time. EventRepository = durable storage.

### 4. Plugin Lifecycle

**Interfaces**: `pluginsdk.PluginContext`, `pluginsdk.CommandContext`

**Stages**: Registration → Init → Command execution → Query/Update → Event emission → Shutdown

**PluginContext provides**: Logger, EventRepository, CWD

**CommandContext provides**: Logger, CWD, Output, Input, ProjectData

### 5. Configuration, Logging, Commands, Entities

| Feature | Interface | Framework Provides | Plugin Implements |
|---------|-----------|-------------------|-------------------|
| **Config** | Framework-managed | YAML load/save, validation | Document config keys |
| **Logging** | `pluginsdk.Logger` | Leveled logging, formatting | Use Logger for all output |
| **Commands** | `pluginsdk.ICommandProvider` | Discovery, routing, help | Return commands from `GetCommands()` |
| **Entities** | `pluginsdk.IEntityProvider` | Aggregation, routing | Define entities, implement queries |

### 6. Database Infrastructure

- Centralized SQLite for all plugins
- Schema management, migrations, indexing, full-text search
- Access via `EventRepository` interface

**Don't**: Create plugin-specific tables. **Do**: Store as events/entities.

### 7. External Plugin Support (RPC)

- JSON-RPC 2.0 over stdin/stdout
- Language-agnostic (Python, Node.js, Rust, Java, etc.)
- Standard RPC methods: `init`, `get_info`, `get_capabilities`, `query_entities`, etc.

**See**: `pkg/pluginsdk/rpc.go` for protocol details

---

## Framework vs Plugin Responsibilities

| Framework Handles | Plugin Handles |
|------------------|----------------|
| Event storage & queries | Event types & payloads |
| Centralized analysis | Domain logic |
| Cross-plugin communication | Event handlers |
| Configuration management | Business rules |
| Logging infrastructure | Entity definitions |
| Command routing | Custom commands |
| Entity aggregation | External API integration |
| Database management | Validation |
| RPC protocol | - |

**Decision Guide**:
- Cross-plugin visibility? → Framework
- Infrastructure? → Framework
- Shared across plugins? → Framework
- Domain-specific? → Plugin

---

## Template Usage

### Building

```bash
make build      # Build both plugin and CLI
make build-cli  # CLI wrapper only
make test-cli   # Test all commands
make clean      # Clean artifacts
```

### Local Testing (CLI Mode)

```bash
./bin/myplugin-cli list                      # List items
./bin/myplugin-cli get item-1                # Get by ID
./bin/myplugin-cli update item-1 name "New"  # Update field
./bin/myplugin-cli info                      # Plugin info
./bin/myplugin-cli types                     # Entity types
```

**CLI vs RPC**:
- **CLI**: Direct calls, stderr events, local testing
- **RPC**: JSON-RPC stdin/stdout, production use

### Integration with DarwinFlow

**Register in `~/.darwinflow/plugins.yaml`**:
```yaml
plugins:
  myplugin:
    command: /path/to/your-plugin/bin/myplugin
    enabled: true
    config:
      custom_setting: "value"
```

**Use via CLI**:
```bash
dw plugins list
dw entities query --type item
dw entities get item-1
dw entities update item-1 --field name="Updated"
```

---

## Customization

### Change Entity Type

1. Edit `internal/entity.go` - rename struct, update fields
2. Update `ToMap()` method
3. Update handlers in `internal/handlers.go` (type name, queries)
4. Update sample data in `cmd/myplugin/main.go`

### Add Persistence

Default: in-memory. To persist:
1. Add `db *sql.DB` to plugin struct
2. Initialize in constructor
3. Update handlers to query/update database

### Add Capabilities

1. Update `GetCapabilities()` to include new capability
2. Implement required RPC methods (see `pkg/pluginsdk/rpc.go`)
3. Add handlers and register in `handleRequest()`

---

## RPC Protocol Reference

### Methods

| Method | Params | Response |
|--------|--------|----------|
| `init` | `InitParams` | `null` |
| `get_info` | none | `PluginInfo` |
| `get_capabilities` | none | `[]string` |
| `get_entity_types` | none | `[]EntityTypeInfo` |
| `query_entities` | `EntityQuery` | `[]map[string]interface{}` |
| `get_entity` | `GetEntityParams` | `map[string]interface{}` |
| `update_entity` | `UpdateEntityParams` | `map[string]interface{}` |
| `start_event_stream` | none | `null` |
| `stop_event_stream` | none | `null` |

### Request Format (JSON-RPC 2.0)

```json
{"jsonrpc":"2.0","id":1,"method":"query_entities","params":{"entity_type":"item","limit":10}}
```

### Response Format

**Success**: `{"jsonrpc":"2.0","id":1,"result":[...]}`

**Error**: `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params"}}`

### Event Format

```json
{"event":"event","type":"item.updated","source":"myplugin","timestamp":"2025-10-22T15:30:00Z","payload":{"item_id":"item-1"}}
```

### Error Codes

- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

---

## Project Structure

```
template/go-plugin/
├── cmd/
│   ├── myplugin/main.go      # RPC plugin entry point
│   └── myplugin-cli/main.go  # CLI testing tool
├── internal/
│   ├── entity.go             # Entity definition
│   ├── plugin.go             # RPC server, event emission
│   └── handlers.go           # RPC method handlers
├── go.mod, Makefile, .gitignore
└── README.md
```

---

## SDK Interface Reference

**Core**: `pluginsdk.Plugin` (main interface)

**Providers** (plugins implement):
- `IEntityProvider` - Queryable entities
- `IEntityUpdater` - Update entities
- `ICommandProvider` - CLI commands
- `IEventEmitter` - Emit events

**Services** (plugins receive):
- `EventRepository` - Event storage/queries
- `EventBus` - Cross-plugin communication
- `Logger` - Leveled logging
- `PluginContext` - Init context
- `CommandContext` - Command context

**Entity Capabilities**:
- `IExtensible` - Required for all entities
- `ITrackable`, `IHasContext`, `ISchedulable`, `IRelatable` - Optional

**See**: `pkg/pluginsdk/CLAUDE.md` for complete API docs

---

## Testing & Debugging

### Manual Testing

```bash
./bin/myplugin  # Start RPC server
echo '{"jsonrpc":"2.0","id":1,"method":"get_info"}' | ./bin/myplugin  # Send request
```

### Integration Testing

```bash
cd /path/to/darwinflow
go test ./internal/infra -run SubprocessPlugin -v
```

### Debugging

Add logging to **stderr** (stdout reserved for RPC):
```go
log.SetOutput(os.Stderr)
log.Printf("[DEBUG] Received method: %s\n", req.Method)
```

Run: `./bin/myplugin 2> debug.log`

### Common Issues

- **Plugin doesn't start**: Check `chmod +x`, verify `go.mod` replace directive, run `go mod tidy`
- **RPC errors**: Validate JSON, check method names match constants, verify param types
- **Events not appearing**: Ensure `start_event_stream` called, check `eventStreaming` flag, verify stdout output

---

## Advanced Topics

### Multi-Language Plugins

Protocol is language-agnostic. Implement in Python, Node.js, Rust, Java, etc. using JSON-RPC 2.0 over stdin/stdout.

### Performance

- Buffer management: Increase scanner buffer sizes
- Connection pooling: Reuse database connections
- Caching: Cache frequently accessed entities
- Batch operations: Process multiple entities at once

### Security

- Validate all inputs
- Prevent unbounded memory/CPU
- Respect working directory
- Never log secrets (use stderr, not stdout)

---

## Resources

- **Plugin SDK**: `pkg/pluginsdk/` - Interface definitions
- **RPC Protocol**: `pkg/pluginsdk/rpc.go` - Method signatures
- **SDK Docs**: `pkg/pluginsdk/CLAUDE.md` - Complete API reference
- **Example Plugin**: `pkg/plugins/claude_code/` - Reference implementation
- **DarwinFlow Docs**: Main repository README

---

## Support

1. Check DarwinFlow documentation
2. Review reference implementations
3. Consult `pkg/pluginsdk/rpc.go` for protocol
4. Open issue in DarwinFlow repository

**License**: Same as DarwinFlow

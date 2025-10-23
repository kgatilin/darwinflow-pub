# Package: pluginsdk

**Path**: `pkg/pluginsdk`

**Role**: Public plugin SDK - Single source of truth for all plugin contracts

---

## Quick Reference

- **Files**: 8
- **Exports**: 31
- **Dependencies**: None (zero internal imports)
- **Layer**: Foundation (bottom of dependency graph)

---

## Generated Documentation

### Exported API

#### Interfaces

**Core Plugin Interface**:
- `Plugin` - Main plugin interface (GetInfo, GetCapabilities)

**Provider Interfaces**:
- `IEntityProvider` - Provides queryable entities
- `IEntityUpdater` - Updates entities
- `ICommandProvider` - Provides CLI commands
- `IEventEmitter` - Emits events for event sourcing
- `EventBus` - Cross-plugin communication (publish/subscribe)

**Entity Capabilities** (optional interfaces):
- `IExtensible` - **REQUIRED** (GetID, GetType, GetCapabilities, GetField, GetAllFields)
- `ITrackable` - Status tracking (GetStatus, GetProgress, IsBlocked, GetBlockReason)
- `IHasContext` - Contextual relationships (GetContext)
- `ISchedulable` - Time-based scheduling
- `IRelatable` - Entity relationships

#### Key Types

**Context Types**:
- `PluginContext` - Runtime context for plugins (Logger, CWD, EventRepository)
- `CommandContext` - Command execution context (Logger, CWD, ProjectData, Output, Input)
- `EntityContext` - Entity metadata (RelatedEntities, LinkedFiles, RecentActivity, Metadata)

**Query Types**:
- `EntityQuery` - Query entities (Type, Filters, Limit, Offset, SortBy)
- `EventQuery` - Query events (Time range, Types, Metadata, SearchText)
- `QueryResult` - Raw query results (Columns, Rows)

**Core Types**:
- `Event` - Event sourcing primitive (Type, Source, Timestamp, Payload, Metadata, Version)
- `Command` - CLI command definition
- `PluginInfo` - Plugin metadata
- `EntityTypeInfo` - Entity type metadata
- `ActivityRecord` - Activity tracking

#### Repository Interfaces

- `EventRepository` - Event storage and retrieval
- `RawQueryExecutor` - Direct SQL queries

#### Standard Errors

- `ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidArgument`
- `ErrPermissionDenied`, `ErrNotImplemented`, `ErrReadOnly`, `ErrInternal`

---

## Architectural Principles

### What MUST Be Here

✅ **All plugin contracts** - Interfaces that plugins implement
✅ **Entity capability interfaces** - IExtensible, ITrackable, IHasContext, etc.
✅ **Shared types** - Events, queries, context objects used across plugin boundaries
✅ **Standard errors** - Common error types for plugin operations
✅ **Pure Go standard library** - time, context, errors, io only

### What MUST NOT Be Here

❌ **Internal imports** - `internal/*` packages (SDK must be fully public)
❌ **Implementation code** - Only interfaces and data structures
❌ **Plugin-specific types** - Event types, payloads belong in plugin packages
❌ **Framework logic** - Business logic belongs in `internal/domain`
❌ **External dependencies** - Keep dependencies minimal (stdlib only preferred)

### Critical Rules

1. **Zero Internal Dependencies**: This package CANNOT import `internal/*`
2. **Single Source of Truth**: If an interface exists here, don't duplicate it elsewhere
3. **Backward Compatibility**: Changes here affect all plugins - be conservative
4. **Minimal Surface**: Only add what's needed for plugin contracts
5. **Documentation**: Every exported type must have godoc comments

---

## Entity Capability Model

**Required Interface**: Every entity MUST implement `IExtensible`
- `GetID()` - Unique identifier
- `GetType()` - Entity type name
- `GetCapabilities()` - List of capability names
- `GetField(name)` - Get any field by name
- `GetAllFields()` - Get all fields as map

**Optional Capabilities**: Declare only what you implement
- `"trackable"` → Implement `ITrackable` (status, progress, blocking)
- `"contextual"` → Implement `IHasContext` (relationships, files, activity)
- `"schedulable"` → Implement `ISchedulable` (deadlines, schedules)
- `"relatable"` → Implement `IRelatable` (parent/child relationships)

**Example**:
```go
// Minimal entity
func (e *MyEntity) GetCapabilities() []string {
    return []string{} // Only IExtensible
}

// Full-featured entity
func (e *MyEntity) GetCapabilities() []string {
    return []string{"trackable", "contextual", "schedulable"}
}
```

---

## Plugin Event Bus

The **EventBus** enables cross-plugin communication through publish/subscribe patterns. Plugins can emit events and subscribe to events from other plugins with flexible filtering.

### EventBus Interface

```go
type EventBus interface {
    // Publish sends an event to all matching subscribers
    Publish(ctx context.Context, event BusEvent) error

    // Subscribe registers a handler for events matching the filter
    Subscribe(filter EventFilter, handler EventHandler) (string, error)

    // Unsubscribe removes a subscription by its ID
    Unsubscribe(subscriptionID string) error
}
```

### BusEvent Structure

Events carry metadata, labels for filtering, and a JSON-encoded payload:

```go
type BusEvent struct {
    ID        string                 // Unique event ID (UUID)
    Type      string                 // Event type (e.g., "gmail.email_received")
    Source    string                 // Plugin ID that emitted this event
    Timestamp time.Time
    Labels    map[string]string      // Filterable key-value pairs
    Metadata  map[string]interface{} // Additional context
    Payload   []byte                 // JSON-encoded event data
}
```

**Creating Events**:
```go
// Create a new event with JSON-encoded payload
event, err := pluginsdk.NewBusEvent(
    "gmail.email_received",  // Event type
    "gmail-plugin",          // Source plugin
    emailData,               // Payload (will be JSON-marshaled)
)

// Add labels for filtering
event.Labels["category"] = "school_notification"
event.Labels["priority"] = "high"

// Publish the event
eventBus.Publish(ctx, event)
```

### Event Filtering

Subscribe to events with flexible filtering criteria:

```go
type EventFilter struct {
    TypePattern  string            // Glob pattern or exact match
    Labels       map[string]string // Required label key-value pairs
    SourcePlugin string            // Filter by source plugin ID
}
```

**Filter Examples**:
```go
// Subscribe to all Gmail events
filter := EventFilter{
    TypePattern: "gmail.*",
}

// Subscribe to school notifications from Gmail
filter := EventFilter{
    TypePattern: "gmail.*",
    Labels: map[string]string{
        "category": "school_notification",
    },
}

// Subscribe to events from specific plugin
filter := EventFilter{
    TypePattern: "*",
    SourcePlugin: "gmail-plugin",
}
```

**Type Pattern Matching**:
- `gmail.*` - All events starting with "gmail."
- `*.event_detected` - All events ending with ".event_detected"
- `gmail.email_received` - Exact match
- `*` - All events

### Event Handlers

Implement `EventHandler` to process events:

```go
type EventHandler interface {
    HandleEvent(ctx context.Context, event BusEvent) error
}
```

**Example Handler**:
```go
type TelegramNotifier struct {
    botAPI *telegram.Bot
}

func (h *TelegramNotifier) HandleEvent(ctx context.Context, event BusEvent) error {
    // Decode the payload
    var emailData EmailPayload
    if err := json.Unmarshal(event.Payload, &emailData); err != nil {
        return err
    }

    // Send notification
    message := fmt.Sprintf("New %s: %s",
        event.Labels["category"],
        emailData.Subject)
    return h.botAPI.SendMessage(ctx, message)
}
```

### Complete Usage Example

```go
// Plugin initialization
type MyPlugin struct {
    eventBus pluginsdk.EventBus
}

func (p *MyPlugin) Init(ctx context.Context, pluginCtx pluginsdk.PluginContext, eventBus pluginsdk.EventBus) error {
    p.eventBus = eventBus

    // Subscribe to events
    filter := pluginsdk.EventFilter{
        TypePattern: "gmail.*",
        Labels: map[string]string{
            "category": "school_notification",
        },
    }

    handler := &TelegramNotifier{botAPI: p.bot}
    subscriptionID, err := eventBus.Subscribe(filter, handler)
    if err != nil {
        return err
    }

    // Store subscription ID for cleanup
    p.subscriptionID = subscriptionID
    return nil
}

// Publishing events
func (p *MyPlugin) processEmail(email Email) error {
    // Create event
    event, err := pluginsdk.NewBusEvent(
        "gmail.email_received",
        "gmail-plugin",
        email,
    )
    if err != nil {
        return err
    }

    // Add labels
    event.Labels["category"] = email.Category
    event.Labels["priority"] = email.Priority

    // Publish
    return p.eventBus.Publish(context.Background(), event)
}
```

### Key Features

**Async Delivery**:
- Events are delivered to subscribers asynchronously (non-blocking)
- Each handler has a 30-second timeout
- Publisher doesn't wait for handlers to complete

**Thread-Safe**:
- Concurrent Publish/Subscribe/Unsubscribe operations are safe
- Multiple goroutines can publish simultaneously
- Handlers may be called concurrently (must be thread-safe)

**Event Persistence** (Optional):
- Events can be persisted to SQLite for replay
- Late-subscribing plugins can catch up on historical events
- Useful for rebuilding state or audit trails

**Event Replay**:
- Replay events from a specific timestamp
- Useful when plugins start after events were published
- Filter replay by event type, labels, or source

### Best Practices

1. **Event Type Naming**: Use dot notation: `<plugin>.<event_name>`
   - Examples: `gmail.email_received`, `calendar.event_created`

2. **Label Design**: Use labels for filtering, metadata for context
   - Labels: Categorical values for routing (category, priority, type)
   - Metadata: Additional context not used for filtering

3. **Payload Schema**: Document your event payload structures
   - Use consistent JSON schemas across event types
   - Version your payloads if structure may change

4. **Error Handling**: Handlers should handle errors gracefully
   - Don't crash on unexpected payloads
   - Log errors but continue processing

5. **Thread Safety**: Handlers may be called concurrently
   - Use mutexes for shared state
   - Prefer immutable data structures

6. **Cleanup**: Unsubscribe when plugin shuts down
   - Store subscription IDs
   - Call Unsubscribe in cleanup/shutdown hooks

---

## Plugin Development Workflow

1. **Import SDK**: `import "darwinflow/pkg/pluginsdk"`
2. **Implement Plugin**: Satisfy `pluginsdk.Plugin` interface
3. **Define Entities**: Implement `IExtensible` + optional capability interfaces
4. **Define Commands**: Return from `GetCommands()` if implementing `ICommandProvider`
5. **Register**: Call `pluginRegistry.RegisterPlugin(myPlugin)` in `cmd/dw`

**See**: `pkg/plugins/claude_code/` for reference implementation

---

## Files

- `capability.go` - Entity capability constants
- `command.go` - Command interfaces and types
- `entity.go` - Entity interfaces (IExtensible, ITrackable, IHasContext, etc.)
- `errors.go` - Standard error definitions
- `event.go` - Event type and EventQuery
- `event_migration.go` - Event migration helpers
- `plugin.go` - Core Plugin and PluginInfo interfaces
- `repository.go` - EventRepository and RawQueryExecutor interfaces

---

*Generated by `go-arch-lint -format=package pkg/pluginsdk`*

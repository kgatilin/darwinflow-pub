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

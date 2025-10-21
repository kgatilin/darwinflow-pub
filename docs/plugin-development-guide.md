# Plugin Development Guide

**DarwinFlow Plugin Architecture - Best Practices and Guidelines**

Last Updated: 2025-10-21

---

## Table of Contents

1. [Architecture Principles](#architecture-principles)
2. [Creating a New Plugin](#creating-a-new-plugin)
3. [Plugin Interface Implementation](#plugin-interface-implementation)
4. [Best Practices](#best-practices)
5. [Architecture Maintenance](#architecture-maintenance)
6. [Common Pitfalls](#common-pitfalls)
7. [Testing Guidelines](#testing-guidelines)

---

## Architecture Principles

### Single Source of Truth

**The Golden Rule**: `pkg/pluginsdk` is the **single source of truth** for all plugin contracts.

- ✅ **DO**: Import `pkg/pluginsdk` for all plugin interfaces
- ❌ **DON'T**: Duplicate interfaces from SDK in domain or plugin packages
- ❌ **DON'T**: Create adaptation layers between SDK and plugin code

### Dependency Rules

```
pkg/pluginsdk/          → No dependencies (fully public)
pkg/plugins/*           → Import pkg/pluginsdk only
internal/domain         → May import pkg/pluginsdk (for contracts)
internal/app            → May import pkg/pluginsdk
internal/infra          → May import pkg/pluginsdk
```

**Verification**: Run `go-arch-lint .` - must show **zero violations**.

### Framework vs Plugin Boundaries

**Framework** (`internal/domain`, `internal/app`, `internal/infra`):
- Generic, plugin-agnostic logic
- Event storage (generic `Event` struct with `Type: string`)
- Repository interfaces for storage
- Zero knowledge of specific plugin schemas

**Plugin** (`pkg/plugins/claude_code`, etc.):
- Plugin-specific event types (e.g., `ChatStarted`, `ToolInvoked`)
- Plugin-specific payload schemas (`ChatPayload`, `ToolPayload`)
- Plugin-specific analysis logic
- Hooks and integration points

**Key Insight**: If the framework needs to know about it, it's not plugin-specific. If only one plugin uses it, it belongs in the plugin package.

---

## Creating a New Plugin

### Step 1: Create Plugin Package

```bash
mkdir -p pkg/plugins/myplugin
```

### Step 2: Define Plugin Struct

```go
// pkg/plugins/myplugin/plugin.go
package myplugin

import "github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"

type MyPlugin struct {
    logger pluginsdk.Logger
    // Add dependencies (repositories, config, etc.)
}

func NewMyPlugin(logger pluginsdk.Logger) *MyPlugin {
    return &MyPlugin{
        logger: logger,
    }
}
```

### Step 3: Implement Plugin Interface

```go
func (p *MyPlugin) GetInfo() pluginsdk.PluginInfo {
    return pluginsdk.PluginInfo{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "Description of what this plugin does",
        IsCore:      true, // or false for external plugins
    }
}

func (p *MyPlugin) GetCapabilities() []string {
    return []string{
        "IEntityProvider",
        "ICommandProvider",
    }
}
```

### Step 4: Implement Capability Interfaces

#### IEntityProvider Example

```go
func (p *MyPlugin) GetEntityTypes() []pluginsdk.EntityTypeInfo {
    return []pluginsdk.EntityTypeInfo{
        {
            Type:              "mytask",
            DisplayName:       "My Task",
            DisplayNamePlural: "My Tasks",
            Capabilities:      []string{"IExtensible", "ITrackable"},
            Icon:              "📋",
            Description:       "Tasks tracked by my plugin",
        },
    }
}

func (p *MyPlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
    // Implement entity querying logic
    // Return entities that implement IExtensible (and optionally ITrackable, IHasContext, etc.)
}

func (p *MyPlugin) GetEntity(ctx context.Context, entityID string) (pluginsdk.IExtensible, error) {
    // Implement entity retrieval by ID
}
```

#### ICommandProvider Example

```go
func (p *MyPlugin) GetCommands() []pluginsdk.Command {
    return []pluginsdk.Command{
        &InitCommand{plugin: p},
        &StatusCommand{plugin: p},
    }
}
```

### Step 5: Define Plugin-Specific Types

```go
// pkg/plugins/myplugin/event_types.go
package myplugin

const (
    TaskCreated  = "myplugin.task.created"
    TaskUpdated  = "myplugin.task.updated"
    TaskDeleted  = "myplugin.task.deleted"
)

// pkg/plugins/myplugin/payloads.go
package myplugin

type TaskPayload struct {
    TaskID      string `json:"task_id"`
    Title       string `json:"title"`
    Status      string `json:"status"`
    Priority    int    `json:"priority"`
}
```

### Step 6: Create Entity Implementation

```go
// pkg/plugins/myplugin/task.go
package myplugin

type Task struct {
    ID       string
    Title    string
    Status   string
    Priority int
}

// IExtensible implementation (REQUIRED)
func (t *Task) GetID() string { return t.ID }
func (t *Task) GetType() string { return "mytask" }
func (t *Task) GetCapabilities() []string {
    return []string{"IExtensible", "ITrackable"}
}
func (t *Task) GetField(name string) interface{} {
    switch name {
    case "title": return t.Title
    case "status": return t.Status
    case "priority": return t.Priority
    default: return nil
    }
}
func (t *Task) GetAllFields() map[string]interface{} {
    return map[string]interface{}{
        "id":       t.ID,
        "title":    t.Title,
        "status":   t.Status,
        "priority": t.Priority,
    }
}

// ITrackable implementation (OPTIONAL)
func (t *Task) GetStatus() string { return t.Status }
func (t *Task) GetProgress() float64 {
    if t.Status == "done" { return 1.0 }
    if t.Status == "in_progress" { return 0.5 }
    return 0.0
}
func (t *Task) IsBlocked() bool { return t.Status == "blocked" }
func (t *Task) GetBlockReason() string {
    if t.IsBlocked() { return "Waiting for dependencies" }
    return ""
}
```

### Step 7: Register Plugin

```go
// cmd/dw/plugin_registration.go
import "github.com/kgatilin/darwinflow-pub/pkg/plugins/myplugin"

func initializePlugins(services AppServices) []domain.Plugin {
    // ...existing plugins...

    myPlugin := myplugin.NewMyPlugin(services.Logger)

    return []domain.Plugin{
        claudePlugin,
        myPlugin, // Add your plugin
    }
}
```

---

## Plugin Interface Implementation

### Required vs Optional Capabilities

**Required for all plugins:**
- `Plugin` interface (GetInfo, GetCapabilities)

**Optional capabilities** (declare in GetCapabilities):
- `IEntityProvider` - Provide queryable entities
- `IEntityUpdater` - Allow entity updates
- `ICommandProvider` - Provide CLI commands
- `IEventEmitter` - Emit events (use PluginContext.EmitEvent)

**Required for all entities:**
- `IExtensible` - Base capability (GetID, GetType, GetCapabilities, GetField, GetAllFields)

**Optional entity capabilities:**
- `IHasContext` - Contextual data (files, activity, metadata)
- `ITrackable` - Status and progress tracking
- `ISchedulable` - Time-based scheduling
- `IRelatable` - Explicit entity relationships

### Event Emission Pattern

```go
// In your plugin's command or service
func (p *MyPlugin) doSomething(ctx context.Context, pluginCtx pluginsdk.PluginContext) error {
    // Emit event using SDK Event struct
    event := pluginsdk.Event{
        Type:      myplugin.TaskCreated,
        Source:    "my-plugin",
        Timestamp: time.Now(),
        Payload: map[string]interface{}{
            "task_id": "task-123",
            "title":   "New task",
        },
        Metadata: map[string]string{
            "user_id": "user-456",
        },
        Version: "1.0",
    }

    return pluginCtx.EmitEvent(ctx, event)
}
```

---

## Best Practices

### 1. Zero Interface Duplication

❌ **Bad**: Duplicating SDK interfaces in your plugin
```go
// DON'T DO THIS
package myplugin

type IExtensible interface { // ❌ Already in SDK!
    GetID() string
    GetType() string
}
```

✅ **Good**: Import and use SDK interfaces
```go
package myplugin

import "github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"

func (p *MyPlugin) Query(...) ([]pluginsdk.IExtensible, error) {
    // Use SDK types directly
}
```

### 2. Plugin-Specific Types Stay in Plugin

❌ **Bad**: Plugin-specific types in framework
```go
// internal/domain/event.go
const ChatStarted = "chat.started" // ❌ Claude-specific!
```

✅ **Good**: Plugin-specific types in plugin package
```go
// pkg/plugins/claude_code/event_types.go
const ChatStarted = "claude.chat.started" // ✅ Plugin-owned
```

### 3. Use Dot Notation for Event Types

```go
// Format: <plugin>.<domain>.<action>
const (
    TaskCreated  = "myplugin.task.created"
    TaskUpdated  = "myplugin.task.updated"
    UserLoggedIn = "myplugin.user.logged_in"
)
```

### 4. Implement Only What You Need

Don't implement optional capabilities unless you actually use them:

```go
// If you don't have entities, don't implement IEntityProvider
func (p *MyPlugin) GetCapabilities() []string {
    return []string{
        "ICommandProvider", // Only what you need
    }
}
```

### 5. Entity Capabilities Match Reality

```go
// Only declare capabilities you actually implement
func (t *Task) GetCapabilities() []string {
    return []string{
        "IExtensible",  // Always required
        "ITrackable",   // Only if you implement it
        // Don't add ISchedulable unless you implement GetStartDate, GetDueDate, IsOverdue
    }
}
```

### 6. Context Usage

Always use `context.Context` for cancellation and timeouts:

```go
func (p *MyPlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Your logic...
}
```

### 7. Error Handling

Use SDK error constants when appropriate:

```go
import "github.com/kgatilin/darwinflow-pub/internal/domain"

func (p *MyPlugin) GetEntity(ctx context.Context, id string) (pluginsdk.IExtensible, error) {
    entity, found := p.findEntity(id)
    if !found {
        return nil, domain.ErrNotFound
    }
    return entity, nil
}
```

---

## Architecture Maintenance

### Preventing Duplication

**Before adding any interface or type, ask:**
1. Does this already exist in `pkg/pluginsdk`?
2. Should this be in SDK (plugin contract) or plugin-specific?
3. Will external plugins need this?

**Verification checklist:**
```bash
# After any changes, verify architecture
go-arch-lint .                    # Must show zero violations
go test ./...                     # All tests must pass
go-arch-lint docs                 # Regenerate documentation
```

### Updating SDK

If you need to add to SDK (for external plugin use):

1. **Add to pkg/pluginsdk** first
2. Update all plugins to use it
3. Never duplicate - SDK is single source of truth
4. Consider backward compatibility (versioning)

### Code Review Checklist

- [ ] Plugin imports only `pkg/pluginsdk` (no `internal/*`)
- [ ] No interface duplication between SDK and plugin
- [ ] Plugin-specific types are in plugin package
- [ ] `go-arch-lint .` shows zero violations
- [ ] All tests pass
- [ ] Documentation updated if needed

---

## Common Pitfalls

### ❌ Pitfall 1: Importing Internal Packages

```go
// ❌ DON'T DO THIS
package myplugin

import "github.com/kgatilin/darwinflow-pub/internal/domain"

// Violates architecture rules - plugins can't import internal/*
```

**Fix**: Use only `pkg/pluginsdk` types.

### ❌ Pitfall 2: Creating Adaptation Layers

```go
// ❌ DON'T DO THIS
func adaptToSDK(domainEntity *domain.Entity) pluginsdk.IExtensible {
    // Unnecessary conversion
}
```

**Fix**: Work with SDK types directly. No adaptation needed.

### ❌ Pitfall 3: Framework Depending on Plugin

```go
// internal/domain/event.go
import "github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"

const ChatStarted = claude_code.ChatStarted // ❌ Framework depends on plugin!
```

**Fix**: Framework should be plugin-agnostic. Move types to SDK or keep them plugin-specific.

### ❌ Pitfall 4: Mixing Framework and Plugin Concerns

```go
// pkg/plugins/myplugin/repository.go
type TaskRepository interface {
    GetAllTasks() []Task // ✅ Plugin-specific
    SaveEvent(*domain.Event) error // ❌ Framework concern!
}
```

**Fix**: Plugin repositories are plugin-specific. Don't mix with framework repositories.

### ❌ Pitfall 5: Not Declaring Capabilities

```go
func (t *Task) GetStatus() string { return t.status }

func (t *Task) GetCapabilities() []string {
    return []string{"IExtensible"} // ❌ Implements ITrackable but doesn't declare it
}
```

**Fix**: Always declare all capabilities you implement.

---

## Testing Guidelines

### Unit Test Structure

```go
// pkg/plugins/myplugin/plugin_test.go
package myplugin_test // Use blackbox testing

import (
    "testing"
    "github.com/kgatilin/darwinflow-pub/pkg/plugins/myplugin"
    "github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

func TestPlugin_GetInfo(t *testing.T) {
    plugin := myplugin.NewMyPlugin(nil)
    info := plugin.GetInfo()

    if info.Name != "my-plugin" {
        t.Errorf("Expected name 'my-plugin', got %q", info.Name)
    }
}
```

### Test Entity Capabilities

```go
func TestTask_ImplementsIExtensible(t *testing.T) {
    task := &myplugin.Task{ID: "test-1", Title: "Test"}

    // Verify interface compliance at compile time
    var _ pluginsdk.IExtensible = task

    // Verify capabilities are declared
    caps := task.GetCapabilities()
    if !contains(caps, "IExtensible") {
        t.Error("Task should declare IExtensible capability")
    }
}
```

### Integration Tests

Test plugin integration with framework:

```go
func TestPlugin_QueryEntities(t *testing.T) {
    plugin := myplugin.NewMyPlugin(logger)

    entities, err := plugin.Query(context.Background(), pluginsdk.EntityQuery{
        EntityType: "mytask",
        Limit:      10,
    })

    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }

    for _, entity := range entities {
        if entity.GetType() != "mytask" {
            t.Errorf("Expected type 'mytask', got %q", entity.GetType())
        }
    }
}
```

---

## Quick Reference

### Plugin Checklist

- [ ] Plugin struct implements `pluginsdk.Plugin`
- [ ] Implements declared capabilities (IEntityProvider, ICommandProvider, etc.)
- [ ] Entities implement `pluginsdk.IExtensible` (required)
- [ ] Entities declare all implemented capabilities
- [ ] Plugin-specific types in plugin package (event types, payloads, etc.)
- [ ] Imports only `pkg/pluginsdk` (no `internal/*`)
- [ ] Tests verify interface compliance
- [ ] `go-arch-lint .` passes with zero violations

### SDK Interface Summary

**Plugin Capabilities:**
- `Plugin` - Base interface (required)
- `IEntityProvider` - Provide queryable entities
- `IEntityUpdater` - Update entity fields
- `ICommandProvider` - Provide CLI commands
- `IEventEmitter` - Emit events (marker interface)

**Entity Capabilities:**
- `IExtensible` - Base entity (required)
- `IHasContext` - Contextual data
- `ITrackable` - Status and progress
- `ISchedulable` - Time-based scheduling
- `IRelatable` - Entity relationships

**Support Types:**
- `Event` - Event emission
- `PluginInfo` - Plugin metadata
- `EntityTypeInfo` - Entity type metadata
- `EntityQuery` - Entity querying
- `Logger` - Logging interface

---

## Additional Resources

- **Architecture Index**: `docs/arch-index.md` - Current package structure
- **SDK Reference**: `pkg/pluginsdk/` - Godoc documentation
- **Example Plugin**: `pkg/plugins/claude_code/` - Reference implementation
- **Architecture Linter**: Run `go-arch-lint .` to validate compliance

---

**Remember**: The SDK is the single source of truth. When in doubt, check `pkg/pluginsdk` first!

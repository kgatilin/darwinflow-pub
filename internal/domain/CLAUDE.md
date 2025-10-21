# Package: domain

**Path**: `internal/domain`

**Role**: Framework business logic and domain models (plugin-agnostic)

---

## Quick Reference

- **Files**: 10
- **Exports**: 43
- **Dependencies**: None (pure domain layer)
- **Layer**: Business logic (depends on nothing)

---

## Generated Documentation

### Exported API

#### Core Domain Types

**Events**:
- `Event` - Core event type (ID, Timestamp, Type, SessionID, Payload, Content, Version)
- `EventType` - Event type enumeration
- `EventQuery` - Query events with filters

**Analysis**:
- `SessionAnalysis` - Analysis result entity
- `AnalysisRepository` - Analysis persistence interface
- `ToolSuggestion` - AI-generated tool suggestions

**Configuration**:
- `Config` - Application configuration
- `AnalysisConfig` - Analysis settings (TokenLimit, Model, ParallelLimit, EnabledPrompts, etc.)
- `ClaudeOptions` - Claude CLI options (AllowedTools, SystemPromptMode)
- `UIConfig` - TUI configuration

**Repositories** (interfaces):
- `EventRepository` - Event storage interface
- `AnalysisRepository` - Analysis storage interface
- `RawQueryExecutor` - Direct query interface

**Query Types**:
- `EventQuery` - Filter events (Time, Types, SessionID, Context, SearchText)
- `QueryResult` - Query result data (Columns, Rows)

#### Functions

- `DefaultConfig()` - Creates default configuration
- `NewEvent()` - Event factory
- `NewSessionAnalysis()` - Analysis factory
- `ValidateModel()` - Model validation

#### Constants

**Trigger Types**:
- `TriggerUserInput`, `TriggerUserPromptSubmit`, `TriggerSessionStart`, `TriggerSessionEnd`
- `TriggerBeforeToolUse`, `TriggerAfterToolUse`, `TriggerPreToolUse`, `TriggerPostToolUse`
- `TriggerNotification`, `TriggerStop`, `TriggerSubagentStop`, `TriggerPreCompact`

**Default Prompts**:
- `DefaultAnalysisPrompt` - Default session analysis prompt
- `DefaultSessionSummaryPrompt` - Session summary prompt
- `DefaultToolAnalysisPrompt` - Tool usage analysis prompt

**Standard Errors**:
- `ErrNotFound`, `ErrAlreadyExists`, `ErrInvalidArgument`
- `ErrPermissionDenied`, `ErrNotImplemented`, `ErrReadOnly`, `ErrInternal`

---

## Architectural Principles

### What MUST Be Here

✅ **Framework domain models** - Events, Config, Analysis (framework-level entities)
✅ **Repository interfaces** - Define what infrastructure must implement
✅ **Business logic** - Domain rules, validation, factories
✅ **Framework constants** - Trigger types, default prompts, error types
✅ **Framework-level queries** - Event queries, analysis queries
✅ **Pure domain types** - No infrastructure concerns

### What MUST NOT Be Here

❌ **Plugin-specific types** - Event types, payloads, analysis types belong in plugin packages
❌ **Infrastructure code** - DB, file I/O, HTTP clients belong in `internal/infra`
❌ **Application logic** - Orchestration belongs in `internal/app`
❌ **Implementation details** - Only interfaces, not concrete implementations
❌ **External dependencies** - Minimize external imports (stdlib + UUID only)
❌ **Duplicate SDK interfaces** - Use `pkg/pluginsdk` interfaces directly

### Critical Rules

1. **Plugin Agnostic**: This layer knows NOTHING about specific plugins
2. **No Infrastructure**: Never import `internal/infra` or `internal/app`
3. **Define Interfaces**: Infrastructure implements, domain defines
4. **Dependency Inversion**: Depend on abstractions, not concretions
5. **Pure Business Logic**: No side effects, testable without I/O

---

## Domain Design Patterns

### Repository Pattern

**Domain defines the interface**:
```go
type EventRepository interface {
    Save(ctx context.Context, event *Event) error
    FindByQuery(ctx context.Context, query EventQuery) ([]*Event, error)
}
```

**Infrastructure implements**:
```go
// internal/infra/sqlite_repository.go
type SQLiteEventRepository struct { ... }
func (r *SQLiteEventRepository) Save(ctx context.Context, event *Event) error { ... }
```

**Application injects**:
```go
// internal/app/analysis.go
type AnalysisService struct {
    repo domain.EventRepository // Interface, not concrete type
}
```

### Factory Pattern

Use constructors for domain entities:
- `NewEvent()` - Ensures valid event creation
- `NewSessionAnalysis()` - Sets defaults and validates
- `DefaultConfig()` - Provides sensible defaults

### Value Objects

Immutable types with value semantics:
- `EventQuery` - Query parameters
- `ClaudeOptions` - Configuration options
- `TriggerType` - Enumeration type

---

## Framework vs Plugin Separation

**Framework (this package)**:
- Generic `Event` type (ID, Type, Timestamp, Payload)
- Generic `EventQuery` (filters applicable to any event)
- Generic `SessionAnalysis` (analysis results)
- Trigger types (hook points in event lifecycle)

**Plugins** (`pkg/plugins/*/`):
- Specific event types: `ChatStarted`, `ToolInvoked`, `FileWritten`
- Specific payloads: `ChatPayload`, `ToolPayload`, `FilePayload`
- Plugin-specific analysis types
- Plugin-specific commands

**Rule**: If it's specific to Claude Code or any other tool → Plugin package, not here

---

## Event Versioning

All events have a `Version` field for schema evolution:
- Version 1: Original schema
- Version 2+: Breaking changes require migration

When adding new event fields:
1. Add to domain Event struct
2. Update version constant
3. Add migration in `internal/infra`
4. Test backward compatibility

---

## Configuration Management

**Config hierarchy**:
```
Config
├── Analysis (AnalysisConfig)
│   ├── TokenLimit, Model, ParallelLimit
│   ├── EnabledPrompts
│   └── ClaudeOptions (AllowedTools, SystemPromptMode)
└── UI (UIConfig)
    ├── DefaultOutputDir
    ├── FilenameTemplate
    └── AutoRefreshInterval
```

**Validation**: Use `ValidateModel()` to ensure valid model names

---

## Files

- `analysis.go` - SessionAnalysis domain type
- `config.go` - Configuration domain models
- `event.go` - Event domain type and factories
- `plugin.go` - (Legacy - being migrated to SDK)
- `repository.go` - Repository interfaces
- `trigger.go` - Trigger type constants
- `*_test.go` - Domain logic tests

---

*Generated by `go-arch-lint -format=package internal/domain`*

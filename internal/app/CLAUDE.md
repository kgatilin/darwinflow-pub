# Package: app

**Path**: `internal/app`

**Role**: Application layer - orchestration, services, command handlers

---

## Quick Reference

- **Files**: 26
- **Exports**: 96
- **Dependencies**: `internal/domain`, `pkg/pluginsdk`
- **Layer**: Application (orchestration and coordination)

---

## Generated Documentation

### Exported API

#### Core Services

**AnalysisService**:
- Session analysis orchestration
- Methods: `AnalyzeSession`, `AnalyzeSessionWithPrompt`, `AnalyzeMultipleSessions`, `GetAnalysis`
- Token estimation and budget management
- Parallel analysis support

**LogsService**:
- Event query and formatting
- Methods: `ListRecentLogs`, `ExecuteRawQuery`
- Multiple output formats (plain, CSV, markdown)

**SetupService**:
- Database initialization
- Method: `Initialize`

**LoggerService**:
- Event logging from hooks
- Method: `LogEvent`

#### Plugin System

**PluginRegistry**:
- Central plugin management
- Methods: `RegisterPlugin`, `GetPlugin`, `GetEntity`, `Query`, `UpdateEntity`
- Entity provider aggregation
- Command provider aggregation
- Event emitter coordination

**CommandRegistry**:
- Command routing and execution
- Methods: `ExecuteCommand`, `GetCommand`, `ListCommands`
- Plugin-scoped command dispatch

#### Command Handlers

**AnalyzeCommandHandler**:
- `dw analyze` command
- Options: SessionID, Last, ViewOnly, AnalyzeAll, Refresh, PromptNames

**LogsCommandHandler**:
- `dw logs` command
- Query execution and formatting

**ConfigCommandHandler**:
- `dw config` commands
- Methods: `Init`, `Show`

**RefreshCommandHandler**:
- `dw refresh` command
- Reanalyze unanalyzed sessions

#### LLM Integration

**ClaudeCLIExecutor**:
- Executes Claude CLI for AI analysis
- Methods: `Execute`, `ExecuteWithOptions`
- Configurable tools and system prompts

#### Utilities

**Formatting Functions**:
- `FormatLogRecord`, `FormatLogsAsCSV`, `FormatLogsAsMarkdown`
- `FormatQueryValue` - Query result formatting

**Context Builders**:
- `NewCommandContext` - Command execution context
- `NewPluginContext` - Plugin initialization context

#### Types

**Options structs**:
- `AnalyzeOptions` - Analysis parameters
- `LogsOptions` - (Defined in cmd/dw)

**Data structures**:
- `LogRecord` - Formatted event record
- `SessionGroup` - Grouped session events
- `AnalysisResult` - Analysis with error handling
- `FilenameTmplData` - Template data for filenames
- `ProjectContext` - Project-level dependencies

---

## Architectural Principles

### What MUST Be Here

✅ **Application services** - Business workflow orchestration
✅ **Command handlers** - CLI command implementations
✅ **Plugin coordination** - Registry, routing, aggregation
✅ **Use case implementations** - Analysis, logging, querying
✅ **Service composition** - Combine domain + infrastructure
✅ **External integration** - Claude CLI, file exports

### What MUST NOT Be Here

❌ **Domain logic** - Business rules belong in `internal/domain`
❌ **Infrastructure** - DB, file I/O belong in `internal/infra`
❌ **UI logic** - TUI belongs in `internal/app/tui`
❌ **Main entry points** - Belong in `cmd/dw`
❌ **Plugin implementations** - Belong in `pkg/plugins/*`

### Critical Rules

1. **Orchestration, Not Logic**: Coordinate domain + infra, don't duplicate logic
2. **Service Layer**: Services implement use cases, not domain rules
3. **Dependency Injection**: Services receive dependencies via constructors
4. **Interface Usage**: Depend on `domain` interfaces, not `infra` concrete types
5. **Testability**: All services should be testable with mock dependencies

---

## Service Architecture

### Dependency Injection Pattern

```go
// Service constructor
func NewAnalysisService(
    eventRepo domain.EventRepository,      // Interface from domain
    analysisRepo domain.AnalysisRepository, // Interface from domain
    logsService *LogsService,               // Concrete from app
    llm LLMExecutor,                        // Interface from app
    logger Logger,                          // Interface from app
    config *domain.Config,                  // Domain type
) *AnalysisService {
    return &AnalysisService{...}
}
```

**Benefits**:
- Easy to test (inject mocks)
- Flexible implementations
- Clear dependencies

### Service Composition

Services may depend on:
- **Domain repositories** (via interfaces)
- **Other app services** (composition)
- **Infrastructure utilities** (logger, config)

```go
type AnalysisService struct {
    eventRepo    domain.EventRepository   // Domain interface
    analysisRepo domain.AnalysisRepository // Domain interface
    logsService  *LogsService              // App service
    llm          LLMExecutor               // App interface
    logger       Logger                    // App interface
    config       *domain.Config            // Domain type
}
```

---

## Plugin System

### PluginRegistry

**Central hub for all plugins**:
- Registers plugins at startup
- Routes entity queries to correct plugin
- Aggregates commands from all plugins
- Coordinates event emitters

**Key operations**:
```go
// Registration
registry.RegisterPlugin(myPlugin)

// Entity access
entity, err := registry.GetEntity(ctx, entityID)

// Querying
entities, err := registry.Query(ctx, query)

// Updating
updated, err := registry.UpdateEntity(ctx, entityID, updates)
```

### CommandRegistry

**Command routing**:
- Maps command names to plugin commands
- Handles namespaced commands (`dw plugin-name command`)
- Provides command help and listing

**Execution flow**:
```
User: dw claude init
  ↓
CommandRegistry.ExecuteCommand("claude", ["init"], ctx)
  ↓
ClaudeCodePlugin.GetCommands() → InitCommand
  ↓
InitCommand.Execute(ctx, ["init"])
```

---

## Analysis Service

### Workflow

1. **Query events**: Fetch session events from repository
2. **Estimate tokens**: Calculate token count for budget
3. **Build context**: Assemble events into analysis context
4. **Execute LLM**: Call Claude CLI with prompt
5. **Parse result**: Extract analysis from LLM response
6. **Save analysis**: Persist to repository
7. **Return result**: Provide analysis to caller

### Features

**Single session analysis**:
```go
analysis, err := service.AnalyzeSession(ctx, sessionID)
```

**Multiple prompts**:
```go
analyses, errs := service.AnalyzeSessionWithMultiplePrompts(ctx, sessionID, promptNames)
```

**Parallel batch**:
```go
analyses, errs := service.AnalyzeMultipleSessionsParallel(ctx, sessionIDs, promptName)
```

**Token budgeting**:
```go
selected, totalTokens, err := service.SelectSessionsWithinTokenLimit(ctx, sessionIDs, tokenLimit)
```

---

## Command Handler Pattern

### Structure

```go
type AnalyzeCommandHandler struct {
    service AnalysisServiceInterface  // Service interface
    logger  Logger                     // Logger interface
    output  io.Writer                  // Output destination
}

func (h *AnalyzeCommandHandler) Execute(ctx context.Context, opts AnalyzeOptions) error {
    // 1. Validate options
    // 2. Call service
    // 3. Format output
    // 4. Handle errors
}
```

### Responsibilities

- Parse and validate command options
- Call appropriate service methods
- Format results for user output
- Handle errors gracefully
- Provide user feedback

---

## Logger Service

**Purpose**: Capture events from Claude Code hooks

**Input**: Hook data (transcript, tool use, user input)

**Output**: Domain events persisted to repository

**Process**:
1. Parse hook input
2. Detect context (Git repo)
3. Extract relevant content
4. Normalize content
5. Create domain event
6. Save to repository

---

## Context Builders

### CommandContext

For plugin commands:
```go
ctx := NewCommandContext(
    logger,      // Logger
    cwd,         // Working directory
    projectData, // Project-specific data
    output,      // Output writer
    input,       // Input reader
)
```

### PluginContext

For plugin initialization:
```go
ctx := NewPluginContext(
    logger,    // Logger
    cwd,       // Working directory
    eventRepo, // Event repository
)
```

---

## Formatting

### Log Records

**Plain text**:
```go
FormatLogRecord(index, record)
```

**CSV**:
```go
FormatLogsAsCSV(writer, records)
```

**Markdown**:
```go
FormatLogsAsMarkdown(writer, records)
```

### Analysis Output

**Markdown export**:
```go
path, err := service.SaveToMarkdown(ctx, analysis, outputDir)
```

**Template support**:
- Session ID
- Prompt name
- Date/time
- Custom filename templates

---

## Testing Strategy

**Service testing**:
- Mock repositories (domain interfaces)
- Mock LLM executor
- Mock logger
- Test business workflows

**Command handler testing**:
- Mock services
- Capture output
- Verify formatting
- Test error handling

**Integration testing**:
- Real repositories (temp database)
- Real formatters
- End-to-end workflows

---

## Files

- `analysis.go` - AnalysisService implementation
- `analysis_prompt.go` - Default prompts
- `analyze_cmd.go` - Analyze command handler
- `command_registry.go` - Command routing
- `config_handler.go` - Config command handler
- `logger.go` - NoOpLogger utility
- `logger_service.go` - Event logging service
- `logs.go` - LogsService implementation
- `logs_cmd.go` - Logs command handler
- `plugin_context.go` - Context builders
- `plugin_registry.go` - Plugin registration and routing
- `refresh_cmd.go` - Refresh command handler
- `setup.go` - SetupService implementation
- `*_test.go` - Service and handler tests

---

*Generated by `go-arch-lint -format=package internal/app`*

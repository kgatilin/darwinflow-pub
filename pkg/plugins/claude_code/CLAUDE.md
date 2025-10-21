# Package: claude_code

**Path**: `pkg/plugins/claude_code`

**Role**: Claude Code plugin - event capture, analysis, and tooling for Claude Code interactions

---

## Quick Reference

- **Files**: 17
- **Exports**: 102
- **Dependencies**: `pkg/pluginsdk` only
- **Layer**: Plugin implementation (reference plugin)
- **Plugin Type**: Core (ships with DarwinFlow)

---

## Generated Documentation

### Exported API

#### Plugin Implementation

**ClaudeCodePlugin**:
- Implements `pluginsdk.Plugin`, `pluginsdk.IEntityProvider`, `pluginsdk.ICommandProvider`
- Capabilities: `"entity_provider"`, `"command_provider"`
- Entity type: `"session"`
- Commands: `init`, `log`, `session-summary`, `auto-summary`, `auto-summary-exec`, `emit-event`

#### Entities

**SessionEntity**:
- Implements `pluginsdk.IExtensible`, `pluginsdk.ITrackable`, `pluginsdk.IHasContext`
- Capabilities: `"trackable"`, `"contextual"`
- Properties: SessionID, FirstEvent, EventCount, Analyses, LatestAnalysis, TokenCount
- Methods: `GetStatus`, `GetProgress`, `IsBlocked`, `GetContext`, `GetAnalyses`, `GetLatestAnalysis`

**SessionAnalysisData**:
- Analysis metadata for entity
- Properties: ID, SessionID, PromptName, ModelUsed, PatternsSummary, CreatedAt

#### Commands

**InitCommand**:
- `dw claude init` - Install Claude Code hooks
- Usage: `dw claude init [--force]`

**LogCommand**:
- `dw claude log` - Manual event logging
- Usage: `dw claude log --type <type> --payload <json>`

**SessionSummaryCommand**:
- `dw claude session-summary` - Summarize session
- Usage: `dw claude session-summary --session-id <id>`

**AutoSummaryCommand**:
- `dw claude auto-summary` - Enable auto-summary
- Usage: `dw claude auto-summary --enable`

**AutoSummaryExecCommand**:
- `dw claude auto-summary-exec` - Execute auto-summary hook
- Usage: Internal (called by hook)

**EmitEventCommand**:
- `dw claude emit-event` - Emit custom events
- Usage: `dw claude emit-event --type <type> --data <json>`

#### Event Types (Plugin-Specific)

**Constants**:
- `ChatStarted`, `ChatMessageUser`, `ChatMessageAssistant`
- `ToolInvoked`, `ToolResult`
- `FileRead`, `FileWritten`
- `ContextChanged`, `Error`
- `TriggerUserInput`, `TriggerSessionEnd`, `TriggerBeforeToolUse`

#### Payloads (Plugin-Specific)

**ChatPayload**:
- Properties: Message, Context

**ToolPayload**:
- Properties: Tool, Parameters, Result, DurationMs, Context

**FilePayload**:
- Properties: FilePath, Changes, DurationMs, Context

**ContextPayload**:
- Properties: Context, Description

**ErrorPayload**:
- Properties: Error, StackTrace, Context

#### Hook Management

**HookConfigManager**:
- Methods: `InstallDarwinFlowHooks`, `ReadSettings`, `WriteSettings`, `GetSettingsPath`
- Manages `.claude/settings.json` hooks

**HookConfig**:
- Hook configuration structure
- Properties: Hooks (map of trigger → matchers)

**HookMatcher**:
- Hook trigger matching
- Properties: Matcher (pattern), Hooks (actions)

**HookAction**:
- Hook action definition
- Properties: Type, Command, Timeout

**HookInputParser**:
- Parse hook input from Claude Code
- Extracts SessionID, TranscriptPath, ToolName, etc.

#### Analysis

**SessionAnalysis**:
- Plugin-specific analysis type
- Properties: ID, SessionID, AnalyzedAt, AnalysisResult, ModelUsed, PromptUsed, PatternsSummary, AnalysisType, PromptName

**ToolSuggestion**:
- AI-generated tool suggestions
- Properties: Name, Description, Rationale, Examples

#### Repositories (Interfaces)

**AnalysisRepository**:
- Plugin-specific analysis storage
- Methods: SaveAnalysis, GetAnalysisBySessionID, GetAllAnalyses

#### Services (Interfaces)

**AnalysisService**:
- Analysis orchestration
- Methods: AnalyzeSession, GetAnalysis, SaveToMarkdown

**LogsService**:
- Event query and logging
- Methods: ListRecentLogs, ExecuteRawQuery

**SetupService**:
- Database setup
- Method: Initialize

#### Utilities

**Functions**:
- `NewClaudeCodePlugin` - Plugin constructor
- `DefaultDarwinFlowConfig` - Default hook configuration
- `MergeHookConfigs` - Merge hook configs
- `HookInputToEvent` - Convert hook input to event
- `NewHookConfigManager` - Hook manager constructor
- `NewHookInputParser` - Parser constructor

---

## Architectural Principles

### What MUST Be Here

✅ **Plugin implementation** - ClaudeCodePlugin
✅ **Plugin-specific entities** - SessionEntity, SessionAnalysisData
✅ **Plugin-specific commands** - Init, Log, SessionSummary, etc.
✅ **Plugin-specific event types** - ChatStarted, ToolInvoked, FileWritten, etc.
✅ **Plugin-specific payloads** - ChatPayload, ToolPayload, FilePayload, etc.
✅ **Plugin-specific analysis** - SessionAnalysis, ToolSuggestion
✅ **Hook integration** - Hook config, parsing, installation

### What MUST NOT Be Here

❌ **Framework logic** - Belongs in `internal/domain`, `internal/app`
❌ **Infrastructure** - DB, file I/O belong in `internal/infra`
❌ **Generic interfaces** - Belong in `pkg/pluginsdk`
❌ **Other plugin code** - Each plugin is isolated
❌ **Internal dependencies** - Only import `pkg/pluginsdk`

### Critical Rules

1. **SDK Only**: Import `pkg/pluginsdk` only (no `internal/*`)
2. **Plugin-Specific Types**: All Claude Code specific types belong here
3. **Dependency Injection**: Receive services via constructor
4. **Self-Contained**: Plugin should be extractable to external package
5. **Follow SDK Contracts**: Implement SDK interfaces correctly

---

## Plugin Architecture

### Plugin Lifecycle

1. **Registration**: `RegisterBuiltInPlugins()` in `cmd/dw/plugin_registration.go`
2. **Initialization**: Constructor receives services
3. **Discovery**: PluginRegistry queries capabilities
4. **Execution**: Commands/queries routed to plugin

### Dependency Injection

```go
plugin := claude_code.NewClaudeCodePlugin(
    analysisService,  // *app.AnalysisService
    logsService,      // *app.LogsService
    logger,           // pluginsdk.Logger
    setupService,     // *app.SetupService
    configLoader,     // app.ConfigLoader
    dbPath,           // string
)
```

**Services injected from application layer**

---

## Entity Implementation

### SessionEntity Capabilities

**IExtensible** (required):
- `GetID()` → Session ID
- `GetType()` → `"session"`
- `GetCapabilities()` → `["trackable", "contextual"]`
- `GetField(name)` → Dynamic field access
- `GetAllFields()` → All fields as map

**ITrackable** (optional):
- `GetStatus()` → Status based on analysis count
- `GetProgress()` → Analysis progress (0.0 - 1.0)
- `IsBlocked()` → Always false
- `GetBlockReason()` → Empty string

**IHasContext** (optional):
- `GetContext()` → EntityContext with related analyses, files, activity

### Entity Fields

Accessible via `GetField()`:
- `session_id`, `first_event`, `last_event`
- `event_count`, `analysis_count`, `token_count`
- `has_analysis`, `latest_analysis`, `analyses`

---

## Command Implementation

### Command Pattern

```go
type InitCommand struct {
    plugin *ClaudeCodePlugin
}

func (c *InitCommand) GetName() string {
    return "init"
}

func (c *InitCommand) GetDescription() string {
    return "Install DarwinFlow hooks for Claude Code"
}

func (c *InitCommand) GetUsage() string {
    return "dw claude init [--force]"
}

func (c *InitCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
    // Implementation
}
```

### Command Execution

**Context**: `pluginsdk.CommandContext`
- Logger
- CWD (current working directory)
- ProjectData (plugin-specific data)
- Output (io.Writer)
- Input (io.Reader)

**Arguments**: Parsed by command
- Flag parsing
- Validation
- Error handling

---

## Event Types

### Plugin-Specific Events

All event types are **plugin-specific** (not in framework):

**Chat Events**:
- `ChatStarted` - New chat session
- `ChatMessageUser` - User message
- `ChatMessageAssistant` - Assistant response

**Tool Events**:
- `ToolInvoked` - Tool called
- `ToolResult` - Tool completed

**File Events**:
- `FileRead` - File read operation
- `FileWritten` - File write operation

**Context Events**:
- `ContextChanged` - Working context changed
- `Error` - Error occurred

### Payload Mapping

Event type → Payload type:
- `ChatStarted` → `ChatPayload`
- `ToolInvoked` → `ToolPayload`
- `FileRead` / `FileWritten` → `FilePayload`
- `ContextChanged` → `ContextPayload`
- `Error` → `ErrorPayload`

---

## Hook Integration

### Hook Flow

```
Claude Code Hook Triggered
    ↓
Hook script: ./dw claude auto-summary-exec
    ↓
AutoSummaryExecCommand.Execute()
    ↓
Parse hook input (HookInputParser)
    ↓
Convert to Event (HookInputToEvent)
    ↓
Save to repository
```

### Hook Configuration

**Location**: `.claude/settings.json`

**Structure**:
```json
{
  "hooks": {
    "user-prompt-submit": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "bash",
            "command": "./dw claude auto-summary-exec",
            "timeout": 5000
          }
        ]
      }
    ]
  }
}
```

### Installation

```go
manager, _ := NewHookConfigManager()
err := manager.InstallDarwinFlowHooks()
```

**Actions**:
1. Read existing `.claude/settings.json`
2. Parse current hooks
3. Merge with DarwinFlow hooks
4. Write back to file

---

## Analysis Implementation

### Session Analysis

**SessionAnalysis** is plugin-specific (not framework):
- Stores Claude Code session analysis
- Links to session via SessionID
- Supports multiple prompts per session
- Tracks model, prompt name, analysis type

### Analysis Types

- `"default"` - Default analysis
- `"tool-analysis"` - Tool usage analysis
- `"summary"` - Session summary
- Custom prompt names

### Storage

Via `AnalysisRepository` interface:
- `SaveAnalysis()` - Persist analysis
- `GetAnalysisBySessionID()` - Retrieve by session
- `GetAllAnalyses()` - List all analyses

---

## Testing Strategy

**Plugin testing**:
- Test plugin implementation
- Mock services (AnalysisService, LogsService)
- Test entity capabilities
- Test command execution

**Command testing**:
- Test argument parsing
- Test execution logic
- Mock services
- Verify output

**Hook testing**:
- Test hook parsing
- Test hook installation
- Test event conversion
- Mock file system

**Integration testing**:
- Real services
- Real database
- Full command flow

---

## External Plugin Template

This package serves as reference for external plugins:

1. **Import SDK**: `import "darwinflow/pkg/pluginsdk"`
2. **Implement Plugin**: Satisfy `pluginsdk.Plugin`
3. **Define entities**: Implement capability interfaces
4. **Define commands**: Implement command interface
5. **Define events**: Plugin-specific event types
6. **Define payloads**: Plugin-specific payload types

**Key difference**: External plugins live outside repo, import only `pkg/pluginsdk`

---

## Files

- `plugin.go` - ClaudeCodePlugin implementation
- `session_entity.go` - SessionEntity implementation
- `commands.go` - Command implementations
- `event_types.go` - Event type constants
- `payloads.go` - Payload type definitions
- `analysis.go` - SessionAnalysis type
- `analysis_repository.go` - Repository interface
- `hook_manager.go` - Hook installation and config
- `hook_input_parser.go` - Hook input parsing
- `interfaces.go` - Plugin-specific interfaces
- `*_test.go` - Plugin tests

---

*Generated by `go-arch-lint -format=package pkg/plugins/claude_code`*

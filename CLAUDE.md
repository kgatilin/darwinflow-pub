# DarwinFlow - Claude Code Logging System

## Project Overview

**DarwinFlow** is a lightweight logging system that captures Claude Code interactions as structured events using event sourcing principles. The system stores events in SQLite and enables future pattern detection and workflow optimization.

### Key Components

- **CLI Tool (`dw`)**: Main entry point with multiple subcommands
  - `dw claude-code init` (or `dw claude init`) - Initialize logging infrastructure
  - `dw refresh` - Update database schema and hooks to latest version (run after upgrading)
  - `dw claude-code log` (or `dw claude log`) - Log events (called by hooks)
  - `dw logs` - View and query logged events
  - `dw ui` - Interactive terminal UI for browsing and analyzing sessions
    - Session list view with analysis status indicators
    - Session detail view with analysis previews
    - Quick actions: analyze, re-analyze, save to markdown, view full analysis
    - Keyboard-driven navigation (arrow keys, Enter, shortcuts)
  - `dw analyze` - AI-powered session analysis with configurable prompts
    - `--last` - Analyze most recent session
    - `--session-id <id>` - Analyze specific session
    - `--all` - Analyze all unanalyzed sessions
    - `--refresh` - Re-analyze already analyzed sessions
    - `--limit N` - Limit refresh/analyze to latest N sessions
    - `--prompt <name>` - Use specific prompt from config (e.g., tool_analysis, session_summary)
    - `--model <model>` - Override model from config
    - `--token-limit <num>` - Override token limit from config
    - `--view` - View existing analysis without re-analyzing
- **Event Logging**: Captures tool invocations and user prompts via Claude Code hooks
- **SQLite Storage**: Fast, file-based event storage with full-text search capability
- **Database Migration**: Safe schema migrations with automatic duplicate cleanup and version upgrades
  - Multi-step migration process: base tables → column additions → duplicate cleanup → indexes
  - Handles missing columns, duplicates, and schema inconsistencies
  - Backwards compatible with existing databases
- **Hook Management**: Claude Code plugin manages its own hooks (`.claude/settings.json`)
- **Log Viewer**: Query interface with SQL support for exploring captured events
- **AI Analysis**: Uses Claude CLI to analyze sessions and suggest workflow optimizations

### Plugin System Architecture

**DarwinFlow uses a capability-driven plugin system** with a public SDK for extensibility.

**Core Concepts**:
- **SDK Contract** (`pkg/pluginsdk`): Public API that all plugins (internal and external) import
- **Plugin Capabilities**: Interfaces defining what plugins can DO (IEntityProvider, ICommandProvider, IEventEmitter)
- **Entity Capabilities**: Interfaces defining what entities ARE (IExtensible, IHasContext, ITrackable, etc.)
- **Plugin Registry**: Central manager that routes queries to appropriate plugins based on capabilities
- **Adaptation Layer**: Converts between SDK types and internal domain types

**Plugin SDK** (`pkg/pluginsdk/`):
- Public API accessible to external Go plugins
- Zero dependencies on internal packages
- Defines all plugin and entity interfaces
- Self-contained types (Event, EntityContext, PluginInfo, etc.)

**Plugin Capability Interfaces** (defined in `pkg/pluginsdk/capability.go`):
- `IEntityProvider` - Plugins that provide queryable entities
- `IEntityUpdater` - Plugins that support entity updates
- `ICommandProvider` - Plugins that provide CLI commands
- `IEventEmitter` - Plugins that emit real-time events (planned)

**Entity Capability Interfaces** (defined in `pkg/pluginsdk/entity.go`):
- `IExtensible` - **Required**: Base capability (ID, type, fields) - all entities must implement
- `IHasContext` - **Optional**: Entities with related data (files, activity, metadata)
- `ITrackable` - **Optional**: Status and progress tracking
- `ISchedulable` - **Optional**: Time-based scheduling (planned)
- `IRelatable` - **Optional**: Explicit entity relationships (planned)

**Plugin Types**:
- **Internal plugins** (`pkg/plugins/`): Built-in plugins that ship with the tool
  - Import `pkg/pluginsdk` for interfaces
  - Receive adapted internal services via cmd layer
  - Example: `pkg/plugins/claude_code`
- **External plugins** (planned): User-created plugins
  - Import `pkg/pluginsdk` (public API)
  - No access to `internal/*` packages
  - Communicate via SDK contracts only

**How It Works**:
1. Plugins import `pkg/pluginsdk` and implement required interfaces
2. Plugin declares capabilities via `GetCapabilities()` method
3. CLI routes commands to PluginRegistry
4. Registry uses capability-based routing to find correct plugin
5. Adaptation layer converts SDK types ↔ domain types
6. TUI/commands work with entities through SDK interfaces
7. All plugins use same SDK contract (internal and external)

**Example Plugin**:
```go
package myplugin

import "github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"

type MyPlugin struct {
    logger pluginsdk.Logger
    // ... other dependencies
}

func (p *MyPlugin) GetInfo() pluginsdk.PluginInfo {
    return pluginsdk.PluginInfo{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "My custom plugin",
    }
}

func (p *MyPlugin) GetCapabilities() []string {
    return []string{"IEntityProvider", "ICommandProvider"}
}

// Implement capability interfaces...
```

**Benefits**:
- External Go plugins can import public SDK
- Internal plugins use same contracts
- Type-safe capability detection
- Clean separation: plugins know only SDK, not internal implementation
- UI automatically supports new entity types with known capabilities

**Command System**:
Plugins provide CLI commands via the `ICommandProvider` capability (defined in `internal/domain/plugin.go`):
- **Command Interface**: Defines command name, description, usage, and execution
- **CommandRegistry** (`internal/app/command_registry.go`): Discovers and executes commands from registered plugins
- **Built-in Commands** (from claude-code plugin):
  - `dw claude-code init` (or `dw claude init`): Initialize Claude Code logging infrastructure
  - `dw claude-code log <event-type>` (or `dw claude log`): Log a Claude Code event
- **Usage**: `dw <plugin-name> <command> [args]` - Plugin commands under plugin namespace
- **Backward Compatibility**: `dw claude` maps to `dw claude-code` for historical compatibility

**Tool System**:
Plugins can also provide project-specific tools via the `IToolProvider` capability:
- **Tool Interface** (`internal/domain/plugin.go`): Defines tool name, description, usage, and execution
- **ToolRegistry** (`internal/app/tool_registry.go`): Discovers and executes tools from registered plugins
- **ProjectContext**: Tools receive access to event repo, analysis repo, config, and working directory
- **Built-in Tools** (from claude-code plugin):
  - `dw project session-summary --last`: Display summary of most recent Claude Code session
  - `dw project session-summary --session-id <id>`: Display summary of specific session
- **Usage**: `dw project <toolname> [args]` - Clear namespace separation from plugin commands

**How It Works**:
1. Plugin implements `ICommandProvider` (for CLI commands) and/or `IToolProvider` (for project tools)
2. Registries discover commands/tools from all registered plugins
3. `dw <plugin-name> <command>` routes to plugin commands (e.g., init, log)
4. `dw project <toolname>` routes to project tools (e.g., session-summary)
5. Commands/tools execute with appropriate context (I/O, repos, config, etc.)

**SDK as Single Source of Truth**:

The plugin SDK (`pkg/pluginsdk`) is the **single source of truth** for all plugin contracts:

**Architecture**:
- `pkg/pluginsdk/` defines all interfaces: Plugin capabilities (IEntityProvider, ICommandProvider) and Entity capabilities (IExtensible, ITrackable, IHasContext)
- Plugins import `pkg/pluginsdk` exclusively - no `internal/*` imports allowed
- Internal domain layer may import SDK types when needed (one-way dependency: domain → SDK)
- Framework layers (`internal/app`, `cmd`) work with SDK types directly

**Why SDK is public**:
- External plugins need stable, versioned contracts
- Single source of truth eliminates interface duplication
- Clean separation: SDK defines "what", internal implementation defines "how"
- Minimal adaptation needed - framework and plugins speak same language

**Current State**:
- ✅ SDK in `pkg/pluginsdk/` (public, self-contained, single source of truth)
- ✅ Claude-code plugin uses SDK exclusively
- ✅ Capability-based routing in PluginRegistry
- ✅ Internal plugins ready (pkg/plugins/)
- ✅ TUI using plugin system for entity queries
- ✅ Command system with plugin-namespaced CLI commands
- ✅ Tool system with project-scoped tools
- ✅ Clean bounded context separation (no plugin logic in framework)
- ✅ Plugin commands are fully self-contained
- 🔄 **Refactoring in progress**: Interface deduplication (Phases 1-7)
  - Remove duplicate interfaces from `internal/domain`
  - Eliminate adaptation layer boilerplate
  - Move plugin-specific types to plugin package
- 🔄 External plugin discovery (planned)
- 🔄 JSON-RPC for non-Go plugins (planned)

**Bounded Context Separation**:

The architecture maintains strict separation between framework and plugins:

- **Framework layers** (`internal/app`, `internal/domain`, `internal/infra`): Zero knowledge of specific plugins
- **Plugin packages** (`pkg/plugins/claude_code`): Self-contained with all plugin-specific logic
- **cmd layer**: Only 2 places reference plugins:
  1. `cmd/dw/main.go` - Backward compatibility mapping (`dw claude` → `dw claude-code`)
  2. `cmd/dw/plugin_registration.go` - Plugin registration (composition root)

Plugin-specific features (like hook management for claude-code) are handled entirely within the plugin package. The framework provides only generic services through well-defined interfaces.

### Architecture Documentation

For detailed architecture and API information, see:
- @docs/arch-index.md - Architecture summary and package details
- `pkg/pluginsdk/` - Public plugin SDK (godoc for API reference)
- Run `go-arch-lint .` to validate architecture compliance

**Key architectural principles**:
1. **SDK is single source of truth**: `pkg/pluginsdk` defines all plugin and entity contracts
2. **Plugins import SDK only**: Both internal and external plugins use pkg/pluginsdk exclusively
3. **Domain may import SDK**: Framework domain can use SDK types (one-way: domain → SDK)
4. **No duplicate interfaces**: If SDK has it, domain doesn't duplicate it
5. **Minimal adaptation**: Framework works with SDK types directly when possible
6. **DDD compliance**: Maintained with go-arch-lint validation

### Current Implementation Status

**Active Hooks**:
- `PreToolUse`: Logs all tool invocations (Read, Write, Bash, etc.)
- `UserPromptSubmit`: Logs user message submissions

**Event Types**: Defined in `internal/domain/event.go`
- `tool.invoked`, `tool.result`
- `chat.message.user`, `chat.message.assistant`
- `chat.started`, `file.read`, `file.written`, etc.

**Phase 3: Plugin SDK Event Integration** ✅
Hooks now emit events through the plugin SDK rather than direct command-line logging:

- **Event Emission Flow**:
  - Hook scripts → JSON event on stdin
  - `dw claude emit-event` command → parses JSON
  - EmitEventCommand.Execute() → calls PluginContext.EmitEvent()
  - PluginContext.EmitEvent() → converts SDK Event to domain Event
  - EventRepository.Save() → stores in SQLite

- **Hook Command**: `dw claude emit-event` accepts JSON event on stdin
  - Replaces deprecated `dw claude log` (still supported for backward compatibility)
  - Type-safe via pluginsdk.Event struct with validation
  - Graceful error handling: invalid events silently ignored (non-blocking)
  - 5-second timeout for hook execution to prevent blocking Claude Code

- **Event Schema Versioning**:
  - Events have optional `Version` field (default: "1.0")
  - Supports schema evolution without breaking backward compatibility
  - Current version stores events with version metadata for future migration support

- **Event Sourcing Pattern**:
  - All events stored sequentially in SQLite with timestamps
  - Full audit trail of Claude Code interactions
  - Enable analytics, pattern detection, and workflow optimization

- **Migration Path**:
  - Old `dw claude log` command still works (deprecated)
  - `dw refresh` updates hooks to new `emit-event` format automatically
  - Both commands coexist during transition period for smooth upgrades

- **Backward Compatibility**:
  - Existing hooks with `dw claude log` continue to work
  - EventRepository queries work identically regardless of command source
  - No breaking changes to event storage or schema

**Analysis Features**:
- **Multi-Prompt Analysis System**: Support for multiple analysis types with different prompts
  - **Session Summary** (`session_summary`): Auto-triggered summaries capturing user intent, outcomes, and session quality
  - **Tool Analysis** (`tool_analysis`): Agent-focused analysis identifying needed tools and workflow improvements
  - Custom prompts can be added to `.darwinflow.yaml` config
- **Auto-Triggered Session Summaries**: Optional automatic analysis on session end
  - Controlled via `analysis.auto_summary_enabled` (default: false)
  - **Fully asynchronous**: Hook returns immediately (~50ms), analysis runs in detached background process
  - Zero blocking - Claude Code session ends without waiting for analysis
  - Uses configurable prompt via `analysis.auto_summary_prompt`
  - Requires `dw claude init` to install SessionEnd hook
- **Configuration-Based Execution**: Analysis settings in `.darwinflow.yaml`
  - `analysis.token_limit`: Max tokens for analysis context (default: 100000)
  - `analysis.model`: Claude model to use - alias (sonnet, opus, haiku) or full name (default: sonnet)
    - **Allowed models**: sonnet, opus, haiku (latest versions), or specific versions
    - Whitelist includes: claude-sonnet-4-5-20250929, claude-opus-4-20250514, etc.
    - Invalid models fall back to default with warning
  - `analysis.parallel_limit`: Max parallel analysis executions (default: 3)
  - `analysis.enabled_prompts`: Array of prompts to run during analysis (default: ["tool_analysis"])
    - Prompts run in parallel when multiple are specified
    - Must reference prompts defined in `prompts` section
    - `--prompt` CLI flag overrides this setting
    - Invalid prompts are filtered out with warning
  - `analysis.auto_summary_enabled`: Enable auto session summaries (default: false)
  - `analysis.auto_summary_prompt`: Prompt for auto summaries (default: session_summary)
  - `analysis.claude_options.allowed_tools`: Tools available during analysis (default: empty = no tools)
  - `analysis.claude_options.system_prompt_mode`: "replace" or "append" (default: replace)
- **Smart Session Selection**: Token-aware batch analysis
  - `EstimateTokenCount`: Uses chars/4 heuristic to estimate session size
  - `SelectSessionsWithinTokenLimit`: Automatically selects sessions that fit within token budget
  - Reserves 20% of tokens for prompt overhead and response
- **Parallel Execution**: Concurrent analysis with semaphore-based concurrency control
  - `AnalyzeMultipleSessionsParallel`: Analyze multiple sessions in parallel
  - `AnalyzeSessionWithMultiplePrompts`: Run multiple prompts on one session in parallel
  - Respects `analysis.parallel_limit` from config
- **CLI Flag Overrides**: All config settings can be overridden via flags
- **Clean Execution**: Uses `claude --system-prompt` for pure analysis without tool invocations
- **Refresh Capability**: Re-analyze sessions with updated prompts using `--refresh --limit N`
- **Multiple Analyses Per Session**: Each session can have multiple analyses (one per prompt type)
- Persistent storage in `session_analyses` table with `analysis_type` and `prompt_name` fields

**Interactive UI Features**:
- **Terminal UI**: Built with Bubble Tea framework for responsive, keyboard-driven interaction
- **Session Management**: Browse all sessions with analysis status indicators (✓ analyzed, ✗ not analyzed, ⟳N multiple analyses)
- **Session Details**: View metadata, event counts, and analysis previews in dedicated detail view
- **Analysis Actions**: Quick actions to analyze, re-analyze, view full analysis, or save to markdown
- **Markdown Export**: Save analyses to configurable directory with customizable filename templates
- **UI Configuration** (`.darwinflow.yaml`):
  - `ui.default_output_dir`: Directory for saved markdown files (default: "./analysis-outputs")
  - `ui.filename_template`: Template for filenames with placeholders {{.SessionID}}, {{.PromptName}}, {{.Date}}, {{.Time}}
  - `ui.auto_refresh_interval`: Auto-refresh interval (e.g., "30s", empty = disabled)

### Development Workflow

**Note**: When the user refers to "workflow", they mean these CLAUDE.md instructions.

**Handling Large Tasks with Task Tool Delegation**:

For large features or refactorings that require substantial context, use the Task tool to ensure complete implementation:

1. **Decompose First**: Break the large task into smaller, context-sized chunks
   - Each chunk should be independently implementable
   - Aim for chunks that fit comfortably within context limits
   - Consider natural boundaries (packages, features, components)

2. **Sequential Delegation Pattern**:
   - Launch Task tool for the first chunk with detailed instructions
   - Wait for sub-agent completion report
   - Review the report and results
   - Launch next Task tool for the next chunk
   - Continue until the entire feature is fully implemented

3. **Benefits**:
   - Ensures complete implementation without hitting context limits
   - Each chunk gets full attention and thorough implementation
   - Natural checkpoints for review and validation
   - Avoids incomplete features due to context exhaustion

**Example**: Adding a complex plugin system:
- Chunk 1: Define domain interfaces and capability contracts
- Chunk 2: Implement plugin registry and discovery
- Chunk 3: Create core plugin implementation
- Chunk 4: Add CLI commands and integrate with existing code
- Chunk 5: Write comprehensive tests for all components

**Final Checklist for Large Changes**:

After completing implementation work (especially when using Task tool delegation), **ALWAYS** create a final checklist using TodoWrite with these items:

1. **Run tests**: `go test ./...` - ensure all tests pass
2. **Run linter**: `go-arch-lint .` - verify zero violations
3. **Update README.md**: If functionality changed (commands, features, behavior)
4. **Update CLAUDE.md**: If workflow or architecture changed
5. **Regenerate docs**: Run `go-arch-lint docs` if architecture/API changed
6. **Commit changes**: Create commit with concise, informative message

**Why this matters**:
- Ensures nothing is forgotten after complex implementations
- Provides clear completion criteria for the task
- Makes the workflow systematic and repeatable
- Gives visibility to the user on final steps

**When to create this checklist**:
- After completing Task tool delegation work
- After finishing a multi-step refactoring or feature
- Before marking a large task as complete
- Any time substantial code changes were made

Use TodoWrite to track each item and mark them complete as you go.

When working on this project:
1. Understand the DDD layered architecture (see below)
2. Check @docs/arch-index.md to see current packages and dependencies
3. Follow the architecture guidelines strictly
4. Write tests for new functionality (aim for 70-80% coverage)
5. **Update documentation** (README.md and CLAUDE.md) when adding features
6. Run tests and linter before committing
7. Regenerate architecture docs if needed (`go-arch-lint docs`)
8. **Commit after each iteration** - After completing each logical task/iteration, commit all changes with a concise, informative commit message (e.g., "add session refresh feature" rather than long explanations)

**Working with the Plugin System**:
- **SDK interfaces** are defined in `pkg/pluginsdk/` (Plugin, IEntityProvider, ICommandProvider, IToolProvider)
- **Entity capability interfaces** are defined in `pkg/pluginsdk/entity.go` (IExtensible, ITrackable, IHasContext, etc.)
- **Internal plugins** live in `pkg/plugins/` (e.g., `pkg/plugins/claude_code/`)
- Plugins import `pkg/pluginsdk` and implement SDK interfaces
- **Adaptation layer** in `internal/app/` and `cmd/dw/` converts SDK ↔ domain types
- Plugin tests should verify SDK interface compliance and capability support
- TUI should render based on entity capabilities, not entity types
- **Entity capabilities**:
  - IExtensible is **required** - all entities must implement
  - IHasContext, ITrackable, ISchedulable, IRelatable are **optional**
  - Entities declare capabilities via `GetCapabilities()` method
  - TUI only uses capabilities that are declared
- Commands are plugin-scoped: `dw <plugin-name> <command>`
- Tools are project-scoped: `dw project <toolname>`

---

# go-arch-lint - Architecture Linting

**CRITICAL**: The .goarchlint configuration is IMMUTABLE - AI agents must NOT modify it.

## Architecture (DDD Layered Architecture)

This project uses **Domain-Driven Design (DDD)** with strict dependency rules:

```
         ┌─────────────┐
         │     cmd     │  Entry points
         └──────┬──────┘
                │
        ┌───────┴────────┐
        │                │
        ▼                ▼
┌──────────────┐  ┌──────────────┐
│ internal/app │  │internal/infra│  Application & Infrastructure layers
└──────┬───────┘  └──────┬───────┘
       │                 │
       └────────┬────────┘
                ▼
        ┌──────────────┐
        │internal/domain│  Pure business logic (NO dependencies)
        └──────────────┘
```

**Layer Responsibilities**:
- **cmd**: Application entry points → can import `internal/app`, `internal/infra`, `pkg/pluginsdk`
- **internal/app**: Use cases, orchestration → can import `internal/domain`, `pkg/pluginsdk`
- **internal/domain**: Framework business logic → can import `pkg/pluginsdk` for contracts
- **internal/infra**: Infrastructure (DB, APIs) → can import `internal/domain`, `pkg/pluginsdk`
- **pkg/pluginsdk**: Public plugin contracts → imports NOTHING internal

**Key Rules**:
- SDK has zero internal dependencies (fully public)
- Domain may import SDK for contracts (SDK is the authority on plugin interfaces)
- All dependencies still flow toward domain (domain is still the core)
- Plugin SDK is an exception: it's public API, not an external dependency

## DDD Core Principles

1. **Domain Isolation**: `internal/domain` contains framework business logic with minimal dependencies
2. **SDK First**: Public plugin contracts live in `pkg/pluginsdk` (single source of truth)
3. **Dependency Inversion**: Domain/SDK define interfaces; infrastructure implements them
4. **Structural Typing**: Go's implicit interfaces - types satisfy interfaces via matching methods
5. **One-Way Dependencies**: Domain ← App/Infra ← Cmd (SDK is public, not a dependency violation)

## Documentation Generation (Run Regularly)

Keep documentation synchronized with code changes:

```bash
# Generate comprehensive architecture documentation
go-arch-lint docs
```

This generates `docs/arch-generated.md` with:
- Project structure and architectural rules
- Complete dependency graph with method-level details
- Public API documentation
- Statistics and validation status

**When to regenerate**:
- After adding/removing packages or files
- After changing public APIs (exported functions, types, methods)
- After modifying package dependencies
- Before committing architectural changes
- Run regularly during development to track changes

## Before Every Commit

1. `go test ./...` - all tests must pass
2. `go-arch-lint .` - ZERO violations required (non-negotiable)
3. Regenerate docs if architecture/API changed (see above)
4. Update README.md and CLAUDE.md if functionality changed

## When Linter Reports Violations

**Do NOT mechanically fix imports.** Violations reveal architectural issues. Process:
1. **Reflect**: Why does this violation exist? What dependency is wrong?
2. **Plan**: Which layer should own this logic? What's the right structure?
3. **Refactor**: Move code to correct layer following DDD principles
4. **Verify**: Run `go-arch-lint .` - confirm zero violations

**Common Violation Scenarios**:

- `internal/domain` imports `internal/app` or `internal/infra` → **VIOLATION**
  - Fix: Define interface in domain or SDK, implement in infra, inject via app layer

- `internal/app` imports `internal/infra` → **VIOLATION**
  - Fix: Use dependency injection through interfaces defined in domain or SDK

- `pkg/pluginsdk` imports `internal/*` → **VIOLATION**
  - Fix: SDK must be fully public with zero internal dependencies

- `internal/domain` imports `pkg/pluginsdk` → **OK** (SDK is public contracts)
- `internal/infra` needs to call domain logic → **OK** (infra → domain allowed)

Example: Domain needs database access:
- ❌ `internal/domain/user.go` imports `internal/infra/sqlite_repository.go`
- ✅ `internal/domain/user.go` or `pkg/pluginsdk/repository.go` defines `UserRepository` interface
- ✅ `internal/infra/sqlite_repository.go` implements `UserRepository`
- ✅ `internal/app/user_service.go` receives injected repository

Example: Plugin contracts:
- ✅ `pkg/pluginsdk/capability.go` defines `IEntityProvider` interface
- ✅ `pkg/plugins/claude_code/plugin.go` implements `IEntityProvider`
- ✅ `internal/app/plugin_registry.go` works with `pluginsdk.IEntityProvider`
- ❌ Duplicate interface in both `pkg/pluginsdk` and `internal/domain`

## Code Guidelines

**DO**:
- Keep framework business logic in `internal/domain`
- Define public plugin contracts in `pkg/pluginsdk` (single source of truth)
- Define framework-internal interfaces in `internal/domain`
- Implement infrastructure in `internal/infra`
- Orchestrate use cases in `internal/app`
- Use dependency injection to wire layers together
- Use black-box tests (`package pkgname_test`) to test through public APIs
- Import `pkg/pluginsdk` when working with plugin contracts

**DON'T**:
- Duplicate interfaces between `pkg/pluginsdk` and `internal/domain`
- Import `internal/*` from `pkg/pluginsdk` (SDK must be fully public)
- Import upward (domain importing app/infra is still a violation)
- Import between internal packages at the same level
- Put business logic in cmd, app, or infra (belongs in domain or plugins)
- Modify .goarchlint (immutable architectural contract)

**Layer-Specific Rules**:
- **cmd**: Keep thin - just CLI parsing and wiring
- **internal/app**: Orchestration only - coordinate domain objects
- **internal/domain**: Pure functions, value objects, domain services
- **internal/infra**: Technical implementations - DB, external APIs, file I/O

Run `go-arch-lint .` frequently during development. Zero violations required.

---

# Testing Conventions

## Test Organization

**Coverage Target**: 70-80% for all packages

**Package Naming**:
- Use black-box testing: `package pkgname_test` (not `package pkgname`)
- This enforces testing only the public API, ensuring good API design

**File Naming**:
- Test files: `*_test.go` in same directory as code under test
- Example: `internal/infra/sqlite_store.go` → `internal/infra/sqlite_store_test.go`

## Test Function Naming

**Format**: `TestFunctionName` or `TestType_Method`

Examples:
- `TestNewLogger` - tests the NewLogger function
- `TestSQLiteStore_Init` - tests the Init method on SQLiteStore type
- `TestDetectContext_FromEnv` - tests DetectContext with specific scenario

## Test Structure

**Setup and Cleanup**:
```go
func TestExample(t *testing.T) {
    // Use t.TempDir() for temporary directories/files
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    // Setup code...
    resource, err := setupFunction()
    if err != nil {
        t.Fatalf("Setup failed: %v", err)
    }
    defer resource.Close()  // Use defer for cleanup

    // Test logic...
}
```

**Error Handling**:
- `t.Fatalf(format, args...)` - Fatal errors that prevent test from continuing (setup failures)
- `t.Errorf(format, args...)` - Assertion failures (test should continue to report all failures)

**Table-Driven Tests**:
Use for multiple test cases with same logic:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
    }{
        {name: "case1", input: "a", want: "A"},
        {name: "case2", input: "b", want: "B"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Function(tt.input)
            if got != tt.want {
                t.Errorf("Function() = %q, want %q", got, tt.want)
            }
        })
    }
}
```

**Assertions**:
- Use simple if-checks (no external assertion libraries)
- Provide descriptive error messages with actual vs expected values
- Format: `t.Errorf("Expected X, got Y")` or `t.Errorf("Function() = %v, want %v", got, want)`

## Running Tests

**Run all tests**:
```bash
go test ./...
```

**Run with coverage**:
```bash
go test -cover ./...
```

**Generate coverage report**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View in browser
```

**Run specific test**:
```bash
go test -run TestFunctionName ./internal/infra
```

## Testing Best Practices

1. **Isolation**: Each test should be independent and not rely on other tests
2. **Temp Resources**: Always use `t.TempDir()` for file/directory operations
3. **Cleanup**: Use `defer` for cleanup to ensure resources are released even on failure
4. **Fast Tests**: Tests should run quickly (< 1s per test typically)
5. **Clear Names**: Test names should describe what they're testing
6. **Test Public API**: Focus on testing exported functions/methods (black-box testing)
7. **Edge Cases**: Test boundary conditions, empty inputs, nil values, errors

---

# Documentation Workflow

**CRITICAL**: Documentation must be updated whenever functionality changes. This is not optional.

## When to Update Documentation

Update documentation when you:
1. Add new commands or subcommands
2. Add new flags or options
3. Change public APIs or behavior
4. Add new features or capabilities
5. Modify architecture or package structure

## What to Update

### README.md
Update the user-facing README when:
- Adding commands: Update **Commands** section with usage examples
- Adding features: Update **Features** list
- Changing structure: Update **Project Structure** section
- Completing roadmap items: Move items from Planned to Current in **Roadmap**

### CLAUDE.md
Update the development documentation when:
- Adding functionality: Update **Key Components** section
- Changing workflow: Update **Development Workflow** section
- Adding architectural patterns: Document in relevant sections
- Changing test conventions: Update **Testing Conventions** section

### Generated Documentation
Regenerate when architecture or API changes:
```bash
# After modifying package dependencies or exports
go-arch-lint docs
```

## Documentation Checklist

When adding a feature, follow this checklist:

- [ ] Code is implemented and tested
- [ ] README.md updated with user-facing changes
- [ ] CLAUDE.md updated with development notes
- [ ] Architecture docs regenerated if needed
- [ ] Examples added to demonstrate usage
- [ ] All tests pass
- [ ] Architecture linter passes

**Example**: Adding `dw logs` command required:
- ✅ README.md: Added to Commands, Log Viewing Examples, Features, Project Structure, Roadmap
- ✅ CLAUDE.md: Updated Key Components, Development Workflow
- ✅ Test coverage: 88% (exceeds 70-80% target)
- ✅ Documentation: Comprehensive testing conventions added

# DarwinFlow

**Capture, store, and analyze your Claude Code interactions**

DarwinFlow is a lightweight logging system that automatically captures all Claude Code interactions as structured events. Built with event sourcing principles, it enables pattern detection, workflow optimization, and deep insights into your AI-assisted development sessions.

## Features

- **Automatic Logging**: Captures all Claude Code events via hooks (tool invocations, user prompts, etc.)
- **AI-Powered Analysis**: Plugin-agnostic analysis framework using view-based architecture
  - **View-Based Pattern**: Plugins provide views of their events, framework analyzes any view type
  - **Multi-Prompt Support**: Session summaries, tool analysis, and custom prompts
  - **Auto-Triggered Summaries**: Optional automatic analysis on session end (configurable)
  - **Parallel Execution**: Concurrent analysis with semaphore-based concurrency control
  - **Token-Aware Selection**: Smart session selection based on token limits
  - **LLM Abstraction**: Swappable LLM providers (Claude CLI, Anthropic API, OpenAI, etc.)
- **Interactive TUI**: Browse sessions, view analyses, and manage workflows with a keyboard-driven interface
- **Plugin System**: Extensible architecture with public SDK for building plugins
  - **Public SDK**: `pkg/pluginsdk` provides contracts for plugin development
  - **Bounded Context**: Framework is plugin-agnostic, zero knowledge of specific plugins
  - **Self-Contained Plugins**: All plugin logic isolated in plugin packages
  - **Capability-Driven**: Entities implement capabilities (IExtensible, ITrackable, IHasContext, etc.)
  - **Core Plugins**: Claude Code sessions built-in, extensible for custom entity types
  - **Plugin Event Bus**: Cross-plugin communication via publish/subscribe pattern
- **Log Viewer**: Query and explore captured events with `dw logs` command
- **Event Sourcing**: Immutable event log enabling replay and analysis
- **SQLite Storage**: Fast, file-based storage with full-text search
- **Zero Performance Impact**: Non-blocking, concurrent-safe logging
- **Context-Aware**: Automatically detects project context from environment
- **Clean Architecture**: Strict 3-layer DDD design enforced by go-arch-lint

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/kgatilin/darwinflow-pub.git
cd darwinflow-pub

# Build the CLI
go build -o dw ./cmd/dw

# Install to your PATH (optional)
go install ./cmd/dw
```

### Initialize Logging

```bash
# Set up Claude Code logging infrastructure
dw claude-code init    # Or 'dw claude init' for backward compatibility
```

This will:
- Create the SQLite database at `.darwinflow/logs/events.db` (project-relative)
- Configure Claude Code hooks in `.claude/settings.json` (plugin-managed)
- Enable automatic event capture via PreToolUse, UserPromptSubmit, and SessionEnd hooks

### Start Using Claude Code

After running `dw claude init`, restart Claude Code. All your interactions will now be automatically logged!

### Upgrading DarwinFlow

When you upgrade to a new version of DarwinFlow (e.g., after `git pull`), run the refresh command to update your installation:

```bash
# Rebuild the binary
go build -o dw ./cmd/dw

# Update database schema and hooks to latest version
dw refresh
```

The `dw refresh` command:
- Updates the database schema with new columns, indexes, and tables
- Migrates existing data safely (e.g., removes duplicates, adds default values)
- Reinstalls/updates hooks to the latest version
- Verifies configuration integrity

**When to use `dw refresh`:**
- After updating DarwinFlow to a new version
- When you see database schema errors (e.g., "no such column")
- When new hooks are added to DarwinFlow
- To fix database inconsistencies

## Architecture

DarwinFlow follows a strict Domain-Driven Design (DDD) architecture enforced by [go-arch-lint](https://github.com/fdaines/go-arch-lint):

```
cmd → internal/app + internal/infra → internal/domain
```

**Layers:**
- **cmd/dw**: CLI entry points and command handlers
- **internal/app**: Application services, use cases, and plugins
- **internal/infra**: Infrastructure (database, file I/O, external APIs)
- **internal/domain**: Pure business logic (entities, capabilities, interfaces)

**Dependency Rule**: Dependencies flow inward only. Domain has zero dependencies.

### Plugin System

**Architecture:**
- **Public SDK** (`pkg/pluginsdk`): Self-contained plugin API with zero internal dependencies
- **Plugins** (`pkg/plugins/`): Isolated packages implementing SDK interfaces
- **Bounded Context**: Framework layers have zero knowledge of specific plugins
- **Adaptation Layer**: cmd/app layers convert SDK ↔ domain types at boundaries

**Core Concepts:**
- **Plugin Capabilities**: IEntityProvider, ICommandProvider, IEventEmitter (defined in SDK)
- **Entity Capabilities**: IExtensible (required), ITrackable, IHasContext (optional)
- **Plugin Registry**: Routes queries to appropriate plugins based on capabilities
- **Command Registry**: Discovers and executes commands from registered plugins
- **Self-Contained Commands**: Plugins manage all their own logic (e.g., hooks, config)

**Current Plugins:**
- **claude-code** (`pkg/plugins/claude_code`): Claude Code session tracking
  - Entity: `session` (IExtensible + IHasContext + ITrackable)
  - Commands: `init`, `emit-event` (for hook integration)
  - CLI: `dw claude-code <command>` or `dw claude <command>` (backward compat)
- **task-manager** (`pkg/plugins/task_manager`): Hierarchical roadmap management
  - Entities: `roadmap`, `track`, `task`, `iteration` (all IExtensible)
  - Capabilities: Event sourcing, SQLite persistence, interactive TUI
  - Commands: Full CRUD operations on all entities with CLI and TUI
  - CLI: `dw task-manager <command>` or `dw task-manager tui` for interactive mode

### Framework Capabilities for Plugin Developers

**Key Principle**: Framework handles infrastructure and cross-plugin concerns. Plugins handle domain logic.

**What the Framework Provides:**
- **Event Storage** (`EventRepository`): Persistent SQLite storage, queries, full-text search, versioning
- **Centralized Analysis**: AI-powered session analysis across ALL plugin events with configurable prompts
- **Cross-Plugin Communication** (`EventBus`): Publish/subscribe real-time communication with filtering
- **Plugin Lifecycle**: Registration, initialization, context injection, shutdown management
- **Configuration Management**: YAML config loading/saving, validation
- **Logging Infrastructure**: Leveled logging with consistent formatting
- **Command Registry**: Automatic command discovery, routing, help generation
- **Entity Management**: Cross-plugin entity aggregation and routing
- **Database Infrastructure**: Centralized SQLite with schema management, migrations, indexing
- **External Plugin Support**: JSON-RPC 2.0 protocol for language-agnostic plugins

**What Plugins Implement:**
- Domain logic and business rules
- Entity definitions and validation
- Event types and payloads
- Custom CLI commands
- External API integrations
- Event handlers for cross-plugin reactions

**Decision Guide:**
- Cross-plugin visibility? → Framework (analysis, entity aggregation)
- Infrastructure? → Framework (database, config, logging, RPC)
- Shared across plugins? → Framework (event storage, EventBus, Logger)
- Domain-specific? → Plugin (entities, commands, business rules)

**For Plugin Development:**
- **Template**: `template/go-plugin/README.md` - Complete plugin template with framework capabilities reference
- **SDK Docs**: `pkg/pluginsdk/CLAUDE.md` - Full API documentation
- **Example**: `pkg/plugins/claude_code/` - Reference implementation

### Real-Time Event Streaming (Phase 4)

DarwinFlow now supports real-time event streaming from multiple plugins simultaneously:

- **EventDispatcher**: Background event processing with buffered channels (100 event buffer)
- **Multi-Plugin Support**: 2+ plugins can emit events concurrently without blocking
- **High Throughput**: Validated at >30,000 events/sec (far exceeds 1,000/sec requirement)
- **TUI Real-Time Updates**: Auto-refresh session list when new events arrive
- **Event Counter Badge**: Shows "+N new" indicator in TUI status bar

#### Task Manager Commands

The task-manager plugin provides comprehensive hierarchical roadmap management with Roadmap → Track → Task → Iteration structure.

**Multi-Project Support:**

The task-manager supports multiple isolated projects (e.g., separate "production" and "test" roadmaps). Each project has its own database and complete isolation.

```bash
# Create a new project
dw task-manager project create test

# List all projects (* marks active project)
dw task-manager project list

# Switch active project
dw task-manager project switch test

# Show current active project
dw task-manager project show

# Delete a project
dw task-manager project delete test --force

# Use --project flag to override active project on any command
dw task-manager track list --project production
```

**Project Notes:**
- First run auto-creates a "default" project with any existing data
- All commands use the active project by default
- Use `--project <name>` flag to override on any command
- Cannot delete the currently active project (switch first)

**Roadmap Commands:**

```bash
# Initialize a new roadmap
dw task-manager roadmap init \
  --vision "Build extensible event-sourced framework" \
  --success-criteria "10 external plugins, 90% test coverage"

# Show current roadmap
dw task-manager roadmap show

# Update roadmap
dw task-manager roadmap update \
  --vision "Updated vision" \
  --success-criteria "New criteria"
```

**Track Commands (Major Work Areas):**

```bash
# Create a track
dw task-manager track create \
  --id track-framework-core \
  --title "Framework Core" \
  --description "Core framework implementation" \
  --priority high

# List all tracks (with filtering)
dw task-manager track list
dw task-manager track list --status in-progress --priority critical

# Show track details
dw task-manager track show track-framework-core

# Update track
dw task-manager track update track-framework-core --status in-progress

# Manage dependencies
dw task-manager track add-dependency track-plugin-system track-framework-core
dw task-manager track remove-dependency track-plugin-system track-framework-core

# Delete track
dw task-manager track delete track-framework-core --force
```

**Task Commands (Concrete Work Items):**

```bash
# Create a task
dw task-manager task create \
  --track track-framework-core \
  --title "Implement LLM abstraction" \
  --description "Create LLM interface in domain layer" \
  --priority high

# List tasks (with filtering)
dw task-manager task list
dw task-manager task list --track track-framework-core --status todo

# Show task details
dw task-manager task show task-fc-001

# Update task
dw task-manager task update task-fc-001 \
  --status in-progress \
  --branch feat/llm-abstraction

# Move task to different track
dw task-manager task move task-fc-001 --track track-plugin-system

# Delete task
dw task-manager task delete task-fc-001 --force
```

**Iteration Commands (Time-Boxed Sprints):**

```bash
# Create an iteration
dw task-manager iteration create \
  --name "Foundation Sprint" \
  --goal "Complete view-based analysis" \
  --deliverable "Plugin-agnostic analysis framework"

# List iterations
dw task-manager iteration list

# Show iteration details
dw task-manager iteration show 1

# Show current iteration
dw task-manager iteration current

# Update iteration
dw task-manager iteration update 1 --name "Sprint 1"

# Add/remove tasks
dw task-manager iteration add-task 1 task-fc-003 task-fc-005
dw task-manager iteration remove-task 1 task-fc-003

# Start iteration (mark as current)
dw task-manager iteration start 2

# Complete iteration
dw task-manager iteration complete 1

# Delete iteration
dw task-manager iteration delete 1 --force
```

**Interactive TUI (Terminal User Interface):**

```bash
# Launch interactive TUI
dw task-manager tui
```

Navigation:
- `j/k` or `↑/↓` - Navigate lists
- `enter` - Select/drill down
- `i` - Switch to iteration view
- `r` - Refresh data
- `esc` - Go back
- `q` - Quit

**Features:**
- Roadmap overview with tracks and task counts
- Track details with nested task lists
- Iteration planning and progress visualization
- Dependency visualization
- Status and priority filtering

### Plugin Event Bus

DarwinFlow includes a powerful **Plugin Event Bus** for cross-plugin communication using publish/subscribe patterns. Plugins can emit events and subscribe to events from other plugins, enabling rich integrations and workflows.

**Key Features:**
- **Publish/Subscribe**: Plugins publish events; other plugins subscribe with filters
- **Event Filtering**: Subscribe to specific event types, labels, or sources
- **Async Delivery**: Non-blocking event delivery with 30-second timeout per subscriber
- **Thread-Safe**: Concurrent publish/subscribe operations fully supported
- **Event Persistence**: Optional SQLite persistence for replay and audit trails
- **Event Replay**: Late-subscribing plugins can replay historical events

**Event Structure:**
```go
type BusEvent struct {
    ID        string                 // Unique event ID
    Type      string                 // e.g., "gmail.email_received"
    Source    string                 // Plugin ID that emitted the event
    Timestamp time.Time
    Labels    map[string]string      // Filterable labels
    Metadata  map[string]interface{} // Additional metadata
    Payload   []byte                 // JSON-encoded payload
}
```

**Example Use Cases:**
- **Gmail → Telegram**: Gmail plugin detects school event emails → Telegram bot sends notification
- **Calendar → Multiple Plugins**: Calendar event created → notifications, task creation, reminders
- **Cross-Plugin Workflows**: Build complex workflows across plugin boundaries

**Example: Publishing Events**
```go
// Plugin publishes an event
event, _ := pluginsdk.NewBusEvent("gmail.email_received", "gmail-plugin", emailData)
event.Labels["category"] = "school_notification"
event.Labels["priority"] = "high"

eventBus.Publish(ctx, event)
```

**Example: Subscribing to Events**
```go
// Plugin subscribes to events with filtering
filter := pluginsdk.EventFilter{
    TypePattern: "gmail.*",           // All Gmail events
    Labels: map[string]string{
        "category": "school_notification",
    },
}

handler := &MyEventHandler{}
subscriptionID, _ := eventBus.Subscribe(filter, handler)
```

**Event Filtering:**
- **Type Patterns**: Glob patterns (`gmail.*`, `*.event_detected`) or exact matches
- **Label Matching**: Filter by label key-value pairs (subset matching)
- **Source Plugin**: Filter events by originating plugin

**Persistence & Replay:**
- Events are optionally stored in SQLite for durability
- Late-subscribing plugins can replay historical events
- Useful for rebuilding state or catching up on missed events

**Architecture Documentation:**
- See `docs/arch-generated.md` for complete dependency graph
- See `CLAUDE.md` for plugin development guide
- Run `go-arch-lint docs` to regenerate after changes

## Usage

### Commands

```bash
# Initialize logging infrastructure
dw claude-code init                        # Or 'dw claude init' (backward compat)

# Update to latest version (run after upgrading DarwinFlow)
dw refresh                                 # Update database schema and hooks

# Log an event (typically called by hooks - backward compat)
dw claude log <event-type>
dw claude-code log <event-type>

# View logged events
dw logs                                    # Show 20 most recent logs
dw logs --limit 50                         # Show 50 most recent logs
dw logs --help                             # Show database schema and help

# Execute arbitrary SQL queries
dw logs --query "SELECT event_type, COUNT(*) FROM events GROUP BY event_type"

# Interactive UI for session management
dw ui                                      # Launch interactive terminal UI
dw ui --debug                              # Launch with debug logging
dw ui --db /path/to/db                     # Use custom database path

# Analyze sessions using AI
dw analyze --last                          # Analyze the most recent session
dw analyze --session-id <id>               # Analyze a specific session
dw analyze --view --session-id <id>        # View existing analysis
dw analyze --all                           # Analyze all unanalyzed sessions
dw analyze --refresh                       # Re-analyze all sessions (even already analyzed)
dw analyze --refresh --limit 5             # Re-analyze only latest 5 sessions

# Use different analysis prompts
dw analyze --last --prompt session_summary    # Factual session summary
dw analyze --last --prompt tool_analysis      # Agent-focused tool suggestions (default)

# Override config settings
dw analyze --last --model sonnet              # Use different model
dw analyze --last --token-limit 50000         # Use custom token limit

# Run plugin tools
dw project session-summary --last             # Display summary of last session
dw project session-summary --session-id <id>  # Display summary of specific session
dw project list                               # List all available plugin tools
```

#### Log Viewing Examples

```bash
# Show recent activity
dw logs --limit 10

# Count events by type
dw logs --query "SELECT event_type, COUNT(*) as count FROM events GROUP BY event_type ORDER BY count DESC"

# Find tool invocations in the last hour
dw logs --query "SELECT * FROM events WHERE event_type = 'tool.invoked' AND timestamp > strftime('%s', 'now', '-1 hour') * 1000"

# Search for specific content
dw logs --query "SELECT * FROM events WHERE content LIKE '%sqlite%' LIMIT 10"

# View database schema
dw logs --help
```

#### Session Analysis Examples

```bash
# Analyze your last session to identify patterns
dw analyze --last

# Analyze a specific session
dw logs --session-id 8518c46d-51fd-49ec-99ac-b41a613f33ac  # First, view the session
dw analyze --session-id 8518c46d-51fd-49ec-99ac-b41a613f33ac

# View a previously saved analysis
dw analyze --view --session-id 8518c46d-51fd-49ec-99ac-b41a613f33ac

# Batch analyze all unanalyzed sessions
dw analyze --all

# Re-analyze all sessions with updated prompts
dw analyze --refresh

# Re-analyze only the latest 10 sessions
dw analyze --refresh --limit 10
```

The analysis uses an **agent-focused prompt** where Claude Code analyzes its own work from a first-person perspective, identifying:
- **Tool gaps**: Where the agent lacked the right tools
- **Repetitive operations**: Where multiple primitive operations could be a single tool
- **Missing specialized agents**: Task types that would benefit from dedicated subagents
- **Workflow inefficiencies**: Multi-step sequences that should be automated

The output includes specific tool suggestions categorized as:
- Specialized Agents (e.g., test generation, refactoring)
- CLI Tools (command-line utilities to augment capabilities)
- Claude Code Features (new built-in capabilities)
- Workflow Automations (multi-step operations as single tools)

#### Interactive UI

The `dw ui` command launches an interactive terminal UI for browsing and managing sessions:

**Features:**
- **Session List View**: Browse all sessions with analysis status indicators
  - ✓ = Analyzed
  - ✗ = Not analyzed
  - ⟳N = Multiple analyses (N = count)
- **Session Details**: View session metadata, event counts, and analysis previews
- **Quick Actions**: Analyze, re-analyze, view, or save analyses to markdown
- **Keyboard Navigation**: Fast, keyboard-driven interface

**Keyboard Controls:**

*Session List:*
- `↑/↓` - Navigate sessions
- `Enter` - View session details
- `r` - Refresh session list
- `Esc` - Quit

*Session Detail:*
- `a` - Analyze session (run new analysis)
- `r` - Re-analyze session (refresh existing analysis)
- `s` - Save analysis to markdown file
- `v` - View full analysis
- `Esc` - Back to list

**Example Workflow:**
```bash
# Launch interactive UI
dw ui

# Navigate to a session and press Enter
# View analysis status and previews
# Press 'a' to analyze if not analyzed
# Press 's' to save analysis to markdown
# Press Esc to return to list
```

Markdown files are saved to the directory configured in `.darwinflow.yaml` (default: `./analysis-outputs/`) with customizable filename templates.

### Configuration

DarwinFlow uses `.darwinflow.yaml` for configuration. Create this file in your project root or home directory:

```yaml
analysis:
  token_limit: 100000                      # Max tokens for analysis context
  # Allowed models: sonnet, opus, haiku (aliases for latest versions)
  # Or specific versions: claude-sonnet-4-5-20250929, claude-opus-4-20250514,
  # claude-3-5-sonnet-20241022, claude-3-5-haiku-20241022
  model: "sonnet"                          # Model alias or full name
  parallel_limit: 3                        # Max parallel analysis executions
  # Prompts to run during analysis (runs in parallel)
  enabled_prompts:
    - tool_analysis                        # Default: run tool_analysis
    # - session_summary                    # Uncomment to also run session summaries
  auto_summary_enabled: false              # Enable auto session summaries
  auto_summary_prompt: "session_summary"   # Prompt for auto summaries
  claude_options:
    allowed_tools: []                      # Tools available during analysis (empty = none)
    system_prompt_mode: "replace"          # "replace" or "append"

ui:
  default_output_dir: "./analysis-outputs" # Directory for saved markdown files
  # Filename template for saved analyses
  # Available: {{.SessionID}}, {{.PromptName}}, {{.Date}}, {{.Time}}
  filename_template: "{{.SessionID}}-{{.PromptName}}-{{.Date}}.md"
  auto_refresh_interval: ""                # e.g., "30s" for auto-refresh (empty = disabled)

prompts:
  session_summary: |
    # Your custom session summary prompt here
  tool_analysis: |
    # Your custom tool analysis prompt here
```

**Key Configuration Options**:
- `enabled_prompts`: Array of prompts to run during analysis (runs in parallel)
  - `dw analyze --last` runs all enabled prompts
  - `dw analyze --last --prompt X` runs only prompt X (overrides config)
  - Prompts must exist in the `prompts` section
- `auto_summary_enabled`: Set to `true` to automatically analyze sessions when they end
- `token_limit`: Controls how many sessions can be batch-analyzed together
- `parallel_limit`: Controls concurrency for parallel analysis
- CLI flags can override any config setting

### Event Types

Currently captured events:

- `tool.invoked` - Claude Code tool invocation (Read, Write, Bash, etc.)
- `chat.message.user` - User prompt submission

### Environment Variables

- `DW_CONTEXT` - Set the current context (e.g., `project/myapp`)
- `DW_MAX_PARAM_LENGTH` - Maximum parameter length for logging (default: 30)

## Development

### Prerequisites

- Go 1.25.1 or later
- [go-arch-lint](https://github.com/fdaines/go-arch-lint) for architecture validation

### Building

```bash
# Build the CLI
make

# Run tests
make test
```

### Architecture Compliance

Before committing, ensure:

1. All tests pass: `go test ./...`
2. Zero architecture violations: `go-arch-lint .`
3. Documentation is up-to-date (see [CLAUDE.md](./CLAUDE.md))

### Generated Documentation

Architecture and API documentation is generated automatically:

```bash
# Generate comprehensive architecture documentation
go-arch-lint docs
```

This generates `docs/arch-generated.md` with the complete dependency graph, public APIs, and architectural validation.

## Project Structure

```
darwinflow-pub/
├── cmd/dw/                           # CLI entry points
│   ├── main.go                       # Main command router
│   ├── bootstrap.go                  # Dependency injection
│   ├── plugin_registration.go        # Plugin registration & adapters
│   ├── logs.go                       # Logs command handlers
│   └── analyze.go                    # Analysis command handlers
├── pkg/                              # Public APIs
│   ├── pluginsdk/                    # Plugin SDK (public contract)
│   │   ├── plugin.go                 # Plugin interface
│   │   ├── capability.go             # Capability interfaces
│   │   ├── entity.go                 # Entity interfaces
│   │   ├── event.go                  # Event types
│   │   └── context.go                # Plugin context
│   └── plugins/                      # Plugin implementations
│       └── claude_code/              # Claude Code plugin
│           ├── plugin.go             # Plugin implementation
│           ├── commands.go           # InitCommand, EmitEventCommand
│           └── session_entity.go     # Session entity
├── internal/                         # Internal packages
│   ├── domain/                       # Core domain types
│   │   ├── event.go                  # Domain event definitions
│   │   ├── analysis.go               # Analysis domain types
│   │   └── plugin.go                 # Domain plugin interfaces
│   ├── app/                          # Application services
│   │   ├── plugin_registry.go        # Plugin routing
│   │   ├── command_registry.go       # Command routing
│   │   ├── plugin_context.go         # SDK context implementation
│   │   ├── logs.go                   # Logs query service
│   │   └── analysis.go               # Analysis service
│   └── infra/                        # Infrastructure implementations
│       ├── sqlite_repository.go      # SQLite storage
│       ├── hook_config.go            # Hook configuration (plugin use)
│       ├── transcript.go             # Transcript parsing
│       └── context.go                # Context detection
├── docs/                             # Generated documentation
│   ├── arch-index.md                 # Architecture index
│   └── arch-generated.md             # Dependency graph
├── CLAUDE.md                         # AI agent instructions
└── README.md                         # This file
```

## Roadmap

### V1 (Current)
- ✅ Basic event capture (PreToolUse, UserPromptSubmit)
- ✅ SQLite storage with full-text search
- ✅ Hook management and merging
- ✅ Log viewer with SQL query support (`dw logs`)
- ✅ AI-powered session analysis (`dw analyze`)
- ✅ Agent-focused analysis prompt (Claude Code identifies what tools IT needs)
- ✅ Refresh functionality to re-analyze sessions with updated prompts
- ✅ Pattern detection and tool suggestions

### V2 (Planned)
- Vector embeddings for semantic search
- Cross-session pattern analysis
- Enhanced context extraction
- Automated tool generation from patterns

### V3 (Future)
- Workflow optimization suggestions
- Self-modifying commands based on patterns
- Advanced analytics and insights
- Real-time pattern detection during sessions

## Contributing

Contributions are welcome! Please ensure:

1. Code follows the 3-layer architecture
2. All tests pass (`go test ./...`)
3. Architecture linter passes (`go-arch-lint .`)
4. Documentation is updated for API/architecture changes

## License

MIT License - See LICENSE file for details

## Acknowledgments

Built to enhance [Claude Code](https://www.anthropic.com/claude/code) workflows and enable AI-assisted development insights.

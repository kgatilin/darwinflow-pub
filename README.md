# DarwinFlow

**Capture, store, and analyze your Claude Code interactions**

DarwinFlow is a lightweight logging system that automatically captures all Claude Code interactions as structured events. Built with event sourcing principles, it enables pattern detection, workflow optimization, and deep insights into your AI-assisted development sessions.

## Features

- **Automatic Logging**: Captures all Claude Code events via hooks (tool invocations, user prompts, etc.)
- **AI-Powered Analysis**: Analyze sessions using Claude to identify patterns and suggest optimizations
- **Log Viewer**: Query and explore captured events with `dw logs` command
- **Event Sourcing**: Immutable event log enabling replay and analysis
- **SQLite Storage**: Fast, file-based storage with full-text search
- **Zero Performance Impact**: Non-blocking, concurrent-safe logging
- **Context-Aware**: Automatically detects project context from environment
- **Clean Architecture**: Strict 3-layer design (`cmd → pkg → internal`)

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
dw claude init
```

This will:
- Create the SQLite database at `~/.darwinflow/logs/events.db`
- Add hooks to your Claude Code settings (typically `~/.claude/settings.json`)
- Configure automatic event capture for PreToolUse and UserPromptSubmit hooks

### Start Using Claude Code

After running `dw claude init`, restart Claude Code. All your interactions will now be automatically logged!

## Architecture

DarwinFlow follows a strict 3-layer architecture enforced by [go-arch-lint](https://github.com/fdaines/go-arch-lint):

```
cmd → pkg → internal
```

- **cmd/dw**: CLI entry points (`dw claude init`, `dw claude log`, `dw logs`)
- **pkg/claude**: Orchestration layer (settings management, logging coordination)
- **internal/**: Domain primitives (events, hooks config, storage interfaces)

### Key Components

- **Events** (`internal/events`): Event types and payload definitions
- **Hooks** (`internal/hooks`): Claude Code hook configuration and merging logic
- **Storage** (`internal/storage`): Storage interface definitions
- **Logger** (`pkg/claude`): Event logging and database interaction
- **Settings Manager** (`pkg/claude`): Claude Code settings file management

## Usage

### Commands

```bash
# Initialize logging infrastructure
dw claude init

# Log an event (typically called by hooks)
dw claude log <event-type>

# View logged events
dw logs                                    # Show 20 most recent logs
dw logs --limit 50                         # Show 50 most recent logs
dw logs --help                             # Show database schema and help

# Execute arbitrary SQL queries
dw logs --query "SELECT event_type, COUNT(*) FROM events GROUP BY event_type"

# Analyze sessions using AI
dw analyze --last                          # Analyze the most recent session
dw analyze --session-id <id>               # Analyze a specific session
dw analyze --view --session-id <id>        # View existing analysis
dw analyze --all                           # Analyze all unanalyzed sessions
dw analyze --refresh                       # Re-analyze all sessions (even already analyzed)
dw analyze --refresh --limit 5             # Re-analyze only latest 5 sessions
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
├── cmd/dw/                    # CLI entry points
│   ├── main.go                # Main command router
│   ├── claude.go              # Claude subcommand handlers
│   ├── logs.go                # Logs command handlers
│   ├── analyze.go             # Analysis command handlers
│   └── logs_test.go           # Logs command tests
├── internal/                  # Domain primitives
│   ├── domain/                # Core domain types
│   │   ├── event.go           # Event definitions
│   │   ├── analysis.go        # Analysis domain types
│   │   └── repository.go      # Repository interfaces
│   ├── app/                   # Application services
│   │   ├── logger.go          # Event logging service
│   │   ├── logs.go            # Logs query service
│   │   ├── analysis.go        # Analysis service
│   │   ├── analysis_prompt.go # Analysis prompt template
│   │   └── setup.go           # Initialization service
│   └── infra/                 # Infrastructure implementations
│       ├── sqlite_repository.go       # SQLite event & analysis storage
│       ├── hook_config.go             # Hook configuration management
│       ├── transcript.go              # Transcript parsing
│       └── context.go                 # Context detection
├── docs/                      # Generated documentation
│   ├── arch-generated.md      # Dependency graph
│   └── public-api-generated.md # Public API reference
├── CLAUDE.md                  # AI agent instructions
└── README.md                  # This file
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

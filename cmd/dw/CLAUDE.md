# Package: main (dw CLI)

**Path**: `cmd/dw`

**Role**: CLI entry points and application bootstrap

---

## Quick Reference

- **Files**: 14
- **Exports**: 6
- **Dependencies**: `internal/app`, `internal/app/tui`, `internal/infra`, `pkg/plugins/claude_code`
- **Layer**: Entry point (top of dependency graph)
- **Binary**: `dw` (DarwinFlow CLI)

---

## Generated Documentation

### Exported API

#### Bootstrap

**InitializeApp()**:
- Initialize application dependencies
- Parameters: dbPath (string), verbose (bool)
- Returns: `*AppServices`, error
- Creates all services, repositories, registries

**AppServices**:
- Container for all application services
- Properties: PluginRegistry, CommandRegistry, LogsService, AnalysisService, SetupService, ConfigLoader, Logger, EventRepo, DBPath, WorkingDir

**RegisterBuiltInPlugins()**:
- Register core plugins (ClaudeCodePlugin)
- Parameters: PluginRegistry, services, logger, dbPath
- Returns: error

#### Command Options

**LogsOptions**:
- CLI options for `dw logs` command
- Properties: Limit, SessionLimit, Query, SessionID, Ordered, Format, Help

#### Utilities

**ParseLogsFlags()**:
- Parse `dw logs` command flags
- Parameters: args ([]string)
- Returns: `*LogsOptions`, error

**PrintLogsHelp()**:
- Print help for `dw logs` command

---

## Architectural Principles

### What MUST Be Here

✅ **Main entry point** - `main()` function
✅ **CLI routing** - Subcommand dispatch
✅ **Bootstrap logic** - Dependency wiring
✅ **Plugin registration** - Register built-in plugins
✅ **Flag parsing** - Command-line argument parsing
✅ **Help text** - Usage information
✅ **Error handling** - Top-level error display

### What MUST NOT Be Here

❌ **Business logic** - Belongs in `internal/app`, `internal/domain`
❌ **Infrastructure** - Belongs in `internal/infra`
❌ **Plugin implementations** - Belong in `pkg/plugins/*`
❌ **UI components** - Belong in `internal/app/tui`
❌ **Heavy computation** - Delegate to services

### Critical Rules

1. **Thin Layer**: Minimal logic, delegate to services
2. **Dependency Injection**: Wire dependencies, don't create them inline
3. **Error Display**: Show user-friendly errors
4. **Exit Codes**: Use appropriate exit codes (0 = success, 1+ = error)
5. **Help Text**: Always provide `-h` / `--help` flags

---

## Application Bootstrap

### Initialization Flow

```
main()
  ↓
InitializeApp(dbPath, verbose)
  ↓
1. Determine DB path
2. Create Logger
3. Create SQLiteEventRepository
4. Create ConfigLoader
5. Load Config
6. Create PluginRegistry
7. Create Services (Logs, Analysis, Setup)
8. Create CommandRegistry
9. RegisterBuiltInPlugins()
  ↓
Return AppServices
```

### Dependency Graph

```
AppServices
├── PluginRegistry
│   └── Plugins (ClaudeCodePlugin, ...)
├── CommandRegistry
│   └── Commands from plugins
├── LogsService
│   └── EventRepository
├── AnalysisService
│   ├── EventRepository
│   ├── AnalysisRepository
│   ├── LogsService
│   ├── LLMExecutor
│   └── Config
├── SetupService
│   └── EventRepository
├── ConfigLoader
├── Logger
└── EventRepo (SQLiteEventRepository)
```

---

## Command Routing

### Main Command Structure

```
dw [global-flags] <command> [command-flags] [args]

Global flags:
  --db <path>     Database path
  --verbose       Enable verbose logging
  --help          Show help

Commands:
  claude <cmd>    Claude Code plugin commands
  logs            View event logs
  analyze         Analyze sessions
  ui              Launch TUI
  refresh         Refresh analyses
  config          Configure DarwinFlow
  help            Show help
```

### Routing Logic

```go
switch command {
case "claude":
    // Route to ClaudeCodePlugin via CommandRegistry
    registry.ExecuteCommand(ctx, "claude", args[1:], cmdCtx)

case "logs":
    // Parse flags, create handler, execute
    opts, _ := ParseLogsFlags(args[1:])
    handler := app.NewLogsCommandHandler(services.LogsService, os.Stdout)
    handler.ListLogs(ctx, opts.Limit, opts.SessionID, opts.Ordered, opts.Format)

case "analyze":
    // Create handler, execute
    handler := app.NewAnalyzeCommandHandler(services.AnalysisService, logger, os.Stdout)
    handler.Execute(ctx, analyzeOpts)

case "ui":
    // Launch TUI
    tui.Run(ctx, services.PluginRegistry, services.AnalysisService, services.LogsService, config)

// ... other commands
}
```

---

## Plugin Registration

### Built-in Plugins

```go
func RegisterBuiltInPlugins(
    registry *app.PluginRegistry,
    analysisService *app.AnalysisService,
    logsService *app.LogsService,
    logger app.Logger,
    setupService *app.SetupService,
    configLoader app.ConfigLoader,
    dbPath string,
) error {
    // Create ClaudeCodePlugin
    claudePlugin := claude_code.NewClaudeCodePlugin(
        analysisService,
        logsService,
        logger,
        setupService,
        configLoader,
        dbPath,
    )

    // Register with PluginRegistry
    return registry.RegisterPlugin(claudePlugin)
}
```

### Plugin Discovery

Once registered, plugins are discoverable via:
- `registry.GetPlugin(name)` - Get plugin by name
- `registry.GetCommandProvider(name)` - Get command provider
- `registry.GetAllPlugins()` - List all plugins

---

## Command Implementation Pattern

### Example: Analyze Command

```go
// analyze.go
func runAnalyzeCommand(services *AppServices, args []string) error {
    // 1. Parse flags
    opts := parseAnalyzeFlags(args)

    // 2. Create handler
    handler := app.NewAnalyzeCommandHandler(
        services.AnalysisService,
        services.Logger,
        os.Stdout,
    )

    // 3. Execute
    ctx := context.Background()
    return handler.Execute(ctx, opts)
}
```

**Pattern**:
1. Parse flags → Options struct
2. Create handler (with dependencies)
3. Execute handler
4. Handle errors

---

## Configuration

### Database Path Resolution

```go
// Priority:
1. --db flag
2. DARWINFLOW_DB environment variable
3. Default: ./.darwinflow/events.db
```

### Config File

```go
// Load from:
configLoader.LoadConfig("./.darwinflow/config.yaml")

// Or create default:
configLoader.InitializeDefaultConfig(configDir)
```

---

## Logging

### Verbose Mode

```bash
dw --verbose logs --limit 10
```

**Effect**:
- Sets logger level to Debug
- Shows detailed operation logs
- Useful for troubleshooting

### Standard Mode

- Info level
- Shows warnings and errors
- Minimal output

---

## Error Handling

### Top-Level Error Display

```go
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

### User-Friendly Errors

Map technical errors to user messages:
```go
if errors.Is(err, domain.ErrNotFound) {
    fmt.Fprintf(os.Stderr, "Session not found. Use 'dw logs' to list sessions.\n")
    os.Exit(1)
}
```

### Exit Codes

- `0` - Success
- `1` - General error
- `2` - Usage error (invalid flags, etc.)

---

## Help System

### Global Help

```bash
dw --help
dw help
```

**Shows**:
- Available commands
- Global flags
- Usage examples

### Command Help

```bash
dw logs --help
dw analyze --help
dw claude --help
```

**Shows**:
- Command description
- Command-specific flags
- Usage examples

### Command Help Implementation

```go
func PrintLogsHelp() {
    fmt.Println("Usage: dw logs [options]")
    fmt.Println("\nOptions:")
    fmt.Println("  --limit <n>        Limit output to n entries")
    fmt.Println("  --session-id <id>  Filter by session ID")
    // ...
}
```

---

## Testing Strategy

### Bootstrap Testing

Test application initialization:
- Verify dependency wiring
- Test database path resolution
- Test logger configuration
- Mock file system

### Command Routing Testing

Test command dispatch:
- Verify correct handler called
- Test flag parsing
- Test error propagation

### Plugin Registration Testing

Test plugin lifecycle:
- Verify plugins registered
- Test command discovery
- Test entity provider discovery

### Integration Testing

End-to-end tests:
- Full command execution
- Real database (temp)
- Real services
- Capture stdout/stderr

---

## Performance Considerations

- **Lazy initialization**: Only create services when needed
- **Connection pooling**: Reuse database connections
- **Early validation**: Fail fast on invalid flags
- **Minimal startup**: Keep bootstrap fast

---

## Files

- `main.go` - Main entry point
- `bootstrap.go` - Application initialization
- `plugin_registration.go` - Plugin registration
- `analyze.go` - Analyze command
- `logs.go` - Logs command
- `ui.go` - UI command (TUI launcher)
- `refresh.go` - Refresh command
- `config.go` - Config command
- `init.go` - Init command (legacy)
- `usage_test.go` - Help text tests
- `*_test.go` - Command tests

---

## Adding New Commands

### Steps

1. **Create command file**: `cmd/dw/mycommand.go`
2. **Parse flags**: Create options struct
3. **Create handler**: In `internal/app/mycommand_cmd.go`
4. **Wire in main**: Add case to command router
5. **Add help**: Implement help function
6. **Test**: Add command tests

### Example

```go
// cmd/dw/mycommand.go
func runMyCommand(services *AppServices, args []string) error {
    opts := parseMyCommandFlags(args)
    handler := app.NewMyCommandHandler(services.SomeService, os.Stdout)
    return handler.Execute(context.Background(), opts)
}

// main.go
case "mycommand":
    return runMyCommand(services, args[1:])
```

---

*Generated by `go-arch-lint -format=package cmd/dw`*

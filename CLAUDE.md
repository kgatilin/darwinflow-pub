# DarwinFlow - Claude Code Logging System

## Project Overview

**DarwinFlow** is a lightweight logging system that captures Claude Code interactions as structured events using event sourcing principles. The system stores events in SQLite and enables future pattern detection and workflow optimization.

### Key Components

- **CLI Tool (`dw`)**: Main entry point with multiple subcommands
  - `dw claude init` - Initialize logging infrastructure
  - `dw claude log` - Log events (called by hooks)
  - `dw logs` - View and query logged events
  - `dw analyze` - AI-powered session analysis with agent-focused prompts
    - `--last` - Analyze most recent session
    - `--session-id <id>` - Analyze specific session
    - `--all` - Analyze all unanalyzed sessions
    - `--refresh` - Re-analyze already analyzed sessions
    - `--limit N` - Limit refresh to latest N sessions
- **Event Logging**: Captures tool invocations and user prompts via Claude Code hooks
- **SQLite Storage**: Fast, file-based event storage with full-text search capability
- **Hook Management**: Automatically configures and merges Claude Code hooks
- **Log Viewer**: Query interface with SQL support for exploring captured events
- **AI Analysis**: Uses Claude CLI to analyze sessions and suggest workflow optimizations

### Architecture Documentation

For detailed architecture and API information, see:
- @docs/arch-generated.md - Complete dependency graph with method-level details and public APIs

### Current Implementation Status

**Active Hooks**:
- `PreToolUse`: Logs all tool invocations (Read, Write, Bash, etc.)
- `UserPromptSubmit`: Logs user message submissions

**Event Types**: Defined in `internal/domain/event.go`
- `tool.invoked`, `tool.result`
- `chat.message.user`, `chat.message.assistant`
- `chat.started`, `file.read`, `file.written`, etc.

**Analysis Features**:
- **Agent-Focused Analysis**: Uses a first-person prompt where Claude Code analyzes its own work
  - Identifies tools the agent needs (not human-facing suggestions)
  - Categories: Specialized Agents, CLI Tools, Claude Code Features, Workflow Automations
  - Configured via `.darwinflow.yaml` (`prompts.analysis`)
- Session analysis using Claude CLI (`claude -p`)
- **Refresh capability**: Re-analyze sessions with updated prompts using `--refresh --limit N`
- Pattern detection and tool gap identification
- Support for analyzing individual sessions, batch analysis, or all sessions
- Persistent storage of analysis results in `session_analyses` table

### Development Workflow

When working on this project:
1. Understand the 3-layer architecture (see below)
2. Check @docs/arch-generated.md to see current package dependencies
3. Check @docs/public-api-generated.md to see what's exported
4. Follow the architecture guidelines strictly
5. Write tests for new functionality (aim for 70-80% coverage)
6. **Update documentation** (README.md and CLAUDE.md) when adding features
7. Run tests and linter before committing
8. Regenerate architecture docs if needed

---

# go-arch-lint - Architecture Linting

**CRITICAL**: The .goarchlint configuration is IMMUTABLE - AI agents must NOT modify it.

## Architecture (3-layer strict dependency flow)

```
cmd → pkg → internal
```

**cmd**: Entry points (imports only pkg) | **pkg**: Orchestration & adapters (imports only internal) | **internal**: Domain primitives (NO imports between internal packages)

## Core Principles

1. **Dependency Inversion**: Internal packages define interfaces. Adapters bridge them in pkg layer.
2. **Structural Typing**: Types satisfy interfaces via matching methods (no explicit implements)
3. **No Slice Covariance**: Create adapters to convert []ConcreteType → []InterfaceType

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
3. **Refactor**: Move code to correct layer or add interfaces/adapters in pkg
4. **Verify**: Run `go-arch-lint .` - confirm zero violations

Example: `internal/A` imports `internal/B` → Should B's logic move to A? Should both define interfaces with pkg adapter? Architecture enforces intentional design.

## Code Guidelines

**DO**:
- Add domain logic to internal/ packages
- Define interfaces in consumer packages
- Create adapters in pkg/ to bridge internal packages
- Use black-box tests (`package pkgname_test`) for pkg packages

**DON'T**:
- Import between internal/ packages (violation!) or pass []ConcreteType as []InterfaceType
- Put business logic in pkg/ or cmd/ (belongs in internal/)
- Modify .goarchlint (immutable architectural contract)

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
- Example: `pkg/claude/logger.go` → `pkg/claude/logger_test.go`

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
go test -run TestFunctionName ./pkg/claude
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
go-arch-lint -detailed -format=markdown . > docs/arch-generated.md 2>&1
go-arch-lint -format=api . > docs/public-api-generated.md 2>&1
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

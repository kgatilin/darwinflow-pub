# DarwinFlow - Claude Code Logging System

## Project Overview

**DarwinFlow** captures Claude Code interactions as structured events using event sourcing. Events are stored in SQLite for pattern detection and workflow optimization.

### Key Components

- **CLI**: `dw claude init`, `dw refresh`, `dw logs`, `dw ui`, `dw analyze`
- **Event Logging**: Captures tool invocations and user prompts via hooks
- **SQLite Storage**: Event storage with full-text search
- **AI Analysis**: Session analysis with configurable prompts
- **Interactive TUI**: Browse sessions, view analyses, export to markdown

### Plugin Architecture

**SDK is Single Source of Truth** (`pkg/pluginsdk`):
- All plugin interfaces defined in `pkg/pluginsdk/` (Plugin, IEntityProvider, ICommandProvider, etc.)
- All entity interfaces defined in `pkg/pluginsdk/entity.go` (IExtensible, ITrackable, IHasContext, etc.)
- Zero interface duplication - if SDK has it, don't duplicate elsewhere
- Direct SDK usage (no adaptation layer, zero overhead)

**Plugin Types**:
- Internal plugins: `pkg/plugins/claude_code/` (ships with tool)
- External plugins: Planned (import `pkg/pluginsdk` only)

**Key Principle**: Framework is plugin-agnostic. Plugin-specific types (event types, payloads, analysis) belong in plugin packages, not in framework domain.

**For detailed plugin development**: See @docs/plugin-development-guide.md

---

## Development Workflow

**Note**: When the user refers to "workflow", they mean these CLAUDE.md instructions.

### Working on Features

1. Check @docs/arch-index.md for current package structure
2. Follow architecture guidelines (DDD layers, dependency rules)
3. Write tests for new functionality (target 70-80% coverage)
4. Update documentation when adding features
5. Run tests and linter before committing
6. Commit after each logical task/iteration

### Large Tasks - Use Task Tool Delegation

For substantial refactorings or multi-package features:

1. **Decompose** into context-sized chunks (packages, features, components)
2. **Delegate** each chunk sequentially using Task tool
3. **Review** sub-agent reports between chunks
4. **Verify** all tests/linter pass after completion

**Final Checklist** (use TodoWrite):
- [ ] Run `go test ./...` - all tests pass
- [ ] Run `go-arch-lint .` - zero violations
- [ ] Update README.md (if commands/features changed)
- [ ] Update CLAUDE.md (if workflow/architecture changed)
- [ ] Run `go-arch-lint docs` (if architecture/API changed)
- [ ] Commit with concise message

### Plugin Development

**SDK interfaces**: `pkg/pluginsdk/` (Plugin, IEntityProvider, ICommandProvider, IToolProvider, IExtensible, ITrackable, etc.)

**Creating entities**:
- IExtensible is **required** (GetID, GetType, GetCapabilities, GetField, GetAllFields)
- ITrackable, IHasContext, ISchedulable, IRelatable are **optional**
- Declare only capabilities you actually implement

**Plugin-specific types**:
- Event types, payloads, analysis → `pkg/plugins/yourplugin/`
- NOT in `internal/domain` (framework must be plugin-agnostic)

**Commands**: `dw <plugin-name> <command>` (plugin-scoped)
**Tools**: `dw project <toolname>` (project-scoped)

See @docs/plugin-development-guide.md for complete guide.

---

## Architecture (DDD Layered)

### Layer Structure

```
         ┌─────────────┐
         │     cmd     │  Entry points
         └──────┬──────┘
                │
        ┌───────┴────────┐
        │                │
        ▼                ▼
┌──────────────┐  ┌──────────────┐
│ internal/app │  │internal/infra│  Application & Infrastructure
└──────┬───────┘  └──────┬───────┘
       │                 │
       └────────┬────────┘
                ▼
        ┌──────────────┐
        │internal/domain│  Framework business logic
        └──────────────┘
                ▲
                │
        ┌──────────────┐
        │pkg/pluginsdk │  Public plugin contracts (zero internal deps)
        └──────────────┘
```

### Dependency Rules

- **pkg/pluginsdk**: Imports NOTHING (fully public)
- **internal/domain**: May import `pkg/pluginsdk` (for contracts)
- **internal/app**: May import `internal/domain`, `pkg/pluginsdk`
- **internal/infra**: May import `internal/domain`, `pkg/pluginsdk`
- **cmd**: May import `internal/app`, `internal/infra`, `pkg/plugins`

**Key Rule**: SDK is single source of truth. No interface duplication.

### Core Principles

1. **SDK First**: Public plugin contracts in `pkg/pluginsdk` (single source of truth)
2. **Zero Duplication**: If SDK has an interface, don't duplicate it
3. **Framework Agnostic**: `internal/domain` has zero plugin-specific knowledge
4. **Plugin-Specific Types**: Event types, payloads, analysis → plugin packages
5. **Dependency Inversion**: Define interfaces in domain/SDK, implement in infra/plugins

---

## Before Every Commit

1. `go test ./...` - all tests must pass
2. `go-arch-lint .` - **ZERO violations required** (non-negotiable)
3. Regenerate docs if needed: `go-arch-lint docs`
4. Update README.md / CLAUDE.md if functionality changed

---

## When Linter Reports Violations

**Don't mechanically fix imports.** Violations reveal architectural issues.

**Process**:
1. **Reflect**: Why does this violation exist? Wrong dependency?
2. **Plan**: Which layer should own this? Right structure?
3. **Refactor**: Move code to correct layer
4. **Verify**: `go-arch-lint .` → zero violations

**Common Violations**:

❌ `internal/domain` imports `internal/app` or `internal/infra`
→ Fix: Define interface in domain/SDK, implement in infra, inject via app

❌ `pkg/pluginsdk` imports `internal/*`
→ Fix: SDK must be fully public with zero internal dependencies

❌ Duplicate interfaces in both `pkg/pluginsdk` and `internal/domain`
→ Fix: Delete from domain, use SDK (single source of truth)

❌ `internal/domain` imports `pkg/plugins/claude_code`
→ Fix: Framework must be plugin-agnostic; move types to SDK or keep in plugin

✅ `internal/domain` imports `pkg/pluginsdk` - OK (SDK is public contracts)
✅ `internal/infra` → `internal/domain` - OK (allowed)

**Example - Domain needs database**:
- ✅ `internal/domain/repository.go` defines `EventRepository` interface
- ✅ `internal/infra/sqlite_repository.go` implements `EventRepository`
- ✅ `internal/app/` receives injected repository

---

## Code Guidelines

**DO**:
- Define plugin contracts in `pkg/pluginsdk` (single source of truth)
- Define framework interfaces in `internal/domain`
- Implement infrastructure in `internal/infra`
- Work with SDK types directly (zero adaptation)
- Use black-box tests (`package pkgname_test`)

**DON'T**:
- Duplicate interfaces between `pkg/pluginsdk` and `internal/domain`
- Import `internal/*` from `pkg/pluginsdk` (must be fully public)
- Create adaptation layers (work with SDK types directly)
- Put plugin-specific types in `internal/domain` (framework is plugin-agnostic)
- Modify `.goarchlint` (immutable)

---

## Testing

**Coverage Target**: 70-80%

**Package Naming**: `package pkgname_test` (black-box testing)

**File Naming**: `*_test.go` in same directory

**Test Naming**:
- `TestFunctionName` or `TestType_Method`
- Examples: `TestNewLogger`, `TestSQLiteStore_Init`

**Running Tests**:
```bash
go test ./...                               # All tests
go test -cover ./...                        # With coverage
go test -coverprofile=coverage.out ./...    # Coverage report
go tool cover -html=coverage.out            # View in browser
```

**Best Practices**:
- Each test is independent
- Use `t.TempDir()` for file operations
- Use `defer` for cleanup
- Test public API only (black-box)
- `t.Fatalf()` for setup failures, `t.Errorf()` for assertions

---

## Documentation

**CRITICAL**: Documentation must be updated when functionality changes.

### When to Update

**README.md** (user-facing):
- New commands or flags
- New features
- Changed behavior

**CLAUDE.md** (development):
- Workflow changes
- Architecture changes
- New patterns or conventions

**Generated docs**:
```bash
go-arch-lint docs  # After modifying packages or APIs
```

### Checklist

- [ ] Code implemented and tested
- [ ] README.md updated (if user-facing changes)
- [ ] CLAUDE.md updated (if workflow changes)
- [ ] Architecture docs regenerated (if needed)
- [ ] All tests pass
- [ ] Linter passes (zero violations)

---

## Key References

- **Architecture Index**: @docs/arch-index.md - Package structure and dependencies
- **Plugin Development**: @docs/plugin-development-guide.md - Complete plugin guide
- **SDK Reference**: `pkg/pluginsdk/` - Godoc for all plugin interfaces
- **Example Plugin**: `pkg/plugins/claude_code/` - Reference implementation
- **Linter**: `go-arch-lint .` - Validate architecture compliance

---

**Remember**: SDK is single source of truth. Framework is plugin-agnostic. Zero interface duplication.

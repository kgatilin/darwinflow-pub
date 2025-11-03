# DarwinFlow - Claude Code Logging System

## Project Overview

**DarwinFlow** captures Claude Code interactions as structured events using event sourcing. Events are stored in SQLite for pattern detection and workflow optimization.

### Key Components

- **CLI**: `dw claude init`, `dw refresh`, `dw logs`, `dw ui`, `dw analyze`
- **Event Logging**: Captures tool invocations and user prompts via hooks
- **SQLite Storage**: Event storage with full-text search
- **AI Analysis**: Session analysis with configurable prompts
- **Interactive TUI**: Browse sessions, view analyses, export to markdown
- **Plugin Event Bus**: Cross-plugin communication via publish/subscribe (all plugins have access)

### Plugin Architecture

DarwinFlow uses a plugin-based architecture:
- **SDK** (`pkg/pluginsdk`) - Public plugin contracts (single source of truth)
- **Core Plugin** (`pkg/plugins/claude_code`) - Claude Code event capture and analysis
- **Framework** (`internal/*`) - Plugin-agnostic event processing and storage

**Key Principle**: Framework is plugin-agnostic. Plugin-specific types belong in plugin packages.

**Framework vs Plugin Responsibilities**:
- **Framework handles**: Event storage, centralized analysis, cross-plugin communication (EventBus), logging, config, command routing, entity aggregation, database infrastructure, RPC protocol
- **Plugins handle**: Domain logic, entity definitions, event types/payloads, custom commands, external API integrations, event handlers

**Decision Guide**: Cross-plugin visibility → Framework. Infrastructure → Framework. Domain-specific → Plugin.

**For Plugin Development**:
- `template/go-plugin/README.md` - Complete plugin template with framework capabilities reference
- `pkg/pluginsdk/CLAUDE.md` - Full SDK API documentation
- `pkg/plugins/claude_code/` - Reference implementation
- `README.md` (main) - Framework capabilities overview

---

## Task Manager - Core Workflow

**Understanding What to Work On:**

```bash
# Check current iteration and tasks
dw task-manager iteration current

# View all tasks (with status filtering)
dw task-manager task list                    # All tasks
dw task-manager task list --status todo      # Backlog
dw task-manager task list --status in-progress  # Active work

# View track details and tasks
dw task-manager track show TM-track-X

# View task details (including acceptance criteria)
dw task-manager task show TM-task-X
```

**Working on Tasks:**

```bash
# Start a task (todo → in-progress)
dw task-manager task update TM-task-X --status in-progress

# Complete a task (in-progress → done)
dw task-manager task update TM-task-X --status done

# Return to backlog (in-progress → todo)
dw task-manager task update TM-task-X --status todo
```

**Acceptance Criteria:**

```bash
# Add acceptance criterion to task
dw task-manager ac add TM-task-X --description "..."

# List task acceptance criteria
dw task-manager ac list TM-task-X

# Mark as verified (when done)
dw task-manager ac verify TM-ac-X

# Mark as failed with feedback
dw task-manager ac fail TM-ac-X --feedback "..."

# List all failed ACs (for debugging/fixing)
dw task-manager ac failed
dw task-manager ac failed --task TM-task-X     # Filter by task
dw task-manager ac failed --iteration 11       # Filter by iteration
```

**Creating New Work:**

```bash
# 1. Create a new track
dw task-manager track create --title "..." --description "..." --rank 100

# 2. Create ADR for the track (REQUIRED before implementation)
dw task-manager adr create <track-id> \
  --title "..." \
  --context "..." \
  --decision "..." \
  --consequences "..." \
  --alternatives "..."

# 3. Update ADR status to 'accepted' after review
dw task-manager adr update <adr-id> --status accepted

# 4. Create tasks in the track with acceptance criteria
dw task-manager task create --track TM-track-X --title "..." --rank 100
dw task-manager ac add TM-task-X --description "..."

# 5. Create iteration and add tasks
dw task-manager iteration create --name "..." --goal "..." --deliverable "..."
dw task-manager iteration add-task <iter-num> TM-task-1 TM-task-2

# 6. Start working on iteration
dw task-manager iteration start <iter-num>
```

**Architecture Decision Records (ADRs) - Additional Commands:**

```bash
# List all ADRs
dw task-manager adr list

# Show ADR details
dw task-manager adr show TM-adr-X

# Mark ADR as superseded
dw task-manager adr supersede TM-adr-X --superseded-by TM-adr-Y

# Mark ADR as deprecated
dw task-manager adr deprecate TM-adr-X

# Check if track has required ADR
dw task-manager track check <track-id>

# List tracks without ADRs
dw task-manager track list-missing-adrs
```

**Priority Guidance**: Work on current iteration first → critical/high priority tracks → planned iterations.

**Best Practices**:
- **ALWAYS create ADR immediately after creating a track** (before any implementation)
- Update task status as you work (don't batch updates)
- Verify all acceptance criteria before marking task "done"
- Use `dw task-manager iteration current` to stay focused
- Check track dependencies before starting new tracks

### Writing Good Acceptance Criteria

**Core Principle**: Acceptance criteria must describe **end-user verifiable functionality** that focuses on **core business logic**, not implementation details or edge cases.

**Good AC Characteristics**:
- ✅ Describes WHAT the user can verify, not HOW it's implemented
- ✅ Focuses on observable behavior and outcomes
- ✅ Can be tested/verified by an end user
- ✅ Addresses core business logic
- ✅ Written from user perspective

**Bad AC Characteristics**:
- ❌ Implementation details (repositories, services, internal methods)
- ❌ Edge cases and technical minutiae
- ❌ Things only developers care about
- ❌ Internal code structure or architecture

**Examples**:

Good ACs:
- ✅ "User can export session data to CSV format"
- ✅ "Command displays error message when file not found"
- ✅ "TUI shows task status with color coding"
- ✅ "Testing instructions are saved and displayed in AC details"

Bad ACs:
- ❌ "ExportRepository saves data to database"
- ❌ "AcceptanceCriterionRepository updated to save and retrieve testing_instructions field"
- ❌ "Service validates input parameters"
- ❌ "Error handling covers nil pointer edge case"

**Testing Instructions**: When adding acceptance criteria, include step-by-step testing instructions that allow anyone to verify the criterion is met.

### Task Granularity

**Core Principle**: If you cannot write end-user verifiable acceptance criteria for a task, the task is likely **too granular** and should be merged into a larger, user-facing task.

**Good Task Granularity**:
- ✅ Represents a complete user-facing feature or capability
- ✅ Has at least 3-5 end-user verifiable acceptance criteria
- ✅ Delivers observable value to the end user
- ✅ Can be demonstrated and tested independently

**Too Granular (merge into larger task)**:
- ❌ "Create X entity" - implementation detail, merge into command that uses it
- ❌ "Add database migration" - implementation detail, happens as part of feature
- ❌ "Define X interface" - implementation detail, merge into service that implements it
- ❌ "Add X field to entity" - implementation detail, merge into feature that uses it

**Examples**:

Good Tasks:
- ✅ "Add task validation and comment commands" (includes entity creation, repository, commands)
- ✅ "Implement iteration locking workflow" (includes status fields, entities, commands)
- ✅ "Show iteration membership in task details" (includes repository method, CLI, TUI)

Too Granular (should be merged):
- ❌ "Create TaskComment entity" → Merge into "Add task validation commands"
- ❌ "Database migration for task planning" → Merge into "Add task planning commands"
- ❌ "Add iteration status fields" → Merge into "Add iteration locking commands"

**Guidelines**:
- Tasks should represent **features**, not implementation steps
- Implementation details (entities, migrations, repositories) are part of feature delivery
- Each task should answer: "What can the user now do that they couldn't before?"
- If the answer is "nothing visible", the task is too granular

---

## Package Structure

**Foundation Layer**:
- `pkg/pluginsdk` - Public plugin SDK (zero internal dependencies) → See `pkg/pluginsdk/CLAUDE.md`

**Framework Layer**:
- `internal/domain` - Framework business logic (plugin-agnostic) → See `internal/domain/CLAUDE.md`
- `internal/infra` - Infrastructure implementations (DB, config, logging) → See `internal/infra/CLAUDE.md`
- `internal/app` - Application services and orchestration → See `internal/app/CLAUDE.md`
- `internal/app/tui` - Terminal user interface (Bubble Tea) → See `internal/app/tui/CLAUDE.md`

**Plugin Layer**:
- `pkg/plugins/claude_code` - Claude Code plugin implementation → See `pkg/plugins/claude_code/CLAUDE.md`

**Entry Layer**:
- `cmd/dw` - CLI entry points and bootstrap → See `cmd/dw/CLAUDE.md`

**Architecture Reference**: `@docs/arch-index.md` - Full dependency graph and package details

**Package-Level Docs**: Each package has a `CLAUDE.md` with architectural guidance. Claude will read these automatically when working in those packages.

---

---

## Analysis Architecture

**View-Based Analysis**: The framework provides AI analysis capabilities for any view type through a plugin-agnostic architecture.

### Core Components

**LLM Abstraction**:
- `LLM` interface (`internal/domain`) - Abstract LLM provider contract
- `ClaudeCodeLLM` (`internal/infra`) - Claude Code CLI implementation
- Swappable implementations (Claude, Anthropic API, OpenAI, etc.)

**View-Based Pattern**:
- `AnalysisView` interface (`pkg/pluginsdk`) - Plugin contract for providing analysis views
- `Analysis` type (`internal/domain`) - Generic analysis results (view-agnostic)
- `AnalysisService.AnalyzeView()` - Analyzes any view implementing AnalysisView

**Plugin Implementation**:
- Plugins implement `AnalysisView` to provide views of their events
- Example: `SessionView` in Claude Code plugin provides session-based views
- Framework analyzes any view using `AnalysisService.AnalyzeView()`
- Plugins control how their events are formatted for LLM analysis

**Storage**:
- Generic `analyses` table stores all analysis types
- View metadata stored as JSON (flexible, plugin-specific)
- Migrated from old `session_analyses` table (Phase 3)

### Backward Compatibility

**SessionAnalysis Type**:
- Exists in `internal/domain` for backward compatibility with internal code
- Wraps the generic `Analysis` type (converts SessionAnalysis ↔ Analysis)
- New features should use `Analysis + AnalysisView` pattern
- Maintained for internal framework code written before refactoring

**Repository Interface**:
- Generic methods: `SaveGenericAnalysis()`, `FindAnalysisByViewID()` (primary)
- Session methods: `SaveAnalysis()`, `GetAnalysisBySessionID()` (compatibility layer)
- Both interfaces operate on same underlying `analyses` table

### Architecture Flow

```
Plugin (e.g., claude-code)
  ↓ implements
AnalysisView interface (SDK)
  ↓ provides to
AnalysisService.AnalyzeView()
  ↓ uses
LLM interface (domain)
  ↓ implemented by
ClaudeCodeLLM (infra)
  ↓ stores
Analysis (generic, domain)
  ↓ persisted in
AnalysisRepository
  ↓ stores as
analyses table (SQLite)
```

### Benefits

- **Plugin-Agnostic**: Framework has no knowledge of "sessions" or other plugin entities
- **Extensible**: Any plugin can leverage analysis (Task Manager, Gmail, Calendar)
- **Swappable LLM**: Easy to swap LLM providers (Claude CLI → Anthropic API)
- **Cross-Plugin Analysis**: Future support for analyzing events from multiple plugins
- **Clean Architecture**: Plugins own views, framework provides capability

---

## Development Workflow

**Note**: When the user refers to "workflow", they mean these CLAUDE.md instructions.

### Working on Features

1. Check `@docs/arch-index.md` for current package structure
2. Read relevant package `CLAUDE.md` for architectural guidance
3. Follow DDD layer rules and dependency constraints
4. Write tests for new functionality (target 70-80% coverage)
5. Update documentation when adding features
6. Run tests and linter before committing
7. Commit after each logical task/iteration

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

**Note**: ADRs should be created during task preparation (before implementation), not after.

### Roadmap Tracking

**Purpose**: Track active work items, enable task switching, prevent losing context.

**Files**:
- `@.agent/roadmap.md` - Active work items with status and next steps
- `.agent/roadmap_done.md` - Completed items archive
- `.agent/details/<item-name>.md` - Detailed implementation docs with checklists

**Workflow**:

1. **Starting new work**:
   - Add item to `.agent/roadmap.md` with status and description
   - Create `.agent/details/<item-name>.md` from template
   - Add detailed requirements, implementation checklist, technical notes

2. **During work**:
   - Update checklist in details doc as you progress
   - Add progress log entries with what was done/learned
   - Update "Next Steps" in roadmap.md when switching tasks

3. **Switching tasks**:
   - Leave clear notes in roadmap.md about current state
   - Update status (In Progress → On Hold if needed)
   - Can work on refactoring, tests, or small features and return

4. **Completing work**:
   - Mark all checklist items complete in details doc
   - Move item from `roadmap.md` to `roadmap_done.md` with completion date
   - Update CLAUDE.md/README.md if needed

**Template**: See `.agent/details/template.md` for detailed doc structure

**Benefits**:
- Quick context recovery when returning to work
- Track multiple parallel efforts (features, refactoring, tests)
- Historical record of completed work
- Prevents "what was I doing?" moments

---

## Architecture Quick Reference

### Dependency Rules

- **pkg/pluginsdk**: Imports NOTHING (fully public)
- **internal/domain**: May import `pkg/pluginsdk` (no other internal packages)
- **internal/infra**: May import `internal/domain`, `pkg/pluginsdk`
- **internal/app**: May import `internal/domain`, `internal/infra`, `pkg/pluginsdk`
- **pkg/plugins/***: May import `pkg/pluginsdk` ONLY (no internal packages)
- **cmd/***: May import `internal/app`, `internal/infra`, `pkg/plugins`

**Key Rule**: SDK is single source of truth. No interface duplication.

### Core Principles

1. **SDK First**: Public plugin contracts in `pkg/pluginsdk`
2. **Zero Duplication**: If SDK has it, don't duplicate elsewhere
3. **Framework Agnostic**: `internal/domain` has zero plugin-specific knowledge
4. **Plugin-Specific Types**: Event types, payloads, analysis → plugin packages
5. **Dependency Inversion**: Define interfaces in domain/SDK, implement in infra/plugins

**Details**: See package-level `CLAUDE.md` files for layer-specific guidance.

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

### Documentation Types

**README.md** (user-facing):
- New commands or flags
- New features
- Changed behavior

**CLAUDE.md** (this file - development workflow):
- Workflow changes
- Architecture changes
- New patterns or conventions

**Package CLAUDE.md** (package-level architectural guidance):
- What belongs in this package vs elsewhere
- Layer-specific patterns and rules
- Testing strategies
- **Update with**: `/utility:update_package_docs`

**Generated docs**:
```bash
go-arch-lint docs  # Regenerates docs/arch-index.md
```

### Documentation Checklist

- [ ] Code implemented and tested
- [ ] README.md updated (if user-facing changes)
- [ ] CLAUDE.md updated (if workflow changes)
- [ ] Package CLAUDE.md updated (if package responsibilities changed)
- [ ] Architecture docs regenerated (if needed)
- [ ] All tests pass
- [ ] Linter passes (zero violations)

**Note**: ADRs should be created during track creation (before implementation), not during documentation phase.

---

## Key References

- **Architecture Index**: `@docs/arch-index.md` - Package structure and dependencies
- **Plugin Development**: `@docs/plugin-development-guide.md` - Complete plugin guide
- **Package Documentation**: `<package>/CLAUDE.md` - Package-specific architectural guidance
- **Linter**: `go-arch-lint .` - Validate architecture compliance

---

**Remember**: SDK is single source of truth. Framework is plugin-agnostic. Zero interface duplication.

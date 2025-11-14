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

**Supported Plugins**:
- `claude_code` - Event capture, session analysis, TUI
- `task_manager` - Roadmap/task tracking (workflow below, architecture in `pkg/plugins/task_manager/CLAUDE.md`)

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

# 2. Create ADR document for the track (RECOMMENDED before implementation)
dw task-manager doc create \
  --title "ADR: ..." \
  --type adr \
  --content "# Context\n...\n\n# Decision\n...\n\n# Consequences\n..." \
  --track <track-id>

# Or from file
dw task-manager doc create \
  --title "ADR: ..." \
  --type adr \
  --from-file ./docs/adr.md \
  --track <track-id>

# 3. Create tasks in the track with acceptance criteria
dw task-manager task create --track TM-track-X --title "..." --rank 100
dw task-manager ac add TM-task-X --description "..."

# 4. Create iteration and add tasks
dw task-manager iteration create --name "..." --goal "..." --deliverable "..."
dw task-manager iteration add-task <iter-num> TM-task-1 TM-task-2

# 5. Start working on iteration
dw task-manager iteration start <iter-num>
```

**Document Management Commands:**

```bash
# Create document (ADR, plan, retrospective, etc.)
dw task-manager doc create \
  --title "..." \
  --type adr \
  --from-file ./docs/adr.md \
  --track TM-track-X

# Or create inline
dw task-manager doc create \
  --title "..." \
  --type plan \
  --content "# Planning doc..."

# List documents (filter by type)
dw task-manager doc list
dw task-manager doc list --type adr

# Show document
dw task-manager doc show TM-doc-X

# Update document
dw task-manager doc update TM-doc-X --from-file ./updated.md

# Attach to track or iteration
dw task-manager doc attach TM-doc-X --track TM-track-Y
dw task-manager doc attach TM-doc-X --iteration 5

# Detach document
dw task-manager doc detach TM-doc-X

# Delete document
dw task-manager doc delete TM-doc-X [--force]
```

**Priority Guidance**: Work on current iteration first → critical/high priority tracks → planned iterations.

**Best Practices**:
- **Create ADR document for tracks before implementation** (use `doc create --type adr`)
- Update task status as you work (don't batch updates)
- Verify all acceptance criteria before marking task "done"
- Use `dw task-manager iteration current` to stay focused
- Check track dependencies before starting new tracks
- Use documents for planning, retrospectives, and decision records

### Writing Good Acceptance Criteria

**Core Principle**: Acceptance criteria must describe **end-user verifiable functionality** that focuses on **core business logic**, not implementation details or edge cases.

**Command Structure**:
```bash
dw task-manager ac add <task-id> \
  --description "What must be verified (end-user observable)" \
  --testing-instructions "Step-by-step instructions to verify"
```

**CRITICAL**: Use separate fields:
- `--description`: The acceptance criterion itself (WHAT needs to be verified)
- `--testing-instructions`: Step-by-step instructions (HOW to verify it)

**Good AC Characteristics**:
- ✅ Describes WHAT the user can verify, not HOW it's implemented
- ✅ Focuses on observable behavior and outcomes
- ✅ Can be tested/verified by an end user
- ✅ Addresses core business logic
- ✅ Written from user perspective
- ✅ Testing instructions in separate field with numbered steps

**Bad AC Characteristics**:
- ❌ Implementation details (repositories, services, internal methods)
- ❌ Edge cases and technical minutiae
- ❌ Things only developers care about
- ❌ Internal code structure or architecture
- ❌ Testing instructions mixed into description field

**Examples**:

Good AC with proper separation:
```bash
dw task-manager ac add TM-task-X \
  --description "Domain layer has 90%+ test coverage with all tests passing" \
  --testing-instructions "1. Run: cd pkg/plugins/task_manager/domain
2. Run: go test ./... -coverprofile=coverage.out
3. Run: go tool cover -func=coverage.out | grep total
4. Verify: total coverage >= 90%
5. Run: go test ./... -v
6. Verify: All tests pass with zero failures"
```

Bad AC (everything in description):
```bash
dw task-manager ac add TM-task-X \
  --description "Domain layer has 90%+ test coverage

Testing instructions:
1. Run: go test ./...
2. Verify coverage >= 90%"
```

**Testing Instructions Best Practices**:
- Start each step with a number
- Use exact commands (copy-paste ready)
- Include verification steps ("Verify: X should show Y")
- Make it reproducible by anyone
- Focus on observable outcomes, not internal state

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

### Reporting Completed Work

**CRITICAL**: When reporting work completion to the user, follow these guidelines:

**DO**:
- ✅ Highlight deviations from the original plan
- ✅ Emphasize questions that require user decision
- ✅ Call out issues that need user attention
- ✅ Note any blockers or unexpected challenges
- ✅ Use clear, separate blocks for user action items

**DON'T**:
- ❌ Provide detailed summaries of what was implemented (user knows the tasks)
- ❌ Include standard "verify acceptance criteria" instructions (goes without saying)
- ❌ Repeat task descriptions or requirements back to user
- ❌ List every file changed or every test added (user can see git commit)
- ❌ Explain what was supposed to happen (user defined the tasks)

**Report Format**:

```markdown
# Implementation Complete: [Task/Iteration Name]

## Status
[One line: Complete / Complete with deviations / Blocked]

## Deviations from Plan
[Only if there were deviations - explain what and why]

## Questions for User
[Only if decisions needed - clear, actionable questions]

## Issues Requiring Attention
[Only if there are blockers or problems]

## Commit
[Commit hash and one-line summary]
```

**Example - Good Report**:
```markdown
# Implementation Complete: Iteration #27 TODO Tasks

## Status
Complete with minor deviations

## Deviations from Plan
1. Mocks placed in `application/mocks/` instead of `domain/repositories/mocks/` (AC-482 specifies domain layer)
2. Infrastructure coverage at 52.3% vs 60% target (7.7% gap)

## Questions for User
1. Should mocks stay in application/ or move to domain/repositories/? (affects linter violations)
2. Is 52.3% infrastructure coverage acceptable, or should I add more tests?

## Commit
8b103f8 feat: complete iteration #27 TODO tasks - test architecture refactoring
```

**Example - Bad Report**:
```markdown
# Implementation Complete: Iteration #27 TODO Tasks

## Summary
Successfully implemented all 4 TODO tasks...
[3 paragraphs explaining what was done]

## Implementation Phases
Phase 1: Created mocks...
Phase 2: Refactored tests...
[Detailed breakdown of every step]

## Files Modified
- Created: application/mocks/ (6 files)
- Modified: 10 files
[Long list of every file]

## Next Steps - USER ACTION REQUIRED
1. Review All Acceptance Criteria
2. Verify Each AC
3. Close Tasks
[Standard AC verification instructions]
```

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

### End-to-End (E2E) Testing

**Location**: `pkg/plugins/task_manager/e2e_test/`

**Purpose**: Validate the complete system by building the binary from source and executing actual CLI commands. E2E tests catch issues that unit tests miss:
- Build failures
- CLI command parsing bugs
- Database connection lifecycle issues
- Integration between layers
- Real-world command workflows

**CRITICAL PRINCIPLE**: E2E tests MUST build and test the actual codebase, not rely on system-installed binaries.

**Running E2E Tests**:
```bash
# All E2E tests (builds binary automatically)
go test ./pkg/plugins/task_manager/e2e_test/... -v

# Specific test suite
go test ./pkg/plugins/task_manager/e2e_test/... -v -run TestProjectSuite
go test ./pkg/plugins/task_manager/e2e_test/... -v -run TestWorkflowSuite

# With parallel execution
go test ./pkg/plugins/task_manager/e2e_test/... -v -parallel 4

# Single test
go test ./pkg/plugins/task_manager/e2e_test/... -v -run TestProjectSuite/TestProjectCreate
```

**Test Structure**:
```
e2e_test/
├── e2e_test.go          # Base suite with binary build + shared setup
├── project_test.go      # Project commands (create, switch, delete)
├── track_test.go        # Track commands and dependencies
├── adr_test.go          # ADR lifecycle
├── task_test.go         # Task management
├── ac_test.go           # Acceptance criteria
├── iteration_test.go    # Iteration lifecycle
└── workflow_test.go     # Integration workflows
```

**How E2E Tests Work**:

1. **Binary Build** (SetupSuite):
   - Builds `dw` binary from current source: `go build -o /tmp/dw-e2e-test ./cmd/dw`
   - Fails fast if code doesn't compile
   - Binary used for ALL test commands

2. **Project Isolation** (SetupSuite):
   - Each test suite creates unique project (e.g., `e2e-test-1234567890`)
   - Initializes roadmap once per suite
   - Prevents test conflicts

3. **Command Execution** (run() helper):
   - Uses built binary: `exec.Command(dwBinaryPath, "task-manager", args...)`
   - Captures stdout/stderr
   - Returns output for assertions

4. **ID Extraction** (parseID() helper):
   - Extracts entity IDs from command output
   - Supports patterns: `ID: TM-track-123` or `-track-` in output
   - Usage: `trackID := s.parseID(output, "track")`

**E2E Test Best Practices**:

✅ **DO**:
- Test complete workflows (create → update → verify → delete)
- Use `parseID()` helper with entity type: `parseID(output, "track")`
- Test error cases (missing flags, invalid IDs, constraint violations)
- Verify command output format matches expectations
- Use suite-level setup for shared resources (project, roadmap)
- Test parallel operations when relevant

❌ **DON'T**:
- Use system-installed `dw` binary from PATH
- Call `defer cleanup()` in GetCommands() or similar lifecycle methods
- Create test data in production database
- Assume IDs - always extract from output
- Skip building the binary (creates false positives)
- Use hardcoded project names (causes conflicts)

**Common E2E Test Patterns**:

```go
// 1. Create entity and extract ID
output, err := s.run("track", "create", "--title", "Test", "--description", "Test", "--rank", "100")
s.Require().NoError(err)
trackID := s.parseID(output, "track")  // Extracts "TM-track-123"
s.Require().NotEmpty(trackID)

// 2. Update entity
output, err = s.run("track", "update", trackID, "--status", "in-progress")
s.Require().NoError(err)

// 3. Verify entity state
output, err = s.run("track", "show", trackID)
s.Require().NoError(err)
s.Require().Contains(output, "in-progress")

// 4. Test error cases
output, err = s.run("track", "create", "--title", "Missing description")
s.Require().Error(err)  // Should fail due to missing --description flag
s.Require().Contains(output, "required")
```

**Test Isolation Rules**:

1. **Suite-level isolation**: Each test suite uses a unique project
2. **Test-level independence**: Tests should not depend on execution order
3. **No cleanup needed**: Tests operate on isolated projects, garbage collected by OS
4. **Parallel safety**: Tests can run in parallel if properly isolated

**Debugging E2E Test Failures**:

```bash
# 1. Check binary builds
go build -o /tmp/dw-test ./cmd/dw
/tmp/dw-test task-manager --help

# 2. Run single failing test with verbose output
go test ./pkg/plugins/task_manager/e2e_test/... -v -run TestWorkflowSuite/TestCompleteTaskLifecycle

# 3. Check command output in test logs
# Look for "Output:" sections in test failure messages

# 4. Test command manually
cd /tmp && mkdir test-debug && cd test-debug
dw task-manager project create test-debug
dw task-manager project switch test-debug
dw task-manager roadmap init --vision "Test" --success-criteria "Test"
# ... reproduce failing command sequence
```

**Key Lessons Learned**:

1. **Binary Must Be Built**: Using system PATH creates false positives
2. **Database Lifecycle**: Never close connections in initialization methods
3. **ID Extraction**: Use entity type names ("track"), not full phrases ("Created track:")
4. **Project Isolation**: Unique names prevent parallel test conflicts
5. **Error Testing**: Verify both success and failure cases

**When E2E Tests Fail**:

- ✅ Indicates real bugs (build issues, CLI bugs, integration problems)
- ✅ Catches what unit tests miss
- ✅ Validates end-user experience

**IMPORTANT - E2E Test Suite Integrity**:

The E2E test suite MUST be run to verify that all changes work correctly. However:

⚠️ **DO NOT modify E2E tests unless**:
- Explicitly defined in task acceptance criteria, OR
- Explicitly requested by the user

If neither condition is met, the agent MUST NOT change the test suite. E2E tests define the contract for how the system should behave.

**When to Add E2E Tests**:

- New CLI command added
- Command behavior changes
- Integration workflow added
- Bug found in production usage
- Complex multi-command scenarios

**Detailed E2E Documentation**: See `pkg/plugins/task_manager/e2e_test/CLAUDE.md`

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

---
allowed-tools:
  - Read
  - Edit
  - Task
  - TodoWrite
  - Bash(dw:*)
  - Bash(go:*)
  - Bash(go-arch-lint:*)
  - Bash(make:*)
  - Bash(git:*)
argument-hint: "[iteration-number-or-id]"
description: Implement TODO tasks in iteration with multi-agent workflow (plan → implement → verify). Ignores in-progress/review tasks.
---

# Iteration Implementation Command

You orchestrate a multi-agent workflow to implement TODO tasks in an iteration.

**Critical Rules**:
- ONLY implement tasks with status "todo"
- IGNORE tasks with status "in-progress", "review", "done", or "blocked"
- You orchestrate; sub-agents execute
- You do NOT verify acceptance criteria or close iterations

---

## Workflow Overview

1. **Identify target iteration** (current, next by rank, or specified)
2. **Create/switch git branch** (`iteration-<number>`)
3. **Start iteration** (if status "planned")
4. **Filter TODO tasks** (ignore in-progress/review/done/blocked)
5. **Read attached documents** (MANDATORY if documents exist)
6. **Planning agent** → Reads documents, explores codebase, retrieves full context, creates implementation plan
7. **Implementation agents** → Execute phases (sequential or parallel)
8. **Verification agent** → Tests, linter, code quality, task status
9. **Git commit** → Save all work
10. **Report to user** → Deviations, questions, issues only (see CLAUDE.md reporting guidelines)

---

## Phase 1: Identify Target Iteration

```bash
dw task-manager iteration current
```

**If argument provided**: Use iteration number or ID
**If no argument**: Use current iteration, or find first "planned"/"in-progress" by rank
**If none found**: Exit with message to create iteration

---

## Phase 2: Create/Switch Git Branch

Extract iteration number from target → construct `iteration-<number>` branch name

```bash
# Check current branch
git branch --show-current

# Create or checkout iteration branch
git checkout iteration-<number> 2>/dev/null || git checkout -b iteration-<number>
```

**If already on iteration branch**: Continue (supports multiple runs)
**If on different branch**: Create and switch (or switch if exists)

---

## Phase 3: Start Iteration (if needed)

If iteration status is "planned":
```bash
dw task-manager iteration start $TARGET_ITERATION
```

---

## Phase 4: Filter TODO Tasks

```bash
dw task-manager iteration show $TARGET_ITERATION --full
```

**Parse output** to extract tasks with status "todo"
**Store as**: `TODO_TASKS` list (e.g., ["TM-task-12", "TM-task-15"])
**If no TODO tasks**: Report and exit gracefully

---

## Phase 4.5: Read Attached Documents (MANDATORY)

**CRITICAL**: If the iteration has attached documents, you MUST read them before planning.

```bash
# List documents to get IDs
dw task-manager doc list --iteration $TARGET_ITERATION

# Read each document
dw task-manager doc show <doc-id>
```

**Why mandatory**:
- Documents contain planning details, ADRs, design decisions
- Planning without reading documents leads to misalignment
- Documents provide critical context for implementation

**If no documents**: Proceed to planning
**If documents exist**: Read ALL documents before Phase 5

---

## Phase 5: Launch Planning Agent

**Agent**: `general-purpose` (requires design and planning)

**Prompt structure**:

```
You are the planning agent for iteration [number/id].

## Step 1: Explore Codebase (if needed)

Use Task tool with Explore agent to understand relevant architecture:
- For new features: Explore similar implementations, patterns, domain models
- For refactoring: Explore code being modified, current architecture
- Thoroughness: "quick" (1-2 tasks) | "medium" (3-4 tasks) | "very thorough" (5+ or complex)

May run multiple explorations for different areas.

## Step 2: Retrieve Full Iteration Context

**CRITICAL**: You must independently retrieve ALL iteration data.

```bash
# Get iteration details with all tasks and attached documents
dw task-manager iteration show [iteration-number] --full

# Get ALL acceptance criteria for ALL tasks in iteration
dw task-manager ac list-iteration [iteration-number]
```

Parse output to understand:
- Iteration name, goal, deliverable
- ALL tasks (IDs, titles, descriptions, statuses)
- ALL acceptance criteria (descriptions, testing instructions)
- Attached documents (if any)

**If iteration has attached documents (MANDATORY)**:
```bash
# Read EACH attached document
dw task-manager doc show <doc-id>
```

**Why documents are critical**:
- Documents contain planning details, ADRs, architectural decisions
- Planning without reading documents leads to incorrect implementation
- Documents provide essential context that overrides assumptions

Parse document contents to understand:
- Design decisions and rationale
- Architecture patterns to follow
- Implementation constraints and guidelines
- Phase breakdown (if planning document exists)

## Step 3: Filter to TODO Tasks Only

**CRITICAL**: You are planning ONLY for tasks with status "todo".

TODO tasks in this iteration: [TODO_TASKS from Phase 4]

Ignore tasks with status: "in-progress", "review", "done", "blocked"

## Step 4: Create Implementation Plan

Using exploration findings + iteration context, create plan with:

1. **Phases**: Decompose TODO tasks into implementation phases (context-sized)
2. **Dependencies**: Identify phase dependencies
3. **Parallelization**: Mark which phases can run concurrently
4. **Architecture**: Follow DarwinFlow patterns (see @docs/arch-index.md, CLAUDE.md)
5. **Testing**: Each phase has verification steps (70-80% coverage target)

**Plan format**:

### Phase 1: [Name]
- Objective: [What it accomplishes]
- TODO tasks: [Task IDs from TODO_TASKS list]
- Requirements: [Specific checklist]
- Files: [Expected files]
- Testing: [Verification approach]
- Parallel with: [None / Phase X]

### Phase 2: [Name]
...

### Verification Phase
- All tests pass (go test ./...)
- Zero linter violations (go-arch-lint .)
- All TODO task AC verifiable by user
- Implementation matches plan

**Final report**: Number of phases, parallel opportunities, dependencies, complexity, risks
```

**Execute**:
```
Task tool:
- subagent_type: "general-purpose"
- description: "Plan TODO tasks implementation"
- prompt: [above, with TODO_TASKS list filled in]
```

**Review report**: Parse phases, note parallelization, create TodoWrite with all phases

---

## Phase 6: Execute Implementation Phases

**For each phase** (respect dependencies):

### Select Agent Type
- **junior-dev-executor**: Clear, well-specified requirements (even if complex)
- **general-purpose**: Research, exploration, design decisions

### Construct Prompt

```
Implement Phase [N]: [Name] for iteration [number]

## Context
- Iteration goal: [2-3 sentences]
- TODO tasks scope: [TODO_TASKS list]
- This phase addresses: [Task IDs for this phase from TODO_TASKS]

## Objective
[What this phase accomplishes]

## Requirements
[Checklist from planning agent]

## Acceptance Criteria Context
[Relevant AC from TODO tasks - for awareness, NOT for you to verify]

## Architecture Constraints
- Follow DarwinFlow structure (@docs/arch-index.md, CLAUDE.md)
- Maintain dependency rules
- Zero linter violations
- 70-80% test coverage

## Verification
After implementation:
- Run: go test ./...
- Run: go-arch-lint .
- Update task status if complete
- Note blockers/issues

## Deliverables
[What to create/modify]

**Report**: What implemented, files, test results, linter results, task updates, issues, recommendations
```

### Execute

**Sequential** (default):
```
Task tool: subagent_type, description, prompt
Wait for completion before next phase
```

**Parallel** (if planning agent identified):
```
Launch multiple Task tools in single message
Wait for ALL before dependent phases
```

### Review Phase Report

- Mark TodoWrite completed
- Note files modified
- Check tests/linter
- Update task statuses:
  ```bash
  dw task-manager task update <task-id> --status done  # If phase completes TODO task
  ```
- **If issues**: Don't proceed, report blocker, ask user

---

## Phase 7: Launch Verification Agent

**Agent**: `general-purpose` (requires analysis)

**Prompt structure**:

```
Verify iteration [number] implementation.

## Implementation Context
- Plan summary: [Brief phases summary]
- Phases completed: [Phase names]
- Files modified: [Key files]
- TODO tasks implemented: [TODO_TASKS from Phase 4]

**Scope**: This implementation focused ONLY on "todo" tasks. Other statuses unchanged.

## Step 1: Gather Current State

```bash
dw task-manager iteration show [number] --full
dw task-manager ac list-iteration [number]
```

Filter: Focus on TODO_TASKS, confirm now "done" (if fully implemented)

## Step 2: Verification Checklist

### Tests
- Run: go test ./...
- All pass, coverage 70-80%

### Linter
- Run: go-arch-lint .
- Zero violations

### Code Quality
- Read modified files
- Check clean architecture, error handling, no duplication

### Implementation vs Plan
- All phase objectives met
- No missing functionality

### Task Status
- All TODO tasks (from TODO_TASKS) marked "done"
- Other statuses unchanged

### AC Readiness
- Implementation enables user to verify each AC for TODO tasks
- Testing instructions clear
- Note: You do NOT verify AC (user does this)

**Report**: Test results, linter results, code quality, completeness, task status, AC readiness, overall assessment (ready: YES/NO), issues, recommendations
```

**Execute**:
```
Task tool:
- subagent_type: "general-purpose"
- description: "Verify iteration implementation"
- prompt: [above with TODO_TASKS filled in]
```

**Review report**: Check tests/linter, note issues

**If issues found**: Create fix phase, delegate, re-verify

---

## Phase 8: Documentation Check

- Functionality changed? → Check README.md
- Workflow changed? → Check CLAUDE.md
- Architecture changed? → Consider `go-arch-lint docs`

Update if needed (or note for user if major changes).

---

## Phase 9: Create Git Commit

**MANDATORY**: Commit all work after implementation + verification.

```bash
git status
git diff --stat
git add .
git commit -m "$(cat <<'EOF'
feat: implement iteration #N - [brief deliverable]

[1-2 sentence summary of what was implemented]
EOF
)"
git log -1 --oneline
```

**Guidelines**: Conventional commit prefix, include iteration number, use HEREDOC format
**Don't push**: User pushes after AC verification

---

## Phase 10: Final Report

**CRITICAL**: Follow CLAUDE.md reporting guidelines.

**Report deviations, questions, issues ONLY** - not what was implemented (user knows tasks).

```markdown
# Implementation Complete: [Iteration Name]

## Status
[Complete / Complete with deviations / Blocked]

## Deviations from Plan
[Only if deviations occurred - what and why]

## Questions for User
[Only if decisions needed - clear, actionable]

## Issues Requiring Attention
[Only if blockers or problems]

## Commit
[Commit hash and summary from git log -1 --oneline]
```

**DO**: Highlight deviations, questions, issues, blockers
**DON'T**: Summarize implementation, repeat tasks, list files, explain what should happen, add standard "verify AC" instructions

---

## Error Handling

**Planning fails**: Report to user, don't proceed, ask guidance
**Phase fails**: Don't proceed to dependents, report blocker, ask how to proceed
**Verification fails**: Create fix phase, resolve, re-verify
**Tests/linter fail**: Fix before final report

---

## Success Criteria

- ✅ TODO tasks filtered correctly
- ✅ Planning agent created plan for TODO tasks
- ✅ All phases completed
- ✅ Verification passed (tests, linter, quality)
- ✅ TODO tasks marked "done"
- ✅ Non-TODO tasks unchanged
- ✅ Git commit created
- ✅ User notified with clear report

**Not your responsibility**:
- ❌ Verifying acceptance criteria (user does this)
- ❌ Closing iteration (user does this)

---

## Key Principles

1. **Branch first** - Ensure on iteration branch before any work
2. **TODO only** - Focus exclusively on "todo" tasks
3. **Read documents first** - MANDATORY if iteration has attached documents (orchestrator AND planning agent)
4. **Orchestrate** - Coordinate agents, don't code yourself
5. **Plan first** - Planning agent reads docs + explores + retrieves full context independently
6. **Respect dependencies** - Sequential unless parallel identified
7. **Verify thoroughly** - Tests, linter, quality before commit
8. **Track progress** - TodoWrite and task status updates
9. **Commit always** - Save all work
10. **User owns acceptance** - Never verify AC or close iteration
11. **Report clearly** - Deviations/questions/issues only

---

Remember: First ensure iteration branch. Work ONLY on TODO tasks. READ ALL ATTACHED DOCUMENTS (orchestrator + planning agent). Planning agent independently retrieves all iteration data (tasks + AC + documents). Implementation agents execute. Verification agent checks. Commit all work. Report deviations/questions/issues to user. User verifies AC and closes iteration.

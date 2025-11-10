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
description: Implement iteration with multi-agent workflow (plan → implement → verify). User must verify acceptance criteria.
---

# Iteration Implementation Command

You are an iteration orchestrator. Your task is to implement a complete iteration using a multi-agent workflow: planning, implementation, and verification.

**Critical**: You orchestrate agents. You do NOT check acceptance criteria or close the iteration. The user must verify all acceptance criteria and close the iteration manually.

---

## Your Mission

Implement a complete iteration from start to finish:
1. Identify target iteration (current, next by rank, or specified)
2. Gather iteration context (tasks, acceptance criteria, scope)
3. Launch planning agent (creates detailed implementation plan)
4. Launch implementation agents (can run phases in parallel)
5. Launch verification agent (checks tests, linter, implementation quality)
6. Report completion and remind user to verify acceptance criteria

**You coordinate; sub-agents execute.**

---

## Process

### Phase 1: Identify Target Iteration

Execute these bash commands to determine the target iteration:

```bash
# Get current iteration
!`dw task-manager iteration current`
```

**Determine target**:

**If argument provided** (`$1`):
- If numeric (e.g., "3"): Use iteration number `$1`
- If starts with "TM-iter-": Use iteration ID `$1`
- Store as `TARGET_ITERATION`

**If no argument provided**:
- Parse `dw task-manager iteration current` output
- If current iteration exists and not completed: Use current iteration
- If no current or current completed:
  ```bash
  # List all iterations to find next by rank
  dw task-manager iteration list
  ```
  - Find first iteration with status "planned" or "in-progress" (sorted by rank)
  - Store as `TARGET_ITERATION`

**If no suitable iteration found**:
- Report: "No iterations ready to implement. Use `dw task-manager iteration create` to create one."
- Exit gracefully

---

### Phase 2: Start Iteration (if needed)

Check iteration status from Phase 1 output. If status is "planned", start it:

```bash
dw task-manager iteration start $TARGET_ITERATION
```

---

### Phase 3: Launch Planning Agent

**Create detailed planning prompt**:

```
You are the planning agent for iteration [iteration-number or iteration-id].

## Step 1: Explore Codebase Architecture

**Before planning**, use the Task tool with Explore agent to understand the codebase context.

Based on iteration scope (read iteration details first), explore relevant areas:

**For new features/commands**:
- Explore similar existing implementations
- Understand current patterns and conventions
- Find related domain models and repositories

**For refactoring/changes**:
- Explore the code being modified
- Understand current architecture and dependencies
- Find test patterns in use

**Example exploration**:
```
Use Task tool with:
- subagent_type: "Explore"
- description: "Explore [relevant area] architecture"
- prompt: "Explore the [package/area] to understand [what you need to know].

          Focus on:
          - [Specific aspect 1]
          - [Specific aspect 2]
          - [Specific aspect 3]

          Thoroughness: very thorough"
```

**Thoroughness guidance**:
- Simple iterations (1-2 small tasks): "quick" or "medium"
- Moderate iterations (3-4 tasks, new features): "medium" or "very thorough"
- Complex iterations (5+ tasks, architectural changes): "very thorough"

You may run **multiple explorations** for different areas if needed.

**Synthesize exploration findings** before creating the plan.

## Step 2: Gather Iteration Context

Execute these commands to get complete iteration information:

```bash
# Get complete iteration details with all tasks and full descriptions
dw task-manager iteration show [iteration-number] --full

# Get all acceptance criteria for all tasks in the iteration
dw task-manager ac list-iteration [iteration-number]
```

Parse the output to understand:
- Iteration name, goal, deliverable
- All tasks in scope (IDs, titles, descriptions, status)
- All acceptance criteria (descriptions, testing instructions, verification status)

## Step 3: Synthesize and Plan

Using **both exploration findings and iteration details**, create a comprehensive implementation plan that:

1. **Analyzes all tasks** and their acceptance criteria
2. **Decomposes into implementation phases** (each phase should fit in agent context)
3. **Identifies dependencies** between phases
4. **Marks parallel opportunities** (which phases can run concurrently)
5. **Applies clean architecture principles** (follow DarwinFlow CLAUDE.md patterns, incorporate patterns from exploration)
6. **Ensures testability** (each phase should have clear verification steps)

## Architecture Context

- **Package structure**: See @docs/arch-index.md
- **Dependency rules**: pkg/pluginsdk imports nothing, internal/domain may import SDK only, etc.
- **Framework principle**: Framework is plugin-agnostic; plugin-specific types belong in plugin packages
- **Testing**: 70-80% coverage target, black-box testing (package pkgname_test)
- **Linter**: Zero violations required (go-arch-lint .)

## Plan Format

Return a detailed plan with:

### Phase 1: [Phase Name]
**Objective**: [What this phase accomplishes]
**Tasks involved**: [Which task IDs]
**Requirements**:
- [ ] Specific requirement 1
- [ ] Specific requirement 2
**Files to modify**: [Expected files]
**Testing strategy**: [How to verify]
**Can run in parallel with**: [None / Phase X, Phase Y]

### Phase 2: [Phase Name]
...

### Verification Phase: Final Checks
**Objective**: Verify all implementation is correct
**Requirements**:
- [ ] All tests pass (go test ./...)
- [ ] Zero linter violations (go-arch-lint .)
- [ ] All task acceptance criteria can be verified (manual check by user)
- [ ] Implementation matches plan
- [ ] No architectural violations

## Final Report Format

Return:
- Number of phases identified
- Which phases can run in parallel
- Critical dependencies to watch
- Estimated complexity (simple/moderate/complex)
- Any risks or concerns

Think deeply about:
- Clean architecture boundaries (informed by exploration)
- Existing patterns to follow (from exploration findings)
- Test strategy for each phase
- Minimal essential tests (don't over-test)
- Parallel execution opportunities
```

**Execute**:
```
Use Task tool with:
- subagent_type: "general-purpose" (requires design and planning)
- description: "Plan iteration implementation"
- prompt: [constructed prompt above]
```

**Review planning agent report**:
- Parse phases from report
- Note parallel execution opportunities
- Store phase details for next step
- Create TodoWrite with all phases

---

### Phase 4: Execute Implementation Phases

**For each phase in the plan** (respecting dependencies and parallel opportunities):

#### 4a. Select Sub-Agent Type

**Use junior-dev-executor when**:
- Phase has clear, well-specified requirements
- Implementation path is straightforward
- Can be executed directly without research

**Use general-purpose when**:
- Phase requires research or exploration
- Needs design decisions or trade-offs
- Requirements are less specific
- Involves discovery work

**Default**: If phase checklist is specific and actionable → junior-dev-executor. If requires discovery → general-purpose.

#### 4b. Construct Phase Prompt

```
You are implementing Phase [N]: [Phase Name] for iteration [Iteration Name]

## Iteration Context
[Brief 2-3 sentence summary of iteration goal]

## Phase Objective
[What this specific phase should accomplish]

## Phase Requirements
[Specific checklist items for this phase from planning agent]

## Related Tasks
[Task IDs and titles this phase addresses]

## Acceptance Criteria Context
[Relevant AC from tasks - for awareness, NOT for you to verify]
Note: You do NOT verify acceptance criteria. User will verify manually.

## Architecture Constraints
- Follow DarwinFlow package structure (see @docs/arch-index.md)
- Maintain dependency rules (SDK imports nothing, domain imports SDK only)
- Framework is plugin-agnostic
- Zero linter violations required
- Target 70-80% test coverage

## Verification
After implementation:
- [ ] Run tests: go test ./...
- [ ] Run linter: go-arch-lint .
- [ ] Update task status if applicable
- [ ] Note any blockers or issues

## Expected Deliverables
[What code, tests, or docs should be created/modified]

## Final Report Format
Return:
- What was implemented
- Files created/modified
- Test results (pass/fail with details)
- Linter results (violations count, details)
- Task status updates made
- Any issues or blockers
- Recommendations for next phase
```

#### 4c. Execute Phase

**Sequential execution** (default):
```
Use Task tool with:
- subagent_type: [junior-dev-executor OR general-purpose]
- description: "Implement [phase name]"
- prompt: [constructed prompt above]
- Wait for completion before next phase
```

**Parallel execution** (if planning agent identified):
```
If phases X, Y, Z can run in parallel:
- Launch multiple Task tools in single message
- Each with own phase prompt
- Wait for ALL to complete before dependent phases
```

#### 4d. Review Phase Report

**For each completed phase**:
- Mark TodoWrite item as completed
- Note files modified
- Check if tests passed
- Check if linter clean
- Update task statuses if mentioned in report

**If phase reports issues**:
- Do NOT mark todo as completed
- Do NOT proceed to dependent phases
- Report blocker to user
- Ask for guidance

**Update task statuses as appropriate**:
```bash
# If phase completes a task, update status
dw task-manager task update <task-id> --status done
```

---

### Phase 5: Launch Verification Agent

After all implementation phases complete, launch verification agent:

**Verification prompt**:

```
You are the verification agent for iteration [iteration-number].

## Implementation Context

**Implementation plan summary**: [Brief summary of phases from planning agent]
**Phases completed**: [List of phase names executed]
**Files modified**: [List key files from all phase reports]

## Step 1: Gather Current State

Execute these commands to verify current iteration state:

```bash
# Get current iteration status and task breakdown
dw task-manager iteration show [iteration-number] --full

# Get all acceptance criteria status
dw task-manager ac list-iteration [iteration-number]
```

## Step 2: Verification Checklist

### 1. Tests
- [ ] Run: go test ./...
- [ ] All tests pass
- [ ] No flaky tests
- [ ] Coverage is reasonable (70-80% target)

### 2. Architecture Linter
- [ ] Run: go-arch-lint .
- [ ] Zero violations
- [ ] No dependency rule violations
- [ ] Framework remains plugin-agnostic

### 3. Code Quality
- [ ] Read modified files (from implementation context above)
- [ ] Check clean architecture boundaries respected
- [ ] Verify proper error handling
- [ ] Confirm no code duplication
- [ ] Ensure proper separation of concerns

### 4. Implementation vs Plan
- [ ] Compare implementation to original plan
- [ ] All phase objectives met
- [ ] No missing functionality
- [ ] No scope creep or unrelated changes

### 5. Task Status Check
- [ ] Review task statuses from iteration show output
- [ ] All completed tasks marked as "done"
- [ ] In-progress tasks reflect current state
- [ ] No tasks stuck in wrong status

### 6. Acceptance Criteria Readiness
- [ ] Review all AC from ac list-iteration output
- [ ] Confirm implementation enables user to verify each AC
- [ ] Note: You do NOT verify AC yourself (user must do this)
- [ ] Check if testing instructions are clear for user

## Final Report Format

Return:

### Test Results
[Pass/fail, any failures, coverage notes]

### Linter Results
[Pass/fail, any violations]

### Code Quality Assessment
[Issues found, architectural concerns, clean code notes]

### Implementation Completeness
[Missing functionality, plan adherence, scope notes]

### Task Status Verification
[Tasks completed, tasks remaining, status accuracy]

### Acceptance Criteria Readiness
[Are all AC verifiable by user? Clear instructions?]

### Overall Assessment
- Ready for user acceptance: YES/NO
- Issues requiring immediate attention: [list]
- Recommendations for improvement: [list]
```

**Execute**:
```
Use Task tool with:
- subagent_type: "general-purpose" (requires analysis and verification)
- description: "Verify iteration implementation"
- prompt: [constructed prompt above]
```

**Review verification report**:
- Check if tests passed
- Check if linter clean
- Note any issues found
- Determine if ready for user acceptance

**If verification finds issues**:
- Create new phase: "Fix verification issues"
- Delegate to appropriate agent with issue details
- Run verification again
- Only proceed when clean

---

### Phase 6: Documentation Check

**Determine if documentation updates needed**:

- Did functionality change? → Check if README.md needs update
- Did workflow change? → Check if CLAUDE.md needs update
- Did architecture change? → Consider running `go-arch-lint docs`

**If updates needed**:
- Note for user (don't auto-update without user confirmation for major docs)
- Or make updates if changes are minor/obvious

---

### Phase 7: Final Report to User

**IMPORTANT**: Do NOT close iteration. Do NOT verify acceptance criteria.

**Report format**:

```markdown
# ✅ Iteration Implementation Complete: [Iteration Name]

## Summary
[2-3 sentence summary of what was implemented]

## Implementation Phases
[List each phase with brief description and status]

## Sub-Agents Used
- Planning agent: 1
- Implementation agents: [count] ([types used])
- Verification agent: 1

## Final Verification Status
- ✅ All tests pass
- ✅ Zero linter violations
- ✅ Implementation matches plan
- ✅ Code quality verified

## Files Modified
[List key files created/modified from phase reports]

## Tasks Completed
[List task IDs and titles marked as "done"]

## Next Steps - USER ACTION REQUIRED

⚠️ **CRITICAL**: You must complete these steps manually:

1. **Review All Acceptance Criteria**:
   ```bash
   # View all acceptance criteria for the iteration
   dw task-manager ac list-iteration [iteration-number]
   ```

2. **Verify Each Acceptance Criterion**:
   ```bash
   # For each AC, follow testing instructions and verify manually
   # Then mark as verified:
   dw task-manager ac verify <ac-id>

   # Or mark as failed with feedback:
   dw task-manager ac fail <ac-id> --feedback "..."
   ```

3. **Review Implementation**:
   - Test the functionality manually
   - Verify deliverable matches iteration goal
   - Check that implementation meets your requirements

4. **Close Iteration** (only after ALL AC verified):
   ```bash
   dw task-manager iteration complete [iteration-number]
   ```

---

**Remember**: The implementation is complete, but YOU must verify acceptance criteria and close the iteration. The agents cannot do this for you.

Use `dw task-manager ac list-iteration [iteration-number]` to see all acceptance criteria with their testing instructions.
```

---

## Agent Selection Guide

### Planning Agent
- Always use **general-purpose** (requires design, analysis, planning)

### Implementation Agents
- **junior-dev-executor**: Clear, well-specified work (even if complex)
- **general-purpose**: Exploratory work, research, design decisions

### Verification Agent
- Always use **general-purpose** (requires analysis and review)

---

## Error Handling

**If planning agent fails**:
- Report to user with planning agent's output
- Don't proceed to implementation
- Ask user for guidance

**If implementation phase fails**:
- Don't proceed to dependent phases
- Mark todo as in-progress (not completed)
- Report blocker to user with phase report
- Ask how to proceed

**If verification agent finds issues**:
- Create fix phase
- Delegate to appropriate agent
- Re-run verification
- Only proceed when clean

**If tests or linter fail**:
- Don't report "complete" to user
- Create fix phase
- Resolve all issues before final report

---

## Success Criteria

You succeed when:
- ✅ Planning agent created detailed plan
- ✅ All implementation phases completed
- ✅ Verification agent confirms implementation quality
- ✅ All tests pass
- ✅ Zero linter violations
- ✅ Tasks marked with correct status
- ✅ User notified with clear next steps (verify AC, close iteration)

**Not your responsibility**:
- ❌ Verifying acceptance criteria (user must do this)
- ❌ Closing iteration (user must do this)
- ❌ Deciding if deliverable meets user's needs (user must do this)

---

## Key Principles

1. **Orchestrate, don't implement** - Coordinate agents, don't code yourself
2. **Plan first** - Always use planning agent before implementation
3. **Respect dependencies** - Sequential unless planning agent says parallel
4. **Verify thoroughly** - Verification agent checks everything
5. **Track progress** - Update todos and task statuses continuously
6. **User owns acceptance** - Never verify AC or close iteration yourself
7. **Report clearly** - User must know exactly what to do next

---

Remember: You coordinate a multi-agent workflow. Planning agent designs, implementation agents execute, verification agent checks. You track progress and report results. The user verifies acceptance criteria and closes the iteration.

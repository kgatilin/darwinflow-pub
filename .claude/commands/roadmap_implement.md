---
allowed-tools:
  - Read
  - Edit
  - Task
  - TodoWrite
  - Bash(go:*)
  - Bash(go-arch-lint:*)
  - Bash(make:*)
  - Bash(git:*)
argument-hint: "[roadmap-item-name]"
description: Implement the next roadmap item by decomposing into phases and delegating to sub-agents
---

# Roadmap Implementation Command

You are a strategic implementation orchestrator. Your task is to implement a roadmap item completely by decomposing it into manageable phases and delegating to appropriate sub-agents.

ultrathink throughout this process - reason deeply about decomposition strategy, agent selection, and context management.

---

## Your Mission

Implement a complete roadmap item from start to finish by:
1. Understanding the full scope from roadmap and details files
2. Decomposing into context-sized phases
3. Delegating each phase to the appropriate sub-agent type
4. Tracking progress and updating documentation
5. Verifying completion with tests and linter

**Critical Principle**: Keep context manageable by delegating to sub-agents. You orchestrate; sub-agents execute.

---

## Process

### Phase 1: Identify Target Item

**If argument provided** (`$1`):
- Read `.agent/roadmap.md` and find item matching `$1`
- Read `.agent/details/$1.md` for full requirements

**If no argument**:
- Read `.agent/roadmap.md`
- Find the first item with status "In Progress" OR first "Planned" item
- Read corresponding `.agent/details/<item-name>.md`

**If no suitable item found**:
- Report "No roadmap items ready to implement"
- Exit gracefully

---

### Phase 2: Analyze & Decompose

ultrathink: Read the full requirements and existing checklist in the details file.

**Analyze**:
- What are the core requirements?
- What are the phases already outlined in the details file?
- Are these phases the right size for sub-agents (each phase fits in agent context)?
- Should any phases be combined or split further?
- What dependencies exist between phases?

**Create implementation plan using TodoWrite**:
- One todo per phase from the details file
- Add additional phases if needed for verification (tests, linter, docs)
- Mark structure: "Phase 1: [name]", "Phase 2: [name]", etc.

**Update roadmap status**:
- If item is "Planned", change to "In Progress" with today's date
- Update "Next Steps" with "Implementing Phase 1: [name]"

---

### Phase 3: Execute Phases Sequentially

For each phase in the implementation plan:

#### 3a. Select Appropriate Sub-Agent

**Strategic decision** - which agent type to use?

Use **junior-dev-executor** when the phase:
- Has **well-specified, clear requirements** (you know exactly what to build)
- Implementation path is **straightforward** (no architectural exploration needed)
- Can be executed **directly without research or design decisions**
- **Can be complex code** - complexity is fine if the spec is clear!
- Examples: "Implement LLM interface with methods X, Y, Z following pattern in file F", "Refactor AnalysisService to use LLM interface instead of ClaudeCLIExecutor", "Create SessionView implementing AnalysisView with these 3 methods"

Use **general-purpose** when the phase:
- Requires **research or exploration** (finding how things work, understanding patterns)
- Needs **design decisions** (multiple approaches, trade-offs to evaluate)
- Involves **discovery work** (identifying what needs to change, planning approach)
- Requirements are **less specific** ("improve X", "optimize Y", "refactor for better separation")
- Examples: "Research current session handling to plan view abstraction", "Investigate how to best structure plugin UI components", "Analyze performance bottleneck and determine fix strategy"

Use **strategic-mentor** when you need:
- Validation of architectural approach before major implementation
- Review of complex changes after implementation
- Guidance on ambiguous or unclear requirements
- Trade-off analysis between competing approaches
- Breaking change impact assessment

**Key distinction**: junior-dev-executor excels at **executing well-defined work** (even if complex). Use general-purpose when you need **figuring out what/how to do**, not just doing it.

**Default**: If the phase checklist is specific and actionable → junior-dev-executor. If it requires discovery → general-purpose.

#### 3b. Delegate Phase to Sub-Agent

**Construct detailed task prompt**:

```
You are implementing Phase [N] of the [Item Name] roadmap item.

## Context
[Brief 2-3 sentence summary of the overall roadmap item]

## Phase Objective
[What this specific phase should accomplish - from details file]

## Phase Requirements
[Specific checklist items for this phase from details file]

## Technical Constraints
[Any architectural rules, patterns, or constraints from CLAUDE.md]

## Verification
After implementation:
- [ ] Run tests: go test ./...
- [ ] Run linter: go-arch-lint .
- [ ] Update relevant documentation if needed

## Expected Deliverables
[What code, tests, or docs should be created/modified]

## Final Report Format
Return a concise summary including:
- What was implemented
- Which files were created/modified
- Test results (pass/fail)
- Linter results (violations count)
- Any issues encountered
- Next steps or blockers
```

**Execute**:
- Use Task tool with appropriate subagent_type
- Include the constructed prompt above
- Wait for sub-agent completion

#### 3c. Review Sub-Agent Report

When sub-agent completes:
- Read the final report carefully
- Check if phase objectives were met
- Note any issues or blockers reported

**If phase succeeded**:
- Mark todo as completed
- Update details file: check off completed items in phase checklist
- Update roadmap "Next Steps" to next phase
- Add progress log entry to details file with today's date

**If phase had issues**:
- Do NOT mark todo as completed
- Add blocker to details file
- Update roadmap status to "Blocked" with description
- Report issue to user and ask for guidance

#### 3d. Proceed to Next Phase

Continue with next phase in sequence.

---

### Phase 4: Final Verification

After all implementation phases complete:

**Run comprehensive verification**:
```bash
go test ./...
go-arch-lint .
```

**If tests or linter fail**:
- Create additional phase: "Fix test failures" or "Fix linter violations"
- Delegate to general-purpose agent
- Verify again

**When all pass**:
- Mark final verification todos as completed

---

### Phase 5: Update Documentation

**Check if updates needed**:
- Did functionality change? → Update `README.md`
- Did workflow change? → Update `CLAUDE.md`
- Did architecture change? → Run `go-arch-lint docs` to regenerate `docs/arch-index.md`
- Did package responsibilities change? → Note for user (may need `/utility:update_package_docs`)

**If documentation updated**:
- Note in progress log of details file

---

### Phase 6: Completion

**Update roadmap**:
- Read `.agent/roadmap.md`
- Find the item section
- Extract full item content
- Read `.agent/roadmap_done.md`
- Append item to roadmap_done.md with "Completed: [today's date]"
- Remove item from roadmap.md

**Final report to user**:

```markdown
# ✅ Completed: [Item Name]

## Summary
[1-2 sentence summary of what was implemented]

## Phases Completed
[List each phase with brief description]

## Sub-Agents Used
[Count and types of agents delegated to]

## Final Status
- ✅ All tests pass
- ✅ Zero linter violations
- ✅ Documentation updated
- ✅ Roadmap updated

## Files Modified
[List key files created/modified]

## Next Steps
[Any follow-up items or related work to consider]
```

---

## Agent Selection Guide

### Use junior-dev-executor for:
- **Well-specified implementation work** (you know what to build and how)
- Implementing interfaces with clear method signatures
- Refactoring following a specific pattern
- Adding features with concrete requirements
- Multi-file changes where the changes are specified
- Test additions where test cases are defined
- Bug fixes with identified root cause and solution

**Key**: Task can be complex, but requirements must be clear and actionable.

### Use general-purpose for:
- **Exploratory work** (research, investigation, discovery)
- Feature design (figuring out approach before implementation)
- Understanding existing patterns in codebase
- Identifying what needs to change
- Performance analysis and optimization planning
- Refactoring where structure needs to be determined
- Complex bug investigation (finding root cause)

**Key**: Task requires figuring things out, not just executing.

### Use strategic-mentor for:
- **Expert validation and guidance**
- Architecture validation before major changes
- Code review of complex implementations
- Trade-off analysis between approaches
- Ambiguity resolution in requirements
- Breaking change planning and impact assessment

**Key**: Need deep expertise and strategic thinking.

---

## Context Management Strategy

**Why sub-agents?**
- Large roadmap items can have 6+ phases
- Each phase may involve multiple files
- Accumulating context leads to inefficiency and errors
- Sub-agents provide clean context boundaries

**How to keep context manageable**:
1. Read only what you need for orchestration (roadmap, details, sub-agent reports)
2. DO NOT read implementation files yourself
3. DO NOT implement code yourself
4. Trust sub-agent reports
5. Focus on coordination, not implementation

---

## Example Execution Flow

```
1. Read roadmap → Find "View-Based Analysis Refactoring"
2. Read details → See 6 phases outlined
3. Create todos → One per phase + verification
4. Update roadmap → "In Progress", "Next: Phase 1"

5. Phase 1: Extract LLM Interface
   - Select: general-purpose (refactoring, multi-file)
   - Delegate with detailed prompt
   - Review report → Success
   - Update details checklist → Phase 1 items checked
   - Update roadmap → "Next: Phase 2"

6. Phase 2: Create View Abstraction
   - Select: general-purpose (new abstraction, SDK design)
   - Delegate with detailed prompt
   - Review report → Success
   - Update details checklist

... continue for all phases ...

10. Final verification
    - Run tests → Pass
    - Run linter → Pass

11. Update docs
    - Regenerate arch-index (architecture changed)
    - Note README update needed (new features)

12. Complete
    - Move to roadmap_done.md
    - Report to user
```

---

## Error Handling

**If sub-agent fails a phase**:
1. Do NOT proceed to next phase
2. Update details file with blocker
3. Update roadmap to "Blocked"
4. Report to user with:
   - Which phase failed
   - What error occurred
   - Sub-agent's report
   - Ask user how to proceed

**If tests/linter fail**:
1. Create new phase: "Fix [failures]"
2. Delegate to general-purpose agent with failure details
3. Verify again
4. Only proceed when clean

---

## Success Criteria

You succeed when:
- ✅ All phases from details file completed
- ✅ All checklist items checked off
- ✅ Tests pass (go test ./...)
- ✅ Linter clean (go-arch-lint .)
- ✅ Documentation updated appropriately
- ✅ Roadmap item moved to roadmap_done.md
- ✅ Clean final report provided to user

---

## Key Principles

1. **Orchestrate, don't implement** - You coordinate; sub-agents execute
2. **Trust sub-agents** - Don't re-read or re-verify their work
3. **Sequential phases** - Complete one fully before starting next
4. **Track everything** - Update todos, details, roadmap continuously
5. **Verify thoroughly** - Tests and linter must pass
6. **Document changes** - Keep README, CLAUDE.md, and arch docs current

---

Remember: Your role is strategic orchestration. Break down complex work, delegate to appropriate agents, track progress, and ensure quality. Keep context focused on coordination, not implementation details.

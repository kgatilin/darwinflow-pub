# Evolution Roadmap

**Author**: Bob Martin
**Date**: 2025-10-09
**Status**: Target Architecture - Evolution Strategy
**Version**: 1.0

---

## Purpose

This document shows how DarwinFlow evolves from a simple single-node MVP to a sophisticated multi-workflow system WITHOUT rewriting the core.

**Key principle**: Grow through addition, not modification.

---

## Evolution Philosophy

### The Right Way to Grow

```
❌ WRONG: Predict complexity upfront
   - Design 20 node types we might need
   - Build complex routing before it's needed
   - Add features "just in case"
   - Result: Overengineered, never-used code

✅ RIGHT: Let usage drive complexity
   - Start with one node type
   - Use the system for real work
   - Add features when pain is felt
   - Result: Lean, battle-tested code
```

### Growth Pattern

```
1. Ship minimal version
2. Use it for real work (dogfooding)
3. Feel the pain points
4. Add ONE feature to address pain
5. Verify it helps
6. Repeat
```

**This is how Linux grew. This is how Unix grew. This is how great software grows.**

---

## Phase 1: MVP (Weeks 1-4)

### Capabilities

**Single node type**: `CLAUDE_CODE`

```yaml
# workflows/assistant_v1.yaml
id: assistant
version: 1
nodes:
  - id: call_claude
    type: CLAUDE_CODE
    config:
      pass_full_request: true
edges:
  - from: START
    to: call_claude
  - from: call_claude
    to: END
```

**That's it.** One node. Barely better than calling Claude Code directly.

### What Works

✓ Execute workflow (basically a wrapper around Claude Code)
✓ Log everything to files
✓ Track metrics (duration, success rate)
✓ Manual reflection command
✓ LLM analyzes logs and generates v2

### What Doesn't Work Yet

✗ No branching (single path only)
✗ No multiple node types
✗ No conditional logic
✗ No parallel execution
✗ Queries are slow (scan all files)

### Code Size

~500-1000 LOC for core framework

**Modules**:
```
darwinflow/
  ├── core/
  │   ├── workflow.go          # Domain entities
  │   ├── executor.go          # Workflow executor
  │   ├── interfaces.go        # All interfaces
  │   └── registry.go          # Node registry
  ├── storage/
  │   ├── file_workflow.go     # File workflow repo
  │   ├── file_log.go          # File log repo
  │   └── file_metrics.go      # File metrics repo
  ├── integrations/
  │   └── claude_code.go       # ClaudeCode node
  ├── reflection/
  │   ├── analyzer.go          # Pattern analyzer
  │   └── generator.go         # Workflow generator
  └── cmd/
      └── darwinflow/
          └── main.go          # CLI entry point
```

### Success Criteria

1. Can execute workflow and see it work
2. 20+ executions logged to files
3. Can query logs: "Show me what happened in run X"
4. Reflection creates v2 workflow based on patterns
5. v2 shows measurable improvement (even if small)

### Timeline

- Week 1: Core interfaces + file storage + basic executor
- Week 2: ClaudeCode node + logging
- Week 3: Metrics tracking + reflection
- Week 4: Testing + dogfooding + first v2 generation

### Learnings Expected

- How logging should work
- What metrics actually matter
- How reflection prompt should be structured
- What patterns are actually learnable

---

## Phase 2: Multi-Node Workflows (Months 2-3)

### What We've Learned (from Phase 1)

From dogfooding, we discover:

- "Always reads vision.md first" → need FILE_READ node
- "Makes decisions about feature type" → need LLM_DECISION node
- "Runs tests after code changes" → need TOOL_CALL node
- "Sometimes asks user for clarification" → need HUMAN_INPUT node

**Notice**: We didn't predict these. Usage revealed them.

### New Capabilities

**5-10 node types**:

```
CLAUDE_CODE      (already exists)
FILE_READ        (read file contents)
FILE_WRITE       (write file contents)
LLM_DECISION     (make decision via LLM)
TOOL_CALL        (execute external tool)
HUMAN_INPUT      (wait for user response)
```

**Conditional edges**:

```yaml
# workflows/assistant_v3.yaml
nodes:
  - id: read_vision
    type: FILE_READ
    config:
      path: docs/product/vision.md

  - id: classify_feature
    type: LLM_DECISION
    config:
      prompt: "Is this Core, Integration, or Workflow?"
      options: ["Core", "Integration", "Workflow"]
      context: [read_vision.output, user_request]

  - id: core_workflow
    type: CLAUDE_CODE
    config:
      context: "This is a Core framework feature"

  - id: integration_workflow
    type: CLAUDE_CODE
    config:
      context: "This is an Integration layer feature"

edges:
  - from: START
    to: read_vision
  - from: read_vision
    to: classify_feature
  - from: classify_feature
    to: core_workflow
    condition: "decision == 'Core'"
  - from: classify_feature
    to: integration_workflow
    condition: "decision == 'Integration'"
  # ...
```

### What Changes

**Core executor enhancements**:
- Add condition evaluation logic (selectNextNode method)
- Evaluate edge conditions to choose next node
- Support conditional branching based on node results

**New node implementations**:
- FILE_READ node
- FILE_WRITE node
- LLM_DECISION node
- TOOL_CALL node
- HUMAN_INPUT node

All additions - no modifications to existing code.

### What Doesn't Change

✓ Core interfaces (same as Phase 1)
✓ Storage implementations (same files/structure)
✓ Workflow loading (just more node types registered)
✓ Logging (same interface, logs more node types)

### Code Size

~2000-3000 LOC total (doubled, but still small)

### Success Criteria

1. Can create workflows with 5+ nodes
2. Conditional branching works
3. LLM decisions are accurate (>80%)
4. Reflection generates more sophisticated workflows
5. Measurable improvement: 30% fewer corrections

### Timeline

- Week 5-6: Add 3 new node types (FILE_READ, FILE_WRITE, LLM_DECISION)
- Week 7-8: Add conditional edge evaluation
- Week 9-10: Add TOOL_CALL and HUMAN_INPUT nodes
- Week 11-12: Dogfooding + refinement

### Learnings Expected

- Which node types are actually useful
- How to structure conditional logic
- How LLM decisions perform in practice
- What context is needed between nodes

---

## Phase 3: Advanced Execution (Months 4-6)

### What We've Learned (from Phase 2)

From extended usage:

- "Some nodes can run in parallel" → need parallel execution
- "Want to reuse workflow sections" → need sub-workflows
- "SQLite would speed up log queries" → need better storage
- "Need better error handling" → need retry/fallback

### New Capabilities

**Parallel execution**:

```yaml
nodes:
  - id: read_vision
    type: FILE_READ
    config:
      path: docs/product/vision.md

  - id: read_mvp
    type: FILE_READ
    config:
      path: docs/mvp_simple.md

  - id: combine_context
    type: CLAUDE_CODE
    config:
      context: [read_vision.output, read_mvp.output]

edges:
  - from: START
    to: [read_vision, read_mvp]  # PARALLEL
  - from: [read_vision, read_mvp]
    to: combine_context  # Wait for both
```

**Sub-workflows**:

```yaml
nodes:
  - id: check_tests
    type: SUB_WORKFLOW
    config:
      workflow: workflows/check_tests_v1.yaml
      inputs:
        project_path: "."
```

**Retry/fallback**:

```yaml
nodes:
  - id: call_api
    type: TOOL_CALL
    config:
      command: "curl https://api.example.com"
      retry:
        max_attempts: 3
        backoff: exponential
      fallback:
        node_id: use_cached_data
```

**SQLite storage** (optional):

```yaml
# config.yaml
storage:
  logs:
    type: sqlite
    path: darwinflow.db
```

### What Changes

**Core executor enhancements**:
- Detect multiple outgoing edges (parallel execution)
- Execute parallel nodes concurrently
- Collect and merge parallel results
- Wait for all parallel branches to complete

**New capabilities**:
- Sub-workflow node (compose workflows)
- Retry decorator (handle failures gracefully)
- SQLite storage implementations (faster queries)

All additions - core interfaces remain unchanged.

### What Doesn't Change

✓ Core interfaces (still the same!)
✓ Basic node types (FILE_READ, etc. unchanged)
✓ Workflow YAML format (just new fields)
✓ Logging interface (same contract)

### Code Size

~4000-5000 LOC total

### Success Criteria

1. Parallel execution works and is faster
2. Sub-workflows compose correctly
3. Retry logic handles failures gracefully
4. SQLite queries are 10x+ faster than files
5. 50% reduction in corrections from v1

### Timeline

- Month 4: Parallel execution + sub-workflows
- Month 5: Retry/fallback + SQLite storage
- Month 6: Dogfooding + optimization

### Learnings Expected

- Performance characteristics of parallel execution
- How to handle errors in complex workflows
- When to use files vs SQLite
- How workflows compose in practice

---

## Phase 4: Mature System (Months 7-12)

### What We've Learned (from Phase 3)

From production usage:

- "Need multiple LLM providers" → OpenAI, Anthropic, local
- "Want to share workflows" → export/import system
- "Need automatic learning" → real-time pattern detection
- "Community wants to contribute" → plugin system

### New Capabilities

**Multiple LLM providers**:

```yaml
nodes:
  - id: classify
    type: LLM_DECISION
    config:
      provider: anthropic  # or openai, local, etc.
      model: claude-3-5-sonnet
```

**Community plugins**:

```
darwinflow-integrations/
  ├── official/
  │   ├── claude-code/
  │   └── telegram/
  └── community/
      ├── jira/
      ├── slack/
      └── notion/
```

**Real-time learning**:

```go
// During execution, learn patterns automatically
func (e *Executor) Execute(workflow Workflow, input Input) {
    // ... execute nodes ...

    // NEW: Real-time pattern detection
    if pattern := e.detectPattern(executionLog); pattern != nil {
        e.suggestWorkflowChange(pattern)
    }
}
```

**Workflow marketplace**:

```bash
# Share workflow
darwinflow publish workflows/assistant_v5.yaml

# Import someone else's workflow
darwinflow import darwinflow://community/pm-workflow
```

**Analytics dashboard** (optional):

```
View at: http://localhost:8080/dashboard

- Execution trends
- Token usage over time
- Success rates by workflow version
- Most common patterns
```

### What Changes

**New abstractions**:
- **LLM provider layer**: Support multiple LLM backends (Anthropic, OpenAI, local)
- **Plugin system**: Load community-contributed node types from shared libraries
- **Real-time learning**: Detect patterns during execution, suggest improvements online

**Architecture stays the same**:
- Providers implement LLMProvider interface
- Plugins implement NodeExecutor interface
- Learning engine extends PatternAnalyzer interface

### What Doesn't Change

✓ Core interfaces (STILL the same!)
✓ Existing node types (unchanged)
✓ Storage abstraction (add backends, don't change interface)
✓ Workflow YAML format (extended, not modified)

### Code Size

~8000-10000 LOC (core + integrations + tools)

### Success Criteria

1. Multiple LLM providers work seamlessly
2. Community contributes 10+ integrations
3. Real-time learning reduces corrections during execution
4. 70% token reduction from v1
5. Workflows shared and reused across users

### Timeline

- Month 7-8: LLM provider abstraction
- Month 9-10: Plugin system + community integrations
- Month 11-12: Real-time learning + workflow marketplace

### Learnings Expected

- Which LLM providers work best for which tasks
- What makes workflows reusable across users
- How to balance real-time vs batch learning
- What governance is needed for community plugins

---

## Key Architectural Invariants

Throughout all phases, these NEVER change:

### 1. Core Interfaces

**WorkflowExecutor** and **NodeExecutor** interfaces remain IDENTICAL across all phases (1-4).

The contract is stable. Only implementations evolve.

### 2. Dependency Rule

```
Integrations  →  Core Interfaces  →  Domain Entities
  (changes)         (stable)           (stable)
```

Integrations can change, add, remove. Core stays stable.

### 3. Extension Points

Adding capabilities is always:

1. **New node type** → Implement `NodeExecutor` + register
2. **New storage** → Implement `*Repository` interface
3. **New workflow** → Write YAML, no code changes

### 4. Testing Strategy

Tests written in Phase 1 continue to pass in Phase 4 without modification.

Because interfaces don't change, existing tests remain valid as new capabilities are added.

---

## Capability Matrix

| Capability | Phase 1 | Phase 2 | Phase 3 | Phase 4 |
|------------|---------|---------|---------|---------|
| **Node Types** | 1 | 6 | 10 | 20+ |
| **Conditional Logic** | ❌ | ✅ | ✅ | ✅ |
| **Parallel Execution** | ❌ | ❌ | ✅ | ✅ |
| **Sub-workflows** | ❌ | ❌ | ✅ | ✅ |
| **Storage Options** | Files | Files | Files, SQLite | Files, SQLite, Postgres |
| **LLM Providers** | Claude | Claude | Claude | Claude, OpenAI, Local |
| **Learning Mode** | Batch | Batch | Batch | Batch, Real-time |
| **Community Plugins** | ❌ | ❌ | ❌ | ✅ |
| **Code Size (LOC)** | ~1K | ~3K | ~5K | ~10K |
| **Token Reduction** | 0% | 30% | 50% | 70% |

---

## Risk Management

### Risk: Feature Creep

**Mitigation**: Every feature must solve a felt pain from dogfooding.

Before adding:
1. Document the pain point
2. Show evidence from usage logs
3. Demonstrate this is the simplest solution
4. Ship it, measure impact

### Risk: Premature Optimization

**Mitigation**: Start with simple, dumb implementations.

Files are "slow"? So what - optimize when it hurts:
- 100 logs: Files fine
- 1,000 logs: Files still ok
- 10,000 logs: Consider SQLite
- 100,000 logs: Need Postgres

### Risk: Abstraction Failure

**Mitigation**: Test with multiple implementations early.

In Phase 1:
- Implement file storage
- Write SQLite implementation (even if not used)
- Verify both work with same tests

This proves abstraction isn't leaky.

### Risk: Community Plugins Break Things

**Mitigation**: Strong interface contracts + sandboxing.

- Plugins must implement stable NodeExecutor interface
- Optional: Run plugins in sandbox that enforces timeouts, resource limits
- Validate plugin config before execution

---

## Success Metrics Evolution

### Phase 1
- ✅ System executes workflows
- ✅ 20+ executions logged
- ✅ Reflection generates v2

### Phase 2
- ✅ 5+ node types working
- ✅ Conditional flows work
- ✅ 30% reduction in corrections

### Phase 3
- ✅ Parallel execution faster
- ✅ Sub-workflows compose
- ✅ 50% reduction in corrections
- ✅ SQLite 10x faster for queries

### Phase 4
- ✅ Multiple LLM providers
- ✅ Community contributes integrations
- ✅ 70% reduction in corrections
- ✅ Workflows reused across users

---

## When to Stop Growing

DarwinFlow reaches maturity when:

1. **Core interfaces haven't changed in 6+ months**
2. **New features are all integrations/workflows, not core**
3. **Community contributing more than core team**
4. **Token reduction plateaus** (diminishing returns)

At that point, focus shifts from framework to ecosystem.

---

## Anti-Pattern: The Rewrite Trap

```
❌ WRONG PATH:
Phase 1 → Phase 2 → "Let's rewrite with better architecture"
   ↓
Phase 2.5 (rewrite) → Never ships
   ↓
Abandoned project

✅ RIGHT PATH:
Phase 1 → Phase 2 → Phase 3 → Phase 4
   ↓         ↓         ↓         ↓
 Ship     Ship      Ship      Ship
```

**Rule**: If you're rewriting, your architecture was wrong from the start.

**This architecture avoids rewrites** through:
- Stable interfaces (nothing breaks)
- Dependency inversion (swap implementations)
- Plugin system (add without modify)
- Clean boundaries (change one layer at a time)

---

## Evolutionary Architecture Checklist

Before claiming "the architecture supports evolution", verify:

- [ ] Can add new node type in <100 LOC
- [ ] Can swap storage without changing core
- [ ] Can add new workflow without code changes
- [ ] Tests from Phase 1 still pass in Phase 4
- [ ] New feature doesn't modify existing interfaces
- [ ] Can deploy changes without breaking existing workflows

If ANY of these fail, architecture needs work.

---

## Conclusion

Evolution isn't about predicting the future. It's about:

1. **Building stable foundations** (interfaces that don't change)
2. **Starting simple** (one node type)
3. **Feeling real pain** (dogfooding)
4. **Adding carefully** (one feature at a time)
5. **Measuring impact** (did it help?)
6. **Never rewriting** (grow through addition)

This is how you build software that lasts 10+ years.

This is clean architecture in practice.

---

**You've now read**:
- [00_architecture_overview.md](00_architecture_overview.md) - Principles and boundaries
- [01_core_interfaces.md](01_core_interfaces.md) - Stable contracts
- [02_integration_architecture.md](02_integration_architecture.md) - Plugin system
- [03_storage_architecture.md](03_storage_architecture.md) - Persistence abstraction
- [04_evolution_roadmap.md](04_evolution_roadmap.md) - This document

**You're ready to build DarwinFlow.**

Start with Phase 1. Ship it. Use it. Let it evolve.

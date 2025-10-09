# DarwinFlow Target Architecture

**Author**: Bob Martin (well, his spirit anyway)
**Date**: 2025-10-09
**Status**: Target Architecture
**Version**: 1.0

---

## Executive Summary

DarwinFlow is a learning workflow framework that gets better through use. This document defines the target architecture that enables progressive evolution from a simple single-node MVP to a sophisticated multi-workflow system, without rewriting the foundations.

The architecture is built on three principles:
1. **Dependency inversion at every boundary** - interfaces own the architecture
2. **Separate what changes from what doesn't** - isolate volatility
3. **Start simple, grow organically** - complexity emerges from usage, not prediction

## Architectural Principles

### 1. The Dependency Rule

Dependencies flow inward, toward business logic:

```
┌─────────────────────────────────────┐
│   External Tools & Integrations     │  ← Adapters (Claude Code, Telegram, etc.)
│   (Most volatile, changes often)    │
└─────────────────────────────────────┘
            ↓ depends on
┌─────────────────────────────────────┐
│   Core Framework Interfaces         │  ← Contracts (WorkflowEngine, Storage, etc.)
│   (Stable, changes rarely)          │
└─────────────────────────────────────┘
            ↓ depends on
┌─────────────────────────────────────┐
│   Domain Entities & Business Rules  │  ← Pure logic (Workflow, Node, Pattern, etc.)
│   (Most stable, almost never)       │
└─────────────────────────────────────┘
```

**Key insight**: Reference workflows (domain-specific) live OUTSIDE core. They compose primitives but don't pollute the framework.

### 2. Interface Segregation

Each component depends only on interfaces it actually uses. No fat interfaces.

**Bad**: Single kitchen-sink interface with Execute(), Log(), Analyze(), Reflect(), Store(), Load() - violates ISP

**Good**: Separate focused interfaces:
- `WorkflowExecutor` - executes workflows
- `ExecutionLogger` - logs steps
- `PatternAnalyzer` - finds patterns

Each interface does ONE thing. Clients compose what they need.

### 3. Open-Closed Principle

The system must be open for extension (new node types, storage backends, integrations) but closed for modification (core framework never changes when adding nodes).

This is achieved through:
- **Plugin architecture** for node types
- **Strategy pattern** for storage backends
- **Adapter pattern** for external integrations

### 4. Single Responsibility at Architectural Level

Each layer has ONE reason to change:

| Layer | Responsibility | Changes When |
|-------|---------------|--------------|
| **Core Framework** | Workflow execution, state management, pattern learning | Learning algorithms improve, execution model evolves |
| **Integrations** | Tool adapters, external service connectors | Tool APIs change, new tools added |
| **Reference Workflows** | Domain-specific orchestration | Domain practices change, new patterns discovered |
| **Storage** | Persistence, retrieval | Storage technology changes (file → DB) |
| **Logging** | Observation, metrics | What we observe changes |

### 5. Liskov Substitution

Any storage implementation can replace another. Any node type can replace another. Any LLM provider can replace another.

This means:
- Interfaces define contracts with preconditions and postconditions
- Implementations honor these contracts without surprises
- Tests can verify substitutability

## Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    REFERENCE WORKFLOWS                           │
│  (Domain-specific orchestration - lives outside core)            │
│                                                                  │
│  • software_engineer_workflow.yaml                              │
│  • product_manager_workflow.yaml                                │
│  • architecture_review_workflow.yaml                            │
│                                                                  │
│  Compose: Core primitives + Integrations                        │
└─────────────────────────────────────────────────────────────────┘
                            ↓ uses
┌─────────────────────────────────────────────────────────────────┐
│                        INTEGRATIONS                              │
│  (Reusable tools - pluggable adapters)                          │
│                                                                  │
│  • ClaudeCodeAdapter (wraps Claude Code CLI)                    │
│  • TelegramBotAdapter (message I/O)                             │
│  • GitHubAdapter (issues, PRs, repos)                           │
│  • FileSystemAdapter (read/write operations)                    │
│  • LLMProviderAdapter (OpenAI, Anthropic, local)                │
│                                                                  │
│  Implements: NodeExecutor interface from Core                   │
└─────────────────────────────────────────────────────────────────┘
                            ↓ depends on
┌─────────────────────────────────────────────────────────────────┐
│                    CORE FRAMEWORK                                │
│  (Domain-agnostic engine - pure business logic)                 │
│                                                                  │
│  • WorkflowEngine (execution, state management)                 │
│  • PatternAnalyzer (learning from logs)                         │
│  • ReflectionEngine (generates new workflows)                   │
│  • NodeRegistry (plugin system for node types)                  │
│  • Storage abstractions (logs, metrics, workflows)              │
│                                                                  │
│  Defines: All interfaces, enforces rules                        │
└─────────────────────────────────────────────────────────────────┘
```

### Decision Criteria Per Layer

**Core Framework**: "Would ALL workflows benefit from this?"
- Yes → Core
- No → Keep it out

**Integration**: "Can multiple workflow types use this tool?"
- Yes → Integration
- No → Might be workflow-specific config

**Reference Workflow**: "Is this domain-specific orchestration?"
- Yes → Reference Workflow
- No → Reconsider which layer it belongs in

## Key Interfaces (High-Level)

See detailed interface definitions in `01_core_interfaces.md`, `02_integration_architecture.md`, and `03_storage_interfaces.md`.

### Core Framework Interfaces

**Domain entities**: Workflow, Node, Edge, ExecutionState

**Execution**: WorkflowExecutor, NodeExecutor

**Storage**: WorkflowRepository, LogRepository, MetricsRepository

**Learning**: PatternAnalyzer, ReflectionEngine, WorkflowGenerator

### Integration Layer

**Node implementations**: ClaudeCodeNode, LLMDecisionNode, FileReadNode

**Tool adapters**: TelegramAdapter, GitHubAdapter

### Reference Workflows

Workflows are data (YAML), not code - they drive the engine without being compiled into it.

## Dependency Flow

```
┌─────────────────┐
│ Reference       │
│ Workflows       │ ── uses ──→ Integration Adapters
│ (.yaml files)   │                    ↓
└─────────────────┘              implements
                                       ↓
                              NodeExecutor interface
                                       ↓
                              ┌─────────────────┐
                              │  Core Framework │
                              │   Interfaces    │
                              └─────────────────┘
                                       ↑
                              ┌─────────────────┐
                              │ Storage Impls   │
                              │ (File, DB, etc) │
                              └─────────────────┘
```

**Notice**:
- Workflows don't import code - they're data
- Integrations implement core interfaces - dependency points inward
- Storage implementations also point inward
- Core defines contracts, everyone else implements

## Progressive Evolution Strategy

The architecture enables three growth phases without rewrites:

### Phase 1: MVP (Single Node, Files)

**Characteristics**:
- One node type: `CLAUDE_CODE`
- File-based storage (logs, metrics, workflows)
- Manual reflection trigger
- Simple linear execution (no branching)

**What exists**:
- Core interfaces (defined but simple implementations)
- File storage implementation
- ClaudeCode adapter
- Basic workflow executor
- LLM-based reflection

**Complexity**: ~500-1000 LOC for core

### Phase 2: Multi-Node, Conditional (2-3 months)

**Characteristics**:
- 5-10 node types (LLM_DECISION, TOOL_CALL, FILE_READ, etc.)
- Conditional edges (branching based on outputs)
- SQLite storage option (behind same interface)
- Improved pattern detection

**What changes**:
- Add node executors (register with NodeRegistry)
- Add edge condition evaluator
- Add SQLite storage implementation
- NO CHANGES to core interfaces (only additions)

**Complexity**: ~2000-3000 LOC total

### Phase 3: Mature System (6-12 months)

**Characteristics**:
- Parallel node execution
- Sub-workflows (composition)
- Multiple LLM providers
- Real-time learning (during execution)
- Community-contributed integrations

**What changes**:
- Execution engine supports parallelism
- Workflow entity supports sub-workflows
- More storage backends (Postgres, etc.)
- Still NO breaking changes to interfaces

**Complexity**: ~5000-8000 LOC total

### The Magic: Interfaces Don't Change

The `WorkflowExecutor` interface from MVP (`Execute(workflow, input) → output`) still works in Phase 3 with parallel execution. Implementation gets smarter, contract stays stable.

## Architectural Boundaries

### Core ↔ Integration Boundary

**Core defines**: `NodeExecutor` interface with Execute() method

**Integration implements**: ClaudeCodeNode, FileReadNode, etc. - each implements NodeExecutor

**Rule**: Core never imports integration packages. Integration imports core.

### Core ↔ Storage Boundary

**Core defines**: `LogRepository` interface with StoreLog(), GetLog(), QueryLogs() methods

**Storage implements**: FileLogRepository, SQLiteLogRepository, PostgresLogRepository

**Rule**: Core never knows about files, databases, or any persistence mechanism.

### Core ↔ Reference Workflow Boundary

**Core loads**: Workflows from YAML files (data, not code)

**Core validates**:
- Node types registered?
- Edges form valid graph?
- Config matches node type schema?

**Rule**: Core executes workflows but doesn't create domain logic.

## Testing Strategy

### Unit Tests

Each layer tested in isolation:
- **Core**: Use mock storage, mock node executors to test execution logic
- **Integration**: Test adapters with mocked external tools (e.g., mock Claude Code CLI)

### Integration Tests

Test boundaries:
- **Core + Storage**: Test with real file storage or SQLite
- **Core + Nodes**: Test with real node executors

### End-to-End Tests

Full system tests:
- Load real workflow YAML
- Execute with real storage + real nodes
- Verify logs, metrics, state

## Metrics & Observability

Built into architecture through interfaces:

**ExecutionLogger**: LogStepStart(), LogStepEnd(), LogError()

**MetricsCollector**: RecordDuration(), RecordSuccess(), RecordFailure(), RecordPatternUsage()

These interfaces sit at architectural boundaries, capturing everything without coupling.

## Migration Path (File → Database)

Because storage is abstracted, you can swap backends without changing executor code:

1. **MVP**: File-based LogRepository
2. **Later**: SQLite-based LogRepository (same interface, no executor changes)
3. **Scale**: Postgres-based LogRepository (same interface, no executor changes)

Same executor code. Different storage. That's the power of dependency inversion.

## What Good Architecture Looks Like

If the architecture is right:

1. **Adding new node type** = Implement interface + register plugin (~50 LOC)
2. **Adding new storage backend** = Implement interface (~200 LOC)
3. **Adding new workflow** = Write YAML (0 LOC in codebase)
4. **Changing reflection algorithm** = Update one component (~100 LOC changed)
5. **Switching LLM provider** = Swap adapter (~0 LOC changed, config only)

If any of these require changing core framework code, architecture is wrong.

## Anti-Patterns to Avoid

### ❌ Leaky Abstractions

**Bad**: Storage interface exposes GetConnection() returning SQL database connection

**Good**: Pure contract with GetLog(runID) returning ExecutionLog - no implementation details leak

### ❌ God Objects

**Bad**: Single DarwinFlow object with ExecuteWorkflow(), StoreLog(), AnalyzePatterns(), GenerateNewWorkflow() - too many responsibilities

**Good**: Separate objects - WorkflowExecutor, LogRepository, PatternAnalyzer, WorkflowGenerator - each with single responsibility

### ❌ Hardcoded Dependencies

**Bad**: WorkflowExecutor depends on concrete FileLogRepository - coupled to file storage

**Good**: WorkflowExecutor depends on LogRepository interface - can be any implementation

### ❌ Domain Logic in Core

**Bad**: WorkflowExecutor has hardcoded logic like "if request contains 'bug', check tests first"

**Good**: Domain logic lives in workflow YAML - nodes for LLM_DECISION, TOOL_CALL with conditional edges based on decision output

## Success Metrics

Architecture is succeeding if:

1. **Adding new node types requires < 100 LOC** (just interface implementation)
2. **Storage backend swap requires 0 core changes** (proves abstraction works)
3. **Test coverage > 80%** (interfaces make testing easy)
4. **New workflow creation requires 0 code** (just YAML)
5. **Learning improvements touch < 5% of codebase** (proper boundaries)

If these metrics degrade, architecture is degrading.

## Implementation Priorities

### Must Have for MVP

1. Core interfaces defined (all of them, even if simple)
2. File storage implementation
3. CLAUDE_CODE node implementation
4. Basic workflow executor
5. LLM reflection engine

### Should Have for MVP

1. Metrics collection interfaces
2. Error handling strategy
3. Logging framework
4. Plugin registry for nodes

### Nice to Have for MVP

1. SQLite storage option
2. Multiple node types
3. Conditional edges

## Next Steps

Read the detailed interface definitions:

1. **[01_core_interfaces.md](01_core_interfaces.md)** - Complete interface definitions for core framework
2. **[02_integration_architecture.md](02_integration_architecture.md)** - How integrations plug in, node types, adapters
3. **[03_storage_interfaces.md](03_storage_interfaces.md)** - Storage abstraction, file/DB implementations
4. **[04_evolution_roadmap.md](04_evolution_roadmap.md)** - Detailed evolution from MVP → mature system

---

## Appendix: Key Design Decisions

### Why YAML for Workflows?

- **Human-readable** - developers can read/edit
- **Version-controllable** - track changes over time
- **Diff-friendly** - see what changed between versions
- **Data, not code** - keeps domain logic out of framework

### Why Plugin Architecture for Nodes?

- **Open-Closed** - add nodes without modifying core
- **Testing** - easy to mock node types
- **Community** - users can contribute node types

### Why Abstract Storage?

- **Start Simple** - files for MVP
- **Grow Later** - databases when needed
- **Test Easily** - in-memory storage for tests
- **Never Locked In** - swap backends anytime

### Why LLM-Based Reflection?

- **Flexibility** - can analyze any pattern
- **Adaptability** - improves as LLMs improve
- **Explainability** - can explain why changes made
- **No Hardcoding** - discovers patterns, doesn't assume them

---

**Remember**: Architecture is about managing dependencies and handling change. If you can add capabilities without modifying existing code, you've achieved the goal.

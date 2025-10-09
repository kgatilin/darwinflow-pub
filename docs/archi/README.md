# DarwinFlow Architecture Documentation

**Status**: Target Architecture
**Version**: 1.0
**Date**: 2025-10-09
**Author**: Bob Martin (well, his architectural spirit anyway)

---

## What's Here

This directory contains the complete target architecture for DarwinFlow - a learning workflow framework that improves through use.

**The architecture enables**:
- Starting simple (one node type, file storage)
- Growing organically (add capabilities as needed)
- Never rewriting (stable interfaces, growing implementations)

---

## Reading Order

### For First-Time Readers

**Start here**: [00_architecture_overview.md](00_architecture_overview.md)

Then read in order:
1. **[00_architecture_overview.md](00_architecture_overview.md)** - Principles, layers, boundaries
2. **[01_core_interfaces.md](01_core_interfaces.md)** - Stable contracts that enable evolution
3. **[02_integration_architecture.md](02_integration_architecture.md)** - How tools plug in
4. **[03_storage_architecture.md](03_storage_architecture.md)** - Persistence abstraction
5. **[04_evolution_roadmap.md](04_evolution_roadmap.md)** - How system grows over time
6. **[05_architecture_diagrams.md](05_architecture_diagrams.md)** - Visual reference (PlantUML)

**Time investment**: 2-3 hours to read everything thoroughly.

### For Quick Reference

**Need to add a feature?** → [04_evolution_roadmap.md](04_evolution_roadmap.md#quick-reference-adding-capabilities)

**Need interface definitions?** → [01_core_interfaces.md](01_core_interfaces.md)

**Need visual diagrams?** → [05_architecture_diagrams.md](05_architecture_diagrams.md)

**Need storage design?** → [03_storage_architecture.md](03_storage_architecture.md)

---

## Key Architectural Decisions

### 1. Three-Layer Architecture

```
┌─────────────────────────────────┐
│   Reference Workflows (YAML)    │  ← Domain-specific orchestration
└─────────────────────────────────┘
              ↓ uses
┌─────────────────────────────────┐
│   Integrations (Plugins)        │  ← Reusable tools
└─────────────────────────────────┘
              ↓ implements
┌─────────────────────────────────┐
│   Core Framework (Interfaces)   │  ← Domain-agnostic engine
└─────────────────────────────────┘
```

**Why?** Keeps domain logic out of framework. Enables reuse. Allows community contributions.

### 2. Dependency Inversion Everywhere

```
Integrations  →  Core Interfaces  →  Domain Entities
  (volatile)       (stable)           (stable)
```

**Why?** Outer layers can change without breaking inner layers. Enables testing. Allows swapping implementations.

### 3. Interface Segregation

Small, focused interfaces instead of kitchen-sink interfaces.

**Why?** Clients depend only on what they use. Easy to mock. Clear responsibilities.

### 4. Progressive Evolution

```
Phase 1 (MVP):     1 node type,  files,      1K LOC
Phase 2 (Growth):  6 node types, files,      3K LOC
Phase 3 (Scale):   10 node types, SQLite,    5K LOC
Phase 4 (Mature):  20+ node types, Postgres, 10K LOC
```

**Why?** Start simple, grow based on real usage. No premature optimization. No predicted complexity.

### 5. Plugin Architecture for Nodes

Any node type can be added by implementing `NodeExecutor` interface.

**Why?** Open-Closed Principle. Add capabilities without modifying core. Community can contribute.

### 6. Storage Abstraction

Three separate repositories: Workflows, Logs, Metrics.

**Why?** Different access patterns. Different scaling needs. Swappable backends (files → SQLite → Postgres).

---

## Core Principles Applied

### SOLID Principles

| Principle | How Applied | Where to Read |
|-----------|-------------|---------------|
| **Single Responsibility** | Each class/interface has ONE reason to change | [00_architecture_overview.md](00_architecture_overview.md#4-single-responsibility-at-architectural-level) |
| **Open-Closed** | Add node types without modifying core | [02_integration_architecture.md](02_integration_architecture.md) |
| **Liskov Substitution** | Any storage implementation can replace another | [03_storage_architecture.md](03_storage_architecture.md) |
| **Interface Segregation** | Small, focused interfaces | [01_core_interfaces.md](01_core_interfaces.md) |
| **Dependency Inversion** | Core depends on abstractions, not implementations | [00_architecture_overview.md](00_architecture_overview.md#1-the-dependency-rule) |

### Clean Architecture Principles

- **Screaming Architecture** - The structure tells you what it does (workflow framework, not "microservices")
- **Independence** - Framework, UI, database, external tools are independent
- **Testability** - Business rules testable without external dependencies
- **Framework Independence** - Not tied to any framework or tool

---

## Success Criteria

The architecture is succeeding if:

1. ✅ **Adding new node type requires < 100 LOC** (just implement interface)
2. ✅ **Storage backend swap requires 0 core changes** (abstraction works)
3. ✅ **New workflow creation requires 0 code** (just YAML)
4. ✅ **Tests from Phase 1 still pass in Phase 4** (backward compatible)
5. ✅ **Learning improvements touch < 5% of codebase** (proper boundaries)

If these metrics degrade, the architecture is degrading.

---

## Common Questions

### Q: Why not use a workflow engine like Temporal or Airflow?

**A**: Those are task orchestration engines. DarwinFlow is a *learning* workflow framework. The key innovation is reflection - the system learns from usage and generates improved workflow versions. That's not what Temporal does.

Also: Starting simple. If we need workflow engine features later, we can integrate them or build minimal versions.

### Q: Why YAML for workflows instead of code?

**A**:
1. **Data > Code** - Workflows should be data that drives the engine, not imperative code
2. **Version control friendly** - Easy to diff versions
3. **LLM-friendly** - LLMs can generate and modify YAML
4. **Human-readable** - Developers can read and edit
5. **Separation** - Keeps domain logic out of framework code

### Q: Won't file storage be too slow?

**A**: For MVP (20-100 executions), files are fine and debuggable. When it hurts, we add SQLite (already designed in interfaces). Don't optimize prematurely.

### Q: Why not just use LangChain or similar?

**A**: LangChain is about LLM chaining. DarwinFlow is about learning workflow orchestration. Different problems. Also, we want full control over the learning mechanism.

### Q: Is this overengineered for MVP?

**A**: No. The *interfaces* are comprehensive, but the *implementations* are minimal. This lets us start simple (file storage, one node type) but grow without rewrites. The architecture enables simplicity, not complexity.

### Q: How do you know this will work?

**A**: These patterns have been proven over decades:
- **Unix philosophy** - small tools, clean interfaces
- **Linux kernel** - stable syscalls, growing capabilities
- **Postgres** - stable protocol, extensible plugins
- **Clean Architecture** - used in countless successful systems

We're not inventing new patterns, we're applying proven ones.

---

## Anti-Patterns Avoided

### ❌ God Objects

No single class does everything. Clear separation of concerns.

### ❌ Leaky Abstractions

Storage interfaces don't leak SQL or file details. Node interfaces don't leak tool specifics.

### ❌ Premature Optimization

Start with files, not distributed database. Start with one node type, not twenty.

### ❌ Feature Bloat

Only add features when pain is felt during dogfooding. No "nice to haves" without proven need.

### ❌ Framework Lock-in

Not tied to any framework, database, or tool. All dependencies are abstracted.

### ❌ Hardcoded Domain Logic

Domain logic lives in workflows (YAML), not core framework (Go code).

---

## Implementation Checklist

Before calling Phase 1 (MVP) complete:

- [ ] Core interfaces defined (`WorkflowExecutor`, `NodeExecutor`, etc.)
- [ ] File-based storage implementations work
- [ ] `CLAUDE_CODE` node type works
- [ ] Workflow executor executes simple workflows
- [ ] Logging captures all execution details
- [ ] Metrics track duration, success, patterns
- [ ] Reflection command analyzes logs and generates v2
- [ ] Tests cover core functionality (>80% coverage)
- [ ] Can dogfood the system (use it to build itself)

Before calling Phase 2 complete:

- [ ] 5+ node types implemented (`FILE_READ`, `FILE_WRITE`, `LLM_DECISION`, etc.)
- [ ] Conditional edges work
- [ ] Edge condition evaluation is robust
- [ ] Reflection generates multi-node workflows
- [ ] Measurable improvement: 30% fewer corrections

---

## Relationship to Other Docs

```
docs/
├── product/
│   └── vision.md              ← Product vision (WHY we're building this)
├── mvp_simple.md              ← MVP specification (WHAT we're building first)
└── archi/                     ← Target architecture (HOW we're building it)
    ├── 00_architecture_overview.md
    ├── 01_core_interfaces.md
    ├── 02_integration_architecture.md
    ├── 03_storage_architecture.md
    ├── 04_evolution_roadmap.md
    └── 05_architecture_diagrams.md
```

**Vision** defines the goal. **MVP** defines the first version. **Architecture** defines the path from MVP to mature system.

---

## Next Steps

1. **Read the docs** (you are here)
2. **Set up development environment** (Go, dependencies)
3. **Implement Phase 1** (MVP - core interfaces + file storage + one node type)
4. **Write tests** (interfaces make this easy)
5. **Dogfood** (use DarwinFlow to build DarwinFlow)
6. **Reflect** (let system learn and generate v2)
7. **Validate** (does v2 actually improve things?)
8. **Iterate** (move to Phase 2 based on learnings)

---

## Contributing to Architecture

These documents are living artifacts. As we learn, they evolve.

**When to update architecture docs**:
- Discovered a design flaw → Document it + propose fix
- Added significant capability → Update relevant doc
- Changed core interface → Update interface doc + explain why
- Learned from dogfooding → Add to evolution roadmap

**How to update**:
1. Open PR with proposed changes
2. Explain rationale (what problem does this solve?)
3. Show evidence (logs, metrics, pain points)
4. Get review from team
5. Update docs after implementation proves design

**Remember**: Architecture evolves based on reality, not prediction.

---

## Quotes to Live By

> "The architecture should scream its intent." - Bob Martin

> "The best architecture is the one that allows the system to evolve." - Bob Martin

> "Make it work, make it right, make it fast - in that order." - Kent Beck

> "First, solve the problem. Then, write the code." - John Johnson

> "Premature optimization is the root of all evil." - Donald Knuth

> "The most important property of a program is whether it accomplishes the intention of its user." - C.A.R. Hoare

---

## License

This architecture documentation is part of the DarwinFlow project.

---

**Last Updated**: 2025-10-09

**Version**: 1.0

**Status**: Target Architecture - Ready for Implementation

---

**Now go build it. Start with Phase 1. Ship it. Use it. Let it evolve.**

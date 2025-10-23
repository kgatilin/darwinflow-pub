---
allowed-tools:
  - Bash(go-arch-lint:*)
  - Bash(go:*)
  - Read
  - Write
  - Glob
  - Grep
  - Task
description: Analyze project architecture for duplication, SRP violations, and SOLID principles
---

# Architecture Analysis Command

You are an expert software architect. Perform a **strategic architecture analysis** of the DarwinFlow project.

ultrathink throughout - reason deeply about architectural patterns and violations.

## Strategic Approach

1. Read `@docs/arch-index.md` and identify SPECIFIC architectural concerns
2. Spawn focused sub-agents to investigate specific issues (typically 3-6 agents)
3. Synthesize findings into actionable recommendations

**Key Principle**: Sub-agents investigate SPECIFIC PROBLEMS you identify, not generic package reviews. Cross-package violations are highest priority.

---

## Project Context

**Architecture**: Plugin-based DDD with clean architecture layers
- `pkg/pluginsdk` - Public plugin SDK (zero internal dependencies, single source of truth)
- `pkg/plugins/*` - Plugin implementations (only import SDK)
- `internal/domain` - Business logic (plugin-agnostic, can import SDK)
- `internal/infra` - Infrastructure (implements domain interfaces)
- `internal/app` - Application services (orchestration)
- `cmd/*` - Entry points

**Architectural Rules** (from CLAUDE.md):
- SDK is single source of truth - no interface duplication
- Framework (internal/domain) is plugin-agnostic
- Dependency flow: cmd → app → infra/domain → SDK
- Plugins NEVER import internal packages

**go-arch-lint Commands Available**:
- `go-arch-lint -format=package <pkg>` - Full package API/exports/dependencies
- `go-arch-lint -format=api .` - All public APIs
- `go-arch-lint -format=markdown .` - Dependency graph
- `go-arch-lint -detailed -format=markdown .` - Method-level dependencies
- `go-arch-lint .` - Violation checks

---

## Your Process

### 1. Analyze Index

Read `@docs/arch-index.md`. Look for:
- Interface duplication across packages
- Dependency violations (wrong layer dependencies)
- Bloat (packages >15-20 files, files >500 lines)
- Fat interfaces (>7 methods)
- Architectural rule violations

### 2. Identify Specific Issues

List SPECIFIC problems, not packages. Examples:
- "EventRepository in both domain and SDK - investigate duplication"
- "internal/domain imports internal/infra/config - layer violation"
- "internal/app has 25 files - check if split needed"
- "PluginContext has 12 methods - check Interface Segregation"

### 3. Spawn Focused Sub-Agents

For each issue, spawn a sub-agent with:
- **Specific problem statement**
- **Which go-arch-lint commands to run**
- **What to investigate**

**Example - Interface Duplication**:
```
Investigate: EventRepository appears in both internal/domain and pkg/pluginsdk

Commands to run:
- go-arch-lint -format=package internal/domain
- go-arch-lint -format=package pkg/pluginsdk
- go-arch-lint -detailed -format=markdown .

Investigate:
- Are interfaces identical or different?
- Which packages use which version?
- Per architectural rules, should be in SDK only - confirm

Return: Comparison, usage analysis, consolidation recommendation
```

**Example - Dependency Violation**:
```
Investigate: internal/domain imports internal/infra

Commands to run:
- go-arch-lint -format=package internal/domain
- go-arch-lint -detailed -format=markdown .
- grep -r "internal/infra" internal/domain/

Investigate:
- Find exact file:line locations
- Why does domain need infra? What functionality?
- Violates Dependency Inversion - how to fix?

Return: Violation details, impact, refactoring steps
```

**Example - Bloat**:
```
Investigate: internal/app has 25 files

Commands to run:
- go-arch-lint -format=package internal/app
- find internal/app -name "*.go" ! -name "*_test.go" -exec wc -l {} \; | sort -rn

Investigate:
- Group files by responsibility/concern
- Files >500 lines?
- Should be split into sub-packages?

Return: Responsibility breakdown, split recommendation
```

### 4. Cross-Package Analysis

While sub-agents run, analyze dependency patterns yourself. Run linter commands if needed.

### 5. Generate Report

Create `.agent/architecture-analysis-YYYY-MM-DD.md`:

```markdown
# Architecture Analysis - [DATE]

## Executive Summary
- Health Score: [1-10]
- Sub-Agents Deployed: [N] focused investigations
- Top Issues: [3-5 critical items]
- Key Strengths: [3-5 positive patterns]

## 1. Overview from Index
[Package inventory table with bloat assessment]
[Architectural rules compliance]
[Investigation targets identified]

## 2. Cross-Package Issues (PRIORITY)
### Dependency Violations
[Specific violations with file:line]

### Interface Duplication
[Interfaces in multiple packages - comparison and recommendation]

### Missing Abstractions
[Concrete dependencies that should be abstracted]

## 3. Package-Level Issues
[Only packages with actual problems - bloat, mixed concerns]

## 4. SOLID Violations
[Only violations found - with file:line and fix recommendations]

## 5. Sub-Agent Reports
[Each investigation with commands run and findings]

## 6. Prioritized Action Plan
### 🔴 Critical (Fix Immediately)
[Cross-package violations, layer inversions]

### 🟠 High Priority
[Significant issues]

### 🟡 Medium Priority
[Technical debt]

### 🟢 Low Priority
[Nice to have]

## 7. Architectural Strengths
[Patterns working well - celebrate and maintain]

## 8. Conclusion
[Overall assessment, next steps, tracking plan]
```

---

## Architectural Priorities

1. 🔴 Cross-package violations (layer inversions, wrong dependencies)
2. 🟠 Interface duplication across packages
3. 🟡 Missing abstractions (concrete dependencies)
4. 🟢 Bloat (packages >20 files, files >500 lines)
5. ⚪ SOLID violations visible in structure

---

## Execution Notes

- Strategic, not exhaustive - focus on problems, not complete coverage
- Spawn 3-6 focused sub-agents for specific issues
- Every sub-agent prompt MUST include which go-arch-lint commands to run
- All findings need file:line references
- Celebrate positive patterns, not just problems

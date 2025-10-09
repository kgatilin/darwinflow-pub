# DarwinFlow - Product Vision

## Vision Statement

A learning workflow framework that transforms personal AI assistants from static prompt executors into adaptive systems that improve through use, reducing cost and latency by learning from human feedback to replace expensive reasoning with efficient patterns.

## The Job We're Hired For

**When** people use AI assistants for repetitive domain-specific work,
**they struggle with** providing the same context repeatedly, correcting the same mistakes, and paying for the same reasoning over and over,
**so they want** a system that learns their patterns and gets better over time,
**enabling them to** work more efficiently while spending less on AI.

## User Personas & Their Jobs

### Software Engineer
- **Job**: Code features, fix bugs, review PRs
- **Current pain**: Repeatedly explaining project context, file structures, coding standards to AI
- **How DarwinFlow helps**: Learns project-specific patterns like "For bug fixes in auth/, check security logs first" or "Always run test suite before committing"
- **Outcome**: AI gathers right context automatically, fewer corrections needed over time

### Product Manager
- **Job**: Validate problems, iterate on solutions, maintain roadmap
- **Current pain**: Re-explaining product context, user research patterns, validation steps
- **How DarwinFlow helps**: Learns validation workflows like "For feature requests, check analytics first" or "Always validate against target persona"
- **Outcome**: Consistent problem validation, less time re-teaching AI product context

### Software Architect
- **Job**: Evaluate designs, ensure consistency, review technical decisions
- **Current pain**: Repeatedly enforcing architectural patterns, explaining design principles
- **How DarwinFlow helps**: Learns architecture checks like "For API changes, verify backwards compatibility" or "Always consider scalability implications"
- **Outcome**: Automated architecture guardrails, consistent technical reviews

## What This Is / Is Not

### This IS:
✓ A general-purpose learning workflow framework
✓ Extensible for any domain-specific workflow
✓ Focused on pattern learning from human feedback
✓ Designed to reduce cost/latency through workflow compilation
✓ Multi-persona, multi-domain capable
✓ Human-in-the-loop by design

### This IS NOT:
✗ A domain-specific tool (not "just for coding" or "just for PM work")
✗ A static workflow automation tool (like Zapier/n8n)
✗ A prompt engineering playground
✗ A fully autonomous agent (human remains in the loop)
✗ A visual workflow builder (v1)
✗ A multi-agent orchestration system

## Product Architecture

DarwinFlow consists of three conceptual layers:

### 1. Core Framework (Domain-Agnostic)
The learning engine that powers workflow adaptation:
- Workflow execution and state management
- Pattern recognition from human feedback
- Compilation of learned patterns into efficient operations
- Node types (LLM calls, tool calls, decisions, human input)

**Philosophy**: Never hardcode domain logic. Provide primitives others build upon.

### 2. Integrations (Reusable Tools)
Tools the framework orchestrates, not business logic:
- Communication platforms (Telegram, Slack, etc.)
- Development tools (GitHub, file systems, etc.)
- LLM providers and other AI services
- Data stores and APIs

**Philosophy**: Low coupling to domains. If multiple workflow types can use it, it's an integration.

### 3. Reference Workflows (Domain Demonstrations)
Domain-specific orchestrations that showcase the framework:
- Software engineering workflows (primary dogfooding use case)
- Product management workflows
- Architecture review workflows
- Others as needed

**Philosophy**: Ships with the product to demonstrate capabilities and reduce friction. Users can fork, customize, or build their own.

## Decision Criteria for Features/PRs

### ✅ ACCEPT as Core Framework if the feature:
- Enables workflow execution/learning regardless of domain
- Improves pattern recognition or feedback processing
- Enhances state management, node execution, or routing
- Reduces cost/latency for all workflow types
- **Examples**: Feedback capture API, pattern compiler, confidence scoring

### ✅ ACCEPT as Integration if the feature:
- Provides reusable tool across multiple workflow types
- Has low coupling to specific domain logic
- Extends framework's tool-calling capabilities
- **Examples**: Telegram bot, GitHub API, file system ops, LLM providers

### ✅ ACCEPT as Reference Workflow if the feature:
- Demonstrates domain-specific orchestration of framework + integrations
- Needed for dogfooding the framework
- Serves as template for others building similar workflows
- **Examples**: Software engineer workflow, PM workflow, architect workflow

### ❌ REJECT if the feature:
- Hardcodes domain logic into core framework
- Can't be generalized even as a reference workflow
- Doesn't contribute to learning/optimization
- Makes framework less general-purpose
- Is a one-off script that doesn't demonstrate learning

### 🤔 QUESTION if the feature:
- Could be configuration/data vs code
- Belongs in reference workflow vs integration vs core
- Is a tool the framework orchestrates vs framework capability
- Should be a plugin/extension vs included in the product

## Core Principles

1. **Learn, Don't Hardcode**: Solutions emerge from patterns, not predefined rules
2. **General Over Specific**: Framework capabilities over domain solutions
3. **Efficiency Through Learning**: Every interaction should reduce future cost
4. **Human-Guided Evolution**: Learn from feedback, don't replace judgment
5. **Composable Workflows**: Personal + Project + Org workflows layer together
6. **Tools, Not Logic**: Integrations are tools the framework orchestrates, not business logic

## Strategic Tradeoffs

When features conflict, priority order:

1. **Framework generality** > Case-specific functionality
2. **Automatic learning** > Manual configuration
3. **Cost/latency reduction** > Feature richness
4. **Extensibility** > Built-in solutions
5. **Reusable integration** > Workflow-specific tool

## Example PR Validations

### Scenario 1: "Add Telegram bot with message parsing"
- **Decision**: ✅ ACCEPT as Integration
- **Reason**: Reusable tool, multiple workflows can use it for task capture
- **Key test**: Can PM workflow and Engineer workflow both use this?

### Scenario 2: "Add 'always check tests' rule for bug fixes"
- **Decision**: ✅ ACCEPT as Reference Workflow
- **Reason**: Domain-specific logic for coding workflow
- **Key test**: Does this apply to all workflows? No → it's workflow-specific

### Scenario 3: "Add JavaScript linting to workflow"
- **Decision**: ✅ ACCEPT (split across layers)
  - Generic tool execution → Core Framework
  - JS linter integration → Integration
  - Usage in coding workflow → Reference Workflow
- **Reason**: Separates general capability from specific tool from domain usage

### Scenario 4: "Optimize LLM token usage in context gathering"
- **Decision**: ✅ ACCEPT as Core Framework
- **Reason**: Benefits all workflows, core learning mechanism
- **Key test**: Does every workflow benefit? Yes → it's core

### Scenario 5: "Add specialized PR review checklist for React"
- **Decision**: 🤔 QUESTION first:
  - Is this configurable data in software-engineer workflow? → ✅ Accept as config
  - Is this hardcoded React logic? → ❌ Reject, make it configurable
- **Reason**: Avoid hardcoding specific frameworks, prefer data-driven approaches

### Scenario 6: "Add visual workflow builder UI"
- **Decision**: ❌ REJECT (marked as non-goal for v1)
- **Reason**: Conflicts with "learn from usage" approach, adds complexity
- **Alternative**: Focus on learning from human feedback instead

## Feature Evolution

Features naturally mature and migrate across layers:

```
1. Experiment in reference workflow
   ↓ (works well, other workflows could use it)
2. Extract reusable parts as integrations
   ↓ (pattern emerges across all workflows)
3. Generalize into core framework capability
   ↓ (future: workflows mature)
4. Reference workflows become shareable templates
```

### Example Journey:
- **Start**: "Add file context gathering to coding workflow"
- **Extract**: "File system integration" becomes reusable tool
- **Generalize**: "Context gathering patterns" moves to core learning
- **Share**: "Software engineer workflow" becomes template for others

This path encourages experimentation while keeping the framework clean.

## Success Metrics

### Framework Success:
- **Token Reduction**: % decrease in LLM tokens per task over time
- **Pattern Reuse**: % of tasks using learned workflows vs full LLM
- **Cost Efficiency**: $ per task completed (trending down)
- **Learning Rate**: New patterns identified per week

### Workflow Success:
- **Time to Context**: Seconds from task start to context ready
- **Accuracy**: % of tasks completed without human correction
- **Adoption**: Number of active workflows in production

### Integration Success:
- **Reuse**: Number of workflows using each integration
- **Reliability**: Uptime and error rates
- **Extensibility**: Ease of adding new integrations

## What Success Looks Like

### Month 1:
- Software engineer workflow capturing context with full LLM
- All operations logged, feedback mechanisms in place
- First patterns identified (repeated file accesses)

### Month 3:
- Common patterns compiled to deterministic nodes
- "For auth bugs, check security logs" learned automatically
- 30% reduction in tokens for repeated task types

### Month 6:
- PM workflow operational alongside engineering workflow
- Shared integrations (Telegram, GitHub) working for both
- 50% token reduction, new task types routed to appropriate workflow

### Month 12:
- Multiple reference workflows demonstrating framework
- Community starting to build custom workflows
- 70% token reduction, most tasks use learned patterns
- Framework mature enough for external adoption

## Open Questions

These questions guide our roadmap but don't block current work:

1. How to handle conflicting feedback from different team members?
2. When to garbage collect rarely-used workflow branches?
3. How to version workflows as they evolve?
4. How to resolve conflicts between personal and project workflows?
5. Should personal workflows be exportable to become project/org workflows?
6. What's the plugin/extension model for community-contributed integrations?
7. How to handle context that changes (files deleted, APIs updated)?

---

**Last Updated**: 2025-10-09
**Version**: 1.0
**Status**: Living document - updated as product evolves

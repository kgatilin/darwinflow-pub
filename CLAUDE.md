# DarwinFlow - Project Context

## What This Project Is

DarwinFlow is a **learning workflow framework** that makes AI assistants improve through use by learning from user corrections.

**Core concept**: Instead of manually correcting the AI every time, the system observes corrections, identifies patterns, and automatically creates improved workflow versions.

## The MVP

### What We're Building (3 Steps)

**Step 1: Workflow Execution + Logging**
- Execute workflows defined in YAML
- Single node type initially: `CLAUDE_CODE` (delegates to Claude Code)
- Log everything to files (one file per execution)
- Abstract storage interface (not tied to specific database)

**Step 2: Metrics Tracking**
- Separate from detailed logs
- Track patterns, durations, success rates
- File-based storage (can add DB later)

**Step 3: Reflection & Learning**
- Manual command: `darwinflow reflect`
- LLM analyzes execution logs to find patterns
- Automatically creates new workflow versions (v2, v3, etc.)
- User chooses which version to use

### How It Works

```
1. User runs workflow v1 → system logs everything
2. User corrects: "No, read vision.md first"
3. After multiple corrections, user runs: darwinflow reflect
4. LLM finds pattern: "read vision.md first" corrected 3 times
5. System creates workflows/assistant_v2.yaml automatically
6. User chooses to use v1 or v2 for next task
```

## Current File Structure

```
darwinflow-pub/
├── docs/
│   ├── mvp_simple.md              # THE MVP specification
│   └── product/
│       └── vision.md              # Product vision & principles
├── workflows/
│   └── assistant_v1.yaml          # Initial simple workflow
└── README.md                      # Project overview
```

## Key Design Decisions

### Start Simple, Evolve Through Use
- Begin with single-node workflow (just call Claude Code)
- Don't predict complexity upfront
- Let complexity emerge from actual usage patterns

### No Approval Flow
- Reflection creates new versions automatically
- User picks which version to use (v1, v2, v3, etc.)
- No "approve/reject" prompts

### Abstract Storage
- Logs: Files (one JSON per execution)
- Metrics: Separate aggregated files
- Can add SQLite/Postgres later behind same interface

### Workflow Format: YAML
- Human-readable
- Version controlled
- Easy to diff versions

## Important Files

### docs/mvp_simple.md
Complete MVP specification. Read this to understand:
- What needs to be built
- Workflow structure
- Logging requirements
- Reflection process
- Success criteria

### docs/product/vision.md
Product vision and architecture:
- Three-layer architecture (Core / Integrations / Reference Workflows)
- Decision criteria for features
- "Learn, Don't Hardcode" principle

### workflows/assistant_v1.yaml
The starting workflow:
- Single node: `CLAUDE_CODE`
- Just delegates to Claude Code
- Will evolve based on usage

## What NOT to Do

❌ Don't design complex workflows upfront
❌ Don't hardcode domain logic in core framework
❌ Don't require user approval for every change
❌ Don't tie implementation to specific database (SQLite, etc.)
❌ Don't add features that aren't in mvp_simple.md

## What TO Do

✅ Keep workflows simple (start with 1 node)
✅ Use abstract interfaces (storage, logging)
✅ Let LLM generate new workflow versions
✅ Store everything as files initially
✅ Follow mvp_simple.md specification

## Development Approach

### Dogfooding
Use DarwinFlow to build DarwinFlow itself:
- Run workflow for each feature
- Make corrections naturally
- Run reflection periodically
- Use evolved workflows
- Validate learning works in real usage

### Success Criteria

**Step 1**: Can execute workflows and review logs
**Step 2**: Metrics showing patterns in usage
**Step 3**: Reflection creates improved workflow versions

**MVP Success**:
- 3+ workflow evolutions from real usage
- 30%+ reduction in corrections
- System measurably improves through use

## Workflow Evolution Example

**v1**: `START → [CLAUDE_CODE] → END`

**v2** (after learning "read vision.md first"):
```yaml
START → [READ_FILE: vision.md] → [CLAUDE_CODE] → END
```

**v3** (after learning "classify features"):
```yaml
START → [READ vision.md] → [CLASSIFY_FEATURE] → [CLAUDE_CODE] → END
```

Each version created by reflection, not designed upfront.

## Key Principles

1. **Learn, Don't Hardcode** - Patterns emerge from corrections
2. **General Over Specific** - Framework, not domain solutions
3. **Start Simple** - Single node → complexity emerges
4. **User Choice** - No forced changes, user picks versions
5. **Transparent** - All changes visible and understandable

## Questions During Development

When building, refer to `docs/mvp_simple.md` for:
- Exact workflow structure
- Logging interface requirements
- Reflection process details
- Open questions and TBD items

---

**Remember**: We're building a system that learns. Start simple, use it for real work, let patterns emerge naturally.

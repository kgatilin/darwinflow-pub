# DarwinFlow MVP - Simple Description

## What Needs to Exist

### 1. Workflow Definition

**File format**: YAML

**Structure**:
```
Workflow:
  - id: unique identifier
  - version: number
  - nodes: list of nodes
  - edges: connections between nodes

Node:
  - id: unique identifier
  - type: node type (initially just "CLAUDE_CODE")
  - config: node-specific configuration
  - logging_config: TBD (should be configurable per node somehow)

Edge:
  - from: source node id
  - to: target node id
  - condition: optional (for branching)
```

**Initial node type**: `CLAUDE_CODE`
- Delegates to Claude Code
- Passes user request
- Returns result

**Example v1 workflow**:
```yaml
id: assistant
version: 1

nodes:
  - id: call_claude
    type: CLAUDE_CODE
    config:
      pass_full_request: true
    # logging_config: TBD - needs to support configurable logging per node

edges:
  - from: START
    to: call_claude
  - from: call_claude
    to: END
```

**Workflow storage**:
```
workflows/
  assistant_v1.yaml
  assistant_v2.yaml  # created by learning process
  assistant_v3.yaml  # etc.
```

### 2. Workflow Execution + Logging

**How to invoke workflow**:

Option 1: Custom binary/CLI
```bash
darwinflow run --workflow workflows/assistant_v1.yaml "Add Telegram bot"
```

Option 2: As library
```python
from darwinflow import WorkflowExecutor
executor = WorkflowExecutor("workflows/assistant_v1.yaml")
executor.run("Add Telegram bot")
```

Option 3: Wrapper around existing tool
```bash
# Wraps Claude Code
darwinflow-claude "Add Telegram bot"
# Internally: loads workflow, executes with logging
```

**Decision needed**: Which invocation method for MVP?

**Execution**:
- Load workflow definition (YAML)
- Execute nodes in sequence (following edges)
- Pass state between nodes
- Handle node-specific execution (initially just Claude Code)

**Logging**:

**What to log**:
- Execution level:
  - run_id (unique identifier)
  - workflow_id, workflow_version
  - user_request (what user asked)
  - started_at, completed_at
  - success/failure

- Node level:
  - step_id (unique identifier)
  - run_id (links to execution)
  - node_id (which node executed)
  - started_at, completed_at
  - inputs (what went into node)
  - outputs (what came out)
  - node_action_log (node-specific details)

**Node-specific logging** (CLAUDE_CODE node):
- Tool calls made (Read, Write, Edit, Bash, etc.)
- Files accessed (read/written)
- Commands executed
- LLM calls (prompt, response, tokens, cost)
- User interactions during execution

**Storage** (abstract interface):
- Must support: read/write logs, query by run_id
- Initial implementation: Files (JSON or structured text)
  - One file per execution for simplicity
- Future: Can add SQLite, Postgres, etc. behind same interface
- Must be queryable: "show me what happened in run X"

**Example file structure**:
```
logs/
  runs/
    2024-10-09_run_abc123.json    # Full execution log
    2024-10-09_run_def456.json
    2024-10-09_run_ghi789.json
```

**Example log file content**:
```json
{
  "run_id": "abc123",
  "workflow_id": "assistant",
  "workflow_version": 1,
  "user_request": "Add Telegram bot",
  "started_at": "2024-10-09T10:00:00Z",
  "completed_at": "2024-10-09T10:15:00Z",
  "success": true,
  "steps": [
    {
      "step_id": "step_001",
      "node_id": "call_claude",
      "started_at": "2024-10-09T10:00:01Z",
      "completed_at": "2024-10-09T10:15:00Z",
      "actions": [
        {"type": "file_read", "file": "docs/product/vision.md"},
        {"type": "file_write", "file": "src/bot/telegram.py"},
        {"type": "tool_call", "tool": "bash", "command": "make test"}
      ]
    }
  ]
}
```

**Logging configuration**:
- TBD: Should support per-node configuration of what to log
- Could be in workflow YAML or separate config file
- Should be flexible enough to add new logging types as node types grow

### Metrics (separate from logging)

**What to track**:
- Execution metrics:
  - Total duration per workflow execution
  - Duration per node
  - Success/failure rates
  - Correction frequency (if captured)

- Pattern metrics:
  - Most common operations
  - Most accessed files
  - Most used tool calls
  - Bottleneck identification

- Learning metrics:
  - Workflow versions created
  - Changes applied over time
  - Effectiveness of changes (did corrections decrease?)

**Storage** (also abstract):
- Initial implementation: Aggregated metrics in files
  - Updated after each execution
- Can be separate from detailed logs
- Used by reflection process to find patterns

**Example metrics file structure**:
```
metrics/
  assistant_v1_summary.json     # Aggregated metrics for workflow v1
  assistant_v2_summary.json     # Metrics for v2 (after evolution)
```

**Example metrics file content**:
```json
{
  "workflow_id": "assistant",
  "workflow_version": 1,
  "total_executions": 25,
  "successful_executions": 23,
  "failed_executions": 2,
  "avg_duration_seconds": 180,
  "patterns": {
    "most_common_files_read": [
      "docs/product/vision.md (15 times)",
      "README.md (8 times)"
    ],
    "most_common_tool_calls": [
      "Read (45 times)",
      "Write (32 times)",
      "Bash (18 times)"
    ],
    "most_common_corrections": [
      "Read vision.md first (3 times)",
      "Run tests before commit (2 times)"
    ]
  },
  "last_updated": "2024-10-09T15:00:00Z"
}
```

**Metrics interface**:
- `record_execution(run_id, duration, success)`
- `record_node_execution(node_id, duration)`
- `get_summary(workflow_id, version, recent_n_runs)`
- `get_patterns(workflow_id)`

### 3. Reflection and Learning

**Reflection process** (manual trigger initially):

**How it works**:
```bash
# User manually triggers reflection
darwinflow reflect --workflow workflows/assistant_v1.yaml --recent-runs 20
```

**Input**:
- Execution logs from N recent runs
- Current workflow definition (YAML)
- User feedback/corrections (if captured in logs)

**Process** (fully LLM-based):
1. **LLM analyzes logs**:
   - What patterns repeat?
   - What gets corrected frequently?
   - Where are bottlenecks?
   - What steps always happen together?
   - Example findings:
     - "User corrected to read vision.md first - 3 times"
     - "Always runs tests after code changes"
     - "Classifies features in every feature request"

2. **LLM generates new workflow**:
   - Creates modified workflow YAML
   - Adds/removes/modifies nodes and edges
   - Includes reasoning in comments or separate file
   - Increments version number

3. **Save new version**:
   - Write `workflows/assistant_v2.yaml`
   - Write `workflows/assistant_v2_reasoning.md` (explains what changed and why)
   - Keep old version (v1) intact

**Output**:
```
workflows/
  assistant_v1.yaml          # original
  assistant_v2.yaml          # new version from reflection
  assistant_v2_reasoning.md  # why changes were made
```

**User workflow**:
1. Use v1 for a while → logs accumulate
2. Run reflection → v2 created automatically
3. Review v2 and reasoning
4. Decide: keep using v1, try v2, or edit v2 manually
5. Use whichever version preferred
6. Repeat: run reflection on v2 → creates v3

**No approval needed**: User chooses which version to use, reflection just creates options

## MVP Scope

### Must Have

**Workflow Definition**:
- ✓ YAML format for workflows
- ✓ Node structure (id, type, config)
- ✓ Edge structure (from, to, optional condition)
- ✓ One node type: CLAUDE_CODE
- ✓ Decide: How to invoke workflows (CLI, library, wrapper)

**Execution + Logging**:
- ✓ Execute workflow (load YAML, run nodes, follow edges)
- ✓ Log executions (run-level data)
- ✓ Log node steps (node-level data)
- ✓ Logging config TBD (should support per-node configuration)
- ✓ Abstract storage interface (files for MVP, can add DB later)
- ✓ Store logs as files (one file per execution)

**Metrics**:
- ✓ Separate from logging (aggregated summary data)
- ✓ Track: duration, success/failure, patterns
- ✓ Abstract storage interface (files for MVP)
- ✓ Used by reflection to find patterns

**Reflection + Learning**:
- ✓ Manual trigger: `darwinflow reflect`
- ✓ LLM analyzes execution logs (find patterns)
- ✓ LLM generates new workflow YAML (v2, v3, etc.)
- ✓ Save new version + reasoning document
- ✓ User chooses which version to use (no approval flow needed)

### Nice to Have (Later)

- Multiple node types (LLM_DECISION, TOOL_CALL, HUMAN_INPUT, etc.)
- Automatic reflection triggers
- Real-time learning (during execution)
- Branching/conditional flows
- Parallel execution
- Sub-workflows
- Visual workflow editor
- Metrics dashboard

## Implementation Order

### Step 1: Workflow Definition & Execution
- Define workflow YAML structure
- Decide on invocation method (CLI, library, or wrapper)
- Build workflow executor (reads YAML, executes nodes)
- Implement CLAUDE_CODE node (delegates to Claude Code)
- Test: Can execute simple workflow end-to-end

### Step 2: Add Logging & Metrics
- Define abstract storage interface (read/write logs, query by run_id)
- Implement file-based storage (one file per execution)
- Implement execution logger
- Implement metrics tracker (separate from logs)
- Integrate with workflow executor
- Note: Logging config mechanism is TBD
- Test: Can query "what happened in run X?" from log files

### Step 3: Add Reflection
- Build `darwinflow reflect` command
- LLM analyzes logs to find patterns
- LLM generates new workflow YAML (v2, v3, etc.)
- Save new version + reasoning document
- Test: Reflection creates new workflow version from real logs

## Success Criteria

**Step 1 Success**:
- Can define workflow in YAML
- Can invoke and execute workflow with CLAUDE_CODE node
- Workflow behaves same as current Claude Code
- Decision made: How workflows are invoked

**Step 2 Success**:
- Abstract storage interface defined
- All executions logged to files
- Metrics tracked separately from logs
- Can query: "What did system do in run X?" from log files
- Can review: tool calls, files accessed, duration, timing
- 20+ executions fully logged
- Metrics showing patterns (most common operations, durations, etc.)

**Step 3 Success**:
- `darwinflow reflect` command works
- LLM analyzes logs and identifies patterns
- New workflow version created automatically
- Reasoning document explains what changed and why
- Can use new version and see different behavior
- Measurable: 1+ workflow evolution from real usage patterns

## What This Enables

### Short term:
- Transparent workflow execution (see what happens)
- Full execution history (replay, debug)
- Foundation for learning

### Medium term:
- Workflow improvements from usage patterns
- Reduced manual corrections
- Personalized workflow evolution

### Long term:
- Multiple specialized workflows
- Automatic optimization
- Community-shared workflows
- Cross-user learning

## Open Questions

1. **Invocation method**: CLI, library, or wrapper? Which is best for MVP?
2. **Logging config**: How to make it configurable per node? YAML field or separate file?
3. **Node type expansion**: What's the 2nd node type to add after CLAUDE_CODE?
4. **Reflection improvements**: How to guide LLM to generate good workflow changes?
5. **Version selection**: Should there be a "default" version or always explicit?

## Example Evolution

**v1 (Initial)** - `workflows/assistant_v1.yaml`:
```yaml
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

**v2 (After learning "read vision.md first")** - `workflows/assistant_v2.yaml`:
```yaml
id: assistant
version: 2
nodes:
  - id: read_vision
    type: READ_FILE
    config:
      file: docs/product/vision.md
  - id: call_claude
    type: CLAUDE_CODE
    config:
      pass_full_request: true
      context_includes: [read_vision.output]
edges:
  - from: START
    to: read_vision
  - from: read_vision
    to: call_claude
  - from: call_claude
    to: END
```

**v3 (After learning "classify features")** - `workflows/assistant_v3.yaml`:
```yaml
id: assistant
version: 3
nodes:
  - id: read_vision
    type: READ_FILE
    config:
      file: docs/product/vision.md
  - id: classify_feature
    type: LLM_DECISION
    config:
      prompt: "Is this Core/Integration/Workflow?"
      context: [read_vision.output, user_request]
  - id: call_claude
    type: CLAUDE_CODE
    config:
      pass_full_request: true
      context_includes: [classification]
edges:
  - from: START
    to: read_vision
  - from: read_vision
    to: classify_feature
  - from: classify_feature
    to: call_claude
  - from: call_claude
    to: END
```

Each evolution emerges from actual usage patterns found in logs, not upfront design.

**Reasoning files** (created alongside):
- `workflows/assistant_v2_reasoning.md`: Explains why read_vision node was added
- `workflows/assistant_v3_reasoning.md`: Explains why classify_feature node was added

---

**Core principle**: Start with simplest possible (one node), add structure to support evolution, let complexity emerge from real patterns.

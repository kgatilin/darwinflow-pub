# Core Framework Interfaces

**Author**: Bob Martin
**Date**: 2025-10-09
**Status**: Target Architecture - Interface Definitions
**Version**: 1.0

---

## Purpose

This document defines the stable contracts that form the core framework. These interfaces enable progressive evolution without breaking changes.

**Key principle**: Interfaces are stable. Implementations evolve.

---

## Domain Entities

### Workflow Definition

The workflow is pure data, not behavior. It's the "program" that the engine executes.

```
Workflow
  - ID: WorkflowID
  - Version: int
  - Description: string
  - Nodes: []Node
  - Edges: []Edge
  - Metadata: WorkflowMetadata

WorkflowMetadata
  - CreatedAt: timestamp
  - CreatedBy: string (user/reflection)
  - ParentID: *WorkflowID (version lineage)
  - Reasoning: string
  - Tags: map[string]string

Operations:
  - Validate() -> error
  - GetNode(id) -> Node, error
  - GetOutgoingEdges(nodeID) -> []Edge
```

**Design notes**:
- Workflow is a value object - immutable once created
- Version is explicit - v1, v2, v3... no implicit "latest"
- Metadata tracks evolution - parent chain shows history
- Validation happens at load time, not execution time

### Node Definition

Nodes are the units of work in a workflow.

```
Node
  - ID: NodeID
  - Type: NodeType
  - Description: string
  - Config: NodeConfig (type-specific configuration)
  - LogConfig: *LoggingConfig (optional per-node logging control)

Standard NodeTypes:
  - CLAUDE_CODE
  - LLM_DECISION
  - TOOL_CALL
  - FILE_READ
  - FILE_WRITE
  - HUMAN_INPUT
  - SUB_WORKFLOW

NodeConfig: map[string]interface{} (extensible, type-specific)

LoggingConfig
  - LogInputs: bool
  - LogOutputs: bool
  - LogDuration: bool
  - LogErrors: bool
  - CustomFields: map[string]bool
```

**Design notes**:
- NodeType is a string - allows dynamic registration
- Config is flexible map - each node type interprets differently
- LogConfig is optional - defaults to workflow-level settings
- Node is immutable value object

### Edge Definition

Edges define the flow between nodes.

```
Edge
  - ID: EdgeID
  - From: NodeID
  - To: NodeID
  - Condition: *Condition (nil = unconditional)

Condition
  - Expression: string (e.g., "output.decision == 'bug_fix'")
  - Type: ConditionType (EXPRESSION, ALWAYS, SCRIPT)

ConditionTypes:
  - ALWAYS (default, no evaluation)
  - EXPRESSION (simple expression evaluation)
  - SCRIPT (full script evaluation)

Operations:
  - Evaluate(context) -> bool, error
```

**Design notes**:
- Edges start simple (unconditional) but support conditions later
- Condition evaluation is extensible
- MVP only needs `ConditionAlways` - others added in Phase 2

### Special Nodes

```
Reserved NodeIDs:
  - START (every workflow must have)
  - END (every workflow must have)
```

---

## Execution Interfaces

### WorkflowExecutor

The engine that runs workflows.

```
Interface: WorkflowExecutor
  - Execute(ctx, workflow, input) -> output, error
  - ExecuteWithCallback(ctx, workflow, input, callback) -> output, error

Input
  - Request: string (user's request/task)
  - Context: map[string]interface{} (additional context)
  - RunID: RunID (unique execution identifier)

Output
  - Result: interface{} (final result)
  - Status: ExecutionStatus (SUCCESS, FAILED, CANCELLED)
  - Error: error (nil if success)
  - Duration: duration
  - Metadata: map[string]interface{} (extensible)

Interface: ExecutionCallback
  - OnStepStart(step)
  - OnStepComplete(step, result)
  - OnStepError(step, error)
  - OnWorkflowComplete(output)
```

**Design notes**:
- Context support for cancellation
- Callback pattern for observability without coupling
- Input/Output are value objects
- Executor doesn't know about storage - that's injected

### NodeExecutor

Interface that all node types implement.

```
Interface: NodeExecutor
  - Type() -> NodeType
  - Execute(ctx, config) -> result, error
  - ValidateConfig(config) -> error

Interface: ExecutionContext (provided to nodes)
  - RunID() -> RunID
  - GetInput() -> Input
  - GetNodeOutput(nodeID) -> interface{}, error
  - SetNodeOutput(nodeID, output)
  - Logger() -> ExecutionLogger
  - Metrics() -> MetricsCollector
  - IsCancelled() -> bool

NodeResult
  - Output: interface{} (node-specific output)
  - Metadata: map[string]interface{} (duration, tokens, cost, etc.)
  - Error: error (nil if success)
```

**Design notes**:
- NodeExecutor is the plugin interface
- ExecutionContext provides everything a node needs
- Context is interface - can be mocked for testing
- Nodes are stateless - all state in context

### NodeRegistry

Plugin system for node types.

```
Interface: NodeRegistry
  - Register(executor) -> error
  - Get(nodeType) -> executor, error
  - List() -> []NodeType
  - Unregister(nodeType) -> error

Usage pattern:
  registry.Register(NewClaudeCodeNode())
  executor := registry.Get("CLAUDE_CODE")
```

**Design notes**:
- Registry is a service locator (acceptable here - limited scope)
- Executors registered at startup
- Get() returns error if type not found
- Allows testing with mock executors

---

## Storage Interfaces

### WorkflowRepository

Manages workflow definitions.

```
Interface: WorkflowRepository
  - Store(workflow) -> error
  - Get(id, version) -> workflow, error
  - GetLatest(id) -> workflow, error
  - List(id) -> []workflow, error
  - ListAll() -> []workflow, error
  - Delete(id, version) -> error
```

**Design notes**:
- Versioning is explicit - no implicit "latest"
- Get() requires version number
- Storage mechanism is opaque (files, DB, S3, etc.)

### LogRepository

Manages execution logs.

```
Interface: LogRepository
  - StoreLog(log) -> error
  - GetLog(runID) -> log, error
  - QueryLogs(filter) -> []log, error
  - GetRecent(workflowID, version, limit) -> []log, error
  - Delete(runID) -> error

ExecutionLog
  - RunID: RunID
  - WorkflowID: WorkflowID
  - WorkflowVersion: int
  - Input: Input
  - Output: Output
  - StartTime: timestamp
  - EndTime: timestamp
  - Steps: []StepLog
  - Metadata: map[string]interface{}

StepLog
  - StepID: string
  - RunID: RunID
  - NodeID: NodeID
  - NodeType: NodeType
  - StartTime: timestamp
  - EndTime: timestamp
  - Input: interface{}
  - Output: interface{}
  - Error: error
  - Actions: []ActionLog
  - Metadata: map[string]interface{}

ActionLog (extensible)
  - Type: ActionType (TOOL_CALL, FILE_READ, FILE_WRITE, LLM_CALL, USER_INPUT)
  - Timestamp: timestamp
  - Details: map[string]interface{}

LogFilter
  - WorkflowID: *WorkflowID
  - WorkflowVersion: *int
  - Status: *ExecutionStatus
  - StartTime: *timestamp
  - EndTime: *timestamp
  - Limit: int
```

**Design notes**:
- Logs are comprehensive - everything needed for reflection
- Actions are extensible - each node type can log custom actions
- Query interface supports analysis
- Separation: detailed logs vs aggregated metrics

### MetricsRepository

Manages aggregated metrics (separate from detailed logs).

```
Interface: MetricsRepository
  - RecordExecution(metric) -> error
  - GetSummary(workflowID, version) -> summary, error
  - GetPatterns(workflowID) -> []pattern, error
  - RecordPattern(pattern) -> error

ExecutionMetric
  - RunID: RunID
  - WorkflowID: WorkflowID
  - WorkflowVersion: int
  - Duration: duration
  - Success: bool
  - NodesExecuted: int
  - TokensUsed: int
  - Cost: float
  - Timestamp: timestamp

WorkflowSummary
  - WorkflowID: WorkflowID
  - WorkflowVersion: int
  - TotalExecutions: int
  - SuccessfulExecutions: int
  - FailedExecutions: int
  - AvgDuration: duration
  - TotalTokens: int
  - TotalCost: float
  - CommonOperations: []OperationFrequency
  - LastUpdated: timestamp

OperationFrequency
  - Operation: string (e.g., "file_read: vision.md")
  - Count: int

Pattern
  - ID: PatternID
  - WorkflowID: WorkflowID
  - Type: PatternType (REPEATED_ACTION, CORRECTION, SEQUENCE, BOTTLENECK)
  - Description: string
  - Occurrences: int
  - Confidence: float (0.0-1.0)
  - FirstSeen: timestamp
  - LastSeen: timestamp
  - Details: map[string]interface{}
```

**Design notes**:
- Metrics are aggregated, not raw logs
- Updated after each execution (can be async)
- Patterns tracked separately for reflection
- Summary gives high-level view without log details

---

## Learning Interfaces

### PatternAnalyzer

Identifies patterns in execution logs.

```
Interface: PatternAnalyzer
  - Analyze(logs) -> []pattern, error
  - AnalyzeWithContext(logs, context) -> []pattern, error

AnalysisContext
  - CurrentWorkflow: Workflow
  - ExistingPatterns: []Pattern
  - UserFeedback: []Feedback
  - Thresholds: AnalysisThresholds

Feedback
  - RunID: RunID
  - Type: FeedbackType (CORRECTION, SUGGESTION, APPROVAL)
  - Message: string
  - Timestamp: timestamp
  - NodeID: *NodeID

AnalysisThresholds
  - MinOccurrences: int (pattern must appear N times)
  - MinConfidence: float (pattern must have X confidence)
```

**Design notes**:
- Analyzer is stateless - pure function
- Takes logs, produces patterns
- Context allows richer analysis (user feedback!)
- Thresholds make it configurable

### ReflectionEngine

Generates new workflow versions based on patterns.

```
Interface: ReflectionEngine
  - Reflect(request) -> result, error

ReflectionRequest
  - CurrentWorkflow: Workflow
  - Logs: []ExecutionLog
  - Patterns: []Pattern
  - Constraints: ReflectionConstraints

ReflectionConstraints
  - MaxNodes: int
  - AllowedNodeTypes: []NodeType
  - PreserveNodes: []NodeID
  - Hints: []string

ReflectionResult
  - NewWorkflow: Workflow
  - Reasoning: ReflectionReasoning
  - Confidence: float

ReflectionReasoning
  - Summary: string
  - PatternsApplied: []PatternID
  - Changes: []WorkflowChange
  - Risks: []string

WorkflowChange
  - Type: ChangeType (NODE_ADDED, NODE_REMOVED, NODE_MODIFIED, EDGE_ADDED, EDGE_REMOVED, EDGE_MODIFIED)
  - Description: string
  - NodeID: *NodeID
  - EdgeID: *EdgeID
```

**Design notes**:
- Reflection is stateless
- Takes current state, produces new version
- Reasoning is first-class - explains changes
- Constraints allow controlling output
- Confidence score allows filtering low-quality generations

### WorkflowGenerator

Low-level interface for creating workflows (used by ReflectionEngine).

```
Interface: WorkflowGenerator
  - Generate(spec) -> workflow, error
  - Validate(workflow) -> error
  - Optimize(workflow, goals) -> workflow, error

WorkflowSpec
  - Description: string
  - BaseWorkflow: *Workflow (optional starting point)
  - RequiredNodes: []NodeType
  - Constraints: map[string]interface{}

OptimizationGoals
  - MinimizeDuration: bool
  - MinimizeCost: bool
  - MaximizeReliability: bool
  - Weights: map[string]float
```

**Design notes**:
- Generator is lower-level than ReflectionEngine
- Can be used independently for creating workflows
- Optimization is explicit - choose what to optimize for

---

## Logging & Observability Interfaces

### ExecutionLogger

Captures execution details.

```
Interface: ExecutionLogger
  - LogStepStart(runID, nodeID) -> error
  - LogStepComplete(runID, nodeID, result) -> error
  - LogStepError(runID, nodeID, error) -> error
  - LogAction(runID, nodeID, action) -> error
  - Flush() -> error
```

**Design notes**:
- Logger is separate from repository
- Logger buffers, repository persists
- Nodes call LogAction for custom events

### MetricsCollector

Captures metrics during execution.

```
Interface: MetricsCollector
  - RecordDuration(nodeID, duration)
  - RecordSuccess(runID)
  - RecordFailure(runID, reason)
  - RecordTokens(nodeID, tokens)
  - RecordCost(nodeID, cost)
  - RecordCustom(name, value)
```

**Design notes**:
- Metrics are recorded during execution
- Separate from logging (different concerns)
- Custom metrics for extensibility

---

## Utility Interfaces

### ConfigLoader

Loads workflow definitions from storage.

```
Interface: ConfigLoader
  - LoadFromFile(path) -> workflow, error
  - LoadFromString(yaml) -> workflow, error
  - LoadFromRepository(repo, id, version) -> workflow, error
```

### ConfigWriter

Persists workflow definitions.

```
Interface: ConfigWriter
  - WriteToFile(workflow, path) -> error
  - WriteToString(workflow) -> string, error
  - WriteToRepository(workflow, repo) -> error
```

---

## Extension Points

### Custom Node Types

Anyone can create new node types by implementing the NodeExecutor interface.

**Pattern**:
1. Implement `NodeExecutor` interface
2. Implement `Type()`, `Execute()`, `ValidateConfig()`
3. Register with `NodeRegistry`

### Custom Storage Backends

Anyone can implement storage by implementing repository interfaces.

**Pattern**:
1. Implement `WorkflowRepository`, `LogRepository`, or `MetricsRepository`
2. Honor interface contracts
3. Inject into core framework

---

## Interface Stability Guarantees

### Will NEVER Change (Stable)

- `WorkflowExecutor.Execute()`
- `NodeExecutor.Execute()`
- `LogRepository.GetLog()`
- `MetricsRepository.GetSummary()`

These are the foundation. Breaking these breaks everything.

### MAY Grow (Backward Compatible)

- `ExecutionContext` (new methods added)
- `NodeConfig` (new fields added)
- `ActionType` (new types added)
- `PatternType` (new types added)

Growth is acceptable, removal is not.

### CAN Change (Unstable Until Proven)

- `ReflectionConstraints`
- `OptimizationGoals`
- `AnalysisThresholds`

These may evolve as we learn what works.

---

## Testing Strategy

All interfaces are designed for testability through mocking:

- Use mock contexts for testing nodes
- Use mock storage for testing executors
- Use mock executors for testing workflows
- Interfaces make all dependencies injectable

---

## Summary

These interfaces form the stable core that enables evolution:

1. **Domain entities** (Workflow, Node, Edge) are immutable value objects
2. **Execution interfaces** (Executor, NodeExecutor) define behavior contracts
3. **Storage interfaces** (Repositories) abstract persistence
4. **Learning interfaces** (Analyzer, ReflectionEngine) enable adaptation
5. **Extension points** (Registry, custom nodes) allow growth

If implementations respect these contracts, the system can evolve indefinitely without breaking changes.

---

**Next**: Read [02_integration_architecture.md](02_integration_architecture.md) for how integrations plug into these interfaces.

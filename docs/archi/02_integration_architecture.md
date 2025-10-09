# Integration Architecture

**Author**: Bob Martin
**Date**: 2025-10-09
**Status**: Target Architecture - Integration Layer
**Version**: 1.0

---

## Purpose

Integrations are the adapters between the core framework and external tools. They implement core interfaces but live outside the framework.

**Key principle**: Integrations are plugins, not core components.

---

## Integration Layer Responsibilities

### What Integrations DO:

✓ Implement `NodeExecutor` interface
✓ Wrap external tools (Claude Code, Telegram, GitHub, etc.)
✓ Handle tool-specific logic (authentication, rate limiting, error handling)
✓ Translate between tool formats and framework contracts
✓ Provide reusable capabilities across workflows

### What Integrations DON'T DO:

✗ Contain domain logic (that's in workflows)
✗ Make assumptions about which workflows use them
✗ Depend on each other (loose coupling)
✗ Modify core framework code

---

## Integration Types

### 1. Tool Adapters (External Services)

Wrap existing tools and services.

**Dependency flow**:
```
External Tool (Claude Code, etc.)
        ↑ wraps
Tool Adapter (implements NodeExecutor)
        ↑ uses
Core Framework (WorkflowExecutor)
```

**Examples**:
- ClaudeCodeAdapter
- TelegramBotAdapter
- GitHubAdapter
- SlackAdapter

### 2. Primitive Operations (Built-in)

Implement common operations directly without external tools.

**Examples**:
- FileReadNode (reads files)
- FileWriteNode (writes files)
- LLMDecisionNode (calls LLM for decision)
- ScriptNode (runs shell commands)

### 3. Composite Operations (High-Level)

Combine other node types or delegate to sub-systems.

**Examples**:
- SubWorkflowNode (runs nested workflow)
- ParallelNode (executes nodes in parallel)
- RetryNode (wraps node with retry logic)

---

## Node Type Contract

All integrations implement the same interface:

```
Interface: NodeExecutor
  - Type() -> NodeType
  - Execute(ctx, config) -> result, error
  - ValidateConfig(config) -> error
```

**Contract requirements**:
- `Type()` must return unique node type identifier
- `Execute()` must be idempotent when possible
- `ValidateConfig()` must validate config at load time, not execution time
- Must log all external calls via `ctx.Logger()`
- Must record metrics via `ctx.Metrics()`
- Must respect cancellation via `ctx.IsCancelled()`

---

## Standard Integration Library

### MVP Node Types (Phase 1)

Start with minimum viable set:

| Node Type | Purpose | External Tool? |
|-----------|---------|----------------|
| `CLAUDE_CODE` | Delegate to Claude Code | Yes (Claude Code CLI) |

That's it for MVP. One node type. Ship it.

### Early Expansion (Phase 2)

Add common operations:

| Node Type | Purpose | External Tool? |
|-----------|---------|----------------|
| `FILE_READ` | Read file contents | No |
| `FILE_WRITE` | Write file contents | No |
| `LLM_DECISION` | Make decision via LLM | Yes (LLM API) |
| `TOOL_CALL` | Execute tool command | Yes (varies) |
| `HUMAN_INPUT` | Wait for user input | Yes (UI/chat) |

### Mature System (Phase 3+)

Full library:

| Node Type | Purpose | External Tool? |
|-----------|---------|----------------|
| `SUB_WORKFLOW` | Execute nested workflow | No |
| `PARALLEL` | Run nodes in parallel | No |
| `CONDITIONAL` | Branch based on condition | No |
| `LOOP` | Iterate over collection | No |
| `RETRY` | Retry on failure | No |
| `DEBOUNCE` | Rate limit execution | No |
| `CACHE` | Cache node output | No |
| `TRANSFORM` | Transform data | No |

---

## Integration Configuration

### Config Schema per Node Type

Each node type defines its own config schema:

**ClaudeCodeNode**:
```
{
  "request": "string (required)",
  "interactive": "boolean (optional, default: true)",
  "timeout": "duration (optional, default: 30m)"
}
```

**FileReadNode**:
```
{
  "path": "string (required)",
  "encoding": "string (optional, default: utf-8)"
}
```

**LLMDecisionNode**:
```
{
  "prompt": "string (required)",
  "options": "[]string (required)",
  "context": "[]nodeID (optional)",
  "model": "string (optional, default: claude-3-5-sonnet)"
}
```

### Config Validation

Validation occurs at workflow load time, not execution time:

**Validation flow**:
1. Workflow loaded from YAML
2. For each node, get executor from registry
3. Call `executor.ValidateConfig(node.Config)`
4. If any validation fails, reject workflow

**This ensures**: Workflows fail fast if misconfigured, not during execution.

---

## Integration Patterns

### Pattern 1: External Tool Adapter

Wraps an external CLI tool or API.

**Responsibilities**:
- Execute external tool with proper arguments
- Parse tool output into structured format
- Handle tool-specific errors
- Log tool invocations
- Respect timeouts

**Example**: ClaudeCodeNode wraps Claude Code CLI

### Pattern 2: Primitive Operation

Implements operation directly, no external tool.

**Responsibilities**:
- Perform operation using standard libraries
- Handle errors gracefully
- Log actions taken
- Return structured results

**Example**: FileReadNode reads files using filesystem operations

### Pattern 3: LLM-Powered Operation

Uses LLM for logic, but within framework.

**Responsibilities**:
- Build prompts from config and context
- Call LLM provider abstraction
- Parse LLM responses
- Log LLM calls and token usage
- Record costs

**Example**: LLMDecisionNode makes decisions using LLM provider interface

---

## Integration Registration

Integrations register themselves at startup:

**Registration flow**:
1. Create node executor instance
2. Call `registry.Register(executor)`
3. Registry validates unique type
4. Executor available for workflows

**Example workflow**:
```
registry := core.DefaultRegistry()
registry.Register(NewClaudeCodeNode("/path/to/claude"))
registry.Register(&FileReadNode{})
registry.Register(NewLLMDecisionNode(llmProvider))
```

---

## Integration Guidelines

### DO:

✓ **Keep integrations thin** - wrap, don't implement business logic
✓ **Log all external calls** - transparency is critical
✓ **Validate config thoroughly** - fail fast at load time
✓ **Handle errors gracefully** - return structured errors
✓ **Use timeouts** - don't hang forever on external calls
✓ **Make them reusable** - multiple workflows should benefit
✓ **Be stateless** - use ExecutionContext for state

### DON'T:

✗ **Make assumptions about usage** - don't couple to specific workflows
✗ **Depend on other integrations** - stay independent
✗ **Store state** - nodes should be stateless (use ExecutionContext)
✗ **Perform expensive operations in ValidateConfig** - that's load time
✗ **Silently swallow errors** - propagate them up
✗ **Hardcode credentials** - use config or environment variables

---

## Security Considerations

### Credential Management

**Never hardcode credentials**:
- Use environment variables
- Use configuration files (outside version control)
- Use credential providers/vaults
- Abstract through interfaces (TokenProvider, etc.)

### Input Validation

**Always validate inputs**:
- Sanitize file paths (prevent directory traversal)
- Validate command arguments (prevent injection)
- Check URL formats (prevent SSRF)
- Limit input sizes (prevent DoS)

### Sandboxing

**Consider sandboxing for unsafe operations**:
- Script execution (use containers, VMs, or chroot)
- File operations (restrict to allowed directories)
- Network calls (whitelist domains)

---

## Integration Lifecycle

### Initialization

Executed once at startup:
- Verify external tools exist
- Check credentials/API keys
- Establish connections if needed
- Register with node registry

### Execution

Executed per node invocation:
1. Validate context
2. Log action start
3. Perform operation
4. Log action complete
5. Return structured result

### Cleanup

Nodes are stateless, so minimal cleanup needed.

**If cleanup required**: Implement optional `Close()` method for resource cleanup.

---

## Community Integrations

Allow community to contribute integrations:

**Structure**:
```
darwinflow-integrations/
  ├── official/          # Maintained by core team
  │   ├── claude-code/
  │   ├── telegram/
  │   └── github/
  └── community/         # Community contributions
      ├── jira/
      ├── slack/
      ├── notion/
      └── linear/
```

**Each integration is a separate module**:
- Self-contained Go module
- Implements NodeExecutor interface
- Has its own dependencies
- Can be installed independently

---

## Extension Points

### Adding New Node Types

**Process**:
1. Create new struct implementing `NodeExecutor`
2. Implement `Type()`, `Execute()`, `ValidateConfig()`
3. Register with `NodeRegistry` at startup
4. Document config schema
5. Write tests

**Estimated effort**: 50-100 LOC per node type

### LLM Provider Abstraction

For nodes that use LLMs:

```
Interface: LLMProvider
  - Complete(prompt) -> response, tokens, error
  - Model() -> string

Implementations:
  - AnthropicProvider (Claude)
  - OpenAIProvider (GPT)
  - LocalProvider (Ollama, etc.)
```

This allows swapping LLM providers without changing node implementations.

---

## Testing Integrations

Each integration should have comprehensive tests:

**Test coverage**:
- Unit tests for Execute() with various configs
- Unit tests for ValidateConfig() edge cases
- Integration tests with mock external tools
- Error handling tests

**Testing strategy**:
- Use mock ExecutionContext
- Mock external tool calls
- Verify logging calls
- Verify metrics recorded
- Test timeout handling
- Test error propagation

---

## Summary

Integrations are plugins that:

1. **Implement NodeExecutor interface** - standard contract
2. **Wrap external tools** - adapt to framework
3. **Stay thin** - no business logic
4. **Are reusable** - multiple workflows benefit
5. **Register at startup** - pluggable architecture

This design allows unlimited growth without touching core framework.

---

**Next**: Read [03_storage_architecture.md](03_storage_architecture.md) for storage implementation details.

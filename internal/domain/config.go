package domain

// Config represents the DarwinFlow configuration
type Config struct {
	// Analysis contains analysis execution settings
	Analysis AnalysisConfig `yaml:"analysis" json:"analysis"`

	// Prompts contains named prompts for different use cases
	Prompts map[string]string `yaml:"prompts" json:"prompts"`
}

// AnalysisConfig contains settings for analysis execution
type AnalysisConfig struct {
	// TokenLimit is the maximum tokens for analysis context (default: 100000)
	TokenLimit int `yaml:"token_limit" json:"token_limit"`

	// Model is the Claude model to use (default: "claude-sonnet-4-5-20250929")
	Model string `yaml:"model" json:"model"`

	// ParallelLimit is the max parallel analysis executions (default: 3)
	ParallelLimit int `yaml:"parallel_limit" json:"parallel_limit"`

	// AutoSummaryEnabled enables auto-triggered session summaries on session end (default: false)
	AutoSummaryEnabled bool `yaml:"auto_summary_enabled" json:"auto_summary_enabled"`

	// AutoSummaryPrompt is the prompt name to use for auto summaries (default: "session_summary")
	AutoSummaryPrompt string `yaml:"auto_summary_prompt" json:"auto_summary_prompt"`

	// ClaudeOptions contains options for Claude CLI execution
	ClaudeOptions ClaudeOptions `yaml:"claude_options" json:"claude_options"`
}

// ClaudeOptions contains Claude CLI execution options
type ClaudeOptions struct {
	// AllowedTools is the list of allowed tools (empty = no tools)
	AllowedTools []string `yaml:"allowed_tools" json:"allowed_tools"`

	// SystemPromptMode determines how to use prompts: "replace" or "append"
	SystemPromptMode string `yaml:"system_prompt_mode" json:"system_prompt_mode"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Analysis: AnalysisConfig{
			TokenLimit:         100000,
			Model:              "claude-sonnet-4-5-20250929",
			ParallelLimit:      3,
			AutoSummaryEnabled: false,           // Disabled by default - user must opt in
			AutoSummaryPrompt:  "session_summary", // Use session_summary prompt for auto-analysis
			ClaudeOptions: ClaudeOptions{
				AllowedTools:     []string{}, // No tools for pure analysis
				SystemPromptMode: "replace",  // Use --system-prompt
			},
		},
		Prompts: map[string]string{
			"session_summary": DefaultSessionSummaryPrompt,
			"tool_analysis":   DefaultToolAnalysisPrompt,
		},
	}
}

// DefaultSessionSummaryPrompt is used for auto-triggered session summaries
const DefaultSessionSummaryPrompt = `Analyze this Claude Code session and provide a concise summary.

## Your Task

Summarize what happened in this session, focusing on:

1. **User Intent**: What did the user want to accomplish?
2. **Goal Achievement**: Was the goal achieved? Fully, partially, or not at all?
3. **Approach Taken**: What approach/steps were taken?
4. **Outcomes**: What was actually accomplished?
5. **Issues Encountered**: Any errors, blockers, or challenges?
6. **Session Quality**: Was this session efficient and successful?

## Output Format

### User Intent
[What the user wanted to accomplish]

### Goal Achievement
[Achieved | Partially Achieved | Not Achieved]

### Approach Summary
[Brief summary of the approach taken]

### Outcomes
[What was accomplished - be specific and factual]

### Issues Encountered
[Any problems, errors, or challenges - or "None" if smooth]

### Session Assessment
[Brief assessment of efficiency and success]

---

## Session to Analyze

`

// DefaultToolAnalysisPrompt analyzes tool usage patterns across sessions
const DefaultToolAnalysisPrompt = `You are Claude Code, an AI agent analyzing your own work across multiple sessions.

## Your Task

Analyze the sessions below and identify what tools YOU need to make YOUR work FASTER and more efficient.

Review these sessions and identify where YOU (the agent) were inefficient due to lack of tools. **Your goal is to minimize execution time and tool call count while completing tasks correctly.**

Specifically look for:

1. **Repetitive Low-Level Operations**: Where you had to perform multiple primitive operations that could be a single tool
   - Example: Multiple Read calls to gather project context (same files read repeatedly across sessions)
   - Example: Sequential Grep/Glob operations that could be one complex search

2. **Missing Specialized Agents**: Task types that would benefit from dedicated subagents with specialized capabilities

3. **Tool Gaps**: Operations you struggled with or had to work around due to missing functionality

4. **Workflow Inefficiencies**: Multi-step sequences you repeat that should be automated

5. **Performance Bottlenecks**: Analyze the relationship between:
   - Task complexity (simple vs complex)
   - Execution time (how long it took)
   - Tool call count (how many tools you invoked)
   - Pattern: If you see the same context-gathering reads across sessions, this is a pattern to optimize

**OPTIMIZATION PRINCIPLE**: Favor creating fewer, more complex tools over using many simple tools. If a pattern requires 10 tool calls, consider whether a single specialized tool could do it in 1-2 calls.

## Tool Categories to Consider

- **Specialized Agents**: Subagents with specific expertise (e.g., test generation, refactoring, documentation)
- **CLI Tools**: Command-line utilities that could be invoked via Bash to augment your capabilities
- **Claude Code Features**: New tools or capabilities that should be built into Claude Code itself
- **Workflow Automations**: Multi-step operations that should be single tool calls

## Output Format

Write your analysis from YOUR perspective as the agent. Use first person.

### What Made Me Inefficient

Describe specific moments where you were slow or used too many tool calls across these sessions. Include:
- **Repetitive patterns**: Did you perform the same operations multiple times?
- **Tool call bloat**: Where did you use many tools when fewer complex tools would suffice?
- **Speed bottlenecks**: What took longer than it should for the task complexity?

Provide concrete examples from the sessions with tool counts and patterns.

### Tools I Need

For each tool you need, state:

**Tool: [Name]**
- **What I Need**: Clear description of the capability
- **Why I Need It**: How it would make YOUR work FASTER and reduce tool calls
  - Current: X tool calls taking Y time
  - With tool: Z tool calls taking W time
- **Type**: [Specialized Agent | CLI Tool | Claude Code Feature | Workflow Automation]
- **How I Would Use It**: Concrete example showing how you would invoke it
- **Implementation Note**: Brief technical approach

### Priority Order

List the tools in priority order based on:
1. **Speed impact**: How much faster would this tool make you?
2. **Tool call reduction**: How many tool calls would this eliminate?
3. **Frequency of need**: How often do you encounter this pattern?
4. **Time saved per invocation**: Seconds/minutes saved each use

Write as: "To make my work faster and more efficient, I need: [ordered list]"

---

## Sessions to Analyze

`

// DefaultAnalysisPrompt is kept for backward compatibility
const DefaultAnalysisPrompt = DefaultToolAnalysisPrompt

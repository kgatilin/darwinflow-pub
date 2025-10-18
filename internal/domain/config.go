package domain

// Config represents the DarwinFlow configuration
type Config struct {
	// Prompts contains named prompts for different use cases
	Prompts map[string]string `yaml:"prompts" json:"prompts"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Prompts: map[string]string{
			"analysis": DefaultAnalysisPrompt,
		},
	}
}

// DefaultAnalysisPrompt is the default template for session analysis
const DefaultAnalysisPrompt = `You are analyzing a Claude Code session to identify patterns and suggest optimizations.

## Your Task

Analyze the session below and identify:

1. **Repetitive Patterns**: Operations that occur multiple times in similar ways
2. **Tool Usage Patterns**: Common sequences of tool invocations
3. **Workflow Patterns**: Recurring workflows or task structures
4. **Consolidation Opportunities**: Where multiple low-level operations could be combined into higher-level tools

## Focus Areas

Look for opportunities to create:
- **Conceptual Tools**: High-level operations that accomplish specific goals (e.g., "add domain component", "implement feature with tests")
- **Workflow Tools**: Tools that execute common multi-step workflows
- **Aggregation Tools**: Tools that combine multiple similar operations into one
- **Context-Aware Tools**: Tools that understand project structure and conventions

## Output Format

Provide your analysis in the following structure:

### Patterns Identified
[List the patterns you observed, with examples from the session]

### Tool Suggestions
For each suggested tool, provide:
- **Name**: Clear, action-oriented name
- **Description**: What the tool does
- **Rationale**: Why this tool would be valuable based on observed patterns
- **Example Usage**: How it would be used in practice
- **Implementation Approach**: High-level approach (can delegate to Claude, run script, etc.)

### Priority Recommendations
[Rank the suggestions by potential impact]

---

## Session to Analyze

`

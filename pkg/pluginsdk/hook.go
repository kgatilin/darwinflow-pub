package pluginsdk

// HookConfiguration describes a single hook provided by a plugin
type HookConfiguration struct {
	// TriggerType is the event type that triggers this hook
	// Examples: "trigger.tool.before", "trigger.user.input", "trigger.session.end"
	TriggerType string

	// Name is a human-readable name for the hook
	// Examples: "PreToolUse", "UserPromptSubmit"
	Name string

	// Description explains what this hook does
	Description string

	// Command is the CLI command to execute when the hook triggers
	// Example: "dw claude-code emit-event"
	Command string

	// Timeout is the maximum seconds this hook should take (0 = no timeout)
	Timeout int
}

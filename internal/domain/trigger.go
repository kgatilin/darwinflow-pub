package domain

// TriggerType represents system hook event types that can trigger event logging
// These are domain concepts representing the points in the interaction where hooks execute
type TriggerType string

const (
	// TriggerBeforeToolUse fires before any tool execution
	// Used by: Claude Code (PreToolUse hook)
	TriggerBeforeToolUse TriggerType = "trigger.tool.before"

	// TriggerAfterToolUse fires after tool execution
	TriggerAfterToolUse TriggerType = "trigger.tool.after"

	// TriggerUserInput fires when user provides input
	// Used by: Claude Code (UserPromptSubmit hook)
	TriggerUserInput TriggerType = "trigger.user.input"

	// TriggerSessionStart fires when session begins
	TriggerSessionStart TriggerType = "trigger.session.start"

	// TriggerSessionEnd fires when session ends
	// Used by: Claude Code (SessionEnd hook)
	TriggerSessionEnd TriggerType = "trigger.session.end"

	// Legacy names for backward compatibility - deprecated, do not use
	// These will be removed in a future version
	TriggerPreToolUse       TriggerType = "PreToolUse"
	TriggerPostToolUse      TriggerType = "PostToolUse"
	TriggerNotification     TriggerType = "Notification"
	TriggerUserPromptSubmit TriggerType = "UserPromptSubmit"
	TriggerStop             TriggerType = "Stop"
	TriggerSubagentStop     TriggerType = "SubagentStop"
	TriggerPreCompact       TriggerType = "PreCompact"
)

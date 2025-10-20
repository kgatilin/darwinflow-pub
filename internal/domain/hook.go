package domain

// HookInputData represents the data from a Claude Code hook invocation.
// This is a domain value object that captures hook metadata.
type HookInputData struct {
	SessionID      string
	TranscriptPath string
	CWD            string
	PermissionMode string
	HookEventName  string
	ToolName       string
	ToolInput      map[string]interface{}
	ToolOutput     interface{}
	Error          interface{}
	UserMessage    string
	Prompt         string
}

package claude_code

// Plugin-specific payload schemas for Claude Code events
// These define the structure of data attached to each event type

// ChatPayload contains data for chat-related events
// Used with ChatStarted, ChatMessageUser, and ChatMessageAssistant events
type ChatPayload struct {
	Message string `json:"message,omitempty"`
	Context string `json:"context,omitempty"`
}

// ToolPayload contains data for tool invocation and result events
// Used with ToolInvoked and ToolResult events
type ToolPayload struct {
	Tool       string      `json:"tool"`
	Parameters interface{} `json:"parameters,omitempty"` // Can be object, array, or string
	Result     interface{} `json:"result,omitempty"`     // Can be object, array, or string
	DurationMs int64       `json:"duration_ms,omitempty"`
	Context    string      `json:"context,omitempty"`
}

// FilePayload contains data for file access events
// Used with FileRead and FileWritten events
type FilePayload struct {
	FilePath   string `json:"file_path"`
	Changes    string `json:"changes,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Context    string `json:"context,omitempty"`
}

// ContextPayload contains data for context change events
// Used with ContextChanged events
type ContextPayload struct {
	Context     string `json:"context"`
	Description string `json:"description,omitempty"`
}

// ErrorPayload contains data for error events
// Used with Error events
type ErrorPayload struct {
	Error      string `json:"error"`
	StackTrace string `json:"stack_trace,omitempty"`
	Context    string `json:"context,omitempty"`
}

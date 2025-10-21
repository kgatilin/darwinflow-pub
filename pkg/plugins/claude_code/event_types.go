package claude_code

// Event type constants - claude-code specific event schemas
// These define the types of events that the Claude Code plugin emits
const (
	// ChatStarted is emitted when a Claude Code session starts
	ChatStarted = "claude.chat.started"

	// ChatMessageUser is emitted when a user submits a message in Claude Code
	ChatMessageUser = "claude.chat.message.user"

	// ChatMessageAssistant is emitted when Claude responds
	ChatMessageAssistant = "claude.chat.message.assistant"

	// ToolInvoked is emitted when a tool (Read, Write, Bash, etc.) is invoked
	ToolInvoked = "claude.tool.invoked"

	// ToolResult is emitted when a tool completes and returns a result
	ToolResult = "claude.tool.result"

	// FileRead is emitted when a file is read
	FileRead = "claude.file.read"

	// FileWritten is emitted when a file is written
	FileWritten = "claude.file.written"

	// ContextChanged is emitted when the context changes
	ContextChanged = "claude.context.changed"

	// Error is emitted when an error occurs
	Error = "claude.error"
)

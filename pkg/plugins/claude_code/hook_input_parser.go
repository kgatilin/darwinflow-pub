package claude_code

import (
	"encoding/json"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// HookInput represents the standardized input passed to hooks via stdin from Claude Code.
// This is a plugin-local struct that captures all Claude Code-specific hook fields.
type HookInput struct {
	SessionID      string                 `json:"session_id"`
	TranscriptPath string                 `json:"transcript_path"`
	CWD            string                 `json:"cwd"`
	PermissionMode string                 `json:"permission_mode,omitempty"`
	HookEventName  string                 `json:"hook_event_name"`
	ToolName       string                 `json:"tool_name,omitempty"`        // For PreToolUse/PostToolUse hooks
	ToolInput      map[string]interface{} `json:"tool_input,omitempty"`       // For PreToolUse/PostToolUse hooks
	ToolOutput     interface{}            `json:"tool_output,omitempty"`      // For PostToolUse hooks
	Error          interface{}            `json:"error,omitempty"`            // For error-related hooks
	UserMessage    string                 `json:"user_message,omitempty"`     // For UserPromptSubmit hook
	Prompt         string                 `json:"prompt,omitempty"`           // Alternative field for user message
}

// hookInputParser implements HookInputParser interface for the plugin.
// This parser is self-contained and has no framework dependencies.
type hookInputParser struct{}

// newHookInputParser creates a new hook input parser for the plugin
func newHookInputParser() HookInputParser {
	return &hookInputParser{}
}

// NewHookInputParser creates a new hook input parser (exported for testing)
func NewHookInputParser() HookInputParser {
	return newHookInputParser()
}

// Parse parses hook input from stdin data and returns plugin's HookInputData.
// Extracts all fields from Claude Code hook input for event conversion.
func (p *hookInputParser) Parse(data []byte) (*HookInputData, error) {
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	// Extract all hook input fields for event conversion
	return &HookInputData{
		SessionID:      input.SessionID,
		TranscriptPath: input.TranscriptPath,
		CWD:            input.CWD,
		PermissionMode: input.PermissionMode,
		HookEventName:  input.HookEventName,
		ToolName:       input.ToolName,
		ToolInput:      input.ToolInput,
		ToolOutput:     input.ToolOutput,
		Error:          input.Error,
		UserMessage:    input.UserMessage,
		Prompt:         input.Prompt,
	}, nil
}

// hookEventNameToEventType converts Claude Code hook event names to SDK event types
func hookEventNameToEventType(hookEventName string) string {
	switch hookEventName {
	case "PreToolUse":
		return "tool.invoked"
	case "PostToolUse":
		return "tool.completed"
	case "UserPromptSubmit":
		return "chat.message.user"
	case "SessionStart":
		return "session.started"
	case "SessionEnd":
		return "session.ended"
	case "Notification":
		return "notification"
	default:
		return "unknown"
	}
}

// HookInputToEvent converts a parsed HookInputData to a pluginsdk.Event
// This handles the format conversion from Claude Code hooks to plugin SDK events
func HookInputToEvent(hookData *HookInputData) *pluginsdk.Event {
	if hookData == nil {
		return nil
	}

	event := &pluginsdk.Event{
		Type:      hookEventNameToEventType(hookData.HookEventName),
		Source:    "claude-code",
		Timestamp: time.Now(),
		Version:   "1.0",
		Metadata: map[string]string{
			"session_id": hookData.SessionID,
		},
		Payload: make(map[string]interface{}),
	}

	// Add optional metadata
	if hookData.CWD != "" {
		event.Metadata["cwd"] = hookData.CWD
	}
	if hookData.PermissionMode != "" {
		event.Metadata["permission_mode"] = hookData.PermissionMode
	}
	if hookData.TranscriptPath != "" {
		event.Metadata["transcript_path"] = hookData.TranscriptPath
	}

	// Add hook event name for reference
	event.Metadata["hook_event_name"] = hookData.HookEventName

	// Add payload based on hook type
	if hookData.ToolName != "" {
		event.Payload["tool"] = hookData.ToolName
	}
	if hookData.ToolInput != nil {
		event.Payload["tool_input"] = hookData.ToolInput
	}
	if hookData.ToolOutput != nil {
		event.Payload["tool_output"] = hookData.ToolOutput
	}
	if hookData.Error != nil {
		event.Payload["error"] = hookData.Error
	}
	// Check both Prompt and UserMessage fields (Prompt takes precedence as it's the current field name)
	if hookData.Prompt != "" {
		event.Payload["message"] = hookData.Prompt
	} else if hookData.UserMessage != "" {
		event.Payload["message"] = hookData.UserMessage
	}

	return event
}

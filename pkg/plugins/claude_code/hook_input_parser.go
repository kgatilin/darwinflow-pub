package claude_code

import (
	"encoding/json"
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

// Parse parses hook input from stdin data and returns plugin's HookInputData.
// Only extracts the SessionID field, which is all the plugin currently needs.
func (p *hookInputParser) Parse(data []byte) (*HookInputData, error) {
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	// Extract only what the plugin needs (currently just SessionID)
	return &HookInputData{
		SessionID: input.SessionID,
	}, nil
}

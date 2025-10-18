package infra

import (
	"encoding/json"
	"io"
)

// HookInput represents the standardized input passed to hooks via stdin from Claude Code
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

// ParseHookInput reads and parses hook input from a reader (typically stdin)
func ParseHookInput(reader io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	return &input, nil
}

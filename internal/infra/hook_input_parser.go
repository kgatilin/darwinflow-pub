package infra

import (
	"bytes"
	"io"

	"github.com/kgatilin/darwinflow-pub/internal/app"
)

// HookInputParserAdapter adapts infra.ParseHookInput to app.HookInputParser interface
type HookInputParserAdapter struct{}

// NewHookInputParserAdapter creates a new hook input parser adapter
func NewHookInputParserAdapter() *HookInputParserAdapter {
	return &HookInputParserAdapter{}
}

// Parse parses hook input from stdin data
func (p *HookInputParserAdapter) Parse(data []byte) (*app.HookInputData, error) {
	// Parse using infra function
	hookInput, err := ParseHookInput(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return nil, err
	}

	// Convert to app.HookInputData
	return &app.HookInputData{
		SessionID:      hookInput.SessionID,
		TranscriptPath: hookInput.TranscriptPath,
		CWD:            hookInput.CWD,
		PermissionMode: hookInput.PermissionMode,
		HookEventName:  hookInput.HookEventName,
		ToolName:       hookInput.ToolName,
		ToolInput:      hookInput.ToolInput,
		ToolOutput:     hookInput.ToolOutput,
		Error:          hookInput.Error,
		UserMessage:    hookInput.UserMessage,
		Prompt:         hookInput.Prompt,
	}, nil
}

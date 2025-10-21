package claude_code_test

import (
	"encoding/json"
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

func TestHookInputParser_UserPromptSubmit_WithPromptField(t *testing.T) {
	// Simulate what Claude Code actually sends for UserPromptSubmit
	hookInput := map[string]interface{}{
		"session_id":       "test-session-123",
		"transcript_path":  "/path/to/transcript.jsonl",
		"cwd":              "/workspace",
		"permission_mode":  "bypassPermissions",
		"hook_event_name":  "UserPromptSubmit",
		"prompt":           "This is the user's actual message",
	}

	data, err := json.Marshal(hookInput)
	if err != nil {
		t.Fatalf("Failed to marshal hook input: %v", err)
	}

	parser := claude_code.NewHookInputParser()
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Prompt != "This is the user's actual message" {
		t.Errorf("Expected Prompt='This is the user's actual message', got %q", result.Prompt)
	}

	if result.SessionID != "test-session-123" {
		t.Errorf("Expected SessionID='test-session-123', got %q", result.SessionID)
	}
}

func TestHookInputParser_UserPromptSubmit_WithUserMessageField(t *testing.T) {
	// Test legacy user_message field for backward compatibility
	hookInput := map[string]interface{}{
		"session_id":       "test-session-456",
		"transcript_path":  "/path/to/transcript.jsonl",
		"cwd":              "/workspace",
		"permission_mode":  "bypassPermissions",
		"hook_event_name":  "UserPromptSubmit",
		"user_message":     "Legacy user message format",
	}

	data, err := json.Marshal(hookInput)
	if err != nil {
		t.Fatalf("Failed to marshal hook input: %v", err)
	}

	parser := claude_code.NewHookInputParser()
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.UserMessage != "Legacy user message format" {
		t.Errorf("Expected UserMessage='Legacy user message format', got %q", result.UserMessage)
	}
}

func TestHookInputToEvent_PreferPromptOverUserMessage(t *testing.T) {
	// Test that Prompt field takes precedence
	hookData := &claude_code.HookInputData{
		SessionID:     "test-session-789",
		HookEventName: "UserPromptSubmit",
		Prompt:        "This should be used",
		UserMessage:   "This should be ignored",
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	message, ok := event.Payload["message"]
	if !ok {
		t.Fatal("Event payload should contain 'message' field")
	}

	if message != "This should be used" {
		t.Errorf("Expected message='This should be used', got %q", message)
	}
}

func TestHookInputToEvent_FallbackToUserMessage(t *testing.T) {
	// Test that UserMessage is used when Prompt is empty
	hookData := &claude_code.HookInputData{
		SessionID:     "test-session-999",
		HookEventName: "UserPromptSubmit",
		Prompt:        "",
		UserMessage:   "Fallback message",
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	message, ok := event.Payload["message"]
	if !ok {
		t.Fatal("Event payload should contain 'message' field")
	}

	if message != "Fallback message" {
		t.Errorf("Expected message='Fallback message', got %q", message)
	}
}

func TestHookInputToEvent_UserPromptSubmit_EventType(t *testing.T) {
	hookData := &claude_code.HookInputData{
		SessionID:     "test-session-eventtype",
		HookEventName: "UserPromptSubmit",
		Prompt:        "Test message",
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	if event.Type != "chat.message.user" {
		t.Errorf("Expected event type='chat.message.user', got %q", event.Type)
	}

	if event.Source != "claude-code" {
		t.Errorf("Expected source='claude-code', got %q", event.Source)
	}
}

func TestHookInputToEvent_NilInput(t *testing.T) {
	event := claude_code.HookInputToEvent(nil)
	if event != nil {
		t.Error("Expected nil event for nil input")
	}
}

func TestHookInputToEvent_AllHookEventTypes(t *testing.T) {
	tests := []struct {
		name          string
		hookEventName string
		expectedType  string
	}{
		{"PreToolUse", "PreToolUse", "tool.invoked"},
		{"PostToolUse", "PostToolUse", "tool.completed"},
		{"UserPromptSubmit", "UserPromptSubmit", "chat.message.user"},
		{"SessionStart", "SessionStart", "session.started"},
		{"SessionEnd", "SessionEnd", "session.ended"},
		{"Notification", "Notification", "notification"},
		{"Unknown", "UnknownEvent", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookData := &claude_code.HookInputData{
				SessionID:     "test-session",
				HookEventName: tt.hookEventName,
			}

			event := claude_code.HookInputToEvent(hookData)
			if event == nil {
				t.Fatal("Event should not be nil")
			}

			if event.Type != tt.expectedType {
				t.Errorf("Expected event type=%q, got %q", tt.expectedType, event.Type)
			}

			if event.Source != "claude-code" {
				t.Errorf("Expected source='claude-code', got %q", event.Source)
			}

			// Verify session_id is in metadata
			if event.Metadata["session_id"] != "test-session" {
				t.Errorf("Expected session_id in metadata, got %q", event.Metadata["session_id"])
			}
		})
	}
}

func TestHookInputToEvent_WithToolData(t *testing.T) {
	toolInput := map[string]interface{}{
		"file_path": "/test/file.txt",
		"content":   "test content",
	}

	hookData := &claude_code.HookInputData{
		SessionID:     "test-session",
		HookEventName: "PreToolUse",
		ToolName:      "Write",
		ToolInput:     toolInput,
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	if event.Payload["tool"] != "Write" {
		t.Errorf("Expected tool='Write', got %v", event.Payload["tool"])
	}

	if event.Payload["tool_input"] == nil {
		t.Fatal("Expected tool_input in payload")
	}
}

func TestHookInputToEvent_WithToolOutput(t *testing.T) {
	hookData := &claude_code.HookInputData{
		SessionID:     "test-session",
		HookEventName: "PostToolUse",
		ToolName:      "Read",
		ToolOutput:    "File contents here",
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	if event.Payload["tool_output"] != "File contents here" {
		t.Errorf("Expected tool_output='File contents here', got %v", event.Payload["tool_output"])
	}
}

func TestHookInputToEvent_WithError(t *testing.T) {
	hookData := &claude_code.HookInputData{
		SessionID:     "test-session",
		HookEventName: "PostToolUse",
		ToolName:      "Bash",
		Error:         "Command failed with exit code 1",
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	if event.Payload["error"] != "Command failed with exit code 1" {
		t.Errorf("Expected error in payload, got %v", event.Payload["error"])
	}
}

func TestHookInputToEvent_WithAllMetadata(t *testing.T) {
	hookData := &claude_code.HookInputData{
		SessionID:      "test-session",
		HookEventName:  "UserPromptSubmit",
		CWD:            "/workspace",
		PermissionMode: "bypassPermissions",
		TranscriptPath: "/path/to/transcript.jsonl",
		Prompt:         "Test prompt",
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	// Verify all metadata fields
	if event.Metadata["cwd"] != "/workspace" {
		t.Errorf("Expected cwd in metadata, got %q", event.Metadata["cwd"])
	}

	if event.Metadata["permission_mode"] != "bypassPermissions" {
		t.Errorf("Expected permission_mode in metadata, got %q", event.Metadata["permission_mode"])
	}

	if event.Metadata["transcript_path"] != "/path/to/transcript.jsonl" {
		t.Errorf("Expected transcript_path in metadata, got %q", event.Metadata["transcript_path"])
	}

	if event.Metadata["hook_event_name"] != "UserPromptSubmit" {
		t.Errorf("Expected hook_event_name in metadata, got %q", event.Metadata["hook_event_name"])
	}
}

func TestHookInputToEvent_EmptyOptionalFields(t *testing.T) {
	// Test that empty optional fields are not added to metadata
	hookData := &claude_code.HookInputData{
		SessionID:     "test-session",
		HookEventName: "SessionStart",
		// All optional fields are empty
	}

	event := claude_code.HookInputToEvent(hookData)
	if event == nil {
		t.Fatal("Event should not be nil")
	}

	// Verify that empty optional metadata fields are not added
	if _, exists := event.Metadata["cwd"]; exists && event.Metadata["cwd"] == "" {
		t.Error("Empty cwd should not be in metadata")
	}

	if _, exists := event.Metadata["permission_mode"]; exists && event.Metadata["permission_mode"] == "" {
		t.Error("Empty permission_mode should not be in metadata")
	}

	if _, exists := event.Metadata["transcript_path"]; exists && event.Metadata["transcript_path"] == "" {
		t.Error("Empty transcript_path should not be in metadata")
	}
}

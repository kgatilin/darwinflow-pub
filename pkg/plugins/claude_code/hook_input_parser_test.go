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

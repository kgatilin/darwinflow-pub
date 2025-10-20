package infra_test

import (
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestParseHookInput_Valid(t *testing.T) {
	jsonInput := `{
		"session_id": "test-session-123",
		"transcript_path": "/path/to/transcript.json",
		"cwd": "/working/directory",
		"hook_event_name": "PreToolUse",
		"tool_name": "Read",
		"tool_input": {"file": "test.go"}
	}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.SessionID != "test-session-123" {
		t.Errorf("Expected SessionID = %q, got %q", "test-session-123", input.SessionID)
	}

	if input.TranscriptPath != "/path/to/transcript.json" {
		t.Errorf("Expected TranscriptPath = %q, got %q", "/path/to/transcript.json", input.TranscriptPath)
	}

	if input.CWD != "/working/directory" {
		t.Errorf("Expected CWD = %q, got %q", "/working/directory", input.CWD)
	}

	if input.HookEventName != "PreToolUse" {
		t.Errorf("Expected HookEventName = %q, got %q", "PreToolUse", input.HookEventName)
	}

	if input.ToolName != "Read" {
		t.Errorf("Expected ToolName = %q, got %q", "Read", input.ToolName)
	}

	if input.ToolInput == nil {
		t.Fatal("Expected non-nil ToolInput")
	}

	if file, ok := input.ToolInput["file"].(string); !ok || file != "test.go" {
		t.Errorf("Expected ToolInput[file] = %q, got %v", "test.go", input.ToolInput["file"])
	}
}

func TestParseHookInput_UserPromptSubmit(t *testing.T) {
	jsonInput := `{
		"session_id": "session-456",
		"transcript_path": "/transcript.json",
		"cwd": "/dir",
		"hook_event_name": "UserPromptSubmit",
		"user_message": "Hello, how are you?"
	}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.UserMessage != "Hello, how are you?" {
		t.Errorf("Expected UserMessage = %q, got %q", "Hello, how are you?", input.UserMessage)
	}
}

func TestParseHookInput_WithPromptField(t *testing.T) {
	jsonInput := `{
		"session_id": "session-789",
		"transcript_path": "/transcript.json",
		"cwd": "/dir",
		"hook_event_name": "UserPromptSubmit",
		"prompt": "Alternative prompt field"
	}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.Prompt != "Alternative prompt field" {
		t.Errorf("Expected Prompt = %q, got %q", "Alternative prompt field", input.Prompt)
	}
}

func TestParseHookInput_WithToolOutput(t *testing.T) {
	jsonInput := `{
		"session_id": "session-output",
		"transcript_path": "/transcript.json",
		"cwd": "/dir",
		"hook_event_name": "PostToolUse",
		"tool_name": "Bash",
		"tool_output": {"stdout": "command output", "exit_code": 0}
	}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.ToolOutput == nil {
		t.Fatal("Expected non-nil ToolOutput")
	}
}

func TestParseHookInput_WithError(t *testing.T) {
	jsonInput := `{
		"session_id": "session-error",
		"transcript_path": "/transcript.json",
		"cwd": "/dir",
		"hook_event_name": "Error",
		"error": {"message": "Something went wrong", "code": 500}
	}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.Error == nil {
		t.Fatal("Expected non-nil Error")
	}
}

func TestParseHookInput_InvalidJSON(t *testing.T) {
	jsonInput := `{invalid json`

	reader := strings.NewReader(jsonInput)
	_, err := infra.ParseHookInput(reader)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseHookInput_EmptyJSON(t *testing.T) {
	jsonInput := `{}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	// Should have zero values
	if input.SessionID != "" {
		t.Error("Expected empty SessionID")
	}
}

func TestParseHookInput_WithPermissionMode(t *testing.T) {
	jsonInput := `{
		"session_id": "session-perm",
		"transcript_path": "/transcript.json",
		"cwd": "/dir",
		"hook_event_name": "PreToolUse",
		"permission_mode": "auto"
	}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.PermissionMode != "auto" {
		t.Errorf("Expected PermissionMode = %q, got %q", "auto", input.PermissionMode)
	}
}

func TestParseHookInput_ComplexToolInput(t *testing.T) {
	jsonInput := `{
		"session_id": "session-complex",
		"transcript_path": "/transcript.json",
		"cwd": "/dir",
		"hook_event_name": "PreToolUse",
		"tool_name": "Edit",
		"tool_input": {
			"file_path": "/path/to/file.go",
			"old_string": "old text",
			"new_string": "new text",
			"replace_all": false
		}
	}`

	reader := strings.NewReader(jsonInput)
	input, err := infra.ParseHookInput(reader)

	if err != nil {
		t.Fatalf("ParseHookInput failed: %v", err)
	}

	if input.ToolName != "Edit" {
		t.Error("ToolName mismatch")
	}

	if input.ToolInput == nil {
		t.Fatal("Expected non-nil ToolInput")
	}

	if filePath, ok := input.ToolInput["file_path"].(string); !ok || filePath != "/path/to/file.go" {
		t.Error("file_path mismatch in ToolInput")
	}

	if replaceAll, ok := input.ToolInput["replace_all"].(bool); !ok || replaceAll != false {
		t.Error("replace_all mismatch in ToolInput")
	}
}

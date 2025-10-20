package infra_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestNewContextDetector(t *testing.T) {
	detector := infra.NewContextDetector()
	if detector == nil {
		t.Fatal("NewContextDetector returned nil")
	}
}

func TestContextDetector_DetectContext_FromEnv(t *testing.T) {
	// Save and restore original env var
	oldCtx := os.Getenv("DW_CONTEXT")
	defer os.Setenv("DW_CONTEXT", oldCtx)

	os.Setenv("DW_CONTEXT", "test/myproject")

	detector := infra.NewContextDetector()
	ctx := detector.DetectContext()

	if ctx != "test/myproject" {
		t.Errorf("Expected context 'test/myproject', got %q", ctx)
	}
}

func TestContextDetector_DetectContext_Default(t *testing.T) {
	// Save and restore original env var
	oldCtx := os.Getenv("DW_CONTEXT")
	defer os.Setenv("DW_CONTEXT", oldCtx)

	// Clear the env var
	os.Unsetenv("DW_CONTEXT")

	detector := infra.NewContextDetector()
	ctx := detector.DetectContext()

	// Should return something (either from path or "unknown")
	if ctx == "" {
		t.Error("Expected non-empty context")
	}
}

func TestNormalizeContent_ToolInvoked(t *testing.T) {
	payload := `{"tool":"Read","parameters":{"file":"test.go"}}`
	content := infra.NormalizeContent("tool.invoked", payload)

	if !strings.Contains(content, "Tool: Read") {
		t.Errorf("Expected 'Tool: Read' in content, got: %q", content)
	}
	if !strings.Contains(content, "Parameters:") {
		t.Errorf("Expected 'Parameters:' in content, got: %q", content)
	}
	if !strings.Contains(content, "test.go") {
		t.Errorf("Expected 'test.go' in content, got: %q", content)
	}
}

func TestNormalizeContent_ToolInvoked_NoParams(t *testing.T) {
	payload := `{"tool":"Write"}`
	content := infra.NormalizeContent("tool.invoked", payload)

	if !strings.Contains(content, "Tool: Write") {
		t.Errorf("Expected 'Tool: Write' in content, got: %q", content)
	}
	// Should not have Parameters section
	lines := strings.Split(content, "\n")
	if len(lines) > 1 {
		t.Errorf("Expected single line output for tool without params, got: %q", content)
	}
}

func TestNormalizeContent_ToolInvoked_LongParams(t *testing.T) {
	// Create a payload with very long parameters
	longParam := strings.Repeat("a", 600)
	payload := `{"tool":"Bash","parameters":"` + longParam + `"}`
	content := infra.NormalizeContent("tool.invoked", payload)

	if !strings.Contains(content, "Tool: Bash") {
		t.Error("Expected 'Tool: Bash' in content")
	}
	// Should be truncated with "..."
	if !strings.Contains(content, "...") {
		t.Errorf("Expected truncation marker '...' in long content, got: %q", content)
	}
	// Content should be limited
	if len(content) > 600 {
		t.Errorf("Expected content to be truncated, got length %d", len(content))
	}
}

func TestNormalizeContent_ToolResult(t *testing.T) {
	payload := `{"tool":"Read","result":"file contents here"}`
	content := infra.NormalizeContent("tool.result", payload)

	if !strings.Contains(content, "Tool: Read") {
		t.Error("Expected 'Tool: Read' in content")
	}
	if !strings.Contains(content, "Result:") {
		t.Error("Expected 'Result:' in content")
	}
	if !strings.Contains(content, "file contents here") {
		t.Error("Expected result text in content")
	}
}

func TestNormalizeContent_ToolResult_NoResult(t *testing.T) {
	payload := `{"tool":"Bash"}`
	content := infra.NormalizeContent("tool.result", payload)

	if !strings.Contains(content, "Tool: Bash") {
		t.Error("Expected 'Tool: Bash' in content")
	}
	if !strings.Contains(content, "completed") {
		t.Error("Expected 'completed' in content for tool without result")
	}
}

func TestNormalizeContent_ChatMessages(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		payload   string
		expected  string
	}{
		{
			name:      "user message",
			eventType: "chat.message.user",
			payload:   `{"message":"Hello, how are you?"}`,
			expected:  "Hello, how are you?",
		},
		{
			name:      "assistant message",
			eventType: "chat.message.assistant",
			payload:   `{"message":"I'm doing well, thanks!"}`,
			expected:  "I'm doing well, thanks!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := infra.NormalizeContent(tt.eventType, tt.payload)
			if content != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, content)
			}
		})
	}
}

func TestNormalizeContent_FileEvents(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		payload   string
		wantFile  string
		wantChanges bool
	}{
		{
			name:      "file read",
			eventType: "file.read",
			payload:   `{"file_path":"/path/to/file.go"}`,
			wantFile:  "/path/to/file.go",
			wantChanges: false,
		},
		{
			name:      "file written with changes",
			eventType: "file.written",
			payload:   `{"file_path":"/path/to/file.go","changes":"added function Foo"}`,
			wantFile:  "/path/to/file.go",
			wantChanges: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := infra.NormalizeContent(tt.eventType, tt.payload)

			if !strings.Contains(content, "File:") || !strings.Contains(content, tt.wantFile) {
				t.Errorf("Expected file path %q in content, got: %q", tt.wantFile, content)
			}

			hasChanges := strings.Contains(content, "Changes:")
			if hasChanges != tt.wantChanges {
				t.Errorf("Expected Changes section=%v, got=%v. Content: %q", tt.wantChanges, hasChanges, content)
			}
		})
	}
}

func TestNormalizeContent_UnknownEventType(t *testing.T) {
	payload := `{"key1":"value1","key2":"value2","key3":"value3"}`
	content := infra.NormalizeContent("custom.event", payload)

	// Should contain key-value pairs
	if !strings.Contains(content, "key1: value1") {
		t.Error("Expected 'key1: value1' in content")
	}
	if !strings.Contains(content, "key2: value2") {
		t.Error("Expected 'key2: value2' in content")
	}
}

func TestNormalizeContent_UnknownEventType_SkipsContext(t *testing.T) {
	payload := `{"field":"value","context":"should be skipped"}`
	content := infra.NormalizeContent("custom.event", payload)

	// Should contain the field
	if !strings.Contains(content, "field: value") {
		t.Error("Expected 'field: value' in content")
	}
	// Should NOT contain context field
	if strings.Contains(content, "context:") || strings.Contains(content, "should be skipped") {
		t.Errorf("Context field should be skipped, got: %q", content)
	}
}

func TestNormalizeContent_InvalidJSON(t *testing.T) {
	payload := `invalid json {`
	content := infra.NormalizeContent("some.event", payload)

	// Should fall back to simple combination
	if !strings.Contains(content, "some.event:") {
		t.Error("Expected event type in fallback content")
	}
	if !strings.Contains(content, payload) {
		t.Error("Expected payload in fallback content")
	}
}

func TestNormalizeContent_EmptyPayload(t *testing.T) {
	content := infra.NormalizeContent("test.event", "{}")

	// Empty JSON object results in no content, which is fine
	// The important thing is it doesn't panic
	_ = content
}

func TestNormalizeContent_NullValues(t *testing.T) {
	payload := `{"key1":"value1","key2":null,"key3":""}`
	content := infra.NormalizeContent("test.event", payload)

	// Should contain non-null values
	if !strings.Contains(content, "key1: value1") {
		t.Error("Expected 'key1: value1' in content")
	}

	// Should handle null and empty gracefully (not output them)
	// The function should skip <nil> values
}

func TestNormalizeContent_ComplexParameters(t *testing.T) {
	payload := `{"tool":"Edit","parameters":{"file":"/path/to/file.go","old":"text","new":"updated"}}`
	content := infra.NormalizeContent("tool.invoked", payload)

	if !strings.Contains(content, "Tool: Edit") {
		t.Error("Expected 'Tool: Edit' in content")
	}
	if !strings.Contains(content, "Parameters:") {
		t.Error("Expected 'Parameters:' in content")
	}
	// Should have JSON representation of complex params
	if !strings.Contains(content, "file") {
		t.Error("Expected 'file' key in parameters")
	}
}

func TestContextDetector_ParseContextFromPath_WithDarwinFlow(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	darwinFlowDir := filepath.Join(projectDir, ".darwinflow")

	err := os.MkdirAll(darwinFlowDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Save current dir and restore later
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to a subdirectory within the project
	subDir := filepath.Join(projectDir, "src", "pkg")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	err = os.Chdir(subDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Clear env var to test path detection
	oldCtx := os.Getenv("DW_CONTEXT")
	defer os.Setenv("DW_CONTEXT", oldCtx)
	os.Unsetenv("DW_CONTEXT")

	detector := infra.NewContextDetector()
	ctx := detector.DetectContext()

	// Should detect "project/myproject" from .darwinflow directory
	if !strings.Contains(ctx, "myproject") {
		t.Errorf("Expected context to contain 'myproject', got: %q", ctx)
	}
	if !strings.HasPrefix(ctx, "project/") {
		t.Errorf("Expected context to start with 'project/', got: %q", ctx)
	}
}

func TestContextDetector_ParseContextFromPath_Fallback(t *testing.T) {
	// Create a temporary directory without .darwinflow
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")

	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Save current dir and restore later
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err = os.Chdir(projectDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Clear env var
	oldCtx := os.Getenv("DW_CONTEXT")
	defer os.Setenv("DW_CONTEXT", oldCtx)
	os.Unsetenv("DW_CONTEXT")

	detector := infra.NewContextDetector()
	ctx := detector.DetectContext()

	// Should use fallback: project/testproject
	if !strings.Contains(ctx, "testproject") {
		t.Errorf("Expected context to contain 'testproject', got: %q", ctx)
	}
}

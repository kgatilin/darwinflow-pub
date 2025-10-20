package claude_code_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// mockCommandContext implements pluginsdk.CommandContext for testing
type mockCommandContext struct {
	stdin   io.Reader
	stdout  io.Writer
	emitErr error
	events  []pluginsdk.Event
}

func (m *mockCommandContext) GetLogger() pluginsdk.Logger {
	return &mockLogger{}
}

func (m *mockCommandContext) GetWorkingDir() string {
	return "/workspace"
}

func (m *mockCommandContext) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	m.events = append(m.events, event)
	return m.emitErr
}

func (m *mockCommandContext) GetStdout() io.Writer {
	return m.stdout
}

func (m *mockCommandContext) GetStdin() io.Reader {
	return m.stdin
}

// newMockCommandContext creates a new mock context with JSON input
func newMockCommandContext(jsonInput string) *mockCommandContext {
	return &mockCommandContext{
		stdin:  strings.NewReader(jsonInput),
		stdout: &bytes.Buffer{},
		events: []pluginsdk.Event{},
	}
}

// TestNewEmitEventCommand verifies the command can be created
func TestNewEmitEventCommand(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	if cmd == nil {
		t.Fatal("NewEmitEventCommand returned nil")
	}

	if cmd.GetName() != "emit-event" {
		t.Errorf("GetName() = %q, want %q", cmd.GetName(), "emit-event")
	}

	if cmd.GetDescription() == "" {
		t.Error("GetDescription() returned empty string")
	}

	if cmd.GetUsage() == "" {
		t.Error("GetUsage() returned empty string")
	}
}

// TestEmitEventCommand_ValidEvent verifies a valid event is emitted
func TestEmitEventCommand_ValidEvent(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:   "tool.invoked",
		Source: "claude-code",
		Timestamp: time.Date(2025, 10, 20, 10, 30, 0, 0, time.UTC),
		Payload: map[string]interface{}{
			"tool": "Read",
		},
		Metadata: map[string]string{
			"session_id": "abc123",
			"cwd":        "/workspace",
		},
		Version: "1.0",
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	// Execute the command
	err := cmd.Execute(context.Background(), mockCtx, nil)

	// Should not return error (silently fails internally)
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// Event should be emitted
	if len(mockCtx.events) != 1 {
		t.Errorf("Expected 1 event emitted, got %d", len(mockCtx.events))
	} else {
		emitted := mockCtx.events[0]
		if emitted.Type != "tool.invoked" {
			t.Errorf("Event type = %q, want %q", emitted.Type, "tool.invoked")
		}
		if emitted.Source != "claude-code" {
			t.Errorf("Event source = %q, want %q", emitted.Source, "claude-code")
		}
		if emitted.Metadata["session_id"] != "abc123" {
			t.Errorf("Session ID = %q, want %q", emitted.Metadata["session_id"], "abc123")
		}
	}
}

// TestEmitEventCommand_InvalidJSON verifies invalid JSON is silently ignored
func TestEmitEventCommand_InvalidJSON(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	mockCtx := newMockCommandContext("{invalid json")

	err := cmd.Execute(context.Background(), mockCtx, nil)

	// Should not return error
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// No events should be emitted
	if len(mockCtx.events) != 0 {
		t.Errorf("Expected 0 events emitted, got %d", len(mockCtx.events))
	}
}

// TestEmitEventCommand_MissingSessionID verifies missing session_id is silently ignored
func TestEmitEventCommand_MissingSessionID(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:   "tool.invoked",
		Source: "claude-code",
		Payload: map[string]interface{}{
			"tool": "Read",
		},
		Metadata: map[string]string{
			"cwd": "/workspace",
			// session_id is missing
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// No events should be emitted
	if len(mockCtx.events) != 0 {
		t.Errorf("Expected 0 events emitted, got %d", len(mockCtx.events))
	}
}

// TestEmitEventCommand_MissingType verifies missing type is silently ignored
func TestEmitEventCommand_MissingType(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		// Type is missing
		Source: "claude-code",
		Payload: map[string]interface{}{
			"tool": "Read",
		},
		Metadata: map[string]string{
			"session_id": "abc123",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// No events should be emitted
	if len(mockCtx.events) != 0 {
		t.Errorf("Expected 0 events emitted, got %d", len(mockCtx.events))
	}
}

// TestEmitEventCommand_MissingSource verifies missing source is silently ignored
func TestEmitEventCommand_MissingSource(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type: "tool.invoked",
		// Source is missing
		Payload: map[string]interface{}{
			"tool": "Read",
		},
		Metadata: map[string]string{
			"session_id": "abc123",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// No events should be emitted
	if len(mockCtx.events) != 0 {
		t.Errorf("Expected 0 events emitted, got %d", len(mockCtx.events))
	}
}

// TestEmitEventCommand_DefaultTimestamp verifies missing timestamp is set to current time
func TestEmitEventCommand_DefaultTimestamp(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:   "tool.invoked",
		Source: "claude-code",
		// Timestamp is missing (zero value)
		Payload: map[string]interface{}{
			"tool": "Read",
		},
		Metadata: map[string]string{
			"session_id": "abc123",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	beforeTime := time.Now()
	err := cmd.Execute(context.Background(), mockCtx, nil)
	afterTime := time.Now()

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}

	emitted := mockCtx.events[0]

	// Verify timestamp was set to approximately now
	if emitted.Timestamp.Before(beforeTime) || emitted.Timestamp.After(afterTime.Add(1*time.Second)) {
		t.Errorf("Timestamp = %v, should be between %v and %v", emitted.Timestamp, beforeTime, afterTime)
	}
}

// TestEmitEventCommand_DefaultVersion verifies missing version is set to "1.0"
func TestEmitEventCommand_DefaultVersion(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:   "tool.invoked",
		Source: "claude-code",
		Payload: map[string]interface{}{
			"tool": "Read",
		},
		Metadata: map[string]string{
			"session_id": "abc123",
		},
		// Version is missing
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}

	emitted := mockCtx.events[0]
	if emitted.Version != "1.0" {
		t.Errorf("Version = %q, want %q", emitted.Version, "1.0")
	}
}

// TestEmitEventCommand_ExplicitVersion verifies explicit version is preserved
func TestEmitEventCommand_ExplicitVersion(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:    "tool.invoked",
		Source:  "claude-code",
		Payload: map[string]interface{}{"tool": "Read"},
		Metadata: map[string]string{
			"session_id": "abc123",
		},
		Version: "2.0",
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}

	emitted := mockCtx.events[0]
	if emitted.Version != "2.0" {
		t.Errorf("Version = %q, want %q", emitted.Version, "2.0")
	}
}

// TestEmitEventCommand_EmptyStdin verifies empty stdin is silently ignored
func TestEmitEventCommand_EmptyStdin(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	mockCtx := newMockCommandContext("")

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// No events should be emitted
	if len(mockCtx.events) != 0 {
		t.Errorf("Expected 0 events emitted, got %d", len(mockCtx.events))
	}
}

// TestEmitEventCommand_NilMetadata verifies nil metadata is initialized
func TestEmitEventCommand_NilMetadata(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:     "tool.invoked",
		Source:   "claude-code",
		Payload:  map[string]interface{}{"tool": "Read"},
		Metadata: nil,
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// No events should be emitted (no session_id in nil metadata)
	if len(mockCtx.events) != 0 {
		t.Errorf("Expected 0 events emitted (nil metadata means no session_id), got %d", len(mockCtx.events))
	}
}

// TestEmitEventCommand_NilPayload verifies nil payload is initialized
func TestEmitEventCommand_NilPayload(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:    "tool.invoked",
		Source:  "claude-code",
		Payload: nil,
		Metadata: map[string]string{
			"session_id": "abc123",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}

	emitted := mockCtx.events[0]
	if emitted.Payload == nil {
		t.Error("Payload should not be nil (should be initialized to empty map)")
	}
}

// TestEmitEventCommand_EmitError verifies emit errors are silently handled
func TestEmitEventCommand_EmitError(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:    "tool.invoked",
		Source:  "claude-code",
		Payload: map[string]interface{}{"tool": "Read"},
		Metadata: map[string]string{
			"session_id": "abc123",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))
	mockCtx.emitErr = io.EOF // Simulate emit failure

	err := cmd.Execute(context.Background(), mockCtx, nil)

	// Should not return error (silently fails)
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}
}

// TestEmitEventCommand_StdinReadError verifies stdin read errors are silently handled
func TestEmitEventCommand_StdinReadError(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	// Create a reader that returns an error
	errReader := &errorReader{}
	mockCtx := &mockCommandContext{
		stdin:   errReader,
		stdout:  &bytes.Buffer{},
		events:  []pluginsdk.Event{},
	}

	err := cmd.Execute(context.Background(), mockCtx, nil)

	// Should not return error (silently fails)
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// No events should be emitted
	if len(mockCtx.events) != 0 {
		t.Errorf("Expected 0 events emitted, got %d", len(mockCtx.events))
	}
}

// errorReader always returns an error on Read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// TestEmitEventCommand_LargePayload verifies large payloads are handled
func TestEmitEventCommand_LargePayload(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	// Create a large payload
	largePayload := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largePayload[strings.Repeat("x", 100)] = strings.Repeat("y", 1000)
	}

	event := pluginsdk.Event{
		Type:     "tool.invoked",
		Source:   "claude-code",
		Payload:  largePayload,
		Metadata: map[string]string{"session_id": "abc123"},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}
}

// TestEmitEventCommand_SpecialCharactersInSessionID verifies special characters in session_id
func TestEmitEventCommand_SpecialCharactersInSessionID(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:    "tool.invoked",
		Source:  "claude-code",
		Payload: map[string]interface{}{"tool": "Read"},
		Metadata: map[string]string{
			"session_id": "abc-123_456.789",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}

	emitted := mockCtx.events[0]
	if emitted.Metadata["session_id"] != "abc-123_456.789" {
		t.Errorf("Session ID not preserved: got %q", emitted.Metadata["session_id"])
	}
}

// TestEmitEventCommand_MultipleMetadataFields verifies multiple metadata fields are preserved
func TestEmitEventCommand_MultipleMetadataFields(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:   "tool.invoked",
		Source: "claude-code",
		Payload: map[string]interface{}{
			"tool": "Read",
		},
		Metadata: map[string]string{
			"session_id": "abc123",
			"cwd":        "/workspace",
			"user_id":    "user-456",
			"env":        "test",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}

	emitted := mockCtx.events[0]
	expectedMetadata := map[string]string{
		"session_id": "abc123",
		"cwd":        "/workspace",
		"user_id":    "user-456",
		"env":        "test",
	}

	for key, expectedValue := range expectedMetadata {
		if emitted.Metadata[key] != expectedValue {
			t.Errorf("Metadata[%q] = %q, want %q", key, emitted.Metadata[key], expectedValue)
		}
	}
}

// TestEmitEventCommand_ComplexPayload verifies complex nested payloads are handled
func TestEmitEventCommand_ComplexPayload(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	event := pluginsdk.Event{
		Type:   "tool.invoked",
		Source: "claude-code",
		Payload: map[string]interface{}{
			"tool": "Read",
			"parameters": map[string]interface{}{
				"file_path": "/workspace/test.go",
				"options": map[string]interface{}{
					"follow_symlinks": true,
					"timeout":         30,
				},
			},
		},
		Metadata: map[string]string{
			"session_id": "abc123",
		},
	}

	jsonData, _ := json.Marshal(event)
	mockCtx := newMockCommandContext(string(jsonData))

	err := cmd.Execute(context.Background(), mockCtx, nil)

	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if len(mockCtx.events) != 1 {
		t.Fatalf("Expected 1 event emitted, got %d", len(mockCtx.events))
	}

	emitted := mockCtx.events[0]
	if emitted.Payload["tool"] != "Read" {
		t.Errorf("Payload tool = %q, want %q", emitted.Payload["tool"], "Read")
	}
}

// TestEmitEventCommand_CommandImplementsSDKInterface verifies the command implements the SDK interface
func TestEmitEventCommand_CommandImplementsSDKInterface(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")
	cmd := claude_code.NewEmitEventCommand(plugin)

	// Verify the command implements pluginsdk.Command interface
	var _ pluginsdk.Command = cmd
}

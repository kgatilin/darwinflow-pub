package app_test

import (
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
)

func TestEventMapper_MapEventType(t *testing.T) {
	mapper := &app.EventMapper{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Chat events
		{name: "chat.started", input: "chat.started", expected: "claude.chat.started"},
		{name: "chat.ended", input: "chat.ended", expected: "claude.chat.started"},
		{name: "chat.end", input: "chat.end", expected: "claude.chat.started"},

		// User messages
		{name: "chat.message.user", input: "chat.message.user", expected: "claude.chat.message.user"},
		{name: "user.message", input: "user.message", expected: "claude.chat.message.user"},

		// Assistant messages
		{name: "chat.message.assistant", input: "chat.message.assistant", expected: "claude.chat.message.assistant"},
		{name: "assistant.message", input: "assistant.message", expected: "claude.chat.message.assistant"},

		// Tool events
		{name: "tool.invoked", input: "tool.invoked", expected: "claude.tool.invoked"},
		{name: "tool.invoke", input: "tool.invoke", expected: "claude.tool.invoked"},
		{name: "tool.result", input: "tool.result", expected: "claude.tool.result"},

		// File events
		{name: "file.read", input: "file.read", expected: "claude.file.read"},
		{name: "file.written", input: "file.written", expected: "claude.file.written"},
		{name: "file.write", input: "file.write", expected: "claude.file.written"},

		// Context events
		{name: "context.changed", input: "context.changed", expected: "claude.context.changed"},
		{name: "context.change", input: "context.change", expected: "claude.context.changed"},

		// Error events
		{name: "error", input: "error", expected: "claude.error"},

		// Case insensitive
		{name: "uppercase CHAT.STARTED", input: "CHAT.STARTED", expected: "claude.chat.started"},
		{name: "mixed case Chat.Started", input: "Chat.Started", expected: "claude.chat.started"},

		// Underscore normalization
		{name: "underscore chat_started", input: "chat_started", expected: "claude.chat.started"},
		{name: "underscore tool_invoked", input: "tool_invoked", expected: "claude.tool.invoked"},
		{name: "underscore file_read", input: "file_read", expected: "claude.file.read"},

		// Unknown event types (returns as-is, normalized)
		{name: "unknown event", input: "custom.event", expected: "claude.custom.event"},
		{name: "unknown with underscore", input: "custom_event", expected: "claude.custom.event"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapEventType(tt.input)
			if result != tt.expected {
				t.Errorf("MapEventType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEventMapper_MapEventType_Normalization(t *testing.T) {
	mapper := &app.EventMapper{}

	// Test that normalization converts underscores to dots and lowercases
	tests := []struct {
		input    string
		expected string
	}{
		{"Chat_Started", "claude.chat.started"},
		{"TOOL_INVOKED", "claude.tool.invoked"},
		{"File_Read", "claude.file.read"},
		{"Tool_Result", "claude.tool.result"},
	}

	for _, tt := range tests {
		result := mapper.MapEventType(tt.input)
		if result != tt.expected {
			t.Errorf("MapEventType(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNewLoggerService(t *testing.T) {
	repo := &MockEventRepository{}
	transcriptParser := &MockTranscriptParser{}
	contextDetector := &MockContextDetector{context: "test-context"}
	normalizer := func(eventType, payload string) string {
		return eventType + ":" + payload
	}

	logger := app.NewLoggerService(repo, transcriptParser, contextDetector, normalizer)

	if logger == nil {
		t.Error("Expected non-nil LoggerService")
	}
}

// MockTranscriptParser for testing
type MockTranscriptParser struct {
	toolName      string
	toolParams    string
	userMessage   string
	assistantMsg  string
	extractError  error
}

func (m *MockTranscriptParser) ExtractLastToolUse(transcriptPath string, maxParamLength int) (string, string, error) {
	if m.extractError != nil {
		return "", "", m.extractError
	}
	return m.toolName, m.toolParams, nil
}

func (m *MockTranscriptParser) ExtractLastUserMessage(transcriptPath string) (string, error) {
	if m.extractError != nil {
		return "", m.extractError
	}
	return m.userMessage, nil
}

func (m *MockTranscriptParser) ExtractLastAssistantMessage(transcriptPath string) (string, error) {
	if m.extractError != nil {
		return "", m.extractError
	}
	return m.assistantMsg, nil
}

// MockContextDetector for testing
type MockContextDetector struct {
	context string
}

func (m *MockContextDetector) DetectContext() string {
	return m.context
}

func TestLoggerService_Creation(t *testing.T) {
	repo := &MockEventRepository{}
	parser := &MockTranscriptParser{}
	detector := &MockContextDetector{context: "/test/path"}
	normalizer := func(eventType, payload string) string {
		return payload
	}

	service := app.NewLoggerService(repo, parser, detector, normalizer)

	if service == nil {
		t.Fatal("LoggerService should not be nil")
	}

	// Verify service was created (we can't access private fields, but creation is enough)
}

func TestLoggerService_ContentNormalizer(t *testing.T) {
	// Test that content normalizer is used
	repo := &MockEventRepository{}
	parser := &MockTranscriptParser{}
	detector := &MockContextDetector{context: "ctx"}

	called := false

	normalizer := func(eventType, payload string) string {
		called = true
		return "normalized"
	}

	service := app.NewLoggerService(repo, parser, detector, normalizer)

	// We can't directly test this without calling LogEvent, which requires
	// more complex setup. This test verifies the service accepts the normalizer.
	if service == nil {
		t.Error("Service should be created with normalizer")
	}

	// Verify normalizer wasn't called during construction
	if called {
		t.Error("Normalizer should not be called during construction")
	}
}

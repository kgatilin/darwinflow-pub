package app_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// mockEventRepo is a mock implementation of domain.EventRepository for testing
type mockEventRepo struct {
	events []*domain.Event
	saveError error
}

func (m *mockEventRepo) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockEventRepo) Save(ctx context.Context, event *domain.Event) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventRepo) FindByQuery(ctx context.Context, query domain.EventQuery) ([]*domain.Event, error) {
	return m.events, nil
}

func (m *mockEventRepo) Close() error {
	return nil
}

// mockPluginContextLogger is a mock implementation of app.Logger for testing plugin context
type mockPluginContextLogger struct{}

func (m *mockPluginContextLogger) Debug(format string, args ...interface{}) {}
func (m *mockPluginContextLogger) Info(format string, args ...interface{})  {}
func (m *mockPluginContextLogger) Warn(format string, args ...interface{})  {}
func (m *mockPluginContextLogger) Error(format string, args ...interface{}) {}

func TestPluginContext_GetLogger(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}

	ctx := app.NewPluginContext(logger, "/test/db", "/test/dir", eventRepo)

	if got := ctx.GetLogger(); got == nil {
		t.Error("GetLogger() returned nil")
	}
}

func TestPluginContext_GetWorkingDir(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}
	workingDir := "/test/working/dir"

	ctx := app.NewPluginContext(logger, "/test/db", workingDir, eventRepo)

	if got := ctx.GetWorkingDir(); got != workingDir {
		t.Errorf("GetWorkingDir() = %q, want %q", got, workingDir)
	}
}

func TestPluginContext_EmitEvent(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}

	pluginCtx := app.NewPluginContext(logger, "/test/db", "/test/dir", eventRepo)

	event := domain.PluginEvent{
		Type:      "test.event",
		Source:    "test-plugin",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"key": "value",
		},
		Metadata: map[string]string{
			"session_id": "test-session-123",
			"meta":       "data",
		},
	}

	ctx := context.Background()
	err := pluginCtx.EmitEvent(ctx, event)
	if err != nil {
		t.Fatalf("EmitEvent() error = %v", err)
	}

	// Verify event was stored
	if len(eventRepo.events) != 1 {
		t.Fatalf("Expected 1 event stored, got %d", len(eventRepo.events))
	}

	stored := eventRepo.events[0]
	if string(stored.Type) != event.Type {
		t.Errorf("Event type = %q, want %q", stored.Type, event.Type)
	}
	if stored.SessionID != "test-session-123" {
		t.Errorf("Event session ID = %q, want %q", stored.SessionID, "test-session-123")
	}

	// Verify payload contains source and data
	payload, ok := stored.Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected payload to be map[string]interface{}, got %T", stored.Payload)
	}

	if payload["source"] != event.Source {
		t.Errorf("Payload source = %q, want %q", payload["source"], event.Source)
	}

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected payload data to be map[string]interface{}, got %T", payload["data"])
	}

	if data["key"] != "value" {
		t.Errorf("Payload data key = %q, want %q", data["key"], "value")
	}
}

func TestPluginContext_EmitEvent_NoMetadata(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}

	pluginCtx := app.NewPluginContext(logger, "/test/db", "/test/dir", eventRepo)

	event := domain.PluginEvent{
		Type:      "test.event",
		Source:    "test-plugin",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"key": "value",
		},
		// No metadata
	}

	ctx := context.Background()
	err := pluginCtx.EmitEvent(ctx, event)
	if err != nil {
		t.Fatalf("EmitEvent() error = %v", err)
	}

	// Verify event was stored
	if len(eventRepo.events) != 1 {
		t.Fatalf("Expected 1 event stored, got %d", len(eventRepo.events))
	}

	stored := eventRepo.events[0]
	// Session ID should be empty when no metadata provided
	if stored.SessionID != "" {
		t.Errorf("Event session ID = %q, want empty string", stored.SessionID)
	}
}

func TestPluginContext_EmitEvent_RepositoryError(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{
		saveError: context.DeadlineExceeded,
	}

	pluginCtx := app.NewPluginContext(logger, "/test/db", "/test/dir", eventRepo)

	event := domain.PluginEvent{
		Type:      "test.event",
		Source:    "test-plugin",
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{},
	}

	ctx := context.Background()
	err := pluginCtx.EmitEvent(ctx, event)
	if err == nil {
		t.Fatal("EmitEvent() expected error, got nil")
	}
}

func TestCommandContext_GetStdout(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}

	stdout := &bytes.Buffer{}
	stdin := &bytes.Buffer{}

	cmdCtx := app.NewCommandContext(logger, "/test/db", "/test/dir", eventRepo, stdout, stdin)

	if got := cmdCtx.GetStdout(); got != stdout {
		t.Errorf("GetStdout() returned wrong writer")
	}
}

func TestCommandContext_GetStdin(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}

	stdout := &bytes.Buffer{}
	stdin := &bytes.Buffer{}

	cmdCtx := app.NewCommandContext(logger, "/test/db", "/test/dir", eventRepo, stdout, stdin)

	if got := cmdCtx.GetStdin(); got != stdin {
		t.Errorf("GetStdin() returned wrong reader")
	}
}

func TestCommandContext_InheritsPluginContext(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}
	workingDir := "/test/dir"

	stdout := &bytes.Buffer{}
	stdin := &bytes.Buffer{}

	cmdCtx := app.NewCommandContext(logger, "/test/db", workingDir, eventRepo, stdout, stdin)

	// Verify CommandContext has access to PluginContext methods
	if got := cmdCtx.GetLogger(); got == nil {
		t.Error("CommandContext.GetLogger() returned nil")
	}

	if got := cmdCtx.GetWorkingDir(); got != workingDir {
		t.Errorf("CommandContext.GetWorkingDir() = %q, want %q", got, workingDir)
	}

	// Test event emission through CommandContext
	event := domain.PluginEvent{
		Type:      "cmd.test",
		Source:    "test",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"test": "data",
		},
	}

	ctx := context.Background()
	if err := cmdCtx.EmitEvent(ctx, event); err != nil {
		t.Fatalf("CommandContext.EmitEvent() error = %v", err)
	}

	if len(eventRepo.events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(eventRepo.events))
	}
}

func TestLoggerAdapter_Methods(t *testing.T) {
	logger := &mockPluginContextLogger{}
	eventRepo := &mockEventRepo{}

	pluginCtx := app.NewPluginContext(logger, "/test/db", "/test/dir", eventRepo)
	sdkLogger := pluginCtx.GetLogger()

	// These should not panic
	sdkLogger.Debug("debug message")
	sdkLogger.Info("info message")
	sdkLogger.Warn("warn message")
	sdkLogger.Error("error message")
}

package claude_code_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// simpleCommandContext is a minimal implementation of pluginsdk.CommandContext for testing
type simpleCommandContext struct {
	stdin      io.Reader
	stdout     io.Writer
	workingDir string
}

func (m *simpleCommandContext) GetLogger() pluginsdk.Logger {
	return &mockLogger{}
}

func (m *simpleCommandContext) GetWorkingDir() string {
	// No default - tests must explicitly set workingDir
	// This prevents accidental pollution of production or system paths
	return m.workingDir
}

func (m *simpleCommandContext) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	return nil
}

func (m *simpleCommandContext) GetStdin() io.Reader {
	return m.stdin
}

func (m *simpleCommandContext) GetStdout() io.Writer {
	return m.stdout
}

// mockCommandContext implements pluginsdk.CommandContext for testing with event tracking
type mockCommandContext struct {
	stdin      io.Reader
	stdout     io.Writer
	emitErr    error
	events     []pluginsdk.Event
	workingDir string
}

func (m *mockCommandContext) GetLogger() pluginsdk.Logger {
	return &mockLogger{}
}

func (m *mockCommandContext) GetWorkingDir() string {
	// No default - tests must explicitly set workingDir
	// This prevents accidental pollution of production or system paths
	return m.workingDir
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

// errorReader always returns an error on Read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// mockConfigLoader implements claude_code.ConfigLoader for testing
type mockConfigLoader struct {
	config *claude_code.Config
	err    error
}

func (m *mockConfigLoader) LoadConfig(path string) (*claude_code.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.config, nil
}

// newTestEmitEventCommand creates an EmitEventCommand with default test config
func newTestEmitEventCommand() *claude_code.EmitEventCommand {
	configLoader := &mockConfigLoader{
		config: &claude_code.Config{
			Logging: claude_code.LoggingConfig{
				FileLogLevel: "debug", // Use debug for tests to capture all logging
			},
		},
	}
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, configLoader, "")
	return claude_code.NewEmitEventCommand(plugin)
}

// newTestCommandContext creates a mockCommandContext with a temp directory for isolated testing
func newTestCommandContext(t *testing.T, stdin io.Reader, stdout io.Writer) *mockCommandContext {
	return &mockCommandContext{
		stdin:      stdin,
		stdout:     stdout,
		workingDir: t.TempDir(), // Each test gets its own isolated temp directory
	}
}

// TestLogCommand_Execute tests the deprecated LogCommand.Execute
// This command is kept for backward compatibility and simply logs a warning
func TestLogCommand_Execute(t *testing.T) {
	ctx := context.Background()
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	// Get the command from the plugin's GetCommands() - this is the proper way to test
	commands := plugin.GetCommands()
	var logCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "log" {
			logCmd = cmd
			break
		}
	}

	if logCmd == nil {
		t.Fatal("LogCommand not found in plugin commands")
	}

	// Create mock context
	cmdCtx := &simpleCommandContext{
		stdin:  strings.NewReader(""),
		stdout: &bytes.Buffer{},
	}

	// Execute should return nil (deprecated, fails silently)
	err := logCmd.Execute(ctx, cmdCtx, []string{"test-event"})
	if err != nil {
		t.Errorf("LogCommand.Execute returned error: %v", err)
	}
}

// TestInitCommand_Execute tests InitCommand.Execute
func TestInitCommand_Execute(t *testing.T) {
	tests := []struct {
		name             string
		setupService     claude_code.SetupService
		expectError      bool
		expectInOutput   string
	}{
		{
			name: "successful_initialization",
			setupService: &mockSetupService{
				err: nil,
			},
			expectError:    false,
			expectInOutput: "Initializing Claude Code logging",
		},
		{
			name: "database_initialization_error",
			setupService: &mockSetupService{
				err: io.ErrClosedPipe,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			plugin := claude_code.NewClaudeCodePlugin(
				nil, nil, &mockLogger{}, tt.setupService, nil, "/tmp/test.db",
			)

			// Get init command from plugin
			commands := plugin.GetCommands()
			var initCmd pluginsdk.Command
			for _, cmd := range commands {
				if cmd.GetName() == "init" {
					initCmd = cmd
					break
				}
			}

			if initCmd == nil {
				t.Fatal("InitCommand not found in plugin commands")
			}

			// Create mock context
			stdout := &bytes.Buffer{}
			cmdCtx := &simpleCommandContext{
				stdin:  strings.NewReader(""),
				stdout: stdout,
			}

			// Execute the command
			err := initCmd.Execute(ctx, cmdCtx, []string{})

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("InitCommand.Execute error = %v, expectError = %v", err, tt.expectError)
			}

			// Check output if no error
			if !tt.expectError && tt.expectInOutput != "" {
				output := stdout.String()
				if !strings.Contains(output, tt.expectInOutput) {
					t.Errorf("Expected output to contain %q, got: %s", tt.expectInOutput, output)
				}
			}
		})
	}
}

// mockSetupService implements claude_code.SetupService for testing
type mockSetupService struct {
	err error
}

func (m *mockSetupService) Initialize(ctx context.Context, dbPath string) error {
	return m.err
}

// TestEmitEventCommand_parseAsHookInput tests the parseAsHookInput method
// This is a wrapper around the hook input parser
func TestEmitEventCommand_parseAsHookInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name: "valid_hook_input",
			input: `{
				"session_id": "sess-123",
				"hook_event_name": "PreToolUse",
				"tool_name": "Read"
			}`,
			expectErr: false,
		},
		{
			name:      "invalid_json",
			input:     `{invalid json`,
			expectErr: true,
		},
		{
			name: "minimal_valid_hook_input",
			input: `{
				"session_id": "sess-456",
				"hook_event_name": "SessionStart"
			}`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			configLoader := &mockConfigLoader{
				config: &claude_code.Config{
					Logging: claude_code.LoggingConfig{
						FileLogLevel: "debug",
					},
				},
			}
			plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, configLoader, "")

			// Create emit event command
			emitCmd := claude_code.NewEmitEventCommand(plugin)

			// Create mock context
			cmdCtx := &simpleCommandContext{
				stdin:  strings.NewReader(tt.input),
				stdout: &bytes.Buffer{},
			}

			// Execute with HookInput format in stdin
			// The command internally calls parseAsHookInput
			err := emitCmd.Execute(ctx, cmdCtx, []string{})

			// parseAsHookInput is only called when the input doesn't have type/source fields
			// So we just verify the command handles it without panicking
			if err != nil && !tt.expectErr {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSessionSummaryCommand_Execute_WithSessionID tests SessionSummaryCommand with --session-id
func TestSessionSummaryCommand_Execute_WithSessionID(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		logsService   claude_code.LogsService
		buildError    bool
		expectInOutput string
	}{
		{
			name:      "valid_session_id",
			sessionID: "session-1",
			logsService: &mockLogsService{
				logs: []*claude_code.LogRecord{
					{
						ID:        "event-1",
						SessionID: "session-1",
						EventType: "tool.invoked",
					},
				},
			},
			expectInOutput: "Session Summary",
		},
		{
			name:      "invalid_session_id",
			sessionID: "nonexistent",
			logsService: &mockLogsService{
				logs: []*claude_code.LogRecord{},
			},
			buildError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := &mockLogger{}

			// Create plugin with mock services
			plugin := claude_code.NewClaudeCodePlugin(
				&mockAnalysisService{
					sessionIDs: []string{"session-1"},
					analyses:   map[string][]*claude_code.SessionAnalysis{},
				},
				tt.logsService,
				logger,
				nil,
				nil,
				"",
			)

			// Get session summary command from plugin
			commands := plugin.GetCommands()
			var summaryCmd pluginsdk.Command
			for _, cmd := range commands {
				if cmd.GetName() == "session-summary" {
					summaryCmd = cmd
					break
				}
			}

			if summaryCmd == nil {
				t.Fatal("SessionSummaryCommand not found in plugin commands")
			}

			// Create mock context
			stdout := &bytes.Buffer{}
			cmdCtx := &simpleCommandContext{
				stdin:  strings.NewReader(""),
				stdout: stdout,
			}

			// Execute with --session-id flag
			args := []string{"--session-id", tt.sessionID}
			err := summaryCmd.Execute(ctx, cmdCtx, args)

			// Check error
			if (err != nil) != tt.buildError {
				t.Errorf("SessionSummaryCommand.Execute error = %v, buildError = %v", err, tt.buildError)
			}

			// Check output if no error
			if !tt.buildError && tt.expectInOutput != "" {
				output := stdout.String()
				if !strings.Contains(output, tt.expectInOutput) {
					t.Errorf("Expected output to contain %q, got: %s", tt.expectInOutput, output)
				}
			}
		})
	}
}

// TestSessionSummaryCommand_Execute_WithLastFlag tests SessionSummaryCommand with --last
func TestSessionSummaryCommand_Execute_WithLastFlag(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}
	sessionID := "session-latest"

	// Create plugin with mock services
	plugin := claude_code.NewClaudeCodePlugin(
		&mockAnalysisService{
			sessionIDs: []string{sessionID},
			analyses:   map[string][]*claude_code.SessionAnalysis{},
		},
		&mockLogsService{
			logs: []*claude_code.LogRecord{
				{
					ID:        "event-1",
					SessionID: sessionID,
					EventType: "tool.invoked",
				},
			},
		},
		logger,
		nil,
		nil,
		"",
	)

	// Get session summary command from plugin
	commands := plugin.GetCommands()
	var summaryCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "session-summary" {
			summaryCmd = cmd
			break
		}
	}

	if summaryCmd == nil {
		t.Fatal("SessionSummaryCommand not found in plugin commands")
	}

	// Create mock context
	stdout := &bytes.Buffer{}
	cmdCtx := &simpleCommandContext{
		stdin:  strings.NewReader(""),
		stdout: stdout,
	}

	// Execute with --last flag
	err := summaryCmd.Execute(ctx, cmdCtx, []string{"--last"})
	if err != nil {
		t.Errorf("SessionSummaryCommand.Execute with --last returned error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Session Summary") {
		t.Errorf("Expected output to contain 'Session Summary', got: %s", output)
	}
}

// TestSessionSummaryCommand_Execute_MissingArguments tests error handling
func TestSessionSummaryCommand_Execute_MissingArguments(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	plugin := claude_code.NewClaudeCodePlugin(nil, nil, logger, nil, nil, "")

	// Get session summary command from plugin
	commands := plugin.GetCommands()
	var summaryCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "session-summary" {
			summaryCmd = cmd
			break
		}
	}

	if summaryCmd == nil {
		t.Fatal("SessionSummaryCommand not found in plugin commands")
	}

	cmdCtx := &simpleCommandContext{
		stdin:  strings.NewReader(""),
		stdout: &bytes.Buffer{},
	}

	// Execute with no arguments
	err := summaryCmd.Execute(ctx, cmdCtx, []string{})
	if err == nil {
		t.Error("Expected error when no arguments provided, got nil")
	}
	if !strings.Contains(err.Error(), "must specify either") {
		t.Errorf("Expected error message about arguments, got: %v", err)
	}
}

// TestAutoSummaryExecCommand_Execute tests AutoSummaryExecCommand.Execute
func TestAutoSummaryExecCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		configLoader   claude_code.ConfigLoader
		analysisEngine claude_code.AnalysisService
		expectError    bool
	}{
		{
			name:      "successful_exec",
			sessionID: "session-1",
			configLoader: &mockConfigLoader{
				config: &claude_code.Config{
					Analysis: claude_code.AnalysisConfig{
						AutoSummaryEnabled: true,
						AutoSummaryPrompt:  "session_summary",
					},
				},
			},
			analysisEngine: &mockAnalysisService{
				sessionIDs: []string{"session-1"},
				analyses:   map[string][]*claude_code.SessionAnalysis{},
			},
			expectError: false,
		},
		{
			name:      "no_session_id",
			sessionID: "",
			configLoader: &mockConfigLoader{
				config: &claude_code.Config{
					Analysis: claude_code.AnalysisConfig{
						AutoSummaryEnabled: true,
					},
				},
			},
			expectError: false, // Returns nil when no session ID
		},
		{
			name:         "config_load_error",
			sessionID:    "session-2",
			configLoader: &mockConfigLoader{err: io.ErrClosedPipe},
			expectError:  false, // Returns nil when config load fails (best-effort)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			plugin := claude_code.NewClaudeCodePlugin(
				tt.analysisEngine,
				nil,
				&mockLogger{},
				nil,
				tt.configLoader,
				"",
			)

			// Get auto summary exec command from plugin
			commands := plugin.GetCommands()
			var autoExecCmd pluginsdk.Command
			for _, cmd := range commands {
				if cmd.GetName() == "auto-summary-exec" {
					autoExecCmd = cmd
					break
				}
			}

			if autoExecCmd == nil {
				t.Fatal("AutoSummaryExecCommand not found in plugin commands")
			}

			// Create mock context
			cmdCtx := &simpleCommandContext{
				stdin:  strings.NewReader(""),
				stdout: &bytes.Buffer{},
			}

			// Prepare args
			var args []string
			if tt.sessionID != "" {
				args = []string{tt.sessionID}
			}

			// Execute the command
			err := autoExecCmd.Execute(ctx, cmdCtx, args)

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("AutoSummaryExecCommand.Execute error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}

// TestAutoSummaryCommand_Execute tests AutoSummaryCommand.Execute
func TestAutoSummaryCommand_Execute(t *testing.T) {
	tests := []struct {
		name         string
		hookInput    map[string]interface{}
		configLoader claude_code.ConfigLoader
		expectError  bool
	}{
		{
			name: "auto_summary_enabled",
			hookInput: map[string]interface{}{
				"session_id":      "session-1",
				"hook_event_name": "SessionEnd",
			},
			configLoader: &mockConfigLoader{
				config: &claude_code.Config{
					Analysis: claude_code.AnalysisConfig{
						AutoSummaryEnabled: true,
					},
				},
			},
			expectError: false,
		},
		{
			name: "auto_summary_disabled",
			hookInput: map[string]interface{}{
				"session_id":      "session-2",
				"hook_event_name": "SessionEnd",
			},
			configLoader: &mockConfigLoader{
				config: &claude_code.Config{
					Analysis: claude_code.AnalysisConfig{
						AutoSummaryEnabled: false,
					},
				},
			},
			expectError: false, // Returns nil when disabled
		},
		{
			name:         "invalid_hook_input",
			hookInput:    map[string]interface{}{},
			configLoader: &mockConfigLoader{
				config: &claude_code.Config{
					Analysis: claude_code.AnalysisConfig{
						AutoSummaryEnabled: true,
					},
				},
			},
			expectError: false, // Returns nil when parsing fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			plugin := claude_code.NewClaudeCodePlugin(
				&mockAnalysisService{
					sessionIDs: []string{"session-1", "session-2"},
					analyses:   map[string][]*claude_code.SessionAnalysis{},
				},
				nil,
				&mockLogger{},
				nil,
				tt.configLoader,
				"",
			)

			// Get auto summary command from plugin
			commands := plugin.GetCommands()
			var autoSummaryCmd pluginsdk.Command
			for _, cmd := range commands {
				if cmd.GetName() == "auto-summary" {
					autoSummaryCmd = cmd
					break
				}
			}

			if autoSummaryCmd == nil {
				t.Fatal("AutoSummaryCommand not found in plugin commands")
			}

			// Serialize hook input
			hookData, _ := json.Marshal(tt.hookInput)

			// Create mock context
			cmdCtx := &simpleCommandContext{
				stdin:  bytes.NewReader(hookData),
				stdout: &bytes.Buffer{},
			}

			// Execute the command
			err := autoSummaryCmd.Execute(ctx, cmdCtx, []string{})

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("AutoSummaryCommand.Execute error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}

// TestAutoSummaryCommand_Execute_EmptyStdin tests handling of empty stdin
func TestAutoSummaryCommand_Execute_EmptyStdin(t *testing.T) {
	ctx := context.Background()
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	// Get auto summary command from plugin
	commands := plugin.GetCommands()
	var autoSummaryCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "auto-summary" {
			autoSummaryCmd = cmd
			break
		}
	}

	if autoSummaryCmd == nil {
		t.Fatal("AutoSummaryCommand not found in plugin commands")
	}

	cmdCtx := &simpleCommandContext{
		stdin:  strings.NewReader(""),
		stdout: &bytes.Buffer{},
	}

	// Execute should return nil (fails silently on empty input)
	err := autoSummaryCmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Errorf("AutoSummaryCommand.Execute with empty stdin returned error: %v", err)
	}
}

// TestInitCommand_WithHookInstallationWarning tests that hook installation warnings don't fail
func TestInitCommand_WithHookInstallationWarning(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	os.Chdir(tmpDir)

	ctx := context.Background()
	plugin := claude_code.NewClaudeCodePlugin(
		nil, nil, &mockLogger{}, &mockSetupService{}, nil, "/tmp/test.db",
	)

	// Get init command from plugin
	commands := plugin.GetCommands()
	var initCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "init" {
			initCmd = cmd
			break
		}
	}

	if initCmd == nil {
		t.Fatal("InitCommand not found in plugin commands")
	}

	stdout := &bytes.Buffer{}
	cmdCtx := &simpleCommandContext{
		stdin:  strings.NewReader(""),
		stdout: stdout,
	}

	// Execute should succeed even if hook manager creation fails
	err := initCmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Errorf("InitCommand.Execute returned error even with hook failures: %v", err)
	}
}

// TestAutoSummaryExecCommand_DefaultPrompt tests default prompt name when not configured
func TestAutoSummaryExecCommand_DefaultPrompt(t *testing.T) {
	ctx := context.Background()

	// Config with empty prompt name (should use default)
	configLoader := &mockConfigLoader{
		config: &claude_code.Config{
			Analysis: claude_code.AnalysisConfig{
				AutoSummaryEnabled: true,
				AutoSummaryPrompt:  "", // Empty - should use default
			},
		},
	}

	analysisService := &mockAnalysisService{
		sessionIDs: []string{"session-1"},
		analyses:   map[string][]*claude_code.SessionAnalysis{},
	}

	plugin := claude_code.NewClaudeCodePlugin(
		analysisService,
		nil,
		&mockLogger{},
		nil,
		configLoader,
		"",
	)

	// Get auto summary exec command from plugin
	commands := plugin.GetCommands()
	var autoExecCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "auto-summary-exec" {
			autoExecCmd = cmd
			break
		}
	}

	if autoExecCmd == nil {
		t.Fatal("AutoSummaryExecCommand not found in plugin commands")
	}

	cmdCtx := &simpleCommandContext{
		stdin:  strings.NewReader(""),
		stdout: &bytes.Buffer{},
	}

	// Execute should use "session_summary" as default prompt
	err := autoExecCmd.Execute(ctx, cmdCtx, []string{"session-1"})
	if err != nil {
		t.Errorf("AutoSummaryExecCommand.Execute returned error: %v", err)
	}
}

// TestNewEmitEventCommand verifies the command can be created
func TestNewEmitEventCommand(t *testing.T) {
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

	// Create a reader that returns an error
	errReader := &errorReader{}
	stdout := &bytes.Buffer{}
	mockCtx := newTestCommandContext(t, errReader, stdout)

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

// TestEmitEventCommand_LargePayload verifies large payloads are handled
func TestEmitEventCommand_LargePayload(t *testing.T) {
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

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
	cmd := newTestEmitEventCommand()

	// Verify the command implements pluginsdk.Command interface
	var _ pluginsdk.Command = cmd
}

// TestEmitEventCommand_FileLogging verifies that errors are logged to .darwinflow/claude-code.log
func TestEmitEventCommand_FileLogging(t *testing.T) {
	// Test cases for different error scenarios
	testCases := []struct {
		name        string
		input       string
		expectLog   bool
		logContains string
	}{
		{
			name:        "empty stdin",
			input:       "",
			expectLog:   true,
			logContains: "empty stdin",
		},
		{
			name:        "invalid JSON",
			input:       "{invalid json}",
			expectLog:   true,
			logContains: "invalid JSON",
		},
		{
			name:        "missing session_id",
			input:       `{"type":"test","source":"test"}`,
			expectLog:   true,
			logContains: "missing required field: metadata.session_id",
		},
		{
			name:        "successful event",
			input:       `{"type":"test","source":"test","metadata":{"session_id":"test123"}}`,
			expectLog:   true,
			logContains: "successfully emitted event",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Each subtest gets its own isolated temp directory
			tmpDir := t.TempDir()

			// Create config loader with debug logging to capture all messages
			configLoader := &mockConfigLoader{
				config: &claude_code.Config{
					Logging: claude_code.LoggingConfig{
						FileLogLevel: "debug",
					},
				},
			}

			// Create plugin and command
			plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, configLoader, "")
			cmd := claude_code.NewEmitEventCommand(plugin)

			// Create mock context with temp directory
			mockCtx := &mockCommandContext{
				stdin:      strings.NewReader(tc.input),
				stdout:     &bytes.Buffer{},
				workingDir: tmpDir,
			}

			// Execute command
			err := cmd.Execute(context.Background(), mockCtx, []string{})
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			// Check if log file was created
			logPath := tmpDir + "/.darwinflow/claude-code.log"
			if tc.expectLog {
				// Read log file
				logContent, err := os.ReadFile(logPath)
				if err != nil {
					t.Fatalf("Failed to read log file: %v", err)
				}

				// Verify log contains expected message
				if !strings.Contains(string(logContent), tc.logContains) {
					t.Errorf("Log does not contain %q. Log content:\n%s", tc.logContains, string(logContent))
				}

				// Verify log has timestamp
				if !strings.Contains(string(logContent), "[202") {
					t.Errorf("Log missing timestamp. Log content:\n%s", string(logContent))
				}
			}
		})
	}
}

// TestEmitEventCommand_LogLevelFiltering verifies that log level filtering works correctly
func TestEmitEventCommand_LogLevelFiltering(t *testing.T) {
	testCases := []struct {
		name             string
		logLevel         string
		input            string
		shouldLogDebug   bool
		shouldLogInfo    bool
		shouldLogError   bool
	}{
		{
			name:           "log level off - no logging",
			logLevel:       "off",
			input:          `{"type":"test","source":"test"}`, // Missing session_id - DEBUG error
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogError: false,
		},
		{
			name:           "log level error - only errors",
			logLevel:       "error",
			input:          `{"type":"test","source":"test"}`, // Missing session_id - DEBUG error
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogError: false, // This is DEBUG level, not ERROR
		},
		{
			name:           "log level info - info and error",
			logLevel:       "info",
			input:          `{"type":"test","source":"test","metadata":{"session_id":"test123"}}`,
			shouldLogDebug: false,
			shouldLogInfo:  true,  // Success message is INFO
			shouldLogError: false, // No actual ERROR occurs with valid input
		},
		{
			name:           "log level debug - all messages",
			logLevel:       "debug",
			input:          `{"type":"test","source":"test"}`, // Missing session_id - DEBUG error
			shouldLogDebug: true,
			shouldLogInfo:  false, // No successful event, so no INFO
			shouldLogError: false, // DEBUG messages, not ERROR
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directory for each test
			tmpDir := t.TempDir()

			// Create config loader with specified log level
			configLoader := &mockConfigLoader{
				config: &claude_code.Config{
					Logging: claude_code.LoggingConfig{
						FileLogLevel: tc.logLevel,
					},
				},
			}

			// Create plugin and command
			plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, configLoader, "")
			cmd := claude_code.NewEmitEventCommand(plugin)

			// Create mock context
			mockCtx := &mockCommandContext{
				stdin:      strings.NewReader(tc.input),
				stdout:     &bytes.Buffer{},
				workingDir: tmpDir,
			}

			// Execute command
			err := cmd.Execute(context.Background(), mockCtx, []string{})
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			// Check log file
			logPath := tmpDir + "/.darwinflow/claude-code.log"
			logContent := ""
			if data, err := os.ReadFile(logPath); err == nil {
				logContent = string(data)
			}

			// Verify logging behavior
			hasDebug := strings.Contains(logContent, "DEBUG:")
			hasInfo := strings.Contains(logContent, "INFO:")
			hasError := strings.Contains(logContent, "ERROR:")

			if hasDebug != tc.shouldLogDebug {
				t.Errorf("DEBUG logging: got %v, want %v. Log:\n%s", hasDebug, tc.shouldLogDebug, logContent)
			}
			if hasInfo != tc.shouldLogInfo {
				t.Errorf("INFO logging: got %v, want %v. Log:\n%s", hasInfo, tc.shouldLogInfo, logContent)
			}
			if hasError != tc.shouldLogError {
				t.Errorf("ERROR logging: got %v, want %v. Log:\n%s", hasError, tc.shouldLogError, logContent)
			}
		})
	}
}

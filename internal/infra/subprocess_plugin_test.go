package infra_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// TestSubprocessPlugin_Initialize tests plugin initialization.
func TestSubprocessPlugin_Initialize(t *testing.T) {
	pluginPath := buildExternalPlugin(t)

	plugin := infra.NewSubprocessPlugin(pluginPath)
	ctx := context.Background()

	err := plugin.Initialize(ctx, "/tmp", nil)
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}
	defer plugin.Shutdown()

	// Verify plugin info
	info := plugin.GetInfo()
	if info.Name != "test-external" {
		t.Errorf("expected name 'test-external', got %s", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", info.Version)
	}

	// Verify capabilities
	capabilities := plugin.GetCapabilities()
	if len(capabilities) < 1 {
		t.Error("expected at least one capability")
	}
}

// TestSubprocessPlugin_EntityProvider tests entity querying.
func TestSubprocessPlugin_EntityProvider(t *testing.T) {
	pluginPath := buildExternalPlugin(t)

	plugin := infra.NewSubprocessPlugin(pluginPath)
	ctx := context.Background()

	if err := plugin.Initialize(ctx, "/tmp", nil); err != nil {
		t.Fatalf("initialization failed: %v", err)
	}
	defer plugin.Shutdown()

	// Get entity types
	entityTypes := plugin.GetEntityTypes()
	if len(entityTypes) == 0 {
		t.Fatal("expected at least one entity type")
	}

	// Query entities
	query := pluginsdk.EntityQuery{
		EntityType: "note",
		Limit:      10,
	}
	entities, err := plugin.Query(ctx, query)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(entities))
	}

	// Get specific entity
	entity, err := plugin.GetEntity(ctx, "note-1")
	if err != nil {
		t.Fatalf("get entity failed: %v", err)
	}

	if entity.GetID() != "note-1" {
		t.Errorf("expected entity ID 'note-1', got %s", entity.GetID())
	}
	if entity.GetType() != "note" {
		t.Errorf("expected entity type 'note', got %s", entity.GetType())
	}

	// Check field access
	title := entity.GetField("title")
	if title != "First Note" {
		t.Errorf("expected title 'First Note', got %v", title)
	}
}

// TestSubprocessPlugin_EntityUpdater tests entity updates.
func TestSubprocessPlugin_EntityUpdater(t *testing.T) {
	pluginPath := buildExternalPlugin(t)

	plugin := infra.NewSubprocessPlugin(pluginPath)
	ctx := context.Background()

	if err := plugin.Initialize(ctx, "/tmp", nil); err != nil {
		t.Fatalf("initialization failed: %v", err)
	}
	defer plugin.Shutdown()

	// Update entity
	fields := map[string]interface{}{
		"title": "Updated Title",
	}
	updated, err := plugin.UpdateEntity(ctx, "note-1", fields)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Verify update
	title := updated.GetField("title")
	if title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %v", title)
	}
}

// TestSubprocessPlugin_CommandProvider tests command execution.
func TestSubprocessPlugin_CommandProvider(t *testing.T) {
	pluginPath := buildExternalPlugin(t)

	plugin := infra.NewSubprocessPlugin(pluginPath)
	ctx := context.Background()

	if err := plugin.Initialize(ctx, "/tmp", nil); err != nil {
		t.Fatalf("initialization failed: %v", err)
	}
	defer plugin.Shutdown()

	// Get commands
	commands := plugin.GetCommands()
	if len(commands) == 0 {
		t.Fatal("expected at least one command")
	}

	// Find test command
	var testCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "test" {
			testCmd = cmd
			break
		}
	}
	if testCmd == nil {
		t.Fatal("test command not found")
	}

	// Execute command
	cmdCtx := &mockCommandContext{
		output: &bytes.Buffer{},
	}
	err := testCmd.Execute(context.Background(), cmdCtx, []string{"arg1", "arg2"})
	if err != nil {
		t.Errorf("command execution failed: %v", err)
	}

	// Verify output
	if cmdCtx.output.Len() == 0 {
		t.Error("expected command output")
	}
}

// TestSubprocessPlugin_EventEmitter tests event streaming.
func TestSubprocessPlugin_EventEmitter(t *testing.T) {
	pluginPath := buildExternalPlugin(t)

	plugin := infra.NewSubprocessPlugin(pluginPath)
	ctx := context.Background()

	if err := plugin.Initialize(ctx, "/tmp", nil); err != nil {
		t.Fatalf("initialization failed: %v", err)
	}
	defer plugin.Shutdown()

	// Start event stream
	eventChan := make(chan pluginsdk.Event, 10)
	err := plugin.StartEventStream(ctx, eventChan)
	if err != nil {
		t.Fatalf("failed to start event stream: %v", err)
	}
	defer plugin.StopEventStream()

	// Wait for events (the plugin emits a test event on stream start)
	timeout := time.After(2 * time.Second)
	select {
	case event := <-eventChan:
		if event.Type != "test.started" {
			t.Errorf("expected event type 'test.started', got %s", event.Type)
		}
	case <-timeout:
		t.Fatal("timeout waiting for event")
	}
}

// buildExternalPlugin creates a test external plugin executable.
func buildExternalPlugin(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-plugin")

	// Plugin source code (implements JSON-RPC protocol)
	pluginSrc := `package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Request struct {
	JSONRPC string          ` + "`json:\"jsonrpc\"`" + `
	ID      interface{}     ` + "`json:\"id\"`" + `
	Method  string          ` + "`json:\"method\"`" + `
	Params  json.RawMessage ` + "`json:\"params,omitempty\"`" + `
}

type Response struct {
	JSONRPC string          ` + "`json:\"jsonrpc\"`" + `
	ID      interface{}     ` + "`json:\"id\"`" + `
	Result  json.RawMessage ` + "`json:\"result,omitempty\"`" + `
	Error   *RPCError       ` + "`json:\"error,omitempty\"`" + `
}

type RPCError struct {
	Code    int    ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

type Event struct {
	Event     string                 ` + "`json:\"event\"`" + `
	Type      string                 ` + "`json:\"type\"`" + `
	Source    string                 ` + "`json:\"source\"`" + `
	Timestamp string                 ` + "`json:\"timestamp\"`" + `
	Payload   map[string]interface{} ` + "`json:\"payload,omitempty\"`" + `
}

var entities = []map[string]interface{}{
	{"id": "note-1", "type": "note", "title": "First Note", "capabilities": []string{}},
	{"id": "note-2", "type": "note", "title": "Second Note", "capabilities": []string{}},
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}

		var result interface{}
		var err *RPCError

		switch req.Method {
		case "init":
			result = nil
		case "get_info":
			result = map[string]interface{}{
				"name":        "test-external",
				"version":     "1.0.0",
				"description": "Test external plugin",
				"is_core":     false,
			}
		case "get_capabilities":
			result = []string{"IEntityProvider", "IEntityUpdater", "ICommandProvider", "IEventEmitter"}
		case "get_entity_types":
			result = []map[string]interface{}{
				{
					"type":                "note",
					"display_name":        "Note",
					"display_name_plural": "Notes",
					"capabilities":        []string{},
					"icon":                "📝",
					"description":         "A simple note",
				},
			}
		case "query_entities":
			result = entities
		case "get_entity":
			var params map[string]string
			json.Unmarshal(req.Params, &params)
			entityID := params["entity_id"]
			for _, e := range entities {
				if e["id"] == entityID {
					result = e
					break
				}
			}
			if result == nil {
				err = &RPCError{Code: -32000, Message: "entity not found"}
			}
		case "update_entity":
			var params map[string]interface{}
			json.Unmarshal(req.Params, &params)
			entityID := params["entity_id"].(string)
			fields := params["fields"].(map[string]interface{})
			for _, e := range entities {
				if e["id"] == entityID {
					for k, v := range fields {
						e[k] = v
					}
					result = e
					break
				}
			}
		case "get_commands":
			result = []map[string]interface{}{
				{
					"name":        "test",
					"description": "Test command",
					"usage":       "test [args...]",
					"help":        "A test command",
				},
			}
		case "execute_command":
			result = map[string]interface{}{
				"exit_code": 0,
				"output":    "Command executed successfully\n",
				"error":     "",
			}
		case "start_event_stream":
			result = nil
			// Emit a test event
			event := Event{
				Event:     "event",
				Type:      "test.started",
				Source:    "test-external",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(os.Stdout, "%s\n", string(data))
		case "stop_event_stream":
			result = nil
		default:
			err = &RPCError{Code: -32601, Message: "method not found"}
		}

		// Send response
		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   err,
		}
		if err == nil && result != nil {
			resp.Result, _ = json.Marshal(result)
		}

		data, _ := json.Marshal(resp)
		fmt.Fprintf(os.Stdout, "%s\n", string(data))
	}
}
`

	srcPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(pluginSrc), 0644); err != nil {
		t.Fatalf("failed to write plugin source: %v", err)
	}

	// Compile plugin
	cmd := exec.Command("go", "build", "-o", pluginPath, srcPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build plugin: %v\nOutput: %s", err, output)
	}

	return pluginPath
}

// mockCommandContext is a test implementation of CommandContext.
type mockCommandContext struct {
	output *bytes.Buffer
}

func (m *mockCommandContext) GetLogger() pluginsdk.Logger {
	return &mockLogger{}
}

func (m *mockCommandContext) GetWorkingDir() string {
	return "/tmp"
}

func (m *mockCommandContext) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	return nil
}

func (m *mockCommandContext) GetStdout() io.Writer {
	return m.output
}

func (m *mockCommandContext) GetStdin() io.Reader {
	return bytes.NewReader(nil)
}

// mockLogger is a no-op logger for testing.
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {}

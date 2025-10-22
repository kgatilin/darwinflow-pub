package infra_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// TestRPCClient_StartStop tests basic subprocess lifecycle.
func TestRPCClient_StartStop(t *testing.T) {
	pluginPath := buildTestPlugin(t)

	client := infra.NewRPCClient(pluginPath, "echo")
	ctx := context.Background()

	// Start client
	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Stop client
	if err := client.Stop(); err != nil {
		t.Fatalf("failed to stop client: %v", err)
	}
}

// TestRPCClient_CallSuccess tests successful RPC call with response.
func TestRPCClient_CallSuccess(t *testing.T) {
	pluginPath := buildTestPlugin(t)

	client := infra.NewRPCClient(pluginPath, "echo")
	ctx := context.Background()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Stop()

	// Make RPC call
	params := map[string]string{"message": "hello"}
	result, err := client.Call(context.Background(), "echo", params)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}

	// Verify result
	var response map[string]interface{}
	if err := json.Unmarshal(result, &response); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if response["message"] != "hello" {
		t.Errorf("expected message 'hello', got %v", response["message"])
	}
}

// TestRPCClient_CallError tests RPC call that returns error.
func TestRPCClient_CallError(t *testing.T) {
	pluginPath := buildTestPlugin(t)

	client := infra.NewRPCClient(pluginPath, "error")
	ctx := context.Background()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Stop()

	// Make RPC call that should error
	_, err := client.Call(context.Background(), "fail", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "rpc error -32603: intentional test error" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestRPCClient_CallTimeout tests RPC call timeout.
func TestRPCClient_CallTimeout(t *testing.T) {
	pluginPath := buildTestPlugin(t)

	client := infra.NewRPCClient(pluginPath, "slow")
	ctx := context.Background()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Stop()

	// Make RPC call with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Call(ctx, "slow_method", nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}
}

// TestRPCClient_MultipleRequests tests concurrent RPC calls.
func TestRPCClient_MultipleRequests(t *testing.T) {
	pluginPath := buildTestPlugin(t)

	client := infra.NewRPCClient(pluginPath, "echo")
	ctx := context.Background()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Stop()

	// Make multiple concurrent calls
	results := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			params := map[string]interface{}{"id": id}
			result, err := client.Call(context.Background(), "echo", params)
			if err != nil {
				results <- err
				return
			}

			var response map[string]interface{}
			if err := json.Unmarshal(result, &response); err != nil {
				results <- err
				return
			}

			if int(response["id"].(float64)) != id {
				results <- fmt.Errorf("expected id %d, got %v", id, response["id"])
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all results
	for i := 0; i < 10; i++ {
		if err := <-results; err != nil {
			t.Errorf("concurrent call failed: %v", err)
		}
	}
}

// TestRPCClient_ProcessCrash tests handling of subprocess crash.
func TestRPCClient_ProcessCrash(t *testing.T) {
	pluginPath := buildTestPlugin(t)

	client := infra.NewRPCClient(pluginPath, "crash")
	ctx := context.Background()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Stop()

	// Wait a moment for process to start
	time.Sleep(50 * time.Millisecond)

	// Make a call that will crash the process
	_, err := client.Call(context.Background(), "crash", nil)
	if err == nil {
		t.Fatal("expected error after crash, got nil")
	}

	// Subsequent calls should also fail
	_, err = client.Call(context.Background(), "echo", nil)
	if err == nil {
		t.Fatal("expected error on subsequent call, got nil")
	}
}

// TestRPCClient_EventStreaming tests event emission from plugin.
func TestRPCClient_EventStreaming(t *testing.T) {
	pluginPath := buildTestPlugin(t)

	client := infra.NewRPCClient(pluginPath, "events")
	ctx := context.Background()

	if err := client.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Stop()

	// Setup event channel
	eventChan := make(chan pluginsdk.Event, 10)
	client.SetEventChannel(eventChan)

	// Trigger event emission
	params := map[string]int{"count": 3}
	_, err := client.Call(context.Background(), "emit_events", params)
	if err != nil {
		t.Fatalf("call failed: %v", err)
	}

	// Receive events
	timeout := time.After(2 * time.Second)
	receivedCount := 0
	for receivedCount < 3 {
		select {
		case event := <-eventChan:
			if event.Type != "test.event" {
				t.Errorf("unexpected event type: %s", event.Type)
			}
			if event.Source != "test-plugin" {
				t.Errorf("unexpected event source: %s", event.Source)
			}
			receivedCount++
		case <-timeout:
			t.Fatalf("timeout waiting for events (received %d/3)", receivedCount)
		}
	}
}

// buildTestPlugin compiles the test plugin executable.
func buildTestPlugin(t *testing.T) string {
	t.Helper()

	// Create temporary directory for plugin
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-plugin")

	// Write test plugin source
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

func main() {
	if len(os.Args) < 2 {
		return
	}

	mode := os.Args[1]

	switch mode {
	case "echo":
		echoMode()
	case "error":
		errorMode()
	case "slow":
		slowMode()
	case "crash":
		crashMode()
	case "events":
		eventsMode()
	}
}

func echoMode() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}

		// Echo params back as result
		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  req.Params,
		}

		data, _ := json.Marshal(resp)
		fmt.Fprintf(os.Stdout, "%s\n", string(data))
	}
}

func errorMode() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}

		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32603,
				Message: "intentional test error",
			},
		}

		data, _ := json.Marshal(resp)
		fmt.Fprintf(os.Stdout, "%s\n", string(data))
	}
}

func slowMode() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Delay to trigger timeout
		time.Sleep(2 * time.Second)

		var req Request
		json.Unmarshal(scanner.Bytes(), &req)

		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage("{}"),
		}

		data, _ := json.Marshal(resp)
		fmt.Fprintf(os.Stdout, "%s\n", string(data))
	}
}

func crashMode() {
	// Immediately exit
	os.Exit(1)
}

func eventsMode() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			continue
		}

		// Parse params to get event count
		var params map[string]int
		json.Unmarshal(req.Params, &params)
		count := params["count"]

		// Emit events
		for i := 0; i < count; i++ {
			event := Event{
				Event:     "event",
				Type:      "test.event",
				Source:    "test-plugin",
				Timestamp: time.Now().Format(time.RFC3339),
				Payload: map[string]interface{}{
					"index": i,
				},
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(os.Stdout, "%s\n", string(data))
		}

		// Send response
		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage("{}"),
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

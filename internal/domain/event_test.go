package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		sessionID string
		payload   interface{}
		content   string
	}{
		{
			name:      "creates chat started event",
			eventType: "claude.chat.started",
			sessionID: "test-session-1",
			payload:   map[string]string{"message": "Hello", "context": "greeting"},
			content:   "Hello greeting",
		},
		{
			name:      "creates tool invoked event",
			eventType: "claude.tool.invoked",
			sessionID: "test-session-2",
			payload:   map[string]interface{}{"tool": "Read", "parameters": map[string]string{"file": "test.go"}},
			content:   "Reading test.go",
		},
		{
			name:      "creates event with nil payload",
			eventType: "claude.file.read",
			sessionID: "test-session-3",
			payload:   nil,
			content:   "file read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := domain.NewEvent(tt.eventType, tt.sessionID, tt.payload, tt.content)

			// Verify required fields are set
			if event.ID == "" {
				t.Error("Expected ID to be generated, got empty string")
			}
			if event.SessionID != tt.sessionID {
				t.Errorf("Expected SessionID = %q, got %q", tt.sessionID, event.SessionID)
			}
			if event.Type != tt.eventType {
				t.Errorf("Expected Type = %q, got %q", tt.eventType, event.Type)
			}
			if event.Content != tt.content {
				t.Errorf("Expected Content = %q, got %q", tt.content, event.Content)
			}

			// Verify timestamp is recent (within last second)
			if time.Since(event.Timestamp) > time.Second {
				t.Errorf("Expected recent timestamp, got %v", event.Timestamp)
			}
		})
	}
}

func TestEvent_MarshalPayload(t *testing.T) {
	tests := []struct {
		name       string
		payload    interface{}
		wantErr    bool
		validateFn func([]byte) error
	}{
		{
			name: "marshals chat payload",
			payload: map[string]string{
				"message": "test message",
				"context": "test context",
			},
			wantErr: false,
			validateFn: func(data []byte) error {
				var p map[string]string
				if err := json.Unmarshal(data, &p); err != nil {
					return err
				}
				if p["message"] != "test message" {
					t.Errorf("Expected message = %q, got %q", "test message", p["message"])
				}
				return nil
			},
		},
		{
			name: "marshals tool payload",
			payload: map[string]interface{}{
				"tool":       "Bash",
				"parameters": map[string]string{"command": "ls"},
				"duration_ms": 100,
			},
			wantErr: false,
			validateFn: func(data []byte) error {
				var p map[string]interface{}
				if err := json.Unmarshal(data, &p); err != nil {
					return err
				}
				if p["tool"] != "Bash" {
					t.Errorf("Expected tool = %q, got %q", "Bash", p["tool"])
				}
				return nil
			},
		},
		{
			name: "marshals file payload",
			payload: map[string]interface{}{
				"file_path":   "/test/path.go",
				"changes":     "added function",
				"duration_ms": 50,
			},
			wantErr: false,
			validateFn: func(data []byte) error {
				var p map[string]interface{}
				if err := json.Unmarshal(data, &p); err != nil {
					return err
				}
				if p["file_path"] != "/test/path.go" {
					t.Errorf("Expected file_path = %q, got %q", "/test/path.go", p["file_path"])
				}
				return nil
			},
		},
		{
			name: "marshals error payload",
			payload: map[string]interface{}{
				"error":       "test error",
				"stack_trace": "line 1\nline 2",
				"context":     "during test",
			},
			wantErr: false,
			validateFn: func(data []byte) error {
				var p map[string]interface{}
				if err := json.Unmarshal(data, &p); err != nil {
					return err
				}
				if p["error"] != "test error" {
					t.Errorf("Expected error = %q, got %q", "test error", p["error"])
				}
				return nil
			},
		},
		{
			name:    "marshals nil payload",
			payload: nil,
			wantErr: false,
			validateFn: func(data []byte) error {
				if string(data) != "null" {
					t.Errorf("Expected JSON null, got %q", string(data))
				}
				return nil
			},
		},
		{
			name: "marshals complex nested payload",
			payload: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
					"arr": []int{1, 2, 3},
				},
			},
			wantErr: false,
			validateFn: func(data []byte) error {
				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					return err
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := domain.NewEvent("claude.chat.started", "test-session", tt.payload, "content")

			data, err := event.MarshalPayload()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFn != nil {
				if err := tt.validateFn(data); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

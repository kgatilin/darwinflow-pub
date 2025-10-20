package claude_code_test

import (
	"context"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

// Minimal test focusing on what can be tested without complex mocking
// Full integration tests would require a test database

func TestNewClaudeCodePlugin(t *testing.T) {
	// This test verifies the constructor works
	// We use nil services since we're only testing construction
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &app.NoOpLogger{}, nil, nil, "")

	if plugin == nil {
		t.Fatal("NewClaudeCodePlugin returned nil")
	}
}

func TestGetInfo(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &app.NoOpLogger{}, nil, nil, "")

	info := plugin.GetInfo()

	if info.Name != "claude-code" {
		t.Errorf("Expected Name 'claude-code', got %q", info.Name)
	}
	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if !info.IsCore {
		t.Error("Expected IsCore to be true for claude-code plugin")
	}
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestGetEntityTypes(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &app.NoOpLogger{}, nil, nil, "")

	entityTypes := plugin.GetEntityTypes()

	if len(entityTypes) != 1 {
		t.Fatalf("Expected 1 entity type, got %d", len(entityTypes))
	}

	sessionType := entityTypes[0]
	if sessionType.Type != "session" {
		t.Errorf("Expected Type 'session', got %q", sessionType.Type)
	}
	if sessionType.DisplayName == "" {
		t.Error("DisplayName should not be empty")
	}
	if sessionType.DisplayNamePlural == "" {
		t.Error("DisplayNamePlural should not be empty")
	}
	if len(sessionType.Capabilities) == 0 {
		t.Error("Should have capabilities defined")
	}

	// Verify expected capabilities
	expectedCaps := map[string]bool{
		"IExtensible": true,
		"IHasContext": true,
		"ITrackable":  true,
	}

	for _, cap := range sessionType.Capabilities {
		if !expectedCaps[cap] {
			t.Errorf("Unexpected capability: %s", cap)
		}
		delete(expectedCaps, cap)
	}

	if len(expectedCaps) > 0 {
		t.Errorf("Missing expected capabilities: %v", expectedCaps)
	}

	if sessionType.Icon == "" {
		t.Error("Icon should not be empty")
	}
}

func TestUpdateEntity_ReadOnly(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &app.NoOpLogger{}, nil, nil, "")

	ctx := context.Background()
	_, err := plugin.UpdateEntity(ctx, "session-1", map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for read-only update, got nil")
	}
	if err.Error() != "entity is read-only" {
		t.Errorf("Expected 'entity is read-only' error, got: %v", err)
	}
}

func TestGetCommands(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &app.NoOpLogger{}, nil, nil, "")

	commands := plugin.GetCommands()

	// Verify we get exactly 5 commands (including session-summary)
	if len(commands) != 5 {
		t.Fatalf("Expected 5 commands, got %d", len(commands))
	}

	// Verify expected command names
	expectedCommands := map[string]bool{
		"init":              false,
		"log":               false,
		"auto-summary":      false,
		"auto-summary-exec": false,
		"session-summary":   false,
	}

	for _, cmd := range commands {
		name := cmd.GetName()
		if _, exists := expectedCommands[name]; !exists {
			t.Errorf("Unexpected command: %s", name)
		}
		expectedCommands[name] = true

		// Verify each command has required metadata
		if cmd.GetDescription() == "" {
			t.Errorf("Command %s has empty description", name)
		}
		if cmd.GetUsage() == "" {
			t.Errorf("Command %s has empty usage", name)
		}
	}

	// Verify all expected commands were found
	for name, found := range expectedCommands {
		if !found {
			t.Errorf("Expected command %s not found", name)
		}
	}
}

func TestCommandProvider_Interface(t *testing.T) {
	// Verify that ClaudeCodePlugin implements SDK ICommandProvider
	var _ domain.ICommandProvider = (*claude_code.ClaudeCodePlugin)(nil)
}

// Note: Full Query, GetEntity, and tool execution tests would require:
// - Test database setup
// - Mock/test repositories
// - Event/analysis data fixtures
// These are better suited for integration tests rather than unit tests.
// The session_entity_test.go file provides comprehensive coverage of the
// SessionEntity logic which is the core functionality.

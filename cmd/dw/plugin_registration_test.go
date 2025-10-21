package main_test

import (
	"path/filepath"
	"testing"

	main "github.com/kgatilin/darwinflow-pub/cmd/dw"
)

// TestRegisterBuiltInPlugins_Integration tests plugin registration with real services
// This is an integration test that verifies plugins can be registered successfully
func TestRegisterBuiltInPlugins_Integration(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := ""

	// Initialize app - this will register built-in plugins
	services, err := main.InitializeApp(dbPath, configPath, false)
	if err != nil {
		t.Fatalf("InitializeApp() failed: %v", err)
	}

	// Verify plugin was registered
	if services.PluginRegistry == nil {
		t.Fatal("PluginRegistry is nil")
	}

	infos := services.PluginRegistry.GetPluginInfos()
	if len(infos) == 0 {
		t.Error("Expected at least one plugin to be registered")
	}

	// Verify claude-code plugin is registered
	foundClaudeCode := false
	for _, info := range infos {
		if info.Name == "claude-code" {
			foundClaudeCode = true
			if info.Version == "" {
				t.Error("Plugin version should not be empty")
			}
			break
		}
	}

	if !foundClaudeCode {
		t.Error("claude-code plugin should be registered")
	}
}

func TestRegisterBuiltInPlugins_MultipleCallsIdempotent(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := ""

	// Initialize app multiple times
	for i := 0; i < 3; i++ {
		services, err := main.InitializeApp(dbPath, configPath, false)
		if err != nil {
			t.Fatalf("InitializeApp() call %d failed: %v", i+1, err)
		}

		// Each time, plugins should be registered
		infos := services.PluginRegistry.GetPluginInfos()
		if len(infos) == 0 {
			t.Errorf("Expected plugins on call %d", i+1)
		}
	}
}

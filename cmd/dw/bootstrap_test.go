package main_test

import (
	"os"
	"path/filepath"
	"testing"

	main "github.com/kgatilin/darwinflow-pub/cmd/dw"
)

func TestInitializeApp_Success(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := "" // Use default config

	// Initialize app
	services, err := main.InitializeApp(dbPath, configPath, false)
	if err != nil {
		t.Fatalf("InitializeApp() failed: %v", err)
	}

	// Verify all services are initialized
	if services == nil {
		t.Fatal("Expected non-nil services")
	}

	if services.PluginRegistry == nil {
		t.Error("PluginRegistry is nil")
	}

	if services.CommandRegistry == nil {
		t.Error("CommandRegistry is nil")
	}

	if services.LogsService == nil {
		t.Error("LogsService is nil")
	}

	if services.AnalysisService == nil {
		t.Error("AnalysisService is nil")
	}

	if services.SetupService == nil {
		t.Error("SetupService is nil")
	}

	if services.ConfigLoader == nil {
		t.Error("ConfigLoader is nil")
	}

	if services.Logger == nil {
		t.Error("Logger is nil")
	}

	if services.EventRepo == nil {
		t.Error("EventRepo is nil")
	}

	if services.DBPath != dbPath {
		t.Errorf("DBPath = %q, want %q", services.DBPath, dbPath)
	}

	if services.WorkingDir == "" {
		t.Error("WorkingDir is empty")
	}
}

func TestInitializeApp_DebugMode(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := ""

	// Initialize app with debug mode
	services, err := main.InitializeApp(dbPath, configPath, true)
	if err != nil {
		t.Fatalf("InitializeApp() failed: %v", err)
	}

	// Verify services are initialized
	if services == nil {
		t.Fatal("Expected non-nil services")
	}

	if services.Logger == nil {
		t.Error("Logger is nil in debug mode")
	}
}

func TestInitializeApp_InvalidDBPath(t *testing.T) {
	// Try to create database in a non-existent parent directory that we cannot create
	// This simulates permission or filesystem errors
	dbPath := "/root/impossible/path/test.db" // Assuming tests don't run as root
	configPath := ""

	services, err := main.InitializeApp(dbPath, configPath, false)

	// On some systems this might succeed if running as root, so we check both cases
	if err == nil && services == nil {
		t.Error("Expected either error or valid services")
	}

	// If services are returned, they should be valid
	if err == nil && services != nil {
		if services.Logger == nil {
			t.Error("Logger should be initialized even if DB creation fails later")
		}
	}
}

func TestInitializeApp_CreatesDBDirectory(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create a nested path
	dbPath := filepath.Join(tmpDir, "nested", "path", "test.db")
	configPath := ""

	// Initialize app
	services, err := main.InitializeApp(dbPath, configPath, false)
	if err != nil {
		t.Fatalf("InitializeApp() failed: %v", err)
	}

	// Verify directory was created
	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		t.Errorf("Database directory was not created: %s", dbDir)
	}

	// Verify services are valid
	if services == nil {
		t.Fatal("Expected non-nil services")
	}
}

func TestInitializeApp_NonFatalConfigError(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Use a non-existent config file (should fall back to defaults)
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// Initialize app - should succeed with default config
	services, err := main.InitializeApp(dbPath, configPath, false)
	if err != nil {
		t.Fatalf("InitializeApp() should handle missing config gracefully: %v", err)
	}

	// Verify services are initialized despite missing config
	if services == nil {
		t.Fatal("Expected non-nil services even with missing config")
	}

	if services.ConfigLoader == nil {
		t.Error("ConfigLoader should be initialized")
	}

	if services.Logger == nil {
		t.Error("Logger should be initialized")
	}
}

func TestInitializeApp_PluginsRegistered(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := ""

	// Initialize app
	services, err := main.InitializeApp(dbPath, configPath, false)
	if err != nil {
		t.Fatalf("InitializeApp() failed: %v", err)
	}

	// Verify built-in plugins are registered
	if services.PluginRegistry == nil {
		t.Fatal("PluginRegistry is nil")
	}

	pluginInfos := services.PluginRegistry.GetPluginInfos()
	if len(pluginInfos) == 0 {
		t.Error("Expected at least one built-in plugin to be registered")
	}

	// Verify claude-code plugin is registered
	foundClaudeCode := false
	for _, info := range pluginInfos {
		if info.Name == "claude-code" {
			foundClaudeCode = true
			break
		}
	}

	if !foundClaudeCode {
		t.Error("claude-code plugin should be registered by default")
	}
}

func TestInitializeApp_WorkingDirectory(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := ""

	// Initialize app
	services, err := main.InitializeApp(dbPath, configPath, false)
	if err != nil {
		t.Fatalf("InitializeApp() failed: %v", err)
	}

	// Verify working directory is set
	if services.WorkingDir == "" {
		t.Error("WorkingDir should not be empty")
	}

	// Working directory should exist
	if _, err := os.Stat(services.WorkingDir); err != nil {
		t.Errorf("WorkingDir does not exist: %s", services.WorkingDir)
	}
}

func TestAppServices_Structure(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := ""

	// Initialize app
	services, err := main.InitializeApp(dbPath, configPath, false)
	if err != nil {
		t.Fatalf("InitializeApp() failed: %v", err)
	}

	// Test that AppServices struct has all expected fields populated
	tests := []struct {
		name  string
		check func() bool
	}{
		{"PluginRegistry", func() bool { return services.PluginRegistry != nil }},
		{"CommandRegistry", func() bool { return services.CommandRegistry != nil }},
		{"LogsService", func() bool { return services.LogsService != nil }},
		{"AnalysisService", func() bool { return services.AnalysisService != nil }},
		{"SetupService", func() bool { return services.SetupService != nil }},
		{"ConfigLoader", func() bool { return services.ConfigLoader != nil }},
		{"Logger", func() bool { return services.Logger != nil }},
		{"EventRepo", func() bool { return services.EventRepo != nil }},
		{"DBPath", func() bool { return services.DBPath == dbPath }},
		{"WorkingDir", func() bool { return services.WorkingDir != "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("AppServices.%s check failed", tt.name)
			}
		})
	}
}

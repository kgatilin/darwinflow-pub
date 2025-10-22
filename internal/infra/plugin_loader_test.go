package infra_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

// TestPluginLoader_LoadFromConfig_Success tests loading a valid configuration.
func TestPluginLoader_LoadFromConfig_Success(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Create a mock executable
	pluginPath := filepath.Join(tempDir, "test-plugin")
	createMockExecutable(t, pluginPath)

	// Create valid plugins.yaml
	configPath := filepath.Join(tempDir, "plugins.yaml")
	configContent := `
plugins:
  test-plugin:
    command: test-plugin
    args: ["--verbose"]
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create loader
	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Load plugins
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %v", err)
	}

	// Verify results
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_MissingFile tests behavior when config file doesn't exist.
func TestPluginLoader_LoadFromConfig_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.yaml")

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should return empty list without error
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error for missing file, got: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins for missing file, got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_InvalidYAML tests handling of malformed YAML.
func TestPluginLoader_LoadFromConfig_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Write invalid YAML
	invalidYAML := `
plugins:
  test-plugin:
    command: "foo"
    invalid: [unclosed
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should return error
	_, err := loader.LoadFromConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestPluginLoader_LoadFromConfig_InvalidCommand tests handling of non-existent command.
func TestPluginLoader_LoadFromConfig_InvalidCommand(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create config with non-existent command
	configContent := `
plugins:
  bad-plugin:
    command: /nonexistent/path/to/plugin
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should skip plugin and return empty list (with warning logged)
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error (should skip invalid plugin), got: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins (invalid command should be skipped), got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_DisabledPlugin tests that disabled plugins are skipped.
func TestPluginLoader_LoadFromConfig_DisabledPlugin(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create mock executable
	pluginPath := filepath.Join(tempDir, "test-plugin")
	createMockExecutable(t, pluginPath)

	// Create config with disabled plugin
	configContent := `
plugins:
  test-plugin:
    command: test-plugin
    enabled: false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should skip disabled plugin
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins (disabled should be skipped), got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_RelativePath tests resolution of relative paths.
func TestPluginLoader_LoadFromConfig_RelativePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory for plugin
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}

	// Create mock executable in subdirectory
	pluginPath := filepath.Join(binDir, "my-plugin")
	createMockExecutable(t, pluginPath)

	// Create config with relative path
	configPath := filepath.Join(tempDir, "plugins.yaml")
	configContent := `
plugins:
  my-plugin:
    command: bin/my-plugin
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should resolve relative path and load plugin
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin (relative path should resolve), got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_MultiplePlugins tests loading multiple plugins.
func TestPluginLoader_LoadFromConfig_MultiplePlugins(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple mock executables
	plugin1Path := filepath.Join(tempDir, "plugin1")
	plugin2Path := filepath.Join(tempDir, "plugin2")
	createMockExecutable(t, plugin1Path)
	createMockExecutable(t, plugin2Path)

	// Create config with multiple plugins
	configPath := filepath.Join(tempDir, "plugins.yaml")
	configContent := `
plugins:
  plugin1:
    command: plugin1
    enabled: true
  plugin2:
    command: plugin2
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should load both plugins
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %v", err)
	}

	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_MixedValidInvalid tests loading with mix of valid and invalid plugins.
func TestPluginLoader_LoadFromConfig_MixedValidInvalid(t *testing.T) {
	tempDir := t.TempDir()

	// Create one valid executable
	validPluginPath := filepath.Join(tempDir, "valid-plugin")
	createMockExecutable(t, validPluginPath)

	// Create config with valid, invalid, and disabled plugins
	configPath := filepath.Join(tempDir, "plugins.yaml")
	configContent := `
plugins:
  valid-plugin:
    command: valid-plugin
    enabled: true
  invalid-plugin:
    command: /nonexistent/plugin
    enabled: true
  disabled-plugin:
    command: valid-plugin
    enabled: false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should load only the valid, enabled plugin
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error (should skip invalid/disabled), got: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin (only valid enabled one), got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_EmptyCommand tests handling of missing command field.
func TestPluginLoader_LoadFromConfig_EmptyCommand(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create config with empty command
	configContent := `
plugins:
  bad-plugin:
    command: ""
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should skip plugin with empty command
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins (empty command should be skipped), got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_DefaultEnabled tests that enabled defaults to true.
func TestPluginLoader_LoadFromConfig_DefaultEnabled(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create mock executable
	pluginPath := filepath.Join(tempDir, "test-plugin")
	createMockExecutable(t, pluginPath)

	// Create config without explicit enabled field
	configContent := `
plugins:
  test-plugin:
    command: test-plugin
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should load plugin (enabled defaults to true)
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin (enabled should default to true), got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_WithArgs tests loading plugin with arguments.
func TestPluginLoader_LoadFromConfig_WithArgs(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create mock executable
	pluginPath := filepath.Join(tempDir, "test-plugin")
	createMockExecutable(t, pluginPath)

	// Create config with args
	configContent := `
plugins:
  test-plugin:
    command: test-plugin
    args: ["--verbose", "--config", "foo.yaml"]
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should load plugin with args
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_WithEnv tests loading plugin with environment variables.
func TestPluginLoader_LoadFromConfig_WithEnv(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create mock executable
	pluginPath := filepath.Join(tempDir, "test-plugin")
	createMockExecutable(t, pluginPath)

	// Create config with env vars
	configContent := `
plugins:
  test-plugin:
    command: test-plugin
    env:
      DEBUG: "true"
      API_KEY: "secret123"
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should load plugin (env vars noted for future support)
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_AbsolutePath tests loading with absolute path.
func TestPluginLoader_LoadFromConfig_AbsolutePath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create mock executable with absolute path
	pluginPath := filepath.Join(tempDir, "abs-plugin")
	createMockExecutable(t, pluginPath)

	// Create config with absolute path
	configContent := "plugins:\n  abs-plugin:\n    command: " + pluginPath + "\n    enabled: true\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should load plugin with absolute path
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Fatalf("LoadFromConfig failed: %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin (absolute path), got %d", len(plugins))
	}
}

// TestPluginLoader_LoadFromConfig_NonExecutableFile tests handling of non-executable file.
func TestPluginLoader_LoadFromConfig_NonExecutableFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugins.yaml")

	// Create non-executable file (no execute permission)
	pluginPath := filepath.Join(tempDir, "non-exec")
	if err := os.WriteFile(pluginPath, []byte("#!/bin/sh\necho test"), 0644); err != nil {
		t.Fatalf("Failed to create non-executable file: %v", err)
	}

	// Create config pointing to non-executable file
	configContent := `
plugins:
  non-exec:
    command: non-exec
    enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewPluginLoader(logger)

	// Should skip non-executable file
	plugins, err := loader.LoadFromConfig(configPath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins (non-executable should be skipped), got %d", len(plugins))
	}
}

// Helper function to create a mock executable file for testing
func createMockExecutable(t *testing.T, path string) {
	t.Helper()

	// Write minimal executable script
	content := "#!/bin/sh\necho 'mock plugin'\n"
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create mock executable: %v", err)
	}
}

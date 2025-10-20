package infra_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestNewConfigLoader(t *testing.T) {
	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	if loader == nil {
		t.Fatal("NewConfigLoader returned nil")
	}
}

func TestConfigLoader_LoadConfig_DefaultsWhenNoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp dir where no config file exists
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	config, err := loader.LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Should return default config
	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	// Verify some default values
	if config.Analysis.TokenLimit != 100000 {
		t.Errorf("Expected default TokenLimit 100000, got %d", config.Analysis.TokenLimit)
	}
	if config.Analysis.Model != "sonnet" {
		t.Errorf("Expected default Model 'sonnet', got %q", config.Analysis.Model)
	}
}

func TestConfigLoader_LoadConfig_FromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test config file
	configPath := filepath.Join(tmpDir, ".darwinflow.yaml")
	configContent := `
analysis:
  token_limit: 50000
  model: "opus"
  parallel_limit: 5
  enabled_prompts: ["test_prompt"]
prompts:
  test_prompt: "Test prompt content"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to temp dir
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelDebug)
	loader := infra.NewConfigLoader(logger)

	config, err := loader.LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify custom values were loaded
	if config.Analysis.TokenLimit != 50000 {
		t.Errorf("Expected TokenLimit 50000, got %d", config.Analysis.TokenLimit)
	}
	if config.Analysis.Model != "opus" {
		t.Errorf("Expected Model 'opus', got %q", config.Analysis.Model)
	}
	if config.Analysis.ParallelLimit != 5 {
		t.Errorf("Expected ParallelLimit 5, got %d", config.Analysis.ParallelLimit)
	}

	// Verify custom prompt exists
	if _, ok := config.Prompts["test_prompt"]; !ok {
		t.Error("Expected test_prompt to exist in config")
	}

	// Verify default prompts were added
	if _, ok := config.Prompts["session_summary"]; !ok {
		t.Error("Expected default session_summary prompt to be present")
	}
}

func TestConfigLoader_LoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid YAML file
	configPath := filepath.Join(tmpDir, ".darwinflow.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: [not: closed"), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	_, err = loader.LoadConfig("")
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestConfigLoader_LoadConfig_ExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config in a specific path
	configPath := filepath.Join(tmpDir, "custom-config.yaml")
	configContent := `
analysis:
  token_limit: 75000
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	config, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Analysis.TokenLimit != 75000 {
		t.Errorf("Expected TokenLimit 75000, got %d", config.Analysis.TokenLimit)
	}
}

func TestConfigLoader_LoadConfig_AppliesDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create partial config
	configPath := filepath.Join(tmpDir, ".darwinflow.yaml")
	configContent := `
analysis:
  token_limit: 60000
  # model not specified, should use default
prompts:
  custom: "Custom prompt"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	config, err := loader.LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Custom value
	if config.Analysis.TokenLimit != 60000 {
		t.Errorf("Expected TokenLimit 60000, got %d", config.Analysis.TokenLimit)
	}

	// Default value
	if config.Analysis.Model != "sonnet" {
		t.Errorf("Expected default Model 'sonnet', got %q", config.Analysis.Model)
	}

	// Default parallel limit
	if config.Analysis.ParallelLimit != 3 {
		t.Errorf("Expected default ParallelLimit 3, got %d", config.Analysis.ParallelLimit)
	}
}

func TestConfigLoader_LoadConfig_WithNilLogger(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create loader with nil logger (should not panic)
	loader := infra.NewConfigLoader(nil)

	config, err := loader.LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig with nil logger failed: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}
}

func TestValidateModelAlias(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "sonnet alias", model: "sonnet", want: true},
		{name: "opus alias", model: "opus", want: true},
		{name: "haiku alias", model: "haiku", want: true},
		{name: "full model name", model: "claude-sonnet-4-5-20250929", want: true},
		{name: "invalid model", model: "invalid", want: false},
		{name: "empty model", model: "", want: true}, // empty is valid (uses default)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := infra.ValidateModelAlias(tt.model)
			if got != tt.want {
				t.Errorf("ValidateModelAlias(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestConfigLoader_SaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	// Create a config
	config := domain.DefaultConfig()
	config.Analysis.TokenLimit = 80000

	// Save it (note: signature is SaveConfig(config, path))
	savedPath, err := loader.SaveConfig(config, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	if savedPath != configPath {
		t.Errorf("SaveConfig returned path %s, expected %s", savedPath, configPath)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load it back and verify
	loadedConfig, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedConfig.Analysis.TokenLimit != 80000 {
		t.Errorf("Expected TokenLimit 80000 after round-trip, got %d", loadedConfig.Analysis.TokenLimit)
	}
}

func TestConfigLoader_GetPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".darwinflow.yaml")

	// Create config with custom prompt
	configContent := `
prompts:
  custom_prompt: "This is a custom prompt"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	// Load config
	config, err := loader.LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Get existing prompt
	prompt, exists := loader.GetPrompt(config, "custom_prompt")
	if !exists {
		t.Fatal("GetPrompt: custom_prompt should exist")
	}
	if prompt != "This is a custom prompt" {
		t.Errorf("Expected custom prompt, got %q", prompt)
	}

	// Get default prompt
	sessionSummary, exists := loader.GetPrompt(config, "session_summary")
	if !exists {
		t.Fatal("GetPrompt: session_summary should exist")
	}
	if sessionSummary == "" {
		t.Error("Expected non-empty session_summary prompt")
	}

	// Get non-existent prompt
	_, exists = loader.GetPrompt(config, "nonexistent")
	if exists {
		t.Error("Expected nonexistent prompt to not exist")
	}
}

func TestConfigLoader_InitializeDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".darwinflow.yaml")

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	// Initialize default config
	createdPath, err := loader.InitializeDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("InitializeDefaultConfig failed: %v", err)
	}
	if createdPath != configPath {
		t.Errorf("InitializeDefaultConfig returned path %s, expected %s", createdPath, configPath)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify it's valid
	config, err := loader.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig after init failed: %v", err)
	}

	// Should have default values
	if config.Analysis.TokenLimit != 100000 {
		t.Error("Initialized config doesn't have default values")
	}
}

func TestConfigLoader_InitializeDefaultConfig_InCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	logger := infra.NewDefaultLogger()
	loader := infra.NewConfigLoader(logger)

	// Initialize with empty path (should use current dir)
	createdPath, err := loader.InitializeDefaultConfig("")
	if err != nil {
		t.Fatalf("InitializeDefaultConfig failed: %v", err)
	}

	// Verify file exists in current dir
	configPath := filepath.Join(tmpDir, ".darwinflow.yaml")
	if createdPath != configPath {
		t.Errorf("InitializeDefaultConfig returned path %s, expected %s", createdPath, configPath)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created in current directory")
	}
}

func TestDefaultConfigFileName(t *testing.T) {
	// Verify the constant is defined
	if infra.DefaultConfigFileName == "" {
		t.Error("DefaultConfigFileName should not be empty")
	}

	expected := ".darwinflow.yaml"
	if infra.DefaultConfigFileName != expected {
		t.Errorf("Expected DefaultConfigFileName = %s, got %s", expected, infra.DefaultConfigFileName)
	}
}

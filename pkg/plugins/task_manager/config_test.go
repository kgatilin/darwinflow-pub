package task_manager_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

func TestDefaultConfig(t *testing.T) {
	cfg := task_manager.DefaultConfig()

	if cfg.ADR.Required {
		t.Error("ADR.Required should be false by default")
	}
	if cfg.ADR.EnforceOnTaskCompletion {
		t.Error("ADR.EnforceOnTaskCompletion should be false by default")
	}
}

func TestLoadConfigNoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Should return defaults when no config file exists
	cfg, err := task_manager.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ADR.Required {
		t.Error("ADR.Required should default to false")
	}
	if cfg.ADR.EnforceOnTaskCompletion {
		t.Error("ADR.EnforceOnTaskCompletion should default to false")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".darwinflow")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `
task_manager:
  adr:
    required: false
    enforce_on_task_completion: false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := task_manager.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ADR.Required {
		t.Error("ADR.Required should be false from config")
	}
	if cfg.ADR.EnforceOnTaskCompletion {
		t.Error("ADR.EnforceOnTaskCompletion should be false from config")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".darwinflow")
	configPath := filepath.Join(configDir, "config.yaml")

	cfg := &task_manager.Config{
		ADR: task_manager.ADRConfig{
			Required:                false,
			EnforceOnTaskCompletion: true,
		},
	}

	if err := task_manager.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Load it back and verify
	loadedCfg, err := task_manager.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loadedCfg.ADR.Required {
		t.Error("ADR.Required should be false")
	}
	if !loadedCfg.ADR.EnforceOnTaskCompletion {
		t.Error("ADR.EnforceOnTaskCompletion should be true")
	}
}

func TestLoadConfigPartialOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".darwinflow")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Config that only sets one value
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `
task_manager:
  adr:
    required: false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := task_manager.LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should override required but keep other defaults
	if cfg.ADR.Required {
		t.Error("ADR.Required should be false from config")
	}
	if cfg.ADR.EnforceOnTaskCompletion {
		t.Error("ADR.EnforceOnTaskCompletion should stay false (default)")
	}
}

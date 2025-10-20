package claude_code_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

func TestNewHookConfigManager(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	mgr, err := claude_code.NewHookConfigManager()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if mgr == nil {
		t.Fatal("Expected manager, got nil")
	}

	// Should default to .claude/settings.json
	if mgr.GetSettingsPath() != ".claude/settings.json" {
		t.Errorf("Expected path '.claude/settings.json', got %q", mgr.GetSettingsPath())
	}
}

func TestReadSettings_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	mgr, _ := claude_code.NewHookConfigManager()
	settings, err := mgr.ReadSettings()

	if err != nil {
		t.Errorf("Expected no error for missing file, got %v", err)
	}

	if settings == nil {
		t.Fatal("Expected settings, got nil")
	}

	if len(settings.Hooks) != 0 {
		t.Errorf("Expected empty hooks, got %d", len(settings.Hooks))
	}
}

func TestWriteSettings(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	mgr, _ := claude_code.NewHookConfigManager()

	// Create settings with hooks
	settings := &claude_code.ClaudeSettings{
		Hooks: map[string][]claude_code.HookMatcher{
			"TestEvent": {
				{
					Matcher: "*",
					Hooks: []claude_code.HookAction{
						{
							Type:    "command",
							Command: "dw test",
							Timeout: 5,
						},
					},
				},
			},
		},
		Other: make(map[string]interface{}),
	}

	// Write settings
	err = mgr.WriteSettings(settings)
	if err != nil {
		t.Fatalf("Failed to write settings: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(mgr.GetSettingsPath()); os.IsNotExist(err) {
		t.Fatal("Settings file was not created")
	}

	// Read back and verify
	readSettings, err := mgr.ReadSettings()
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	if len(readSettings.Hooks) != 1 {
		t.Errorf("Expected 1 hook event, got %d", len(readSettings.Hooks))
	}

	if _, ok := readSettings.Hooks["TestEvent"]; !ok {
		t.Fatal("TestEvent hook not found")
	}
}

func TestDefaultDarwinFlowConfig(t *testing.T) {
	config := claude_code.DefaultDarwinFlowConfig()

	expectedEvents := []string{"PreToolUse", "UserPromptSubmit", "SessionEnd"}
	for _, event := range expectedEvents {
		if _, ok := config.Hooks[event]; !ok {
			t.Errorf("Expected hook event %q not found", event)
		}
	}

	// Verify hook structure
	if len(config.Hooks["PreToolUse"]) == 0 {
		t.Fatal("PreToolUse hooks should not be empty")
	}

	preToolHook := config.Hooks["PreToolUse"][0]
	if preToolHook.Matcher != "*" {
		t.Errorf("Expected matcher '*', got %q", preToolHook.Matcher)
	}

	if len(preToolHook.Hooks) == 0 {
		t.Fatal("PreToolUse actions should not be empty")
	}

	action := preToolHook.Hooks[0]
	if action.Command != "dw claude emit-event" {
		t.Errorf("Expected command 'dw claude emit-event', got %q", action.Command)
	}
}

func TestMergeHookConfigs(t *testing.T) {
	existing := claude_code.HookConfig{
		Hooks: map[string][]claude_code.HookMatcher{
			"PreToolUse": {
				{
					Matcher: "*",
					Hooks: []claude_code.HookAction{
						{
							Type:    "command",
							Command: "existing-command",
							Timeout: 5,
						},
					},
				},
			},
		},
	}

	new := claude_code.HookConfig{
		Hooks: map[string][]claude_code.HookMatcher{
			"PreToolUse": {
				{
					Matcher: "*",
					Hooks: []claude_code.HookAction{
						{
							Type:    "command",
							Command: "new-command",
							Timeout: 10,
						},
					},
				},
			},
			"UserPromptSubmit": {
				{
					Hooks: []claude_code.HookAction{
						{
							Type:    "command",
							Command: "dw claude emit-event",
							Timeout: 5,
						},
					},
				},
			},
		},
	}

	merged := claude_code.MergeHookConfigs(existing, new)

	// Should have both events
	if len(merged.Hooks) != 2 {
		t.Errorf("Expected 2 hook events, got %d", len(merged.Hooks))
	}

	// PreToolUse should have both commands
	preToolHooks := merged.Hooks["PreToolUse"]
	if len(preToolHooks) != 1 {
		t.Errorf("Expected 1 matcher for PreToolUse, got %d", len(preToolHooks))
	}

	if len(preToolHooks[0].Hooks) != 2 {
		t.Errorf("Expected 2 actions in PreToolUse matcher, got %d", len(preToolHooks[0].Hooks))
	}

	// Check both commands are present
	commands := make(map[string]bool)
	for _, action := range preToolHooks[0].Hooks {
		commands[action.Command] = true
	}

	if !commands["existing-command"] {
		t.Fatal("existing-command not found after merge")
	}
	if !commands["new-command"] {
		t.Fatal("new-command not found after merge")
	}
}

func TestInstallDarwinFlowHooks(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	mgr, _ := claude_code.NewHookConfigManager()

	// Install hooks
	err = mgr.InstallDarwinFlowHooks()
	if err != nil {
		t.Fatalf("Failed to install hooks: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(mgr.GetSettingsPath()); os.IsNotExist(err) {
		t.Fatal("Settings file was not created")
	}

	// Read back and verify default hooks are present
	settings, err := mgr.ReadSettings()
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	defaultConfig := claude_code.DefaultDarwinFlowConfig()
	for event := range defaultConfig.Hooks {
		if _, ok := settings.Hooks[event]; !ok {
			t.Errorf("Expected hook event %q not found after install", event)
		}
	}
}

func TestWriteSettingsPreservesOtherFields(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	mgr, _ := claude_code.NewHookConfigManager()

	// Write initial settings with other fields
	settings := &claude_code.ClaudeSettings{
		Hooks: make(map[string][]claude_code.HookMatcher),
		Other: map[string]interface{}{
			"customField": "customValue",
			"nested": map[string]interface{}{
				"key": "value",
			},
		},
	}

	err = mgr.WriteSettings(settings)
	if err != nil {
		t.Fatalf("Failed to write settings: %v", err)
	}

	// Read raw JSON to verify other fields are preserved
	data, err := os.ReadFile(mgr.GetSettingsPath())
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var rawSettings map[string]interface{}
	if err := json.Unmarshal(data, &rawSettings); err != nil {
		t.Fatalf("Failed to parse settings JSON: %v", err)
	}

	if customField, ok := rawSettings["customField"]; !ok || customField != "customValue" {
		t.Error("customField was not preserved")
	}

	if _, ok := rawSettings["nested"]; !ok {
		t.Error("nested field was not preserved")
	}
}

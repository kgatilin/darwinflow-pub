package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// HookConfig represents the hooks configuration for Claude Code
type HookConfig struct {
	Hooks map[string][]HookMatcher `json:"hooks"`
}

// HookMatcher represents a hook matcher and its associated hooks
type HookMatcher struct {
	Matcher string       `json:"matcher,omitempty"`
	Hooks   []HookAction `json:"hooks"`
}

// HookAction represents a single hook action
type HookAction struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// DefaultDarwinFlowConfig returns the default hooks configuration for DarwinFlow logging
func DefaultDarwinFlowConfig() HookConfig {
	return HookConfig{
		Hooks: map[string][]HookMatcher{
			"PreToolUse": {
				{
					Matcher: "*", // Match all tools
					Hooks: []HookAction{
						{
							Type:    "command",
							Command: "dw claude log tool.invoked",
							Timeout: 5,
						},
					},
				},
			},
			"UserPromptSubmit": {
				{
					Hooks: []HookAction{
						{
							Type:    "command",
							Command: "dw claude log chat.message.user",
							Timeout: 5,
						},
					},
				},
			},
			"SessionEnd": {
				{
					Hooks: []HookAction{
						{
							Type:    "command",
							Command: "dw claude auto-summary",
							Timeout: 60, // Longer timeout for analysis
						},
					},
				},
			},
		},
	}
}

// MergeHookConfigs merges new hooks into existing configuration
func MergeHookConfigs(existing, new HookConfig) HookConfig {
	merged := HookConfig{
		Hooks: make(map[string][]HookMatcher),
	}

	// Copy existing hooks
	for event, matchers := range existing.Hooks {
		merged.Hooks[event] = append([]HookMatcher{}, matchers...)
	}

	// Add new hooks
	for event, newMatchers := range new.Hooks {
		if existingMatchers, ok := merged.Hooks[event]; ok {
			// Event exists, merge matchers
			merged.Hooks[event] = mergeMatchers(existingMatchers, newMatchers)
		} else {
			// New event, add all matchers
			merged.Hooks[event] = newMatchers
		}
	}

	return merged
}

// mergeMatchers merges new matchers into existing ones
func mergeMatchers(existing, new []HookMatcher) []HookMatcher {
	result := append([]HookMatcher{}, existing...)

	for _, newMatcher := range new {
		found := false
		for i, existingMatcher := range result {
			if existingMatcher.Matcher == newMatcher.Matcher {
				// Matcher exists, merge hooks
				result[i].Hooks = mergeHooks(existingMatcher.Hooks, newMatcher.Hooks)
				found = true
				break
			}
		}
		if !found {
			// New matcher, add it
			result = append(result, newMatcher)
		}
	}

	return result
}

// mergeHooks merges new hooks into existing ones, avoiding duplicates
func mergeHooks(existing, new []HookAction) []HookAction {
	result := append([]HookAction{}, existing...)

	for _, newHook := range new {
		duplicate := false
		for _, existingHook := range result {
			if existingHook.Command == newHook.Command {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result = append(result, newHook)
		}
	}

	return result
}

// HookConfigManager handles reading and writing Claude Code settings
type HookConfigManager struct {
	settingsPath string
}

// NewHookConfigManager creates a new hook configuration manager
func NewHookConfigManager() (*HookConfigManager, error) {
	settingsPath, err := findSettingsFile()
	if err != nil {
		return nil, err
	}

	return &HookConfigManager{
		settingsPath: settingsPath,
	}, nil
}

// findSettingsFile locates the Claude Code settings file
// Only returns local project settings files, never global settings
func findSettingsFile() (string, error) {
	// Check local settings first
	localSettings := []string{
		".claude/settings.local.json",
		".claude/settings.json",
	}

	for _, path := range localSettings {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Always default to local settings.json if none exist
	// This ensures we never modify global ~/.claude/settings.json
	return ".claude/settings.json", nil
}

// ClaudeSettings represents the Claude Code settings structure
type ClaudeSettings struct {
	Hooks map[string][]HookMatcher `json:"hooks,omitempty"`
	Other map[string]interface{}   `json:"-"` // For preserving unknown fields
}

// ReadSettings reads the current settings file
func (m *HookConfigManager) ReadSettings() (*ClaudeSettings, error) {
	// Check if file exists
	data, err := os.ReadFile(m.settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty settings
			return &ClaudeSettings{
				Hooks: make(map[string][]HookMatcher),
				Other: make(map[string]interface{}),
			}, nil
		}
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	// Parse JSON, preserving unknown fields
	var rawSettings map[string]interface{}
	if err := json.Unmarshal(data, &rawSettings); err != nil {
		return nil, fmt.Errorf("failed to parse settings JSON: %w", err)
	}

	settings := &ClaudeSettings{
		Other: make(map[string]interface{}),
	}

	// Extract hooks
	if hooksData, ok := rawSettings["hooks"]; ok {
		hooksJSON, _ := json.Marshal(hooksData)
		if err := json.Unmarshal(hooksJSON, &settings.Hooks); err != nil {
			return nil, fmt.Errorf("failed to parse hooks: %w", err)
		}
		delete(rawSettings, "hooks")
	} else {
		settings.Hooks = make(map[string][]HookMatcher)
	}

	// Store other fields
	settings.Other = rawSettings

	return settings, nil
}

// WriteSettings writes settings to the file
func (m *HookConfigManager) WriteSettings(settings *ClaudeSettings) error {
	// Ensure directory exists
	dir := filepath.Dir(m.settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Merge hooks with other settings
	output := make(map[string]interface{})
	for k, v := range settings.Other {
		output[k] = v
	}
	if len(settings.Hooks) > 0 {
		output["hooks"] = settings.Hooks
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// InstallDarwinFlowHooks adds DarwinFlow logging hooks to settings
func (m *HookConfigManager) InstallDarwinFlowHooks() error {
	// Read existing settings
	settings, err := m.ReadSettings()
	if err != nil {
		return err
	}

	// Create backup if file exists
	if _, err := os.Stat(m.settingsPath); err == nil {
		backupPath := m.settingsPath + ".backup"
		if err := copyFile(m.settingsPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Merge with default DarwinFlow hooks
	defaultConfig := DefaultDarwinFlowConfig()
	merged := MergeHookConfigs(
		HookConfig{Hooks: settings.Hooks},
		defaultConfig,
	)
	settings.Hooks = merged.Hooks

	// Write updated settings
	if err := m.WriteSettings(settings); err != nil {
		return err
	}

	return nil
}

// GetSettingsPath returns the path to the settings file
func (m *HookConfigManager) GetSettingsPath() string {
	return m.settingsPath
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

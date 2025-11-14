package task_manager

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ADRConfig holds configuration for ADR (Architecture Decision Records) requirements
type ADRConfig struct {
	Required                bool `yaml:"required" json:"required"`
	EnforceOnTaskCompletion bool `yaml:"enforce_on_task_completion" json:"enforce_on_task_completion"`
}

// Config holds all task-manager plugin configuration
type Config struct {
	ADR ADRConfig `yaml:"adr" json:"adr"`
}

// DefaultConfig returns the default configuration for the task-manager plugin
func DefaultConfig() *Config {
	return &Config{
		ADR: ADRConfig{
			Required:                false,
			EnforceOnTaskCompletion: false,
		},
	}
}

// LoadConfig loads configuration from file if it exists, otherwise returns default config
// It searches for config in order:
// 1. DW_CONFIG_PATH environment variable
// 2. Project-specific config at .darwinflow/config.yaml
// 3. Global config at ~/.darwinflow/config.yaml
// 4. Built-in defaults
func LoadConfig(workingDir string) (*Config, error) {
	// Start with default config
	cfg := DefaultConfig()

	// Try environment variable first
	if envPath := os.Getenv("DW_CONFIG_PATH"); envPath != "" {
		if err := loadConfigFromFile(envPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config from DW_CONFIG_PATH: %w", err)
		}
		return cfg, nil
	}

	// Try project-specific config
	projectConfigPath := filepath.Join(workingDir, ".darwinflow", "config.yaml")
	if _, err := os.Stat(projectConfigPath); err == nil {
		if err := loadConfigFromFile(projectConfigPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to load project config: %w", err)
		}
		return cfg, nil
	}

	// Try global config
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfigPath := filepath.Join(homeDir, ".darwinflow", "config.yaml")
		if _, err := os.Stat(globalConfigPath); err == nil {
			if err := loadConfigFromFile(globalConfigPath, cfg); err != nil {
				return nil, fmt.Errorf("failed to load global config: %w", err)
			}
			return cfg, nil
		}
	}

	// Return defaults if no config files found
	return cfg, nil
}

// loadConfigFromFile loads configuration from a YAML file
func loadConfigFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Create a map to hold raw config values to allow partial overrides
	rawCfg := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &rawCfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract task_manager section if it exists
	if taskManagerCfgRaw, ok := rawCfg["task_manager"]; ok {
		var taskManagerCfg map[interface{}]interface{}
		// Handle both interface{} and map types
		switch v := taskManagerCfgRaw.(type) {
		case map[interface{}]interface{}:
			taskManagerCfg = v
		case map[string]interface{}:
			// Convert string keys to interface{} keys
			taskManagerCfg = make(map[interface{}]interface{})
			for k, v := range v {
				taskManagerCfg[k] = v
			}
		default:
			return nil
		}

		// Apply ADR config if present
		if adrCfgRaw, ok := taskManagerCfg["adr"]; ok {
			var adrCfg map[interface{}]interface{}
			// Handle both interface{} and map types
			switch v := adrCfgRaw.(type) {
			case map[interface{}]interface{}:
				adrCfg = v
			case map[string]interface{}:
				// Convert string keys to interface{} keys
				adrCfg = make(map[interface{}]interface{})
				for k, v := range v {
					adrCfg[k] = v
				}
			default:
				return nil
			}

			if required, ok := adrCfg["required"].(bool); ok {
				cfg.ADR.Required = required
			}
			if enforce, ok := adrCfg["enforce_on_task_completion"].(bool); ok {
				cfg.ADR.EnforceOnTaskCompletion = enforce
			}
		}
	}

	return nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(path string, cfg *Config) error {
	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	cfgMap := map[string]interface{}{
		"task_manager": map[string]interface{}{
			"adr": map[string]interface{}{
				"required":                  cfg.ADR.Required,
				"enforce_on_task_completion": cfg.ADR.EnforceOnTaskCompletion,
			},
		},
	}

	data, err := yaml.Marshal(cfgMap)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

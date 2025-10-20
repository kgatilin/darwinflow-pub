package infra

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

const (
	// DefaultConfigFileName is the name of the config file
	DefaultConfigFileName = ".darwinflow.yaml"
)

// ConfigLoader loads DarwinFlow configuration
type ConfigLoader struct {
	logger *Logger
}

// NewConfigLoader creates a new config loader
func NewConfigLoader(logger *Logger) *ConfigLoader {
	return &ConfigLoader{
		logger: logger,
	}
}

// LoadConfig loads configuration from the specified directory
// If configPath is empty, it looks for .darwinflow.yaml in the current directory
// Falls back to default config if file doesn't exist
func (c *ConfigLoader) LoadConfig(configPath string) (*domain.Config, error) {
	// Determine config file path
	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		configPath = filepath.Join(cwd, DefaultConfigFileName)
	}

	if c.logger != nil {
		c.logger.Debug("Looking for config at: %s", configPath)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if c.logger != nil {
			c.logger.Debug("Config file not found, using defaults")
		}
		return domain.DefaultConfig(), nil
	}

	if c.logger != nil {
		c.logger.Debug("Loading config from file")
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config domain.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure prompts map is initialized
	if config.Prompts == nil {
		config.Prompts = make(map[string]string)
	}

	// Fill in missing default prompts
	defaults := domain.DefaultConfig()
	for key, value := range defaults.Prompts {
		if _, exists := config.Prompts[key]; !exists {
			config.Prompts[key] = value
		}
	}

	// Apply defaults for analysis config fields if not set
	if config.Analysis.TokenLimit == 0 {
		config.Analysis.TokenLimit = defaults.Analysis.TokenLimit
	}
	if config.Analysis.Model == "" {
		config.Analysis.Model = defaults.Analysis.Model
	}
	if config.Analysis.ParallelLimit == 0 {
		config.Analysis.ParallelLimit = defaults.Analysis.ParallelLimit
	}
	if len(config.Analysis.EnabledPrompts) == 0 {
		config.Analysis.EnabledPrompts = defaults.Analysis.EnabledPrompts
	}
	if config.Analysis.AutoSummaryPrompt == "" {
		config.Analysis.AutoSummaryPrompt = defaults.Analysis.AutoSummaryPrompt
	}
	if config.Analysis.ClaudeOptions.SystemPromptMode == "" {
		config.Analysis.ClaudeOptions.SystemPromptMode = defaults.Analysis.ClaudeOptions.SystemPromptMode
	}

	// Validate model is in whitelist
	if !domain.ValidateModel(config.Analysis.Model) {
		if c.logger != nil {
			c.logger.Warn("Invalid model '%s', using default '%s'", config.Analysis.Model, defaults.Analysis.Model)
		}
		config.Analysis.Model = defaults.Analysis.Model
	}

	// Validate enabled prompts exist
	validPrompts := []string{}
	for _, promptName := range config.Analysis.EnabledPrompts {
		if _, exists := config.Prompts[promptName]; exists {
			validPrompts = append(validPrompts, promptName)
		} else {
			if c.logger != nil {
				c.logger.Warn("Enabled prompt '%s' not found in prompts section, ignoring", promptName)
			}
		}
	}
	config.Analysis.EnabledPrompts = validPrompts

	// If all enabled prompts were invalid, use default
	if len(config.Analysis.EnabledPrompts) == 0 {
		if c.logger != nil {
			c.logger.Warn("No valid enabled prompts, using default: %v", defaults.Analysis.EnabledPrompts)
		}
		config.Analysis.EnabledPrompts = defaults.Analysis.EnabledPrompts
	}

	if c.logger != nil {
		c.logger.Info("Config loaded successfully with %d prompt(s), enabled: %v, model: %s",
			len(config.Prompts), config.Analysis.EnabledPrompts, config.Analysis.Model)
	}
	return &config, nil
}

// ValidateModelAlias validates a model alias or full name against the whitelist
func ValidateModelAlias(model string) bool {
	return domain.ValidateModel(model)
}

// SaveConfig saves configuration to the specified path and returns the actual path used
func (c *ConfigLoader) SaveConfig(config *domain.Config, configPath string) (string, error) {
	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		configPath = filepath.Join(cwd, DefaultConfigFileName)
	}

	if c.logger != nil {
		c.logger.Debug("Saving config to: %s", configPath)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	if c.logger != nil {
		c.logger.Info("Config saved to %s", configPath)
	}
	return configPath, nil
}

// GetPrompt retrieves a named prompt from the config
// Returns the prompt and true if found, empty string and false otherwise
func (c *ConfigLoader) GetPrompt(config *domain.Config, promptName string) (string, bool) {
	if config == nil || config.Prompts == nil {
		return "", false
	}
	prompt, exists := config.Prompts[promptName]
	return prompt, exists
}

// InitializeDefaultConfig creates and saves a default config file and returns the path used
func (c *ConfigLoader) InitializeDefaultConfig(configPath string) (string, error) {
	config := domain.DefaultConfig()
	return c.SaveConfig(config, configPath)
}

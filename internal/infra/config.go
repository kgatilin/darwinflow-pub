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

	if c.logger != nil {
		c.logger.Info("Config loaded successfully with %d prompt(s)", len(config.Prompts))
	}
	return &config, nil
}

// SaveConfig saves configuration to the specified path
func (c *ConfigLoader) SaveConfig(config *domain.Config, configPath string) error {
	if configPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		configPath = filepath.Join(cwd, DefaultConfigFileName)
	}

	if c.logger != nil {
		c.logger.Debug("Saving config to: %s", configPath)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if c.logger != nil {
		c.logger.Info("Config saved to %s", configPath)
	}
	return nil
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

// InitializeDefaultConfig creates and saves a default config file
func (c *ConfigLoader) InitializeDefaultConfig(configPath string) error {
	config := domain.DefaultConfig()
	return c.SaveConfig(config, configPath)
}

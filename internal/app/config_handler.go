package app

import (
	"context"
	"fmt"
	"io"
)

// ConfigCommandHandler handles config command operations
type ConfigCommandHandler struct {
	configLoader ConfigLoader
	logger       Logger
	output       io.Writer
}

// NewConfigCommandHandler creates a new config command handler
func NewConfigCommandHandler(
	configLoader ConfigLoader,
	logger Logger,
	output io.Writer,
) *ConfigCommandHandler {
	return &ConfigCommandHandler{
		configLoader: configLoader,
		logger:       logger,
		output:       output,
	}
}

// Init creates a default configuration file
func (h *ConfigCommandHandler) Init(ctx context.Context, configPath string, force bool) error {
	// Create and save default config
	// The config loader will check if file exists and handle force flag
	createdPath, err := h.configLoader.InitializeDefaultConfig(configPath)
	if err != nil {
		// Check if error is because file already exists
		if !force {
			return fmt.Errorf("failed to create config: %w (use --force to overwrite existing file)", err)
		}
		return fmt.Errorf("failed to create config: %w", err)
	}

	fmt.Fprintf(h.output, "Created config file: %s\n", createdPath)
	fmt.Fprintln(h.output, "\nYou can now customize the prompts in this file for your project.")

	return nil
}

// Show displays the current configuration
func (h *ConfigCommandHandler) Show(ctx context.Context) error {
	config, err := h.configLoader.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Fprintln(h.output, "=== DarwinFlow Configuration ===")
	fmt.Fprintf(h.output, "\nPrompts defined: %d\n", len(config.Prompts))
	for name := range config.Prompts {
		fmt.Fprintf(h.output, "  - %s\n", name)
	}
	fmt.Fprintln(h.output, "\nTo edit prompts, modify .darwinflow.yaml in your project root")

	return nil
}

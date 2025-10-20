package app

import (
	"context"
	"fmt"
	"io"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// RefreshCommandHandler handles the refresh command logic
type RefreshCommandHandler struct {
	repo              domain.EventRepository
	hookConfigManager HookConfigManager
	configLoader      ConfigLoader
	logger            Logger
	out               io.Writer
}

// ConfigLoader interface for config loading
type ConfigLoader interface {
	LoadConfig(path string) (*domain.Config, error)
	InitializeDefaultConfig(path string) (string, error)
}

// NewRefreshCommandHandler creates a new refresh command handler
func NewRefreshCommandHandler(
	repo domain.EventRepository,
	hookConfigManager HookConfigManager,
	configLoader ConfigLoader,
	logger Logger,
	out io.Writer,
) *RefreshCommandHandler {
	return &RefreshCommandHandler{
		repo:              repo,
		hookConfigManager: hookConfigManager,
		configLoader:      configLoader,
		logger:            logger,
		out:               out,
	}
}

// Execute runs the refresh operation
func (h *RefreshCommandHandler) Execute(ctx context.Context, dbPath string) error {
	fmt.Fprintln(h.out, "Refreshing DarwinFlow to latest version...")
	fmt.Fprintln(h.out)

	// Step 1: Update database schema
	fmt.Fprintln(h.out, "Updating database schema...")
	if err := h.repo.Initialize(ctx); err != nil {
		return fmt.Errorf("error updating database schema: %w", err)
	}
	fmt.Fprintf(h.out, "✓ Database schema updated: %s\n", dbPath)

	// Step 2: Update hooks
	fmt.Fprintln(h.out)
	fmt.Fprintln(h.out, "Updating Claude Code hooks...")
	if err := h.hookConfigManager.InstallDarwinFlowHooks(); err != nil {
		return fmt.Errorf("error updating hooks: %w", err)
	}
	fmt.Fprintf(h.out, "✓ Hooks updated: %s\n", h.hookConfigManager.GetSettingsPath())

	// Step 3: Initialize config if needed
	fmt.Fprintln(h.out)
	fmt.Fprintln(h.out, "Checking configuration...")

	// Try to load config
	config, err := h.configLoader.LoadConfig("")
	if err != nil || config == nil {
		// Config doesn't exist or is invalid, create default
		fmt.Fprintln(h.out, "Creating default configuration...")
		configPath, err := h.configLoader.InitializeDefaultConfig("")
		if err != nil {
			fmt.Fprintf(h.out, "Warning: Could not create default config: %v\n", err)
		} else {
			fmt.Fprintf(h.out, "✓ Configuration initialized: %s\n", configPath)
		}
	} else {
		fmt.Fprintln(h.out, "✓ Configuration is valid")
	}

	// Done
	fmt.Fprintln(h.out)
	fmt.Fprintln(h.out, "DarwinFlow has been refreshed successfully!")
	fmt.Fprintln(h.out)
	fmt.Fprintln(h.out, "Changes applied:")
	fmt.Fprintln(h.out, "  • Database schema updated with latest migrations")
	fmt.Fprintln(h.out, "  • Hooks updated to latest version")
	fmt.Fprintln(h.out, "  • Configuration verified")
	fmt.Fprintln(h.out)
	fmt.Fprintln(h.out, "You may need to restart Claude Code for changes to take effect.")

	return nil
}

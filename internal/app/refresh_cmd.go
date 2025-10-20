package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// RefreshCommandHandler handles the refresh command logic.
// This handler is plugin-agnostic: it queries all plugins for IHookProvider capability
// and refreshes hooks for each one that provides them.
type RefreshCommandHandler struct {
	repo          domain.EventRepository
	pluginRegistry *PluginRegistry
	configLoader  ConfigLoader
	logger        Logger
	out           io.Writer
}

// ConfigLoader interface for config loading
type ConfigLoader interface {
	LoadConfig(path string) (*domain.Config, error)
	InitializeDefaultConfig(path string) (string, error)
}

// NewRefreshCommandHandler creates a new refresh command handler
func NewRefreshCommandHandler(
	repo domain.EventRepository,
	pluginRegistry *PluginRegistry,
	configLoader ConfigLoader,
	logger Logger,
	out io.Writer,
) *RefreshCommandHandler {
	return &RefreshCommandHandler{
		repo:          repo,
		pluginRegistry: pluginRegistry,
		configLoader:  configLoader,
		logger:        logger,
		out:           out,
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

	// Step 2: Update hooks from all plugins
	fmt.Fprintln(h.out)
	fmt.Fprintln(h.out, "Updating hooks for all plugins...")

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// Get all registered plugins
	plugins := h.pluginRegistry.GetAllPlugins()
	hasPluginErrors := false

	for _, plugin := range plugins {
		// Check if plugin implements IHookProvider capability
		hookProvider, ok := plugin.(pluginsdk.IHookProvider)
		if !ok {
			// Plugin doesn't provide hooks, skip it
			continue
		}

		// Refresh hooks for this plugin
		fmt.Fprintf(h.out, "  → Refreshing hooks for plugin: %s\n", plugin.GetInfo().Name)
		if err := hookProvider.RefreshHooks(workingDir); err != nil {
			fmt.Fprintf(h.out, "    ⚠️  Hook refresh failed for %s: %v\n", plugin.GetInfo().Name, err)
			h.logger.Warn("Hook refresh failed for plugin %s: %v", plugin.GetInfo().Name, err)
			hasPluginErrors = true
			// Continue to other plugins, don't fail completely
		} else {
			fmt.Fprintf(h.out, "    ✓ Hooks refreshed for %s\n", plugin.GetInfo().Name)
		}
	}

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
	fmt.Fprintln(h.out, "  • Hooks refreshed for all plugins")
	fmt.Fprintln(h.out, "  • Configuration verified")
	fmt.Fprintln(h.out)

	if hasPluginErrors {
		fmt.Fprintln(h.out, "Note: Some plugins had errors during hook refresh. See output above.")
	}

	fmt.Fprintln(h.out, "You may need to restart Claude Code for changes to take effect.")

	return nil
}

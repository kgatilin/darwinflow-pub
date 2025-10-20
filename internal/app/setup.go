package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

const (
	// DefaultDBPath is the default location for the event database
	DefaultDBPath = ".darwinflow/logs/events.db"
)

// SetupService orchestrates initialization of the DarwinFlow logging infrastructure.
// This service is plugin-agnostic: it queries plugins for IHookProvider capability
// rather than hardcoding plugin-specific logic. This allows any plugin to provide hooks.
type SetupService struct {
	repository domain.EventRepository
	logger     Logger
}

// NewSetupService creates a new setup service
func NewSetupService(
	repository domain.EventRepository,
	logger Logger,
) *SetupService {
	return &SetupService{
		repository: repository,
		logger:     logger,
	}
}

// Initialize sets up the complete logging infrastructure.
// It accepts a list of plugins and queries each one for IHookProvider capability.
// Each plugin that provides hooks has its hooks installed.
func (s *SetupService) Initialize(ctx context.Context, dbPath string, plugins []pluginsdk.Plugin) error {
	// Create database directory
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Initialize repository (create schema, indexes, etc.)
	if err := s.repository.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Install hooks from all plugins that provide them
	var hookErrors []error
	for _, plugin := range plugins {
		// Check if plugin implements IHookProvider capability
		hookProvider, ok := plugin.(pluginsdk.IHookProvider)
		if !ok {
			// Plugin doesn't provide hooks, skip it
			continue
		}

		// Install hooks for this plugin
		if err := hookProvider.InstallHooks(dir); err != nil {
			hookErrors = append(hookErrors, fmt.Errorf("plugin %s hook install: %w", plugin.GetInfo().Name, err))
		}
	}

	// If any plugin had hook installation errors, return them
	if len(hookErrors) > 0 {
		// Log the error but don't fail initialization - one plugin's failure shouldn't block others
		for _, err := range hookErrors {
			s.logger.Warn("Hook installation warning: %v", err)
		}
		// Still return the errors so caller can decide how to handle
		return fmt.Errorf("hook installation errors: %v", hookErrors)
	}

	return nil
}

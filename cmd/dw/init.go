package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kgatilin/darwinflow-pub/internal/app"
)

// handleInit orchestrates the initialization of DarwinFlow:
// 1. Creates the event database
// 2. Initializes repositories
// 3. Discovers and registers plugins
// 4. Calls each plugin's init command (if they provide one)
func handleInit(args []string) {
	ctx := context.Background()

	fmt.Println("Initializing DarwinFlow...")
	fmt.Println()

	// 1. Create event database
	dbPath := app.DefaultDBPath
	if err := createEventDatabase(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating database: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Created event database: %s\n", dbPath)

	// 2. Initialize app services (which creates repository and registers plugins)
	services, err := InitializeApp(dbPath, "", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Initialized event repository")

	// 3. Display discovered plugins
	pluginInfos := services.PluginRegistry.GetPluginInfos()
	fmt.Println()
	fmt.Println("Discovered plugins:")
	for _, info := range pluginInfos {
		coreLabel := ""
		if info.IsCore {
			coreLabel = " (core)"
		}
		fmt.Printf("  - %s v%s%s\n", info.Name, info.Version, coreLabel)
	}

	// 4. Initialize each plugin
	fmt.Println()
	fmt.Println("Initializing plugins...")
	for _, info := range pluginInfos {
		if err := initializePlugin(ctx, services, info.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing plugin %s: %v\n", info.Name, err)
			os.Exit(1)
		}
	}

	fmt.Println()
	fmt.Println("✓ DarwinFlow initialization complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Restart Claude Code to activate the hooks")
	fmt.Println("  2. Events will be automatically logged to", dbPath)
	fmt.Println("  3. Use 'dw logs' to view logged events")
	fmt.Println("  4. Use 'dw ui' to browse sessions interactively")
	fmt.Println("  5. Use 'dw analyze' to analyze sessions")
}

// createEventDatabase ensures the database directory exists and is writable
func createEventDatabase(dbPath string) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		// Database already exists, that's fine (idempotent)
		return nil
	}

	// Database will be created by repository Init, we just ensure directory exists
	return nil
}

// initializePlugin checks if a plugin provides an "init" command and executes it
// Uses the command registry to execute the plugin's init command
func initializePlugin(ctx context.Context, services *AppServices, pluginName string) error {
	// Create command context for plugin
	cmdCtx := app.NewCommandContext(
		services.Logger,
		services.DBPath,
		services.WorkingDir,
		services.EventRepo,
		os.Stdout,
		os.Stdin,
	)

	// Try to execute the init command via the command registry
	fmt.Printf("  → Running: dw %s init\n", pluginName)
	if err := services.CommandRegistry.ExecuteCommand(ctx, pluginName, "init", []string{}, cmdCtx); err != nil {
		// If the error is "command not found", the plugin doesn't have an init command
		// This is fine - just skip silently
		if err.Error() == fmt.Sprintf("command not found: %s init", pluginName) {
			return nil
		}
		return fmt.Errorf("init command failed: %w", err)
	}

	return nil
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

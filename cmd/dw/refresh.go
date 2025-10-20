package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

// handleRefresh updates DarwinFlow to the latest version
// This includes:
// - Updating database schema (adding new columns, indexes, etc.)
// - Refreshing hooks for all plugins
// - Updating configuration if needed
func handleRefresh(args []string) {
	dbPath := app.DefaultDBPath

	// Initialize app to get plugin registry
	services, err := InitializeApp(dbPath, "", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}

	// The EventRepo is stored as interface{}, but we need to cast it to EventRepository
	// for RefreshCommandHandler
	repo, ok := services.EventRepo.(*infra.SQLiteEventRepository)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: Invalid repository type\n")
		os.Exit(1)
	}

	logger := services.Logger
	configLoader := services.ConfigLoader
	pluginRegistry := services.PluginRegistry

	// Create handler with plugin registry
	handler := app.NewRefreshCommandHandler(
		repo,
		pluginRegistry,
		configLoader,
		logger,
		os.Stdout,
	)

	// Execute
	ctx := context.Background()
	if err := handler.Execute(ctx, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

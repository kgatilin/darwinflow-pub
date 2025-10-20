package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

// handleRefresh updates DarwinFlow to the latest version
// This includes:
// - Updating database schema (adding new columns, indexes, etc.)
// - Reinstalling/updating hooks
// - Updating configuration if needed
func handleRefresh(args []string) {
	dbPath := app.DefaultDBPath

	// Create infrastructure dependencies
	repository, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repository.Close()

	hookConfigManager, err := claude_code.NewHookConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating hook config manager: %v\n", err)
		os.Exit(1)
	}

	logger := infra.NewDefaultLogger()
	configLoader := infra.NewConfigLoader(logger)

	// Create handler
	handler := app.NewRefreshCommandHandler(repository, hookConfigManager, configLoader, logger, os.Stdout)

	// Execute
	ctx := context.Background()
	if err := handler.Execute(ctx, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

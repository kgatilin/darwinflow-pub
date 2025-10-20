package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func configCmd(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: dw config <subcommand>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Subcommands:")
		fmt.Fprintln(os.Stderr, "  init    Create a default .darwinflow.yaml config file")
		fmt.Fprintln(os.Stderr, "  show    Display the current configuration")
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "init":
		configInitCmd(subArgs)
	case "show":
		configShowCmd(subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown config subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func configInitCmd(args []string) {
	fs := flag.NewFlagSet("config init", flag.ContinueOnError)
	force := fs.Bool("force", false, "Overwrite existing config file")
	debug := fs.Bool("debug", false, "Enable debug logging")
	debugShort := fs.Bool("d", false, "Enable debug logging (short flag)")

	if err := fs.Parse(args); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create logger
	var logger *infra.Logger
	if *debug || *debugShort {
		logger = infra.NewDebugLogger()
		logger.Info("Debug logging enabled")
	} else {
		logger = infra.NewDefaultLogger()
	}

	configLoader := infra.NewConfigLoader(logger)

	// Check if config already exists
	configPath := infra.DefaultConfigFileName
	if _, err := os.Stat(configPath); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "Config file %s already exists. Use --force to overwrite.\n", configPath)
		os.Exit(1)
	}

	// Create handler
	handler := app.NewConfigCommandHandler(configLoader, logger, os.Stdout)

	// Execute
	ctx := context.Background()
	if err := handler.Init(ctx, "", *force); err != nil {
		logger.Error("Failed to create config: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to create config: %v\n", err)
		os.Exit(1)
	}
}

func configShowCmd(args []string) {
	fs := flag.NewFlagSet("config show", flag.ContinueOnError)
	debug := fs.Bool("debug", false, "Enable debug logging")
	debugShort := fs.Bool("d", false, "Enable debug logging (short flag)")

	if err := fs.Parse(args); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create logger
	var logger *infra.Logger
	if *debug || *debugShort {
		logger = infra.NewDebugLogger()
	} else {
		logger = infra.NewDefaultLogger()
	}

	configLoader := infra.NewConfigLoader(logger)

	// Create handler
	handler := app.NewConfigCommandHandler(configLoader, logger, os.Stdout)

	// Execute
	ctx := context.Background()
	if err := handler.Show(ctx); err != nil {
		logger.Error("Failed to load config: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
}

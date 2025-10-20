package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Handle help first
	if command == "help" || command == "--help" || command == "-h" {
		printUsage()
		return
	}

	// Handle init command specially - it bootstraps the system
	if command == "init" {
		handleInit(args)
		return
	}

	// Handle ui command specially - it has its own initialization with custom flags
	if command == "ui" {
		uiCommand(args)
		return
	}

	// Initialize app (includes plugin registration)
	// Use default DB path, can be overridden by command flags
	services, err := InitializeApp(app.DefaultDBPath, "", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Route command
	switch command {
	case "logs":
		handleLogs(args)
	case "analyze":
		analyzeCmd(args)
	case "refresh":
		handleRefresh(args)
	case "config":
		configCmd(args)
	case "claude":
		// Backward compatibility: "dw claude <command>" -> "dw claude-code <command>"
		if len(args) > 0 {
			cmdCtx := app.NewCommandContext(services.Logger, services.DBPath, services.WorkingDir, services.EventRepo, os.Stdout, os.Stdin)
			if err := services.CommandRegistry.ExecuteCommand(ctx, "claude-code", args[0], args[1:], cmdCtx); err != nil {
				fmt.Fprintf(os.Stderr, "Error executing claude-code command: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: claude subcommand required\n")
			fmt.Fprintf(os.Stderr, "Usage: dw claude <subcommand>\n")
			os.Exit(1)
		}
	default:
		// Try plugin commands: dw <plugin-name> <command> [args]
		cmdCtx := app.NewCommandContext(services.Logger, services.DBPath, services.WorkingDir, services.EventRepo, os.Stdout, os.Stdin)
		if len(args) > 0 {
			// Try as: dw <plugin> <command> [args]
			if err := services.CommandRegistry.ExecuteCommand(ctx, command, args[0], args[1:], cmdCtx); err == nil {
				return
			}
		}
		// Try as: dw <command> (single-word plugin command)
		if err := services.CommandRegistry.ExecuteCommand(ctx, command, "", args, cmdCtx); err == nil {
			return
		}
		// Unknown command
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("dw - DarwinFlow CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dw <plugin> <command> [args]   Run a plugin command")
	fmt.Println()
	fmt.Println("Built-in Commands:")
	fmt.Println("  dw init              Initialize DarwinFlow and all plugins")
	fmt.Println("  dw logs              View logged events from the database")
	fmt.Println("  dw analyze           Analyze sessions to identify tool gaps and inefficiencies")
	fmt.Println("  dw ui                Interactive UI for browsing and analyzing sessions")
	fmt.Println("  dw config            Manage DarwinFlow configuration")
	fmt.Println("  dw refresh           Update database schema and hooks to latest version")
	fmt.Println("  dw help              Show this help message")
	fmt.Println()
	fmt.Println("Plugin Commands:")
	fmt.Println("  dw claude init                          Initialize Claude Code logging (backward compat)")
	fmt.Println("  dw claude log <event-type>              Log a Claude Code event (backward compat)")
	fmt.Println("  dw claude-code init                     Initialize Claude Code logging")
	fmt.Println("  dw claude-code log <event-type>         Log a Claude Code event")
	fmt.Println("  dw claude-code session-summary [flags]  Display session summary")
	fmt.Println()
	fmt.Println("For command-specific help:")
	fmt.Println("  dw logs --help       Show logs command help and database schema")
	fmt.Println("  dw analyze --help    Show analyze command options")
	fmt.Println("  dw config --help     Show config command options")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DW_CONTEXT           Set the current context (e.g., project/myapp)")
	fmt.Println()
}

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "claude":
		handleClaudeCommand(os.Args[2:])
	case "logs":
		handleLogs(os.Args[2:])
	case "analyze":
		analyzeCmd(os.Args[2:])
	case "config":
		configCmd(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("dw - DarwinFlow CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dw claude init       Initialize Claude Code logging")
	fmt.Println("  dw claude log        Log a Claude Code event")
	fmt.Println("  dw logs              View logged events from the database")
	fmt.Println("  dw analyze           Analyze sessions to identify tool gaps and inefficiencies")
	fmt.Println("  dw config            Manage DarwinFlow configuration")
	fmt.Println("  dw help              Show this help message")
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

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kgatilin/darwinflow-pub/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		printUsageWithPlugins()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Handle help first
	if command == "help" || command == "--help" || command == "-h" {
		printUsageWithPlugins()
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
	case "plugin":
		pluginCmd(args)
		return
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
		// Check if this is a plugin help request: dw <plugin> --help
		if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
			if printPluginHelp(services, command) {
				return
			}
		}
		// If no args and it's a known plugin, show plugin help
		if len(args) == 0 {
			if printPluginHelp(services, command) {
				return
			}
		}

		// Try plugin commands: dw <plugin-name> <command> [args]
		cmdCtx := app.NewCommandContext(services.Logger, services.DBPath, services.WorkingDir, services.EventRepo, os.Stdout, os.Stdin)
		if len(args) > 0 {
			// Try as: dw <plugin> <command> [args]
			err := services.CommandRegistry.ExecuteCommand(ctx, command, args[0], args[1:], cmdCtx)
			if err == nil {
				return
			}
			// If command execution failed, check if it's because plugin/command doesn't exist
			// or if it's an actual execution error
			if isPluginOrCommandNotFound(err) {
				// Try single-word fallback
				if err := services.CommandRegistry.ExecuteCommand(ctx, command, "", args, cmdCtx); err == nil {
					return
				}
			} else {
				// Command exists but failed - show error and help
				fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
				printCommandHelp(services, command, args[0])
				os.Exit(1)
			}
		}
		// Try as: dw <command> (single-word plugin command)
		if err := services.CommandRegistry.ExecuteCommand(ctx, command, "", args, cmdCtx); err == nil {
			return
		}
		// Unknown command - show full help with loaded plugins
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printFullUsage(services)
		os.Exit(1)
	}
}

// printUsageWithPlugins initializes the app to load plugins, then prints full usage
func printUsageWithPlugins() {
	services, err := InitializeApp(app.DefaultDBPath, "", false)
	if err != nil {
		// Fallback to basic usage if can't initialize
		printBasicUsage()
		return
	}
	printFullUsage(services)
}

// printBasicUsage shows help without plugin commands (fallback)
func printBasicUsage() {
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
	fmt.Println("  dw plugin            Manage plugins (list, reload)")
	fmt.Println("  dw help              Show this help message")
	fmt.Println()
	fmt.Println("For command-specific help:")
	fmt.Println("  dw logs --help       Show logs command help and database schema")
	fmt.Println("  dw analyze --help    Show analyze command options")
	fmt.Println("  dw config --help     Show config command options")
	fmt.Println("  dw plugin --help     Show plugin command options")
	fmt.Println()
}

// printFullUsage shows help with all registered plugin commands
func printFullUsage(services *AppServices) {
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
	fmt.Println("  dw plugin            Manage plugins (list, reload)")
	fmt.Println("  dw help              Show this help message")
	fmt.Println()

	// Dynamically list all plugin commands
	fmt.Println("Plugin Commands:")

	// Get all commands from all plugins
	allCommands := services.CommandRegistry.GetAllCommands()
	if len(allCommands) == 0 {
		fmt.Println("  (no plugins with commands registered)")
	} else {
		for pluginName, commands := range allCommands {
			pluginInfo, err := services.PluginRegistry.GetPlugin(pluginName)
			if err != nil {
				continue
			}

			info := pluginInfo.GetInfo()
			fmt.Printf("\n  %s (%s):\n", info.Name, info.Description)

			for _, cmd := range commands {
				// Format: "    dw <plugin> <command>    <description>"
				cmdLine := fmt.Sprintf("    dw %s %s", pluginName, cmd.GetName())
				fmt.Printf("%-45s %s\n", cmdLine, cmd.GetDescription())
			}
		}
	}

	fmt.Println()
	fmt.Println("Backward Compatibility:")
	fmt.Println("  dw claude <command>  Alias for 'dw claude-code <command>'")
	fmt.Println()
	fmt.Println("For command-specific help:")
	fmt.Println("  dw logs --help       Show logs command help and database schema")
	fmt.Println("  dw analyze --help    Show analyze command options")
	fmt.Println("  dw config --help     Show config command options")
	fmt.Println("  dw plugin --help     Show plugin command options")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DW_CONTEXT           Set the current context (e.g., project/myapp)")
	fmt.Println()
}

// printPluginHelp shows help for a specific plugin and its commands
func printPluginHelp(services *AppServices, pluginName string) bool {
	// Check if plugin exists
	plugin, err := services.PluginRegistry.GetPlugin(pluginName)
	if err != nil {
		return false
	}

	// Get plugin info
	info := plugin.GetInfo()

	// Print plugin header
	fmt.Printf("Plugin: %s (version %s)\n", info.Name, info.Version)
	fmt.Printf("Description: %s\n\n", info.Description)

	// Get and display commands
	commands := services.CommandRegistry.GetCommandsForPlugin(pluginName)
	if len(commands) == 0 {
		fmt.Println("This plugin provides no commands.")
		return true
	}

	fmt.Println("Available Commands:")
	fmt.Println()

	for _, cmd := range commands {
		// Format: "  dw <plugin> <command>    <description>"
		cmdLine := fmt.Sprintf("  dw %s %s", pluginName, cmd.GetName())
		fmt.Printf("%-40s %s\n", cmdLine, cmd.GetDescription())
	}

	fmt.Println()
	fmt.Println("For command-specific help:")
	fmt.Printf("  dw %s <command> --help\n", pluginName)
	fmt.Println()

	return true
}

// printCommandHelp shows help for a specific command, or plugin help if command not found
func printCommandHelp(services *AppServices, pluginName, commandName string) {
	// Try to get the specific command
	cmd, err := services.CommandRegistry.GetCommand(pluginName, commandName)
	if err == nil && cmd != nil {
		// Show command-specific help
		fmt.Printf("Command: dw %s %s\n\n", pluginName, commandName)
		fmt.Printf("Description:\n  %s\n\n", cmd.GetDescription())
		fmt.Printf("Usage:\n  %s\n\n", cmd.GetUsage())
		if help := cmd.GetHelp(); help != "" {
			fmt.Printf("%s\n", help)
		}
		return
	}

	// Command not found, show plugin help instead
	if !printPluginHelp(services, pluginName) {
		// Plugin doesn't exist either
		fmt.Fprintf(os.Stderr, "Plugin '%s' not found\n", pluginName)
	}
}

// isPluginOrCommandNotFound checks if error indicates missing plugin/command
func isPluginOrCommandNotFound(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "plugin not found") ||
		strings.Contains(errMsg, "command not found") ||
		strings.Contains(errMsg, "does not provide commands")
}

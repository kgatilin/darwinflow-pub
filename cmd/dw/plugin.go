package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

// pluginCmd handles the "dw plugin" command and its subcommands
func pluginCmd(args []string) {
	if len(args) == 0 {
		printPluginCmdHelp()
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "list":
		handlePluginList(subArgs)
	case "reload":
		handlePluginReload(subArgs)
	case "--help", "-h", "help":
		printPluginCmdHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown plugin subcommand: %s\n\n", subcommand)
		printPluginCmdHelp()
		os.Exit(1)
	}
}

// handlePluginList lists all registered plugins
func handlePluginList(args []string) {
	// Parse flags
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printPluginListHelp()
		return
	}

	// Initialize app to get plugin registry
	services, err := InitializeApp(app.DefaultDBPath, "", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}

	// Get all plugins
	allPlugins := services.PluginRegistry.GetAllPlugins()
	if len(allPlugins) == 0 {
		fmt.Println("No plugins registered.")
		return
	}

	// Print header
	fmt.Println("Registered Plugins:")

	// Track counts
	coreCount := 0
	externalCount := 0

	// List plugins
	for _, plugin := range allPlugins {
		info := plugin.GetInfo()

		// Determine if core or external
		// Built-in plugins: claude-code, task-manager
		pluginType := "external"
		if isBuiltInPlugin(info.Name) {
			pluginType = "core"
			coreCount++
		} else {
			externalCount++
		}

		// Format: "  ✓ <name> (<type>)    - <description>"
		fmt.Printf("  ✓ %-20s (%s)   - %s\n", info.Name, pluginType, info.Description)
	}

	fmt.Println()
	fmt.Printf("Total: %d plugin(s) (%d core, %d external)\n", len(allPlugins), coreCount, externalCount)
}

// handlePluginReload reloads external plugins from config
func handlePluginReload(args []string) {
	// Parse flags
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printPluginReloadHelp()
		return
	}

	// Construct plugins.yaml path
	// DefaultDBPath is .darwinflow/logs/events.db, so we need to go up two levels to .darwinflow
	darwinflowDir := filepath.Dir(filepath.Dir(app.DefaultDBPath))
	pluginsConfigPath := filepath.Join(darwinflowDir, "plugins.yaml")

	// Check if config file exists
	if _, err := os.Stat(pluginsConfigPath); os.IsNotExist(err) {
		fmt.Printf("No plugins.yaml file found at: %s\n", pluginsConfigPath)
		fmt.Println("No external plugins to reload.")
		return
	}

	fmt.Printf("Reloading external plugins from %s...\n", pluginsConfigPath)

	// Create a temporary logger for plugin loading
	logger := infra.NewDefaultLogger()

	// Load external plugins
	loader := infra.NewPluginLoader(logger)
	externalPlugins, err := loader.LoadFromConfig(pluginsConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading plugins: %v\n", err)
		os.Exit(1)
	}

	// Display loaded plugins
	successCount := 0
	for _, plugin := range externalPlugins {
		info := plugin.GetInfo()
		fmt.Printf("✓ Loaded %s (external subprocess)\n", info.Name)
		successCount++
	}

	fmt.Printf("\nTotal: %d external plugin(s) reloaded\n", successCount)
	fmt.Println("\nNote: Plugins will be active on next command execution.")
}

// isBuiltInPlugin returns true if the plugin is a built-in core plugin
func isBuiltInPlugin(name string) bool {
	builtInPlugins := []string{"claude-code", "task-manager"}
	for _, builtIn := range builtInPlugins {
		if name == builtIn {
			return true
		}
	}
	return false
}

// printPluginCmdHelp prints help for the plugin command
func printPluginCmdHelp() {
	fmt.Println("Usage: dw plugin <subcommand> [options]")
	fmt.Println()
	fmt.Println("Manage DarwinFlow plugins")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  list      List all registered plugins (core and external)")
	fmt.Println("  reload    Reload external plugins from .darwinflow/plugins.yaml")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("For subcommand-specific help:")
	fmt.Println("  dw plugin list --help")
	fmt.Println("  dw plugin reload --help")
	fmt.Println()
}

// printPluginListHelp prints help for the plugin list command
func printPluginListHelp() {
	fmt.Println("Usage: dw plugin list")
	fmt.Println()
	fmt.Println("List all registered plugins (core and external)")
	fmt.Println()
	fmt.Println("This command shows:")
	fmt.Println("  - Plugin name")
	fmt.Println("  - Plugin type (core or external)")
	fmt.Println("  - Plugin description")
	fmt.Println("  - Total count by type")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  dw plugin list")
	fmt.Println()
}

// printPluginReloadHelp prints help for the plugin reload command
func printPluginReloadHelp() {
	fmt.Println("Usage: dw plugin reload")
	fmt.Println()
	fmt.Println("Reload external plugins from .darwinflow/plugins.yaml")
	fmt.Println()
	fmt.Println("This command:")
	fmt.Println("  1. Unregisters all currently loaded external plugins")
	fmt.Println("  2. Re-reads .darwinflow/plugins.yaml")
	fmt.Println("  3. Loads and registers plugins from the config")
	fmt.Println()
	fmt.Println("Note: Core plugins (claude-code, task-manager) are never unloaded")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  dw plugin reload")
	fmt.Println()
}

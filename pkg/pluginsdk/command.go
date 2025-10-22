package pluginsdk

import "context"

// Command represents a CLI command provided by a plugin.
// Commands are executed via `dw project <command> [args...]`.
type Command interface {
	// GetName returns the command name (used in CLI routing).
	// Example: "init", "refresh", "status"
	GetName() string

	// GetDescription returns a human-readable description of what the command does.
	// Used in help text and command listings.
	GetDescription() string

	// GetUsage returns usage information for the command.
	// Example: "init [--force]", "status <entity-id>"
	GetUsage() string

	// GetHelp returns detailed help text for the command.
	// This should include:
	// - Full description of what the command does
	// - Usage examples
	// - Available flags/options
	// - Any important notes or warnings
	// Return empty string to use default help (description + usage).
	GetHelp() string

	// Execute runs the command with the given arguments.
	// The CommandContext provides access to I/O streams and plugin context.
	// Arguments are passed as a string slice (similar to os.Args).
	// If args contains "--help" or "-h", the framework will display help automatically.
	Execute(ctx context.Context, cmdCtx CommandContext, args []string) error
}

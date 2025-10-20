package domain

import (
	"context"
	"errors"
	"io"
	"time"
)

// Common errors used throughout the plugin system
var (
	ErrNotFound          = errors.New("entity not found")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrAlreadyExists     = errors.New("already exists")
	ErrNotImplemented    = errors.New("not implemented")
	ErrInternal          = errors.New("internal error")
	ErrReadOnly          = errors.New("entity is read-only")
)

// Plugin is the base interface that ALL plugins must implement.
// It provides basic plugin metadata and declares what capabilities the plugin supports.
type Plugin interface {
	// GetInfo returns basic metadata about the plugin
	GetInfo() PluginInfo

	// GetCapabilities returns a list of capability interface names this plugin implements.
	// Examples: "IEntityProvider", "ICommandProvider", "IEventEmitter"
	// The framework uses this to route requests to the appropriate plugin methods.
	GetCapabilities() []string
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	// Name is the unique identifier for the plugin (e.g., "claude-code", "task-manager")
	Name string

	// Version is the semantic version of the plugin (e.g., "1.0.0")
	Version string

	// Description is a human-readable description of what the plugin does
	Description string

	// IsCore indicates whether this is a built-in plugin shipped with DarwinFlow.
	// Core plugins are loaded automatically, while external plugins are discovered.
	IsCore bool
}

// IEntityProvider is a plugin capability for providing queryable entities.
// Plugins that implement this can be queried for entities via the framework's registry.
type IEntityProvider interface {
	Plugin

	// GetEntityTypes returns metadata about all entity types this plugin provides
	GetEntityTypes() []EntityTypeInfo

	// Query returns entities matching the given query criteria
	Query(ctx context.Context, query EntityQuery) ([]IExtensible, error)

	// GetEntity retrieves a specific entity by ID
	GetEntity(ctx context.Context, entityID string) (IExtensible, error)
}

// IEntityUpdater is a plugin capability for supporting entity updates.
// It extends IEntityProvider with the ability to modify entities.
type IEntityUpdater interface {
	IEntityProvider

	// UpdateEntity modifies an entity's fields and returns the updated entity.
	// The fields map contains field names as keys and new values.
	UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (IExtensible, error)
}

// ICommandProvider is a plugin capability for providing CLI commands.
// Plugins that implement this can register commands accessible via `dw project <command>`.
type ICommandProvider interface {
	Plugin

	// GetCommands returns all commands provided by this plugin
	GetCommands() []Command
}

// IEventEmitter is a plugin capability for emitting real-time events.
// Plugins that implement this can stream events to the framework's event store.
type IEventEmitter interface {
	Plugin
	// Event emission is handled via PluginContext.EmitEvent() for built-in plugins
	// or stdout JSON streams for subprocess plugins.
}

// EntityTypeInfo describes an entity type provided by a plugin
type EntityTypeInfo struct {
	// Type is the unique identifier for this entity type (e.g., "session", "task")
	Type string

	// DisplayName is the human-readable singular name (e.g., "Claude Session", "Task")
	DisplayName string

	// DisplayNamePlural is the human-readable plural name (e.g., "Claude Sessions", "Tasks")
	DisplayNamePlural string

	// Capabilities is a list of entity capability interfaces this type implements
	// Examples: ["IExtensible", "ITrackable"]
	Capabilities []string

	// Icon is an optional emoji or symbol representing this entity type.
	// Used in UI displays.
	Icon string

	// Description is a human-readable description of this entity type
	Description string
}

// EntityQuery represents a query for entities from a plugin.
// Plugins receive this query and return matching entities.
type EntityQuery struct {
	// EntityType is the type of entities to query (e.g., "session", "task")
	EntityType string

	// Filters contains query filters as key-value pairs.
	// The structure and supported filters depend on the plugin and entity type.
	// Common filters: "status", "created_after", "tag", etc.
	Filters map[string]interface{}

	// Limit is the maximum number of entities to return.
	// 0 means no limit.
	Limit int

	// Offset is the number of entities to skip (for pagination)
	Offset int

	// SortBy specifies the field to sort results by.
	// Empty string means no specific sorting (plugin default).
	SortBy string

	// SortDesc indicates whether to sort in descending order.
	// False means ascending order.
	SortDesc bool
}

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

	// Execute runs the command with the given arguments.
	// The CommandContext provides access to I/O streams and plugin context.
	// Arguments are passed as a string slice (similar to os.Args).
	Execute(ctx context.Context, cmdCtx CommandContext, args []string) error
}

// PluginContext is the runtime context provided to plugins by the framework.
// It provides access to logging, working directory, and event emission.
type PluginContext interface {
	// GetLogger returns a logger for the plugin to use
	GetLogger() Logger

	// GetWorkingDir returns the current working directory of the DarwinFlow project
	GetWorkingDir() string

	// EmitEvent sends an event to the framework's event store.
	// This is the primary way plugins communicate events to the framework.
	EmitEvent(ctx context.Context, event PluginEvent) error
}

// CommandContext extends PluginContext with I/O streams for command execution.
// It is provided to commands when they are executed via the CLI.
type CommandContext interface {
	PluginContext

	// GetStdout returns the output stream for the command.
	// Commands should write their output here.
	GetStdout() io.Writer

	// GetStdin returns the input stream for the command.
	// Commands can read user input from here.
	GetStdin() io.Reader
}

// Logger is the interface for plugin logging.
// The framework provides an implementation that plugins use to log messages.
type Logger interface {
	// Debug logs a debug-level message.
	// Debug messages are only shown when debug logging is enabled.
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an info-level message.
	// Info messages are shown during normal operation.
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning-level message.
	// Warnings indicate potential issues that don't prevent operation.
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error-level message.
	// Errors indicate failures that may affect functionality.
	Error(msg string, keysAndValues ...interface{})
}

// PluginEvent is the standard event structure for plugin-emitted events.
// This is different from the main Event type which is for system events.
// Events are emitted by plugins and stored in the framework's event database.
type PluginEvent struct {
	// Type is the event type identifier (e.g., "tool.invoked", "task.created").
	// Use dot notation to namespace events: "<domain>.<action>".
	Type string

	// Source is the name of the plugin that emitted this event
	Source string

	// Timestamp is when the event occurred
	Timestamp time.Time

	// Payload contains the event-specific data.
	// Structure depends on the event type.
	Payload map[string]interface{}

	// Metadata contains additional context about the event.
	// Common fields: session_id, user_id, environment, etc.
	Metadata map[string]string
}

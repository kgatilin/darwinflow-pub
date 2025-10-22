package pluginsdk

import "encoding/json"

// JSON-RPC 2.0 Protocol Types
//
// This file defines the JSON-RPC protocol for external plugin communication.
// External plugins run as separate processes and communicate with the main
// process via newline-delimited JSON over stdin/stdout.
//
// Protocol specification:
// - Newline-delimited JSON messages
// - Request/response correlation by ID
// - Event emission via stdout with "event" field

// RPCRequest represents a JSON-RPC 2.0 request.
// External plugins receive these on stdin from the main process.
type RPCRequest struct {
	// JSONRPC is the protocol version (always "2.0")
	JSONRPC string `json:"jsonrpc"`

	// ID is a unique identifier for correlating request and response.
	// Can be string or number. Required for requests that expect a response.
	ID interface{} `json:"id,omitempty"`

	// Method is the RPC method name to invoke
	Method string `json:"method"`

	// Params contains method-specific parameters
	Params json.RawMessage `json:"params,omitempty"`
}

// RPCResponse represents a JSON-RPC 2.0 response.
// External plugins send these on stdout to the main process.
type RPCResponse struct {
	// JSONRPC is the protocol version (always "2.0")
	JSONRPC string `json:"jsonrpc"`

	// ID is the request ID this response corresponds to
	ID interface{} `json:"id"`

	// Result contains the method result (present on success)
	Result json.RawMessage `json:"result,omitempty"`

	// Error contains error information (present on failure)
	Error *RPCError `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error object.
type RPCError struct {
	// Code is the error code
	// Standard codes:
	//   -32700: Parse error (invalid JSON)
	//   -32600: Invalid request (malformed request object)
	//   -32601: Method not found
	//   -32602: Invalid params
	//   -32603: Internal error
	//   -32000 to -32099: Server-defined errors
	Code int `json:"code"`

	// Message is a short error description
	Message string `json:"message"`

	// Data contains additional error information (optional)
	Data interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	RPCErrorParseError     = -32700
	RPCErrorInvalidRequest = -32600
	RPCErrorMethodNotFound = -32601
	RPCErrorInvalidParams  = -32602
	RPCErrorInternal       = -32603
)

// RPCEvent represents an event emitted by the plugin to the main process.
// Events are sent on stdout with the "event" field to distinguish them
// from RPC responses.
type RPCEvent struct {
	// Event marker to distinguish from RPC responses (always "event")
	Event string `json:"event"`

	// Type is the event type (e.g., "tool.invoked", "task.created")
	Type string `json:"type"`

	// Source is the plugin name emitting the event
	Source string `json:"source"`

	// Timestamp is when the event occurred (ISO 8601 format)
	Timestamp string `json:"timestamp"`

	// Payload contains event-specific data
	Payload map[string]interface{} `json:"payload,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`

	// Version is the event schema version
	Version string `json:"version,omitempty"`
}

// RPC Method Names
//
// These constants define the standard RPC methods that external plugins
// can implement. The main process calls these methods via JSON-RPC.

const (
	// Core plugin methods (all plugins must implement)

	// RPCMethodInit initializes the plugin with configuration.
	// Request params: InitParams
	// Response result: (none)
	RPCMethodInit = "init"

	// RPCMethodGetInfo returns plugin metadata.
	// Request params: (none)
	// Response result: PluginInfo
	RPCMethodGetInfo = "get_info"

	// RPCMethodGetCapabilities returns plugin capability list.
	// Request params: (none)
	// Response result: []string (capability names)
	RPCMethodGetCapabilities = "get_capabilities"

	// IEntityProvider methods

	// RPCMethodGetEntityTypes returns entity type metadata.
	// Request params: (none)
	// Response result: []EntityTypeInfo
	RPCMethodGetEntityTypes = "get_entity_types"

	// RPCMethodQueryEntities queries entities.
	// Request params: EntityQuery
	// Response result: []map[string]interface{} (serialized IExtensible entities)
	RPCMethodQueryEntities = "query_entities"

	// RPCMethodGetEntity retrieves a specific entity by ID.
	// Request params: GetEntityParams { EntityID string }
	// Response result: map[string]interface{} (serialized IExtensible entity)
	RPCMethodGetEntity = "get_entity"

	// IEntityUpdater methods

	// RPCMethodUpdateEntity updates an entity's fields.
	// Request params: UpdateEntityParams { EntityID string, Fields map[string]interface{} }
	// Response result: map[string]interface{} (serialized updated entity)
	RPCMethodUpdateEntity = "update_entity"

	// ICommandProvider methods

	// RPCMethodGetCommands returns all commands provided by the plugin.
	// Request params: (none)
	// Response result: []CommandInfo (metadata about commands)
	RPCMethodGetCommands = "get_commands"

	// RPCMethodExecuteCommand executes a command.
	// Request params: ExecuteCommandParams { CommandName string, Args []string }
	// Response result: ExecuteCommandResult { ExitCode int, Output string, Error string }
	RPCMethodExecuteCommand = "execute_command"

	// IEventEmitter methods

	// RPCMethodStartEventStream starts the event stream.
	// Request params: (none)
	// Response result: (none)
	// After this call, the plugin should start sending RPCEvent messages on stdout.
	RPCMethodStartEventStream = "start_event_stream"

	// RPCMethodStopEventStream stops the event stream.
	// Request params: (none)
	// Response result: (none)
	RPCMethodStopEventStream = "stop_event_stream"
)

// RPC Parameter Types
//
// These types define the structure of parameters sent to RPC methods.

// InitParams contains initialization parameters for the plugin.
type InitParams struct {
	// WorkingDir is the current working directory
	WorkingDir string `json:"working_dir"`

	// Config contains plugin-specific configuration
	Config map[string]interface{} `json:"config,omitempty"`
}

// GetEntityParams contains parameters for get_entity method.
type GetEntityParams struct {
	// EntityID is the ID of the entity to retrieve
	EntityID string `json:"entity_id"`
}

// UpdateEntityParams contains parameters for update_entity method.
type UpdateEntityParams struct {
	// EntityID is the ID of the entity to update
	EntityID string `json:"entity_id"`

	// Fields contains the fields to update
	Fields map[string]interface{} `json:"fields"`
}

// ExecuteCommandParams contains parameters for execute_command method.
type ExecuteCommandParams struct {
	// CommandName is the name of the command to execute
	CommandName string `json:"command_name"`

	// Args are the command arguments
	Args []string `json:"args"`
}

// CommandInfo contains metadata about a command (serializable version of Command interface).
type CommandInfo struct {
	// Name is the command name
	Name string `json:"name"`

	// Description is a human-readable description
	Description string `json:"description"`

	// Usage is usage information
	Usage string `json:"usage"`

	// Help is detailed help text
	Help string `json:"help"`
}

// ExecuteCommandResult contains the result of command execution.
type ExecuteCommandResult struct {
	// ExitCode is the command exit code (0 for success)
	ExitCode int `json:"exit_code"`

	// Output is the command's stdout output
	Output string `json:"output"`

	// Error is the command's stderr output or error message
	Error string `json:"error,omitempty"`
}

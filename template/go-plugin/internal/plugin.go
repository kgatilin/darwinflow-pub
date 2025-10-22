package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ItemPlugin implements a simple item management plugin.
// It demonstrates the DarwinFlow external plugin protocol using JSON-RPC over stdin/stdout.
type ItemPlugin struct {
	// workingDir is the current working directory provided during initialization
	workingDir string

	// items stores all items in memory (keyed by ID)
	items map[string]*Item

	// eventStreaming indicates whether event emission is active
	eventStreaming bool
}

// NewItemPlugin creates a new ItemPlugin instance.
func NewItemPlugin() *ItemPlugin {
	return &ItemPlugin{
		items: make(map[string]*Item),
	}
}

// AddItem adds an item to the plugin's storage.
// This is a convenience method for initializing sample data.
func (p *ItemPlugin) AddItem(item *Item) {
	p.items[item.ID] = item
}

// Serve runs the JSON-RPC server loop.
// It reads newline-delimited JSON requests from stdin and writes responses to stdout.
// This method blocks until stdin is closed.
func (p *ItemPlugin) Serve() {
	// Create a buffered scanner for reading stdin
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 64KB initial, 1MB max

	// Read and process requests line by line
	for scanner.Scan() {
		var req pluginsdk.RPCRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			// Send parse error response
			p.sendError(req.ID, pluginsdk.RPCErrorParseError, "parse error: "+err.Error())
			continue
		}

		// Dispatch request to appropriate handler
		p.handleRequest(&req)
	}
}

// handleRequest routes an RPC request to the appropriate handler method.
func (p *ItemPlugin) handleRequest(req *pluginsdk.RPCRequest) {
	switch req.Method {
	case pluginsdk.RPCMethodInit:
		p.handleInit(req)
	case pluginsdk.RPCMethodGetInfo:
		p.handleGetInfo(req)
	case pluginsdk.RPCMethodGetCapabilities:
		p.handleGetCapabilities(req)
	case pluginsdk.RPCMethodGetEntityTypes:
		p.handleGetEntityTypes(req)
	case pluginsdk.RPCMethodQueryEntities:
		p.handleQueryEntities(req)
	case pluginsdk.RPCMethodGetEntity:
		p.handleGetEntity(req)
	case pluginsdk.RPCMethodUpdateEntity:
		p.handleUpdateEntity(req)
	case pluginsdk.RPCMethodStartEventStream:
		p.handleStartEventStream(req)
	case pluginsdk.RPCMethodStopEventStream:
		p.handleStopEventStream(req)
	default:
		p.sendError(req.ID, pluginsdk.RPCErrorMethodNotFound, "method not found: "+req.Method)
	}
}

// sendResult sends a successful RPC response.
// If result is nil, an empty result is sent.
func (p *ItemPlugin) sendResult(id interface{}, result interface{}) {
	var resultJSON json.RawMessage
	if result != nil {
		data, err := json.Marshal(result)
		if err != nil {
			p.sendError(id, pluginsdk.RPCErrorInternal, "failed to marshal result: "+err.Error())
			return
		}
		resultJSON = data
	}

	resp := pluginsdk.RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}

	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", string(data))
}

// sendError sends an RPC error response.
// Use standard error codes from pluginsdk (e.g., RPCErrorInvalidParams).
func (p *ItemPlugin) sendError(id interface{}, code int, message string) {
	resp := pluginsdk.RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &pluginsdk.RPCError{
			Code:    code,
			Message: message,
		},
	}

	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", string(data))
}

// emitEvent sends an event to the main process.
// Events are only sent when event streaming is active.
// The event is written to stdout with an "event" field to distinguish it from RPC responses.
func (p *ItemPlugin) emitEvent(eventType string, payload map[string]interface{}) {
	if !p.eventStreaming {
		return
	}

	event := pluginsdk.RPCEvent{
		Event:     "event",
		Type:      eventType,
		Source:    "myplugin",
		Timestamp: time.Now().Format(time.RFC3339),
		Payload:   payload,
	}

	data, _ := json.Marshal(event)
	fmt.Fprintf(os.Stdout, "%s\n", string(data))
}

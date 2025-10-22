package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// main implements a simple external plugin that provides note entities.
// This plugin runs as a separate process and communicates via JSON-RPC over stdin/stdout.
func main() {
	plugin := &NotesPlugin{
		notes: make(map[string]*Note),
	}

	// Create sample notes
	plugin.notes["note-1"] = &Note{
		ID:        "note-1",
		Title:     "Example Note",
		Content:   "This is an example note from the external plugin.",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	plugin.notes["note-2"] = &Note{
		ID:        "note-2",
		Title:     "Another Note",
		Content:   "External plugins can run in any language!",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// Run RPC server
	plugin.Serve()
}

// NotesPlugin implements a simple notes management plugin.
type NotesPlugin struct {
	workingDir    string
	notes         map[string]*Note
	eventStreaming bool
}

// Note represents a note entity.
type Note struct {
	ID        string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ToMap converts a Note to a map for JSON serialization.
func (n *Note) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":           n.ID,
		"type":         "note",
		"title":        n.Title,
		"content":      n.Content,
		"created_at":   n.CreatedAt.Format(time.RFC3339),
		"updated_at":   n.UpdatedAt.Format(time.RFC3339),
		"capabilities": []string{},
	}
}

// Serve runs the JSON-RPC server loop.
func (p *NotesPlugin) Serve() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 64KB initial, 1MB max

	for scanner.Scan() {
		var req pluginsdk.RPCRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			p.sendError(req.ID, pluginsdk.RPCErrorParseError, "parse error: "+err.Error())
			continue
		}

		p.handleRequest(&req)
	}
}

// handleRequest processes an RPC request.
func (p *NotesPlugin) handleRequest(req *pluginsdk.RPCRequest) {
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

// handleInit initializes the plugin.
func (p *NotesPlugin) handleInit(req *pluginsdk.RPCRequest) {
	var params pluginsdk.InitParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid params: "+err.Error())
		return
	}

	p.workingDir = params.WorkingDir
	p.sendResult(req.ID, nil)
}

// handleGetInfo returns plugin metadata.
func (p *NotesPlugin) handleGetInfo(req *pluginsdk.RPCRequest) {
	info := pluginsdk.PluginInfo{
		Name:        "notes-external",
		Version:     "1.0.0",
		Description: "External notes plugin (Go subprocess example)",
		IsCore:      false,
	}
	p.sendResult(req.ID, info)
}

// handleGetCapabilities returns supported capabilities.
func (p *NotesPlugin) handleGetCapabilities(req *pluginsdk.RPCRequest) {
	capabilities := []string{"IEntityProvider", "IEntityUpdater", "IEventEmitter"}
	p.sendResult(req.ID, capabilities)
}

// handleGetEntityTypes returns entity type metadata.
func (p *NotesPlugin) handleGetEntityTypes(req *pluginsdk.RPCRequest) {
	types := []pluginsdk.EntityTypeInfo{
		{
			Type:               "note",
			DisplayName:        "Note",
			DisplayNamePlural:  "Notes",
			Capabilities:       []string{},
			Icon:               "📝",
			Description:        "A text note from external plugin",
		},
	}
	p.sendResult(req.ID, types)
}

// handleQueryEntities queries notes.
func (p *NotesPlugin) handleQueryEntities(req *pluginsdk.RPCRequest) {
	var query pluginsdk.EntityQuery
	if err := json.Unmarshal(req.Params, &query); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid query: "+err.Error())
		return
	}

	if query.EntityType != "note" {
		p.sendResult(req.ID, []interface{}{})
		return
	}

	// Return all notes as maps
	notes := make([]map[string]interface{}, 0, len(p.notes))
	for _, note := range p.notes {
		notes = append(notes, note.ToMap())
	}

	// Apply limit if specified
	if query.Limit > 0 && len(notes) > query.Limit {
		notes = notes[:query.Limit]
	}

	p.sendResult(req.ID, notes)
}

// handleGetEntity retrieves a specific note.
func (p *NotesPlugin) handleGetEntity(req *pluginsdk.RPCRequest) {
	var params pluginsdk.GetEntityParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid params: "+err.Error())
		return
	}

	note, ok := p.notes[params.EntityID]
	if !ok {
		p.sendError(req.ID, -32000, "note not found")
		return
	}

	p.sendResult(req.ID, note.ToMap())
}

// handleUpdateEntity updates a note.
func (p *NotesPlugin) handleUpdateEntity(req *pluginsdk.RPCRequest) {
	var params pluginsdk.UpdateEntityParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid params: "+err.Error())
		return
	}

	note, ok := p.notes[params.EntityID]
	if !ok {
		p.sendError(req.ID, -32000, "note not found")
		return
	}

	// Update fields
	if title, ok := params.Fields["title"].(string); ok {
		note.Title = title
	}
	if content, ok := params.Fields["content"].(string); ok {
		note.Content = content
	}
	note.UpdatedAt = time.Now()

	// Emit update event if streaming
	if p.eventStreaming {
		p.emitEvent("note.updated", map[string]interface{}{
			"note_id": note.ID,
			"title":   note.Title,
		})
	}

	p.sendResult(req.ID, note.ToMap())
}

// handleStartEventStream starts event streaming.
func (p *NotesPlugin) handleStartEventStream(req *pluginsdk.RPCRequest) {
	p.eventStreaming = true
	p.sendResult(req.ID, nil)

	// Emit initial event
	p.emitEvent("stream.started", map[string]interface{}{
		"note_count": len(p.notes),
	})
}

// handleStopEventStream stops event streaming.
func (p *NotesPlugin) handleStopEventStream(req *pluginsdk.RPCRequest) {
	p.eventStreaming = false
	p.sendResult(req.ID, nil)
}

// sendResult sends a successful RPC response.
func (p *NotesPlugin) sendResult(id interface{}, result interface{}) {
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
func (p *NotesPlugin) sendError(id interface{}, code int, message string) {
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
func (p *NotesPlugin) emitEvent(eventType string, payload map[string]interface{}) {
	event := pluginsdk.RPCEvent{
		Event:     "event",
		Type:      eventType,
		Source:    "notes-external",
		Timestamp: time.Now().Format(time.RFC3339),
		Payload:   payload,
	}

	data, _ := json.Marshal(event)
	fmt.Fprintf(os.Stdout, "%s\n", string(data))
}

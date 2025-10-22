package internal

import (
	"encoding/json"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// handleInit initializes the plugin with configuration.
// This is called once when the plugin starts, providing the working directory
// and optional configuration parameters.
func (p *ItemPlugin) handleInit(req *pluginsdk.RPCRequest) {
	var params pluginsdk.InitParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid params: "+err.Error())
		return
	}

	p.workingDir = params.WorkingDir
	p.sendResult(req.ID, nil)
}

// handleGetInfo returns plugin metadata.
// This provides basic information about the plugin for display and identification.
func (p *ItemPlugin) handleGetInfo(req *pluginsdk.RPCRequest) {
	info := pluginsdk.PluginInfo{
		Name:        "myplugin",
		Version:     "1.0.0",
		Description: "Example external plugin template",
		IsCore:      false,
	}
	p.sendResult(req.ID, info)
}

// handleGetCapabilities returns the list of capabilities this plugin supports.
// Capabilities determine which optional interfaces are implemented:
//   - IEntityProvider: Query and get entities
//   - IEntityUpdater: Update entity fields
//   - IEventEmitter: Emit events when changes occur
//   - ICommandProvider: Provide CLI commands (not implemented in this template)
func (p *ItemPlugin) handleGetCapabilities(req *pluginsdk.RPCRequest) {
	capabilities := []string{
		"IEntityProvider",
		"IEntityUpdater",
		"IEventEmitter",
	}
	p.sendResult(req.ID, capabilities)
}

// handleGetEntityTypes returns metadata about entity types provided by this plugin.
// Each entity type includes display information and optional features.
func (p *ItemPlugin) handleGetEntityTypes(req *pluginsdk.RPCRequest) {
	types := []pluginsdk.EntityTypeInfo{
		{
			Type:              "item",
			DisplayName:       "Item",
			DisplayNamePlural: "Items",
			Capabilities:      []string{}, // Add entity-level capabilities if needed
			Icon:              "📦",
			Description:       "A generic item entity",
		},
	}
	p.sendResult(req.ID, types)
}

// handleQueryEntities queries items based on filters and pagination.
// This implements the IEntityProvider capability.
func (p *ItemPlugin) handleQueryEntities(req *pluginsdk.RPCRequest) {
	var query pluginsdk.EntityQuery
	if err := json.Unmarshal(req.Params, &query); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid query: "+err.Error())
		return
	}

	// Only handle queries for "item" entity type
	if query.EntityType != "item" {
		p.sendResult(req.ID, []interface{}{})
		return
	}

	// Convert all items to maps for serialization
	items := make([]map[string]interface{}, 0, len(p.items))
	for _, item := range p.items {
		items = append(items, item.ToMap())
	}

	// Apply pagination limit if specified
	if query.Limit > 0 && len(items) > query.Limit {
		items = items[:query.Limit]
	}

	p.sendResult(req.ID, items)
}

// handleGetEntity retrieves a specific item by ID.
// This implements the IEntityProvider capability.
func (p *ItemPlugin) handleGetEntity(req *pluginsdk.RPCRequest) {
	var params pluginsdk.GetEntityParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid params: "+err.Error())
		return
	}

	// Look up the item by ID
	item, ok := p.items[params.EntityID]
	if !ok {
		p.sendError(req.ID, -32000, "item not found")
		return
	}

	p.sendResult(req.ID, item.ToMap())
}

// handleUpdateEntity updates an item's fields.
// This implements the IEntityUpdater capability.
// After updating, it emits an "item.updated" event if event streaming is active.
func (p *ItemPlugin) handleUpdateEntity(req *pluginsdk.RPCRequest) {
	var params pluginsdk.UpdateEntityParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		p.sendError(req.ID, pluginsdk.RPCErrorInvalidParams, "invalid params: "+err.Error())
		return
	}

	// Look up the item by ID
	item, ok := p.items[params.EntityID]
	if !ok {
		p.sendError(req.ID, -32000, "item not found")
		return
	}

	// Update fields if provided
	if name, ok := params.Fields["name"].(string); ok {
		item.Name = name
	}
	if description, ok := params.Fields["description"].(string); ok {
		item.Description = description
	}
	if tags, ok := params.Fields["tags"].([]interface{}); ok {
		// Convert []interface{} to []string
		item.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				item.Tags = append(item.Tags, tagStr)
			}
		}
	}

	// Update the timestamp
	item.UpdatedAt = time.Now()

	// Emit update event if streaming is active
	if p.eventStreaming {
		p.emitEvent("item.updated", map[string]interface{}{
			"item_id": item.ID,
			"name":    item.Name,
		})
	}

	p.sendResult(req.ID, item.ToMap())
}

// handleStartEventStream starts event streaming.
// This implements the IEventEmitter capability.
// After starting, the plugin will emit events when entities are modified.
func (p *ItemPlugin) handleStartEventStream(req *pluginsdk.RPCRequest) {
	p.eventStreaming = true
	p.sendResult(req.ID, nil)

	// Emit initial event to confirm streaming started
	p.emitEvent("stream.started", map[string]interface{}{
		"item_count": len(p.items),
	})
}

// handleStopEventStream stops event streaming.
// This implements the IEventEmitter capability.
func (p *ItemPlugin) handleStopEventStream(req *pluginsdk.RPCRequest) {
	p.eventStreaming = false
	p.sendResult(req.ID, nil)
}

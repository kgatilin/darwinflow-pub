package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// PluginRegistry manages all registered plugins and routes operations to them.
type PluginRegistry struct {
	plugins    map[string]domain.Plugin // key: plugin name
	entityMap  map[string]string        // key: entity type, value: plugin name
	logger     Logger
	mu         sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(logger Logger) *PluginRegistry {
	return &PluginRegistry{
		plugins:   make(map[string]domain.Plugin),
		entityMap: make(map[string]string),
		logger:    logger,
	}
}

// RegisterPlugin registers a plugin with the system.
// Returns error if plugin name already exists or entity type conflicts.
func (r *PluginRegistry) RegisterPlugin(plugin domain.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := plugin.GetInfo()

	// Check for duplicate plugin name
	if _, exists := r.plugins[info.Name]; exists {
		return fmt.Errorf("plugin already registered: %s", info.Name)
	}

	// Check for entity type conflicts
	entityTypes := plugin.GetEntityTypes()
	for _, et := range entityTypes {
		if existingPlugin, exists := r.entityMap[et.Type]; exists {
			return fmt.Errorf("entity type %s already provided by plugin %s", et.Type, existingPlugin)
		}
	}

	// Register plugin
	r.plugins[info.Name] = plugin

	// Map entity types to plugin
	for _, et := range entityTypes {
		r.entityMap[et.Type] = info.Name
	}

	r.logger.Info("Registered plugin: %s (version %s)", info.Name, info.Version)
	for _, et := range entityTypes {
		r.logger.Debug("  - Entity type: %s (capabilities: %v)", et.Type, et.Capabilities)
	}

	return nil
}

// GetPlugin retrieves a plugin by name
func (r *PluginRegistry) GetPlugin(name string) (domain.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// GetPluginForEntityType retrieves the plugin that provides a given entity type
func (r *PluginRegistry) GetPluginForEntityType(entityType string) (domain.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pluginName, exists := r.entityMap[entityType]
	if !exists {
		return nil, fmt.Errorf("no plugin found for entity type: %s", entityType)
	}

	return r.plugins[pluginName], nil
}

// GetAllPlugins returns all registered plugins
func (r *PluginRegistry) GetAllPlugins() []domain.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]domain.Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// GetAllEntityTypes returns all entity types from all plugins
func (r *PluginRegistry) GetAllEntityTypes() []domain.EntityTypeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var entityTypes []domain.EntityTypeInfo
	for _, plugin := range r.plugins {
		entityTypes = append(entityTypes, plugin.GetEntityTypes()...)
	}

	return entityTypes
}

// Query executes a query across one or more plugins
func (r *PluginRegistry) Query(ctx context.Context, query domain.EntityQuery) ([]domain.IExtensible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If entity type is specified, route to specific plugin
	if query.EntityType != "" {
		pluginName, exists := r.entityMap[query.EntityType]
		if !exists {
			return nil, fmt.Errorf("no plugin found for entity type: %s", query.EntityType)
		}

		plugin := r.plugins[pluginName]
		return plugin.Query(ctx, query)
	}

	// Otherwise, query all plugins and combine results
	var allEntities []domain.IExtensible
	for _, plugin := range r.plugins {
		entities, err := plugin.Query(ctx, query)
		if err != nil {
			r.logger.Warn("Plugin %s query failed: %v", plugin.GetInfo().Name, err)
			continue
		}
		allEntities = append(allEntities, entities...)
	}

	return allEntities, nil
}

// GetEntity retrieves a single entity by ID.
// If entity type is known, specify it in the ID format: "type:id"
// Otherwise, searches all plugins (slower).
func (r *PluginRegistry) GetEntity(ctx context.Context, entityID string) (domain.IExtensible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try each plugin until we find the entity
	for _, plugin := range r.plugins {
		entity, err := plugin.GetEntity(ctx, entityID)
		if err == nil {
			return entity, nil
		}
	}

	return nil, fmt.Errorf("entity not found: %s", entityID)
}

// UpdateEntity updates an entity's fields
func (r *PluginRegistry) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (domain.IExtensible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try each plugin until we find the entity
	for _, plugin := range r.plugins {
		entity, err := plugin.UpdateEntity(ctx, entityID, fields)
		if err == nil {
			return entity, nil
		}
	}

	return nil, fmt.Errorf("entity not found: %s", entityID)
}

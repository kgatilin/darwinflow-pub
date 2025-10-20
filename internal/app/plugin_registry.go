package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// PluginRegistry manages all registered plugins and routes operations to them.
// It uses SDK plugin interfaces and adapts between SDK and domain types.
// Routing is capability-based: plugins declare capabilities, registry routes accordingly.
type PluginRegistry struct {
	plugins          map[string]domain.Plugin       // key: plugin name (uses SDK interface)
	entityProviders  map[string]domain.IEntityProvider  // key: entity type, value: provider
	commandProviders map[string]domain.ICommandProvider // key: plugin name, value: provider
	eventEmitters    []domain.IEventEmitter
	entityUpdaters   map[string]domain.IEntityUpdater // key: entity type, value: updater
	logger           Logger
	mu               sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(logger Logger) *PluginRegistry {
	return &PluginRegistry{
		plugins:          make(map[string]domain.Plugin),
		entityProviders:  make(map[string]domain.IEntityProvider),
		commandProviders: make(map[string]domain.ICommandProvider),
		eventEmitters:    make([]domain.IEventEmitter, 0),
		entityUpdaters:   make(map[string]domain.IEntityUpdater),
		logger:           logger,
	}
}

// RegisterPlugin registers a plugin with the system.
// Accepts plugins implementing the SDK Plugin interface.
// Returns error if plugin name already exists or entity type conflicts.
// Uses capability-based routing to map plugins to their provided capabilities.
func (r *PluginRegistry) RegisterPlugin(plugin domain.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := plugin.GetInfo()

	// Check for duplicate plugin name
	if _, exists := r.plugins[info.Name]; exists {
		return fmt.Errorf("plugin already registered: %s", info.Name)
	}

	// Get plugin capabilities
	capabilities := plugin.GetCapabilities()

	// Route based on capabilities
	if contains(capabilities, "IEntityProvider") {
		entityProvider, ok := plugin.(domain.IEntityProvider)
		if !ok {
			return fmt.Errorf("plugin %s declares IEntityProvider capability but doesn't implement it", info.Name)
		}

		entityTypes := entityProvider.GetEntityTypes()
		// Check for entity type conflicts
		for _, et := range entityTypes {
			if existingProvider, exists := r.entityProviders[et.Type]; exists {
				existingInfo := existingProvider.(domain.Plugin).GetInfo()
				return fmt.Errorf("entity type %s already provided by plugin %s", et.Type, existingInfo.Name)
			}
		}

		// Map entity types to provider
		for _, et := range entityTypes {
			r.entityProviders[et.Type] = entityProvider
			r.logger.Debug("  - Entity type: %s (capabilities: %v)", et.Type, et.Capabilities)
		}
	}

	if contains(capabilities, "ICommandProvider") {
		commandProvider, ok := plugin.(domain.ICommandProvider)
		if !ok {
			return fmt.Errorf("plugin %s declares ICommandProvider capability but doesn't implement it", info.Name)
		}
		r.commandProviders[info.Name] = commandProvider
	}

	if contains(capabilities, "IEventEmitter") {
		eventEmitter, ok := plugin.(domain.IEventEmitter)
		if !ok {
			return fmt.Errorf("plugin %s declares IEventEmitter capability but doesn't implement it", info.Name)
		}
		r.eventEmitters = append(r.eventEmitters, eventEmitter)
	}

	if contains(capabilities, "IEntityUpdater") {
		entityUpdater, ok := plugin.(domain.IEntityUpdater)
		if !ok {
			return fmt.Errorf("plugin %s declares IEntityUpdater capability but doesn't implement it", info.Name)
		}

		entityTypes := entityUpdater.GetEntityTypes()
		// Map entity types to updater
		for _, et := range entityTypes {
			r.entityUpdaters[et.Type] = entityUpdater
		}
	}

	// Register plugin
	r.plugins[info.Name] = plugin
	r.logger.Info("Registered plugin: %s (version %s) with capabilities: %v", info.Name, info.Version, capabilities)

	return nil
}

// GetPlugin retrieves a plugin by name (returns SDK plugin)
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

	provider, exists := r.entityProviders[entityType]
	if !exists {
		return nil, fmt.Errorf("no plugin found for entity type: %s", entityType)
	}

	// IEntityProvider extends Plugin, so we can cast safely
	return provider.(domain.Plugin), nil
}

// GetAllPlugins returns all registered plugins (SDK plugins)
func (r *PluginRegistry) GetAllPlugins() []domain.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]domain.Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// GetPluginInfos returns metadata for all registered plugins (adapted to domain)
func (r *PluginRegistry) GetPluginInfos() []domain.PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]domain.PluginInfo, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		sdkInfo := plugin.GetInfo()
		infos = append(infos, domain.PluginInfo{
			Name:        sdkInfo.Name,
			Version:     sdkInfo.Version,
			Description: sdkInfo.Description,
			IsCore:      sdkInfo.IsCore,
		})
	}

	return infos
}

// GetAllEntityTypes returns all entity types from all plugins (adapted to domain)
func (r *PluginRegistry) GetAllEntityTypes() []domain.EntityTypeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var entityTypes []domain.EntityTypeInfo
	for _, plugin := range r.plugins {
		// Only get entity types from plugins that provide entities
		if entityProvider, ok := plugin.(domain.IEntityProvider); ok {
			sdkTypes := entityProvider.GetEntityTypes()
			entityTypes = append(entityTypes, adaptEntityTypeInfos(sdkTypes)...)
		}
	}

	return entityTypes
}

// Query executes a query across one or more plugins (adapts between domain and SDK)
func (r *PluginRegistry) Query(ctx context.Context, query domain.EntityQuery) ([]domain.IExtensible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Convert domain query to SDK query
	sdkQuery := adaptEntityQuery(query)

	// If entity type is specified, route to specific provider
	if query.EntityType != "" {
		provider, exists := r.entityProviders[query.EntityType]
		if !exists {
			return nil, fmt.Errorf("no provider for entity type: %s", query.EntityType)
		}

		sdkEntities, err := provider.Query(ctx, sdkQuery)
		if err != nil {
			return nil, err
		}

		// Adapt SDK entities to domain entities
		return adaptEntities(sdkEntities), nil
	}

	// Otherwise, query all entity providers and combine results
	var allEntities []domain.IExtensible
	for _, provider := range r.entityProviders {
		sdkEntities, err := provider.Query(ctx, sdkQuery)
		if err != nil {
			pluginInfo := provider.(domain.Plugin).GetInfo()
			r.logger.Warn("Plugin %s query failed: %v", pluginInfo.Name, err)
			continue
		}

		// Adapt and append
		allEntities = append(allEntities, adaptEntities(sdkEntities)...)
	}

	return allEntities, nil
}

// GetEntity retrieves a single entity by ID.
// Searches all entity providers until the entity is found.
func (r *PluginRegistry) GetEntity(ctx context.Context, entityID string) (domain.IExtensible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try each entity provider until we find the entity
	for _, provider := range r.entityProviders {
		sdkEntity, err := provider.GetEntity(ctx, entityID)
		if err == nil {
			// Adapt SDK entity to domain entity
			return newEntityAdapter(sdkEntity), nil
		}
	}

	return nil, fmt.Errorf("entity not found: %s", entityID)
}

// UpdateEntity updates an entity's fields
func (r *PluginRegistry) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (domain.IExtensible, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try each entity updater until we find and update the entity
	for _, updater := range r.entityUpdaters {
		sdkEntity, err := updater.UpdateEntity(ctx, entityID, fields)
		if err == nil {
			// Adapt SDK entity to domain entity
			return newEntityAdapter(sdkEntity), nil
		}
	}

	return nil, fmt.Errorf("entity not found or not updatable: %s", entityID)
}

// GetCommandProvider retrieves a command provider for a plugin
func (r *PluginRegistry) GetCommandProvider(pluginName string) (domain.ICommandProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.commandProviders[pluginName]
	if !exists {
		return nil, fmt.Errorf("no command provider for plugin: %s", pluginName)
	}

	return provider, nil
}

// GetAllCommandProviders returns all registered command providers
func (r *PluginRegistry) GetAllCommandProviders() []domain.ICommandProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]domain.ICommandProvider, 0, len(r.commandProviders))
	for _, provider := range r.commandProviders {
		providers = append(providers, provider)
	}

	return providers
}

// GetEventEmitters returns all registered event emitters
func (r *PluginRegistry) GetEventEmitters() []domain.IEventEmitter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid concurrent modification
	emitters := make([]domain.IEventEmitter, len(r.eventEmitters))
	copy(emitters, r.eventEmitters)
	return emitters
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

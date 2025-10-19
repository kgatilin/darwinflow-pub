package domain

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrNotFound = errors.New("entity not found")
)

// Plugin represents a plugin that provides entities to the system.
// Plugins can be built-in (core) or external (loaded from project directory).
type Plugin interface {
	// GetInfo returns metadata about this plugin
	GetInfo() PluginInfo

	// GetEntityTypes returns the types of entities this plugin provides
	GetEntityTypes() []EntityTypeInfo

	// Query returns entities matching the given query
	Query(ctx context.Context, query EntityQuery) ([]IExtensible, error)

	// GetEntity retrieves a single entity by ID
	GetEntity(ctx context.Context, entityID string) (IExtensible, error)

	// UpdateEntity updates an entity's fields
	UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (IExtensible, error)
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	// Name is the unique identifier for this plugin (e.g., "claude-code")
	Name string

	// Version is the plugin version
	Version string

	// Description is a human-readable description
	Description string

	// IsCore indicates if this is a built-in core plugin
	IsCore bool
}

// EntityTypeInfo describes an entity type provided by a plugin
type EntityTypeInfo struct {
	// Type is the entity type name (e.g., "session", "task")
	Type string

	// DisplayName is the human-readable name (e.g., "Claude Session")
	DisplayName string

	// DisplayNamePlural is the plural form (e.g., "Claude Sessions")
	DisplayNamePlural string

	// Capabilities lists the capability interfaces this entity type supports
	Capabilities []string

	// Icon is an optional emoji or symbol for UI display
	Icon string
}

// EntityQuery specifies criteria for querying entities
type EntityQuery struct {
	// EntityType filters by entity type (empty = all types from plugin)
	EntityType string

	// Capabilities filters to entities supporting ALL listed capabilities
	Capabilities []string

	// Filters are field-level filters (field name -> expected value)
	Filters map[string]interface{}

	// Limit restricts the number of results (0 = no limit)
	Limit int

	// Offset for pagination
	Offset int

	// OrderBy specifies sort field (empty = plugin default order)
	OrderBy string

	// OrderDesc reverses sort order
	OrderDesc bool
}

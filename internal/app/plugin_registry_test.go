package app_test

import (
	"context"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// MockPlugin is a test plugin
type MockPlugin struct {
	name         string
	version      string
	entityTypes  []domain.EntityTypeInfo
	entities     []domain.IExtensible
	queryError   error
	getError     error
	updateError  error
}

func NewMockPlugin(name string, entityTypes []domain.EntityTypeInfo) *MockPlugin {
	return &MockPlugin{
		name:        name,
		version:     "1.0.0",
		entityTypes: entityTypes,
		entities:    []domain.IExtensible{},
	}
}

func (p *MockPlugin) GetInfo() domain.PluginInfo {
	return domain.PluginInfo{
		Name:        p.name,
		Version:     p.version,
		Description: "Mock plugin for testing",
		IsCore:      true,
	}
}

func (p *MockPlugin) GetEntityTypes() []domain.EntityTypeInfo {
	return p.entityTypes
}

func (p *MockPlugin) Query(ctx context.Context, query domain.EntityQuery) ([]domain.IExtensible, error) {
	if p.queryError != nil {
		return nil, p.queryError
	}
	return p.entities, nil
}

func (p *MockPlugin) GetEntity(ctx context.Context, entityID string) (domain.IExtensible, error) {
	if p.getError != nil {
		return nil, p.getError
	}

	for _, e := range p.entities {
		if e.GetID() == entityID {
			return e, nil
		}
	}

	return nil, domain.ErrNotFound
}

func (p *MockPlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (domain.IExtensible, error) {
	if p.updateError != nil {
		return nil, p.updateError
	}
	return p.GetEntity(ctx, entityID)
}

// MockEntity is a test entity
type MockEntity struct {
	id           string
	entityType   string
	capabilities []string
	fields       map[string]interface{}
}

func NewMockEntity(id, entityType string, capabilities []string) *MockEntity {
	return &MockEntity{
		id:           id,
		entityType:   entityType,
		capabilities: capabilities,
		fields:       make(map[string]interface{}),
	}
}

func (e *MockEntity) GetID() string {
	return e.id
}

func (e *MockEntity) GetType() string {
	return e.entityType
}

func (e *MockEntity) GetCapabilities() []string {
	return e.capabilities
}

func (e *MockEntity) GetField(name string) interface{} {
	return e.fields[name]
}

func (e *MockEntity) GetAllFields() map[string]interface{} {
	return e.fields
}

func TestPluginRegistry_RegisterPlugin(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []domain.EntityTypeInfo{
		{
			Type:         "task",
			DisplayName:  "Task",
			Capabilities: []string{"IExtensible"},
		},
	}

	plugin := NewMockPlugin("test-plugin", entityTypes)

	err := registry.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Try to get the plugin
	retrieved, err := registry.GetPlugin("test-plugin")
	if err != nil {
		t.Fatalf("Failed to get plugin: %v", err)
	}

	if retrieved.GetInfo().Name != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got %s", retrieved.GetInfo().Name)
	}
}

func TestPluginRegistry_RegisterPlugin_Duplicate(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []domain.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	plugin1 := NewMockPlugin("test-plugin", entityTypes)
	plugin2 := NewMockPlugin("test-plugin", entityTypes)

	err := registry.RegisterPlugin(plugin1)
	if err != nil {
		t.Fatalf("Failed to register first plugin: %v", err)
	}

	err = registry.RegisterPlugin(plugin2)
	if err == nil {
		t.Error("Expected error when registering duplicate plugin, got nil")
	}
}

func TestPluginRegistry_RegisterPlugin_EntityTypeConflict(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []domain.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	plugin1 := NewMockPlugin("plugin1", entityTypes)
	plugin2 := NewMockPlugin("plugin2", entityTypes) // Same entity type

	err := registry.RegisterPlugin(plugin1)
	if err != nil {
		t.Fatalf("Failed to register first plugin: %v", err)
	}

	err = registry.RegisterPlugin(plugin2)
	if err == nil {
		t.Error("Expected error when registering plugin with conflicting entity type, got nil")
	}
}

func TestPluginRegistry_GetPluginForEntityType(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []domain.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	plugin := NewMockPlugin("test-plugin", entityTypes)
	err := registry.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	retrieved, err := registry.GetPluginForEntityType("task")
	if err != nil {
		t.Fatalf("Failed to get plugin for entity type: %v", err)
	}

	if retrieved.GetInfo().Name != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got %s", retrieved.GetInfo().Name)
	}
}

func TestPluginRegistry_Query(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []domain.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	plugin := NewMockPlugin("test-plugin", entityTypes)
	plugin.entities = []domain.IExtensible{
		NewMockEntity("task-1", "task", []string{"IExtensible"}),
		NewMockEntity("task-2", "task", []string{"IExtensible"}),
	}

	err := registry.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Query for entities
	ctx := context.Background()
	entities, err := registry.Query(ctx, domain.EntityQuery{
		EntityType: "task",
	})

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(entities) != 2 {
		t.Errorf("Expected 2 entities, got %d", len(entities))
	}
}

func TestPluginRegistry_GetAllEntityTypes(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	plugin1 := NewMockPlugin("plugin1", []domain.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	})

	plugin2 := NewMockPlugin("plugin2", []domain.EntityTypeInfo{
		{Type: "note", DisplayName: "Note", Capabilities: []string{"IExtensible"}},
	})

	registry.RegisterPlugin(plugin1)
	registry.RegisterPlugin(plugin2)

	entityTypes := registry.GetAllEntityTypes()

	if len(entityTypes) != 2 {
		t.Errorf("Expected 2 entity types, got %d", len(entityTypes))
	}

	// Check that both types are present
	types := make(map[string]bool)
	for _, et := range entityTypes {
		types[et.Type] = true
	}

	if !types["task"] || !types["note"] {
		t.Error("Expected both 'task' and 'note' entity types")
	}
}

package app_test

import (
	"context"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// MockPlugin is a test plugin implementing SDK interfaces
type MockPlugin struct {
	name         string
	version      string
	capabilities []string
	entityTypes  []pluginsdk.EntityTypeInfo
	entities     []pluginsdk.IExtensible
	queryError   error
	getError     error
	updateError  error
}

func NewMockPlugin(name string, entityTypes []pluginsdk.EntityTypeInfo) *MockPlugin {
	return &MockPlugin{
		name:         name,
		version:      "1.0.0",
		capabilities: []string{"IEntityProvider", "IEntityUpdater"},
		entityTypes:  entityTypes,
		entities:     []pluginsdk.IExtensible{},
	}
}

func (p *MockPlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        p.name,
		Version:     p.version,
		Description: "Mock plugin for testing",
		IsCore:      true,
	}
}

func (p *MockPlugin) GetCapabilities() []string {
	return p.capabilities
}

func (p *MockPlugin) GetEntityTypes() []pluginsdk.EntityTypeInfo {
	return p.entityTypes
}

func (p *MockPlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
	if p.queryError != nil {
		return nil, p.queryError
	}
	return p.entities, nil
}

func (p *MockPlugin) GetEntity(ctx context.Context, entityID string) (pluginsdk.IExtensible, error) {
	if p.getError != nil {
		return nil, p.getError
	}

	for _, e := range p.entities {
		if e.GetID() == entityID {
			return e, nil
		}
	}

	return nil, pluginsdk.ErrNotFound
}

func (p *MockPlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (pluginsdk.IExtensible, error) {
	if p.updateError != nil {
		return nil, p.updateError
	}
	return p.GetEntity(ctx, entityID)
}

// MockEntity is a test entity implementing SDK interfaces
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

	entityTypes := []pluginsdk.EntityTypeInfo{
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

	entityTypes := []pluginsdk.EntityTypeInfo{
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

	entityTypes := []pluginsdk.EntityTypeInfo{
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

	entityTypes := []pluginsdk.EntityTypeInfo{
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

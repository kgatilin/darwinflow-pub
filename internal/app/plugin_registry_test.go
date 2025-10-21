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

func TestPluginRegistry_GetPluginInfos(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes1 := []pluginsdk.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}
	entityTypes2 := []pluginsdk.EntityTypeInfo{
		{Type: "note", DisplayName: "Note", Capabilities: []string{"IExtensible"}},
	}

	plugin1 := NewMockPlugin("plugin1", entityTypes1)
	plugin2 := NewMockPlugin("plugin2", entityTypes2)

	registry.RegisterPlugin(plugin1)
	registry.RegisterPlugin(plugin2)

	infos := registry.GetPluginInfos()

	if len(infos) != 2 {
		t.Errorf("Expected 2 plugin infos, got %d", len(infos))
	}
}

func TestPluginRegistry_GetAllEntityTypes(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []pluginsdk.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
		{Type: "note", DisplayName: "Note", Capabilities: []string{"IExtensible"}},
	}

	plugin := NewMockPlugin("test-plugin", entityTypes)
	registry.RegisterPlugin(plugin)

	allTypes := registry.GetAllEntityTypes()

	if len(allTypes) != 2 {
		t.Errorf("Expected 2 entity types, got %d", len(allTypes))
	}
}

func TestPluginRegistry_Query(t *testing.T) {
	ctx := context.Background()
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []pluginsdk.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	entity := NewMockEntity("entity-1", "task", []string{"IExtensible"})
	plugin := NewMockPlugin("test-plugin", entityTypes)
	plugin.entities = []pluginsdk.IExtensible{entity}

	registry.RegisterPlugin(plugin)

	query := pluginsdk.EntityQuery{
		EntityType: "task",
	}

	results, err := registry.Query(ctx, query)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].GetID() != "entity-1" {
		t.Errorf("Expected entity ID 'entity-1', got %s", results[0].GetID())
	}
}

func TestPluginRegistry_Query_AllProviders(t *testing.T) {
	ctx := context.Background()
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes1 := []pluginsdk.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}
	entityTypes2 := []pluginsdk.EntityTypeInfo{
		{Type: "note", DisplayName: "Note", Capabilities: []string{"IExtensible"}},
	}

	entity1 := NewMockEntity("entity-1", "task", []string{"IExtensible"})
	entity2 := NewMockEntity("entity-2", "note", []string{"IExtensible"})

	plugin1 := NewMockPlugin("plugin1", entityTypes1)
	plugin1.entities = []pluginsdk.IExtensible{entity1}

	plugin2 := NewMockPlugin("plugin2", entityTypes2)
	plugin2.entities = []pluginsdk.IExtensible{entity2}

	registry.RegisterPlugin(plugin1)
	registry.RegisterPlugin(plugin2)

	// Query without entity type should return all
	query := pluginsdk.EntityQuery{}

	results, err := registry.Query(ctx, query)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestPluginRegistry_GetEntity(t *testing.T) {
	ctx := context.Background()
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []pluginsdk.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	entity := NewMockEntity("entity-123", "task", []string{"IExtensible"})
	plugin := NewMockPlugin("test-plugin", entityTypes)
	plugin.entities = []pluginsdk.IExtensible{entity}

	registry.RegisterPlugin(plugin)

	retrieved, err := registry.GetEntity(ctx, "entity-123")
	if err != nil {
		t.Fatalf("GetEntity failed: %v", err)
	}

	if retrieved.GetID() != "entity-123" {
		t.Errorf("Expected entity ID 'entity-123', got %s", retrieved.GetID())
	}
}

func TestPluginRegistry_GetEntity_NotFound(t *testing.T) {
	ctx := context.Background()
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []pluginsdk.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	plugin := NewMockPlugin("test-plugin", entityTypes)
	registry.RegisterPlugin(plugin)

	_, err := registry.GetEntity(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when entity not found")
	}
}

func TestPluginRegistry_UpdateEntity(t *testing.T) {
	ctx := context.Background()
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	entityTypes := []pluginsdk.EntityTypeInfo{
		{Type: "task", DisplayName: "Task", Capabilities: []string{"IExtensible"}},
	}

	entity := NewMockEntity("entity-123", "task", []string{"IExtensible"})
	plugin := NewMockPlugin("test-plugin", entityTypes)
	plugin.entities = []pluginsdk.IExtensible{entity}

	registry.RegisterPlugin(plugin)

	fields := map[string]interface{}{
		"title": "Updated Title",
	}

	updated, err := registry.UpdateEntity(ctx, "entity-123", fields)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if updated.GetID() != "entity-123" {
		t.Errorf("Expected entity ID 'entity-123', got %s", updated.GetID())
	}
}

func TestPluginRegistry_GetAllCommandProviders(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	// GetAllCommandProviders returns empty since our mock doesn't implement ICommandProvider interface
	providers := registry.GetAllCommandProviders()

	// Should return empty list for plugins without ICommandProvider
	if providers == nil {
		t.Error("Expected non-nil slice")
	}
}

// Tests from plugin_registry_capability_test.go
func TestPluginRegistry_CapabilityBasedRouting(t *testing.T) {
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

	// Test that the plugin was registered with its capabilities
	capabilities := plugin.GetCapabilities()
	if len(capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(capabilities))
	}

	// Check IEntityProvider capability
	if !containsString(capabilities, "IEntityProvider") {
		t.Error("Expected IEntityProvider capability")
	}

	// Check IEntityUpdater capability
	if !containsString(capabilities, "IEntityUpdater") {
		t.Error("Expected IEntityUpdater capability")
	}
}

func TestPluginRegistry_GetCommandProvider(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	// Create a mock plugin with ICommandProvider capability
	plugin := &MockCommandPlugin{
		name:         "cmd-plugin",
		version:      "1.0.0",
		capabilities: []string{"ICommandProvider"},
	}

	err := registry.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Get the command provider
	provider, err := registry.GetCommandProvider("cmd-plugin")
	if err != nil {
		t.Fatalf("Failed to get command provider: %v", err)
	}

	if provider == nil {
		t.Error("Expected non-nil command provider")
	}
}

func TestPluginRegistry_GetCommandProvider_NotFound(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	// Try to get a non-existent command provider
	_, err := registry.GetCommandProvider("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent command provider")
	}
}

func TestPluginRegistry_MissingCapabilityImplementation(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	// Create a plugin that declares ICommandProvider but doesn't implement it
	plugin := &MockBadPlugin{
		name:         "bad-plugin",
		version:      "1.0.0",
		capabilities: []string{"ICommandProvider"},
	}

	err := registry.RegisterPlugin(plugin)
	if err == nil {
		t.Error("Expected error when plugin declares capability but doesn't implement it")
	}
}

func TestPluginRegistry_EventEmitters(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	// Create a mock plugin with IEventEmitter capability
	plugin := &MockEventEmitterPlugin{
		name:         "event-plugin",
		version:      "1.0.0",
		capabilities: []string{"IEventEmitter"},
	}

	err := registry.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Get event emitters
	emitters := registry.GetEventEmitters()
	if len(emitters) != 1 {
		t.Errorf("Expected 1 event emitter, got %d", len(emitters))
	}
}

func TestPluginRegistry_MultipleCapabilities(t *testing.T) {
	logger := &app.NoOpLogger{}
	registry := app.NewPluginRegistry(logger)

	// Register plugin with multiple capabilities
	entityTypes := []pluginsdk.EntityTypeInfo{
		{Type: "session", DisplayName: "Session", Capabilities: []string{"IExtensible"}},
	}

	plugin := &MockMultiCapabilityPlugin{
		name:         "multi-plugin",
		version:      "1.0.0",
		capabilities: []string{"IEntityProvider", "ICommandProvider", "IEventEmitter"},
		entityTypes:  entityTypes,
	}

	err := registry.RegisterPlugin(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Verify entity provider was registered
	_, err = registry.GetPluginForEntityType("session")
	if err != nil {
		t.Errorf("Expected plugin to be registered as entity provider: %v", err)
	}

	// Verify command provider was registered
	_, err = registry.GetCommandProvider("multi-plugin")
	if err != nil {
		t.Errorf("Expected plugin to be registered as command provider: %v", err)
	}

	// Verify event emitter was registered
	emitters := registry.GetEventEmitters()
	if len(emitters) != 1 {
		t.Errorf("Expected 1 event emitter, got %d", len(emitters))
	}
}

// MockCommandPlugin is a test plugin that implements ICommandProvider
type MockCommandPlugin struct {
	name         string
	version      string
	capabilities []string
}

func (p *MockCommandPlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        p.name,
		Version:     p.version,
		Description: "Mock command plugin",
		IsCore:      true,
	}
}

func (p *MockCommandPlugin) GetCapabilities() []string {
	return p.capabilities
}

func (p *MockCommandPlugin) GetCommands() []pluginsdk.Command {
	return []pluginsdk.Command{}
}

// MockBadPlugin declares capabilities but doesn't implement them
type MockBadPlugin struct {
	name         string
	version      string
	capabilities []string
}

func (p *MockBadPlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        p.name,
		Version:     p.version,
		Description: "Mock bad plugin",
		IsCore:      true,
	}
}

func (p *MockBadPlugin) GetCapabilities() []string {
	return p.capabilities
}

// MockEventEmitterPlugin implements IEventEmitter
type MockEventEmitterPlugin struct {
	name         string
	version      string
	capabilities []string
}

func (p *MockEventEmitterPlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        p.name,
		Version:     p.version,
		Description: "Mock event emitter plugin",
		IsCore:      true,
	}
}

func (p *MockEventEmitterPlugin) GetCapabilities() []string {
	return p.capabilities
}

func (p *MockEventEmitterPlugin) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	return nil
}

// MockMultiCapabilityPlugin implements multiple capabilities
type MockMultiCapabilityPlugin struct {
	name         string
	version      string
	capabilities []string
	entityTypes  []pluginsdk.EntityTypeInfo
}

func (p *MockMultiCapabilityPlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        p.name,
		Version:     p.version,
		Description: "Mock multi-capability plugin",
		IsCore:      true,
	}
}

func (p *MockMultiCapabilityPlugin) GetCapabilities() []string {
	return p.capabilities
}

func (p *MockMultiCapabilityPlugin) GetEntityTypes() []pluginsdk.EntityTypeInfo {
	result := make([]pluginsdk.EntityTypeInfo, len(p.entityTypes))
	for i, et := range p.entityTypes {
		result[i] = pluginsdk.EntityTypeInfo{
			Type:         et.Type,
			DisplayName:  et.DisplayName,
			Capabilities: et.Capabilities,
		}
	}
	return result
}

func (p *MockMultiCapabilityPlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
	return []pluginsdk.IExtensible{}, nil
}

func (p *MockMultiCapabilityPlugin) GetEntity(ctx context.Context, entityID string) (pluginsdk.IExtensible, error) {
	return nil, pluginsdk.ErrNotFound
}

func (p *MockMultiCapabilityPlugin) GetCommands() []pluginsdk.Command {
	return []pluginsdk.Command{}
}

func (p *MockMultiCapabilityPlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (pluginsdk.IExtensible, error) {
	return nil, nil
}

func (p *MockMultiCapabilityPlugin) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	return nil
}

// Helper function for tests
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

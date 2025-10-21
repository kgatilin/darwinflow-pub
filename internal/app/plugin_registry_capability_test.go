package app_test

import (
	"context"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

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

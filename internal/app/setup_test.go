package app_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// MockEventRepository for testing
type MockEventRepository struct {
	initError error
	saveError error
	events    []*domain.Event
}

func (m *MockEventRepository) Initialize(ctx context.Context) error {
	if m.initError != nil {
		return m.initError
	}
	return nil
}

func (m *MockEventRepository) Save(ctx context.Context, event *domain.Event) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventRepository) FindByQuery(ctx context.Context, query domain.EventQuery) ([]*domain.Event, error) {
	if query.SessionID != "" {
		var result []*domain.Event
		for _, e := range m.events {
			if e.SessionID == query.SessionID {
				result = append(result, e)
			}
		}
		return result, nil
	}
	return m.events, nil
}

func (m *MockEventRepository) Close() error {
	return nil
}

// MockLogger for testing
type MockLogger struct {
	warnings []string
	errors   []string
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
}

func (m *MockLogger) Info(format string, args ...interface{}) {
}

func (m *MockLogger) Warn(format string, args ...interface{}) {
	m.warnings = append(m.warnings, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Error(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

// MockHookProvider implements IHookProvider for testing
type MockHookProvider struct {
	pluginInfo       pluginsdk.PluginInfo
	hooks            []pluginsdk.HookConfiguration
	installError     error
	refreshError     error
	installCalled    bool
	refreshCalled    bool
	installWorkdir   string
	refreshWorkdir   string
}

func (m *MockHookProvider) GetInfo() pluginsdk.PluginInfo {
	return m.pluginInfo
}

func (m *MockHookProvider) GetCapabilities() []string {
	return []string{"IHookProvider"}
}

func (m *MockHookProvider) GetHooks() []pluginsdk.HookConfiguration {
	return m.hooks
}

func (m *MockHookProvider) InstallHooks(workingDir string) error {
	m.installCalled = true
	m.installWorkdir = workingDir
	return m.installError
}

func (m *MockHookProvider) RefreshHooks(workingDir string) error {
	m.refreshCalled = true
	m.refreshWorkdir = workingDir
	return m.refreshError
}

// MockPlainPlugin implements Plugin for testing (without IHookProvider)
type MockPlainPlugin struct {
	pluginInfo pluginsdk.PluginInfo
}

func (m *MockPlainPlugin) GetInfo() pluginsdk.PluginInfo {
	return m.pluginInfo
}

func (m *MockPlainPlugin) GetCapabilities() []string {
	return []string{}
}

func TestNewSetupService(t *testing.T) {
	repo := &MockEventRepository{}
	logger := &MockLogger{}

	service := app.NewSetupService(repo, logger)
	if service == nil {
		t.Error("Expected non-nil SetupService")
	}
}

func TestSetupService_Initialize_WithHookProvider(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &MockEventRepository{}
	logger := &MockLogger{}
	service := app.NewSetupService(repo, logger)

	// Create mock plugin with IHookProvider
	mockHook := &MockHookProvider{
		pluginInfo: pluginsdk.PluginInfo{Name: "test-plugin", Version: "1.0.0"},
		hooks: []pluginsdk.HookConfiguration{
			{TriggerType: "trigger.test", Name: "TestHook", Command: "test"},
		},
	}

	ctx := context.Background()
	plugins := []pluginsdk.Plugin{mockHook}

	err := service.Initialize(ctx, fmt.Sprintf("%s/test.db", tmpDir), plugins)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify InstallHooks was called
	if !mockHook.installCalled {
		t.Error("Expected InstallHooks to be called")
	}

	// Verify working dir was passed correctly
	if mockHook.installWorkdir != tmpDir {
		t.Errorf("Expected workdir %s, got %s", tmpDir, mockHook.installWorkdir)
	}
}

func TestSetupService_Initialize_SkipsPluginsWithoutHooks(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &MockEventRepository{}
	logger := &MockLogger{}
	service := app.NewSetupService(repo, logger)

	// Create mock plugin without IHookProvider
	mockPlugin := &MockPlainPlugin{
		pluginInfo: pluginsdk.PluginInfo{Name: "no-hooks", Version: "1.0.0"},
	}

	// Create mock plugin with IHookProvider
	mockHook := &MockHookProvider{
		pluginInfo: pluginsdk.PluginInfo{Name: "with-hooks", Version: "1.0.0"},
	}

	ctx := context.Background()
	plugins := []pluginsdk.Plugin{mockPlugin, mockHook}

	err := service.Initialize(ctx, fmt.Sprintf("%s/test.db", tmpDir), plugins)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify only hook provider's InstallHooks was called
	if !mockHook.installCalled {
		t.Error("Expected InstallHooks to be called on hook provider")
	}
}

func TestSetupService_Initialize_PluginHookError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &MockEventRepository{}
	logger := &MockLogger{}
	service := app.NewSetupService(repo, logger)

	// Create mock plugin with hook installation error
	hookError := fmt.Errorf("hook install failed")
	mockHook := &MockHookProvider{
		pluginInfo:   pluginsdk.PluginInfo{Name: "failing-plugin", Version: "1.0.0"},
		installError: hookError,
	}

	ctx := context.Background()
	plugins := []pluginsdk.Plugin{mockHook}

	err := service.Initialize(ctx, fmt.Sprintf("%s/test.db", tmpDir), plugins)
	if err == nil {
		t.Error("Expected Initialize to return error when plugin hook install fails")
	}

	// Verify error contains plugin name
	if err != nil {
		if errStr := err.Error(); len(errStr) < 15 || errStr[:15] != "hook installati" {
			// Check if error contains "failing-plugin"
			if !strings.Contains(errStr, "failing-plugin") {
				t.Errorf("Expected error to contain plugin name, got: %v", err)
			}
		}
	}

	// Verify warning was logged
	if len(logger.warnings) == 0 {
		t.Error("Expected warning to be logged")
	}
}

func TestSetupService_Initialize_NoPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &MockEventRepository{}
	logger := &MockLogger{}
	service := app.NewSetupService(repo, logger)

	ctx := context.Background()
	plugins := []pluginsdk.Plugin{}

	err := service.Initialize(ctx, fmt.Sprintf("%s/test.db", tmpDir), plugins)
	if err != nil {
		t.Fatalf("Initialize failed with no plugins: %v", err)
	}
}

func TestSetupService_Initialize_RepositoryError(t *testing.T) {
	tmpDir := t.TempDir()
	expectedErr := fmt.Errorf("repository init failed")
	repo := &MockEventRepository{initError: expectedErr}
	logger := &MockLogger{}
	service := app.NewSetupService(repo, logger)

	ctx := context.Background()
	plugins := []pluginsdk.Plugin{}

	err := service.Initialize(ctx, fmt.Sprintf("%s/test.db", tmpDir), plugins)
	if err == nil {
		t.Error("Expected Initialize to return error when repository fails")
	}
}

func TestDefaultDBPath(t *testing.T) {
	// Verify the constant is defined
	if app.DefaultDBPath == "" {
		t.Error("DefaultDBPath should not be empty")
	}

	expectedPath := ".darwinflow/logs/events.db"
	if app.DefaultDBPath != expectedPath {
		t.Errorf("Expected DefaultDBPath = %s, got %s", expectedPath, app.DefaultDBPath)
	}
}

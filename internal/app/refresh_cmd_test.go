package app_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// mockEventRepository is a mock for testing
type mockEventRepository struct {
	initializeFunc func(ctx context.Context) error
}

func (m *mockEventRepository) Initialize(ctx context.Context) error {
	if m.initializeFunc != nil {
		return m.initializeFunc(ctx)
	}
	return nil
}

func (m *mockEventRepository) Close() error {
	return nil
}

func (m *mockEventRepository) Save(ctx context.Context, event *domain.Event) error {
	return nil
}

func (m *mockEventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	return nil
}

func (m *mockEventRepository) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) FindByQuery(ctx context.Context, query domain.EventQuery) ([]*domain.Event, error) {
	return nil, nil
}

// mockConfigLoader is a mock for testing
type mockConfigLoader struct {
	loadConfigFunc              func(path string) (*domain.Config, error)
	initializeDefaultConfigFunc func(path string) (string, error)
}

func (m *mockConfigLoader) LoadConfig(path string) (*domain.Config, error) {
	if m.loadConfigFunc != nil {
		return m.loadConfigFunc(path)
	}
	return &domain.Config{}, nil
}

func (m *mockConfigLoader) InitializeDefaultConfig(path string) (string, error) {
	if m.initializeDefaultConfigFunc != nil {
		return m.initializeDefaultConfigFunc(path)
	}
	return ".darwinflow.yaml", nil
}

func TestRefreshCommandHandler_Execute(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockEventRepository{}
	// Create a minimal plugin registry with no plugins for testing
	registry := app.NewPluginRegistry(&mockLogger{})
	mockConfigLdr := &mockConfigLoader{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}

	handler := app.NewRefreshCommandHandler(mockRepo, registry, mockConfigLdr, logger, out)

	err := handler.Execute(ctx, "/test/db/path.db")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Refreshing DarwinFlow") {
		t.Errorf("Output should indicate refreshing, got: %s", output)
	}
	if !strings.Contains(output, "Database schema updated") {
		t.Errorf("Output should confirm database update, got: %s", output)
	}
	if !strings.Contains(output, "Updating hooks for all plugins") {
		t.Errorf("Output should confirm hooks update, got: %s", output)
	}
	if !strings.Contains(output, "refreshed successfully") {
		t.Errorf("Output should indicate success, got: %s", output)
	}
}

func TestRefreshCommandHandler_Execute_WithMissingConfig(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockEventRepository{}
	registry := app.NewPluginRegistry(&mockLogger{})
	mockConfigLdr := &mockConfigLoader{
		loadConfigFunc: func(path string) (*domain.Config, error) {
			return nil, nil // Config doesn't exist
		},
	}
	logger := &mockLogger{}
	out := &bytes.Buffer{}

	handler := app.NewRefreshCommandHandler(mockRepo, registry, mockConfigLdr, logger, out)

	err := handler.Execute(ctx, "/test/db/path.db")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Creating default configuration") {
		t.Errorf("Output should indicate creating default config, got: %s", output)
	}
}

func TestRefreshCommandHandler_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockEventRepository{
		initializeFunc: func(ctx context.Context) error {
			return fmt.Errorf("database initialization failed")
		},
	}
	registry := app.NewPluginRegistry(&mockLogger{})
	mockConfigLdr := &mockConfigLoader{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}

	handler := app.NewRefreshCommandHandler(mockRepo, registry, mockConfigLdr, logger, out)

	err := handler.Execute(ctx, "/test/db/path.db")
	if err == nil {
		t.Error("Execute should fail when repository initialization fails")
	}
	if !strings.Contains(err.Error(), "error updating database schema") {
		t.Errorf("Error should mention database schema, got: %v", err)
	}
}

func TestRefreshCommandHandler_Execute_WithPlugins(t *testing.T) {
	ctx := context.Background()
	mockRepo := &mockEventRepository{}
	registry := app.NewPluginRegistry(&mockLogger{})
	// Add no plugins to test the empty plugin case
	mockConfigLdr := &mockConfigLoader{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}

	handler := app.NewRefreshCommandHandler(mockRepo, registry, mockConfigLdr, logger, out)

	err := handler.Execute(ctx, "/test/db/path.db")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Updating hooks for all plugins") {
		t.Errorf("Output should mention updating hooks, got: %s", output)
	}
}

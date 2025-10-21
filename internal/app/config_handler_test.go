package app_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// MockConfigLoader is a mock implementation for testing
type MockConfigLoader struct {
	config         *domain.Config
	initPath       string
	initError      error
	loadError      error
}

func (m *MockConfigLoader) LoadConfig(configPath string) (*domain.Config, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	if m.config != nil {
		return m.config, nil
	}
	return domain.DefaultConfig(), nil
}

func (m *MockConfigLoader) InitializeDefaultConfig(configPath string) (string, error) {
	if m.initError != nil {
		return "", m.initError
	}
	m.initPath = configPath
	if configPath == "" {
		return ".darwinflow.yaml", nil
	}
	return configPath, nil
}

func TestNewConfigCommandHandler(t *testing.T) {
	loader := &MockConfigLoader{}
	logger := &app.NoOpLogger{}
	output := &bytes.Buffer{}

	handler := app.NewConfigCommandHandler(loader, logger, output)

	if handler == nil {
		t.Fatal("Expected non-nil ConfigCommandHandler")
	}
}

func TestConfigCommandHandler_Init(t *testing.T) {
	ctx := context.Background()
	loader := &MockConfigLoader{}
	logger := &app.NoOpLogger{}
	output := &bytes.Buffer{}

	handler := app.NewConfigCommandHandler(loader, logger, output)

	err := handler.Init(ctx, "", false)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	outputStr := output.String()
	if !contains(outputStr, "Created config file") {
		t.Error("Output should indicate config file was created")
	}
}

func TestConfigCommandHandler_Init_WithPath(t *testing.T) {
	ctx := context.Background()
	loader := &MockConfigLoader{}
	logger := &app.NoOpLogger{}
	output := &bytes.Buffer{}

	handler := app.NewConfigCommandHandler(loader, logger, output)

	err := handler.Init(ctx, "/custom/path/.darwinflow.yaml", false)
	if err != nil {
		t.Fatalf("Init with custom path failed: %v", err)
	}

	if loader.initPath != "/custom/path/.darwinflow.yaml" {
		t.Errorf("Expected init path '/custom/path/.darwinflow.yaml', got %s", loader.initPath)
	}
}

func TestConfigCommandHandler_Init_WithError(t *testing.T) {
	ctx := context.Background()
	expectedErr := fmt.Errorf("init failed")
	loader := &MockConfigLoader{
		initError: expectedErr,
	}
	logger := &app.NoOpLogger{}
	output := &bytes.Buffer{}

	handler := app.NewConfigCommandHandler(loader, logger, output)

	err := handler.Init(ctx, "", false)
	if err == nil {
		t.Error("Expected error when init fails")
	}
}

func TestConfigCommandHandler_Show(t *testing.T) {
	ctx := context.Background()

	config := domain.DefaultConfig()
	config.Prompts["custom_prompt"] = "Test prompt"

	loader := &MockConfigLoader{
		config: config,
	}
	logger := &app.NoOpLogger{}
	output := &bytes.Buffer{}

	handler := app.NewConfigCommandHandler(loader, logger, output)

	err := handler.Show(ctx)
	if err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	outputStr := output.String()
	if !contains(outputStr, "DarwinFlow Configuration") {
		t.Error("Output should contain configuration title")
	}

	if !contains(outputStr, "custom_prompt") {
		t.Error("Output should list custom prompt")
	}
}

func TestConfigCommandHandler_Show_WithLoadError(t *testing.T) {
	ctx := context.Background()

	expectedErr := fmt.Errorf("load failed")
	loader := &MockConfigLoader{
		loadError: expectedErr,
	}
	logger := &app.NoOpLogger{}
	output := &bytes.Buffer{}

	handler := app.NewConfigCommandHandler(loader, logger, output)

	err := handler.Show(ctx)
	if err == nil {
		t.Error("Expected error when load fails")
	}
}

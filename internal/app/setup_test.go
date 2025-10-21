package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// MockEventRepository for testing
type MockEventRepository struct {
	initError     error
	saveError     error
	queryError    error
	rawQueryError error
	events        []*domain.Event
	savedEvents   []*domain.Event
	queryResult   *domain.QueryResult
	closed        bool
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
	m.savedEvents = append(m.savedEvents, event)
	return nil
}

func (m *MockEventRepository) FindByQuery(ctx context.Context, query domain.EventQuery) ([]*domain.Event, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
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
	m.closed = true
	return nil
}

func (m *MockEventRepository) ExecuteRawQuery(ctx context.Context, query string) (*domain.QueryResult, error) {
	if m.rawQueryError != nil {
		return nil, m.rawQueryError
	}
	if m.queryResult != nil {
		return m.queryResult, nil
	}
	return &domain.QueryResult{
		Columns: []string{"id", "event_type"},
		Rows:    [][]interface{}{{"1", "test"}},
	}, nil
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

func TestNewSetupService(t *testing.T) {
	repo := &MockEventRepository{}
	logger := &MockLogger{}

	service := app.NewSetupService(repo, logger)
	if service == nil {
		t.Error("Expected non-nil SetupService")
	}
}

func TestSetupService_Initialize_Success(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &MockEventRepository{}
	logger := &MockLogger{}
	service := app.NewSetupService(repo, logger)

	ctx := context.Background()

	err := service.Initialize(ctx, fmt.Sprintf("%s/test.db", tmpDir))
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify repository Initialize was called (no error returned means it was called)
}

func TestSetupService_Initialize_RepositoryError(t *testing.T) {
	tmpDir := t.TempDir()
	expectedErr := fmt.Errorf("repository init failed")
	repo := &MockEventRepository{initError: expectedErr}
	logger := &MockLogger{}
	service := app.NewSetupService(repo, logger)

	ctx := context.Background()

	err := service.Initialize(ctx, fmt.Sprintf("%s/test.db", tmpDir))
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

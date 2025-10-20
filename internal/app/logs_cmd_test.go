package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// mockLogsService is a mock implementation for testing
type mockLogsService struct {
	listRecentLogsFunc  func(ctx context.Context, limit, sessionLimit int, sessionID string, ordered bool) ([]*app.LogRecord, error)
	executeRawQueryFunc func(ctx context.Context, query string) (*domain.QueryResult, error)
}

func (m *mockLogsService) ListRecentLogs(ctx context.Context, limit, sessionLimit int, sessionID string, ordered bool) ([]*app.LogRecord, error) {
	if m.listRecentLogsFunc != nil {
		return m.listRecentLogsFunc(ctx, limit, sessionLimit, sessionID, ordered)
	}
	return []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "tool.invoked",
			SessionID: "session-123",
			Content:   "Read /test.go",
		},
		{
			ID:        "event-2",
			Timestamp: time.Now(),
			EventType: "chat.message.user",
			SessionID: "session-123",
			Content:   "Hello",
		},
	}, nil
}

func (m *mockLogsService) ExecuteRawQuery(ctx context.Context, query string) (*domain.QueryResult, error) {
	if m.executeRawQueryFunc != nil {
		return m.executeRawQueryFunc(ctx, query)
	}
	return &domain.QueryResult{
		Columns: []string{"id", "event_type", "count"},
		Rows: [][]interface{}{
			{"1", "tool.invoked", 10},
			{"2", "chat.message.user", 5},
		},
	}, nil
}

func TestLogsCommandHandler_ListLogs(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ListLogs(ctx, 20, 0, "", false, "text")
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Showing 2 most recent logs") {
		t.Errorf("Output should show log count, got: %s", output)
	}
	if !strings.Contains(output, "event-1") || !strings.Contains(output, "event-2") {
		t.Errorf("Output should contain event IDs, got: %s", output)
	}
}

func TestLogsCommandHandler_ListLogsWithSessionID(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ListLogs(ctx, 20, 0, "session-123", false, "text")
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "session-123") {
		t.Errorf("Output should contain session ID, got: %s", output)
	}
}

func TestLogsCommandHandler_ListLogsNoResults(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{
		listRecentLogsFunc: func(ctx context.Context, limit, sessionLimit int, sessionID string, ordered bool) ([]*app.LogRecord, error) {
			return []*app.LogRecord{}, nil
		},
	}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ListLogs(ctx, 20, 0, "", false, "text")
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No logs found") {
		t.Errorf("Output should indicate no logs found, got: %s", output)
	}
}

func TestLogsCommandHandler_ListLogsCSVFormat(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ListLogs(ctx, 20, 0, "", false, "csv")
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}

	output := out.String()
	// CSV format should have headers and data rows
	if !strings.Contains(output, "ID,Timestamp,EventType") {
		t.Errorf("CSV output should contain headers, got: %s", output)
	}
	if !strings.Contains(output, "event-1") {
		t.Errorf("CSV output should contain event data, got: %s", output)
	}
}

func TestLogsCommandHandler_ListLogsMarkdownFormat(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ListLogs(ctx, 20, 0, "", false, "markdown")
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}

	output := out.String()
	// Markdown format should have event markers or sections
	if !strings.Contains(output, "Event Timeline") {
		t.Errorf("Markdown output should contain event timeline section, got: %s", output)
	}
	if !strings.Contains(output, "tool.invoked") {
		t.Errorf("Markdown output should contain event types, got: %s", output)
	}
}

func TestLogsCommandHandler_ListLogsInvalidFormat(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ListLogs(ctx, 20, 0, "", false, "invalid")
	if err == nil {
		t.Error("ListLogs should fail with invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Error should mention invalid format, got: %v", err)
	}
}

func TestLogsCommandHandler_ExecuteRawQuery(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ExecuteRawQuery(ctx, "SELECT * FROM events")
	if err != nil {
		t.Fatalf("ExecuteRawQuery failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "id") || !strings.Contains(output, "event_type") {
		t.Errorf("Output should contain column names, got: %s", output)
	}
	if !strings.Contains(output, "2 rows") {
		t.Errorf("Output should show row count, got: %s", output)
	}
}

func TestLogsCommandHandler_ListLogsWithSessionLimit(t *testing.T) {
	ctx := context.Background()
	mockService := &mockLogsService{}
	out := &bytes.Buffer{}
	handler := app.NewLogsCommandHandler(mockService, out)

	err := handler.ListLogs(ctx, 0, 3, "", false, "text")
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "from 3 most recent sessions") {
		t.Errorf("Output should mention session limit, got: %s", output)
	}
}

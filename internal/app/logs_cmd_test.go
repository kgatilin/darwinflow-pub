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

// Tests from logs_service_test.go
func TestNewLogsService(t *testing.T) {
	eventRepo := &MockEventRepository{}
	service := app.NewLogsService(eventRepo, eventRepo)

	if service == nil {
		t.Fatal("Expected non-nil LogsService")
	}
}

func TestLogsService_ListRecentLogs(t *testing.T) {
	ctx := context.Background()

	event1 := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{
		"tool": "Read",
	}, "Read file")
	event2 := domain.NewEvent("claude.chat.message.user", "session-123", map[string]interface{}{
		"message": "Hello",
	}, "Hello")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event1, event2},
	}

	service := app.NewLogsService(eventRepo, eventRepo)

	records, err := service.ListRecentLogs(ctx, 10, 0, "", false)
	if err != nil {
		t.Fatalf("ListRecentLogs failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if records[0].EventType != "claude.tool.invoked" {
		t.Errorf("Expected first event type 'claude.tool.invoked', got %s", records[0].EventType)
	}
}

func TestLogsService_ListRecentLogs_WithSessionID(t *testing.T) {
	ctx := context.Background()

	event1 := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{}, "test1")
	event2 := domain.NewEvent("claude.tool.invoked", "session-456", map[string]interface{}{}, "test2")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event1, event2},
	}

	service := app.NewLogsService(eventRepo, eventRepo)

	records, err := service.ListRecentLogs(ctx, 10, 0, "session-123", false)
	if err != nil {
		t.Fatalf("ListRecentLogs failed: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record for session-123, got %d", len(records))
	}

	if records[0].SessionID != "session-123" {
		t.Errorf("Expected session ID 'session-123', got %s", records[0].SessionID)
	}
}

func TestLogsService_ListRecentLogs_WithSessionLimit(t *testing.T) {
	ctx := context.Background()

	eventRepo := &MockEventRepository{
		queryResult: &domain.QueryResult{
			Columns: []string{"session_id"},
			Rows: [][]interface{}{
				{"session-123"},
				{"session-456"},
			},
		},
		events: []*domain.Event{
			domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{}, "test1"),
			domain.NewEvent("claude.tool.invoked", "session-456", map[string]interface{}{}, "test2"),
		},
	}

	service := app.NewLogsService(eventRepo, eventRepo)

	records, err := service.ListRecentLogs(ctx, 0, 2, "", false)
	if err != nil {
		t.Fatalf("ListRecentLogs with session limit failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

func TestLogsService_ExecuteRawQuery(t *testing.T) {
	ctx := context.Background()

	eventRepo := &MockEventRepository{
		queryResult: &domain.QueryResult{
			Columns: []string{"id", "event_type", "count"},
			Rows: [][]interface{}{
				{"1", "claude.tool.invoked", 10},
				{"2", "claude.chat.message.user", 5},
			},
		},
	}

	service := app.NewLogsService(eventRepo, eventRepo)

	result, err := service.ExecuteRawQuery(ctx, "SELECT * FROM events")
	if err != nil {
		t.Fatalf("ExecuteRawQuery failed: %v", err)
	}

	if len(result.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(result.Columns))
	}

	if len(result.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(result.Rows))
	}
}

func TestFormatLogRecord(t *testing.T) {
	record := &app.LogRecord{
		ID:        "event-123",
		Timestamp: time.Now(),
		EventType: "claude.tool.invoked",
		SessionID: "session-123",
		Payload:   []byte(`{"tool":"Read","file":"test.go"}`),
		Content:   "Read test.go",
	}

	output := app.FormatLogRecord(0, record)

	if output == "" {
		t.Error("Expected non-empty output")
	}

	if !contains(output, "event-123") {
		t.Error("Output should contain event ID")
	}

	if !contains(output, "claude.tool.invoked") {
		t.Error("Output should contain event type")
	}

	if !contains(output, "session-123") {
		t.Error("Output should contain session ID")
	}
}

func TestFormatQueryValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "NULL",
		},
		{
			name:     "string value",
			input:    "test",
			expected: "test",
		},
		{
			name:     "int64 value",
			input:    int64(42),
			expected: "42",
		},
		{
			name:     "bytes value",
			input:    []byte("test"),
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.FormatQueryValue(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatLogsAsCSV(t *testing.T) {
	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "claude.tool.invoked",
			SessionID: "session-123",
			Payload:   []byte(`{"tool":"Read"}`),
			Content:   "Read file",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsCSV(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsCSV failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "ID,Timestamp,EventType") {
		t.Error("CSV output should contain headers")
	}

	if !contains(output, "event-1") {
		t.Error("CSV output should contain event ID")
	}
}

func TestFormatLogsAsMarkdown(t *testing.T) {
	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "claude.tool.invoked",
			SessionID: "session-123",
			Payload:   []byte(`{"tool":"Read","file":"test.go"}`),
			Content:   "Read test.go",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "# Event Logs") {
		t.Error("Markdown output should contain title")
	}

	if !contains(output, "Event Timeline") {
		t.Error("Markdown output should contain event timeline")
	}

	if !contains(output, "claude.tool.invoked") {
		t.Error("Markdown output should contain event type")
	}
}

func TestFormatLogsAsMarkdown_NoEvents(t *testing.T) {
	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, []*app.LogRecord{})
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "No events found") {
		t.Error("Markdown output should indicate no events found")
	}
}

func TestFormatQueryValue_Timestamp(t *testing.T) {
	// Test timestamp value (13 digits)
	timestamp := int64(1609459200000) // 2021-01-01 00:00:00 UTC
	result := app.FormatQueryValue(timestamp)

	if !contains(result, "1609459200000") {
		t.Errorf("Result should contain timestamp, got %s", result)
	}
}

func TestFormatQueryValue_JSONBytes(t *testing.T) {
	jsonBytes := []byte(`{"tool":"Read","file":"test.go"}`)
	result := app.FormatQueryValue(jsonBytes)

	if !contains(result, "tool") {
		t.Errorf("Result should contain JSON content, got %s", result)
	}
}

func TestFormatQueryValue_LongString(t *testing.T) {
	longString := ""
	for i := 0; i < 150; i++ {
		longString += "a"
	}

	result := app.FormatQueryValue(longString)

	if len(result) > 105 {
		t.Errorf("Long string should be truncated")
	}
}

func TestFormatLogRecord_LongContent(t *testing.T) {
	longContent := ""
	for i := 0; i < 300; i++ {
		longContent += "x"
	}

	record := &app.LogRecord{
		ID:        "event-123",
		Timestamp: time.Now(),
		EventType: "test.event",
		SessionID: "session-123",
		Payload:   []byte(`{}`),
		Content:   longContent,
	}

	output := app.FormatLogRecord(0, record)

	// Content should be truncated
	if !contains(output, "...") {
		t.Error("Long content should be truncated with ...")
	}
}

func TestFormatLogRecord_WithInvalidJSON(t *testing.T) {
	record := &app.LogRecord{
		ID:        "event-123",
		Timestamp: time.Now(),
		EventType: "test.event",
		SessionID: "session-123",
		Payload:   []byte(`{invalid json`),
		Content:   "test content",
	}

	output := app.FormatLogRecord(0, record)

	// Should still produce output even with invalid JSON
	if output == "" {
		t.Error("Expected output even with invalid JSON payload")
	}

	if !contains(output, "event-123") {
		t.Error("Output should contain event ID")
	}
}

func TestFormatLogRecord_WithNestedJSON(t *testing.T) {
	record := &app.LogRecord{
		ID:        "event-123",
		Timestamp: time.Now(),
		EventType: "test.event",
		SessionID: "session-123",
		Payload:   []byte(`{"tool":"Read","params":"{\"file\":\"test.go\"}"}`),
		Content:   "test content",
	}

	output := app.FormatLogRecord(0, record)

	// Should expand nested JSON
	if !contains(output, "test.go") || !contains(output, "Read") {
		t.Error("Output should contain expanded nested JSON")
	}
}

func TestFormatLogsAsMarkdown_WithArrayPayload(t *testing.T) {
	// Test with array in payload to cover expandNestedJSON array branch
	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "test.event",
			SessionID: "session-123",
			Payload:   []byte(`{"items":["{\"key\":\"value\"}","plain string"]}`),
			Content:   "test",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "test.event") {
		t.Error("Output should contain event type")
	}
}

func TestFormatLogsAsMarkdown_WithMessageField(t *testing.T) {
	// Test with message field to cover getTruncateLimit message case
	longMessage := ""
	for i := 0; i < 250; i++ {
		longMessage += "x"
	}

	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "test.event",
			SessionID: "session-123",
			Payload:   []byte(`{"message":"` + longMessage + `"}`),
			Content:   "test",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	// Message should not be truncated (getTruncateLimit returns -1 for message)
	if !contains(output, "message") {
		t.Error("Output should contain message field")
	}
}

func TestFormatLogsAsMarkdown_WithWriteToolContent(t *testing.T) {
	// Test Write tool with content field to cover getTruncateLimit Write tool case
	longContent := ""
	for i := 0; i < 200; i++ {
		longContent += "x"
	}

	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "test.event",
			SessionID: "session-123",
			Payload:   []byte(`{"tool":"Write","content":"` + longContent + `"}`),
			Content:   "test",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	// Content should be truncated to 100 chars for Write tool
	if !contains(output, "content") {
		t.Error("Output should contain content field")
	}
}

func TestFormatLogsAsMarkdown_WithFilePath(t *testing.T) {
	// Test file_path field to cover getTruncateLimit file_path case
	longPath := "/very/long/path/"
	for i := 0; i < 30; i++ {
		longPath += "subdir/"
	}

	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "test.event",
			SessionID: "session-123",
			Payload:   []byte(`{"file_path":"` + longPath + `"}`),
			Content:   "test",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "file_path") {
		t.Error("Output should contain file_path field")
	}
}

func TestFormatLogsAsMarkdown_WithNullValue(t *testing.T) {
	// Test null value in payload to cover formatMarkdownPayloadWithContext null case
	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "test.event",
			SessionID: "session-123",
			Payload:   []byte(`{"field":null}`),
			Content:   "test",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "null") {
		t.Error("Output should contain null value")
	}
}

func TestFormatLogsAsMarkdown_WithEmptyString(t *testing.T) {
	// Test empty string value to cover formatMarkdownPayloadWithContext empty string case
	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "test.event",
			SessionID: "session-123",
			Payload:   []byte(`{"field":""}`),
			Content:   "test",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "empty") {
		t.Error("Output should indicate empty field")
	}
}

func TestFormatLogsAsMarkdown_WithMetadataField(t *testing.T) {
	// Test metadata field which should be skipped in formatting
	records := []*app.LogRecord{
		{
			ID:        "event-1",
			Timestamp: time.Now(),
			EventType: "test.event",
			SessionID: "session-123",
			Payload:   []byte(`{"metadata":{"key":"value"},"data":"content"}`),
			Content:   "test",
		},
	}

	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, records)
	if err != nil {
		t.Fatalf("FormatLogsAsMarkdown failed: %v", err)
	}

	output := buf.String()
	// metadata field should be skipped, but data should be present
	if !contains(output, "data") {
		t.Error("Output should contain data field")
	}
}

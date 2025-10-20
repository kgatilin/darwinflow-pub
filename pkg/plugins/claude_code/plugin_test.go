package claude_code_test

import (
	"context"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// mockLogger implements pluginsdk.Logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {}

// Minimal test focusing on what can be tested without complex mocking
// Full integration tests would require a test database

func TestNewClaudeCodePlugin(t *testing.T) {
	// This test verifies the constructor works
	// We use nil services since we're only testing construction
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	if plugin == nil {
		t.Fatal("NewClaudeCodePlugin returned nil")
	}
}

func TestGetInfo(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	info := plugin.GetInfo()

	if info.Name != "claude-code" {
		t.Errorf("Expected Name 'claude-code', got %q", info.Name)
	}
	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if !info.IsCore {
		t.Error("Expected IsCore to be true for claude-code plugin")
	}
	if info.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestGetEntityTypes(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	entityTypes := plugin.GetEntityTypes()

	if len(entityTypes) != 1 {
		t.Fatalf("Expected 1 entity type, got %d", len(entityTypes))
	}

	sessionType := entityTypes[0]
	if sessionType.Type != "session" {
		t.Errorf("Expected Type 'session', got %q", sessionType.Type)
	}
	if sessionType.DisplayName == "" {
		t.Error("DisplayName should not be empty")
	}
	if sessionType.DisplayNamePlural == "" {
		t.Error("DisplayNamePlural should not be empty")
	}
	if len(sessionType.Capabilities) == 0 {
		t.Error("Should have capabilities defined")
	}

	// Verify expected capabilities
	expectedCaps := map[string]bool{
		"IExtensible": true,
		"IHasContext": true,
		"ITrackable":  true,
	}

	for _, cap := range sessionType.Capabilities {
		if !expectedCaps[cap] {
			t.Errorf("Unexpected capability: %s", cap)
		}
		delete(expectedCaps, cap)
	}

	if len(expectedCaps) > 0 {
		t.Errorf("Missing expected capabilities: %v", expectedCaps)
	}

	if sessionType.Icon == "" {
		t.Error("Icon should not be empty")
	}
}

func TestUpdateEntity_ReadOnly(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	ctx := context.Background()
	_, err := plugin.UpdateEntity(ctx, "session-1", map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for read-only update, got nil")
	}
	// Check against SDK error constant
	if err != pluginsdk.ErrReadOnly {
		t.Errorf("Expected pluginsdk.ErrReadOnly, got: %v", err)
	}
}

func TestGetCommands(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	commands := plugin.GetCommands()

	// Verify we get exactly 6 commands (including emit-event and session-summary)
	if len(commands) != 6 {
		t.Fatalf("Expected 6 commands, got %d", len(commands))
	}

	// Verify expected command names
	expectedCommands := map[string]bool{
		"init":              false,
		"emit-event":        false,
		"log":               false,
		"auto-summary":      false,
		"auto-summary-exec": false,
		"session-summary":   false,
	}

	for _, cmd := range commands {
		name := cmd.GetName()
		if _, exists := expectedCommands[name]; !exists {
			t.Errorf("Unexpected command: %s", name)
		}
		expectedCommands[name] = true

		// Verify each command has required metadata
		if cmd.GetDescription() == "" {
			t.Errorf("Command %s has empty description", name)
		}
		if cmd.GetUsage() == "" {
			t.Errorf("Command %s has empty usage", name)
		}
	}

	// Verify all expected commands were found
	for name, found := range expectedCommands {
		if !found {
			t.Errorf("Expected command %s not found", name)
		}
	}
}

func TestCommandProvider_Interface(t *testing.T) {
	// Verify that ClaudeCodePlugin implements SDK ICommandProvider
	var _ pluginsdk.ICommandProvider = (*claude_code.ClaudeCodePlugin)(nil)
}

// mockLogsService implements claude_code.LogsService for testing
type mockLogsService struct {
	logs []*claude_code.LogRecord
}

func (m *mockLogsService) ListRecentLogs(ctx context.Context, limit, offset int, sessionID string, ordered bool) ([]*claude_code.LogRecord, error) {
	if sessionID != "" {
		// Filter logs by session ID
		var filtered []*claude_code.LogRecord
		for _, log := range m.logs {
			if log.SessionID == sessionID {
				filtered = append(filtered, log)
			}
		}
		m.logs = filtered
	}
	return m.logs, nil
}

// mockAnalysisService implements claude_code.AnalysisService for testing
type mockAnalysisService struct {
	sessionIDs []string
	analyses   map[string][]*claude_code.SessionAnalysis
}

func (m *mockAnalysisService) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	if limit > 0 && limit < len(m.sessionIDs) {
		return m.sessionIDs[:limit], nil
	}
	return m.sessionIDs, nil
}

func (m *mockAnalysisService) GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*claude_code.SessionAnalysis, error) {
	if analyses, ok := m.analyses[sessionID]; ok {
		return analyses, nil
	}
	return []*claude_code.SessionAnalysis{}, nil
}

func (m *mockAnalysisService) EstimateTokenCount(ctx context.Context, sessionID string) (int, error) {
	return 1000, nil
}

func (m *mockAnalysisService) GetLastSession(ctx context.Context) (string, error) {
	if len(m.sessionIDs) > 0 {
		return m.sessionIDs[0], nil
	}
	return "", nil
}

func (m *mockAnalysisService) AnalyzeSessionWithPrompt(ctx context.Context, sessionID string, promptName string) (*claude_code.SessionAnalysis, error) {
	// Return a mock analysis
	return &claude_code.SessionAnalysis{
		ID:              "analysis-1",
		SessionID:       sessionID,
		PromptName:      promptName,
		ModelUsed:       "claude-sonnet-4",
		PatternsSummary: "Mock analysis",
	}, nil
}

// TestClaudeCodePlugin_QueryBuildsSessionsFromEvents demonstrates event sourcing:
// Sessions are reconstructed from historical events stored in the event repository.
//
// This test verifies that:
// 1. Events are queried from the repository via LogsService
// 2. Sessions are built from event streams (not retrieved directly)
// 3. Session state is derived from events (event sourcing pattern)
// 4. Multiple events can be reconstructed into a single session
func TestClaudeCodePlugin_QueryBuildsSessionsFromEvents(t *testing.T) {
	ctx := context.Background()
	sessionID := "session-1"
	logger := &mockLogger{}

	// Create sample events that represent a session timeline
	now := time.Now()
	events := []*claude_code.LogRecord{
		{
			ID:        "event-1",
			Timestamp: now.Add(-2 * 60 * 1000),
			EventType: "chat.started",
			SessionID: sessionID,
			Payload:   []byte(`{"session_id":"session-1"}`),
			Content:   "Session started",
		},
		{
			ID:        "event-2",
			Timestamp: now.Add(-1 * 60 * 1000),
			EventType: "tool.invoked",
			SessionID: sessionID,
			Payload:   []byte(`{"tool":"Read","file":"/test.txt"}`),
			Content:   "Reading file",
		},
		{
			ID:        "event-3",
			Timestamp: now,
			EventType: "chat.message.user",
			SessionID: sessionID,
			Payload:   []byte(`{"message":"Hello"}`),
			Content:   "User message",
		},
	}

	// Create mock services
	logsService := &mockLogsService{logs: events}
	analysisService := &mockAnalysisService{
		sessionIDs: []string{sessionID},
		analyses:   map[string][]*claude_code.SessionAnalysis{},
	}

	// Create plugin with mock services
	plugin := claude_code.NewClaudeCodePlugin(
		analysisService,
		logsService,
		logger,
		nil,
		nil,
		"",
	)

	// Query for sessions - this triggers event sourcing
	query := pluginsdk.EntityQuery{Limit: 10}
	sessions, err := plugin.Query(ctx, query)

	// Verify the query succeeded
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Verify we got one session
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	// Cast to SessionEntity to verify it was built from events
	sessionEntity, ok := sessions[0].(*claude_code.SessionEntity)
	if !ok {
		t.Fatal("Expected SessionEntity type")
	}

	// Verify session ID matches
	if sessionEntity.GetID() != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, sessionEntity.GetID())
	}

	// Verify session type is correct
	if sessionEntity.GetType() != "session" {
		t.Errorf("Expected type 'session', got %s", sessionEntity.GetType())
	}

	// Verify event count matches events we put in the repository
	fields := sessionEntity.GetAllFields()
	eventCount, ok := fields["event_count"].(int)
	if !ok {
		t.Fatal("event_count field missing or not an int")
	}
	if eventCount != 3 {
		t.Errorf("Expected 3 events, got %d (events sourced from repository)", eventCount)
	}

	// Verify first and last event timestamps match the events we stored
	firstEventTime, ok := fields["first_event"].(time.Time)
	if !ok {
		t.Fatal("first_event field missing or not time.Time")
	}
	expectedFirstTime := now.Add(-2 * 60 * 1000)
	if !firstEventTime.Equal(expectedFirstTime) {
		t.Errorf("Expected first event at %v, got %v", expectedFirstTime, firstEventTime)
	}

	lastEventTime, ok := fields["last_event"].(time.Time)
	if !ok {
		t.Fatal("last_event field missing or not time.Time")
	}
	if !lastEventTime.Equal(now) {
		t.Errorf("Expected last event at %v, got %v", now, lastEventTime)
	}

	// Verify session status is active (no analyses)
	status, ok := fields["status"].(string)
	if !ok {
		t.Fatal("status field missing or not string")
	}
	if status != "active" {
		t.Errorf("Expected status 'active', got %s", status)
	}

	t.Log("✓ Session successfully reconstructed from event stream (event sourcing verified)")
	t.Log("✓ Event count matches repository:", eventCount)
	t.Log("✓ Session boundaries derived from first/last events")
	t.Log("✓ Event sourcing pattern confirmed: sessions built from events, not stored directly")
}

// TestClaudeCodePlugin_GetEntity_RebuildsSameSessionFromEvents verifies that
// GetEntity also rebuilds sessions from events, demonstrating consistency.
func TestClaudeCodePlugin_GetEntity_RebuildsSameSessionFromEvents(t *testing.T) {
	ctx := context.Background()
	sessionID := "session-2"
	logger := &mockLogger{}

	// Create events
	now := time.Now()
	events := []*claude_code.LogRecord{
		{
			ID:        "event-1",
			Timestamp: now,
			EventType: "tool.invoked",
			SessionID: sessionID,
			Payload:   []byte(`{"tool":"Bash"}`),
			Content:   "",
		},
	}

	// Create mock services
	logsService := &mockLogsService{logs: events}
	analysisService := &mockAnalysisService{
		sessionIDs: []string{sessionID},
		analyses:   map[string][]*claude_code.SessionAnalysis{},
	}

	plugin := claude_code.NewClaudeCodePlugin(
		analysisService,
		logsService,
		logger,
		nil,
		nil,
		"",
	)

	// GetEntity should also rebuild from events
	entity, err := plugin.GetEntity(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetEntity failed: %v", err)
	}

	if entity == nil {
		t.Fatal("Expected entity, got nil")
	}

	if entity.GetID() != sessionID {
		t.Errorf("Expected ID %s, got %s", sessionID, entity.GetID())
	}

	t.Log("✓ GetEntity confirms event sourcing: retrieved entity matches event stream")
}

// TestClaudeCodePlugin_QueryWithAnalyses verifies sessions include analysis state
// when available, demonstrating complete derived state from events + analyses.
func TestClaudeCodePlugin_QueryWithAnalyses(t *testing.T) {
	ctx := context.Background()
	sessionID := "session-3"
	logger := &mockLogger{}

	// Create events
	now := time.Now()
	events := []*claude_code.LogRecord{
		{
			ID:        "event-1",
			Timestamp: now,
			EventType: "chat.started",
			SessionID: sessionID,
			Payload:   []byte(`{}`),
			Content:   "",
		},
	}

	// Create analyses
	analyses := []*claude_code.SessionAnalysis{
		{
			ID:              "analysis-1",
			SessionID:       sessionID,
			PromptName:      "tool_analysis",
			ModelUsed:       "claude-sonnet",
			PatternsSummary: "User frequently uses Read tool",
			AnalyzedAt:      now.Add(1 * time.Minute),
		},
	}

	// Create mock services
	logsService := &mockLogsService{logs: events}
	analysisService := &mockAnalysisService{
		sessionIDs: []string{sessionID},
		analyses:   map[string][]*claude_code.SessionAnalysis{sessionID: analyses},
	}

	plugin := claude_code.NewClaudeCodePlugin(
		analysisService,
		logsService,
		logger,
		nil,
		nil,
		"",
	)

	// Query for sessions
	query := pluginsdk.EntityQuery{Limit: 10}
	sessions, err := plugin.Query(ctx, query)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	// Verify analyses are included in derived state
	fields := sessions[0].GetAllFields()
	analysisCount, ok := fields["analysis_count"].(int)
	if !ok {
		t.Fatal("analysis_count field missing")
	}
	if analysisCount != 1 {
		t.Errorf("Expected 1 analysis, got %d", analysisCount)
	}

	// Verify status reflects analyzed state
	status, ok := fields["status"].(string)
	if !ok {
		t.Fatal("status field missing")
	}
	if status != "analyzed" {
		t.Errorf("Expected status 'analyzed', got %s", status)
	}

	t.Log("✓ Session includes analysis data in derived state")
	t.Log("✓ Status correctly reflects analysis presence")
	t.Log("✓ Complete derived state: events + analyses = current session")
}

// Note: Full integration tests with real database would be valuable but require:
// - SQLite test database setup
// - Real EventRepository implementation
// - Migration setup
// The tests above demonstrate the event sourcing pattern in isolation with mocks.
// The session_entity_test.go file provides comprehensive coverage of the
// SessionEntity logic which is the core functionality.

// TestClaudeCodePlugin_GetCapabilities verifies expected capabilities
func TestClaudeCodePlugin_GetCapabilities(t *testing.T) {
	plugin := claude_code.NewClaudeCodePlugin(nil, nil, &mockLogger{}, nil, nil, "")

	capabilities := plugin.GetCapabilities()

	// Verify we have the expected capabilities (no IHookProvider - hooks are plugin-specific, not framework)
	expectedCapabilities := map[string]bool{
		"IEntityProvider":  false,
		"IEntityUpdater":   false,
		"ICommandProvider": false,
	}

	for _, cap := range capabilities {
		if _, exists := expectedCapabilities[cap]; !exists {
			t.Errorf("Unexpected capability: %s", cap)
		}
		expectedCapabilities[cap] = true
	}

	// Verify all expected capabilities were found
	for cap, found := range expectedCapabilities {
		if !found {
			t.Errorf("Missing expected capability: %s", cap)
		}
	}
}

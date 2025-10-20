package infra_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteEventRepository_EmptyDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create empty database
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	store.Close()

	// Verify empty database can be queried
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	query := "SELECT id, timestamp, event_type, payload, content FROM events ORDER BY timestamp DESC LIMIT ?"
	rows, err := db.Query(query, 10)
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 rows in empty database, got %d", count)
	}
}

func TestSQLiteEventRepository_WithData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database with test data
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Insert test records
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	testRecords := []struct {
		id        string
		timestamp int64
		eventType string
		payload   string
		content   string
	}{
		{
			id:        "test-1",
			timestamp: time.Now().UnixMilli(),
			eventType: "chat.message.user",
			payload:   `{"message":"hello"}`,
			content:   "hello",
		},
		{
			id:        "test-2",
			timestamp: time.Now().UnixMilli() + 1000,
			eventType: "tool.invoked",
			payload:   `{"tool":"Read"}`,
			content:   "Read tool",
		},
		{
			id:        "test-3",
			timestamp: time.Now().UnixMilli() + 2000,
			eventType: "tool.result",
			payload:   `{"result":"success"}`,
			content:   "success",
		},
	}

	for _, r := range testRecords {
		_, err := db.Exec(
			"INSERT INTO events (id, timestamp, event_type, payload, content) VALUES (?, ?, ?, ?, ?)",
			r.id, r.timestamp, r.eventType, r.payload, r.content,
		)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Test query with limit
	query := "SELECT id, timestamp, event_type, payload, content FROM events ORDER BY timestamp DESC LIMIT ?"
	rows, err := db.Query(query, 2)
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id string
		var timestamp int64
		var eventType string
		var payload string
		var content string

		if err := rows.Scan(&id, &timestamp, &eventType, &payload, &content); err != nil {
			t.Errorf("Scan failed: %v", err)
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 rows with limit=2, got %d", count)
	}

	// Test query without limit (should get all)
	rows2, err := db.Query("SELECT id FROM events")
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	defer rows2.Close()

	count2 := 0
	for rows2.Next() {
		count2++
	}

	if count2 != 3 {
		t.Errorf("Expected 3 total rows, got %d", count2)
	}

	store.Close()
}

func TestSQLiteEventRepository_ExecuteRawQuery_SelectQuery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database with test data
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	store.Close()

	// Insert test data
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, payload, content) VALUES (?, ?, ?, ?, ?)",
		"test-1", time.Now().UnixMilli(), "chat.started", `{"session":"123"}`, "session started",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test that query executes successfully
	rows, err := db.Query("SELECT event_type, COUNT(*) as count FROM events GROUP BY event_type")
	if err != nil {
		t.Errorf("Aggregate query failed: %v", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			t.Errorf("Scan failed: %v", err)
		}
		if eventType == "chat.started" && count == 1 {
			found = true
		}
	}

	if !found {
		t.Error("Expected to find chat.started event with count=1")
	}
}

func TestSQLiteEventRepository_ExecuteRawQuery_InvalidQuery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	store.Close()

	// Test invalid query
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	_, err = db.Query("SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("Expected error for invalid query, got nil")
	}
}

func TestSQLiteEventRepository_TimestampFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database with test data
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	store.Close()

	// Insert test data with known timestamp
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	knownTime := time.Date(2025, 10, 17, 12, 30, 45, 0, time.UTC)
	timestamp := knownTime.UnixMilli()

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, payload, content) VALUES (?, ?, ?, ?, ?)",
		"test-ts", timestamp, "test.event", `{}`, "test",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Query and verify timestamp can be read
	rows, err := db.Query("SELECT timestamp FROM events WHERE id = ?", "test-ts")
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var ts int64
		if err := rows.Scan(&ts); err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		if ts != timestamp {
			t.Errorf("Expected timestamp %d, got %d", timestamp, ts)
		}

		// Verify we can convert it back to time
		retrievedTime := time.UnixMilli(ts)
		if !retrievedTime.Equal(knownTime) {
			t.Errorf("Expected time %v, got %v", knownTime, retrievedTime)
		}
	} else {
		t.Error("Expected to find row with timestamp")
	}
}

func TestSQLiteEventRepository_DatabaseSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	store.Close()

	// Verify schema
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	// Check that events table exists with correct columns
	rows, err := db.Query("PRAGMA table_info(events)")
	if err != nil {
		t.Fatalf("PRAGMA table_info failed: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]bool{
		"id":         false,
		"timestamp":  false,
		"event_type": false,
		"payload":    false,
		"content":    false,
	}

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Errorf("Scan failed: %v", err)
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
		}
	}

	// Verify all expected columns were found
	for col, found := range expectedColumns {
		if !found {
			t.Errorf("Expected column %q not found in events table", col)
		}
	}

	// Check that indexes exist
	rows2, err := db.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='events'")
	if err != nil {
		t.Fatalf("Query for indexes failed: %v", err)
	}
	defer rows2.Close()

	indexCount := 0
	for rows2.Next() {
		var name string
		if err := rows2.Scan(&name); err != nil {
			t.Errorf("Scan failed: %v", err)
		}
		indexCount++
	}

	// Should have at least a few indexes (excluding auto-created ones)
	if indexCount < 3 {
		t.Errorf("Expected at least 3 indexes, got %d", indexCount)
	}
}
func TestSQLiteEventRepository_FindByQuery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert test data directly
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	testTime := time.Now()
	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-1", testTime.UnixMilli(), "tool.invoked", "session-123", `{"tool":"Read"}`, "Read tool",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-2", testTime.Add(time.Second).UnixMilli(), "tool.result", "session-123", `{"result":"ok"}`, "result ok",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test FindByQuery with session filter
	query := domain.EventQuery{SessionID: "session-123"}
	events, err := store.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

func TestSQLiteEventRepository_GetAllSessionIDs(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert events with different session IDs
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	sessions := []string{"session-1", "session-2", "session-3"}
	for i, sessionID := range sessions {
		_, err = db.Exec(
			"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
			"evt-"+sessionID, time.Now().UnixMilli()+int64(i), "test.event", sessionID, `{}`, "test",
		)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Test GetAllSessionIDs
	sessionIDs, err := store.GetAllSessionIDs(ctx, 10)
	if err != nil {
		t.Fatalf("GetAllSessionIDs failed: %v", err)
	}

	if len(sessionIDs) != 3 {
		t.Errorf("Expected 3 session IDs, got %d", len(sessionIDs))
	}
}

func TestSQLiteEventRepository_GetAnalysesBySessionID(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert test analyses directly
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	sessionID := "test-session-456"
	_, err = db.Exec(
		"INSERT INTO session_analyses (id, session_id, analyzed_at, analysis_result, model_used, prompt_used, analysis_type, prompt_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"analysis-1", sessionID, time.Now().UnixMilli(), "Analysis result 1", "sonnet", "prompt1", "tool_analysis", "tool_analysis",
	)
	if err != nil {
		t.Fatalf("Insert analysis 1 failed: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO session_analyses (id, session_id, analyzed_at, analysis_result, model_used, prompt_used, analysis_type, prompt_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"analysis-2", sessionID, time.Now().Add(time.Hour).UnixMilli(), "Analysis result 2", "opus", "prompt2", "session_summary", "session_summary",
	)
	if err != nil {
		t.Fatalf("Insert analysis 2 failed: %v", err)
	}

	// Test GetAnalysesBySessionID
	analyses, err := store.GetAnalysesBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetAnalysesBySessionID failed: %v", err)
	}

	if len(analyses) != 2 {
		t.Errorf("Expected 2 analyses, got %d", len(analyses))
	}

	// Verify they're ordered by analyzed_at DESC (newest first)
	if len(analyses) >= 2 && analyses[0].AnalysisResult != "Analysis result 2" {
		t.Error("Expected analyses to be ordered by analyzed_at DESC")
	}
}

func TestSQLiteEventRepository_ExecuteRawQuery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert test data
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-1", time.Now().UnixMilli(), "tool.invoked", "session-789", `{"tool":"Bash"}`, "Bash command",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test ExecuteRawQuery
	result, err := store.ExecuteRawQuery(ctx, "SELECT COUNT(*) as count FROM events WHERE session_id = 'session-789'")
	if err != nil {
		t.Fatalf("ExecuteRawQuery failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Columns) == 0 {
		t.Error("Expected columns in result")
	}

	if len(result.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(result.Rows))
	}
}

func TestSQLiteEventRepository_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test Close
	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Calling Close again should be safe
	err = store.Close()
	if err != nil {
		t.Errorf("Second Close failed: %v", err)
	}
}

func TestSQLiteEventRepository_Save(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Create an event
	event := domain.NewEvent(domain.ToolInvoked, "session-save-test", map[string]string{"tool": "Read"}, "Read tool")

	// Save it
	err = store.Save(ctx, event)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify it was saved
	query := domain.EventQuery{SessionID: "session-save-test"}
	events, err := store.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].ID != event.ID {
		t.Errorf("Expected event ID %s, got %s", event.ID, events[0].ID)
	}
}

func TestSQLiteEventRepository_FindByQuery_WithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert multiple events
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		_, err = db.Exec(
			"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
			"evt-"+string(rune('0'+i)), time.Now().UnixMilli()+int64(i), "test.event", "limit-session", `{}`, "test",
		)
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Query with limit
	query := domain.EventQuery{
		SessionID: "limit-session",
		Limit:     5,
	}
	events, err := store.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery failed: %v", err)
	}

	if len(events) != 5 {
		t.Errorf("Expected 5 events with limit=5, got %d", len(events))
	}
}

func TestSQLiteEventRepository_FindByQuery_WithEventTypes(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert events with different types
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-tool", time.Now().UnixMilli(), "tool.invoked", "type-session", `{}`, "tool",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-chat", time.Now().UnixMilli(), "chat.message.user", "type-session", `{}`, "chat",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Query for specific event types
	query := domain.EventQuery{
		SessionID:  "type-session",
		EventTypes: []domain.EventType{domain.ToolInvoked},
	}
	events, err := store.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event of type tool.invoked, got %d", len(events))
	}

	if len(events) > 0 && events[0].Type != domain.ToolInvoked {
		t.Errorf("Expected event type tool.invoked, got %s", events[0].Type)
	}
}

func TestSQLiteEventRepository_FindByQuery_WithTimeRange(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert events at different times
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	baseTime := time.Now()
	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-old", baseTime.Add(-2*time.Hour).UnixMilli(), "test.event", "time-session", `{}`, "old",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-new", baseTime.UnixMilli(), "test.event", "time-session", `{}`, "new",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Query for events after 1 hour ago
	oneHourAgo := baseTime.Add(-1 * time.Hour)
	query := domain.EventQuery{
		SessionID: "time-session",
		StartTime: &oneHourAgo,
	}
	events, err := store.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event after time filter, got %d", len(events))
	}

	if len(events) > 0 && events[0].ID != "evt-new" {
		t.Errorf("Expected evt-new, got %s", events[0].ID)
	}
}

func TestSQLiteEventRepository_FindByQuery_Ordered(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer store.Close()

	// Insert events in non-chronological order
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	baseTime := time.Now()
	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-2", baseTime.Add(2*time.Second).UnixMilli(), "test.event", "order-session", `{}`, "second",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-1", baseTime.Add(1*time.Second).UnixMilli(), "test.event", "order-session", `{}`, "first",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Query with ordering
	query := domain.EventQuery{
		SessionID:   "order-session",
		OrderByTime: true,
	}
	events, err := store.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}

	// Should be ordered by time ASC
	if events[0].ID != "evt-1" {
		t.Errorf("Expected first event to be evt-1, got %s", events[0].ID)
	}
	if events[1].ID != "evt-2" {
		t.Errorf("Expected second event to be evt-2, got %s", events[1].ID)
	}
}

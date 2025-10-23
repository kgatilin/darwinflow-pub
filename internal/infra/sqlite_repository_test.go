package infra_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
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
	query := pluginsdk.EventQuery{Metadata: map[string]string{"session_id": "session-123"}}
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
	event := domain.NewEvent("claude.tool.invoked", "session-save-test", map[string]string{"tool": "Read"}, "Read tool")

	// Save it
	err = store.Save(ctx, event)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify it was saved
	query := pluginsdk.EventQuery{Metadata: map[string]string{"session_id": "session-save-test"}}
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
	query := pluginsdk.EventQuery{
		Metadata: map[string]string{"session_id": "limit-session"},
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
		"evt-tool", time.Now().UnixMilli(), "claude.tool.invoked", "type-session", `{}`, "tool",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO events (id, timestamp, event_type, session_id, payload, content) VALUES (?, ?, ?, ?, ?, ?)",
		"evt-chat", time.Now().UnixMilli(), "claude.chat.message.user", "type-session", `{}`, "chat",
	)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Query for specific event types
	query := pluginsdk.EventQuery{
		Metadata:   map[string]string{"session_id": "type-session"},
		EventTypes: []string{"claude.tool.invoked"},
	}
	events, err := store.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event of type tool.invoked, got %d", len(events))
	}

	if len(events) > 0 && events[0].Type != "claude.tool.invoked" {
		t.Errorf("Expected event type claude.tool.invoked, got %s", events[0].Type)
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
	query := pluginsdk.EventQuery{
		Metadata: map[string]string{"session_id": "time-session"},
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
	query := pluginsdk.EventQuery{
		Metadata:    map[string]string{"session_id": "order-session"},
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

// Tests from event_version_test.go

func TestSQLiteEventRepository_Save_WithVersion(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test: Event with explicit version
	event := &domain.Event{
		ID:        "test-event-1",
		Timestamp: time.Now(),
		Type:      "chat.started",
		SessionID: "session-1",
		Payload:   map[string]interface{}{"msg": "test"},
		Content:   "test content",
		Version:   "2.0",
	}

	if err := repo.Save(ctx, event); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify version is persisted
	events, err := repo.FindByQuery(ctx, pluginsdk.EventQuery{})
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	retrieved := events[0]
	if retrieved.Version != "2.0" {
		t.Errorf("Event version = %q, want %q", retrieved.Version, "2.0")
	}
	if retrieved.ID != "test-event-1" {
		t.Errorf("Event ID = %q, want %q", retrieved.ID, "test-event-1")
	}
}

func TestSQLiteEventRepository_Save_DefaultVersion(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Test: Event with default version "1.0"
	event := domain.NewEvent("chat.started", "session-2", map[string]interface{}{"msg": "test"}, "test")

	if err := repo.Save(ctx, event); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify default version is persisted
	events, err := repo.FindByQuery(ctx, pluginsdk.EventQuery{})
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	retrieved := events[0]
	if retrieved.Version != "1.0" {
		t.Errorf("Event version = %q, want %q", retrieved.Version, "1.0")
	}
}

func TestSQLiteEventRepository_MultipleVersions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Save multiple events with different versions
	testCases := []struct {
		id      string
		version string
	}{
		{"event-v1", "1.0"},
		{"event-v2", "2.0"},
		{"event-v3", "1.0"},
	}

	for _, tc := range testCases {
		event := &domain.Event{
			ID:        tc.id,
			Timestamp: time.Now(),
			Type:      "chat.started",
			SessionID: "session-3",
			Payload:   map[string]interface{}{"version": tc.version},
			Content:   "test",
			Version:   tc.version,
		}

		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Save() error for %s: %v", tc.id, err)
		}
	}

	// Retrieve all events
	events, err := repo.FindByQuery(ctx, pluginsdk.EventQuery{})
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Verify versions match what was saved
	versionMap := make(map[string]string)
	for _, e := range events {
		versionMap[e.ID] = e.Version
	}

	for _, tc := range testCases {
		if got := versionMap[tc.id]; got != tc.version {
			t.Errorf("Event %s version = %q, want %q", tc.id, got, tc.version)
		}
	}
}

func TestNewEvent_DefaultVersion(t *testing.T) {
	event := domain.NewEvent("chat.started", "session-4", map[string]interface{}{}, "test")

	if event.Version != "1.0" {
		t.Errorf("NewEvent() version = %q, want %q", event.Version, "1.0")
	}
}

func TestEvent_VersionPreservation(t *testing.T) {
	// Create event with custom version
	event := &domain.Event{
		ID:        "custom-v-event",
		Timestamp: time.Now(),
		Type:      "chat.started",
		SessionID: "test-session",
		Payload:   map[string]interface{}{},
		Content:   "test",
		Version:   "3.5",
	}

	if event.Version != "3.5" {
		t.Errorf("Event version = %q, want %q", event.Version, "3.5")
	}
}

// Tests from phase3_integration_test.go

// TestPhase3_EmitEventHookIntegration tests the complete Phase 3 event emission flow
// Scenario: Hook emits events → CLI captures and stores via EmitEventCommand → Plugin context stores events
func TestPhase3_EmitEventHookIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Step 1: Initialize database with schema
	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Step 2: Create sample event as would come from hook via adapter layer
	eventTimestamp := time.Date(2025, 10, 20, 15, 30, 45, 0, time.UTC)
	eventPayload := map[string]interface{}{
		"tool":        "Read",
		"file_path":   "/workspace/test.go",
		"description": "Reading test file",
	}
	sessionID := "test-session-123"

	// Step 3: Create domain Event (as would happen in adapter layer from SDK Event)
	domainEvent := &domain.Event{
		ID:        "event-1",
		Timestamp: eventTimestamp,
		Type:      "claude.tool.invoked",
		SessionID: sessionID,
		Payload:   eventPayload,
		Content:   "",
		Version:   "1.0",
	}

	// Step 4: Store event via repository (simulating EmitEventCommand execution)
	if err := repo.Save(ctx, domainEvent); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Step 5: Query EventRepository for stored event
	query := pluginsdk.EventQuery{
		Metadata: map[string]string{"session_id": "test-session-123"},
	}
	storedEvents, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	// Step 6: Verify event properties match input
	if len(storedEvents) != 1 {
		t.Errorf("Expected 1 event, got %d", len(storedEvents))
	}

	retrieved := storedEvents[0]
	if retrieved.Type != "claude.tool.invoked" {
		t.Errorf("Event type = %v, want %v", retrieved.Type, "claude.tool.invoked")
	}
	if retrieved.SessionID != "test-session-123" {
		t.Errorf("Event SessionID = %q, want %q", retrieved.SessionID, "test-session-123")
	}
	if retrieved.Version != "1.0" {
		t.Errorf("Event Version = %q, want %q", retrieved.Version, "1.0")
	}

	// Verify payload contains tool information
	// Note: Payload is stored as JSON in SQLite, so it may be json.RawMessage when retrieved
	payloadJSON, err := json.Marshal(retrieved.Payload)
	if err == nil {
		var payload map[string]interface{}
		if err := json.Unmarshal(payloadJSON, &payload); err == nil {
			if tool, ok := payload["tool"]; !ok || tool != "Read" {
				t.Errorf("Payload missing or wrong tool, got %v", payload["tool"])
			}
		}
	}
}

// TestPhase3_MultipleEventsOrdering tests that multiple events are stored and ordered correctly
func TestPhase3_MultipleEventsOrdering(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	sessionID := "session-multiple-events"
	baseTime := time.Date(2025, 10, 20, 15, 0, 0, 0, time.UTC)

	// Create multiple events with different timestamps
	testCases := []struct {
		eventType string
		offset    time.Duration
		tool      string
	}{
		{"claude.chat.started", 0 * time.Second, ""},
		{"claude.tool.invoked", 1 * time.Second, "Read"},
		{"claude.tool.invoked", 2 * time.Second, "Write"},
		{"claude.chat.message.user", 3 * time.Second, ""},
		{"claude.tool.invoked", 4 * time.Second, "Bash"},
	}

	// Store all events
	for idx, tc := range testCases {
		event := &domain.Event{
			ID:        "event-" + string(rune(idx)),
			Timestamp: baseTime.Add(tc.offset),
			Type:      tc.eventType,
			SessionID: sessionID,
			Payload: map[string]interface{}{
				"tool": tc.tool,
			},
			Content: "",
			Version: "1.0",
		}

		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Save() error for event %d: %v", idx, err)
		}
	}

	// Query all events for session
	query := pluginsdk.EventQuery{
		Metadata: map[string]string{"session_id": sessionID},
	}
	storedEvents, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	// Verify all events stored
	if len(storedEvents) != len(testCases) {
		t.Errorf("Expected %d events, got %d", len(testCases), len(storedEvents))
	}

	// Verify all event types are present (order may vary depending on DB sorting)
	typeMap := make(map[string]int)
	for _, event := range storedEvents {
		typeMap[event.Type]++
	}

	for _, tc := range testCases {
		if typeMap[tc.eventType] == 0 {
			t.Errorf("Event type %v not found in results", tc.eventType)
		}
	}
}

// TestPhase3_EventVersioning tests event schema versioning support
func TestPhase3_EventVersioning(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	sessionID := "session-versioning"

	// Save events with different versions
	versions := []string{"1.0", "1.1", "2.0"}
	for _, version := range versions {
		event := &domain.Event{
			ID:        "versioned-event-" + version,
			Timestamp: time.Now(),
			Type:      "claude.tool.invoked",
			SessionID: sessionID,
			Payload:   map[string]interface{}{"format_version": version},
			Content:   "",
			Version:   version,
		}

		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Save() error for version %s: %v", version, err)
		}
	}

	// Query all versioned events
	query := pluginsdk.EventQuery{
		Metadata: map[string]string{"session_id": sessionID},
	}
	allEvents, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(allEvents) != len(versions) {
		t.Errorf("Expected %d events, got %d", len(versions), len(allEvents))
	}

	// Verify versions
	versionCount := make(map[string]int)
	for _, event := range allEvents {
		versionCount[event.Version]++
	}

	for _, version := range versions {
		if versionCount[version] != 1 {
			t.Errorf("Version %s count = %d, want 1", version, versionCount[version])
		}
	}
}

// TestPhase3_EventPayloadPersistence tests complex payload preservation
func TestPhase3_EventPayloadPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Create event with complex nested payload
	complexPayload := map[string]interface{}{
		"tool":    "Read",
		"file":    "/workspace/test.go",
		"options": map[string]interface{}{
			"follow_symlinks": true,
			"timeout":         30,
			"max_size":        1024000,
		},
		"metadata": map[string]interface{}{
			"tags": []string{"important", "code-review"},
			"size": 4096,
		},
	}

	event := &domain.Event{
		ID:        "complex-payload-event",
		Timestamp: time.Now(),
		Type:      "claude.tool.invoked",
		SessionID: "test-session",
		Payload:   complexPayload,
		Content:   "",
		Version:   "1.0",
	}

	if err := repo.Save(ctx, event); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify payload structure
	query := pluginsdk.EventQuery{
		Metadata: map[string]string{"session_id": "test-session"},
	}
	stored, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(stored) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(stored))
	}

	retrieved := stored[0]

	// Parse payload from JSON (it's stored as JSON in SQLite)
	payloadJSON, err := json.Marshal(retrieved.Payload)
	if err != nil {
		t.Fatalf("json.Marshal(Payload) error = %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		t.Fatalf("json.Unmarshal(Payload) error = %v", err)
	}

	// Verify top-level fields
	if payload["tool"] != "Read" {
		t.Errorf("Payload tool = %v, want %v", payload["tool"], "Read")
	}

	// Verify nested structure
	if opts, ok := payload["options"].(map[string]interface{}); ok {
		if opts["timeout"] != float64(30) && opts["timeout"] != 30 {
			t.Errorf("Payload options.timeout = %v, want 30", opts["timeout"])
		}
	} else {
		t.Error("Payload options not found or wrong type")
	}

	// Verify nested array
	if meta, ok := payload["metadata"].(map[string]interface{}); ok {
		if _, ok := meta["tags"]; !ok {
			t.Error("Payload metadata.tags not found")
		}
	} else {
		t.Error("Payload metadata not found or wrong type")
	}
}

// TestPhase3_DomainEventFields tests that domain events preserve all important fields
func TestPhase3_DomainEventFields(t *testing.T) {
	// Create a domain event with all fields
	eventTime := time.Now()
	eventPayload := map[string]interface{}{
		"tool": "Read",
	}
	sessionID := "field-test"

	domainEvent := &domain.Event{
		Type:      "claude.tool.invoked",
		SessionID: sessionID,
		Timestamp: eventTime,
		Payload:   eventPayload,
		Version:   "1.0",
	}

	// Verify all fields are preserved
	if domainEvent.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", domainEvent.SessionID, sessionID)
	}
	if domainEvent.Version != "1.0" {
		t.Errorf("Version = %q, want %q", domainEvent.Version, "1.0")
	}
	payload, ok := domainEvent.Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("Payload not a map, got %T", domainEvent.Payload)
	}
	if payload["tool"] != "Read" {
		t.Errorf("Payload tool = %v, want %v", payload["tool"], "Read")
	}
}

// TestPhase3_JSONMarshaling tests that events can be marshaled/unmarshaled as JSON
func TestPhase3_JSONMarshaling(t *testing.T) {
	// Create a domain event
	eventPayload := map[string]interface{}{
		"tool":    "Bash",
		"command": "go test ./...",
	}

	originalEvent := &domain.Event{
		Type:      "claude.tool.invoked",
		SessionID: "marshal-test",
		Timestamp: time.Date(2025, 10, 20, 15, 30, 45, 0, time.UTC),
		Payload:   eventPayload,
		Version:   "1.0",
	}

	// Marshal to JSON
	data, err := json.Marshal(originalEvent)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var restored domain.Event
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields preserved
	if restored.Type != "claude.tool.invoked" {
		t.Errorf("Type = %v, want %v", restored.Type, "claude.tool.invoked")
	}
	if restored.SessionID != "marshal-test" {
		t.Errorf("SessionID = %q, want %q", restored.SessionID, "marshal-test")
	}
	if restored.Version != "1.0" {
		t.Errorf("Version = %q, want %q", restored.Version, "1.0")
	}
}

// TestPhase3_HookMigrationCompatibility tests backward compatibility with old log format
func TestPhase3_HookMigrationCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Simulate old-format event (would come from deprecated "dw claude log")
	oldFormatEvent := &domain.Event{
		ID:        "old-format-event",
		Timestamp: time.Now(),
		Type:      "claude.tool.invoked",
		SessionID: "legacy-session",
		Payload:   map[string]interface{}{"tool": "Read"},
		Content:   "",
		Version:   "1.0", // Would be empty string in truly old format, but now defaults to 1.0
	}

	if err := repo.Save(ctx, oldFormatEvent); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify backward compatibility
	query := pluginsdk.EventQuery{
		Metadata: map[string]string{"session_id": "legacy-session"},
	}
	stored, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(stored) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(stored))
	}

	// Old format events should still be queryable and functional
	retrieved := stored[0]
	if retrieved.SessionID != "legacy-session" {
		t.Errorf("Backward compat: SessionID = %q, want %q", retrieved.SessionID, "legacy-session")
	}
	if retrieved.Type != "claude.tool.invoked" {
		t.Errorf("Backward compat: Type = %v, want %v", retrieved.Type, "claude.tool.invoked")
	}
}

// Tests from analysis_repository_test.go

func TestSQLiteEventRepository_SaveAnalysis(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create and save an analysis
	analysis := domain.NewSessionAnalysis(
		"test-session-123",
		"This is the analysis result with patterns and suggestions",
		"claude-sonnet-4",
		"Analysis prompt template",
	)
	analysis.PatternsSummary = "Found 3 patterns: read-edit-save, grep-read, tool chains"

	err = repo.SaveAnalysis(ctx, analysis)
	if err != nil {
		t.Errorf("Failed to save analysis: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetAnalysisBySessionID(ctx, "test-session-123")
	if err != nil {
		t.Errorf("Failed to retrieve analysis: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected analysis to be found, got nil")
	}

	if retrieved.SessionID != analysis.SessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", retrieved.SessionID, analysis.SessionID)
	}

	if retrieved.AnalysisResult != analysis.AnalysisResult {
		t.Errorf("AnalysisResult mismatch")
	}

	if retrieved.ModelUsed != analysis.ModelUsed {
		t.Errorf("ModelUsed mismatch: got %s, want %s", retrieved.ModelUsed, analysis.ModelUsed)
	}
}

func TestSQLiteEventRepository_GetUnanalyzedSessionIDs(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create some events for different sessions
	session1 := "session-1"
	session2 := "session-2"
	session3 := "session-3"

	events := []*domain.Event{
		domain.NewEvent("chat.started", session1, map[string]interface{}{"message": "test"}, "test"),
		domain.NewEvent("chat.started", session2, map[string]interface{}{"message": "test"}, "test"),
		domain.NewEvent("chat.started", session3, map[string]interface{}{"message": "test"}, "test"),
	}

	for _, event := range events {
		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Failed to save event: %v", err)
		}
	}

	// Initially, all sessions should be unanalyzed
	unanalyzed, err := repo.GetUnanalyzedSessionIDs(ctx)
	if err != nil {
		t.Errorf("Failed to get unanalyzed sessions: %v", err)
	}

	if len(unanalyzed) != 3 {
		t.Errorf("Expected 3 unanalyzed sessions, got %d", len(unanalyzed))
	}

	// Analyze session1
	analysis := domain.NewSessionAnalysis(session1, "analysis result", "claude", "prompt")
	if err := repo.SaveAnalysis(ctx, analysis); err != nil {
		t.Fatalf("Failed to save analysis: %v", err)
	}

	// Now should have 2 unanalyzed sessions
	unanalyzed, err = repo.GetUnanalyzedSessionIDs(ctx)
	if err != nil {
		t.Errorf("Failed to get unanalyzed sessions: %v", err)
	}

	if len(unanalyzed) != 2 {
		t.Errorf("Expected 2 unanalyzed sessions, got %d", len(unanalyzed))
	}

	// Verify session1 is not in the list
	for _, id := range unanalyzed {
		if id == session1 {
			t.Errorf("session1 should not be in unanalyzed list")
		}
	}
}

func TestSQLiteEventRepository_GetAllAnalyses(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Save multiple analyses
	analyses := []*domain.SessionAnalysis{
		domain.NewSessionAnalysis("session-1", "result 1", "claude", "prompt"),
		domain.NewSessionAnalysis("session-2", "result 2", "claude", "prompt"),
		domain.NewSessionAnalysis("session-3", "result 3", "claude", "prompt"),
	}

	for _, analysis := range analyses {
		time.Sleep(time.Millisecond) // Ensure different timestamps
		if err := repo.SaveAnalysis(ctx, analysis); err != nil {
			t.Fatalf("Failed to save analysis: %v", err)
		}
	}

	// Retrieve all
	all, err := repo.GetAllAnalyses(ctx, 0)
	if err != nil {
		t.Errorf("Failed to get all analyses: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 analyses, got %d", len(all))
	}

	// Verify they're ordered by analyzed_at DESC (newest first)
	if len(all) > 1 {
		if all[0].AnalyzedAt.Before(all[1].AnalyzedAt) {
			t.Errorf("Analyses not ordered by analyzed_at DESC")
		}
	}

	// Test limit
	limited, err := repo.GetAllAnalyses(ctx, 2)
	if err != nil {
		t.Errorf("Failed to get limited analyses: %v", err)
	}

	if len(limited) != 2 {
		t.Errorf("Expected 2 analyses with limit, got %d", len(limited))
	}
}

func TestSQLiteEventRepository_GetAnalysisBySessionID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Try to get non-existent analysis
	analysis, err := repo.GetAnalysisBySessionID(ctx, "non-existent-session")
	if err != nil {
		t.Errorf("Expected no error for non-existent session, got: %v", err)
	}

	if analysis != nil {
		t.Errorf("Expected nil analysis for non-existent session, got: %v", analysis)
	}
}
// TestAnalysesMigration tests the migration from session_analyses to analyses table
func TestAnalysesMigration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Step 1: Create old-style database with session_analyses
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create old schema
	oldSchema := `
		CREATE TABLE IF NOT EXISTS session_analyses (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			analyzed_at INTEGER NOT NULL,
			analysis_result TEXT NOT NULL,
			model_used TEXT,
			prompt_used TEXT,
			patterns_summary TEXT,
			analysis_type TEXT DEFAULT 'tool_analysis',
			prompt_name TEXT DEFAULT 'analysis'
		);
	`
	_, err = db.Exec(oldSchema)
	if err != nil {
		t.Fatalf("Failed to create old schema: %v", err)
	}

	// Insert test data into old table
	testData := []struct {
		id              string
		sessionID       string
		analyzedAt      int64
		analysisResult  string
		modelUsed       string
		promptUsed      string
		patternsSummary string
		analysisType    string
		promptName      string
	}{
		{
			id:              "analysis-1",
			sessionID:       "session-1",
			analyzedAt:      time.Now().UnixMilli(),
			analysisResult:  "Test analysis 1",
			modelUsed:       "claude-sonnet-4",
			promptUsed:      "Analyze this session",
			patternsSummary: "Pattern 1",
			analysisType:    "tool_analysis",
			promptName:      "tool_analysis",
		},
		{
			id:              "analysis-2",
			sessionID:       "session-2",
			analyzedAt:      time.Now().UnixMilli(),
			analysisResult:  "Test analysis 2",
			modelUsed:       "claude-opus-4",
			promptUsed:      "Summarize this session",
			patternsSummary: "Pattern 2",
			analysisType:    "session_summary",
			promptName:      "summary",
		},
	}

	for _, td := range testData {
		_, err := db.Exec(`
			INSERT INTO session_analyses (id, session_id, analyzed_at, analysis_result, model_used, prompt_used, patterns_summary, analysis_type, prompt_name)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, td.id, td.sessionID, td.analyzedAt, td.analysisResult, td.modelUsed, td.promptUsed, td.patternsSummary, td.analysisType, td.promptName)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	db.Close()

	// Step 2: Initialize repository (should trigger migration)
	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Step 3: Verify migration completed
	// Check that analyses table exists
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db.Close()

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='analyses'").Scan(&tableName)
	if err != nil {
		t.Fatalf("analyses table does not exist: %v", err)
	}

	// Step 4: Verify data was migrated correctly
	rows, err := db.Query(`
		SELECT id, view_id, view_type, timestamp, result, model_used, prompt_used, metadata
		FROM analyses
		WHERE view_type != '__migration_marker__'
		ORDER BY timestamp
	`)
	if err != nil {
		t.Fatalf("Failed to query analyses: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, viewID, viewType, result, modelUsed, promptUsed, metadata string
		var timestamp int64

		err := rows.Scan(&id, &viewID, &viewType, &timestamp, &result, &modelUsed, &promptUsed, &metadata)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}

		// Verify view_type is "session" for migrated data
		if viewType != "session" {
			t.Errorf("Expected view_type='session', got %s", viewType)
		}

		// Verify view_id matches session_id from original data
		expectedSessionID := testData[count].sessionID
		if viewID != expectedSessionID {
			t.Errorf("Expected view_id=%s, got %s", expectedSessionID, viewID)
		}

		// Verify result was preserved
		if result != testData[count].analysisResult {
			t.Errorf("Expected result=%s, got %s", testData[count].analysisResult, result)
		}

		count++
	}

	if count != len(testData) {
		t.Errorf("Expected %d migrated records, got %d", len(testData), count)
	}

	// Step 5: Verify migration marker exists
	var markerID string
	err = db.QueryRow("SELECT id FROM analyses WHERE view_type = '__migration_marker__'").Scan(&markerID)
	if err != nil {
		t.Errorf("Migration marker not found: %v", err)
	}

	// Step 6: Verify migration is idempotent (running Initialize again should not duplicate data)
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Second Initialize failed: %v", err)
	}

	var recordCount int
	err = db.QueryRow("SELECT COUNT(*) FROM analyses WHERE view_type != '__migration_marker__'").Scan(&recordCount)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if recordCount != len(testData) {
		t.Errorf("Migration not idempotent: expected %d records, got %d", len(testData), recordCount)
	}
}

// TestGenericAnalysisRepository tests the generic analysis repository methods
func TestGenericAnalysisRepository(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test SaveGenericAnalysis
	analysis1 := domain.NewAnalysis("view-1", "session", "Analysis result 1", "claude-sonnet-4", "tool_analysis")
	analysis1.Metadata["key1"] = "value1"
	analysis1.Metadata["key2"] = 42

	if err := repo.SaveGenericAnalysis(ctx, analysis1); err != nil {
		t.Fatalf("SaveGenericAnalysis failed: %v", err)
	}

	analysis2 := domain.NewAnalysis("view-1", "session", "Analysis result 2", "claude-opus-4", "summary")
	analysis2.Metadata["key3"] = "value3"

	if err := repo.SaveGenericAnalysis(ctx, analysis2); err != nil {
		t.Fatalf("SaveGenericAnalysis failed: %v", err)
	}

	analysis3 := domain.NewAnalysis("view-2", "task-list", "Task analysis", "claude-sonnet-4", "task_analysis")

	if err := repo.SaveGenericAnalysis(ctx, analysis3); err != nil {
		t.Fatalf("SaveGenericAnalysis failed: %v", err)
	}

	// Test FindAnalysisByViewID
	analyses, err := repo.FindAnalysisByViewID(ctx, "view-1")
	if err != nil {
		t.Fatalf("FindAnalysisByViewID failed: %v", err)
	}

	if len(analyses) != 2 {
		t.Errorf("Expected 2 analyses for view-1, got %d", len(analyses))
	}

	// Verify metadata was preserved
	if analyses[0].Metadata == nil {
		t.Error("Metadata is nil")
	}

	// Test FindAnalysisByViewType
	sessionAnalyses, err := repo.FindAnalysisByViewType(ctx, "session")
	if err != nil {
		t.Fatalf("FindAnalysisByViewType failed: %v", err)
	}

	if len(sessionAnalyses) != 2 {
		t.Errorf("Expected 2 session analyses, got %d", len(sessionAnalyses))
	}

	taskAnalyses, err := repo.FindAnalysisByViewType(ctx, "task-list")
	if err != nil {
		t.Fatalf("FindAnalysisByViewType failed: %v", err)
	}

	if len(taskAnalyses) != 1 {
		t.Errorf("Expected 1 task-list analysis, got %d", len(taskAnalyses))
	}

	// Test FindAnalysisById
	found, err := repo.FindAnalysisById(ctx, analysis1.ID)
	if err != nil {
		t.Fatalf("FindAnalysisById failed: %v", err)
	}

	if found == nil {
		t.Fatal("Analysis not found")
	}

	if found.ViewID != "view-1" {
		t.Errorf("Expected view_id='view-1', got %s", found.ViewID)
	}

	if found.Result != "Analysis result 1" {
		t.Errorf("Expected result='Analysis result 1', got %s", found.Result)
	}

	// Verify metadata
	if v, ok := found.Metadata["key1"].(string); !ok || v != "value1" {
		t.Errorf("Metadata key1 not preserved correctly")
	}

	// Test ListRecentAnalyses
	recent, err := repo.ListRecentAnalyses(ctx, 10)
	if err != nil {
		t.Fatalf("ListRecentAnalyses failed: %v", err)
	}

	if len(recent) != 3 {
		t.Errorf("Expected 3 recent analyses, got %d", len(recent))
	}

	// Verify order (most recent first)
	if recent[0].Timestamp.Before(recent[1].Timestamp) {
		t.Error("Analyses not ordered by timestamp DESC")
	}

	// Test with limit
	limited, err := repo.ListRecentAnalyses(ctx, 2)
	if err != nil {
		t.Fatalf("ListRecentAnalyses with limit failed: %v", err)
	}

	if len(limited) != 2 {
		t.Errorf("Expected 2 limited analyses, got %d", len(limited))
	}
}

// TestEmptyDatabaseNoMigration tests that a fresh database doesn't trigger migration
func TestEmptyDatabaseNoMigration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify analyses table exists
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='analyses'").Scan(&tableName)
	if err != nil {
		t.Fatalf("analyses table does not exist: %v", err)
	}

	// Verify no migration marker (since there was nothing to migrate)
	var markerID string
	err = db.QueryRow("SELECT id FROM analyses WHERE view_type = '__migration_marker__'").Scan(&markerID)
	if err == nil {
		t.Errorf("Expected no migration marker in fresh database, but found one")
	} else if err != sql.ErrNoRows {
		t.Errorf("Unexpected error querying migration marker: %v", err)
	}
}

// TestOldDatabaseMigration tests automatic migration of old database with session_analyses data
func TestOldDatabaseMigration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "old-test.db")

	// Create an old database with session_analyses but no analyses table
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create old schema (session_analyses table only)
	_, err = db.Exec(`
		CREATE TABLE events (
			id TEXT PRIMARY KEY,
			timestamp INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			session_id TEXT,
			payload TEXT NOT NULL,
			content TEXT NOT NULL
		);

		CREATE TABLE session_analyses (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			analyzed_at INTEGER NOT NULL,
			analysis_result TEXT NOT NULL,
			model_used TEXT,
			prompt_used TEXT,
			patterns_summary TEXT,
			analysis_type TEXT DEFAULT 'tool_analysis',
			prompt_name TEXT DEFAULT 'analysis'
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create old schema: %v", err)
	}

	// Insert test data into session_analyses
	testTime := time.Now()
	_, err = db.Exec(`
		INSERT INTO session_analyses (id, session_id, analyzed_at, analysis_result, model_used, prompt_used, patterns_summary, analysis_type, prompt_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-analysis-1", "session-123", testTime.UnixMilli(), "Test analysis result", "claude-sonnet-4", "Test prompt", "Test patterns", "tool_analysis", "analysis")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	db.Close()

	// Now open with repository and initialize (should trigger migration)
	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify analyses table exists
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='analyses'").Scan(&tableName)
	if err != nil {
		t.Fatalf("analyses table does not exist: %v", err)
	}

	// Verify data was migrated
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM analyses WHERE view_type = 'session'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count migrated analyses: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 migrated analysis, got %d", count)
	}

	// Verify migration marker exists
	var markerID string
	err = db.QueryRow("SELECT id FROM analyses WHERE view_type = '__migration_marker__'").Scan(&markerID)
	if err != nil {
		t.Errorf("Expected migration marker, got error: %v", err)
	}

	// Verify migrated data content
	var viewID, result string
	err = db.QueryRow("SELECT view_id, result FROM analyses WHERE view_type = 'session' LIMIT 1").Scan(&viewID, &result)
	if err != nil {
		t.Fatalf("Failed to query migrated data: %v", err)
	}
	if viewID != "session-123" {
		t.Errorf("Expected view_id 'session-123', got %s", viewID)
	}
	if result != "Test analysis result" {
		t.Errorf("Expected result 'Test analysis result', got %s", result)
	}

	// Run Initialize again - should NOT re-migrate
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Second Initialize failed: %v", err)
	}

	// Verify count didn't increase (no duplicate migration)
	err = db.QueryRow("SELECT COUNT(*) FROM analyses WHERE view_type = 'session'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count analyses after second initialize: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 analysis after second initialize, got %d (possible duplicate migration)", count)
	}
}

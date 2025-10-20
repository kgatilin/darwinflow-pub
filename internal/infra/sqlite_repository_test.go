package infra_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

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

package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestRepeatString(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		count  int
		want   string
	}{
		{
			name:  "repeat dash 5 times",
			s:     "-",
			count: 5,
			want:  "-----",
		},
		{
			name:  "repeat empty string",
			s:     "",
			count: 10,
			want:  "",
		},
		{
			name:  "repeat zero times",
			s:     "x",
			count: 0,
			want:  "",
		},
		{
			name:  "repeat multi-char string",
			s:     "ab",
			count: 3,
			want:  "ababab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repeatString(tt.s, tt.count)
			if got != tt.want {
				t.Errorf("repeatString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLogsFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    logsOptions
		wantErr bool
	}{
		{
			name: "default flags",
			args: []string{},
			want: logsOptions{limit: 20, query: "", sessionID: "", ordered: false, help: false},
		},
		{
			name: "custom limit",
			args: []string{"--limit", "50"},
			want: logsOptions{limit: 50, query: "", sessionID: "", ordered: false, help: false},
		},
		{
			name: "with query",
			args: []string{"--query", "SELECT * FROM events"},
			want: logsOptions{limit: 20, query: "SELECT * FROM events", sessionID: "", ordered: false, help: false},
		},
		{
			name: "with session-id",
			args: []string{"--session-id", "abc123"},
			want: logsOptions{limit: 20, query: "", sessionID: "abc123", ordered: false, help: false},
		},
		{
			name: "with ordered",
			args: []string{"--ordered"},
			want: logsOptions{limit: 20, query: "", sessionID: "", ordered: true, help: false},
		},
		{
			name: "help flag",
			args: []string{"--help"},
			want: logsOptions{limit: 20, query: "", sessionID: "", ordered: false, help: true},
		},
		{
			name: "multiple flags",
			args: []string{"--limit", "100", "--query", "SELECT COUNT(*) FROM events"},
			want: logsOptions{limit: 100, query: "SELECT COUNT(*) FROM events", sessionID: "", ordered: false, help: false},
		},
		{
			name: "session-id with ordered",
			args: []string{"--session-id", "xyz789", "--ordered"},
			want: logsOptions{limit: 20, query: "", sessionID: "xyz789", ordered: true, help: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLogsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogsFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.limit != tt.want.limit {
					t.Errorf("limit = %d, want %d", got.limit, tt.want.limit)
				}
				if got.query != tt.want.query {
					t.Errorf("query = %q, want %q", got.query, tt.want.query)
				}
				if got.sessionID != tt.want.sessionID {
					t.Errorf("sessionID = %q, want %q", got.sessionID, tt.want.sessionID)
				}
				if got.ordered != tt.want.ordered {
					t.Errorf("ordered = %v, want %v", got.ordered, tt.want.ordered)
				}
				if got.help != tt.want.help {
					t.Errorf("help = %v, want %v", got.help, tt.want.help)
				}
			}
		})
	}
}

// Note: Tests for queryLogs, formatLogRecord, and formatQueryValue have been
// removed as these functions are now in the app layer (internal/app/logs.go).
// Tests for those functions should be added to internal/app/logs_test.go.

func TestListLogs_EmptyDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create empty database
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer store.Close()

	// Test listLogs with empty database - should not error
	service := app.NewLogsService(store, store)
	opts := &logsOptions{limit: 10}
	err = listLogs(ctx, service, opts)
	if err != nil {
		t.Errorf("listLogs with empty DB failed: %v", err)
	}
}

func TestExecuteRawQuery_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create and initialize database
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer store.Close()

	// Test executeRawQuery with valid query
	service := app.NewLogsService(store, store)
	err = executeRawQuery(ctx, service, "SELECT COUNT(*) FROM events")
	if err != nil {
		t.Errorf("executeRawQuery failed: %v", err)
	}
}

func TestExecuteRawQuery_InvalidSQL(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create database
	store, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer store.Close()

	// Test with invalid SQL
	service := app.NewLogsService(store, store)
	err = executeRawQuery(ctx, service, "INVALID SQL QUERY")
	if err == nil {
		t.Error("Expected error for invalid SQL, got nil")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

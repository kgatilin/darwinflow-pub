package main_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	main "github.com/kgatilin/darwinflow-pub/cmd/dw"
	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestParseLogsFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    main.LogsOptions
		wantErr bool
	}{
		{
			name: "default flags",
			args: []string{},
			want: main.LogsOptions{Limit: 20, SessionLimit: 0, Query: "", SessionID: "", Ordered: false, Format: "text", Help: false},
		},
		{
			name: "custom limit",
			args: []string{"--limit", "50"},
			want: main.LogsOptions{Limit: 50, SessionLimit: 0, Query: "", SessionID: "", Ordered: false, Format: "text", Help: false},
		},
		{
			name: "custom session-limit",
			args: []string{"--session-limit", "3"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 3, Query: "", SessionID: "", Ordered: false, Format: "text", Help: false},
		},
		{
			name: "with query",
			args: []string{"--query", "SELECT * FROM events"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 0, Query: "SELECT * FROM events", SessionID: "", Ordered: false, Format: "text", Help: false},
		},
		{
			name: "with session-id",
			args: []string{"--session-id", "abc123"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 0, Query: "", SessionID: "abc123", Ordered: false, Format: "text", Help: false},
		},
		{
			name: "with ordered",
			args: []string{"--ordered"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 0, Query: "", SessionID: "", Ordered: true, Format: "text", Help: false},
		},
		{
			name: "help flag",
			args: []string{"--help"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 0, Query: "", SessionID: "", Ordered: false, Format: "text", Help: true},
		},
		{
			name: "multiple flags",
			args: []string{"--limit", "100", "--query", "SELECT COUNT(*) FROM events"},
			want: main.LogsOptions{Limit: 100, SessionLimit: 0, Query: "SELECT COUNT(*) FROM events", SessionID: "", Ordered: false, Format: "text", Help: false},
		},
		{
			name: "session-id with ordered",
			args: []string{"--session-id", "xyz789", "--ordered"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 0, Query: "", SessionID: "xyz789", Ordered: true, Format: "text", Help: false},
		},
		{
			name: "session-limit with format markdown",
			args: []string{"--session-limit", "5", "--format", "markdown"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 5, Query: "", SessionID: "", Ordered: false, Format: "markdown", Help: false},
		},
		{
			name: "csv format",
			args: []string{"--format", "csv"},
			want: main.LogsOptions{Limit: 20, SessionLimit: 0, Query: "", SessionID: "", Ordered: false, Format: "csv", Help: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := main.ParseLogsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogsFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Limit != tt.want.Limit {
					t.Errorf("Limit = %d, want %d", got.Limit, tt.want.Limit)
				}
				if got.SessionLimit != tt.want.SessionLimit {
					t.Errorf("SessionLimit = %d, want %d", got.SessionLimit, tt.want.SessionLimit)
				}
				if got.Query != tt.want.Query {
					t.Errorf("Query = %q, want %q", got.Query, tt.want.Query)
				}
				if got.SessionID != tt.want.SessionID {
					t.Errorf("SessionID = %q, want %q", got.SessionID, tt.want.SessionID)
				}
				if got.Ordered != tt.want.Ordered {
					t.Errorf("Ordered = %v, want %v", got.Ordered, tt.want.Ordered)
				}
				if got.Format != tt.want.Format {
					t.Errorf("Format = %q, want %q", got.Format, tt.want.Format)
				}
				if got.Help != tt.want.Help {
					t.Errorf("Help = %v, want %v", got.Help, tt.want.Help)
				}
			}
		})
	}
}

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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test ListLogs with empty database - should not error
	service := app.NewLogsService(store, store)
	opts := &main.LogsOptions{Limit: 10, SessionLimit: 0}
	err = main.ListLogs(ctx, service, opts)

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("ListLogs with empty DB failed: %v", err)
	}

	// Should display "No logs found" message
	if !strings.Contains(string(output), "No logs found") {
		t.Errorf("Expected 'No logs found' message, got: %s", output)
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test ExecuteRawQuery with valid query
	service := app.NewLogsService(store, store)
	err = main.ExecuteRawQuery(ctx, service, "SELECT COUNT(*) FROM events")

	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)

	if err != nil {
		t.Errorf("ExecuteRawQuery failed: %v", err)
	}

	// Should have column headers and row count
	outputStr := string(output)
	if !strings.Contains(outputStr, "rows") {
		t.Errorf("Expected row count in output, got: %s", outputStr)
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
	err = main.ExecuteRawQuery(ctx, service, "INVALID SQL QUERY")
	if err == nil {
		t.Error("Expected error for invalid SQL, got nil")
	}
}

func TestParseLogsFlags_InvalidFlag(t *testing.T) {
	// Test with an invalid flag
	_, err := main.ParseLogsFlags([]string{"--invalid-flag"})
	if err == nil {
		t.Error("Expected error for invalid flag, got nil")
	}
}

func TestParseLogsFlags_AllFormats(t *testing.T) {
	formats := []string{"text", "csv", "markdown"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			opts, err := main.ParseLogsFlags([]string{"--format", format})
			if err != nil {
				t.Errorf("ParseLogsFlags failed for format %s: %v", format, err)
			}
			if opts.Format != format {
				t.Errorf("Expected format %s, got %s", format, opts.Format)
			}
		})
	}
}

package main_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	main "github.com/kgatilin/darwinflow-pub/cmd/dw"
)

// Helper function to capture stdout
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

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

// Note: ListLogs and ExecuteRawQuery functionality is now tested in internal/app/logs_cmd_test.go
// The logic has been moved to the app layer for better testability

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

// PrintLogsHelp tests

func TestPrintLogsHelp(t *testing.T) {
	output := captureStdout(func() {
		main.PrintLogsHelp()
	})

	expectedStrings := []string{
		"DarwinFlow Logs",
		"DATABASE STRUCTURE",
		"Table: events",
		"Columns:",
		"event_type",
		"session_id",
		"COMMON EVENT TYPES",
		"QUERY EXAMPLES",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("PrintLogsHelp output should contain %q, but it doesn't", expected)
		}
	}
}

func TestPrintLogsHelp_ContainsSQLExamples(t *testing.T) {
	output := captureStdout(func() {
		main.PrintLogsHelp()
	})

	// Should have SQL query examples
	if !strings.Contains(output, "SELECT") {
		t.Error("PrintLogsHelp should include SQL query examples with SELECT")
	}
	if !strings.Contains(output, "COUNT") {
		t.Error("PrintLogsHelp should include examples using COUNT")
	}
}

func TestPrintLogsHelp_DescribesEventTypes(t *testing.T) {
	output := captureStdout(func() {
		main.PrintLogsHelp()
	})

	// Should describe common event types
	expectedTypes := []string{"tool.invoked", "chat.message.user", "file.read"}
	for _, eventType := range expectedTypes {
		if !strings.Contains(output, eventType) {
			t.Errorf("PrintLogsHelp should mention event type %q", eventType)
		}
	}
}

// Additional comprehensive tests

func TestFlagCombinations(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "all flags together",
			args: []string{
				"--limit", "100",
				"--session-limit", "5",
				"--session-id", "test-123",
				"--ordered",
				"--format", "csv",
			},
		},
		{
			name: "query with other flags",
			args: []string{
				"--query", "SELECT * FROM events",
				"--format", "text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := main.ParseLogsFlags(tt.args)
			if err != nil {
				t.Errorf("ParseLogsFlags() failed: %v", err)
			}
			if opts == nil {
				t.Error("Expected non-nil options")
			}
		})
	}
}

func TestParseLogsFlags_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "negative limit",
			args:    []string{"--limit", "-1"},
			wantErr: false, // Flag parsing doesn't validate negative values
		},
		{
			name:    "very large limit",
			args:    []string{"--limit", "999999"},
			wantErr: false,
		},
		{
			name:    "empty query",
			args:    []string{"--query", ""},
			wantErr: false,
		},
		{
			name:    "empty session id",
			args:    []string{"--session-id", ""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := main.ParseLogsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogsFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && opts == nil {
				t.Error("Expected non-nil options when no error")
			}
		})
	}
}

func TestParseLogsFlags_BooleanFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		checkFn  func(*main.LogsOptions) bool
		expected bool
	}{
		{
			name:     "ordered flag true",
			args:     []string{"--ordered"},
			checkFn:  func(o *main.LogsOptions) bool { return o.Ordered },
			expected: true,
		},
		{
			name:     "ordered flag false (default)",
			args:     []string{},
			checkFn:  func(o *main.LogsOptions) bool { return o.Ordered },
			expected: false,
		},
		{
			name:     "help flag true",
			args:     []string{"--help"},
			checkFn:  func(o *main.LogsOptions) bool { return o.Help },
			expected: true,
		},
		{
			name:     "help flag false (default)",
			args:     []string{},
			checkFn:  func(o *main.LogsOptions) bool { return o.Help },
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := main.ParseLogsFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseLogsFlags() failed: %v", err)
			}
			got := tt.checkFn(opts)
			if got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestParseLogsFlags_Defaults(t *testing.T) {
	opts, err := main.ParseLogsFlags([]string{})
	if err != nil {
		t.Fatalf("ParseLogsFlags() failed: %v", err)
	}

	// Test all default values
	if opts.Limit != 20 {
		t.Errorf("Default Limit should be 20, got %d", opts.Limit)
	}
	if opts.SessionLimit != 0 {
		t.Errorf("Default SessionLimit should be 0, got %d", opts.SessionLimit)
	}
	if opts.Query != "" {
		t.Errorf("Default Query should be empty, got %q", opts.Query)
	}
	if opts.SessionID != "" {
		t.Errorf("Default SessionID should be empty, got %q", opts.SessionID)
	}
	if opts.Ordered {
		t.Errorf("Default Ordered should be false, got true")
	}
	if opts.Format != "text" {
		t.Errorf("Default Format should be 'text', got %q", opts.Format)
	}
	if opts.Help {
		t.Errorf("Default Help should be false, got true")
	}
}

func TestParseLogsFlags_FlagValues(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "limit with non-numeric value",
			args:    []string{"--limit", "abc"},
			wantErr: true,
		},
		{
			name:    "session-limit with non-numeric value",
			args:    []string{"--session-limit", "xyz"},
			wantErr: true,
		},
		{
			name:    "valid numeric limit",
			args:    []string{"--limit", "50"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := main.ParseLogsFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLogsFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseLogsFlags_QueryWithSpecialChars(t *testing.T) {
	specialQueries := []struct {
		name  string
		query string
	}{
		{name: "with quotes", query: "SELECT * FROM events WHERE event_type = 'tool.invoked'"},
		{name: "with newlines", query: "SELECT *\nFROM events\nWHERE session_id = 'abc'"},
		{name: "with semicolons", query: "SELECT COUNT(*); SELECT MAX(timestamp);"},
		{name: "with special chars", query: "SELECT * FROM events WHERE content LIKE '%test%'"},
	}

	for _, tt := range specialQueries {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := main.ParseLogsFlags([]string{"--query", tt.query})
			if err != nil {
				t.Errorf("ParseLogsFlags() failed: %v", err)
			}
			if opts.Query != tt.query {
				t.Errorf("Query not preserved: expected %q, got %q", tt.query, opts.Query)
			}
		})
	}
}

func TestParseLogsFlags_SessionIDFormats(t *testing.T) {
	sessionIDs := []string{
		"simple",
		"with-dashes",
		"with_underscores",
		"123-456-789",
		"UUID-LIKE-FORMAT-HERE",
		"mixedCaseID",
	}

	for _, sid := range sessionIDs {
		t.Run(sid, func(t *testing.T) {
			opts, err := main.ParseLogsFlags([]string{"--session-id", sid})
			if err != nil {
				t.Errorf("ParseLogsFlags() failed: %v", err)
			}
			if opts.SessionID != sid {
				t.Errorf("SessionID not preserved: expected %q, got %q", sid, opts.SessionID)
			}
		})
	}
}

func TestParseLogsFlags_OrderIndependence(t *testing.T) {
	// These should produce the same result
	args1 := []string{"--limit", "50", "--format", "csv", "--ordered"}
	args2 := []string{"--ordered", "--format", "csv", "--limit", "50"}
	args3 := []string{"--format", "csv", "--ordered", "--limit", "50"}

	opts1, err1 := main.ParseLogsFlags(args1)
	opts2, err2 := main.ParseLogsFlags(args2)
	opts3, err3 := main.ParseLogsFlags(args3)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("ParseLogsFlags() failed: %v, %v, %v", err1, err2, err3)
	}

	// Check all produce same result
	if opts1.Limit != opts2.Limit || opts2.Limit != opts3.Limit {
		t.Error("Limit differs based on flag order")
	}
	if opts1.Format != opts2.Format || opts2.Format != opts3.Format {
		t.Error("Format differs based on flag order")
	}
	if opts1.Ordered != opts2.Ordered || opts2.Ordered != opts3.Ordered {
		t.Error("Ordered differs based on flag order")
	}
}

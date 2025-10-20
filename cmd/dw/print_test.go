package main_test

import (
	"bytes"
	"io"
	"os"
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

func TestRepeatString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		count  int
		want   string
	}{
		{name: "repeat dash 5 times", input: "-", count: 5, want: "-----"},
		{name: "repeat equals 3 times", input: "=", count: 3, want: "==="},
		{name: "repeat zero times", input: "x", count: 0, want: ""},
		{name: "repeat multiple chars", input: "ab", count: 3, want: "ababab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: repeatString is not exported, so we can't test it directly
			// This test is here as documentation of what we'd like to test
			// The function is already tested indirectly via ExecuteRawQuery tests
			t.Skip("repeatString is not exported, tested indirectly")
		})
	}
}

// Test that we can parse all supported formats
func TestAllSupportedFormats(t *testing.T) {
	supportedFormats := []string{"text", "csv", "markdown"}

	for _, format := range supportedFormats {
		opts, err := main.ParseLogsFlags([]string{"--format", format})
		if err != nil {
			t.Errorf("Format %q should be supported but got error: %v", format, err)
		}
		if opts.Format != format {
			t.Errorf("Expected format %q, got %q", format, opts.Format)
		}
	}
}

// Test flag combinations
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

// Test edge cases
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

// Test boolean flags
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

// Test default values
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

// Test that unknown flags produce errors
func TestParseLogsFlags_UnknownFlags(t *testing.T) {
	unknownFlags := []string{
		"--unknown",
		"--invalid-flag",
		"--does-not-exist",
	}

	for _, flag := range unknownFlags {
		t.Run(flag, func(t *testing.T) {
			_, err := main.ParseLogsFlags([]string{flag})
			if err == nil {
				t.Errorf("Expected error for unknown flag %q, got nil", flag)
			}
		})
	}
}

// Test flag value validation
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

// Test that format values are case-sensitive
func TestParseLogsFlags_FormatCaseSensitive(t *testing.T) {
	// These should all work (lowercase is standard)
	validFormats := []string{"text", "csv", "markdown"}
	for _, format := range validFormats {
		opts, err := main.ParseLogsFlags([]string{"--format", format})
		if err != nil {
			t.Errorf("Format %q should be valid, got error: %v", format, err)
		}
		if opts.Format != format {
			t.Errorf("Format should be %q, got %q", format, opts.Format)
		}
	}

	// These are not validated by the flag parser but by ListLogs
	// They will be accepted by ParseLogsFlags but rejected by ListLogs
	upperFormats := []string{"TEXT", "CSV", "MARKDOWN"}
	for _, format := range upperFormats {
		opts, err := main.ParseLogsFlags([]string{"--format", format})
		if err != nil {
			t.Errorf("ParseLogsFlags should accept any string, got error: %v", err)
		}
		if opts.Format != format {
			t.Errorf("Format should be %q, got %q", format, opts.Format)
		}
	}
}

// Test multi-value flags
func TestParseLogsFlags_MultiValueFlags(t *testing.T) {
	// Test that providing a flag multiple times uses the last value
	opts, err := main.ParseLogsFlags([]string{
		"--limit", "10",
		"--limit", "20",
		"--limit", "30",
	})
	if err != nil {
		t.Fatalf("ParseLogsFlags() failed: %v", err)
	}
	if opts.Limit != 30 {
		t.Errorf("Expected last limit value 30, got %d", opts.Limit)
	}
}

// Test query with special characters
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

// Test that session-id accepts various formats
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

// Test flag order independence
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

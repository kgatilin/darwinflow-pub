package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

type logsOptions struct {
	limit     int
	query     string
	sessionID string
	ordered   bool
	format    string
	help      bool
}

func parseLogsFlags(args []string) (*logsOptions, error) {
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	opts := &logsOptions{}

	fs.IntVar(&opts.limit, "limit", 20, "Number of most recent logs to display")
	fs.StringVar(&opts.query, "query", "", "Arbitrary SQL query to execute")
	fs.StringVar(&opts.sessionID, "session-id", "", "Filter logs by session ID")
	fs.BoolVar(&opts.ordered, "ordered", false, "Order by timestamp ASC and session ID (chronological)")
	fs.StringVar(&opts.format, "format", "text", "Output format: text, csv, or markdown")
	fs.BoolVar(&opts.help, "help", false, "Show help and database schema")

	fs.Usage = printLogsUsage

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opts, nil
}

func handleLogs(args []string) {
	opts, err := parseLogsFlags(args)
	if err != nil {
		os.Exit(1)
	}

	// Show help if requested
	if opts.help {
		printLogsHelp()
		return
	}

	dbPath := app.DefaultDBPath

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Database not found at %s\n", dbPath)
		fmt.Fprintf(os.Stderr, "Run 'dw claude init' to initialize logging.\n")
		os.Exit(1)
	}

	// Initialize repository and service
	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	service := app.NewLogsService(repo, repo)
	ctx := context.Background()

	// Handle arbitrary SQL query
	if opts.query != "" {
		if err := executeRawQuery(ctx, service, opts.query); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle standard log listing
	if err := listLogs(ctx, service, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printLogsUsage() {
	fmt.Println("Usage: dw logs [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --limit N          Number of most recent logs to display (default: 20)")
	fmt.Println("  --session-id ID    Filter logs by session ID")
	fmt.Println("  --ordered          Order by timestamp ASC and session ID (chronological)")
	fmt.Println("  --format FORMAT    Output format: text, csv, or markdown (default: text)")
	fmt.Println("  --query SQL        Execute an arbitrary SQL query")
	fmt.Println("  --help             Show help and database schema")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dw logs                                          # Show 20 most recent logs")
	fmt.Println("  dw logs --limit 50                               # Show 50 most recent logs")
	fmt.Println("  dw logs --session-id abc123                      # Show logs for session abc123")
	fmt.Println("  dw logs --session-id abc123 --ordered            # Show session abc123 in chronological order")
	fmt.Println("  dw logs --format csv --limit 100                 # Export 100 logs as CSV")
	fmt.Println("  dw logs --format markdown --limit 50             # Export 50 logs as Markdown for LLM analysis")
	fmt.Println("  dw logs --query \"SELECT * FROM events\"           # Run custom SQL query")
	fmt.Println()
}

func printLogsHelp() {
	fmt.Println("DarwinFlow Logs - Database Schema")
	fmt.Println()
	fmt.Println("DATABASE STRUCTURE:")
	fmt.Println()
	fmt.Println("Table: events")
	fmt.Println("  Columns:")
	fmt.Println("    - id           TEXT PRIMARY KEY    Unique event identifier (UUID)")
	fmt.Println("    - timestamp    INTEGER NOT NULL    Unix timestamp in milliseconds")
	fmt.Println("    - event_type   TEXT NOT NULL       Event type (e.g., 'tool.invoked', 'chat.message.user')")
	fmt.Println("    - session_id   TEXT                Claude Code session identifier")
	fmt.Println("    - payload      TEXT NOT NULL       JSON payload with event-specific data")
	fmt.Println("    - content      TEXT NOT NULL       Normalized searchable content")
	fmt.Println()
	fmt.Println("  Indexes:")
	fmt.Println("    - idx_events_timestamp          ON events(timestamp)")
	fmt.Println("    - idx_events_type               ON events(event_type)")
	fmt.Println("    - idx_events_timestamp_type     ON events(timestamp, event_type)")
	fmt.Println("    - idx_events_session_id         ON events(session_id)")
	fmt.Println("    - idx_events_timestamp_session  ON events(timestamp, session_id)")
	fmt.Println()
	fmt.Println("FTS5 Virtual Table: events_fts (if available)")
	fmt.Println("  Full-text search on content field")
	fmt.Println()
	fmt.Println("COMMON EVENT TYPES:")
	fmt.Println("  - tool.invoked              Tool was invoked (Read, Write, Bash, etc.)")
	fmt.Println("  - tool.result               Tool execution completed")
	fmt.Println("  - chat.message.user         User sent a message")
	fmt.Println("  - chat.message.assistant    Assistant sent a message")
	fmt.Println("  - chat.started              Chat session started")
	fmt.Println("  - file.read                 File was read")
	fmt.Println("  - file.written              File was written")
	fmt.Println("  - context.changed           Context changed")
	fmt.Println("  - error                     Error occurred")
	fmt.Println()
	fmt.Println("QUERY EXAMPLES:")
	fmt.Println("  # Count events by type")
	fmt.Println("  dw logs --query \"SELECT event_type, COUNT(*) as count FROM events GROUP BY event_type ORDER BY count DESC\"")
	fmt.Println()
	fmt.Println("  # Find all tool invocations in last hour")
	fmt.Printf("  dw logs --query \"SELECT * FROM events WHERE event_type = 'tool.invoked' AND timestamp > strftime('%%s', 'now', '-1 hour') * 1000\"\n")
	fmt.Println()
	fmt.Println("  # Search content for specific text")
	fmt.Printf("  dw logs --query \"SELECT * FROM events WHERE content LIKE '%%sqlite%%' LIMIT 10\"\n")
	fmt.Println()
	fmt.Println("Database location:", app.DefaultDBPath)
	fmt.Println()
}

func listLogs(ctx context.Context, service *app.LogsService, opts *logsOptions) error {
	records, err := service.ListRecentLogs(ctx, opts.limit, opts.sessionID, opts.ordered)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("No logs found.")
		fmt.Println("Run 'dw claude init' to initialize logging, then use Claude Code to generate events.")
		return nil
	}

	// Handle CSV format
	if opts.format == "csv" {
		return app.FormatLogsAsCSV(os.Stdout, records)
	}

	// Handle Markdown format
	if opts.format == "markdown" {
		return app.FormatLogsAsMarkdown(os.Stdout, records)
	}

	// Validate format
	if opts.format != "text" && opts.format != "" {
		fmt.Fprintf(os.Stderr, "Error: Invalid format '%s'. Valid formats: text, csv, markdown\n", opts.format)
		os.Exit(1)
	}

	// Display logs in text format
	if opts.sessionID != "" {
		fmt.Printf("Showing %d logs for session %s:\n\n", len(records), opts.sessionID)
	} else {
		fmt.Printf("Showing %d most recent logs:\n\n", len(records))
	}

	for i, record := range records {
		fmt.Print(app.FormatLogRecord(i, record))
	}

	return nil
}

func executeRawQuery(ctx context.Context, service *app.LogsService, query string) error {
	result, err := service.ExecuteRawQuery(ctx, query)
	if err != nil {
		return err
	}

	// Print column headers
	for i, col := range result.Columns {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Print(col)
	}
	fmt.Println()
	fmt.Println(repeatString("-", 80))

	// Print rows
	for _, row := range result.Rows {
		for i, val := range row {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Print(app.FormatQueryValue(val))
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Printf("(%d rows)\n", len(result.Rows))
	return nil
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

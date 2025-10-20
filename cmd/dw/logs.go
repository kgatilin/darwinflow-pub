package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

// LogsOptions contains options for the logs command
type LogsOptions struct {
	Limit        int
	SessionLimit int
	Query        string
	SessionID    string
	Ordered      bool
	Format       string
	Help         bool
}

// ParseLogsFlags parses command line flags for the logs command
func ParseLogsFlags(args []string) (*LogsOptions, error) {
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	opts := &LogsOptions{}

	fs.IntVar(&opts.Limit, "limit", 20, "Number of most recent logs to display")
	fs.IntVar(&opts.SessionLimit, "session-limit", 0, "Limit by number of sessions instead of logs (0 = use --limit)")
	fs.StringVar(&opts.Query, "query", "", "Arbitrary SQL query to execute")
	fs.StringVar(&opts.SessionID, "session-id", "", "Filter logs by session ID")
	fs.BoolVar(&opts.Ordered, "ordered", false, "Order by timestamp ASC and session ID (chronological)")
	fs.StringVar(&opts.Format, "format", "text", "Output format: text, csv, or markdown")
	fs.BoolVar(&opts.Help, "help", false, "Show help and database schema")

	fs.Usage = printLogsUsage

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opts, nil
}

func handleLogs(args []string) {
	opts, err := ParseLogsFlags(args)
	if err != nil {
		os.Exit(1)
	}

	// Show help if requested
	if opts.Help {
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
	if opts.Query != "" {
		if err := ExecuteRawQuery(ctx, service, opts.Query); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle standard log listing
	if err := ListLogs(ctx, service, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printLogsUsage() {
	fmt.Println("Usage: dw logs [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --limit N            Number of most recent logs to display (default: 20)")
	fmt.Println("  --session-limit N    Limit by number of sessions instead of logs (0 = use --limit)")
	fmt.Println("  --session-id ID      Filter logs by session ID")
	fmt.Println("  --ordered            Order by timestamp ASC and session ID (chronological)")
	fmt.Println("  --format FORMAT      Output format: text, csv, or markdown (default: text)")
	fmt.Println("  --query SQL          Execute an arbitrary SQL query")
	fmt.Println("  --help               Show help and database schema")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dw logs                                          # Show 20 most recent logs")
	fmt.Println("  dw logs --limit 50                               # Show 50 most recent logs")
	fmt.Println("  dw logs --session-limit 3                        # Show all logs from 3 most recent sessions")
	fmt.Println("  dw logs --session-id abc123                      # Show logs for session abc123")
	fmt.Println("  dw logs --session-id abc123 --ordered            # Show session abc123 in chronological order")
	fmt.Println("  dw logs --format csv --limit 100                 # Export 100 logs as CSV")
	fmt.Println("  dw logs --format markdown --session-limit 5      # Export 5 most recent sessions as Markdown")
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

// ListLogs displays logs based on the provided options
func ListLogs(ctx context.Context, service *app.LogsService, opts *LogsOptions) error {
	records, err := service.ListRecentLogs(ctx, opts.Limit, opts.SessionLimit, opts.SessionID, opts.Ordered)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("No logs found.")
		fmt.Println("Run 'dw claude init' to initialize logging, then use Claude Code to generate events.")
		return nil
	}

	// Handle CSV format
	if opts.Format == "csv" {
		return app.FormatLogsAsCSV(os.Stdout, records)
	}

	// Handle Markdown format
	if opts.Format == "markdown" {
		return app.FormatLogsAsMarkdown(os.Stdout, records)
	}

	// Validate format
	if opts.Format != "text" && opts.Format != "" {
		fmt.Fprintf(os.Stderr, "Error: Invalid format '%s'. Valid formats: text, csv, markdown\n", opts.Format)
		os.Exit(1)
	}

	// Display logs in text format
	if opts.SessionID != "" {
		fmt.Printf("Showing %d logs for session %s:\n\n", len(records), opts.SessionID)
	} else if opts.SessionLimit > 0 {
		fmt.Printf("Showing %d logs from %d most recent sessions:\n\n", len(records), opts.SessionLimit)
	} else {
		fmt.Printf("Showing %d most recent logs:\n\n", len(records))
	}

	for i, record := range records {
		fmt.Print(app.FormatLogRecord(i, record))
	}

	return nil
}

// ExecuteRawQuery executes a raw SQL query and displays the results
func ExecuteRawQuery(ctx context.Context, service *app.LogsService, query string) error {
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

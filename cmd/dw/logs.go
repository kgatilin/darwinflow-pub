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
		PrintLogsHelp()
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

	// Initialize database schema (including migration from old databases)
	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	service := app.NewLogsService(repo, repo)
	handler := app.NewLogsCommandHandler(service, os.Stdout)

	// Handle arbitrary SQL query
	if opts.Query != "" {
		if err := handler.ExecuteRawQuery(ctx, opts.Query); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle standard log listing
	if err := handler.ListLogs(ctx, opts.Limit, opts.SessionLimit, opts.SessionID, opts.Ordered, opts.Format); err != nil {
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

// PrintLogsHelp displays detailed help for the logs command
func PrintLogsHelp() {
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


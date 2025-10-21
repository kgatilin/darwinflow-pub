package pluginsdk

import (
	"context"
	"time"
)

// EventRepository defines the interface for persisting and retrieving events.
// The framework provides implementations of this interface to allow plugins
// to access the event store for queries and analysis.
type EventRepository interface {
	// Initialize initializes the repository (creates schema, indexes, etc.)
	Initialize(ctx context.Context) error

	// Save persists an event to the repository
	Save(ctx context.Context, event *Event) error

	// Query retrieves events based on query criteria
	Query(ctx context.Context, query EventQuery) ([]*Event, error)

	// Close closes the repository connection
	Close() error
}

// EventQuery defines query parameters for retrieving events.
// Plugins use this to filter and search events from the repository.
type EventQuery struct {
	// Time range for filtering events
	StartTime *time.Time
	EndTime   *time.Time

	// EventTypes filters by event type strings (e.g., "claude.tool.invoked")
	EventTypes []string

	// Metadata filters events by metadata key-value pairs (e.g., session_id)
	Metadata map[string]string

	// SearchText enables full-text search on event content
	SearchText string

	// OrderByTime if true, orders results by timestamp ASC then session_id
	// otherwise returns in descending timestamp order (most recent first)
	OrderByTime bool

	// Pagination parameters
	Limit  int // 0 means no limit
	Offset int // Number of results to skip
}

// QueryResult represents the result of a raw SQL query execution.
// Used by RawQueryExecutor for admin/debug queries.
type QueryResult struct {
	Columns []string        // Column names from the query
	Rows    [][]interface{} // Query result rows
}

// RawQueryExecutor defines the interface for executing arbitrary SQL queries.
// This is separate from EventRepository to keep normal queries structured,
// while allowing debug/admin capabilities through raw SQL access.
type RawQueryExecutor interface {
	// ExecuteRawQuery executes an arbitrary SQL query and returns results
	ExecuteRawQuery(ctx context.Context, query string) (*QueryResult, error)
}

package domain

import (
	"context"
	"time"
)

// EventRepository defines the interface for persisting and retrieving events (repository pattern)
type EventRepository interface {
	// Initialize initializes the repository (creates schema, indexes, etc.)
	Initialize(ctx context.Context) error

	// Save persists an event
	Save(ctx context.Context, event *Event) error

	// FindByQuery retrieves events based on query criteria
	FindByQuery(ctx context.Context, query EventQuery) ([]*Event, error)

	// Close closes the repository connection
	Close() error
}

// EventQuery defines query parameters for retrieving events (specification pattern)
type EventQuery struct {
	// Time range
	StartTime *time.Time
	EndTime   *time.Time

	// Event type filtering
	EventTypes []EventType

	// Session filtering
	SessionID string

	// Context filtering
	Context string

	// Full-text search
	SearchText string

	// Ordering
	OrderByTime bool // If true, order by timestamp ASC, session_id

	// Pagination
	Limit  int
	Offset int
}

// QueryResult represents the result of a raw query execution
type QueryResult struct {
	Columns []string
	Rows    [][]interface{}
}

// RawQueryExecutor defines the interface for executing arbitrary SQL queries
// This is separate from EventRepository to keep domain queries pure while
// allowing debug/admin capabilities
type RawQueryExecutor interface {
	// ExecuteRawQuery executes an arbitrary SQL query and returns results
	ExecuteRawQuery(ctx context.Context, query string) (*QueryResult, error)
}

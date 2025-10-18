package infra

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// SQLiteEventRepository implements domain.EventRepository using SQLite
type SQLiteEventRepository struct {
	db   *sql.DB
	path string
}

// NewSQLiteEventRepository creates a new SQLite-backed event repository
func NewSQLiteEventRepository(dbPath string) (*SQLiteEventRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	return &SQLiteEventRepository{
		db:   db,
		path: dbPath,
	}, nil
}

// Initialize initializes the database schema
func (r *SQLiteEventRepository) Initialize(ctx context.Context) error {
	// Create base schema
	baseSchema := `
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			timestamp INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			session_id TEXT,
			payload TEXT NOT NULL,
			content TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
		CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
		CREATE INDEX IF NOT EXISTS idx_events_timestamp_type ON events(timestamp, event_type);
		CREATE INDEX IF NOT EXISTS idx_events_session_id ON events(session_id);
		CREATE INDEX IF NOT EXISTS idx_events_timestamp_session ON events(timestamp, session_id);
	`

	_, err := r.db.ExecContext(ctx, baseSchema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Try to create FTS5 virtual table (optional, may not be available)
	ftsSchema := `
		CREATE VIRTUAL TABLE IF NOT EXISTS events_fts USING fts5(
			content,
			content=events,
			content_rowid=rowid
		);

		CREATE TRIGGER IF NOT EXISTS events_fts_insert AFTER INSERT ON events BEGIN
			INSERT INTO events_fts(rowid, content) VALUES (new.rowid, new.content);
		END;

		CREATE TRIGGER IF NOT EXISTS events_fts_delete AFTER DELETE ON events BEGIN
			DELETE FROM events_fts WHERE rowid = old.rowid;
		END;

		CREATE TRIGGER IF NOT EXISTS events_fts_update AFTER UPDATE ON events BEGIN
			DELETE FROM events_fts WHERE rowid = old.rowid;
			INSERT INTO events_fts(rowid, content) VALUES (new.rowid, new.content);
		END;
	`

	// Attempt FTS5, but don't fail if unavailable
	_, _ = r.db.ExecContext(ctx, ftsSchema)

	return nil
}

// Save persists an event
func (r *SQLiteEventRepository) Save(ctx context.Context, event *domain.Event) error {
	payloadJSON, err := event.MarshalPayload()
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO events (id, timestamp, event_type, session_id, payload, content)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.Timestamp.UnixMilli(),
		string(event.Type),
		event.SessionID,
		string(payloadJSON),
		event.Content,
	)

	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

// FindByQuery retrieves events based on query criteria
func (r *SQLiteEventRepository) FindByQuery(ctx context.Context, query domain.EventQuery) ([]*domain.Event, error) {
	var conditions []string
	var args []interface{}

	// Build WHERE clause
	if query.StartTime != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, query.StartTime.UnixMilli())
	}

	if query.EndTime != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, query.EndTime.UnixMilli())
	}

	if len(query.EventTypes) > 0 {
		placeholders := make([]string, len(query.EventTypes))
		for i, et := range query.EventTypes {
			placeholders[i] = "?"
			args = append(args, string(et))
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	if query.SessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, query.SessionID)
	}

	// Build SQL query
	sqlQuery := "SELECT id, timestamp, event_type, session_id, payload, content FROM events"

	if query.SearchText != "" {
		// Try FTS search first, fall back to LIKE if FTS not available
		ftsQuery := `
			SELECT e.id, e.timestamp, e.event_type, e.session_id, e.payload, e.content
			FROM events e
			JOIN events_fts fts ON fts.rowid = e.rowid
			WHERE fts.content MATCH ?
		`
		ftsArgs := append([]interface{}{query.SearchText}, args...)

		if len(conditions) > 0 {
			ftsQuery += " AND " + strings.Join(conditions, " AND ")
		}

		// Try FTS query
		_, err := r.db.QueryContext(ctx, ftsQuery+" LIMIT 1", ftsArgs...)
		if err == nil {
			// FTS is available
			sqlQuery = ftsQuery
			args = ftsArgs
		} else {
			// Fall back to LIKE search
			conditions = append([]string{"content LIKE ?"}, conditions...)
			args = append([]interface{}{"%" + query.SearchText + "%"}, args...)
			if len(conditions) > 0 {
				sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
			}
		}
	} else if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY clause
	if query.OrderByTime {
		sqlQuery += " ORDER BY timestamp ASC, session_id"
	} else {
		sqlQuery += " ORDER BY timestamp DESC"
	}

	if query.Limit > 0 {
		sqlQuery += " LIMIT ?"
		args = append(args, query.Limit)

		if query.Offset > 0 {
			sqlQuery += " OFFSET ?"
			args = append(args, query.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var id, eventType, payloadStr, content string
		var sessionID sql.NullString
		var timestampMs int64

		if err := rows.Scan(&id, &timestampMs, &eventType, &sessionID, &payloadStr, &content); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Reconstruct domain event
		// Note: We unmarshal into json.RawMessage to preserve the original payload structure
		var payload json.RawMessage
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		event := &domain.Event{
			ID:        id,
			Timestamp: millisecondsToTime(timestampMs),
			Type:      domain.EventType(eventType),
			SessionID: sessionID.String,
			Payload:   payload,
			Content:   content,
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// Close closes the database connection
func (r *SQLiteEventRepository) Close() error {
	return r.db.Close()
}

// ExecuteRawQuery executes an arbitrary SQL query and returns results
// Implements domain.RawQueryExecutor interface
func (r *SQLiteEventRepository) ExecuteRawQuery(ctx context.Context, query string) (*domain.QueryResult, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Scan all rows
	var resultRows [][]interface{}
	for rows.Next() {
		// Prepare scan destinations
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Copy values to result
		row := make([]interface{}, len(columns))
		copy(row, values)
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &domain.QueryResult{
		Columns: columns,
		Rows:    resultRows,
	}, nil
}

// Helper function to convert milliseconds to time.Time
func millisecondsToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

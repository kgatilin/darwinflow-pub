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
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
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
	// Step 1: Create base tables (minimal schema for old versions)
	baseTablesSchema := `
		CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			timestamp INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			session_id TEXT,
			payload TEXT NOT NULL,
			content TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS session_analyses (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			analyzed_at INTEGER NOT NULL,
			analysis_result TEXT NOT NULL,
			model_used TEXT,
			prompt_used TEXT,
			patterns_summary TEXT
		);
	`

	_, err := r.db.ExecContext(ctx, baseTablesSchema)
	if err != nil {
		return fmt.Errorf("failed to create base tables: %w", err)
	}

	// Step 2: Run migrations to add new columns if they don't exist
	// These will fail silently if columns already exist
	migrationSQL := `ALTER TABLE session_analyses ADD COLUMN analysis_type TEXT DEFAULT 'tool_analysis';`
	_, _ = r.db.ExecContext(ctx, migrationSQL)

	migrationSQL2 := `ALTER TABLE session_analyses ADD COLUMN prompt_name TEXT DEFAULT 'analysis';`
	_, _ = r.db.ExecContext(ctx, migrationSQL2)

	// Add version column to events table for schema evolution support
	migrationSQL3 := `ALTER TABLE events ADD COLUMN version TEXT DEFAULT '1.0';`
	_, _ = r.db.ExecContext(ctx, migrationSQL3)

	// Step 3: Clean up duplicate analyses (keep only the most recent one per session_id/analysis_type)
	// This handles the case where old databases have multiple analyses with the same analysis_type
	cleanupSQL := `
		DELETE FROM session_analyses
		WHERE id NOT IN (
			SELECT id FROM (
				SELECT id, session_id, analysis_type,
				       ROW_NUMBER() OVER (PARTITION BY session_id, analysis_type ORDER BY analyzed_at DESC) as rn
				FROM session_analyses
			)
			WHERE rn = 1
		);
	`
	// Ignore errors - older SQLite versions might not support window functions
	_, _ = r.db.ExecContext(ctx, cleanupSQL)

	// Fallback cleanup for older SQLite without window functions
	fallbackCleanup := `
		DELETE FROM session_analyses
		WHERE id IN (
			SELECT a1.id
			FROM session_analyses a1
			INNER JOIN session_analyses a2
			ON a1.session_id = a2.session_id
			   AND a1.analysis_type = a2.analysis_type
			   AND a1.analyzed_at < a2.analyzed_at
		);
	`
	_, _ = r.db.ExecContext(ctx, fallbackCleanup)

	// Step 4: Create indexes (including those on new columns)
	indexSchema := `
		CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
		CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
		CREATE INDEX IF NOT EXISTS idx_events_timestamp_type ON events(timestamp, event_type);
		CREATE INDEX IF NOT EXISTS idx_events_session_id ON events(session_id);
		CREATE INDEX IF NOT EXISTS idx_events_timestamp_session ON events(timestamp, session_id);

		CREATE INDEX IF NOT EXISTS idx_analyses_session_id ON session_analyses(session_id);
		CREATE INDEX IF NOT EXISTS idx_analyses_analyzed_at ON session_analyses(analyzed_at);
	`

	_, err = r.db.ExecContext(ctx, indexSchema)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Step 5: Create unique index (after cleanup)
	uniqueIndexSQL := `CREATE UNIQUE INDEX IF NOT EXISTS idx_analyses_session_type ON session_analyses(session_id, analysis_type);`
	_, err = r.db.ExecContext(ctx, uniqueIndexSQL)
	if err != nil {
		return fmt.Errorf("failed to create unique index: %w", err)
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

	// Step 6: Create bus_events table for event bus persistence
	busEventsSchema := `
		CREATE TABLE IF NOT EXISTS bus_events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			labels TEXT,
			metadata TEXT,
			payload BLOB
		);

		CREATE INDEX IF NOT EXISTS idx_bus_events_type ON bus_events(type);
		CREATE INDEX IF NOT EXISTS idx_bus_events_source ON bus_events(source);
		CREATE INDEX IF NOT EXISTS idx_bus_events_timestamp ON bus_events(timestamp);
	`

	_, err = r.db.ExecContext(ctx, busEventsSchema)
	if err != nil {
		return fmt.Errorf("failed to create bus_events table: %w", err)
	}

	return nil
}

// Save persists an event
func (r *SQLiteEventRepository) Save(ctx context.Context, event *domain.Event) error {
	payloadJSON, err := event.MarshalPayload()
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO events (id, timestamp, event_type, session_id, payload, content, version)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.Timestamp.UnixMilli(),
		string(event.Type),
		event.SessionID,
		string(payloadJSON),
		event.Content,
		event.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

// FindByQuery retrieves events based on query criteria
// Uses pluginsdk.EventQuery as the single source of truth for query structure
func (r *SQLiteEventRepository) FindByQuery(ctx context.Context, query pluginsdk.EventQuery) ([]*domain.Event, error) {
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

	// Map Metadata["session_id"] to session_id column
	if sessionID, ok := query.Metadata["session_id"]; ok && sessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, sessionID)
	}

	// Build SQL query
	sqlQuery := "SELECT id, timestamp, event_type, session_id, payload, content, COALESCE(version, '1.0') as version FROM events"

	if query.SearchText != "" {
		// Try FTS search first, fall back to LIKE if FTS not available
		ftsQuery := `
			SELECT e.id, e.timestamp, e.event_type, e.session_id, e.payload, e.content, COALESCE(e.version, '1.0') as version
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
		var id, eventType, payloadStr, content, version string
		var sessionID sql.NullString
		var timestampMs int64

		if err := rows.Scan(&id, &timestampMs, &eventType, &sessionID, &payloadStr, &content, &version); err != nil {
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
			Type:      eventType,
			SessionID: sessionID.String,
			Payload:   payload,
			Content:   content,
			Version:   version,
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
// Implements pluginsdk.RawQueryExecutor interface
func (r *SQLiteEventRepository) ExecuteRawQuery(ctx context.Context, query string) (*pluginsdk.QueryResult, error) {
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

	return &pluginsdk.QueryResult{
		Columns: columns,
		Rows:    resultRows,
	}, nil
}

// SaveAnalysis persists a session analysis
func (r *SQLiteEventRepository) SaveAnalysis(ctx context.Context, analysis *domain.SessionAnalysis) error {
	query := `
		INSERT INTO session_analyses (id, session_id, analyzed_at, analysis_result, model_used, prompt_used, patterns_summary, analysis_type, prompt_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id, analysis_type) DO UPDATE SET
			analyzed_at = excluded.analyzed_at,
			analysis_result = excluded.analysis_result,
			model_used = excluded.model_used,
			prompt_used = excluded.prompt_used,
			patterns_summary = excluded.patterns_summary,
			prompt_name = excluded.prompt_name
	`

	_, err := r.db.ExecContext(ctx, query,
		analysis.ID,
		analysis.SessionID,
		analysis.AnalyzedAt.UnixMilli(),
		analysis.AnalysisResult,
		analysis.ModelUsed,
		analysis.PromptUsed,
		analysis.PatternsSummary,
		analysis.AnalysisType,
		analysis.PromptName,
	)

	if err != nil {
		return fmt.Errorf("failed to store analysis: %w", err)
	}

	return nil
}

// GetAnalysisBySessionID retrieves the most recent analysis for a session
func (r *SQLiteEventRepository) GetAnalysisBySessionID(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error) {
	query := `
		SELECT id, session_id, analyzed_at, analysis_result, model_used, prompt_used, patterns_summary,
		       COALESCE(analysis_type, 'tool_analysis') as analysis_type,
		       COALESCE(prompt_name, 'analysis') as prompt_name
		FROM session_analyses
		WHERE session_id = ?
		ORDER BY analyzed_at DESC
		LIMIT 1
	`

	var analysis domain.SessionAnalysis
	var analyzedAtMs int64
	var modelUsed, promptUsed, patternsSummary sql.NullString

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&analysis.ID,
		&analysis.SessionID,
		&analyzedAtMs,
		&analysis.AnalysisResult,
		&modelUsed,
		&promptUsed,
		&patternsSummary,
		&analysis.AnalysisType,
		&analysis.PromptName,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	analysis.AnalyzedAt = millisecondsToTime(analyzedAtMs)
	analysis.ModelUsed = modelUsed.String
	analysis.PromptUsed = promptUsed.String
	analysis.PatternsSummary = patternsSummary.String

	return &analysis, nil
}

// GetAnalysesBySessionID retrieves all analyses for a session, ordered by analyzed_at DESC
func (r *SQLiteEventRepository) GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionAnalysis, error) {
	query := `
		SELECT id, session_id, analyzed_at, analysis_result, model_used, prompt_used, patterns_summary,
		       COALESCE(analysis_type, 'tool_analysis') as analysis_type,
		       COALESCE(prompt_name, 'analysis') as prompt_name
		FROM session_analyses
		WHERE session_id = ?
		ORDER BY analyzed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query analyses: %w", err)
	}
	defer rows.Close()

	var analyses []*domain.SessionAnalysis
	for rows.Next() {
		var analysis domain.SessionAnalysis
		var analyzedAtMs int64
		var modelUsed, promptUsed, patternsSummary sql.NullString

		err := rows.Scan(
			&analysis.ID,
			&analysis.SessionID,
			&analyzedAtMs,
			&analysis.AnalysisResult,
			&modelUsed,
			&promptUsed,
			&patternsSummary,
			&analysis.AnalysisType,
			&analysis.PromptName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis: %w", err)
		}

		analysis.AnalyzedAt = millisecondsToTime(analyzedAtMs)
		analysis.ModelUsed = modelUsed.String
		analysis.PromptUsed = promptUsed.String
		analysis.PatternsSummary = patternsSummary.String

		analyses = append(analyses, &analysis)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return analyses, nil
}

// GetUnanalyzedSessionIDs retrieves session IDs that have not been analyzed
func (r *SQLiteEventRepository) GetUnanalyzedSessionIDs(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT session_id
		FROM events
		WHERE session_id IS NOT NULL
		  AND session_id != ''
		  AND session_id NOT IN (SELECT DISTINCT session_id FROM session_analyses)
		ORDER BY session_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unanalyzed sessions: %w", err)
	}
	defer rows.Close()

	var sessionIDs []string
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, fmt.Errorf("failed to scan session ID: %w", err)
		}
		sessionIDs = append(sessionIDs, sessionID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return sessionIDs, nil
}

// GetAllAnalyses retrieves all analyses, ordered by analyzed_at DESC
func (r *SQLiteEventRepository) GetAllAnalyses(ctx context.Context, limit int) ([]*domain.SessionAnalysis, error) {
	query := `
		SELECT id, session_id, analyzed_at, analysis_result, model_used, prompt_used, patterns_summary,
		       COALESCE(analysis_type, 'tool_analysis') as analysis_type,
		       COALESCE(prompt_name, 'analysis') as prompt_name
		FROM session_analyses
		ORDER BY analyzed_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get analyses: %w", err)
	}
	defer rows.Close()

	var analyses []*domain.SessionAnalysis
	for rows.Next() {
		var analysis domain.SessionAnalysis
		var analyzedAtMs int64
		var modelUsed, promptUsed, patternsSummary sql.NullString

		err := rows.Scan(
			&analysis.ID,
			&analysis.SessionID,
			&analyzedAtMs,
			&analysis.AnalysisResult,
			&modelUsed,
			&promptUsed,
			&patternsSummary,
			&analysis.AnalysisType,
			&analysis.PromptName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis: %w", err)
		}

		analysis.AnalyzedAt = millisecondsToTime(analyzedAtMs)
		analysis.ModelUsed = modelUsed.String
		analysis.PromptUsed = promptUsed.String
		analysis.PatternsSummary = patternsSummary.String

		analyses = append(analyses, &analysis)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return analyses, nil
}

// GetAllSessionIDs retrieves all session IDs, ordered by most recent first
func (r *SQLiteEventRepository) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	query := `
		SELECT session_id
		FROM events
		WHERE session_id IS NOT NULL AND session_id != ''
		GROUP BY session_id
		ORDER BY MAX(timestamp) DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get session IDs: %w", err)
	}
	defer rows.Close()

	var sessionIDs []string
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return nil, fmt.Errorf("failed to scan session ID: %w", err)
		}
		sessionIDs = append(sessionIDs, sessionID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return sessionIDs, nil
}

// Helper function to convert milliseconds to time.Time
func millisecondsToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

package infra

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// SQLiteEventBusRepository implements domain.EventBusRepository using SQLite.
// It stores bus events in a dedicated table separate from the main events table.
type SQLiteEventBusRepository struct {
	db *sql.DB
}

// NewSQLiteEventBusRepository creates a new SQLite event bus repository.
// It reuses an existing database connection.
func NewSQLiteEventBusRepository(db *sql.DB) (*SQLiteEventBusRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	return &SQLiteEventBusRepository{
		db: db,
	}, nil
}

// NewSQLiteEventBusRepositoryFromRepo creates a new SQLite event bus repository
// from an existing SQLiteEventRepository, sharing the same database connection.
func NewSQLiteEventBusRepositoryFromRepo(repo *SQLiteEventRepository) *SQLiteEventBusRepository {
	return &SQLiteEventBusRepository{
		db: repo.db,
	}
}

// Initialize creates the bus_events table and indexes if they don't exist.
func (r *SQLiteEventBusRepository) Initialize(ctx context.Context) error {
	schema := `
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

	_, err := r.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create bus_events schema: %w", err)
	}

	return nil
}

// StoreEvent persists a bus event to SQLite.
func (r *SQLiteEventBusRepository) StoreEvent(ctx context.Context, event pluginsdk.BusEvent) error {
	// Serialize labels and metadata to JSON
	labelsJSON, err := json.Marshal(event.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO bus_events (id, type, source, timestamp, labels, metadata, payload)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		event.Source,
		event.Timestamp.UnixMilli(),
		string(labelsJSON),
		string(metadataJSON),
		event.Payload,
	)

	if err != nil {
		return fmt.Errorf("failed to store bus event: %w", err)
	}

	return nil
}

// GetEvents retrieves events matching the filter criteria.
func (r *SQLiteEventBusRepository) GetEvents(ctx context.Context, filter pluginsdk.EventFilter, limit int) ([]pluginsdk.BusEvent, error) {
	return r.queryEvents(ctx, filter, limit, nil)
}

// GetEventsSince retrieves events since a given timestamp for replay.
func (r *SQLiteEventBusRepository) GetEventsSince(ctx context.Context, since interface{}, filter pluginsdk.EventFilter, limit int) ([]pluginsdk.BusEvent, error) {
	// Convert since to time.Time
	var sinceTime time.Time
	switch v := since.(type) {
	case time.Time:
		sinceTime = v
	case *time.Time:
		if v != nil {
			sinceTime = *v
		}
	case int64:
		sinceTime = time.Unix(0, v*int64(time.Millisecond))
	default:
		return nil, fmt.Errorf("unsupported since type: %T", since)
	}

	return r.queryEvents(ctx, filter, limit, &sinceTime)
}

// queryEvents is the internal implementation for event queries.
func (r *SQLiteEventBusRepository) queryEvents(ctx context.Context, filter pluginsdk.EventFilter, limit int, since *time.Time) ([]pluginsdk.BusEvent, error) {
	var conditions []string
	var args []interface{}

	// Add timestamp filter if provided
	if since != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, since.UnixMilli())
	}

	// Add type pattern filter
	if filter.TypePattern != "" {
		// Check if it's a simple pattern or requires glob matching
		if strings.Contains(filter.TypePattern, "*") {
			// Use LIKE with % instead of *
			likePattern := strings.ReplaceAll(filter.TypePattern, "*", "%")
			conditions = append(conditions, "type LIKE ?")
			args = append(args, likePattern)
		} else {
			// Exact match
			conditions = append(conditions, "type = ?")
			args = append(args, filter.TypePattern)
		}
	}

	// Add source plugin filter
	if filter.SourcePlugin != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, filter.SourcePlugin)
	}

	// Build SQL query
	sqlQuery := "SELECT id, type, source, timestamp, labels, metadata, payload FROM bus_events"

	if len(conditions) > 0 {
		sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Order by timestamp ascending for replay
	sqlQuery += " ORDER BY timestamp ASC"

	if limit > 0 {
		sqlQuery += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query bus events: %w", err)
	}
	defer rows.Close()

	var events []pluginsdk.BusEvent
	for rows.Next() {
		var id, eventType, source, labelsStr, metadataStr string
		var timestampMs int64
		var payload []byte

		if err := rows.Scan(&id, &eventType, &source, &timestampMs, &labelsStr, &metadataStr, &payload); err != nil {
			return nil, fmt.Errorf("failed to scan bus event: %w", err)
		}

		// Deserialize labels
		var labels map[string]string
		if err := json.Unmarshal([]byte(labelsStr), &labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}

		// Deserialize metadata
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		event := pluginsdk.BusEvent{
			ID:        id,
			Type:      eventType,
			Source:    source,
			Timestamp: time.Unix(0, timestampMs*int64(time.Millisecond)),
			Labels:    labels,
			Metadata:  metadata,
			Payload:   payload,
		}

		// Apply label filters (post-filter since SQLite doesn't support JSON queries easily)
		if r.matchesLabelFilter(event, filter) {
			events = append(events, event)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bus events: %w", err)
	}

	return events, nil
}

// matchesLabelFilter checks if an event matches the label filter criteria.
// This is done post-query since SQLite JSON queries can be complex.
func (r *SQLiteEventBusRepository) matchesLabelFilter(event pluginsdk.BusEvent, filter pluginsdk.EventFilter) bool {
	if len(filter.Labels) == 0 {
		return true
	}

	// All filter labels must match event labels
	for key, value := range filter.Labels {
		eventValue, exists := event.Labels[key]
		if !exists || eventValue != value {
			return false
		}
	}

	return true
}

// Verify SQLiteEventBusRepository implements domain.EventBusRepository
var _ domain.EventBusRepository = (*SQLiteEventBusRepository)(nil)

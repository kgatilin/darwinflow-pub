package domain

import (
	"context"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// EventRepository defines the interface for persisting and retrieving events (repository pattern)
// This interface works with domain.Event (storage format) but uses pluginsdk.EventQuery (query format).
// The SDK defines the query structure as the single source of truth.
type EventRepository interface {
	// Initialize initializes the repository (creates schema, indexes, etc.)
	Initialize(ctx context.Context) error

	// Save persists an event
	Save(ctx context.Context, event *Event) error

	// FindByQuery retrieves events based on query criteria
	// Uses pluginsdk.EventQuery for query structure (single source of truth)
	// Returns domain.Event (storage format)
	FindByQuery(ctx context.Context, query pluginsdk.EventQuery) ([]*Event, error)

	// Close closes the repository connection
	Close() error
}

// Note: EventQuery, QueryResult, and RawQueryExecutor are now defined in pkg/pluginsdk
// to serve as the single source of truth. Import from pluginsdk to use them.

// EventBusRepository defines the interface for persisting and retrieving event bus events.
// This provides optional persistence for the in-memory event bus.
type EventBusRepository interface {
	// StoreEvent persists a bus event to storage
	StoreEvent(ctx context.Context, event pluginsdk.BusEvent) error

	// GetEvents retrieves events matching the filter criteria
	GetEvents(ctx context.Context, filter pluginsdk.EventFilter, limit int) ([]pluginsdk.BusEvent, error)

	// GetEventsSince retrieves events since a given timestamp for replay
	GetEventsSince(ctx context.Context, since interface{}, filter pluginsdk.EventFilter, limit int) ([]pluginsdk.BusEvent, error)
}

// AnalysisRepository defines the interface for persisting and retrieving analyses.
// Supports both generic Analysis (new, plugin-agnostic) and SessionAnalysis (backward compatibility).
type AnalysisRepository interface {
	// Generic analysis methods (plugin-agnostic)
	SaveGenericAnalysis(ctx context.Context, analysis *Analysis) error
	FindAnalysisByViewID(ctx context.Context, viewID string) ([]*Analysis, error)
	FindAnalysisByViewType(ctx context.Context, viewType string) ([]*Analysis, error)
	FindAnalysisById(ctx context.Context, id string) (*Analysis, error)
	ListRecentAnalyses(ctx context.Context, limit int) ([]*Analysis, error)

	// Session-specific methods (backward compatibility)
	// NOTE: These exist for backward compatibility with internal code.
	// Analysis storage is semantically owned by the claude-code plugin.
	SaveAnalysis(ctx context.Context, analysis *SessionAnalysis) error
	GetAnalysisBySessionID(ctx context.Context, sessionID string) (*SessionAnalysis, error)
	GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*SessionAnalysis, error)
	GetUnanalyzedSessionIDs(ctx context.Context) ([]string, error)
	GetAllAnalyses(ctx context.Context, limit int) ([]*SessionAnalysis, error)
	GetAllSessionIDs(ctx context.Context, limit int) ([]string, error)
}

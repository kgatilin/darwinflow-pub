package pluginsdk

// AnalysisView represents a view of events that can be analyzed.
// Plugins implement this interface to provide different views of their events
// (e.g., sessions, tasks, date ranges, cross-plugin views).
type AnalysisView interface {
	// GetID returns a unique identifier for this view
	GetID() string

	// GetType returns the type of this view (e.g., "session", "task-list", "date-range")
	GetType() string

	// GetEvents returns the events contained in this view
	GetEvents() []Event

	// FormatForAnalysis formats the events as text for LLM analysis
	FormatForAnalysis() string

	// GetMetadata returns additional context for this view
	GetMetadata() map[string]interface{}
}

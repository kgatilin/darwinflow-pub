package claude_code

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// SessionView represents a view of all events in a single Claude Code session.
// It implements the pluginsdk.AnalysisView interface.
type SessionView struct {
	SessionID string
	Events    []pluginsdk.Event
}

// NewSessionView creates a new SessionView for a given session ID and events.
func NewSessionView(sessionID string, events []pluginsdk.Event) *SessionView {
	return &SessionView{
		SessionID: sessionID,
		Events:    events,
	}
}

// GetID returns the unique identifier for this view (the session ID).
func (v *SessionView) GetID() string {
	return v.SessionID
}

// GetType returns the type of this view.
func (v *SessionView) GetType() string {
	return "session"
}

// GetEvents returns the events contained in this view.
func (v *SessionView) GetEvents() []pluginsdk.Event {
	return v.Events
}

// FormatForAnalysis formats the events as markdown text suitable for LLM analysis.
// This uses a simple markdown format that groups events by type and includes timestamps.
func (v *SessionView) FormatForAnalysis() string {
	if len(v.Events) == 0 {
		return "# Event Logs\n\nNo events found.\n"
	}

	var buf bytes.Buffer

	// Write header
	buf.WriteString("# Event Logs\n\n")
	buf.WriteString("## Session: `" + v.SessionID + "`\n\n")
	buf.WriteString("This document contains event data for session analysis.\n\n")

	// Group events by type
	eventsByType := make(map[string][]pluginsdk.Event)
	for _, event := range v.Events {
		eventsByType[event.Type] = append(eventsByType[event.Type], event)
	}

	// Write events grouped by type
	buf.WriteString("## Event Summary\n\n")
	buf.WriteString("**Total Events**: ")
	fmt.Fprintf(&buf, "%d", len(v.Events))
	buf.WriteString("\n\n")

	buf.WriteString("**Event Types**:\n\n")
	for eventType, events := range eventsByType {
		buf.WriteString("- `")
		buf.WriteString(eventType)
		buf.WriteString("`: ")
		fmt.Fprintf(&buf, "%d", len(events))
		buf.WriteString(" event(s)\n")
	}

	// Write events in chronological order
	buf.WriteString("\n## Event Details\n\n")

	for i, event := range v.Events {
		buf.WriteString("### Event ")
		fmt.Fprintf(&buf, "%d", i+1)
		buf.WriteString("\n\n")

		buf.WriteString("**Type**: `")
		buf.WriteString(event.Type)
		buf.WriteString("`\n\n")

		buf.WriteString("**Timestamp**: ")
		buf.WriteString(event.Timestamp.Format("2006-01-02 15:04:05 MST"))
		buf.WriteString("\n\n")

		buf.WriteString("**Source**: ")
		buf.WriteString(event.Source)
		buf.WriteString("\n\n")

		if len(event.Metadata) > 0 {
			buf.WriteString("**Metadata**:\n\n")
			for key, value := range event.Metadata {
				buf.WriteString("- `")
				buf.WriteString(key)
				buf.WriteString("`: ")
				buf.WriteString(value)
				buf.WriteString("\n")
			}
			buf.WriteString("\n")
		}

		if len(event.Payload) > 0 {
			buf.WriteString("**Payload**:\n\n")
			buf.WriteString("```json\n")

			// Pretty-print JSON payload
			payloadJSON, _ := json.MarshalIndent(event.Payload, "", "  ")
			buf.Write(payloadJSON)

			buf.WriteString("\n```\n\n")
		}
	}

	return buf.String()
}

// GetMetadata returns additional context for this view.
func (v *SessionView) GetMetadata() map[string]interface{} {
	return map[string]interface{}{
		"session_id":  v.SessionID,
		"event_count": len(v.Events),
		"view_type":   "session",
	}
}

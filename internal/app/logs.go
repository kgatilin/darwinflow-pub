package app

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// LogRecord represents a formatted log entry for display
type LogRecord struct {
	ID        string
	Timestamp time.Time
	EventType string
	SessionID string
	Payload   json.RawMessage
	Content   string
}

// LogsService provides methods for querying and displaying logs
type LogsService struct {
	repo          domain.EventRepository
	rawExecutor   domain.RawQueryExecutor
}

// NewLogsService creates a new logs service
func NewLogsService(repo domain.EventRepository, rawExecutor domain.RawQueryExecutor) *LogsService {
	return &LogsService{
		repo:        repo,
		rawExecutor: rawExecutor,
	}
}

// ListRecentLogs retrieves the most recent N logs, optionally filtered by session ID and ordered chronologically
func (s *LogsService) ListRecentLogs(ctx context.Context, limit int, sessionID string, ordered bool) ([]*LogRecord, error) {
	query := domain.EventQuery{
		Limit:       limit,
		SessionID:   sessionID,
		OrderByTime: ordered,
	}

	events, err := s.repo.FindByQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}

	records := make([]*LogRecord, len(events))
	for i, event := range events {
		payloadBytes, err := event.MarshalPayload()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}

		records[i] = &LogRecord{
			ID:        event.ID,
			Timestamp: event.Timestamp,
			EventType: string(event.Type),
			SessionID: event.SessionID,
			Payload:   payloadBytes,
			Content:   event.Content,
		}
	}

	return records, nil
}

// ExecuteRawQuery executes an arbitrary SQL query
func (s *LogsService) ExecuteRawQuery(ctx context.Context, query string) (*domain.QueryResult, error) {
	return s.rawExecutor.ExecuteRawQuery(ctx, query)
}

// FormatLogRecord formats a single log record for display
func FormatLogRecord(index int, record *LogRecord) string {
	var output string

	output += fmt.Sprintf("[%d] %s\n", index+1, record.Timestamp.Format("2006-01-02 15:04:05.000"))
	output += fmt.Sprintf("    Event: %s\n", record.EventType)
	output += fmt.Sprintf("    ID: %s\n", record.ID)
	if record.SessionID != "" {
		output += fmt.Sprintf("    Session: %s\n", record.SessionID)
	}

	// Pretty print JSON payload with nested JSON expansion
	var payload interface{}
	if err := json.Unmarshal(record.Payload, &payload); err == nil {
		// Expand nested JSON strings in the payload
		expandedPayload := expandNestedJSON(payload)
		prettyPayload, _ := json.MarshalIndent(expandedPayload, "    ", "  ")
		output += fmt.Sprintf("    Payload: %s\n", string(prettyPayload))
	} else {
		output += fmt.Sprintf("    Payload: %s\n", string(record.Payload))
	}

	if record.Content != "" {
		// Truncate content if too long
		content := record.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		output += fmt.Sprintf("    Content: %s\n", content)
	}

	output += "\n"
	return output
}

// expandNestedJSON recursively expands JSON strings within a data structure
func expandNestedJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Recursively process map values
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = expandNestedJSON(value)
		}
		return result

	case []interface{}:
		// Recursively process array elements
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = expandNestedJSON(value)
		}
		return result

	case string:
		// Try to parse string as JSON
		if len(v) > 0 && (v[0] == '{' || v[0] == '[') {
			var parsed interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				// Successfully parsed, recursively expand
				return expandNestedJSON(parsed)
			}
		}
		// Not JSON or parsing failed, return as-is
		return v

	default:
		// Return other types as-is
		return v
	}
}

// FormatQueryValue formats a value from a raw query result for display
func FormatQueryValue(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "NULL"
	case []byte:
		// Try to parse as JSON for pretty printing
		var jsonObj interface{}
		if err := json.Unmarshal(v, &jsonObj); err == nil {
			jsonBytes, _ := json.Marshal(jsonObj)
			str := string(jsonBytes)
			if len(str) > 100 {
				str = str[:100] + "..."
			}
			return str
		}
		str := string(v)
		if len(str) > 100 {
			str = str[:100] + "..."
		}
		return str
	case string:
		if len(v) > 100 {
			return v[:100] + "..."
		}
		return v
	case int64:
		// Check if it might be a timestamp (13 digits for milliseconds)
		if v > 1000000000000 && v < 9999999999999 {
			t := time.UnixMilli(v)
			return fmt.Sprintf("%d (%s)", v, t.Format("2006-01-02 15:04:05"))
		}
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FormatLogsAsCSV writes log records as CSV to the provided writer
func FormatLogsAsCSV(w io.Writer, records []*LogRecord) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Timestamp", "EventType", "SessionID", "Payload", "Content"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write records
	for _, record := range records {
		row := []string{
			record.ID,
			record.Timestamp.Format(time.RFC3339),
			record.EventType,
			record.SessionID,
			string(record.Payload),
			record.Content,
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// FormatLogsAsMarkdown writes log records as Markdown to the provided writer
// Groups events by session and orders chronologically for LLM-friendly reading
func FormatLogsAsMarkdown(w io.Writer, records []*LogRecord) error {
	if len(records) == 0 {
		fmt.Fprintln(w, "# Event Logs")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No events found.")
		return nil
	}

	// Group records by session
	sessions := groupRecordsBySession(records)

	// Sort sessions by first event timestamp
	sortSessionsByTime(sessions)

	// Write output
	fmt.Fprintln(w, "# Event Logs")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "This document contains %d event(s) across %d session(s).\n", len(records), len(sessions))
	fmt.Fprintln(w)

	for sessionIdx, session := range sessions {
		// Session header
		if session.SessionID != "" {
			fmt.Fprintf(w, "## Session %d: `%s`\n\n", sessionIdx+1, session.SessionID)
		} else {
			fmt.Fprintf(w, "## Session %d: (No Session ID)\n\n", sessionIdx+1)
		}

		// Session metadata
		fmt.Fprintf(w, "**Session started**: %s\n\n", session.StartTime.Format("2006-01-02 15:04:05 MST"))
		fmt.Fprintf(w, "**Session duration**: %s\n\n", session.EndTime.Sub(session.StartTime).Round(time.Millisecond))
		fmt.Fprintf(w, "**Total events**: %d\n\n", len(session.Records))

		// Events in chronological order
		fmt.Fprintln(w, "### Event Timeline")
		fmt.Fprintln(w)

		for eventIdx, record := range session.Records {
			// Use narrative language
			var prefix string
			if eventIdx == 0 {
				prefix = "**Session started** - "
			} else {
				prefix = fmt.Sprintf("**Step %d** - ", eventIdx+1)
			}

			// Event type and timestamp
			duration := record.Timestamp.Sub(session.StartTime)
			fmt.Fprintf(w, "%s`%s` *(+%s)*\n\n", prefix, record.EventType, duration.Round(time.Millisecond))

			// Event ID
			fmt.Fprintf(w, "- **Event ID**: `%s`\n", record.ID)
			fmt.Fprintln(w)

			// Payload - expanded and formatted
			fmt.Fprintln(w, "**Details**:")
			fmt.Fprintln(w)

			var payload interface{}
			if err := json.Unmarshal(record.Payload, &payload); err == nil {
				// Expand nested JSON strings
				expandedPayload := expandNestedJSON(payload)
				if err := formatMarkdownPayload(w, expandedPayload, 0); err != nil {
					return fmt.Errorf("failed to format payload: %w", err)
				}
			} else {
				fmt.Fprintf(w, "```\n%s\n```\n", string(record.Payload))
			}
			fmt.Fprintln(w)

			// Content
			if record.Content != "" {
				fmt.Fprintln(w, "**Content Summary**:")
				fmt.Fprintln(w)
				fmt.Fprintf(w, "```\n%s\n```\n", record.Content)
				fmt.Fprintln(w)
			}
		}

		fmt.Fprintln(w, "---")
		fmt.Fprintln(w)
	}

	return nil
}

// SessionGroup represents events grouped by session
type SessionGroup struct {
	SessionID string
	Records   []*LogRecord
	StartTime time.Time
	EndTime   time.Time
}

// groupRecordsBySession groups log records by session ID and sorts chronologically
func groupRecordsBySession(records []*LogRecord) []*SessionGroup {
	sessionMap := make(map[string]*SessionGroup)

	for _, record := range records {
		sessionID := record.SessionID
		if sessionID == "" {
			sessionID = "(no-session)"
		}

		session, exists := sessionMap[sessionID]
		if !exists {
			session = &SessionGroup{
				SessionID: record.SessionID,
				Records:   []*LogRecord{},
				StartTime: record.Timestamp,
				EndTime:   record.Timestamp,
			}
			sessionMap[sessionID] = session
		}

		session.Records = append(session.Records, record)

		// Update start/end times
		if record.Timestamp.Before(session.StartTime) {
			session.StartTime = record.Timestamp
		}
		if record.Timestamp.After(session.EndTime) {
			session.EndTime = record.Timestamp
		}
	}

	// Sort records within each session chronologically
	for _, session := range sessionMap {
		sort.Slice(session.Records, func(i, j int) bool {
			return session.Records[i].Timestamp.Before(session.Records[j].Timestamp)
		})
	}

	// Convert map to slice
	sessions := make([]*SessionGroup, 0, len(sessionMap))
	for _, session := range sessionMap {
		sessions = append(sessions, session)
	}

	return sessions
}

// sortSessionsByTime sorts sessions by their start time
func sortSessionsByTime(sessions []*SessionGroup) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.Before(sessions[j].StartTime)
	})
}

// formatMarkdownPayload recursively formats a payload structure as Markdown
func formatMarkdownPayload(w io.Writer, data interface{}, depth int) error {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			// Skip the context field as it's not very useful in markdown
			if key == "context" {
				continue
			}

			switch val := value.(type) {
			case map[string]interface{}:
				fmt.Fprintf(w, "%s- **%s**:\n", indent, key)
				if err := formatMarkdownPayload(w, val, depth+1); err != nil {
					return err
				}
			case []interface{}:
				fmt.Fprintf(w, "%s- **%s**:\n", indent, key)
				if err := formatMarkdownPayload(w, val, depth+1); err != nil {
					return err
				}
			case string:
				if val == "" {
					fmt.Fprintf(w, "%s- **%s**: *(empty)*\n", indent, key)
				} else if len(val) > 200 {
					fmt.Fprintf(w, "%s- **%s**: `%s...`\n", indent, key, val[:200])
				} else {
					fmt.Fprintf(w, "%s- **%s**: `%s`\n", indent, key, val)
				}
			case nil:
				fmt.Fprintf(w, "%s- **%s**: `null`\n", indent, key)
			default:
				fmt.Fprintf(w, "%s- **%s**: `%v`\n", indent, key, val)
			}
		}
	case []interface{}:
		for i, item := range v {
			switch val := item.(type) {
			case map[string]interface{}:
				fmt.Fprintf(w, "%s%d.\n", indent, i+1)
				if err := formatMarkdownPayload(w, val, depth+1); err != nil {
					return err
				}
			case []interface{}:
				fmt.Fprintf(w, "%s%d.\n", indent, i+1)
				if err := formatMarkdownPayload(w, val, depth+1); err != nil {
					return err
				}
			default:
				fmt.Fprintf(w, "%s- `%v`\n", indent, val)
			}
		}
	default:
		fmt.Fprintf(w, "%s`%v`\n", indent, v)
	}

	return nil
}

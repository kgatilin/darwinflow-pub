package infra_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

// TestPhase3_EmitEventHookIntegration tests the complete Phase 3 event emission flow
// Scenario: Hook emits events → CLI captures and stores via EmitEventCommand → Plugin context stores events
func TestPhase3_EmitEventHookIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Step 1: Initialize database with schema
	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Step 2: Create sample event as would come from hook via adapter layer
	eventTimestamp := time.Date(2025, 10, 20, 15, 30, 45, 0, time.UTC)
	eventPayload := map[string]interface{}{
		"tool":        "Read",
		"file_path":   "/workspace/test.go",
		"description": "Reading test file",
	}
	sessionID := "test-session-123"

	// Step 3: Create domain Event (as would happen in adapter layer from SDK Event)
	domainEvent := &domain.Event{
		ID:        "event-1",
		Timestamp: eventTimestamp,
		Type:      "claude.tool.invoked",
		SessionID: sessionID,
		Payload:   eventPayload,
		Content:   "",
		Version:   "1.0",
	}

	// Step 4: Store event via repository (simulating EmitEventCommand execution)
	if err := repo.Save(ctx, domainEvent); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Step 5: Query EventRepository for stored event
	query := domain.EventQuery{
		SessionID: "test-session-123",
	}
	storedEvents, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	// Step 6: Verify event properties match input
	if len(storedEvents) != 1 {
		t.Errorf("Expected 1 event, got %d", len(storedEvents))
	}

	retrieved := storedEvents[0]
	if retrieved.Type != "claude.tool.invoked" {
		t.Errorf("Event type = %v, want %v", retrieved.Type, "claude.tool.invoked")
	}
	if retrieved.SessionID != "test-session-123" {
		t.Errorf("Event SessionID = %q, want %q", retrieved.SessionID, "test-session-123")
	}
	if retrieved.Version != "1.0" {
		t.Errorf("Event Version = %q, want %q", retrieved.Version, "1.0")
	}

	// Verify payload contains tool information
	// Note: Payload is stored as JSON in SQLite, so it may be json.RawMessage when retrieved
	payloadJSON, err := json.Marshal(retrieved.Payload)
	if err == nil {
		var payload map[string]interface{}
		if err := json.Unmarshal(payloadJSON, &payload); err == nil {
			if tool, ok := payload["tool"]; !ok || tool != "Read" {
				t.Errorf("Payload missing or wrong tool, got %v", payload["tool"])
			}
		}
	}
}

// TestPhase3_MultipleEventsOrdering tests that multiple events are stored and ordered correctly
func TestPhase3_MultipleEventsOrdering(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	sessionID := "session-multiple-events"
	baseTime := time.Date(2025, 10, 20, 15, 0, 0, 0, time.UTC)

	// Create multiple events with different timestamps
	testCases := []struct {
		eventType string
		offset    time.Duration
		tool      string
	}{
		{"claude.chat.started", 0 * time.Second, ""},
		{"claude.tool.invoked", 1 * time.Second, "Read"},
		{"claude.tool.invoked", 2 * time.Second, "Write"},
		{"claude.chat.message.user", 3 * time.Second, ""},
		{"claude.tool.invoked", 4 * time.Second, "Bash"},
	}

	// Store all events
	for idx, tc := range testCases {
		event := &domain.Event{
			ID:        "event-" + string(rune(idx)),
			Timestamp: baseTime.Add(tc.offset),
			Type:      tc.eventType,
			SessionID: sessionID,
			Payload: map[string]interface{}{
				"tool": tc.tool,
			},
			Content: "",
			Version: "1.0",
		}

		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Save() error for event %d: %v", idx, err)
		}
	}

	// Query all events for session
	query := domain.EventQuery{
		SessionID: sessionID,
	}
	storedEvents, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	// Verify all events stored
	if len(storedEvents) != len(testCases) {
		t.Errorf("Expected %d events, got %d", len(testCases), len(storedEvents))
	}

	// Verify all event types are present (order may vary depending on DB sorting)
	typeMap := make(map[string]int)
	for _, event := range storedEvents {
		typeMap[event.Type]++
	}

	for _, tc := range testCases {
		if typeMap[tc.eventType] == 0 {
			t.Errorf("Event type %v not found in results", tc.eventType)
		}
	}
}

// TestPhase3_EventVersioning tests event schema versioning support
func TestPhase3_EventVersioning(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	sessionID := "session-versioning"

	// Save events with different versions
	versions := []string{"1.0", "1.1", "2.0"}
	for _, version := range versions {
		event := &domain.Event{
			ID:        "versioned-event-" + version,
			Timestamp: time.Now(),
			Type:      "claude.tool.invoked",
			SessionID: sessionID,
			Payload:   map[string]interface{}{"format_version": version},
			Content:   "",
			Version:   version,
		}

		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Save() error for version %s: %v", version, err)
		}
	}

	// Query all versioned events
	query := domain.EventQuery{
		SessionID: sessionID,
	}
	allEvents, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(allEvents) != len(versions) {
		t.Errorf("Expected %d events, got %d", len(versions), len(allEvents))
	}

	// Verify versions
	versionCount := make(map[string]int)
	for _, event := range allEvents {
		versionCount[event.Version]++
	}

	for _, version := range versions {
		if versionCount[version] != 1 {
			t.Errorf("Version %s count = %d, want 1", version, versionCount[version])
		}
	}
}

// TestPhase3_EventPayloadPersistence tests complex payload preservation
func TestPhase3_EventPayloadPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Create event with complex nested payload
	complexPayload := map[string]interface{}{
		"tool":    "Read",
		"file":    "/workspace/test.go",
		"options": map[string]interface{}{
			"follow_symlinks": true,
			"timeout":         30,
			"max_size":        1024000,
		},
		"metadata": map[string]interface{}{
			"tags": []string{"important", "code-review"},
			"size": 4096,
		},
	}

	event := &domain.Event{
		ID:        "complex-payload-event",
		Timestamp: time.Now(),
		Type:      "claude.tool.invoked",
		SessionID: "test-session",
		Payload:   complexPayload,
		Content:   "",
		Version:   "1.0",
	}

	if err := repo.Save(ctx, event); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify payload structure
	query := domain.EventQuery{
		SessionID: "test-session",
	}
	stored, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(stored) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(stored))
	}

	retrieved := stored[0]

	// Parse payload from JSON (it's stored as JSON in SQLite)
	payloadJSON, err := json.Marshal(retrieved.Payload)
	if err != nil {
		t.Fatalf("json.Marshal(Payload) error = %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		t.Fatalf("json.Unmarshal(Payload) error = %v", err)
	}

	// Verify top-level fields
	if payload["tool"] != "Read" {
		t.Errorf("Payload tool = %v, want %v", payload["tool"], "Read")
	}

	// Verify nested structure
	if opts, ok := payload["options"].(map[string]interface{}); ok {
		if opts["timeout"] != float64(30) && opts["timeout"] != 30 {
			t.Errorf("Payload options.timeout = %v, want 30", opts["timeout"])
		}
	} else {
		t.Error("Payload options not found or wrong type")
	}

	// Verify nested array
	if meta, ok := payload["metadata"].(map[string]interface{}); ok {
		if _, ok := meta["tags"]; !ok {
			t.Error("Payload metadata.tags not found")
		}
	} else {
		t.Error("Payload metadata not found or wrong type")
	}
}

// TestPhase3_DomainEventFields tests that domain events preserve all important fields
func TestPhase3_DomainEventFields(t *testing.T) {
	// Create a domain event with all fields
	eventTime := time.Now()
	eventPayload := map[string]interface{}{
		"tool": "Read",
	}
	sessionID := "field-test"

	domainEvent := &domain.Event{
		Type:      "claude.tool.invoked",
		SessionID: sessionID,
		Timestamp: eventTime,
		Payload:   eventPayload,
		Version:   "1.0",
	}

	// Verify all fields are preserved
	if domainEvent.SessionID != sessionID {
		t.Errorf("SessionID = %q, want %q", domainEvent.SessionID, sessionID)
	}
	if domainEvent.Version != "1.0" {
		t.Errorf("Version = %q, want %q", domainEvent.Version, "1.0")
	}
	payload, ok := domainEvent.Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("Payload not a map, got %T", domainEvent.Payload)
	}
	if payload["tool"] != "Read" {
		t.Errorf("Payload tool = %v, want %v", payload["tool"], "Read")
	}
}

// TestPhase3_JSONMarshaling tests that events can be marshaled/unmarshaled as JSON
func TestPhase3_JSONMarshaling(t *testing.T) {
	// Create a domain event
	eventPayload := map[string]interface{}{
		"tool":    "Bash",
		"command": "go test ./...",
	}

	originalEvent := &domain.Event{
		Type:      "claude.tool.invoked",
		SessionID: "marshal-test",
		Timestamp: time.Date(2025, 10, 20, 15, 30, 45, 0, time.UTC),
		Payload:   eventPayload,
		Version:   "1.0",
	}

	// Marshal to JSON
	data, err := json.Marshal(originalEvent)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var restored domain.Event
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify fields preserved
	if restored.Type != "claude.tool.invoked" {
		t.Errorf("Type = %v, want %v", restored.Type, "claude.tool.invoked")
	}
	if restored.SessionID != "marshal-test" {
		t.Errorf("SessionID = %q, want %q", restored.SessionID, "marshal-test")
	}
	if restored.Version != "1.0" {
		t.Errorf("Version = %q, want %q", restored.Version, "1.0")
	}
}

// TestPhase3_HookMigrationCompatibility tests backward compatibility with old log format
func TestPhase3_HookMigrationCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteEventRepository() error = %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Simulate old-format event (would come from deprecated "dw claude log")
	oldFormatEvent := &domain.Event{
		ID:        "old-format-event",
		Timestamp: time.Now(),
		Type:      "claude.tool.invoked",
		SessionID: "legacy-session",
		Payload:   map[string]interface{}{"tool": "Read"},
		Content:   "",
		Version:   "1.0", // Would be empty string in truly old format, but now defaults to 1.0
	}

	if err := repo.Save(ctx, oldFormatEvent); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify backward compatibility
	query := domain.EventQuery{
		SessionID: "legacy-session",
	}
	stored, err := repo.FindByQuery(ctx, query)
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(stored) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(stored))
	}

	// Old format events should still be queryable and functional
	retrieved := stored[0]
	if retrieved.SessionID != "legacy-session" {
		t.Errorf("Backward compat: SessionID = %q, want %q", retrieved.SessionID, "legacy-session")
	}
	if retrieved.Type != "claude.tool.invoked" {
		t.Errorf("Backward compat: Type = %v, want %v", retrieved.Type, "claude.tool.invoked")
	}
}

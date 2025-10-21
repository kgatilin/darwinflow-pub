package infra_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestSQLiteEventRepository_Save_WithVersion(t *testing.T) {
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

	// Test: Event with explicit version
	event := &domain.Event{
		ID:        "test-event-1",
		Timestamp: time.Now(),
		Type:      string(domain.ChatStarted),
		SessionID: "session-1",
		Payload:   map[string]interface{}{"msg": "test"},
		Content:   "test content",
		Version:   "2.0",
	}

	if err := repo.Save(ctx, event); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify version is persisted
	events, err := repo.FindByQuery(ctx, domain.EventQuery{})
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	retrieved := events[0]
	if retrieved.Version != "2.0" {
		t.Errorf("Event version = %q, want %q", retrieved.Version, "2.0")
	}
	if retrieved.ID != "test-event-1" {
		t.Errorf("Event ID = %q, want %q", retrieved.ID, "test-event-1")
	}
}

func TestSQLiteEventRepository_Save_DefaultVersion(t *testing.T) {
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

	// Test: Event with default version "1.0"
	event := domain.NewEvent(string(domain.ChatStarted), "session-2", map[string]interface{}{"msg": "test"}, "test")

	if err := repo.Save(ctx, event); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Retrieve and verify default version is persisted
	events, err := repo.FindByQuery(ctx, domain.EventQuery{})
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	retrieved := events[0]
	if retrieved.Version != "1.0" {
		t.Errorf("Event version = %q, want %q", retrieved.Version, "1.0")
	}
}

func TestSQLiteEventRepository_MultipleVersions(t *testing.T) {
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

	// Save multiple events with different versions
	testCases := []struct {
		id      string
		version string
	}{
		{"event-v1", "1.0"},
		{"event-v2", "2.0"},
		{"event-v3", "1.0"},
	}

	for _, tc := range testCases {
		event := &domain.Event{
			ID:        tc.id,
			Timestamp: time.Now(),
			Type:      string(domain.ChatStarted),
			SessionID: "session-3",
			Payload:   map[string]interface{}{"version": tc.version},
			Content:   "test",
			Version:   tc.version,
		}

		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Save() error for %s: %v", tc.id, err)
		}
	}

	// Retrieve all events
	events, err := repo.FindByQuery(ctx, domain.EventQuery{})
	if err != nil {
		t.Fatalf("FindByQuery() error = %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Verify versions match what was saved
	versionMap := make(map[string]string)
	for _, e := range events {
		versionMap[e.ID] = e.Version
	}

	for _, tc := range testCases {
		if got := versionMap[tc.id]; got != tc.version {
			t.Errorf("Event %s version = %q, want %q", tc.id, got, tc.version)
		}
	}
}

func TestNewEvent_DefaultVersion(t *testing.T) {
	event := domain.NewEvent(string(domain.ChatStarted), "session-4", map[string]interface{}{}, "test")

	if event.Version != "1.0" {
		t.Errorf("NewEvent() version = %q, want %q", event.Version, "1.0")
	}
}

func TestEvent_VersionPreservation(t *testing.T) {
	// Create event with custom version
	event := &domain.Event{
		ID:        "custom-v-event",
		Timestamp: time.Now(),
		Type:      string(domain.ChatStarted),
		SessionID: "test-session",
		Payload:   map[string]interface{}{},
		Content:   "test",
		Version:   "3.5",
	}

	if event.Version != "3.5" {
		t.Errorf("Event version = %q, want %q", event.Version, "3.5")
	}
}

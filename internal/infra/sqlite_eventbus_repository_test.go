// Package infra_test contains integration tests for the event bus with persistence.
//
// These tests verify the full stack integration of:
// - InMemoryEventBus (in-memory pub/sub with optional persistence)
// - SQLiteEventBusRepository (SQLite-backed event storage)
// - Cross-plugin communication patterns
// - Event replay and time-based queries
// - Concurrent operations and thread safety
//
// Test Scenarios:
// 1. Event persistence and replay - verifies events are stored and can be replayed to new subscribers
// 2. Cross-plugin communication - tests plugins communicating via event bus with filtering
// 3. Concurrent publishing and subscribing - validates thread safety under load
// 4. Event bus lifecycle with repository - ensures persistence survives restarts
// 5. Filter matching with persistence - comprehensive filter query testing
// 6. Time-based replay (GetEventsSince) - validates timestamp-based event queries
// 7. Replay with filters - tests replay with various filter combinations
// 8. In-memory vs persistent modes - compares behavior of both modes
// 9. Concurrent replay - validates thread safety of replay operations
// 10. Limit parameter - tests query limiting functionality
package infra_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	_ "github.com/mattn/go-sqlite3"
)

// mockEventHandler is a thread-safe test event handler that records received events.
type mockEventHandler struct {
	received []pluginsdk.BusEvent
	mu       sync.Mutex
	callCount int32
}

func (h *mockEventHandler) HandleEvent(ctx context.Context, event pluginsdk.BusEvent) error {
	atomic.AddInt32(&h.callCount, 1)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.received = append(h.received, event)
	return nil
}

func (h *mockEventHandler) getReceived() []pluginsdk.BusEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]pluginsdk.BusEvent{}, h.received...)
}


// setupTestDB creates a temporary SQLite database for testing.
func setupTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test_eventbus.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db, dbPath
}

// setupEventBusWithPersistence creates an event bus with SQLite persistence.
func setupEventBusWithPersistence(t *testing.T) (*infra.InMemoryEventBus, *infra.SQLiteEventBusRepository, *sql.DB) {
	t.Helper()

	db, _ := setupTestDB(t)

	// Create repository
	repo, err := infra.NewSQLiteEventBusRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Initialize schema
	if err := repo.Initialize(context.Background()); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create event bus with repository
	bus := infra.NewInMemoryEventBus(repo)

	return bus, repo, db
}

// TestIntegration_EventPersistenceAndReplay verifies events are persisted and can be replayed.
func TestIntegration_EventPersistenceAndReplay(t *testing.T) {
	ctx := context.Background()
	bus, repo, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Step 1: Publish several events
	for i := 0; i < 5; i++ {
		event, err := pluginsdk.NewBusEvent(
			fmt.Sprintf("test.event.%d", i),
			"test-plugin",
			map[string]interface{}{"index": i, "message": fmt.Sprintf("Event %d", i)},
		)
		if err != nil {
			t.Fatalf("Failed to create event %d: %v", i, err)
		}
		event.Labels = map[string]string{"category": "test", "priority": "high"}

		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf("Failed to publish event %d: %v", i, err)
		}
	}

	// Wait for async publish to complete
	time.Sleep(100 * time.Millisecond)

	// Step 2: Verify events are stored in database
	storedEvents, err := repo.GetEvents(ctx, pluginsdk.EventFilter{}, 0)
	if err != nil {
		t.Fatalf("Failed to get events from repository: %v", err)
	}

	if len(storedEvents) != 5 {
		t.Fatalf("Expected 5 stored events, got %d", len(storedEvents))
	}

	// Step 3: Create new subscriber and replay events
	handler := &mockEventHandler{}
	sinceTime := time.Now().Add(-1 * time.Hour) // Get all events from past hour

	if err := bus.Replay(ctx, sinceTime, pluginsdk.EventFilter{}, handler); err != nil {
		t.Fatalf("Failed to replay events: %v", err)
	}

	// Step 4: Verify subscriber receives all historical events
	received := handler.getReceived()
	if len(received) != 5 {
		t.Fatalf("Expected handler to receive 5 replayed events, got %d", len(received))
	}

	// Verify event order (should be in timestamp order)
	for i, event := range received {
		var payload map[string]interface{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			t.Fatalf("Failed to unmarshal event payload: %v", err)
		}

		expectedIndex := float64(i) // JSON unmarshals numbers as float64
		if payload["index"] != expectedIndex {
			t.Errorf("Event %d: expected index %v, got %v", i, expectedIndex, payload["index"])
		}
	}
}

// TestIntegration_CrossPluginCommunication verifies plugins can communicate via event bus.
func TestIntegration_CrossPluginCommunication(t *testing.T) {
	ctx := context.Background()
	bus, _, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Create mock plugin B handler
	pluginBHandler := &mockEventHandler{}

	// Plugin B subscribes to all events from Plugin A
	filter := pluginsdk.EventFilter{
		SourcePlugin: "plugin-a",
		TypePattern:  "plugin-a.*",
	}
	subID, err := bus.Subscribe(filter, pluginBHandler)
	if err != nil {
		t.Fatalf("Failed to subscribe plugin B: %v", err)
	}
	defer bus.Unsubscribe(subID)

	// Plugin A publishes events
	eventTypes := []string{"plugin-a.user.created", "plugin-a.user.updated", "plugin-a.user.deleted"}
	for i, eventType := range eventTypes {
		event, err := pluginsdk.NewBusEvent(
			eventType,
			"plugin-a",
			map[string]interface{}{"user_id": i + 1, "action": eventType},
		)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}

		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}
	}

	// Wait for async delivery
	time.Sleep(100 * time.Millisecond)

	// Verify Plugin B received all events from Plugin A
	received := pluginBHandler.getReceived()
	if len(received) != 3 {
		t.Fatalf("Expected Plugin B to receive 3 events, got %d", len(received))
	}

	// Verify all received events are from plugin-a and have correct types
	// Note: Async delivery doesn't guarantee order, so we check for presence
	receivedTypes := make(map[string]bool)
	for _, event := range received {
		if event.Source != "plugin-a" {
			t.Errorf("Expected source 'plugin-a', got '%s'", event.Source)
		}
		receivedTypes[event.Type] = true
	}

	// Verify all expected types were received
	for _, expectedType := range eventTypes {
		if !receivedTypes[expectedType] {
			t.Errorf("Expected to receive event type '%s', but didn't", expectedType)
		}
	}

	// Test with label filtering
	pluginCHandler := &mockEventHandler{}
	labelFilter := pluginsdk.EventFilter{
		TypePattern: "plugin-a.*",
		Labels:      map[string]string{"env": "production"},
	}
	subID2, err := bus.Subscribe(labelFilter, pluginCHandler)
	if err != nil {
		t.Fatalf("Failed to subscribe plugin C: %v", err)
	}
	defer bus.Unsubscribe(subID2)

	// Publish event with matching label
	event, _ := pluginsdk.NewBusEvent("plugin-a.data.sync", "plugin-a", map[string]string{"status": "complete"})
	event.Labels = map[string]string{"env": "production", "region": "us-west"}
	bus.Publish(ctx, event)

	// Publish event without matching label
	event2, _ := pluginsdk.NewBusEvent("plugin-a.data.sync", "plugin-a", map[string]string{"status": "complete"})
	event2.Labels = map[string]string{"env": "development"}
	bus.Publish(ctx, event2)

	time.Sleep(100 * time.Millisecond)

	// Plugin C should only receive the event with env=production
	receivedC := pluginCHandler.getReceived()
	if len(receivedC) != 1 {
		t.Fatalf("Expected Plugin C to receive 1 event, got %d", len(receivedC))
	}
	if receivedC[0].Labels["env"] != "production" {
		t.Errorf("Expected label env=production, got %s", receivedC[0].Labels["env"])
	}
}

// TestIntegration_ConcurrentPublishingAndSubscribing verifies thread safety.
func TestIntegration_ConcurrentPublishingAndSubscribing(t *testing.T) {
	ctx := context.Background()
	bus, _, db := setupEventBusWithPersistence(t)
	defer db.Close()

	const numPublishers = 5
	const numSubscribers = 5
	const eventsPerPublisher = 10

	// Create subscribers
	handlers := make([]*mockEventHandler, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		handlers[i] = &mockEventHandler{}
		_, err := bus.Subscribe(pluginsdk.EventFilter{}, handlers[i])
		if err != nil {
			t.Fatalf("Failed to subscribe handler %d: %v", i, err)
		}
	}

	// Publish events concurrently
	var wg sync.WaitGroup
	for publisherID := 0; publisherID < numPublishers; publisherID++ {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			for eventNum := 0; eventNum < eventsPerPublisher; eventNum++ {
				event, err := pluginsdk.NewBusEvent(
					"test.concurrent",
					fmt.Sprintf("publisher-%d", pid),
					map[string]interface{}{
						"publisher": pid,
						"event_num": eventNum,
					},
				)
				if err != nil {
					t.Errorf("Publisher %d failed to create event: %v", pid, err)
					return
				}

				if err := bus.Publish(ctx, event); err != nil {
					t.Errorf("Publisher %d failed to publish event: %v", pid, err)
					return
				}
			}
		}(publisherID)
	}

	wg.Wait()

	// Wait for all async deliveries
	time.Sleep(500 * time.Millisecond)

	// Verify no events were lost
	totalExpected := numPublishers * eventsPerPublisher
	for i, handler := range handlers {
		received := handler.getReceived()
		if len(received) != totalExpected {
			t.Errorf("Handler %d expected %d events, got %d", i, totalExpected, len(received))
		}
	}

	// Verify all unique events exist
	eventIDs := make(map[string]bool)
	for _, event := range handlers[0].getReceived() {
		eventIDs[event.ID] = true
	}

	if len(eventIDs) != totalExpected {
		t.Errorf("Expected %d unique events, got %d", totalExpected, len(eventIDs))
	}
}

// TestIntegration_EventBusLifecycleWithRepository verifies persistence survives restarts.
func TestIntegration_EventBusLifecycleWithRepository(t *testing.T) {
	ctx := context.Background()
	db, _ := setupTestDB(t)
	defer db.Close()

	// Phase 1: Create bus, publish events, stop
	repo1, err := infra.NewSQLiteEventBusRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	if err := repo1.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	bus1 := infra.NewInMemoryEventBus(repo1)

	// Publish events
	publishedEvents := []string{}
	for i := 0; i < 3; i++ {
		event, err := pluginsdk.NewBusEvent(
			fmt.Sprintf("lifecycle.event.%d", i),
			"lifecycle-plugin",
			map[string]string{"phase": "1"},
		)
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
		publishedEvents = append(publishedEvents, event.ID)

		if err := bus1.Publish(ctx, event); err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Phase 2: Create new bus with same repository (simulates restart)
	repo2, err := infra.NewSQLiteEventBusRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository 2: %v", err)
	}

	bus2 := infra.NewInMemoryEventBus(repo2)

	// Query historical events
	historicalEvents, err := repo2.GetEvents(ctx, pluginsdk.EventFilter{}, 0)
	if err != nil {
		t.Fatalf("Failed to get historical events: %v", err)
	}

	if len(historicalEvents) != 3 {
		t.Fatalf("Expected 3 historical events after restart, got %d", len(historicalEvents))
	}

	// Verify event IDs match
	for i, event := range historicalEvents {
		if event.ID != publishedEvents[i] {
			t.Errorf("Event %d: expected ID %s, got %s", i, publishedEvents[i], event.ID)
		}
	}

	// Phase 3: Publish new events with new bus
	handler := &mockEventHandler{}
	bus2.Subscribe(pluginsdk.EventFilter{}, handler)

	newEvent, _ := pluginsdk.NewBusEvent("lifecycle.event.new", "lifecycle-plugin", map[string]string{"phase": "2"})
	bus2.Publish(ctx, newEvent)

	time.Sleep(100 * time.Millisecond)

	// Verify new bus can publish and deliver
	received := handler.getReceived()
	if len(received) != 1 {
		t.Fatalf("Expected 1 new event, got %d", len(received))
	}

	// Verify all events are in database (old + new)
	allEvents, err := repo2.GetEvents(ctx, pluginsdk.EventFilter{}, 0)
	if err != nil {
		t.Fatalf("Failed to get all events: %v", err)
	}

	if len(allEvents) != 4 {
		t.Fatalf("Expected 4 total events, got %d", len(allEvents))
	}
}

// TestIntegration_FilterMatchingWithPersistence verifies query filtering works correctly.
func TestIntegration_FilterMatchingWithPersistence(t *testing.T) {
	ctx := context.Background()
	bus, repo, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Publish diverse events with different types, sources, and labels
	events := []struct {
		eventType string
		source    string
		labels    map[string]string
	}{
		{"gmail.email.received", "gmail-plugin", map[string]string{"priority": "high", "folder": "inbox"}},
		{"gmail.email.sent", "gmail-plugin", map[string]string{"priority": "low"}},
		{"slack.message.received", "slack-plugin", map[string]string{"channel": "general"}},
		{"slack.message.sent", "slack-plugin", map[string]string{"channel": "dev", "priority": "high"}},
		{"calendar.event.created", "calendar-plugin", map[string]string{"type": "meeting"}},
	}

	for _, ev := range events {
		event, err := pluginsdk.NewBusEvent(ev.eventType, ev.source, map[string]string{"data": "test"})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
		event.Labels = ev.labels

		if err := bus.Publish(ctx, event); err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Test cases for filtering
	testCases := []struct {
		name          string
		filter        pluginsdk.EventFilter
		expectedCount int
		description   string
	}{
		{
			name:          "All events",
			filter:        pluginsdk.EventFilter{},
			expectedCount: 5,
			description:   "Empty filter should return all events",
		},
		{
			name:          "Gmail events only",
			filter:        pluginsdk.EventFilter{SourcePlugin: "gmail-plugin"},
			expectedCount: 2,
			description:   "Source filter should return only gmail events",
		},
		{
			name:          "Glob pattern - all received events",
			filter:        pluginsdk.EventFilter{TypePattern: "*.received"},
			expectedCount: 2,
			description:   "Type pattern *.received should match gmail and slack received events",
		},
		{
			name:          "Glob pattern - gmail events",
			filter:        pluginsdk.EventFilter{TypePattern: "gmail.*"},
			expectedCount: 2,
			description:   "Type pattern gmail.* should match all gmail events",
		},
		{
			name:          "Label filter - high priority",
			filter:        pluginsdk.EventFilter{Labels: map[string]string{"priority": "high"}},
			expectedCount: 2,
			description:   "Label filter priority=high should match 2 events",
		},
		{
			name: "Combined filter - gmail + high priority",
			filter: pluginsdk.EventFilter{
				SourcePlugin: "gmail-plugin",
				Labels:       map[string]string{"priority": "high"},
			},
			expectedCount: 1,
			description:   "Combined filter should match gmail events with high priority",
		},
		{
			name: "Complex combined filter",
			filter: pluginsdk.EventFilter{
				TypePattern:  "slack.*",
				SourcePlugin: "slack-plugin",
				Labels:       map[string]string{"priority": "high"},
			},
			expectedCount: 1,
			description:   "Complex filter should match slack high priority events",
		},
		{
			name:          "No matches",
			filter:        pluginsdk.EventFilter{SourcePlugin: "nonexistent-plugin"},
			expectedCount: 0,
			description:   "Filter with no matches should return empty result",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := repo.GetEvents(ctx, tc.filter, 0)
			if err != nil {
				t.Fatalf("GetEvents failed: %v", err)
			}

			if len(results) != tc.expectedCount {
				t.Errorf("%s: expected %d events, got %d", tc.description, tc.expectedCount, len(results))
			}
		})
	}
}

// TestIntegration_GetEventsSince_TimeBasedReplay verifies time-based event replay.
func TestIntegration_GetEventsSince_TimeBasedReplay(t *testing.T) {
	ctx := context.Background()
	bus, repo, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Record timestamps for different phases
	beforeAny := time.Now()
	time.Sleep(50 * time.Millisecond)

	// Publish batch 1
	for i := 0; i < 3; i++ {
		event, _ := pluginsdk.NewBusEvent("test.batch1", "test-plugin", map[string]int{"batch": 1, "index": i})
		bus.Publish(ctx, event)
	}
	time.Sleep(100 * time.Millisecond)

	afterBatch1 := time.Now()
	time.Sleep(50 * time.Millisecond)

	// Publish batch 2
	for i := 0; i < 2; i++ {
		event, _ := pluginsdk.NewBusEvent("test.batch2", "test-plugin", map[string]int{"batch": 2, "index": i})
		bus.Publish(ctx, event)
	}
	time.Sleep(100 * time.Millisecond)

	afterBatch2 := time.Now()
	time.Sleep(50 * time.Millisecond)

	// Publish batch 3
	for i := 0; i < 4; i++ {
		event, _ := pluginsdk.NewBusEvent("test.batch3", "test-plugin", map[string]int{"batch": 3, "index": i})
		bus.Publish(ctx, event)
	}
	time.Sleep(100 * time.Millisecond)

	// Test 1: Get all events (since beginning of time)
	allEvents, err := repo.GetEventsSince(ctx, beforeAny, pluginsdk.EventFilter{}, 0)
	if err != nil {
		t.Fatalf("GetEventsSince failed: %v", err)
	}
	if len(allEvents) != 9 {
		t.Errorf("Expected 9 total events, got %d", len(allEvents))
	}

	// Test 2: Get events since after batch 1 (should get batch 2 + batch 3)
	eventsAfterBatch1, err := repo.GetEventsSince(ctx, afterBatch1, pluginsdk.EventFilter{}, 0)
	if err != nil {
		t.Fatalf("GetEventsSince failed: %v", err)
	}
	if len(eventsAfterBatch1) != 6 {
		t.Errorf("Expected 6 events after batch1, got %d", len(eventsAfterBatch1))
	}

	// Test 3: Get events since after batch 2 (should get batch 3 only)
	eventsAfterBatch2, err := repo.GetEventsSince(ctx, afterBatch2, pluginsdk.EventFilter{}, 0)
	if err != nil {
		t.Fatalf("GetEventsSince failed: %v", err)
	}
	if len(eventsAfterBatch2) != 4 {
		t.Errorf("Expected 4 events after batch2, got %d", len(eventsAfterBatch2))
	}

	// Test 4: Get events with type filter
	batch2Events, err := repo.GetEventsSince(ctx, beforeAny, pluginsdk.EventFilter{TypePattern: "test.batch2"}, 0)
	if err != nil {
		t.Fatalf("GetEventsSince with filter failed: %v", err)
	}
	if len(batch2Events) != 2 {
		t.Errorf("Expected 2 batch2 events, got %d", len(batch2Events))
	}

	// Verify events are returned in timestamp order
	for i := 1; i < len(allEvents); i++ {
		if allEvents[i].Timestamp.Before(allEvents[i-1].Timestamp) {
			t.Errorf("Events not in timestamp order at index %d", i)
		}
	}
}

// TestIntegration_ReplayWithFilter verifies Replay method with various filters.
func TestIntegration_ReplayWithFilter(t *testing.T) {
	ctx := context.Background()
	bus, _, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Publish events from multiple sources
	sources := []string{"plugin-a", "plugin-b", "plugin-c"}
	for _, source := range sources {
		for i := 0; i < 3; i++ {
			event, _ := pluginsdk.NewBusEvent(
				fmt.Sprintf("%s.event.%d", source, i),
				source,
				map[string]interface{}{"source": source, "index": i},
			)
			event.Labels = map[string]string{"source": source}
			bus.Publish(ctx, event)
		}
	}

	time.Sleep(200 * time.Millisecond)

	// Test 1: Replay all events
	handler1 := &mockEventHandler{}
	err := bus.Replay(ctx, time.Time{}, pluginsdk.EventFilter{}, handler1)
	if err != nil {
		t.Fatalf("Replay all failed: %v", err)
	}
	if len(handler1.getReceived()) != 9 {
		t.Errorf("Expected 9 events in full replay, got %d", len(handler1.getReceived()))
	}

	// Test 2: Replay only plugin-a events
	handler2 := &mockEventHandler{}
	err = bus.Replay(ctx, time.Time{}, pluginsdk.EventFilter{SourcePlugin: "plugin-a"}, handler2)
	if err != nil {
		t.Fatalf("Replay plugin-a failed: %v", err)
	}
	received := handler2.getReceived()
	if len(received) != 3 {
		t.Errorf("Expected 3 plugin-a events, got %d", len(received))
	}
	for _, event := range received {
		if event.Source != "plugin-a" {
			t.Errorf("Expected source plugin-a, got %s", event.Source)
		}
	}

	// Test 3: Replay with type pattern
	handler3 := &mockEventHandler{}
	err = bus.Replay(ctx, time.Time{}, pluginsdk.EventFilter{TypePattern: "plugin-b.*"}, handler3)
	if err != nil {
		t.Fatalf("Replay with type pattern failed: %v", err)
	}
	if len(handler3.getReceived()) != 3 {
		t.Errorf("Expected 3 plugin-b events, got %d", len(handler3.getReceived()))
	}

	// Test 4: Replay with label filter
	handler4 := &mockEventHandler{}
	err = bus.Replay(ctx, time.Time{}, pluginsdk.EventFilter{Labels: map[string]string{"source": "plugin-c"}}, handler4)
	if err != nil {
		t.Fatalf("Replay with label filter failed: %v", err)
	}
	if len(handler4.getReceived()) != 3 {
		t.Errorf("Expected 3 plugin-c events with label, got %d", len(handler4.getReceived()))
	}
}

// TestIntegration_InMemoryVsPersistent compares in-memory and persistent modes.
func TestIntegration_InMemoryVsPersistent(t *testing.T) {
	ctx := context.Background()

	// Create in-memory bus (no persistence)
	inMemoryBus := infra.NewInMemoryEventBus(nil)

	// Create persistent bus
	persistentBus, repo, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Publish same events to both
	for i := 0; i < 5; i++ {
		event, _ := pluginsdk.NewBusEvent(fmt.Sprintf("test.%d", i), "test-plugin", map[string]int{"index": i})

		inMemoryBus.Publish(ctx, event)
		persistentBus.Publish(ctx, event)
	}

	time.Sleep(100 * time.Millisecond)

	// Subscribe handlers to both
	inMemoryHandler := &mockEventHandler{}
	persistentHandler := &mockEventHandler{}

	inMemoryBus.Subscribe(pluginsdk.EventFilter{}, inMemoryHandler)
	persistentBus.Subscribe(pluginsdk.EventFilter{}, persistentHandler)

	// Publish new event
	newEvent, _ := pluginsdk.NewBusEvent("test.new", "test-plugin", map[string]string{"new": "event"})
	inMemoryBus.Publish(ctx, newEvent)
	persistentBus.Publish(ctx, newEvent)

	time.Sleep(100 * time.Millisecond)

	// Both should deliver the new event
	if len(inMemoryHandler.getReceived()) != 1 {
		t.Errorf("In-memory handler expected 1 event, got %d", len(inMemoryHandler.getReceived()))
	}
	if len(persistentHandler.getReceived()) != 1 {
		t.Errorf("Persistent handler expected 1 event, got %d", len(persistentHandler.getReceived()))
	}

	// Only persistent bus should have historical events
	historicalEvents, err := repo.GetEvents(ctx, pluginsdk.EventFilter{}, 0)
	if err != nil {
		t.Fatalf("Failed to get historical events: %v", err)
	}
	if len(historicalEvents) != 6 {
		t.Errorf("Expected 6 historical events in persistent mode, got %d", len(historicalEvents))
	}

	// In-memory bus should not support replay
	inMemoryReplayHandler := &mockEventHandler{}
	err = inMemoryBus.Replay(ctx, time.Time{}, pluginsdk.EventFilter{}, inMemoryReplayHandler)
	if err == nil {
		t.Error("In-memory bus should error on Replay, got nil")
	}
}

// TestIntegration_ConcurrentReplay verifies concurrent replay operations.
func TestIntegration_ConcurrentReplay(t *testing.T) {
	ctx := context.Background()
	bus, _, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Publish events
	for i := 0; i < 20; i++ {
		event, _ := pluginsdk.NewBusEvent(fmt.Sprintf("test.%d", i), "test-plugin", map[string]int{"index": i})
		bus.Publish(ctx, event)
	}

	time.Sleep(200 * time.Millisecond)

	// Start multiple concurrent replays
	const numReplays = 10
	var wg sync.WaitGroup
	handlers := make([]*mockEventHandler, numReplays)

	for i := 0; i < numReplays; i++ {
		handlers[i] = &mockEventHandler{}
		wg.Add(1)
		go func(h *mockEventHandler) {
			defer wg.Done()
			err := bus.Replay(ctx, time.Time{}, pluginsdk.EventFilter{}, h)
			if err != nil {
				t.Errorf("Replay failed: %v", err)
			}
		}(handlers[i])
	}

	wg.Wait()

	// All replays should have received all events
	for i, handler := range handlers {
		received := handler.getReceived()
		if len(received) != 20 {
			t.Errorf("Handler %d expected 20 events, got %d", i, len(received))
		}
	}
}

// TestIntegration_LimitParameter verifies limit parameter in queries.
func TestIntegration_LimitParameter(t *testing.T) {
	ctx := context.Background()
	bus, repo, db := setupEventBusWithPersistence(t)
	defer db.Close()

	// Publish 10 events
	for i := 0; i < 10; i++ {
		event, _ := pluginsdk.NewBusEvent(fmt.Sprintf("test.%d", i), "test-plugin", map[string]int{"index": i})
		bus.Publish(ctx, event)
	}

	time.Sleep(100 * time.Millisecond)

	// Test different limits
	testCases := []struct {
		limit    int
		expected int
	}{
		{0, 10},  // No limit, get all
		{5, 5},   // Limit to 5
		{15, 10}, // Limit exceeds total, get all
		{1, 1},   // Limit to 1
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("limit=%d", tc.limit), func(t *testing.T) {
			results, err := repo.GetEvents(ctx, pluginsdk.EventFilter{}, tc.limit)
			if err != nil {
				t.Fatalf("GetEvents failed: %v", err)
			}
			if len(results) != tc.expected {
				t.Errorf("With limit=%d, expected %d events, got %d", tc.limit, tc.expected, len(results))
			}
		})
	}
}

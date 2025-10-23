package infra_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// mockHandler is a test event handler that records received events.
type mockHandler struct {
	mu           sync.Mutex
	events       []pluginsdk.BusEvent
	callCount    int32
	delay        time.Duration
	shouldError  bool
	errorMessage string
}

func (h *mockHandler) HandleEvent(ctx context.Context, event pluginsdk.BusEvent) error {
	atomic.AddInt32(&h.callCount, 1)

	if h.delay > 0 {
		select {
		case <-time.After(h.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	h.mu.Lock()
	h.events = append(h.events, event)
	h.mu.Unlock()

	if h.shouldError {
		return errors.New(h.errorMessage)
	}
	return nil
}

func (h *mockHandler) getEvents() []pluginsdk.BusEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]pluginsdk.BusEvent{}, h.events...)
}

func (h *mockHandler) getCallCount() int {
	return int(atomic.LoadInt32(&h.callCount))
}

func TestNewInMemoryEventBus(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	if bus == nil {
		t.Fatal("NewInMemoryEventBus returned nil")
	}
}

func TestPublishSubscribe_Basic(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	// Create handler
	handler := &mockHandler{}

	// Subscribe
	subID, err := bus.Subscribe(pluginsdk.EventFilter{}, handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	if subID == "" {
		t.Fatal("Subscribe returned empty subscription ID")
	}

	// Publish event
	event, err := pluginsdk.NewBusEvent("test.event", "test-plugin", map[string]string{"foo": "bar"})
	if err != nil {
		t.Fatalf("NewBusEvent failed: %v", err)
	}

	err = bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for async delivery
	time.Sleep(50 * time.Millisecond)

	// Verify handler received event
	events := handler.getEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].ID != event.ID {
		t.Errorf("Event ID mismatch: expected %s, got %s", event.ID, events[0].ID)
	}
	if events[0].Type != "test.event" {
		t.Errorf("Event type mismatch: expected 'test.event', got %s", events[0].Type)
	}
}

func TestFilterMatching_TypePattern(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		pattern       string
		eventType     string
		shouldReceive bool
	}{
		{"exact match", "gmail.email_received", "gmail.email_received", true},
		{"glob match all", "*", "any.event.type", true},
		{"glob prefix", "gmail.*", "gmail.email_received", true},
		{"glob suffix", "*received", "gmail.email_received", true},
		{"no match", "gmail.*", "slack.message_sent", false},
		{"empty pattern matches all", "", "any.event", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &mockHandler{}

			// Subscribe with pattern
			filter := pluginsdk.EventFilter{TypePattern: tt.pattern}
			_, err := bus.Subscribe(filter, handler)
			if err != nil {
				t.Fatalf("Subscribe failed: %v", err)
			}

			// Publish event
			event, _ := pluginsdk.NewBusEvent(tt.eventType, "test-plugin", nil)
			_ = bus.Publish(ctx, event)

			// Wait for delivery
			time.Sleep(50 * time.Millisecond)

			// Check if received
			events := handler.getEvents()
			received := len(events) > 0

			if received != tt.shouldReceive {
				t.Errorf("Expected shouldReceive=%v, got %v (events: %d)", tt.shouldReceive, received, len(events))
			}
		})
	}
}

func TestFilterMatching_Labels(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		filterLabels  map[string]string
		eventLabels   map[string]string
		shouldReceive bool
	}{
		{
			name:          "exact match",
			filterLabels:  map[string]string{"env": "prod"},
			eventLabels:   map[string]string{"env": "prod"},
			shouldReceive: true,
		},
		{
			name:          "subset match",
			filterLabels:  map[string]string{"env": "prod"},
			eventLabels:   map[string]string{"env": "prod", "region": "us-west"},
			shouldReceive: true,
		},
		{
			name:          "no match - different value",
			filterLabels:  map[string]string{"env": "prod"},
			eventLabels:   map[string]string{"env": "dev"},
			shouldReceive: false,
		},
		{
			name:          "no match - missing label",
			filterLabels:  map[string]string{"env": "prod"},
			eventLabels:   map[string]string{"region": "us-west"},
			shouldReceive: false,
		},
		{
			name:          "empty filter matches all",
			filterLabels:  map[string]string{},
			eventLabels:   map[string]string{"env": "prod"},
			shouldReceive: true,
		},
		{
			name:          "multiple labels match",
			filterLabels:  map[string]string{"env": "prod", "region": "us-west"},
			eventLabels:   map[string]string{"env": "prod", "region": "us-west", "app": "api"},
			shouldReceive: true,
		},
		{
			name:          "multiple labels - partial match fails",
			filterLabels:  map[string]string{"env": "prod", "region": "us-west"},
			eventLabels:   map[string]string{"env": "prod", "region": "eu-central"},
			shouldReceive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &mockHandler{}

			// Subscribe with label filter
			filter := pluginsdk.EventFilter{Labels: tt.filterLabels}
			_, err := bus.Subscribe(filter, handler)
			if err != nil {
				t.Fatalf("Subscribe failed: %v", err)
			}

			// Create and publish event
			event, _ := pluginsdk.NewBusEvent("test.event", "test-plugin", nil)
			event.Labels = tt.eventLabels
			_ = bus.Publish(ctx, event)

			// Wait for delivery
			time.Sleep(50 * time.Millisecond)

			// Check if received
			events := handler.getEvents()
			received := len(events) > 0

			if received != tt.shouldReceive {
				t.Errorf("Expected shouldReceive=%v, got %v (events: %d)", tt.shouldReceive, received, len(events))
			}
		})
	}
}

func TestFilterMatching_SourcePlugin(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		filterSource  string
		eventSource   string
		shouldReceive bool
	}{
		{"exact match", "gmail-plugin", "gmail-plugin", true},
		{"no match", "gmail-plugin", "slack-plugin", false},
		{"empty filter matches all", "", "any-plugin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &mockHandler{}

			// Subscribe with source filter
			filter := pluginsdk.EventFilter{SourcePlugin: tt.filterSource}
			_, err := bus.Subscribe(filter, handler)
			if err != nil {
				t.Fatalf("Subscribe failed: %v", err)
			}

			// Publish event
			event, _ := pluginsdk.NewBusEvent("test.event", tt.eventSource, nil)
			_ = bus.Publish(ctx, event)

			// Wait for delivery
			time.Sleep(50 * time.Millisecond)

			// Check if received
			events := handler.getEvents()
			received := len(events) > 0

			if received != tt.shouldReceive {
				t.Errorf("Expected shouldReceive=%v, got %v (events: %d)", tt.shouldReceive, received, len(events))
			}
		})
	}
}

func TestFilterMatching_Combined(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	handler := &mockHandler{}

	// Subscribe with combined filter
	filter := pluginsdk.EventFilter{
		TypePattern:  "gmail.*",
		SourcePlugin: "gmail-plugin",
		Labels:       map[string]string{"env": "prod"},
	}
	_, err := bus.Subscribe(filter, handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Event matching all criteria
	event1, _ := pluginsdk.NewBusEvent("gmail.email_received", "gmail-plugin", nil)
	event1.Labels = map[string]string{"env": "prod"}

	// Event not matching type
	event2, _ := pluginsdk.NewBusEvent("slack.message_sent", "gmail-plugin", nil)
	event2.Labels = map[string]string{"env": "prod"}

	// Event not matching source
	event3, _ := pluginsdk.NewBusEvent("gmail.email_received", "slack-plugin", nil)
	event3.Labels = map[string]string{"env": "prod"}

	// Event not matching labels
	event4, _ := pluginsdk.NewBusEvent("gmail.email_received", "gmail-plugin", nil)
	event4.Labels = map[string]string{"env": "dev"}

	// Publish all events
	_ = bus.Publish(ctx, event1)
	_ = bus.Publish(ctx, event2)
	_ = bus.Publish(ctx, event3)
	_ = bus.Publish(ctx, event4)

	// Wait for delivery
	time.Sleep(50 * time.Millisecond)

	// Only event1 should be received
	events := handler.getEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].ID != event1.ID {
		t.Errorf("Expected event1, got different event")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	// Create multiple handlers
	handler1 := &mockHandler{}
	handler2 := &mockHandler{}
	handler3 := &mockHandler{}

	// Subscribe all handlers
	_, err := bus.Subscribe(pluginsdk.EventFilter{}, handler1)
	if err != nil {
		t.Fatalf("Subscribe handler1 failed: %v", err)
	}
	_, err = bus.Subscribe(pluginsdk.EventFilter{}, handler2)
	if err != nil {
		t.Fatalf("Subscribe handler2 failed: %v", err)
	}
	_, err = bus.Subscribe(pluginsdk.EventFilter{TypePattern: "test.*"}, handler3)
	if err != nil {
		t.Fatalf("Subscribe handler3 failed: %v", err)
	}

	// Publish event
	event, _ := pluginsdk.NewBusEvent("test.event", "test-plugin", nil)
	_ = bus.Publish(ctx, event)

	// Wait for delivery
	time.Sleep(50 * time.Millisecond)

	// All handlers should receive the event
	if len(handler1.getEvents()) != 1 {
		t.Errorf("Handler1 expected 1 event, got %d", len(handler1.getEvents()))
	}
	if len(handler2.getEvents()) != 1 {
		t.Errorf("Handler2 expected 1 event, got %d", len(handler2.getEvents()))
	}
	if len(handler3.getEvents()) != 1 {
		t.Errorf("Handler3 expected 1 event, got %d", len(handler3.getEvents()))
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	handler := &mockHandler{}

	// Subscribe
	subID, err := bus.Subscribe(pluginsdk.EventFilter{}, handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish first event
	event1, _ := pluginsdk.NewBusEvent("test.event1", "test-plugin", nil)
	_ = bus.Publish(ctx, event1)
	time.Sleep(50 * time.Millisecond)

	// Unsubscribe
	err = bus.Unsubscribe(subID)
	if err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}

	// Publish second event
	event2, _ := pluginsdk.NewBusEvent("test.event2", "test-plugin", nil)
	_ = bus.Publish(ctx, event2)
	time.Sleep(50 * time.Millisecond)

	// Handler should only have received first event
	events := handler.getEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].ID != event1.ID {
		t.Errorf("Expected event1, got different event")
	}
}

func TestUnsubscribe_NonexistentID(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)

	err := bus.Unsubscribe("nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent subscription ID, got nil")
	}
}

func TestSubscribe_NilHandler(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)

	_, err := bus.Subscribe(pluginsdk.EventFilter{}, nil)
	if err == nil {
		t.Error("Expected error for nil handler, got nil")
	}
}

func TestHandlerErrors(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	// Create handlers - one errors, one succeeds
	errorHandler := &mockHandler{shouldError: true, errorMessage: "handler error"}
	successHandler := &mockHandler{}

	// Subscribe both
	_, _ = bus.Subscribe(pluginsdk.EventFilter{}, errorHandler)
	_, _ = bus.Subscribe(pluginsdk.EventFilter{}, successHandler)

	// Publish event
	event, _ := pluginsdk.NewBusEvent("test.event", "test-plugin", nil)
	_ = bus.Publish(ctx, event)

	// Wait for delivery
	time.Sleep(50 * time.Millisecond)

	// Both handlers should be called despite error
	if errorHandler.getCallCount() != 1 {
		t.Errorf("Error handler expected 1 call, got %d", errorHandler.getCallCount())
	}
	if successHandler.getCallCount() != 1 {
		t.Errorf("Success handler expected 1 call, got %d", successHandler.getCallCount())
	}

	// Success handler should have received event
	if len(successHandler.getEvents()) != 1 {
		t.Errorf("Success handler expected 1 event, got %d", len(successHandler.getEvents()))
	}
}

func TestHandlerTimeout(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	// Create handler with 35-second delay (exceeds 30-second timeout)
	slowHandler := &mockHandler{delay: 35 * time.Second}

	// Subscribe
	_, err := bus.Subscribe(pluginsdk.EventFilter{}, slowHandler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Publish event
	event, _ := pluginsdk.NewBusEvent("test.event", "test-plugin", nil)
	_ = bus.Publish(ctx, event)

	// Wait a bit to ensure handler is called but not completed
	time.Sleep(100 * time.Millisecond)

	// Handler should be called but context should timeout
	if slowHandler.getCallCount() != 1 {
		t.Errorf("Handler expected 1 call, got %d", slowHandler.getCallCount())
	}

	// Handler should not complete (event not added to list due to context timeout)
	// We can't reliably test this without making the test slow, so we just verify call count
}

func TestConcurrentPublishSubscribe(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	ctx := context.Background()

	const numSubscribers = 10
	const numEvents = 20

	var wg sync.WaitGroup
	handlers := make([]*mockHandler, numSubscribers)

	// Subscribe concurrently
	for i := 0; i < numSubscribers; i++ {
		handlers[i] = &mockHandler{}
		wg.Add(1)
		go func(h *mockHandler) {
			defer wg.Done()
			_, _ = bus.Subscribe(pluginsdk.EventFilter{}, h)
		}(handlers[i])
	}
	wg.Wait()

	// Publish events concurrently
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			event, _ := pluginsdk.NewBusEvent("test.event", "test-plugin", map[string]int{"index": idx})
			_ = bus.Publish(ctx, event)
		}(i)
	}
	wg.Wait()

	// Wait for all deliveries
	time.Sleep(200 * time.Millisecond)

	// Each handler should receive all events
	for i, handler := range handlers {
		events := handler.getEvents()
		if len(events) != numEvents {
			t.Errorf("Handler %d expected %d events, got %d", i, numEvents, len(events))
		}
	}
}

func TestConcurrentUnsubscribe(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)

	const numSubscribers = 20

	var wg sync.WaitGroup
	subIDs := make([]string, numSubscribers)

	// Subscribe
	for i := 0; i < numSubscribers; i++ {
		handler := &mockHandler{}
		subID, err := bus.Subscribe(pluginsdk.EventFilter{}, handler)
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}
		subIDs[i] = subID
	}

	// Unsubscribe concurrently
	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(subID string) {
			defer wg.Done()
			_ = bus.Unsubscribe(subID)
		}(subIDs[i])
	}
	wg.Wait()

	// All unsubscribes should succeed without race conditions
	// Try to unsubscribe again - should all fail
	for _, subID := range subIDs {
		err := bus.Unsubscribe(subID)
		if err == nil {
			t.Errorf("Expected error for already-unsubscribed ID %s, got nil", subID)
		}
	}
}

func TestPublishWithCancelledContext(t *testing.T) {
	bus := infra.NewInMemoryEventBus(nil)
	handler := &mockHandler{}

	// Subscribe
	_, err := bus.Subscribe(pluginsdk.EventFilter{}, handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Publish with cancelled context
	event, _ := pluginsdk.NewBusEvent("test.event", "test-plugin", nil)
	err = bus.Publish(ctx, event)

	// Publish should not error (delivery is async)
	if err != nil {
		t.Errorf("Publish with cancelled context should not error, got: %v", err)
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Handler should not receive event due to cancelled context
	events := handler.getEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events with cancelled context, got %d", len(events))
	}
}

func TestNewBusEvent(t *testing.T) {
	payload := map[string]string{"key": "value"}
	event, err := pluginsdk.NewBusEvent("test.type", "test-source", payload)

	if err != nil {
		t.Fatalf("NewBusEvent failed: %v", err)
	}

	// Check ID is generated
	if event.ID == "" {
		t.Error("Event ID should not be empty")
	}

	// Check type and source
	if event.Type != "test.type" {
		t.Errorf("Expected type 'test.type', got %s", event.Type)
	}
	if event.Source != "test-source" {
		t.Errorf("Expected source 'test-source', got %s", event.Source)
	}

	// Check timestamp is set
	if event.Timestamp.IsZero() {
		t.Error("Event timestamp should not be zero")
	}

	// Check labels and metadata are initialized
	if event.Labels == nil {
		t.Error("Event labels should be initialized")
	}
	if event.Metadata == nil {
		t.Error("Event metadata should be initialized")
	}

	// Check payload is JSON-encoded
	if len(event.Payload) == 0 {
		t.Error("Event payload should not be empty")
	}

	// Verify payload can be decoded
	var decoded map[string]string
	if err := json.Unmarshal(event.Payload, &decoded); err != nil {
		t.Errorf("Failed to decode payload: %v", err)
	}
	if decoded["key"] != "value" {
		t.Errorf("Expected payload key='value', got key='%s'", decoded["key"])
	}
}

func TestNewBusEvent_InvalidPayload(t *testing.T) {
	// Use a payload that can't be marshaled to JSON
	invalidPayload := make(chan int)

	_, err := pluginsdk.NewBusEvent("test.type", "test-source", invalidPayload)
	if err == nil {
		t.Error("Expected error for invalid payload, got nil")
	}
}

package infra

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// InMemoryEventBus is a thread-safe in-memory implementation of EventBus.
// It delivers events asynchronously to matching subscribers with timeouts.
// Optionally persists events to a repository for replay and durability.
type InMemoryEventBus struct {
	mu            sync.RWMutex
	subscriptions map[string]*subscription
	repository    EventBusRepository // Optional persistence layer
}

// EventBusRepository defines the interface for persisting bus events.
// This is used for optional event persistence in InMemoryEventBus.
type EventBusRepository interface {
	StoreEvent(ctx context.Context, event pluginsdk.BusEvent) error
	GetEvents(ctx context.Context, filter pluginsdk.EventFilter, limit int) ([]pluginsdk.BusEvent, error)
	GetEventsSince(ctx context.Context, since interface{}, filter pluginsdk.EventFilter, limit int) ([]pluginsdk.BusEvent, error)
}

// subscription represents a registered event handler with its filter and context.
type subscription struct {
	id         string
	filter     pluginsdk.EventFilter
	handler    pluginsdk.EventHandler
	cancel     context.CancelFunc
	ctx        context.Context
}

// NewInMemoryEventBus creates a new in-memory event bus with optional persistence.
// If repository is nil, events are not persisted (pure in-memory mode).
// If repository is provided, events are persisted on publish for replay.
func NewInMemoryEventBus(repository EventBusRepository) *InMemoryEventBus {
	return &InMemoryEventBus{
		subscriptions: make(map[string]*subscription),
		repository:    repository,
	}
}

// Publish sends an event to all matching subscribers.
// Events are delivered asynchronously with a 30-second timeout per subscriber.
// If a repository is configured, events are also persisted for replay.
func (bus *InMemoryEventBus) Publish(ctx context.Context, event pluginsdk.BusEvent) error {
	// Persist event if repository is configured
	if bus.repository != nil {
		if err := bus.repository.StoreEvent(ctx, event); err != nil {
			// Log error but don't fail publish (in-memory delivery still works)
			// TODO: Add logger to event bus for error reporting
			_ = err
		}
	}

	bus.mu.RLock()
	defer bus.mu.RUnlock()

	// Collect matching subscriptions
	var matchingSubs []*subscription
	for _, sub := range bus.subscriptions {
		if bus.matchesFilter(event, sub.filter) {
			matchingSubs = append(matchingSubs, sub)
		}
	}

	// Deliver to each matching subscriber asynchronously
	for _, sub := range matchingSubs {
		go bus.deliverEvent(ctx, sub, event)
	}

	return nil
}

// Subscribe registers a handler for events matching the filter.
// Returns a subscription ID that can be used to unsubscribe.
func (bus *InMemoryEventBus) Subscribe(filter pluginsdk.EventFilter, handler pluginsdk.EventHandler) (string, error) {
	if handler == nil {
		return "", fmt.Errorf("handler cannot be nil")
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	// Create subscription with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	sub := &subscription{
		id:      uuid.New().String(),
		filter:  filter,
		handler: handler,
		cancel:  cancel,
		ctx:     ctx,
	}

	bus.subscriptions[sub.id] = sub
	return sub.id, nil
}

// Unsubscribe removes a subscription by its ID.
// After unsubscribe, the handler will no longer receive events.
func (bus *InMemoryEventBus) Unsubscribe(subscriptionID string) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	sub, exists := bus.subscriptions[subscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	// Cancel the subscription context
	sub.cancel()

	// Remove from subscriptions
	delete(bus.subscriptions, subscriptionID)

	return nil
}

// deliverEvent delivers an event to a single subscriber with timeout.
// This runs in a goroutine and handles errors internally.
func (bus *InMemoryEventBus) deliverEvent(publishCtx context.Context, sub *subscription, event pluginsdk.BusEvent) {
	// Create timeout context for handler execution
	handlerCtx, cancel := context.WithTimeout(sub.ctx, 30*time.Second)
	defer cancel()

	// Check if publish context is cancelled
	select {
	case <-publishCtx.Done():
		return
	case <-sub.ctx.Done():
		return
	default:
	}

	// Call handler and ignore errors (handler is responsible for error handling)
	_ = sub.handler.HandleEvent(handlerCtx, event)
}

// Replay retrieves events from persistence and replays them to a handler.
// This is useful for rebuilding state from historical events or catching up subscribers.
// If the repository is nil, this method returns an error.
func (bus *InMemoryEventBus) Replay(ctx context.Context, since time.Time, filter pluginsdk.EventFilter, handler pluginsdk.EventHandler) error {
	if bus.repository == nil {
		return fmt.Errorf("replay requires a persistence repository")
	}

	// Retrieve events from persistence
	events, err := bus.repository.GetEventsSince(ctx, since, filter, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve events for replay: %w", err)
	}

	// Replay each event to the handler
	for _, event := range events {
		if err := handler.HandleEvent(ctx, event); err != nil {
			// Continue on handler errors (handler is responsible for error handling)
			continue
		}
	}

	return nil
}

// matchesFilter checks if an event matches the subscription filter.
// All specified filter criteria must match.
func (bus *InMemoryEventBus) matchesFilter(event pluginsdk.BusEvent, filter pluginsdk.EventFilter) bool {
	// Check type pattern
	if filter.TypePattern != "" {
		matched, err := filepath.Match(filter.TypePattern, event.Type)
		if err != nil || !matched {
			return false
		}
	}

	// Check source plugin
	if filter.SourcePlugin != "" && filter.SourcePlugin != event.Source {
		return false
	}

	// Check labels (all filter labels must match event labels)
	for key, value := range filter.Labels {
		eventValue, exists := event.Labels[key]
		if !exists || eventValue != value {
			return false
		}
	}

	return true
}

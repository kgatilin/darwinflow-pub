package pluginsdk

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventBus enables cross-plugin communication through publish/subscribe pattern.
// Plugins can publish events and subscribe to events from other plugins with filtering.
type EventBus interface {
	// Publish sends an event to all matching subscribers.
	// Events are delivered asynchronously with timeouts.
	Publish(ctx context.Context, event BusEvent) error

	// Subscribe registers a handler for events matching the filter.
	// Returns a subscription ID that can be used to unsubscribe.
	Subscribe(filter EventFilter, handler EventHandler) (string, error)

	// Unsubscribe removes a subscription by its ID.
	// After unsubscribe, the handler will no longer receive events.
	Unsubscribe(subscriptionID string) error
}

// BusEvent represents an event on the event bus.
// Events carry metadata and a JSON-encoded payload.
type BusEvent struct {
	// ID is a unique identifier for this event (UUID).
	ID string `json:"id"`

	// Type is the event type (e.g., "gmail.email_received").
	// Use dot notation for namespacing: "<plugin>.<event_name>".
	Type string `json:"type"`

	// Source is the plugin ID that emitted this event.
	Source string `json:"source"`

	// Timestamp is when the event was created.
	Timestamp time.Time `json:"timestamp"`

	// Labels are filterable key-value pairs for routing.
	// Subscribers can filter events by label matching.
	Labels map[string]string `json:"labels,omitempty"`

	// Metadata contains additional event metadata.
	// This is not used for filtering but provides context.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Payload is the JSON-encoded event data.
	// Subscribers decode this based on the event type.
	Payload []byte `json:"payload"`
}

// EventFilter defines criteria for selecting events.
// All specified criteria must match for an event to be delivered.
type EventFilter struct {
	// TypePattern is a glob pattern or exact match for event types.
	// Examples: "gmail.*", "slack.message_received", "*"
	TypePattern string

	// Labels specifies required label key-value pairs.
	// All specified labels must match (subset matching).
	// If empty, all label combinations match.
	Labels map[string]string

	// SourcePlugin filters events by source plugin ID.
	// If empty, events from all plugins match.
	SourcePlugin string
}

// EventHandler processes events.
// Handlers should be thread-safe as they may be called concurrently.
type EventHandler interface {
	// HandleEvent processes a single event.
	// Implementations should handle errors gracefully.
	// The context may be cancelled if handler execution exceeds timeout.
	HandleEvent(ctx context.Context, event BusEvent) error
}

// NewBusEvent creates a new bus event with JSON-encoded payload.
// The payload is marshaled to JSON and stored in the event.
// ID and timestamp are automatically generated.
func NewBusEvent(eventType, source string, payload interface{}) (BusEvent, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return BusEvent{}, err
	}

	return BusEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now(),
		Labels:    make(map[string]string),
		Metadata:  make(map[string]interface{}),
		Payload:   payloadBytes,
	}, nil
}

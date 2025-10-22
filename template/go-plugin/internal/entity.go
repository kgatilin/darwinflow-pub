package internal

import "time"

// Item represents a generic item entity.
// This is a simple example entity that can be customized for your specific use case.
type Item struct {
	// ID is the unique identifier for this item
	ID string

	// Name is the display name of the item
	Name string

	// Description provides additional details about the item
	Description string

	// Tags are labels for categorizing and filtering items
	Tags []string

	// CreatedAt is when this item was created
	CreatedAt time.Time

	// UpdatedAt is when this item was last modified
	UpdatedAt time.Time
}

// ToMap converts an Item to a map for JSON serialization.
// This format is used for entity queries and responses in the RPC protocol.
// The "type" field identifies the entity type, and "capabilities" lists
// optional features this entity supports (e.g., "trackable", "schedulable").
func (i *Item) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          i.ID,
		"type":        "item",
		"name":        i.Name,
		"description": i.Description,
		"tags":        i.Tags,
		"created_at":  i.CreatedAt.Format(time.RFC3339),
		"updated_at":  i.UpdatedAt.Format(time.RFC3339),
		"capabilities": []string{}, // Add capabilities like "trackable" or "schedulable" if implemented
	}
}

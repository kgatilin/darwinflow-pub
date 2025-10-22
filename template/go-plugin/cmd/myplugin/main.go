package main

import (
	"time"

	"example.com/myplugin/internal"
)

// main is the entry point for the plugin.
// It creates the plugin instance, initializes sample data, and starts the RPC server.
func main() {
	// Create a new plugin instance
	plugin := internal.NewItemPlugin()

	// Create sample items for demonstration
	// In a real plugin, you might load data from a database or file system
	plugin.AddItem(&internal.Item{
		ID:          "item-1",
		Name:        "Example Item",
		Description: "This is an example item from the external plugin.",
		Tags:        []string{"example", "demo"},
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	})

	plugin.AddItem(&internal.Item{
		ID:          "item-2",
		Name:        "Another Item",
		Description: "External plugins can run in any language!",
		Tags:        []string{"external", "plugin"},
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		UpdatedAt:   time.Now().Add(-1 * time.Hour),
	})

	plugin.AddItem(&internal.Item{
		ID:          "item-3",
		Name:        "Third Item",
		Description: "Demonstrates multiple entities and querying.",
		Tags:        []string{"query", "demo"},
		CreatedAt:   time.Now().Add(-5 * time.Hour),
		UpdatedAt:   time.Now().Add(-3 * time.Hour),
	})

	// Run the JSON-RPC server loop
	// This blocks until stdin is closed
	plugin.Serve()
}

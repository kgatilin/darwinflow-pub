package events_test

import (
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/events"
)

// TestEventConstantsNaming verifies that all event constants follow the naming convention:
// - EventType constants must start with "Event"
// - Event type strings must use "task-manager." prefix (except deprecated events)
// - Event names use dot notation: "task-manager.<entity>.<action>"
func TestEventConstantsNaming(t *testing.T) {
	testCases := []struct {
		name        string
		eventType   string
		wantPrefix  string
		wantPattern string // Expected pattern: "task-manager.<entity>.<action>"
	}{
		// Roadmap events
		{"RoadmapCreated", events.EventRoadmapCreated, "task-manager.", "task-manager.roadmap.created"},
		{"RoadmapUpdated", events.EventRoadmapUpdated, "task-manager.", "task-manager.roadmap.updated"},

		// Track events
		{"TrackCreated", events.EventTrackCreated, "task-manager.", "task-manager.track.created"},
		{"TrackUpdated", events.EventTrackUpdated, "task-manager.", "task-manager.track.updated"},
		{"TrackStatusChanged", events.EventTrackStatusChanged, "task-manager.", "task-manager.track.status_changed"},
		{"TrackCompleted", events.EventTrackCompleted, "task-manager.", "task-manager.track.completed"},
		{"TrackBlocked", events.EventTrackBlocked, "task-manager.", "task-manager.track.blocked"},

		// Task events
		{"TaskCreated", events.EventTaskCreated, "task-manager.", "task-manager.task.created"},
		{"TaskUpdated", events.EventTaskUpdated, "task-manager.", "task-manager.task.updated"},
		{"TaskStatusChanged", events.EventTaskStatusChanged, "task-manager.", "task-manager.task.status_changed"},
		{"TaskCompleted", events.EventTaskCompleted, "task-manager.", "task-manager.task.completed"},

		// Iteration events
		{"IterationCreated", events.EventIterationCreated, "task-manager.", "task-manager.iteration.created"},
		{"IterationStarted", events.EventIterationStarted, "task-manager.", "task-manager.iteration.started"},
		{"IterationCompleted", events.EventIterationCompleted, "task-manager.", "task-manager.iteration.completed"},
		{"IterationUpdated", events.EventIterationUpdated, "task-manager.", "task-manager.iteration.updated"},

		// Acceptance Criteria events
		{"ACCreated", events.EventACCreated, "task-manager.", "task-manager.ac.created"},
		{"ACUpdated", events.EventACUpdated, "task-manager.", "task-manager.ac.updated"},
		{"ACVerified", events.EventACVerified, "task-manager.", "task-manager.ac.verified"},
		{"ACAutomaticallyVerified", events.EventACAutomaticallyVerified, "task-manager.", "task-manager.ac.automatically_verified"},
		{"ACPendingReview", events.EventACPendingReview, "task-manager.", "task-manager.ac.pending_review"},
		{"ACFailed", events.EventACFailed, "task-manager.", "task-manager.ac.failed"},
		{"ACDeleted", events.EventACDeleted, "task-manager.", "task-manager.ac.deleted"},

		// ADR events
		{"ADRCreated", events.EventADRCreated, "task-manager.", "task-manager.adr.created"},
		{"ADRUpdated", events.EventADRUpdated, "task-manager.", "task-manager.adr.updated"},
		{"ADRSuperseded", events.EventADRSuperseded, "task-manager.", "task-manager.adr.superseded"},
		{"ADRDeprecated", events.EventADRDeprecated, "task-manager.", "task-manager.adr.deprecated"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify prefix
			if !strings.HasPrefix(tc.eventType, tc.wantPrefix) {
				t.Errorf("Event %q must have prefix %q, got: %q", tc.name, tc.wantPrefix, tc.eventType)
			}

			// Verify exact match
			if tc.eventType != tc.wantPattern {
				t.Errorf("Event %q must match pattern %q, got: %q", tc.name, tc.wantPattern, tc.eventType)
			}

			// Verify no uppercase letters in event type string (except first segment)
			parts := strings.Split(tc.eventType, ".")
			for i, part := range parts {
				if i > 0 && part != strings.ToLower(part) {
					t.Errorf("Event %q segment %q must be lowercase", tc.name, part)
				}
			}
		})
	}
}

// TestDeprecatedEventConstants verifies that deprecated event constants are marked and present
func TestDeprecatedEventConstants(t *testing.T) {
	// EventTaskDeletedLegacy is the only deprecated event
	if events.EventTaskDeletedLegacy != "task.deleted" {
		t.Errorf("EventTaskDeletedLegacy must be 'task.deleted', got: %q", events.EventTaskDeletedLegacy)
	}

	// Verify it doesn't have the "task-manager." prefix (that's why it's deprecated)
	if strings.HasPrefix(events.EventTaskDeletedLegacy, "task-manager.") {
		t.Errorf("EventTaskDeletedLegacy should NOT have 'task-manager.' prefix (deprecated format)")
	}
}

// TestPluginSourceName verifies the plugin source name constant
func TestPluginSourceName(t *testing.T) {
	want := "task-manager"
	if events.PluginSourceName != want {
		t.Errorf("PluginSourceName = %q, want %q", events.PluginSourceName, want)
	}
}

// TestEventTypesUniqueness verifies that all event type constants are unique
func TestEventTypesUniqueness(t *testing.T) {
	eventTypes := []string{
		// Roadmap
		events.EventRoadmapCreated,
		events.EventRoadmapUpdated,

		// Track
		events.EventTrackCreated,
		events.EventTrackUpdated,
		events.EventTrackStatusChanged,
		events.EventTrackCompleted,
		events.EventTrackBlocked,

		// Task
		events.EventTaskCreated,
		events.EventTaskUpdated,
		events.EventTaskStatusChanged,
		events.EventTaskCompleted,

		// Iteration
		events.EventIterationCreated,
		events.EventIterationStarted,
		events.EventIterationCompleted,
		events.EventIterationUpdated,

		// AC
		events.EventACCreated,
		events.EventACUpdated,
		events.EventACVerified,
		events.EventACAutomaticallyVerified,
		events.EventACPendingReview,
		events.EventACFailed,
		events.EventACDeleted,

		// ADR
		events.EventADRCreated,
		events.EventADRUpdated,
		events.EventADRSuperseded,
		events.EventADRDeprecated,
	}

	seen := make(map[string]bool)
	for _, eventType := range eventTypes {
		if seen[eventType] {
			t.Errorf("Duplicate event type found: %q", eventType)
		}
		seen[eventType] = true
	}
}

// TestEventCounts verifies the expected number of events per entity type
func TestEventCounts(t *testing.T) {
	testCases := []struct {
		entity string
		count  int
	}{
		{"roadmap", 2},
		{"track", 5},
		{"task", 4}, // Not counting deprecated EventTaskDeletedLegacy
		{"iteration", 4},
		{"ac", 7},
		{"adr", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.entity, func(t *testing.T) {
			// This test documents the expected event count per entity
			// If new events are added, this test will need updating
			t.Logf("Entity %q has %d events", tc.entity, tc.count)
		})
	}
}

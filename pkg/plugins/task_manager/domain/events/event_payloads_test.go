package events_test

import (
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/events"
)

// TestPayloadTypeAliases verifies that all payload types are correctly aliased to entity types.
// This ensures that event payloads are always consistent with the underlying domain entities.
func TestPayloadTypeAliases(t *testing.T) {
	// Roadmap payloads
	t.Run("RoadmapPayloads", func(t *testing.T) {
		var created events.RoadmapCreatedPayload
		var updated events.RoadmapUpdatedPayload

		// Verify they're assignable to RoadmapEntity
		_ = entities.RoadmapEntity(created)
		_ = entities.RoadmapEntity(updated)
	})

	// Track payloads
	t.Run("TrackPayloads", func(t *testing.T) {
		var created events.TrackCreatedPayload
		var updated events.TrackUpdatedPayload
		var statusChanged events.TrackStatusChangedPayload
		var completed events.TrackCompletedPayload
		var blocked events.TrackBlockedPayload

		// Verify they're assignable to TrackEntity
		_ = entities.TrackEntity(created)
		_ = entities.TrackEntity(updated)
		_ = entities.TrackEntity(statusChanged)
		_ = entities.TrackEntity(completed)
		_ = entities.TrackEntity(blocked)
	})

	// Task payloads
	t.Run("TaskPayloads", func(t *testing.T) {
		var created events.TaskCreatedPayload
		var updated events.TaskUpdatedPayload
		var statusChanged events.TaskStatusChangedPayload
		var completed events.TaskCompletedPayload

		// Verify they're assignable to TaskEntity
		_ = entities.TaskEntity(created)
		_ = entities.TaskEntity(updated)
		_ = entities.TaskEntity(statusChanged)
		_ = entities.TaskEntity(completed)
	})

	// Iteration payloads
	t.Run("IterationPayloads", func(t *testing.T) {
		var created events.IterationCreatedPayload
		var started events.IterationStartedPayload
		var completed events.IterationCompletedPayload
		var updated events.IterationUpdatedPayload

		// Verify they're assignable to IterationEntity
		_ = entities.IterationEntity(created)
		_ = entities.IterationEntity(started)
		_ = entities.IterationEntity(completed)
		_ = entities.IterationEntity(updated)
	})

	// AC payloads
	t.Run("ACPayloads", func(t *testing.T) {
		var created events.ACCreatedPayload
		var updated events.ACUpdatedPayload
		var verified events.ACVerifiedPayload
		var autoVerified events.ACAutomaticallyVerifiedPayload
		var pendingReview events.ACPendingReviewPayload
		var failed events.ACFailedPayload
		var deleted events.ACDeletedPayload

		// Verify they're assignable to AcceptanceCriteriaEntity
		_ = entities.AcceptanceCriteriaEntity(created)
		_ = entities.AcceptanceCriteriaEntity(updated)
		_ = entities.AcceptanceCriteriaEntity(verified)
		_ = entities.AcceptanceCriteriaEntity(autoVerified)
		_ = entities.AcceptanceCriteriaEntity(pendingReview)
		_ = entities.AcceptanceCriteriaEntity(failed)
		_ = entities.AcceptanceCriteriaEntity(deleted)
	})

	// ADR payloads
	t.Run("ADRPayloads", func(t *testing.T) {
		var created events.ADRCreatedPayload
		var updated events.ADRUpdatedPayload
		var superseded events.ADRSupersededPayload
		var deprecated events.ADRDeprecatedPayload

		// Verify they're assignable to ADREntity
		_ = entities.ADREntity(created)
		_ = entities.ADREntity(updated)
		_ = entities.ADREntity(superseded)
		_ = entities.ADREntity(deprecated)
	})
}

// TestPayloadUsageExample demonstrates how event payloads are used with actual entities
func TestPayloadUsageExample(t *testing.T) {
	// Create a sample task entity
	task := entities.TaskEntity{
		ID:          "TM-task-1",
		TrackID:     "TM-track-1",
		Title:       "Sample Task",
		Description: "A sample task for testing",
		Status:      "todo", // String type, not TaskStatus constant
	}

	// Use it as a payload
	var payload events.TaskCreatedPayload = task

	// Verify we can access entity fields through the payload
	if payload.ID != "TM-task-1" {
		t.Errorf("Expected payload.ID = 'TM-task-1', got %q", payload.ID)
	}

	if payload.Status != "todo" {
		t.Errorf("Expected payload.Status = 'todo', got %v", payload.Status)
	}
}

// TestPayloadConstruction demonstrates constructing payloads for all event types
func TestPayloadConstruction(t *testing.T) {
	t.Run("RoadmapPayload", func(t *testing.T) {
		roadmap := entities.RoadmapEntity{
			ID:              "TM-roadmap-1",
			Vision:          "Build great software",
			SuccessCriteria: "Quality, Speed", // String, not []string
		}
		var _ events.RoadmapCreatedPayload = roadmap
		var _ events.RoadmapUpdatedPayload = roadmap
	})

	t.Run("TrackPayload", func(t *testing.T) {
		track := entities.TrackEntity{
			ID:          "TM-track-1",
			RoadmapID:   "TM-roadmap-1",
			Title:       "Feature Track",
			Description: "A feature track",
			Status:      "not-started", // String type
			Rank:        100,
		}
		var _ events.TrackCreatedPayload = track
		var _ events.TrackUpdatedPayload = track
		var _ events.TrackStatusChangedPayload = track
		var _ events.TrackCompletedPayload = track
		var _ events.TrackBlockedPayload = track
	})

	t.Run("TaskPayload", func(t *testing.T) {
		task := entities.TaskEntity{
			ID:          "TM-task-1",
			TrackID:     "TM-track-1",
			Title:       "Implement feature",
			Description: "Implement the feature",
			Status:      "todo", // String type
			Rank:        100,
		}
		var _ events.TaskCreatedPayload = task
		var _ events.TaskUpdatedPayload = task
		var _ events.TaskStatusChangedPayload = task
		var _ events.TaskCompletedPayload = task
	})

	t.Run("IterationPayload", func(t *testing.T) {
		iteration := entities.IterationEntity{
			Number:      1,
			Name:        "Sprint 1",
			Goal:        "Complete MVP",
			Deliverable: "Working prototype",
			Status:      "planned", // String type
		}
		var _ events.IterationCreatedPayload = iteration
		var _ events.IterationStartedPayload = iteration
		var _ events.IterationCompletedPayload = iteration
		var _ events.IterationUpdatedPayload = iteration
	})

	t.Run("ACPayload", func(t *testing.T) {
		ac := entities.AcceptanceCriteriaEntity{
			ID:                  "TM-ac-1",
			TaskID:              "TM-task-1",
			Description:         "Feature works correctly",
			TestingInstructions: "Run test suite",
			Status:              "not-started", // String type
		}
		var _ events.ACCreatedPayload = ac
		var _ events.ACUpdatedPayload = ac
		var _ events.ACVerifiedPayload = ac
		var _ events.ACAutomaticallyVerifiedPayload = ac
		var _ events.ACPendingReviewPayload = ac
		var _ events.ACFailedPayload = ac
		var _ events.ACDeletedPayload = ac
	})

	t.Run("ADRPayload", func(t *testing.T) {
		adr := entities.ADREntity{
			ID:           "TM-adr-1",
			TrackID:      "TM-track-1",
			Title:        "Use SQLite",
			Context:      "Need local storage",
			Decision:     "Use SQLite for local data",
			Consequences: "Fast, Portable", // String, not []string
			Status:       "proposed",        // String type
		}
		var _ events.ADRCreatedPayload = adr
		var _ events.ADRUpdatedPayload = adr
		var _ events.ADRSupersededPayload = adr
		var _ events.ADRDeprecatedPayload = adr
	})
}

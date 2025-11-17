package dto_test

import (
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/dto"
)

// TestCreateTrackDTO verifies the CreateTrackDTO structure
func TestCreateTrackDTO(t *testing.T) {
	createDTO := dto.CreateTrackDTO{
		RoadmapID:   "TM-roadmap-1",
		Title:       "Feature Track",
		Description: "A feature track",
		Status:      "not-started",
		Rank:        100,
	}

	// Verify required fields are set
	if createDTO.RoadmapID == "" {
		t.Error("RoadmapID should not be empty")
	}
	if createDTO.Title == "" {
		t.Error("Title should not be empty")
	}
	if createDTO.Status == "" {
		t.Error("Status should not be empty")
	}

	// Verify values match
	if createDTO.RoadmapID != "TM-roadmap-1" {
		t.Errorf("RoadmapID = %q, want 'TM-roadmap-1'", createDTO.RoadmapID)
	}
	if createDTO.Rank != 100 {
		t.Errorf("Rank = %d, want 100", createDTO.Rank)
	}
}

// TestUpdateTrackDTO verifies the UpdateTrackDTO structure with optional fields
func TestUpdateTrackDTO(t *testing.T) {
	t.Run("all fields set", func(t *testing.T) {
		updateDTO := dto.UpdateTrackDTO{
			ID:          "TM-track-1",
			Title:       dto.StringPtr("New Title"),
			Description: dto.StringPtr("New Description"),
			Status:      dto.StringPtr("in-progress"),
			Rank:        dto.IntPtr(200),
		}

		if updateDTO.ID == "" {
			t.Error("ID should not be empty")
		}
		if updateDTO.Title == nil {
			t.Error("Title should not be nil")
		}
		if updateDTO.Description == nil {
			t.Error("Description should not be nil")
		}
		if updateDTO.Status == nil {
			t.Error("Status should not be nil")
		}
		if updateDTO.Rank == nil {
			t.Error("Rank should not be nil")
		}
	})

	t.Run("partial update", func(t *testing.T) {
		updateDTO := dto.UpdateTrackDTO{
			ID:    "TM-track-1",
			Title: dto.StringPtr("Only Title Updated"),
			// Other fields are nil (no change)
		}

		if updateDTO.ID == "" {
			t.Error("ID should not be empty")
		}
		if updateDTO.Title == nil {
			t.Error("Title should be set")
		}
		if updateDTO.Description != nil {
			t.Error("Description should be nil (no update)")
		}
		if updateDTO.Status != nil {
			t.Error("Status should be nil (no update)")
		}
		if updateDTO.Rank != nil {
			t.Error("Rank should be nil (no update)")
		}
	})
}

// TestTrackListFilters verifies the TrackListFilters structure
func TestTrackListFilters(t *testing.T) {
	t.Run("with status filter", func(t *testing.T) {
		filters := dto.TrackListFilters{
			Status: []string{"in-progress", "blocked"},
		}

		if len(filters.Status) != 2 {
			t.Errorf("Status filter count = %d, want 2", len(filters.Status))
		}
	})

	t.Run("empty filters", func(t *testing.T) {
		filters := dto.TrackListFilters{}

		if filters.Status != nil {
			t.Error("Empty filters should have nil Status")
		}
	})
}

// TestCreateTaskDTO verifies the CreateTaskDTO structure
func TestCreateTaskDTO(t *testing.T) {
	createDTO := dto.CreateTaskDTO{
		TrackID:     "TM-track-1",
		Title:       "Implement feature",
		Description: "Implement the feature",
		Status:      "todo",
		Rank:        100,
	}

	if createDTO.TrackID == "" {
		t.Error("TrackID should not be empty")
	}
	if createDTO.Title == "" {
		t.Error("Title should not be empty")
	}
	if createDTO.Status == "" {
		t.Error("Status should not be empty")
	}
}

// TestUpdateTaskDTO verifies the UpdateTaskDTO structure
func TestUpdateTaskDTO(t *testing.T) {
	t.Run("update status only", func(t *testing.T) {
		updateDTO := dto.UpdateTaskDTO{
			ID:     "TM-task-1",
			Status: dto.StringPtr("in-progress"),
		}

		if updateDTO.ID == "" {
			t.Error("ID should not be empty")
		}
		if updateDTO.Status == nil {
			t.Error("Status should be set")
		}
		if updateDTO.Title != nil {
			t.Error("Title should be nil (no update)")
		}
	})

	t.Run("move task to different track", func(t *testing.T) {
		updateDTO := dto.UpdateTaskDTO{
			ID:      "TM-task-1",
			TrackID: dto.StringPtr("TM-track-2"),
		}

		if updateDTO.TrackID == nil {
			t.Error("TrackID should be set")
		}
		if *updateDTO.TrackID != "TM-track-2" {
			t.Errorf("TrackID = %q, want 'TM-track-2'", *updateDTO.TrackID)
		}
	})
}

// TestTaskListFilters verifies the TaskListFilters structure
func TestTaskListFilters(t *testing.T) {
	t.Run("filter by status", func(t *testing.T) {
		filters := dto.TaskListFilters{
			Status: []string{"todo", "in-progress"},
		}

		if len(filters.Status) != 2 {
			t.Errorf("Status filter count = %d, want 2", len(filters.Status))
		}
	})

	t.Run("filter by track", func(t *testing.T) {
		trackID := "TM-track-1"
		filters := dto.TaskListFilters{
			TrackID: &trackID,
		}

		if filters.TrackID == nil {
			t.Error("TrackID should be set")
		}
		if *filters.TrackID != "TM-track-1" {
			t.Errorf("TrackID = %q, want 'TM-track-1'", *filters.TrackID)
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		trackID := "TM-track-1"
		filters := dto.TaskListFilters{
			Status:  []string{"todo"},
			TrackID: &trackID,
		}

		if filters.Status == nil {
			t.Error("Status should be set")
		}
		if filters.TrackID == nil {
			t.Error("TrackID should be set")
		}
	})
}

// TestDTOZeroValues verifies that DTOs handle zero values correctly
func TestDTOZeroValues(t *testing.T) {
	t.Run("CreateTrackDTO zero values", func(t *testing.T) {
		// Should be able to create DTO with zero values
		createDTO := dto.CreateTrackDTO{}

		// Zero values should be distinguishable from set values
		if createDTO.RoadmapID != "" {
			t.Error("Zero value RoadmapID should be empty string")
		}
		if createDTO.Rank != 0 {
			t.Error("Zero value Rank should be 0")
		}
	})

	t.Run("UpdateTrackDTO nil vs empty string", func(t *testing.T) {
		// nil means "no update", empty string means "update to empty"
		updateEmpty := dto.UpdateTrackDTO{
			ID:          "TM-track-1",
			Description: dto.StringPtr(""), // Update to empty string
		}

		updateNil := dto.UpdateTrackDTO{
			ID: "TM-track-1",
			// Description is nil (no update)
		}

		if updateEmpty.Description == nil {
			t.Error("Empty string update should not be nil")
		}
		if *updateEmpty.Description != "" {
			t.Error("Empty string update should be empty string")
		}

		if updateNil.Description != nil {
			t.Error("Nil update should be nil")
		}
	})
}

// TestDTOPointerSemantics verifies that pointer fields in update DTOs work correctly
func TestDTOPointerSemantics(t *testing.T) {
	t.Run("modifying pointer after DTO creation", func(t *testing.T) {
		title := "Original Title"
		updateDTO := dto.UpdateTrackDTO{
			ID:    "TM-track-1",
			Title: &title,
		}

		// Modify the original variable
		title = "Modified Title"

		// DTO should reflect the change (shared pointer)
		if *updateDTO.Title != "Modified Title" {
			t.Error("DTO should share pointer with original variable")
		}

		// Better approach: use helper to create independent pointer
		safeDTO := dto.UpdateTrackDTO{
			ID:    "TM-track-1",
			Title: dto.StringPtr("Safe Title"),
		}

		// Modifying through helper-created pointer doesn't affect original
		*safeDTO.Title = "Changed"
		// No original variable to be affected - this is the safe pattern
	})
}

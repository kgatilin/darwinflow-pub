package entities_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
)

func TestNewTaskEntity(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		id          string
		trackID     string
		title       string
		description string
		status      string
		rank        int
		branch      string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid task",
			id:          "DW-task-1",
			trackID:     "DW-track-1",
			title:       "Test Task",
			description: "Test Description",
			status:      "todo",
			rank:        500,
			branch:      "feature/test",
			wantErr:     false,
		},
		{
			name:        "valid in-progress",
			id:          "DW-task-2",
			trackID:     "DW-track-1",
			title:       "Test Task",
			description: "Test Description",
			status:      "in-progress",
			rank:        100,
			branch:      "",
			wantErr:     false,
		},
		{
			name:        "valid done",
			id:          "DW-task-3",
			trackID:     "DW-track-1",
			title:       "Test Task",
			description: "Test Description",
			status:      "done",
			rank:        1000,
			branch:      "",
			wantErr:     false,
		},
		{
			name:        "valid review",
			id:          "DW-task-4",
			trackID:     "DW-track-1",
			title:       "Test Task",
			description: "Test Description",
			status:      "review",
			rank:        500,
			branch:      "",
			wantErr:     false,
		},
		{
			name:         "invalid status",
			id:           "DW-task-1",
			trackID:      "DW-track-1",
			title:        "Test Task",
			description:  "Test Description",
			status:       "invalid-status",
			rank:         500,
			branch:       "",
			wantErr:      true,
			errContains:  "invalid task status",
		},
		{
			name:         "rank too low",
			id:           "DW-task-1",
			trackID:      "DW-track-1",
			title:        "Test Task",
			description:  "Test Description",
			status:       "todo",
			rank:         0,
			branch:       "",
			wantErr:      true,
			errContains:  "invalid task rank",
		},
		{
			name:         "rank too high",
			id:           "DW-task-1",
			trackID:      "DW-track-1",
			title:        "Test Task",
			description:  "Test Description",
			status:       "todo",
			rank:         1001,
			branch:       "",
			wantErr:      true,
			errContains:  "invalid task rank",
		},
		{
			name:         "empty title",
			id:           "DW-task-1",
			trackID:      "DW-track-1",
			title:        "",
			description:  "Test Description",
			status:       "todo",
			rank:         500,
			branch:       "",
			wantErr:      true,
			errContains:  "task title must be non-empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := entities.NewTaskEntity(
				tt.id, tt.trackID, tt.title, tt.description,
				tt.status, tt.rank, tt.branch, now, now,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if task == nil {
					t.Fatal("expected non-nil task")
				}
				if task.ID != tt.id {
					t.Errorf("ID = %q, want %q", task.ID, tt.id)
				}
				if task.Status != tt.status {
					t.Errorf("Status = %q, want %q", task.Status, tt.status)
				}
				if task.Rank != tt.rank {
					t.Errorf("Rank = %d, want %d", task.Rank, tt.rank)
				}
			}
		})
	}
}

func TestTaskEntity_TransitionTo(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  string
		toStatus    string
		wantErr     bool
		errContains string
	}{
		// Valid transitions
		{"todo to in-progress", "todo", "in-progress", false, ""},
		{"todo to done", "todo", "done", false, ""},
		{"in-progress to done", "in-progress", "done", false, ""},
		{"in-progress to review", "in-progress", "review", false, ""},
		{"in-progress to todo", "in-progress", "todo", false, ""},
		{"review to done", "review", "done", false, ""},
		{"review to in-progress", "review", "in-progress", false, ""},
		{"done to todo (reopen)", "done", "todo", false, ""},
		{"done to in-progress", "done", "in-progress", false, ""},

		// Invalid status
		{"to invalid status", "todo", "invalid-status", true, "invalid task status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &entities.TaskEntity{
				ID:          "DW-task-1",
				TrackID:     "DW-track-1",
				Title:       "Test Task",
				Description: "Test Description",
				Status:      tt.fromStatus,
				Rank:        500,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			err := task.TransitionTo(tt.toStatus)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if task.Status != tt.toStatus {
					t.Errorf("Status = %q, want %q", task.Status, tt.toStatus)
				}
			}
		})
	}
}

func TestTaskEntity_GetProgress(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected float64
	}{
		{"done", "done", 1.0},
		{"review", "review", 0.8},
		{"in-progress", "in-progress", 0.5},
		{"todo", "todo", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &entities.TaskEntity{Status: tt.status}
			progress := task.GetProgress()
			if progress != tt.expected {
				t.Errorf("GetProgress() = %v, want %v", progress, tt.expected)
			}
		})
	}
}

func TestTaskEntity_IsBlocked(t *testing.T) {
	task := &entities.TaskEntity{}
	if task.IsBlocked() {
		t.Error("expected IsBlocked() to return false for tasks")
	}
}

func TestTaskEntity_GetBlockReason(t *testing.T) {
	task := &entities.TaskEntity{}
	if task.GetBlockReason() != "" {
		t.Errorf("expected empty block reason, got %q", task.GetBlockReason())
	}
}

// SDK Interface Tests

func TestTaskEntity_GetID(t *testing.T) {
	task := &entities.TaskEntity{ID: "DW-task-42"}
	if got := task.GetID(); got != "DW-task-42" {
		t.Errorf("GetID() = %q, want %q", got, "DW-task-42")
	}
}

func TestTaskEntity_GetType(t *testing.T) {
	task := &entities.TaskEntity{}
	if got := task.GetType(); got != "task" {
		t.Errorf("GetType() = %q, want %q", got, "task")
	}
}

func TestTaskEntity_GetCapabilities(t *testing.T) {
	task := &entities.TaskEntity{}
	capabilities := task.GetCapabilities()

	expected := []string{"IExtensible", "ITrackable"}
	if len(capabilities) != len(expected) {
		t.Errorf("GetCapabilities() length = %d, want %d", len(capabilities), len(expected))
		return
	}

	for i, cap := range capabilities {
		if cap != expected[i] {
			t.Errorf("GetCapabilities()[%d] = %q, want %q", i, cap, expected[i])
		}
	}
}

func TestTaskEntity_GetField(t *testing.T) {
	task := &entities.TaskEntity{
		ID:          "DW-task-1",
		TrackID:     "DW-track-1",
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "in-progress",
		Rank:        500,
		Branch:      "feature/test",
	}

	tests := []struct {
		field    string
		expected interface{}
	}{
		{"id", "DW-task-1"},
		{"track_id", "DW-track-1"},
		{"title", "Test Task"},
		{"description", "Test Description"},
		{"status", "in-progress"},
		{"rank", 500},
		{"branch", "feature/test"},
		{"progress", 0.5},
		{"is_blocked", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := task.GetField(tt.field)
			if got != tt.expected {
				t.Errorf("GetField(%q) = %v, want %v", tt.field, got, tt.expected)
			}
		})
	}
}

func TestTaskEntity_GetAllFields(t *testing.T) {
	now := time.Now()
	task := &entities.TaskEntity{
		ID:          "DW-task-1",
		TrackID:     "DW-track-1",
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "in-progress",
		Rank:        500,
		Branch:      "feature/test",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	fields := task.GetAllFields()

	// Verify all expected fields are present
	expectedFields := []string{
		"id", "track_id", "title", "description",
		"status", "rank", "branch",
		"created_at", "updated_at", "progress", "is_blocked",
	}

	for _, field := range expectedFields {
		if _, exists := fields[field]; !exists {
			t.Errorf("GetAllFields() missing field %q", field)
		}
	}

	// Verify some key values
	if fields["id"] != "DW-task-1" {
		t.Errorf("GetAllFields()[\"id\"] = %v, want %v", fields["id"], "DW-task-1")
	}
	if fields["status"] != "in-progress" {
		t.Errorf("GetAllFields()[\"status\"] = %v, want %v", fields["status"], "in-progress")
	}
	if fields["progress"] != 0.5 {
		t.Errorf("GetAllFields()[\"progress\"] = %v, want %v", fields["progress"], 0.5)
	}
}

func TestTaskEntity_GetStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"todo", "todo"},
		{"in-progress", "in-progress"},
		{"review", "review"},
		{"done", "done"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &entities.TaskEntity{Status: tt.status}
			if got := task.GetStatus(); got != tt.status {
				t.Errorf("GetStatus() = %q, want %q", got, tt.status)
			}
		})
	}
}

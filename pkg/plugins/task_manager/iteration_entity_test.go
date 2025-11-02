package task_manager_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// TestNewIterationEntity tests iteration entity creation
func TestNewIterationEntity(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		number      int
		status      string
		expectedErr bool
	}{
		{
			name:        "valid iteration",
			number:      1,
			status:      "planned",
			expectedErr: false,
		},
		{
			name:        "zero number",
			number:      0,
			status:      "planned",
			expectedErr: true,
		},
		{
			name:        "negative number",
			number:      -1,
			status:      "planned",
			expectedErr: true,
		},
		{
			name:        "invalid status",
			number:      1,
			status:      "unknown",
			expectedErr: true,
		},
		{
			name:        "current status",
			number:      2,
			status:      "current",
			expectedErr: false,
		},
		{
			name:        "complete status",
			number:      3,
			status:      "complete",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iteration, err := task_manager.NewIterationEntity(
				tt.number,
				"Test Iteration",
				"Test Goal",
				"Test Deliverable",
				[]string{},
				tt.status,
				500,
				time.Time{},
				time.Time{},
				now,
				now,
			)

			if tt.expectedErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedErr && iteration == nil {
				t.Error("expected iteration, got nil")
			}
		})
	}
}

// TestIterationEntityGetters tests iteration entity field getters
func TestIterationEntityGetters(t *testing.T) {
	now := time.Now().UTC()
	number := 1

	iteration, err := task_manager.NewIterationEntity(
		number,
		"Foundation Sprint",
		"Complete core framework",
		"Working framework",
		[]string{},
		"planned",
		500,
		time.Time{},
		time.Time{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}

	if iteration.GetID() != "1" {
		t.Errorf("expected ID '1', got %q", iteration.GetID())
	}

	if iteration.GetType() != "iteration" {
		t.Errorf("expected type 'iteration', got %q", iteration.GetType())
	}

	caps := iteration.GetCapabilities()
	if len(caps) != 1 || caps[0] != "IExtensible" {
		t.Errorf("expected capabilities ['IExtensible'], got %v", caps)
	}
}

// TestIterationEntityFields tests GetAllFields method
func TestIterationEntityFields(t *testing.T) {
	now := time.Now().UTC()

	iteration, err := task_manager.NewIterationEntity(
		1,
		"Foundation Sprint",
		"Complete core framework",
		"Working framework",
		[]string{"task-1", "task-2"},
		"planned",
		500,
		time.Time{},
		time.Time{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}

	fields := iteration.GetAllFields()

	expectedFields := []string{
		"number", "name", "goal", "task_ids",
		"status", "deliverable", "created_at", "updated_at",
		"started_at", "completed_at",
	}

	for _, field := range expectedFields {
		if fields[field] == nil && field != "started_at" && field != "completed_at" {
			t.Errorf("field %q is nil", field)
		}
	}

	// Test GetField
	if iteration.GetField("number") != 1 {
		t.Error("GetField('number') mismatch")
	}

	if iteration.GetField("name") != "Foundation Sprint" {
		t.Error("GetField('name') mismatch")
	}

	if iteration.GetField("status") != "planned" {
		t.Error("GetField('status') mismatch")
	}
}

// TestIterationEntityTaskManagement tests task add/remove functionality
func TestIterationEntityTaskManagement(t *testing.T) {
	now := time.Now().UTC()

	iteration, err := task_manager.NewIterationEntity(
		1,
		"Test Iteration",
		"Test Goal",
		"Test Deliverable",
		[]string{},
		"planned",
		500,
		time.Time{},
		time.Time{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}

	// Test HasTask on empty iteration
	if iteration.HasTask("task-1") {
		t.Error("expected HasTask to return false for empty iteration")
	}

	// Test AddTask
	err = iteration.AddTask("task-1")
	if err != nil {
		t.Errorf("failed to add task: %v", err)
	}

	if !iteration.HasTask("task-1") {
		t.Error("expected HasTask to return true after AddTask")
	}

	// Test duplicate task
	err = iteration.AddTask("task-1")
	if err == nil {
		t.Error("expected error when adding duplicate task")
	}

	// Test GetTaskCount
	if iteration.GetTaskCount() != 1 {
		t.Errorf("expected task count 1, got %d", iteration.GetTaskCount())
	}

	// Test AddTask for multiple tasks
	err = iteration.AddTask("task-2")
	if err != nil {
		t.Errorf("failed to add second task: %v", err)
	}

	if iteration.GetTaskCount() != 2 {
		t.Errorf("expected task count 2, got %d", iteration.GetTaskCount())
	}

	// Test RemoveTask
	err = iteration.RemoveTask("task-1")
	if err != nil {
		t.Errorf("failed to remove task: %v", err)
	}

	if iteration.HasTask("task-1") {
		t.Error("expected HasTask to return false after RemoveTask")
	}

	if iteration.GetTaskCount() != 1 {
		t.Errorf("expected task count 1 after removal, got %d", iteration.GetTaskCount())
	}

	// Test removing non-existent task
	err = iteration.RemoveTask("task-999")
	if err == nil {
		t.Error("expected error when removing non-existent task")
	}
}

// TestIterationEntityWithTasks tests creation with initial tasks
func TestIterationEntityWithTasks(t *testing.T) {
	now := time.Now().UTC()
	taskIDs := []string{"task-1", "task-2", "task-3"}

	iteration, err := task_manager.NewIterationEntity(
		1,
		"Test Iteration",
		"Test Goal",
		"Test Deliverable",
		taskIDs,
		"planned",
		500,
		time.Time{},
		time.Time{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}

	if iteration.GetTaskCount() != 3 {
		t.Errorf("expected task count 3, got %d", iteration.GetTaskCount())
	}

	for _, taskID := range taskIDs {
		if !iteration.HasTask(taskID) {
			t.Errorf("expected iteration to have task %s", taskID)
		}
	}
}

// TestIterationEntityTimestamps tests timestamp handling
func TestIterationEntityTimestamps(t *testing.T) {
	createdAt := time.Now().UTC().Add(-2 * time.Hour)
	updatedAt := time.Now().UTC().Add(-1 * time.Hour)
	startedAt := time.Now().UTC()
	completedAt := time.Now().UTC().Add(1 * time.Hour)

	iteration, err := task_manager.NewIterationEntity(
		1,
		"Test Iteration",
		"Test Goal",
		"Test Deliverable",
		[]string{},
		"complete",
		500,
		startedAt,
		completedAt,
		createdAt,
		updatedAt,
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}

	fields := iteration.GetAllFields()

	if fields["created_at"].(time.Time) != createdAt {
		t.Error("created_at timestamp mismatch")
	}

	if fields["updated_at"].(time.Time) != updatedAt {
		t.Error("updated_at timestamp mismatch")
	}

	if fields["started_at"] != nil && fields["started_at"].(*time.Time).Before(startedAt) {
		t.Error("started_at timestamp mismatch")
	}

	if fields["completed_at"] != nil && fields["completed_at"].(*time.Time).Before(completedAt) {
		t.Error("completed_at timestamp mismatch")
	}
}

// TestIterationEntityIExtensible verifies IExtensible interface implementation
func TestIterationEntityIExtensible(t *testing.T) {
	now := time.Now().UTC()

	iteration, err := task_manager.NewIterationEntity(
		1,
		"Test Iteration",
		"Test Goal",
		"Test Deliverable",
		[]string{},
		"planned",
		500,
		time.Time{},
		time.Time{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}

	// Verify it implements IExtensible
	var _ pluginsdk.IExtensible = iteration

	// Verify required methods
	if iteration.GetID() == "" {
		t.Error("GetID() returned empty string")
	}

	if iteration.GetType() != "iteration" {
		t.Error("GetType() didn't return 'iteration'")
	}

	if iteration.GetCapabilities() == nil || len(iteration.GetCapabilities()) == 0 {
		t.Error("GetCapabilities() returned empty list")
	}

	fields := iteration.GetAllFields()
	if len(fields) == 0 {
		t.Error("GetAllFields() returned empty map")
	}

	if iteration.GetField("number") == nil {
		t.Error("GetField('number') returned nil")
	}
}

// TestIterationEntityValidStatuses tests all valid status values
func TestIterationEntityValidStatuses(t *testing.T) {
	now := time.Now().UTC()
	statuses := []string{"planned", "current", "complete"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			iteration, err := task_manager.NewIterationEntity(
				1,
				"Test",
				"Test",
				"Test",
				[]string{},
				status,
				500,
				time.Time{},
				time.Time{},
				now,
				now,
			)

			if err != nil {
				t.Errorf("failed to create iteration with status %q: %v", status, err)
			}
			if iteration == nil {
				t.Errorf("iteration is nil for status %q", status)
			}
		})
	}
}

// TestIterationEntityValidation tests validation logic
func TestIterationEntityValidation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name            string
		number          int
		status          string
		expectedErr     bool
		expectedErrType error
	}{
		{
			name:            "invalid number error",
			number:          0,
			status:          "planned",
			expectedErr:     true,
			expectedErrType: pluginsdk.ErrInvalidArgument,
		},
		{
			name:            "invalid status error",
			number:          1,
			status:          "invalid",
			expectedErr:     true,
			expectedErrType: pluginsdk.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := task_manager.NewIterationEntity(
				tt.number,
				"Test",
				"Test",
				"Test",
				[]string{},
				tt.status,
				500,
				time.Time{},
				time.Time{},
				now,
				now,
			)

			if tt.expectedErr && err == nil {
				t.Error("expected error, got nil")
			}

			if tt.expectedErr && err != nil {
				// Check if error wraps the expected error type
				if err != tt.expectedErrType && !isWrappedError(err, tt.expectedErrType) {
					t.Errorf("expected error type %v, got %v", tt.expectedErrType, err)
				}
			}
		})
	}
}

// TestIterationEntityLargeNumbers tests with large iteration numbers
func TestIterationEntityLargeNumbers(t *testing.T) {
	now := time.Now().UTC()

	iteration, err := task_manager.NewIterationEntity(
		999,
		"Test",
		"Test",
		"Test",
		[]string{},
		"planned",
		500,
		time.Time{},
		time.Time{},
		now,
		now,
	)

	if err != nil {
		t.Errorf("failed to create iteration with large number: %v", err)
	}

	if iteration.GetID() != "999" {
		t.Errorf("expected ID '999', got %q", iteration.GetID())
	}
}

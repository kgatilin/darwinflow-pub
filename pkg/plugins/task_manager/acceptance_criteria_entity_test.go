package task_manager_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

func TestNewAcceptanceCriteriaEntity(t *testing.T) {
	now := time.Now().UTC()
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"User can login",
		task_manager.VerificationTypeManual,
		"",
		now,
		now,
	)

	if ac.ID != "DW-ac-1" {
		t.Errorf("Expected ID 'DW-ac-1', got '%s'", ac.ID)
	}
	if ac.TaskID != "DW-task-1" {
		t.Errorf("Expected TaskID 'DW-task-1', got '%s'", ac.TaskID)
	}
	if ac.Description != "User can login" {
		t.Errorf("Expected description 'User can login', got '%s'", ac.Description)
	}
	if ac.Status != task_manager.ACStatusNotStarted {
		t.Errorf("Expected status 'not_started', got '%s'", ac.Status)
	}
}

func TestAcceptanceCriteriaIsVerified(t *testing.T) {
	tests := []struct {
		name     string
		status   task_manager.AcceptanceCriteriaStatus
		verified bool
	}{
		{"verified", task_manager.ACStatusVerified, true},
		{"automatically verified", task_manager.ACStatusAutomaticallyVerified, true},
		{"pending review", task_manager.ACStatusPendingHumanReview, false},
		{"failed", task_manager.ACStatusFailed, false},
		{"not started", task_manager.ACStatusNotStarted, false},
	}

	now := time.Now().UTC()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := task_manager.NewAcceptanceCriteriaEntity(
				"DW-ac-1",
				"DW-task-1",
				"Test AC",
				task_manager.VerificationTypeManual,
				"",
				now,
				now,
			)
			ac.Status = tt.status

			if ac.IsVerified() != tt.verified {
				t.Errorf("IsVerified() returned %v, expected %v", ac.IsVerified(), tt.verified)
			}
		})
	}
}

func TestAcceptanceCriteriaStatusIndicator(t *testing.T) {
	tests := []struct {
		name      string
		status    task_manager.AcceptanceCriteriaStatus
		indicator string
	}{
		{"verified", task_manager.ACStatusVerified, "✓"},
		{"automatically verified", task_manager.ACStatusAutomaticallyVerified, "✓"},
		{"pending review", task_manager.ACStatusPendingHumanReview, "⏸"},
		{"failed", task_manager.ACStatusFailed, "✗"},
		{"not started", task_manager.ACStatusNotStarted, "○"},
	}

	now := time.Now().UTC()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := task_manager.NewAcceptanceCriteriaEntity(
				"DW-ac-1",
				"DW-task-1",
				"Test AC",
				task_manager.VerificationTypeManual,
				"",
				now,
				now,
			)
			ac.Status = tt.status

			if ac.StatusIndicator() != tt.indicator {
				t.Errorf("StatusIndicator() returned %s, expected %s", ac.StatusIndicator(), tt.indicator)
			}
		})
	}
}

func TestAcceptanceCriteriaIsPendingReview(t *testing.T) {
	now := time.Now().UTC()
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"Test AC",
		task_manager.VerificationTypeManual,
		"",
		now,
		now,
	)

	ac.Status = task_manager.ACStatusPendingHumanReview
	if !ac.IsPendingReview() {
		t.Errorf("IsPendingReview() returned false, expected true")
	}

	ac.Status = task_manager.ACStatusVerified
	if ac.IsPendingReview() {
		t.Errorf("IsPendingReview() returned true, expected false")
	}
}

func TestAcceptanceCriteriaIsFailed(t *testing.T) {
	now := time.Now().UTC()
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"Test AC",
		task_manager.VerificationTypeManual,
		"",
		now,
		now,
	)

	ac.Status = task_manager.ACStatusFailed
	if !ac.IsFailed() {
		t.Errorf("IsFailed() returned false, expected true")
	}

	ac.Status = task_manager.ACStatusVerified
	if ac.IsFailed() {
		t.Errorf("IsFailed() returned true, expected false")
	}
}

func TestAcceptanceCriteriaGetID(t *testing.T) {
	now := time.Now().UTC()
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-123",
		"DW-task-1",
		"Test AC",
		task_manager.VerificationTypeManual,
		"",
		now,
		now,
	)

	if ac.GetID() != "DW-ac-123" {
		t.Errorf("GetID() returned %s, expected DW-ac-123", ac.GetID())
	}
}

func TestAcceptanceCriteriaGetType(t *testing.T) {
	now := time.Now().UTC()
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"Test AC",
		task_manager.VerificationTypeManual,
		"",
		now,
		now,
	)

	if ac.GetType() != "acceptance_criteria" {
		t.Errorf("GetType() returned %s, expected acceptance_criteria", ac.GetType())
	}
}

func TestAcceptanceCriteriaWithTestingInstructions(t *testing.T) {
	now := time.Now().UTC()
	testingInstructions := "1. Open the login page\n2. Enter credentials\n3. Click submit"
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"User can login",
		task_manager.VerificationTypeManual,
		testingInstructions,
		now,
		now,
	)

	if ac.TestingInstructions != testingInstructions {
		t.Errorf("Expected TestingInstructions '%s', got '%s'", testingInstructions, ac.TestingInstructions)
	}
}

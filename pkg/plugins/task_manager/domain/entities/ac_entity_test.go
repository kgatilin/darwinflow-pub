package entities_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestNewAcceptanceCriteriaEntity(t *testing.T) {
	now := time.Now()

	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"Feature works correctly",
		entities.VerificationTypeManual,
		"1. Open app\n2. Click button\n3. Verify output",
		now,
		now,
	)

	assert.NotNil(t, ac)
	assert.Equal(t, "TM-ac-1", ac.GetID())
	assert.Equal(t, "TM-task-1", ac.TaskID)
	assert.Equal(t, "Feature works correctly", ac.Description)
	assert.Equal(t, entities.VerificationTypeManual, ac.VerificationType)
	assert.Equal(t, "1. Open app\n2. Click button\n3. Verify output", ac.TestingInstructions)
	assert.Equal(t, entities.ACStatusNotStarted, ac.Status)
	assert.Equal(t, "", ac.Notes)
	assert.Equal(t, now, ac.CreatedAt)
	assert.Equal(t, now, ac.UpdatedAt)
}

func TestNewAcceptanceCriteriaEntity_AutomatedVerification(t *testing.T) {
	now := time.Now()

	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-2",
		"TM-task-1",
		"Tests pass with 90% coverage",
		entities.VerificationTypeAutomated,
		"Run: go test ./... -coverprofile=coverage.out",
		now,
		now,
	)

	assert.NotNil(t, ac)
	assert.Equal(t, entities.VerificationTypeAutomated, ac.VerificationType)
	assert.Equal(t, entities.ACStatusNotStarted, ac.Status)
}

func TestAcceptanceCriteriaEntity_GetID(t *testing.T) {
	now := time.Now()
	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-123",
		"TM-task-1",
		"description",
		entities.VerificationTypeManual,
		"steps",
		now,
		now,
	)

	assert.Equal(t, "TM-ac-123", ac.GetID())
}

func TestAcceptanceCriteriaEntity_GetType(t *testing.T) {
	now := time.Now()
	ac := entities.NewAcceptanceCriteriaEntity(
		"id",
		"task",
		"desc",
		entities.VerificationTypeManual,
		"steps",
		now,
		now,
	)

	assert.Equal(t, "acceptance_criteria", ac.GetType())
}

func TestAcceptanceCriteriaEntity_IsVerified(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		status     entities.AcceptanceCriteriaStatus
		wantResult bool
	}{
		{"verified manually", entities.ACStatusVerified, true},
		{"verified automatically", entities.ACStatusAutomaticallyVerified, true},
		{"not started", entities.ACStatusNotStarted, false},
		{"pending review", entities.ACStatusPendingHumanReview, false},
		{"failed", entities.ACStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := entities.NewAcceptanceCriteriaEntity(
				"id",
				"task",
				"desc",
				entities.VerificationTypeManual,
				"steps",
				now,
				now,
			)
			ac.Status = tt.status

			assert.Equal(t, tt.wantResult, ac.IsVerified())
		})
	}
}

func TestAcceptanceCriteriaEntity_IsFailed(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		status     entities.AcceptanceCriteriaStatus
		wantResult bool
	}{
		{"failed", entities.ACStatusFailed, true},
		{"verified", entities.ACStatusVerified, false},
		{"automatically verified", entities.ACStatusAutomaticallyVerified, false},
		{"not started", entities.ACStatusNotStarted, false},
		{"pending review", entities.ACStatusPendingHumanReview, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := entities.NewAcceptanceCriteriaEntity(
				"id",
				"task",
				"desc",
				entities.VerificationTypeManual,
				"steps",
				now,
				now,
			)
			ac.Status = tt.status

			assert.Equal(t, tt.wantResult, ac.IsFailed())
		})
	}
}

func TestAcceptanceCriteriaEntity_IsPendingReview(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		status     entities.AcceptanceCriteriaStatus
		wantResult bool
	}{
		{"pending review", entities.ACStatusPendingHumanReview, true},
		{"verified", entities.ACStatusVerified, false},
		{"automatically verified", entities.ACStatusAutomaticallyVerified, false},
		{"failed", entities.ACStatusFailed, false},
		{"not started", entities.ACStatusNotStarted, false},
		{"skipped", entities.ACStatusSkipped, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := entities.NewAcceptanceCriteriaEntity(
				"id",
				"task",
				"desc",
				entities.VerificationTypeManual,
				"steps",
				now,
				now,
			)
			ac.Status = tt.status

			assert.Equal(t, tt.wantResult, ac.IsPendingReview())
		})
	}
}

func TestAcceptanceCriteriaEntity_IsSkipped(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		status     entities.AcceptanceCriteriaStatus
		wantResult bool
	}{
		{"skipped", entities.ACStatusSkipped, true},
		{"verified", entities.ACStatusVerified, false},
		{"automatically verified", entities.ACStatusAutomaticallyVerified, false},
		{"failed", entities.ACStatusFailed, false},
		{"not started", entities.ACStatusNotStarted, false},
		{"pending review", entities.ACStatusPendingHumanReview, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := entities.NewAcceptanceCriteriaEntity(
				"id",
				"task",
				"desc",
				entities.VerificationTypeManual,
				"steps",
				now,
				now,
			)
			ac.Status = tt.status

			assert.Equal(t, tt.wantResult, ac.IsSkipped())
		})
	}
}

func TestAcceptanceCriteriaEntity_StatusIndicator(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		status        entities.AcceptanceCriteriaStatus
		wantIndicator string
	}{
		{"verified", entities.ACStatusVerified, "✓"},
		{"auto-verified", entities.ACStatusAutomaticallyVerified, "✓"},
		{"pending", entities.ACStatusPendingHumanReview, "⏸"},
		{"failed", entities.ACStatusFailed, "✗"},
		{"skipped", entities.ACStatusSkipped, "⊘"},
		{"not started", entities.ACStatusNotStarted, "○"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := entities.NewAcceptanceCriteriaEntity(
				"id",
				"task",
				"desc",
				entities.VerificationTypeManual,
				"steps",
				now,
				now,
			)
			ac.Status = tt.status

			assert.Equal(t, tt.wantIndicator, ac.StatusIndicator())
		})
	}
}

// Test edge cases
func TestAcceptanceCriteriaEntity_EmptyFields(t *testing.T) {
	now := time.Now()

	ac := entities.NewAcceptanceCriteriaEntity(
		"",
		"",
		"",
		entities.VerificationTypeManual,
		"",
		now,
		now,
	)

	assert.NotNil(t, ac)
	assert.Equal(t, "", ac.GetID())
	assert.Equal(t, "", ac.TaskID)
	assert.Equal(t, "", ac.Description)
	assert.Equal(t, "", ac.TestingInstructions)
	assert.Equal(t, entities.ACStatusNotStarted, ac.Status)
	assert.Equal(t, "", ac.Notes)
}

func TestAcceptanceCriteriaEntity_LongDescription(t *testing.T) {
	now := time.Now()
	longDesc := "This is a very long description that contains detailed information about what needs to be verified. " +
		"It spans multiple lines and provides comprehensive acceptance criteria. " +
		"The system should handle long descriptions without issues."

	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		longDesc,
		entities.VerificationTypeManual,
		"steps",
		now,
		now,
	)

	assert.Equal(t, longDesc, ac.Description)
}

func TestAcceptanceCriteriaEntity_MultilineTestingInstructions(t *testing.T) {
	now := time.Now()
	instructions := `1. Run: cd pkg/plugins/task_manager/domain
2. Run: go test ./... -coverprofile=coverage.out
3. Run: go tool cover -func=coverage.out | grep total
4. Verify: total coverage >= 90%
5. Run: go test ./... -v
6. Verify: All tests pass with zero failures`

	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"Domain layer has 90%+ test coverage",
		entities.VerificationTypeAutomated,
		instructions,
		now,
		now,
	)

	assert.Equal(t, instructions, ac.TestingInstructions)
	assert.Equal(t, entities.VerificationTypeAutomated, ac.VerificationType)
}

func TestAcceptanceCriteriaEntity_StatusTransitions(t *testing.T) {
	now := time.Now()
	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"Feature works",
		entities.VerificationTypeManual,
		"Test it",
		now,
		now,
	)

	// Initial state
	assert.Equal(t, entities.ACStatusNotStarted, ac.Status)
	assert.False(t, ac.IsVerified())
	assert.False(t, ac.IsFailed())
	assert.False(t, ac.IsPendingReview())
	assert.Equal(t, "○", ac.StatusIndicator())

	// Transition to pending review
	ac.Status = entities.ACStatusPendingHumanReview
	assert.False(t, ac.IsVerified())
	assert.False(t, ac.IsFailed())
	assert.True(t, ac.IsPendingReview())
	assert.Equal(t, "⏸", ac.StatusIndicator())

	// Transition to verified
	ac.Status = entities.ACStatusVerified
	assert.True(t, ac.IsVerified())
	assert.False(t, ac.IsFailed())
	assert.False(t, ac.IsPendingReview())
	assert.Equal(t, "✓", ac.StatusIndicator())

	// Transition to failed
	ac.Status = entities.ACStatusFailed
	assert.False(t, ac.IsVerified())
	assert.True(t, ac.IsFailed())
	assert.False(t, ac.IsPendingReview())
	assert.Equal(t, "✗", ac.StatusIndicator())

	// Transition to automatically verified
	ac.Status = entities.ACStatusAutomaticallyVerified
	assert.True(t, ac.IsVerified())
	assert.False(t, ac.IsFailed())
	assert.False(t, ac.IsPendingReview())
	assert.False(t, ac.IsSkipped())
	assert.Equal(t, "✓", ac.StatusIndicator())

	// Transition to skipped
	ac.Status = entities.ACStatusSkipped
	assert.False(t, ac.IsVerified())
	assert.False(t, ac.IsFailed())
	assert.False(t, ac.IsPendingReview())
	assert.True(t, ac.IsSkipped())
	assert.Equal(t, "⊘", ac.StatusIndicator())
}

func TestAcceptanceCriteriaEntity_VerificationTypes(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		verificationType entities.AcceptanceCriteriaVerificationType
	}{
		{"manual verification", entities.VerificationTypeManual},
		{"automated verification", entities.VerificationTypeAutomated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := entities.NewAcceptanceCriteriaEntity(
				"TM-ac-1",
				"TM-task-1",
				"Test criterion",
				tt.verificationType,
				"Test steps",
				now,
				now,
			)

			assert.Equal(t, tt.verificationType, ac.VerificationType)
		})
	}
}

func TestAcceptanceCriteriaEntity_NotesField(t *testing.T) {
	now := time.Now()
	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"Test feature",
		entities.VerificationTypeManual,
		"Test steps",
		now,
		now,
	)

	// Initially empty
	assert.Equal(t, "", ac.Notes)

	// Can be updated
	ac.Notes = "Failed due to missing dependency"
	assert.Equal(t, "Failed due to missing dependency", ac.Notes)

	// Can be cleared
	ac.Notes = ""
	assert.Equal(t, "", ac.Notes)
}

func TestAcceptanceCriteriaEntity_Timestamps(t *testing.T) {
	createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 1, 2, 15, 30, 0, 0, time.UTC)

	ac := entities.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"Test feature",
		entities.VerificationTypeManual,
		"Test steps",
		createdAt,
		updatedAt,
	)

	assert.Equal(t, createdAt, ac.CreatedAt)
	assert.Equal(t, updatedAt, ac.UpdatedAt)
	assert.True(t, ac.UpdatedAt.After(ac.CreatedAt))
}

// SDK Interface Tests
// Note: AcceptanceCriteriaEntity only implements GetID() and GetType() from SDK interfaces

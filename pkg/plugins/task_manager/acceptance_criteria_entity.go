package task_manager

import (
	"time"
)

// AcceptanceCriteriaStatus represents the current status of an acceptance criterion
type AcceptanceCriteriaStatus string

const (
	// ACStatusNotStarted - AC has not been verified yet
	ACStatusNotStarted AcceptanceCriteriaStatus = "not_started"
	// ACStatusAutomaticallyVerified - AC was verified by automated process
	ACStatusAutomaticallyVerified AcceptanceCriteriaStatus = "automatically_verified"
	// ACStatusPendingHumanReview - AC is awaiting human verification
	ACStatusPendingHumanReview AcceptanceCriteriaStatus = "pending_human_review"
	// ACStatusVerified - AC has been manually verified by human
	ACStatusVerified AcceptanceCriteriaStatus = "verified"
	// ACStatusFailed - AC did not meet verification requirements
	ACStatusFailed AcceptanceCriteriaStatus = "failed"
)

// AcceptanceCriteriaVerificationType indicates who should verify this AC
type AcceptanceCriteriaVerificationType string

const (
	// VerificationTypeManual - Requires manual human verification
	VerificationTypeManual AcceptanceCriteriaVerificationType = "manual"
	// VerificationTypeAutomated - Can be automatically verified by coding agent
	VerificationTypeAutomated AcceptanceCriteriaVerificationType = "automated"
)

// AcceptanceCriteriaEntity represents a single acceptance criterion for a task
type AcceptanceCriteriaEntity struct {
	ID                  string                            `json:"id"`
	TaskID              string                            `json:"task_id"`           // Parent task ID
	Description         string                            `json:"description"`       // What must be verified
	VerificationType    AcceptanceCriteriaVerificationType `json:"verification_type"` // manual or automated
	Status              AcceptanceCriteriaStatus          `json:"status"`            // Current verification status
	Notes               string                            `json:"notes"`             // Additional notes (reason, feedback, etc.)
	TestingInstructions string                            `json:"testing_instructions"` // Step-by-step testing guidance
	CreatedAt           time.Time                         `json:"created_at"`
	UpdatedAt           time.Time                         `json:"updated_at"`
}

// NewAcceptanceCriteriaEntity creates a new acceptance criterion entity
func NewAcceptanceCriteriaEntity(
	id string,
	taskID string,
	description string,
	verificationType AcceptanceCriteriaVerificationType,
	testingInstructions string,
	createdAt time.Time,
	updatedAt time.Time,
) *AcceptanceCriteriaEntity {
	return &AcceptanceCriteriaEntity{
		ID:                  id,
		TaskID:              taskID,
		Description:         description,
		VerificationType:    verificationType,
		Status:              ACStatusNotStarted,
		Notes:               "",
		TestingInstructions: testingInstructions,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}
}

// GetID returns the unique identifier for this entity
func (ac *AcceptanceCriteriaEntity) GetID() string {
	return ac.ID
}

// GetType returns the entity type
func (ac *AcceptanceCriteriaEntity) GetType() string {
	return "acceptance_criteria"
}

// IsVerified returns true if the AC has been verified (manually or automatically)
func (ac *AcceptanceCriteriaEntity) IsVerified() bool {
	return ac.Status == ACStatusVerified || ac.Status == ACStatusAutomaticallyVerified
}

// IsFailed returns true if the AC has failed verification
func (ac *AcceptanceCriteriaEntity) IsFailed() bool {
	return ac.Status == ACStatusFailed
}

// IsPendingReview returns true if the AC is awaiting human review
func (ac *AcceptanceCriteriaEntity) IsPendingReview() bool {
	return ac.Status == ACStatusPendingHumanReview
}

// StatusIndicator returns a visual indicator for the AC status
func (ac *AcceptanceCriteriaEntity) StatusIndicator() string {
	switch ac.Status {
	case ACStatusVerified, ACStatusAutomaticallyVerified:
		return "✓"
	case ACStatusPendingHumanReview:
		return "⏸"
	case ACStatusFailed:
		return "✗"
	default: // not_started
		return "○"
	}
}

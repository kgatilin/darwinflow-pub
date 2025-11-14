package entities

import (
	"fmt"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// TrackStatus represents valid status values for tracks
type TrackStatus string

const (
	TrackStatusNotStarted TrackStatus = "not-started"
	TrackStatusInProgress TrackStatus = "in-progress"
	TrackStatusComplete   TrackStatus = "complete"
	TrackStatusBlocked    TrackStatus = "blocked"
	TrackStatusWaiting    TrackStatus = "waiting"
)

// Valid status values for tracks
var validTrackStatuses = map[string]bool{
	string(TrackStatusNotStarted): true,
	string(TrackStatusInProgress): true,
	string(TrackStatusComplete):   true,
	string(TrackStatusBlocked):    true,
	string(TrackStatusWaiting):    true,
}

// IsValidTrackStatus validates a track status string
func IsValidTrackStatus(status string) bool {
	return validTrackStatuses[status]
}

// TaskStatus represents valid status values for tasks
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in-progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

// Valid status values for tasks
var validTaskStatuses = map[string]bool{
	string(TaskStatusTodo):       true,
	string(TaskStatusInProgress): true,
	string(TaskStatusReview):     true,
	string(TaskStatusDone):       true,
}

// IsValidTaskStatus validates a task status string
func IsValidTaskStatus(status string) bool {
	return validTaskStatuses[status]
}

// IterationStatus represents valid status values for iterations
type IterationStatus string

const (
	IterationStatusPlanned  IterationStatus = "planned"
	IterationStatusCurrent  IterationStatus = "current"
	IterationStatusComplete IterationStatus = "complete"
)

// Valid status values for iterations
var validIterationStatuses = map[string]bool{
	string(IterationStatusPlanned):  true,
	string(IterationStatusCurrent):  true,
	string(IterationStatusComplete): true,
}

// IsValidIterationStatus validates an iteration status string
func IsValidIterationStatus(status string) bool {
	return validIterationStatuses[status]
}

// ADRStatus represents the lifecycle status of an ADR
type ADRStatus string

const (
	ADRStatusProposed    ADRStatus = "proposed"
	ADRStatusAccepted    ADRStatus = "accepted"
	ADRStatusDeprecated  ADRStatus = "deprecated"
	ADRStatusSuperseded  ADRStatus = "superseded"
)

// Valid ADR statuses
var validADRStatuses = map[string]bool{
	string(ADRStatusProposed):   true,
	string(ADRStatusAccepted):   true,
	string(ADRStatusDeprecated): true,
	string(ADRStatusSuperseded): true,
}

// IsValidADRStatus validates an ADR status string
func IsValidADRStatus(status string) bool {
	return validADRStatuses[status]
}

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
	// ACStatusSkipped - AC was intentionally skipped with a reason
	ACStatusSkipped AcceptanceCriteriaStatus = "skipped"
)

// AcceptanceCriteriaVerificationType indicates who should verify this AC
type AcceptanceCriteriaVerificationType string

const (
	// VerificationTypeManual - Requires manual human verification
	VerificationTypeManual AcceptanceCriteriaVerificationType = "manual"
	// VerificationTypeAutomated - Can be automatically verified by coding agent
	VerificationTypeAutomated AcceptanceCriteriaVerificationType = "automated"
)

// Filter types for queries

// TrackFilters represents filter criteria for track queries
type TrackFilters struct {
	Status   []string // Filter by status values (e.g., "not-started", "in-progress")
	Priority []string // Legacy - not used
}

// TaskFilters represents filter criteria for task queries
type TaskFilters struct {
	TrackID  string   // Filter by parent track ID
	Status   []string // Filter by status values (e.g., "todo", "in-progress", "review", "done")
	Priority []string // Legacy - not used
}

// ACFilters represents filter criteria for acceptance criteria queries
type ACFilters struct {
	IterationNum *int   // Filter by iteration number
	TrackID      string // Filter by track ID (via tasks)
	TaskID       string // Filter by task ID
}

// DocumentType represents valid document type values
type DocumentType string

const (
	DocumentTypeADR            DocumentType = "adr"
	DocumentTypePlan           DocumentType = "plan"
	DocumentTypeRetrospective  DocumentType = "retrospective"
	DocumentTypeOther          DocumentType = "other"
)

// Valid document types
var validDocumentTypes = map[string]bool{
	string(DocumentTypeADR):            true,
	string(DocumentTypePlan):           true,
	string(DocumentTypeRetrospective):  true,
	string(DocumentTypeOther):          true,
}

// NewDocumentType creates a DocumentType with validation
func NewDocumentType(docType string) (DocumentType, error) {
	if !validDocumentTypes[docType] {
		return "", fmt.Errorf("%w: invalid document type: must be one of adr, plan, retrospective, other", pluginsdk.ErrInvalidArgument)
	}
	return DocumentType(docType), nil
}

// String returns the string representation of the document type
func (dt DocumentType) String() string {
	return string(dt)
}

// IsValid checks if the document type is valid
func (dt DocumentType) IsValid() bool {
	return validDocumentTypes[string(dt)]
}

// DocumentStatus represents valid document status values
type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusPublished DocumentStatus = "published"
	DocumentStatusArchived  DocumentStatus = "archived"
)

// Valid document statuses
var validDocumentStatuses = map[string]bool{
	string(DocumentStatusDraft):     true,
	string(DocumentStatusPublished): true,
	string(DocumentStatusArchived):  true,
}

// NewDocumentStatus creates a DocumentStatus with validation
func NewDocumentStatus(status string) (DocumentStatus, error) {
	if !validDocumentStatuses[status] {
		return "", fmt.Errorf("%w: invalid document status: must be one of draft, published, archived", pluginsdk.ErrInvalidArgument)
	}
	return DocumentStatus(status), nil
}

// String returns the string representation of the document status
func (ds DocumentStatus) String() string {
	return string(ds)
}

// IsValid checks if the document status is valid
func (ds DocumentStatus) IsValid() bool {
	return validDocumentStatuses[string(ds)]
}

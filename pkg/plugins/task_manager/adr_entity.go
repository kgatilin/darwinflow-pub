package task_manager

import (
	"fmt"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

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

// ADREntity represents an Architecture Decision Record for a track
type ADREntity struct {
	ID           string    `json:"id"`
	TrackID      string    `json:"track_id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"` // proposed, accepted, deprecated, superseded
	Context      string    `json:"context"`
	Decision     string    `json:"decision"`
	Consequences string    `json:"consequences"`
	Alternatives string    `json:"alternatives"` // Optional
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	SupersededBy *string   `json:"superseded_by"` // Optional: ID of ADR that superseded this one
}

// NewADREntity creates a new ADR entity with validation
func NewADREntity(
	id, trackID, title, status, context, decision, consequences, alternatives string,
	createdAt, updatedAt time.Time,
	supersededBy *string,
) (*ADREntity, error) {
	// Validate status
	if !validADRStatuses[status] {
		return nil, fmt.Errorf("%w: invalid ADR status: must be one of proposed, accepted, deprecated, superseded", pluginsdk.ErrInvalidArgument)
	}

	// Validate required fields
	if title == "" {
		return nil, fmt.Errorf("%w: ADR title is required", pluginsdk.ErrInvalidArgument)
	}
	if context == "" {
		return nil, fmt.Errorf("%w: ADR context is required", pluginsdk.ErrInvalidArgument)
	}
	if decision == "" {
		return nil, fmt.Errorf("%w: ADR decision is required", pluginsdk.ErrInvalidArgument)
	}
	if consequences == "" {
		return nil, fmt.Errorf("%w: ADR consequences are required", pluginsdk.ErrInvalidArgument)
	}

	// If superseded, SupersededBy must be set
	if status == string(ADRStatusSuperseded) && supersededBy == nil {
		return nil, fmt.Errorf("%w: superseded ADR must specify superseded_by", pluginsdk.ErrInvalidArgument)
	}

	return &ADREntity{
		ID:           id,
		TrackID:      trackID,
		Title:        title,
		Status:       status,
		Context:      context,
		Decision:     decision,
		Consequences: consequences,
		Alternatives: alternatives,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		SupersededBy: supersededBy,
	}, nil
}

// IExtensible implementation

// GetID returns the unique identifier for this entity
func (a *ADREntity) GetID() string {
	return a.ID
}

// GetType returns the entity type
func (a *ADREntity) GetType() string {
	return "adr"
}

// GetCapabilities returns list of capability names this entity supports
func (a *ADREntity) GetCapabilities() []string {
	return []string{"IExtensible"}
}

// GetField retrieves a named field value
func (a *ADREntity) GetField(name string) interface{} {
	fields := a.GetAllFields()
	return fields[name]
}

// GetAllFields returns all fields as a map
func (a *ADREntity) GetAllFields() map[string]interface{} {
	supersededBy := ""
	if a.SupersededBy != nil {
		supersededBy = *a.SupersededBy
	}
	return map[string]interface{}{
		"id":            a.ID,
		"track_id":      a.TrackID,
		"title":         a.Title,
		"status":        a.Status,
		"context":       a.Context,
		"decision":      a.Decision,
		"consequences":  a.Consequences,
		"alternatives":  a.Alternatives,
		"created_at":    a.CreatedAt,
		"updated_at":    a.UpdatedAt,
		"superseded_by": supersededBy,
	}
}

// IsAccepted returns true if the ADR is in accepted status
func (a *ADREntity) IsAccepted() bool {
	return a.Status == string(ADRStatusAccepted)
}

// IsSuperseded returns true if the ADR is in superseded status
func (a *ADREntity) IsSuperseded() bool {
	return a.Status == string(ADRStatusSuperseded)
}

// IsDeprecated returns true if the ADR is in deprecated status
func (a *ADREntity) IsDeprecated() bool {
	return a.Status == string(ADRStatusDeprecated)
}

// ToMarkdown formats the ADR as markdown
func (a *ADREntity) ToMarkdown() string {
	md := fmt.Sprintf("# ADR %s: %s\n\n", a.ID, a.Title)
	md += fmt.Sprintf("**Status**: %s\n\n", a.Status)

	if a.SupersededBy != nil {
		md += fmt.Sprintf("**Superseded by**: %s\n\n", *a.SupersededBy)
	}

	md += "## Context\n\n"
	md += a.Context + "\n\n"

	md += "## Decision\n\n"
	md += a.Decision + "\n\n"

	md += "## Consequences\n\n"
	md += a.Consequences + "\n\n"

	if a.Alternatives != "" {
		md += "## Alternatives\n\n"
		md += a.Alternatives + "\n\n"
	}

	md += fmt.Sprintf("**Created**: %s\n", a.CreatedAt.Format(time.RFC3339))
	md += fmt.Sprintf("**Updated**: %s\n", a.UpdatedAt.Format(time.RFC3339))

	return md
}

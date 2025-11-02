package task_manager_test

import (
	"testing"
	"time"

	tm "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

func TestNewADREntity(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		id        string
		trackID   string
		title     string
		status    string
		context   string
		decision  string
		consequences string
		alternatives string
		shouldErr bool
	}{
		{
			name:      "valid ADR",
			id:        "DW-adr-1",
			trackID:   "track-1",
			title:     "Use PostgreSQL",
			status:    string(tm.ADRStatusProposed),
			context:   "Need persistent storage",
			decision:  "Choose PostgreSQL",
			consequences: "Must maintain database migrations",
			alternatives: "Consider MySQL or MongoDB",
			shouldErr: false,
		},
		{
			name:      "missing title",
			id:        "DW-adr-1",
			trackID:   "track-1",
			title:     "",
			status:    string(tm.ADRStatusProposed),
			context:   "Need persistent storage",
			decision:  "Choose PostgreSQL",
			consequences: "Must maintain database migrations",
			shouldErr: true,
		},
		{
			name:      "missing context",
			id:        "DW-adr-1",
			trackID:   "track-1",
			title:     "Use PostgreSQL",
			status:    string(tm.ADRStatusProposed),
			context:   "",
			decision:  "Choose PostgreSQL",
			consequences: "Must maintain database migrations",
			shouldErr: true,
		},
		{
			name:      "invalid status",
			id:        "DW-adr-1",
			trackID:   "track-1",
			title:     "Use PostgreSQL",
			status:    "invalid",
			context:   "Need persistent storage",
			decision:  "Choose PostgreSQL",
			consequences: "Must maintain database migrations",
			shouldErr: true,
		},
		{
			name:      "superseded without superseded_by",
			id:        "DW-adr-1",
			trackID:   "track-1",
			title:     "Use PostgreSQL",
			status:    string(tm.ADRStatusSuperseded),
			context:   "Need persistent storage",
			decision:  "Choose PostgreSQL",
			consequences: "Must maintain database migrations",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var supersededBy *string
			if tt.status == string(tm.ADRStatusSuperseded) && tt.shouldErr == false {
				id := "DW-adr-2"
				supersededBy = &id
			}

			adr, err := tm.NewADREntity(
				tt.id,
				tt.trackID,
				tt.title,
				tt.status,
				tt.context,
				tt.decision,
				tt.consequences,
				tt.alternatives,
				now,
				now,
				supersededBy,
			)

			if (err != nil) != tt.shouldErr {
				t.Errorf("expected error: %v, got: %v", tt.shouldErr, err != nil)
				return
			}

			if !tt.shouldErr && adr == nil {
				t.Error("expected non-nil ADR")
			}
		})
	}
}

func TestADREntity_ToMarkdown(t *testing.T) {
	now := time.Now().UTC()
	adr, _ := tm.NewADREntity(
		"DW-adr-1",
		"track-1",
		"Use PostgreSQL",
		string(tm.ADRStatusAccepted),
		"Need persistent storage",
		"Choose PostgreSQL for reliability",
		"Must maintain database migrations",
		"Could use MongoDB or MySQL",
		now,
		now,
		nil,
	)

	markdown := adr.ToMarkdown()

	// Check that markdown contains expected sections
	if !containsStr(markdown, "# ADR DW-adr-1: Use PostgreSQL") {
		t.Error("markdown missing title section")
	}
	if !containsStr(markdown, "**Status**: accepted") {
		t.Error("markdown missing status")
	}
	if !containsStr(markdown, "## Context") {
		t.Error("markdown missing context section")
	}
	if !containsStr(markdown, "## Decision") {
		t.Error("markdown missing decision section")
	}
	if !containsStr(markdown, "## Consequences") {
		t.Error("markdown missing consequences section")
	}
	if !containsStr(markdown, "## Alternatives") {
		t.Error("markdown missing alternatives section")
	}
}

func TestADREntity_IExtensible(t *testing.T) {
	now := time.Now().UTC()
	adr, _ := tm.NewADREntity(
		"DW-adr-1",
		"track-1",
		"Use PostgreSQL",
		string(tm.ADRStatusAccepted),
		"Need persistent storage",
		"Choose PostgreSQL",
		"Must maintain database migrations",
		"",
		now,
		now,
		nil,
	)

	// Test IExtensible interface
	if adr.GetID() != "DW-adr-1" {
		t.Error("GetID failed")
	}

	if adr.GetType() != "adr" {
		t.Error("GetType failed")
	}

	capabilities := adr.GetCapabilities()
	if len(capabilities) == 0 {
		t.Error("GetCapabilities returned empty list")
	}

	fields := adr.GetAllFields()
	if _, ok := fields["id"]; !ok {
		t.Error("GetAllFields missing 'id'")
	}

	if adr.GetField("title") != "Use PostgreSQL" {
		t.Error("GetField failed")
	}
}

func TestADREntity_StatusMethods(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		status   string
		wantAccepted bool
		wantDeprecated bool
		wantSuperseded bool
	}{
		{string(tm.ADRStatusAccepted), true, false, false},
		{string(tm.ADRStatusDeprecated), false, true, false},
		{string(tm.ADRStatusProposed), false, false, false},
	}

	for _, tt := range tests {
		adr, _ := tm.NewADREntity(
			"DW-adr-1",
			"track-1",
			"Test",
			tt.status,
			"Context",
			"Decision",
			"Consequences",
			"",
			now,
			now,
			nil,
		)

		if adr.IsAccepted() != tt.wantAccepted {
			t.Errorf("IsAccepted() for %s = %v, want %v", tt.status, adr.IsAccepted(), tt.wantAccepted)
		}

		if adr.IsDeprecated() != tt.wantDeprecated {
			t.Errorf("IsDeprecated() for %s = %v, want %v", tt.status, adr.IsDeprecated(), tt.wantDeprecated)
		}

		if adr.IsSuperseded() != tt.wantSuperseded {
			t.Errorf("IsSuperseded() for %s = %v, want %v", tt.status, adr.IsSuperseded(), tt.wantSuperseded)
		}
	}
}

// Helper function to check if string contains substring
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

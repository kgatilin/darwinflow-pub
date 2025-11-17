package dto_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/dto"
)

// TestStringPtr verifies that StringPtr returns a pointer to the given string
func TestStringPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"non-empty string", "test value"},
		{"string with spaces", "test value with spaces"},
		{"string with special chars", "test@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := dto.StringPtr(tt.input)

			if ptr == nil {
				t.Fatal("StringPtr returned nil")
			}

			if *ptr != tt.input {
				t.Errorf("StringPtr() = %q, want %q", *ptr, tt.input)
			}

			// Verify it's a unique pointer (modifying original doesn't affect pointer)
			original := tt.input
			*ptr = "modified"
			if original != tt.input {
				t.Error("StringPtr should create a copy, not reference the original")
			}
		})
	}
}

// TestIntPtr verifies that IntPtr returns a pointer to the given int
func TestIntPtr(t *testing.T) {
	tests := []struct {
		name  string
		input int
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -42},
		{"max int", int(^uint(0) >> 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := dto.IntPtr(tt.input)

			if ptr == nil {
				t.Fatal("IntPtr returned nil")
			}

			if *ptr != tt.input {
				t.Errorf("IntPtr() = %d, want %d", *ptr, tt.input)
			}
		})
	}
}

// TestBoolPtr verifies that BoolPtr returns a pointer to the given bool
func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name  string
		input bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := dto.BoolPtr(tt.input)

			if ptr == nil {
				t.Fatal("BoolPtr returned nil")
			}

			if *ptr != tt.input {
				t.Errorf("BoolPtr() = %v, want %v", *ptr, tt.input)
			}
		})
	}
}

// TestTimePtr verifies that TimePtr returns a pointer to the given time
func TestTimePtr(t *testing.T) {
	now := time.Now()
	epoch := time.Unix(0, 0)
	future := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name  string
		input time.Time
	}{
		{"now", now},
		{"epoch", epoch},
		{"future", future},
		{"zero time", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := dto.TimePtr(tt.input)

			if ptr == nil {
				t.Fatal("TimePtr returned nil")
			}

			if !ptr.Equal(tt.input) {
				t.Errorf("TimePtr() = %v, want %v", *ptr, tt.input)
			}
		})
	}
}

// TestHelperUsageInUpdateDTOs demonstrates how helpers are used in update DTOs
func TestHelperUsageInUpdateDTOs(t *testing.T) {
	// Example: Creating an update DTO with optional fields
	updateTrack := dto.UpdateTrackDTO{
		ID:          "TM-track-1",
		Title:       dto.StringPtr("New Title"),
		Description: dto.StringPtr("New Description"),
		Status:      dto.StringPtr("in-progress"),
		Rank:        dto.IntPtr(100),
	}

	// Verify all fields are set
	if updateTrack.Title == nil {
		t.Error("Title should not be nil")
	}
	if *updateTrack.Title != "New Title" {
		t.Errorf("Title = %q, want 'New Title'", *updateTrack.Title)
	}

	if updateTrack.Rank == nil {
		t.Error("Rank should not be nil")
	}
	if *updateTrack.Rank != 100 {
		t.Errorf("Rank = %d, want 100", *updateTrack.Rank)
	}

	// Example: Partial update (only some fields set)
	partialUpdate := dto.UpdateTrackDTO{
		ID:    "TM-track-2",
		Title: dto.StringPtr("Only Title Updated"),
		// Description, Status, Rank are nil (no change)
	}

	if partialUpdate.Description != nil {
		t.Error("Description should be nil for partial update")
	}
	if partialUpdate.Status != nil {
		t.Error("Status should be nil for partial update")
	}
	if partialUpdate.Rank != nil {
		t.Error("Rank should be nil for partial update")
	}
}

// TestHelperReturnsDifferentPointers verifies that helpers create new pointers each time
func TestHelperReturnsDifferentPointers(t *testing.T) {
	// String pointers
	s1 := dto.StringPtr("test")
	s2 := dto.StringPtr("test")
	if s1 == s2 {
		t.Error("StringPtr should return different pointers for same value")
	}

	// Int pointers
	i1 := dto.IntPtr(42)
	i2 := dto.IntPtr(42)
	if i1 == i2 {
		t.Error("IntPtr should return different pointers for same value")
	}

	// Bool pointers
	b1 := dto.BoolPtr(true)
	b2 := dto.BoolPtr(true)
	if b1 == b2 {
		t.Error("BoolPtr should return different pointers for same value")
	}

	// Time pointers
	now := time.Now()
	t1 := dto.TimePtr(now)
	t2 := dto.TimePtr(now)
	if t1 == t2 {
		t.Error("TimePtr should return different pointers for same value")
	}
}

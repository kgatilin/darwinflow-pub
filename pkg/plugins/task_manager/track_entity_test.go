package task_manager_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// TestNewTrackEntity tests track entity creation
func TestNewTrackEntity(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		id          string
		status      string
		rank        int
		deps        []string
		expectedErr bool
	}{
		{
			name:        "valid track",
			id:          "track-framework-core",
			status:      "not-started",
			rank:        100,
			deps:        []string{},
			expectedErr: false,
		},
		{
			name:        "invalid track ID format",
			id:          "framework-core", // missing "track-" prefix
			status:      "not-started",
			rank:        100,
			deps:        []string{},
			expectedErr: true,
		},
		{
			name:        "invalid status",
			id:          "track-framework-core",
			status:      "unknown",
			rank:        100,
			deps:        []string{},
			expectedErr: true,
		},
		{
			name:        "invalid rank",
			id:          "track-framework-core",
			status:      "not-started",
			rank:        -1, // negative rank is invalid
			deps:        []string{},
			expectedErr: true,
		},
		{
			name:        "self-dependency",
			id:          "track-framework-core",
			status:      "not-started",
			rank:        100,
			deps:        []string{"track-framework-core"},
			expectedErr: true,
		},
		{
			name:        "valid with dependencies",
			id:          "track-plugin-system",
			status:      "not-started",
			rank:        200,
			deps:        []string{"track-framework-core", "track-database"},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track, err := task_manager.NewTrackEntity(
				tt.id,
				"roadmap-001",
				"Test Track",
				"Test Description",
				tt.status,
				tt.rank,
				tt.deps,
				now,
				now,
			)

			if tt.expectedErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedErr && track == nil {
				t.Error("expected track, got nil")
			}
		})
	}
}

// TestTrackEntityGetters tests track entity field getters
func TestTrackEntityGetters(t *testing.T) {
	now := time.Now().UTC()
	id := "track-framework-core"

	track, err := task_manager.NewTrackEntity(
		id,
		"roadmap-001",
		"Framework Core",
		"Core framework work",
		"in-progress",
		100,
		[]string{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}

	if track.GetID() != id {
		t.Errorf("expected ID %q, got %q", id, track.GetID())
	}

	if track.GetType() != "track" {
		t.Errorf("expected type 'track', got %q", track.GetType())
	}

	if track.GetStatus() != "in-progress" {
		t.Errorf("expected status 'in-progress', got %q", track.GetStatus())
	}

	caps := track.GetCapabilities()
	if len(caps) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(caps))
	}
}

// TestTrackEntityProgress tests GetProgress method
func TestTrackEntityProgress(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		status   string
		expected float64
	}{
		{"not-started", 0.0},
		{"in-progress", 0.5},
		{"complete", 1.0},
		{"blocked", 0.0},
		{"waiting", 0.0},
	}

	for _, tt := range tests {
		track, err := task_manager.NewTrackEntity(
			"track-test",
			"roadmap-001",
			"Test",
			"Test",
			tt.status,
			300,
			[]string{},
			now,
			now,
		)
		if err != nil {
			t.Fatalf("failed to create track: %v", err)
		}

		progress := track.GetProgress()
		if progress != tt.expected {
			t.Errorf("for status %q, expected progress %.1f, got %.1f", tt.status, tt.expected, progress)
		}
	}
}

// TestTrackEntityIsBlocked tests IsBlocked method
func TestTrackEntityIsBlocked(t *testing.T) {
	now := time.Now().UTC()

	blockedTrack, err := task_manager.NewTrackEntity(
		"track-blocked",
		"roadmap-001",
		"Blocked Track",
		"Test",
		"blocked",
		100,
		[]string{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create blocked track: %v", err)
	}

	if !blockedTrack.IsBlocked() {
		t.Error("track with status 'blocked' should return true for IsBlocked()")
	}

	notBlockedTrack, err := task_manager.NewTrackEntity(
		"track-active",
		"roadmap-001",
		"Active Track",
		"Test",
		"in-progress",
		100,
		[]string{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create active track: %v", err)
	}

	if notBlockedTrack.IsBlocked() {
		t.Error("track with status 'in-progress' should return false for IsBlocked()")
	}
}

// TestTrackEntityDependencies tests dependency management
func TestTrackEntityDependencies(t *testing.T) {
	now := time.Now().UTC()

	track, err := task_manager.NewTrackEntity(
		"track-plugin-system",
		"roadmap-001",
		"Plugin System",
		"Test",
		"not-started",
		200,
		[]string{"track-framework-core"},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}

	// Test HasDependency
	if !track.HasDependency("track-framework-core") {
		t.Error("expected track to have dependency on track-framework-core")
	}

	if track.HasDependency("track-nonexistent") {
		t.Error("expected track to not have dependency on track-nonexistent")
	}

	// Test AddDependency
	err = track.AddDependency("track-database")
	if err != nil {
		t.Errorf("failed to add dependency: %v", err)
	}

	if !track.HasDependency("track-database") {
		t.Error("expected track to have dependency after AddDependency")
	}

	// Test duplicate dependency
	err = track.AddDependency("track-database")
	if err == nil {
		t.Error("expected error when adding duplicate dependency")
	}

	// Test self-dependency
	err = track.AddDependency("track-plugin-system")
	if err == nil {
		t.Error("expected error when adding self-dependency")
	}

	// Test RemoveDependency
	err = track.RemoveDependency("track-database")
	if err != nil {
		t.Errorf("failed to remove dependency: %v", err)
	}

	if track.HasDependency("track-database") {
		t.Error("expected track to not have dependency after RemoveDependency")
	}

	// Test removing non-existent dependency
	err = track.RemoveDependency("track-nonexistent")
	if err == nil {
		t.Error("expected error when removing non-existent dependency")
	}
}

// TestTrackEntityITrackable tests ITrackable interface implementation
func TestTrackEntityITrackable(t *testing.T) {
	now := time.Now().UTC()

	track, err := task_manager.NewTrackEntity(
		"track-test",
		"roadmap-001",
		"Test Track",
		"Test",
		"in-progress",
		100,
		[]string{},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}

	// Verify it implements ITrackable
	var _ pluginsdk.ITrackable = track

	// Test ITrackable methods
	if track.GetStatus() != "in-progress" {
		t.Error("GetStatus() failed")
	}

	if track.GetProgress() != 0.5 {
		t.Error("GetProgress() failed for in-progress status")
	}

	if track.IsBlocked() {
		t.Error("IsBlocked() should return false for in-progress status")
	}

	blockReason := track.GetBlockReason()
	if blockReason != "" {
		t.Error("GetBlockReason() should return empty string for non-blocked track")
	}
}

// TestTrackEntityFields tests GetAllFields method
func TestTrackEntityFields(t *testing.T) {
	now := time.Now().UTC()

	track, err := task_manager.NewTrackEntity(
		"track-test",
		"roadmap-001",
		"Test Track",
		"Test Description",
		"in-progress",
		200,
		[]string{"track-dep"},
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}

	fields := track.GetAllFields()

	expectedFields := []string{
		"id", "roadmap_id", "title", "description",
		"status", "rank", "dependencies",
		"created_at", "updated_at", "progress", "is_blocked",
	}

	for _, field := range expectedFields {
		if fields[field] == nil && field != "is_blocked" {
			t.Errorf("field %q is nil", field)
		}
	}

	// Test GetField
	if track.GetField("id") != "track-test" {
		t.Error("GetField('id') mismatch")
	}

	if track.GetField("status") != "in-progress" {
		t.Error("GetField('status') mismatch")
	}
}

// TestTrackIDValidation tests track ID format validation
func TestTrackIDValidation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		id          string
		expectedErr bool
	}{
		{"valid simple", "track-core", false},
		{"valid with hyphens", "track-framework-core", false},
		{"valid with numbers", "track-api-v2", false},
		{"missing track prefix", "framework-core", true},
		{"no hyphen", "trackcore", true},
		{"uppercase", "Track-Core", true},
		{"underscore", "track_core", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := task_manager.NewTrackEntity(
				tt.id,
				"roadmap-001",
				"Test",
				"Test",
				"not-started",
				300,
				[]string{},
				now,
				now,
			)

			if tt.expectedErr && err == nil {
				t.Error("expected error for invalid ID, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error for valid ID: %v", err)
			}
		})
	}
}

// TestTrackEntityValidRanks tests all valid rank values
func TestTrackEntityValidRanks(t *testing.T) {
	now := time.Now().UTC()
	ranks := []int{100, 200, 300, 400}

	for _, rank := range ranks {
		t.Run(string(rune(rank)), func(t *testing.T) {
			track, err := task_manager.NewTrackEntity(
				"track-test",
				"roadmap-001",
				"Test",
				"Test",
				"not-started",
				rank,
				[]string{},
				now,
				now,
			)

			if err != nil {
				t.Errorf("failed to create track with rank %d: %v", rank, err)
			}
			if track == nil {
				t.Errorf("track is nil for rank %d", rank)
			}
		})
	}
}

// TestTrackEntityValidStatuses tests all valid status values
func TestTrackEntityValidStatuses(t *testing.T) {
	now := time.Now().UTC()
	statuses := []string{"not-started", "in-progress", "complete", "blocked", "waiting"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			track, err := task_manager.NewTrackEntity(
				"track-test",
				"roadmap-001",
				"Test",
				"Test",
				status,
				300,
				[]string{},
				now,
				now,
			)

			if err != nil {
				t.Errorf("failed to create track with status %q: %v", status, err)
			}
			if track == nil {
				t.Errorf("track is nil for status %q", status)
			}
		})
	}
}

package task_manager_e2e_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TrackTestSuite tests track CRUD and dependency commands
type TrackTestSuite struct {
	E2ETestSuite
}

func TestTrackSuite(t *testing.T) {
	suite.Run(t, new(TrackTestSuite))
}

// TestTrackCreate tests creating a track with required flags
func (s *TrackTestSuite) TestTrackCreate() {
	output, err := s.run("track", "create", "--title", "Backend API", "--description", "REST API implementation", "--rank", "100")
	s.requireSuccess(output, err, "failed to create track")

	// Extract track ID from output (format: "ID: XXX-track-X")
	trackID := s.parseID(output, "-track-")
	s.NotEmpty(trackID, "track ID should be extracted from output")
	s.Contains(trackID, "-track-", "track ID should have correct format")
}

// TestTrackCreateMinimal tests creating track with minimal required flags
func (s *TrackTestSuite) TestTrackCreateMinimal() {
	output, err := s.run("track", "create", "--title", "Test Track", "--description", "Test description")
	s.requireSuccess(output, err, "failed to create track with minimal flags")

	trackID := s.parseID(output, "-track-")
	s.NotEmpty(trackID, "track ID should be created")
}

// TestTrackList tests listing all tracks
func (s *TrackTestSuite) TestTrackList() {
	// Create a track first
	createOutput, err := s.run("track", "create", "--title", "Frontend UI", "--description", "React components")
	s.requireSuccess(createOutput, err, "failed to create track for listing test")

	trackID := s.parseID(createOutput, "-track-")

	// List tracks
	listOutput, err := s.run("track", "list")
	s.requireSuccess(listOutput, err, "failed to list tracks")

	// Verify the created track appears in the list
	s.Contains(listOutput, trackID, "created track should appear in list")
	s.Contains(listOutput, "Frontend UI", "track title should appear in list")
}

// TestTrackShow tests showing track details
func (s *TrackTestSuite) TestTrackShow() {
	// Create a track
	createOutput, err := s.run("track", "create", "--title", "Database Layer", "--description", "PostgreSQL schema and migrations")
	s.requireSuccess(createOutput, err, "failed to create track for show test")

	trackID := s.parseID(createOutput, "-track-")

	// Show track details
	showOutput, err := s.run("track", "show", trackID)
	s.requireSuccess(showOutput, err, "failed to show track")

	// Verify details
	s.Contains(showOutput, trackID, "track ID should be in output")
	s.Contains(showOutput, "Database Layer", "track title should be in output")
	s.Contains(showOutput, "PostgreSQL schema", "track description should be in output")
}

// TestTrackUpdate tests updating track fields
func (s *TrackTestSuite) TestTrackUpdate() {
	// Create a track
	createOutput, err := s.run("track", "create", "--title", "Original Title", "--description", "Original description")
	s.requireSuccess(createOutput, err, "failed to create track for update test")

	trackID := s.parseID(createOutput, "-track-")

	// Update the track
	updateOutput, err := s.run("track", "update", trackID, "--title", "Updated Title", "--description", "Updated description")
	s.requireSuccess(updateOutput, err, "failed to update track")

	// Verify the update
	showOutput, err := s.run("track", "show", trackID)
	s.requireSuccess(showOutput, err, "failed to show updated track")

	s.Contains(showOutput, "Updated Title", "track title should be updated")
	s.Contains(showOutput, "Updated description", "track description should be updated")
}

// TestTrackDelete tests deleting a track
func (s *TrackTestSuite) TestTrackDelete() {
	// Create a track
	createOutput, err := s.run("track", "create", "--title", "Temporary Track", "--description", "To be deleted")
	s.requireSuccess(createOutput, err, "failed to create track for deletion test")

	trackID := s.parseID(createOutput, "-track-")
	s.NotEmpty(trackID, "track ID should be extracted")

	// Delete the track
	deleteOutput, err := s.run("track", "delete", trackID, "--force")
	s.requireSuccess(deleteOutput, err, "failed to delete track")

	// Verify deletion was initiated successfully
	// Note: deletion behavior depends on implementation
	s.NotEmpty(deleteOutput, "delete command should produce output")
}

// TestTrackSetDependency tests setting a dependency between tracks
func (s *TrackTestSuite) TestTrackSetDependency() {
	// Create two tracks
	output1, err := s.run("track", "create", "--title", "Foundation Track", "--description", "Must be done first")
	s.requireSuccess(output1, err, "failed to create first track")
	trackID1 := s.parseID(output1, "-track-")

	output2, err := s.run("track", "create", "--title", "Dependent Track", "--description", "Depends on foundation")
	s.requireSuccess(output2, err, "failed to create second track")
	trackID2 := s.parseID(output2, "-track-")

	// Set dependency: track2 depends on track1
	depOutput, err := s.run("track", "add-dependency", trackID2, trackID1)
	s.requireSuccess(depOutput, err, "failed to set track dependency")

	// Verify dependency in track show
	showOutput, err := s.run("track", "show", trackID2)
	s.requireSuccess(showOutput, err, "failed to show track with dependency")

	s.Contains(showOutput, trackID1, "dependency should be visible in track details")
}

// TestTrackRemoveDependency tests removing a dependency between tracks
func (s *TrackTestSuite) TestTrackRemoveDependency() {
	// Create two tracks
	output1, err := s.run("track", "create", "--title", "Base Track", "--description", "Independent")
	s.requireSuccess(output1, err, "failed to create first track")
	trackID1 := s.parseID(output1, "-track-")

	output2, err := s.run("track", "create", "--title", "Related Track", "--description", "Depends on base")
	s.requireSuccess(output2, err, "failed to create second track")
	trackID2 := s.parseID(output2, "-track-")

	// Set dependency
	depOutput, err := s.run("track", "add-dependency", trackID2, trackID1)
	s.requireSuccess(depOutput, err, "failed to set dependency before removal")

	// Remove dependency
	removeOutput, err := s.run("track", "remove-dependency", trackID2, trackID1)
	s.requireSuccess(removeOutput, err, "failed to remove dependency")

	// Verify dependency is gone
	showOutput, err := s.run("track", "show", trackID2)
	s.requireSuccess(showOutput, err, "failed to show track after dependency removal")

	// The dependency should no longer be listed (or show "no dependencies")
	s.NotContains(showOutput, "depends on "+trackID1, "removed dependency should not appear")
}

// TestCircularDependencyPrevention tests that circular dependencies are prevented
func (s *TrackTestSuite) TestCircularDependencyPrevention() {
	// Create two tracks
	output1, err := s.run("track", "create", "--title", "Track A", "--description", "First track")
	s.requireSuccess(output1, err, "failed to create track A")
	trackIDA := s.parseID(output1, "-track-")

	output2, err := s.run("track", "create", "--title", "Track B", "--description", "Second track")
	s.requireSuccess(output2, err, "failed to create track B")
	trackIDB := s.parseID(output2, "-track-")

	// Set dependency: B depends on A
	depOutput, err := s.run("track", "add-dependency", trackIDB, trackIDA)
	s.requireSuccess(depOutput, err, "failed to set initial dependency")

	// Try to create circular dependency: A depends on B
	// This should fail because it would create a cycle: A -> B -> A
	circularOutput, err := s.run("track", "add-dependency", trackIDA, trackIDB)
	s.requireError(err, "should prevent circular dependency")
	s.Contains(circularOutput, "circular", "error message should mention circular dependency")
}

// TestTrackListWithStatus tests listing tracks filtered by status
func (s *TrackTestSuite) TestTrackListWithStatus() {
	// Create a track
	createOutput, err := s.run("track", "create", "--title", "Status Test Track", "--description", "For status filtering")
	s.requireSuccess(createOutput, err, "failed to create track for status test")

	trackID := s.parseID(createOutput, "-track-")

	// List tracks with status filter
	listOutput, err := s.run("track", "list", "--status", "not-started")
	s.requireSuccess(listOutput, err, "failed to list tracks with status filter")

	// Track should appear in the list
	s.Contains(listOutput, trackID, "newly created track should be in not-started status")
}

// TestTrackMultipleDependencies tests a track with multiple dependencies
func (s *TrackTestSuite) TestTrackMultipleDependencies() {
	// Create three tracks
	output1, err := s.run("track", "create", "--title", "Foundation 1", "--description", "First dependency")
	s.requireSuccess(output1, err, "failed to create foundation track 1")
	trackID1 := s.parseID(output1, "-track-")

	output2, err := s.run("track", "create", "--title", "Foundation 2", "--description", "Second dependency")
	s.requireSuccess(output2, err, "failed to create foundation track 2")
	trackID2 := s.parseID(output2, "-track-")

	output3, err := s.run("track", "create", "--title", "Dependent on Both", "--description", "Depends on both")
	s.requireSuccess(output3, err, "failed to create dependent track")
	trackID3 := s.parseID(output3, "-track-")

	// Set multiple dependencies
	dep1Output, err := s.run("track", "add-dependency", trackID3, trackID1)
	s.requireSuccess(dep1Output, err, "failed to set first dependency")

	dep2Output, err := s.run("track", "add-dependency", trackID3, trackID2)
	s.requireSuccess(dep2Output, err, "failed to set second dependency")

	// Verify both dependencies appear in show
	showOutput, err := s.run("track", "show", trackID3)
	s.requireSuccess(showOutput, err, "failed to show track with multiple dependencies")

	s.Contains(showOutput, trackID1, "first dependency should appear")
	s.Contains(showOutput, trackID2, "second dependency should appear")
}

// TestTrackWorkflowWithoutADR verifies that tracks can be created and used without ADRs (AC-564, AC-566)
func (s *TrackTestSuite) TestTrackWorkflowWithoutADR() {
	// Create a track without ADR
	trackOutput, err := s.run("track", "create", "--title", "No ADR Track", "--description", "Track without ADR", "--rank", "100")
	s.requireSuccess(trackOutput, err, "failed to create track without ADR")
	trackID := s.parseID(trackOutput, "-track-")
	s.NotEmpty(trackID, "track ID should be created")

	// Verify track can be shown
	showOutput, err := s.run("track", "show", trackID)
	s.requireSuccess(showOutput, err, "should be able to show track without ADR")
	s.Contains(showOutput, "No ADR Track", "track title should appear")

	// Update track status without ADR (no validation errors expected - AC-566)
	updateOutput, err := s.run("track", "update", trackID, "--status", "in-progress")
	s.requireSuccess(updateOutput, err, "should be able to update track status without ADR")

	// Create task in track without ADR
	taskOutput, err := s.run("task", "create", "--track", trackID, "--title", "Task without ADR")
	s.requireSuccess(taskOutput, err, "should be able to create task in track without ADR")
	taskID := s.parseID(taskOutput, "-task-")
	s.NotEmpty(taskID, "task should be created in track without ADR")

	// Update task status without track having ADR
	taskUpdateOutput, err := s.run("task", "update", taskID, "--status", "in-progress")
	s.requireSuccess(taskUpdateOutput, err, "should be able to update task status without track ADR")

	// Complete track without ADR (no validation errors expected)
	completeOutput, err := s.run("track", "update", trackID, "--status", "complete")
	s.requireSuccess(completeOutput, err, "should be able to complete track without ADR")
}

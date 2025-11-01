package task_manager_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// ============================================================================
// TrackCreateCommand Tests
// ============================================================================

func TestTrackCreateCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap first in project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create track command
	cmd := &task_manager.TrackCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--id", "track-plugin-system",
		"--title", "Plugin System",
		"--description", "Implement plugin architecture",
		"--priority", "high",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify output
	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Track created successfully") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify track was saved
	track, err := repo.GetTrack(ctx, "track-plugin-system")
	if err != nil {
		t.Errorf("failed to retrieve saved track: %v", err)
	}
	if track.Title != "Plugin System" {
		t.Errorf("expected title 'Plugin System', got '%s'", track.Title)
	}
}

func TestTrackCreateCommand_MissingId(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	cmd := &task_manager.TrackCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err := cmd.Execute(ctx, cmdCtx, []string{
		"--title", "Plugin System",
	})

	if err == nil || !strings.Contains(err.Error(), "--id and --title are required") {
		t.Errorf("expected error about missing id, got: %v", err)
	}
}

func TestTrackCreateCommand_NoActiveRoadmap(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	cmd := &task_manager.TrackCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err := cmd.Execute(ctx, cmdCtx, []string{
		"--id", "track-plugin-system",
		"--title", "Plugin System",
	})

	if err == nil || !strings.Contains(err.Error(), "no active roadmap found") {
		t.Errorf("expected error about no roadmap, got: %v", err)
	}
}

// ============================================================================
// TrackListCommand Tests
// ============================================================================

func TestTrackListCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap and tracks in project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create a track
	track1, _ := task_manager.NewTrackEntity(
		"track-plugin-system",
		roadmap.ID,
		"Plugin System",
		"",
		"in-progress",
		"high",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track1)

	cmd := &task_manager.TrackListCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "track-plugin-system") {
		t.Errorf("expected track ID in output, got: %s", output)
	}
}

// ============================================================================
// TrackShowCommand Tests
// ============================================================================

func TestTrackShowCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap and track in project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, _ := task_manager.NewTrackEntity(
		"track-plugin-system",
		roadmap.ID,
		"Plugin System",
		"Implement plugin architecture",
		"in-progress",
		"high",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track)

	cmd := &task_manager.TrackShowCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"track-plugin-system"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "track-plugin-system") {
		t.Errorf("expected track ID in output, got: %s", output)
	}
	if !strings.Contains(output, "Plugin System") {
		t.Errorf("expected title in output, got: %s", output)
	}
}

// ============================================================================
// TrackUpdateCommand Tests
// ============================================================================

func TestTrackUpdateCommand_UpdateTitle(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap and track in project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, _ := task_manager.NewTrackEntity(
		"track-test",
		roadmap.ID,
		"Old Title",
		"",
		"not-started",
		"medium",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track)

	cmd := &task_manager.TrackUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-test",
		"--title", "New Title",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify track was updated
	updated, err := repo.GetTrack(ctx, "track-test")
	if err != nil {
		t.Errorf("failed to retrieve updated track: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", updated.Title)
	}
}

// ============================================================================
// TrackDeleteCommand Tests
// ============================================================================

func TestTrackDeleteCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap and track in project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, _ := task_manager.NewTrackEntity(
		"track-test",
		roadmap.ID,
		"Test Track",
		"",
		"not-started",
		"medium",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track)

	cmd := &task_manager.TrackDeleteCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-test",
		"--force",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify track was deleted
	_, err = repo.GetTrack(ctx, "track-test")
	if err == nil {
		t.Error("track should not exist after deletion")
	}
}

// ============================================================================
// TrackAddDependencyCommand Tests
// ============================================================================

func TestTrackAddDependencyCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap and tracks in project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track1, _ := task_manager.NewTrackEntity(
		"track-a",
		roadmap.ID,
		"Track A",
		"",
		"not-started",
		"medium",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track1)

	track2, _ := task_manager.NewTrackEntity(
		"track-b",
		roadmap.ID,
		"Track B",
		"",
		"not-started",
		"medium",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track2)

	cmd := &task_manager.TrackAddDependencyCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-a",
		"track-b",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify dependency was added
	deps, err := repo.GetTrackDependencies(ctx, "track-a")
	if err != nil {
		t.Errorf("failed to get dependencies: %v", err)
	}
	if len(deps) != 1 || deps[0] != "track-b" {
		t.Errorf("expected dependencies ['track-b'], got %v", deps)
	}
}

// ============================================================================
// TrackRemoveDependencyCommand Tests
// ============================================================================

func TestTrackRemoveDependencyCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap and tracks in project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track1, _ := task_manager.NewTrackEntity(
		"track-a",
		roadmap.ID,
		"Track A",
		"",
		"not-started",
		"medium",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track1)

	track2, _ := task_manager.NewTrackEntity(
		"track-b",
		roadmap.ID,
		"Track B",
		"",
		"not-started",
		"medium",
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveTrack(ctx, track2)

	// Add dependency first
	repo.AddTrackDependency(ctx, "track-a", "track-b")

	cmd := &task_manager.TrackRemoveDependencyCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-a",
		"track-b",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify dependency was removed
	deps, err := repo.GetTrackDependencies(ctx, "track-a")
	if err != nil {
		t.Errorf("failed to get dependencies: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected no dependencies, got %v", deps)
	}
}

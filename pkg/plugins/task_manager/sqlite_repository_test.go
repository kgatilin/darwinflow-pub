package task_manager_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// Helper to create a test database
func createTestDB(t *testing.T) *sql.DB {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Initialize schema
	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	return db
}

// Helper to create a test logger
func createTestLogger() pluginsdk.Logger {
	return &testLogger{}
}

type testLogger struct{}

func (l *testLogger) Debug(msg string, fields ...interface{})   {}
func (l *testLogger) Info(msg string, fields ...interface{})    {}
func (l *testLogger) Warn(msg string, fields ...interface{})    {}
func (l *testLogger) Error(msg string, fields ...interface{})   {}
func (l *testLogger) WithFields(fields ...interface{}) pluginsdk.Logger { return l }

// ============================================================================
// Roadmap Tests
// ============================================================================

func TestSaveAndGetRoadmap(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create a roadmap
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-1",
		"Build the best system",
		"Deliver on time and quality",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap entity: %v", err)
	}

	// Save roadmap
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Get roadmap
	retrieved, err := repo.GetRoadmap(ctx, "roadmap-1")
	if err != nil {
		t.Fatalf("failed to get roadmap: %v", err)
	}

	if retrieved.ID != roadmap.ID {
		t.Errorf("expected roadmap ID %s, got %s", roadmap.ID, retrieved.ID)
	}
	if retrieved.Vision != roadmap.Vision {
		t.Errorf("expected vision %s, got %s", roadmap.Vision, retrieved.Vision)
	}
}

func TestSaveRoadmapDuplicate(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())

	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Try to save duplicate
	if err := repo.SaveRoadmap(ctx, roadmap); err == nil {
		t.Error("expected error when saving duplicate roadmap")
	} else if !errors.Is(err, pluginsdk.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got: %v", err)
	}
}

func TestGetRoadmapNotFound(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	_, err := repo.GetRoadmap(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent roadmap")
	} else if !errors.Is(err, pluginsdk.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestUpdateRoadmap(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	// Update roadmap
	roadmap.Vision = "new vision"
	roadmap.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to update roadmap: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetRoadmap(ctx, "roadmap-1")
	if retrieved.Vision != "new vision" {
		t.Errorf("expected vision to be updated, got %s", retrieved.Vision)
	}
}

func TestGetActiveRoadmap(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	now := time.Now().UTC()

	// Create first roadmap
	roadmap1, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision1", "criteria1", now, now)
	repo.SaveRoadmap(ctx, roadmap1)

	time.Sleep(10 * time.Millisecond)

	// Create second roadmap (more recent)
	roadmap2, _ := task_manager.NewRoadmapEntity("roadmap-2", "vision2", "criteria2", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap2)

	// Get active roadmap should return the most recent one
	active, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		t.Fatalf("failed to get active roadmap: %v", err)
	}

	if active.ID != "roadmap-2" {
		t.Errorf("expected roadmap-2, got %s", active.ID)
	}
}

// ============================================================================
// Track Tests
// ============================================================================

func TestSaveAndGetTrack(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create roadmap first
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	// Create track
	track, err := task_manager.NewTrackEntity(
		"track-core",
		"roadmap-1",
		"Core Features",
		"Essential features",
		"not-started",
		200,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track entity: %v", err)
	}

	// Save track
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Get track
	retrieved, err := repo.GetTrack(ctx, "track-core")
	if err != nil {
		t.Fatalf("failed to get track: %v", err)
	}

	if retrieved.ID != track.ID {
		t.Errorf("expected track ID %s, got %s", track.ID, retrieved.ID)
	}
}

func TestListTracks(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create roadmap
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	// Create tracks
	for i := 1; i <= 3; i++ {
		id := "track-" + string(rune(48+i))
		track, _ := task_manager.NewTrackEntity(id, "roadmap-1", "Track "+string(rune(48+i)), "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
		repo.SaveTrack(ctx, track)
	}

	// List all tracks
	tracks, err := repo.ListTracks(ctx, "roadmap-1", task_manager.TrackFilters{})
	if err != nil {
		t.Fatalf("failed to list tracks: %v", err)
	}

	if len(tracks) != 3 {
		t.Errorf("expected 3 tracks, got %d", len(tracks))
	}
}

func TestListTracksWithFilters(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create roadmap
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	// Create tracks with different statuses
	track1, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track 1", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	track2, _ := task_manager.NewTrackEntity("track-2", "roadmap-1", "Track 2", "", "in-progress", 200, []string{}, time.Now().UTC(), time.Now().UTC())

	repo.SaveTrack(ctx, track1)
	repo.SaveTrack(ctx, track2)

	// Filter by status
	tracks, err := repo.ListTracks(ctx, "roadmap-1", task_manager.TrackFilters{Status: []string{"in-progress"}})
	if err != nil {
		t.Fatalf("failed to list tracks: %v", err)
	}

	if len(tracks) != 1 || tracks[0].Status != "in-progress" {
		t.Errorf("expected 1 in-progress track, got %d", len(tracks))
	}
}

func TestTrackDependencies(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track1, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track 1", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	track2, _ := task_manager.NewTrackEntity("track-2", "roadmap-1", "Track 2", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())

	repo.SaveTrack(ctx, track1)
	repo.SaveTrack(ctx, track2)

	// Add dependency
	if err := repo.AddTrackDependency(ctx, "track-2", "track-1"); err != nil {
		t.Fatalf("failed to add dependency: %v", err)
	}

	// Get dependencies
	deps, err := repo.GetTrackDependencies(ctx, "track-2")
	if err != nil {
		t.Fatalf("failed to get dependencies: %v", err)
	}

	if len(deps) != 1 || deps[0] != "track-1" {
		t.Errorf("expected track-1 dependency, got %v", deps)
	}

	// Remove dependency
	if err := repo.RemoveTrackDependency(ctx, "track-2", "track-1"); err != nil {
		t.Fatalf("failed to remove dependency: %v", err)
	}

	deps, _ = repo.GetTrackDependencies(ctx, "track-2")
	if len(deps) != 0 {
		t.Errorf("expected no dependencies, got %v", deps)
	}
}

func TestValidateNoCycles(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track1, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track 1", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	track2, _ := task_manager.NewTrackEntity("track-2", "roadmap-1", "Track 2", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	track3, _ := task_manager.NewTrackEntity("track-3", "roadmap-1", "Track 3", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())

	repo.SaveTrack(ctx, track1)
	repo.SaveTrack(ctx, track2)
	repo.SaveTrack(ctx, track3)

	// Create a cycle: 1 -> 2 -> 3 -> 1
	repo.AddTrackDependency(ctx, "track-2", "track-1")
	repo.AddTrackDependency(ctx, "track-3", "track-2")
	repo.AddTrackDependency(ctx, "track-1", "track-3")

	// Validate should detect cycle
	err := repo.ValidateNoCycles(ctx, "track-1")
	if err == nil {
		t.Error("expected error for cycle detection")
	} else if !errors.Is(err, pluginsdk.ErrInvalidArgument) {
		t.Errorf("expected ErrInvalidArgument, got: %v", err)
	}
}

// ============================================================================
// Task Tests
// ============================================================================

func TestSaveAndGetTask(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create and save task
	task := task_manager.NewTaskEntity("task-1", "track-1", "Implement feature", "Do something", "todo", 200, "feat/impl", time.Now().UTC(), time.Now().UTC())

	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Get task
	retrieved, err := repo.GetTask(ctx, "task-1")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if retrieved.ID != task.ID || retrieved.Title != task.Title {
		t.Errorf("task mismatch")
	}
	if retrieved.Branch != "feat/impl" {
		t.Errorf("expected branch feat/impl, got %s", retrieved.Branch)
	}
}

func TestListTasks(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create multiple tasks
	for i := 1; i <= 3; i++ {
		id := "task-" + string(rune(48+i))
		task := task_manager.NewTaskEntity(id, "track-1", "Task "+string(rune(48+i)), "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
		repo.SaveTask(ctx, task)
	}

	// List tasks
	tasks, err := repo.ListTasks(ctx, task_manager.TaskFilters{})
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestListTasksWithFilters(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create tasks with different statuses
	task1 := task_manager.NewTaskEntity("task-1", "track-1", "Task 1", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	task2 := task_manager.NewTaskEntity("task-2", "track-1", "Task 2", "", "done", 200, "", time.Now().UTC(), time.Now().UTC())

	repo.SaveTask(ctx, task1)
	repo.SaveTask(ctx, task2)

	// Filter by status
	tasks, err := repo.ListTasks(ctx, task_manager.TaskFilters{Status: []string{"done"}})
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 1 || tasks[0].Status != "done" {
		t.Errorf("expected 1 done task, got %d", len(tasks))
	}
}

func TestMoveTaskToTrack(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track1, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track 1", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	track2, _ := task_manager.NewTrackEntity("track-2", "roadmap-1", "Track 2", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())

	repo.SaveTrack(ctx, track1)
	repo.SaveTrack(ctx, track2)

	task := task_manager.NewTaskEntity("task-1", "track-1", "Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	// Move task to track-2
	if err := repo.MoveTaskToTrack(ctx, "task-1", "track-2"); err != nil {
		t.Fatalf("failed to move task: %v", err)
	}

	// Verify move
	updated, _ := repo.GetTask(ctx, "task-1")
	if updated.TrackID != "track-2" {
		t.Errorf("expected track-2, got %s", updated.TrackID)
	}
}

// ============================================================================
// Iteration Tests
// ============================================================================

func TestSaveAndGetIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iteration
	iteration, err := task_manager.NewIterationEntity(
		1,
		"Sprint 1",
		"Build MVP",
		"",
		[]string{},
		"planned",
		500,
		time.Time{},
		time.Time{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}

	// Save iteration
	if err := repo.SaveIteration(ctx, iteration); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Get iteration
	retrieved, err := repo.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration: %v", err)
	}

	if retrieved.Number != 1 || retrieved.Name != "Sprint 1" {
		t.Errorf("iteration mismatch")
	}
}

func TestIterationTaskManagement(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup iteration
	iteration, _ := task_manager.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveIteration(ctx, iteration)

	// Setup task
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("task-1", "track-1", "Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	// Add task to iteration
	if err := repo.AddTaskToIteration(ctx, 1, "task-1"); err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	// Get tasks
	tasks, err := repo.GetIterationTasks(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration tasks: %v", err)
	}

	if len(tasks) != 1 || tasks[0].ID != "task-1" {
		t.Errorf("expected task-1, got %v", tasks)
	}

	// Remove task
	if err := repo.RemoveTaskFromIteration(ctx, 1, "task-1"); err != nil {
		t.Fatalf("failed to remove task: %v", err)
	}

	tasks, _ = repo.GetIterationTasks(ctx, 1)
	if len(tasks) != 0 {
		t.Errorf("expected no tasks, got %d", len(tasks))
	}
}

func TestStartAndCompleteIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iteration
	iteration, _ := task_manager.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveIteration(ctx, iteration)

	// Start iteration
	if err := repo.StartIteration(ctx, 1); err != nil {
		t.Fatalf("failed to start iteration: %v", err)
	}

	// Verify status
	updated, _ := repo.GetIteration(ctx, 1)
	if updated.Status != "current" {
		t.Errorf("expected current, got %s", updated.Status)
	}
	if updated.StartedAt == nil {
		t.Error("expected started_at to be set")
	}

	// Complete iteration
	if err := repo.CompleteIteration(ctx, 1); err != nil {
		t.Fatalf("failed to complete iteration: %v", err)
	}

	// Verify completion
	completed, _ := repo.GetIteration(ctx, 1)
	if completed.Status != "complete" {
		t.Errorf("expected complete, got %s", completed.Status)
	}
	if completed.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestGetCurrentIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iterations
	iter1, _ := task_manager.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	iter2, _ := task_manager.NewIterationEntity(2, "Sprint 2", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())

	repo.SaveIteration(ctx, iter1)
	repo.SaveIteration(ctx, iter2)

	// Start iteration 1
	repo.StartIteration(ctx, 1)

	// Get current
	current, err := repo.GetCurrentIteration(ctx)
	if err != nil {
		t.Fatalf("failed to get current iteration: %v", err)
	}

	if current.Number != 1 {
		t.Errorf("expected iteration 1, got %d", current.Number)
	}
}

func TestListIterations(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create multiple iterations
	for i := 1; i <= 3; i++ {
		iter, _ := task_manager.NewIterationEntity(i, "Sprint "+string(rune(48+i)), "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
		repo.SaveIteration(ctx, iter)
	}

	// List all
	iterations, err := repo.ListIterations(ctx)
	if err != nil {
		t.Fatalf("failed to list iterations: %v", err)
	}

	if len(iterations) != 3 {
		t.Errorf("expected 3 iterations, got %d", len(iterations))
	}
}

// ============================================================================
// Error Cases
// ============================================================================

func TestAddDependencyToNonexistentTrack(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	err := repo.AddTrackDependency(ctx, "nonexistent", "also-nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent track")
	} else if !errors.Is(err, pluginsdk.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSelfDependency(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Try self dependency
	err := repo.AddTrackDependency(ctx, "track-1", "track-1")
	if err == nil {
		t.Error("expected error for self dependency")
	} else if !errors.Is(err, pluginsdk.ErrInvalidArgument) {
		t.Errorf("expected ErrInvalidArgument, got: %v", err)
	}
}

func TestAddTaskToNonexistentIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	err := repo.AddTaskToIteration(ctx, 999, "task-1")
	if err == nil {
		t.Error("expected error for nonexistent iteration")
	} else if !errors.Is(err, pluginsdk.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// ============================================================================
// Aggregate Query Tests
// ============================================================================

func TestGetRoadmapWithTracks(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	for i := 1; i <= 2; i++ {
		track, _ := task_manager.NewTrackEntity(
			"track-"+string(rune(48+i)),
			"roadmap-1",
			"Track "+string(rune(48+i)),
			"",
			"not-started",
			200,
			[]string{},
			time.Now().UTC(),
			time.Now().UTC(),
		)
		repo.SaveTrack(ctx, track)
	}

	// Get roadmap with tracks
	retrieved, err := repo.GetRoadmapWithTracks(ctx, "roadmap-1")
	if err != nil {
		t.Fatalf("failed to get roadmap with tracks: %v", err)
	}

	if retrieved.ID != "roadmap-1" {
		t.Errorf("expected roadmap-1, got %s", retrieved.ID)
	}
}

func TestGetTrackWithTasks(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup
	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-1", "roadmap-1", "Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create tasks
	for i := 1; i <= 2; i++ {
		task := task_manager.NewTaskEntity(
			"task-"+string(rune(48+i)),
			"track-1",
			"Task "+string(rune(48+i)),
			"",
			"todo",
			200,
			"",
			time.Now().UTC(),
			time.Now().UTC(),
		)
		repo.SaveTask(ctx, task)
	}

	// Get track with tasks
	retrieved, err := repo.GetTrackWithTasks(ctx, "track-1")
	if err != nil {
		t.Fatalf("failed to get track with tasks: %v", err)
	}

	if retrieved.ID != "track-1" {
		t.Errorf("expected track-1, got %s", retrieved.ID)
	}
}

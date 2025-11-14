package persistence_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/infrastructure/persistence"
)

// ============================================================================
// Iteration Tests
// ============================================================================

func TestSaveAndGetIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iteration
	iteration, err := entities.NewIterationEntity(
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

	roadmapRepo := persistence.NewSQLiteRoadmapRepository(db, createTestLogger())
	trackRepo := persistence.NewSQLiteTrackRepository(db, createTestLogger())
	taskRepo := persistence.NewSQLiteTaskRepository(db, createTestLogger())
	iterationRepo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Setup iteration
	iteration, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	iterationRepo.SaveIteration(ctx, iteration)

	// Setup task
	roadmap, _ := entities.NewRoadmapEntity("roadmap-1", "vision", "criteria", time.Now().UTC(), time.Now().UTC())
	roadmapRepo.SaveRoadmap(ctx, roadmap)

	track, _ := entities.NewTrackEntity("track-1", "roadmap-1", "Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	trackRepo.SaveTrack(ctx, track)

	task, _ := entities.NewTaskEntity("task-1", "track-1", "Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	taskRepo.SaveTask(ctx, task)

	// Add task to iteration
	if err := iterationRepo.AddTaskToIteration(ctx, 1, "task-1"); err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	// Get tasks
	tasks, err := iterationRepo.GetIterationTasks(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration tasks: %v", err)
	}

	if len(tasks) != 1 || tasks[0].ID != "task-1" {
		t.Errorf("expected task-1, got %v", tasks)
	}

	// Remove task
	if err := iterationRepo.RemoveTaskFromIteration(ctx, 1, "task-1"); err != nil {
		t.Fatalf("failed to remove task: %v", err)
	}

	tasks, _ = iterationRepo.GetIterationTasks(ctx, 1)
	if len(tasks) != 0 {
		t.Errorf("expected no tasks, got %d", len(tasks))
	}
}

func TestStartAndCompleteIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iteration
	iteration, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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

	repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iterations
	iter1, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	iter2, _ := entities.NewIterationEntity(2, "Sprint 2", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())

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

func TestGetIterationTasksWithWarnings_MissingTask(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	roadmapRepo := persistence.NewSQLiteRoadmapRepository(db, createTestLogger())
	trackRepo := persistence.NewSQLiteTrackRepository(db, createTestLogger())
	taskRepo := persistence.NewSQLiteTaskRepository(db, createTestLogger())
	iterationRepo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Create roadmap and track
	roadmap, _ := entities.NewRoadmapEntity("roadmap-1", "Vision", "Success criteria", time.Now().UTC(), time.Now().UTC())
	roadmapRepo.SaveRoadmap(ctx, roadmap)

	track, _ := entities.NewTrackEntity("track-1", "roadmap-1", "Track 1", "Description", "not-started", 500, []string{}, time.Now().UTC(), time.Now().UTC())
	trackRepo.SaveTrack(ctx, track)

	// Create three tasks
	task1, _ := entities.NewTaskEntity("task-1", "track-1", "Task 1", "Description", "todo", 500, "", time.Now().UTC(), time.Now().UTC())
	task2, _ := entities.NewTaskEntity("task-2", "track-1", "Task 2", "Description", "todo", 500, "", time.Now().UTC(), time.Now().UTC())
	task3, _ := entities.NewTaskEntity("task-3", "track-1", "Task 3", "Description", "todo", 500, "", time.Now().UTC(), time.Now().UTC())
	taskRepo.SaveTask(ctx, task1)
	taskRepo.SaveTask(ctx, task2)
	taskRepo.SaveTask(ctx, task3)

	// Create iteration and add all three tasks
	iteration, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	iterationRepo.SaveIteration(ctx, iteration)
	iterationRepo.AddTaskToIteration(ctx, 1, "task-1")
	iterationRepo.AddTaskToIteration(ctx, 1, "task-2")
	iterationRepo.AddTaskToIteration(ctx, 1, "task-3")

	// Delete task-2 directly from database to simulate a missing task
	_, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", "task-2")
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	// Call GetIterationTasksWithWarnings
	tasks, missingTaskIDs, err := iterationRepo.GetIterationTasksWithWarnings(ctx, 1)
	if err != nil {
		t.Fatalf("GetIterationTasksWithWarnings failed: %v", err)
	}

	// Verify results
	if len(tasks) != 2 {
		t.Errorf("expected 2 found tasks, got %d", len(tasks))
	}

	if len(missingTaskIDs) != 1 {
		t.Errorf("expected 1 missing task ID, got %d", len(missingTaskIDs))
	}

	if len(missingTaskIDs) > 0 && missingTaskIDs[0] != "task-2" {
		t.Errorf("expected missing task 'task-2', got '%s'", missingTaskIDs[0])
	}

	// Verify the found tasks are correct
	foundIDs := make(map[string]bool)
	for _, task := range tasks {
		foundIDs[task.ID] = true
	}

	if !foundIDs["task-1"] || !foundIDs["task-3"] {
		t.Errorf("expected to find task-1 and task-3, got: %v", foundIDs)
	}

	if foundIDs["task-2"] {
		t.Error("task-2 should not be in found tasks")
	}
}

func TestListIterations(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Create multiple iterations
	for i := 1; i <= 3; i++ {
		iter, _ := entities.NewIterationEntity(i, "Sprint "+string(rune(48+i)), "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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

func TestAddTaskToNonexistentIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	err := repo.AddTaskToIteration(ctx, 999, "task-1")
	if err == nil {
		t.Error("expected error for nonexistent iteration")
	} else if !errors.Is(err, pluginsdk.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestUpdateIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iteration
	iteration, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveIteration(ctx, iteration)

	// Update iteration
	iteration.Name = "Sprint 1 - Updated"
	iteration.Goal = "New Goal"
	iteration.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateIteration(ctx, iteration); err != nil {
		t.Fatalf("failed to update iteration: %v", err)
	}

	// Verify update
	retrieved, _ := repo.GetIteration(ctx, 1)
	if retrieved.Name != "Sprint 1 - Updated" {
		t.Errorf("expected name to be updated, got %s", retrieved.Name)
	}
	if retrieved.Goal != "New Goal" {
		t.Errorf("expected goal to be updated, got %s", retrieved.Goal)
	}
}

func TestDeleteIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())
	ctx := context.Background()

	// Create iteration
	iteration, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveIteration(ctx, iteration)

	// Delete iteration
	if err := repo.DeleteIteration(ctx, 1); err != nil {
		t.Fatalf("failed to delete iteration: %v", err)
	}

	// Verify deletion
	_, err := repo.GetIteration(ctx, 1)
	if !errors.Is(err, pluginsdk.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetNextPlannedIteration(t *testing.T) {
	ctx := context.Background()

	t.Run("returns first planned iteration ordered by rank", func(t *testing.T) {
		db := createTestDB(t)
		defer db.Close()
		repo := persistence.NewSQLiteIterationRepository(db, createTestLogger())

		// Create multiple planned iterations with different ranks
		iter1, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 300, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
		iter2, _ := entities.NewIterationEntity(2, "Sprint 2", "Goal 2", "", []string{}, "planned", 100, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
		iter3, _ := entities.NewIterationEntity(3, "Sprint 3", "Goal 3", "", []string{}, "planned", 200, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())

		repo.SaveIteration(ctx, iter1)
		repo.SaveIteration(ctx, iter2)
		repo.SaveIteration(ctx, iter3)

		// Get next planned iteration
		next, err := repo.GetNextPlannedIteration(ctx)
		if err != nil {
			t.Fatalf("failed to get next planned iteration: %v", err)
		}

		// Should return iteration 2 (lowest rank = 100)
		if next.Number != 2 {
			t.Errorf("expected iteration 2, got %d", next.Number)
		}
		if next.Name != "Sprint 2" {
			t.Errorf("expected 'Sprint 2', got %s", next.Name)
		}
	})

	t.Run("returns ErrNotFound when no planned iterations", func(t *testing.T) {
		// Create a new DB for this test
		db2 := createTestDB(t)
		defer db2.Close()
		repo2 := persistence.NewSQLiteIterationRepository(db2, createTestLogger())

		// Create only completed iterations
		iter1, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "complete", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
		repo2.SaveIteration(ctx, iter1)

		// Try to get next planned
		_, err := repo2.GetNextPlannedIteration(ctx)
		if !errors.Is(err, pluginsdk.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("ignores current and complete iterations", func(t *testing.T) {
		// Create a new DB for this test
		db3 := createTestDB(t)
		defer db3.Close()
		repo3 := persistence.NewSQLiteIterationRepository(db3, createTestLogger())

		// Create iterations with different statuses
		iter1, _ := entities.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "current", 100, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
		iter2, _ := entities.NewIterationEntity(2, "Sprint 2", "Goal 2", "", []string{}, "complete", 200, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
		iter3, _ := entities.NewIterationEntity(3, "Sprint 3", "Goal 3", "", []string{}, "planned", 300, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())

		repo3.SaveIteration(ctx, iter1)
		repo3.SaveIteration(ctx, iter2)
		repo3.SaveIteration(ctx, iter3)

		// Get next planned
		next, err := repo3.GetNextPlannedIteration(ctx)
		if err != nil {
			t.Fatalf("failed to get next planned iteration: %v", err)
		}

		// Should return only the planned iteration
		if next.Number != 3 {
			t.Errorf("expected iteration 3, got %d", next.Number)
		}
	})
}


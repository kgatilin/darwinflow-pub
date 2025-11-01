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
// IterationCreateCommand Tests
// ============================================================================

func TestIterationCreateCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.IterationCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(context.Background(), cmdCtx, []string{
		"--name", "Foundation Sprint",
		"--goal", "Complete view-based analysis",
		"--deliverable", "Plugin-agnostic framework",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify output
	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Iteration created successfully") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Number:       1") {
		t.Errorf("expected iteration number 1, got: %s", output)
	}
}

func TestIterationCreateCommand_AutoIncrementNumber(t *testing.T) {
	tmpDir := t.TempDir()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	ctx := context.Background()

	// Create project command to set up default project
	projectCmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	projectCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}
	if err := projectCmd.Execute(ctx, projectCtx, []string{"default"}); err != nil {
		t.Fatalf("failed to create default project: %v", err)
	}

	// Create first iteration using command
	cmd1 := &task_manager.IterationCreateCommand{Plugin: plugin}
	cmdCtx1 := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd1.Execute(ctx, cmdCtx1, []string{
		"--name", "Sprint 1",
		"--goal", "Goal 1",
	})
	if err != nil {
		t.Fatalf("failed to create first iteration: %v", err)
	}

	// Create second iteration via command
	cmd2 := &task_manager.IterationCreateCommand{Plugin: plugin}
	cmdCtx2 := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd2.Execute(ctx, cmdCtx2, []string{
		"--name", "Sprint 2",
		"--goal", "Goal 2",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify output shows number 2
	output := cmdCtx2.stdout.String()
	if !strings.Contains(output, "Number:       2") {
		t.Errorf("expected iteration number 2, got: %s", output)
	}
}

func TestIterationCreateCommand_MissingName(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.IterationCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(context.Background(), cmdCtx, []string{
		"--goal", "Goal",
	})
	if err == nil {
		t.Error("expected error for missing --name flag")
	}
}

func TestIterationCreateCommand_MissingGoal(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.IterationCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(context.Background(), cmdCtx, []string{
		"--name", "Sprint 1",
	})
	if err == nil {
		t.Error("expected error for missing --goal flag")
	}
}

// ============================================================================
// IterationListCommand Tests
// ============================================================================

func TestIterationListCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create iterations
	iter1, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter1); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// List iterations
	cmd := &task_manager.IterationListCommand{Plugin: plugin}
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
	if !strings.Contains(output, "Sprint 1") {
		t.Errorf("expected 'Sprint 1' in output, got: %s", output)
	}
}

func TestIterationListCommand_NoIterations(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.IterationListCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(context.Background(), cmdCtx, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "No iterations found") {
		t.Errorf("expected 'No iterations found', got: %s", output)
	}
}

// ============================================================================
// IterationShowCommand Tests
// ============================================================================

func TestIterationShowCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create roadmap
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create track
	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", "medium", []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create task
	task := task_manager.NewTaskEntity("task-fc-001", "track-test", "Test Task", "Description", "todo", "medium", "", time.Now().UTC(), time.Now().UTC())
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}
	if err := repo.AddTaskToIteration(ctx, 1, "task-fc-001"); err != nil {
		t.Fatalf("failed to add task to iteration: %v", err)
	}

	// Show iteration
	cmd := &task_manager.IterationShowCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Sprint 1") {
		t.Errorf("expected 'Sprint 1' in output, got: %s", output)
	}
	if !strings.Contains(output, "Goal 1") {
		t.Errorf("expected 'Goal 1' in output, got: %s", output)
	}
}

func TestIterationShowCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.IterationShowCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(context.Background(), cmdCtx, []string{"999"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found' message, got: %s", output)
	}
}

// ============================================================================
// IterationCurrentCommand Tests
// ============================================================================

func TestIterationCurrentCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Start iteration
	if err := repo.StartIteration(ctx, 1); err != nil {
		t.Fatalf("failed to start iteration: %v", err)
	}

	// Get current
	cmd := &task_manager.IterationCurrentCommand{Plugin: plugin}
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
	if !strings.Contains(output, "Current Iteration") {
		t.Errorf("expected 'Current Iteration', got: %s", output)
	}
}

func TestIterationCurrentCommand_NoCurrentIteration(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.IterationCurrentCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(context.Background(), cmdCtx, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "No current iteration") {
		t.Errorf("expected 'No current iteration', got: %s", output)
	}
}

// ============================================================================
// IterationUpdateCommand Tests
// ============================================================================

func TestIterationUpdateCommand_UpdateName(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Update iteration
	cmd := &task_manager.IterationUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1", "--name", "Updated Sprint"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Updated Sprint") {
		t.Errorf("expected 'Updated Sprint' in output, got: %s", output)
	}

	// Verify in database
	updated, err := repo.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration: %v", err)
	}
	if updated.Name != "Updated Sprint" {
		t.Errorf("expected name 'Updated Sprint', got '%s'", updated.Name)
	}
}

func TestIterationUpdateCommand_NoFlags(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Try to update without flags
	cmd := &task_manager.IterationUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1"})
	if err == nil {
		t.Error("expected error when no flags provided")
	}
}

// ============================================================================
// IterationDeleteCommand Tests
// ============================================================================

func TestIterationDeleteCommand_ForceDelete(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Delete iteration
	cmd := &task_manager.IterationDeleteCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1", "--force"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "deleted successfully") {
		t.Errorf("expected 'deleted successfully', got: %s", output)
	}

	// Verify deleted
	_, err = repo.GetIteration(ctx, 1)
	if err == nil {
		t.Error("expected iteration to be deleted")
	}
}

func TestIterationDeleteCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.IterationDeleteCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(context.Background(), cmdCtx, []string{"999", "--force"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found', got: %s", output)
	}
}

// ============================================================================
// IterationAddTaskCommand Tests
// ============================================================================

func TestIterationAddTaskCommand_SingleTask(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Setup: Create roadmap, track, task, iteration
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", "medium", []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	task := task_manager.NewTaskEntity("task-fc-001", "track-test", "Test Task", "Description", "todo", "medium", "", time.Now().UTC(), time.Now().UTC())
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Add task to iteration
	cmd := &task_manager.IterationAddTaskCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1", "task-fc-001"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "1 task(s) added") {
		t.Errorf("expected '1 task(s) added', got: %s", output)
	}
}

func TestIterationAddTaskCommand_MultipleTasksSuccess(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Setup
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", "medium", []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create tasks
	for i := 1; i <= 3; i++ {
		id := "task-fc-" + string(rune('0'+byte(i)))
		task := task_manager.NewTaskEntity(id, "track-test", "Task "+string(rune('0'+byte(i))), "Description", "todo", "medium", "", time.Now().UTC(), time.Now().UTC())
		if err := repo.SaveTask(ctx, task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Add multiple tasks
	cmd := &task_manager.IterationAddTaskCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1", "task-fc-1", "task-fc-2", "task-fc-3"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "3 task(s) added") {
		t.Errorf("expected '3 task(s) added', got: %s", output)
	}
}

func TestIterationAddTaskCommand_TaskNotFound(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	cmd := &task_manager.IterationAddTaskCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1", "nonexistent-task"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found', got: %s", output)
	}
}

// ============================================================================
// IterationRemoveTaskCommand Tests
// ============================================================================

func TestIterationRemoveTaskCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Setup
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", "medium", []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	task := task_manager.NewTaskEntity("task-fc-001", "track-test", "Test Task", "Description", "todo", "medium", "", time.Now().UTC(), time.Now().UTC())
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}
	if err := repo.AddTaskToIteration(ctx, 1, "task-fc-001"); err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	// Remove task from iteration
	cmd := &task_manager.IterationRemoveTaskCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1", "task-fc-001"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "removed") {
		t.Errorf("expected 'removed', got: %s", output)
	}
}

// ============================================================================
// IterationStartCommand Tests
// ============================================================================

func TestIterationStartCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Start iteration
	cmd := &task_manager.IterationStartCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "started") {
		t.Errorf("expected 'started', got: %s", output)
	}

	// Verify status
	updated, err := repo.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration: %v", err)
	}
	if updated.Status != "current" {
		t.Errorf("expected status 'current', got '%s'", updated.Status)
	}
}

// ============================================================================
// IterationCompleteCommand Tests
// ============================================================================

func TestIterationCompleteCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Start iteration
	if err := repo.StartIteration(ctx, 1); err != nil {
		t.Fatalf("failed to start iteration: %v", err)
	}

	// Complete iteration
	cmd := &task_manager.IterationCompleteCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "completed") {
		t.Errorf("expected 'completed', got: %s", output)
	}

	// Verify status
	updated, err := repo.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration: %v", err)
	}
	if updated.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", updated.Status)
	}
}

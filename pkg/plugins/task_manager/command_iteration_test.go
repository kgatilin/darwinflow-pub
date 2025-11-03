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
	iter1, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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
	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", 300, []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create task
	task := task_manager.NewTaskEntity("task-fc-001", "track-test", "Test Task", "Description", "todo", 300, "", time.Now().UTC(), time.Now().UTC())
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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

	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", 300, []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	task := task_manager.NewTaskEntity("task-fc-001", "track-test", "Test Task", "Description", "todo", 300, "", time.Now().UTC(), time.Now().UTC())
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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

	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", 300, []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create tasks
	for i := 1; i <= 3; i++ {
		id := "task-fc-" + string(rune('0'+byte(i)))
		task := task_manager.NewTaskEntity(id, "track-test", "Task "+string(rune('0'+byte(i))), "Description", "todo", 300, "", time.Now().UTC(), time.Now().UTC())
		if err := repo.SaveTask(ctx, task); err != nil {
			t.Fatalf("failed to save task: %v", err)
		}
	}

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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

	track, err := task_manager.NewTrackEntity("track-test", roadmap.ID, "Test Track", "Description", "not-started", 300, []string{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	task := task_manager.NewTaskEntity("task-fc-001", "track-test", "Test Task", "Description", "todo", 300, "", time.Now().UTC(), time.Now().UTC())
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
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

// ============================================================================
// IterationViewCommand Tests
// ============================================================================

func TestIterationViewCommand_Success(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create a roadmap
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test Roadmap",
		"Success criteria",
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
	track, err := task_manager.NewTrackEntity(
		"TM-track-1",
		"roadmap-test",
		"Test Track",
		"Track description",
		"in-progress",
		500,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create tasks
	task1 := task_manager.NewTaskEntity(
		"TM-task-1",
		"TM-track-1",
		"Implement feature",
		"This is a detailed description of the task",
		"in-progress",
		100,
		"feature-branch",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveTask(ctx, task1); err != nil {
		t.Fatalf("failed to save task 1: %v", err)
	}

	task2 := task_manager.NewTaskEntity(
		"TM-task-2",
		"TM-track-1",
		"Write tests",
		"Write comprehensive tests",
		"done",
		200,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveTask(ctx, task2); err != nil {
		t.Fatalf("failed to save task 2: %v", err)
	}

	// Create iteration (without tasks first)
	iteration, err := task_manager.NewIterationEntity(
		1,
		"Sprint 1",
		"Complete core features",
		"Working prototype",
		[]string{},
		"current",
		500,
		time.Now().UTC(),
		time.Time{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iteration); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Add tasks to iteration
	if err := repo.AddTaskToIteration(ctx, 1, task1.ID); err != nil {
		t.Fatalf("failed to add task 1 to iteration: %v", err)
	}
	if err := repo.AddTaskToIteration(ctx, 1, task2.ID); err != nil {
		t.Fatalf("failed to add task 2 to iteration: %v", err)
	}

	// Create acceptance criteria for task1
	ac1 := task_manager.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		task1.ID,
		"Feature works correctly",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac1.Status = task_manager.ACStatusVerified
	if err := repo.SaveAC(ctx, ac1); err != nil {
		t.Fatalf("failed to save AC 1: %v", err)
	}

	ac2 := task_manager.NewAcceptanceCriteriaEntity(
		"TM-ac-2",
		task1.ID,
		"Error handling implemented",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveAC(ctx, ac2); err != nil {
		t.Fatalf("failed to save AC 2: %v", err)
	}

	// Execute view command
	cmd := &task_manager.IterationViewCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify markdown output
	output := cmdCtx.stdout.String()

	// Check header
	if !strings.Contains(output, "# Iteration #1: Sprint 1") {
		t.Errorf("expected iteration header, got: %s", output)
	}

	// Check metadata
	if !strings.Contains(output, "**Goal**: Complete core features") {
		t.Errorf("expected goal in output, got: %s", output)
	}
	if !strings.Contains(output, "**Deliverable**: Working prototype") {
		t.Errorf("expected deliverable in output, got: %s", output)
	}
	if !strings.Contains(output, "**Status**: current") {
		t.Errorf("expected status in output, got: %s", output)
	}

	// Check tasks section
	if !strings.Contains(output, "## Tasks (2 total, 1 completed)") {
		t.Errorf("expected tasks summary, got: %s", output)
	}
	if !strings.Contains(output, "### TM-task-1: Implement feature") {
		t.Errorf("expected task 1 header, got: %s", output)
	}
	if !strings.Contains(output, "### TM-task-2: Write tests") {
		t.Errorf("expected task 2 header, got: %s", output)
	}

	// Check acceptance criteria
	if !strings.Contains(output, "**Acceptance Criteria**:") {
		t.Errorf("expected AC section, got: %s", output)
	}
	if !strings.Contains(output, "[x] **TM-ac-1**: Feature works correctly") {
		t.Errorf("expected verified AC 1, got: %s", output)
	}
	if !strings.Contains(output, "[ ] **TM-ac-2**: Error handling implemented") {
		t.Errorf("expected unverified AC 2, got: %s", output)
	}

	// Check summary
	if !strings.Contains(output, "## Summary") {
		t.Errorf("expected summary section, got: %s", output)
	}
	if !strings.Contains(output, "- Completed: 1 (50%)") {
		t.Errorf("expected completion percentage, got: %s", output)
	}
}

func TestIterationViewCommand_FullFlag(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create minimal setup
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test Roadmap",
		"Success criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, err := task_manager.NewTrackEntity(
		"TM-track-1",
		"roadmap-test",
		"Test Track",
		"Track description",
		"in-progress",
		500,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	task := task_manager.NewTaskEntity(
		"TM-task-1",
		"TM-track-1",
		"Test Task",
		"This is a very long description that should be displayed in full when --full flag is used",
		"todo",
		100,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	iteration, err := task_manager.NewIterationEntity(
		1,
		"Sprint 1",
		"Goal",
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
	if err := repo.SaveIteration(ctx, iteration); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}
	if err := repo.AddTaskToIteration(ctx, 1, task.ID); err != nil {
		t.Fatalf("failed to add task to iteration: %v", err)
	}

	// Create AC with notes
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		task.ID,
		"Verify feature",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac.Notes = "Some important notes about verification"
	if err := repo.SaveAC(ctx, ac); err != nil {
		t.Fatalf("failed to save AC: %v", err)
	}

	// Execute with --full flag
	cmd := &task_manager.IterationViewCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1", "--full"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()

	// Check that full description is shown
	if !strings.Contains(output, "This is a very long description that should be displayed in full when --full flag is used") {
		t.Errorf("expected full description in output with --full flag, got: %s", output)
	}

	// Check that AC notes are shown
	if !strings.Contains(output, "*Notes*: Some important notes about verification") {
		t.Errorf("expected AC notes in output with --full flag, got: %s", output)
	}
}

func TestIterationViewCommand_NotFound(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &task_manager.IterationViewCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err := cmd.Execute(context.Background(), cmdCtx, []string{"999"})
	if err == nil {
		t.Error("expected error for non-existent iteration")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestIterationViewCommand_EmptyIteration(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create roadmap
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test Roadmap",
		"Success criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create iteration with no tasks
	iteration, err := task_manager.NewIterationEntity(
		1,
		"Empty Sprint",
		"No tasks yet",
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
	if err := repo.SaveIteration(ctx, iteration); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Execute view command
	cmd := &task_manager.IterationViewCommand{Plugin: plugin}
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

	// Check that empty iteration message is displayed
	if !strings.Contains(output, "*No tasks in this iteration*") {
		t.Errorf("expected empty iteration message, got: %s", output)
	}
}

// ============================================================================
// IterationComplete AC Enforcement Tests
// ============================================================================

func TestIterationCompleteCommand_BlocksWithUnverifiedAC(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create roadmap first
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Create a track
	track, err := task_manager.NewTrackEntity(
		"TM-track-1",
		"roadmap-test",
		"Test Track",
		"Track description",
		"in-progress",
		500,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create a task
	task := task_manager.NewTaskEntity(
		"TM-task-1",
		"TM-track-1",
		"Test Task",
		"Task description",
		"todo",
		500,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Add task to iteration
	if err := repo.AddTaskToIteration(ctx, 1, "TM-task-1"); err != nil {
		t.Fatalf("failed to add task to iteration: %v", err)
	}

	// Create AC for task - not verified
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"User can login",
		task_manager.VerificationTypeManual,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveAC(ctx, ac); err != nil {
		t.Fatalf("failed to save AC: %v", err)
	}

	// Start iteration
	if err := repo.StartIteration(ctx, 1); err != nil {
		t.Fatalf("failed to start iteration: %v", err)
	}

	// Try to complete iteration with unverified AC - should fail
	cmd := &task_manager.IterationCompleteCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1"})
	if err == nil {
		t.Errorf("expected error due to unverified AC, but got none")
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Cannot complete iteration") {
		t.Errorf("expected 'Cannot complete iteration' message, got: %s", output)
	}
	if !strings.Contains(output, "Unverified acceptance criteria") {
		t.Errorf("expected 'Unverified acceptance criteria' message, got: %s", output)
	}
	if !strings.Contains(output, "TM-ac-1") {
		t.Errorf("expected AC ID in message, got: %s", output)
	}

	// Verify iteration was NOT completed
	updated, err := repo.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration: %v", err)
	}
	if updated.Status != "current" {
		t.Errorf("expected status 'current' (unchanged), got '%s'", updated.Status)
	}
}

func TestIterationCompleteCommand_AllowsWithVerifiedAC(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create roadmap first
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Create a track
	track, err := task_manager.NewTrackEntity(
		"TM-track-1",
		"roadmap-test",
		"Test Track",
		"Track description",
		"in-progress",
		500,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create a task
	task := task_manager.NewTaskEntity(
		"TM-task-1",
		"TM-track-1",
		"Test Task",
		"Task description",
		"todo",
		500,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Add task to iteration
	if err := repo.AddTaskToIteration(ctx, 1, "TM-task-1"); err != nil {
		t.Fatalf("failed to add task to iteration: %v", err)
	}

	// Create AC for task and mark as verified
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"User can login",
		task_manager.VerificationTypeManual,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac.Status = task_manager.ACStatusVerified // Mark as verified
	if err := repo.SaveAC(ctx, ac); err != nil {
		t.Fatalf("failed to save AC: %v", err)
	}

	// Start iteration
	if err := repo.StartIteration(ctx, 1); err != nil {
		t.Fatalf("failed to start iteration: %v", err)
	}

	// Complete iteration - should succeed
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

	// Verify iteration status changed to complete
	updated, err := repo.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get iteration: %v", err)
	}
	if updated.Status != "complete" {
		t.Errorf("expected status 'complete', got '%s'", updated.Status)
	}
}

func TestIterationCompleteCommand_AutoMarkTasksDone(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create roadmap first
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Create a track
	track, err := task_manager.NewTrackEntity(
		"TM-track-1",
		"roadmap-test",
		"Test Track",
		"Track description",
		"in-progress",
		500,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create multiple tasks
	task1 := task_manager.NewTaskEntity("TM-task-1", "TM-track-1", "Task 1", "", "todo", 500, "", time.Now().UTC(), time.Now().UTC())
	task2 := task_manager.NewTaskEntity("TM-task-2", "TM-track-1", "Task 2", "", "in-progress", 500, "", time.Now().UTC(), time.Now().UTC())
	task3 := task_manager.NewTaskEntity("TM-task-3", "TM-track-1", "Task 3", "", "done", 500, "", time.Now().UTC(), time.Now().UTC())

	if err := repo.SaveTask(ctx, task1); err != nil {
		t.Fatalf("failed to save task1: %v", err)
	}
	if err := repo.SaveTask(ctx, task2); err != nil {
		t.Fatalf("failed to save task2: %v", err)
	}
	if err := repo.SaveTask(ctx, task3); err != nil {
		t.Fatalf("failed to save task3: %v", err)
	}

	// Add tasks to iteration
	if err := repo.AddTaskToIteration(ctx, 1, "TM-task-1"); err != nil {
		t.Fatalf("failed to add task1: %v", err)
	}
	if err := repo.AddTaskToIteration(ctx, 1, "TM-task-2"); err != nil {
		t.Fatalf("failed to add task2: %v", err)
	}
	if err := repo.AddTaskToIteration(ctx, 1, "TM-task-3"); err != nil {
		t.Fatalf("failed to add task3: %v", err)
	}

	// Create and verify ACs for all tasks
	for _, taskID := range []string{"TM-task-1", "TM-task-2", "TM-task-3"} {
		ac := task_manager.NewAcceptanceCriteriaEntity(
			"TM-ac-"+taskID,
			taskID,
			"AC for "+taskID,
			task_manager.VerificationTypeManual,
			"",
			time.Now().UTC(),
			time.Now().UTC(),
		)
		ac.Status = task_manager.ACStatusVerified
		if err := repo.SaveAC(ctx, ac); err != nil {
			t.Fatalf("failed to save AC for %s: %v", taskID, err)
		}
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

	// Verify all tasks are marked done
	for _, taskID := range []string{"TM-task-1", "TM-task-2", "TM-task-3"} {
		updatedTask, err := repo.GetTask(ctx, taskID)
		if err != nil {
			t.Fatalf("failed to get task %s: %v", taskID, err)
		}
		if updatedTask.Status != "done" {
			t.Errorf("expected task %s status 'done', got '%s'", taskID, updatedTask.Status)
		}
	}

	// Verify iteration shows all tasks completed
	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Tasks:      3/3 completed") {
		t.Errorf("expected 'Tasks:      3/3 completed', got: %s", output)
	}
}

func TestIterationCompleteCommand_AllowsWithFailedAC(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	// Create repository for setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})

	// Create roadmap first
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-test", "Vision", "Criteria", time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Sprint 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter); err != nil {
		t.Fatalf("failed to save iteration: %v", err)
	}

	// Create a track
	track, err := task_manager.NewTrackEntity(
		"TM-track-1",
		"roadmap-test",
		"Test Track",
		"Track description",
		"in-progress",
		500,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Create a task
	task := task_manager.NewTaskEntity(
		"TM-task-1",
		"TM-track-1",
		"Test Task",
		"Task description",
		"todo",
		500,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Add task to iteration
	if err := repo.AddTaskToIteration(ctx, 1, "TM-task-1"); err != nil {
		t.Fatalf("failed to add task to iteration: %v", err)
	}

	// Create AC and mark as failed (failed ACs don't block completion)
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"TM-ac-1",
		"TM-task-1",
		"User can login",
		task_manager.VerificationTypeManual,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac.Status = task_manager.ACStatusFailed // Mark as failed
	ac.Notes = "Feature not implemented"
	if err := repo.SaveAC(ctx, ac); err != nil {
		t.Fatalf("failed to save AC: %v", err)
	}

	// Start iteration
	if err := repo.StartIteration(ctx, 1); err != nil {
		t.Fatalf("failed to start iteration: %v", err)
	}

	// Complete iteration - should now BLOCK with failed AC (new behavior)
	cmd := &task_manager.IterationCompleteCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"1"})
	if err == nil {
		t.Errorf("expected error due to failed AC, but got none")
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Cannot complete iteration") {
		t.Errorf("expected 'Cannot complete iteration' message, got: %s", output)
	}
	if !strings.Contains(output, "Failed acceptance criteria") {
		t.Errorf("expected 'Failed acceptance criteria' in output, got: %s", output)
	}
}
func TestIterationReordering(t *testing.T) {
	// Setup
	db := createTestDB(t)
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create roadmap
	now := time.Now().UTC()
	roadmap, err := task_manager.NewRoadmapEntity("roadmap-1", "Test roadmap", "Test criteria", now, now)
	if err != nil {
		t.Fatalf("Failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("Failed to save roadmap: %v", err)
	}

	// Create 3 iterations with same default rank (500)
	iter1, _ := task_manager.NewIterationEntity(1, "Iteration 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)
	iter2, _ := task_manager.NewIterationEntity(2, "Iteration 2", "Goal 2", "", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)
	iter3, _ := task_manager.NewIterationEntity(3, "Iteration 3", "Goal 3", "", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)

	if err := repo.SaveIteration(ctx, iter1); err != nil {
		t.Fatalf("Failed to save iteration 1: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter2); err != nil {
		t.Fatalf("Failed to save iteration 2: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter3); err != nil {
		t.Fatalf("Failed to save iteration 3: %v", err)
	}

	// Verify initial order (by number since ranks are equal)
	iterations, err := repo.ListIterations(ctx)
	if err != nil {
		t.Fatalf("Failed to list iterations: %v", err)
	}
	if len(iterations) != 3 {
		t.Fatalf("Expected 3 iterations, got %d", len(iterations))
	}
	if iterations[0].Number != 1 || iterations[1].Number != 2 || iterations[2].Number != 3 {
		t.Errorf("Initial order wrong: got %d, %d, %d", iterations[0].Number, iterations[1].Number, iterations[2].Number)
	}

	// Log initial ranks
	t.Logf("Initial ranks: iter1=%d, iter2=%d, iter3=%d", iterations[0].Rank, iterations[1].Rank, iterations[2].Rank)

	// Simulate TUI reordering: move iteration 1 down (swap with iteration 2)
	// This is what the TUI code does
	firstIter := iterations[0]
	secondIter := iterations[1]

	t.Logf("Before swap: firstIter.Rank=%d, secondIter.Rank=%d", firstIter.Rank, secondIter.Rank)

	// Swap ranks - this is what the TUI does now
	if firstIter.Rank == secondIter.Rank {
		// Moving down, so increase rank
		firstIter.Rank = firstIter.Rank + 1
	} else {
		firstIter.Rank, secondIter.Rank = secondIter.Rank, firstIter.Rank
	}
	firstIter.UpdatedAt = time.Now().UTC()
	secondIter.UpdatedAt = time.Now().UTC()

	t.Logf("After swap in memory: firstIter.Rank=%d, secondIter.Rank=%d", firstIter.Rank, secondIter.Rank)

	// Update both iterations
	if err := repo.UpdateIteration(ctx, firstIter); err != nil {
		t.Fatalf("Failed to update first iteration: %v", err)
	}
	if err := repo.UpdateIteration(ctx, secondIter); err != nil {
		t.Fatalf("Failed to update second iteration: %v", err)
	}

	// Reload iterations to verify order changed
	iterations, err = repo.ListIterations(ctx)
	if err != nil {
		t.Fatalf("Failed to list iterations after reorder: %v", err)
	}

	t.Logf("After reload from DB: iter[0]=%d (rank=%d), iter[1]=%d (rank=%d), iter[2]=%d (rank=%d)",
		iterations[0].Number, iterations[0].Rank,
		iterations[1].Number, iterations[1].Rank,
		iterations[2].Number, iterations[2].Rank)

	// After moving iter1 down:
	// - iter1 should now have rank=501
	// - iter2 should still have rank=500
	// - iter3 should still have rank=500
	// Order by rank, then number: iter2 (500, #2), iter3 (500, #3), iter1 (501, #1)

	// Expected: iteration 2 should come before iteration 1 now
	if iterations[0].Number != 2 {
		t.Errorf("After reordering down, first iteration should be 2, got %d (rank=%d)", iterations[0].Number, iterations[0].Rank)
	}
	if iterations[1].Number != 3 {
		t.Errorf("After reordering down, second iteration should be 3, got %d (rank=%d)", iterations[1].Number, iterations[1].Rank)
	}
	if iterations[2].Number != 1 {
		t.Errorf("After reordering down, third iteration should be 1, got %d (rank=%d)", iterations[2].Number, iterations[2].Rank)
	}

	// Verify ranks are different now
	if iterations[2].Rank != 501 {
		t.Errorf("Iteration 1 should have rank 501, got %d", iterations[2].Rank)
	}
}

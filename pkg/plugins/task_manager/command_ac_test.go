package task_manager_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// ============================================================================
// ACAddCommand Tests
// ============================================================================

func TestACAddCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup: Create roadmap, track, and task
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

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

	track, err := task_manager.NewTrackEntity(
		"DW-track-1",
		"roadmap-test",
		"Test Track",
		"Test description",
		"not-started",
		200,
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
		"DW-task-1",
		"DW-track-1",
		"Test Task",
		"Test description",
		"todo",
		200,
		"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err := repo.SaveTask(ctx, task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Execute command
	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-task-1",
		"--description", "User can login with email",
		"--type", "manual",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Acceptance criterion created successfully") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "User can login with email") {
		t.Errorf("expected description in output, got: %s", output)
	}
	if !strings.Contains(output, "manual") {
		t.Errorf("expected verification type in output, got: %s", output)
	}

	// Verify AC was created
	acs, err := repo.ListAC(ctx, "DW-task-1")
	if err != nil {
		t.Fatalf("failed to list ACs: %v", err)
	}
	if len(acs) != 1 {
		t.Errorf("expected 1 AC, got %d", len(acs))
	}
	if len(acs) > 0 {
		if acs[0].Description != "User can login with email" {
			t.Errorf("expected description 'User can login with email', got '%s'", acs[0].Description)
		}
		if acs[0].VerificationType != "manual" {
			t.Errorf("expected verification type 'manual', got '%s'", acs[0].VerificationType)
		}
	}
}

func TestACAddCommand_AutomatedType(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	// Execute command with automated type
	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-task-1",
		"--description", "Tests pass with 80% coverage",
		"--type", "automated",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify AC type
	acs, err := repo.ListAC(ctx, "DW-task-1")
	if err != nil {
		t.Fatalf("failed to list ACs: %v", err)
	}
	if len(acs) > 0 && acs[0].VerificationType != "automated" {
		t.Errorf("expected verification type 'automated', got '%s'", acs[0].VerificationType)
	}
}

func TestACAddCommand_MissingTaskID(t *testing.T) {
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

	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--description", "Some description",
	})
	if err == nil {
		t.Errorf("expected error for missing task ID")
	}
	if !strings.Contains(err.Error(), "task-id") {
		t.Errorf("expected error about task-id, got: %v", err)
	}
}

func TestACAddCommand_MissingDescription(t *testing.T) {
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

	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-task-1",
	})
	if err == nil {
		t.Errorf("expected error for missing description")
	}
	if !strings.Contains(err.Error(), "description") {
		t.Errorf("expected error about description, got: %v", err)
	}
}

func TestACAddCommand_InvalidType(t *testing.T) {
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

	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-task-1",
		"--description", "Some description",
		"--type", "invalid",
	})
	if err == nil {
		t.Errorf("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid verification type") {
		t.Errorf("expected error about invalid type, got: %v", err)
	}
}

func TestACAddCommand_TaskNotFound(t *testing.T) {
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

	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"nonexistent-task",
		"--description", "Some description",
	})
	if err == nil {
		t.Errorf("expected error for nonexistent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("expected 'task not found' error, got: %v", err)
	}
}

// ============================================================================
// ACListCommand Tests
// ============================================================================

func TestACListCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	// Create ACs
	for i := 1; i <= 2; i++ {
		ac := task_manager.NewAcceptanceCriteriaEntity(
			"DW-ac-"+string(rune(48+i)),
			"DW-task-1",
			"AC "+string(rune(48+i)),
			task_manager.VerificationTypeManual,
			time.Now().UTC(),
			time.Now().UTC(),
		)
		if err := repo.SaveAC(ctx, ac); err != nil {
			t.Fatalf("failed to save AC: %v", err)
		}
	}

	// Execute command
	cmd := &task_manager.ACListCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-task-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Test Task") {
		t.Errorf("expected task title in output, got: %s", output)
	}
	if !strings.Contains(output, "Summary: 0/2 verified") {
		t.Errorf("expected summary in output, got: %s", output)
	}
}

func TestACListCommand_NoACs(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	// Execute command
	cmd := &task_manager.ACListCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-task-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "No acceptance criteria found") {
		t.Errorf("expected 'No acceptance criteria found', got: %s", output)
	}
}

func TestACListCommand_MissingTaskID(t *testing.T) {
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

	cmd := &task_manager.ACListCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{})
	if err == nil {
		t.Errorf("expected error for missing task ID")
	}
	if !strings.Contains(err.Error(), "task-id") {
		t.Errorf("expected error about task-id, got: %v", err)
	}
}

func TestACListCommand_TaskNotFound(t *testing.T) {
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

	cmd := &task_manager.ACListCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"nonexistent-task"})
	if err == nil {
		t.Errorf("expected error for nonexistent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("expected 'task not found' error, got: %v", err)
	}
}

// ============================================================================
// ACVerifyCommand Tests
// ============================================================================

func TestACVerifyCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC description",
		task_manager.VerificationTypeManual,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute command
	cmd := &task_manager.ACVerifyCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Acceptance criterion verified") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify status changed
	updated, err := repo.GetAC(ctx, "DW-ac-1")
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}
	if updated.Status != task_manager.ACStatusVerified {
		t.Errorf("expected status 'verified', got '%s'", updated.Status)
	}
}

func TestACVerifyCommand_WithNotes(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC description",
		task_manager.VerificationTypeManual,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute command with notes
	cmd := &task_manager.ACVerifyCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-ac-1",
		"--notes", "Tested on Chrome and Firefox",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify notes were saved
	updated, err := repo.GetAC(ctx, "DW-ac-1")
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}
	if updated.Notes != "Tested on Chrome and Firefox" {
		t.Errorf("expected notes 'Tested on Chrome and Firefox', got '%s'", updated.Notes)
	}
}

func TestACVerifyCommand_NotFound(t *testing.T) {
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

	cmd := &task_manager.ACVerifyCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"nonexistent-ac"})
	if err == nil {
		t.Errorf("expected error for nonexistent AC")
	}
	if !strings.Contains(err.Error(), "AC not found") {
		t.Errorf("expected 'AC not found' error, got: %v", err)
	}
}

// ============================================================================
// ACFailCommand Tests
// ============================================================================

func TestACFailCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC description",
		task_manager.VerificationTypeManual,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute command
	cmd := &task_manager.ACFailCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-1", "--reason", "Tests failing on Safari"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Acceptance criterion marked as failed") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Tests failing on Safari") {
		t.Errorf("expected reason in output, got: %s", output)
	}

	// Verify status changed
	updated, err := repo.GetAC(ctx, "DW-ac-1")
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}
	if updated.Status != task_manager.ACStatusFailed {
		t.Errorf("expected status 'failed', got '%s'", updated.Status)
	}
}

func TestACFailCommand_NotFound(t *testing.T) {
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

	cmd := &task_manager.ACFailCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"nonexistent-ac"})
	if err == nil {
		t.Errorf("expected error for nonexistent AC")
	}
	if !strings.Contains(err.Error(), "AC not found") {
		t.Errorf("expected 'AC not found' error, got: %v", err)
	}
}

// ============================================================================
// ACUpdateCommand Tests
// ============================================================================

func TestACUpdateCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"Original description",
		task_manager.VerificationTypeManual,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute command
	cmd := &task_manager.ACUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-ac-1",
		"--description", "Updated description",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Acceptance criterion updated") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Updated description") {
		t.Errorf("expected updated description in output, got: %s", output)
	}

	// Verify description changed
	updated, err := repo.GetAC(ctx, "DW-ac-1")
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}
	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got '%s'", updated.Description)
	}
}

func TestACUpdateCommand_MissingDescription(t *testing.T) {
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

	cmd := &task_manager.ACUpdateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-1"})
	if err == nil {
		t.Errorf("expected error for missing description")
	}
	if !strings.Contains(err.Error(), "description") {
		t.Errorf("expected error about description, got: %v", err)
	}
}

// ============================================================================
// ACDeleteCommand Tests
// ============================================================================

func TestACDeleteCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC description",
		task_manager.VerificationTypeManual,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute command with --force
	cmd := &task_manager.ACDeleteCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-1", "--force"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Acceptance criterion deleted") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify AC was deleted
	_, err = repo.GetAC(ctx, "DW-ac-1")
	if err == nil {
		t.Errorf("expected AC to be deleted")
	}
}

func TestACDeleteCommand_NoForce(t *testing.T) {
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

	cmd := &task_manager.ACDeleteCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-1"})
	if err == nil {
		t.Errorf("expected error for missing --force flag")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("expected error about --force flag, got: %v", err)
	}
}

// ============================================================================
// ACVerifyAutoCommand Tests
// ============================================================================

func TestACVerifyAutoCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC description",
		task_manager.VerificationTypeAutomated,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute command
	cmd := &task_manager.ACVerifyAutoCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Acceptance criterion marked as automatically verified") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify status changed
	updated, err := repo.GetAC(ctx, "DW-ac-1")
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}
	if updated.Status != task_manager.ACStatusAutomaticallyVerified {
		t.Errorf("expected status 'automatically_verified', got '%s'", updated.Status)
	}
}

// ============================================================================
// ACRequestReviewCommand Tests
// ============================================================================

func TestACRequestReviewCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task)

	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC description",
		task_manager.VerificationTypeManual,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute command
	cmd := &task_manager.ACRequestReviewCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Human review requested") {
		t.Errorf("expected success message, got: %s", output)
	}

	// Verify status changed
	updated, err := repo.GetAC(ctx, "DW-ac-1")
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}
	if updated.Status != task_manager.ACStatusPendingHumanReview {
		t.Errorf("expected status 'pending_human_review', got '%s'", updated.Status)
	}
}

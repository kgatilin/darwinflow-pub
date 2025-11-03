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
			"",
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
			"",
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
			"",
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
			"",
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
			"",
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
			"",
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
		"",
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
			"",
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

// ============================================================================
// ACFailedCommand Tests
// ============================================================================

func TestACFailedCommand_NoFilters(t *testing.T) {
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

	task1 := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task 1", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task1)

	task2 := task_manager.NewTaskEntity("DW-task-2", "DW-track-1", "Test Task 2", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task2)

	// Create failed ACs
	ac1 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC1 description",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac1.Status = task_manager.ACStatusFailed
	ac1.Notes = "Test failure reason 1"
	repo.SaveAC(ctx, ac1)

	ac2 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-2",
		"DW-task-2",
		"AC2 description",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac2.Status = task_manager.ACStatusFailed
	ac2.Notes = "Test failure reason 2"
	repo.SaveAC(ctx, ac2)

	// Create a verified AC (should not appear)
	ac3 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-3",
		"DW-task-1",
		"AC3 description",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac3.Status = task_manager.ACStatusVerified
	repo.SaveAC(ctx, ac3)

	// Execute command with no filters
	cmd := &task_manager.ACFailedCommand{Plugin: plugin}
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
	if !strings.Contains(output, "Total: 2") {
		t.Errorf("expected 2 failed ACs, got output: %s", output)
	}
	if !strings.Contains(output, "DW-ac-1") || !strings.Contains(output, "DW-ac-2") {
		t.Errorf("expected both failed ACs in output, got: %s", output)
	}
	if !strings.Contains(output, "Test failure reason 1") || !strings.Contains(output, "Test failure reason 2") {
		t.Errorf("expected failure reasons in output, got: %s", output)
	}
	if strings.Contains(output, "DW-ac-3") {
		t.Errorf("verified AC should not appear in output, got: %s", output)
	}
}

func TestACFailedCommand_FilterByIteration(t *testing.T) {
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

	task1 := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task 1", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task1)

	task2 := task_manager.NewTaskEntity("DW-task-2", "DW-track-1", "Test Task 2", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task2)

	// Create iteration
	iter, err := task_manager.NewIterationEntity(1, "Iteration 1", "Test goal", "Test deliverable", []string{}, "planned", 100, time.Time{}, time.Time{}, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to create iteration: %v", err)
	}
	repo.SaveIteration(ctx, iter)
	repo.AddTaskToIteration(ctx, 1, "DW-task-1")

	// Create failed ACs
	ac1 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC1 in iteration",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac1.Status = task_manager.ACStatusFailed
	ac1.Notes = "Iteration failure"
	repo.SaveAC(ctx, ac1)

	ac2 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-2",
		"DW-task-2",
		"AC2 not in iteration",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac2.Status = task_manager.ACStatusFailed
	ac2.Notes = "Not in iteration"
	repo.SaveAC(ctx, ac2)

	// Execute command with iteration filter
	cmd := &task_manager.ACFailedCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"--iteration", "1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Iteration 1") {
		t.Errorf("expected iteration filter in header, got: %s", output)
	}
	if !strings.Contains(output, "Total: 1") {
		t.Errorf("expected 1 failed AC in iteration, got: %s", output)
	}
	if !strings.Contains(output, "DW-ac-1") {
		t.Errorf("expected AC1 in output, got: %s", output)
	}
	if strings.Contains(output, "DW-ac-2") {
		t.Errorf("AC2 should not appear (not in iteration), got: %s", output)
	}
}

func TestACFailedCommand_FilterByTrack(t *testing.T) {
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

	track1, _ := task_manager.NewTrackEntity("DW-track-1", "roadmap-test", "Test Track 1", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track1)

	track2, _ := task_manager.NewTrackEntity("DW-track-2", "roadmap-test", "Test Track 2", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track2)

	task1 := task_manager.NewTaskEntity("DW-task-1", "DW-track-1", "Test Task 1", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task1)

	task2 := task_manager.NewTaskEntity("DW-task-2", "DW-track-2", "Test Task 2", "", "todo", 200, "", time.Now().UTC(), time.Now().UTC())
	repo.SaveTask(ctx, task2)

	// Create failed ACs
	ac1 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC1 in track 1",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac1.Status = task_manager.ACStatusFailed
	repo.SaveAC(ctx, ac1)

	ac2 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-2",
		"DW-task-2",
		"AC2 in track 2",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac2.Status = task_manager.ACStatusFailed
	repo.SaveAC(ctx, ac2)

	// Execute command with track filter
	cmd := &task_manager.ACFailedCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"--track", "DW-track-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Track: DW-track-1") {
		t.Errorf("expected track filter in header, got: %s", output)
	}
	if !strings.Contains(output, "Total: 1") {
		t.Errorf("expected 1 failed AC in track, got: %s", output)
	}
	if !strings.Contains(output, "DW-ac-1") {
		t.Errorf("expected AC1 in output, got: %s", output)
	}
	if strings.Contains(output, "DW-ac-2") {
		t.Errorf("AC2 should not appear (different track), got: %s", output)
	}
}

func TestACFailedCommand_FilterByTask(t *testing.T) {
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

	// Create multiple failed ACs for same task
	ac1 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"AC1 for task",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac1.Status = task_manager.ACStatusFailed
	repo.SaveAC(ctx, ac1)

	ac2 := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-2",
		"DW-task-1",
		"AC2 for task",
		task_manager.VerificationTypeManual,
			"",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	ac2.Status = task_manager.ACStatusFailed
	repo.SaveAC(ctx, ac2)

	// Execute command with task filter
	cmd := &task_manager.ACFailedCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"--task", "DW-task-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Task: DW-task-1") {
		t.Errorf("expected task filter in header, got: %s", output)
	}
	if !strings.Contains(output, "Total: 2") {
		t.Errorf("expected 2 failed ACs for task, got: %s", output)
	}
	if !strings.Contains(output, "DW-ac-1") || !strings.Contains(output, "DW-ac-2") {
		t.Errorf("expected both ACs in output, got: %s", output)
	}
}

func TestACFailedCommand_NoResults(t *testing.T) {
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

	// Setup with no failed ACs
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	// Execute command
	cmd := &task_manager.ACFailedCommand{Plugin: plugin}
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
	if !strings.Contains(output, "No failed acceptance criteria found") {
		t.Errorf("expected 'no failed ACs' message, got: %s", output)
	}
}

// ============================================================================
// ACAddCommand Tests - Testing Instructions
// ============================================================================

func TestACAddCommand_WithTestingInstructions(t *testing.T) {
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

	// Execute command with testing instructions
	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	testingInstructions := "1. Open the app\n2. Navigate to login\n3. Verify email field exists"

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-task-1",
		"--description", "User can login with email",
		"--testing-instructions", testingInstructions,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify AC was created with testing instructions
	acs, err := repo.ListAC(ctx, "DW-task-1")
	if err != nil {
		t.Fatalf("failed to list ACs: %v", err)
	}
	if len(acs) != 1 {
		t.Errorf("expected 1 AC, got %d", len(acs))
	}
	if len(acs) > 0 {
		if acs[0].TestingInstructions != testingInstructions {
			t.Errorf("expected testing instructions '%s', got '%s'", testingInstructions, acs[0].TestingInstructions)
		}
	}
}

func TestACAddCommand_WithoutTestingInstructions(t *testing.T) {
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

	// Execute command without testing instructions
	cmd := &task_manager.ACAddCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-task-1",
		"--description", "User can login with email",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify AC was created with empty testing instructions
	acs, err := repo.ListAC(ctx, "DW-task-1")
	if err != nil {
		t.Fatalf("failed to list ACs: %v", err)
	}
	if len(acs) > 0 && acs[0].TestingInstructions != "" {
		t.Errorf("expected empty testing instructions, got '%s'", acs[0].TestingInstructions)
	}
}

// ============================================================================
// ACListCommand Tests - Testing Instructions Display
// ============================================================================

func TestACListCommand_DisplaysTestingInstructions(t *testing.T) {
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

	testingInstructions := "1. Click the button\n2. Verify result"
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"Button works correctly",
		task_manager.VerificationTypeManual,
		testingInstructions,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute list command
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
	if !strings.Contains(output, "Testing instructions:") {
		t.Errorf("expected 'Testing instructions:' in output, got: %s", output)
	}
	if !strings.Contains(output, testingInstructions) {
		t.Errorf("expected testing instructions content in output, got: %s", output)
	}
}

// ============================================================================
// ACShowCommand Tests
// ============================================================================

func TestACShowCommand_Success(t *testing.T) {
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

	testingInstructions := "1. Click the button\n2. Verify result"
	ac := task_manager.NewAcceptanceCriteriaEntity(
		"DW-ac-1",
		"DW-task-1",
		"Button works correctly",
		task_manager.VerificationTypeManual,
		testingInstructions,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute show command
	cmd := &task_manager.ACShowCommand{Plugin: plugin}
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
	if !strings.Contains(output, "DW-ac-1") {
		t.Errorf("expected AC ID in output, got: %s", output)
	}
	if !strings.Contains(output, "Button works correctly") {
		t.Errorf("expected AC description in output, got: %s", output)
	}
	if !strings.Contains(output, "Testing Instructions:") {
		t.Errorf("expected 'Testing Instructions:' in output, got: %s", output)
	}
	if !strings.Contains(output, testingInstructions) {
		t.Errorf("expected testing instructions content in output, got: %s", output)
	}
}

func TestACShowCommand_WithoutTestingInstructions(t *testing.T) {
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
		"Button works correctly",
		task_manager.VerificationTypeManual,
		"", // no testing instructions
		time.Now().UTC(),
		time.Now().UTC(),
	)
	repo.SaveAC(ctx, ac)

	// Execute show command
	cmd := &task_manager.ACShowCommand{Plugin: plugin}
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
	if !strings.Contains(output, "Testing Instructions: (none)") {
		t.Errorf("expected '(none)' message for testing instructions, got: %s", output)
	}
}

func TestACShowCommand_NotFound(t *testing.T) {
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

	// Execute show command with non-existent AC
	cmd := &task_manager.ACShowCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-ac-999"})
	if err == nil {
		t.Errorf("expected error for non-existent AC, got nil")
	}

	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found' message, got: %s", output)
	}
}

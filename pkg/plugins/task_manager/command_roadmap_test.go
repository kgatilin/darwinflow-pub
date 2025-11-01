package task_manager_test

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// mockCommandContext implements pluginsdk.CommandContext for testing
type mockCommandContext struct {
	workingDir string
	stdout     *bytes.Buffer
	stdin      io.Reader
	logger     pluginsdk.Logger
}

func (m *mockCommandContext) GetWorkingDir() string {
	return m.workingDir
}

func (m *mockCommandContext) GetStdout() io.Writer {
	return m.stdout
}

func (m *mockCommandContext) GetStdin() io.Reader {
	if m.stdin == nil {
		return &bytes.Buffer{}
	}
	return m.stdin
}

func (m *mockCommandContext) GetLogger() pluginsdk.Logger {
	return m.logger
}

func (m *mockCommandContext) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	return nil
}

// createRoadmapTestDB creates a test database with schema for roadmap tests
func createRoadmapTestDB(t *testing.T) *sql.DB {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	return db
}

// TestRoadmapInitCommand tests successful roadmap creation
func TestRoadmapInitCommand_Success(t *testing.T) {
	// Setup
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	cmd := &task_manager.RoadmapInitCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	// Execute
	err := cmd.Execute(ctx, cmdCtx, []string{
		"--vision", "Build extensible framework",
		"--success-criteria", "Support 10 plugins",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify output
	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Roadmap created successfully") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Build extensible framework") {
		t.Errorf("expected vision in output, got: %s", output)
	}

	// Verify roadmap was saved - get database for default project
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		t.Errorf("failed to retrieve saved roadmap: %v", err)
	}
	if roadmap == nil {
		t.Error("roadmap should not be nil")
	}
	if roadmap != nil {
		if roadmap.Vision != "Build extensible framework" {
			t.Errorf("expected vision 'Build extensible framework', got '%s'", roadmap.Vision)
		}
		if roadmap.SuccessCriteria != "Support 10 plugins" {
			t.Errorf("expected success criteria 'Support 10 plugins', got '%s'", roadmap.SuccessCriteria)
		}
	}
}

// TestRoadmapInitCommand_MissingVision tests error when vision is missing
func TestRoadmapInitCommand_MissingVision(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.RoadmapInitCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	// Execute without --vision
	err = cmd.Execute(ctx, cmdCtx, []string{
		"--success-criteria", "Support 10 plugins",
	})

	if err == nil || !strings.Contains(err.Error(), "--vision is required") {
		t.Errorf("expected error about missing vision, got: %v", err)
	}
}

// TestRoadmapInitCommand_MissingSuccessCriteria tests error when success criteria is missing
func TestRoadmapInitCommand_MissingSuccessCriteria(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.RoadmapInitCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	// Execute without --success-criteria
	err = cmd.Execute(ctx, cmdCtx, []string{
		"--vision", "Build extensible framework",
	})

	if err == nil || !strings.Contains(err.Error(), "--success-criteria is required") {
		t.Errorf("expected error about missing success criteria, got: %v", err)
	}
}

// TestRoadmapInitCommand_DuplicateRoadmap tests error when roadmap already exists
func TestRoadmapInitCommand_DuplicateRoadmap(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	// Create first roadmap using database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()
	roadmap1, err := task_manager.NewRoadmapEntity(
		"roadmap-1",
		"First vision",
		"First criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap entity: %v", err)
	}

	if err := repo.SaveRoadmap(ctx, roadmap1); err != nil {
		t.Fatalf("failed to save first roadmap: %v", err)
	}

	// Try to create second roadmap via command
	cmd := &task_manager.RoadmapInitCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--vision", "Build extensible framework",
		"--success-criteria", "Support 10 plugins",
	})

	if err == nil || !strings.Contains(err.Error(), "roadmap already exists") {
		t.Errorf("expected error about duplicate roadmap, got: %v", err)
	}
}

// TestRoadmapShowCommand_WithExistingRoadmap tests showing an existing roadmap
func TestRoadmapShowCommand_WithExistingRoadmap(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap in the project database
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
		t.Fatalf("failed to create roadmap entity: %v", err)
	}

	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Execute command
	cmd := &task_manager.RoadmapShowCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify output
	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "Test vision") {
		t.Errorf("expected vision in output, got: %s", output)
	}
	if !strings.Contains(output, "Test criteria") {
		t.Errorf("expected criteria in output, got: %s", output)
	}
	if !strings.Contains(output, "roadmap-test") {
		t.Errorf("expected roadmap ID in output, got: %s", output)
	}
}

// TestRoadmapShowCommand_NoExistingRoadmap tests showing when no roadmap exists
func TestRoadmapShowCommand_NoExistingRoadmap(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Execute command without creating roadmap
	cmd := &task_manager.RoadmapShowCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify helpful message
	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "No roadmap found") {
		t.Errorf("expected 'No roadmap found' message, got: %s", output)
	}
	if !strings.Contains(output, "roadmap init") {
		t.Errorf("expected suggestion to create roadmap, got: %s", output)
	}
}

// TestRoadmapUpdateCommand_UpdateVision tests updating only vision
func TestRoadmapUpdateCommand_UpdateVision(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap in the project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Old vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap entity: %v", err)
	}

	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Execute command
	cmd := &task_manager.RoadmapUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--vision", "New vision",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify update
	updated, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		t.Errorf("failed to retrieve updated roadmap: %v", err)
	}
	if updated.Vision != "New vision" {
		t.Errorf("expected vision 'New vision', got '%s'", updated.Vision)
	}
	// Criteria should not change
	if updated.SuccessCriteria != "Test criteria" {
		t.Errorf("expected unchanged criteria 'Test criteria', got '%s'", updated.SuccessCriteria)
	}
}

// TestRoadmapUpdateCommand_UpdateSuccessCriteria tests updating only success criteria
func TestRoadmapUpdateCommand_UpdateSuccessCriteria(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap in the project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Old criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap entity: %v", err)
	}

	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Execute command
	cmd := &task_manager.RoadmapUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--success-criteria", "New criteria",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify update
	updated, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		t.Errorf("failed to retrieve updated roadmap: %v", err)
	}
	if updated.SuccessCriteria != "New criteria" {
		t.Errorf("expected criteria 'New criteria', got '%s'", updated.SuccessCriteria)
	}
	// Vision should not change
	if updated.Vision != "Test vision" {
		t.Errorf("expected unchanged vision 'Test vision', got '%s'", updated.Vision)
	}
}

// TestRoadmapUpdateCommand_UpdateBoth tests updating both vision and criteria
func TestRoadmapUpdateCommand_UpdateBoth(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Setup: Create roadmap in the project database
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Old vision",
		"Old criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap entity: %v", err)
	}

	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Execute command
	cmd := &task_manager.RoadmapUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--vision", "New vision",
		"--success-criteria", "New criteria",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify update
	updated, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		t.Errorf("failed to retrieve updated roadmap: %v", err)
	}
	if updated.Vision != "New vision" {
		t.Errorf("expected vision 'New vision', got '%s'", updated.Vision)
	}
	if updated.SuccessCriteria != "New criteria" {
		t.Errorf("expected criteria 'New criteria', got '%s'", updated.SuccessCriteria)
	}
}

// TestRoadmapUpdateCommand_NoFlags tests error when no update flags provided
func TestRoadmapUpdateCommand_NoFlags(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create roadmap
	repo := plugin.GetRepository()
	ctx := context.Background()
	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap entity: %v", err)
	}

	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	// Execute command without flags
	cmd := &task_manager.RoadmapUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{})
	if err == nil || !strings.Contains(err.Error(), "at least one flag must be provided") {
		t.Errorf("expected error about missing flags, got: %v", err)
	}
}

// TestRoadmapUpdateCommand_NoRoadmap tests error when no roadmap exists
func TestRoadmapUpdateCommand_NoRoadmap(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Execute command without creating roadmap
	cmd := &task_manager.RoadmapUpdateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--vision", "New vision",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should show helpful message
	output := cmdCtx.stdout.String()
	if !strings.Contains(output, "No roadmap found") {
		t.Errorf("expected 'No roadmap found' message, got: %s", output)
	}
}

// TestCommandGetters tests that commands return correct metadata
func TestRoadmapCommandGetters(t *testing.T) {
	tests := []struct {
		name        string
		cmd         pluginsdk.Command
		expectedName string
		expectedDesc string
	}{
		{
			name:        "RoadmapInitCommand",
			cmd:         &task_manager.RoadmapInitCommand{},
			expectedName: "roadmap.init",
			expectedDesc: "Initialize a new roadmap",
		},
		{
			name:        "RoadmapShowCommand",
			cmd:         &task_manager.RoadmapShowCommand{},
			expectedName: "roadmap.show",
			expectedDesc: "Display the current roadmap",
		},
		{
			name:        "RoadmapUpdateCommand",
			cmd:         &task_manager.RoadmapUpdateCommand{},
			expectedName: "roadmap.update",
			expectedDesc: "Update the current roadmap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.GetName(); got != tt.expectedName {
				t.Errorf("GetName() = %s, want %s", got, tt.expectedName)
			}
			if got := tt.cmd.GetDescription(); got != tt.expectedDesc {
				t.Errorf("GetDescription() = %s, want %s", got, tt.expectedDesc)
			}
		})
	}
}

// stubLogger is a simple logger implementation for testing
type stubLogger struct{}

func (s *stubLogger) Info(msg string, args ...interface{}) {}
func (s *stubLogger) Warn(msg string, args ...interface{}) {}
func (s *stubLogger) Error(msg string, args ...interface{}) {}
func (s *stubLogger) Debug(msg string, args ...interface{}) {}

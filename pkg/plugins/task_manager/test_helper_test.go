package task_manager_test

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// setupTestPlugin creates a plugin and initializes the default project
func setupTestPlugin(t *testing.T) (*task_manager.TaskManagerPlugin, string) {
	tmpDir := t.TempDir()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create default project
	projectCmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	projectCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}
	if err := projectCmd.Execute(context.Background(), projectCtx, []string{"default"}); err != nil {
		t.Fatalf("failed to create default project: %v", err)
	}

	return plugin, tmpDir
}

// getProjectDB gets the database for a project for direct data setup in tests
func getProjectDB(t *testing.T, workingDir, projectName string) *sql.DB {
	projectDir := filepath.Join(workingDir, ".darwinflow", "projects", projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	dbPath := filepath.Join(projectDir, "roadmap.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := task_manager.InitSchema(db); err != nil {
		db.Close()
		t.Fatalf("failed to initialize schema: %v", err)
	}

	return db
}

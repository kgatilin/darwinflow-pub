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

// setupTestWithRoadmapAndTrack creates a plugin with roadmap and track
func setupTestWithRoadmapAndTrack(t *testing.T) (*task_manager.TaskManagerPlugin, string, string) {
	plugin, tmpDir := setupTestPlugin(t)
	ctx := context.Background()

	// Create roadmap
	roadmapCmd := &task_manager.RoadmapInitCommand{Plugin: plugin}
	roadmapCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}
	err := roadmapCmd.Execute(ctx, roadmapCtx, []string{
		"--vision", "Test vision",
		"--success-criteria", "Test criteria",
	})
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}

	// Create track
	trackCmd := &task_manager.TrackCreateCommand{Plugin: plugin}
	trackCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}
	err = trackCmd.Execute(ctx, trackCtx, []string{
		"--title", "Test Track",
		"--description", "Test description",
		"--rank", "200",
	})
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}

	// Extract track ID from output
	trackOutput := trackCtx.stdout.String()
	trackIDPrefix := "ID:"
	trackIDStart := indexOf(trackOutput, trackIDPrefix)
	if trackIDStart == -1 {
		t.Fatalf("failed to find track ID in output: %s", trackOutput)
	}
	trackIDStart += len(trackIDPrefix)
	trackIDEnd := indexOf(trackOutput[trackIDStart:], "\n")
	if trackIDEnd == -1 {
		trackIDEnd = len(trackOutput)
	} else {
		trackIDEnd += trackIDStart
	}
	trackID := trimSpace(trackOutput[trackIDStart:trackIDEnd])

	return plugin, tmpDir, trackID
}

// indexOf returns the index of substring in string, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

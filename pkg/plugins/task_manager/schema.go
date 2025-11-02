package task_manager

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// SchemaVersion is the current database schema version
	SchemaVersion = 3
)

// SQL table creation statements
const (
	createRoadmapsTable = `
CREATE TABLE IF NOT EXISTS roadmaps (
    id TEXT PRIMARY KEY,
    vision TEXT NOT NULL,
    success_criteria TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
)
`

	createTracksTable = `
CREATE TABLE IF NOT EXISTS tracks (
    id TEXT PRIMARY KEY,
    roadmap_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY(roadmap_id) REFERENCES roadmaps(id) ON DELETE CASCADE
)
`

	createTrackDependenciesTable = `
CREATE TABLE IF NOT EXISTS track_dependencies (
    track_id TEXT NOT NULL,
    depends_on_id TEXT NOT NULL,
    PRIMARY KEY (track_id, depends_on_id),
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_id) REFERENCES tracks(id) ON DELETE CASCADE
)
`

	createTasksTable = `
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    track_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    branch TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE
)
`

	createIterationsTable = `
CREATE TABLE IF NOT EXISTS iterations (
    number INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    goal TEXT,
    status TEXT NOT NULL,
    deliverable TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
)
`

	createIterationTasksTable = `
CREATE TABLE IF NOT EXISTS iteration_tasks (
    iteration_number INTEGER NOT NULL,
    task_id TEXT NOT NULL,
    PRIMARY KEY (iteration_number, task_id),
    FOREIGN KEY (iteration_number) REFERENCES iterations(number) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
)
`

	createProjectMetadataTable = `
CREATE TABLE IF NOT EXISTS project_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
)
`

	createAcceptanceCriteriaTable = `
CREATE TABLE IF NOT EXISTS acceptance_criteria (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    description TEXT NOT NULL,
    verification_type TEXT NOT NULL,
    status TEXT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
)
`

	// Indexes for common queries
	createTracksRoadmapIDIndex = `
CREATE INDEX IF NOT EXISTS idx_tracks_roadmap_id ON tracks(roadmap_id)
`

	createTracksStatusIndex = `
CREATE INDEX IF NOT EXISTS idx_tracks_status ON tracks(status)
`

	createTasksTrackIDIndex = `
CREATE INDEX IF NOT EXISTS idx_tasks_track_id ON tasks(track_id)
`

	createTasksStatusIndex = `
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)
`

	createIterationsStatusIndex = `
CREATE INDEX IF NOT EXISTS idx_iterations_status ON iterations(status)
`

	createIterationTasksIterationIndex = `
CREATE INDEX IF NOT EXISTS idx_iteration_tasks_iteration ON iteration_tasks(iteration_number)
`

	createIterationTasksTaskIndex = `
CREATE INDEX IF NOT EXISTS idx_iteration_tasks_task ON iteration_tasks(task_id)
`

	createAcceptanceCriteriaTaskIDIndex = `
CREATE INDEX IF NOT EXISTS idx_ac_task_id ON acceptance_criteria(task_id)
`

	createAcceptanceCriteriaStatusIndex = `
CREATE INDEX IF NOT EXISTS idx_ac_status ON acceptance_criteria(status)
`

	createADRsTable = `
CREATE TABLE IF NOT EXISTS adrs (
    id TEXT PRIMARY KEY,
    track_id TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    context TEXT NOT NULL,
    decision TEXT NOT NULL,
    consequences TEXT NOT NULL,
    alternatives TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    superseded_by TEXT,
    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY(superseded_by) REFERENCES adrs(id) ON DELETE SET NULL
)
`

	createADRsTrackIDIndex = `
CREATE INDEX IF NOT EXISTS idx_adrs_track_id ON adrs(track_id)
`

	createADRsStatusIndex = `
CREATE INDEX IF NOT EXISTS idx_adrs_status ON adrs(status)
`
)

// InitSchema initializes the database schema with all required tables and indexes.
// It's safe to call multiple times (uses IF NOT EXISTS).
func InitSchema(db *sql.DB) error {
	statements := []string{
		createRoadmapsTable,
		createTracksTable,
		createTrackDependenciesTable,
		createTasksTable,
		createIterationsTable,
		createIterationTasksTable,
		createProjectMetadataTable,
		createAcceptanceCriteriaTable,
		createADRsTable,
		createTracksRoadmapIDIndex,
		createTracksStatusIndex,
		createTasksTrackIDIndex,
		createTasksStatusIndex,
		createIterationsStatusIndex,
		createIterationTasksIterationIndex,
		createIterationTasksTaskIndex,
		createAcceptanceCriteriaTaskIDIndex,
		createAcceptanceCriteriaStatusIndex,
		createADRsTrackIDIndex,
		createADRsStatusIndex,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	return nil
}

// MigrateFromFileStorage migrates existing task JSON files to the database.
// It creates a "legacy-tasks" track if needed and imports all tasks from the file storage directory.
func MigrateFromFileStorage(db *sql.DB, tasksDir string) error {
	// Check if tasks directory exists
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		// No existing tasks to migrate
		return nil
	}

	// First, check if there are any tasks already in the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing tasks: %w", err)
	}

	// If database already has tasks, skip migration
	if count > 0 {
		return nil
	}

	// Read task files from directory
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return fmt.Errorf("failed to read tasks directory: %w", err)
	}

	// Check if there are any task files
	taskFiles := []os.DirEntry{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			taskFiles = append(taskFiles, entry)
		}
	}

	// If no task files, nothing to migrate
	if len(taskFiles) == 0 {
		return nil
	}

	// Create a default roadmap for legacy tasks
	legacyRoadmapID := "legacy-roadmap"
	legacyTrackID := "track-legacy-tasks"

	// Check if legacy roadmap exists
	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM roadmaps WHERE id = ?", legacyRoadmapID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check for legacy roadmap: %w", err)
	}

	// Only create if doesn't exist
	if exists == 0 {
		roadmap, err := NewRoadmapEntity(
			legacyRoadmapID,
			"Legacy Tasks from File Storage",
			"Migrate existing tasks to database",
			GetCurrentTime(),
			GetCurrentTime(),
		)
		if err != nil {
			return fmt.Errorf("failed to create legacy roadmap: %w", err)
		}

		_, err = db.Exec(
			"INSERT INTO roadmaps (id, vision, success_criteria, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			roadmap.ID, roadmap.Vision, roadmap.SuccessCriteria, roadmap.CreatedAt, roadmap.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert legacy roadmap: %w", err)
		}

		// Create a legacy track
		track, err := NewTrackEntity(
			legacyTrackID,
			legacyRoadmapID,
			"Legacy Tasks",
			"Tasks migrated from file-based storage",
			"not-started",
			300, // low priority = 300 rank
			[]string{},
			GetCurrentTime(),
			GetCurrentTime(),
		)
		if err != nil {
			return fmt.Errorf("failed to create legacy track: %w", err)
		}

		_, err = db.Exec(
			"INSERT INTO tracks (id, roadmap_id, title, description, status, priority, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			track.ID, track.RoadmapID, track.Title, track.Description, track.Status, track.Rank, track.CreatedAt, track.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert legacy track: %w", err)
		}
	}

	// Migrate task files
	migratedCount := 0
	for _, entry := range taskFiles {
		taskPath := filepath.Join(tasksDir, entry.Name())
		data, err := os.ReadFile(taskPath)
		if err != nil {
			// Log error but continue with next file
			fmt.Printf("Warning: failed to read task file %s: %v\n", entry.Name(), err)
			continue
		}

		// Unmarshal JSON
		var oldTask TaskEntity
		if err := json.Unmarshal(data, &oldTask); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to parse task file %s: %v\n", entry.Name(), err)
			continue
		}

		// Insert into database (force legacy track assignment)
		_, err = db.Exec(
			"INSERT INTO tasks (id, track_id, title, description, status, priority, branch, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			oldTask.ID, legacyTrackID, oldTask.Title, oldTask.Description, oldTask.Status, oldTask.Rank, oldTask.Branch, oldTask.CreatedAt, oldTask.UpdatedAt,
		)
		if err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to migrate task %s: %v\n", oldTask.ID, err)
			continue
		}

		migratedCount++
	}

	if migratedCount > 0 {
		fmt.Printf("Migrated %d tasks to database\n", migratedCount)
	}

	return nil
}

// GetCurrentTime returns the current time in UTC.
// This is a helper function for consistent timestamp handling.
func GetCurrentTime() time.Time {
	return time.Now().UTC()
}

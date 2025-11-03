package task_manager_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// TestMigrateV4ToV5 tests the v4 to v5 migration (adding testing_instructions column)
func TestMigrateV4ToV5(t *testing.T) {
	// Create a v4 database schema (without testing_instructions)
	dbPath := filepath.Join(t.TempDir(), "test_migrate.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	// Create v4 schema manually (without testing_instructions column)
	createV4Schema := `
	CREATE TABLE IF NOT EXISTS project_metadata (
	    key TEXT PRIMARY KEY,
	    value TEXT NOT NULL
	);
	INSERT INTO project_metadata (key, value) VALUES ('schema_version', '4');

	CREATE TABLE IF NOT EXISTS roadmaps (
	    id TEXT PRIMARY KEY,
	    vision TEXT NOT NULL,
	    success_criteria TEXT NOT NULL,
	    created_at TIMESTAMP NOT NULL,
	    updated_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tracks (
	    id TEXT PRIMARY KEY,
	    roadmap_id TEXT NOT NULL,
	    title TEXT NOT NULL,
	    description TEXT,
	    status TEXT NOT NULL,
	    rank INTEGER NOT NULL DEFAULT 500,
	    created_at TIMESTAMP NOT NULL,
	    updated_at TIMESTAMP NOT NULL,
	    FOREIGN KEY(roadmap_id) REFERENCES roadmaps(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS track_dependencies (
	    track_id TEXT NOT NULL,
	    depends_on_id TEXT NOT NULL,
	    PRIMARY KEY (track_id, depends_on_id),
	    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
	    FOREIGN KEY (depends_on_id) REFERENCES tracks(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS tasks (
	    id TEXT PRIMARY KEY,
	    track_id TEXT NOT NULL,
	    title TEXT NOT NULL,
	    description TEXT,
	    status TEXT NOT NULL,
	    rank INTEGER NOT NULL DEFAULT 500,
	    branch TEXT,
	    created_at TIMESTAMP NOT NULL,
	    updated_at TIMESTAMP NOT NULL,
	    FOREIGN KEY(track_id) REFERENCES tracks(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS iterations (
	    number INTEGER PRIMARY KEY,
	    name TEXT NOT NULL,
	    goal TEXT,
	    status TEXT NOT NULL,
	    rank INTEGER NOT NULL DEFAULT 500,
	    deliverable TEXT,
	    started_at TIMESTAMP,
	    completed_at TIMESTAMP,
	    created_at TIMESTAMP NOT NULL,
	    updated_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS iteration_tasks (
	    iteration_number INTEGER NOT NULL,
	    task_id TEXT NOT NULL,
	    PRIMARY KEY (iteration_number, task_id),
	    FOREIGN KEY (iteration_number) REFERENCES iterations(number) ON DELETE CASCADE,
	    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

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
	);

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
	);

	CREATE INDEX IF NOT EXISTS idx_ac_task_id ON acceptance_criteria(task_id);
	CREATE INDEX IF NOT EXISTS idx_ac_status ON acceptance_criteria(status);
	`

	if _, err := db.Exec(createV4Schema); err != nil {
		t.Fatalf("failed to create v4 schema: %v", err)
	}

	// Insert sample data
	now := time.Now().UTC()
	roadmapID := "roadmap-1"
	trackID := "track-1"
	taskID := "task-1"
	acID := "ac-1"

	if _, err := db.Exec("INSERT INTO roadmaps (id, vision, success_criteria, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		roadmapID, "Test vision", "Test criteria", now, now); err != nil {
		t.Fatalf("failed to insert roadmap: %v", err)
	}

	if _, err := db.Exec("INSERT INTO tracks (id, roadmap_id, title, description, status, rank, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		trackID, roadmapID, "Test track", "Description", "not-started", 500, now, now); err != nil {
		t.Fatalf("failed to insert track: %v", err)
	}

	if _, err := db.Exec("INSERT INTO tasks (id, track_id, title, description, status, rank, branch, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		taskID, trackID, "Test task", "Description", "todo", 500, "", now, now); err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	if _, err := db.Exec("INSERT INTO acceptance_criteria (id, task_id, description, verification_type, status, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		acID, taskID, "Test AC", "manual", "not_started", "Notes", now, now); err != nil {
		t.Fatalf("failed to insert AC: %v", err)
	}

	// Verify AC doesn't have testing_instructions column before migration
	rows, err := db.Query("PRAGMA table_info(acceptance_criteria)")
	if err != nil {
		t.Fatalf("failed to check table info: %v", err)
	}
	defer rows.Close()

	hasTestingInstructions := false
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed to scan column info: %v", err)
		}
		if name == "testing_instructions" {
			hasTestingInstructions = true
			break
		}
	}
	rows.Close()

	if hasTestingInstructions {
		t.Fatal("testing_instructions column should not exist before migration")
	}

	// Run migration
	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	// Verify testing_instructions column exists after migration
	rows, err = db.Query("PRAGMA table_info(acceptance_criteria)")
	if err != nil {
		t.Fatalf("failed to check table info after migration: %v", err)
	}
	defer rows.Close()

	hasTestingInstructions = false
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed to scan column info: %v", err)
		}
		if name == "testing_instructions" {
			hasTestingInstructions = true
			break
		}
	}
	rows.Close()

	if !hasTestingInstructions {
		t.Fatal("testing_instructions column should exist after migration")
	}

	// Verify data was preserved
	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	ac, err := repo.GetAC(ctx, acID)
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}

	if ac.ID != acID {
		t.Errorf("expected AC ID %s, got %s", acID, ac.ID)
	}
	if ac.Description != "Test AC" {
		t.Errorf("expected description 'Test AC', got '%s'", ac.Description)
	}
	if ac.TestingInstructions != "" {
		t.Errorf("expected empty TestingInstructions, got '%s'", ac.TestingInstructions)
	}

	// Verify schema version is updated to 5
	var version string
	err = db.QueryRow("SELECT value FROM project_metadata WHERE key = 'schema_version'").Scan(&version)
	if err != nil {
		t.Fatalf("failed to get schema version: %v", err)
	}
	if version != "5" {
		t.Errorf("expected schema version 5, got %s", version)
	}
}

// TestAcceptanceCriteriaSaveAndRetrieveWithTestingInstructions tests saving and retrieving AC with testing_instructions
func TestAcceptanceCriteriaSaveAndRetrieveWithTestingInstructions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_ac.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := task_manager.InitSchema(db); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	repo := task_manager.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create test data
	now := time.Now().UTC()
	roadmapID := "roadmap-1"
	trackID := "track-1"
	taskID := "task-1"

	roadmap, err := task_manager.NewRoadmapEntity(roadmapID, "Vision", "Criteria", now, now)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	repo.SaveRoadmap(ctx, roadmap)

	track, err := task_manager.NewTrackEntity(trackID, roadmapID, "Track", "Description", "not-started", 500, []string{}, now, now)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	repo.SaveTrack(ctx, track)

	task := task_manager.NewTaskEntity(taskID, trackID, "Task", "Description", "todo", 500, "", now, now)
	repo.SaveTask(ctx, task)

	// Create AC with testing_instructions
	testingInstructions := "1. Do this\n2. Then that\n3. Verify result"
	ac := task_manager.NewAcceptanceCriteriaEntity("ac-1", taskID, "Test AC", task_manager.VerificationTypeManual, testingInstructions, now, now)
	if err := repo.SaveAC(ctx, ac); err != nil {
		t.Fatalf("failed to save AC: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetAC(ctx, "ac-1")
	if err != nil {
		t.Fatalf("failed to get AC: %v", err)
	}

	if retrieved.TestingInstructions != testingInstructions {
		t.Errorf("expected TestingInstructions '%s', got '%s'", testingInstructions, retrieved.TestingInstructions)
	}

	// Update testing_instructions
	newInstructions := "Updated instructions"
	retrieved.TestingInstructions = newInstructions
	if err := repo.UpdateAC(ctx, retrieved); err != nil {
		t.Fatalf("failed to update AC: %v", err)
	}

	// Verify update
	updated, err := repo.GetAC(ctx, "ac-1")
	if err != nil {
		t.Fatalf("failed to get updated AC: %v", err)
	}

	if updated.TestingInstructions != newInstructions {
		t.Errorf("expected updated TestingInstructions '%s', got '%s'", newInstructions, updated.TestingInstructions)
	}
}

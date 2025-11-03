package task_manager

import (
	"database/sql"
	"fmt"
)

// migrateV4ToV5 migrates database from schema version 4 to version 5
// Adds testing_instructions column to acceptance_criteria table
func migrateV4ToV5(db *sql.DB) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if acceptance_criteria table has testing_instructions column (version 5)
	var hasTestingInstructions bool
	rows, err := tx.Query("PRAGMA table_info(acceptance_criteria)")
	if err != nil {
		return fmt.Errorf("failed to check acceptance_criteria table: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("failed to scan column info: %w", err)
		}
		if name == "testing_instructions" {
			hasTestingInstructions = true
			break
		}
	}
	rows.Close()

	if hasTestingInstructions {
		// Already migrated or new database
		return tx.Commit()
	}

	fmt.Println("Migrating database from schema v4 to v5 (adding testing_instructions column)...")

	// MIGRATE ACCEPTANCE_CRITERIA TABLE
	// 1. Create new acceptance_criteria table with testing_instructions
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS acceptance_criteria_new (
			id TEXT PRIMARY KEY,
			task_id TEXT NOT NULL,
			description TEXT NOT NULL,
			verification_type TEXT NOT NULL,
			status TEXT NOT NULL,
			notes TEXT,
			testing_instructions TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new acceptance_criteria table: %w", err)
	}

	// 2. Copy data from old table to new table
	_, err = tx.Exec(`
		INSERT INTO acceptance_criteria_new (id, task_id, description, verification_type, status, notes, testing_instructions, created_at, updated_at)
		SELECT
			id,
			task_id,
			description,
			verification_type,
			status,
			notes,
			'' as testing_instructions,
			created_at,
			updated_at
		FROM acceptance_criteria
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate acceptance_criteria data: %w", err)
	}

	// 3. Drop old table and rename new one
	if _, err = tx.Exec("DROP TABLE acceptance_criteria"); err != nil {
		return fmt.Errorf("failed to drop old acceptance_criteria table: %w", err)
	}
	if _, err = tx.Exec("ALTER TABLE acceptance_criteria_new RENAME TO acceptance_criteria"); err != nil {
		return fmt.Errorf("failed to rename new acceptance_criteria table: %w", err)
	}

	// 4. Recreate indexes
	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_ac_task_id ON acceptance_criteria(task_id)")
	if err != nil {
		return fmt.Errorf("failed to create ac_task_id index: %w", err)
	}

	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_ac_status ON acceptance_criteria(status)")
	if err != nil {
		return fmt.Errorf("failed to create ac_status index: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	fmt.Println("✓ Migration to schema v5 complete!")
	return nil
}

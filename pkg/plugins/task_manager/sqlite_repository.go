package task_manager

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// SQLiteRoadmapRepository implements RoadmapRepository using SQLite as the backend.
type SQLiteRoadmapRepository struct {
	db     *sql.DB
	logger pluginsdk.Logger
}

// NewSQLiteRoadmapRepository creates a new SQLite-backed repository.
func NewSQLiteRoadmapRepository(db *sql.DB, logger pluginsdk.Logger) *SQLiteRoadmapRepository {
	return &SQLiteRoadmapRepository{
		db:     db,
		logger: logger,
	}
}

// ============================================================================
// Roadmap Operations
// ============================================================================

// SaveRoadmap persists a new roadmap to storage.
func (r *SQLiteRoadmapRepository) SaveRoadmap(ctx context.Context, roadmap *RoadmapEntity) error {
	// Check if roadmap already exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM roadmaps WHERE id = ?", roadmap.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check roadmap existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: roadmap %s already exists", pluginsdk.ErrAlreadyExists, roadmap.ID)
	}

	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO roadmaps (id, vision, success_criteria, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		roadmap.ID, roadmap.Vision, roadmap.SuccessCriteria, roadmap.CreatedAt, roadmap.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert roadmap: %w", err)
	}

	return nil
}

// GetRoadmap retrieves a roadmap by its ID.
func (r *SQLiteRoadmapRepository) GetRoadmap(ctx context.Context, id string) (*RoadmapEntity, error) {
	var roadmap RoadmapEntity

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, vision, success_criteria, created_at, updated_at FROM roadmaps WHERE id = ?",
		id,
	).Scan(&roadmap.ID, &roadmap.Vision, &roadmap.SuccessCriteria, &roadmap.CreatedAt, &roadmap.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: roadmap %s not found", pluginsdk.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to query roadmap: %w", err)
	}

	return &roadmap, nil
}

// GetActiveRoadmap retrieves the most recently created roadmap.
func (r *SQLiteRoadmapRepository) GetActiveRoadmap(ctx context.Context) (*RoadmapEntity, error) {
	var roadmap RoadmapEntity

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, vision, success_criteria, created_at, updated_at FROM roadmaps ORDER BY created_at DESC LIMIT 1",
	).Scan(&roadmap.ID, &roadmap.Vision, &roadmap.SuccessCriteria, &roadmap.CreatedAt, &roadmap.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: no active roadmap found", pluginsdk.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to query active roadmap: %w", err)
	}

	return &roadmap, nil
}

// UpdateRoadmap updates an existing roadmap.
func (r *SQLiteRoadmapRepository) UpdateRoadmap(ctx context.Context, roadmap *RoadmapEntity) error {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE roadmaps SET vision = ?, success_criteria = ?, updated_at = ? WHERE id = ?",
		roadmap.Vision, roadmap.SuccessCriteria, roadmap.UpdatedAt, roadmap.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update roadmap: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: roadmap %s not found", pluginsdk.ErrNotFound, roadmap.ID)
	}

	return nil
}

// ============================================================================
// Track Operations
// ============================================================================

// SaveTrack persists a new track to storage.
func (r *SQLiteRoadmapRepository) SaveTrack(ctx context.Context, track *TrackEntity) error {
	// Check if track already exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tracks WHERE id = ?", track.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check track existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: track %s already exists", pluginsdk.ErrAlreadyExists, track.ID)
	}

	// Check if roadmap exists
	var roadmapExists int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM roadmaps WHERE id = ?", track.RoadmapID).Scan(&roadmapExists)
	if err != nil {
		return fmt.Errorf("failed to check roadmap existence: %w", err)
	}
	if roadmapExists == 0 {
		return fmt.Errorf("%w: roadmap %s not found", pluginsdk.ErrNotFound, track.RoadmapID)
	}

	// Start transaction for track and dependencies
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert track
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO tracks (id, roadmap_id, title, description, status, rank, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		track.ID, track.RoadmapID, track.Title, track.Description, track.Status, track.Rank, track.CreatedAt, track.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert track: %w", err)
	}

	// Insert dependencies
	for _, depID := range track.Dependencies {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO track_dependencies (track_id, depends_on_id) VALUES (?, ?)",
			track.ID, depID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert track dependency: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTrack retrieves a track by its ID.
func (r *SQLiteRoadmapRepository) GetTrack(ctx context.Context, id string) (*TrackEntity, error) {
	var track TrackEntity

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, roadmap_id, title, description, status, rank, created_at, updated_at FROM tracks WHERE id = ?",
		id,
	).Scan(&track.ID, &track.RoadmapID, &track.Title, &track.Description, &track.Status, &track.Rank, &track.CreatedAt, &track.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: track %s not found", pluginsdk.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to query track: %w", err)
	}

	// Load dependencies
	deps, err := r.GetTrackDependencies(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load track dependencies: %w", err)
	}
	track.Dependencies = deps

	return &track, nil
}

// ListTracks returns all tracks for a roadmap, optionally filtered.
func (r *SQLiteRoadmapRepository) ListTracks(ctx context.Context, roadmapID string, filters TrackFilters) ([]*TrackEntity, error) {
	query := "SELECT id, roadmap_id, title, description, status, rank, created_at, updated_at FROM tracks WHERE roadmap_id = ?"
	args := []interface{}{roadmapID}

	// Add status filter if provided
	if len(filters.Status) > 0 {
		placeholders := ""
		for i := range filters.Status {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, filters.Status[i])
		}
		query += " AND status IN (" + placeholders + ")"
	}

	// Add priority filter if provided
	if len(filters.Priority) > 0 {
		placeholders := ""
		for i := range filters.Priority {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, filters.Priority[i])
		}
		query += " AND rank IN (" + placeholders + ")"
	}

	query += " ORDER BY id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracks: %w", err)
	}
	defer rows.Close()

	var tracks []*TrackEntity
	for rows.Next() {
		var track TrackEntity
		err := rows.Scan(&track.ID, &track.RoadmapID, &track.Title, &track.Description, &track.Status, &track.Rank, &track.CreatedAt, &track.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}

		// Load dependencies
		deps, err := r.GetTrackDependencies(ctx, track.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load track dependencies: %w", err)
		}
		track.Dependencies = deps

		tracks = append(tracks, &track)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tracks: %w", err)
	}

	return tracks, nil
}

// UpdateTrack updates an existing track.
func (r *SQLiteRoadmapRepository) UpdateTrack(ctx context.Context, track *TrackEntity) error {
	// Start transaction for track and dependencies update
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update track fields
	result, err := tx.ExecContext(
		ctx,
		"UPDATE tracks SET title = ?, description = ?, status = ?, rank = ?, updated_at = ? WHERE id = ?",
		track.Title, track.Description, track.Status, track.Rank, track.UpdatedAt, track.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update track: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: track %s not found", pluginsdk.ErrNotFound, track.ID)
	}

	// Delete existing dependencies
	_, err = tx.ExecContext(ctx, "DELETE FROM track_dependencies WHERE track_id = ?", track.ID)
	if err != nil {
		return fmt.Errorf("failed to delete dependencies: %w", err)
	}

	// Insert new dependencies
	for _, depID := range track.Dependencies {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO track_dependencies (track_id, depends_on_id) VALUES (?, ?)",
			track.ID, depID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert dependency: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteTrack removes a track from storage.
func (r *SQLiteRoadmapRepository) DeleteTrack(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tracks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete track: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: track %s not found", pluginsdk.ErrNotFound, id)
	}

	return nil
}

// AddTrackDependency adds a dependency from trackID to dependsOnID.
func (r *SQLiteRoadmapRepository) AddTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	// Check for self-dependency
	if trackID == dependsOnID {
		return fmt.Errorf("%w: track cannot depend on itself", pluginsdk.ErrInvalidArgument)
	}

	// Check both tracks exist
	for _, id := range []string{trackID, dependsOnID} {
		var exists int
		err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tracks WHERE id = ?", id).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check track existence: %w", err)
		}
		if exists == 0 {
			return fmt.Errorf("%w: track %s not found", pluginsdk.ErrNotFound, id)
		}
	}

	// Check if dependency already exists
	var exists int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM track_dependencies WHERE track_id = ? AND depends_on_id = ?",
		trackID, dependsOnID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check dependency existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: dependency already exists", pluginsdk.ErrAlreadyExists)
	}

	// Insert dependency
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO track_dependencies (track_id, depends_on_id) VALUES (?, ?)",
		trackID, dependsOnID,
	)
	if err != nil {
		return fmt.Errorf("failed to add dependency: %w", err)
	}

	return nil
}

// RemoveTrackDependency removes a dependency from trackID to dependsOnID.
func (r *SQLiteRoadmapRepository) RemoveTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	result, err := r.db.ExecContext(
		ctx,
		"DELETE FROM track_dependencies WHERE track_id = ? AND depends_on_id = ?",
		trackID, dependsOnID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove dependency: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: dependency not found", pluginsdk.ErrNotFound)
	}

	return nil
}

// GetTrackDependencies returns the IDs of all tracks that trackID depends on.
func (r *SQLiteRoadmapRepository) GetTrackDependencies(ctx context.Context, trackID string) ([]string, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT depends_on_id FROM track_dependencies WHERE track_id = ? ORDER BY depends_on_id",
		trackID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependencies: %w", err)
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var depID string
		if err := rows.Scan(&depID); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		deps = append(deps, depID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dependencies: %w", err)
	}

	return deps, nil
}

// ValidateNoCycles checks if adding/updating the track would create a circular dependency.
// Uses depth-first search to detect cycles.
func (r *SQLiteRoadmapRepository) ValidateNoCycles(ctx context.Context, trackID string) error {
	// Use DFS to detect cycles
	visited := make(map[string]bool)
	return r.detectCycleDFS(ctx, trackID, visited)
}

// detectCycleDFS performs depth-first search to detect cycles.
func (r *SQLiteRoadmapRepository) detectCycleDFS(ctx context.Context, trackID string, visited map[string]bool) error {
	if visited[trackID] {
		return fmt.Errorf("%w: circular dependency detected", pluginsdk.ErrInvalidArgument)
	}

	visited[trackID] = true

	deps, err := r.GetTrackDependencies(ctx, trackID)
	if err != nil {
		return err
	}

	for _, depID := range deps {
		if err := r.detectCycleDFS(ctx, depID, visited); err != nil {
			return err
		}
	}

	visited[trackID] = false
	return nil
}

// ============================================================================
// Task Operations
// ============================================================================

// SaveTask persists a new task to storage.
func (r *SQLiteRoadmapRepository) SaveTask(ctx context.Context, task *TaskEntity) error {
	// Check if task already exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", task.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: task %s already exists", pluginsdk.ErrAlreadyExists, task.ID)
	}

	// Check if track exists
	var trackExists int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tracks WHERE id = ?", task.TrackID).Scan(&trackExists)
	if err != nil {
		return fmt.Errorf("failed to check track existence: %w", err)
	}
	if trackExists == 0 {
		return fmt.Errorf("%w: track %s not found", pluginsdk.ErrNotFound, task.TrackID)
	}

	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO tasks (id, track_id, title, description, status, rank, branch, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		task.ID, task.TrackID, task.Title, task.Description, task.Status, task.Rank, task.Branch, task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

// GetTask retrieves a task by its ID.
func (r *SQLiteRoadmapRepository) GetTask(ctx context.Context, id string) (*TaskEntity, error) {
	var task TaskEntity
	var branch sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, track_id, title, description, status, rank, branch, created_at, updated_at FROM tasks WHERE id = ?",
		id,
	).Scan(&task.ID, &task.TrackID, &task.Title, &task.Description, &task.Status, &task.Rank, &branch, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: task %s not found", pluginsdk.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	if branch.Valid {
		task.Branch = branch.String
	}

	return &task, nil
}

// ListTasks returns all tasks matching the filters.
func (r *SQLiteRoadmapRepository) ListTasks(ctx context.Context, filters TaskFilters) ([]*TaskEntity, error) {
	query := "SELECT id, track_id, title, description, status, rank, branch, created_at, updated_at FROM tasks WHERE 1=1"
	args := []interface{}{}

	// Add track filter if provided
	if filters.TrackID != "" {
		query += " AND track_id = ?"
		args = append(args, filters.TrackID)
	}

	// Add status filter if provided
	if len(filters.Status) > 0 {
		placeholders := ""
		for i := range filters.Status {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, filters.Status[i])
		}
		query += " AND status IN (" + placeholders + ")"
	}

	// Add priority filter if provided
	if len(filters.Priority) > 0 {
		placeholders := ""
		for i := range filters.Priority {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, filters.Priority[i])
		}
		query += " AND rank IN (" + placeholders + ")"
	}

	query += " ORDER BY id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*TaskEntity
	for rows.Next() {
		var task TaskEntity
		var branch sql.NullString

		err := rows.Scan(&task.ID, &task.TrackID, &task.Title, &task.Description, &task.Status, &task.Rank, &branch, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if branch.Valid {
			task.Branch = branch.String
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTask updates an existing task.
func (r *SQLiteRoadmapRepository) UpdateTask(ctx context.Context, task *TaskEntity) error {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE tasks SET track_id = ?, title = ?, description = ?, status = ?, rank = ?, branch = ?, updated_at = ? WHERE id = ?",
		task.TrackID, task.Title, task.Description, task.Status, task.Rank, task.Branch, task.UpdatedAt, task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: task %s not found", pluginsdk.ErrNotFound, task.ID)
	}

	return nil
}

// DeleteTask removes a task from storage.
func (r *SQLiteRoadmapRepository) DeleteTask(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: task %s not found", pluginsdk.ErrNotFound, id)
	}

	return nil
}

// MoveTaskToTrack moves a task from its current track to a new track.
func (r *SQLiteRoadmapRepository) MoveTaskToTrack(ctx context.Context, taskID, newTrackID string) error {
	// Check if task exists
	task, err := r.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	// Check if new track exists
	var trackExists int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tracks WHERE id = ?", newTrackID).Scan(&trackExists)
	if err != nil {
		return fmt.Errorf("failed to check track existence: %w", err)
	}
	if trackExists == 0 {
		return fmt.Errorf("%w: track %s not found", pluginsdk.ErrNotFound, newTrackID)
	}

	// Update task's track
	task.TrackID = newTrackID
	task.UpdatedAt = time.Now().UTC()
	return r.UpdateTask(ctx, task)
}

// ============================================================================
// Iteration Operations
// ============================================================================

// SaveIteration persists a new iteration to storage.
func (r *SQLiteRoadmapRepository) SaveIteration(ctx context.Context, iteration *IterationEntity) error {
	// Check if iteration already exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM iterations WHERE number = ?", iteration.Number).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check iteration existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: iteration %d already exists", pluginsdk.ErrAlreadyExists, iteration.Number)
	}

	// Start transaction for iteration and tasks
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert iteration
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO iterations (number, name, goal, status, deliverable, started_at, completed_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		iteration.Number, iteration.Name, iteration.Goal, iteration.Status, iteration.Deliverable, iteration.StartedAt, iteration.CompletedAt, iteration.CreatedAt, iteration.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert iteration: %w", err)
	}

	// Insert task associations
	for _, taskID := range iteration.TaskIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO iteration_tasks (iteration_number, task_id) VALUES (?, ?)",
			iteration.Number, taskID,
		)
		if err != nil {
			return fmt.Errorf("failed to add task to iteration: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetIteration retrieves an iteration by its number.
func (r *SQLiteRoadmapRepository) GetIteration(ctx context.Context, number int) (*IterationEntity, error) {
	var iteration IterationEntity
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		"SELECT number, name, goal, status, rank, deliverable, started_at, completed_at, created_at, updated_at FROM iterations WHERE number = ?",
		number,
	).Scan(&iteration.Number, &iteration.Name, &iteration.Goal, &iteration.Status, &iteration.Rank, &iteration.Deliverable, &startedAt, &completedAt, &iteration.CreatedAt, &iteration.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: iteration %d not found", pluginsdk.ErrNotFound, number)
		}
		return nil, fmt.Errorf("failed to query iteration: %w", err)
	}

	if startedAt.Valid {
		iteration.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		iteration.CompletedAt = &completedAt.Time
	}

	// Load task IDs
	taskIDs, err := r.getIterationTaskIDs(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to load iteration tasks: %w", err)
	}
	iteration.TaskIDs = taskIDs

	return &iteration, nil
}

// GetCurrentIteration returns the iteration with status "current".
func (r *SQLiteRoadmapRepository) GetCurrentIteration(ctx context.Context) (*IterationEntity, error) {
	var iteration IterationEntity
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		"SELECT number, name, goal, status, rank, deliverable, started_at, completed_at, created_at, updated_at FROM iterations WHERE status = ? LIMIT 1",
		"current",
	).Scan(&iteration.Number, &iteration.Name, &iteration.Goal, &iteration.Status, &iteration.Rank, &iteration.Deliverable, &startedAt, &completedAt, &iteration.CreatedAt, &iteration.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: no current iteration found", pluginsdk.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to query current iteration: %w", err)
	}

	if startedAt.Valid {
		iteration.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		iteration.CompletedAt = &completedAt.Time
	}

	// Load task IDs
	taskIDs, err := r.getIterationTaskIDs(ctx, iteration.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to load iteration tasks: %w", err)
	}
	iteration.TaskIDs = taskIDs

	return &iteration, nil
}

// ListIterations returns all iterations, ordered by rank (then number).
func (r *SQLiteRoadmapRepository) ListIterations(ctx context.Context) ([]*IterationEntity, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT number, name, goal, status, rank, deliverable, started_at, completed_at, created_at, updated_at FROM iterations ORDER BY rank, number",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query iterations: %w", err)
	}
	defer rows.Close()

	var iterations []*IterationEntity
	for rows.Next() {
		var iteration IterationEntity
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(&iteration.Number, &iteration.Name, &iteration.Goal, &iteration.Status, &iteration.Rank, &iteration.Deliverable, &startedAt, &completedAt, &iteration.CreatedAt, &iteration.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan iteration: %w", err)
		}

		if startedAt.Valid {
			iteration.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			iteration.CompletedAt = &completedAt.Time
		}

		// Load task IDs
		taskIDs, err := r.getIterationTaskIDs(ctx, iteration.Number)
		if err != nil {
			return nil, fmt.Errorf("failed to load iteration tasks: %w", err)
		}
		iteration.TaskIDs = taskIDs

		iterations = append(iterations, &iteration)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating iterations: %w", err)
	}

	return iterations, nil
}

// UpdateIteration updates an existing iteration.
func (r *SQLiteRoadmapRepository) UpdateIteration(ctx context.Context, iteration *IterationEntity) error {
	// Start transaction for iteration and tasks update
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update iteration fields
	result, err := tx.ExecContext(
		ctx,
		"UPDATE iterations SET name = ?, goal = ?, status = ?, rank = ?, deliverable = ?, started_at = ?, completed_at = ?, updated_at = ? WHERE number = ?",
		iteration.Name, iteration.Goal, iteration.Status, iteration.Rank, iteration.Deliverable, iteration.StartedAt, iteration.CompletedAt, iteration.UpdatedAt, iteration.Number,
	)
	if err != nil {
		return fmt.Errorf("failed to update iteration: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: iteration %d not found", pluginsdk.ErrNotFound, iteration.Number)
	}

	// Delete existing task associations
	_, err = tx.ExecContext(ctx, "DELETE FROM iteration_tasks WHERE iteration_number = ?", iteration.Number)
	if err != nil {
		return fmt.Errorf("failed to delete task associations: %w", err)
	}

	// Insert new task associations
	for _, taskID := range iteration.TaskIDs {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO iteration_tasks (iteration_number, task_id) VALUES (?, ?)",
			iteration.Number, taskID,
		)
		if err != nil {
			return fmt.Errorf("failed to add task to iteration: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteIteration removes an iteration from storage.
func (r *SQLiteRoadmapRepository) DeleteIteration(ctx context.Context, number int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM iterations WHERE number = ?", number)
	if err != nil {
		return fmt.Errorf("failed to delete iteration: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: iteration %d not found", pluginsdk.ErrNotFound, number)
	}

	return nil
}

// AddTaskToIteration adds a task to an iteration.
func (r *SQLiteRoadmapRepository) AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error {
	// Check if iteration exists
	var iterExists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM iterations WHERE number = ?", iterationNum).Scan(&iterExists)
	if err != nil {
		return fmt.Errorf("failed to check iteration existence: %w", err)
	}
	if iterExists == 0 {
		return fmt.Errorf("%w: iteration %d not found", pluginsdk.ErrNotFound, iterationNum)
	}

	// Check if task exists
	var taskExists int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&taskExists)
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if taskExists == 0 {
		return fmt.Errorf("%w: task %s not found", pluginsdk.ErrNotFound, taskID)
	}

	// Check if task already in iteration
	var alreadyExists int
	err = r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM iteration_tasks WHERE iteration_number = ? AND task_id = ?",
		iterationNum, taskID,
	).Scan(&alreadyExists)
	if err != nil {
		return fmt.Errorf("failed to check task in iteration: %w", err)
	}
	if alreadyExists > 0 {
		return fmt.Errorf("%w: task already in iteration", pluginsdk.ErrAlreadyExists)
	}

	// Insert task association
	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO iteration_tasks (iteration_number, task_id) VALUES (?, ?)",
		iterationNum, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to add task to iteration: %w", err)
	}

	return nil
}

// RemoveTaskFromIteration removes a task from an iteration.
func (r *SQLiteRoadmapRepository) RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error {
	result, err := r.db.ExecContext(
		ctx,
		"DELETE FROM iteration_tasks WHERE iteration_number = ? AND task_id = ?",
		iterationNum, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove task from iteration: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: task not in iteration", pluginsdk.ErrNotFound)
	}

	return nil
}

// GetIterationTasks returns all tasks in an iteration.
func (r *SQLiteRoadmapRepository) GetIterationTasks(ctx context.Context, iterationNum int) ([]*TaskEntity, error) {
	// Check if iteration exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM iterations WHERE number = ?", iterationNum).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check iteration existence: %w", err)
	}
	if exists == 0 {
		return nil, fmt.Errorf("%w: iteration %d not found", pluginsdk.ErrNotFound, iterationNum)
	}

	taskIDs, err := r.getIterationTaskIDs(ctx, iterationNum)
	if err != nil {
		return nil, err
	}

	var tasks []*TaskEntity
	for _, taskID := range taskIDs {
		task, err := r.GetTask(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// StartIteration marks an iteration as current and sets started_at timestamp.
func (r *SQLiteRoadmapRepository) StartIteration(ctx context.Context, iterationNum int) error {
	// Get iteration first
	iteration, err := r.GetIteration(ctx, iterationNum)
	if err != nil {
		return err
	}

	// Check if status is "planned"
	if iteration.Status != "planned" {
		return fmt.Errorf("%w: iteration must be in planned status to start", pluginsdk.ErrInvalidArgument)
	}

	// Update status to current and set started_at
	now := time.Now().UTC()
	iteration.Status = "current"
	iteration.StartedAt = &now
	iteration.UpdatedAt = now

	return r.UpdateIteration(ctx, iteration)
}

// CompleteIteration marks an iteration as complete and sets completed_at timestamp.
func (r *SQLiteRoadmapRepository) CompleteIteration(ctx context.Context, iterationNum int) error {
	// Get iteration first
	iteration, err := r.GetIteration(ctx, iterationNum)
	if err != nil {
		return err
	}

	// Check if status is "current"
	if iteration.Status != "current" {
		return fmt.Errorf("%w: iteration must be in current status to complete", pluginsdk.ErrInvalidArgument)
	}

	// Update status to complete and set completed_at
	now := time.Now().UTC()
	iteration.Status = "complete"
	iteration.CompletedAt = &now
	iteration.UpdatedAt = now

	return r.UpdateIteration(ctx, iteration)
}

// ============================================================================
// Aggregate Queries
// ============================================================================

// GetRoadmapWithTracks retrieves a roadmap with all its tracks.
func (r *SQLiteRoadmapRepository) GetRoadmapWithTracks(ctx context.Context, roadmapID string) (*RoadmapEntity, error) {
	roadmap, err := r.GetRoadmap(ctx, roadmapID)
	if err != nil {
		return nil, err
	}

	// Load all tracks for this roadmap
	tracks, err := r.ListTracks(ctx, roadmapID, TrackFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to load tracks: %w", err)
	}

	// Store tracks in roadmap for retrieval (though RoadmapEntity doesn't have a Tracks field)
	// This is for the aggregate query to ensure all related data is loaded
	_ = tracks

	return roadmap, nil
}

// GetTrackWithTasks retrieves a track with all its tasks.
func (r *SQLiteRoadmapRepository) GetTrackWithTasks(ctx context.Context, trackID string) (*TrackEntity, error) {
	track, err := r.GetTrack(ctx, trackID)
	if err != nil {
		return nil, err
	}

	// Load all tasks for this track
	tasks, err := r.ListTasks(ctx, TaskFilters{TrackID: trackID})
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Store tasks for aggregate query (though TrackEntity doesn't have a Tasks field)
	_ = tasks

	return track, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// getIterationTaskIDs retrieves all task IDs for an iteration.
func (r *SQLiteRoadmapRepository) getIterationTaskIDs(ctx context.Context, iterationNum int) ([]string, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT task_id FROM iteration_tasks WHERE iteration_number = ? ORDER BY task_id",
		iterationNum,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query iteration tasks: %w", err)
	}
	defer rows.Close()

	var taskIDs []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return nil, fmt.Errorf("failed to scan task ID: %w", err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task IDs: %w", err)
	}

	return taskIDs, nil
}

// ============================================================================
// Project Metadata Operations
// ============================================================================

// GetProjectMetadata retrieves a metadata value by key.
func (r *SQLiteRoadmapRepository) GetProjectMetadata(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM project_metadata WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%w: metadata key %s not found", pluginsdk.ErrNotFound, key)
		}
		return "", fmt.Errorf("failed to query metadata: %w", err)
	}
	return value, nil
}

// SetProjectMetadata sets a metadata value by key.
func (r *SQLiteRoadmapRepository) SetProjectMetadata(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT OR REPLACE INTO project_metadata (key, value) VALUES (?, ?)",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}
	return nil
}

// GetProjectCode retrieves the project code (e.g., "DW" for darwinflow).
// Returns "DW" as default if not set.
func (r *SQLiteRoadmapRepository) GetProjectCode(ctx context.Context) string {
	code, err := r.GetProjectMetadata(ctx, "project_code")
	if err != nil {
		// Return default if not set
		return "DW"
	}
	return code
}

// GetNextSequenceNumber retrieves the next sequence number for an entity type.
// Entity types: "task", "track", "iter", "ac", "adr"
func (r *SQLiteRoadmapRepository) GetNextSequenceNumber(ctx context.Context, entityType string) (int, error) {
	var maxNum int
	var query string

	switch entityType {
	case "task":
		// Parse existing task IDs to find max number
		query = "SELECT id FROM tasks"
	case "track":
		// Parse existing track IDs to find max number
		query = "SELECT id FROM tracks"
	case "iter":
		// For iterations, use the number column directly
		err := r.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(number), 0) FROM iterations").Scan(&maxNum)
		if err != nil {
			return 0, fmt.Errorf("failed to get max iteration number: %w", err)
		}
		return maxNum + 1, nil
	case "ac":
		// Parse existing AC IDs to find max number
		query = "SELECT id FROM acceptance_criteria"
	case "adr":
		// Parse existing ADR IDs to find max number
		query = "SELECT id FROM adrs"
	default:
		return 0, fmt.Errorf("%w: invalid entity type: %s", pluginsdk.ErrInvalidArgument, entityType)
	}

	// For tasks and tracks, we need to parse IDs
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to query %s IDs: %w", entityType, err)
	}
	defer rows.Close()

	maxNum = 0
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("failed to scan ID: %w", err)
		}

		// Parse the numeric part from IDs like "DW-task-123" or "DW-track-5"
		// Format: {CODE}-{entity}-{number}
		// Split by "-" and parse the last part
		parts := strings.Split(id, "-")
		if len(parts) >= 3 {
			var num int
			_, err := fmt.Sscanf(parts[len(parts)-1], "%d", &num)
			if err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	if err = rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating IDs: %w", err)
	}

	return maxNum + 1, nil
}

// ============================================================================
// ADR Operations
// ============================================================================

// SaveADR persists a new ADR to storage.
func (r *SQLiteRoadmapRepository) SaveADR(ctx context.Context, adr *ADREntity) error {
	// Check if ADR already exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM adrs WHERE id = ?", adr.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check ADR existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: ADR %s already exists", pluginsdk.ErrAlreadyExists, adr.ID)
	}

	// Check if track exists
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tracks WHERE id = ?", adr.TrackID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check track existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("%w: track %s does not exist", pluginsdk.ErrNotFound, adr.TrackID)
	}

	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO adrs (id, track_id, title, status, context, decision, consequences, alternatives, created_at, updated_at, superseded_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		adr.ID, adr.TrackID, adr.Title, adr.Status, adr.Context, adr.Decision, adr.Consequences, adr.Alternatives, adr.CreatedAt, adr.UpdatedAt, adr.SupersededBy,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ADR: %w", err)
	}

	return nil
}

// GetADR retrieves an ADR by its ID.
func (r *SQLiteRoadmapRepository) GetADR(ctx context.Context, id string) (*ADREntity, error) {
	row := r.db.QueryRowContext(
		ctx,
		"SELECT id, track_id, title, status, context, decision, consequences, alternatives, created_at, updated_at, superseded_by FROM adrs WHERE id = ?",
		id,
	)

	var adr ADREntity
	var supersededBy sql.NullString
	err := row.Scan(
		&adr.ID, &adr.TrackID, &adr.Title, &adr.Status, &adr.Context, &adr.Decision, &adr.Consequences, &adr.Alternatives, &adr.CreatedAt, &adr.UpdatedAt, &supersededBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: ADR %s not found", pluginsdk.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to query ADR: %w", err)
	}

	if supersededBy.Valid {
		adr.SupersededBy = &supersededBy.String
	}

	return &adr, nil
}

// ListADRs returns all ADRs, optionally filtered by track.
func (r *SQLiteRoadmapRepository) ListADRs(ctx context.Context, trackID *string) ([]*ADREntity, error) {
	query := "SELECT id, track_id, title, status, context, decision, consequences, alternatives, created_at, updated_at, superseded_by FROM adrs"
	var args []interface{}

	if trackID != nil {
		query += " WHERE track_id = ?"
		args = append(args, *trackID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ADRs: %w", err)
	}
	defer rows.Close()

	var adrs []*ADREntity
	for rows.Next() {
		var adr ADREntity
		var supersededBy sql.NullString
		err := rows.Scan(
			&adr.ID, &adr.TrackID, &adr.Title, &adr.Status, &adr.Context, &adr.Decision, &adr.Consequences, &adr.Alternatives, &adr.CreatedAt, &adr.UpdatedAt, &supersededBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ADR: %w", err)
		}

		if supersededBy.Valid {
			adr.SupersededBy = &supersededBy.String
		}

		adrs = append(adrs, &adr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ADRs: %w", err)
	}

	return adrs, nil
}

// UpdateADR updates an existing ADR.
func (r *SQLiteRoadmapRepository) UpdateADR(ctx context.Context, adr *ADREntity) error {
	// Check if ADR exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM adrs WHERE id = ?", adr.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check ADR existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("%w: ADR %s not found", pluginsdk.ErrNotFound, adr.ID)
	}

	_, err = r.db.ExecContext(
		ctx,
		"UPDATE adrs SET title = ?, status = ?, context = ?, decision = ?, consequences = ?, alternatives = ?, updated_at = ?, superseded_by = ? WHERE id = ?",
		adr.Title, adr.Status, adr.Context, adr.Decision, adr.Consequences, adr.Alternatives, adr.UpdatedAt, adr.SupersededBy, adr.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update ADR: %w", err)
	}

	return nil
}

// SupersedeADR marks an ADR as superseded by another ADR.
func (r *SQLiteRoadmapRepository) SupersedeADR(ctx context.Context, adrID, supersededByID string) error {
	// Check if both ADRs exist
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM adrs WHERE id = ?", adrID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check ADR existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("%w: ADR %s not found", pluginsdk.ErrNotFound, adrID)
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM adrs WHERE id = ?", supersededByID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check superseding ADR existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("%w: ADR %s not found", pluginsdk.ErrNotFound, supersededByID)
	}

	now := time.Now().UTC()
	_, err = r.db.ExecContext(
		ctx,
		"UPDATE adrs SET status = ?, superseded_by = ?, updated_at = ? WHERE id = ?",
		string(ADRStatusSuperseded), supersededByID, now, adrID,
	)
	if err != nil {
		return fmt.Errorf("failed to supersede ADR: %w", err)
	}

	return nil
}

// DeprecateADR marks an ADR as deprecated.
func (r *SQLiteRoadmapRepository) DeprecateADR(ctx context.Context, adrID string) error {
	// Check if ADR exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM adrs WHERE id = ?", adrID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check ADR existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("%w: ADR %s not found", pluginsdk.ErrNotFound, adrID)
	}

	now := time.Now().UTC()
	_, err = r.db.ExecContext(
		ctx,
		"UPDATE adrs SET status = ?, updated_at = ? WHERE id = ?",
		string(ADRStatusDeprecated), now, adrID,
	)
	if err != nil {
		return fmt.Errorf("failed to deprecate ADR: %w", err)
	}

	return nil
}

// GetADRsByTrack returns all ADRs for a specific track.
func (r *SQLiteRoadmapRepository) GetADRsByTrack(ctx context.Context, trackID string) ([]*ADREntity, error) {
	return r.ListADRs(ctx, &trackID)
}

// ============================================================================
// Acceptance Criteria Operations
// ============================================================================

// SaveAC persists a new acceptance criterion to storage.
func (r *SQLiteRoadmapRepository) SaveAC(ctx context.Context, ac *AcceptanceCriteriaEntity) error {
	// Check if AC already exists
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM acceptance_criteria WHERE id = ?", ac.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check AC existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: AC %s already exists", pluginsdk.ErrAlreadyExists, ac.ID)
	}

	// Verify task exists
	var taskExists int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", ac.TaskID).Scan(&taskExists)
	if err != nil {
		return fmt.Errorf("failed to verify task: %w", err)
	}
	if taskExists == 0 {
		return fmt.Errorf("%w: task %s not found", pluginsdk.ErrNotFound, ac.TaskID)
	}

	_, err = r.db.ExecContext(
		ctx,
		"INSERT INTO acceptance_criteria (id, task_id, description, verification_type, status, notes, testing_instructions, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ac.ID, ac.TaskID, ac.Description, string(ac.VerificationType), string(ac.Status), ac.Notes, ac.TestingInstructions, ac.CreatedAt, ac.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert AC: %w", err)
	}

	return nil
}

// GetAC retrieves an acceptance criterion by its ID.
func (r *SQLiteRoadmapRepository) GetAC(ctx context.Context, id string) (*AcceptanceCriteriaEntity, error) {
	var ac AcceptanceCriteriaEntity

	var testingInstructions sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, task_id, description, verification_type, status, notes, testing_instructions, created_at, updated_at FROM acceptance_criteria WHERE id = ?",
		id,
	).Scan(&ac.ID, &ac.TaskID, &ac.Description, (*string)(&ac.VerificationType), (*string)(&ac.Status), &ac.Notes, &testingInstructions, &ac.CreatedAt, &ac.UpdatedAt)

	if testingInstructions.Valid {
		ac.TestingInstructions = testingInstructions.String
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: AC %s not found", pluginsdk.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to query AC: %w", err)
	}

	return &ac, nil
}

// ListAC returns all acceptance criteria for a task.
func (r *SQLiteRoadmapRepository) ListAC(ctx context.Context, taskID string) ([]*AcceptanceCriteriaEntity, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, task_id, description, verification_type, status, notes, testing_instructions, created_at, updated_at FROM acceptance_criteria WHERE task_id = ? ORDER BY created_at ASC",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ACs: %w", err)
	}
	defer rows.Close()

	var acs []*AcceptanceCriteriaEntity
	for rows.Next() {
		var ac AcceptanceCriteriaEntity
		var testingInstructions sql.NullString
		err := rows.Scan(&ac.ID, &ac.TaskID, &ac.Description, (*string)(&ac.VerificationType), (*string)(&ac.Status), &ac.Notes, &testingInstructions, &ac.CreatedAt, &ac.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AC: %w", err)
		}
		if testingInstructions.Valid {
			ac.TestingInstructions = testingInstructions.String
		}
		acs = append(acs, &ac)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ACs: %w", err)
	}

	return acs, nil
}

// UpdateAC updates an existing acceptance criterion.
func (r *SQLiteRoadmapRepository) UpdateAC(ctx context.Context, ac *AcceptanceCriteriaEntity) error {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE acceptance_criteria SET task_id = ?, description = ?, verification_type = ?, status = ?, notes = ?, testing_instructions = ?, updated_at = ? WHERE id = ?",
		ac.TaskID, ac.Description, string(ac.VerificationType), string(ac.Status), ac.Notes, ac.TestingInstructions, ac.UpdatedAt, ac.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update AC: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: AC %s not found", pluginsdk.ErrNotFound, ac.ID)
	}

	return nil
}

// DeleteAC removes an acceptance criterion from storage.
func (r *SQLiteRoadmapRepository) DeleteAC(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM acceptance_criteria WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete AC: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: AC %s not found", pluginsdk.ErrNotFound, id)
	}

	return nil
}

// ListACByTrack returns all acceptance criteria for all tasks in a track.
func (r *SQLiteRoadmapRepository) ListACByTrack(ctx context.Context, trackID string) ([]*AcceptanceCriteriaEntity, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT ac.id, ac.task_id, ac.description, ac.verification_type, ac.status, ac.notes, ac.testing_instructions, ac.created_at, ac.updated_at
		 FROM acceptance_criteria ac
		 JOIN tasks t ON ac.task_id = t.id
		 WHERE t.track_id = ?
		 ORDER BY ac.created_at ASC`,
		trackID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ACs by track: %w", err)
	}
	defer rows.Close()

	var acs []*AcceptanceCriteriaEntity
	for rows.Next() {
		var ac AcceptanceCriteriaEntity
		var testingInstructions sql.NullString
		err := rows.Scan(&ac.ID, &ac.TaskID, &ac.Description, (*string)(&ac.VerificationType), (*string)(&ac.Status), &ac.Notes, &testingInstructions, &ac.CreatedAt, &ac.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AC: %w", err)
		}
		if testingInstructions.Valid {
			ac.TestingInstructions = testingInstructions.String
		}
		acs = append(acs, &ac)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ACs: %w", err)
	}

	return acs, nil
}

// ListACByIteration returns all acceptance criteria for all tasks in an iteration.
func (r *SQLiteRoadmapRepository) ListACByIteration(ctx context.Context, iterationNum int) ([]*AcceptanceCriteriaEntity, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT ac.id, ac.task_id, ac.description, ac.verification_type, ac.status, ac.notes, ac.testing_instructions, ac.created_at, ac.updated_at
		 FROM acceptance_criteria ac
		 JOIN tasks t ON ac.task_id = t.id
		 JOIN iteration_tasks it ON t.id = it.task_id
		 WHERE it.iteration_number = ?
		 ORDER BY ac.created_at ASC`,
		iterationNum,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ACs by iteration: %w", err)
	}
	defer rows.Close()

	var acs []*AcceptanceCriteriaEntity
	for rows.Next() {
		var ac AcceptanceCriteriaEntity
		var testingInstructions sql.NullString
		err := rows.Scan(&ac.ID, &ac.TaskID, &ac.Description, (*string)(&ac.VerificationType), (*string)(&ac.Status), &ac.Notes, &testingInstructions, &ac.CreatedAt, &ac.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AC: %w", err)
		}
		if testingInstructions.Valid {
			ac.TestingInstructions = testingInstructions.String
		}
		acs = append(acs, &ac)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ACs: %w", err)
	}

	return acs, nil
}

// ============================================================================
// New Query Methods for LLM Agent Integration
// ============================================================================

// GetIterationsForTask returns all iterations that contain a specific task.
func (r *SQLiteRoadmapRepository) GetIterationsForTask(ctx context.Context, taskID string) ([]*IterationEntity, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT i.number, i.name, i.goal, i.status, i.rank, i.deliverable, i.started_at, i.completed_at, i.created_at, i.updated_at
		 FROM iterations i
		 JOIN iteration_tasks it ON i.number = it.iteration_number
		 WHERE it.task_id = ?
		 ORDER BY i.number ASC`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query iterations for task: %w", err)
	}
	defer rows.Close()

	var iterations []*IterationEntity
	for rows.Next() {
		var iteration IterationEntity
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(&iteration.Number, &iteration.Name, &iteration.Goal, &iteration.Status, &iteration.Rank, &iteration.Deliverable, &startedAt, &completedAt, &iteration.CreatedAt, &iteration.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan iteration: %w", err)
		}

		if startedAt.Valid {
			iteration.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			iteration.CompletedAt = &completedAt.Time
		}

		// Load task IDs for each iteration
		taskIDs, err := r.getIterationTaskIDs(ctx, iteration.Number)
		if err != nil {
			return nil, fmt.Errorf("failed to load iteration tasks: %w", err)
		}
		iteration.TaskIDs = taskIDs

		iterations = append(iterations, &iteration)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating iterations: %w", err)
	}

	return iterations, nil
}

// GetBacklogTasks returns all tasks that are not in any iteration and not done.
func (r *SQLiteRoadmapRepository) GetBacklogTasks(ctx context.Context) ([]*TaskEntity, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT t.id, t.track_id, t.title, t.description, t.status, t.rank, t.branch, t.created_at, t.updated_at
		 FROM tasks t
		 LEFT JOIN iteration_tasks it ON t.id = it.task_id
		 WHERE it.task_id IS NULL AND t.status != 'done'
		 ORDER BY t.created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query backlog tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*TaskEntity
	for rows.Next() {
		var task TaskEntity
		var branch sql.NullString

		err := rows.Scan(&task.ID, &task.TrackID, &task.Title, &task.Description, &task.Status, &task.Rank, &branch, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if branch.Valid {
			task.Branch = branch.String
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}

// ListFailedAC returns all acceptance criteria with status "failed".
func (r *SQLiteRoadmapRepository) ListFailedAC(ctx context.Context, filters ACFilters) ([]*AcceptanceCriteriaEntity, error) {
	query := `SELECT ac.id, ac.task_id, ac.description, ac.verification_type, ac.status, ac.notes, ac.testing_instructions, ac.created_at, ac.updated_at
		      FROM acceptance_criteria ac`

	var joins []string
	var conditions []string
	var args []interface{}

	// Base condition: status = failed
	conditions = append(conditions, "ac.status = ?")
	args = append(args, string(ACStatusFailed))

	// Add iteration filter
	if filters.IterationNum != nil {
		joins = append(joins, "JOIN iteration_tasks it ON ac.task_id = it.task_id")
		conditions = append(conditions, "it.iteration_number = ?")
		args = append(args, *filters.IterationNum)
	}

	// Add track filter
	if filters.TrackID != "" {
		joins = append(joins, "JOIN tasks t ON ac.task_id = t.id")
		conditions = append(conditions, "t.track_id = ?")
		args = append(args, filters.TrackID)
	}

	// Add task filter
	if filters.TaskID != "" {
		conditions = append(conditions, "ac.task_id = ?")
		args = append(args, filters.TaskID)
	}

	// Build final query
	for _, join := range joins {
		query += " " + join
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY ac.created_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed ACs: %w", err)
	}
	defer rows.Close()

	var acs []*AcceptanceCriteriaEntity
	for rows.Next() {
		var ac AcceptanceCriteriaEntity
		var testingInstructions sql.NullString
		err := rows.Scan(&ac.ID, &ac.TaskID, &ac.Description, (*string)(&ac.VerificationType), (*string)(&ac.Status), &ac.Notes, &testingInstructions, &ac.CreatedAt, &ac.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan AC: %w", err)
		}
		if testingInstructions.Valid {
			ac.TestingInstructions = testingInstructions.String
		}
		acs = append(acs, &ac)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ACs: %w", err)
	}

	return acs, nil
}

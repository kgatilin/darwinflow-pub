package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/repositories"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// Compile-time check that SQLiteIterationRepository implements repositories.IterationRepository
var _ repositories.IterationRepository = (*SQLiteIterationRepository)(nil)

// SQLiteIterationRepository implements repositories.IterationRepository using SQLite as the backend.
type SQLiteIterationRepository struct {
	DB     *sql.DB
	logger pluginsdk.Logger
}

// NewSQLiteIterationRepository creates a new SQLite-backed repository.
func NewSQLiteIterationRepository(db *sql.DB, logger pluginsdk.Logger) *SQLiteIterationRepository {
	return &SQLiteIterationRepository{
		DB:     db,
		logger: logger,
	}
}

// ============================================================================
// Iteration Operations
// ============================================================================

// SaveIteration persists a new iteration to storage.
func (r *SQLiteIterationRepository) SaveIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	// Check if iteration already exists
	var exists int
	err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM iterations WHERE number = ?", iteration.Number).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check iteration existence: %w", err)
	}
	if exists > 0 {
		return fmt.Errorf("%w: iteration %d already exists", pluginsdk.ErrAlreadyExists, iteration.Number)
	}

	// Start transaction for iteration and tasks
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert iteration
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO iterations (number, name, goal, status, rank, deliverable, started_at, completed_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		iteration.Number, iteration.Name, iteration.Goal, iteration.Status, iteration.Rank, iteration.Deliverable, iteration.StartedAt, iteration.CompletedAt, iteration.CreatedAt, iteration.UpdatedAt,
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
func (r *SQLiteIterationRepository) GetIteration(ctx context.Context, number int) (*entities.IterationEntity, error) {
	var iteration entities.IterationEntity
	var startedAt, completedAt sql.NullTime

	err := r.DB.QueryRowContext(
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
func (r *SQLiteIterationRepository) GetCurrentIteration(ctx context.Context) (*entities.IterationEntity, error) {
	var iteration entities.IterationEntity
	var startedAt, completedAt sql.NullTime

	err := r.DB.QueryRowContext(
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
func (r *SQLiteIterationRepository) ListIterations(ctx context.Context) ([]*entities.IterationEntity, error) {
	rows, err := r.DB.QueryContext(
		ctx,
		"SELECT number, name, goal, status, rank, deliverable, started_at, completed_at, created_at, updated_at FROM iterations ORDER BY rank, number",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query iterations: %w", err)
	}
	defer rows.Close()

	var iterations []*entities.IterationEntity
	for rows.Next() {
		var iteration entities.IterationEntity
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
func (r *SQLiteIterationRepository) UpdateIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	// Start transaction for iteration and tasks update
	tx, err := r.DB.BeginTx(ctx, nil)
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
func (r *SQLiteIterationRepository) DeleteIteration(ctx context.Context, number int) error {
	result, err := r.DB.ExecContext(ctx, "DELETE FROM iterations WHERE number = ?", number)
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
func (r *SQLiteIterationRepository) AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error {
	// Check if iteration exists
	var iterExists int
	err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM iterations WHERE number = ?", iterationNum).Scan(&iterExists)
	if err != nil {
		return fmt.Errorf("failed to check iteration existence: %w", err)
	}
	if iterExists == 0 {
		return fmt.Errorf("%w: iteration %d not found", pluginsdk.ErrNotFound, iterationNum)
	}

	// Check if task exists
	var taskExists int
	err = r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&taskExists)
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if taskExists == 0 {
		return fmt.Errorf("%w: task %s not found", pluginsdk.ErrNotFound, taskID)
	}

	// Check if task already in iteration
	var alreadyExists int
	err = r.DB.QueryRowContext(
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
	_, err = r.DB.ExecContext(
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
func (r *SQLiteIterationRepository) RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error {
	result, err := r.DB.ExecContext(
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
func (r *SQLiteIterationRepository) GetIterationTasks(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
	// Check if iteration exists
	var exists int
	err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM iterations WHERE number = ?", iterationNum).Scan(&exists)
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

	var tasks []*entities.TaskEntity
	for _, taskID := range taskIDs {
		task, err := r.getTask(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetIterationTasksWithWarnings retrieves all tasks for an iteration,
// gracefully handling missing tasks by returning them separately.
// Returns: found tasks, missing task IDs, error
func (r *SQLiteIterationRepository) GetIterationTasksWithWarnings(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, []string, error) {
	// Check if iteration exists
	var exists int
	err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM iterations WHERE number = ?", iterationNum).Scan(&exists)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check iteration existence: %w", err)
	}
	if exists == 0 {
		return nil, nil, fmt.Errorf("%w: iteration %d not found", pluginsdk.ErrNotFound, iterationNum)
	}

	// Get all task IDs from iteration_tasks table
	taskIDs, err := r.getIterationTaskIDs(ctx, iterationNum)
	if err != nil {
		return nil, nil, err
	}

	var foundTasks []*entities.TaskEntity
	var missingTaskIDs []string

	// Try to fetch each task, collecting missing ones separately
	for _, taskID := range taskIDs {
		task, err := r.getTask(ctx, taskID)
		if err != nil {
			if errors.Is(err, pluginsdk.ErrNotFound) {
				// Task was deleted or missing - add to missing list
				missingTaskIDs = append(missingTaskIDs, taskID)
				continue
			}
			// Other errors should still fail the operation
			return nil, nil, fmt.Errorf("failed to get task %s: %w", taskID, err)
		}
		foundTasks = append(foundTasks, task)
	}

	return foundTasks, missingTaskIDs, nil
}

// StartIteration marks an iteration as current and sets started_at timestamp.
func (r *SQLiteIterationRepository) StartIteration(ctx context.Context, iterationNum int) error {
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
func (r *SQLiteIterationRepository) CompleteIteration(ctx context.Context, iterationNum int) error {
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

// GetIterationByNumber is an alias for GetIteration for consistency with other repositories.
func (r *SQLiteIterationRepository) GetIterationByNumber(ctx context.Context, number int) (*entities.IterationEntity, error) {
	return r.GetIteration(ctx, number)
}

// GetNextPlannedIteration returns the first planned iteration ordered by rank.
func (r *SQLiteIterationRepository) GetNextPlannedIteration(ctx context.Context) (*entities.IterationEntity, error) {
	var iteration entities.IterationEntity
	var startedAt, completedAt sql.NullTime

	err := r.DB.QueryRowContext(
		ctx,
		"SELECT number, name, goal, status, rank, deliverable, started_at, completed_at, created_at, updated_at FROM iterations WHERE status = ? ORDER BY rank, number LIMIT 1",
		"planned",
	).Scan(&iteration.Number, &iteration.Name, &iteration.Goal, &iteration.Status, &iteration.Rank, &iteration.Deliverable, &startedAt, &completedAt, &iteration.CreatedAt, &iteration.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: no planned iteration found", pluginsdk.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to query next planned iteration: %w", err)
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

// ============================================================================
// Helper Methods
// ============================================================================

// getIterationTaskIDs retrieves all task IDs for an iteration.
func (r *SQLiteIterationRepository) getIterationTaskIDs(ctx context.Context, iterationNum int) ([]string, error) {
	rows, err := r.DB.QueryContext(
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

// getTask retrieves a task by its ID.
func (r *SQLiteIterationRepository) getTask(ctx context.Context, id string) (*entities.TaskEntity, error) {
	var task entities.TaskEntity
	var branch sql.NullString

	err := r.DB.QueryRowContext(
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

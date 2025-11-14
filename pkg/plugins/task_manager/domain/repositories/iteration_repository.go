package repositories

import (
	"context"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
)

// IterationRepository defines the contract for persistent storage of iteration entities.
type IterationRepository interface {
	// SaveIteration persists a new iteration to storage.
	// Returns ErrAlreadyExists if an iteration with the same number already exists.
	SaveIteration(ctx context.Context, iteration *entities.IterationEntity) error

	// GetIteration retrieves an iteration by its number.
	// Returns ErrNotFound if the iteration doesn't exist.
	GetIteration(ctx context.Context, number int) (*entities.IterationEntity, error)

	// GetCurrentIteration returns the iteration with status "current".
	// Returns ErrNotFound if no current iteration exists.
	GetCurrentIteration(ctx context.Context) (*entities.IterationEntity, error)

	// ListIterations returns all iterations, ordered by number.
	// Returns empty slice if no iterations exist.
	ListIterations(ctx context.Context) ([]*entities.IterationEntity, error)

	// UpdateIteration updates an existing iteration.
	// Returns ErrNotFound if the iteration doesn't exist.
	UpdateIteration(ctx context.Context, iteration *entities.IterationEntity) error

	// DeleteIteration removes an iteration from storage.
	// Returns ErrNotFound if the iteration doesn't exist.
	DeleteIteration(ctx context.Context, number int) error

	// AddTaskToIteration adds a task to an iteration.
	// Returns ErrNotFound if the iteration or task doesn't exist.
	// Returns ErrAlreadyExists if the task is already in the iteration.
	AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error

	// RemoveTaskFromIteration removes a task from an iteration.
	// Returns ErrNotFound if the iteration doesn't exist or the task is not in it.
	RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error

	// GetIterationTasks returns all tasks in an iteration.
	// Returns empty slice if the iteration has no tasks.
	GetIterationTasks(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error)

	// GetIterationTasksWithWarnings retrieves all tasks for an iteration,
	// gracefully handling missing tasks by returning them separately.
	// Returns: found tasks, missing task IDs, error
	GetIterationTasksWithWarnings(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, []string, error)

	// StartIteration marks an iteration as current and sets started_at timestamp.
	// Returns ErrNotFound if the iteration doesn't exist.
	// Returns ErrInvalidArgument if the iteration status is not "planned".
	StartIteration(ctx context.Context, iterationNum int) error

	// CompleteIteration marks an iteration as complete and sets completed_at timestamp.
	// Returns ErrNotFound if the iteration doesn't exist.
	// Returns ErrInvalidArgument if the iteration status is not "current".
	CompleteIteration(ctx context.Context, iterationNum int) error

	// GetIterationByNumber is an alias for GetIteration for consistency with other repositories.
	GetIterationByNumber(ctx context.Context, number int) (*entities.IterationEntity, error)

	// GetNextPlannedIteration returns the first planned iteration ordered by rank.
	// Returns ErrNotFound if no planned iterations exist.
	GetNextPlannedIteration(ctx context.Context) (*entities.IterationEntity, error)
}

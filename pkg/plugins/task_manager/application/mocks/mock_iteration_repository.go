package mocks

import (
	"context"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
)

// MockIterationRepository is a mock implementation of repositories.IterationRepository for testing.
type MockIterationRepository struct {
	// In-memory storage for testing
	iterations map[int]*entities.IterationEntity

	// SaveIterationFunc is called by SaveIteration. If nil, uses default implementation.
	SaveIterationFunc func(ctx context.Context, iteration *entities.IterationEntity) error

	// GetIterationFunc is called by GetIteration. If nil, uses default implementation.
	GetIterationFunc func(ctx context.Context, number int) (*entities.IterationEntity, error)

	// GetCurrentIterationFunc is called by GetCurrentIteration. If nil, returns nil, nil.
	GetCurrentIterationFunc func(ctx context.Context) (*entities.IterationEntity, error)

	// ListIterationsFunc is called by ListIterations. If nil, returns empty slice, nil.
	ListIterationsFunc func(ctx context.Context) ([]*entities.IterationEntity, error)

	// UpdateIterationFunc is called by UpdateIteration. If nil, returns nil.
	UpdateIterationFunc func(ctx context.Context, iteration *entities.IterationEntity) error

	// DeleteIterationFunc is called by DeleteIteration. If nil, returns nil.
	DeleteIterationFunc func(ctx context.Context, number int) error

	// AddTaskToIterationFunc is called by AddTaskToIteration. If nil, returns nil.
	AddTaskToIterationFunc func(ctx context.Context, iterationNum int, taskID string) error

	// RemoveTaskFromIterationFunc is called by RemoveTaskFromIteration. If nil, returns nil.
	RemoveTaskFromIterationFunc func(ctx context.Context, iterationNum int, taskID string) error

	// GetIterationTasksFunc is called by GetIterationTasks. If nil, returns empty slice, nil.
	GetIterationTasksFunc func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error)

	// GetIterationTasksWithWarningsFunc is called by GetIterationTasksWithWarnings. If nil, returns empty slice, empty slice, nil.
	GetIterationTasksWithWarningsFunc func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, []string, error)

	// StartIterationFunc is called by StartIteration. If nil, returns nil.
	StartIterationFunc func(ctx context.Context, iterationNum int) error

	// CompleteIterationFunc is called by CompleteIteration. If nil, returns nil.
	CompleteIterationFunc func(ctx context.Context, iterationNum int) error

	// GetIterationByNumberFunc is called by GetIterationByNumber. If nil, returns nil, nil.
	GetIterationByNumberFunc func(ctx context.Context, number int) (*entities.IterationEntity, error)

	// GetNextPlannedIterationFunc is called by GetNextPlannedIteration. If nil, returns nil, nil.
	GetNextPlannedIterationFunc func(ctx context.Context) (*entities.IterationEntity, error)
}

// NewMockIterationRepository creates a new mock iteration repository with in-memory storage
func NewMockIterationRepository() *MockIterationRepository {
	return &MockIterationRepository{
		iterations: make(map[int]*entities.IterationEntity),
	}
}

// SaveIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) SaveIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	if m.SaveIterationFunc != nil {
		return m.SaveIterationFunc(ctx, iteration)
	}
	// Default implementation: store in memory
	m.iterations[iteration.Number] = iteration
	return nil
}

// GetIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) GetIteration(ctx context.Context, number int) (*entities.IterationEntity, error) {
	if m.GetIterationFunc != nil {
		return m.GetIterationFunc(ctx, number)
	}
	// Default implementation: get from memory
	iteration, exists := m.iterations[number]
	if !exists {
		return nil, nil
	}
	return iteration, nil
}

// GetCurrentIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) GetCurrentIteration(ctx context.Context) (*entities.IterationEntity, error) {
	if m.GetCurrentIterationFunc != nil {
		return m.GetCurrentIterationFunc(ctx)
	}
	// Default implementation: find iteration with status "current"
	for _, iteration := range m.iterations {
		if iteration.Status == "current" {
			return iteration, nil
		}
	}
	return nil, nil
}

// ListIterations implements repositories.IterationRepository.
func (m *MockIterationRepository) ListIterations(ctx context.Context) ([]*entities.IterationEntity, error) {
	if m.ListIterationsFunc != nil {
		return m.ListIterationsFunc(ctx)
	}
	// Default implementation: return all iterations
	var result []*entities.IterationEntity
	for _, iteration := range m.iterations {
		result = append(result, iteration)
	}
	return result, nil
}

// UpdateIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) UpdateIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	if m.UpdateIterationFunc != nil {
		return m.UpdateIterationFunc(ctx, iteration)
	}
	return nil
}

// DeleteIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) DeleteIteration(ctx context.Context, number int) error {
	if m.DeleteIterationFunc != nil {
		return m.DeleteIterationFunc(ctx, number)
	}
	return nil
}

// AddTaskToIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error {
	if m.AddTaskToIterationFunc != nil {
		return m.AddTaskToIterationFunc(ctx, iterationNum, taskID)
	}
	return nil
}

// RemoveTaskFromIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error {
	if m.RemoveTaskFromIterationFunc != nil {
		return m.RemoveTaskFromIterationFunc(ctx, iterationNum, taskID)
	}
	return nil
}

// GetIterationTasks implements repositories.IterationRepository.
func (m *MockIterationRepository) GetIterationTasks(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
	if m.GetIterationTasksFunc != nil {
		return m.GetIterationTasksFunc(ctx, iterationNum)
	}
	return []*entities.TaskEntity{}, nil
}

// GetIterationTasksWithWarnings implements repositories.IterationRepository.
func (m *MockIterationRepository) GetIterationTasksWithWarnings(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, []string, error) {
	if m.GetIterationTasksWithWarningsFunc != nil {
		return m.GetIterationTasksWithWarningsFunc(ctx, iterationNum)
	}
	return []*entities.TaskEntity{}, []string{}, nil
}

// StartIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) StartIteration(ctx context.Context, iterationNum int) error {
	if m.StartIterationFunc != nil {
		return m.StartIterationFunc(ctx, iterationNum)
	}
	return nil
}

// CompleteIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) CompleteIteration(ctx context.Context, iterationNum int) error {
	if m.CompleteIterationFunc != nil {
		return m.CompleteIterationFunc(ctx, iterationNum)
	}
	return nil
}

// GetIterationByNumber implements repositories.IterationRepository.
func (m *MockIterationRepository) GetIterationByNumber(ctx context.Context, number int) (*entities.IterationEntity, error) {
	if m.GetIterationByNumberFunc != nil {
		return m.GetIterationByNumberFunc(ctx, number)
	}
	return nil, nil
}

// GetNextPlannedIteration implements repositories.IterationRepository.
func (m *MockIterationRepository) GetNextPlannedIteration(ctx context.Context) (*entities.IterationEntity, error) {
	if m.GetNextPlannedIterationFunc != nil {
		return m.GetNextPlannedIterationFunc(ctx)
	}
	// Default implementation: find first planned iteration
	for _, iteration := range m.iterations {
		if iteration.Status == "planned" {
			return iteration, nil
		}
	}
	return nil, nil
}

// Reset clears all configured behavior.
func (m *MockIterationRepository) Reset() {
	m.SaveIterationFunc = nil
	m.GetIterationFunc = nil
	m.GetCurrentIterationFunc = nil
	m.ListIterationsFunc = nil
	m.UpdateIterationFunc = nil
	m.DeleteIterationFunc = nil
	m.AddTaskToIterationFunc = nil
	m.RemoveTaskFromIterationFunc = nil
	m.GetIterationTasksFunc = nil
	m.GetIterationTasksWithWarningsFunc = nil
	m.StartIterationFunc = nil
	m.CompleteIterationFunc = nil
	m.GetIterationByNumberFunc = nil
	m.GetNextPlannedIterationFunc = nil
}

// WithError configures the mock to return the specified error for all methods.
func (m *MockIterationRepository) WithError(err error) *MockIterationRepository {
	m.SaveIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error { return err }
	m.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) { return nil, err }
	m.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) { return nil, err }
	m.ListIterationsFunc = func(ctx context.Context) ([]*entities.IterationEntity, error) { return nil, err }
	m.UpdateIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error { return err }
	m.DeleteIterationFunc = func(ctx context.Context, number int) error { return err }
	m.AddTaskToIterationFunc = func(ctx context.Context, iterationNum int, taskID string) error { return err }
	m.RemoveTaskFromIterationFunc = func(ctx context.Context, iterationNum int, taskID string) error { return err }
	m.GetIterationTasksFunc = func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
		return nil, err
	}
	m.GetIterationTasksWithWarningsFunc = func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, []string, error) {
		return nil, nil, err
	}
	m.StartIterationFunc = func(ctx context.Context, iterationNum int) error { return err }
	m.CompleteIterationFunc = func(ctx context.Context, iterationNum int) error { return err }
	m.GetIterationByNumberFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, err
	}
	m.GetNextPlannedIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return nil, err
	}
	return m
}

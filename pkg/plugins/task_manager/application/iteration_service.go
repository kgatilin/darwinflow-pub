package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/dto"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/repositories"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/services"
)

// IterationApplicationService orchestrates iteration operations, delegating to domain services and repositories.
// It handles all iteration lifecycle operations including creation, updates, state transitions, and task management.
type IterationApplicationService struct {
	iterationRepo     repositories.IterationRepository
	taskRepo          repositories.TaskRepository
	aggregateRepo     repositories.AggregateRepository
	iterationService  *services.IterationService
	validationService *services.ValidationService
}

// NewIterationApplicationService creates a new iteration application service.
func NewIterationApplicationService(
	iterationRepo repositories.IterationRepository,
	taskRepo repositories.TaskRepository,
	aggregateRepo repositories.AggregateRepository,
	iterationService *services.IterationService,
	validationService *services.ValidationService,
) *IterationApplicationService {
	return &IterationApplicationService{
		iterationRepo:     iterationRepo,
		taskRepo:          taskRepo,
		aggregateRepo:     aggregateRepo,
		iterationService:  iterationService,
		validationService: validationService,
	}
}

// ============================================================================
// Write Operations
// ============================================================================

// CreateIteration creates a new iteration with validation.
// Default status is "planned" if not specified.
// Iteration number is auto-generated based on existing iterations.
func (s *IterationApplicationService) CreateIteration(ctx context.Context, input dto.CreateIterationDTO) (*entities.IterationEntity, error) {
	// Generate iteration number (max + 1)
	iterations, err := s.iterationRepo.ListIterations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate iteration number: %w", err)
	}

	// Find max iteration number
	maxNumber := 0
	for _, iter := range iterations {
		if iter.Number > maxNumber {
			maxNumber = iter.Number
		}
	}
	nextNumber := maxNumber + 1

	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(nextNumber); err != nil {
		return nil, err
	}

	// Validate name is non-empty
	if err := s.validationService.ValidateNonEmpty("name", input.Name); err != nil {
		return nil, err
	}

	// Check if iteration already exists
	_, err = s.iterationRepo.GetIteration(ctx, nextNumber)
	if err == nil {
		return nil, fmt.Errorf("%w: iteration %d already exists", pluginsdk.ErrAlreadyExists, nextNumber)
	}
	// If error is not ErrNotFound, it's an unexpected error
	if !errors.Is(err, pluginsdk.ErrNotFound) {
		return nil, fmt.Errorf("failed to check iteration existence: %w", err)
	}

	// Default status to "planned"
	status := input.Status
	if status == "" {
		status = string(entities.IterationStatusPlanned)
	}

	// Validate status
	if !entities.IsValidIterationStatus(status) {
		return nil, fmt.Errorf("%w: invalid iteration status: %s", pluginsdk.ErrInvalidArgument, status)
	}

	// Create iteration entity
	now := time.Now().UTC()
	iteration, err := entities.NewIterationEntity(
		nextNumber,
		input.Name,
		input.Goal,
		input.Deliverable,
		[]string{},
		status,
		500, // Default rank
		time.Time{},
		time.Time{},
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create iteration entity: %w", err)
	}

	// Persist iteration
	if err := s.iterationRepo.SaveIteration(ctx, iteration); err != nil {
		return nil, fmt.Errorf("failed to save iteration: %w", err)
	}

	return iteration, nil
}

// UpdateIteration updates an existing iteration.
// Only non-nil fields in the DTO are updated.
func (s *IterationApplicationService) UpdateIteration(ctx context.Context, input dto.UpdateIterationDTO) (*entities.IterationEntity, error) {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(input.Number); err != nil {
		return nil, err
	}

	// Retrieve existing iteration
	iteration, err := s.iterationRepo.GetIteration(ctx, input.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to get iteration: %w", err)
	}

	// Apply updates
	if input.Name != nil {
		if err := s.validationService.ValidateNonEmpty("name", *input.Name); err != nil {
			return nil, err
		}
		iteration.Name = *input.Name
	}

	if input.Goal != nil {
		iteration.Goal = *input.Goal
	}

	if input.Deliverable != nil {
		iteration.Deliverable = *input.Deliverable
	}

	iteration.UpdatedAt = time.Now().UTC()

	// Persist changes
	if err := s.iterationRepo.UpdateIteration(ctx, iteration); err != nil {
		return nil, fmt.Errorf("failed to update iteration: %w", err)
	}

	return iteration, nil
}

// DeleteIteration removes an iteration from storage.
func (s *IterationApplicationService) DeleteIteration(ctx context.Context, iterationNum int) error {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(iterationNum); err != nil {
		return err
	}

	// Verify iteration exists
	_, err := s.iterationRepo.GetIteration(ctx, iterationNum)
	if err != nil {
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Delete iteration
	if err := s.iterationRepo.DeleteIteration(ctx, iterationNum); err != nil {
		return fmt.Errorf("failed to delete iteration: %w", err)
	}

	return nil
}

// ============================================================================
// Lifecycle Operations
// ============================================================================

// StartIteration transitions an iteration from "planned" to "current".
// Validates that no other iteration is current before starting.
func (s *IterationApplicationService) StartIteration(ctx context.Context, iterationNum int) error {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(iterationNum); err != nil {
		return err
	}

	// Retrieve iteration
	iteration, err := s.iterationRepo.GetIteration(ctx, iterationNum)
	if err != nil {
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Validate start transition using domain service
	if err := s.iterationService.CanStartIteration(ctx, iteration, s.iterationRepo.GetCurrentIteration); err != nil {
		return err
	}

	// Transition to current status
	if err := iteration.TransitionTo(string(entities.IterationStatusCurrent)); err != nil {
		return fmt.Errorf("failed to transition iteration: %w", err)
	}

	// Persist changes
	if err := s.iterationRepo.UpdateIteration(ctx, iteration); err != nil {
		return fmt.Errorf("failed to update iteration: %w", err)
	}

	return nil
}

// CompleteIteration transitions an iteration from "current" to "complete".
func (s *IterationApplicationService) CompleteIteration(ctx context.Context, iterationNum int) error {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(iterationNum); err != nil {
		return err
	}

	// Retrieve iteration
	iteration, err := s.iterationRepo.GetIteration(ctx, iterationNum)
	if err != nil {
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Validate complete transition using domain service
	if err := s.iterationService.CanCompleteIteration(iteration); err != nil {
		return err
	}

	// Transition to complete status
	if err := iteration.TransitionTo(string(entities.IterationStatusComplete)); err != nil {
		return fmt.Errorf("failed to transition iteration: %w", err)
	}

	// Persist changes
	if err := s.iterationRepo.UpdateIteration(ctx, iteration); err != nil {
		return fmt.Errorf("failed to update iteration: %w", err)
	}

	return nil
}

// ============================================================================
// Task Management
// ============================================================================

// AddTask adds a task to an iteration.
func (s *IterationApplicationService) AddTask(ctx context.Context, iterationNum int, taskID string) error {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(iterationNum); err != nil {
		return err
	}

	// Verify iteration exists
	_, err := s.iterationRepo.GetIteration(ctx, iterationNum)
	if err != nil {
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Verify task exists
	_, err = s.taskRepo.GetTask(ctx, taskID)
	if err != nil {
		if err == pluginsdk.ErrNotFound {
			return fmt.Errorf("%w: task %s not found", pluginsdk.ErrNotFound, taskID)
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Add task to iteration
	if err := s.iterationRepo.AddTaskToIteration(ctx, iterationNum, taskID); err != nil {
		return fmt.Errorf("failed to add task to iteration: %w", err)
	}

	return nil
}

// RemoveTask removes a task from an iteration.
func (s *IterationApplicationService) RemoveTask(ctx context.Context, iterationNum int, taskID string) error {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(iterationNum); err != nil {
		return err
	}

	// Verify iteration exists
	_, err := s.iterationRepo.GetIteration(ctx, iterationNum)
	if err != nil {
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Remove task from iteration
	if err := s.iterationRepo.RemoveTaskFromIteration(ctx, iterationNum, taskID); err != nil {
		return fmt.Errorf("failed to remove task from iteration: %w", err)
	}

	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// GetIteration retrieves an iteration by its number.
func (s *IterationApplicationService) GetIteration(ctx context.Context, iterationNum int) (*entities.IterationEntity, error) {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(iterationNum); err != nil {
		return nil, err
	}

	iteration, err := s.iterationRepo.GetIteration(ctx, iterationNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get iteration: %w", err)
	}

	return iteration, nil
}

// GetCurrentIteration returns the iteration with status "current".
// If no current iteration exists, returns the first planned iteration (fallback).
// Returns result with IsFallback=true and FallbackMsg when using fallback.
func (s *IterationApplicationService) GetCurrentIteration(ctx context.Context) (*dto.CurrentIterationResult, error) {
	// Try to get current iteration first
	iteration, err := s.iterationRepo.GetCurrentIteration(ctx)
	if err == nil {
		// Found current iteration - return without fallback
		return &dto.CurrentIterationResult{
			Iteration:   iteration,
			IsFallback:  false,
			FallbackMsg: "",
		}, nil
	}

	// Check if error is "not found" - if not, it's a real error
	if !errors.Is(err, pluginsdk.ErrNotFound) {
		return nil, fmt.Errorf("failed to get current iteration: %w", err)
	}

	// No current iteration - try fallback to next planned iteration
	plannedIter, err := s.iterationRepo.GetNextPlannedIteration(ctx)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			// No current and no planned iterations - return nil with fallback message
			return &dto.CurrentIterationResult{
				Iteration:   nil,
				IsFallback:  true,
				FallbackMsg: "No current or planned iterations found",
			}, nil
		}
		return nil, fmt.Errorf("failed to get next planned iteration: %w", err)
	}

	// Return planned iteration as fallback
	return &dto.CurrentIterationResult{
		Iteration:   plannedIter,
		IsFallback:  true,
		FallbackMsg: fmt.Sprintf("No current iteration. Showing next planned iteration: %s", plannedIter.Name),
	}, nil
}

// ListIterations returns all iterations ordered by number.
func (s *IterationApplicationService) ListIterations(ctx context.Context) ([]*entities.IterationEntity, error) {
	iterations, err := s.iterationRepo.ListIterations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list iterations: %w", err)
	}

	return iterations, nil
}

// GetIterationTasks returns all tasks in an iteration.
func (s *IterationApplicationService) GetIterationTasks(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
	// Validate iteration number
	if err := s.validationService.ValidateIterationNumber(iterationNum); err != nil {
		return nil, err
	}

	tasks, err := s.iterationRepo.GetIterationTasks(ctx, iterationNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get iteration tasks: %w", err)
	}

	return tasks, nil
}

package application

import (
	"context"
	"fmt"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/dto"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/repositories"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/services"
)

// ACApplicationService handles all Acceptance Criteria operations
type ACApplicationService struct {
	acRepo            repositories.AcceptanceCriteriaRepository
	taskRepo          repositories.TaskRepository
	aggregateRepo     repositories.AggregateRepository
	validationService *services.ValidationService
}

// NewACApplicationService creates a new AC service
func NewACApplicationService(
	acRepo repositories.AcceptanceCriteriaRepository,
	taskRepo repositories.TaskRepository,
	aggregateRepo repositories.AggregateRepository,
	validationService *services.ValidationService,
) *ACApplicationService {
	return &ACApplicationService{
		acRepo:            acRepo,
		taskRepo:          taskRepo,
		aggregateRepo:     aggregateRepo,
		validationService: validationService,
	}
}

// CreateAC creates a new acceptance criterion
func (s *ACApplicationService) CreateAC(ctx context.Context, input dto.CreateACDTO) (*entities.AcceptanceCriteriaEntity, error) {
	// Generate AC ID
	projectCode := s.aggregateRepo.GetProjectCode(ctx)
	nextNum, err := s.aggregateRepo.GetNextSequenceNumber(ctx, "ac")
	if err != nil {
		return nil, fmt.Errorf("failed to generate AC ID: %w", err)
	}
	id := fmt.Sprintf("%s-ac-%d", projectCode, nextNum)

	// Validate AC ID
	if err := s.validationService.ValidateNonEmpty("AC ID", id); err != nil {
		return nil, err
	}

	// Validate description
	if err := s.validationService.ValidateNonEmpty("description", input.Description); err != nil {
		return nil, err
	}

	// Verify task exists
	_, err = s.taskRepo.GetTask(ctx, input.TaskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	now := time.Now().UTC()

	// Create AC entity (default status: not-started, default type: manual)
	ac := entities.NewAcceptanceCriteriaEntity(
		id,
		input.TaskID,
		input.Description,
		entities.VerificationTypeManual,
		input.TestingInstructions,
		now,
		now,
	)

	// Persist AC
	if err := s.acRepo.SaveAC(ctx, ac); err != nil {
		return nil, fmt.Errorf("failed to save AC: %w", err)
	}

	return ac, nil
}

// UpdateAC updates an existing acceptance criterion
func (s *ACApplicationService) UpdateAC(ctx context.Context, input dto.UpdateACDTO) (*entities.AcceptanceCriteriaEntity, error) {
	// Fetch existing AC
	ac, err := s.acRepo.GetAC(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AC: %w", err)
	}

	// Apply updates
	if input.Description != nil {
		if err := s.validationService.ValidateNonEmpty("description", *input.Description); err != nil {
			return nil, err
		}
		ac.Description = *input.Description
	}

	if input.TestingInstructions != nil {
		ac.TestingInstructions = *input.TestingInstructions
	}

	// Update timestamp
	ac.UpdatedAt = time.Now().UTC()

	// Persist updates
	if err := s.acRepo.UpdateAC(ctx, ac); err != nil {
		return nil, fmt.Errorf("failed to update AC: %w", err)
	}

	return ac, nil
}

// VerifyAC marks an acceptance criterion as verified
func (s *ACApplicationService) VerifyAC(ctx context.Context, input dto.VerifyACDTO) error {
	// Fetch existing AC
	ac, err := s.acRepo.GetAC(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("AC not found: %w", err)
	}

	// Update status to verified
	ac.Status = entities.ACStatusVerified
	ac.Notes = fmt.Sprintf("Verified by: %s at %s", input.VerifiedBy, input.VerifiedAt)
	ac.UpdatedAt = time.Now().UTC()

	// Persist updates
	if err := s.acRepo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to verify AC: %w", err)
	}

	return nil
}

// FailAC marks an acceptance criterion as failed
func (s *ACApplicationService) FailAC(ctx context.Context, input dto.FailACDTO) error {
	// Validate feedback
	if err := s.validationService.ValidateNonEmpty("feedback", input.Feedback); err != nil {
		return err
	}

	// Fetch existing AC
	ac, err := s.acRepo.GetAC(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("AC not found: %w", err)
	}

	// Update status to failed
	ac.Status = entities.ACStatusFailed
	ac.Notes = input.Feedback
	ac.UpdatedAt = time.Now().UTC()

	// Persist updates
	if err := s.acRepo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to mark AC as failed: %w", err)
	}

	return nil
}

// SkipAC marks an acceptance criterion as skipped with a reason
func (s *ACApplicationService) SkipAC(ctx context.Context, input dto.SkipACDTO) error {
	// Validate reason
	if err := s.validationService.ValidateNonEmpty("reason", input.Reason); err != nil {
		return err
	}

	// Fetch existing AC
	ac, err := s.acRepo.GetAC(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("AC not found: %w", err)
	}

	// Update status to skipped
	ac.Status = entities.ACStatusSkipped
	ac.Notes = input.Reason
	ac.UpdatedAt = time.Now().UTC()

	// Persist updates
	if err := s.acRepo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to skip AC: %w", err)
	}

	return nil
}

// DeleteAC removes an acceptance criterion
func (s *ACApplicationService) DeleteAC(ctx context.Context, acID string) error {
	if err := s.acRepo.DeleteAC(ctx, acID); err != nil {
		return fmt.Errorf("failed to delete AC: %w", err)
	}
	return nil
}

// GetAC retrieves an acceptance criterion by ID
func (s *ACApplicationService) GetAC(ctx context.Context, acID string) (*entities.AcceptanceCriteriaEntity, error) {
	ac, err := s.acRepo.GetAC(ctx, acID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AC: %w", err)
	}
	return ac, nil
}

// ListAC returns all acceptance criteria for a task
func (s *ACApplicationService) ListAC(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
	acs, err := s.acRepo.ListAC(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list ACs: %w", err)
	}
	return acs, nil
}

// ListACByIteration returns all acceptance criteria for tasks in an iteration
func (s *ACApplicationService) ListACByIteration(ctx context.Context, iterationNum int) ([]*entities.AcceptanceCriteriaEntity, error) {
	acs, err := s.acRepo.ListACByIteration(ctx, iterationNum)
	if err != nil {
		return nil, fmt.Errorf("failed to list ACs by iteration: %w", err)
	}
	return acs, nil
}

// ListFailedAC returns all acceptance criteria with status "failed"
func (s *ACApplicationService) ListFailedAC(ctx context.Context, filters entities.ACFilters) ([]*entities.AcceptanceCriteriaEntity, error) {
	acs, err := s.acRepo.ListFailedAC(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list failed ACs: %w", err)
	}
	return acs, nil
}

package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/dto"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/mocks"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/services"
)

// setupACTestService creates a test service with mock repositories
func setupACTestService(t *testing.T) (*application.ACApplicationService, context.Context, *mocks.MockAcceptanceCriteriaRepository, *mocks.MockTaskRepository, *mocks.MockAggregateRepository) {
	mockACRepo := &mocks.MockAcceptanceCriteriaRepository{}
	mockTaskRepo := &mocks.MockTaskRepository{}
	mockAggregateRepo := &mocks.MockAggregateRepository{}
	validationService := services.NewValidationService()

	service := application.NewACApplicationService(mockACRepo, mockTaskRepo, mockAggregateRepo, validationService)
	ctx := context.Background()

	return service, ctx, mockACRepo, mockTaskRepo, mockAggregateRepo
}

// createTestACEntity creates a test AC entity for mock configuration
func createTestACEntity(t *testing.T, id, taskID string) *entities.AcceptanceCriteriaEntity {
	now := time.Now().UTC()
	ac := entities.NewAcceptanceCriteriaEntity(id, taskID, "Test AC", entities.VerificationTypeManual, "Test instructions", now, now)
	return ac
}

// createTestTaskEntityForAC creates a test task entity for mock configuration
func createTestTaskEntityForAC(t *testing.T, taskID string) *entities.TaskEntity {
	now := time.Now().UTC()
	task, err := entities.NewTaskEntity(taskID, "TM-track-1", "Test Task", "", "todo", 500, "", now, now)
	if err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}
	return task
}

// TestACService_CreateAC_Success tests successful AC creation
func TestACService_CreateAC_Success(t *testing.T) {
	service, ctx, mockACRepo, mockTaskRepo, _ := setupACTestService(t)

	task := createTestTaskEntityForAC(t, "TM-task-1")

	// Configure mocks
	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == "TM-task-1" {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.SaveACFunc = func(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
		return nil // Success
	}

	input := dto.CreateACDTO{
		TaskID:              "TM-task-1",
		Description:         "Test acceptance criterion",
		TestingInstructions: "Step 1: Do this\nStep 2: Do that",
	}

	ac, err := service.CreateAC(ctx, input)
	if err != nil {
		t.Fatalf("CreateAC() failed: %v", err)
	}

	if ac.ID == "" {
		t.Error("ac.ID should not be empty (auto-generated)")
	}
	if ac.Description != input.Description {
		t.Errorf("ac.Description = %q, want %q", ac.Description, input.Description)
	}
	if ac.Status != entities.ACStatusNotStarted {
		t.Errorf("ac.Status = %q, want %q", ac.Status, entities.ACStatusNotStarted)
	}
}

// TestACService_CreateAC_InvalidID tests AC creation with invalid ID
// NOTE: This test is now obsolete because CreateACDTO no longer has an ID field.
// The service auto-generates IDs internally, so there's no "invalid ID" scenario for create operations.
// Keeping this test as a stub for documentation purposes.
func TestACService_CreateAC_InvalidID(t *testing.T) {
	t.Skip("Test obsolete: CreateACDTO no longer accepts ID field (service auto-generates)")
}

// TestACService_CreateAC_EmptyDescription tests AC creation with empty description
func TestACService_CreateAC_EmptyDescription(t *testing.T) {
	service, ctx, _, mockTaskRepo, _ := setupACTestService(t)

	task := createTestTaskEntityForAC(t, "TM-task-1")

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == "TM-task-1" {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	input := dto.CreateACDTO{
		TaskID:      "TM-task-1",
		Description: "", // Empty description
	}

	_, err := service.CreateAC(ctx, input)
	if err == nil {
		t.Fatal("CreateAC() should fail with empty description")
	}
}

// TestACService_CreateAC_TaskNotFound tests AC creation with non-existent task
func TestACService_CreateAC_TaskNotFound(t *testing.T) {
	service, ctx, _, mockTaskRepo, _ := setupACTestService(t)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	input := dto.CreateACDTO{
		TaskID:      "nonexistent",
		Description: "Test acceptance criterion",
	}

	_, err := service.CreateAC(ctx, input)
	if err == nil {
		t.Fatal("CreateAC() should fail with non-existent task")
	}
}

// TestACService_CreateAC_DefaultStatus tests AC creation with default status
func TestACService_CreateAC_DefaultStatus(t *testing.T) {
	service, ctx, mockACRepo, mockTaskRepo, _ := setupACTestService(t)

	task := createTestTaskEntityForAC(t, "TM-task-1")

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == "TM-task-1" {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.SaveACFunc = func(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
		return nil
	}

	input := dto.CreateACDTO{
		TaskID:      "TM-task-1",
		Description: "Test acceptance criterion",
	}

	ac, err := service.CreateAC(ctx, input)
	if err != nil {
		t.Fatalf("CreateAC() failed: %v", err)
	}

	if ac.Status != entities.ACStatusNotStarted {
		t.Errorf("ac.Status = %q, want %q", ac.Status, entities.ACStatusNotStarted)
	}
}

// TestACService_UpdateAC_Success tests successful AC update
func TestACService_UpdateAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	original := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	// Configure mocks
	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return original, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.UpdateACFunc = func(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
		return nil
	}

	// Update AC
	newDescription := "Updated description"
	newInstructions := "Updated instructions"
	updateInput := dto.UpdateACDTO{
		ID:                  original.ID, // MUST set ID for update operations
		Description:         &newDescription,
		TestingInstructions: &newInstructions,
	}

	ac, err := service.UpdateAC(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateAC() failed: %v", err)
	}

	if ac.Description != newDescription {
		t.Errorf("ac.Description = %q, want %q", ac.Description, newDescription)
	}
	if ac.TestingInstructions != newInstructions {
		t.Errorf("ac.TestingInstructions = %q, want %q", ac.TestingInstructions, newInstructions)
	}
}

// TestACService_UpdateAC_NotFound tests updating non-existent AC
func TestACService_UpdateAC_NotFound(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	newDescription := "Updated description"
	updateInput := dto.UpdateACDTO{
		ID:          "nonexistent",
		Description: &newDescription,
	}

	_, err := service.UpdateAC(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateAC() should fail for non-existent AC")
	}
}

// TestACService_UpdateAC_PartialUpdate tests partial AC update
func TestACService_UpdateAC_PartialUpdate(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	original := createTestACEntity(t, "TM-ac-1", "TM-task-1")
	originalInstructions := original.TestingInstructions

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return original, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.UpdateACFunc = func(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
		return nil
	}

	// Update only description
	newDescription := "Updated description"
	updateInput := dto.UpdateACDTO{
		ID:          original.ID, // MUST set ID for update operations
		Description: &newDescription,
	}

	ac, err := service.UpdateAC(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateAC() failed: %v", err)
	}

	if ac.Description != newDescription {
		t.Errorf("ac.Description = %q, want %q", ac.Description, newDescription)
	}
	// Other fields should remain unchanged
	if ac.TestingInstructions != originalInstructions {
		t.Errorf("ac.TestingInstructions changed: got %q, want %q", ac.TestingInstructions, originalInstructions)
	}
}

// TestACService_VerifyAC_Success tests successful AC verification
func TestACService_VerifyAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return ac, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.UpdateACFunc = func(ctx context.Context, updatedAC *entities.AcceptanceCriteriaEntity) error {
		ac.Status = updatedAC.Status
		ac.Notes = updatedAC.Notes
		return nil
	}

	// Verify AC
	verifyInput := dto.VerifyACDTO{
		ID:         ac.ID, // MUST set ID for verify operations
		VerifiedBy: "test-user",
		VerifiedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err := service.VerifyAC(ctx, verifyInput)
	if err != nil {
		t.Fatalf("VerifyAC() failed: %v", err)
	}

	// Check status changed
	gotAC, err := service.GetAC(ctx, "TM-ac-1")
	if err != nil {
		t.Fatalf("GetAC() failed: %v", err)
	}

	if gotAC.Status != entities.ACStatusVerified {
		t.Errorf("ac.Status = %q, want %q", gotAC.Status, entities.ACStatusVerified)
	}
}

// TestACService_VerifyAC_NotFound tests verifying non-existent AC
func TestACService_VerifyAC_NotFound(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	verifyInput := dto.VerifyACDTO{
		ID:         "nonexistent",
		VerifiedBy: "test-user",
		VerifiedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err := service.VerifyAC(ctx, verifyInput)
	if err == nil {
		t.Fatal("VerifyAC() should fail for non-existent AC")
	}
}

// TestACService_FailAC_Success tests successful AC failure marking
func TestACService_FailAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return ac, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.UpdateACFunc = func(ctx context.Context, updatedAC *entities.AcceptanceCriteriaEntity) error {
		ac.Status = updatedAC.Status
		ac.Notes = updatedAC.Notes
		return nil
	}

	// Fail AC
	failInput := dto.FailACDTO{
		ID:       ac.ID, // MUST set ID for fail operations
		Feedback: "Failed because XYZ",
	}

	err := service.FailAC(ctx, failInput)
	if err != nil {
		t.Fatalf("FailAC() failed: %v", err)
	}

	// Check status changed
	gotAC, err := service.GetAC(ctx, "TM-ac-1")
	if err != nil {
		t.Fatalf("GetAC() failed: %v", err)
	}

	if gotAC.Status != entities.ACStatusFailed {
		t.Errorf("ac.Status = %q, want %q", gotAC.Status, entities.ACStatusFailed)
	}
	if gotAC.Notes != failInput.Feedback {
		t.Errorf("ac.Notes = %q, want %q", gotAC.Notes, failInput.Feedback)
	}
}

// TestACService_FailAC_NotFound tests failing non-existent AC
func TestACService_FailAC_NotFound(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	failInput := dto.FailACDTO{
		ID:       "nonexistent",
		Feedback: "Failed because XYZ",
	}

	err := service.FailAC(ctx, failInput)
	if err == nil {
		t.Fatal("FailAC() should fail for non-existent AC")
	}
}

// TestACService_FailAC_EmptyFeedback tests failing AC with empty feedback
func TestACService_FailAC_EmptyFeedback(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return ac, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Fail AC with empty feedback
	failInput := dto.FailACDTO{
		ID:       ac.ID, // MUST set ID for fail operations
		Feedback: "", // Empty feedback
	}

	err := service.FailAC(ctx, failInput)
	if err == nil {
		t.Fatal("FailAC() should fail with empty feedback")
	}
}

// TestACService_SkipAC_Success tests successful AC skip marking
func TestACService_SkipAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return ac, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.UpdateACFunc = func(ctx context.Context, updatedAC *entities.AcceptanceCriteriaEntity) error {
		ac.Status = updatedAC.Status
		ac.Notes = updatedAC.Notes
		return nil
	}

	// Skip AC
	skipInput := dto.SkipACDTO{
		ID:     ac.ID,
		Reason: "No longer applicable due to architecture change",
	}

	err := service.SkipAC(ctx, skipInput)
	if err != nil {
		t.Fatalf("SkipAC() failed: %v", err)
	}

	// Check status changed
	gotAC, err := service.GetAC(ctx, "TM-ac-1")
	if err != nil {
		t.Fatalf("GetAC() failed: %v", err)
	}

	if gotAC.Status != entities.ACStatusSkipped {
		t.Errorf("ac.Status = %q, want %q", gotAC.Status, entities.ACStatusSkipped)
	}
	if gotAC.Notes != skipInput.Reason {
		t.Errorf("ac.Notes = %q, want %q", gotAC.Notes, skipInput.Reason)
	}
}

// TestACService_SkipAC_NotFound tests skipping non-existent AC
func TestACService_SkipAC_NotFound(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	skipInput := dto.SkipACDTO{
		ID:     "nonexistent",
		Reason: "Some reason",
	}

	err := service.SkipAC(ctx, skipInput)
	if err == nil {
		t.Fatal("SkipAC() should fail for non-existent AC")
	}
}

// TestACService_SkipAC_EmptyReason tests skipping AC with empty reason
func TestACService_SkipAC_EmptyReason(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return ac, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Skip AC with empty reason
	skipInput := dto.SkipACDTO{
		ID:     ac.ID,
		Reason: "", // Empty reason
	}

	err := service.SkipAC(ctx, skipInput)
	if err == nil {
		t.Fatal("SkipAC() should fail with empty reason")
	}
}

// TestACService_DeleteAC_Success tests successful AC deletion
func TestACService_DeleteAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	// Track deletion state
	deleted := false

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" && !deleted {
			return ac, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.DeleteACFunc = func(ctx context.Context, id string) error {
		if id == "TM-ac-1" {
			deleted = true
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	// Delete AC
	err := service.DeleteAC(ctx, "TM-ac-1")
	if err != nil {
		t.Fatalf("DeleteAC() failed: %v", err)
	}

	// Verify AC is deleted
	_, err = service.GetAC(ctx, "TM-ac-1")
	if err == nil {
		t.Fatal("GetAC() should fail after deletion")
	}
}

// TestACService_DeleteAC_NotFound tests deleting non-existent AC
func TestACService_DeleteAC_NotFound(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	mockACRepo.DeleteACFunc = func(ctx context.Context, id string) error {
		return pluginsdk.ErrNotFound
	}

	err := service.DeleteAC(ctx, "nonexistent")
	if err == nil {
		t.Fatal("DeleteAC() should fail for non-existent AC")
	}
}

// TestACService_GetAC_Success tests successful AC retrieval
func TestACService_GetAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		if id == "TM-ac-1" {
			return ac, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Get AC
	gotAC, err := service.GetAC(ctx, "TM-ac-1")
	if err != nil {
		t.Fatalf("GetAC() failed: %v", err)
	}

	if gotAC.ID != ac.ID {
		t.Errorf("ac.ID = %q, want %q", gotAC.ID, ac.ID)
	}
	if gotAC.Description != ac.Description {
		t.Errorf("ac.Description = %q, want %q", gotAC.Description, ac.Description)
	}
}

// TestACService_GetAC_NotFound tests retrieving non-existent AC
func TestACService_GetAC_NotFound(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.GetACFunc = func(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	_, err := service.GetAC(ctx, "nonexistent")
	if err == nil {
		t.Fatal("GetAC() should fail for non-existent AC")
	}
}

// TestACService_ListAC_Success tests successful AC listing
func TestACService_ListAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac1 := createTestACEntity(t, "TM-ac-1", "TM-task-1")
	ac2 := createTestACEntity(t, "TM-ac-2", "TM-task-1")
	ac3 := createTestACEntity(t, "TM-ac-3", "TM-task-1")

	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == "TM-task-1" {
			return []*entities.AcceptanceCriteriaEntity{ac1, ac2, ac3}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// List ACs
	acs, err := service.ListAC(ctx, "TM-task-1")
	if err != nil {
		t.Fatalf("ListAC() failed: %v", err)
	}

	if len(acs) != 3 {
		t.Fatalf("ListAC() returned %d ACs, want 3", len(acs))
	}
}

// TestACService_ListAC_Empty tests listing ACs for task with no ACs
func TestACService_ListAC_Empty(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// List ACs (should be empty)
	acs, err := service.ListAC(ctx, "TM-task-1")
	if err != nil {
		t.Fatalf("ListAC() failed: %v", err)
	}

	if len(acs) != 0 {
		t.Fatalf("ListAC() returned %d ACs, want 0", len(acs))
	}
}

// TestACService_ListACByIteration_Success tests listing ACs by iteration
func TestACService_ListACByIteration_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.ListACByIterationFunc = func(ctx context.Context, iterationNum int) ([]*entities.AcceptanceCriteriaEntity, error) {
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// Note: We would need to create an iteration and add the task to it
	// For simplicity, this test just checks the method doesn't error on empty result
	acs, err := service.ListACByIteration(ctx, 1)
	if err != nil {
		t.Fatalf("ListACByIteration() failed: %v", err)
	}

	// Should be empty since we didn't create iteration
	if len(acs) != 0 {
		t.Fatalf("ListACByIteration() returned %d ACs, want 0", len(acs))
	}
}

// TestACService_ListACByIteration_Empty tests listing ACs for iteration with no tasks
func TestACService_ListACByIteration_Empty(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.ListACByIterationFunc = func(ctx context.Context, iterationNum int) ([]*entities.AcceptanceCriteriaEntity, error) {
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// List ACs for non-existent iteration
	acs, err := service.ListACByIteration(ctx, 999)
	if err != nil {
		t.Fatalf("ListACByIteration() failed: %v", err)
	}

	if len(acs) != 0 {
		t.Fatalf("ListACByIteration() returned %d ACs, want 0", len(acs))
	}
}

// TestACService_ListFailedAC_Success tests listing failed ACs
func TestACService_ListFailedAC_Success(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	ac := createTestACEntity(t, "TM-ac-1", "TM-task-1")
	ac.Status = entities.ACStatusFailed
	ac.Notes = "Failed because XYZ"

	mockACRepo.ListFailedACFunc = func(ctx context.Context, filters entities.ACFilters) ([]*entities.AcceptanceCriteriaEntity, error) {
		return []*entities.AcceptanceCriteriaEntity{ac}, nil
	}

	// List failed ACs
	filters := entities.ACFilters{}
	acs, err := service.ListFailedAC(ctx, filters)
	if err != nil {
		t.Fatalf("ListFailedAC() failed: %v", err)
	}

	if len(acs) != 1 {
		t.Fatalf("ListFailedAC() returned %d ACs, want 1", len(acs))
	}
	if acs[0].Status != entities.ACStatusFailed {
		t.Errorf("acs[0].Status = %q, want %q", acs[0].Status, entities.ACStatusFailed)
	}
}

// TestACService_ListFailedAC_Empty tests listing failed ACs with no failures
func TestACService_ListFailedAC_Empty(t *testing.T) {
	service, ctx, mockACRepo, _, _ := setupACTestService(t)

	mockACRepo.ListFailedACFunc = func(ctx context.Context, filters entities.ACFilters) ([]*entities.AcceptanceCriteriaEntity, error) {
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// List failed ACs (should be empty)
	filters := entities.ACFilters{}
	acs, err := service.ListFailedAC(ctx, filters)
	if err != nil {
		t.Fatalf("ListFailedAC() failed: %v", err)
	}

	if len(acs) != 0 {
		t.Fatalf("ListFailedAC() returned %d ACs, want 0", len(acs))
	}
}

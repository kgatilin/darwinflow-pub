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

// setupIterationTestService creates a test service with mock repositories
func setupIterationTestService(t *testing.T) (*application.IterationApplicationService, context.Context, *mocks.MockIterationRepository, *mocks.MockTaskRepository, *mocks.MockAggregateRepository, *services.IterationService) {
	mockIterationRepo := &mocks.MockIterationRepository{}
	mockTaskRepo := &mocks.MockTaskRepository{}
	mockAggregateRepo := &mocks.MockAggregateRepository{}

	iterationService := services.NewIterationService()
	validationService := services.NewValidationService()

	service := application.NewIterationApplicationService(mockIterationRepo, mockTaskRepo, mockAggregateRepo, iterationService, validationService)
	ctx := context.Background()

	return service, ctx, mockIterationRepo, mockTaskRepo, mockAggregateRepo, iterationService
}

// createTestIterationEntity creates a test iteration entity for mock configuration
func createTestIterationEntity(t *testing.T, number int, status string) *entities.IterationEntity {
	now := time.Now().UTC()
	iteration, err := entities.NewIterationEntity(number, "Test Iteration", "Test goal", "Test deliverable", []string{}, status, 100, time.Time{}, time.Time{}, now, now)
	if err != nil {
		t.Fatalf("failed to create test iteration: %v", err)
	}
	return iteration
}

// createTestTaskEntity creates a test task entity for mock configuration
func createTestTaskEntity(t *testing.T, taskID string) *entities.TaskEntity {
	now := time.Now().UTC()
	task, err := entities.NewTaskEntity(taskID, "TM-track-1", "Test Task", "", "todo", 500, "", now, now)
	if err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}
	return task
}

// ============================================================================
// CreateIteration Tests
// ============================================================================

func TestIterationService_CreateIteration_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	// Configure mocks
	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound // Iteration doesn't exist yet
	}

	mockIterationRepo.SaveIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		return nil // Success
	}

	input := dto.CreateIterationDTO{
		Number:      1,
		Name:        "Sprint 1",
		Goal:        "Foundation",
		Deliverable: "Core features",
		Status:      "planned",
	}

	iteration, err := service.CreateIteration(ctx, input)
	if err != nil {
		t.Fatalf("CreateIteration() failed: %v", err)
	}

	if iteration.Number != input.Number {
		t.Errorf("iteration.Number = %d, want %d", iteration.Number, input.Number)
	}
	if iteration.Name != input.Name {
		t.Errorf("iteration.Name = %q, want %q", iteration.Name, input.Name)
	}
	if iteration.Goal != input.Goal {
		t.Errorf("iteration.Goal = %q, want %q", iteration.Goal, input.Goal)
	}
	if iteration.Status != input.Status {
		t.Errorf("iteration.Status = %q, want %q", iteration.Status, input.Status)
	}
}

func TestIterationService_CreateIteration_DuplicateNumber(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	// Track whether iteration was created
	created := false

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if created {
			// Return existing iteration on second call
			return createTestIterationEntity(t, 1, "planned"), nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.SaveIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		if !created {
			created = true
			return nil // First call succeeds
		}
		return pluginsdk.ErrAlreadyExists // Should not be reached
	}

	input := dto.CreateIterationDTO{
		Number:      1,
		Name:        "Sprint 1",
		Goal:        "Foundation",
		Deliverable: "Core features",
		Status:      "planned",
	}

	// Create first iteration
	_, err := service.CreateIteration(ctx, input)
	if err != nil {
		t.Fatalf("CreateIteration() failed: %v", err)
	}

	// Try to create duplicate
	_, err = service.CreateIteration(ctx, input)
	if err == nil {
		t.Fatal("CreateIteration() should fail with duplicate number")
	}
}

func TestIterationService_CreateIteration_EmptyName(t *testing.T) {
	service, ctx, _, _, _, _ := setupIterationTestService(t)

	input := dto.CreateIterationDTO{
		Number:      1,
		Name:        "", // Empty name
		Goal:        "Foundation",
		Deliverable: "Core features",
		Status:      "planned",
	}

	_, err := service.CreateIteration(ctx, input)
	if err == nil {
		t.Fatal("CreateIteration() should fail with empty name")
	}
}

func TestIterationService_CreateIteration_DefaultStatus(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.SaveIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		return nil
	}

	input := dto.CreateIterationDTO{
		Number:      1,
		Name:        "Sprint 1",
		Goal:        "Foundation",
		Deliverable: "Core features",
		Status:      "", // Empty status should default to "planned"
	}

	iteration, err := service.CreateIteration(ctx, input)
	if err != nil {
		t.Fatalf("CreateIteration() failed: %v", err)
	}

	if iteration.Status != "planned" {
		t.Errorf("iteration.Status = %q, want %q", iteration.Status, "planned")
	}
}

func TestIterationService_CreateIteration_InvalidNumber(t *testing.T) {
	service, ctx, _, _, _, _ := setupIterationTestService(t)

	input := dto.CreateIterationDTO{
		Number:      0, // Invalid: must be positive
		Name:        "Sprint 0",
		Goal:        "Foundation",
		Deliverable: "Core features",
		Status:      "planned",
	}

	_, err := service.CreateIteration(ctx, input)
	if err == nil {
		t.Fatal("CreateIteration() should fail with invalid number")
	}
}

func TestIterationService_CreateIteration_InvalidStatus(t *testing.T) {
	service, ctx, _, _, _, _ := setupIterationTestService(t)

	input := dto.CreateIterationDTO{
		Number:      1,
		Name:        "Sprint 1",
		Goal:        "Foundation",
		Deliverable: "Core features",
		Status:      "invalid-status", // Invalid status
	}

	_, err := service.CreateIteration(ctx, input)
	if err == nil {
		t.Fatal("CreateIteration() should fail with invalid status")
	}
}

// ============================================================================
// UpdateIteration Tests
// ============================================================================

func TestIterationService_UpdateIteration_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	original := createTestIterationEntity(t, 1, "planned")

	// Configure mocks
	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return original, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.UpdateIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		return nil
	}

	// Update iteration
	newName := "Updated Name"
	newGoal := "Updated Goal"
	updateInput := dto.UpdateIterationDTO{
		Number: 1,
		Name:   &newName,
		Goal:   &newGoal,
	}

	iteration, err := service.UpdateIteration(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateIteration() failed: %v", err)
	}

	if iteration.Name != newName {
		t.Errorf("iteration.Name = %q, want %q", iteration.Name, newName)
	}
	if iteration.Goal != newGoal {
		t.Errorf("iteration.Goal = %q, want %q", iteration.Goal, newGoal)
	}
}

func TestIterationService_UpdateIteration_NotFound(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	newName := "Updated Name"
	updateInput := dto.UpdateIterationDTO{
		Number: 999,
		Name:   &newName,
	}

	_, err := service.UpdateIteration(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateIteration() should fail for non-existent iteration")
	}
}

func TestIterationService_UpdateIteration_PartialUpdate(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	original := createTestIterationEntity(t, 1, "planned")
	originalGoal := original.Goal
	originalDeliverable := original.Deliverable

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return original, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.UpdateIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		return nil
	}

	// Update only name
	newName := "Updated Name"
	updateInput := dto.UpdateIterationDTO{
		Number: 1,
		Name:   &newName,
	}

	iteration, err := service.UpdateIteration(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateIteration() failed: %v", err)
	}

	if iteration.Name != newName {
		t.Errorf("iteration.Name = %q, want %q", iteration.Name, newName)
	}
	// Other fields should remain unchanged
	if iteration.Goal != originalGoal {
		t.Errorf("iteration.Goal changed: got %q, want %q", iteration.Goal, originalGoal)
	}
	if iteration.Deliverable != originalDeliverable {
		t.Errorf("iteration.Deliverable changed: got %q, want %q", iteration.Deliverable, originalDeliverable)
	}
}

func TestIterationService_UpdateIteration_EmptyName(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	original := createTestIterationEntity(t, 1, "planned")

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return original, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to update with empty name
	emptyName := ""
	updateInput := dto.UpdateIterationDTO{
		Number: 1,
		Name:   &emptyName,
	}

	_, err := service.UpdateIteration(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateIteration() should fail with empty name")
	}
}

// ============================================================================
// DeleteIteration Tests
// ============================================================================

func TestIterationService_DeleteIteration_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "planned")

	// Configure mocks for creation and deletion
	callCount := 0
	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		callCount++
		if callCount == 1 {
			return iteration, nil // First call (before delete) returns iteration
		}
		return nil, pluginsdk.ErrNotFound // After delete, not found
	}

	mockIterationRepo.DeleteIterationFunc = func(ctx context.Context, number int) error {
		if number == 1 {
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	// Delete iteration
	err := service.DeleteIteration(ctx, 1)
	if err != nil {
		t.Fatalf("DeleteIteration() failed: %v", err)
	}

	// Verify iteration is deleted
	_, err = service.GetIteration(ctx, 1)
	if err == nil {
		t.Fatal("GetIteration() should fail after deletion")
	}
}

func TestIterationService_DeleteIteration_NotFound(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.DeleteIterationFunc = func(ctx context.Context, number int) error {
		return pluginsdk.ErrNotFound
	}

	err := service.DeleteIteration(ctx, 999)
	if err == nil {
		t.Fatal("DeleteIteration() should fail for non-existent iteration")
	}
}

// ============================================================================
// Lifecycle Tests
// ============================================================================

func TestIterationService_StartIteration_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "planned")

	// Configure mocks
	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound // No current iteration
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return iteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.StartIterationFunc = func(ctx context.Context, iterationNum int) error {
		if iterationNum == 1 {
			iteration.Status = "current"
			now := time.Now().UTC()
			iteration.StartedAt = &now
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	// Start iteration
	err := service.StartIteration(ctx, 1)
	if err != nil {
		t.Fatalf("StartIteration() failed: %v", err)
	}

	// Verify status changed
	gotIteration, err := service.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("GetIteration() failed: %v", err)
	}

	if gotIteration.Status != "current" {
		t.Errorf("iteration.Status = %q, want %q", gotIteration.Status, "current")
	}
	if gotIteration.StartedAt == nil {
		t.Error("iteration.StartedAt should be set")
	}
}

func TestIterationService_StartIteration_NotFound(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	err := service.StartIteration(ctx, 999)
	if err == nil {
		t.Fatal("StartIteration() should fail for non-existent iteration")
	}
}

func TestIterationService_StartIteration_AlreadyCurrent(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	currentIteration := createTestIterationEntity(t, 1, "current")
	plannedIteration := createTestIterationEntity(t, 2, "planned")

	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return currentIteration, nil // Iteration 1 is current
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return currentIteration, nil
		}
		if number == 2 {
			return plannedIteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to start second iteration (should fail because first is current)
	err := service.StartIteration(ctx, 2)
	if err == nil {
		t.Fatal("StartIteration() should fail when another iteration is already current")
	}
}

func TestIterationService_StartIteration_AlreadyStarted(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	currentIteration := createTestIterationEntity(t, 1, "current")

	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return currentIteration, nil
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return currentIteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to start again (should fail because iteration is already current)
	err := service.StartIteration(ctx, 1)
	if err == nil {
		t.Fatal("StartIteration() should fail when iteration is already current")
	}
}

func TestIterationService_StartIteration_CompleteFirst(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	completedIteration := createTestIterationEntity(t, 1, "complete")

	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return completedIteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to start completed iteration (should fail)
	err := service.StartIteration(ctx, 1)
	if err == nil {
		t.Fatal("StartIteration() should fail for completed iteration")
	}
}

func TestIterationService_CompleteIteration_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "current")

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return iteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.CompleteIterationFunc = func(ctx context.Context, iterationNum int) error {
		if iterationNum == 1 {
			iteration.Status = "complete"
			now := time.Now().UTC()
			iteration.CompletedAt = &now
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	// Complete iteration
	err := service.CompleteIteration(ctx, 1)
	if err != nil {
		t.Fatalf("CompleteIteration() failed: %v", err)
	}

	// Verify status changed
	gotIteration, err := service.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("GetIteration() failed: %v", err)
	}

	if gotIteration.Status != "complete" {
		t.Errorf("iteration.Status = %q, want %q", gotIteration.Status, "complete")
	}
	if gotIteration.CompletedAt == nil {
		t.Error("iteration.CompletedAt should be set")
	}
}

func TestIterationService_CompleteIteration_NotFound(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	err := service.CompleteIteration(ctx, 999)
	if err == nil {
		t.Fatal("CompleteIteration() should fail for non-existent iteration")
	}
}

func TestIterationService_CompleteIteration_NotStarted(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "planned")

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return iteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to complete without starting
	err := service.CompleteIteration(ctx, 1)
	if err == nil {
		t.Fatal("CompleteIteration() should fail for non-started iteration")
	}
}

// ============================================================================
// Task Management Tests
// ============================================================================

func TestIterationService_AddTask_Success(t *testing.T) {
	service, ctx, mockIterationRepo, mockTaskRepo, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "planned")
	task := createTestTaskEntity(t, "TM-task-1")

	// Configure mocks
	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return iteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == "TM-task-1" {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.AddTaskToIterationFunc = func(ctx context.Context, iterationNum int, taskID string) error {
		if iterationNum == 1 && taskID == "TM-task-1" {
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	mockIterationRepo.GetIterationTasksFunc = func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
		if iterationNum == 1 {
			return []*entities.TaskEntity{task}, nil
		}
		return []*entities.TaskEntity{}, nil
	}

	// Add task to iteration
	err := service.AddTask(ctx, 1, "TM-task-1")
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Verify task was added
	tasks, err := service.GetIterationTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetIterationTasks() failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("GetIterationTasks() returned %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "TM-task-1" {
		t.Errorf("tasks[0].ID = %q, want %q", tasks[0].ID, "TM-task-1")
	}
}

func TestIterationService_AddTask_IterationNotFound(t *testing.T) {
	service, ctx, mockIterationRepo, mockTaskRepo, _, _ := setupIterationTestService(t)

	task := createTestTaskEntity(t, "TM-task-1")

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == "TM-task-1" {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to add task to non-existent iteration
	err := service.AddTask(ctx, 999, "TM-task-1")
	if err == nil {
		t.Fatal("AddTask() should fail for non-existent iteration")
	}
}

func TestIterationService_AddTask_TaskNotFound(t *testing.T) {
	service, ctx, mockIterationRepo, mockTaskRepo, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "planned")

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return iteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	// Try to add non-existent task
	err := service.AddTask(ctx, 1, "nonexistent")
	if err == nil {
		t.Fatal("AddTask() should fail for non-existent task")
	}
	if err != pluginsdk.ErrNotFound {
		// Check if it's wrapped
		var found bool
		for e := err; e != nil; {
			if e == pluginsdk.ErrNotFound {
				found = true
				break
			}
			// Try to unwrap
			unwrapped, ok := e.(interface{ Unwrap() error })
			if !ok {
				break
			}
			e = unwrapped.Unwrap()
		}
		if !found {
			t.Errorf("AddTask() error should be or wrap ErrNotFound, got: %v", err)
		}
	}
}

func TestIterationService_RemoveTask_Success(t *testing.T) {
	service, ctx, mockIterationRepo, mockTaskRepo, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "planned")
	task := createTestTaskEntity(t, "TM-task-1")

	// Track whether task is added
	taskAdded := false

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return iteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == "TM-task-1" {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.AddTaskToIterationFunc = func(ctx context.Context, iterationNum int, taskID string) error {
		if iterationNum == 1 && taskID == "TM-task-1" {
			taskAdded = true
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	mockIterationRepo.RemoveTaskFromIterationFunc = func(ctx context.Context, iterationNum int, taskID string) error {
		if iterationNum == 1 && taskID == "TM-task-1" {
			taskAdded = false
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	mockIterationRepo.GetIterationTasksFunc = func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
		if iterationNum == 1 {
			if taskAdded {
				return []*entities.TaskEntity{task}, nil
			}
			return []*entities.TaskEntity{}, nil
		}
		return []*entities.TaskEntity{}, nil
	}

	// Add task
	err := service.AddTask(ctx, 1, "TM-task-1")
	if err != nil {
		t.Fatalf("AddTask() failed: %v", err)
	}

	// Remove task from iteration
	err = service.RemoveTask(ctx, 1, "TM-task-1")
	if err != nil {
		t.Fatalf("RemoveTask() failed: %v", err)
	}

	// Verify task was removed
	tasks, err := service.GetIterationTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetIterationTasks() failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Fatalf("GetIterationTasks() returned %d tasks, want 0", len(tasks))
	}
}

func TestIterationService_RemoveTask_IterationNotFound(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	err := service.RemoveTask(ctx, 999, "TM-task-1")
	if err == nil {
		t.Fatal("RemoveTask() should fail for non-existent iteration")
	}
}

// ============================================================================
// Read Operations Tests
// ============================================================================

func TestIterationService_GetIteration_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	iteration := createTestIterationEntity(t, 1, "planned")

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		if number == 1 {
			return iteration, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Get iteration
	gotIteration, err := service.GetIteration(ctx, 1)
	if err != nil {
		t.Fatalf("GetIteration() failed: %v", err)
	}

	if gotIteration.Number != iteration.Number {
		t.Errorf("iteration.Number = %d, want %d", gotIteration.Number, iteration.Number)
	}
	if gotIteration.Name != iteration.Name {
		t.Errorf("iteration.Name = %q, want %q", gotIteration.Name, iteration.Name)
	}
}

func TestIterationService_GetIteration_NotFound(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	_, err := service.GetIteration(ctx, 999)
	if err == nil {
		t.Fatal("GetIteration() should fail for non-existent iteration")
	}
}

func TestIterationService_GetCurrentIteration_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	currentIteration := createTestIterationEntity(t, 1, "current")

	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return currentIteration, nil
	}

	// Get current iteration
	result, err := service.GetCurrentIteration(ctx)
	if err != nil {
		t.Fatalf("GetCurrentIteration() failed: %v", err)
	}

	// Verify it's not a fallback
	if result.IsFallback {
		t.Error("expected IsFallback = false for current iteration")
	}
	if result.FallbackMsg != "" {
		t.Errorf("expected empty FallbackMsg, got %q", result.FallbackMsg)
	}

	// Verify iteration data
	iteration, ok := result.Iteration.(*entities.IterationEntity)
	if !ok {
		t.Fatal("expected Iteration to be *entities.IterationEntity")
	}
	if iteration.Number != 1 {
		t.Errorf("iteration.Number = %d, want 1", iteration.Number)
	}
	if iteration.Status != "current" {
		t.Errorf("iteration.Status = %q, want %q", iteration.Status, "current")
	}
}

func TestIterationService_GetCurrentIteration_FallbackToPlanned(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	plannedIteration := createTestIterationEntity(t, 2, "planned")

	// No current iteration
	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	// But there is a planned iteration
	mockIterationRepo.GetNextPlannedIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return plannedIteration, nil
	}

	// Get current iteration (should fallback to planned)
	result, err := service.GetCurrentIteration(ctx)
	if err != nil {
		t.Fatalf("GetCurrentIteration() failed: %v", err)
	}

	// Verify it IS a fallback
	if !result.IsFallback {
		t.Error("expected IsFallback = true when falling back to planned")
	}
	if result.FallbackMsg == "" {
		t.Error("expected FallbackMsg to be set")
	}
	if result.FallbackMsg != "No current iteration. Showing next planned iteration: Test Iteration" {
		t.Errorf("unexpected FallbackMsg: %q", result.FallbackMsg)
	}

	// Verify iteration data
	iteration, ok := result.Iteration.(*entities.IterationEntity)
	if !ok {
		t.Fatal("expected Iteration to be *entities.IterationEntity")
	}
	if iteration.Number != 2 {
		t.Errorf("iteration.Number = %d, want 2", iteration.Number)
	}
	if iteration.Status != "planned" {
		t.Errorf("iteration.Status = %q, want %q", iteration.Status, "planned")
	}
}

func TestIterationService_GetCurrentIteration_NoCurrentOrPlanned(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	// No current iteration
	mockIterationRepo.GetCurrentIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	// No planned iterations either
	mockIterationRepo.GetNextPlannedIterationFunc = func(ctx context.Context) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	// Get current iteration (should return nil with message)
	result, err := service.GetCurrentIteration(ctx)
	if err != nil {
		t.Fatalf("GetCurrentIteration() should not error when no iterations exist: %v", err)
	}

	// Verify fallback state
	if !result.IsFallback {
		t.Error("expected IsFallback = true when no iterations")
	}
	if result.Iteration != nil {
		t.Error("expected Iteration to be nil when no iterations")
	}
	if result.FallbackMsg != "No current or planned iterations found" {
		t.Errorf("unexpected FallbackMsg: %q", result.FallbackMsg)
	}
}

func TestIterationService_ListIterations_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	iteration1 := createTestIterationEntity(t, 1, "planned")
	iteration2 := createTestIterationEntity(t, 2, "planned")
	iteration3 := createTestIterationEntity(t, 3, "planned")

	mockIterationRepo.ListIterationsFunc = func(ctx context.Context) ([]*entities.IterationEntity, error) {
		return []*entities.IterationEntity{iteration1, iteration2, iteration3}, nil
	}

	// List iterations
	iterations, err := service.ListIterations(ctx)
	if err != nil {
		t.Fatalf("ListIterations() failed: %v", err)
	}

	if len(iterations) != 3 {
		t.Fatalf("ListIterations() returned %d iterations, want 3", len(iterations))
	}
}

func TestIterationService_ListIterations_Empty(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.ListIterationsFunc = func(ctx context.Context) ([]*entities.IterationEntity, error) {
		return []*entities.IterationEntity{}, nil
	}

	// List iterations from empty database
	iterations, err := service.ListIterations(ctx)
	if err != nil {
		t.Fatalf("ListIterations() failed: %v", err)
	}

	if len(iterations) != 0 {
		t.Fatalf("ListIterations() returned %d iterations, want 0", len(iterations))
	}
}

func TestIterationService_GetIterationTasks_Success(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	task1 := createTestTaskEntity(t, "TM-task-1")
	task2 := createTestTaskEntity(t, "TM-task-2")

	mockIterationRepo.GetIterationTasksFunc = func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
		if iterationNum == 1 {
			return []*entities.TaskEntity{task1, task2}, nil
		}
		return []*entities.TaskEntity{}, nil
	}

	// Get iteration tasks
	tasks, err := service.GetIterationTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetIterationTasks() failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("GetIterationTasks() returned %d tasks, want 2", len(tasks))
	}
}

func TestIterationService_GetIterationTasks_Empty(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	mockIterationRepo.GetIterationTasksFunc = func(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
		return []*entities.TaskEntity{}, nil
	}

	// Get iteration tasks (should be empty)
	tasks, err := service.GetIterationTasks(ctx, 1)
	if err != nil {
		t.Fatalf("GetIterationTasks() failed: %v", err)
	}

	if len(tasks) != 0 {
		t.Fatalf("GetIterationTasks() returned %d tasks, want 0", len(tasks))
	}
}

// ============================================================================
// Iteration Number Generation Tests (TM-ac-522)
// ============================================================================

// TestIterationService_CreateIteration_NumberGenerationSequential tests that
// iteration numbers are generated sequentially starting from 1
func TestIterationService_CreateIteration_NumberGenerationSequential(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	// Track created iterations
	var createdIterations []*entities.IterationEntity

	mockIterationRepo.ListIterationsFunc = func(ctx context.Context) ([]*entities.IterationEntity, error) {
		return createdIterations, nil
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound // Not exists
	}

	mockIterationRepo.SaveIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		createdIterations = append(createdIterations, iteration)
		return nil
	}

	// Create first iteration - should be 1
	input1 := dto.CreateIterationDTO{
		Name:        "Iteration 1",
		Goal:        "First",
		Deliverable: "First deliverable",
	}
	iteration1, err := service.CreateIteration(ctx, input1)
	if err != nil {
		t.Fatalf("CreateIteration() failed for first iteration: %v", err)
	}
	if iteration1.Number != 1 {
		t.Errorf("First iteration number = %d, want 1", iteration1.Number)
	}

	// Create second iteration - should be 2
	input2 := dto.CreateIterationDTO{
		Name:        "Iteration 2",
		Goal:        "Second",
		Deliverable: "Second deliverable",
	}
	iteration2, err := service.CreateIteration(ctx, input2)
	if err != nil {
		t.Fatalf("CreateIteration() failed for second iteration: %v", err)
	}
	if iteration2.Number != 2 {
		t.Errorf("Second iteration number = %d, want 2", iteration2.Number)
	}

	// Create third iteration - should be 3
	input3 := dto.CreateIterationDTO{
		Name:        "Iteration 3",
		Goal:        "Third",
		Deliverable: "Third deliverable",
	}
	iteration3, err := service.CreateIteration(ctx, input3)
	if err != nil {
		t.Fatalf("CreateIteration() failed for third iteration: %v", err)
	}
	if iteration3.Number != 3 {
		t.Errorf("Third iteration number = %d, want 3", iteration3.Number)
	}
}

// TestIterationService_CreateIteration_NumberGenerationWithGaps tests that
// iteration numbering uses MAX+1, not COUNT+1, when there are gaps in numbers
func TestIterationService_CreateIteration_NumberGenerationWithGaps(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	// Simulate iterations 1, 2, 3 exist, then iteration 2 is deleted
	// So we have [1, 3] in the database
	iteration1 := createTestIterationEntity(t, 1, "planned")
	iteration3 := createTestIterationEntity(t, 3, "planned")
	existingIterations := []*entities.IterationEntity{iteration1, iteration3}

	mockIterationRepo.ListIterationsFunc = func(ctx context.Context) ([]*entities.IterationEntity, error) {
		return existingIterations, nil
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound // Not exists
	}

	mockIterationRepo.SaveIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		return nil
	}

	// Create new iteration
	input := dto.CreateIterationDTO{
		Name:        "New Iteration",
		Goal:        "After gap",
		Deliverable: "Should be 4",
	}

	iteration, err := service.CreateIteration(ctx, input)
	if err != nil {
		t.Fatalf("CreateIteration() failed: %v", err)
	}

	// CRITICAL: Should be 4 (MAX+1), NOT 3 (COUNT+1)
	// MAX([1,3]) = 3, so next should be 3+1 = 4
	// COUNT([1,3]) = 2, so next would be 2+1 = 3 (WRONG)
	if iteration.Number != 4 {
		t.Errorf("Iteration number with gaps = %d, want 4 (MAX+1, not COUNT+1)", iteration.Number)
	}
}

// TestIterationService_CreateIteration_NumberGenerationAfterDeleteAll tests that
// iteration numbering starts from 1 when all iterations are deleted
func TestIterationService_CreateIteration_NumberGenerationAfterDeleteAll(t *testing.T) {
	service, ctx, mockIterationRepo, _, _, _ := setupIterationTestService(t)

	// Simulate all iterations deleted - empty list
	mockIterationRepo.ListIterationsFunc = func(ctx context.Context) ([]*entities.IterationEntity, error) {
		return []*entities.IterationEntity{}, nil
	}

	mockIterationRepo.GetIterationFunc = func(ctx context.Context, number int) (*entities.IterationEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	mockIterationRepo.SaveIterationFunc = func(ctx context.Context, iteration *entities.IterationEntity) error {
		return nil
	}

	// Create new iteration after all deleted
	input := dto.CreateIterationDTO{
		Name:        "Fresh Start",
		Goal:        "After delete all",
		Deliverable: "Should be 1",
	}

	iteration, err := service.CreateIteration(ctx, input)
	if err != nil {
		t.Fatalf("CreateIteration() failed: %v", err)
	}

	// Should be 1 when no iterations exist
	// MAX([]) = 0, so next should be 0+1 = 1
	if iteration.Number != 1 {
		t.Errorf("Iteration number after delete all = %d, want 1", iteration.Number)
	}
}

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

// setupTaskTestService creates a test service with mock repositories
func setupTaskTestService(t *testing.T) (*application.TaskApplicationService, context.Context, *mocks.MockTaskRepository, *mocks.MockTrackRepository, *mocks.MockAggregateRepository, *mocks.MockAcceptanceCriteriaRepository) {
	mockTaskRepo := &mocks.MockTaskRepository{}
	mockTrackRepo := &mocks.MockTrackRepository{}
	mockAggregateRepo := &mocks.MockAggregateRepository{}
	mockACRepo := &mocks.MockAcceptanceCriteriaRepository{}
	validationService := services.NewValidationService()

	service := application.NewTaskApplicationService(mockTaskRepo, mockTrackRepo, mockAggregateRepo, mockACRepo, validationService)
	ctx := context.Background()

	return service, ctx, mockTaskRepo, mockTrackRepo, mockAggregateRepo, mockACRepo
}

// createTestTrackForMock creates a test track entity for mock configuration
func createTestTrackForMock(t *testing.T) *entities.TrackEntity {
	now := time.Now().UTC()
	track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Test Track", "Description", "not-started", 500, []string{}, now, now)
	if err != nil {
		t.Fatalf("failed to create test track: %v", err)
	}
	return track
}

// ============================================================================
// CreateTask Tests
// ============================================================================

// TestTaskService_CreateTask_Success tests successful task creation
func TestTaskService_CreateTask_Success(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	// Configure mocks
	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		if id == track.ID {
			return track, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.SaveTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil // Success
	}

	input := dto.CreateTaskDTO{
		TrackID:     track.ID,
		Title:       "Test Task",
		Description: "Test description",
		Status:      "todo",
		Rank:        100,
	}

	task, err := service.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if task.ID == "" {
		t.Error("task.ID should not be empty (auto-generated)")
	}
	if task.Title != input.Title {
		t.Errorf("task.Title = %q, want %q", task.Title, input.Title)
	}
	if task.Status != input.Status {
		t.Errorf("task.Status = %q, want %q", task.Status, input.Status)
	}
	if task.Rank != input.Rank {
		t.Errorf("task.Rank = %d, want %d", task.Rank, input.Rank)
	}
	if task.TrackID != input.TrackID {
		t.Errorf("task.TrackID = %q, want %q", task.TrackID, input.TrackID)
	}
}

// TestTaskService_CreateTask_DuplicateID tests task creation with duplicate ID
func TestTaskService_CreateTask_DuplicateID(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	// First call succeeds, second fails with duplicate
	callCount := 0
	mockTaskRepo.SaveTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		callCount++
		if callCount == 1 {
			return nil // First save succeeds
		}
		return pluginsdk.ErrAlreadyExists // Duplicate
	}

	input := dto.CreateTaskDTO{
		TrackID:     track.ID,
		Title:       "Test Task",
		Description: "Test description",
		Status:      "todo",
		Rank:        100,
	}

	// Create first task
	_, err := service.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	// Try to create duplicate task
	_, err = service.CreateTask(ctx, input)
	if err == nil {
		t.Fatal("CreateTask() should fail with duplicate ID")
	}
}

// TestTaskService_CreateTask_EmptyTitle tests task creation with empty title
func TestTaskService_CreateTask_EmptyTitle(t *testing.T) {
	service, ctx, _, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	input := dto.CreateTaskDTO{
		TrackID:     track.ID,
		Title:       "",
		Description: "Test description",
		Status:      "todo",
		Rank:        100,
	}

	_, err := service.CreateTask(ctx, input)
	if err == nil {
		t.Fatal("CreateTask() should fail with empty title")
	}
}

// TestTaskService_CreateTask_TrackNotFound tests task creation with non-existent track
func TestTaskService_CreateTask_TrackNotFound(t *testing.T) {
	service, ctx, _, mockTrackRepo, _, _ := setupTaskTestService(t)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	input := dto.CreateTaskDTO{
		TrackID:     "nonexistent",
		Title:       "Test Task",
		Description: "Test description",
		Status:      "todo",
		Rank:        100,
	}

	_, err := service.CreateTask(ctx, input)
	if err == nil {
		t.Fatal("CreateTask() should fail with non-existent track")
	}
}

// TestTaskService_CreateTask_InvalidRank tests task creation with invalid rank
func TestTaskService_CreateTask_InvalidRank(t *testing.T) {
	service, ctx, _, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	input := dto.CreateTaskDTO{
		TrackID:     track.ID,
		Title:       "Test Task",
		Description: "Test description",
		Status:      "todo",
		Rank:        9999, // Invalid: must be 1-1000
	}

	_, err := service.CreateTask(ctx, input)
	if err == nil {
		t.Fatal("CreateTask() should fail with invalid rank")
	}
}

// TestTaskService_CreateTask_DefaultStatus tests task creation with default status
func TestTaskService_CreateTask_DefaultStatus(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.SaveTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	input := dto.CreateTaskDTO{
		TrackID:     track.ID,
		Title:       "Test Task",
		Description: "Test description",
		Status:      "", // Empty status should default to todo
		Rank:        100,
	}

	task, err := service.CreateTask(ctx, input)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if task.Status != "todo" {
		t.Errorf("task.Status = %q, want %q", task.Status, "todo")
	}
}

// TestTaskService_CreateTask_InvalidStatus tests task creation with invalid status
func TestTaskService_CreateTask_InvalidStatus(t *testing.T) {
	service, ctx, _, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	input := dto.CreateTaskDTO{
		TrackID:     track.ID,
		Title:       "Test Task",
		Description: "Test description",
		Status:      "invalid-status",
		Rank:        100,
	}

	_, err := service.CreateTask(ctx, input)
	if err == nil {
		t.Fatal("CreateTask() should fail with invalid status")
	}
}

// ============================================================================
// UpdateTask Tests
// ============================================================================

// TestTaskService_UpdateTask_Success tests successful task update
func TestTaskService_UpdateTask_Success(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Original Title", "Original description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update task
	newTitle := "Updated Title"
	newStatus := "in-progress"
	newRank := 200
	updateInput := dto.UpdateTaskDTO{
		ID:     existingTask.ID, // MUST set ID for update operations
		Title:  &newTitle,
		Status: &newStatus,
		Rank:   &newRank,
	}

	task, err := service.UpdateTask(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateTask() failed: %v", err)
	}

	if task.Title != newTitle {
		t.Errorf("task.Title = %q, want %q", task.Title, newTitle)
	}
	if task.Status != newStatus {
		t.Errorf("task.Status = %q, want %q", task.Status, newStatus)
	}
	if task.Rank != newRank {
		t.Errorf("task.Rank = %d, want %d", task.Rank, newRank)
	}
}

// TestTaskService_UpdateTask_NotFound tests updating non-existent task
func TestTaskService_UpdateTask_NotFound(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	newTitle := "Updated Title"
	updateInput := dto.UpdateTaskDTO{
		ID:    "nonexistent",
		Title: &newTitle,
	}

	_, err := service.UpdateTask(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateTask() should fail for non-existent task")
	}
}

// TestTaskService_UpdateTask_PartialUpdate tests partial task update
func TestTaskService_UpdateTask_PartialUpdate(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Original Title", "Original description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update only title
	newTitle := "Updated Title"
	updateInput := dto.UpdateTaskDTO{
		ID:    existingTask.ID, // MUST set ID for update operations
		Title: &newTitle,
	}

	task, err := service.UpdateTask(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateTask() failed: %v", err)
	}

	if task.Title != newTitle {
		t.Errorf("task.Title = %q, want %q", task.Title, newTitle)
	}
	// Other fields should remain unchanged
	if task.Description != existingTask.Description {
		t.Errorf("task.Description changed: got %q, want %q", task.Description, existingTask.Description)
	}
	if task.Status != existingTask.Status {
		t.Errorf("task.Status changed: got %q, want %q", task.Status, existingTask.Status)
	}
}

// TestTaskService_UpdateTask_InvalidStatus tests updating with invalid status
func TestTaskService_UpdateTask_InvalidStatus(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to update with invalid status
	invalidStatus := "invalid-status"
	updateInput := dto.UpdateTaskDTO{
		Status: &invalidStatus,
	}

	_, err := service.UpdateTask(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateTask() should fail with invalid status")
	}
}

// TestTaskService_UpdateTask_UpdateTrackID tests updating task's track
func TestTaskService_UpdateTask_UpdateTrackID(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track1 := createTestTrackForMock(t)

	now := time.Now().UTC()
	track2, _ := entities.NewTrackEntity("TM-track-2", "roadmap-1", "Test Track 2", "Description", "not-started", 500, []string{}, now, now)
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track1.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		if id == track1.ID {
			return track1, nil
		}
		if id == track2.ID {
			return track2, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update task to move to second track
	newTrackID := "TM-track-2"
	updateInput := dto.UpdateTaskDTO{
		ID:      existingTask.ID, // MUST set ID for update operations
		TrackID: &newTrackID,
	}

	task, err := service.UpdateTask(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateTask() failed: %v", err)
	}

	if task.TrackID != newTrackID {
		t.Errorf("task.TrackID = %q, want %q", task.TrackID, newTrackID)
	}
}

// TestTaskService_UpdateTask_InvalidTrackID tests updating with non-existent track
func TestTaskService_UpdateTask_InvalidTrackID(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		if id == track.ID {
			return track, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to update with non-existent track
	invalidTrackID := "nonexistent"
	updateInput := dto.UpdateTaskDTO{
		TrackID: &invalidTrackID,
	}

	_, err := service.UpdateTask(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateTask() should fail with non-existent track")
	}
}

// TestTaskService_UpdateTask_EmptyTitle tests updating with empty title
func TestTaskService_UpdateTask_EmptyTitle(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to update with empty title
	emptyTitle := ""
	updateInput := dto.UpdateTaskDTO{
		Title: &emptyTitle,
	}

	_, err := service.UpdateTask(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateTask() should fail with empty title")
	}
}

// TestTaskService_UpdateTask_InvalidRank tests updating with invalid rank
func TestTaskService_UpdateTask_InvalidRank(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to update with invalid rank
	invalidRank := 9999
	updateInput := dto.UpdateTaskDTO{
		Rank: &invalidRank,
	}

	_, err := service.UpdateTask(ctx, updateInput)
	if err == nil {
		t.Fatal("UpdateTask() should fail with invalid rank")
	}
}

// TestTaskService_UpdateTask_UpdateDescription tests updating description
func TestTaskService_UpdateTask_UpdateDescription(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Test Task", "Original description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update description
	newDescription := "Updated description"
	updateInput := dto.UpdateTaskDTO{
		ID:          existingTask.ID, // MUST set ID for update operations
		Description: &newDescription,
	}

	task, err := service.UpdateTask(ctx, updateInput)
	if err != nil {
		t.Fatalf("UpdateTask() failed: %v", err)
	}

	if task.Description != newDescription {
		t.Errorf("task.Description = %q, want %q", task.Description, newDescription)
	}
}

// ============================================================================
// DeleteTask Tests
// ============================================================================

// TestTaskService_DeleteTask_Success tests successful task deletion
func TestTaskService_DeleteTask_Success(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)

	deleted := false
	mockTaskRepo.DeleteTaskFunc = func(ctx context.Context, id string) error {
		if id == "TM-task-1" {
			deleted = true
			return nil
		}
		return pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if deleted && id == "TM-task-1" {
			return nil, pluginsdk.ErrNotFound
		}
		return nil, nil
	}

	// Delete task
	err := service.DeleteTask(ctx, "TM-task-1")
	if err != nil {
		t.Fatalf("DeleteTask() failed: %v", err)
	}

	// Verify task is deleted
	_, err = service.GetTask(ctx, "TM-task-1")
	if err == nil {
		t.Fatal("GetTask() should fail after deletion")
	}
}

// TestTaskService_DeleteTask_NotFound tests deleting non-existent task
func TestTaskService_DeleteTask_NotFound(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)

	mockTaskRepo.DeleteTaskFunc = func(ctx context.Context, id string) error {
		return pluginsdk.ErrNotFound
	}

	err := service.DeleteTask(ctx, "nonexistent")
	if err == nil {
		t.Fatal("DeleteTask() should fail for non-existent task")
	}
}

// ============================================================================
// MoveTask Tests
// ============================================================================

// TestTaskService_MoveTask_Success tests successful task move
func TestTaskService_MoveTask_Success(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track1 := createTestTrackForMock(t)

	now := time.Now().UTC()
	track2, _ := entities.NewTrackEntity("TM-track-2", "roadmap-1", "Test Track 2", "Description", "not-started", 500, []string{}, now, now)
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track1.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	// Use a pointer to track the current task state
	currentTask := existingTask

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		if id == track1.ID {
			return track1, nil
		}
		if id == track2.ID {
			return track2, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == currentTask.ID {
			return currentTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.MoveTaskToTrackFunc = func(ctx context.Context, taskID, newTrackID string) error {
		// Update the task's TrackID
		currentTask.TrackID = newTrackID
		return nil
	}

	// Move task to second track
	err := service.MoveTask(ctx, "TM-task-1", "TM-track-2")
	if err != nil {
		t.Fatalf("MoveTask() failed: %v", err)
	}

	// Verify task was moved
	task, err := service.GetTask(ctx, "TM-task-1")
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}
	if task.TrackID != "TM-track-2" {
		t.Errorf("task.TrackID = %q, want %q", task.TrackID, "TM-track-2")
	}
}

// TestTaskService_MoveTask_TaskNotFound tests moving non-existent task
func TestTaskService_MoveTask_TaskNotFound(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		return track, nil
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	err := service.MoveTask(ctx, "nonexistent", track.ID)
	if err == nil {
		t.Fatal("MoveTask() should fail for non-existent task")
	}
}

// TestTaskService_MoveTask_TrackNotFound tests moving task to non-existent track
func TestTaskService_MoveTask_TrackNotFound(t *testing.T) {
	service, ctx, mockTaskRepo, mockTrackRepo, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	mockTrackRepo.GetTrackFunc = func(ctx context.Context, id string) (*entities.TrackEntity, error) {
		if id == track.ID {
			return track, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Try to move to non-existent track
	err := service.MoveTask(ctx, "TM-task-1", "nonexistent")
	if err == nil {
		t.Fatal("MoveTask() should fail for non-existent track")
	}
}

// ============================================================================
// GetTask Tests
// ============================================================================

// TestTaskService_GetTask_Success tests successful task retrieval
func TestTaskService_GetTask_Success(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	existingTask, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Test Task", "Test description", "todo", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == existingTask.ID {
			return existingTask, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Get task
	task, err := service.GetTask(ctx, "TM-task-1")
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}

	if task.ID != existingTask.ID {
		t.Errorf("task.ID = %q, want %q", task.ID, existingTask.ID)
	}
	if task.Title != existingTask.Title {
		t.Errorf("task.Title = %q, want %q", task.Title, existingTask.Title)
	}
}

// TestTaskService_GetTask_NotFound tests retrieving non-existent task
func TestTaskService_GetTask_NotFound(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		return nil, pluginsdk.ErrNotFound
	}

	_, err := service.GetTask(ctx, "nonexistent")
	if err == nil {
		t.Fatal("GetTask() should fail for non-existent task")
	}
}

// ============================================================================
// ListTasks Tests
// ============================================================================

// TestTaskService_ListTasks_Success tests successful task listing
func TestTaskService_ListTasks_Success(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	task1, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Task 1", "", "todo", 100, "", now, now)
	task2, _ := entities.NewTaskEntity("TM-task-2", track.ID, "Task 2", "", "in-progress", 200, "", now, now)
	task3, _ := entities.NewTaskEntity("TM-task-3", track.ID, "Task 3", "", "done", 300, "", now, now)

	mockTaskRepo.ListTasksFunc = func(ctx context.Context, filters entities.TaskFilters) ([]*entities.TaskEntity, error) {
		return []*entities.TaskEntity{task1, task2, task3}, nil
	}

	// List all tasks
	filters := entities.TaskFilters{}
	results, err := service.ListTasks(ctx, filters)
	if err != nil {
		t.Fatalf("ListTasks() failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("ListTasks() returned %d tasks, want 3", len(results))
	}
}

// TestTaskService_ListTasks_WithFilters tests task listing with status filter
func TestTaskService_ListTasks_WithFilters(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	task2, _ := entities.NewTaskEntity("TM-task-2", track.ID, "Task 2", "", "in-progress", 200, "", now, now)
	task3, _ := entities.NewTaskEntity("TM-task-3", track.ID, "Task 3", "", "in-progress", 300, "", now, now)

	mockTaskRepo.ListTasksFunc = func(ctx context.Context, filters entities.TaskFilters) ([]*entities.TaskEntity, error) {
		if len(filters.Status) > 0 && filters.Status[0] == "in-progress" {
			return []*entities.TaskEntity{task2, task3}, nil
		}
		return []*entities.TaskEntity{}, nil
	}

	// List tasks with status filter
	filters := entities.TaskFilters{Status: []string{"in-progress"}}
	results, err := service.ListTasks(ctx, filters)
	if err != nil {
		t.Fatalf("ListTasks() failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("ListTasks() returned %d tasks, want 2", len(results))
	}

	for _, task := range results {
		if task.Status != "in-progress" {
			t.Errorf("task.Status = %q, want %q", task.Status, "in-progress")
		}
	}
}

// TestTaskService_ListTasks_Empty tests listing tasks from empty database
func TestTaskService_ListTasks_Empty(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)

	mockTaskRepo.ListTasksFunc = func(ctx context.Context, filters entities.TaskFilters) ([]*entities.TaskEntity, error) {
		return []*entities.TaskEntity{}, nil
	}

	// List tasks from empty database
	filters := entities.TaskFilters{}
	results, err := service.ListTasks(ctx, filters)
	if err != nil {
		t.Fatalf("ListTasks() failed: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("ListTasks() returned %d tasks, want 0", len(results))
	}
}

// ============================================================================
// GetBacklogTasks Tests
// ============================================================================

// TestTaskService_GetBacklogTasks_Success tests successful backlog retrieval
func TestTaskService_GetBacklogTasks_Success(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)
	track := createTestTrackForMock(t)

	now := time.Now().UTC()
	task1, _ := entities.NewTaskEntity("TM-task-1", track.ID, "Task 1", "", "todo", 100, "", now, now)
	task3, _ := entities.NewTaskEntity("TM-task-3", track.ID, "Task 3", "", "todo", 300, "", now, now)

	mockTaskRepo.GetBacklogTasksFunc = func(ctx context.Context) ([]*entities.TaskEntity, error) {
		return []*entities.TaskEntity{task1, task3}, nil
	}

	// Get backlog tasks (should return tasks not in any iteration and not done)
	results, err := service.GetBacklogTasks(ctx)
	if err != nil {
		t.Fatalf("GetBacklogTasks() failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("GetBacklogTasks() returned %d tasks, want 2", len(results))
	}
}

// TestTaskService_GetBacklogTasks_Empty tests backlog retrieval with no tasks
func TestTaskService_GetBacklogTasks_Empty(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, _ := setupTaskTestService(t)

	mockTaskRepo.GetBacklogTasksFunc = func(ctx context.Context) ([]*entities.TaskEntity, error) {
		return []*entities.TaskEntity{}, nil
	}

	// Get backlog tasks from empty database
	results, err := service.GetBacklogTasks(ctx)
	if err != nil {
		t.Fatalf("GetBacklogTasks() failed: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("GetBacklogTasks() returned %d tasks, want 0", len(results))
	}
}

// ============================================================================
// AC Verification Enforcement Tests (Phase 3 - Iteration 36)
// ============================================================================

// TestTaskService_UpdateTask_CannotCompleteTodo_WithPendingACs tests that tasks cannot be marked done with pending ACs
func TestTaskService_UpdateTask_CannotCompleteTodo_WithPendingACs(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "in-progress", 100, "", now, now)

	// Mock task retrieval
	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with pending ACs
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == task.ID {
			return []*entities.AcceptanceCriteriaEntity{
				entities.NewAcceptanceCriteriaEntity("TM-ac-1", task.ID, "AC 1", entities.VerificationTypeManual, "", now, now),
				entities.NewAcceptanceCriteriaEntity("TM-ac-2", task.ID, "AC 2", entities.VerificationTypeManual, "", now, now),
			}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// Try to update task status to "done"
	doneStatus := "done"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &doneStatus,
	}

	_, err := service.UpdateTask(ctx, input)
	if err == nil {
		t.Fatal("UpdateTask() should fail when marking task done with pending ACs")
	}

	// Verify error message contains guidance
	errMsg := err.Error()
	if !contains(errMsg, "unverified acceptance criteria") {
		t.Errorf("error message should mention unverified ACs, got: %s", errMsg)
	}
	if !contains(errMsg, "TM-ac-1") && !contains(errMsg, "TM-ac-2") {
		t.Errorf("error message should list unverified AC IDs, got: %s", errMsg)
	}
}

// TestTaskService_UpdateTask_CannotCompleteTodo_WithFailedACs tests that tasks cannot be marked done with failed ACs
func TestTaskService_UpdateTask_CannotCompleteTodo_WithFailedACs(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "in-progress", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with failed ACs
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == task.ID {
			ac := entities.NewAcceptanceCriteriaEntity("TM-ac-1", task.ID, "AC 1", entities.VerificationTypeManual, "", now, now)
			ac.Status = entities.ACStatusFailed
			return []*entities.AcceptanceCriteriaEntity{ac}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// Try to update task status to "done"
	doneStatus := "done"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &doneStatus,
	}

	_, err := service.UpdateTask(ctx, input)
	if err == nil {
		t.Fatal("UpdateTask() should fail when marking task done with failed ACs")
	}

	// Verify error message
	errMsg := err.Error()
	if !contains(errMsg, "unverified acceptance criteria") {
		t.Errorf("error message should mention unverified ACs, got: %s", errMsg)
	}
}

// TestTaskService_UpdateTask_CanCompleteTodo_WithAllVerifiedACs tests successful completion with all verified ACs
func TestTaskService_UpdateTask_CanCompleteTodo_WithAllVerifiedACs(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "in-progress", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with all verified ACs
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == task.ID {
			ac1 := entities.NewAcceptanceCriteriaEntity("TM-ac-1", task.ID, "AC 1", entities.VerificationTypeManual, "", now, now)
			ac1.Status = entities.ACStatusVerified
			ac2 := entities.NewAcceptanceCriteriaEntity("TM-ac-2", task.ID, "AC 2", entities.VerificationTypeManual, "", now, now)
			ac2.Status = entities.ACStatusVerified
			return []*entities.AcceptanceCriteriaEntity{ac1, ac2}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update task status to "done"
	doneStatus := "done"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &doneStatus,
	}

	updatedTask, err := service.UpdateTask(ctx, input)
	if err != nil {
		t.Fatalf("UpdateTask() should succeed with all verified ACs, got error: %v", err)
	}

	if updatedTask.Status != "done" {
		t.Errorf("task.Status = %q, want %q", updatedTask.Status, "done")
	}
}

// TestTaskService_UpdateTask_CanCompleteTodo_WithAllSkippedACs tests successful completion with all skipped ACs
func TestTaskService_UpdateTask_CanCompleteTodo_WithAllSkippedACs(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "in-progress", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with all skipped ACs
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == task.ID {
			ac1 := entities.NewAcceptanceCriteriaEntity("TM-ac-1", task.ID, "AC 1", entities.VerificationTypeManual, "", now, now)
			ac1.Status = entities.ACStatusSkipped
			ac2 := entities.NewAcceptanceCriteriaEntity("TM-ac-2", task.ID, "AC 2", entities.VerificationTypeManual, "", now, now)
			ac2.Status = entities.ACStatusSkipped
			return []*entities.AcceptanceCriteriaEntity{ac1, ac2}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update task status to "done"
	doneStatus := "done"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &doneStatus,
	}

	updatedTask, err := service.UpdateTask(ctx, input)
	if err != nil {
		t.Fatalf("UpdateTask() should succeed with all skipped ACs, got error: %v", err)
	}

	if updatedTask.Status != "done" {
		t.Errorf("task.Status = %q, want %q", updatedTask.Status, "done")
	}
}

// TestTaskService_UpdateTask_CanCompleteTodo_WithMixedVerifiedAndSkippedACs tests completion with mixed verified/skipped ACs
func TestTaskService_UpdateTask_CanCompleteTodo_WithMixedVerifiedAndSkippedACs(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "in-progress", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with mixed verified and skipped ACs
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == task.ID {
			ac1 := entities.NewAcceptanceCriteriaEntity("TM-ac-1", task.ID, "AC 1", entities.VerificationTypeManual, "", now, now)
			ac1.Status = entities.ACStatusVerified
			ac2 := entities.NewAcceptanceCriteriaEntity("TM-ac-2", task.ID, "AC 2", entities.VerificationTypeManual, "", now, now)
			ac2.Status = entities.ACStatusSkipped
			ac3 := entities.NewAcceptanceCriteriaEntity("TM-ac-3", task.ID, "AC 3", entities.VerificationTypeManual, "", now, now)
			ac3.Status = entities.ACStatusAutomaticallyVerified
			return []*entities.AcceptanceCriteriaEntity{ac1, ac2, ac3}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update task status to "done"
	doneStatus := "done"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &doneStatus,
	}

	updatedTask, err := service.UpdateTask(ctx, input)
	if err != nil {
		t.Fatalf("UpdateTask() should succeed with mixed verified/skipped ACs, got error: %v", err)
	}

	if updatedTask.Status != "done" {
		t.Errorf("task.Status = %q, want %q", updatedTask.Status, "done")
	}
}

// TestTaskService_UpdateTask_CanCompleteTodo_WithNoACs tests completion when task has no ACs
func TestTaskService_UpdateTask_CanCompleteTodo_WithNoACs(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "in-progress", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with no ACs
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update task status to "done"
	doneStatus := "done"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &doneStatus,
	}

	updatedTask, err := service.UpdateTask(ctx, input)
	if err != nil {
		t.Fatalf("UpdateTask() should succeed with no ACs, got error: %v", err)
	}

	if updatedTask.Status != "done" {
		t.Errorf("task.Status = %q, want %q", updatedTask.Status, "done")
	}
}

// TestTaskService_UpdateTask_CannotCompleteTodo_WithMixedStatuses tests failure with mixed AC statuses including unverified
func TestTaskService_UpdateTask_CannotCompleteTodo_WithMixedStatuses(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "in-progress", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with mixed statuses (verified + pending + failed)
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == task.ID {
			ac1 := entities.NewAcceptanceCriteriaEntity("TM-ac-1", task.ID, "AC 1", entities.VerificationTypeManual, "", now, now)
			ac1.Status = entities.ACStatusVerified // OK
			ac2 := entities.NewAcceptanceCriteriaEntity("TM-ac-2", task.ID, "AC 2", entities.VerificationTypeManual, "", now, now)
			ac2.Status = entities.ACStatusNotStarted // BLOCKS
			ac3 := entities.NewAcceptanceCriteriaEntity("TM-ac-3", task.ID, "AC 3", entities.VerificationTypeManual, "", now, now)
			ac3.Status = entities.ACStatusFailed // BLOCKS
			return []*entities.AcceptanceCriteriaEntity{ac1, ac2, ac3}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	// Try to update task status to "done"
	doneStatus := "done"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &doneStatus,
	}

	_, err := service.UpdateTask(ctx, input)
	if err == nil {
		t.Fatal("UpdateTask() should fail when marking task done with mixed AC statuses including unverified")
	}

	// Verify error message lists both blocking ACs
	errMsg := err.Error()
	if !contains(errMsg, "TM-ac-2") || !contains(errMsg, "TM-ac-3") {
		t.Errorf("error message should list both unverified AC IDs (TM-ac-2, TM-ac-3), got: %s", errMsg)
	}
}

// TestTaskService_UpdateTask_AllowsNonDoneTransition_WithPendingACs tests that non-done transitions are allowed even with pending ACs
func TestTaskService_UpdateTask_AllowsNonDoneTransition_WithPendingACs(t *testing.T) {
	service, ctx, mockTaskRepo, _, _, mockACRepo := setupTaskTestService(t)

	now := time.Now().UTC()
	task, _ := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Test Task", "Description", "todo", 100, "", now, now)

	mockTaskRepo.GetTaskFunc = func(ctx context.Context, id string) (*entities.TaskEntity, error) {
		if id == task.ID {
			return task, nil
		}
		return nil, pluginsdk.ErrNotFound
	}

	// Mock AC list with pending ACs
	mockACRepo.ListACFunc = func(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
		if taskID == task.ID {
			return []*entities.AcceptanceCriteriaEntity{
				entities.NewAcceptanceCriteriaEntity("TM-ac-1", task.ID, "AC 1", entities.VerificationTypeManual, "", now, now),
			}, nil
		}
		return []*entities.AcceptanceCriteriaEntity{}, nil
	}

	mockTaskRepo.UpdateTaskFunc = func(ctx context.Context, task *entities.TaskEntity) error {
		return nil
	}

	// Update task status to "in-progress" (not "done")
	inProgressStatus := "in-progress"
	input := dto.UpdateTaskDTO{
		ID:     task.ID,
		Status: &inProgressStatus,
	}

	updatedTask, err := service.UpdateTask(ctx, input)
	if err != nil {
		t.Fatalf("UpdateTask() should allow non-done transitions even with pending ACs, got error: %v", err)
	}

	if updatedTask.Status != "in-progress" {
		t.Errorf("task.Status = %q, want %q", updatedTask.Status, "in-progress")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

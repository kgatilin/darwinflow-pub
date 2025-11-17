package transformers_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/transformers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformToTrackDetailViewModel(t *testing.T) {
	now := time.Now()

	// Create dependency tracks
	dep1, err := entities.NewTrackEntity("TM-track-2", "roadmap-1", "Database Setup", "Setup PostgreSQL", "complete", 50, []string{}, now, now)
	require.NoError(t, err)
	dep2, err := entities.NewTrackEntity("TM-track-3", "roadmap-1", "API Gateway", "Setup API gateway", "in-progress", 75, []string{}, now, now)
	require.NoError(t, err)

	// Create track with dependencies
	track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Authentication System", "Implement user authentication", "in-progress", 100, []string{"TM-track-2", "TM-track-3"}, now, now)
	require.NoError(t, err)

	// Create tasks
	task1, err := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Implement login", "Create login endpoint", "todo", 100, "", now, now)
	require.NoError(t, err)
	task2, err := entities.NewTaskEntity("TM-task-2", "TM-track-1", "Implement signup", "Create signup endpoint", "in-progress", 200, "", now, now)
	require.NoError(t, err)
	task3, err := entities.NewTaskEntity("TM-task-3", "TM-track-1", "Setup database", "Create user table", "done", 300, "", now, now)
	require.NoError(t, err)

	tasks := []*entities.TaskEntity{task1, task2, task3}
	dependencyTracks := []*entities.TrackEntity{dep1, dep2}

	// Transform
	vm := transformers.TransformToTrackDetailViewModel(track, tasks, dependencyTracks)

	// Verify track fields
	assert.Equal(t, "TM-track-1", vm.ID)
	assert.Equal(t, "Authentication System", vm.Title)
	assert.Equal(t, "Implement user authentication", vm.Description)
	assert.Equal(t, "in-progress", vm.Status)
	assert.Equal(t, "In Progress", vm.StatusLabel)
	assert.Equal(t, 100, vm.Rank)
	assert.Equal(t, []string{"TM-track-2", "TM-track-3"}, vm.Dependencies)
	assert.Equal(t, []string{"Database Setup", "API Gateway"}, vm.DependencyLabels)

	// Verify task grouping
	assert.Len(t, vm.TODOTasks, 1)
	assert.Len(t, vm.InProgressTasks, 1)
	assert.Len(t, vm.DoneTasks, 1)

	// Verify TODO tasks
	assert.Equal(t, "TM-task-1", vm.TODOTasks[0].ID)
	assert.Equal(t, "Implement login", vm.TODOTasks[0].Title)
	assert.Equal(t, "todo", vm.TODOTasks[0].Status)

	// Verify In Progress tasks
	assert.Equal(t, "TM-task-2", vm.InProgressTasks[0].ID)
	assert.Equal(t, "Implement signup", vm.InProgressTasks[0].Title)
	assert.Equal(t, "in-progress", vm.InProgressTasks[0].Status)

	// Verify Done tasks
	assert.Equal(t, "TM-task-3", vm.DoneTasks[0].ID)
	assert.Equal(t, "Setup database", vm.DoneTasks[0].Title)
	assert.Equal(t, "done", vm.DoneTasks[0].Status)

	// Verify progress
	assert.Equal(t, 1, vm.Progress.Completed)
	assert.Equal(t, 3, vm.Progress.Total)
	assert.InDelta(t, 0.333, vm.Progress.Percent, 0.01)
}

func TestTransformToTrackDetailViewModel_NoDependencies(t *testing.T) {
	now := time.Now()

	// Create track without dependencies
	track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Authentication System", "Implement user authentication", "not-started", 100, []string{}, now, now)
	require.NoError(t, err)

	tasks := []*entities.TaskEntity{}
	dependencyTracks := []*entities.TrackEntity{}

	// Transform
	vm := transformers.TransformToTrackDetailViewModel(track, tasks, dependencyTracks)

	// Verify no dependencies
	assert.Empty(t, vm.Dependencies)
	assert.Empty(t, vm.DependencyLabels)
	assert.Equal(t, "Not Started", vm.StatusLabel)
}

func TestTransformToTrackDetailViewModel_NoTasks(t *testing.T) {
	now := time.Now()

	// Create track with no tasks
	track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Authentication System", "Implement user authentication", "not-started", 100, []string{}, now, now)
	require.NoError(t, err)

	tasks := []*entities.TaskEntity{}
	dependencyTracks := []*entities.TrackEntity{}

	// Transform
	vm := transformers.TransformToTrackDetailViewModel(track, tasks, dependencyTracks)

	// Verify empty task lists
	assert.Empty(t, vm.TODOTasks)
	assert.Empty(t, vm.InProgressTasks)
	assert.Empty(t, vm.DoneTasks)

	// Verify zero progress
	assert.Equal(t, 0, vm.Progress.Completed)
	assert.Equal(t, 0, vm.Progress.Total)
	assert.Equal(t, 0.0, vm.Progress.Percent)
}

func TestTransformToTrackDetailViewModel_ReviewTasksGroupedWithInProgress(t *testing.T) {
	now := time.Now()

	track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Track 1", "Description", "in-progress", 100, []string{}, now, now)
	require.NoError(t, err)

	// Create task with "review" status
	task1, err := entities.NewTaskEntity("TM-task-1", "TM-track-1", "Task 1", "Description", "review", 100, "", now, now)
	require.NoError(t, err)

	tasks := []*entities.TaskEntity{task1}
	dependencyTracks := []*entities.TrackEntity{}

	// Transform
	vm := transformers.TransformToTrackDetailViewModel(track, tasks, dependencyTracks)

	// Verify review tasks are grouped with in-progress
	assert.Len(t, vm.InProgressTasks, 1)
	assert.Equal(t, "TM-task-1", vm.InProgressTasks[0].ID)
	assert.Equal(t, "review", vm.InProgressTasks[0].Status)
}

func TestTransformToTrackDetailViewModel_AllStatuses(t *testing.T) {
	now := time.Now()

	tests := []struct {
		status       string
		expectedLabel string
	}{
		{"not-started", "Not Started"},
		{"in-progress", "In Progress"},
		{"complete", "Complete"},
		{"blocked", "Blocked"},
		{"waiting", "Waiting"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Track 1", "Description", tt.status, 100, []string{}, now, now)
			require.NoError(t, err)

			vm := transformers.TransformToTrackDetailViewModel(track, []*entities.TaskEntity{}, []*entities.TrackEntity{})

			assert.Equal(t, tt.status, vm.Status)
			assert.Equal(t, tt.expectedLabel, vm.StatusLabel)
		})
	}
}

func TestTransformToTrackDetailViewModel_DependencyTrackNotFound(t *testing.T) {
	now := time.Now()

	// Create track with dependency that doesn't exist in dependencyTracks
	track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Track 1", "Description", "not-started", 100, []string{"TM-track-99"}, now, now)
	require.NoError(t, err)

	dependencyTracks := []*entities.TrackEntity{} // Empty - dependency not found

	// Transform
	vm := transformers.TransformToTrackDetailViewModel(track, []*entities.TaskEntity{}, dependencyTracks)

	// Verify fallback to ID when dependency not found
	assert.Equal(t, []string{"TM-track-99"}, vm.Dependencies)
	assert.Equal(t, []string{"TM-track-99"}, vm.DependencyLabels) // Falls back to ID
}

func TestFormatTrackStatus_UnknownStatus(t *testing.T) {
	// Test the default case directly (unreachable via entity validation)
	result := transformers.FormatTrackStatus("unknown-status")
	assert.Equal(t, "Unknown-status", result)
}

func TestFormatTrackStatus_EmptyStatus(t *testing.T) {
	// Test empty status (defensive programming)
	result := transformers.FormatTrackStatus("")
	assert.Equal(t, "", result)
}

func TestTransformToTrackDetailViewModel_ProgressCalculation(t *testing.T) {
	now := time.Now()

	track, err := entities.NewTrackEntity("TM-track-1", "roadmap-1", "Track 1", "Description", "in-progress", 100, []string{}, now, now)
	require.NoError(t, err)

	// Create 10 tasks: 3 done, 7 not done
	tasks := make([]*entities.TaskEntity, 10)
	for i := 0; i < 10; i++ {
		status := "todo"
		if i < 3 {
			status = "done"
		}
		task, err := entities.NewTaskEntity(
			"TM-task-"+string(rune('0'+i)),
			"TM-track-1",
			"Task",
			"Description",
			status,
			100+i,
			"",
			now,
			now,
		)
		require.NoError(t, err)
		tasks[i] = task
	}

	// Transform
	vm := transformers.TransformToTrackDetailViewModel(track, tasks, []*entities.TrackEntity{})

	// Verify progress: 3/10 = 0.3 (30%)
	assert.Equal(t, 3, vm.Progress.Completed)
	assert.Equal(t, 10, vm.Progress.Total)
	assert.Equal(t, 0.3, vm.Progress.Percent)
}

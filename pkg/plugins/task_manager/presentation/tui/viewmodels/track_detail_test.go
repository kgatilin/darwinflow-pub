package viewmodels_test

import (
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/viewmodels"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackDetailViewModel(t *testing.T) {
	vm := viewmodels.NewTrackDetailViewModel(
		"TM-track-1",
		"Authentication System",
		"Implement user authentication",
		"in-progress",
		"In Progress",
		100,
		[]string{"TM-track-2"},
		[]string{"Database Setup"},
	)

	assert.Equal(t, "TM-track-1", vm.ID)
	assert.Equal(t, "Authentication System", vm.Title)
	assert.Equal(t, "Implement user authentication", vm.Description)
	assert.Equal(t, "in-progress", vm.Status)
	assert.Equal(t, "In Progress", vm.StatusLabel)
	assert.Equal(t, 100, vm.Rank)
	assert.Equal(t, []string{"TM-track-2"}, vm.Dependencies)
	assert.Equal(t, []string{"Database Setup"}, vm.DependencyLabels)

	// Verify collections are initialized
	assert.NotNil(t, vm.TODOTasks)
	assert.NotNil(t, vm.InProgressTasks)
	assert.NotNil(t, vm.DoneTasks)
	assert.NotNil(t, vm.Progress)

	// Verify empty collections
	assert.Empty(t, vm.TODOTasks)
	assert.Empty(t, vm.InProgressTasks)
	assert.Empty(t, vm.DoneTasks)
}

func TestNewTrackDetailViewModel_WithNoDependencies(t *testing.T) {
	vm := viewmodels.NewTrackDetailViewModel(
		"TM-track-1",
		"Authentication System",
		"Implement user authentication",
		"not-started",
		"Not Started",
		100,
		[]string{},
		[]string{},
	)

	assert.Equal(t, "TM-track-1", vm.ID)
	assert.Empty(t, vm.Dependencies)
	assert.Empty(t, vm.DependencyLabels)
}

func TestTrackDetailTaskViewModel_Fields(t *testing.T) {
	task := &viewmodels.TrackDetailTaskViewModel{
		ID:          "TM-task-1",
		Title:       "Implement login",
		Status:      "todo",
		Description: "Create login endpoint",
	}

	assert.Equal(t, "TM-task-1", task.ID)
	assert.Equal(t, "Implement login", task.Title)
	assert.Equal(t, "todo", task.Status)
	assert.Equal(t, "Create login endpoint", task.Description)
}

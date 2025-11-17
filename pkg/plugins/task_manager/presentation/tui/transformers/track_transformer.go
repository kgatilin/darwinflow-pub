package transformers

import (
	"strings"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/viewmodels"
)

// TransformToTrackDetailViewModel transforms track + tasks + dependency tracks to track detail view model
func TransformToTrackDetailViewModel(
	track *entities.TrackEntity,
	tasks []*entities.TaskEntity,
	dependencyTracks []*entities.TrackEntity,
) *viewmodels.TrackDetailViewModel {
	// Build dependency labels map
	dependencyLabels := make([]string, 0, len(track.Dependencies))
	depMap := make(map[string]string)
	for _, depTrack := range dependencyTracks {
		depMap[depTrack.ID] = depTrack.Title
	}
	// Preserve order from track.Dependencies
	for _, depID := range track.Dependencies {
		if title, exists := depMap[depID]; exists {
			dependencyLabels = append(dependencyLabels, title)
		} else {
			dependencyLabels = append(dependencyLabels, depID) // Fallback to ID if not found
		}
	}

	// Create view model with pre-computed display data
	vm := viewmodels.NewTrackDetailViewModel(
		track.ID,
		track.Title,
		track.Description,
		track.Status,
		FormatTrackStatus(track.Status),
		track.Rank,
		track.Dependencies,
		dependencyLabels,
	)

	// Group tasks by status and create task map
	taskMap := make(map[string]*viewmodels.TrackDetailTaskViewModel)
	for _, task := range tasks {
		taskRow := &viewmodels.TrackDetailTaskViewModel{
			ID:          task.ID,
			Title:       task.Title,
			Status:      task.Status,
			Description: task.Description,
		}

		// Store in map
		taskMap[task.ID] = taskRow

		switch task.Status {
		case string(entities.TaskStatusTodo):
			vm.TODOTasks = append(vm.TODOTasks, taskRow)
		case string(entities.TaskStatusInProgress), string(entities.TaskStatusReview):
			vm.InProgressTasks = append(vm.InProgressTasks, taskRow)
		case string(entities.TaskStatusDone):
			vm.DoneTasks = append(vm.DoneTasks, taskRow)
		}
	}

	// Calculate progress (done tasks / total tasks)
	totalTasks := len(tasks)
	doneTasks := len(vm.DoneTasks)
	vm.Progress = viewmodels.NewProgressViewModel(doneTasks, totalTasks)

	return vm
}

// FormatTrackStatus converts raw status to display-friendly label
func FormatTrackStatus(status string) string {
	switch status {
	case string(entities.TrackStatusNotStarted):
		return "Not Started"
	case string(entities.TrackStatusInProgress):
		return "In Progress"
	case string(entities.TrackStatusComplete):
		return "Complete"
	case string(entities.TrackStatusBlocked):
		return "Blocked"
	case string(entities.TrackStatusWaiting):
		return "Waiting"
	default:
		// Fallback for unknown status (defensive programming - unreachable due to entity validation)
		if len(status) == 0 {
			return ""
		}
		return strings.ToUpper(string(status[0])) + strings.ToLower(status[1:])
	}
}

package queries

import (
	"context"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/transformers"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/viewmodels"
)

// LoadTrackDetailData loads track detail data for a specific track.
// Returns track + tasks + dependency tracks transformed into view model ready for presentation.
//
// Pre-loads:
// - Track entity
// - All tasks in the track
// - All dependency tracks (for display labels)
//
// Eliminates N+1 queries by loading all related data upfront.
func LoadTrackDetailData(
	ctx context.Context,
	repo domain.RoadmapRepository,
	trackID string,
) (*viewmodels.TrackDetailViewModel, error) {
	// Fetch track
	track, err := repo.GetTrack(ctx, trackID)
	if err != nil {
		return nil, err
	}

	// Fetch all tasks in the track
	tasks, err := repo.ListTasks(ctx, entities.TaskFilters{TrackID: trackID})
	if err != nil {
		return nil, err
	}

	// Fetch dependency tracks for display labels
	dependencyTracks := make([]*entities.TrackEntity, 0, len(track.Dependencies))
	for _, depID := range track.Dependencies {
		depTrack, err := repo.GetTrack(ctx, depID)
		if err != nil {
			// If dependency track not found, skip it (transformer will fallback to ID)
			continue
		}
		dependencyTracks = append(dependencyTracks, depTrack)
	}

	// Transform to view model
	vm := transformers.TransformToTrackDetailViewModel(track, tasks, dependencyTracks)

	return vm, nil
}

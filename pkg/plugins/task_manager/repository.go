package task_manager

import "context"

// TrackFilters provides optional filtering for track queries
type TrackFilters struct {
	Status   []string // Filter by status (not-started, in-progress, complete, blocked, waiting)
	Priority []string // Filter by priority (critical, high, medium, low)
}

// TaskFilters provides optional filtering for task queries
type TaskFilters struct {
	TrackID  string   // Filter by track ID
	Status   []string // Filter by status (todo, in-progress, done)
	Priority []string // Filter by priority (critical, high, medium, low)
}

// RoadmapRepository defines the contract for persistent storage of roadmap entities.
// It manages roadmaps, tracks, tasks, and iterations in a hierarchical structure.
type RoadmapRepository interface {
	// Roadmap operations

	// SaveRoadmap persists a new roadmap to storage.
	// Returns ErrAlreadyExists if a roadmap with the same ID already exists.
	SaveRoadmap(ctx context.Context, roadmap *RoadmapEntity) error

	// GetRoadmap retrieves a roadmap by its ID.
	// Returns ErrNotFound if the roadmap doesn't exist.
	GetRoadmap(ctx context.Context, id string) (*RoadmapEntity, error)

	// GetActiveRoadmap retrieves the most recently created roadmap.
	// Returns ErrNotFound if no roadmaps exist.
	GetActiveRoadmap(ctx context.Context) (*RoadmapEntity, error)

	// UpdateRoadmap updates an existing roadmap.
	// Returns ErrNotFound if the roadmap doesn't exist.
	UpdateRoadmap(ctx context.Context, roadmap *RoadmapEntity) error

	// Track operations

	// SaveTrack persists a new track to storage.
	// Returns ErrAlreadyExists if a track with the same ID already exists.
	SaveTrack(ctx context.Context, track *TrackEntity) error

	// GetTrack retrieves a track by its ID.
	// Returns ErrNotFound if the track doesn't exist.
	GetTrack(ctx context.Context, id string) (*TrackEntity, error)

	// ListTracks returns all tracks for a roadmap, optionally filtered.
	// Returns empty slice if no tracks match the filters.
	ListTracks(ctx context.Context, roadmapID string, filters TrackFilters) ([]*TrackEntity, error)

	// UpdateTrack updates an existing track.
	// Returns ErrNotFound if the track doesn't exist.
	UpdateTrack(ctx context.Context, track *TrackEntity) error

	// DeleteTrack removes a track from storage.
	// Returns ErrNotFound if the track doesn't exist.
	DeleteTrack(ctx context.Context, id string) error

	// AddTrackDependency adds a dependency from trackID to dependsOnID.
	// Returns ErrNotFound if either track doesn't exist.
	// Returns ErrInvalidArgument if it would create a self-dependency.
	// Returns ErrAlreadyExists if the dependency already exists.
	AddTrackDependency(ctx context.Context, trackID, dependsOnID string) error

	// RemoveTrackDependency removes a dependency from trackID to dependsOnID.
	// Returns ErrNotFound if the dependency doesn't exist.
	RemoveTrackDependency(ctx context.Context, trackID, dependsOnID string) error

	// GetTrackDependencies returns the IDs of all tracks that trackID depends on.
	// Returns empty slice if there are no dependencies.
	GetTrackDependencies(ctx context.Context, trackID string) ([]string, error)

	// ValidateNoCycles checks if adding/updating the track would create a circular dependency.
	// Returns ErrInvalidArgument if a cycle is detected.
	ValidateNoCycles(ctx context.Context, trackID string) error

	// Task operations

	// SaveTask persists a new task to storage.
	// Returns ErrAlreadyExists if a task with the same ID already exists.
	// Returns ErrNotFound if the track doesn't exist.
	SaveTask(ctx context.Context, task *TaskEntity) error

	// GetTask retrieves a task by its ID.
	// Returns ErrNotFound if the task doesn't exist.
	GetTask(ctx context.Context, id string) (*TaskEntity, error)

	// ListTasks returns all tasks matching the filters.
	// Returns empty slice if no tasks match the filters.
	ListTasks(ctx context.Context, filters TaskFilters) ([]*TaskEntity, error)

	// UpdateTask updates an existing task.
	// Returns ErrNotFound if the task doesn't exist.
	UpdateTask(ctx context.Context, task *TaskEntity) error

	// DeleteTask removes a task from storage.
	// Returns ErrNotFound if the task doesn't exist.
	DeleteTask(ctx context.Context, id string) error

	// MoveTaskToTrack moves a task from its current track to a new track.
	// Returns ErrNotFound if the task or new track doesn't exist.
	MoveTaskToTrack(ctx context.Context, taskID, newTrackID string) error

	// Iteration operations

	// SaveIteration persists a new iteration to storage.
	// Returns ErrAlreadyExists if an iteration with the same number already exists.
	SaveIteration(ctx context.Context, iteration *IterationEntity) error

	// GetIteration retrieves an iteration by its number.
	// Returns ErrNotFound if the iteration doesn't exist.
	GetIteration(ctx context.Context, number int) (*IterationEntity, error)

	// GetCurrentIteration returns the iteration with status "current".
	// Returns ErrNotFound if no current iteration exists.
	GetCurrentIteration(ctx context.Context) (*IterationEntity, error)

	// ListIterations returns all iterations, ordered by number.
	// Returns empty slice if no iterations exist.
	ListIterations(ctx context.Context) ([]*IterationEntity, error)

	// UpdateIteration updates an existing iteration.
	// Returns ErrNotFound if the iteration doesn't exist.
	UpdateIteration(ctx context.Context, iteration *IterationEntity) error

	// DeleteIteration removes an iteration from storage.
	// Returns ErrNotFound if the iteration doesn't exist.
	DeleteIteration(ctx context.Context, number int) error

	// AddTaskToIteration adds a task to an iteration.
	// Returns ErrNotFound if the iteration or task doesn't exist.
	// Returns ErrAlreadyExists if the task is already in the iteration.
	AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error

	// RemoveTaskFromIteration removes a task from an iteration.
	// Returns ErrNotFound if the iteration doesn't exist or the task is not in it.
	RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error

	// GetIterationTasks returns all tasks in an iteration.
	// Returns empty slice if the iteration has no tasks.
	GetIterationTasks(ctx context.Context, iterationNum int) ([]*TaskEntity, error)

	// StartIteration marks an iteration as current and sets started_at timestamp.
	// Returns ErrNotFound if the iteration doesn't exist.
	// Returns ErrInvalidArgument if the iteration status is not "planned".
	StartIteration(ctx context.Context, iterationNum int) error

	// CompleteIteration marks an iteration as complete and sets completed_at timestamp.
	// Returns ErrNotFound if the iteration doesn't exist.
	// Returns ErrInvalidArgument if the iteration status is not "current".
	CompleteIteration(ctx context.Context, iterationNum int) error

	// Aggregate queries

	// GetRoadmapWithTracks retrieves a roadmap with all its tracks.
	// The roadmap is returned with Dependencies populated from the database.
	// Returns ErrNotFound if the roadmap doesn't exist.
	GetRoadmapWithTracks(ctx context.Context, roadmapID string) (*RoadmapEntity, error)

	// GetTrackWithTasks retrieves a track with all its tasks.
	// The track is returned with Dependencies populated from the database.
	// Returns ErrNotFound if the track doesn't exist.
	GetTrackWithTasks(ctx context.Context, trackID string) (*TrackEntity, error)

	// Project metadata operations

	// GetProjectMetadata retrieves a metadata value by key.
	// Returns ErrNotFound if the key doesn't exist.
	GetProjectMetadata(ctx context.Context, key string) (string, error)

	// SetProjectMetadata sets a metadata value by key.
	// Creates or updates the key-value pair.
	SetProjectMetadata(ctx context.Context, key, value string) error

	// GetProjectCode retrieves the project code (e.g., "DW" for darwinflow).
	// Returns "DW" as default if not set.
	GetProjectCode(ctx context.Context) string

	// ADR operations

	// SaveADR persists a new ADR to storage.
	// Returns ErrAlreadyExists if an ADR with the same ID already exists.
	// Returns ErrNotFound if the track doesn't exist.
	SaveADR(ctx context.Context, adr *ADREntity) error

	// GetADR retrieves an ADR by its ID.
	// Returns ErrNotFound if the ADR doesn't exist.
	GetADR(ctx context.Context, id string) (*ADREntity, error)

	// ListADRs returns all ADRs, optionally filtered by track.
	// Returns empty slice if no ADRs match the filters.
	ListADRs(ctx context.Context, trackID *string) ([]*ADREntity, error)

	// UpdateADR updates an existing ADR.
	// Returns ErrNotFound if the ADR doesn't exist.
	UpdateADR(ctx context.Context, adr *ADREntity) error

	// SupersedeADR marks an ADR as superseded by another ADR.
	// Returns ErrNotFound if either ADR doesn't exist.
	SupersedeADR(ctx context.Context, adrID, supersededByID string) error

	// DeprecateADR marks an ADR as deprecated.
	// Returns ErrNotFound if the ADR doesn't exist.
	DeprecateADR(ctx context.Context, adrID string) error

	// GetADRsByTrack returns all ADRs for a specific track.
	// Returns empty slice if the track has no ADRs.
	GetADRsByTrack(ctx context.Context, trackID string) ([]*ADREntity, error)

	// GetNextSequenceNumber retrieves the next sequence number for an entity type.
	// Entity types: "task", "track", "iter"
	GetNextSequenceNumber(ctx context.Context, entityType string) (int, error)

	// Acceptance Criteria operations

	// SaveAC persists a new acceptance criterion to storage.
	// Returns ErrAlreadyExists if an AC with the same ID already exists.
	// Returns ErrNotFound if the task doesn't exist.
	SaveAC(ctx context.Context, ac *AcceptanceCriteriaEntity) error

	// GetAC retrieves an acceptance criterion by its ID.
	// Returns ErrNotFound if the AC doesn't exist.
	GetAC(ctx context.Context, id string) (*AcceptanceCriteriaEntity, error)

	// ListAC returns all acceptance criteria for a task.
	// Returns empty slice if the task has no ACs.
	ListAC(ctx context.Context, taskID string) ([]*AcceptanceCriteriaEntity, error)

	// UpdateAC updates an existing acceptance criterion.
	// Returns ErrNotFound if the AC doesn't exist.
	UpdateAC(ctx context.Context, ac *AcceptanceCriteriaEntity) error

	// DeleteAC removes an acceptance criterion from storage.
	// Returns ErrNotFound if the AC doesn't exist.
	DeleteAC(ctx context.Context, id string) error

	// ListACByTrack returns all acceptance criteria for all tasks in a track.
	// Returns empty slice if the track has no ACs.
	ListACByTrack(ctx context.Context, trackID string) ([]*AcceptanceCriteriaEntity, error)

	// ListACByIteration returns all acceptance criteria for all tasks in an iteration.
	// Returns empty slice if the iteration has no ACs.
	ListACByIteration(ctx context.Context, iterationNum int) ([]*AcceptanceCriteriaEntity, error)
}

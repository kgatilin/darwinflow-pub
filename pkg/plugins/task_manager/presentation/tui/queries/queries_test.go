package queries_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/queries"
)

// MockRepository is a mock implementation for testing queries.
type MockRepository struct {
	iterations          []*entities.IterationEntity
	activeRoadmap       *entities.RoadmapEntity
	tracks              []*entities.TrackEntity
	backlogTasks        []*entities.TaskEntity
	iteration           *entities.IterationEntity
	iterationTasks      []*entities.TaskEntity
	acsByIteration      []*entities.AcceptanceCriteriaEntity
	task                *entities.TaskEntity
	acsByTask           []*entities.AcceptanceCriteriaEntity
	track               *entities.TrackEntity
	iterationsForTask   []*entities.IterationEntity
	tasksForTrack       []*entities.TaskEntity
	dependencyTracks    map[string]*entities.TrackEntity
	listTracksErr       error
	listIterationsErr   error
	getActiveRoadmapErr error
	getBacklogTasksErr  error
	getIterationErr     error
	getIterationTasksErr error
	listACByIterationErr error
	getTaskErr          error
	listACErr           error
	getTrackErr         error
	getIterationsForTaskErr error
	listTasksErr        error
}

// ListIterations returns all iterations.
func (m *MockRepository) ListIterations(ctx context.Context) ([]*entities.IterationEntity, error) {
	if m.listIterationsErr != nil {
		return nil, m.listIterationsErr
	}
	return m.iterations, nil
}

// GetActiveRoadmap returns the active roadmap.
func (m *MockRepository) GetActiveRoadmap(ctx context.Context) (*entities.RoadmapEntity, error) {
	if m.getActiveRoadmapErr != nil {
		return nil, m.getActiveRoadmapErr
	}
	return m.activeRoadmap, nil
}

// ListTracks returns all tracks.
func (m *MockRepository) ListTracks(ctx context.Context, roadmapID string, filters entities.TrackFilters) ([]*entities.TrackEntity, error) {
	if m.listTracksErr != nil {
		return nil, m.listTracksErr
	}
	return m.tracks, nil
}

// GetBacklogTasks returns backlog tasks.
func (m *MockRepository) GetBacklogTasks(ctx context.Context) ([]*entities.TaskEntity, error) {
	if m.getBacklogTasksErr != nil {
		return nil, m.getBacklogTasksErr
	}
	return m.backlogTasks, nil
}

// GetIteration returns an iteration.
func (m *MockRepository) GetIteration(ctx context.Context, number int) (*entities.IterationEntity, error) {
	if m.getIterationErr != nil {
		return nil, m.getIterationErr
	}
	return m.iteration, nil
}

// GetIterationTasks returns tasks for an iteration.
func (m *MockRepository) GetIterationTasks(ctx context.Context, iterationNumber int) ([]*entities.TaskEntity, error) {
	if m.getIterationTasksErr != nil {
		return nil, m.getIterationTasksErr
	}
	return m.iterationTasks, nil
}

// ListACByIteration returns ACs for an iteration.
func (m *MockRepository) ListACByIteration(ctx context.Context, iterationNumber int) ([]*entities.AcceptanceCriteriaEntity, error) {
	if m.listACByIterationErr != nil {
		return nil, m.listACByIterationErr
	}
	return m.acsByIteration, nil
}

// GetTask returns a task.
func (m *MockRepository) GetTask(ctx context.Context, taskID string) (*entities.TaskEntity, error) {
	if m.getTaskErr != nil {
		return nil, m.getTaskErr
	}
	return m.task, nil
}

// ListAC returns ACs for a task.
func (m *MockRepository) ListAC(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
	if m.listACErr != nil {
		return nil, m.listACErr
	}
	return m.acsByTask, nil
}

// GetTrack returns a track.
func (m *MockRepository) GetTrack(ctx context.Context, id string) (*entities.TrackEntity, error) {
	if m.getTrackErr != nil {
		return nil, m.getTrackErr
	}
	// Support dependency track resolution
	if m.dependencyTracks != nil {
		if depTrack, exists := m.dependencyTracks[id]; exists {
			return depTrack, nil
		}
	}
	return m.track, nil
}

// GetIterationsForTask returns iterations for a task.
func (m *MockRepository) GetIterationsForTask(ctx context.Context, taskID string) ([]*entities.IterationEntity, error) {
	if m.getIterationsForTaskErr != nil {
		return nil, m.getIterationsForTaskErr
	}
	return m.iterationsForTask, nil
}

// TestLoadRoadmapListDataSuccess verifies that LoadRoadmapListData successfully loads and transforms data.
func TestLoadRoadmapListDataSuccess(t *testing.T) {
	ctx := context.Background()

	roadmap := &entities.RoadmapEntity{
		ID:     "roadmap-1",
		Vision: "Test vision",
	}

	iterations := []*entities.IterationEntity{
		{
			Number: 1,
			Name:   "Iteration 1",
			Status: "planned",
		},
	}

	tracks := []*entities.TrackEntity{
		{
			ID:     "track-1",
			Title:  "Track 1",
			Status: "not-started",
		},
	}

	tasks := []*entities.TaskEntity{
		{
			ID:      "task-1",
			Title:   "Task 1",
			TrackID: "track-1",
			Status:  "todo",
		},
	}

	repo := &MockRepository{
		activeRoadmap:  roadmap,
		iterations:     iterations,
		tracks:         tracks,
		backlogTasks:   tasks,
	}

	vm, err := queries.LoadRoadmapListData(ctx, repo)
	if err != nil {
		t.Fatalf("LoadRoadmapListData failed: %v", err)
	}

	if vm == nil {
		t.Fatal("Expected non-nil ViewModel")
	}

	if vm.ActiveIterations == nil || vm.ActiveTracks == nil || vm.BacklogTasks == nil {
		t.Fatal("Expected ViewModel with initialized slices")
	}

	if vm.Vision != "Test vision" {
		t.Errorf("Expected Vision 'Test vision', got %q", vm.Vision)
	}
}

// TestLoadRoadmapListDataGetActiveRoadmapError verifies error handling when GetActiveRoadmap fails.
func TestLoadRoadmapListDataGetActiveRoadmapError(t *testing.T) {
	ctx := context.Background()

	repo := &MockRepository{
		getActiveRoadmapErr: errors.New("database error"),
	}

	vm, err := queries.LoadRoadmapListData(ctx, repo)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if vm != nil {
		t.Fatal("Expected nil ViewModel on error")
	}

	if err.Error() != "database error" {
		t.Fatalf("Expected 'database error', got %v", err)
	}
}

// TestLoadIterationDetailDataSuccess verifies that LoadIterationDetailData successfully loads and transforms data.
func TestLoadIterationDetailDataSuccess(t *testing.T) {
	ctx := context.Background()

	iteration := &entities.IterationEntity{
		Number: 1,
		Name:   "Iteration 1",
		Status: "planned",
	}

	tasks := []*entities.TaskEntity{
		{
			ID:      "task-1",
			Title:   "Task 1",
			TrackID: "track-1",
			Status:  "todo",
		},
	}

	acs := []*entities.AcceptanceCriteriaEntity{
		{
			ID:          "ac-1",
			TaskID:      "task-1",
			Description: "AC 1",
		},
	}

	repo := &MockRepository{
		iteration:      iteration,
		iterationTasks: tasks,
		acsByIteration: acs,
	}

	vm, err := queries.LoadIterationDetailData(ctx, repo, 1)
	if err != nil {
		t.Fatalf("LoadIterationDetailData failed: %v", err)
	}

	if vm == nil {
		t.Fatal("Expected non-nil ViewModel")
	}

	if vm.Number != 1 {
		t.Fatalf("Expected iteration number 1, got %d", vm.Number)
	}
}

// TestLoadIterationDetailDataGetIterationError verifies error handling when GetIteration fails.
func TestLoadIterationDetailDataGetIterationError(t *testing.T) {
	ctx := context.Background()

	repo := &MockRepository{
		getIterationErr: errors.New("iteration not found"),
	}

	vm, err := queries.LoadIterationDetailData(ctx, repo, 1)
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if vm != nil {
		t.Fatal("Expected nil ViewModel on error")
	}

	if err.Error() != "iteration not found" {
		t.Fatalf("Expected 'iteration not found', got %v", err)
	}
}

// TestLoadTaskDetailDataSuccess verifies that LoadTaskDetailData successfully loads and transforms data.
func TestLoadTaskDetailDataSuccess(t *testing.T) {
	ctx := context.Background()

	task := &entities.TaskEntity{
		ID:      "task-1",
		Title:   "Task 1",
		TrackID: "track-1",
		Status:  "todo",
	}

	track := &entities.TrackEntity{
		ID:    "track-1",
		Title: "Track 1",
	}

	acs := []*entities.AcceptanceCriteriaEntity{
		{
			ID:          "ac-1",
			TaskID:      "task-1",
			Description: "AC 1",
		},
	}

	iterations := []*entities.IterationEntity{
		{
			Number: 1,
			Name:   "Iteration 1",
		},
	}

	repo := &MockRepository{
		task:              task,
		acsByTask:         acs,
		track:             track,
		iterationsForTask: iterations,
	}

	vm, err := queries.LoadTaskDetailData(ctx, repo, "task-1")
	if err != nil {
		t.Fatalf("LoadTaskDetailData failed: %v", err)
	}

	if vm == nil {
		t.Fatal("Expected non-nil ViewModel")
	}

	if vm.ID != "task-1" {
		t.Fatalf("Expected task ID 'task-1', got %s", vm.ID)
	}
}

// TestLoadTaskDetailDataGetTaskError verifies error handling when GetTask fails.
func TestLoadTaskDetailDataGetTaskError(t *testing.T) {
	ctx := context.Background()

	repo := &MockRepository{
		getTaskErr: errors.New("task not found"),
	}

	vm, err := queries.LoadTaskDetailData(ctx, repo, "task-1")
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if vm != nil {
		t.Fatal("Expected nil ViewModel on error")
	}

	if err.Error() != "task not found" {
		t.Fatalf("Expected 'task not found', got %v", err)
	}
}

// Stub implementations for RoadmapRepository interface methods (not used in queries tests)

func (m *MockRepository) SaveRoadmap(ctx context.Context, roadmap *entities.RoadmapEntity) error {
	return nil
}

func (m *MockRepository) GetRoadmap(ctx context.Context, id string) (*entities.RoadmapEntity, error) {
	return nil, nil
}

func (m *MockRepository) UpdateRoadmap(ctx context.Context, roadmap *entities.RoadmapEntity) error {
	return nil
}

func (m *MockRepository) SaveTrack(ctx context.Context, track *entities.TrackEntity) error {
	return nil
}

func (m *MockRepository) UpdateTrack(ctx context.Context, track *entities.TrackEntity) error {
	return nil
}

func (m *MockRepository) DeleteTrack(ctx context.Context, id string) error {
	return nil
}

func (m *MockRepository) AddTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	return nil
}

func (m *MockRepository) RemoveTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	return nil
}

func (m *MockRepository) GetTrackDependencies(ctx context.Context, trackID string) ([]string, error) {
	return nil, nil
}

func (m *MockRepository) ValidateNoCycles(ctx context.Context, trackID string) error {
	return nil
}

func (m *MockRepository) GetTrackWithTasks(ctx context.Context, trackID string) (*entities.TrackEntity, error) {
	return nil, nil
}

func (m *MockRepository) SaveTask(ctx context.Context, task *entities.TaskEntity) error {
	return nil
}

func (m *MockRepository) ListTasks(ctx context.Context, filters entities.TaskFilters) ([]*entities.TaskEntity, error) {
	if m.listTasksErr != nil {
		return nil, m.listTasksErr
	}
	return m.tasksForTrack, nil
}

func (m *MockRepository) UpdateTask(ctx context.Context, task *entities.TaskEntity) error {
	return nil
}

func (m *MockRepository) DeleteTask(ctx context.Context, id string) error {
	return nil
}

func (m *MockRepository) MoveTaskToTrack(ctx context.Context, taskID, newTrackID string) error {
	return nil
}

func (m *MockRepository) SaveIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	return nil
}

func (m *MockRepository) GetCurrentIteration(ctx context.Context) (*entities.IterationEntity, error) {
	return nil, nil
}

func (m *MockRepository) UpdateIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	return nil
}

func (m *MockRepository) DeleteIteration(ctx context.Context, number int) error {
	return nil
}

func (m *MockRepository) AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error {
	return nil
}

func (m *MockRepository) RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error {
	return nil
}

func (m *MockRepository) GetIterationTasksWithWarnings(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, []string, error) {
	return nil, nil, nil
}

func (m *MockRepository) StartIteration(ctx context.Context, iterationNumber int) error {
	return nil
}

func (m *MockRepository) CompleteIteration(ctx context.Context, iterationNumber int) error {
	return nil
}

func (m *MockRepository) GetIterationByNumber(ctx context.Context, iterationNumber int) (*entities.IterationEntity, error) {
	return nil, nil
}

func (m *MockRepository) SaveADR(ctx context.Context, adr *entities.ADREntity) error {
	return nil
}

func (m *MockRepository) GetADR(ctx context.Context, id string) (*entities.ADREntity, error) {
	return nil, nil
}

func (m *MockRepository) ListADRs(ctx context.Context, trackID *string) ([]*entities.ADREntity, error) {
	return nil, nil
}

func (m *MockRepository) UpdateADR(ctx context.Context, adr *entities.ADREntity) error {
	return nil
}

func (m *MockRepository) SupersedeADR(ctx context.Context, adrID, supersededByID string) error {
	return nil
}

func (m *MockRepository) DeprecateADR(ctx context.Context, adrID string) error {
	return nil
}

func (m *MockRepository) GetADRsByTrack(ctx context.Context, trackID string) ([]*entities.ADREntity, error) {
	return nil, nil
}

func (m *MockRepository) SaveAC(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
	return nil
}

func (m *MockRepository) GetAC(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *MockRepository) UpdateAC(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
	return nil
}

func (m *MockRepository) DeleteAC(ctx context.Context, id string) error {
	return nil
}

func (m *MockRepository) ListACByTask(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *MockRepository) ListACByTrack(ctx context.Context, trackID string) ([]*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *MockRepository) ListFailedAC(ctx context.Context, filters entities.ACFilters) ([]*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *MockRepository) GetRoadmapWithTracks(ctx context.Context, roadmapID string) (*entities.RoadmapEntity, error) {
	return nil, nil
}

func (m *MockRepository) GetProjectMetadata(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *MockRepository) SetProjectMetadata(ctx context.Context, key, value string) error {
	return nil
}

func (m *MockRepository) GetProjectCode(ctx context.Context) string {
	return ""
}

func (m *MockRepository) GetNextSequenceNumber(ctx context.Context, entityType string) (int, error) {
	return 0, nil
}

// TestLoadTrackDetailDataSuccess verifies that LoadTrackDetailData successfully loads and transforms data.
func TestLoadTrackDetailDataSuccess(t *testing.T) {
	ctx := context.Background()

	track := &entities.TrackEntity{
		ID:           "TM-track-1",
		Title:        "Authentication System",
		Description:  "Implement user authentication",
		Status:       "in-progress",
		Rank:         100,
		Dependencies: []string{"TM-track-2"},
	}

	tasks := []*entities.TaskEntity{
		{
			ID:      "TM-task-1",
			Title:   "Implement login",
			TrackID: "TM-track-1",
			Status:  "todo",
		},
		{
			ID:      "TM-task-2",
			Title:   "Implement signup",
			TrackID: "TM-track-1",
			Status:  "done",
		},
	}

	depTrack := &entities.TrackEntity{
		ID:    "TM-track-2",
		Title: "Database Setup",
	}

	repo := &MockRepository{
		track:         track,
		tasksForTrack: tasks,
		dependencyTracks: map[string]*entities.TrackEntity{
			"TM-track-2": depTrack,
		},
	}

	vm, err := queries.LoadTrackDetailData(ctx, repo, "TM-track-1")
	if err != nil {
		t.Fatalf("LoadTrackDetailData failed: %v", err)
	}

	if vm == nil {
		t.Fatal("Expected non-nil ViewModel")
	}

	if vm.ID != "TM-track-1" {
		t.Fatalf("Expected track ID 'TM-track-1', got %s", vm.ID)
	}

	if vm.Title != "Authentication System" {
		t.Fatalf("Expected title 'Authentication System', got %s", vm.Title)
	}

	if len(vm.Dependencies) != 1 {
		t.Fatalf("Expected 1 dependency, got %d", len(vm.Dependencies))
	}

	if len(vm.DependencyLabels) != 1 {
		t.Fatalf("Expected 1 dependency label, got %d", len(vm.DependencyLabels))
	}

	if vm.DependencyLabels[0] != "Database Setup" {
		t.Fatalf("Expected dependency label 'Database Setup', got %s", vm.DependencyLabels[0])
	}

	// Verify task grouping
	if len(vm.TODOTasks) != 1 {
		t.Fatalf("Expected 1 TODO task, got %d", len(vm.TODOTasks))
	}

	if len(vm.DoneTasks) != 1 {
		t.Fatalf("Expected 1 done task, got %d", len(vm.DoneTasks))
	}
}

// TestLoadTrackDetailDataGetTrackError verifies error handling when GetTrack fails.
func TestLoadTrackDetailDataGetTrackError(t *testing.T) {
	ctx := context.Background()

	repo := &MockRepository{
		getTrackErr: errors.New("track not found"),
	}

	vm, err := queries.LoadTrackDetailData(ctx, repo, "TM-track-1")
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if vm != nil {
		t.Fatal("Expected nil ViewModel on error")
	}

	if err.Error() != "track not found" {
		t.Fatalf("Expected 'track not found', got %v", err)
	}
}

// TestLoadTrackDetailDataListTasksError verifies error handling when ListTasks fails.
func TestLoadTrackDetailDataListTasksError(t *testing.T) {
	ctx := context.Background()

	track := &entities.TrackEntity{
		ID:    "TM-track-1",
		Title: "Authentication System",
	}

	repo := &MockRepository{
		track:        track,
		listTasksErr: errors.New("database error"),
	}

	vm, err := queries.LoadTrackDetailData(ctx, repo, "TM-track-1")
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if vm != nil {
		t.Fatal("Expected nil ViewModel on error")
	}

	if err.Error() != "database error" {
		t.Fatalf("Expected 'database error', got %v", err)
	}
}

// TestLoadTrackDetailDataNoDependencies verifies loading track with no dependencies.
func TestLoadTrackDetailDataNoDependencies(t *testing.T) {
	ctx := context.Background()

	track := &entities.TrackEntity{
		ID:           "TM-track-1",
		Title:        "Authentication System",
		Dependencies: []string{},
	}

	tasks := []*entities.TaskEntity{}

	repo := &MockRepository{
		track:         track,
		tasksForTrack: tasks,
	}

	vm, err := queries.LoadTrackDetailData(ctx, repo, "TM-track-1")
	if err != nil {
		t.Fatalf("LoadTrackDetailData failed: %v", err)
	}

	if vm == nil {
		t.Fatal("Expected non-nil ViewModel")
	}

	if len(vm.Dependencies) != 0 {
		t.Fatalf("Expected 0 dependencies, got %d", len(vm.Dependencies))
	}

	if len(vm.DependencyLabels) != 0 {
		t.Fatalf("Expected 0 dependency labels, got %d", len(vm.DependencyLabels))
	}
}

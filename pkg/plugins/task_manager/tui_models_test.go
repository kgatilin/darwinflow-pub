package task_manager_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	tm "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// MockRepository is a mock implementation of RoadmapRepository for testing
type MockRepository struct {
	activeRoadmap *tm.RoadmapEntity
	tracks        []*tm.TrackEntity
	tasks         []*tm.TaskEntity
	iterations    []*tm.IterationEntity
	shouldError   bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		tracks:     []*tm.TrackEntity{},
		tasks:      []*tm.TaskEntity{},
		iterations: []*tm.IterationEntity{},
	}
}

func (m *MockRepository) SaveRoadmap(ctx context.Context, roadmap *tm.RoadmapEntity) error {
	if m.shouldError {
		return pluginsdk.ErrInternal
	}
	m.activeRoadmap = roadmap
	return nil
}

func (m *MockRepository) GetRoadmap(ctx context.Context, id string) (*tm.RoadmapEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	if m.activeRoadmap != nil && m.activeRoadmap.ID == id {
		return m.activeRoadmap, nil
	}
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) GetActiveRoadmap(ctx context.Context) (*tm.RoadmapEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	if m.activeRoadmap == nil {
		return nil, pluginsdk.ErrNotFound
	}
	return m.activeRoadmap, nil
}

func (m *MockRepository) UpdateRoadmap(ctx context.Context, roadmap *tm.RoadmapEntity) error {
	return nil
}

func (m *MockRepository) SaveTrack(ctx context.Context, track *tm.TrackEntity) error {
	return nil
}

func (m *MockRepository) GetTrack(ctx context.Context, id string) (*tm.TrackEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	for _, track := range m.tracks {
		if track.ID == id {
			return track, nil
		}
	}
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) ListTracks(ctx context.Context, roadmapID string, filters tm.TrackFilters) ([]*tm.TrackEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	return m.tracks, nil
}

func (m *MockRepository) UpdateTrack(ctx context.Context, track *tm.TrackEntity) error {
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
	return []string{}, nil
}

func (m *MockRepository) ValidateNoCycles(ctx context.Context, trackID string) error {
	return nil
}

func (m *MockRepository) SaveTask(ctx context.Context, task *tm.TaskEntity) error {
	return nil
}

func (m *MockRepository) GetTask(ctx context.Context, id string) (*tm.TaskEntity, error) {
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) ListTasks(ctx context.Context, filters tm.TaskFilters) ([]*tm.TaskEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	return m.tasks, nil
}

func (m *MockRepository) UpdateTask(ctx context.Context, task *tm.TaskEntity) error {
	return nil
}

func (m *MockRepository) DeleteTask(ctx context.Context, id string) error {
	return nil
}

func (m *MockRepository) MoveTaskToTrack(ctx context.Context, taskID, newTrackID string) error {
	return nil
}

func (m *MockRepository) SaveIteration(ctx context.Context, iteration *tm.IterationEntity) error {
	return nil
}

func (m *MockRepository) GetIteration(ctx context.Context, number int) (*tm.IterationEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	for _, iter := range m.iterations {
		if iter.Number == number {
			return iter, nil
		}
	}
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) GetCurrentIteration(ctx context.Context) (*tm.IterationEntity, error) {
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) ListIterations(ctx context.Context) ([]*tm.IterationEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	return m.iterations, nil
}

func (m *MockRepository) UpdateIteration(ctx context.Context, iteration *tm.IterationEntity) error {
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

func (m *MockRepository) GetIterationTasks(ctx context.Context, iterationNum int) ([]*tm.TaskEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	// Return tasks that are in the iteration
	var iterationTasks []*tm.TaskEntity
	for _, iter := range m.iterations {
		if iter.Number == iterationNum {
			for _, taskID := range iter.TaskIDs {
				for _, task := range m.tasks {
					if task.ID == taskID {
						iterationTasks = append(iterationTasks, task)
						break
					}
				}
			}
			break
		}
	}
	return iterationTasks, nil
}

func (m *MockRepository) StartIteration(ctx context.Context, iterationNum int) error {
	return nil
}

func (m *MockRepository) CompleteIteration(ctx context.Context, iterationNum int) error {
	return nil
}

func (m *MockRepository) GetRoadmapWithTracks(ctx context.Context, roadmapID string) (*tm.RoadmapEntity, error) {
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) GetTrackWithTasks(ctx context.Context, trackID string) (*tm.TrackEntity, error) {
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) GetProjectMetadata(ctx context.Context, key string) (string, error) {
	return "", pluginsdk.ErrNotFound
}

func (m *MockRepository) SetProjectMetadata(ctx context.Context, key, value string) error {
	return nil
}

func (m *MockRepository) GetProjectCode(ctx context.Context) string {
	return "TEST"
}

func (m *MockRepository) GetNextSequenceNumber(ctx context.Context, entityType string) (int, error) {
	switch entityType {
	case "task":
		return len(m.tasks) + 1, nil
	case "track":
		return len(m.tracks) + 1, nil
	case "iter":
		return len(m.iterations) + 1, nil
	default:
		return 0, pluginsdk.ErrInvalidArgument
	}
}

// Acceptance Criteria stub methods
func (m *MockRepository) SaveAC(ctx context.Context, ac *tm.AcceptanceCriteriaEntity) error {
	return nil
}

func (m *MockRepository) GetAC(ctx context.Context, id string) (*tm.AcceptanceCriteriaEntity, error) {
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) ListAC(ctx context.Context, taskID string) ([]*tm.AcceptanceCriteriaEntity, error) {
	return []*tm.AcceptanceCriteriaEntity{}, nil
}

func (m *MockRepository) UpdateAC(ctx context.Context, ac *tm.AcceptanceCriteriaEntity) error {
	return nil
}

func (m *MockRepository) DeleteAC(ctx context.Context, id string) error {
	return nil
}

func (m *MockRepository) ListACByTrack(ctx context.Context, trackID string) ([]*tm.AcceptanceCriteriaEntity, error) {
	return []*tm.AcceptanceCriteriaEntity{}, nil
}

func (m *MockRepository) ListACByIteration(ctx context.Context, iterationNum int) ([]*tm.AcceptanceCriteriaEntity, error) {
	return []*tm.AcceptanceCriteriaEntity{}, nil
}

// ADR stub methods
func (m *MockRepository) SaveADR(ctx context.Context, adr *tm.ADREntity) error {
	return nil
}

func (m *MockRepository) GetADR(ctx context.Context, id string) (*tm.ADREntity, error) {
	return nil, pluginsdk.ErrNotFound
}

func (m *MockRepository) ListADRs(ctx context.Context, trackID *string) ([]*tm.ADREntity, error) {
	return []*tm.ADREntity{}, nil
}

func (m *MockRepository) UpdateADR(ctx context.Context, adr *tm.ADREntity) error {
	return nil
}

func (m *MockRepository) SupersedeADR(ctx context.Context, adrID, supersededByID string) error {
	return nil
}

func (m *MockRepository) DeprecateADR(ctx context.Context, adrID string) error {
	return nil
}

func (m *MockRepository) GetADRsByTrack(ctx context.Context, trackID string) ([]*tm.ADREntity, error) {
	return []*tm.ADREntity{}, nil
}

func (m *MockRepository) GetIterationsForTask(ctx context.Context, taskID string) ([]*tm.IterationEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	// Return iterations that contain this task
	var result []*tm.IterationEntity
	for _, iter := range m.iterations {
		for _, id := range iter.TaskIDs {
			if id == taskID {
				result = append(result, iter)
				break
			}
		}
	}
	return result, nil
}

func (m *MockRepository) GetBacklogTasks(ctx context.Context) ([]*tm.TaskEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	// Return tasks that are not in any iteration and status != done
	var backlog []*tm.TaskEntity
	for _, task := range m.tasks {
		if task.Status == "done" {
			continue
		}
		inIteration := false
		for _, iter := range m.iterations {
			for _, taskID := range iter.TaskIDs {
				if taskID == task.ID {
					inIteration = true
					break
				}
			}
			if inIteration {
				break
			}
		}
		if !inIteration {
			backlog = append(backlog, task)
		}
	}
	return backlog, nil
}

func (m *MockRepository) ListFailedAC(ctx context.Context, filters tm.ACFilters) ([]*tm.AcceptanceCriteriaEntity, error) {
	if m.shouldError {
		return nil, pluginsdk.ErrInternal
	}
	// Return empty slice for mock
	return []*tm.AcceptanceCriteriaEntity{}, nil
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Tests

func TestNewAppModel(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)

	if model == nil {
		t.Fatal("NewAppModel returned nil")
	}
}

func TestAppModelInit(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()
	repo.activeRoadmap = &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	model := tm.NewAppModel(ctx, repo, logger)
	cmd := model.Init()

	if cmd == nil {
		t.Fatal("Init returned nil command")
	}
}

func TestAppModelRoadmapListView(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(&tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	model.SetTracks([]*tm.TrackEntity{
		{
			ID:          "track-1",
			RoadmapID:   "roadmap-1",
			Title:       "Track 1",
			Description: "Description 1",
			Status:      "in-progress",
			Rank:        200,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	})
	model.SetCurrentView(tm.ViewRoadmapList)
	model.SetDimensions(80, 24)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "Track 1") {
		t.Fatalf("View should contain track title, got: %s", view)
	}
}

func TestAppModelTrackDetailView(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentTrack(&tm.TrackEntity{
		ID:          "track-1",
		RoadmapID:   "roadmap-1",
		Title:       "Track 1",
		Description: "Description 1",
		Status:      "in-progress",
		Rank:        200,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	model.SetTasks([]*tm.TaskEntity{
		{
			ID:          "task-1",
			TrackID:     "track-1",
			Title:       "Task 1",
			Description: "Description 1",
			Status:      "todo",
			Rank:        300,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	})
	model.SetCurrentView(tm.ViewTrackDetail)
	model.SetDimensions(80, 24)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "Track 1") {
		t.Fatalf("View should contain track title, got: %s", view)
	}
	if !contains(view, "Task 1") {
		t.Fatalf("View should contain task title, got: %s", view)
	}
}

func TestAppModelKeyNavigation(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(&tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	model.SetTracks([]*tm.TrackEntity{
		{
			ID:          "track-1",
			RoadmapID:   "roadmap-1",
			Title:       "Track 1",
			Description: "Description 1",
			Status:      "in-progress",
			Rank:        200,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "track-2",
			RoadmapID:   "roadmap-1",
			Title:       "Track 2",
			Description: "Description 2",
			Status:      "todo",
			Rank:        400,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	})
	model.SetCurrentView(tm.ViewRoadmapList)

	// Test navigation down
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd != nil {
		t.Fatal("Unexpected command returned")
	}

	// Test navigation up
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if cmd != nil {
		t.Fatal("Unexpected command returned")
	}

	// Test quit
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("Expected quit command")
	}
}

func TestAppModelErrorView(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentView(tm.ViewError)
	model.SetError(pluginsdk.ErrNotFound)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "Error") {
		t.Fatalf("Error view should contain 'Error', got: %s", view)
	}
}

func TestAppModelLoadingView(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentView(tm.ViewLoading)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "Loading") {
		t.Fatalf("Loading view should contain 'Loading', got: %s", view)
	}
}

// Iteration view tests

func TestIterationListView(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{"task-1", "task-2"}, "complete", 500, now, now.AddDate(0, 0, 7), now, now)
	iter2, _ := tm.NewIterationEntity(2, "Sprint 2", "Features", "New features", []string{"task-3"}, "current", 500, now.AddDate(0, 0, 7), time.Time{}, now.AddDate(0, 0, 7), now.AddDate(0, 0, 7))

	repo.iterations = []*tm.IterationEntity{iter1, iter2}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetIterations(repo.iterations)
	model.SetCurrentView(tm.ViewIterationList)
	model.SetDimensions(80, 24)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "Iterations") {
		t.Fatalf("View should contain 'Iterations', got: %s", view)
	}
	if !contains(view, "Sprint 1") {
		t.Fatalf("View should contain 'Sprint 1', got: %s", view)
	}
	if !contains(view, "Sprint 2") {
		t.Fatalf("View should contain 'Sprint 2', got: %s", view)
	}
	if !contains(view, "Complete") {
		t.Fatalf("View should contain 'Complete', got: %s", view)
	}
	if !contains(view, "Current") {
		t.Fatalf("View should contain 'Current', got: %s", view)
	}
}

func TestIterationListViewEmpty(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetIterations([]*tm.IterationEntity{})
	model.SetCurrentView(tm.ViewIterationList)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "No iterations yet") {
		t.Fatalf("Empty view should contain 'No iterations yet', got: %s", view)
	}
}

func TestIterationDetailView(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create tasks
	task1 := tm.NewTaskEntity("task-1", "track-1", "Task 1", "Description 1", "done", 200, "", now, now)
	task2 := tm.NewTaskEntity("task-2", "track-1", "Task 2", "Description 2", "in-progress", 300, "", now, now)
	task3 := tm.NewTaskEntity("task-3", "track-1", "Task 3", "Description 3", "todo", 400, "", now, now)

	// Create iteration with tasks
	iter, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{"task-1", "task-2", "task-3"}, "current", 500, now, time.Time{}, now, now)

	repo.tasks = []*tm.TaskEntity{task1, task2, task3}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentIteration(iter)
	model.SetIterationTasks([]*tm.TaskEntity{task1, task2, task3})
	model.SetCurrentView(tm.ViewIterationDetail)
	model.SetDimensions(80, 24)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "Sprint 1") {
		t.Fatalf("View should contain 'Sprint 1', got: %s", view)
	}
	if !contains(view, "Foundation") {
		t.Fatalf("View should contain 'Foundation', got: %s", view)
	}
	if !contains(view, "Task 1") {
		t.Fatalf("View should contain 'Task 1', got: %s", view)
	}
	if !contains(view, "Task 2") {
		t.Fatalf("View should contain 'Task 2', got: %s", view)
	}
	if !contains(view, "Task 3") {
		t.Fatalf("View should contain 'Task 3', got: %s", view)
	}
	if !contains(view, "Progress:") {
		t.Fatalf("View should contain 'Progress:', got: %s", view)
	}
	if !contains(view, "To Do") {
		t.Fatalf("View should contain 'To Do', got: %s", view)
	}
	if !contains(view, "In Progress") {
		t.Fatalf("View should contain 'In Progress', got: %s", view)
	}
	if !contains(view, "Done") {
		t.Fatalf("View should contain 'Done', got: %s", view)
	}
}

func TestIterationDetailViewNoTasks(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentIteration(iter)
	model.SetIterationTasks([]*tm.TaskEntity{})
	model.SetCurrentView(tm.ViewIterationDetail)

	view := model.View()

	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !contains(view, "No tasks in this iteration") {
		t.Fatalf("View should contain 'No tasks in this iteration', got: %s", view)
	}
}

func TestNavigationToIterationList(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", 500, now, now, now, now)
	repo.iterations = []*tm.IterationEntity{iter1}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(&tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	model.SetTracks([]*tm.TrackEntity{})
	model.SetCurrentView(tm.ViewRoadmapList)
	model.SetIterations([]*tm.IterationEntity{iter1})

	// Press 'i' to go to iteration list
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Should switch to iteration list view
	if updatedModel.(*tm.AppModel).GetCurrentView() != tm.ViewIterationList {
		t.Fatalf("Expected ViewIterationList, got %d", updatedModel.(*tm.AppModel).GetCurrentView())
	}
}

func TestNavigationFromIterationListToDetail(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", 500, now, now, now, now)
	iter2, _ := tm.NewIterationEntity(2, "Sprint 2", "Features", "New features", []string{}, "current", 500, now, time.Time{}, now, now)
	repo.iterations = []*tm.IterationEntity{iter1, iter2}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetIterations(repo.iterations)
	model.SetCurrentView(tm.ViewIterationList)
	model.SetSelectedIterationIdx(0) // Select first iteration

	// Press 'enter' to view selected iteration - use tea.KeyEnter for proper handling
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return a command (loadIterationDetail)
	if cmd == nil {
		t.Fatal("Expected command for loading iteration detail")
	}
}

func TestNavigationFromIterationDetailBack(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", 500, now, now, now, now)

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentIteration(iter)
	model.SetCurrentView(tm.ViewIterationDetail)
	model.SetSelectedIterationIdx(0) // Set initial selection

	// Press 'esc' to go back to main view
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should return to main view (ViewRoadmapList) instead of ViewIterationList
	if model.GetCurrentView() != tm.ViewRoadmapList {
		t.Fatalf("Expected ViewRoadmapList after esc, got %v", model.GetCurrentView())
	}

	// Should preserve selected iteration index
	if model.GetSelectedIterationIdx() != 0 {
		t.Fatalf("Expected selected iteration index to be preserved (0), got %d", model.GetSelectedIterationIdx())
	}
}

func TestNavigationFromMainViewIterationSection(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", 500, now, now, now, now)
	iter2, _ := tm.NewIterationEntity(2, "Sprint 2", "Features", "New features", []string{}, "current", 500, now, time.Time{}, now, now)
	iter3, _ := tm.NewIterationEntity(3, "Sprint 3", "Polish", "Final touches", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)

	repo.iterations = []*tm.IterationEntity{iter1, iter2, iter3}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(&tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	model.SetIterations(repo.iterations)
	model.SetCurrentView(tm.ViewRoadmapList)
	model.SetSelectedIterationIdx(1) // Select second iteration (Sprint 2)

	// Press 'enter' to view selected iteration
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return a command (loadIterationDetail)
	if cmd == nil {
		t.Fatal("Expected command for loading iteration detail")
	}

	// Simulate loading iteration detail (would normally happen via message handling)
	model.SetCurrentIteration(iter2)
	model.SetCurrentView(tm.ViewIterationDetail)

	// Press 'esc' to go back to main view
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should return to main view
	if model.GetCurrentView() != tm.ViewRoadmapList {
		t.Fatalf("Expected ViewRoadmapList after esc, got %v", model.GetCurrentView())
	}

	// Should preserve selected iteration index (Sprint 2 = index 1)
	if model.GetSelectedIterationIdx() != 1 {
		t.Fatalf("Expected selected iteration index to be preserved (1), got %d", model.GetSelectedIterationIdx())
	}
}

func TestIterationListNavigation(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", 500, now, now, now, now)
	iter2, _ := tm.NewIterationEntity(2, "Sprint 2", "Features", "New features", []string{}, "current", 500, now, time.Time{}, now, now)
	iter3, _ := tm.NewIterationEntity(3, "Sprint 3", "Polish", "Final touches", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)

	repo.iterations = []*tm.IterationEntity{iter1, iter2, iter3}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetIterations(repo.iterations)
	model.SetCurrentView(tm.ViewIterationList)
	model.SetSelectedIterationIdx(0)

	// Navigate down
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.GetSelectedIterationIdx() != 1 {
		t.Fatalf("Expected selected index 1 after 'j', got %d", model.GetSelectedIterationIdx())
	}

	// Navigate down again
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if model.GetSelectedIterationIdx() != 2 {
		t.Fatalf("Expected selected index 2 after 'j', got %d", model.GetSelectedIterationIdx())
	}

	// Navigate up
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if model.GetSelectedIterationIdx() != 1 {
		t.Fatalf("Expected selected index 1 after 'k', got %d", model.GetSelectedIterationIdx())
	}
}

func TestProgressBarRendering(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)

	// Test various percentages
	tests := []struct {
		percent  float64
		expected string
	}{
		{0.0, "0.0%"},
		{50.0, "50.0%"},
		{100.0, "100.0%"},
	}

	for _, test := range tests {
		bar := model.RenderProgressBar(test.percent, 40)
		if !contains(bar, test.expected) {
			t.Fatalf("Progress bar should contain '%s', got: %s", test.expected, bar)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if len(s[i:]) < len(substr) {
			return false
		}
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestACListDisplay tests that AC list displays full text and task ID
func TestACListDisplay(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentView(tm.ViewACList)

	// Set up test ACs with long descriptions
	acs := []*tm.AcceptanceCriteriaEntity{
		{
			ID:               "ac-1",
			TaskID:           "TM-task-1",
			Description:      "This is a very long acceptance criterion description that should be displayed in full without any truncation",
			VerificationType: tm.VerificationTypeManual,
			Status:           tm.ACStatusNotStarted,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               "ac-2",
			TaskID:           "TM-task-2",
			Description:      "Another acceptance criterion with a long description",
			VerificationType: tm.VerificationTypeAutomated,
			Status:           tm.ACStatusPendingHumanReview,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	// Set up track
	track := &tm.TrackEntity{
		ID:          "TM-track-1",
		RoadmapID:   "roadmap-1",
		Title:       "Test Track",
		Description: "Test Description",
		Status:      "in-progress",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	model.SetCurrentTrack(track)
	model.SetACs(acs)

	// Render the view
	view := model.View()

	// Check that full description is displayed (not truncated)
	if len(view) == 0 {
		t.Fatal("View should not be empty")
	}

	// Check that task IDs are displayed
	if !contains(view, "TM-task-1") {
		t.Error("View should display task ID TM-task-1")
	}

	if !contains(view, "TM-task-2") {
		t.Error("View should display task ID TM-task-2")
	}

	// Check that full descriptions are present (not truncated)
	// With text wrapping, phrases may be split across lines, so check for individual words
	keyWords := []string{
		"This",
		"very",
		"long",
		"acceptance",
		"criterion",
		"description",
		"displayed",
		"truncation",
	}
	missingWords := []string{}
	for _, word := range keyWords {
		if !contains(view, word) {
			missingWords = append(missingWords, word)
		}
	}
	if len(missingWords) > 0 {
		t.Errorf("View is missing AC description words: %v", missingWords)
	}

	// Check that space bar help text is shown
	if !contains(view, "space: Verify selected AC") {
		t.Error("View should show space bar help text")
	}
}

// TestACSpaceBarVerification tests that space bar verifies selected AC
func TestACSpaceBarVerification(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentView(tm.ViewACList)

	// Set up test ACs
	acs := []*tm.AcceptanceCriteriaEntity{
		{
			ID:               "ac-1",
			TaskID:           "TM-task-1",
			Description:      "Test AC 1",
			VerificationType: tm.VerificationTypeManual,
			Status:           tm.ACStatusNotStarted,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               "ac-2",
			TaskID:           "TM-task-2",
			Description:      "Test AC 2",
			VerificationType: tm.VerificationTypeManual,
			Status:           tm.ACStatusPendingHumanReview,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	// Set up track
	track := &tm.TrackEntity{
		ID:          "TM-track-1",
		RoadmapID:   "roadmap-1",
		Title:       "Test Track",
		Description: "Test Description",
		Status:      "in-progress",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	model.SetCurrentTrack(track)
	model.SetACs(acs)
	model.SetSelectedACIdx(0)

	// Press space bar to verify
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	// Should return a command to reload ACs
	if cmd == nil {
		t.Error("Expected command to reload ACs after verification")
	}

	// Verify that the AC status was updated
	if acs[0].Status != tm.ACStatusVerified {
		t.Errorf("Expected AC status to be Verified, got %s", acs[0].Status)
	}
}

// TestACSpaceBarNoOpOnVerified tests that space bar does nothing on already verified AC
func TestACSpaceBarNoOpOnVerified(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentView(tm.ViewACList)

	// Set up test AC that's already verified
	acs := []*tm.AcceptanceCriteriaEntity{
		{
			ID:               "ac-1",
			TaskID:           "TM-task-1",
			Description:      "Test AC 1",
			VerificationType: tm.VerificationTypeManual,
			Status:           tm.ACStatusVerified,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	// Set up track
	track := &tm.TrackEntity{
		ID:          "TM-track-1",
		RoadmapID:   "roadmap-1",
		Title:       "Test Track",
		Description: "Test Description",
		Status:      "in-progress",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	model.SetCurrentTrack(track)
	model.SetACs(acs)
	model.SetSelectedACIdx(0)

	originalStatus := acs[0].Status

	// Press space bar
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	// Should not return a command since AC is already verified
	if cmd != nil {
		t.Error("Expected no command when AC is already verified")
	}

	// Verify that the status didn't change
	if acs[0].Status != originalStatus {
		t.Error("AC status should not change when already verified")
	}
}

// TestBacklogDisplayInMainView tests that backlog tasks are displayed in main view
func TestBacklogDisplayInMainView(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create roadmap
	roadmap := &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create track
	track1 := &tm.TrackEntity{
		ID:          "TM-track-1",
		RoadmapID:   "roadmap-1",
		Title:       "Core Framework",
		Description: "Core features",
		Status:      "in-progress",
		Rank:        100,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create backlog tasks (not in any iteration, status != done)
	task1 := tm.NewTaskEntity("TM-task-1", "TM-track-1", "Backlog Task 1", "Description 1", "todo", 200, "", now, now)
	task2 := tm.NewTaskEntity("TM-task-2", "TM-track-1", "Backlog Task 2", "Description 2", "todo", 300, "", now, now)
	task3 := tm.NewTaskEntity("TM-task-3", "TM-track-1", "Done Task", "Should not appear", "done", 400, "", now, now)

	repo.activeRoadmap = roadmap
	repo.tracks = []*tm.TrackEntity{track1}
	repo.tasks = []*tm.TaskEntity{task1, task2, task3}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(roadmap)
	model.SetTracks(repo.tracks)
	model.SetCurrentView(tm.ViewRoadmapList)
	model.SetDimensions(120, 40)

	// Simulate loading backlog data
	backlogTasks := []*tm.TaskEntity{task1, task2} // Exclude done task
	msg := tm.FullRoadmapDataLoadedMsg{
		IterationTasks: make(map[int][]*tm.TaskEntity),
		TrackTasks:     make(map[string][]*tm.TaskEntity),
		BacklogTasks:   backlogTasks,
	}
	_, _ = model.Update(msg)

	view := model.View()

	// Check that backlog section exists
	if !contains(view, "Backlog") {
		t.Fatal("View should contain 'Backlog' section")
	}

	// Check that backlog count is shown
	if !contains(view, "Backlog (2)") {
		t.Fatal("View should show backlog count (2)")
	}

	// Check that todo tasks are displayed
	if !contains(view, "TM-task-1") {
		t.Fatal("View should contain task TM-task-1")
	}
	if !contains(view, "Backlog Task 1") {
		t.Fatal("View should contain 'Backlog Task 1'")
	}

	if !contains(view, "TM-task-2") {
		t.Fatal("View should contain task TM-task-2")
	}
	if !contains(view, "Backlog Task 2") {
		t.Fatal("View should contain 'Backlog Task 2'")
	}

	// Check that done task is NOT displayed
	if contains(view, "TM-task-3") {
		t.Fatal("View should NOT contain done task TM-task-3")
	}
	if contains(view, "Done Task") {
		t.Fatal("View should NOT contain 'Done Task'")
	}
}

// TestBacklogDisplayWithTrackInfo tests that backlog tasks show track information
func TestBacklogDisplayWithTrackInfo(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create roadmap
	roadmap := &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create track
	track1 := &tm.TrackEntity{
		ID:          "TM-track-1",
		RoadmapID:   "roadmap-1",
		Title:       "Core Framework",
		Description: "Core features",
		Status:      "in-progress",
		Rank:        100,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create backlog task with track
	task1 := tm.NewTaskEntity("TM-task-1", "TM-track-1", "Backlog Task 1", "Description 1", "todo", 200, "", now, now)

	repo.activeRoadmap = roadmap
	repo.tracks = []*tm.TrackEntity{track1}
	repo.tasks = []*tm.TaskEntity{task1}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(roadmap)
	model.SetTracks(repo.tracks)
	model.SetCurrentView(tm.ViewRoadmapList)
	model.SetDimensions(120, 40)

	// Simulate loading backlog data
	backlogTasks := []*tm.TaskEntity{task1}
	msg := tm.FullRoadmapDataLoadedMsg{
		IterationTasks: make(map[int][]*tm.TaskEntity),
		TrackTasks:     make(map[string][]*tm.TaskEntity),
		BacklogTasks:   backlogTasks,
	}
	_, _ = model.Update(msg)

	view := model.View()

	// Check that track ID and title are displayed with task
	if !contains(view, "[TM-track-1: Core Framework]") {
		t.Fatal("View should contain track ID and title [TM-track-1: Core Framework]")
	}
}

// TestBacklogTabNavigation tests that Tab key navigates to backlog section
func TestBacklogTabNavigation(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create roadmap
	roadmap := &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create iteration
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "current", 500, now, time.Time{}, now, now)

	// Create track
	track1 := &tm.TrackEntity{
		ID:          "TM-track-1",
		RoadmapID:   "roadmap-1",
		Title:       "Core Framework",
		Description: "Core features",
		Status:      "in-progress",
		Rank:        100,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create backlog task
	task1 := tm.NewTaskEntity("TM-task-1", "TM-track-1", "Backlog Task 1", "Description 1", "todo", 200, "", now, now)

	repo.activeRoadmap = roadmap
	repo.iterations = []*tm.IterationEntity{iter1}
	repo.tracks = []*tm.TrackEntity{track1}
	repo.tasks = []*tm.TaskEntity{task1}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(roadmap)
	model.SetIterations(repo.iterations)
	model.SetTracks(repo.tracks)
	model.SetCurrentView(tm.ViewRoadmapList)

	// Simulate loading backlog data
	backlogTasks := []*tm.TaskEntity{task1}
	msg := tm.FullRoadmapDataLoadedMsg{
		IterationTasks: make(map[int][]*tm.TaskEntity),
		TrackTasks:     make(map[string][]*tm.TaskEntity),
		BacklogTasks:   backlogTasks,
	}
	_, _ = model.Update(msg)

	// Start in Iterations section (default)
	// Press Tab to move to Tracks
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Press Tab again to move to Backlog
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Now we should be in backlog section
	// Try navigating with j/k
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Navigation should work (no panic/error)
	// This tests that backlog section is accessible via Tab
}

// TestBacklogNavigationUpDown tests j/k navigation in backlog
func TestBacklogNavigationUpDown(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create roadmap
	roadmap := &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create backlog tasks
	task1 := tm.NewTaskEntity("TM-task-1", "TM-track-1", "Backlog Task 1", "Description 1", "todo", 200, "", now, now)
	task2 := tm.NewTaskEntity("TM-task-2", "TM-track-1", "Backlog Task 2", "Description 2", "todo", 300, "", now, now)
	task3 := tm.NewTaskEntity("TM-task-3", "TM-track-1", "Backlog Task 3", "Description 3", "todo", 400, "", now, now)

	repo.activeRoadmap = roadmap
	repo.tasks = []*tm.TaskEntity{task1, task2, task3}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(roadmap)
	model.SetCurrentView(tm.ViewRoadmapList)

	// Simulate loading backlog data
	backlogTasks := []*tm.TaskEntity{task1, task2, task3}
	msg := tm.FullRoadmapDataLoadedMsg{
		IterationTasks: make(map[int][]*tm.TaskEntity),
		TrackTasks:     make(map[string][]*tm.TaskEntity),
		BacklogTasks:   backlogTasks,
	}
	_, _ = model.Update(msg)

	// Navigate to backlog section (Tab twice from iterations)
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // To tracks
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // To backlog

	// Now test navigation
	// Press 'j' to move down
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Should now be at index 1 (second task)

	// Press 'j' again
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Should now be at index 2 (third task)

	// Press 'k' to move up
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	// Should now be at index 1 (second task)

	// Navigation completed without errors
}

// TestBacklogEnterViewsTaskDetail tests that Enter key opens backlog task detail
func TestBacklogEnterViewsTaskDetail(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create roadmap
	roadmap := &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create backlog task
	task1 := tm.NewTaskEntity("TM-task-1", "TM-track-1", "Backlog Task 1", "Description 1", "todo", 200, "", now, now)

	repo.activeRoadmap = roadmap
	repo.tasks = []*tm.TaskEntity{task1}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(roadmap)
	model.SetCurrentView(tm.ViewRoadmapList)

	// Simulate loading backlog data
	backlogTasks := []*tm.TaskEntity{task1}
	msg := tm.FullRoadmapDataLoadedMsg{
		IterationTasks: make(map[int][]*tm.TaskEntity),
		TrackTasks:     make(map[string][]*tm.TaskEntity),
		BacklogTasks:   backlogTasks,
	}
	_, _ = model.Update(msg)

	// Navigate to backlog section
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // To tracks
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // To backlog

	// Press Enter to view task detail
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return a command (loadTaskDetail)
	if cmd == nil {
		t.Fatal("Expected command for loading task detail")
	}
}

// TestBacklogEmptyState tests that empty backlog doesn't show section
func TestBacklogEmptyState(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create roadmap
	roadmap := &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	repo.activeRoadmap = roadmap

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(roadmap)
	model.SetCurrentView(tm.ViewRoadmapList)
	model.SetDimensions(120, 40)

	// Simulate loading empty backlog
	msg := tm.FullRoadmapDataLoadedMsg{
		IterationTasks: make(map[int][]*tm.TaskEntity),
		TrackTasks:     make(map[string][]*tm.TaskEntity),
		BacklogTasks:   []*tm.TaskEntity{}, // Empty backlog
	}
	_, _ = model.Update(msg)

	view := model.View()

	// Backlog section should not be displayed when empty
	if contains(view, "Backlog (0)") {
		t.Fatal("View should NOT show empty backlog section")
	}
}

// TestBacklogSelectionHighlight tests that selected backlog task is highlighted
func TestBacklogSelectionHighlight(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()

	// Create roadmap
	roadmap := &tm.RoadmapEntity{
		ID:              "roadmap-1",
		Vision:          "Test vision",
		SuccessCriteria: "Test criteria",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create backlog tasks
	task1 := tm.NewTaskEntity("TM-task-1", "TM-track-1", "Backlog Task 1", "Description 1", "todo", 200, "", now, now)
	task2 := tm.NewTaskEntity("TM-task-2", "TM-track-1", "Backlog Task 2", "Description 2", "todo", 300, "", now, now)

	repo.activeRoadmap = roadmap
	repo.tasks = []*tm.TaskEntity{task1, task2}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(roadmap)
	model.SetCurrentView(tm.ViewRoadmapList)
	model.SetDimensions(120, 40)

	// Simulate loading backlog data
	backlogTasks := []*tm.TaskEntity{task1, task2}
	msg := tm.FullRoadmapDataLoadedMsg{
		IterationTasks: make(map[int][]*tm.TaskEntity),
		TrackTasks:     make(map[string][]*tm.TaskEntity),
		BacklogTasks:   backlogTasks,
	}
	_, _ = model.Update(msg)

	// Navigate to backlog section
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // To tracks
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // To backlog

	view := model.View()

	// First task should be selected (arrow indicator)
	// Check for selection indicator
	if !contains(view, "→") {
		t.Fatal("View should contain selection indicator (→)")
	}
}

// TestWrapTextBasic tests basic text wrapping functionality
func TestWrapTextBasic(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "Empty string",
			text:     "",
			width:    10,
			expected: "",
		},
		{
			name:     "Text shorter than width",
			text:     "Hello",
			width:    20,
			expected: "Hello",
		},
		{
			name:     "Text equal to width",
			text:     "HelloWorld",
			width:    10,
			expected: "HelloWorld",
		},
		{
			name:     "Text longer than width - single wrap",
			text:     "Hello world",
			width:    8,
			expected: "Hello\nworld",
		},
		{
			name:     "Text with multiple words - multiple wraps",
			text:     "The quick brown fox jumps",
			width:    10,
			expected: "The quick\nbrown fox\njumps",
		},
		{
			name:     "Very long single word",
			text:     "abcdefghijklmnop",
			width:    10,
			expected: "abcdefghijklmnop",
		},
		{
			name:     "Whitespace handling",
			text:     "One  two   three",
			width:    10,
			expected: "One two\nthree",
		},
		{
			name:     "Zero width",
			text:     "Hello world",
			width:    0,
			expected: "Hello world",
		},
		{
			name:     "Negative width",
			text:     "Hello world",
			width:    -1,
			expected: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the wrapped function
			result := tm.WrapText(tt.text, tt.width)
			if result != tt.expected {
				t.Errorf("WrapText(%q, %d)\n  got:      %q\n  expected: %q",
					tt.text, tt.width, result, tt.expected)
			}
		})
	}
}

// TestWrapTextEdgeCases tests edge cases for text wrapping
func TestWrapTextEdgeCases(t *testing.T) {
	// Very long description that should wrap at multiple points
	longText := "This is a comprehensive description that contains multiple words and should wrap to fit within the terminal width without causing overflow issues on the display"
	result := tm.WrapText(longText, 20)
	lines := strings.Split(result, "\n")

	// Check that no line exceeds the width
	for i, line := range lines {
		if len(line) > 20 {
			t.Errorf("Line %d exceeds width: len=%d, content=%q", i, len(line), line)
		}
	}

	// Verify we got multiple lines
	if len(lines) < 5 {
		t.Errorf("Expected multiple lines for long text, got %d", len(lines))
	}
}

// TestWrapTextPreservesWords tests that wrapping preserves words correctly
func TestWrapTextPreservesWords(t *testing.T) {
	text := "The quick brown fox"
	result := tm.WrapText(text, 8)
	lines := strings.Split(result, "\n")

	// All lines should be actual content lines (no empty lines)
	for _, line := range lines {
		if line == "" {
			t.Errorf("WrapText should not produce empty lines")
		}
	}

	// Reconstruct original text (minus whitespace) and verify it matches
	reconstructed := strings.Join(lines, " ")
	if reconstructed != text {
		t.Errorf("Reconstructed text doesn't match original\n  original: %q\n  reconstructed: %q", text, reconstructed)
	}
}


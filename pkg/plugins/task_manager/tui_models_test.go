package task_manager_test

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	tm "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
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
		ID:               "roadmap-1",
		Vision:           "Test vision",
		SuccessCriteria:  "Test criteria",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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
		ID:               "roadmap-1",
		Vision:           "Test vision",
		SuccessCriteria:  "Test criteria",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	model.SetTracks([]*tm.TrackEntity{
		{
			ID:          "track-1",
			RoadmapID:   "roadmap-1",
			Title:       "Track 1",
			Description: "Description 1",
			Status:      "in-progress",
			Priority:    "high",
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
		Priority:    "high",
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
			Priority:    "medium",
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
		ID:               "roadmap-1",
		Vision:           "Test vision",
		SuccessCriteria:  "Test criteria",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	model.SetTracks([]*tm.TrackEntity{
		{
			ID:          "track-1",
			RoadmapID:   "roadmap-1",
			Title:       "Track 1",
			Description: "Description 1",
			Status:      "in-progress",
			Priority:    "high",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "track-2",
			RoadmapID:   "roadmap-1",
			Title:       "Track 2",
			Description: "Description 2",
			Status:      "todo",
			Priority:    "low",
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
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{"task-1", "task-2"}, "complete", now, now.AddDate(0, 0, 7), now, now)
	iter2, _ := tm.NewIterationEntity(2, "Sprint 2", "Features", "New features", []string{"task-3"}, "current", now.AddDate(0, 0, 7), time.Time{}, now.AddDate(0, 0, 7), now.AddDate(0, 0, 7))

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
	task1 := tm.NewTaskEntity("task-1", "track-1", "Task 1", "Description 1", "done", "high", "", now, now)
	task2 := tm.NewTaskEntity("task-2", "track-1", "Task 2", "Description 2", "in-progress", "medium", "", now, now)
	task3 := tm.NewTaskEntity("task-3", "track-1", "Task 3", "Description 3", "todo", "low", "", now, now)

	// Create iteration with tasks
	iter, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{"task-1", "task-2", "task-3"}, "current", now, time.Time{}, now, now)

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
	if !contains(view, "Todo:") {
		t.Fatalf("View should contain 'Todo:', got: %s", view)
	}
	if !contains(view, "In Progress:") {
		t.Fatalf("View should contain 'In Progress:', got: %s", view)
	}
	if !contains(view, "Done:") {
		t.Fatalf("View should contain 'Done:', got: %s", view)
	}
}

func TestIterationDetailViewNoTasks(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "planned", time.Time{}, time.Time{}, now, now)

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
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", now, now, now, now)
	repo.iterations = []*tm.IterationEntity{iter1}

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetRoadmap(&tm.RoadmapEntity{
		ID:               "roadmap-1",
		Vision:           "Test vision",
		SuccessCriteria:  "Test criteria",
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	model.SetTracks([]*tm.TrackEntity{})
	model.SetCurrentView(tm.ViewRoadmapList)

	// Press 'i' to go to iteration list
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	// Should return a command (loadIterations)
	if cmd == nil {
		t.Fatal("Expected command for loading iterations")
	}
}

func TestNavigationFromIterationListToDetail(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", now, now, now, now)
	iter2, _ := tm.NewIterationEntity(2, "Sprint 2", "Features", "New features", []string{}, "current", now, time.Time{}, now, now)
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
	iter, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", now, now, now, now)

	model := tm.NewAppModel(ctx, repo, logger)
	model.SetCurrentIteration(iter)
	model.SetCurrentView(tm.ViewIterationDetail)

	// Press 'esc' to go back to iteration list
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should not return a command (handled in Update)
	if model.GetCurrentView() != tm.ViewIterationList {
		t.Fatalf("Expected ViewIterationList after esc, got %v", model.GetCurrentView())
	}
}

func TestIterationListNavigation(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository()
	logger := NewMockLogger()

	now := time.Now()
	iter1, _ := tm.NewIterationEntity(1, "Sprint 1", "Foundation", "Core features", []string{}, "complete", now, now, now, now)
	iter2, _ := tm.NewIterationEntity(2, "Sprint 2", "Features", "New features", []string{}, "current", now, time.Time{}, now, now)
	iter3, _ := tm.NewIterationEntity(3, "Sprint 3", "Polish", "Final touches", []string{}, "planned", time.Time{}, time.Time{}, now, now)

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

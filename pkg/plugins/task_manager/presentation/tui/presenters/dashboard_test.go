package presenters_test

import (
	"context"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/presenters"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/viewmodels"
)

func TestRoadmapListPresenter_TabNavigation(t *testing.T) {
	// Create test view model with items in all sections
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
		},
		ActiveTracks: []*viewmodels.TrackCardViewModel{
			{ID: "TM-track-1", Title: "Track 1", TaskCount: 2},
		},
		BacklogTasks: []*viewmodels.BacklogTaskViewModel{
			{ID: "TM-task-1", Title: "Task 1"},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Simulate Tab key press
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	_, cmd := presenter.Update(tabMsg)

	// Verify no error occurred
	if cmd != nil {
		t.Errorf("Expected no command from tab navigation, got %v", cmd)
	}

	// Verify section header highlighting in view (visual feedback test)
	view := presenter.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestRoadmapListPresenter_TabNavigationCycle(t *testing.T) {
	// Create test view model with items in all sections
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
			{Number: 2, Name: "Iteration 2", TaskCount: 1},
		},
		ActiveTracks: []*viewmodels.TrackCardViewModel{
			{ID: "TM-track-1", Title: "Track 1", TaskCount: 2},
		},
		BacklogTasks: []*viewmodels.BacklogTaskViewModel{
			{ID: "TM-task-1", Title: "Task 1"},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Press Tab 3 times to cycle through all sections
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	for i := 0; i < 3; i++ {
		p, _ := presenter.Update(tabMsg)
		presenter = p.(*presenters.RoadmapListPresenter)
	}

	// Verify presenter is still functional
	view := presenter.View()
	if view == "" {
		t.Error("Expected non-empty view after cycling sections")
	}
}

func TestRoadmapListPresenter_RefreshKey(t *testing.T) {
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Simulate 'r' key press (refresh)
	rMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := presenter.Update(rMsg)

	// Verify command is returned (refresh request)
	if cmd == nil {
		t.Error("Expected refresh command from 'r' key, got nil")
	}

	// Execute command to verify it returns RefreshDashboardMsg
	msg := cmd()
	if _, ok := msg.(presenters.RefreshDashboardMsg); !ok {
		t.Errorf("Expected RefreshDashboardMsg, got %T", msg)
	}
}

func TestRoadmapListKeyMap_TabKeyBinding(t *testing.T) {
	keys := presenters.NewRoadmapListKeyMap()

	// Verify Tab key binding exists
	if keys.Tab.Keys()[0] != "tab" {
		t.Errorf("Expected tab key binding, got %v", keys.Tab.Keys())
	}

	// Verify Tab key is in short help
	shortHelp := keys.ShortHelp()
	found := false
	for _, binding := range shortHelp {
		if key.Matches(tea.KeyMsg{Type: tea.KeyTab}, binding) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Tab key in short help")
	}
}

func TestRoadmapListKeyMap_RefreshKeyBinding(t *testing.T) {
	keys := presenters.NewRoadmapListKeyMap()

	// Verify Refresh key binding exists
	if keys.Refresh.Keys()[0] != "r" {
		t.Errorf("Expected r key binding, got %v", keys.Refresh.Keys())
	}

	// Verify Refresh key is in short help
	shortHelp := keys.ShortHelp()
	found := false
	for _, binding := range shortHelp {
		if key.Matches(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, binding) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Refresh key in short help")
	}
}

func TestRefreshDashboardMsg_PreservesSelection(t *testing.T) {
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
			{Number: 2, Name: "Iteration 2", TaskCount: 1},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Navigate to second item
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	p, _ := presenter.Update(downMsg)
	presenter = p.(*presenters.RoadmapListPresenter)

	// Press 'r' to refresh
	rMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := presenter.Update(rMsg)

	// Verify refresh message contains selected index
	msg := cmd()
	refreshMsg, ok := msg.(presenters.RefreshDashboardMsg)
	if !ok {
		t.Fatalf("Expected RefreshDashboardMsg, got %T", msg)
	}

	if refreshMsg.SelectedIndex != 1 {
		t.Errorf("Expected SelectedIndex=1, got %d", refreshMsg.SelectedIndex)
	}
}

func TestRoadmapListPresenter_EnterOnBacklogTask(t *testing.T) {
	// Create test view model with items in all sections
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
		},
		ActiveTracks: []*viewmodels.TrackCardViewModel{
			{ID: "TM-track-1", Title: "Track 1", TaskCount: 2},
		},
		BacklogTasks: []*viewmodels.BacklogTaskViewModel{
			{ID: "TM-task-1", Title: "Task 1", Status: "todo"},
			{ID: "TM-task-2", Title: "Task 2", Status: "todo"},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Navigate to first backlog task (index = 1 iteration + 1 track + 0 = 2)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	p, _ := presenter.Update(downMsg) // index=1
	presenter = p.(*presenters.RoadmapListPresenter)
	p, _ = presenter.Update(downMsg) // index=2 (first backlog task)
	presenter = p.(*presenters.RoadmapListPresenter)

	// Press Enter on backlog task
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := presenter.Update(enterMsg)

	// Verify TaskSelectedMsg is returned
	if cmd == nil {
		t.Fatal("Expected command from Enter on backlog task, got nil")
	}

	msg := cmd()
	taskMsg, ok := msg.(presenters.TaskSelectedMsg)
	if !ok {
		t.Fatalf("Expected TaskSelectedMsg, got %T", msg)
	}

	if taskMsg.TaskID != "TM-task-1" {
		t.Errorf("Expected TaskID=TM-task-1, got %s", taskMsg.TaskID)
	}
}

func TestRoadmapListPresenter_EnterOnSecondBacklogTask(t *testing.T) {
	// Create test view model with items in all sections
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
			{Number: 2, Name: "Iteration 2", TaskCount: 1},
		},
		ActiveTracks: []*viewmodels.TrackCardViewModel{
			{ID: "TM-track-1", Title: "Track 1", TaskCount: 2},
			{ID: "TM-track-2", Title: "Track 2", TaskCount: 1},
		},
		BacklogTasks: []*viewmodels.BacklogTaskViewModel{
			{ID: "TM-task-1", Title: "Task 1", Status: "todo"},
			{ID: "TM-task-2", Title: "Task 2", Status: "todo"},
			{ID: "TM-task-3", Title: "Task 3", Status: "todo"},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Navigate to second backlog task (index = 2 iterations + 2 tracks + 1 = 5)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 5; i++ {
		p, _ := presenter.Update(downMsg)
		presenter = p.(*presenters.RoadmapListPresenter)
	}

	// Press Enter on second backlog task
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := presenter.Update(enterMsg)

	// Verify TaskSelectedMsg is returned with correct task ID
	if cmd == nil {
		t.Fatal("Expected command from Enter on backlog task, got nil")
	}

	msg := cmd()
	taskMsg, ok := msg.(presenters.TaskSelectedMsg)
	if !ok {
		t.Fatalf("Expected TaskSelectedMsg, got %T", msg)
	}

	if taskMsg.TaskID != "TM-task-2" {
		t.Errorf("Expected TaskID=TM-task-2, got %s", taskMsg.TaskID)
	}
}

func TestRoadmapListPresenter_EnterOnIteration(t *testing.T) {
	// Create test view model with iteration
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Press Enter on first iteration (index=0)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := presenter.Update(enterMsg)

	// Verify IterationSelectedMsg is returned
	if cmd == nil {
		t.Fatal("Expected command from Enter on iteration, got nil")
	}

	msg := cmd()
	iterMsg, ok := msg.(presenters.IterationSelectedMsg)
	if !ok {
		t.Fatalf("Expected IterationSelectedMsg, got %T", msg)
	}

	if iterMsg.IterationNumber != 1 {
		t.Errorf("Expected IterationNumber=1, got %d", iterMsg.IterationNumber)
	}
}

func TestRoadmapListPresenter_EnterOnTrack(t *testing.T) {
	// Create test view model with iteration and track
	vm := &viewmodels.RoadmapListViewModel{
		ActiveIterations: []*viewmodels.IterationCardViewModel{
			{Number: 1, Name: "Iteration 1", TaskCount: 3},
		},
		ActiveTracks: []*viewmodels.TrackCardViewModel{
			{ID: "TM-track-1", Title: "Track 1", TaskCount: 2},
		},
	}

	presenter := presenters.NewRoadmapListPresenter(vm, nil, context.Background())

	// Navigate to track (index=1)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	p, _ := presenter.Update(downMsg)
	presenter = p.(*presenters.RoadmapListPresenter)

	// Press Enter on track
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := presenter.Update(enterMsg)

	// Verify TrackSelectedMsg is returned
	if cmd == nil {
		t.Fatal("Expected command from Enter on track, got nil")
	}

	msg := cmd()
	trackMsg, ok := msg.(presenters.TrackSelectedMsg)
	if !ok {
		t.Fatalf("Expected TrackSelectedMsg, got %T", msg)
	}

	if trackMsg.TrackID != "TM-track-1" {
		t.Errorf("Expected TrackID=TM-track-1, got %s", trackMsg.TrackID)
	}
}

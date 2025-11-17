package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/presenters"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/queries"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/viewmodels"
)

// ViewStateNew represents the current view in the new MVP TUI
type ViewStateNew int

const (
	ViewLoadingNew ViewStateNew = iota
	ViewErrorNew
	ViewRoadmapListNew
	ViewIterationDetailNew
	ViewTaskDetailNew
	ViewTrackDetailNew
)

// AppModelNew is the root Bubble Tea model for the new MVP TUI
type AppModelNew struct {
	ctx         context.Context
	repo        domain.RoadmapRepository
	logger      pluginsdk.Logger
	projectName string

	currentView     ViewStateNew
	activePresenter presenters.Presenter
	lastError       error

	// Navigation state tracking
	previousView           ViewStateNew
	currentIterationNumber int
	currentTaskID          string
	currentTrackID         string
	currentActiveTab       presenters.IterationDetailTab // Track active tab for AC actions
	dashboardSelectedIndex int                            // Dashboard selected index (for restoring focus on return)

	width  int
	height int
}

// NewAppModelNew creates a new application model for the MVP TUI
func NewAppModelNew(
	ctx context.Context,
	repo domain.RoadmapRepository,
	logger pluginsdk.Logger,
	projectName string,
) *AppModelNew {
	return &AppModelNew{
		ctx:         ctx,
		repo:        repo,
		logger:      logger,
		projectName: projectName,
		currentView: ViewLoadingNew,
	}
}

func (m *AppModelNew) Init() tea.Cmd {
	loadingVM := viewmodels.NewLoadingViewModel("Loading dashboard...")
	m.activePresenter = presenters.NewLoadingPresenter(loadingVM)

	return tea.Batch(
		m.activePresenter.Init(),
		m.loadRoadmapList(),
	)
}

func (m *AppModelNew) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case roadmapListLoadedMsg:
		// Transition to RoadmapListPresenter with loaded data
		m.currentView = ViewRoadmapListNew
		// Use the selected index from message if provided (non-nil)
		if msg.selectedIndex != nil {
			m.activePresenter = presenters.NewRoadmapListPresenterWithSelection(msg.viewModel, m.repo, m.ctx, *msg.selectedIndex)
		} else {
			m.activePresenter = presenters.NewRoadmapListPresenter(msg.viewModel, m.repo, m.ctx)
		}
		return m, m.activePresenter.Init()

	case presenters.ErrorMsg:
		m.lastError = msg.Err
		// Track the view we came from before showing error (so we can navigate back)
		// Only update previousView if we're not already in error view
		if m.currentView != ViewErrorNew {
			m.previousView = m.currentView
		}
		m.currentView = ViewErrorNew
		errorVM := viewmodels.NewErrorViewModel(msg.Err.Error())
		errorVM.CanGoBack = true
		errorVM.RetryAction = "Fix the issue and try again"
		m.activePresenter = presenters.NewErrorPresenter(errorVM)
		return m, m.activePresenter.Init()

	case presenters.BackMsgNew:
		if m.currentView == ViewErrorNew {
			// Navigate back to the view we came from before the error
			if m.previousView == ViewRoadmapListNew {
				m.currentView = ViewLoadingNew
				loadingVM := viewmodels.NewLoadingViewModel("Loading dashboard...")
				m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
				return m, tea.Batch(
					m.activePresenter.Init(),
					m.loadRoadmapList(),
				)
			}
			if m.previousView == ViewIterationDetailNew && m.currentIterationNumber > 0 {
				m.currentView = ViewLoadingNew
				loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading iteration #%d...", m.currentIterationNumber))
				m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
				return m, tea.Batch(
					m.activePresenter.Init(),
					m.loadIterationDetail(m.currentIterationNumber),
				)
			}
			if m.previousView == ViewTrackDetailNew && m.currentTrackID != "" {
				m.currentView = ViewLoadingNew
				loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading track %s...", m.currentTrackID))
				m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
				return m, tea.Batch(
					m.activePresenter.Init(),
					m.loadTrackDetail(m.currentTrackID),
				)
			}
			if m.previousView == ViewTaskDetailNew && m.currentTaskID != "" {
				m.currentView = ViewLoadingNew
				loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading task %s...", m.currentTaskID))
				m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
				return m, tea.Batch(
					m.activePresenter.Init(),
					m.loadTaskDetail(m.currentTaskID),
				)
			}
			// Fallback: if no previous view tracked, go to dashboard
			m.currentView = ViewLoadingNew
			loadingVM := viewmodels.NewLoadingViewModel("Loading dashboard...")
			m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
			return m, tea.Batch(
				m.activePresenter.Init(),
				m.loadRoadmapList(),
			)
		}
		if m.currentView == ViewTrackDetailNew {
			// Go back to dashboard from track detail
			m.currentView = ViewLoadingNew
			loadingVM := viewmodels.NewLoadingViewModel("Loading dashboard...")
			m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
			return m, tea.Batch(
				m.activePresenter.Init(),
				m.loadRoadmapListWithIndex(m.dashboardSelectedIndex),
			)
		}
		if m.currentView == ViewTaskDetailNew {
			// Go back to track detail if we came from there
			if m.previousView == ViewTrackDetailNew && m.currentTrackID != "" {
				m.currentView = ViewLoadingNew
				loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading track %s...", m.currentTrackID))
				m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
				return m, tea.Batch(
					m.activePresenter.Init(),
					m.loadTrackDetail(m.currentTrackID),
				)
			}
			// Go back to iteration detail if we came from there
			if m.previousView == ViewIterationDetailNew && m.currentIterationNumber > 0 {
				m.currentView = ViewLoadingNew
				loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading iteration #%d...", m.currentIterationNumber))
				m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
				return m, tea.Batch(
					m.activePresenter.Init(),
					m.loadIterationDetail(m.currentIterationNumber),
				)
			}
			// Otherwise go back to dashboard (restore selection from backlog navigation)
			m.currentView = ViewLoadingNew
			loadingVM := viewmodels.NewLoadingViewModel("Loading dashboard...")
			m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
			return m, tea.Batch(
				m.activePresenter.Init(),
				m.loadRoadmapListWithIndex(m.dashboardSelectedIndex),
			)
		}
		if m.currentView == ViewIterationDetailNew {
			// Go back to dashboard (restore selection from iteration navigation)
			m.currentView = ViewLoadingNew
			loadingVM := viewmodels.NewLoadingViewModel("Loading dashboard...")
			m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
			return m, tea.Batch(
				m.activePresenter.Init(),
				m.loadRoadmapListWithIndex(m.dashboardSelectedIndex),
			)
		}

	case presenters.IterationSelectedMsg:
		// Load iteration detail
		m.previousView = m.currentView
		m.currentIterationNumber = msg.IterationNumber
		m.dashboardSelectedIndex = msg.SelectedIndex
		m.currentView = ViewLoadingNew
		loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading iteration #%d...", msg.IterationNumber))
		m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
		return m, tea.Batch(
			m.activePresenter.Init(),
			m.loadIterationDetail(msg.IterationNumber),
		)

	case presenters.TrackSelectedMsg:
		// Load track detail
		m.previousView = m.currentView
		m.currentTrackID = msg.TrackID
		m.dashboardSelectedIndex = msg.SelectedIndex
		m.currentView = ViewLoadingNew
		loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading track %s...", msg.TrackID))
		m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
		return m, tea.Batch(
			m.activePresenter.Init(),
			m.loadTrackDetail(msg.TrackID),
		)

	case trackDetailLoadedMsg:
		// Transition to TrackDetailPresenter
		m.currentView = ViewTrackDetailNew
		if msg.selectedIndex != nil {
			m.activePresenter = presenters.NewTrackDetailPresenterWithSelection(msg.viewModel, m.repo, m.ctx, *msg.selectedIndex)
		} else {
			m.activePresenter = presenters.NewTrackDetailPresenter(msg.viewModel, m.repo, m.ctx)
		}
		return m, m.activePresenter.Init()

	case iterationDetailLoadedMsg:
		// Transition to IterationDetailPresenter with saved activeTab and optional selectedIndex
		m.currentView = ViewIterationDetailNew
		if msg.selectedIndex != nil {
			m.activePresenter = presenters.NewIterationDetailPresenterWithSelection(msg.viewModel, m.repo, m.ctx, msg.activeTab, *msg.selectedIndex)
		} else {
			m.activePresenter = presenters.NewIterationDetailPresenterWithTab(msg.viewModel, m.repo, m.ctx, msg.activeTab)
		}
		return m, m.activePresenter.Init()

	case presenters.TaskSelectedMsg:
		// Load task detail
		m.previousView = m.currentView
		m.currentTaskID = msg.TaskID
		m.dashboardSelectedIndex = msg.SelectedIndex
		m.currentView = ViewLoadingNew
		loadingVM := viewmodels.NewLoadingViewModel(fmt.Sprintf("Loading task %s...", msg.TaskID))
		m.activePresenter = presenters.NewLoadingPresenter(loadingVM)
		return m, tea.Batch(
			m.activePresenter.Init(),
			m.loadTaskDetail(msg.TaskID),
		)

	case taskDetailLoadedMsg:
		// Transition to TaskDetailPresenter
		m.currentView = ViewTaskDetailNew
		if msg.selectedIndex != nil {
			m.activePresenter = presenters.NewTaskDetailPresenterWithSelection(msg.viewModel, m.repo, m.ctx, *msg.selectedIndex)
		} else {
			m.activePresenter = presenters.NewTaskDetailPresenter(msg.viewModel, m.repo, m.ctx)
		}
		return m, m.activePresenter.Init()

	case presenters.ACActionCompletedMsg:
		// Save the active tab and reload current view after AC action
		m.currentActiveTab = msg.ActiveTab
		if m.currentView == ViewIterationDetailNew && m.currentIterationNumber > 0 {
			return m, m.loadIterationDetailWithTabAndSelection(m.currentIterationNumber, msg.ActiveTab, msg.SelectedIndex)
		}
		if m.currentView == ViewTaskDetailNew && m.currentTaskID != "" {
			return m, m.loadTaskDetailWithSelection(m.currentTaskID, msg.SelectedIndex)
		}
		return m, nil

	case presenters.ReorderCompletedMsg:
		// Reload dashboard after iteration reordering, preserving selected iteration
		selectedIterationNumber := msg.SelectedIterationNumber
		return m, m.loadRoadmapListWithSelection(selectedIterationNumber)

	case presenters.RefreshDashboardMsg:
		// Reload dashboard data, preserving selected index
		return m, m.loadRoadmapListWithIndex(msg.SelectedIndex)
	}

	if m.activePresenter != nil {
		var cmd tea.Cmd
		m.activePresenter, cmd = m.activePresenter.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *AppModelNew) View() string {
	if m.activePresenter != nil {
		return m.activePresenter.View()
	}
	return "\nInitializing...\n"
}

func (m *AppModelNew) loadRoadmapList() tea.Cmd {
	return func() tea.Msg {
		vm, err := queries.LoadRoadmapListData(m.ctx, m.repo)
		if err != nil {
			return presenters.ErrorMsg{Err: err}
		}
		return roadmapListLoadedMsg{viewModel: vm, selectedIndex: nil}
	}
}

func (m *AppModelNew) loadRoadmapListWithSelection(iterationNumber int) tea.Cmd {
	return func() tea.Msg {
		vm, err := queries.LoadRoadmapListData(m.ctx, m.repo)
		if err != nil {
			return presenters.ErrorMsg{Err: err}
		}
		// Find the index of the iteration in the reloaded view model
		selectedIndex := m.findIterationIndex(vm, iterationNumber)
		return roadmapListLoadedMsg{viewModel: vm, selectedIndex: &selectedIndex}
	}
}

func (m *AppModelNew) loadRoadmapListWithIndex(index int) tea.Cmd {
	return func() tea.Msg {
		vm, err := queries.LoadRoadmapListData(m.ctx, m.repo)
		if err != nil {
			return presenters.ErrorMsg{Err: err}
		}
		// Clamp index to valid range
		totalItems := len(vm.ActiveIterations) + len(vm.ActiveTracks) + len(vm.BacklogTasks)
		if index >= totalItems {
			index = totalItems - 1
		}
		if index < 0 {
			index = 0
		}
		return roadmapListLoadedMsg{viewModel: vm, selectedIndex: &index}
	}
}

// findIterationIndex finds the index of an iteration by number in the view model
func (m *AppModelNew) findIterationIndex(vm *viewmodels.RoadmapListViewModel, iterationNumber int) int {
	for i, iter := range vm.ActiveIterations {
		if iter.Number == iterationNumber {
			return i
		}
	}
	return 0 // Default to first item if not found
}

func (m *AppModelNew) loadIterationDetail(iterationNumber int) tea.Cmd {
	return m.loadIterationDetailWithTab(iterationNumber, presenters.IterationDetailTabTasks)
}

func (m *AppModelNew) loadIterationDetailWithTab(iterationNumber int, activeTab presenters.IterationDetailTab) tea.Cmd {
	return m.loadIterationDetailWithTabAndSelection(iterationNumber, activeTab, 0)
}

func (m *AppModelNew) loadIterationDetailWithTabAndSelection(iterationNumber int, activeTab presenters.IterationDetailTab, selectedIndex int) tea.Cmd {
	return func() tea.Msg {
		vm, err := queries.LoadIterationDetailData(m.ctx, m.repo, iterationNumber)
		if err != nil {
			return presenters.ErrorMsg{Err: err}
		}
		// Only include selectedIndex if it's non-zero
		if selectedIndex >= 0 {
			return iterationDetailLoadedMsg{viewModel: vm, activeTab: activeTab, selectedIndex: &selectedIndex}
		}
		return iterationDetailLoadedMsg{viewModel: vm, activeTab: activeTab, selectedIndex: nil}
	}
}

func (m *AppModelNew) loadTaskDetail(taskID string) tea.Cmd {
	return m.loadTaskDetailWithSelection(taskID, 0)
}

func (m *AppModelNew) loadTaskDetailWithSelection(taskID string, selectedIndex int) tea.Cmd {
	return func() tea.Msg {
		vm, err := queries.LoadTaskDetailData(m.ctx, m.repo, taskID)
		if err != nil {
			return presenters.ErrorMsg{Err: err}
		}
		// Only include selectedIndex if it's non-zero
		if selectedIndex >= 0 {
			return taskDetailLoadedMsg{viewModel: vm, selectedIndex: &selectedIndex}
		}
		return taskDetailLoadedMsg{viewModel: vm, selectedIndex: nil}
	}
}

func (m *AppModelNew) loadTrackDetail(trackID string) tea.Cmd {
	return m.loadTrackDetailWithSelection(trackID, 0)
}

func (m *AppModelNew) loadTrackDetailWithSelection(trackID string, selectedIndex int) tea.Cmd {
	return func() tea.Msg {
		vm, err := queries.LoadTrackDetailData(m.ctx, m.repo, trackID)
		if err != nil {
			return presenters.ErrorMsg{Err: err}
		}
		// Only include selectedIndex if it's non-zero
		if selectedIndex >= 0 {
			return trackDetailLoadedMsg{viewModel: vm, selectedIndex: &selectedIndex}
		}
		return trackDetailLoadedMsg{viewModel: vm, selectedIndex: nil}
	}
}

// Custom messages (app-local only)
// Shared message types defined in presenters/messages.go:
// - presenters.ErrorMsg
// - presenters.IterationSelectedMsg
// - presenters.TrackSelectedMsg
// - presenters.TaskSelectedMsg
// - presenters.ACActionCompletedMsg
// - presenters.ReorderCompletedMsg

type roadmapListLoadedMsg struct {
	viewModel     *viewmodels.RoadmapListViewModel
	selectedIndex *int
}

type iterationDetailLoadedMsg struct {
	viewModel     *viewmodels.IterationDetailViewModel
	activeTab     presenters.IterationDetailTab
	selectedIndex *int // Optional: preserve selected index across reload
}

type taskDetailLoadedMsg struct {
	viewModel     *viewmodels.TaskDetailViewModel
	selectedIndex *int // Optional: preserve selected index across reload
}

type trackDetailLoadedMsg struct {
	viewModel     *viewmodels.TrackDetailViewModel
	selectedIndex *int // Optional: preserve selected index across reload
}


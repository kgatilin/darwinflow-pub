package task_manager

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ViewMode represents the current TUI screen
type ViewMode int

// View constants
const (
	ViewRoadmapList   ViewMode = iota
	ViewTrackDetail
	ViewIterationList
	ViewIterationDetail
	ViewADRList
	ViewError
	ViewLoading
)

// AppModel is the main TUI application model
type AppModel struct {
	ctx        context.Context
	repository RoadmapRepository
	logger     pluginsdk.Logger

	// State
	currentView ViewMode
	error       error
	projectName string // Current project name for display

	// Data - Roadmap/Track/Task views
	roadmap     *RoadmapEntity
	tracks      []*TrackEntity
	currentTrack *TrackEntity
	tasks       []*TaskEntity

	// Data - Iteration views
	iterations       []*IterationEntity
	currentIteration *IterationEntity
	iterationTasks   []*TaskEntity

	// Data - ADR views
	adrs             []*ADREntity
	selectedADRIdx   int
	previousViewMode ViewMode // To return to previous view after ADR list

	// UI state
	selectedTrackIdx     int
	selectedTaskIdx      int
	selectedIterationIdx int
	width                int
	height               int

	// Timestamps for debouncing
	lastUpdate time.Time
}

// Message types for Bubble Tea

type RoadmapLoadedMsg struct {
	Roadmap *RoadmapEntity
	Tracks  []*TrackEntity
	Error   error
}

type TrackDetailLoadedMsg struct {
	Track *TrackEntity
	Tasks []*TaskEntity
	Error error
}

type ErrorMsg struct {
	Error error
}

type BackMsg struct{}

type IterationsLoadedMsg struct {
	Iterations []*IterationEntity
	Error      error
}

type IterationDetailLoadedMsg struct {
	Iteration *IterationEntity
	Tasks     []*TaskEntity
	Error     error
}

type ADRsLoadedMsg struct {
	ADRs  []*ADREntity
	Error error
}

// NewAppModel creates a new TUI app model
func NewAppModel(
	ctx context.Context,
	repository RoadmapRepository,
	logger pluginsdk.Logger,
) *AppModel {
	return &AppModel{
		ctx:             ctx,
		repository:      repository,
		logger:          logger,
		currentView:     ViewLoading,
		selectedTrackIdx: 0,
		selectedTaskIdx: 0,
		lastUpdate:      time.Now(),
		projectName:     "default",
	}
}

// NewAppModelWithProject creates a new TUI app model with project name
func NewAppModelWithProject(
	ctx context.Context,
	repository RoadmapRepository,
	logger pluginsdk.Logger,
	projectName string,
) *AppModel {
	return &AppModel{
		ctx:             ctx,
		repository:      repository,
		logger:          logger,
		currentView:     ViewLoading,
		selectedTrackIdx: 0,
		selectedTaskIdx: 0,
		lastUpdate:      time.Now(),
		projectName:     projectName,
	}
}

// Init initializes the application
func (m *AppModel) Init() tea.Cmd {
	return m.loadRoadmap
}

// loadRoadmap is a tea.Cmd that loads the active roadmap
func (m *AppModel) loadRoadmap() tea.Msg {
	roadmap, err := m.repository.GetActiveRoadmap(m.ctx)
	if err != nil {
		return RoadmapLoadedMsg{Error: err}
	}

	tracks, err := m.repository.ListTracks(m.ctx, roadmap.ID, TrackFilters{})
	if err != nil {
		return RoadmapLoadedMsg{Error: err}
	}

	return RoadmapLoadedMsg{
		Roadmap: roadmap,
		Tracks:  tracks,
	}
}

// loadTrackDetail is a tea.Cmd that loads track details and tasks
func (m *AppModel) loadTrackDetail(trackID string) tea.Cmd {
	return func() tea.Msg {
		track, err := m.repository.GetTrack(m.ctx, trackID)
		if err != nil {
			return TrackDetailLoadedMsg{Error: err}
		}

		tasks, err := m.repository.ListTasks(m.ctx, TaskFilters{TrackID: trackID})
		if err != nil {
			return TrackDetailLoadedMsg{Error: err}
		}

		return TrackDetailLoadedMsg{
			Track: track,
			Tasks: tasks,
		}
	}
}

// loadIterations is a tea.Cmd that loads all iterations
func (m *AppModel) loadIterations() tea.Msg {
	iterations, err := m.repository.ListIterations(m.ctx)
	if err != nil {
		return IterationsLoadedMsg{Error: err}
	}

	return IterationsLoadedMsg{
		Iterations: iterations,
	}
}

// loadIterationDetail is a tea.Cmd that loads iteration details and tasks
func (m *AppModel) loadIterationDetail(iterationNum int) tea.Cmd {
	return func() tea.Msg {
		iteration, err := m.repository.GetIteration(m.ctx, iterationNum)
		if err != nil {
			return IterationDetailLoadedMsg{Error: err}
		}

		tasks, err := m.repository.GetIterationTasks(m.ctx, iterationNum)
		if err != nil {
			return IterationDetailLoadedMsg{Error: err}
		}

		return IterationDetailLoadedMsg{
			Iteration: iteration,
			Tasks:     tasks,
		}
	}
}

// Update processes messages and updates state
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			switch m.currentView {
			case ViewTrackDetail:
				m.currentView = ViewRoadmapList
				m.selectedTaskIdx = 0
				return m, nil
			case ViewIterationDetail:
				m.currentView = ViewIterationList
				m.selectedIterationIdx = 0
				return m, nil
			case ViewIterationList:
				m.currentView = ViewRoadmapList
				m.selectedIterationIdx = 0
				return m, nil
			case ViewADRList:
				m.currentView = m.previousViewMode
				m.selectedADRIdx = 0
				return m, nil
			}
		}

		// View-specific key handling
		switch m.currentView {
		case ViewRoadmapList:
			return m.handleRoadmapListKeys(msg)
		case ViewTrackDetail:
			return m.handleTrackDetailKeys(msg)
		case ViewIterationList:
			return m.handleIterationListKeys(msg)
		case ViewIterationDetail:
			return m.handleIterationDetailKeys(msg)
		case ViewADRList:
			return m.handleADRListKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case RoadmapLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.roadmap = msg.Roadmap
		m.tracks = msg.Tracks
		m.currentView = ViewRoadmapList
		m.selectedTrackIdx = 0
		m.lastUpdate = time.Now()

	case TrackDetailLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.currentTrack = msg.Track
		m.tasks = msg.Tasks
		m.currentView = ViewTrackDetail
		m.selectedTaskIdx = 0
		m.lastUpdate = time.Now()

	case IterationsLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.iterations = msg.Iterations
		m.currentView = ViewIterationList
		m.selectedIterationIdx = 0
		m.lastUpdate = time.Now()

	case IterationDetailLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.currentIteration = msg.Iteration
		m.iterationTasks = msg.Tasks
		m.currentView = ViewIterationDetail
		m.lastUpdate = time.Now()

	case ADRsLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.adrs = msg.ADRs
		m.currentView = ViewADRList
		m.selectedADRIdx = 0
		m.lastUpdate = time.Now()

	case ErrorMsg:
		m.currentView = ViewError
		m.error = msg.Error
	}

	return m, nil
}

// handleRoadmapListKeys processes key presses on roadmap list view
func (m *AppModel) handleRoadmapListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selectedTrackIdx < len(m.tracks)-1 {
			m.selectedTrackIdx++
		}
	case "k", "up":
		if m.selectedTrackIdx > 0 {
			m.selectedTrackIdx--
		}
	case "enter":
		if m.selectedTrackIdx < len(m.tracks) {
			trackID := m.tracks[m.selectedTrackIdx].ID
			m.currentView = ViewLoading
			return m, m.loadTrackDetail(trackID)
		}
	case "r":
		m.currentView = ViewLoading
		return m, m.loadRoadmap
	case "i":
		m.currentView = ViewLoading
		return m, m.loadIterations
	}
	return m, nil
}

// handleTrackDetailKeys processes key presses on track detail view
func (m *AppModel) handleTrackDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selectedTaskIdx < len(m.tasks)-1 {
			m.selectedTaskIdx++
		}
	case "k", "up":
		if m.selectedTaskIdx > 0 {
			m.selectedTaskIdx--
		}
	case "a":
		// View ADRs for the current track
		if m.currentTrack != nil {
			m.previousViewMode = ViewTrackDetail
			m.currentView = ViewLoading
			return m, m.loadADRs(m.currentTrack.ID)
		}
	}
	return m, nil
}

// handleIterationListKeys processes key presses on iteration list view
func (m *AppModel) handleIterationListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selectedIterationIdx < len(m.iterations)-1 {
			m.selectedIterationIdx++
		}
	case "k", "up":
		if m.selectedIterationIdx > 0 {
			m.selectedIterationIdx--
		}
	case "enter":
		if m.selectedIterationIdx < len(m.iterations) {
			iterNum := m.iterations[m.selectedIterationIdx].Number
			m.currentView = ViewLoading
			return m, m.loadIterationDetail(iterNum)
		}
	case "r":
		m.currentView = ViewLoading
		return m, m.loadIterations
	}
	return m, nil
}

// handleIterationDetailKeys processes key presses on iteration detail view
func (m *AppModel) handleIterationDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Iteration detail is read-only, navigation handled by esc in main Update
	return m, nil
}

// View renders the current view
func (m *AppModel) View() string {
	switch m.currentView {
	case ViewLoading:
		return m.renderLoading()
	case ViewError:
		return m.renderError()
	case ViewRoadmapList:
		return m.renderRoadmapList()
	case ViewTrackDetail:
		return m.renderTrackDetail()
	case ViewIterationList:
		return m.renderIterationList()
	case ViewIterationDetail:
		return m.renderIterationDetail()
	case ViewADRList:
		return m.renderADRList()
	default:
		return "Unknown view"
	}
}

// renderLoading renders the loading screen
func (m *AppModel) renderLoading() string {
	return "Loading...\n\nPress q to quit"
}

// renderError renders the error screen
func (m *AppModel) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196"))

	return errorStyle.Render(fmt.Sprintf("Error: %v", m.error)) + "\n\nPress esc to go back"
}

// renderRoadmapList renders the roadmap overview screen
func (m *AppModel) renderRoadmapList() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	trackItemStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0)

	selectedTrackStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("229"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	// Project header
	s += titleStyle.Render(fmt.Sprintf("Project: %s", m.projectName)) + "\n\n"

	// Header
	if m.roadmap != nil {
		s += fmt.Sprintf("Roadmap: %s\n", m.roadmap.ID)
		s += fmt.Sprintf("Vision: %s\n", m.roadmap.Vision)
		s += fmt.Sprintf("Success Criteria: %s\n", m.roadmap.SuccessCriteria)
		s += "\n"
	}

	// Tracks
	s += "Tracks:\n"
	if len(m.tracks) == 0 {
		s += "  No tracks yet\n"
	} else {
		for i, track := range m.tracks {
			statusIcon := getStatusIcon(track.Status)
			priorityIcon := getPriorityIcon(track.Priority)

			line := fmt.Sprintf("%s %s %s - %s", statusIcon, priorityIcon, track.ID, track.Title)

			if i == m.selectedTrackIdx {
				s += selectedTrackStyle.Render(line) + "\n"
			} else {
				s += trackItemStyle.Render(line) + "\n"
			}
		}
	}

	// Help text
	s += "\n"
	s += helpStyle.Render("Navigation: j/k or ↑/↓ | Enter: View track | r: Refresh | q: Quit")

	return s
}

// renderTrackDetail renders the track detail view
func (m *AppModel) renderTrackDetail() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Italic(true).
		MarginBottom(1)

	taskItemStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0)

	selectedTaskStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("229"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	if m.currentTrack == nil {
		return "Loading track details..."
	}

	// Header
	s += titleStyle.Render(fmt.Sprintf("Track: %s", m.currentTrack.ID)) + "\n"
	s += fmt.Sprintf("Title: %s\n", m.currentTrack.Title)
	s += fmt.Sprintf("Description: %s\n", m.currentTrack.Description)
	s += fmt.Sprintf("Status: %s | Priority: %s\n", m.currentTrack.Status, m.currentTrack.Priority)

	// ADR Status
	if adrs, err := m.repository.ListADRs(m.ctx, &m.currentTrack.ID); err == nil {
		if len(adrs) > 0 {
			acceptedCount := 0
			for _, adr := range adrs {
				if adr.Status == string(ADRStatusAccepted) {
					acceptedCount++
				}
			}
			proposedCount := len(adrs) - acceptedCount
			s += fmt.Sprintf("ADRs: %d (%d accepted, %d proposed) [press 'a' to view]\n", len(adrs), acceptedCount, proposedCount)
		} else {
			s += "ADRs: None (required for task completion)\n"
		}
	}

	// Dependencies
	if len(m.currentTrack.Dependencies) > 0 {
		s += "\nDependencies:\n"
		for _, dep := range m.currentTrack.Dependencies {
			s += fmt.Sprintf("  → %s\n", dep)
		}
	}

	// Tasks
	s += fmt.Sprintf("\nTasks (%d):\n", len(m.tasks))
	if len(m.tasks) == 0 {
		s += subtitleStyle.Render("No tasks yet")
	} else {
		for i, task := range m.tasks {
			statusIcon := getStatusIcon(task.Status)
			priorityIcon := getPriorityIcon(task.Priority)

			line := fmt.Sprintf("%s %s %s - %s", statusIcon, priorityIcon, task.ID, task.Title)

			if i == m.selectedTaskIdx {
				s += selectedTaskStyle.Render(line) + "\n"
			} else {
				s += taskItemStyle.Render(line) + "\n"
			}
		}
	}

	// Help
	s += "\n"
	s += helpStyle.Render("Navigation: j/k or ↑/↓ | a: ADRs | esc: Back | q: Quit")

	return s
}

// renderIterationList renders the iteration list view
func (m *AppModel) renderIterationList() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	iterationItemStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0)

	selectedIterationStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("229"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	// Header
	s += titleStyle.Render("Iterations") + "\n"

	// Table header
	if len(m.iterations) > 0 {
		s += fmt.Sprintf("%-5s %-30s %-15s %-12s %-10s\n",
			"#", "Name", "Goal", "Status", "Tasks")
		s += strings.Repeat("-", 80) + "\n"
	}

	if len(m.iterations) == 0 {
		s += "  No iterations yet. Create one with: dw task-manager iteration create\n"
	} else {
		for i, iter := range m.iterations {
			status := m.formatIterationStatus(iter.Status)
			taskCount := len(iter.TaskIDs)

			line := fmt.Sprintf("%-5d %-30s %-15s %-12s %-10d",
				iter.Number,
				truncateString(iter.Name, 30),
				truncateString(iter.Goal, 15),
				status,
				taskCount)

			if i == m.selectedIterationIdx {
				s += selectedIterationStyle.Render("→ " + line) + "\n"
			} else {
				s += iterationItemStyle.Render("  " + line) + "\n"
			}
		}
	}

	// Help
	s += "\n"
	s += helpStyle.Render("Navigation: j/k or ↑/↓ | Enter: View iteration | r: Refresh | esc: Back | q: Quit")

	return s
}

// renderIterationDetail renders the iteration detail view
func (m *AppModel) renderIterationDetail() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Italic(true).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	if m.currentIteration == nil {
		return "Loading iteration details..."
	}

	// Header
	s += titleStyle.Render(fmt.Sprintf("Iteration %d: %s", m.currentIteration.Number, m.currentIteration.Name)) + "\n"
	s += "\n"

	// Details
	s += fmt.Sprintf("Goal: %s\n", m.currentIteration.Goal)
	s += fmt.Sprintf("Status: %s\n", m.formatIterationStatus(m.currentIteration.Status))
	if m.currentIteration.Deliverable != "" {
		s += fmt.Sprintf("Deliverable: %s\n", m.currentIteration.Deliverable)
	}

	// Timestamps
	if m.currentIteration.StartedAt != nil {
		s += fmt.Sprintf("Started: %s\n", m.currentIteration.StartedAt.Format("2006-01-02"))
	}
	if m.currentIteration.CompletedAt != nil {
		s += fmt.Sprintf("Completed: %s\n", m.currentIteration.CompletedAt.Format("2006-01-02"))
	}
	s += "\n"

	// Tasks in iteration
	s += fmt.Sprintf("Tasks (%d):\n", len(m.iterationTasks))
	if len(m.iterationTasks) == 0 {
		s += subtitleStyle.Render("  No tasks in this iteration.") + "\n"
	} else {
		// Group tasks by status
		todoTasks := []*TaskEntity{}
		inProgressTasks := []*TaskEntity{}
		doneTasks := []*TaskEntity{}

		for _, task := range m.iterationTasks {
			switch task.Status {
			case "todo":
				todoTasks = append(todoTasks, task)
			case "in-progress":
				inProgressTasks = append(inProgressTasks, task)
			case "done":
				doneTasks = append(doneTasks, task)
			}
		}

		// Progress bar
		total := len(m.iterationTasks)
		done := len(doneTasks)
		progress := 0.0
		if total > 0 {
			progress = float64(done) / float64(total) * 100.0
		}
		s += fmt.Sprintf("Progress: %d/%d (%.1f%%)\n", done, total, progress)
		s += m.renderProgressBar(progress, 40)
		s += "\n\n"

		// Task breakdown
		if len(todoTasks) > 0 {
			s += "Todo:\n"
			for _, task := range todoTasks {
				s += fmt.Sprintf("  ○ %s\n", task.Title)
			}
		}

		if len(inProgressTasks) > 0 {
			s += "In Progress:\n"
			for _, task := range inProgressTasks {
				s += fmt.Sprintf("  → %s\n", task.Title)
			}
		}

		if len(doneTasks) > 0 {
			s += "Done:\n"
			for _, task := range doneTasks {
				s += fmt.Sprintf("  ✓ %s\n", task.Title)
			}
		}
	}

	// Help
	s += "\n"
	s += helpStyle.Render("Navigation: esc: Back to iteration list | q: Quit")

	return s
}

// renderProgressBar renders a simple progress bar
func (m *AppModel) renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100.0 * float64(width))
	empty := width - filled
	if empty < 0 {
		empty = 0
	}
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s] %.1f%%", bar, percent)
}

// formatIterationStatus formats iteration status for display
func (m *AppModel) formatIterationStatus(status string) string {
	switch status {
	case "current":
		return "→ Current"
	case "complete":
		return "✓ Complete"
	case "planned":
		return "○ Planned"
	default:
		return status
	}
}

// Helper functions for rendering

func getStatusIcon(status string) string {
	switch status {
	case "done", "complete":
		return "✓"
	case "in-progress":
		return "→"
	case "blocked":
		return "✗"
	case "waiting":
		return "⏸"
	default:
		return "○"
	}
}

func getPriorityIcon(priority string) string {
	switch priority {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}

// Test helper methods - exported for testing

// SetRoadmap sets the roadmap for testing
func (m *AppModel) SetRoadmap(roadmap *RoadmapEntity) {
	m.roadmap = roadmap
}

// SetTracks sets the tracks for testing
func (m *AppModel) SetTracks(tracks []*TrackEntity) {
	m.tracks = tracks
}

// SetTasks sets the tasks for testing
func (m *AppModel) SetTasks(tasks []*TaskEntity) {
	m.tasks = tasks
}

// SetCurrentTrack sets the current track for testing
func (m *AppModel) SetCurrentTrack(track *TrackEntity) {
	m.currentTrack = track
}

// SetCurrentView sets the current view for testing
func (m *AppModel) SetCurrentView(view ViewMode) {
	m.currentView = view
}

// SetError sets the error for testing
func (m *AppModel) SetError(err error) {
	m.error = err
}

// SetDimensions sets the width and height for testing
func (m *AppModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// SetIterations sets the iterations for testing
func (m *AppModel) SetIterations(iterations []*IterationEntity) {
	m.iterations = iterations
}

// SetCurrentIteration sets the current iteration for testing
func (m *AppModel) SetCurrentIteration(iteration *IterationEntity) {
	m.currentIteration = iteration
}

// SetIterationTasks sets the iteration tasks for testing
func (m *AppModel) SetIterationTasks(tasks []*TaskEntity) {
	m.iterationTasks = tasks
}

// SetSelectedIterationIdx sets the selected iteration index for testing
func (m *AppModel) SetSelectedIterationIdx(idx int) {
	m.selectedIterationIdx = idx
}

// RenderProgressBar is a public wrapper for testing progress bar rendering
func (m *AppModel) RenderProgressBar(percent float64, width int) string {
	return m.renderProgressBar(percent, width)
}

// GetCurrentView returns the current view mode for testing
func (m *AppModel) GetCurrentView() ViewMode {
	return m.currentView
}

// GetSelectedIterationIdx returns the selected iteration index for testing
func (m *AppModel) GetSelectedIterationIdx() int {
	return m.selectedIterationIdx
}

// renderADRList renders the ADR list view
func (m *AppModel) renderADRList() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	adrItemStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0)

	selectedADRStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("229"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	// Header
	s += titleStyle.Render(fmt.Sprintf("ADRs for Track: %s", m.currentTrack.ID)) + "\n\n"

	// ADRs
	if len(m.adrs) == 0 {
		s += "No ADRs yet. Create one with:\n"
		s += fmt.Sprintf("  dw task-manager adr create %s --title \"...\" --context \"...\" --decision \"...\"\n", m.currentTrack.ID)
	} else {
		for i, adr := range m.adrs {
			statusIcon := getADRStatusIcon(adr.Status)
			line := fmt.Sprintf("%s [%s] %s", statusIcon, adr.ID, adr.Title)

			if i == m.selectedADRIdx {
				s += selectedADRStyle.Render(line) + "\n"
			} else {
				s += adrItemStyle.Render(line) + "\n"
			}
		}
	}

	// Help
	s += "\n"
	s += helpStyle.Render("Navigation: j/k or ↑/↓ | esc: Back | q: Quit")

	return s
}

// handleADRListKeys processes key presses on ADR list view
func (m *AppModel) handleADRListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selectedADRIdx < len(m.adrs)-1 {
			m.selectedADRIdx++
		}
	case "k", "up":
		if m.selectedADRIdx > 0 {
			m.selectedADRIdx--
		}
	}
	return m, nil
}

// loadADRs loads ADRs for a track
func (m *AppModel) loadADRs(trackID string) tea.Cmd {
	return func() tea.Msg {
		adrs, err := m.repository.ListADRs(m.ctx, &trackID)
		if err != nil {
			return ADRsLoadedMsg{
				ADRs:  nil,
				Error: err,
			}
		}
		return ADRsLoadedMsg{
			ADRs:  adrs,
			Error: nil,
		}
	}
}

// getADRStatusIcon returns a visual icon for ADR status
func getADRStatusIcon(status string) string {
	switch status {
	case string(ADRStatusAccepted):
		return "✓"
	case string(ADRStatusProposed):
		return "○"
	case string(ADRStatusDeprecated):
		return "✗"
	case string(ADRStatusSuperseded):
		return "⇒"
	default:
		return "?"
	}
}

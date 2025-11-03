package task_manager

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ViewMode represents the current TUI screen
type ViewMode int

// View constants
const (
	ViewRoadmapList   ViewMode = iota
	ViewTrackDetail
	ViewTaskDetail
	ViewIterationList
	ViewIterationDetail
	ViewADRList
	ViewACList
	ViewACFailInput // New view for capturing failure feedback
	ViewError
	ViewLoading
)

// ItemSelectionType determines which list is active for selection in the roadmap view
type ItemSelectionType int

const (
	SelectTracks ItemSelectionType = iota
	SelectIterations
	SelectBacklog
)

// Hotkey represents a single keyboard shortcut
type Hotkey struct {
	Keys        []string // Key combinations (e.g., ["j", "↓"])
	Description string   // What the hotkey does
}

// HotkeyGroup represents a category of related hotkeys
type HotkeyGroup struct {
	Name    string
	Hotkeys []Hotkey
}

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
	currentTask *TaskEntity

	// Data - Iteration views
	iterations       []*IterationEntity
	currentIteration *IterationEntity
	iterationTasks   []*TaskEntity

	// Data - ADR views
	adrs             []*ADREntity
	selectedADRIdx   int

	// Data - AC views
	acs              []*AcceptanceCriteriaEntity
	selectedACIdx    int

	// Input - for capturing AC failure feedback
	feedbackInput    textinput.Model

	previousViewMode ViewMode // To return to previous view after ADR/AC list

	// UI state
	selectedTrackIdx     int
	selectedTaskIdx      int
	selectedIterationIdx int
	selectedBacklogIdx       int
	selectedIterationTaskIdx int // For navigating tasks in iteration detail view
	selectedIterationACIdx   int // For navigating ACs in iteration detail view
	iterationDetailFocusAC   bool // True = focus on ACs, False = focus on tasks
	selectedItemType     ItemSelectionType // Tracks which list is active for selection (tracks vs iterations vs backlog)
	width                int
	height               int

	// Viewport for scrolling
	roadmapViewport viewport.Model

	// Section line positions (for auto-scrolling on Tab)
	iterationsSectionLine int
	tracksSectionLine     int
	backlogSectionLine    int

	// Track which iteration is being reordered (to maintain selection after reload)
	reorderingIterationNumber int

	// View mode toggles
	showFullRoadmap      bool // Toggle between tracks-only and full roadmap view
	showCompletedTracks  bool // Toggle showing completed tracks
	showCompletedIters   bool // Toggle showing completed iterations

	// Full roadmap data
	iterationTasksByNumber map[int][]*TaskEntity   // Tasks for each iteration
	trackTasksByID         map[string][]*TaskEntity // Tasks for each track
	backlogTasks           []*TaskEntity            // Tasks not in any iteration/track

	// Timestamps for debouncing
	lastUpdate time.Time
}

// Message types for Bubble Tea

type RoadmapLoadedMsg struct {
	Roadmap    *RoadmapEntity
	Tracks     []*TrackEntity
	Iterations []*IterationEntity
	Error      error
}

type TrackDetailLoadedMsg struct {
	Track *TrackEntity
	Tasks []*TaskEntity
	Error error
}

type TaskDetailLoadedMsg struct {
	Task *TaskEntity
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
	ACs       []*AcceptanceCriteriaEntity
	Error     error
}

type ADRsLoadedMsg struct {
	ADRs  []*ADREntity
	Error error
}

type ACsLoadedMsg struct {
	ACs   []*AcceptanceCriteriaEntity
	Error error
}

type FullRoadmapDataLoadedMsg struct {
	IterationTasks map[int][]*TaskEntity
	TrackTasks     map[string][]*TaskEntity
	BacklogTasks   []*TaskEntity
	Error          error
}

// NewAppModel creates a new TUI app model
func NewAppModel(
	ctx context.Context,
	repository RoadmapRepository,
	logger pluginsdk.Logger,
) *AppModel {
	// Initialize text input for AC failure feedback
	ti := textinput.New()
	ti.Placeholder = "Enter failure feedback..."
	ti.CharLimit = 500
	ti.Width = 80

	m := &AppModel{
		ctx:             ctx,
		repository:      repository,
		logger:          logger,
		currentView:     ViewLoading,
		selectedTrackIdx: 0,
		selectedTaskIdx: 0,
		lastUpdate:      time.Now(),
		projectName:     "default",
		selectedItemType:        SelectIterations,
		iterationTasksByNumber:  make(map[int][]*TaskEntity),
		trackTasksByID:          make(map[string][]*TaskEntity),
		backlogTasks:            make([]*TaskEntity, 0),
		feedbackInput:           ti,
	}
	m.roadmapViewport = viewport.New(80, 24) // Default size, will be updated on window resize
	return m
}

// NewAppModelWithProject creates a new TUI app model with project name
func NewAppModelWithProject(
	ctx context.Context,
	repository RoadmapRepository,
	logger pluginsdk.Logger,
	projectName string,
) *AppModel {
	// Initialize text input for AC failure feedback
	ti := textinput.New()
	ti.Placeholder = "Enter failure feedback..."
	ti.CharLimit = 500
	ti.Width = 80

	m := &AppModel{
		ctx:             ctx,
		repository:      repository,
		logger:          logger,
		currentView:     ViewLoading,
		selectedTrackIdx: 0,
		selectedTaskIdx: 0,
		lastUpdate:      time.Now(),
		projectName:     projectName,
		selectedItemType:        SelectIterations,
		iterationTasksByNumber:  make(map[int][]*TaskEntity),
		trackTasksByID:          make(map[string][]*TaskEntity),
		backlogTasks:            make([]*TaskEntity, 0),
		feedbackInput:           ti,
	}
	m.roadmapViewport = viewport.New(80, 24) // Default size, will be updated on window resize
	return m
}

// getRoadmapListHotkeys returns the hotkey groups for the roadmap list view
func (m AppModel) getRoadmapListHotkeys() []HotkeyGroup {
	// Dynamic description based on current selection mode
	var navTarget string
	switch m.selectedItemType {
	case SelectTracks:
		navTarget = "Tracks"
	case SelectIterations:
		navTarget = "Iterations"
	case SelectBacklog:
		navTarget = "Backlog"
	}

	// Dynamic view mode description
	var viewMode string
	if m.showFullRoadmap {
		viewMode = "Hide Details"
	} else {
		viewMode = "Show Details"
	}

	groups := []HotkeyGroup{
		{
			Name: "Navigation",
			Hotkeys: []Hotkey{
				{Keys: []string{"j", "k", "↑", "↓"}, Description: fmt.Sprintf("Select %s", navTarget)},
				{Keys: []string{"Ctrl+j", "Ctrl+k"}, Description: "Scroll View"},
				{Keys: []string{"Tab"}, Description: "Switch Selection Mode"},
				{Keys: []string{"Enter"}, Description: "View Details"},
			},
		},
		{
			Name: "Views",
			Hotkeys: []Hotkey{
				{Keys: []string{"i"}, Description: "Iterations View"},
				{Keys: []string{"v"}, Description: viewMode},
				{Keys: []string{"t"}, Description: "Toggle Completed Tracks"},
				{Keys: []string{"c"}, Description: "Toggle Completed Iterations"},
			},
		},
		{
			Name: "General",
			Hotkeys: []Hotkey{
				{Keys: []string{"r"}, Description: "Refresh"},
				{Keys: []string{"q", "Ctrl+C"}, Description: "Quit"},
			},
		},
	}

	// Add reordering hotkeys when iterations are selected
	if m.selectedItemType == SelectIterations {
		groups[0].Hotkeys = append(groups[0].Hotkeys,
			Hotkey{Keys: []string{"Shift+K", "Shift+J"}, Description: "Reorder Iteration"},
		)
	}

	return groups
}

// formatHotkeyGroups formats hotkey groups for display, with multi-line wrapping
func formatHotkeyGroups(groups []HotkeyGroup, width int, style lipgloss.Style) string {
	// Format all hotkeys first
	var formatted []string
	for _, group := range groups {
		for _, hotkey := range group.Hotkeys {
			// Combine keys with "/"
			keys := strings.Join(hotkey.Keys, "/")
			// Format as "keys:description"
			formatted = append(formatted, fmt.Sprintf("%s:%s", keys, hotkey.Description))
		}
	}

	// Join with " | " separator
	separator := " | "

	// Build output with wrapping
	var lines []string
	var currentLine strings.Builder

	for i, item := range formatted {
		// Check if adding this item would exceed width
		testLine := currentLine.String()
		if i > 0 && len(testLine) > 0 {
			testLine += separator
		}
		testLine += item

		// If first item or fits on current line, add it
		if currentLine.Len() == 0 || len(testLine) <= width {
			if currentLine.Len() > 0 {
				currentLine.WriteString(separator)
			}
			currentLine.WriteString(item)
		} else {
			// Start new line
			lines = append(lines, style.Render(currentLine.String()))
			currentLine.Reset()
			currentLine.WriteString(item)
		}
	}

	// Add final line
	if currentLine.Len() > 0 {
		lines = append(lines, style.Render(currentLine.String()))
	}

	return strings.Join(lines, "\n")
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

	// Also load iterations for display on main view
	iterations, err := m.repository.ListIterations(m.ctx)
	if err != nil {
		// Don't fail if iterations can't be loaded
		iterations = []*IterationEntity{}
	}

	return RoadmapLoadedMsg{
		Roadmap:    roadmap,
		Tracks:     tracks,
		Iterations: iterations,
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

		acs, err := m.repository.ListACByIteration(m.ctx, iterationNum)
		if err != nil {
			return IterationDetailLoadedMsg{Error: err}
		}

		return IterationDetailLoadedMsg{
			Iteration: iteration,
			Tasks:     tasks,
			ACs:       acs,
		}
	}
}

// loadFullRoadmapData loads all tasks for iterations, tracks, and backlog
func (m *AppModel) loadFullRoadmapData() tea.Cmd {
	return func() tea.Msg {
		// Load all tasks
		allTasks, err := m.repository.ListTasks(m.ctx, TaskFilters{})
		if err != nil {
			return FullRoadmapDataLoadedMsg{Error: err}
		}

		// Create a map of tasks by ID for quick lookup
		tasksByID := make(map[string]*TaskEntity)
		for _, task := range allTasks {
			tasksByID[task.ID] = task
		}

		// Initialize maps
		iterationTasks := make(map[int][]*TaskEntity)
		trackTasks := make(map[string][]*TaskEntity)
		assignedTaskIDs := make(map[string]bool) // Track which tasks are assigned to iterations
		var backlogTasks []*TaskEntity

		// Group tasks by iteration (from iteration's task IDs)
		for _, iter := range m.iterations {
			for _, taskID := range iter.TaskIDs {
				if task, exists := tasksByID[taskID]; exists {
					iterationTasks[iter.Number] = append(iterationTasks[iter.Number], task)
					assignedTaskIDs[taskID] = true
				}
			}
		}

		// Group tasks by track
		for _, task := range allTasks {
			trackTasks[task.TrackID] = append(trackTasks[task.TrackID], task)
		}

		// Collect backlog tasks (tasks not assigned to any iteration, excluding 'done' status)
		for _, task := range allTasks {
			if !assignedTaskIDs[task.ID] && task.Status != "done" {
				backlogTasks = append(backlogTasks, task)
			}
		}

		return FullRoadmapDataLoadedMsg{
			IterationTasks: iterationTasks,
			TrackTasks:     trackTasks,
			BacklogTasks:   backlogTasks,
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
			case ViewTaskDetail:
				m.currentView = m.previousViewMode
				return m, nil
			case ViewTrackDetail:
				m.currentView = ViewRoadmapList
				m.selectedTaskIdx = 0
				return m, nil
			case ViewIterationDetail:
				m.currentView = ViewRoadmapList
				// Keep selectedIterationIdx to preserve selection
				return m, nil
			case ViewIterationList:
				m.currentView = ViewRoadmapList
				m.selectedIterationIdx = 0
				return m, nil
			case ViewADRList:
				m.currentView = m.previousViewMode
				m.selectedADRIdx = 0
				return m, nil
			case ViewACList:
				m.currentView = m.previousViewMode
				m.selectedACIdx = 0
				return m, nil
			case ViewACFailInput:
				// Cancel input, return to AC list
				m.feedbackInput.SetValue("")
				m.currentView = ViewACList
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
		case ViewACList:
			return m.handleACListKeys(msg)
		case ViewACFailInput:
			return m.handleACFailInputKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update viewport size (leave room for help text at bottom)
		m.roadmapViewport.Width = msg.Width
		m.roadmapViewport.Height = msg.Height - 3 // Reserve 3 lines for help text
		return m, nil

	case RoadmapLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.roadmap = msg.Roadmap
		m.tracks = msg.Tracks
		m.iterations = msg.Iterations
		m.currentView = ViewRoadmapList
		m.selectedTrackIdx = 0
		m.lastUpdate = time.Now()

		// If we just reordered an iteration, find it and update selection
		if m.reorderingIterationNumber != 0 {
			// Filter to active iterations only (same as in reordering logic)
			activeIterations := []*IterationEntity{}
			for _, iter := range m.iterations {
				if iter.Status != "complete" {
					activeIterations = append(activeIterations, iter)
				}
			}

			// Find the reordered iteration in the new list
			for idx, iter := range activeIterations {
				if iter.Number == m.reorderingIterationNumber {
					m.selectedIterationIdx = idx
					break
				}
			}

			// Reset the flag
			m.reorderingIterationNumber = 0
		}

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

	case TaskDetailLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.currentTask = msg.Task
		m.currentView = ViewTaskDetail
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
		m.acs = msg.ACs // Store ACs for iteration view
		// Sort tasks by status to match display order (todo, in-progress, done)
		// This ensures the selection index matches the visual order
		sort.SliceStable(m.iterationTasks, func(i, j int) bool {
			statusOrder := map[string]int{"todo": 0, "in-progress": 1, "done": 2}
			return statusOrder[m.iterationTasks[i].Status] < statusOrder[m.iterationTasks[j].Status]
		})
		m.selectedIterationTaskIdx = 0 // Reset task selection
		m.selectedIterationACIdx = 0    // Reset AC selection
		m.iterationDetailFocusAC = false // Start with tasks focused
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

	case ACsLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
			return m, nil
		}
		m.acs = msg.ACs
		m.currentView = ViewACList
		m.selectedACIdx = 0
		m.lastUpdate = time.Now()

	case FullRoadmapDataLoadedMsg:
		if msg.Error != nil {
			m.currentView = ViewError
			m.error = msg.Error
		} else {
			m.iterationTasksByNumber = msg.IterationTasks
			m.trackTasksByID = msg.TrackTasks
			m.backlogTasks = msg.BacklogTasks
			m.currentView = ViewRoadmapList
		}
		return m, nil

	case ErrorMsg:
		m.currentView = ViewError
		m.error = msg.Error
	}

	return m, nil
}

// handleRoadmapListKeys processes key presses on roadmap list view
func (m *AppModel) handleRoadmapListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	// Tab: Cycle through selection modes (Iterations → Tracks → Backlog)
	case "tab":
		switch m.selectedItemType {
		case SelectIterations:
			m.selectedItemType = SelectTracks
			m.selectedTrackIdx = 0
			// Scroll to tracks section
			m.roadmapViewport.SetYOffset(m.tracksSectionLine)
		case SelectTracks:
			// Only show backlog option if backlog has items
			if len(m.backlogTasks) > 0 {
				m.selectedItemType = SelectBacklog
				m.selectedBacklogIdx = 0
				// Scroll to backlog section
				m.roadmapViewport.SetYOffset(m.backlogSectionLine)
			} else {
				m.selectedItemType = SelectIterations
				m.selectedIterationIdx = 0
				// Scroll to top (iterations)
				m.roadmapViewport.SetYOffset(m.iterationsSectionLine)
			}
		case SelectBacklog:
			m.selectedItemType = SelectIterations
			m.selectedIterationIdx = 0
			// Scroll to top (iterations)
			m.roadmapViewport.SetYOffset(m.iterationsSectionLine)
		}

	// Navigation: j/k or arrow keys
	case "j", "down":
		switch m.selectedItemType {
		case SelectTracks:
			// Navigate tracks
			activeTracks := []*TrackEntity{}
			for _, track := range m.tracks {
				if track.Status != "complete" && track.Status != "done" {
					activeTracks = append(activeTracks, track)
				}
			}
			if len(activeTracks) > 0 && m.selectedTrackIdx < len(activeTracks)-1 {
				m.selectedTrackIdx++
			}
		case SelectIterations:
			// Navigate iterations
			activeIterations := []*IterationEntity{}
			for _, iter := range m.iterations {
				if iter.Status != "complete" {
					activeIterations = append(activeIterations, iter)
				}
			}
			if len(activeIterations) > 0 && m.selectedIterationIdx < len(activeIterations)-1 {
				m.selectedIterationIdx++
			}
		case SelectBacklog:
			// Navigate backlog
			if len(m.backlogTasks) > 0 && m.selectedBacklogIdx < len(m.backlogTasks)-1 {
				m.selectedBacklogIdx++
			}
		}

	case "k", "up":
		switch m.selectedItemType {
		case SelectTracks:
			// Navigate tracks
			if m.selectedTrackIdx > 0 {
				m.selectedTrackIdx--
			}
		case SelectIterations:
			// Navigate iterations
			if m.selectedIterationIdx > 0 {
				m.selectedIterationIdx--
			}
		case SelectBacklog:
			// Navigate backlog
			if m.selectedBacklogIdx > 0 {
				m.selectedBacklogIdx--
			}
		}

	// Enter: View details of selected item
	case "enter":
		switch m.selectedItemType {
		case SelectTracks:
			// View track detail
			activeTracks := []*TrackEntity{}
			for _, track := range m.tracks {
				if track.Status != "complete" && track.Status != "done" {
					activeTracks = append(activeTracks, track)
				}
			}
			if len(activeTracks) > 0 && m.selectedTrackIdx < len(activeTracks) {
				return m, m.loadTrackDetail(activeTracks[m.selectedTrackIdx].ID)
			}
		case SelectIterations:
			// View iteration detail
			activeIterations := []*IterationEntity{}
			for _, iter := range m.iterations {
				if iter.Status != "complete" {
					activeIterations = append(activeIterations, iter)
				}
			}
			if len(activeIterations) > 0 && m.selectedIterationIdx < len(activeIterations) {
				return m, m.loadIterationDetail(activeIterations[m.selectedIterationIdx].Number)
			}
		case SelectBacklog:
			// View backlog task detail
			if len(m.backlogTasks) > 0 && m.selectedBacklogIdx < len(m.backlogTasks) {
				m.previousViewMode = ViewRoadmapList
				return m, m.loadTaskDetail(m.backlogTasks[m.selectedBacklogIdx].ID)
			}
		}

	// Refresh roadmap data
	case "r":
		return m, func() tea.Msg {
			return m.loadRoadmap()
		}

	// Switch to iterations list view
	case "i":
		m.currentView = ViewIterationList
		return m, nil

	// Toggle full roadmap view (vision/criteria)
	case "v":
		m.showFullRoadmap = !m.showFullRoadmap
		// If turning on full view and data not loaded, load it
		if m.showFullRoadmap && len(m.iterationTasksByNumber) == 0 && len(m.trackTasksByID) == 0 {
			m.currentView = ViewLoading
			return m, m.loadFullRoadmapData()
		}

	// Toggle completed tracks
	case "t":
		m.showCompletedTracks = !m.showCompletedTracks

	// Toggle completed iterations
	case "c":
		m.showCompletedIters = !m.showCompletedIters

	// Viewport scrolling
	case "pgdown":
		m.roadmapViewport.PageDown()
	case "pgup":
		m.roadmapViewport.PageUp()
	case "home":
		m.roadmapViewport.GotoTop()
	case "end":
		m.roadmapViewport.GotoBottom()
	case "ctrl+j":
		m.roadmapViewport.ScrollDown(3)
	case "ctrl+k":
		m.roadmapViewport.ScrollUp(3)

	// Iteration reordering with shift+up/down (also support J/K for reordering)
	case "shift+down", "J":
		if m.selectedItemType == SelectIterations {
			// Get active iterations
			activeIterations := []*IterationEntity{}
			for _, iter := range m.iterations {
				if iter.Status != "complete" {
					activeIterations = append(activeIterations, iter)
				}
			}

			// Move iteration down (increase rank)
			if len(activeIterations) > 0 && m.selectedIterationIdx < len(activeIterations)-1 {
				currentIter := activeIterations[m.selectedIterationIdx]
				nextIter := activeIterations[m.selectedIterationIdx+1]

				// Remember which iteration we're moving so we can find it after reload
				m.reorderingIterationNumber = currentIter.Number

				// Swap ranks - if equal, make them different first
				if currentIter.Rank == nextIter.Rank {
					// Current moves down, so increase its rank
					currentIter.Rank = currentIter.Rank + 1
				} else {
					// Normal swap
					currentIter.Rank, nextIter.Rank = nextIter.Rank, currentIter.Rank
				}
				currentIter.UpdatedAt = time.Now()
				nextIter.UpdatedAt = time.Now()

				// Update both iterations in repository
				if err := m.repository.UpdateIteration(m.ctx, currentIter); err == nil {
					if err := m.repository.UpdateIteration(m.ctx, nextIter); err == nil {
						// Reload roadmap to reflect changes
						// Selection will be updated when RoadmapLoadedMsg is received
						return m, func() tea.Msg {
							return m.loadRoadmap()
						}
					}
				}
			}
		}

	case "shift+up", "K":
		if m.selectedItemType == SelectIterations {
			// Get active iterations
			activeIterations := []*IterationEntity{}
			for _, iter := range m.iterations {
				if iter.Status != "complete" {
					activeIterations = append(activeIterations, iter)
				}
			}

			// Move iteration up (decrease rank)
			if len(activeIterations) > 0 && m.selectedIterationIdx > 0 {
				currentIter := activeIterations[m.selectedIterationIdx]
				prevIter := activeIterations[m.selectedIterationIdx-1]

				// Remember which iteration we're moving so we can find it after reload
				m.reorderingIterationNumber = currentIter.Number

				// Swap ranks - if equal, make them different first
				if currentIter.Rank == prevIter.Rank {
					// Current moves up, so decrease its rank
					currentIter.Rank = currentIter.Rank - 1
				} else {
					// Normal swap
					currentIter.Rank, prevIter.Rank = prevIter.Rank, currentIter.Rank
				}
				currentIter.UpdatedAt = time.Now()
				prevIter.UpdatedAt = time.Now()

				// Update both iterations in repository
				if err := m.repository.UpdateIteration(m.ctx, currentIter); err == nil {
					if err := m.repository.UpdateIteration(m.ctx, prevIter); err == nil {
						// Reload roadmap to reflect changes
						// Selection will be updated when RoadmapLoadedMsg is received
						return m, func() tea.Msg {
							return m.loadRoadmap()
						}
					}
				}
			}
		}
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
	case " ": // Space bar - verify selected AC
		if m.selectedACIdx < len(m.acs) {
			ac := m.acs[m.selectedACIdx]
			// Only verify if not already verified
			if ac.Status != ACStatusVerified && ac.Status != ACStatusAutomaticallyVerified {
				ac.Status = ACStatusVerified
				ac.UpdatedAt = time.Now()
				
				// Update in repository
				if err := m.repository.UpdateAC(m.ctx, ac); err != nil {
					m.logger.Error("Failed to verify AC", "error", err)
				} else {
					// Reload ACs to reflect the change
					return m, m.loadACs(m.currentTrack.ID)
				}
			}
		}
	case "enter":
		// Navigate into selected task
		if m.selectedTaskIdx < len(m.tasks) {
			m.previousViewMode = ViewTrackDetail
			taskID := m.tasks[m.selectedTaskIdx].ID
			m.currentView = ViewLoading
			return m, m.loadTaskDetail(taskID)
		}
	case "a":
		// View ADRs for the current track
		if m.currentTrack != nil {
			m.previousViewMode = ViewTrackDetail
			m.currentView = ViewLoading
			return m, m.loadADRs(m.currentTrack.ID)
		}
	case "c":
		// View ACs for the current track
		if m.currentTrack != nil {
			m.previousViewMode = ViewTrackDetail
			m.currentView = ViewLoading
			return m, m.loadACs(m.currentTrack.ID)
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
	case " ": // Space bar - verify selected AC
		if m.selectedACIdx < len(m.acs) {
			ac := m.acs[m.selectedACIdx]
			// Only verify if not already verified
			if ac.Status != ACStatusVerified && ac.Status != ACStatusAutomaticallyVerified {
				ac.Status = ACStatusVerified
				ac.UpdatedAt = time.Now()
				
				// Update in repository
				if err := m.repository.UpdateAC(m.ctx, ac); err != nil {
					m.logger.Error("Failed to verify AC", "error", err)
				} else {
					// Reload ACs to reflect the change
					return m, m.loadACs(m.currentTrack.ID)
				}
			}
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

// handleTaskDetailKeys processes key presses on task detail view

// handleIterationDetailKeys processes key presses on iteration detail view
func (m *AppModel) handleIterationDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// Toggle between tasks and ACs
		if len(m.acs) > 0 {
			m.iterationDetailFocusAC = !m.iterationDetailFocusAC
		}
	case "j", "down":
		if m.iterationDetailFocusAC {
			// Navigate ACs
			if m.selectedIterationACIdx < len(m.acs)-1 {
				m.selectedIterationACIdx++
			}
		} else {
			// Navigate tasks
			if m.selectedIterationTaskIdx < len(m.iterationTasks)-1 {
				m.selectedIterationTaskIdx++
			}
		}
	case "k", "up":
		if m.iterationDetailFocusAC {
			// Navigate ACs
			if m.selectedIterationACIdx > 0 {
				m.selectedIterationACIdx--
			}
		} else {
			// Navigate tasks
			if m.selectedIterationTaskIdx > 0 {
				m.selectedIterationTaskIdx--
			}
		}
	case "enter":
		if !m.iterationDetailFocusAC && m.selectedIterationTaskIdx < len(m.iterationTasks) {
			// View task detail (only when focused on tasks)
			m.previousViewMode = ViewIterationDetail
			taskID := m.iterationTasks[m.selectedIterationTaskIdx].ID
			m.currentView = ViewLoading
			return m, m.loadTaskDetail(taskID)
		}
	case " ": // Space bar
		if m.iterationDetailFocusAC && m.selectedIterationACIdx < len(m.acs) {
			// Verify selected AC
			ac := m.acs[m.selectedIterationACIdx]
			if ac.Status != ACStatusVerified {
				ac.Status = ACStatusVerified
				ac.UpdatedAt = time.Now()
				if err := m.repository.UpdateAC(m.ctx, ac); err != nil {
					m.error = err
					m.currentView = ViewError
					return m, nil
				}
				// Reload iteration detail to refresh display
				return m, m.loadIterationDetail(m.currentIteration.Number)
			}
		}
	case "f":
		if m.iterationDetailFocusAC && m.selectedIterationACIdx < len(m.acs) {
			// Fail selected AC with feedback
			m.feedbackInput.SetValue("")
			m.feedbackInput.Focus()
			m.previousViewMode = ViewIterationDetail
			m.currentView = ViewACFailInput
		}
	}
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
	case ViewTaskDetail:
		return m.renderTaskDetail()
	case ViewIterationList:
		return m.renderIterationList()
	case ViewIterationDetail:
		return m.renderIterationDetail()
	case ViewADRList:
		return m.renderADRList()
	case ViewACList:
		return m.renderACList()
	case ViewACFailInput:
		return m.renderACFailInput()
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

// countLines counts the number of lines in a string (number of newlines + 1)
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// renderSelectableList renders a list of items with standardized selection highlighting
// selectedIdx: index of selected item (-1 for no selection)
// items: slice of strings to display (already formatted with icons, text, etc.)
// unselectedStyle: style for normal items
// selectedStyle: style for selected item
func renderSelectableList(selectedIdx int, items []string, unselectedStyle, selectedStyle lipgloss.Style) string {
	var result string
	for i, item := range items {
		if i == selectedIdx {
			result += selectedStyle.Render("→ "+item) + "\n"
		} else {
			result += unselectedStyle.Render("  "+item) + "\n"
		}
	}
	return result
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

	// TM-task-45: Vision and success criteria with formatting (hide roadmap ID)
	if m.roadmap != nil {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
		contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

		availableWidth := m.width
		if availableWidth <= 0 {
			availableWidth = 80
		}
		contentWidth := availableWidth - 4

		// Vision as header with content below
		s += headerStyle.Render("# Vision") + "\n"
		visionText := wrapText(m.roadmap.Vision, contentWidth)
		s += contentStyle.Render(visionText) + "\n\n"

		// Success Criteria as header with content below
		s += headerStyle.Render("# Success Criteria") + "\n"
		criteriaText := wrapText(m.roadmap.SuccessCriteria, contentWidth)
		s += contentStyle.Render(criteriaText) + "\n\n"
	}

	// TM-task-48: Show non-completed iterations on main view
	sectionHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141")).MarginTop(1)
	completedStyle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("240"))

	nonCompletedIterations := []*IterationEntity{}
	completedIterations := []*IterationEntity{}
	for _, iter := range m.iterations {
		if iter.Status == "complete" {
			completedIterations = append(completedIterations, iter)
		} else {
			nonCompletedIterations = append(nonCompletedIterations, iter)
		}
	}

	if len(nonCompletedIterations) > 0 {
		// Track iterations section position
		m.iterationsSectionLine = countLines(s)
		s += sectionHeaderStyle.Render("Active Iterations:") + "\n"
		for idx, iter := range nonCompletedIterations {
			statusIcon := m.formatIterationStatus(iter.Status)

			// Apply selection highlighting if this iteration is selected
			var line string
			if m.selectedItemType == SelectIterations && idx == m.selectedIterationIdx {
				line = fmt.Sprintf("→ %s Iteration %d: %s (%d tasks)",
					statusIcon, iter.Number, iter.Name, len(iter.TaskIDs))
				s += selectedTrackStyle.Render(line) + "\n"
			} else {
				line = fmt.Sprintf("  %s Iteration %d: %s (%d tasks)",
					statusIcon, iter.Number, iter.Name, len(iter.TaskIDs))
				s += trackItemStyle.Render(line) + "\n"
			}

			// If full view is enabled, show tasks for this iteration
			if m.showFullRoadmap {
				tasks := m.iterationTasksByNumber[iter.Number]
				if len(tasks) > 0 {
					for _, task := range tasks {
						taskLine := fmt.Sprintf("    %s %s", task.ID, task.Title)
						s += completedStyle.Render(taskLine) + "\n"
					}
				}
			}
		}
		s += "\n"
	}

	if m.showCompletedIters && len(completedIterations) > 0 {
		s += sectionHeaderStyle.Render("Completed Iterations:") + "\n"
		for _, iter := range completedIterations {
			s += completedStyle.Render(fmt.Sprintf("✓ Iteration %d: %s", iter.Number, iter.Name)) + "\n"
		}
		s += "\n"
	}

	// TM-task-47: Better track status visualization with separation
	activeTracks := []*TrackEntity{}
	completedTracks := []*TrackEntity{}
	for _, track := range m.tracks {
		if track.Status == "complete" || track.Status == "done" {
			completedTracks = append(completedTracks, track)
		} else {
			activeTracks = append(activeTracks, track)
		}
	}

	// Active tracks
	// Track tracks section position
	m.tracksSectionLine = countLines(s)
	if len(activeTracks) > 0 {
		s += sectionHeaderStyle.Render("Active Tracks:") + "\n"
		for i, track := range activeTracks {
			statusIcon := getStatusIcon(track.Status)
			priorityIcon := getPriorityIcon(track.Rank)

			// Apply selection highlighting if this track is selected
			var line string
			if m.selectedItemType == SelectTracks && i == m.selectedTrackIdx {
				line = fmt.Sprintf("→ %s %s %s - %s", statusIcon, priorityIcon, track.ID, track.Title)
				s += selectedTrackStyle.Render(line) + "\n"
			} else {
				line = fmt.Sprintf("  %s %s %s - %s", statusIcon, priorityIcon, track.ID, track.Title)
				s += trackItemStyle.Render(line) + "\n"
			}

			// If full view is enabled, show tasks for this track
			if m.showFullRoadmap {
				tasks := m.trackTasksByID[track.ID]
				if len(tasks) > 0 {
					for _, task := range tasks {
						taskLine := fmt.Sprintf("    %s %s", task.ID, task.Title)
						s += completedStyle.Render(taskLine) + "\n"
					}
				}
			}
		}
	} else {
		s += sectionHeaderStyle.Render("Active Tracks:") + "\n"
		s += "  No active tracks\n"
	}

	// Completed tracks (TM-task-47: optional view)
	if m.showCompletedTracks && len(completedTracks) > 0 {
		s += "\n" + sectionHeaderStyle.Render("Completed Tracks:") + "\n"
		for _, track := range completedTracks {
			s += completedStyle.Render(fmt.Sprintf("✓ %s - %s", track.ID, track.Title)) + "\n"
		}
	}

	// Backlog (always visible when tasks exist)
	if len(m.backlogTasks) > 0 {
		// Track backlog section position
		m.backlogSectionLine = countLines(s)
		s += "\n" + sectionHeaderStyle.Render(fmt.Sprintf("Backlog (%d):", len(m.backlogTasks))) + "\n"
		for idx, task := range m.backlogTasks {
			// Get track information for display
			trackInfo := ""
			if task.TrackID != "" {
				track, err := m.repository.GetTrack(m.ctx, task.TrackID)
				if err == nil && track != nil {
					trackInfo = fmt.Sprintf(" [%s]", track.ID)
				}
			}

			var taskLine string
			if m.selectedItemType == SelectBacklog && idx == m.selectedBacklogIdx {
				taskLine = fmt.Sprintf("→ %s %s%s", task.ID, task.Title, trackInfo)
				s += selectedTrackStyle.Render(taskLine) + "\n"
			} else {
				taskLine = fmt.Sprintf("  %s %s%s", task.ID, task.Title, trackInfo)
				s += trackItemStyle.Render(taskLine) + "\n"
			}
		}
	}

	// Set viewport content
	m.roadmapViewport.SetContent(s)

	// Render viewport
	viewportView := m.roadmapViewport.View()

	// Add hotkeys below viewport
	hotkeys := m.getRoadmapListHotkeys()
	helpText := formatHotkeyGroups(hotkeys, m.width-4, helpStyle)
	return viewportView + "\n" + helpText
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
	s += fmt.Sprintf("Status: %s | Rank: %d %s\n", m.currentTrack.Status, m.currentTrack.Rank, getPriorityIcon(m.currentTrack.Rank))

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

	// AC Status (aggregate across all tasks in track)
	if tasks, err := m.repository.ListTasks(m.ctx, TaskFilters{TrackID: m.currentTrack.ID}); err == nil {
		totalACs := 0
		verifiedACs := 0
		pendingReviewACs := 0
		for _, task := range tasks {
			if acs, err := m.repository.ListAC(m.ctx, task.ID); err == nil {
				for _, ac := range acs {
					totalACs++
					if ac.Status == ACStatusVerified || ac.Status == ACStatusAutomaticallyVerified {
						verifiedACs++
					} else if ac.Status == ACStatusPendingHumanReview {
						pendingReviewACs++
					}
				}
			}
		}
		if totalACs > 0 {
			s += fmt.Sprintf("Acceptance Criteria: %d total (%d verified, %d pending review) [press 'c' to view]\n", totalACs, verifiedACs, pendingReviewACs)
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
			priorityIcon := getPriorityIcon(task.Rank)

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
	s += helpStyle.Render("Navigation: j/k or ↑/↓ | Enter: View task | a: ADRs | c: Acceptance Criteria | esc: Back | q: Quit")

	return s
}

// loadTaskDetail loads task details
func (m *AppModel) loadTaskDetail(taskID string) tea.Cmd {
	return func() tea.Msg {
		task, err := m.repository.GetTask(m.ctx, taskID)
		if err != nil {
			return TaskDetailLoadedMsg{Error: err}
		}

		return TaskDetailLoadedMsg{
			Task: task,
		}
	}
}

// renderTaskDetail renders the task detail view
func (m *AppModel) renderTaskDetail() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginTop(1).
		MarginBottom(0)

	contentStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(0)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	if m.currentTask == nil {
		return "Loading task details..."
	}

	// Header
	statusIcon := getStatusIcon(m.currentTask.Status)
	priorityIcon := getPriorityIcon(m.currentTask.Rank)
	s += titleStyle.Render(fmt.Sprintf("%s %s Task: %s", statusIcon, priorityIcon, m.currentTask.ID)) + "\n"

	// Title
	s += sectionStyle.Render("Title") + "\n"
	s += contentStyle.Render(m.currentTask.Title) + "\n"

	// Description
	if m.currentTask.Description != "" {
		s += "\n" + sectionStyle.Render("Description") + "\n"
		s += contentStyle.Render(m.currentTask.Description) + "\n"
	}

	// Status
	s += "\n" + sectionStyle.Render("Status") + "\n"
	s += contentStyle.Render(m.currentTask.Status) + "\n"

	// Rank
	s += "\n" + sectionStyle.Render("Rank") + "\n"
	s += contentStyle.Render(fmt.Sprintf("%d %s", m.currentTask.Rank, getPriorityIcon(m.currentTask.Rank))) + "\n"

	// Track
	s += "\n" + sectionStyle.Render("Track") + "\n"
	// Load track to get name
	if m.currentTask.TrackID != "" {
		track, err := m.repository.GetTrack(m.ctx, m.currentTask.TrackID)
		if err == nil && track != nil {
			s += contentStyle.Render(fmt.Sprintf("%s - %s", track.ID, track.Title)) + "\n"
		} else {
			s += contentStyle.Render(m.currentTask.TrackID) + "\n"
		}
	} else {
		s += contentStyle.Render("(no track assigned)") + "\n"
	}

	// Branch (if set)
	if m.currentTask.Branch != "" {
		s += "\n" + sectionStyle.Render("Branch") + "\n"
		s += contentStyle.Render(m.currentTask.Branch) + "\n"
	}

	// Iterations
	iterations, err := m.repository.GetIterationsForTask(m.ctx, m.currentTask.ID)
	if err == nil {
		s += "\n" + sectionStyle.Render("Iterations") + "\n"
		if len(iterations) == 0 {
			s += contentStyle.Render("Not assigned to any iteration") + "\n"
		} else {
			for _, iter := range iterations {
				s += contentStyle.Render(fmt.Sprintf("Iteration %d: %s (status: %s)", iter.Number, iter.Name, iter.Status)) + "\n"
			}
		}
	}

	// Acceptance Criteria
	acs, acErr := m.repository.ListAC(m.ctx, m.currentTask.ID)
	if acErr == nil && len(acs) > 0 {
		s += "\n" + sectionStyle.Render(fmt.Sprintf("Acceptance Criteria (%d)", len(acs))) + "\n"
		for _, ac := range acs {
			acIcon := getACStatusIcon(string(ac.Status))
			typeStr := "manual"
			if ac.VerificationType == VerificationTypeAutomated {
				typeStr = "auto"
			}
			s += contentStyle.Render(fmt.Sprintf("%s [%s] (%s) %s", acIcon, ac.ID, typeStr, ac.Description)) + "\n"
		}
	}

	// Timestamps
	s += "\n" + sectionStyle.Render("Timestamps") + "\n"
	s += contentStyle.Render(fmt.Sprintf("Created: %s", m.currentTask.CreatedAt.Format("2006-01-02 15:04:05"))) + "\n"
	s += contentStyle.Render(fmt.Sprintf("Updated: %s", m.currentTask.UpdatedAt.Format("2006-01-02 15:04:05"))) + "\n"

	// Help
	s += "\n"
	s += helpStyle.Render("Navigation: esc: Back to track | q: Quit")

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

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

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
		// Separate tasks by status
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

		// Define styles for list rendering
		itemStyle := lipgloss.NewStyle().PaddingLeft(2)
		selectedItemStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("229"))

		// Track cumulative task index across all status groups
		taskIdx := 0

		// Render To Do tasks
		if len(todoTasks) > 0 {
			s += "\n" + sectionStyle.Render(fmt.Sprintf("To Do (%d)", len(todoTasks))) + "\n"
			var todoItems []string
			for _, task := range todoTasks {
				statusIcon := getStatusIcon(task.Status)
				priorityIcon := getPriorityIcon(task.Rank)
				todoItems = append(todoItems, fmt.Sprintf("%s %s %s - %s", statusIcon, priorityIcon, task.ID, task.Title))
			}
			s += renderSelectableList(m.selectedIterationTaskIdx-taskIdx, todoItems, itemStyle, selectedItemStyle)
			taskIdx += len(todoTasks)
		}

		// Render In Progress tasks
		if len(inProgressTasks) > 0 {
			s += "\n" + sectionStyle.Render(fmt.Sprintf("In Progress (%d)", len(inProgressTasks))) + "\n"
			var inProgressItems []string
			for _, task := range inProgressTasks {
				statusIcon := getStatusIcon(task.Status)
				priorityIcon := getPriorityIcon(task.Rank)
				inProgressItems = append(inProgressItems, fmt.Sprintf("%s %s %s - %s", statusIcon, priorityIcon, task.ID, task.Title))
			}
			s += renderSelectableList(m.selectedIterationTaskIdx-taskIdx, inProgressItems, itemStyle, selectedItemStyle)
			taskIdx += len(inProgressTasks)
		}

		// Render Done tasks
		if len(doneTasks) > 0 {
			s += "\n" + sectionStyle.Render(fmt.Sprintf("Done (%d)", len(doneTasks))) + "\n"
			var doneItems []string
			for _, task := range doneTasks {
				statusIcon := getStatusIcon(task.Status)
				priorityIcon := getPriorityIcon(task.Rank)
				doneItems = append(doneItems, fmt.Sprintf("%s %s %s - %s", statusIcon, priorityIcon, task.ID, task.Title))
			}
			s += renderSelectableList(m.selectedIterationTaskIdx-taskIdx, doneItems, itemStyle, selectedItemStyle)
		}
	}

	// Acceptance Criteria section
	if len(m.acs) > 0 {
		s += "\n\n"
		focusIndicator := ""
		if m.iterationDetailFocusAC {
			focusIndicator = " [FOCUSED]"
		}
		s += sectionStyle.Render(fmt.Sprintf("Acceptance Criteria (%d)%s", len(m.acs), focusIndicator)) + "\n"

		// Group ACs by task (but keep flat list for selection)
		acsByTask := make(map[string][]*AcceptanceCriteriaEntity)
		taskMap := make(map[string]*TaskEntity)
		for _, task := range m.iterationTasks {
			taskMap[task.ID] = task
		}
		for _, ac := range m.acs {
			acsByTask[ac.TaskID] = append(acsByTask[ac.TaskID], ac)
		}

		// Styles for AC items
		acItemStyle := lipgloss.NewStyle().PaddingLeft(2)
		acSelectedStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("229"))

		// Track cumulative AC index for selection
		acIdx := 0

		// Display ACs grouped by task
		for _, task := range m.iterationTasks {
			acs, hasACs := acsByTask[task.ID]
			if !hasACs || len(acs) == 0 {
				continue
			}

			s += "\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(fmt.Sprintf("%s: %s", task.ID, task.Title)) + "\n"
			for _, ac := range acs {
				statusIcon := getACStatusIcon(string(ac.Status))
				// Truncate description if too long
				desc := ac.Description
				if len(desc) > 80 {
					desc = desc[:77] + "..."
				}

				// Build AC line
				acLine := fmt.Sprintf("%s [%s] %s", statusIcon, ac.ID, desc)

				// Apply selection style if this AC is selected and in focus mode
				if m.iterationDetailFocusAC && acIdx == m.selectedIterationACIdx {
					s += acSelectedStyle.Render(acLine) + "\n"
				} else {
					s += acItemStyle.Render(acLine) + "\n"
				}
				acIdx++
			}
		}
	}

	// Help - context sensitive
	s += "\n"
	if len(m.acs) > 0 {
		if m.iterationDetailFocusAC {
			s += helpStyle.Render("j/k/↑/↓: Navigate ACs | space: Verify | f: Fail | tab: Switch to Tasks | esc: Back | q: Quit")
		} else {
			s += helpStyle.Render("j/k/↑/↓: Navigate Tasks | Enter: View Details | tab: Switch to ACs | esc: Back | q: Quit")
		}
	} else {
		s += helpStyle.Render("j/k/↑/↓: Navigate | Enter: View Details | esc: Back | q: Quit")
	}

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

func getPriorityIcon(rank int) string {
	// Convert rank to icon: 1-100=critical (red), 101-200=high (orange), 201-300=medium (yellow), 301+=low (green)
	if rank <= 100 {
		return "🔴" // critical
	} else if rank <= 200 {
		return "🟠" // high
	} else if rank <= 300 {
		return "🟡" // medium
	} else {
		return "🟢" // low
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
	case " ": // Space bar - verify selected AC
		if m.selectedACIdx < len(m.acs) {
			ac := m.acs[m.selectedACIdx]
			// Only verify if not already verified
			if ac.Status != ACStatusVerified && ac.Status != ACStatusAutomaticallyVerified {
				ac.Status = ACStatusVerified
				ac.UpdatedAt = time.Now()
				
				// Update in repository
				if err := m.repository.UpdateAC(m.ctx, ac); err != nil {
					m.logger.Error("Failed to verify AC", "error", err)
				} else {
					// Reload ACs to reflect the change
					return m, m.loadACs(m.currentTrack.ID)
				}
			}
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

// renderACList renders the AC list view
func (m *AppModel) renderACList() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	acItemStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(1)

	selectedACStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginBottom(1).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("229"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	// Header
	s += titleStyle.Render(fmt.Sprintf("Acceptance Criteria for Track: %s", m.currentTrack.ID)) + "\n\n"

	// ACs
	if len(m.acs) == 0 {
		s += "No acceptance criteria yet. Create one with:\n"
		s += "  dw task-manager ac add <task-id> --description \"...\" [--type manual|automated]\n"
	} else {
		for i, ac := range m.acs {
			statusIcon := getACStatusIcon(string(ac.Status))
			typeStr := "manual"
			if ac.VerificationType == VerificationTypeAutomated {
				typeStr = "auto"
			}
			line := fmt.Sprintf("%s [%s] Task: %s (%s)\n  %s", statusIcon, ac.ID, ac.TaskID, typeStr, ac.Description)

			// Show feedback indicator for failed ACs
			if ac.Status == ACStatusFailed && ac.Notes != "" {
				line += fmt.Sprintf("\n  Feedback: %s", ac.Notes)
			}

			if i == m.selectedACIdx {
				s += selectedACStyle.Render(line) + "\n"
			} else {
				s += acItemStyle.Render(line) + "\n"
			}
		}
	}

	// Help
	s += "\n"
	s += helpStyle.Render("Navigation: j/k or ↑/↓ | space: Verify selected AC | f: Mark as Failed | esc: Back | q: Quit")

	return s
}

// renderACFailInput renders the AC failure feedback input view
func (m *AppModel) renderACFailInput() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	var s string

	// Get selected AC
	if m.selectedACIdx < len(m.acs) {
		ac := m.acs[m.selectedACIdx]

		// Header
		s += titleStyle.Render(fmt.Sprintf("Mark AC as Failed: %s", ac.ID)) + "\n\n"
		s += fmt.Sprintf("Description: %s\n\n", ac.Description)

		// Input prompt
		s += "Enter failure feedback (reason why AC failed):\n\n"
		s += m.feedbackInput.View() + "\n\n"

		// Help
		s += helpStyle.Render("Enter: Submit | esc: Cancel")
	} else {
		s = "Error: No AC selected"
	}

	return s
}

// handleACListKeys processes key presses on AC list view
func (m *AppModel) handleACListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selectedACIdx < len(m.acs)-1 {
			m.selectedACIdx++
		}
	case "k", "up":
		if m.selectedACIdx > 0 {
			m.selectedACIdx--
		}
	case " ": // Space bar - verify selected AC
		if m.selectedACIdx < len(m.acs) {
			ac := m.acs[m.selectedACIdx]
			// Only verify if not already verified
			if ac.Status != ACStatusVerified && ac.Status != ACStatusAutomaticallyVerified {
				ac.Status = ACStatusVerified
				ac.UpdatedAt = time.Now()

				// Update in repository
				if err := m.repository.UpdateAC(m.ctx, ac); err != nil {
					m.logger.Error("Failed to verify AC", "error", err)
				} else {
					// Reload ACs to reflect the change
					return m, m.loadACs(m.currentTrack.ID)
				}
			}
		}
	case "f": // 'f' key - mark selected AC as failed with feedback
		if m.selectedACIdx < len(m.acs) {
			// Enter feedback input mode
			m.feedbackInput.SetValue("")
			m.feedbackInput.Focus()
			m.currentView = ViewACFailInput
		}
	}
	return m, nil
}

// handleACFailInputKeys processes key presses in AC failure input mode
func (m *AppModel) handleACFailInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		// Submit the feedback and mark AC as failed
		if m.selectedACIdx < len(m.acs) {
			ac := m.acs[m.selectedACIdx]
			feedback := m.feedbackInput.Value()

			// Update AC status to failed with feedback
			ac.Status = ACStatusFailed
			ac.Notes = feedback
			ac.UpdatedAt = time.Now()

			// Update in repository
			if err := m.repository.UpdateAC(m.ctx, ac); err != nil {
				m.logger.Error("Failed to mark AC as failed", "error", err)
			} else {
				// Clear input and return to AC list
				m.feedbackInput.SetValue("")
				m.feedbackInput.Blur()
				m.currentView = ViewACList
				// Reload ACs to reflect the change
				return m, m.loadACs(m.currentTrack.ID)
			}
		}
		return m, nil
	case tea.KeyCtrlC, tea.KeyEsc:
		// Cancel and return to AC list
		m.feedbackInput.SetValue("")
		m.feedbackInput.Blur()
		m.currentView = ViewACList
		return m, nil
	}

	// Update text input with key press
	m.feedbackInput, cmd = m.feedbackInput.Update(msg)
	return m, cmd
}

// loadACs loads ACs for a track (all tasks in track)
func (m *AppModel) loadACs(trackID string) tea.Cmd {
	return func() tea.Msg {
		// Get all tasks in track
		tasks, err := m.repository.ListTasks(m.ctx, TaskFilters{TrackID: trackID})
		if err != nil {
			return ACsLoadedMsg{
				ACs:   nil,
				Error: err,
			}
		}

		// Collect all ACs from all tasks
		var allACs []*AcceptanceCriteriaEntity
		for _, task := range tasks {
			acs, err := m.repository.ListAC(m.ctx, task.ID)
			if err != nil {
				continue // Skip task if error
			}
			allACs = append(allACs, acs...)
		}

		return ACsLoadedMsg{
			ACs:   allACs,
			Error: nil,
		}
	}
}

// getACStatusIcon returns a visual icon for AC status
func getACStatusIcon(status string) string {
	switch status {
	case string(ACStatusVerified):
		return "✓"
	case string(ACStatusAutomaticallyVerified):
		return "✓ₐ"
	case string(ACStatusPendingHumanReview):
		return "⏸"
	case string(ACStatusFailed):
		return "✗"
	case string(ACStatusNotStarted):
		return "○"
	default:
		return "?"
	}
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for _, word := range words {
		wordLen := len(word)

		if lineLen == 0 {
			result.WriteString(word)
			lineLen = wordLen
		} else if lineLen+1+wordLen <= width {
			result.WriteString(" ")
			result.WriteString(word)
			lineLen += 1 + wordLen
		} else {
			result.WriteString("\n")
			result.WriteString(word)
			lineLen = wordLen
		}
	}

	return result.String()
}

// SetACs sets the ACs for testing
func (m *AppModel) SetACs(acs []*AcceptanceCriteriaEntity) {
	m.acs = acs
}

// SetSelectedACIdx sets the selected AC index for testing
func (m *AppModel) SetSelectedACIdx(idx int) {
	m.selectedACIdx = idx
}

// rankToPriority converts rank (1-1000) to priority string for display

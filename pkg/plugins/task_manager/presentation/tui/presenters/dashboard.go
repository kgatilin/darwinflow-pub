package presenters

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/components"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui/viewmodels"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
)

// RoadmapListKeyMap defines keybindings for the dashboard view
type RoadmapListKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Help     key.Binding
	Refresh  key.Binding
	Tab      key.Binding // Tab to cycle between sections
	MoveUp   key.Binding // Shift+up or K for reordering
	MoveDown key.Binding // Shift+down or J for reordering
	PageUp   key.Binding // Page up or b
	PageDown key.Binding // Page down or space
}

// NewRoadmapListKeyMap creates default keybindings for dashboard
func NewRoadmapListKeyMap() RoadmapListKeyMap {
	return RoadmapListKeyMap{
		Up:    components.NewUpKey(),
		Down:  components.NewDownKey(),
		Enter: components.NewEnterKey(),
		Quit:  components.NewQuitKey(),
		Help:  components.NewHelpKey(),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch section"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("shift+up", "K"),
			key.WithHelp("K/shift+↑", "move up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("shift+down", "J"),
			key.WithHelp("J/shift+↓", "move down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdn", "space"),
			key.WithHelp("pgdn/space", "page down"),
		),
	}
}

// ShortHelp returns keybindings to show in short help view
func (k RoadmapListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Tab, k.Refresh, k.Quit}
}

// FullHelp returns all keybindings for full help view
func (k RoadmapListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Tab, k.Refresh},
		{k.PageUp, k.PageDown},
		{k.MoveUp, k.MoveDown},
		{k.Help, k.Quit},
	}
}

// DashboardSection represents the three main sections in the dashboard
type DashboardSection int

const (
	SectionIterations DashboardSection = iota
	SectionTracks
	SectionBacklog
)

// RoadmapListPresenter presents the dashboard view with iterations, tracks, and backlog
type RoadmapListPresenter struct {
	viewModel     *viewmodels.RoadmapListViewModel
	help          components.Help
	keys          RoadmapListKeyMap
	showFullHelp  bool
	selectedIndex int
	activeSection DashboardSection // Track which section has focus
	width         int
	height        int
	repo          domain.RoadmapRepository
	ctx           context.Context
	scrollHelper  *components.ScrollHelper
}

// NewRoadmapListPresenter creates a new dashboard presenter
func NewRoadmapListPresenter(vm *viewmodels.RoadmapListViewModel, repo domain.RoadmapRepository, ctx context.Context) *RoadmapListPresenter {
	return NewRoadmapListPresenterWithSelection(vm, repo, ctx, 0)
}

// NewRoadmapListPresenterWithSelection creates a new dashboard presenter with initial selection
func NewRoadmapListPresenterWithSelection(vm *viewmodels.RoadmapListViewModel, repo domain.RoadmapRepository, ctx context.Context, selectedIndex int) *RoadmapListPresenter {
	return &RoadmapListPresenter{
		viewModel:     vm,
		help:          components.NewHelp(),
		keys:          NewRoadmapListKeyMap(),
		showFullHelp:  false,
		selectedIndex: selectedIndex,
		activeSection: SectionIterations, // Default to iterations section
		repo:          repo,
		ctx:           ctx,
		width:         80, // Default width until WindowSizeMsg arrives
		height:        24,
		scrollHelper:  components.NewScrollHelper(),
	}
}

func (p *RoadmapListPresenter) Init() tea.Cmd {
	// Request terminal size immediately to get actual dimensions
	return tea.WindowSize()
}

func (p *RoadmapListPresenter) Update(msg tea.Msg) (Presenter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.help.SetWidth(msg.Width)

		// Calculate available viewport height for scrolling
		// Account for: title (1) + vision/criteria section (if present, ~5) + section headers (3) + help (2)
		headerHeight := 5
		if p.viewModel.Vision != "" || p.viewModel.SuccessCriteria != "" {
			headerHeight = 9
		}
		footerHeight := 2 // Help text
		availableHeight := msg.Height - headerHeight - footerHeight
		if availableHeight < 5 {
			availableHeight = 5 // Minimum height
		}
		p.scrollHelper.SetViewportHeight(availableHeight)
		p.scrollHelper.EnsureVisible(getTotalItems(p.viewModel), p.selectedIndex)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keys.Quit):
			return p, tea.Quit
		case key.Matches(msg, p.keys.Help):
			p.showFullHelp = !p.showFullHelp
		case key.Matches(msg, p.keys.Tab):
			// Cycle through sections: Iterations → Tracks → Backlog → Iterations
			p.cycleActiveSection()
		case key.Matches(msg, p.keys.Refresh):
			// Reload dashboard data, preserving current selection
			return p, func() tea.Msg {
				return RefreshDashboardMsg{SelectedIndex: p.selectedIndex}
			}
		case key.Matches(msg, p.keys.Up):
			totalItems := getTotalItems(p.viewModel)
			if p.selectedIndex > 0 {
				p.selectedIndex--
				p.scrollHelper.EnsureVisible(totalItems, p.selectedIndex)
			}
		case key.Matches(msg, p.keys.Down):
			totalItems := getTotalItems(p.viewModel)
			if p.selectedIndex < totalItems-1 {
				p.selectedIndex++
				p.scrollHelper.EnsureVisible(totalItems, p.selectedIndex)
			}
		case key.Matches(msg, p.keys.PageUp):
			totalItems := getTotalItems(p.viewModel)
			newIndex := p.scrollHelper.PageUp(totalItems)
			p.selectedIndex = newIndex
		case key.Matches(msg, p.keys.PageDown):
			totalItems := getTotalItems(p.viewModel)
			newIndex := p.scrollHelper.PageDown(totalItems, p.selectedIndex)
			p.selectedIndex = newIndex
		case key.Matches(msg, p.keys.Enter):
			// Navigate to selected item
			if p.selectedIndex < len(p.viewModel.ActiveIterations) {
				// Navigate to iteration detail
				iter := p.viewModel.ActiveIterations[p.selectedIndex]
				return p, func() tea.Msg {
					return IterationSelectedMsg{
						IterationNumber: iter.Number,
						SelectedIndex:   p.selectedIndex,
					}
				}
			}
			// Check if selection is in tracks section
			trackOffset := len(p.viewModel.ActiveIterations)
			if p.selectedIndex >= trackOffset && p.selectedIndex < trackOffset+len(p.viewModel.ActiveTracks) {
				// Navigate to track detail
				trackIndex := p.selectedIndex - trackOffset
				track := p.viewModel.ActiveTracks[trackIndex]
				return p, func() tea.Msg {
					return TrackSelectedMsg{
						TrackID:       track.ID,
						SelectedIndex: p.selectedIndex,
					}
				}
			}
			// Check if selection is in backlog section
			backlogOffset := len(p.viewModel.ActiveIterations) + len(p.viewModel.ActiveTracks)
			if p.selectedIndex >= backlogOffset && p.selectedIndex < backlogOffset+len(p.viewModel.BacklogTasks) {
				// Navigate to task detail
				taskIndex := p.selectedIndex - backlogOffset
				task := p.viewModel.BacklogTasks[taskIndex]
				return p, func() tea.Msg {
					return TaskSelectedMsg{
						TaskID:        task.ID,
						SelectedIndex: p.selectedIndex,
					}
				}
			}
		case key.Matches(msg, p.keys.MoveUp):
			// Reorder iterations (move selected iteration up)
			if p.selectedIndex > 0 && p.selectedIndex < len(p.viewModel.ActiveIterations) {
				return p, p.reorderIterations(p.selectedIndex, p.selectedIndex-1)
			}
		case key.Matches(msg, p.keys.MoveDown):
			// Reorder iterations (move selected iteration down)
			if p.selectedIndex < len(p.viewModel.ActiveIterations)-1 {
				return p, p.reorderIterations(p.selectedIndex, p.selectedIndex+1)
			}
		}
	}

	return p, nil
}

// getTotalItems returns the total number of items across all sections
func getTotalItems(vm *viewmodels.RoadmapListViewModel) int {
	return len(vm.ActiveIterations) + len(vm.ActiveTracks) + len(vm.BacklogTasks)
}

func (p *RoadmapListPresenter) View() string {
	var b strings.Builder

	// Title
	b.WriteString(components.Styles.TitleStyle.Render("Dashboard"))
	b.WriteString("\n\n")

	// Roadmap vision and success criteria header
	if p.viewModel.Vision != "" || p.viewModel.SuccessCriteria != "" {
		b.WriteString(components.Styles.SectionStyle.Render("Roadmap Vision"))
		b.WriteString("\n")
		if p.viewModel.Vision != "" {
			// Use wordwrap + indent for proper ANSI-aware text wrapping (Bubble Tea best practice)
			indentSize := 2
			availableWidth := p.width - indentSize - 2 // Account for indent + right margin
			if availableWidth < 20 {
				availableWidth = 20 // Safety minimum for extremely narrow terminals
			}
			wrappedVision := wordwrap.String(p.viewModel.Vision, availableWidth)
			indentedVision := indent.String(wrappedVision, uint(indentSize))
			b.WriteString(indentedVision)
			b.WriteString("\n")
		}
		b.WriteString("\n")

		b.WriteString(components.Styles.SectionStyle.Render("Success Criteria"))
		b.WriteString("\n")
		if p.viewModel.SuccessCriteria != "" {
			// Use wordwrap + indent for proper ANSI-aware text wrapping (Bubble Tea best practice)
			indentSize := 2
			availableWidth := p.width - indentSize - 2 // Account for indent + right margin
			if availableWidth < 20 {
				availableWidth = 20 // Safety minimum for extremely narrow terminals
			}
			wrappedCriteria := wordwrap.String(p.viewModel.SuccessCriteria, availableWidth)
			indentedCriteria := indent.String(wrappedCriteria, uint(indentSize))
			b.WriteString(indentedCriteria)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Get visible range for scrolling
	totalItems := getTotalItems(p.viewModel)
	start, end := p.scrollHelper.VisibleRange(totalItems)

	// Render visible items from all sections
	currentItemIndex := 0

	// Active Iterations section
	if len(p.viewModel.ActiveIterations) > 0 {
		if start < len(p.viewModel.ActiveIterations) {
			// Only show section header if items in this section are visible
			// Highlight header if this section is active
			if p.activeSection == SectionIterations {
				b.WriteString(components.Styles.SelectedStyle.Render("Active Iterations"))
			} else {
				b.WriteString(components.Styles.SectionStyle.Render("Active Iterations"))
			}
			b.WriteString("\n")
		}

		for _, iter := range p.viewModel.ActiveIterations {
			if currentItemIndex < start {
				currentItemIndex++
				continue
			}
			if currentItemIndex >= end {
				break
			}

			var itemStyle string
			if p.isSelected(currentItemIndex, "iteration") {
				itemStyle = components.Styles.SelectedStyle.Render(
					fmt.Sprintf("  #%d %s (%d tasks) - %s",
						iter.Number, iter.Name, iter.TaskCount, iter.Status))
			} else {
				itemStyle = fmt.Sprintf("  #%d %s (%d tasks) - %s",
					iter.Number, iter.Name, iter.TaskCount, iter.Status)
			}
			b.WriteString(itemStyle)
			b.WriteString("\n")
			currentItemIndex++
		}

		if currentItemIndex > start {
			b.WriteString("\n")
		}
	}

	// Active Tracks section
	if len(p.viewModel.ActiveTracks) > 0 {
		if currentItemIndex < end && start < currentItemIndex+len(p.viewModel.ActiveTracks) {
			// Only show section header if items in this section are visible
			// Highlight header if this section is active
			if p.activeSection == SectionTracks {
				b.WriteString(components.Styles.SelectedStyle.Render("Active Tracks"))
			} else {
				b.WriteString(components.Styles.SectionStyle.Render("Active Tracks"))
			}
			b.WriteString("\n")
		}

		for _, track := range p.viewModel.ActiveTracks {
			if currentItemIndex < start {
				currentItemIndex++
				continue
			}
			if currentItemIndex >= end {
				break
			}

			var itemStyle string
			if p.isSelected(currentItemIndex, "track") {
				itemStyle = components.Styles.SelectedStyle.Render(
					fmt.Sprintf("  %s: %s (%d tasks) - %s",
						track.ID, track.Title, track.TaskCount, track.Status))
			} else {
				itemStyle = fmt.Sprintf("  %s: %s (%d tasks) - %s",
					track.ID, track.Title, track.TaskCount, track.Status)
			}
			b.WriteString(itemStyle)
			b.WriteString("\n")
			currentItemIndex++
		}

		if currentItemIndex > start && currentItemIndex < totalItems {
			b.WriteString("\n")
		}
	}

	// Backlog Tasks section
	if len(p.viewModel.BacklogTasks) > 0 {
		if currentItemIndex < end && start < currentItemIndex+len(p.viewModel.BacklogTasks) {
			// Only show section header if items in this section are visible
			// Highlight header if this section is active
			if p.activeSection == SectionBacklog {
				b.WriteString(components.Styles.SelectedStyle.Render("Backlog Tasks"))
			} else {
				b.WriteString(components.Styles.SectionStyle.Render("Backlog Tasks"))
			}
			b.WriteString("\n")
		}

		for _, task := range p.viewModel.BacklogTasks {
			if currentItemIndex < start {
				currentItemIndex++
				continue
			}
			if currentItemIndex >= end {
				break
			}

			var itemStyle string
			if p.isSelected(currentItemIndex, "task") {
				itemStyle = components.Styles.SelectedStyle.Render(
					fmt.Sprintf("  %s: %s - %s",
						task.ID, task.Title, task.Status))
			} else {
				itemStyle = fmt.Sprintf("  %s: %s - %s",
					task.ID, task.Title, task.Status)
			}
			b.WriteString(itemStyle)
			b.WriteString("\n")
			currentItemIndex++
		}

		if currentItemIndex > start {
			b.WriteString("\n")
		}
	}

	// Scroll indicators (optional but helpful)
	if start > 0 {
		b.WriteString(components.Styles.MetadataStyle.Render("  ↑ More items above\n"))
	}
	if end < totalItems {
		b.WriteString(components.Styles.MetadataStyle.Render("  ↓ More items below\n"))
	}

	// Help view
	if p.showFullHelp {
		b.WriteString(p.help.FullHelpView(p.keys.FullHelp()))
	} else {
		b.WriteString(p.help.ShortHelpView(p.keys.ShortHelp()))
	}

	return b.String()
}

// isSelected checks if the given index is currently selected
func (p *RoadmapListPresenter) isSelected(index int, section string) bool {
	return p.selectedIndex == index
}

// reorderIterations reorders iterations using fractional ranking
func (p *RoadmapListPresenter) reorderIterations(fromIndex, toIndex int) tea.Cmd {
	return func() tea.Msg {
		if fromIndex < 0 || fromIndex >= len(p.viewModel.ActiveIterations) {
			return nil
		}
		if toIndex < 0 || toIndex >= len(p.viewModel.ActiveIterations) {
			return nil
		}

		// Get the iteration being moved
		iterToMove := p.viewModel.ActiveIterations[fromIndex]

		// Fetch full iteration entity
		iteration, err := p.repo.GetIteration(p.ctx, iterToMove.Number)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Calculate new rank using fractional ranking
		// This ensures stable ordering regardless of how many times we move
		var newRank float64

		if toIndex == 0 {
			// Moving to first position: rank = first_item.rank - 1
			firstIter, err := p.repo.GetIteration(p.ctx, p.viewModel.ActiveIterations[0].Number)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			newRank = firstIter.Rank - 1
		} else if toIndex == len(p.viewModel.ActiveIterations)-1 {
			// Moving to last position: rank = last_item.rank + 1
			lastIter, err := p.repo.GetIteration(p.ctx, p.viewModel.ActiveIterations[len(p.viewModel.ActiveIterations)-1].Number)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			newRank = lastIter.Rank + 1
		} else {
			// Moving to middle position: rank = average of adjacent items
			var prevRank, nextRank float64

			if fromIndex < toIndex {
				// Moving down: insert between toIndex and toIndex+1
				targetIter, err := p.repo.GetIteration(p.ctx, p.viewModel.ActiveIterations[toIndex].Number)
				if err != nil {
					return ErrorMsg{Err: err}
				}
				prevRank = targetIter.Rank

				if toIndex+1 < len(p.viewModel.ActiveIterations) {
					nextIter, err := p.repo.GetIteration(p.ctx, p.viewModel.ActiveIterations[toIndex+1].Number)
					if err != nil {
						return ErrorMsg{Err: err}
					}
					nextRank = nextIter.Rank
				} else {
					nextRank = prevRank + 2
				}
			} else {
				// Moving up: insert between toIndex-1 and toIndex
				targetIter, err := p.repo.GetIteration(p.ctx, p.viewModel.ActiveIterations[toIndex].Number)
				if err != nil {
					return ErrorMsg{Err: err}
				}
				nextRank = targetIter.Rank

				if toIndex > 0 {
					prevIter, err := p.repo.GetIteration(p.ctx, p.viewModel.ActiveIterations[toIndex-1].Number)
					if err != nil {
						return ErrorMsg{Err: err}
					}
					prevRank = prevIter.Rank
				} else {
					prevRank = nextRank - 2
				}
			}

			// Calculate intermediate rank
			if prevRank == nextRank {
				// Collision detected (shouldn't happen after v8 migration)
				const epsilon = 0.01
				if fromIndex < toIndex {
					// Moving down - insert after nextRank
					newRank = nextRank + epsilon
				} else {
					// Moving up - insert before prevRank
					newRank = prevRank - epsilon
				}
				// Note: Collisions shouldn't exist after v8 normalization
			} else {
				// No collision - use fractional ranking (average)
				newRank = (prevRank + nextRank) / 2.0
			}
		}

		// Update iteration rank
		iteration.Rank = newRank
		err = p.repo.UpdateIteration(p.ctx, iteration)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		// Update selected index to follow the moved iteration
		p.selectedIndex = toIndex

		return ReorderCompletedMsg{SelectedIterationNumber: iterToMove.Number}
	}
}

// cycleActiveSection cycles through sections: Iterations → Tracks → Backlog → Iterations
// Updates activeSection and adjusts selectedIndex to first item in new section
func (p *RoadmapListPresenter) cycleActiveSection() {
	// Determine next section with available items
	startSection := p.activeSection
	for {
		// Move to next section
		p.activeSection = (p.activeSection + 1) % 3

		// Check if this section has items
		switch p.activeSection {
		case SectionIterations:
			if len(p.viewModel.ActiveIterations) > 0 {
				// Jump to first iteration
				p.selectedIndex = 0
				p.scrollHelper.EnsureVisible(getTotalItems(p.viewModel), p.selectedIndex)
				return
			}
		case SectionTracks:
			if len(p.viewModel.ActiveTracks) > 0 {
				// Jump to first track
				p.selectedIndex = len(p.viewModel.ActiveIterations)
				p.scrollHelper.EnsureVisible(getTotalItems(p.viewModel), p.selectedIndex)
				return
			}
		case SectionBacklog:
			if len(p.viewModel.BacklogTasks) > 0 {
				// Jump to first backlog task
				p.selectedIndex = len(p.viewModel.ActiveIterations) + len(p.viewModel.ActiveTracks)
				p.scrollHelper.EnsureVisible(getTotalItems(p.viewModel), p.selectedIndex)
				return
			}
		}

		// If we've cycled back to start without finding items, break to avoid infinite loop
		if p.activeSection == startSection {
			break
		}
	}
}

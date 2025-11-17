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

// TrackDetailKeyMap defines keybindings for track detail view
type TrackDetailKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Back     key.Binding
	Help     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

// NewTrackDetailKeyMap creates default keybindings for track detail
func NewTrackDetailKeyMap() TrackDetailKeyMap {
	return TrackDetailKeyMap{
		Up:    components.NewUpKey(),
		Down:  components.NewDownKey(),
		Enter: components.NewEnterKey(),
		Quit:  components.NewQuitKey(),
		Back:  components.NewBackKey(),
		Help:  components.NewHelpKey(),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdn"),
			key.WithHelp("pgdn", "page down"),
		),
	}
}

// ShortHelp returns keybindings for short help
func (k TrackDetailKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Quit}
}

// FullHelp returns all keybindings for full help
func (k TrackDetailKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.PageUp, k.PageDown},
		{k.Back, k.Help, k.Quit},
	}
}

// TrackDetailPresenter presents the track detail view
type TrackDetailPresenter struct {
	viewModel      *viewmodels.TrackDetailViewModel
	help           components.Help
	keys           TrackDetailKeyMap
	showFullHelp   bool
	selectedIndex  int
	width          int
	height         int
	repo           domain.RoadmapRepository
	ctx            context.Context
	scrollHelper   *components.ScrollHelper
	terminalHeight int
}

// NewTrackDetailPresenter creates a new track detail presenter
func NewTrackDetailPresenter(vm *viewmodels.TrackDetailViewModel, repo domain.RoadmapRepository, ctx context.Context) *TrackDetailPresenter {
	return NewTrackDetailPresenterWithSelection(vm, repo, ctx, 0)
}

// NewTrackDetailPresenterWithSelection creates a new track detail presenter with a specific selected index
func NewTrackDetailPresenterWithSelection(vm *viewmodels.TrackDetailViewModel, repo domain.RoadmapRepository, ctx context.Context, selectedIndex int) *TrackDetailPresenter {
	return &TrackDetailPresenter{
		viewModel:      vm,
		help:           components.NewHelp(),
		keys:           NewTrackDetailKeyMap(),
		showFullHelp:   false,
		selectedIndex:  selectedIndex,
		repo:           repo,
		ctx:            ctx,
		width:          80,
		height:         24,
		scrollHelper:   components.NewScrollHelper(),
		terminalHeight: 24,
	}
}

func (p *TrackDetailPresenter) Init() tea.Cmd {
	// Request terminal size immediately to get actual dimensions
	return tea.WindowSize()
}

func (p *TrackDetailPresenter) Update(msg tea.Msg) (Presenter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.terminalHeight = msg.Height
		p.help.SetWidth(msg.Width)

		// Calculate available viewport height for scrolling
		// Account for: title (1) + metadata (5-7 lines) + progress (1) + section headers + help (2)
		headerHeight := 12
		footerHeight := 2
		availableHeight := msg.Height - headerHeight - footerHeight
		if availableHeight < 5 {
			availableHeight = 5
		}

		p.scrollHelper.SetViewportHeight(availableHeight)

		// Ensure current selection is visible with new viewport height
		totalTasks := len(p.viewModel.TODOTasks) + len(p.viewModel.InProgressTasks) + len(p.viewModel.DoneTasks)
		p.scrollHelper.EnsureVisible(totalTasks, p.selectedIndex)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keys.Quit):
			return p, tea.Quit
		case key.Matches(msg, p.keys.Back):
			return p, func() tea.Msg { return BackMsgNew{} }
		case key.Matches(msg, p.keys.Help):
			p.showFullHelp = !p.showFullHelp
		case key.Matches(msg, p.keys.Up):
			if p.selectedIndex > 0 {
				p.selectedIndex--
				totalTasks := len(p.viewModel.TODOTasks) + len(p.viewModel.InProgressTasks) + len(p.viewModel.DoneTasks)
				p.scrollHelper.EnsureVisible(totalTasks, p.selectedIndex)
			}
		case key.Matches(msg, p.keys.Down):
			maxIndex := p.getMaxIndex()
			if p.selectedIndex < maxIndex {
				p.selectedIndex++
				totalTasks := len(p.viewModel.TODOTasks) + len(p.viewModel.InProgressTasks) + len(p.viewModel.DoneTasks)
				p.scrollHelper.EnsureVisible(totalTasks, p.selectedIndex)
			}
		case key.Matches(msg, p.keys.PageUp):
			totalTasks := len(p.viewModel.TODOTasks) + len(p.viewModel.InProgressTasks) + len(p.viewModel.DoneTasks)
			newIndex := p.scrollHelper.PageUp(totalTasks)
			p.selectedIndex = newIndex
		case key.Matches(msg, p.keys.PageDown):
			totalTasks := len(p.viewModel.TODOTasks) + len(p.viewModel.InProgressTasks) + len(p.viewModel.DoneTasks)
			newIndex := p.scrollHelper.PageDown(totalTasks, p.selectedIndex)
			p.selectedIndex = newIndex
		case key.Matches(msg, p.keys.Enter):
			// Navigate to task detail
			taskID := p.getSelectedTaskID()
			if taskID != "" {
				return p, func() tea.Msg {
					return TaskSelectedMsg{TaskID: taskID}
				}
			}
		}
	}

	return p, nil
}

func (p *TrackDetailPresenter) View() string {
	var b strings.Builder

	// Title
	b.WriteString(components.Styles.TitleStyle.Render(fmt.Sprintf("Track: %s", p.viewModel.Title)))
	b.WriteString("\n\n")

	// Metadata
	b.WriteString(components.Styles.MetadataStyle.Render(fmt.Sprintf("ID: %s", p.viewModel.ID)))
	b.WriteString("\n")
	b.WriteString(components.Styles.MetadataStyle.Render(fmt.Sprintf("Status: %s", p.viewModel.StatusLabel)))
	b.WriteString("\n")
	b.WriteString(components.Styles.MetadataStyle.Render(fmt.Sprintf("Rank: %d", p.viewModel.Rank)))
	b.WriteString("\n")

	// Description with text wrapping
	if p.viewModel.Description != "" {
		availableWidth := p.width - 14 // Account for "Description: " prefix
		if availableWidth < 20 {
			availableWidth = 20
		}
		wrappedDesc := wordwrap.String(p.viewModel.Description, availableWidth)
		// Apply indentation to all lines AFTER the first
		lines := strings.Split(wrappedDesc, "\n")
		if len(lines) > 1 {
			indented := lines[0] + "\n" + indent.String(strings.Join(lines[1:], "\n"), 14)
			b.WriteString(components.Styles.MetadataStyle.Render(fmt.Sprintf("Description: %s", indented)))
		} else {
			b.WriteString(components.Styles.MetadataStyle.Render(fmt.Sprintf("Description: %s", wrappedDesc)))
		}
		b.WriteString("\n")
	}

	// Dependencies
	if len(p.viewModel.Dependencies) > 0 {
		depStr := strings.Join(p.viewModel.DependencyLabels, ", ")
		b.WriteString(components.Styles.MetadataStyle.Render(fmt.Sprintf("Dependencies: %s", depStr)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Progress bar
	progressText := fmt.Sprintf("Progress: %d/%d tasks (%.0f%%)",
		p.viewModel.Progress.Completed,
		p.viewModel.Progress.Total,
		p.viewModel.Progress.Percent*100)
	b.WriteString(components.Styles.ProgressStyle.Render(progressText))
	b.WriteString("\n\n")

	// Render tasks
	p.renderTasksView(&b)

	// Help view
	b.WriteString("\n")
	if p.showFullHelp {
		b.WriteString(p.help.FullHelpView(p.keys.FullHelp()))
	} else {
		b.WriteString(p.help.ShortHelpView(p.keys.ShortHelp()))
	}

	return b.String()
}

func (p *TrackDetailPresenter) renderTasksView(b *strings.Builder) {
	// Build flat task list with section info
	type taskItem struct {
		task        *viewmodels.TrackDetailTaskViewModel
		section     string
		sectionName string
	}
	allTasks := make([]taskItem, 0)

	for _, task := range p.viewModel.TODOTasks {
		allTasks = append(allTasks, taskItem{task: task, section: "todo", sectionName: "TODO"})
	}
	for _, task := range p.viewModel.InProgressTasks {
		allTasks = append(allTasks, taskItem{task: task, section: "in-progress", sectionName: "IN PROGRESS"})
	}
	for _, task := range p.viewModel.DoneTasks {
		allTasks = append(allTasks, taskItem{task: task, section: "done", sectionName: "DONE"})
	}

	if len(allTasks) == 0 {
		b.WriteString(components.Styles.MetadataStyle.Render("No tasks in this track"))
		return
	}

	// Get visible range from scroll helper
	start, end := p.scrollHelper.VisibleRange(len(allTasks))

	// Scroll indicator (above)
	if start > 0 {
		b.WriteString(components.Styles.MetadataStyle.Render("  ↑ More tasks above"))
		b.WriteString("\n")
	}

	// Render visible tasks with section headers
	currentSection := ""
	for i := start; i < end; i++ {
		item := allTasks[i]

		// Render section header if new section
		if item.section != currentSection {
			currentSection = item.section
			b.WriteString(components.Styles.SectionStyle.Render(item.sectionName))
			b.WriteString("\n")
		}

		// Render task
		var output string
		if i == p.selectedIndex {
			output = components.Styles.SelectedStyle.Render(fmt.Sprintf("  %s: %s", item.task.ID, item.task.Title))
		} else {
			output = fmt.Sprintf("  %s: %s", item.task.ID, item.task.Title)
		}
		b.WriteString(output)
		b.WriteString("\n")
	}

	// Scroll indicator (below)
	if end < len(allTasks) {
		b.WriteString(components.Styles.MetadataStyle.Render("  ↓ More tasks below"))
		b.WriteString("\n")
	}
}

func (p *TrackDetailPresenter) getMaxIndex() int {
	return len(p.viewModel.TODOTasks) +
		len(p.viewModel.InProgressTasks) +
		len(p.viewModel.DoneTasks) - 1
}

// getSelectedTaskID returns the task ID of the currently selected task
func (p *TrackDetailPresenter) getSelectedTaskID() string {
	index := p.selectedIndex
	todoLen := len(p.viewModel.TODOTasks)
	inProgressLen := len(p.viewModel.InProgressTasks)

	if index < todoLen {
		return p.viewModel.TODOTasks[index].ID
	}
	index -= todoLen

	if index < inProgressLen {
		return p.viewModel.InProgressTasks[index].ID
	}
	index -= inProgressLen

	if index < len(p.viewModel.DoneTasks) {
		return p.viewModel.DoneTasks[index].ID
	}

	return ""
}

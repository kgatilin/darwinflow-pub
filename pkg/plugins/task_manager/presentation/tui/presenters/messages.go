package presenters

import tea "github.com/charmbracelet/bubbletea"

// IterationSelectedMsg is sent when a user selects an iteration on the dashboard
type IterationSelectedMsg struct {
	IterationNumber int
	SelectedIndex   int // Dashboard selected index (for restoring focus on return)
}

// TrackSelectedMsg is sent when a user selects a track on the dashboard
type TrackSelectedMsg struct {
	TrackID       string
	SelectedIndex int // Dashboard selected index (for restoring focus on return)
}

// TaskSelectedMsg is sent when a user selects a task
type TaskSelectedMsg struct {
	TaskID        string
	SelectedIndex int // Dashboard selected index (for restoring focus on return)
}

// ErrorMsg is sent when an error occurs during loading or operations
type ErrorMsg struct {
	Err error
}

// ACActionCompletedMsg is sent after a successful AC action (verify/skip/fail)
type ACActionCompletedMsg struct {
	ActiveTab     IterationDetailTab // Preserve active tab (Tasks=0, ACs=1)
	SelectedIndex int                // Preserve selected index across reload
}

// ReorderCompletedMsg is sent after iterations are successfully reordered
type ReorderCompletedMsg struct {
	SelectedIterationNumber int
}

// RefreshDashboardMsg is sent when user requests dashboard refresh (r key)
type RefreshDashboardMsg struct {
	SelectedIndex int // Preserve selected index across reload
}

// Ensure these are valid Bubble Tea messages
var (
	_ tea.Msg = IterationSelectedMsg{}
	_ tea.Msg = TrackSelectedMsg{}
	_ tea.Msg = TaskSelectedMsg{}
	_ tea.Msg = ErrorMsg{}
	_ tea.Msg = ACActionCompletedMsg{}
	_ tea.Msg = ReorderCompletedMsg{}
	_ tea.Msg = RefreshDashboardMsg{}
)

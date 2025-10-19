package tui

import (
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// ViewState represents the current view in the TUI
type ViewState int

const (
	ViewSessionList ViewState = iota
	ViewSessionDetail
	ViewAnalysisAction
	ViewSaveDialog
	ViewProgress
)

// SessionInfo contains displayable information about a session
type SessionInfo struct {
	SessionID       string
	ShortID         string // First 8 chars
	FirstEvent      time.Time
	LastEvent       time.Time
	EventCount      int
	AnalysisCount   int
	Analyses        []*domain.SessionAnalysis
	HasAnalysis     bool
	LatestAnalysis  *domain.SessionAnalysis
	AnalysisTypes   []string // List of analysis prompt names
}

// Message types for Bubble Tea updates

type SessionsLoadedMsg struct {
	Sessions []*SessionInfo
	Error    error
}

type AnalysisCompleteMsg struct {
	SessionID string
	Analysis  *domain.SessionAnalysis
	Error     error
}

type SaveCompleteMsg struct {
	FilePath string
	Error    error
}

type ErrorMsg struct {
	Error error
}

package app

import (
	"context"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// MockLLMExecutor is a mock implementation for testing
type MockLLMExecutor struct {
	Response string
	Error    error
}

func (m *MockLLMExecutor) Execute(ctx context.Context, prompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}

// MockAnalysisRepository is a mock for testing
type MockAnalysisRepository struct {
	SavedAnalyses   []*domain.SessionAnalysis
	UnanalyzedIDs   []string
	AnalysisByID    map[string]*domain.SessionAnalysis
	SaveError       error
	GetError        error
	UnanalyzedError error
}

func NewMockAnalysisRepository() *MockAnalysisRepository {
	return &MockAnalysisRepository{
		AnalysisByID: make(map[string]*domain.SessionAnalysis),
	}
}

func (m *MockAnalysisRepository) SaveAnalysis(ctx context.Context, analysis *domain.SessionAnalysis) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.SavedAnalyses = append(m.SavedAnalyses, analysis)
	m.AnalysisByID[analysis.SessionID] = analysis
	return nil
}

func (m *MockAnalysisRepository) GetAnalysisBySessionID(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	return m.AnalysisByID[sessionID], nil
}

func (m *MockAnalysisRepository) GetUnanalyzedSessionIDs(ctx context.Context) ([]string, error) {
	if m.UnanalyzedError != nil {
		return nil, m.UnanalyzedError
	}
	return m.UnanalyzedIDs, nil
}

func (m *MockAnalysisRepository) GetAllAnalyses(ctx context.Context, limit int) ([]*domain.SessionAnalysis, error) {
	return m.SavedAnalyses, nil
}

func (m *MockAnalysisRepository) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	var sessionIDs []string
	seen := make(map[string]bool)
	for _, analysis := range m.SavedAnalyses {
		if !seen[analysis.SessionID] {
			sessionIDs = append(sessionIDs, analysis.SessionID)
			seen[analysis.SessionID] = true
		}
	}
	return sessionIDs, nil
}

func TestGetAnalysisPrompt(t *testing.T) {
	sessionData := "## Session Data\n- Tool: Read\n- File: test.go"
	prompt := GetAnalysisPrompt(sessionData)

	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	if len(prompt) < len(DefaultAnalysisPrompt)+len(sessionData) {
		t.Error("Prompt should contain both template and session data")
	}

	// Should contain the session data
	if !contains(prompt, sessionData) {
		t.Error("Prompt should contain session data")
	}

	// Should contain key analysis instructions
	if !contains(prompt, "What Made Me Inefficient") {
		t.Error("Prompt should contain analysis structure")
	}
}

func TestClaudeCLIExecutor_Integration(t *testing.T) {
	// This is a basic integration test that just checks the executor can be created
	// We don't actually run claude here as it may not be available in test environment
	executor := NewClaudeCLIExecutor(&NoOpLogger{})
	if executor == nil {
		t.Error("Expected non-nil executor")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

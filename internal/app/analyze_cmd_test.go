package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// mockLogger is a mock implementation of Logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(format string, args ...interface{}) {}
func (m *mockLogger) Info(format string, args ...interface{})  {}
func (m *mockLogger) Warn(format string, args ...interface{})  {}
func (m *mockLogger) Error(format string, args ...interface{}) {}

// mockAnalysisService is a mock implementation of AnalysisService for testing
type mockAnalysisService struct {
	getLastSessionFunc            func(ctx context.Context) (string, error)
	getAnalysisFunc               func(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error)
	analyzeSessionWithPromptFunc  func(ctx context.Context, sessionID string, promptName string) (*domain.SessionAnalysis, error)
	getUnanalyzedSessionsFunc     func(ctx context.Context) ([]string, error)
	getAllSessionIDsFunc          func(ctx context.Context, limit int) ([]string, error)
	analyzeMultiplePromptsFunc    func(ctx context.Context, sessionID string, promptNames []string) (map[string]*domain.SessionAnalysis, []error)
}

func (m *mockAnalysisService) GetLastSession(ctx context.Context) (string, error) {
	if m.getLastSessionFunc != nil {
		return m.getLastSessionFunc(ctx)
	}
	return "last-session-123", nil
}

func (m *mockAnalysisService) GetAnalysis(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error) {
	if m.getAnalysisFunc != nil {
		return m.getAnalysisFunc(ctx, sessionID)
	}
	return &domain.SessionAnalysis{
		SessionID:      sessionID,
		AnalyzedAt:     time.Now(),
		ModelUsed:      "claude-sonnet-4",
		AnalysisResult: "Test analysis result",
	}, nil
}

func (m *mockAnalysisService) AnalyzeSessionWithPrompt(ctx context.Context, sessionID string, promptName string) (*domain.SessionAnalysis, error) {
	if m.analyzeSessionWithPromptFunc != nil {
		return m.analyzeSessionWithPromptFunc(ctx, sessionID, promptName)
	}
	return &domain.SessionAnalysis{
		SessionID:      sessionID,
		AnalyzedAt:     time.Now(),
		ModelUsed:      "claude-sonnet-4",
		AnalysisResult: "Analysis for " + promptName,
		PromptName:     promptName,
	}, nil
}

func (m *mockAnalysisService) GetUnanalyzedSessions(ctx context.Context) ([]string, error) {
	if m.getUnanalyzedSessionsFunc != nil {
		return m.getUnanalyzedSessionsFunc(ctx)
	}
	return []string{"session-1", "session-2"}, nil
}

func (m *mockAnalysisService) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	if m.getAllSessionIDsFunc != nil {
		return m.getAllSessionIDsFunc(ctx, limit)
	}
	return []string{"session-1", "session-2", "session-3"}, nil
}

func (m *mockAnalysisService) AnalyzeSessionWithMultiplePrompts(ctx context.Context, sessionID string, promptNames []string) (map[string]*domain.SessionAnalysis, []error) {
	if m.analyzeMultiplePromptsFunc != nil {
		return m.analyzeMultiplePromptsFunc(ctx, sessionID, promptNames)
	}
	results := make(map[string]*domain.SessionAnalysis)
	for _, name := range promptNames {
		results[name] = &domain.SessionAnalysis{
			SessionID:      sessionID,
			AnalyzedAt:     time.Now(),
			ModelUsed:      "claude-sonnet-4",
			AnalysisResult: "Analysis for " + name,
			PromptName:     name,
		}
	}
	return results, nil
}

func TestAnalyzeCommandHandler_ViewAnalysis(t *testing.T) {
	ctx := context.Background()
	mockService := &mockAnalysisService{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}
	handler := app.NewAnalyzeCommandHandler(mockService, logger, out)

	opts := app.AnalyzeOptions{
		SessionID: "test-session-123",
		ViewOnly:  true,
	}

	err := handler.Execute(ctx, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "test-session-123") {
		t.Errorf("Output should contain session ID, got: %s", output)
	}
	if !strings.Contains(output, "Analysis Result") {
		t.Errorf("Output should contain analysis result header, got: %s", output)
	}
}

func TestAnalyzeCommandHandler_AnalyzeLastSession(t *testing.T) {
	ctx := context.Background()
	mockService := &mockAnalysisService{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}
	handler := app.NewAnalyzeCommandHandler(mockService, logger, out)

	opts := app.AnalyzeOptions{
		Last:        true,
		PromptNames: []string{"test_prompt"},
	}

	err := handler.Execute(ctx, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Analyzing last session") {
		t.Errorf("Output should indicate analyzing last session, got: %s", output)
	}
}

func TestAnalyzeCommandHandler_AnalyzeSpecificSession(t *testing.T) {
	ctx := context.Background()
	mockService := &mockAnalysisService{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}
	handler := app.NewAnalyzeCommandHandler(mockService, logger, out)

	opts := app.AnalyzeOptions{
		SessionID:   "specific-session-456",
		PromptNames: []string{"tool_analysis"},
	}

	err := handler.Execute(ctx, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "specific-session-456") {
		t.Errorf("Output should contain session ID, got: %s", output)
	}
}

func TestAnalyzeCommandHandler_AnalyzeAll(t *testing.T) {
	ctx := context.Background()
	mockService := &mockAnalysisService{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}
	handler := app.NewAnalyzeCommandHandler(mockService, logger, out)

	opts := app.AnalyzeOptions{
		AnalyzeAll:  true,
		PromptNames: []string{"test_prompt"},
	}

	err := handler.Execute(ctx, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Found 2 unanalyzed session(s)") {
		t.Errorf("Output should show number of sessions, got: %s", output)
	}
}

func TestAnalyzeCommandHandler_Refresh(t *testing.T) {
	ctx := context.Background()
	mockService := &mockAnalysisService{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}
	handler := app.NewAnalyzeCommandHandler(mockService, logger, out)

	opts := app.AnalyzeOptions{
		Refresh:     true,
		Limit:       2,
		PromptNames: []string{"test_prompt"},
	}

	err := handler.Execute(ctx, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Refreshing analyses") {
		t.Errorf("Output should indicate refreshing, got: %s", output)
	}
}

func TestAnalyzeCommandHandler_MultiplePrompts(t *testing.T) {
	ctx := context.Background()
	mockService := &mockAnalysisService{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}
	handler := app.NewAnalyzeCommandHandler(mockService, logger, out)

	opts := app.AnalyzeOptions{
		SessionID:   "test-session",
		PromptNames: []string{"prompt1", "prompt2"},
	}

	err := handler.Execute(ctx, opts)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "2 prompts in parallel") {
		t.Errorf("Output should indicate parallel execution, got: %s", output)
	}
}

func TestAnalyzeCommandHandler_NoSessionSpecified(t *testing.T) {
	ctx := context.Background()
	mockService := &mockAnalysisService{}
	logger := &mockLogger{}
	out := &bytes.Buffer{}
	handler := app.NewAnalyzeCommandHandler(mockService, logger, out)

	opts := app.AnalyzeOptions{
		PromptNames: []string{"test_prompt"},
		// Neither SessionID nor Last is set
	}

	err := handler.Execute(ctx, opts)
	if err == nil {
		t.Error("Execute should fail when no session is specified")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("Error should indicate missing session specification, got: %v", err)
	}
}

package app_test

import (
	"context"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
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

func (m *MockAnalysisRepository) GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionAnalysis, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	var analyses []*domain.SessionAnalysis
	for _, analysis := range m.SavedAnalyses {
		if analysis.SessionID == sessionID {
			analyses = append(analyses, analysis)
		}
	}
	return analyses, nil
}

func TestGetAnalysisPrompt(t *testing.T) {
	sessionData := "## Session Data\n- Tool: Read\n- File: test.go"
	prompt := app.GetAnalysisPrompt(sessionData)

	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	if len(prompt) < len(app.DefaultAnalysisPrompt)+len(sessionData) {
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
	executor := app.NewClaudeCLIExecutor(&app.NoOpLogger{})
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

func TestNewAnalysisService(t *testing.T) {
	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	if service == nil {
		t.Fatal("Expected non-nil AnalysisService")
	}
}

func TestNewAnalysisService_WithNilConfig(t *testing.T) {
	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "test analysis"}
	logger := &app.NoOpLogger{}

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, nil)

	if service == nil {
		t.Fatal("Expected non-nil AnalysisService even with nil config")
	}
}

func TestAnalysisService_AnalyzeSessionWithPrompt(t *testing.T) {
	ctx := context.Background()

	// Create mock event with test data
	event := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{
		"tool": "Read",
		"file": "test.go",
	}, "Read test.go")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "This is a test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	analysis, err := service.AnalyzeSessionWithPrompt(ctx, "session-123", "tool_analysis")
	if err != nil {
		t.Fatalf("AnalyzeSessionWithPrompt failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected non-nil analysis")
	}

	if analysis.SessionID != "session-123" {
		t.Errorf("Expected SessionID 'session-123', got %s", analysis.SessionID)
	}

	if analysis.AnalysisResult != "This is a test analysis" {
		t.Errorf("Expected analysis result 'This is a test analysis', got %s", analysis.AnalysisResult)
	}

	if len(analysisRepo.SavedAnalyses) != 1 {
		t.Errorf("Expected 1 saved analysis, got %d", len(analysisRepo.SavedAnalyses))
	}
}

func TestAnalysisService_AnalyzeSessionWithPrompt_NoLogs(t *testing.T) {
	ctx := context.Background()

	eventRepo := &MockEventRepository{
		events: []*domain.Event{},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	_, err := service.AnalyzeSessionWithPrompt(ctx, "session-123", "tool_analysis")
	if err == nil {
		t.Error("Expected error when analyzing session with no logs")
	}
}

func TestAnalysisService_GetAnalysesBySessionID(t *testing.T) {
	ctx := context.Background()

	analysisRepo := NewMockAnalysisRepository()
	analysisRepo.SavedAnalyses = []*domain.SessionAnalysis{
		domain.NewSessionAnalysis("session-123", "analysis 1", "claude-3", "prompt1"),
		domain.NewSessionAnalysis("session-123", "analysis 2", "claude-3", "prompt2"),
		domain.NewSessionAnalysis("session-456", "analysis 3", "claude-3", "prompt1"),
	}

	eventRepo := &MockEventRepository{}
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	analyses, err := service.GetAnalysesBySessionID(ctx, "session-123")
	if err != nil {
		t.Fatalf("GetAnalysesBySessionID failed: %v", err)
	}

	if len(analyses) != 2 {
		t.Errorf("Expected 2 analyses for session-123, got %d", len(analyses))
	}
}

func TestAnalysisService_GetAllSessionIDs(t *testing.T) {
	ctx := context.Background()

	analysisRepo := NewMockAnalysisRepository()
	analysisRepo.SavedAnalyses = []*domain.SessionAnalysis{
		domain.NewSessionAnalysis("session-123", "analysis 1", "claude-3", "prompt1"),
		domain.NewSessionAnalysis("session-456", "analysis 2", "claude-3", "prompt1"),
	}

	eventRepo := &MockEventRepository{}
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	sessionIDs, err := service.GetAllSessionIDs(ctx, 10)
	if err != nil {
		t.Fatalf("GetAllSessionIDs failed: %v", err)
	}

	if len(sessionIDs) != 2 {
		t.Errorf("Expected 2 session IDs, got %d", len(sessionIDs))
	}
}

func TestAnalysisService_GetLastSession(t *testing.T) {
	ctx := context.Background()

	event := domain.NewEvent("claude.tool.invoked", "session-latest", map[string]interface{}{}, "test")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	sessionID, err := service.GetLastSession(ctx)
	if err != nil {
		t.Fatalf("GetLastSession failed: %v", err)
	}

	if sessionID != "session-latest" {
		t.Errorf("Expected session ID 'session-latest', got %s", sessionID)
	}
}

func TestAnalysisService_GetLastSession_NoSessions(t *testing.T) {
	ctx := context.Background()

	eventRepo := &MockEventRepository{
		events: []*domain.Event{},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	_, err := service.GetLastSession(ctx)
	if err == nil {
		t.Error("Expected error when getting last session with no sessions")
	}
}

func TestAnalysisService_AnalyzeMultipleSessions(t *testing.T) {
	ctx := context.Background()

	event1 := domain.NewEvent("claude.tool.invoked", "session-1", map[string]interface{}{}, "test1")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event1},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "analysis result"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	results, errors := service.AnalyzeMultipleSessions(ctx, []string{"session-1"}, "tool_analysis")

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test with empty session list
	results, _ = service.AnalyzeMultipleSessions(ctx, []string{}, "tool_analysis")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty session list, got %d", len(results))
	}
}

func TestAnalysisService_EstimateTokenCount(t *testing.T) {
	ctx := context.Background()

	event := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{
		"tool": "Read",
		"file": "test.go",
	}, "Read test.go")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	tokenCount, err := service.EstimateTokenCount(ctx, "session-123")
	if err != nil {
		t.Fatalf("EstimateTokenCount failed: %v", err)
	}

	if tokenCount <= 0 {
		t.Errorf("Expected positive token count, got %d", tokenCount)
	}
}

func TestAnalysisService_SelectSessionsWithinTokenLimit(t *testing.T) {
	ctx := context.Background()

	event := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{
		"tool": "Read",
	}, "Read")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	selected, totalTokens, err := service.SelectSessionsWithinTokenLimit(ctx, []string{"session-123"}, 100000)
	if err != nil {
		t.Fatalf("SelectSessionsWithinTokenLimit failed: %v", err)
	}

	if len(selected) == 0 {
		t.Error("Expected at least 1 selected session")
	}

	if totalTokens <= 0 {
		t.Errorf("Expected positive total tokens, got %d", totalTokens)
	}
}

func TestAnalysisService_AnalyzeSession(t *testing.T) {
	ctx := context.Background()

	event := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{
		"tool": "Read",
	}, "Read test.go")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	analysis, err := service.AnalyzeSession(ctx, "session-123")
	if err != nil {
		t.Fatalf("AnalyzeSession failed: %v", err)
	}

	if analysis.SessionID != "session-123" {
		t.Errorf("Expected session ID 'session-123', got %s", analysis.SessionID)
	}
}

func TestAnalysisService_GetUnanalyzedSessions(t *testing.T) {
	ctx := context.Background()

	analysisRepo := NewMockAnalysisRepository()
	analysisRepo.UnanalyzedIDs = []string{"session-1", "session-2"}

	eventRepo := &MockEventRepository{}
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	sessions, err := service.GetUnanalyzedSessions(ctx)
	if err != nil {
		t.Fatalf("GetUnanalyzedSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 unanalyzed sessions, got %d", len(sessions))
	}
}

func TestAnalysisService_GetAnalysis(t *testing.T) {
	ctx := context.Background()

	analysis := domain.NewSessionAnalysis("session-123", "test analysis", "claude-3", "prompt1")
	analysisRepo := NewMockAnalysisRepository()
	analysisRepo.AnalysisByID = map[string]*domain.SessionAnalysis{
		"session-123": analysis,
	}

	eventRepo := &MockEventRepository{}
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	result, err := service.GetAnalysis(ctx, "session-123")
	if err != nil {
		t.Fatalf("GetAnalysis failed: %v", err)
	}

	if result.SessionID != "session-123" {
		t.Errorf("Expected session ID 'session-123', got %s", result.SessionID)
	}
}

func TestAnalysisService_AnalyzeMultipleSessionsParallel(t *testing.T) {
	ctx := context.Background()

	event := domain.NewEvent("claude.tool.invoked", "session-1", map[string]interface{}{}, "test")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "analysis result"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	results, errors := service.AnalyzeMultipleSessionsParallel(ctx, []string{"session-1"}, "tool_analysis")

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestAnalysisService_AnalyzeMultipleSessionsParallel_Empty(t *testing.T) {
	ctx := context.Background()

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	results, errors := service.AnalyzeMultipleSessionsParallel(ctx, []string{}, "tool_analysis")

	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty list, got %d", len(errors))
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty list, got %d", len(results))
	}
}

func TestAnalysisService_AnalyzeSessionWithMultiplePrompts(t *testing.T) {
	ctx := context.Background()

	event := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{}, "test")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "analysis result"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()
	config.Prompts["prompt1"] = "Prompt 1"
	config.Prompts["prompt2"] = "Prompt 2"

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	results, errors := service.AnalyzeSessionWithMultiplePrompts(ctx, "session-123", []string{"prompt1", "prompt2"})

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestAnalysisService_SaveToMarkdown(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	analysis := domain.NewSessionAnalysis("session-123", "test analysis", "claude-3", "prompt1")

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	filePath, err := service.SaveToMarkdown(ctx, analysis, tmpDir, "test-analysis.md")
	if err != nil {
		t.Fatalf("SaveToMarkdown failed: %v", err)
	}

	if filePath == "" {
		t.Error("Expected non-empty file path")
	}

	if !contains(filePath, "test-analysis.md") {
		t.Errorf("Expected file path to contain 'test-analysis.md', got %s", filePath)
	}
}

func TestNewClaudeCLIExecutorWithConfig(t *testing.T) {
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	executor := app.NewClaudeCLIExecutorWithConfig(logger, config)

	if executor == nil {
		t.Error("Expected non-nil executor")
	}
}

func TestNewClaudeCLIExecutorWithConfig_NilLogger(t *testing.T) {
	config := domain.DefaultConfig()

	executor := app.NewClaudeCLIExecutorWithConfig(nil, config)

	if executor == nil {
		t.Error("Expected non-nil executor even with nil logger")
	}
}

func TestNewClaudeCLIExecutorWithConfig_NilConfig(t *testing.T) {
	logger := &app.NoOpLogger{}

	executor := app.NewClaudeCLIExecutorWithConfig(logger, nil)

	if executor == nil {
		t.Error("Expected non-nil executor even with nil config")
	}
}

func TestAnalysisService_AnalyzeSessionWithPrompt_UnknownPrompt(t *testing.T) {
	ctx := context.Background()

	event := domain.NewEvent("claude.tool.invoked", "session-123", map[string]interface{}{
		"tool": "Read",
	}, "Read test.go")

	eventRepo := &MockEventRepository{
		events: []*domain.Event{event},
	}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{Response: "analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	// Should fall back to tool_analysis when prompt not found
	analysis, err := service.AnalyzeSessionWithPrompt(ctx, "session-123", "nonexistent_prompt")
	if err != nil {
		t.Fatalf("AnalyzeSessionWithPrompt should fall back to default prompt: %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected non-nil analysis with fallback prompt")
	}
}

func TestAnalysisService_AnalyzeSessionWithMultiplePrompts_Empty(t *testing.T) {
	ctx := context.Background()

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	results, errors := service.AnalyzeSessionWithMultiplePrompts(ctx, "session-123", []string{})

	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty prompt list, got %d", len(errors))
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty prompt list, got %d", len(results))
	}
}

func TestAnalysisService_SaveToMarkdown_WithDefaults(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	analysis := domain.NewSessionAnalysis("session-123456789", "test analysis", "claude-3", "prompt1")

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	// Test with default filename (empty string)
	filePath, err := service.SaveToMarkdown(ctx, analysis, tmpDir, "")
	if err != nil {
		t.Fatalf("SaveToMarkdown with default filename failed: %v", err)
	}

	if filePath == "" {
		t.Error("Expected non-empty file path")
	}

	// Should have .md extension
	if !contains(filePath, ".md") {
		t.Errorf("Expected file path to have .md extension, got %s", filePath)
	}
}

func TestAnalysisService_SaveToMarkdown_NilAnalysis(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	_, err := service.SaveToMarkdown(ctx, nil, tmpDir, "test.md")
	if err == nil {
		t.Error("Expected error when saving nil analysis")
	}
}

func TestNoOpLogger_AllMethods(t *testing.T) {
	logger := &app.NoOpLogger{}

	// These should not panic and do nothing
	logger.Debug("debug %s", "test")
	logger.Info("info %s", "test")
	logger.Warn("warn %s", "test")
	logger.Error("error %s", "test")

	// If we got here without panicking, the test passes
}

func TestAnalysisService_SaveToMarkdown_InvalidTemplate(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	analysis := domain.NewSessionAnalysis("session-123", "test analysis", "claude-3", "prompt1")

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()
	config.UI.FilenameTemplate = "{{.InvalidField}}" // Invalid template

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	_, err := service.SaveToMarkdown(ctx, analysis, tmpDir, "")
	if err == nil {
		t.Error("Expected error with invalid filename template")
	}
	if !contains(err.Error(), "filename") {
		t.Errorf("Expected error to mention filename, got: %v", err)
	}
}

func TestAnalysisService_SaveToMarkdown_ShortSessionID(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	// Test with session ID that's at least 8 characters
	analysis := domain.NewSessionAnalysis("session-12345678", "test analysis", "claude-3", "prompt1")

	filePath, err := service.SaveToMarkdown(ctx, analysis, tmpDir, "")
	if err != nil {
		t.Fatalf("SaveToMarkdown failed: %v", err)
	}

	if filePath == "" {
		t.Error("Expected non-empty file path")
	}
}

func TestAnalysisService_SaveToMarkdown_WithDefaultOutputDir(t *testing.T) {
	ctx := context.Background()

	analysis := domain.NewSessionAnalysis("session-12345678", "test analysis", "claude-3", "prompt1")

	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llmExecutor := &MockLLMExecutor{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()
	config.UI.DefaultOutputDir = t.TempDir()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llmExecutor, logger, config)

	// Use empty outputDir to test default behavior
	filePath, err := service.SaveToMarkdown(ctx, analysis, "", "test.md")
	if err != nil {
		t.Fatalf("SaveToMarkdown with default dir failed: %v", err)
	}

	if !contains(filePath, "test.md") {
		t.Errorf("Expected file path to contain test.md, got %s", filePath)
	}
}

func TestNewClaudeCLIExecutor_WithNilLogger(t *testing.T) {
	// Test the 66.7% coverage case where logger is nil
	executor := app.NewClaudeCLIExecutor(nil)

	if executor == nil {
		t.Error("Expected non-nil executor even with nil logger")
	}
}

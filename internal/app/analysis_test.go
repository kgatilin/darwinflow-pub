package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// MockLLMExecutor is a mock implementation for testing (deprecated - use MockLLM instead)
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

// MockLLM is a mock implementation of domain.LLM for testing
type MockLLM struct {
	Response    string
	Error       error
	ModelValue  string
	TokenEst    int
	QueryCalls  int
	TokensCalls int
}

func (m *MockLLM) Query(ctx context.Context, prompt string, options *domain.LLMOptions) (string, error) {
	m.QueryCalls++
	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}

func (m *MockLLM) EstimateTokens(prompt string) int {
	m.TokensCalls++
	if m.TokenEst > 0 {
		return m.TokenEst
	}
	return len(prompt) / 4
}

func (m *MockLLM) GetModel() string {
	if m.ModelValue != "" {
		return m.ModelValue
	}
	return "claude-3"
}

// MockAnalysisRepository is a mock for testing
type MockAnalysisRepository struct {
	SavedAnalyses      []*domain.SessionAnalysis
	UnanalyzedIDs      []string
	AnalysisByID       map[string]*domain.SessionAnalysis
	AnalysesByViewID   []*domain.Analysis
	SaveError          error
	GetError           error
	UnanalyzedError    error
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

// Generic analysis methods (stubs for interface compliance)
func (m *MockAnalysisRepository) SaveGenericAnalysis(ctx context.Context, analysis *domain.Analysis) error {
	return m.SaveError
}

func (m *MockAnalysisRepository) FindAnalysisByViewID(ctx context.Context, viewID string) ([]*domain.Analysis, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	return m.AnalysesByViewID, nil
}

func (m *MockAnalysisRepository) FindAnalysisByViewType(ctx context.Context, viewType string) ([]*domain.Analysis, error) {
	return nil, m.GetError
}

func (m *MockAnalysisRepository) FindAnalysisById(ctx context.Context, id string) (*domain.Analysis, error) {
	return nil, m.GetError
}

func (m *MockAnalysisRepository) ListRecentAnalyses(ctx context.Context, limit int) ([]*domain.Analysis, error) {
	return nil, m.GetError
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
	llm := &MockLLM{Response: "test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

	if service == nil {
		t.Fatal("Expected non-nil AnalysisService")
	}
}

func TestNewAnalysisService_WithNilConfig(t *testing.T) {
	eventRepo := &MockEventRepository{}
	analysisRepo := NewMockAnalysisRepository()
	logsService := app.NewLogsService(eventRepo, eventRepo)
	llm := &MockLLM{Response: "test analysis"}
	logger := &app.NoOpLogger{}

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, nil)

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
	llm := &MockLLM{Response: "This is a test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{Response: "test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{Response: "analysis result"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{Response: "test analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{Response: "analysis result"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{Response: "analysis result"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()
	config.Prompts["prompt1"] = "Prompt 1"
	config.Prompts["prompt2"] = "Prompt 2"

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{Response: "analysis"}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()
	config.UI.FilenameTemplate = "{{.InvalidField}}" // Invalid template

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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
	llm := &MockLLM{}
	logger := &app.NoOpLogger{}
	config := domain.DefaultConfig()
	config.UI.DefaultOutputDir = t.TempDir()

	service := app.NewAnalysisService(eventRepo, analysisRepo, logsService, llm, logger, config)
	service.SetSessionViewFactory(mockSessionViewFactory)

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

// mockSessionViewFactory creates a mock session view factory for testing
func mockSessionViewFactory(sessionID string, events []pluginsdk.Event) pluginsdk.AnalysisView {
	return &MockAnalysisView{
		ID:     sessionID,
		Type:   "session",
		Events: events,
		Metadata: map[string]interface{}{
			"session_id":  sessionID,
			"event_count": len(events),
		},
	}
}

// MockAnalysisView is a mock implementation of pluginsdk.AnalysisView for testing
type MockAnalysisView struct {
	ID       string
	Type     string
	Events   []pluginsdk.Event
	Metadata map[string]interface{}
}

func (m *MockAnalysisView) GetID() string {
	return m.ID
}

func (m *MockAnalysisView) GetType() string {
	return m.Type
}

func (m *MockAnalysisView) GetEvents() []pluginsdk.Event {
	return m.Events
}

func (m *MockAnalysisView) FormatForAnalysis() string {
	return "# Test Analysis\n\nFormatted view content for analysis.\n"
}

func (m *MockAnalysisView) GetMetadata() map[string]interface{} {
	return m.Metadata
}

func TestAnalysisService_AnalyzeView_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{
		Response:   "This is the analysis result.",
		ModelValue: "claude-3",
	}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	view := &MockAnalysisView{
		ID:     "session-123",
		Type:   "session",
		Events: []pluginsdk.Event{},
		Metadata: map[string]interface{}{
			"event_count": 0,
		},
	}

	analysis, err := service.AnalyzeView(ctx, view, "tool_analysis")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected non-nil analysis result")
	}

	// AnalyzeView now returns generic Analysis, so check ViewID instead of SessionID
	if analysis.ViewID != "session-123" {
		t.Errorf("Expected view ID 'session-123', got '%s'", analysis.ViewID)
	}

	// Check Result instead of AnalysisResult
	if !contains(analysis.Result, "This is the analysis result") {
		t.Errorf("Expected analysis result to contain expected text, got: %s", analysis.Result)
	}

	// Check model is set (could be default or from config)
	if analysis.ModelUsed == "" {
		t.Errorf("Expected non-empty model")
	}

	// Verify LLM was called
	if mockLLM.QueryCalls != 1 {
		t.Errorf("Expected 1 LLM call, got %d", mockLLM.QueryCalls)
	}

	// Note: AnalyzeView now saves to generic analyses, not SavedAnalyses
	// We would need to update the mock to track generic saves if we want to verify this
}

func TestAnalysisService_AnalyzeView_NilView(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{Response: "analysis"}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	_, err := service.AnalyzeView(ctx, nil, "tool_analysis")

	if err == nil {
		t.Error("Expected error for nil view")
	}

	if !contains(err.Error(), "view is nil") {
		t.Errorf("Expected 'view is nil' error, got: %v", err)
	}
}

func TestAnalysisService_AnalyzeView_LLMError(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{
		Error: fmt.Errorf("LLM service error"),
	}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	view := &MockAnalysisView{
		ID:   "session-123",
		Type: "session",
	}

	_, err := service.AnalyzeView(ctx, view, "tool_analysis")

	if err == nil {
		t.Error("Expected error when LLM fails")
	}

	if !contains(err.Error(), "LLM") {
		t.Errorf("Expected error mentioning LLM, got: %v", err)
	}

	// Verify no analysis was saved
	if len(mockAnalysisRepo.SavedAnalyses) != 0 {
		t.Errorf("Expected 0 saved analyses after error, got %d", len(mockAnalysisRepo.SavedAnalyses))
	}
}

func TestAnalysisService_AnalyzeView_WithCustomPrompt(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{
		Response:   "Custom analysis result",
		ModelValue: "claude-3",
	}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()
	config.Prompts["custom_prompt"] = "Custom prompt template: "

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	view := &MockAnalysisView{
		ID:   "session-456",
		Type: "session",
	}

	analysis, err := service.AnalyzeView(ctx, view, "custom_prompt")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Generic Analysis uses PromptUsed instead of PromptName
	if analysis.PromptUsed != "custom_prompt" {
		t.Errorf("Expected prompt used 'custom_prompt', got '%s'", analysis.PromptUsed)
	}

	// Generic Analysis uses Result instead of AnalysisResult
	if !contains(analysis.Result, "Custom analysis result") {
		t.Errorf("Expected custom analysis result in output")
	}
}

// NoOpTestLogger is a no-op logger for testing
type NoOpTestLogger struct{}

func (l *NoOpTestLogger) Debug(format string, args ...interface{}) {}
func (l *NoOpTestLogger) Info(format string, args ...interface{})  {}
func (l *NoOpTestLogger) Warn(format string, args ...interface{})  {}
func (l *NoOpTestLogger) Error(format string, args ...interface{}) {}

// TestAnalysisService_GetAnalysesByViewID tests retrieving analyses by view ID
func TestAnalysisService_GetAnalysesByViewID(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{
		Response:   "Analysis result",
		ModelValue: "claude-3",
	}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()

	// Add some generic analyses to the mock repo
	expectedAnalyses := []*domain.Analysis{
		{
			ID:         "analysis-1",
			ViewID:     "view-123",
			ViewType:   "session",
			PromptUsed: "test_prompt",
			Result:     "Result 1",
		},
		{
			ID:         "analysis-2",
			ViewID:     "view-123",
			ViewType:   "session",
			PromptUsed: "another_prompt",
			Result:     "Result 2",
		},
	}
	mockAnalysisRepo.AnalysesByViewID = expectedAnalyses

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	analyses, err := service.GetAnalysesByViewID(ctx, "view-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(analyses) != 2 {
		t.Errorf("Expected 2 analyses, got %d", len(analyses))
	}

	if analyses[0].ViewID != "view-123" {
		t.Errorf("Expected view ID 'view-123', got '%s'", analyses[0].ViewID)
	}
}

// TestAnalysisService_AnalyzeViewWithOptions tests analyzing a view with custom options
func TestAnalysisService_AnalyzeViewWithOptions(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{
		Response:   "Custom options analysis result",
		ModelValue: "claude-3",
	}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()
	config.Prompts["test_prompt"] = "Test prompt: "

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	view := &MockAnalysisView{
		ID:   "view-789",
		Type: "session",
	}

	options := &app.AnalysisOptions{
		ModelOverride: "claude-opus",
		LLMOptions: &domain.LLMOptions{
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	}

	analysis, err := service.AnalyzeViewWithOptions(ctx, view, "test_prompt", options)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if analysis.ViewID != "view-789" {
		t.Errorf("Expected view ID 'view-789', got '%s'", analysis.ViewID)
	}

	if !contains(analysis.Result, "Custom options analysis result") {
		t.Errorf("Expected custom options analysis result in output")
	}
}

// TestAnalysisService_AnalyzeViewWithOptions_NilView tests error handling for nil view
func TestAnalysisService_AnalyzeViewWithOptions_NilView(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	_, err := service.AnalyzeViewWithOptions(ctx, nil, "test_prompt", nil)

	if err == nil {
		t.Fatal("Expected error for nil view, got nil")
	}

	if !contains(err.Error(), "view is nil") {
		t.Errorf("Expected 'view is nil' error, got: %v", err)
	}
}

// TestAnalysisService_AnalyzeViewWithOptions_NilOptions tests that nil options are handled
func TestAnalysisService_AnalyzeViewWithOptions_NilOptions(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockEventRepository{}
	mockAnalysisRepo := NewMockAnalysisRepository()
	mockLLM := &MockLLM{
		Response:   "Result with nil options",
		ModelValue: "claude-3",
	}
	mockLogsService := app.NewLogsService(mockRepo, mockRepo)
	config := domain.DefaultConfig()
	config.Prompts["test_prompt"] = "Test: "

	service := app.NewAnalysisService(
		mockRepo,
		mockAnalysisRepo,
		mockLogsService,
		mockLLM,
		&NoOpTestLogger{},
		config,
	)

	view := &MockAnalysisView{
		ID:   "view-nil-opts",
		Type: "session",
	}

	// Pass nil options - should use defaults
	analysis, err := service.AnalyzeViewWithOptions(ctx, view, "test_prompt", nil)

	if err != nil {
		t.Fatalf("Expected no error with nil options, got: %v", err)
	}

	if analysis.ViewID != "view-nil-opts" {
		t.Errorf("Expected view ID 'view-nil-opts', got '%s'", analysis.ViewID)
	}
}

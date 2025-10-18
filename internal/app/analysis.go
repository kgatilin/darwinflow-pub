package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// NoOpLogger is a logger that does nothing (for backward compatibility)
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(format string, args ...interface{}) {}
func (l *NoOpLogger) Info(format string, args ...interface{})  {}
func (l *NoOpLogger) Warn(format string, args ...interface{})  {}
func (l *NoOpLogger) Error(format string, args ...interface{}) {}

// LLMExecutor defines the interface for executing LLM queries
type LLMExecutor interface {
	// Execute runs an LLM query with the given prompt and returns the response
	Execute(ctx context.Context, prompt string) (string, error)
}

// Logger interface for dependency injection
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// AnalysisService handles session analysis operations
type AnalysisService struct {
	eventRepo    domain.EventRepository
	analysisRepo domain.AnalysisRepository
	logsService  *LogsService
	llmExecutor  LLMExecutor
	logger       Logger
	config       *domain.Config
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(
	eventRepo domain.EventRepository,
	analysisRepo domain.AnalysisRepository,
	logsService *LogsService,
	llmExecutor LLMExecutor,
	logger Logger,
	config *domain.Config,
) *AnalysisService {
	if config == nil {
		config = domain.DefaultConfig()
	}
	return &AnalysisService{
		eventRepo:    eventRepo,
		analysisRepo: analysisRepo,
		logsService:  logsService,
		llmExecutor:  llmExecutor,
		logger:       logger,
		config:       config,
	}
}

// AnalyzeSession analyzes a specific session and stores the results
func (s *AnalysisService) AnalyzeSession(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error) {
	// Get session logs
	s.logger.Debug("Fetching logs for session %s", sessionID)
	logs, err := s.logsService.ListRecentLogs(ctx, 0, sessionID, true)
	if err != nil {
		s.logger.Error("Failed to get session logs: %v", err)
		return nil, fmt.Errorf("failed to get session logs: %w", err)
	}

	if len(logs) == 0 {
		s.logger.Warn("No logs found for session %s", sessionID)
		return nil, fmt.Errorf("no logs found for session %s", sessionID)
	}
	s.logger.Debug("Found %d log records for session %s", len(logs), sessionID)

	// Format logs as markdown
	s.logger.Debug("Formatting logs as markdown")
	var buf bytes.Buffer
	if err := FormatLogsAsMarkdown(&buf, logs); err != nil {
		s.logger.Error("Failed to format logs: %v", err)
		return nil, fmt.Errorf("failed to format logs: %w", err)
	}

	// Get analysis prompt from config
	promptTemplate, exists := s.config.Prompts["analysis"]
	if !exists || promptTemplate == "" {
		s.logger.Warn("Analysis prompt not found in config, using default")
		promptTemplate = domain.DefaultAnalysisPrompt
	}

	prompt := promptTemplate + buf.String()
	s.logger.Debug("Generated prompt with %d characters (%d KB)", len(prompt), len(prompt)/1024)

	// Execute LLM analysis
	s.logger.Info("Invoking Claude CLI for analysis...")
	analysisResult, err := s.llmExecutor.Execute(ctx, prompt)
	if err != nil {
		s.logger.Error("Failed to execute LLM analysis: %v", err)
		return nil, fmt.Errorf("failed to execute LLM analysis: %w", err)
	}
	s.logger.Debug("Claude CLI returned %d characters", len(analysisResult))

	// Create and save analysis
	s.logger.Debug("Saving analysis to database")
	analysis := domain.NewSessionAnalysis(
		sessionID,
		analysisResult,
		"claude", // model identifier
		promptTemplate,
	)

	if err := s.analysisRepo.SaveAnalysis(ctx, analysis); err != nil {
		s.logger.Error("Failed to save analysis: %v", err)
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	s.logger.Info("Analysis completed successfully")
	return analysis, nil
}

// GetLastSession returns the ID of the most recent session
func (s *AnalysisService) GetLastSession(ctx context.Context) (string, error) {
	logs, err := s.logsService.ListRecentLogs(ctx, 1, "", false)
	if err != nil {
		return "", fmt.Errorf("failed to get last session: %w", err)
	}

	if len(logs) == 0 {
		return "", fmt.Errorf("no sessions found")
	}

	return logs[0].SessionID, nil
}

// GetUnanalyzedSessions returns all session IDs that haven't been analyzed
func (s *AnalysisService) GetUnanalyzedSessions(ctx context.Context) ([]string, error) {
	return s.analysisRepo.GetUnanalyzedSessionIDs(ctx)
}

// GetAnalysis retrieves the analysis for a session
func (s *AnalysisService) GetAnalysis(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error) {
	return s.analysisRepo.GetAnalysisBySessionID(ctx, sessionID)
}

// GetAllSessionIDs retrieves all session IDs, ordered by most recent first
// If limit > 0, returns only the latest N sessions
func (s *AnalysisService) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	return s.analysisRepo.GetAllSessionIDs(ctx, limit)
}

// ClaudeCLIExecutor implements LLMExecutor using the claude CLI tool
type ClaudeCLIExecutor struct {
	logger Logger
}

// NewClaudeCLIExecutor creates a new Claude CLI executor
func NewClaudeCLIExecutor(logger Logger) *ClaudeCLIExecutor {
	if logger == nil {
		logger = &NoOpLogger{}
	}
	return &ClaudeCLIExecutor{
		logger: logger,
	}
}

// Execute runs claude -p with the given prompt
// Streams output to stderr in real-time for progress visibility
func (e *ClaudeCLIExecutor) Execute(ctx context.Context, prompt string) (string, error) {
	e.logger.Debug("Executing: claude -p <prompt of %d chars>", len(prompt))
	cmd := exec.CommandContext(ctx, "claude", "-p", prompt)

	var stdout, stderr bytes.Buffer

	// Use MultiWriter to stream to both the buffer and os.Stderr for real-time feedback
	cmd.Stdout = io.MultiWriter(&stdout, os.Stderr)
	cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)

	e.logger.Debug("Running Claude CLI command...")
	if err := cmd.Run(); err != nil {
		e.logger.Error("Claude CLI command failed: %v", err)
		return "", fmt.Errorf("claude command failed: %w, stderr: %s", err, stderr.String())
	}
	e.logger.Debug("Claude CLI command completed successfully")

	return strings.TrimSpace(stdout.String()), nil
}

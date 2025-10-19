package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

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

// AnalyzeSession analyzes a specific session with the default analysis prompt
// This is kept for backward compatibility - uses "tool_analysis" prompt
func (s *AnalysisService) AnalyzeSession(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error) {
	return s.AnalyzeSessionWithPrompt(ctx, sessionID, "tool_analysis")
}

// AnalyzeSessionWithPrompt analyzes a specific session with a named prompt from config
func (s *AnalysisService) AnalyzeSessionWithPrompt(ctx context.Context, sessionID, promptName string) (*domain.SessionAnalysis, error) {
	// Get session logs
	s.logger.Debug("Fetching logs for session %s", sessionID)
	logs, err := s.logsService.ListRecentLogs(ctx, 0, 0, sessionID, true)
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
	promptTemplate, exists := s.config.Prompts[promptName]
	if !exists || promptTemplate == "" {
		s.logger.Warn("Prompt %s not found in config, using default tool_analysis", promptName)
		promptTemplate = domain.DefaultToolAnalysisPrompt
		promptName = "tool_analysis"
	}

	prompt := promptTemplate + buf.String()
	s.logger.Debug("Generated prompt with %d characters (%d KB)", len(prompt), len(prompt)/1024)

	// Execute LLM analysis
	s.logger.Info("Invoking Claude CLI for %s analysis...", promptName)
	analysisResult, err := s.llmExecutor.Execute(ctx, prompt)
	if err != nil {
		s.logger.Error("Failed to execute LLM analysis: %v", err)
		return nil, fmt.Errorf("failed to execute LLM analysis: %w", err)
	}
	s.logger.Debug("Claude CLI returned %d characters", len(analysisResult))

	// Create and save analysis with type
	s.logger.Debug("Saving analysis to database")
	analysis := domain.NewSessionAnalysisWithType(
		sessionID,
		analysisResult,
		s.config.Analysis.Model,
		promptTemplate,
		promptName, // analysis type matches prompt name
		promptName,
	)

	if err := s.analysisRepo.SaveAnalysis(ctx, analysis); err != nil {
		s.logger.Error("Failed to save analysis: %v", err)
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	s.logger.Info("Analysis completed successfully")
	return analysis, nil
}

// AnalyzeMultipleSessions analyzes multiple sessions with a specific prompt
// Returns a map of sessionID -> analysis, and any errors encountered
func (s *AnalysisService) AnalyzeMultipleSessions(ctx context.Context, sessionIDs []string, promptName string) (map[string]*domain.SessionAnalysis, []error) {
	results := make(map[string]*domain.SessionAnalysis)
	var errors []error

	for _, sessionID := range sessionIDs {
		analysis, err := s.AnalyzeSessionWithPrompt(ctx, sessionID, promptName)
		if err != nil {
			errors = append(errors, fmt.Errorf("session %s: %w", sessionID, err))
			continue
		}
		results[sessionID] = analysis
	}

	return results, errors
}

// AnalysisResult represents the result of a parallel analysis
type AnalysisResult struct {
	SessionID  string
	PromptName string
	Analysis   *domain.SessionAnalysis
	Error      error
}

// AnalyzeMultipleSessionsParallel analyzes multiple sessions in parallel
// Uses a semaphore to limit concurrent executions based on config.Analysis.ParallelLimit
func (s *AnalysisService) AnalyzeMultipleSessionsParallel(ctx context.Context, sessionIDs []string, promptName string) (map[string]*domain.SessionAnalysis, []error) {
	if len(sessionIDs) == 0 {
		return make(map[string]*domain.SessionAnalysis), nil
	}

	parallelLimit := s.config.Analysis.ParallelLimit
	if parallelLimit <= 0 {
		parallelLimit = 1
	}

	s.logger.Info("Starting parallel analysis of %d sessions (max parallel: %d)", len(sessionIDs), parallelLimit)

	// Semaphore to limit concurrent executions
	sem := make(chan struct{}, parallelLimit)
	resultsChan := make(chan AnalysisResult, len(sessionIDs))
	var wg sync.WaitGroup

	// Launch goroutines
	for _, sessionID := range sessionIDs {
		wg.Add(1)
		go func(sid string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			s.logger.Debug("Analyzing session %s in parallel", sid)
			analysis, err := s.AnalyzeSessionWithPrompt(ctx, sid, promptName)

			resultsChan <- AnalysisResult{
				SessionID:  sid,
				PromptName: promptName,
				Analysis:   analysis,
				Error:      err,
			}
		}(sessionID)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := make(map[string]*domain.SessionAnalysis)
	var errors []error

	for result := range resultsChan {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("session %s: %w", result.SessionID, result.Error))
			s.logger.Warn("Parallel analysis failed for session %s: %v", result.SessionID, result.Error)
		} else {
			results[result.SessionID] = result.Analysis
			s.logger.Debug("Parallel analysis completed for session %s", result.SessionID)
		}
	}

	s.logger.Info("Parallel analysis complete: %d/%d successful", len(results), len(sessionIDs))
	return results, errors
}

// AnalyzeSessionWithMultiplePrompts analyzes a single session with multiple prompts in parallel
func (s *AnalysisService) AnalyzeSessionWithMultiplePrompts(ctx context.Context, sessionID string, promptNames []string) (map[string]*domain.SessionAnalysis, []error) {
	if len(promptNames) == 0 {
		return make(map[string]*domain.SessionAnalysis), nil
	}

	parallelLimit := s.config.Analysis.ParallelLimit
	if parallelLimit <= 0 {
		parallelLimit = 1
	}

	s.logger.Info("Analyzing session %s with %d prompts in parallel", sessionID, len(promptNames))

	// Semaphore to limit concurrent executions
	sem := make(chan struct{}, parallelLimit)
	resultsChan := make(chan AnalysisResult, len(promptNames))
	var wg sync.WaitGroup

	// Launch goroutines
	for _, promptName := range promptNames {
		wg.Add(1)
		go func(pname string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			s.logger.Debug("Analyzing session %s with prompt %s", sessionID, pname)
			analysis, err := s.AnalyzeSessionWithPrompt(ctx, sessionID, pname)

			resultsChan <- AnalysisResult{
				SessionID:  sessionID,
				PromptName: pname,
				Analysis:   analysis,
				Error:      err,
			}
		}(promptName)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := make(map[string]*domain.SessionAnalysis)
	var errors []error

	for result := range resultsChan {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("prompt %s: %w", result.PromptName, result.Error))
			s.logger.Warn("Analysis failed for prompt %s: %v", result.PromptName, result.Error)
		} else {
			results[result.PromptName] = result.Analysis
			s.logger.Debug("Analysis completed for prompt %s", result.PromptName)
		}
	}

	s.logger.Info("Multi-prompt analysis complete: %d/%d successful", len(results), len(promptNames))
	return results, errors
}

// EstimateTokenCount estimates the token count for a session's logs
// Uses a simple chars/4 heuristic (common approximation for Claude models)
func (s *AnalysisService) EstimateTokenCount(ctx context.Context, sessionID string) (int, error) {
	logs, err := s.logsService.ListRecentLogs(ctx, 0, 0, sessionID, true)
	if err != nil {
		return 0, fmt.Errorf("failed to get session logs: %w", err)
	}

	var buf bytes.Buffer
	if err := FormatLogsAsMarkdown(&buf, logs); err != nil {
		return 0, fmt.Errorf("failed to format logs: %w", err)
	}

	// Estimate tokens: ~4 characters per token (conservative estimate)
	charCount := buf.Len()
	tokenEstimate := charCount / 4

	s.logger.Debug("Session %s: %d chars ≈ %d tokens", sessionID, charCount, tokenEstimate)
	return tokenEstimate, nil
}

// SelectSessionsWithinTokenLimit selects sessions that fit within the token limit
// Returns selected session IDs and total estimated tokens
func (s *AnalysisService) SelectSessionsWithinTokenLimit(ctx context.Context, sessionIDs []string, tokenLimit int) ([]string, int, error) {
	if tokenLimit <= 0 {
		tokenLimit = s.config.Analysis.TokenLimit
	}

	var selected []string
	totalTokens := 0

	// Reserve 20% of tokens for prompt overhead and response
	availableTokens := int(float64(tokenLimit) * 0.8)

	s.logger.Debug("Selecting sessions within %d tokens (%.0f%% of %d limit)", availableTokens, 80.0, tokenLimit)

	for _, sessionID := range sessionIDs {
		tokenCount, err := s.EstimateTokenCount(ctx, sessionID)
		if err != nil {
			s.logger.Warn("Failed to estimate tokens for session %s: %v", sessionID, err)
			continue
		}

		if totalTokens+tokenCount <= availableTokens {
			selected = append(selected, sessionID)
			totalTokens += tokenCount
			s.logger.Debug("Selected session %s (%d tokens, total: %d/%d)", sessionID, tokenCount, totalTokens, availableTokens)
		} else {
			s.logger.Debug("Skipping session %s (%d tokens would exceed limit: %d + %d > %d)",
				sessionID, tokenCount, totalTokens, tokenCount, availableTokens)
			break
		}
	}

	s.logger.Info("Selected %d/%d sessions (%d tokens, %.1f%% of limit)",
		len(selected), len(sessionIDs), totalTokens, float64(totalTokens)/float64(tokenLimit)*100)

	return selected, totalTokens, nil
}

// GetLastSession returns the ID of the most recent session
func (s *AnalysisService) GetLastSession(ctx context.Context) (string, error) {
	logs, err := s.logsService.ListRecentLogs(ctx, 1, 0, "", false)
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

// GetAnalysis retrieves the most recent analysis for a session
func (s *AnalysisService) GetAnalysis(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error) {
	return s.analysisRepo.GetAnalysisBySessionID(ctx, sessionID)
}

// GetAnalysesBySessionID retrieves all analyses for a session
func (s *AnalysisService) GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionAnalysis, error) {
	return s.analysisRepo.GetAnalysesBySessionID(ctx, sessionID)
}

// GetAllSessionIDs retrieves all session IDs, ordered by most recent first
// If limit > 0, returns only the latest N sessions
func (s *AnalysisService) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	return s.analysisRepo.GetAllSessionIDs(ctx, limit)
}

// FilenameTmplData contains template data for filename generation
type FilenameTmplData struct {
	SessionID  string
	PromptName string
	Date       string
	Time       string
}

// SaveToMarkdown saves an analysis to a markdown file
// outputDir: directory to save the file (empty uses config default)
// filename: filename override (empty uses config template)
func (s *AnalysisService) SaveToMarkdown(ctx context.Context, analysis *domain.SessionAnalysis, outputDir, filename string) (string, error) {
	if analysis == nil {
		return "", fmt.Errorf("analysis is nil")
	}

	// Use config default if outputDir not specified
	if outputDir == "" {
		outputDir = s.config.UI.DefaultOutputDir
		if outputDir == "" {
			outputDir = "./analysis-outputs"
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename if not specified
	if filename == "" {
		tmplStr := s.config.UI.FilenameTemplate
		if tmplStr == "" {
			tmplStr = "{{.SessionID}}-{{.PromptName}}-{{.Date}}.md"
		}

		// Parse and execute template
		tmpl, err := template.New("filename").Parse(tmplStr)
		if err != nil {
			return "", fmt.Errorf("invalid filename template: %w", err)
		}

		now := time.Now()
		data := FilenameTmplData{
			SessionID:  analysis.SessionID[:8], // Use first 8 chars of session ID
			PromptName: analysis.PromptName,
			Date:       now.Format("2006-01-02"),
			Time:       now.Format("15-04-05"),
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to generate filename: %w", err)
		}

		filename = buf.String()
	}

	// Ensure .md extension
	if filepath.Ext(filename) != ".md" {
		filename = filename + ".md"
	}

	// Build full path
	fullPath := filepath.Join(outputDir, filename)

	// Create markdown content
	content := fmt.Sprintf(`# Session Analysis: %s

**Session ID**: %s
**Analysis Type**: %s
**Prompt**: %s
**Model**: %s
**Analyzed At**: %s

---

%s
`,
		analysis.SessionID[:8],
		analysis.SessionID,
		analysis.AnalysisType,
		analysis.PromptName,
		analysis.ModelUsed,
		analysis.AnalyzedAt.Format(time.RFC3339),
		analysis.AnalysisResult,
	)

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	s.logger.Info("Saved analysis to %s", fullPath)
	return fullPath, nil
}

// ClaudeCLIExecutor implements LLMExecutor using the claude CLI tool
type ClaudeCLIExecutor struct {
	logger Logger
	config *domain.Config
}

// NewClaudeCLIExecutor creates a new Claude CLI executor
func NewClaudeCLIExecutor(logger Logger) *ClaudeCLIExecutor {
	if logger == nil {
		logger = &NoOpLogger{}
	}
	return &ClaudeCLIExecutor{
		logger: logger,
		config: domain.DefaultConfig(),
	}
}

// NewClaudeCLIExecutorWithConfig creates a new Claude CLI executor with custom config
func NewClaudeCLIExecutorWithConfig(logger Logger, config *domain.Config) *ClaudeCLIExecutor {
	if logger == nil {
		logger = &NoOpLogger{}
	}
	if config == nil {
		config = domain.DefaultConfig()
	}
	return &ClaudeCLIExecutor{
		logger: logger,
		config: config,
	}
}

// Execute runs claude -p with the given prompt
// Streams output to stderr in real-time for progress visibility
// The prompt parameter is treated as the user prompt unless config specifies system prompt mode
func (e *ClaudeCLIExecutor) Execute(ctx context.Context, prompt string) (string, error) {
	return e.ExecuteWithOptions(ctx, prompt, nil)
}

// ExecuteWithOptions runs claude with custom options
// options can override config settings (model, tokenLimit, etc.)
func (e *ClaudeCLIExecutor) ExecuteWithOptions(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	// Build command arguments (no -p flag, we'll use direct prompt)
	args := []string{}

	// Apply model from config or options
	model := e.config.Analysis.Model
	if opt, ok := options["model"].(string); ok && opt != "" {
		model = opt
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	var userPrompt string

	// Apply system prompt mode from config
	if e.config.Analysis.ClaudeOptions.SystemPromptMode == "replace" {
		args = append(args, "--system-prompt", prompt)
		// When using system prompt, we need a user prompt too
		userPrompt = "Analyze the session data provided in the system prompt."
	} else if e.config.Analysis.ClaudeOptions.SystemPromptMode == "append" {
		args = append(args, "--append-system-prompt", prompt)
		userPrompt = "Analyze the session data."
	} else {
		// No system prompt mode, use prompt directly
		userPrompt = prompt
	}

	// Apply allowed tools from config
	if len(e.config.Analysis.ClaudeOptions.AllowedTools) > 0 {
		args = append(args, "--allowed-tools", strings.Join(e.config.Analysis.ClaudeOptions.AllowedTools, ","))
	}

	// Add the user prompt last
	args = append(args, userPrompt)

	e.logger.Debug("Executing: claude %s", strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "claude", args...)

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

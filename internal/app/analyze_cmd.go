package app

import (
	"context"
	"fmt"
	"io"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// AnalyzeOptions contains options for the analyze command
type AnalyzeOptions struct {
	SessionID     string
	Last          bool
	ViewOnly      bool
	AnalyzeAll    bool
	Refresh       bool
	Limit         int
	PromptNames   []string
	ModelOverride string
	TokenLimit    int
}

// AnalysisServiceInterface defines the interface for analysis operations
type AnalysisServiceInterface interface {
	GetLastSession(ctx context.Context) (string, error)
	GetAnalysis(ctx context.Context, sessionID string) (*domain.SessionAnalysis, error)
	AnalyzeSessionWithPrompt(ctx context.Context, sessionID string, promptName string) (*domain.SessionAnalysis, error)
	GetUnanalyzedSessions(ctx context.Context) ([]string, error)
	GetAllSessionIDs(ctx context.Context, limit int) ([]string, error)
	AnalyzeSessionWithMultiplePrompts(ctx context.Context, sessionID string, promptNames []string) (map[string]*domain.SessionAnalysis, []error)
}

// AnalyzeCommandHandler handles the analyze command logic
type AnalyzeCommandHandler struct {
	analysisService AnalysisServiceInterface
	logger          Logger
	out             io.Writer
}

// NewAnalyzeCommandHandler creates a new analyze command handler
func NewAnalyzeCommandHandler(analysisService AnalysisServiceInterface, logger Logger, out io.Writer) *AnalyzeCommandHandler {
	return &AnalyzeCommandHandler{
		analysisService: analysisService,
		logger:          logger,
		out:             out,
	}
}

// Execute runs the analyze command based on options
func (h *AnalyzeCommandHandler) Execute(ctx context.Context, opts AnalyzeOptions) error {
	// Handle different modes
	if opts.Refresh {
		return h.refreshAnalyses(ctx, opts.Limit, opts.PromptNames)
	}

	if opts.AnalyzeAll {
		return h.analyzeAllSessions(ctx, opts.PromptNames)
	}

	// Determine which session to analyze
	var targetSessionID string
	if opts.SessionID != "" {
		targetSessionID = opts.SessionID
	} else if opts.Last {
		lastSessionID, err := h.analysisService.GetLastSession(ctx)
		if err != nil {
			return fmt.Errorf("failed to get last session: %w", err)
		}
		targetSessionID = lastSessionID
		fmt.Fprintf(h.out, "Analyzing last session: %s\n\n", targetSessionID)
	} else {
		return fmt.Errorf("must specify --session-id or --last")
	}

	// View existing analysis if requested
	if opts.ViewOnly {
		return h.viewAnalysis(ctx, targetSessionID)
	}

	// Perform analysis
	return h.analyzeSession(ctx, targetSessionID, opts.PromptNames)
}

// viewAnalysis displays an existing analysis
func (h *AnalyzeCommandHandler) viewAnalysis(ctx context.Context, sessionID string) error {
	analysis, err := h.analysisService.GetAnalysis(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get analysis: %w", err)
	}

	if analysis == nil {
		fmt.Fprintf(h.out, "No analysis found for session %s\n", sessionID)
		fmt.Fprintln(h.out, "Run without --view to analyze this session")
		return fmt.Errorf("no analysis found")
	}

	fmt.Fprintf(h.out, "Session: %s\n", analysis.SessionID)
	fmt.Fprintf(h.out, "Analyzed at: %s\n", analysis.AnalyzedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(h.out, "Model: %s\n\n", analysis.ModelUsed)
	fmt.Fprintln(h.out, "=== Analysis Result ===")
	fmt.Fprintln(h.out, analysis.AnalysisResult)

	return nil
}

// analyzeSession analyzes a single session with one or more prompts
func (h *AnalyzeCommandHandler) analyzeSession(ctx context.Context, sessionID string, promptNames []string) error {
	if len(promptNames) == 1 {
		// Single prompt - use simple sequential analysis
		fmt.Fprintf(h.out, "Analyzing session %s with prompt '%s'...\n", sessionID, promptNames[0])
		analysis, err := h.analysisService.AnalyzeSessionWithPrompt(ctx, sessionID, promptNames[0])
		if err != nil {
			return fmt.Errorf("failed to analyze session: %w", err)
		}

		fmt.Fprintf(h.out, "\nAnalysis completed at %s\n\n", analysis.AnalyzedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintln(h.out, "=== Analysis Result ===")
		fmt.Fprintln(h.out, analysis.AnalysisResult)
	} else {
		// Multiple prompts - use parallel analysis
		fmt.Fprintf(h.out, "Analyzing session %s with %d prompts in parallel: %v\n", sessionID, len(promptNames), promptNames)
		analyses, errs := h.analysisService.AnalyzeSessionWithMultiplePrompts(ctx, sessionID, promptNames)

		if len(errs) > 0 {
			fmt.Fprintln(h.out, "\nErrors during analysis:")
			for _, err := range errs {
				fmt.Fprintf(h.out, "  - %v\n", err)
			}
		}

		if len(analyses) > 0 {
			fmt.Fprintf(h.out, "\nCompleted %d/%d analyses successfully\n\n", len(analyses), len(promptNames))
			for promptName, analysis := range analyses {
				fmt.Fprintf(h.out, "=== Analysis: %s (completed at %s) ===\n", promptName, analysis.AnalyzedAt.Format("15:04:05"))
				fmt.Fprintln(h.out, analysis.AnalysisResult)
				fmt.Fprintln(h.out)
			}
		} else {
			return fmt.Errorf("all analyses failed")
		}
	}

	return nil
}

// analyzeAllSessions analyzes all unanalyzed sessions
func (h *AnalyzeCommandHandler) analyzeAllSessions(ctx context.Context, promptNames []string) error {
	// Get unanalyzed sessions
	h.logger.Debug("Fetching unanalyzed sessions")
	sessionIDs, err := h.analysisService.GetUnanalyzedSessions(ctx)
	if err != nil {
		h.logger.Error("Failed to get unanalyzed sessions: %v", err)
		return fmt.Errorf("failed to get unanalyzed sessions: %w", err)
	}
	h.logger.Debug("Found %d unanalyzed sessions", len(sessionIDs))

	if len(sessionIDs) == 0 {
		h.logger.Info("No unanalyzed sessions found")
		fmt.Fprintln(h.out, "No unanalyzed sessions found")
		return nil
	}

	fmt.Fprintf(h.out, "Found %d unanalyzed session(s)\n", len(sessionIDs))
	fmt.Fprintf(h.out, "Using prompts: %v\n\n", promptNames)

	// Analyze each session with all prompts
	successCount := 0
	for i, sessionID := range sessionIDs {
		fmt.Fprintf(h.out, "[%d/%d] Analyzing session %s with %d prompt(s)...\n", i+1, len(sessionIDs), sessionID, len(promptNames))
		h.logger.Debug("Starting analysis for session %s (%d/%d)", sessionID, i+1, len(sessionIDs))

		if len(promptNames) == 1 {
			// Single prompt - simple sequential
			analysis, err := h.analysisService.AnalyzeSessionWithPrompt(ctx, sessionID, promptNames[0])
			if err != nil {
				fmt.Fprintf(h.out, "Failed to analyze session %s: %v\n", sessionID, err)
				h.logger.Warn("Analysis failed for session %s: %v", sessionID, err)
				continue
			}
			successCount++
			h.logger.Info("Analysis completed for session %s", sessionID)
			fmt.Fprintf(h.out, "✓ Completed at %s\n\n", analysis.AnalyzedAt.Format("15:04:05"))
		} else {
			// Multiple prompts - parallel
			analyses, errs := h.analysisService.AnalyzeSessionWithMultiplePrompts(ctx, sessionID, promptNames)
			if len(errs) > 0 {
				h.logger.Warn("Some analyses failed for session %s: %v", sessionID, errs)
			}
			if len(analyses) > 0 {
				successCount++
				h.logger.Info("Completed %d/%d analyses for session %s", len(analyses), len(promptNames), sessionID)
				fmt.Fprintf(h.out, "✓ Completed %d/%d analyses\n\n", len(analyses), len(promptNames))
			} else {
				fmt.Fprintf(h.out, "All analyses failed for session %s\n", sessionID)
			}
		}
	}

	fmt.Fprintf(h.out, "\nAnalyzed %d/%d session(s) successfully\n", successCount, len(sessionIDs))
	h.logger.Info("Batch analysis complete: %d/%d successful", successCount, len(sessionIDs))

	return nil
}

// refreshAnalyses re-analyzes existing sessions
func (h *AnalyzeCommandHandler) refreshAnalyses(ctx context.Context, limit int, promptNames []string) error {
	// Get all sessions (or latest N if limit is specified)
	h.logger.Debug("Fetching session IDs for refresh (limit: %d)", limit)
	sessionIDs, err := h.analysisService.GetAllSessionIDs(ctx, limit)
	if err != nil {
		h.logger.Error("Failed to get session IDs: %v", err)
		return fmt.Errorf("failed to get session IDs: %w", err)
	}
	h.logger.Debug("Found %d sessions to refresh", len(sessionIDs))

	if len(sessionIDs) == 0 {
		h.logger.Info("No sessions found to refresh")
		fmt.Fprintln(h.out, "No sessions found to refresh")
		return nil
	}

	if limit > 0 {
		fmt.Fprintf(h.out, "Refreshing analyses for latest %d session(s)\n", len(sessionIDs))
	} else {
		fmt.Fprintf(h.out, "Refreshing analyses for all %d session(s)\n", len(sessionIDs))
	}
	fmt.Fprintf(h.out, "Using prompts: %v\n\n", promptNames)

	// Re-analyze each session with all prompts
	successCount := 0
	for i, sessionID := range sessionIDs {
		fmt.Fprintf(h.out, "[%d/%d] Re-analyzing session %s with %d prompt(s)...\n", i+1, len(sessionIDs), sessionID, len(promptNames))
		h.logger.Debug("Starting re-analysis for session %s (%d/%d)", sessionID, i+1, len(sessionIDs))

		if len(promptNames) == 1 {
			// Single prompt - simple sequential
			analysis, err := h.analysisService.AnalyzeSessionWithPrompt(ctx, sessionID, promptNames[0])
			if err != nil {
				fmt.Fprintf(h.out, "Failed to re-analyze session %s: %v\n", sessionID, err)
				h.logger.Warn("Re-analysis failed for session %s: %v", sessionID, err)
				continue
			}
			successCount++
			h.logger.Info("Re-analysis completed for session %s", sessionID)
			fmt.Fprintf(h.out, "✓ Completed at %s\n\n", analysis.AnalyzedAt.Format("15:04:05"))
		} else {
			// Multiple prompts - parallel
			analyses, errs := h.analysisService.AnalyzeSessionWithMultiplePrompts(ctx, sessionID, promptNames)
			if len(errs) > 0 {
				h.logger.Warn("Some re-analyses failed for session %s: %v", sessionID, errs)
			}
			if len(analyses) > 0 {
				successCount++
				h.logger.Info("Completed %d/%d re-analyses for session %s", len(analyses), len(promptNames), sessionID)
				fmt.Fprintf(h.out, "✓ Completed %d/%d analyses\n\n", len(analyses), len(promptNames))
			} else {
				fmt.Fprintf(h.out, "All re-analyses failed for session %s\n", sessionID)
			}
		}
	}

	fmt.Fprintf(h.out, "\nRefreshed %d/%d session(s) successfully\n", successCount, len(sessionIDs))
	h.logger.Info("Refresh complete: %d/%d successful", successCount, len(sessionIDs))

	return nil
}

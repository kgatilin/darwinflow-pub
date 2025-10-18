package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func analyzeCmd(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	sessionID := fs.String("session-id", "", "Session ID to analyze")
	last := fs.Bool("last", false, "Analyze the last session")
	viewOnly := fs.Bool("view", false, "View existing analysis without re-analyzing")
	analyzeAll := fs.Bool("all", false, "Analyze all unanalyzed sessions")
	refresh := fs.Bool("refresh", false, "Re-analyze sessions even if already analyzed")
	limit := fs.Int("limit", 0, "Limit number of sessions to refresh (0 = all sessions)")
	debug := fs.Bool("debug", false, "Enable debug logging")
	debugShort := fs.Bool("d", false, "Enable debug logging (short flag)")

	if err := fs.Parse(args); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Enable debug if either flag is set
	debugEnabled := *debug || *debugShort

	// Create logger with appropriate level
	var logger *infra.Logger
	if debugEnabled {
		logger = infra.NewDebugLogger()
		logger.Info("Debug logging enabled")
	} else {
		logger = infra.NewDefaultLogger()
	}

	ctx := context.Background()

	// Initialize repository
	logger.Debug("Initializing repository at %s", app.DefaultDBPath)
	repo, err := infra.NewSQLiteEventRepository(app.DefaultDBPath)
	if err != nil {
		logger.Error("Failed to initialize repository: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to initialize repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Load config
	logger.Debug("Loading configuration")
	configLoader := infra.NewConfigLoader(logger)
	config, err := configLoader.LoadConfig("")
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create services
	logger.Debug("Creating analysis services")
	logsService := app.NewLogsService(repo, repo)
	llmExecutor := app.NewClaudeCLIExecutor(logger)
	analysisService := app.NewAnalysisService(repo, repo, logsService, llmExecutor, logger, config)

	// Handle different modes
	if *refresh {
		refreshAnalyses(ctx, analysisService, logger, *limit)
		return
	}

	if *analyzeAll {
		analyzeAllSessions(ctx, analysisService, logger)
		return
	}

	// Determine which session to analyze
	var targetSessionID string
	if *sessionID != "" {
		targetSessionID = *sessionID
	} else if *last {
		lastSessionID, err := analysisService.GetLastSession(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get last session: %v\n", err)
			os.Exit(1)
		}
		targetSessionID = lastSessionID
		fmt.Printf("Analyzing last session: %s\n\n", targetSessionID)
	} else {
		fmt.Fprintf(os.Stderr, "Error: must specify --session-id or --last\n")
		fs.Usage()
		os.Exit(1)
	}

	// View existing analysis if requested
	if *viewOnly {
		viewAnalysis(ctx, analysisService, targetSessionID)
		return
	}

	// Perform analysis
	fmt.Printf("Analyzing session %s...\n", targetSessionID)
	analysis, err := analysisService.AnalyzeSession(ctx, targetSessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to analyze session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAnalysis completed at %s\n\n", analysis.AnalyzedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("=== Analysis Result ===")
	fmt.Println(analysis.AnalysisResult)
}

func viewAnalysis(ctx context.Context, service *app.AnalysisService, sessionID string) {
	analysis, err := service.GetAnalysis(ctx, sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get analysis: %v\n", err)
		os.Exit(1)
	}

	if analysis == nil {
		fmt.Printf("No analysis found for session %s\n", sessionID)
		fmt.Println("Run without --view to analyze this session")
		os.Exit(1)
	}

	fmt.Printf("Session: %s\n", analysis.SessionID)
	fmt.Printf("Analyzed at: %s\n", analysis.AnalyzedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Model: %s\n\n", analysis.ModelUsed)
	fmt.Println("=== Analysis Result ===")
	fmt.Println(analysis.AnalysisResult)
}

func analyzeAllSessions(ctx context.Context, service *app.AnalysisService, logger *infra.Logger) {
	// Get unanalyzed sessions
	logger.Debug("Fetching unanalyzed sessions")
	sessionIDs, err := service.GetUnanalyzedSessions(ctx)
	if err != nil {
		logger.Error("Failed to get unanalyzed sessions: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to get unanalyzed sessions: %v\n", err)
		os.Exit(1)
	}
	logger.Debug("Found %d unanalyzed sessions", len(sessionIDs))

	if len(sessionIDs) == 0 {
		logger.Info("No unanalyzed sessions found")
		fmt.Println("No unanalyzed sessions found")
		return
	}

	fmt.Printf("Found %d unanalyzed session(s)\n\n", len(sessionIDs))

	// Analyze each session
	successCount := 0
	for i, sessionID := range sessionIDs {
		fmt.Printf("[%d/%d] Analyzing session %s...\n", i+1, len(sessionIDs), sessionID)
		logger.Debug("Starting analysis for session %s (%d/%d)", sessionID, i+1, len(sessionIDs))

		analysis, err := service.AnalyzeSession(ctx, sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to analyze session %s: %v\n", sessionID, err)
			logger.Warn("Analysis failed for session %s: %v", sessionID, err)
			continue
		}

		successCount++
		logger.Info("Analysis completed for session %s", sessionID)
		fmt.Printf("✓ Completed at %s\n\n", analysis.AnalyzedAt.Format("15:04:05"))
	}

	fmt.Printf("\nAnalyzed %d/%d session(s) successfully\n", successCount, len(sessionIDs))
	logger.Info("Batch analysis complete: %d/%d successful", successCount, len(sessionIDs))
}

func refreshAnalyses(ctx context.Context, service *app.AnalysisService, logger *infra.Logger, limit int) {
	// Get all sessions (or latest N if limit is specified)
	logger.Debug("Fetching session IDs for refresh (limit: %d)", limit)
	sessionIDs, err := service.GetAllSessionIDs(ctx, limit)
	if err != nil {
		logger.Error("Failed to get session IDs: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to get session IDs: %v\n", err)
		os.Exit(1)
	}
	logger.Debug("Found %d sessions to refresh", len(sessionIDs))

	if len(sessionIDs) == 0 {
		logger.Info("No sessions found to refresh")
		fmt.Println("No sessions found to refresh")
		return
	}

	if limit > 0 {
		fmt.Printf("Refreshing analyses for latest %d session(s)...\n\n", len(sessionIDs))
	} else {
		fmt.Printf("Refreshing analyses for all %d session(s)...\n\n", len(sessionIDs))
	}

	// Re-analyze each session
	successCount := 0
	for i, sessionID := range sessionIDs {
		fmt.Printf("[%d/%d] Re-analyzing session %s...\n", i+1, len(sessionIDs), sessionID)
		logger.Debug("Starting re-analysis for session %s (%d/%d)", sessionID, i+1, len(sessionIDs))

		analysis, err := service.AnalyzeSession(ctx, sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to re-analyze session %s: %v\n", sessionID, err)
			logger.Warn("Re-analysis failed for session %s: %v", sessionID, err)
			continue
		}

		successCount++
		logger.Info("Re-analysis completed for session %s", sessionID)
		fmt.Printf("✓ Completed at %s\n\n", analysis.AnalyzedAt.Format("15:04:05"))
	}

	fmt.Printf("\nRefreshed %d/%d session(s) successfully\n", successCount, len(sessionIDs))
	logger.Info("Refresh complete: %d/%d successful", successCount, len(sessionIDs))
}

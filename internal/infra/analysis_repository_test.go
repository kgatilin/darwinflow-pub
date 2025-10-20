package infra_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestSQLiteEventRepository_SaveAnalysis(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create and save an analysis
	analysis := domain.NewSessionAnalysis(
		"test-session-123",
		"This is the analysis result with patterns and suggestions",
		"claude-sonnet-4",
		"Analysis prompt template",
	)
	analysis.PatternsSummary = "Found 3 patterns: read-edit-save, grep-read, tool chains"

	err = repo.SaveAnalysis(ctx, analysis)
	if err != nil {
		t.Errorf("Failed to save analysis: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetAnalysisBySessionID(ctx, "test-session-123")
	if err != nil {
		t.Errorf("Failed to retrieve analysis: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected analysis to be found, got nil")
	}

	if retrieved.SessionID != analysis.SessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", retrieved.SessionID, analysis.SessionID)
	}

	if retrieved.AnalysisResult != analysis.AnalysisResult {
		t.Errorf("AnalysisResult mismatch")
	}

	if retrieved.ModelUsed != analysis.ModelUsed {
		t.Errorf("ModelUsed mismatch: got %s, want %s", retrieved.ModelUsed, analysis.ModelUsed)
	}
}

func TestSQLiteEventRepository_GetUnanalyzedSessionIDs(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create some events for different sessions
	session1 := "session-1"
	session2 := "session-2"
	session3 := "session-3"

	events := []*domain.Event{
		domain.NewEvent(domain.ChatStarted, session1, domain.ChatPayload{Message: "test"}, "test"),
		domain.NewEvent(domain.ChatStarted, session2, domain.ChatPayload{Message: "test"}, "test"),
		domain.NewEvent(domain.ChatStarted, session3, domain.ChatPayload{Message: "test"}, "test"),
	}

	for _, event := range events {
		if err := repo.Save(ctx, event); err != nil {
			t.Fatalf("Failed to save event: %v", err)
		}
	}

	// Initially, all sessions should be unanalyzed
	unanalyzed, err := repo.GetUnanalyzedSessionIDs(ctx)
	if err != nil {
		t.Errorf("Failed to get unanalyzed sessions: %v", err)
	}

	if len(unanalyzed) != 3 {
		t.Errorf("Expected 3 unanalyzed sessions, got %d", len(unanalyzed))
	}

	// Analyze session1
	analysis := domain.NewSessionAnalysis(session1, "analysis result", "claude", "prompt")
	if err := repo.SaveAnalysis(ctx, analysis); err != nil {
		t.Fatalf("Failed to save analysis: %v", err)
	}

	// Now should have 2 unanalyzed sessions
	unanalyzed, err = repo.GetUnanalyzedSessionIDs(ctx)
	if err != nil {
		t.Errorf("Failed to get unanalyzed sessions: %v", err)
	}

	if len(unanalyzed) != 2 {
		t.Errorf("Expected 2 unanalyzed sessions, got %d", len(unanalyzed))
	}

	// Verify session1 is not in the list
	for _, id := range unanalyzed {
		if id == session1 {
			t.Errorf("session1 should not be in unanalyzed list")
		}
	}
}

func TestSQLiteEventRepository_GetAllAnalyses(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Save multiple analyses
	analyses := []*domain.SessionAnalysis{
		domain.NewSessionAnalysis("session-1", "result 1", "claude", "prompt"),
		domain.NewSessionAnalysis("session-2", "result 2", "claude", "prompt"),
		domain.NewSessionAnalysis("session-3", "result 3", "claude", "prompt"),
	}

	for _, analysis := range analyses {
		time.Sleep(time.Millisecond) // Ensure different timestamps
		if err := repo.SaveAnalysis(ctx, analysis); err != nil {
			t.Fatalf("Failed to save analysis: %v", err)
		}
	}

	// Retrieve all
	all, err := repo.GetAllAnalyses(ctx, 0)
	if err != nil {
		t.Errorf("Failed to get all analyses: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 analyses, got %d", len(all))
	}

	// Verify they're ordered by analyzed_at DESC (newest first)
	if len(all) > 1 {
		if all[0].AnalyzedAt.Before(all[1].AnalyzedAt) {
			t.Errorf("Analyses not ordered by analyzed_at DESC")
		}
	}

	// Test limit
	limited, err := repo.GetAllAnalyses(ctx, 2)
	if err != nil {
		t.Errorf("Failed to get limited analyses: %v", err)
	}

	if len(limited) != 2 {
		t.Errorf("Expected 2 analyses with limit, got %d", len(limited))
	}
}

func TestSQLiteEventRepository_GetAnalysisBySessionID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Try to get non-existent analysis
	analysis, err := repo.GetAnalysisBySessionID(ctx, "non-existent-session")
	if err != nil {
		t.Errorf("Expected no error for non-existent session, got: %v", err)
	}

	if analysis != nil {
		t.Errorf("Expected nil analysis for non-existent session, got: %v", analysis)
	}
}

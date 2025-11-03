package task_manager_test

import (
	"context"
	"testing"
	"time"

	tm "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

func TestIterationReordering(t *testing.T) {
	// Setup
	db := createTestDB(t)
	defer db.Close()

	repo := tm.NewSQLiteRoadmapRepository(db, createTestLogger())
	ctx := context.Background()

	// Create roadmap
	now := time.Now().UTC()
	roadmap, err := tm.NewRoadmapEntity("roadmap-1", "Test roadmap", "Test criteria", now, now)
	if err != nil {
		t.Fatalf("Failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("Failed to save roadmap: %v", err)
	}

	// Create 3 iterations with same default rank (500)
	iter1, _ := tm.NewIterationEntity(1, "Iteration 1", "Goal 1", "", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)
	iter2, _ := tm.NewIterationEntity(2, "Iteration 2", "Goal 2", "", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)
	iter3, _ := tm.NewIterationEntity(3, "Iteration 3", "Goal 3", "", []string{}, "planned", 500, time.Time{}, time.Time{}, now, now)

	if err := repo.SaveIteration(ctx, iter1); err != nil {
		t.Fatalf("Failed to save iteration 1: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter2); err != nil {
		t.Fatalf("Failed to save iteration 2: %v", err)
	}
	if err := repo.SaveIteration(ctx, iter3); err != nil {
		t.Fatalf("Failed to save iteration 3: %v", err)
	}

	// Verify initial order (by number since ranks are equal)
	iterations, err := repo.ListIterations(ctx)
	if err != nil {
		t.Fatalf("Failed to list iterations: %v", err)
	}
	if len(iterations) != 3 {
		t.Fatalf("Expected 3 iterations, got %d", len(iterations))
	}
	if iterations[0].Number != 1 || iterations[1].Number != 2 || iterations[2].Number != 3 {
		t.Errorf("Initial order wrong: got %d, %d, %d", iterations[0].Number, iterations[1].Number, iterations[2].Number)
	}

	// Log initial ranks
	t.Logf("Initial ranks: iter1=%d, iter2=%d, iter3=%d", iterations[0].Rank, iterations[1].Rank, iterations[2].Rank)

	// Simulate TUI reordering: move iteration 1 down (swap with iteration 2)
	// This is what the TUI code does
	firstIter := iterations[0]
	secondIter := iterations[1]

	t.Logf("Before swap: firstIter.Rank=%d, secondIter.Rank=%d", firstIter.Rank, secondIter.Rank)

	// Swap ranks - this is what the TUI does now
	if firstIter.Rank == secondIter.Rank {
		// Moving down, so increase rank
		firstIter.Rank = firstIter.Rank + 1
	} else {
		firstIter.Rank, secondIter.Rank = secondIter.Rank, firstIter.Rank
	}
	firstIter.UpdatedAt = time.Now().UTC()
	secondIter.UpdatedAt = time.Now().UTC()

	t.Logf("After swap in memory: firstIter.Rank=%d, secondIter.Rank=%d", firstIter.Rank, secondIter.Rank)

	// Update both iterations
	if err := repo.UpdateIteration(ctx, firstIter); err != nil {
		t.Fatalf("Failed to update first iteration: %v", err)
	}
	if err := repo.UpdateIteration(ctx, secondIter); err != nil {
		t.Fatalf("Failed to update second iteration: %v", err)
	}

	// Reload iterations to verify order changed
	iterations, err = repo.ListIterations(ctx)
	if err != nil {
		t.Fatalf("Failed to list iterations after reorder: %v", err)
	}

	t.Logf("After reload from DB: iter[0]=%d (rank=%d), iter[1]=%d (rank=%d), iter[2]=%d (rank=%d)",
		iterations[0].Number, iterations[0].Rank,
		iterations[1].Number, iterations[1].Rank,
		iterations[2].Number, iterations[2].Rank)

	// After moving iter1 down:
	// - iter1 should now have rank=501
	// - iter2 should still have rank=500
	// - iter3 should still have rank=500
	// Order by rank, then number: iter2 (500, #2), iter3 (500, #3), iter1 (501, #1)

	// Expected: iteration 2 should come before iteration 1 now
	if iterations[0].Number != 2 {
		t.Errorf("After reordering down, first iteration should be 2, got %d (rank=%d)", iterations[0].Number, iterations[0].Rank)
	}
	if iterations[1].Number != 3 {
		t.Errorf("After reordering down, second iteration should be 3, got %d (rank=%d)", iterations[1].Number, iterations[1].Rank)
	}
	if iterations[2].Number != 1 {
		t.Errorf("After reordering down, third iteration should be 1, got %d (rank=%d)", iterations[2].Number, iterations[2].Rank)
	}

	// Verify ranks are different now
	if iterations[2].Rank != 501 {
		t.Errorf("Iteration 1 should have rank 501, got %d", iterations[2].Rank)
	}
}

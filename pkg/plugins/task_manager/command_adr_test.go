package task_manager_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// ============================================================================
// ADRCreateCommand Tests
// ============================================================================

func TestADRCreateCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup: Create roadmap and track
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, err := task_manager.NewRoadmapEntity(
		"roadmap-test",
		"Test vision",
		"Test criteria",
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create roadmap: %v", err)
	}
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		t.Fatalf("failed to save roadmap: %v", err)
	}

	track, err := task_manager.NewTrackEntity(
		"track-core",
		"roadmap-test",
		"Core Framework",
		"Test description",
		"not-started",
		200,
		[]string{},
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("failed to create track: %v", err)
	}
	if err := repo.SaveTrack(ctx, track); err != nil {
		t.Fatalf("failed to save track: %v", err)
	}

	// Execute command
	cmd := &task_manager.ADRCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-core",
		"--title", "Use gRPC for RPC",
		"--context", "Need efficient RPC mechanism",
		"--decision", "Adopt gRPC with Protocol Buffers",
		"--consequences", "Requires code generation",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify ADR was created by querying the repository
	adrs, err := repo.ListADRs(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list ADRs: %v", err)
	}
	if len(adrs) != 1 {
		t.Errorf("expected 1 ADR, got %d", len(adrs))
	}
	if len(adrs) > 0 {
		if adrs[0].Title != "Use gRPC for RPC" {
			t.Errorf("expected title 'Use gRPC for RPC', got '%s'", adrs[0].Title)
		}
		if adrs[0].Status != string(task_manager.ADRStatusProposed) {
			t.Errorf("expected status 'proposed', got '%s'", adrs[0].Status)
		}
	}
}

func TestADRCreateCommand_WithAlternatives(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-core", "roadmap-test", "Core Framework", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Execute command with alternatives
	cmd := &task_manager.ADRCreateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-core",
		"--title", "Use gRPC for RPC",
		"--context", "Need efficient RPC mechanism",
		"--decision", "Adopt gRPC with Protocol Buffers",
		"--consequences", "Requires code generation",
		"--alternatives", "REST, GraphQL, Thrift",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestADRCreateCommand_MissingTrackID(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRCreateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"--title", "Some title",
		"--context", "Some context",
		"--decision", "Some decision",
		"--consequences", "Some consequences",
	})
	if err == nil {
		t.Errorf("expected error for missing track ID")
	}
	if !strings.Contains(err.Error(), "track ID is required") {
		t.Errorf("expected error about track ID, got: %v", err)
	}
}

func TestADRCreateCommand_MissingTitle(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRCreateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-core",
		"--context", "Some context",
		"--decision", "Some decision",
		"--consequences", "Some consequences",
	})
	if err == nil {
		t.Errorf("expected error for missing title")
	}
	if !strings.Contains(err.Error(), "--title") {
		t.Errorf("expected error about --title, got: %v", err)
	}
}

func TestADRCreateCommand_MissingContext(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRCreateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-core",
		"--title", "Some title",
		"--decision", "Some decision",
		"--consequences", "Some consequences",
	})
	if err == nil {
		t.Errorf("expected error for missing context")
	}
	if !strings.Contains(err.Error(), "--context") {
		t.Errorf("expected error about --context, got: %v", err)
	}
}

func TestADRCreateCommand_MissingDecision(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRCreateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-core",
		"--title", "Some title",
		"--context", "Some context",
		"--consequences", "Some consequences",
	})
	if err == nil {
		t.Errorf("expected error for missing decision")
	}
	if !strings.Contains(err.Error(), "--decision") {
		t.Errorf("expected error about --decision, got: %v", err)
	}
}

func TestADRCreateCommand_MissingConsequences(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRCreateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"track-core",
		"--title", "Some title",
		"--context", "Some context",
		"--decision", "Some decision",
	})
	if err == nil {
		t.Errorf("expected error for missing consequences")
	}
	if !strings.Contains(err.Error(), "--consequences") {
		t.Errorf("expected error about --consequences, got: %v", err)
	}
}

// ============================================================================
// ADRListCommand Tests
// ============================================================================

func TestADRListCommand_NoADRs(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRListCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Command should succeed silently when no ADRs exist
	// (ADRListCommand writes to stdout via fmt.Println, not cmdCtx.GetStdout())
}

func TestADRListCommand_ListAll(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-core", "roadmap-test", "Core Framework", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create ADRs
	for i := 1; i <= 2; i++ {
		now := task_manager.GetCurrentTime()
		adr, _ := task_manager.NewADREntity(
			"DW-adr-"+string(rune(48+i)),
			"track-core",
			"ADR "+string(rune(48+i)),
			"proposed",
			"Context "+string(rune(48+i)),
			"Decision "+string(rune(48+i)),
			"Consequences "+string(rune(48+i)),
			"",
			now,
			now,
			nil,
		)
		repo.SaveADR(ctx, adr)
	}

	// Execute command
	cmd := &task_manager.ADRListCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify ADRs were listed by checking repository
	adrs, err := repo.ListADRs(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list ADRs: %v", err)
	}
	if len(adrs) != 2 {
		t.Errorf("expected 2 ADRs, got %d", len(adrs))
	}
}

func TestADRListCommand_FilterByTrack(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	// Create two tracks
	for i := 1; i <= 2; i++ {
		track, _ := task_manager.NewTrackEntity(
			"track-"+string(rune(48+i)),
			"roadmap-test",
			"Track "+string(rune(48+i)),
			"",
			"not-started",
			200,
			[]string{},
			time.Now().UTC(),
			time.Now().UTC(),
		)
		repo.SaveTrack(ctx, track)
	}

	// Create ADRs for both tracks
	for i := 1; i <= 2; i++ {
		trackID := "track-" + string(rune(48+i))
		now := task_manager.GetCurrentTime()
		adr, _ := task_manager.NewADREntity(
			"DW-adr-"+string(rune(48+i)),
			trackID,
			"ADR "+string(rune(48+i)),
			"proposed",
			"Context",
			"Decision",
			"Consequences",
			"",
			now,
			now,
			nil,
		)
		repo.SaveADR(ctx, adr)
	}

	// Execute command filtering by track
	cmd := &task_manager.ADRListCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"track-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify ADRs were filtered correctly
	adrs, err := repo.GetADRsByTrack(ctx, "track-1")
	if err != nil {
		t.Fatalf("failed to get ADRs by track: %v", err)
	}
	if len(adrs) != 1 {
		t.Errorf("expected 1 ADR, got %d", len(adrs))
	}
}

// ============================================================================
// ADRShowCommand Tests
// ============================================================================

func TestADRShowCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-core", "roadmap-test", "Core Framework", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create ADR
	now := task_manager.GetCurrentTime()
	adr, _ := task_manager.NewADREntity(
		"DW-adr-1",
		"track-core",
		"Use gRPC",
		"proposed",
		"Need efficient RPC",
		"Adopt gRPC",
		"Requires code generation",
		"",
		now,
		now,
		nil,
	)
	repo.SaveADR(ctx, adr)

	// Execute command
	cmd := &task_manager.ADRShowCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-adr-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify ADR was retrieved successfully (command succeeds)
	retrieved, err := repo.GetADR(ctx, "DW-adr-1")
	if err != nil {
		t.Fatalf("failed to retrieve ADR: %v", err)
	}
	if retrieved.Title != "Use gRPC" {
		t.Errorf("expected title 'Use gRPC', got '%s'", retrieved.Title)
	}
}

func TestADRShowCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRShowCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"nonexistent-adr"})
	if err == nil {
		t.Errorf("expected error for nonexistent ADR")
	}
}

// ============================================================================
// ADRUpdateCommand Tests
// ============================================================================

func TestADRUpdateCommand_UpdateTitle(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-core", "roadmap-test", "Core Framework", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create ADR
	now := task_manager.GetCurrentTime()
	adr, _ := task_manager.NewADREntity(
		"DW-adr-1",
		"track-core",
		"Old Title",
		"proposed",
		"Context",
		"Decision",
		"Consequences",
		"",
		now,
		now,
		nil,
	)
	repo.SaveADR(ctx, adr)

	// Execute command
	cmd := &task_manager.ADRUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-adr-1",
		"--title", "New Title",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify update
	updated, err := repo.GetADR(ctx, "DW-adr-1")
	if err != nil {
		t.Fatalf("failed to get ADR: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", updated.Title)
	}
}

func TestADRUpdateCommand_MultipleFields(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-core", "roadmap-test", "Core Framework", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create ADR
	now := task_manager.GetCurrentTime()
	adr, _ := task_manager.NewADREntity(
		"DW-adr-1",
		"track-core",
		"Title",
		"proposed",
		"Context",
		"Decision",
		"Consequences",
		"",
		now,
		now,
		nil,
	)
	repo.SaveADR(ctx, adr)

	// Execute command
	cmd := &task_manager.ADRUpdateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-adr-1",
		"--title", "New Title",
		"--decision", "New Decision",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify updates
	updated, err := repo.GetADR(ctx, "DW-adr-1")
	if err != nil {
		t.Fatalf("failed to get ADR: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", updated.Title)
	}
	if updated.Decision != "New Decision" {
		t.Errorf("expected decision 'New Decision', got '%s'", updated.Decision)
	}
}

func TestADRUpdateCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRUpdateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"nonexistent-adr",
		"--title", "New Title",
	})
	if err == nil {
		t.Errorf("expected error for nonexistent ADR")
	}
}

// ============================================================================
// ADRSupersdeCommand Tests
// ============================================================================

func TestADRSupersdeCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-core", "roadmap-test", "Core Framework", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create ADRs
	now := task_manager.GetCurrentTime()
	for i := 1; i <= 2; i++ {
		adr, _ := task_manager.NewADREntity(
			"DW-adr-"+string(rune(48+i)),
			"track-core",
			"ADR "+string(rune(48+i)),
			"proposed",
			"Context",
			"Decision",
			"Consequences",
			"",
			now,
			now,
			nil,
		)
		repo.SaveADR(ctx, adr)
	}

	// Execute command
	cmd := &task_manager.ADRSupersdeCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{
		"DW-adr-1",
		"--by", "DW-adr-2",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify ADR was superseded by checking the repository
	updated, err := repo.GetADR(ctx, "DW-adr-1")
	if err != nil {
		t.Fatalf("failed to get ADR: %v", err)
	}
	if updated.SupersededBy == nil || *updated.SupersededBy != "DW-adr-2" {
		t.Errorf("expected ADR to be superseded by DW-adr-2")
	}
}

func TestADRSupersdeCommand_MissingByFlag(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRSupersdeCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-adr-1"})
	if err == nil {
		t.Errorf("expected error for missing --by flag")
	}
	if !strings.Contains(err.Error(), "--by") {
		t.Errorf("expected error about --by flag, got: %v", err)
	}
}

// ============================================================================
// ADRDeprecateCommand Tests
// ============================================================================

func TestADRDeprecateCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	db := getProjectDB(t, tmpDir, "default")
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPlugin(
		&stubLogger{},
		tmpDir,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Set active project
	if err := os.WriteFile(filepath.Join(tmpDir, ".darwinflow", "active-project.txt"), []byte("default"), 0644); err != nil {
		t.Fatalf("failed to set active project: %v", err)
	}

	// Setup
	repo := task_manager.NewSQLiteRoadmapRepository(db, &stubLogger{})
	ctx := context.Background()

	roadmap, _ := task_manager.NewRoadmapEntity("roadmap-test", "Test vision", "Test criteria", time.Now().UTC(), time.Now().UTC())
	repo.SaveRoadmap(ctx, roadmap)

	track, _ := task_manager.NewTrackEntity("track-core", "roadmap-test", "Core Framework", "", "not-started", 200, []string{}, time.Now().UTC(), time.Now().UTC())
	repo.SaveTrack(ctx, track)

	// Create ADR
	now := task_manager.GetCurrentTime()
	adr, _ := task_manager.NewADREntity(
		"DW-adr-1",
		"track-core",
		"ADR",
		"proposed",
		"Context",
		"Decision",
		"Consequences",
		"",
		now,
		now,
		nil,
	)
	repo.SaveADR(ctx, adr)

	// Execute command
	cmd := &task_manager.ADRDeprecateCommand{Plugin: plugin}
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"DW-adr-1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify ADR was deprecated by checking the repository
	updated, err := repo.GetADR(ctx, "DW-adr-1")
	if err != nil {
		t.Fatalf("failed to get ADR: %v", err)
	}
	if updated.Status != string(task_manager.ADRStatusDeprecated) {
		t.Errorf("expected status 'deprecated', got '%s'", updated.Status)
	}
}

func TestADRDeprecateCommand_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	db := createRoadmapTestDB(t)
	defer db.Close()

	plugin, err := task_manager.NewTaskManagerPluginWithDatabase(
		&stubLogger{},
		tmpDir,
		db,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	cmd := &task_manager.ADRDeprecateCommand{Plugin: plugin}
	ctx := context.Background()
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     &bytes.Buffer{},
		logger:     &stubLogger{},
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"nonexistent-adr"})
	if err == nil {
		t.Errorf("expected error for nonexistent ADR")
	}
}

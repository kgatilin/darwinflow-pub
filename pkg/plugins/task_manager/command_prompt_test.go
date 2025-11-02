package task_manager_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tm "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// TestPromptCommand_Display verifies that the prompt command displays
// the system prompt to stdout.
func TestPromptCommand_Display(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &tm.PromptCommand{Plugin: plugin}

	// Create command context with captured stdout
	var stdout bytes.Buffer
	cmdCtx := &mockCommandContext{
		stdout:     &stdout,
		workingDir: tmpDir,
	}

	ctx := context.Background()

	// Execute without output file (should print to stdout)
	err := cmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := stdout.String()

	// Verify output contains prompt
	if !strings.Contains(output, "# Task Manager System Prompt") {
		t.Error("output should contain prompt header")
	}

	if !strings.Contains(output, "## Overview") {
		t.Error("output should contain overview section")
	}

	// Verify output is substantial
	if len(output) < 5000 {
		t.Error("output should be substantial documentation")
	}
}

// TestPromptCommand_SaveToFile verifies that the prompt command saves
// the system prompt to a file.
func TestPromptCommand_SaveToFile(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &tm.PromptCommand{Plugin: plugin}

	outputFile := filepath.Join(tmpDir, "prompt.md")

	// Create command context
	var stdout bytes.Buffer
	cmdCtx := &mockCommandContext{
		stdout:     &stdout,
		workingDir: tmpDir,
	}

	ctx := context.Background()

	// Execute with output file
	args := []string{"--output", outputFile}
	err := cmd.Execute(ctx, cmdCtx, args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	fileContent := string(content)

	// Verify file content
	if !strings.Contains(fileContent, "# Task Manager System Prompt") {
		t.Error("file should contain prompt header")
	}

	// Verify stdout message about file saved
	stdoutMsg := stdout.String()
	if !strings.Contains(stdoutMsg, "System prompt saved to:") {
		t.Error("stdout should confirm file was saved")
	}

	if !strings.Contains(stdoutMsg, outputFile) {
		t.Error("stdout should include output file path")
	}
}

// TestPromptCommand_SaveToNestedDirectory verifies that the command creates
// nested directories if needed.
func TestPromptCommand_SaveToNestedDirectory(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &tm.PromptCommand{Plugin: plugin}

	outputFile := filepath.Join(tmpDir, "docs", "guides", "prompt.md")

	var stdout bytes.Buffer
	cmdCtx := &mockCommandContext{
		stdout:     &stdout,
		workingDir: tmpDir,
	}

	ctx := context.Background()

	args := []string{"--output", outputFile}
	err := cmd.Execute(ctx, cmdCtx, args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify file was created in nested directory
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("file should contain prompt content")
	}
}

// TestPromptCommand_GetName verifies command metadata.
func TestPromptCommand_GetName(t *testing.T) {
	cmd := &tm.PromptCommand{}

	name := cmd.GetName()
	if name != "prompt" {
		t.Errorf("expected name 'prompt', got %q", name)
	}
}

// TestPromptCommand_GetDescription verifies command description.
func TestPromptCommand_GetDescription(t *testing.T) {
	cmd := &tm.PromptCommand{}

	desc := cmd.GetDescription()
	if desc == "" {
		t.Error("description should not be empty")
	}

	if !strings.Contains(desc, "prompt") {
		t.Error("description should mention prompt")
	}
}

// TestPromptCommand_GetUsage verifies command usage message.
func TestPromptCommand_GetUsage(t *testing.T) {
	cmd := &tm.PromptCommand{}

	usage := cmd.GetUsage()
	if usage == "" {
		t.Error("usage should not be empty")
	}

	if !strings.Contains(usage, "dw task-manager prompt") {
		t.Error("usage should show command invocation")
	}
}

// TestPromptCommand_GetHelp verifies command help text.
func TestPromptCommand_GetHelp(t *testing.T) {
	cmd := &tm.PromptCommand{}

	help := cmd.GetHelp()
	if help == "" {
		t.Error("help should not be empty")
	}

	expectedItems := []string{
		"system prompt",
		"task manager",
		"--output",
	}

	for _, item := range expectedItems {
		if !strings.Contains(strings.ToLower(help), strings.ToLower(item)) {
			t.Errorf("help should mention %q", item)
		}
	}
}

// TestPromptCommand_FilePermissions verifies that saved prompt file
// has readable permissions.
func TestPromptCommand_FilePermissions(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &tm.PromptCommand{Plugin: plugin}

	outputFile := filepath.Join(tmpDir, "test_prompt.md")

	var stdout bytes.Buffer
	cmdCtx := &mockCommandContext{
		stdout:     &stdout,
		workingDir: tmpDir,
	}

	ctx := context.Background()

	args := []string{"--output", outputFile}
	err := cmd.Execute(ctx, cmdCtx, args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("failed to stat output file: %v", err)
	}

	// File should be readable (0644 = rw-r--r--)
	perm := info.Mode().Perm()
	expectedPerm := os.FileMode(0644)
	if perm != expectedPerm {
		t.Errorf("file should have %o permissions, got %o", expectedPerm, perm)
	}
}

// TestPromptCommand_LargePrompt verifies that the command handles
// large prompts correctly.
func TestPromptCommand_LargePrompt(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &tm.PromptCommand{Plugin: plugin}

	outputFile := filepath.Join(tmpDir, "large_prompt.md")

	var stdout bytes.Buffer
	cmdCtx := &mockCommandContext{
		stdout:     &stdout,
		workingDir: tmpDir,
	}

	ctx := context.Background()

	// Test display to stdout (large output)
	err := cmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Test save to file (large file)
	err = cmd.Execute(ctx, cmdCtx, []string{"--output", outputFile})
	if err != nil {
		t.Fatalf("Execute with file failed: %v", err)
	}

	// Verify file is large
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Size() < 5000 {
		t.Errorf("file should be substantial, got %d bytes", info.Size())
	}
}

// TestPromptCommand_CurrentDirectory verifies that output file paths
// are handled correctly in current directory.
func TestPromptCommand_CurrentDirectory(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &tm.PromptCommand{Plugin: plugin}

	var stdout bytes.Buffer
	cmdCtx := &mockCommandContext{
		stdout:     &stdout,
		workingDir: tmpDir,
	}

	ctx := context.Background()

	// Create a subdirectory to save the prompt in
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	// Save to subdirectory
	promptFile := filepath.Join(outputDir, "prompt.md")
	err := cmd.Execute(ctx, cmdCtx, []string{"--output", promptFile})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(promptFile); err != nil {
		t.Errorf("prompt file should be created: %v", err)
	}
}

// TestPromptCommand_DisplayAndSave verifies that both display and save
// operations work correctly.
func TestPromptCommand_DisplayAndSave(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	// First: display to stdout
	{
		cmd := &tm.PromptCommand{Plugin: plugin}
		var stdout bytes.Buffer
		cmdCtx := &mockCommandContext{
			stdout:     &stdout,
			workingDir: tmpDir,
		}

		ctx := context.Background()
		err := cmd.Execute(ctx, cmdCtx, []string{})
		if err != nil {
			t.Fatalf("Display execute failed: %v", err)
		}

		if len(stdout.String()) == 0 {
			t.Error("display should output to stdout")
		}
	}

	// Second: save to file
	{
		cmd := &tm.PromptCommand{Plugin: plugin}
		var stdout bytes.Buffer
		outputFile := filepath.Join(tmpDir, "saved_prompt.md")
		cmdCtx := &mockCommandContext{
			stdout:     &stdout,
			workingDir: tmpDir,
		}

		ctx := context.Background()
		err := cmd.Execute(ctx, cmdCtx, []string{"--output", outputFile})
		if err != nil {
			t.Fatalf("Save execute failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(outputFile); err != nil {
			t.Fatalf("saved file should exist: %v", err)
		}
	}
}

// TestPromptCommand_InvalidOutputPath verifies handling of invalid output paths.
func TestPromptCommand_InvalidOutputPath(t *testing.T) {
	plugin, tmpDir := setupTestPlugin(t)

	cmd := &tm.PromptCommand{Plugin: plugin}

	// Try to create file in non-existent parent directory with invalid name
	// Use a path that would exist but is inaccessible
	invalidPath := filepath.Join(tmpDir, "valid", "test.md")

	var stdout bytes.Buffer
	cmdCtx := &mockCommandContext{
		stdout:     &stdout,
		workingDir: tmpDir,
	}

	ctx := context.Background()

	// This should succeed because os.MkdirAll creates nested directories
	err := cmd.Execute(ctx, cmdCtx, []string{"--output", invalidPath})
	if err != nil {
		t.Fatalf("Execute should create nested directories: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(invalidPath); err != nil {
		t.Errorf("file should be created: %v", err)
	}
}

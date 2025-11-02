package task_manager_test

import (
	"context"
	"strings"
	"testing"

	tm "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// TestDefaultSystemPrompt verifies that the default system prompt is defined
// and contains all expected sections.
func TestDefaultSystemPrompt(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	// Check that prompt is not empty
	if prompt == "" {
		t.Fatal("DefaultSystemPrompt should not be empty")
	}

	// Check for expected sections and keywords
	expectedSections := []string{
		"# Task Manager System Prompt",
		"## Overview",
		"## Entity Hierarchy",
		"### Roadmap",
		"### Track",
		"### Task",
		"### Iteration",
		"## Standard Workflows",
		"### Workflow 1: Initial Project Setup",
		"### Workflow 2: Iteration Planning",
		"### Workflow 3: Continuous Task Management",
		"### Workflow 4: Checking Track Dependencies",
		"## Best Practices",
		"## Entity Relationships Diagram",
		"## Status Transitions",
		"## Common Pitfalls to Avoid",
		"## Integration with Other Systems",
		"## Getting Help",
	}

	for _, section := range expectedSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("prompt missing expected section: %s", section)
		}
	}
}

// TestDefaultSystemPromptContainsCommands verifies that the prompt includes
// relevant command examples.
func TestDefaultSystemPromptContainsCommands(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	expectedCommands := []string{
		"dw task-manager roadmap init",
		"dw task-manager track create",
		"dw task-manager task create",
		"dw task-manager iteration create",
		"dw task-manager track add-dependency",
		"dw task-manager task list",
		"dw task-manager iteration start",
		"dw task-manager tui",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(prompt, cmd) {
			t.Errorf("prompt missing expected command: %s", cmd)
		}
	}
}

// TestDefaultSystemPromptContainsWorkflows verifies that the prompt includes
// practical workflow examples.
func TestDefaultSystemPromptContainsWorkflows(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	expectedWorkflows := []string{
		"Initial Project Setup",
		"Iteration Planning",
		"Continuous Task Management",
		"Checking Track Dependencies",
	}

	for _, workflow := range expectedWorkflows {
		if !strings.Contains(prompt, workflow) {
			t.Errorf("prompt missing expected workflow: %s", workflow)
		}
	}
}

// TestDefaultSystemPromptContainsBestPractices verifies that the prompt includes
// best practices section with practical guidance.
func TestDefaultSystemPromptContainsBestPractices(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	expectedPractices := []string{
		"Define Tracks Before Tasks",
		"Use Iterations for Structured Planning",
		"Track Priorities to Guide Work Order",
		"Use Status Consistently",
		"Check Dependencies Before Starting Tracks",
		"Use Git Branches with Tasks",
		"Regular Status Updates",
	}

	for _, practice := range expectedPractices {
		if !strings.Contains(prompt, practice) {
			t.Errorf("prompt missing expected best practice: %s", practice)
		}
	}
}

// TestDefaultSystemPromptContainsEntityHierarchy verifies that the prompt
// clearly explains the entity hierarchy.
func TestDefaultSystemPromptContainsEntityHierarchy(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	entities := []string{
		"Roadmap",
		"Track",
		"Task",
		"Iteration",
	}

	for _, entity := range entities {
		if !strings.Contains(prompt, entity) {
			t.Errorf("prompt missing entity: %s", entity)
		}
	}

	// Check that hierarchy is explained
	if !strings.Contains(prompt, "Roadmap → Track → Task → Iteration") {
		t.Error("prompt should explain the entity hierarchy order")
	}
}

// TestDefaultSystemPromptContainsStatusTransitions verifies that the prompt
// explains how entity statuses transition.
func TestDefaultSystemPromptContainsStatusTransitions(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	expectedTransitions := []string{
		"Track Status Flow",
		"Task Status Flow",
		"Iteration Status Flow",
	}

	for _, transition := range expectedTransitions {
		if !strings.Contains(prompt, transition) {
			t.Errorf("prompt missing expected status transition: %s", transition)
		}
	}
}

// TestGetSystemPrompt verifies that GetSystemPrompt returns the default prompt.
func TestGetSystemPrompt(t *testing.T) {
	ctx := context.Background()
	prompt := tm.GetSystemPrompt(ctx)

	// Should return the default prompt
	if prompt != tm.DefaultSystemPrompt {
		t.Error("GetSystemPrompt should return DefaultSystemPrompt")
	}

	// Should not be empty
	if prompt == "" {
		t.Error("GetSystemPrompt should not return empty string")
	}

	// Should contain key sections
	if !strings.Contains(prompt, "# Task Manager System Prompt") {
		t.Error("GetSystemPrompt result should contain header")
	}
}

// TestGetSystemPromptWithNilContext verifies that GetSystemPrompt works
// even with nil context (though context should never be nil in practice).
func TestGetSystemPromptWithContext(t *testing.T) {
	ctx := context.Background()
	prompt := tm.GetSystemPrompt(ctx)

	// Should work fine with background context
	if prompt == "" {
		t.Error("GetSystemPrompt with background context should not be empty")
	}
}

// TestDefaultSystemPromptConsistency verifies that the prompt is consistent
// across multiple calls (no dynamic generation).
func TestDefaultSystemPromptConsistency(t *testing.T) {
	ctx := context.Background()

	prompt1 := tm.GetSystemPrompt(ctx)
	prompt2 := tm.GetSystemPrompt(ctx)

	if prompt1 != prompt2 {
		t.Error("GetSystemPrompt should return consistent results")
	}
}

// TestDefaultSystemPromptLength verifies that the prompt is substantial
// (has comprehensive documentation).
func TestDefaultSystemPromptLength(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	// Prompt should be at least 5KB (substantial documentation)
	minLength := 5000
	if len(prompt) < minLength {
		t.Errorf("DefaultSystemPrompt should be at least %d bytes, got %d", minLength, len(prompt))
	}
}

// TestDefaultSystemPromptMarkdownFormat verifies that the prompt is in
// markdown format.
func TestDefaultSystemPromptMarkdownFormat(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	// Check for markdown headers
	if !strings.Contains(prompt, "# ") {
		t.Error("prompt should contain markdown h1 headers (#)")
	}

	if !strings.Contains(prompt, "## ") {
		t.Error("prompt should contain markdown h2 headers (##)")
	}

	if !strings.Contains(prompt, "### ") {
		t.Error("prompt should contain markdown h3 headers (###)")
	}

	// Check for markdown lists
	if !strings.Contains(prompt, "- ") {
		t.Error("prompt should contain markdown list items")
	}

	// Check for markdown tables
	if !strings.Contains(prompt, "| ") {
		t.Error("prompt should contain markdown tables")
	}
}

// TestDefaultSystemPromptContainsExamples verifies that the prompt includes
// practical examples.
func TestDefaultSystemPromptContainsExamples(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	if !strings.Contains(prompt, "Example") && !strings.Contains(prompt, "example") {
		t.Error("prompt should contain examples")
	}

	if !strings.Contains(prompt, "dw task-manager") {
		t.Error("prompt should contain command examples")
	}
}

// TestDefaultSystemPromptAccessible verifies the prompt is readable and
// not overly technical for new users.
func TestDefaultSystemPromptAccessible(t *testing.T) {
	prompt := tm.DefaultSystemPrompt

	// Should have clear sections
	if !strings.Contains(prompt, "## Overview") {
		t.Error("prompt should start with overview")
	}

	// Should have examples
	exampleCount := strings.Count(prompt, "Example")
	if exampleCount < 5 {
		t.Errorf("prompt should have multiple examples, found %d", exampleCount)
	}

	// Should explain the hierarchy clearly
	if !strings.Contains(prompt, "Roadmap") || !strings.Contains(prompt, "Track") ||
		!strings.Contains(prompt, "Task") || !strings.Contains(prompt, "Iteration") {
		t.Error("prompt should explain all entity types")
	}
}

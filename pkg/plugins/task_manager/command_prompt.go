package task_manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// PromptCommand displays the LLM system prompt for the task manager.
// This prompt explains the task manager's entity hierarchy and workflows to LLMs.
type PromptCommand struct {
	Plugin *TaskManagerPlugin
}

func (c *PromptCommand) GetName() string {
	return "prompt"
}

func (c *PromptCommand) GetDescription() string {
	return "Display LLM system prompt"
}

func (c *PromptCommand) GetUsage() string {
	return "dw task-manager prompt [--output <file>]"
}

func (c *PromptCommand) GetHelp() string {
	return `Displays the system prompt that explains the task manager to LLMs.

This prompt contains comprehensive documentation about the task manager's entity
hierarchy (Roadmap → Track → Task → Iteration), standard workflows, best practices,
and integration with other systems. Use this when working with AI assistants or
documenting the task manager usage.

The prompt explains:
- Entity definitions and relationships
- Required vs optional entities
- Standard workflows for different scenarios
- Best practices and pitfalls to avoid
- Integration with AC and ADR systems
- Command reference and examples

Flags:
  --output <file>  Save prompt to specified file instead of displaying
                   If not specified, prompt is printed to stdout

Examples:
  # Display prompt to terminal
  dw task-manager prompt

  # Save prompt to file for documentation
  dw task-manager prompt --output task-manager-prompt.md

  # Save and pipe to other tools
  dw task-manager prompt --output prompt.md && cat prompt.md

  # Use with LLM (example with Claude CLI)
  dw task-manager prompt | claude --system-prompt -

Notes:
  - Prompt is in markdown format
  - Suitable for saving as documentation
  - Contains complete reference for all workflows
  - Includes best practices and examples
  - Updated with latest features and patterns`
}

func (c *PromptCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	var outputFile string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		}
	}

	// Get the system prompt
	prompt := GetSystemPrompt(ctx)

	// If output file is specified, save to file
	if outputFile != "" {
		// Create directory if needed
		dir := filepath.Dir(outputFile)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		// Write prompt to file
		if err := os.WriteFile(outputFile, []byte(prompt), 0644); err != nil {
			return fmt.Errorf("failed to write prompt to file: %w", err)
		}

		fmt.Fprintf(cmdCtx.GetStdout(), "System prompt saved to: %s\n", outputFile)
		return nil
	}

	// Otherwise, print prompt to stdout
	fmt.Fprint(cmdCtx.GetStdout(), prompt)

	return nil
}

package app

import (
	"context"
	"io"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// ProjectCommandHandler handles project command operations
type ProjectCommandHandler struct {
	toolRegistry *ToolRegistry
	output       io.Writer
}

// NewProjectCommandHandler creates a new project command handler
func NewProjectCommandHandler(
	toolRegistry *ToolRegistry,
	output io.Writer,
) *ProjectCommandHandler {
	return &ProjectCommandHandler{
		toolRegistry: toolRegistry,
		output:       output,
	}
}

// ListTools displays all available project tools
func (h *ProjectCommandHandler) ListTools(ctx context.Context) error {
	listing := h.toolRegistry.ListTools()
	h.output.Write([]byte(listing))
	return nil
}

// ExecuteTool executes a specific tool with given arguments
func (h *ProjectCommandHandler) ExecuteTool(
	ctx context.Context,
	toolName string,
	args []string,
	projectCtx *domain.ProjectContext,
) error {
	return h.toolRegistry.ExecuteTool(ctx, toolName, args, projectCtx)
}

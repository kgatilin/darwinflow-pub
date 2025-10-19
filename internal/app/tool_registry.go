package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// ToolRegistry manages tools provided by plugins
type ToolRegistry struct {
	pluginRegistry *PluginRegistry
	logger         Logger
}

// NewToolRegistry creates a new ToolRegistry
func NewToolRegistry(pluginRegistry *PluginRegistry, logger Logger) *ToolRegistry {
	return &ToolRegistry{
		pluginRegistry: pluginRegistry,
		logger:         logger,
	}
}

// GetTool retrieves a tool by name from registered plugins
func (r *ToolRegistry) GetTool(toolName string) (domain.Tool, error) {
	plugins := r.pluginRegistry.GetAllPlugins()

	for _, plugin := range plugins {
		// Check if plugin implements IToolProvider
		toolProvider, ok := plugin.(domain.IToolProvider)
		if !ok {
			continue
		}

		// Search for tool in this plugin
		tools := toolProvider.GetTools()
		for _, tool := range tools {
			if tool.GetName() == toolName {
				r.logger.Debug("Found tool '%s' from plugin '%s'", toolName, plugin.GetInfo().Name)
				return tool, nil
			}
		}
	}

	return nil, fmt.Errorf("tool not found: %s", toolName)
}

// GetAllTools returns all tools from all registered plugins
func (r *ToolRegistry) GetAllTools() []domain.Tool {
	var allTools []domain.Tool
	plugins := r.pluginRegistry.GetAllPlugins()

	for _, plugin := range plugins {
		// Check if plugin implements IToolProvider
		toolProvider, ok := plugin.(domain.IToolProvider)
		if !ok {
			continue
		}

		tools := toolProvider.GetTools()
		allTools = append(allTools, tools...)
	}

	// Sort tools by name for consistent output
	sort.Slice(allTools, func(i, j int) bool {
		return allTools[i].GetName() < allTools[j].GetName()
	})

	return allTools
}

// ExecuteTool executes a tool with the given context and arguments
func (r *ToolRegistry) ExecuteTool(ctx context.Context, toolName string, args []string, projectCtx *domain.ProjectContext) error {
	tool, err := r.GetTool(toolName)
	if err != nil {
		return err
	}

	r.logger.Info("Executing tool: %s", toolName)
	return tool.Execute(ctx, args, projectCtx)
}

// ExecuteToolWithContext creates a ProjectContext and executes the tool
func (r *ToolRegistry) ExecuteToolWithContext(
	ctx context.Context,
	toolName string,
	args []string,
	eventRepo domain.EventRepository,
	analysisRepo domain.AnalysisRepository,
	config *domain.Config,
	cwd string,
	dbPath string,
) error {
	projectCtx := &domain.ProjectContext{
		EventRepo:    eventRepo,
		AnalysisRepo: analysisRepo,
		Config:       config,
		CWD:          cwd,
		DBPath:       dbPath,
	}

	return r.ExecuteTool(ctx, toolName, args, projectCtx)
}

// ListTools returns a formatted list of all available tools with their descriptions
func (r *ToolRegistry) ListTools() string {
	tools := r.GetAllTools()

	if len(tools) == 0 {
		return "No tools available"
	}

	output := "Available tools:\n"
	for _, tool := range tools {
		output += fmt.Sprintf("  %-15s %s\n", tool.GetName(), tool.GetDescription())
	}

	return output
}

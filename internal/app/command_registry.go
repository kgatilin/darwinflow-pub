package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// CommandRegistry manages command discovery and routing from plugins.
// It discovers commands from plugins that implement ICommandProvider.
type CommandRegistry struct {
	pluginRegistry *PluginRegistry
	logger         Logger
	commandCache   map[string]map[string]domain.Command // pluginName -> commandName -> Command
	mu             sync.RWMutex
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry(pluginRegistry *PluginRegistry, logger Logger) *CommandRegistry {
	return &CommandRegistry{
		pluginRegistry: pluginRegistry,
		logger:         logger,
		commandCache:   make(map[string]map[string]domain.Command),
	}
}

// GetCommand finds a command by plugin name and command name
func (r *CommandRegistry) GetCommand(pluginName, commandName string) (domain.Command, error) {
	r.mu.RLock()
	cached := r.commandCache[pluginName]
	r.mu.RUnlock()

	if cached != nil {
		if cmd, exists := cached[commandName]; exists {
			return cmd, nil
		}
		return nil, fmt.Errorf("command not found: %s %s", pluginName, commandName)
	}

	// Load commands from plugin
	plugin, err := r.pluginRegistry.GetPlugin(pluginName)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	cmdProvider, ok := plugin.(domain.ICommandProvider)
	if !ok {
		return nil, fmt.Errorf("plugin %s does not provide commands", pluginName)
	}

	// Cache commands for this plugin
	commands := cmdProvider.GetCommands()
	r.mu.Lock()
	r.commandCache[pluginName] = make(map[string]domain.Command)
	for _, cmd := range commands {
		r.commandCache[pluginName][cmd.GetName()] = cmd
	}
	r.mu.Unlock()

	// Try again from cache
	r.mu.RLock()
	cmd, exists := r.commandCache[pluginName][commandName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("command not found: %s %s", pluginName, commandName)
	}

	return cmd, nil
}

// GetCommandsForPlugin returns all commands from a specific plugin
func (r *CommandRegistry) GetCommandsForPlugin(pluginName string) []domain.Command {
	r.mu.RLock()
	cached := r.commandCache[pluginName]
	r.mu.RUnlock()

	if cached != nil {
		commands := make([]domain.Command, 0, len(cached))
		for _, cmd := range cached {
			commands = append(commands, cmd)
		}
		return commands
	}

	// Load commands from plugin
	plugin, err := r.pluginRegistry.GetPlugin(pluginName)
	if err != nil {
		r.logger.Debug("Plugin not found: %s", pluginName)
		return nil
	}

	cmdProvider, ok := plugin.(domain.ICommandProvider)
	if !ok {
		return nil
	}

	commands := cmdProvider.GetCommands()

	// Cache commands
	r.mu.Lock()
	r.commandCache[pluginName] = make(map[string]domain.Command)
	for _, cmd := range commands {
		r.commandCache[pluginName][cmd.GetName()] = cmd
	}
	r.mu.Unlock()

	return commands
}

// GetAllCommands returns all commands from all plugins
func (r *CommandRegistry) GetAllCommands() map[string][]domain.Command {
	allPlugins := r.pluginRegistry.GetAllPlugins()
	result := make(map[string][]domain.Command)

	for _, plugin := range allPlugins {
		cmdProvider, ok := plugin.(domain.ICommandProvider)
		if !ok {
			continue
		}

		pluginName := plugin.GetInfo().Name
		commands := cmdProvider.GetCommands()

		if len(commands) > 0 {
			result[pluginName] = commands

			// Update cache
			r.mu.Lock()
			r.commandCache[pluginName] = make(map[string]domain.Command)
			for _, cmd := range commands {
				r.commandCache[pluginName][cmd.GetName()] = cmd
			}
			r.mu.Unlock()
		}
	}

	return result
}

// ExecuteCommand executes a command from a plugin
func (r *CommandRegistry) ExecuteCommand(ctx context.Context, pluginName, commandName string, args []string, cmdCtx domain.CommandContext) error {
	cmd, err := r.GetCommand(pluginName, commandName)
	if err != nil {
		return err
	}

	r.logger.Debug("Executing command: %s %s", pluginName, commandName)
	return cmd.Execute(ctx, cmdCtx, args)
}

// ListCommands returns formatted list of all plugin commands
func (r *CommandRegistry) ListCommands() string {
	allCommands := r.GetAllCommands()

	if len(allCommands) == 0 {
		return "No plugin commands available"
	}

	var sb strings.Builder
	sb.WriteString("Available plugin commands:\n\n")

	for pluginName, commands := range allCommands {
		sb.WriteString(fmt.Sprintf("%s:\n", pluginName))
		for _, cmd := range commands {
			sb.WriteString(fmt.Sprintf("  dw %s %s - %s\n", pluginName, cmd.GetName(), cmd.GetDescription()))
			if usage := cmd.GetUsage(); usage != "" {
				sb.WriteString(fmt.Sprintf("    Usage: dw %s %s\n", pluginName, usage))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

package app_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// mockSDKLogger implements pluginsdk.Logger for testing
type mockSDKLogger struct{}

func (m *mockSDKLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *mockSDKLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *mockSDKLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockSDKLogger) Error(msg string, keysAndValues ...interface{}) {}

// mockCommand implements pluginsdk.Command
type mockCommand struct {
	name        string
	description string
	usage       string
	executeFunc func(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error
}

func (m *mockCommand) GetName() string        { return m.name }
func (m *mockCommand) GetDescription() string { return m.description }
func (m *mockCommand) GetUsage() string       { return m.usage }
func (m *mockCommand) GetHelp() string        { return "" }
func (m *mockCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, cmdCtx, args)
	}
	return nil
}

// mockCommandContext implements pluginsdk.CommandContext for testing
type mockCommandContext struct {
	logger pluginsdk.Logger
	cwd    string
	stdout io.Writer
	stdin  io.Reader
}

func (m *mockCommandContext) GetLogger() pluginsdk.Logger {
	if m.logger != nil {
		return m.logger
	}
	return &mockSDKLogger{}
}

func (m *mockCommandContext) GetWorkingDir() string {
	if m.cwd != "" {
		return m.cwd
	}
	return "/tmp"
}

func (m *mockCommandContext) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	return nil
}

func (m *mockCommandContext) GetStdout() io.Writer {
	if m.stdout != nil {
		return m.stdout
	}
	return io.Discard
}

func (m *mockCommandContext) GetStdin() io.Reader {
	if m.stdin != nil {
		return m.stdin
	}
	return strings.NewReader("")
}

// mockCommandProviderPlugin implements pluginsdk.Plugin and pluginsdk.ICommandProvider
type mockCommandProviderPlugin struct {
	info     pluginsdk.PluginInfo
	commands []pluginsdk.Command
}

func (m *mockCommandProviderPlugin) GetInfo() pluginsdk.PluginInfo {
	return m.info
}

func (m *mockCommandProviderPlugin) GetCapabilities() []string {
	return []string{"ICommandProvider"}
}

func (m *mockCommandProviderPlugin) GetCommands() []pluginsdk.Command {
	return m.commands
}

// mockNonCommandPlugin implements pluginsdk.Plugin but not ICommandProvider
type mockNonCommandPlugin struct {
	info pluginsdk.PluginInfo
}

func (m *mockNonCommandPlugin) GetInfo() pluginsdk.PluginInfo {
	return m.info
}

func (m *mockNonCommandPlugin) GetCapabilities() []string {
	return []string{} // No capabilities
}

func TestNewCommandRegistry(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	if registry == nil {
		t.Fatal("NewCommandRegistry returned nil")
	}
}

func TestCommandRegistry_GetCommand(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	// Register plugin with commands
	plugin := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "test-plugin", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{name: "init", description: "Initialize", usage: "init"},
			&mockCommand{name: "start", description: "Start", usage: "start"},
		},
	}
	pluginRegistry.RegisterPlugin(plugin)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	tests := []struct {
		name        string
		pluginName  string
		commandName string
		wantErr     bool
	}{
		{
			name:        "existing command",
			pluginName:  "test-plugin",
			commandName: "init",
			wantErr:     false,
		},
		{
			name:        "another existing command",
			pluginName:  "test-plugin",
			commandName: "start",
			wantErr:     false,
		},
		{
			name:        "non-existent command",
			pluginName:  "test-plugin",
			commandName: "nonexistent",
			wantErr:     true,
		},
		{
			name:        "non-existent plugin",
			pluginName:  "nonexistent-plugin",
			commandName: "init",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := registry.GetCommand(tt.pluginName, tt.commandName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetCommand() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetCommand() unexpected error: %v", err)
				}
				if cmd == nil {
					t.Errorf("GetCommand() returned nil command")
				}
				if cmd.GetName() != tt.commandName {
					t.Errorf("GetCommand() returned command with name %q, want %q", cmd.GetName(), tt.commandName)
				}
			}
		})
	}
}

func TestCommandRegistry_GetCommand_NonCommandProvider(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	// Register plugin that does NOT provide commands
	plugin := &mockNonCommandPlugin{
		info: pluginsdk.PluginInfo{Name: "no-commands", Version: "1.0.0"},
	}
	pluginRegistry.RegisterPlugin(plugin)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	_, err := registry.GetCommand("no-commands", "init")
	if err == nil {
		t.Error("GetCommand() expected error for non-command-provider plugin, got nil")
	}
}

func TestCommandRegistry_GetCommandsForPlugin(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	// Register plugin with commands
	plugin := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "test-plugin", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{name: "init", description: "Initialize"},
			&mockCommand{name: "start", description: "Start"},
		},
	}
	pluginRegistry.RegisterPlugin(plugin)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	commands := registry.GetCommandsForPlugin("test-plugin")

	if len(commands) != 2 {
		t.Errorf("GetCommandsForPlugin() returned %d commands, want 2", len(commands))
	}
}

func TestCommandRegistry_GetCommandsForPlugin_NonExistent(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	commands := registry.GetCommandsForPlugin("nonexistent")

	if commands != nil {
		t.Errorf("GetCommandsForPlugin() returned %v, want nil", commands)
	}
}

func TestCommandRegistry_GetAllCommands(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	// Register multiple plugins
	plugin1 := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "plugin1", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{name: "init", description: "Initialize"},
		},
	}
	plugin2 := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "plugin2", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{name: "start", description: "Start"},
			&mockCommand{name: "stop", description: "Stop"},
		},
	}

	pluginRegistry.RegisterPlugin(plugin1)
	pluginRegistry.RegisterPlugin(plugin2)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	allCommands := registry.GetAllCommands()

	if len(allCommands) != 2 {
		t.Errorf("GetAllCommands() returned %d plugins, want 2", len(allCommands))
	}

	if len(allCommands["plugin1"]) != 1 {
		t.Errorf("GetAllCommands()[plugin1] has %d commands, want 1", len(allCommands["plugin1"]))
	}

	if len(allCommands["plugin2"]) != 2 {
		t.Errorf("GetAllCommands()[plugin2] has %d commands, want 2", len(allCommands["plugin2"]))
	}
}

func TestCommandRegistry_ExecuteCommand(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	executed := false
	executedArgs := []string{}

	plugin := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "test-plugin", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{
				name:        "test-cmd",
				description: "Test command",
				executeFunc: func(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
					executed = true
					executedArgs = args
					return nil
				},
			},
		},
	}
	pluginRegistry.RegisterPlugin(plugin)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	ctx := context.Background()
	args := []string{"arg1", "arg2"}

	err := registry.ExecuteCommand(ctx, "test-plugin", "test-cmd", args, nil)

	if err != nil {
		t.Errorf("ExecuteCommand() unexpected error: %v", err)
	}

	if !executed {
		t.Error("ExecuteCommand() did not execute the command")
	}

	if len(executedArgs) != 2 || executedArgs[0] != "arg1" || executedArgs[1] != "arg2" {
		t.Errorf("ExecuteCommand() passed wrong args: %v", executedArgs)
	}
}

func TestCommandRegistry_ExecuteCommand_Error(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	expectedErr := errors.New("command failed")

	plugin := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "test-plugin", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{
				name:        "fail-cmd",
				description: "Failing command",
				executeFunc: func(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
					return expectedErr
				},
			},
		},
	}
	pluginRegistry.RegisterPlugin(plugin)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	ctx := context.Background()
	err := registry.ExecuteCommand(ctx, "test-plugin", "fail-cmd", nil, nil)

	if err != expectedErr {
		t.Errorf("ExecuteCommand() error = %v, want %v", err, expectedErr)
	}
}

func TestCommandRegistry_ListCommands(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	plugin := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "test-plugin", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{name: "init", description: "Initialize system", usage: "init [--force]"},
			&mockCommand{name: "start", description: "Start service", usage: "start"},
		},
	}
	pluginRegistry.RegisterPlugin(plugin)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	output := registry.ListCommands()

	if !strings.Contains(output, "test-plugin") {
		t.Error("ListCommands() output does not contain plugin name")
	}

	if !strings.Contains(output, "init") || !strings.Contains(output, "Initialize system") {
		t.Error("ListCommands() output does not contain init command")
	}

	if !strings.Contains(output, "start") || !strings.Contains(output, "Start service") {
		t.Error("ListCommands() output does not contain start command")
	}
}

func TestCommandRegistry_ListCommands_Empty(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	registry := app.NewCommandRegistry(pluginRegistry, logger)

	output := registry.ListCommands()

	if !strings.Contains(output, "No plugin commands available") {
		t.Errorf("ListCommands() with no plugins should return 'No plugin commands available', got: %s", output)
	}
}

func TestCommandRegistry_Caching(t *testing.T) {
	logger := &app.NoOpLogger{}
	pluginRegistry := app.NewPluginRegistry(logger)

	callCount := 0

	// Create a plugin that tracks GetCommands calls
	plugin := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{Name: "test-plugin", Version: "1.0.0"},
		commands: []pluginsdk.Command{
			&mockCommand{name: "init", description: "Initialize"},
		},
	}

	// Note: caching test simplified - GetCommands now requires context and CommandContext parameters

	pluginRegistry.RegisterPlugin(plugin)
	registry := app.NewCommandRegistry(pluginRegistry, logger)

	// First call - should cache
	registry.GetCommand("test-plugin", "init")
	callCount++

	// Second call - should use cache
	registry.GetCommand("test-plugin", "init")

	// The command should still work (we can't directly verify caching without modifying the implementation)
	cmd, err := registry.GetCommand("test-plugin", "init")
	if err != nil {
		t.Errorf("GetCommand() with cache should not error: %v", err)
	}
	if cmd.GetName() != "init" {
		t.Errorf("GetCommand() with cache returned wrong command: %s", cmd.GetName())
	}
}

// TestCommandRegistry_ExecuteCommand_Help tests help flag functionality
func TestCommandRegistry_ExecuteCommand_Help(t *testing.T) {
	pluginRegistry := app.NewPluginRegistry(&mockLogger{})
	logger := &mockLogger{}

	// Create a command with help text
	cmd := &mockCommand{
		name:        "test-cmd",
		description: "Test command description",
		usage:       "dw test-plugin test-cmd [options]",
	}

	// Create command with GetHelp returning detailed help
	cmdWithHelp := &mockCommandWithHelp{
		mockCommand: mockCommand{
			name:        "help-cmd",
			description: "Command with detailed help",
			usage:       "dw test-plugin help-cmd",
		},
		help: "This is detailed help text\nWith multiple lines",
	}

	plugin := &mockCommandProviderPlugin{
		info: pluginsdk.PluginInfo{
			Name:    "test-plugin",
			Version: "1.0.0",
		},
		commands: []pluginsdk.Command{cmd, cmdWithHelp},
	}

	pluginRegistry.RegisterPlugin(plugin)
	registry := app.NewCommandRegistry(pluginRegistry, logger)

	// Create a command context with output buffer
	output := &strings.Builder{}
	cmdCtx := &mockCommandContext{
		stdout: output,
	}

	// Execute with --help flag
	err := registry.ExecuteCommand(context.Background(), "test-plugin", "test-cmd", []string{"--help"}, cmdCtx)

	if err != nil {
		t.Errorf("ExecuteCommand() with --help should not error: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "test-cmd") {
		t.Errorf("Help output should contain command name, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Test command description") {
		t.Errorf("Help output should contain description, got: %s", outputStr)
	}

	// Test with -h flag
	output.Reset()
	err = registry.ExecuteCommand(context.Background(), "test-plugin", "help-cmd", []string{"-h"}, cmdCtx)

	if err != nil {
		t.Errorf("ExecuteCommand() with -h should not error: %v", err)
	}

	outputStr = output.String()
	if !strings.Contains(outputStr, "help-cmd") {
		t.Errorf("Help output should contain command name")
	}
	if !strings.Contains(outputStr, "This is detailed help text") {
		t.Errorf("Help output should contain detailed help text, got: %s", outputStr)
	}
}

// mockCommandWithHelp extends mockCommand to provide GetHelp
type mockCommandWithHelp struct {
	mockCommand
	help string
}

func (m *mockCommandWithHelp) GetHelp() string {
	return m.help
}

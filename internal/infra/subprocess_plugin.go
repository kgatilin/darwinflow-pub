package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// SubprocessPlugin is an adapter that wraps an external plugin process.
// It implements all SDK plugin interfaces and delegates calls to the subprocess via RPC.
type SubprocessPlugin struct {
	client       *RPCClient
	info         pluginsdk.PluginInfo
	capabilities []string

	// Command adapter
	commands map[string]*subprocessCommand

	// Entity type cache
	entityTypes []pluginsdk.EntityTypeInfo
}

// NewSubprocessPlugin creates a new subprocess plugin wrapper.
// The plugin process is not started until Initialize() is called.
func NewSubprocessPlugin(executablePath string, args ...string) *SubprocessPlugin {
	return &SubprocessPlugin{
		client:   NewRPCClient(executablePath, args...),
		commands: make(map[string]*subprocessCommand),
	}
}

// Initialize starts the subprocess and retrieves plugin metadata.
// This must be called before using the plugin.
func (p *SubprocessPlugin) Initialize(ctx context.Context, workingDir string, config map[string]interface{}) error {
	// Start subprocess
	if err := p.client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start subprocess: %w", err)
	}

	// Initialize plugin
	initParams := pluginsdk.InitParams{
		WorkingDir: workingDir,
		Config:     config,
	}
	if _, err := p.client.Call(ctx, pluginsdk.RPCMethodInit, initParams); err != nil {
		p.client.Stop()
		return fmt.Errorf("plugin initialization failed: %w", err)
	}

	// Get plugin info
	result, err := p.client.Call(ctx, pluginsdk.RPCMethodGetInfo, nil)
	if err != nil {
		p.client.Stop()
		return fmt.Errorf("failed to get plugin info: %w", err)
	}
	if err := json.Unmarshal(result, &p.info); err != nil {
		p.client.Stop()
		return fmt.Errorf("failed to parse plugin info: %w", err)
	}

	// Get capabilities
	result, err = p.client.Call(ctx, pluginsdk.RPCMethodGetCapabilities, nil)
	if err != nil {
		p.client.Stop()
		return fmt.Errorf("failed to get capabilities: %w", err)
	}
	if err := json.Unmarshal(result, &p.capabilities); err != nil {
		p.client.Stop()
		return fmt.Errorf("failed to parse capabilities: %w", err)
	}

	// Load commands if plugin supports ICommandProvider
	if p.hasCapability("ICommandProvider") {
		if err := p.loadCommands(ctx); err != nil {
			p.client.Stop()
			return fmt.Errorf("failed to load commands: %w", err)
		}
	}

	// Load entity types if plugin supports IEntityProvider
	if p.hasCapability("IEntityProvider") {
		if err := p.loadEntityTypes(ctx); err != nil {
			p.client.Stop()
			return fmt.Errorf("failed to load entity types: %w", err)
		}
	}

	return nil
}

// Shutdown gracefully stops the subprocess.
func (p *SubprocessPlugin) Shutdown() error {
	return p.client.Stop()
}

// GetInfo returns plugin metadata.
func (p *SubprocessPlugin) GetInfo() pluginsdk.PluginInfo {
	return p.info
}

// GetCapabilities returns the list of capability interfaces this plugin implements.
func (p *SubprocessPlugin) GetCapabilities() []string {
	return p.capabilities
}

// hasCapability checks if the plugin has a specific capability.
func (p *SubprocessPlugin) hasCapability(capability string) bool {
	for _, cap := range p.capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// GetEntityTypes returns entity type metadata (IEntityProvider).
func (p *SubprocessPlugin) GetEntityTypes() []pluginsdk.EntityTypeInfo {
	return p.entityTypes
}

// Query queries entities (IEntityProvider).
func (p *SubprocessPlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
	result, err := p.client.Call(ctx, pluginsdk.RPCMethodQueryEntities, query)
	if err != nil {
		return nil, err
	}

	// Unmarshal to raw entity data
	var rawEntities []map[string]interface{}
	if err := json.Unmarshal(result, &rawEntities); err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	// Wrap in entity adapters
	entities := make([]pluginsdk.IExtensible, len(rawEntities))
	for i, raw := range rawEntities {
		entities[i] = &subprocessEntity{data: raw}
	}

	return entities, nil
}

// GetEntity retrieves a specific entity by ID (IEntityProvider).
func (p *SubprocessPlugin) GetEntity(ctx context.Context, entityID string) (pluginsdk.IExtensible, error) {
	params := pluginsdk.GetEntityParams{EntityID: entityID}
	result, err := p.client.Call(ctx, pluginsdk.RPCMethodGetEntity, params)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(result, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse entity: %w", err)
	}

	return &subprocessEntity{data: raw}, nil
}

// UpdateEntity updates an entity (IEntityUpdater).
func (p *SubprocessPlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (pluginsdk.IExtensible, error) {
	params := pluginsdk.UpdateEntityParams{
		EntityID: entityID,
		Fields:   fields,
	}
	result, err := p.client.Call(ctx, pluginsdk.RPCMethodUpdateEntity, params)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(result, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse updated entity: %w", err)
	}

	return &subprocessEntity{data: raw}, nil
}

// GetCommands returns all commands provided by the plugin (ICommandProvider).
func (p *SubprocessPlugin) GetCommands() []pluginsdk.Command {
	commands := make([]pluginsdk.Command, 0, len(p.commands))
	for _, cmd := range p.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// StartEventStream starts streaming events from the plugin (IEventEmitter).
func (p *SubprocessPlugin) StartEventStream(ctx context.Context, eventChan chan<- pluginsdk.Event) error {
	// Set event channel on RPC client
	p.client.SetEventChannel(eventChan)

	// Call start_event_stream
	_, err := p.client.Call(ctx, pluginsdk.RPCMethodStartEventStream, nil)
	return err
}

// StopEventStream stops streaming events (IEventEmitter).
func (p *SubprocessPlugin) StopEventStream() error {
	_, err := p.client.Call(context.Background(), pluginsdk.RPCMethodStopEventStream, nil)
	return err
}

// loadCommands fetches command metadata from the subprocess.
func (p *SubprocessPlugin) loadCommands(ctx context.Context) error {
	result, err := p.client.Call(ctx, pluginsdk.RPCMethodGetCommands, nil)
	if err != nil {
		return err
	}

	var commandInfos []pluginsdk.CommandInfo
	if err := json.Unmarshal(result, &commandInfos); err != nil {
		return fmt.Errorf("failed to parse commands: %w", err)
	}

	// Create command adapters
	for _, info := range commandInfos {
		p.commands[info.Name] = &subprocessCommand{
			plugin: p,
			info:   info,
		}
	}

	return nil
}

// loadEntityTypes fetches entity type metadata from the subprocess.
func (p *SubprocessPlugin) loadEntityTypes(ctx context.Context) error {
	result, err := p.client.Call(ctx, pluginsdk.RPCMethodGetEntityTypes, nil)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(result, &p.entityTypes); err != nil {
		return fmt.Errorf("failed to parse entity types: %w", err)
	}

	return nil
}

// subprocessCommand is an adapter for external plugin commands.
type subprocessCommand struct {
	plugin *SubprocessPlugin
	info   pluginsdk.CommandInfo
}

func (c *subprocessCommand) GetName() string {
	return c.info.Name
}

func (c *subprocessCommand) GetDescription() string {
	return c.info.Description
}

func (c *subprocessCommand) GetUsage() string {
	return c.info.Usage
}

func (c *subprocessCommand) GetHelp() string {
	return c.info.Help
}

func (c *subprocessCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	params := pluginsdk.ExecuteCommandParams{
		CommandName: c.info.Name,
		Args:        args,
	}

	result, err := c.plugin.client.Call(ctx, pluginsdk.RPCMethodExecuteCommand, params)
	if err != nil {
		return err
	}

	var cmdResult pluginsdk.ExecuteCommandResult
	if err := json.Unmarshal(result, &cmdResult); err != nil {
		return fmt.Errorf("failed to parse command result: %w", err)
	}

	// Write output to command context
	if cmdResult.Output != "" {
		fmt.Fprint(cmdCtx.GetStdout(), cmdResult.Output)
	}

	// Check exit code
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d: %s", cmdResult.ExitCode, cmdResult.Error)
	}

	return nil
}

// subprocessEntity is an adapter for entities from external plugins.
type subprocessEntity struct {
	data map[string]interface{}
}

func (e *subprocessEntity) GetID() string {
	if id, ok := e.data["id"].(string); ok {
		return id
	}
	return ""
}

func (e *subprocessEntity) GetType() string {
	if t, ok := e.data["type"].(string); ok {
		return t
	}
	return ""
}

func (e *subprocessEntity) GetCapabilities() []string {
	if caps, ok := e.data["capabilities"].([]interface{}); ok {
		result := make([]string, len(caps))
		for i, cap := range caps {
			if s, ok := cap.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return []string{}
}

func (e *subprocessEntity) GetField(name string) interface{} {
	if val, ok := e.data[name]; ok {
		return val
	}
	return nil
}

func (e *subprocessEntity) GetAllFields() map[string]interface{} {
	return e.data
}

func (e *subprocessEntity) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.data)
}

func (e *subprocessEntity) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &e.data)
}

// Verify interface implementations at compile time
var _ pluginsdk.Plugin = (*SubprocessPlugin)(nil)
var _ pluginsdk.IEntityProvider = (*SubprocessPlugin)(nil)
var _ pluginsdk.IEntityUpdater = (*SubprocessPlugin)(nil)
var _ pluginsdk.ICommandProvider = (*SubprocessPlugin)(nil)
var _ pluginsdk.IEventEmitter = (*SubprocessPlugin)(nil)
var _ pluginsdk.Command = (*subprocessCommand)(nil)
var _ pluginsdk.IExtensible = (*subprocessEntity)(nil)

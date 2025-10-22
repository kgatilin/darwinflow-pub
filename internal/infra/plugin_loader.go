package infra

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// PluginConfig represents the configuration for a single plugin.
// It is the internal representation loaded from plugins.yaml.
type PluginConfig struct {
	Command        string            `yaml:"command"`
	Args           []string          `yaml:"args"`
	Env            map[string]string `yaml:"env"`
	Enabled        *bool             `yaml:"enabled"`        // Pointer to distinguish between unset and false
	Timeout        int               `yaml:"timeout"`        // seconds
	RestartOnCrash bool              `yaml:"restart_on_crash"`
}

// IsEnabled returns true if the plugin is enabled.
// If Enabled is nil (not set in YAML), defaults to true.
func (c *PluginConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true // Default to enabled
	}
	return *c.Enabled
}

// GetTimeout returns the RPC timeout in seconds.
// If Timeout is 0 (not set), defaults to 30 seconds.
func (c *PluginConfig) GetTimeout() int {
	if c.Timeout == 0 {
		return 30 // Default timeout
	}
	return c.Timeout
}

// pluginsYAML represents the top-level structure of plugins.yaml
type pluginsYAML struct {
	Plugins map[string]PluginConfig `yaml:"plugins"`
}

// PluginLoader loads external plugins from a YAML configuration file.
type PluginLoader struct {
	logger *Logger
}

// NewPluginLoader creates a new plugin loader.
func NewPluginLoader(logger *Logger) *PluginLoader {
	return &PluginLoader{
		logger: logger,
	}
}

// LoadFromConfig loads plugins from a plugins.yaml configuration file.
//
// The configPath should be the full path to the plugins.yaml file.
// If the file doesn't exist, returns an empty list with no error.
//
// Behavior:
// - Skips plugins with enabled=false
// - Skips plugins where the command executable doesn't exist (logs warning)
// - Resolves relative paths relative to the .darwinflow/ directory
// - Returns all successfully loaded plugins
// - Collects and returns warnings for skipped plugins
//
// Returns:
// - List of successfully loaded plugins
// - Error if the YAML is invalid or there's a critical loading issue
func (l *PluginLoader) LoadFromConfig(configPath string) ([]pluginsdk.Plugin, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if l.logger != nil {
			l.logger.Debug("Plugin config file not found: %s (no plugins will be loaded)", configPath)
		}
		return []pluginsdk.Plugin{}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %w", err)
	}

	// Parse YAML
	var config pluginsYAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse plugin config: %w", err)
	}

	// Determine base directory for relative path resolution
	configDir := filepath.Dir(configPath)

	// Load each plugin
	var plugins []pluginsdk.Plugin
	for name, pluginCfg := range config.Plugins {
		// Skip disabled plugins
		if !pluginCfg.IsEnabled() {
			if l.logger != nil {
				l.logger.Debug("Skipping disabled plugin: %s", name)
			}
			continue
		}

		// Validate command is specified
		if pluginCfg.Command == "" {
			if l.logger != nil {
				l.logger.Warn("Skipping plugin '%s': command is required", name)
			}
			continue
		}

		// Resolve command path
		cmdPath := pluginCfg.Command
		if !filepath.IsAbs(cmdPath) {
			// Relative path - resolve relative to config directory
			cmdPath = filepath.Join(configDir, cmdPath)
		}

		// Check if command executable exists
		if err := l.validateCommand(cmdPath); err != nil {
			if l.logger != nil {
				l.logger.Warn("Skipping plugin '%s': %v", name, err)
			}
			continue
		}

		// Create subprocess plugin
		plugin := l.createSubprocessPlugin(name, cmdPath, pluginCfg)
		plugins = append(plugins, plugin)

		if l.logger != nil {
			l.logger.Info("Loaded plugin configuration: %s (command: %s)", name, cmdPath)
		}
	}

	return plugins, nil
}

// validateCommand checks if the command exists and is executable.
func (l *PluginLoader) validateCommand(cmdPath string) error {
	// First check if it's an absolute path that exists
	if filepath.IsAbs(cmdPath) {
		info, err := os.Stat(cmdPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("command not found: %s", cmdPath)
		}
		if err != nil {
			return fmt.Errorf("failed to check command: %w", err)
		}

		// Check if it's a regular file
		if !info.Mode().IsRegular() {
			return fmt.Errorf("command is not a regular file: %s", cmdPath)
		}

		// Check if it's executable
		if info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("command is not executable: %s", cmdPath)
		}

		return nil
	}

	// For non-absolute paths (e.g., "python3"), use exec.LookPath
	_, err := exec.LookPath(cmdPath)
	if err != nil {
		return fmt.Errorf("command not found in PATH: %s", cmdPath)
	}

	return nil
}

// createSubprocessPlugin creates a SubprocessPlugin from the configuration.
func (l *PluginLoader) createSubprocessPlugin(name, cmdPath string, cfg PluginConfig) pluginsdk.Plugin {
	// Create subprocess plugin with command and args
	plugin := NewSubprocessPlugin(cmdPath, cfg.Args...)

	// TODO: Future enhancement - set environment variables on the subprocess
	// This would require extending SubprocessPlugin to accept env vars
	// For now, we document this limitation

	if l.logger != nil && len(cfg.Env) > 0 {
		l.logger.Warn("Plugin '%s': environment variables not yet supported (will be added in future)", name)
	}

	// TODO: Future enhancement - implement timeout configuration
	// This would require extending RPCClient to accept custom timeout
	if l.logger != nil && cfg.Timeout != 30 {
		l.logger.Warn("Plugin '%s': custom timeout not yet supported (will be added in future)", name)
	}

	// TODO: Future enhancement - implement restart on crash
	// This would require process monitoring in SubprocessPlugin
	if l.logger != nil && cfg.RestartOnCrash {
		l.logger.Warn("Plugin '%s': restart_on_crash not yet supported (will be added in future)", name)
	}

	return plugin
}

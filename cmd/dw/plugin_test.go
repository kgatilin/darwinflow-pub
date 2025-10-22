package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestPluginListCommand tests the 'dw plugin list' command
func TestPluginListCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run plugin list command
	pluginCmd([]string{"list"})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected content
	if output == "" {
		t.Error("plugin list produced no output")
	}

	// Should contain core plugins
	expectedPlugins := []string{"claude-code", "task-manager"}
	for _, plugin := range expectedPlugins {
		if !bytes.Contains(buf.Bytes(), []byte(plugin)) {
			t.Errorf("output missing expected plugin: %s", plugin)
		}
	}
}

// TestPluginReloadCommand_NoConfig tests reload when plugins.yaml doesn't exist
func TestPluginReloadCommand_NoConfig(t *testing.T) {
	// Create temp dir that doesn't have plugins.yaml
	tmpDir := t.TempDir()
	darwinflowDir := filepath.Join(tmpDir, ".darwinflow")
	logsDir := filepath.Join(darwinflowDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run plugin reload command
	pluginCmd([]string{"reload"})

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Should mention no plugins.yaml found
	if !bytes.Contains(buf.Bytes(), []byte("No plugins.yaml")) {
		t.Error("expected message about missing plugins.yaml")
	}
}

// TestPluginCommand_Help tests plugin command help
func TestPluginCommand_Help(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{"main help", []string{"--help"}},
		{"list help", []string{"list", "--help"}},
		{"reload help", []string{"reload", "--help"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			pluginCmd(tc.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify output contains help text
			if output == "" {
				t.Error("help command produced no output")
			}
			if !bytes.Contains(buf.Bytes(), []byte("Usage:")) {
				t.Error("help output should contain 'Usage:'")
			}
		})
	}
}

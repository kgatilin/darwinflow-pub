package pluginsdk_test

import (
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

func TestHookConfiguration_Basic(t *testing.T) {
	config := pluginsdk.HookConfiguration{
		TriggerType: "trigger.tool.before",
		Name:        "PreToolUse",
		Description: "Before any tool execution",
		Command:     "dw claude-code emit-event",
		Timeout:     5,
	}

	if config.TriggerType != "trigger.tool.before" {
		t.Errorf("Expected trigger type 'trigger.tool.before', got %s", config.TriggerType)
	}

	if config.Name != "PreToolUse" {
		t.Errorf("Expected name 'PreToolUse', got %s", config.Name)
	}

	if config.Command != "dw claude-code emit-event" {
		t.Errorf("Expected command 'dw claude-code emit-event', got %s", config.Command)
	}

	if config.Timeout != 5 {
		t.Errorf("Expected timeout 5, got %d", config.Timeout)
	}
}

func TestHookConfiguration_AllTriggerTypes(t *testing.T) {
	triggers := []struct {
		name    string
		trigger string
	}{
		{"PreToolUse", "trigger.tool.before"},
		{"PostToolUse", "trigger.tool.after"},
		{"UserInput", "trigger.user.input"},
		{"SessionStart", "trigger.session.start"},
		{"SessionEnd", "trigger.session.end"},
	}

	for _, tt := range triggers {
		t.Run(tt.name, func(t *testing.T) {
			config := pluginsdk.HookConfiguration{
				TriggerType: tt.trigger,
				Name:        tt.name,
			}

			if config.TriggerType != tt.trigger {
				t.Errorf("Expected trigger %s, got %s", tt.trigger, config.TriggerType)
			}
		})
	}
}

func TestHookConfiguration_WithoutTimeout(t *testing.T) {
	config := pluginsdk.HookConfiguration{
		TriggerType: "trigger.tool.before",
		Name:        "PreToolUse",
		Command:     "dw claude emit-event",
		Timeout:     0, // No timeout
	}

	if config.Timeout != 0 {
		t.Errorf("Expected timeout 0 (no limit), got %d", config.Timeout)
	}
}

func TestHookConfiguration_EmptyFields(t *testing.T) {
	config := pluginsdk.HookConfiguration{}

	if config.TriggerType != "" {
		t.Errorf("Expected empty trigger type, got %s", config.TriggerType)
	}

	if config.Name != "" {
		t.Errorf("Expected empty name, got %s", config.Name)
	}

	if config.Command != "" {
		t.Errorf("Expected empty command, got %s", config.Command)
	}

	if config.Timeout != 0 {
		t.Errorf("Expected timeout 0, got %d", config.Timeout)
	}
}

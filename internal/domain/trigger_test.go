package domain_test

import (
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

func TestTriggerTypes(t *testing.T) {
	// Test that all trigger type constants are defined
	triggerTypes := []domain.TriggerType{
		// Legacy types
		domain.TriggerPreToolUse,
		domain.TriggerPostToolUse,
		domain.TriggerNotification,
		domain.TriggerUserPromptSubmit,
		domain.TriggerStop,
		domain.TriggerSubagentStop,
		domain.TriggerPreCompact,
		// New generic types
		domain.TriggerBeforeToolUse,
		domain.TriggerAfterToolUse,
		domain.TriggerUserInput,
		domain.TriggerSessionStart,
		domain.TriggerSessionEnd,
	}

	// Verify each type has a non-empty value
	for _, tt := range triggerTypes {
		if string(tt) == "" {
			t.Errorf("Trigger type %v has empty string value", tt)
		}
	}

	// Verify types are unique
	seen := make(map[domain.TriggerType]bool)
	for _, tt := range triggerTypes {
		if seen[tt] {
			t.Errorf("Duplicate trigger type: %v", tt)
		}
		seen[tt] = true
	}
}

func TestTriggerType_Values(t *testing.T) {
	tests := []struct {
		trigger  domain.TriggerType
		expected string
	}{
		// Legacy trigger types (for backward compatibility)
		{domain.TriggerPreToolUse, "PreToolUse"},
		{domain.TriggerPostToolUse, "PostToolUse"},
		{domain.TriggerNotification, "Notification"},
		{domain.TriggerUserPromptSubmit, "UserPromptSubmit"},
		{domain.TriggerStop, "Stop"},
		{domain.TriggerSubagentStop, "SubagentStop"},
		{domain.TriggerPreCompact, "PreCompact"},
		// New generic trigger types
		{domain.TriggerBeforeToolUse, "trigger.tool.before"},
		{domain.TriggerAfterToolUse, "trigger.tool.after"},
		{domain.TriggerUserInput, "trigger.user.input"},
		{domain.TriggerSessionStart, "trigger.session.start"},
		{domain.TriggerSessionEnd, "trigger.session.end"},
	}

	for _, tt := range tests {
		t.Run(string(tt.trigger), func(t *testing.T) {
			if string(tt.trigger) != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, string(tt.trigger))
			}
		})
	}
}

func TestTriggerType_CanBeUsedAsMapKey(t *testing.T) {
	// Verify TriggerType can be used as a map key
	triggerMap := make(map[domain.TriggerType]string)
	triggerMap[domain.TriggerPreToolUse] = "before tool"
	triggerMap[domain.TriggerPostToolUse] = "after tool"

	if triggerMap[domain.TriggerPreToolUse] != "before tool" {
		t.Error("TriggerType cannot be used as map key")
	}
}

func TestTriggerType_CanBeCompared(t *testing.T) {
	// Verify TriggerType values can be compared
	trigger1 := domain.TriggerPreToolUse
	trigger2 := domain.TriggerPreToolUse
	trigger3 := domain.TriggerPostToolUse

	if trigger1 != trigger2 {
		t.Error("Same trigger types should be equal")
	}
	if trigger1 == trigger3 {
		t.Error("Different trigger types should not be equal")
	}
}

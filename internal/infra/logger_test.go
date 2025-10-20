package infra_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level infra.LogLevel
		want  string
	}{
		{name: "debug level", level: infra.LogLevelDebug, want: "DEBUG"},
		{name: "info level", level: infra.LogLevelInfo, want: "INFO"},
		{name: "warn level", level: infra.LogLevelWarn, want: "WARN"},
		{name: "error level", level: infra.LogLevelError, want: "ERROR"},
		{name: "unknown level", level: infra.LogLevel(99), want: "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("LogLevel.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelInfo)

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
}

func TestNewDefaultLogger(t *testing.T) {
	logger := infra.NewDefaultLogger()

	if logger == nil {
		t.Fatal("NewDefaultLogger returned nil")
	}
}

func TestNewDebugLogger(t *testing.T) {
	logger := infra.NewDebugLogger()

	if logger == nil {
		t.Fatal("NewDebugLogger returned nil")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelInfo)

	// Initially, debug should not log
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("Debug message should not be logged at INFO level")
	}

	// Set to debug level
	logger.SetLevel(infra.LogLevelDebug)

	// Now debug should log
	logger.Debug("debug message")
	if buf.Len() == 0 {
		t.Error("Debug message should be logged at DEBUG level")
	}

	// Verify the message contains expected content
	output := buf.String()
	if !strings.Contains(output, "DEBUG") || !strings.Contains(output, "debug message") {
		t.Errorf("Expected DEBUG log with 'debug message', got: %q", output)
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelDebug)

	logger.Debug("test debug %s", "message")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Error("Expected DEBUG level in output")
	}
	if !strings.Contains(output, "test debug message") {
		t.Error("Expected formatted message in output")
	}
	// Should have timestamp
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Expected timestamp brackets in output")
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelInfo)

	logger.Info("test info %d", 123)

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level in output")
	}
	if !strings.Contains(output, "test info 123") {
		t.Error("Expected formatted message in output")
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelWarn)

	logger.Warn("test warning %v", true)

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Error("Expected WARN level in output")
	}
	if !strings.Contains(output, "test warning true") {
		t.Error("Expected formatted message in output")
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelError)

	logger.Error("test error: %s", "failure")

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected ERROR level in output")
	}
	if !strings.Contains(output, "test error: failure") {
		t.Error("Expected formatted message in output")
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name       string
		minLevel   infra.LogLevel
		logFunc    func(*infra.Logger)
		shouldLog  bool
		levelLabel string
	}{
		{
			name:       "debug at debug level",
			minLevel:   infra.LogLevelDebug,
			logFunc:    func(l *infra.Logger) { l.Debug("msg") },
			shouldLog:  true,
			levelLabel: "DEBUG",
		},
		{
			name:       "debug at info level",
			minLevel:   infra.LogLevelInfo,
			logFunc:    func(l *infra.Logger) { l.Debug("msg") },
			shouldLog:  false,
			levelLabel: "DEBUG",
		},
		{
			name:       "info at info level",
			minLevel:   infra.LogLevelInfo,
			logFunc:    func(l *infra.Logger) { l.Info("msg") },
			shouldLog:  true,
			levelLabel: "INFO",
		},
		{
			name:       "info at warn level",
			minLevel:   infra.LogLevelWarn,
			logFunc:    func(l *infra.Logger) { l.Info("msg") },
			shouldLog:  false,
			levelLabel: "INFO",
		},
		{
			name:       "warn at warn level",
			minLevel:   infra.LogLevelWarn,
			logFunc:    func(l *infra.Logger) { l.Warn("msg") },
			shouldLog:  true,
			levelLabel: "WARN",
		},
		{
			name:       "warn at error level",
			minLevel:   infra.LogLevelError,
			logFunc:    func(l *infra.Logger) { l.Warn("msg") },
			shouldLog:  false,
			levelLabel: "WARN",
		},
		{
			name:       "error at error level",
			minLevel:   infra.LogLevelError,
			logFunc:    func(l *infra.Logger) { l.Error("msg") },
			shouldLog:  true,
			levelLabel: "ERROR",
		},
		{
			name:       "error at debug level",
			minLevel:   infra.LogLevelDebug,
			logFunc:    func(l *infra.Logger) { l.Error("msg") },
			shouldLog:  true,
			levelLabel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := infra.NewLogger(&buf, tt.minLevel)

			tt.logFunc(logger)

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, but got hasOutput=%v. Output: %q", tt.shouldLog, hasOutput, output)
			}

			if tt.shouldLog && !strings.Contains(output, tt.levelLabel) {
				t.Errorf("Expected output to contain %q, got: %q", tt.levelLabel, output)
			}
		})
	}
}

func TestLogger_MessageFormatting(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelInfo)

	// Test various formatting
	logger.Info("string: %s, int: %d, bool: %v", "test", 42, true)

	output := buf.String()
	if !strings.Contains(output, "string: test, int: 42, bool: true") {
		t.Errorf("Message not formatted correctly: %q", output)
	}
}

func TestLogger_OutputFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelInfo)

	logger.Info("test message")

	output := buf.String()

	// Should have timestamp [YYYY-MM-DD HH:MM:SS.mmm]
	if !strings.HasPrefix(output, "[") {
		t.Error("Expected output to start with timestamp bracket")
	}

	// Should have level label
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level label")
	}

	// Should have the message
	if !strings.Contains(output, "test message") {
		t.Error("Expected message content")
	}

	// Should have newline at end
	if !strings.HasSuffix(output, "\n") {
		t.Error("Expected newline at end of output")
	}
}

func TestLogger_ConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelDebug)

	// Test concurrent writes don't panic
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			logger.Debug("debug %d", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			logger.Info("info %d", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			logger.Error("error %d", i)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// Should have output from all three
	output := buf.String()
	if !strings.Contains(output, "DEBUG") || !strings.Contains(output, "INFO") || !strings.Contains(output, "ERROR") {
		t.Error("Expected output from all log levels")
	}
}

func TestLogger_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelInfo)

	logger.Info("")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Empty message should still produce log entry with timestamp and level")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level in output")
	}
}

func TestLogger_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	logger := infra.NewLogger(&buf, infra.LogLevelInfo)

	logger.Info("message with\nnewline and\ttab")

	output := buf.String()
	if !strings.Contains(output, "message with\nnewline and\ttab") {
		t.Error("Special characters should be preserved in output")
	}
}

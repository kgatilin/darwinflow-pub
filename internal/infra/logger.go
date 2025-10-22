package infra

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// LogLevelDebug for detailed debugging information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo for general informational messages
	LogLevelInfo
	// LogLevelWarn for warning messages
	LogLevelWarn
	// LogLevelError for error messages
	LogLevelError
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with levels
type Logger struct {
	mu       sync.Mutex
	output   io.Writer
	minLevel LogLevel
}

// NewLogger creates a new logger with the specified output and minimum level
func NewLogger(output io.Writer, minLevel LogLevel) *Logger {
	return &Logger{
		output:   output,
		minLevel: minLevel,
	}
}

// NewDefaultLogger creates a logger that writes to stderr with INFO level
func NewDefaultLogger() *Logger {
	return NewLogger(os.Stderr, LogLevelInfo)
}

// NewDebugLogger creates a logger that writes to stderr with DEBUG level
func NewDebugLogger() *Logger {
	return NewLogger(os.Stderr, LogLevelDebug)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// log writes a log message if the level is at or above the minimum level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)

	// Format: [TIMESTAMP] LEVEL: message
	fmt.Fprintf(l.output, "[%s] %-5s: %s\n", timestamp, level.String(), message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// ParseLogLevel converts a string log level to LogLevel
// Valid values: "debug", "info", "warn", "error", "off"
// Returns LogLevelError + true if invalid/off
func ParseLogLevel(level string) (LogLevel, bool) {
	switch level {
	case "debug":
		return LogLevelDebug, false
	case "info":
		return LogLevelInfo, false
	case "warn":
		return LogLevelWarn, false
	case "error":
		return LogLevelError, false
	case "off", "":
		// "off" means set to a level higher than Error so nothing logs
		return LogLevelError + 1, true
	default:
		// Invalid - default to error level
		return LogLevelError, false
	}
}

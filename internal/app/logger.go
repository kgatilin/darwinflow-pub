package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// EventMapper maps string event types to domain.EventType
type EventMapper struct{}

// MapEventType maps string event types to domain.EventType
func (m *EventMapper) MapEventType(eventTypeStr string) domain.EventType {
	// Normalize the string
	normalized := strings.ToLower(strings.ReplaceAll(eventTypeStr, "_", "."))

	switch normalized {
	case "chat.started":
		return domain.ChatStarted
	case "chat.ended", "chat.end":
		return domain.ChatStarted // Reuse for now
	case "chat.message.user", "user.message":
		return domain.ChatMessageUser
	case "chat.message.assistant", "assistant.message":
		return domain.ChatMessageAssistant
	case "tool.invoked", "tool.invoke":
		return domain.ToolInvoked
	case "tool.result":
		return domain.ToolResult
	case "file.read":
		return domain.FileRead
	case "file.written", "file.write":
		return domain.FileWritten
	case "context.changed", "context.change":
		return domain.ContextChanged
	case "error":
		return domain.Error
	default:
		// Default to generic event
		return domain.EventType(normalized)
	}
}

// TranscriptParser defines the interface for parsing Claude Code transcripts
type TranscriptParser interface {
	ExtractLastToolUse(transcriptPath string, maxParamLength int) (string, string, error)
	ExtractLastUserMessage(transcriptPath string) (string, error)
	ExtractLastAssistantMessage(transcriptPath string) (string, error)
}

// ContextDetector defines the interface for detecting context
type ContextDetector interface {
	DetectContext() string
}

// ContentNormalizer defines the interface for normalizing content for search
type ContentNormalizer func(eventType, payload string) string

// LoggerService orchestrates event logging for Claude Code interactions
type LoggerService struct {
	repository        domain.EventRepository
	transcriptParser  TranscriptParser
	contextDetector   ContextDetector
	contentNormalizer ContentNormalizer
	context           string
	sessionID         string
}

// NewLoggerService creates a new logger service
func NewLoggerService(
	repository domain.EventRepository,
	transcriptParser TranscriptParser,
	contextDetector ContextDetector,
	contentNormalizer ContentNormalizer,
) *LoggerService {
	return &LoggerService{
		repository:        repository,
		transcriptParser:  transcriptParser,
		contextDetector:   contextDetector,
		contentNormalizer: contentNormalizer,
		context:           contextDetector.DetectContext(),
	}
}

// LogEvent logs a domain event
func (s *LoggerService) LogEvent(ctx context.Context, eventType domain.EventType, payload interface{}) error {
	// Create normalized content for full-text search
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	content := s.contentNormalizer(string(eventType), string(payloadJSON))

	// Create domain event
	event := domain.NewEvent(eventType, s.sessionID, payload, content)

	// Save to repository
	if err := s.repository.Save(ctx, event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// Close closes the logger service and underlying repository
func (s *LoggerService) Close() error {
	return s.repository.Close()
}

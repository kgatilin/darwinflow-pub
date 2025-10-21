# Package: infra

**Path**: `internal/infra`

**Role**: Infrastructure implementations (databases, file I/O, external services)

---

## Quick Reference

- **Files**: 12
- **Exports**: 47
- **Dependencies**: `internal/domain` only
- **Layer**: Infrastructure (implements domain interfaces)

---

## Generated Documentation

### Exported API

#### Repository Implementations

**SQLiteEventRepository**:
- Implements `domain.EventRepository`, `domain.AnalysisRepository`, `domain.RawQueryExecutor`
- Methods: `Save`, `FindByQuery`, `GetAllSessionIDs`, `SaveAnalysis`, `GetAnalysisBySessionID`
- Full-text search support
- Event versioning and migration

#### Configuration

**ConfigLoader**:
- Loads/saves YAML configuration
- Methods: `LoadConfig`, `SaveConfig`, `InitializeDefaultConfig`, `GetPrompt`
- Handles prompt templates

#### Logging

**Logger**:
- Leveled logging (Debug, Info, Warn, Error)
- Log levels: `LogLevelDebug`, `LogLevelInfo`, `LogLevelWarn`, `LogLevelError`
- Thread-safe with mutex
- Constructors: `NewLogger`, `NewDefaultLogger`, `NewDebugLogger`

#### Utilities

**TranscriptParser**:
- Parse Claude Code transcripts
- Extract messages, tool uses
- Methods: `Parse`, `ExtractLastUserMessage`, `ExtractLastAssistantMessage`, `ExtractLastToolUse`

**ContextDetector**:
- Detect Git repository context
- Method: `DetectContext()`

**Helper Functions**:
- `NormalizeContent` - Content normalization
- `ValidateModelAlias` - Model validation

---

## Architectural Principles

### What MUST Be Here

✅ **Repository implementations** - SQLite, file-based, etc.
✅ **External service clients** - API clients, CLI wrappers
✅ **File I/O** - Config loaders, transcript parsers
✅ **Database code** - SQL, migrations, connection management
✅ **Infrastructure utilities** - Logging, context detection
✅ **Third-party integrations** - SQLite, YAML parsing

### What MUST NOT Be Here

❌ **Business logic** - Domain rules belong in `internal/domain`
❌ **Application orchestration** - Workflows belong in `internal/app`
❌ **Domain interfaces** - Define in `internal/domain`, implement here
❌ **Plugin code** - Belongs in `pkg/plugins/*`
❌ **UI code** - Belongs in `internal/app/tui`

### Critical Rules

1. **Implement, Don't Define**: Infrastructure implements domain interfaces
2. **Dependency Direction**: May import `internal/domain`, never vice versa
3. **Separation of Concerns**: Each infrastructure component is independent
4. **Error Handling**: Map infrastructure errors to domain errors
5. **Testability**: Use dependency injection, provide test doubles

---

## Repository Implementation Pattern

**Domain defines interface**:
```go
// internal/domain/repository.go
type EventRepository interface {
    Save(ctx context.Context, event *Event) error
}
```

**Infrastructure implements**:
```go
// internal/infra/sqlite_repository.go
type SQLiteEventRepository struct {
    db *sql.DB
}

func (r *SQLiteEventRepository) Save(ctx context.Context, event *domain.Event) error {
    // SQL implementation
}
```

**Application uses interface**:
```go
// internal/app/analysis.go
type AnalysisService struct {
    repo domain.EventRepository // Not *SQLiteEventRepository
}
```

---

## SQLite Repository

### Schema

**Events table**:
- `id` - UUID primary key
- `timestamp` - Event time
- `event_type` - Event type string
- `session_id` - Session identifier
- `payload` - JSON blob
- `content` - Full-text searchable content
- `version` - Schema version

**Analyses table**:
- `id` - UUID primary key
- `session_id` - Foreign key
- `analyzed_at` - Analysis timestamp
- `analysis_result` - Analysis text
- `model_used` - Model name
- `prompt_used` - Prompt template
- `patterns_summary` - Summary
- `analysis_type` - Type discriminator
- `prompt_name` - Named prompt identifier

### Indexing

- Session ID (frequent queries)
- Timestamp (time-range queries)
- Event type (filtering)
- Full-text index on content

### Migrations

Event versioning handled in:
- `Initialize()` - Creates schema
- Version detection in `Save()`
- Backward compatibility ensured

---

## Configuration Management

**File Format**: YAML (`.darwinflow/config.yaml`)

**Structure**:
```yaml
analysis:
  token_limit: 100000
  model: "claude-sonnet-4"
  enabled_prompts: ["default", "tools"]
  claude_options:
    allowed_tools: ["read", "edit"]

ui:
  default_output_dir: "./analyses"

prompts:
  default: "Analyze this session..."
  custom_prompt: "Custom analysis..."
```

**Thread Safety**: ConfigLoader is stateless (thread-safe)

---

## Logging

### Log Levels

1. **Debug** - Verbose diagnostic info
2. **Info** - Normal operation events
3. **Warn** - Warnings, recoverable errors
4. **Error** - Errors, failures

### Usage

```go
logger := infra.NewLogger(os.Stdout, infra.LogLevelInfo)
logger.Info("Processing event: %s", eventID)
logger.Error("Failed to save: %v", err)
```

**Thread Safety**: Mutex-protected writes

---

## Transcript Parsing

**Input**: Claude Code transcript JSON files

**Output**: Structured `TranscriptEntry` list

**Capabilities**:
- Parse user/assistant messages
- Extract tool invocations
- Extract tool results
- Handle errors

**Usage**:
```go
parser := infra.NewTranscriptParser()
entries, err := parser.Parse(transcriptData)
lastMsg, err := parser.ExtractLastUserMessage(transcriptData)
```

---

## Context Detection

**ContextDetector** identifies Git repository context:
- Detects `.git` directory
- Extracts repository name
- Used for event context tagging

---

## Error Mapping

Map infrastructure errors to domain errors:

```go
// Infrastructure error
if err == sql.ErrNoRows {
    return domain.ErrNotFound
}

// Database constraint violation
if isUniqueViolation(err) {
    return domain.ErrAlreadyExists
}
```

**Standard domain errors**: `ErrNotFound`, `ErrAlreadyExists`, `ErrInternal`, etc.

---

## Testing Strategy

**Black-box testing** (`package infra_test`):
- Test through public interfaces
- Use temporary databases (`t.TempDir()`)
- Test error conditions
- Test migrations

**Integration tests**:
- Real SQLite database
- File system I/O
- Full query lifecycle

---

## Files

- `sqlite_repository.go` - SQLite implementation of repositories
- `config.go` - YAML configuration loader
- `logger.go` - Leveled logger implementation
- `transcript.go` - Transcript parser
- `context.go` - Context detector
- `*_test.go` - Infrastructure tests
- `*_integration_test.go` - Integration tests

---

*Generated by `go-arch-lint -format=package internal/infra`*

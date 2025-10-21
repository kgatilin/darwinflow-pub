# Package: tui

**Path**: `internal/app/tui`

**Role**: Terminal User Interface - interactive session browser

---

## Quick Reference

- **Files**: 22
- **Exports**: 55
- **Dependencies**: `internal/app`, `internal/domain`, `pkg/pluginsdk`
- **Layer**: Presentation (user interface)
- **Framework**: Bubble Tea (Elm architecture)

---

## Generated Documentation

### Exported API

#### Main Entry Point

**Run()**:
- Launches TUI application
- Parameters: context, PluginRegistry, AnalysisService, LogsService, Config
- Returns error on failure

#### Models (Bubble Tea)

**AppModel**:
- Root application model
- Manages view state transitions
- Coordinates sub-models

**SessionListModel**:
- Browse sessions in list view
- Methods: `Init`, `Update`, `View`, `GetSelectedSession`, `UpdateSessions`

**SessionDetailModel**:
- View session details
- Shows analyses, metadata
- Methods: `Init`, `Update`, `View`

**AnalysisViewerModel**:
- Render analysis results
- Markdown rendering with Glamour
- Methods: `Init`, `Update`, `View`

**LogViewerModel**:
- View session event logs
- Methods: `Init`, `Update`, `View`

#### Messages (Bubble Tea)

**Navigation**:
- `SelectedSessionMsg` - Session selected from list
- `BackToListMsg` - Return to session list
- `BackToDetailMsg` - Return to detail view

**Actions**:
- `AnalyzeSessionMsg` - Request analysis
- `ReanalyzeSessionMsg` - Re-run analysis
- `ViewAnalysisMsg` - View existing analysis
- `ViewLogMsg` - View session logs
- `SaveToMarkdownMsg` - Export analysis
- `RefreshRequestMsg` - Refresh data

**Results**:
- `SessionsLoadedMsg` - Sessions loaded (with error handling)
- `AnalysisCompleteMsg` - Analysis finished (with error handling)
- `SaveCompleteMsg` - Export finished (with error handling)
- `ErrorMsg` - Generic error

#### View States

**Constants**:
- `ViewSessionList` - Session browser
- `ViewSessionDetail` - Session details
- `ViewAnalysisViewer` - Analysis display
- `ViewLogViewer` - Event log display
- `ViewAnalysisAction` - Analysis action menu
- `ViewProgress` - Loading/progress
- `ViewSaveDialog` - Save dialog

#### Data Types

**SessionInfo**:
- Session metadata
- Properties: SessionID, ShortID, FirstEvent, LastEvent, EventCount, AnalysisCount, Analyses, TokenCount

**SessionItem**:
- List item wrapper
- Methods: `Title`, `Description`, `FilterValue`

#### Utilities

**FormatTokenCount()**:
- Format token counts (e.g., "1.2K", "15K")

**Constructors**:
- `NewAppModel`, `NewSessionListModel`, `NewSessionDetailModel`, `NewAnalysisViewerModel`, `NewLogViewerModel`

---

## Architectural Principles

### What MUST Be Here

✅ **UI components** - Bubble Tea models and views
✅ **User interactions** - Key bindings, navigation
✅ **View state management** - Screen transitions
✅ **Rendering logic** - Layout, styling, formatting
✅ **Message handling** - Bubble Tea message types
✅ **Display formatting** - Pretty-printing for terminal

### What MUST NOT Be Here

❌ **Business logic** - Belongs in `internal/app` services
❌ **Data access** - Use services, not repositories directly
❌ **Domain models** - Use `internal/domain` types
❌ **Infrastructure code** - No DB, file I/O (use services)
❌ **Heavy computation** - Offload to services

### Critical Rules

1. **Thin UI Layer**: Delegate logic to services
2. **Bubble Tea Patterns**: Follow Elm architecture (Model, Update, View)
3. **Async Operations**: Use Bubble Tea commands for I/O
4. **Error Handling**: Show errors gracefully in UI
5. **Accessibility**: Clear navigation, keyboard-driven

---

## Bubble Tea Architecture

### The Elm Architecture

**Model**: Application state
```go
type AppModel struct {
    viewState    ViewState
    sessionList  SessionListModel
    sessionDetail SessionDetailModel
    // ...
}
```

**Update**: State transitions
```go
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case SelectedSessionMsg:
        // Transition to detail view
    case tea.KeyMsg:
        // Handle key press
    }
}
```

**View**: Rendering
```go
func (m AppModel) View() string {
    switch m.viewState {
    case ViewSessionList:
        return m.sessionList.View()
    case ViewSessionDetail:
        return m.sessionDetail.View()
    }
}
```

### Command Pattern

Async operations return `tea.Cmd`:
```go
func loadSessionsCmd(service *app.AnalysisService) tea.Cmd {
    return func() tea.Msg {
        sessions, err := service.GetAllSessionIDs(ctx, 100)
        return SessionsLoadedMsg{Sessions: sessions, Error: err}
    }
}
```

---

## View State Machine

```
ViewSessionList
    ↓ [Enter] Select session
ViewSessionDetail
    ↓ [a] Analyze
    ↓ [v] View analysis
    ↓ [l] View logs
ViewAnalysisViewer
    ↓ [s] Save
    ↓ [Esc] Back
ViewLogViewer
    ↓ [Esc] Back
ViewSaveDialog
    ↓ [Enter] Confirm
    ↓ [Esc] Cancel
```

**Transitions handled in AppModel.Update()**

---

## Session List View

### Features

- **List display**: Bubble Tea list component
- **Filtering**: Search sessions by ID
- **Sorting**: By timestamp (newest first)
- **Pagination**: Navigate large session lists
- **Status indicators**: Analysis status, token count

### Key Bindings

- `↑/↓` - Navigate
- `Enter` - Select session
- `/` - Filter
- `q` - Quit

### Data Loading

```go
// Initial load
cmd := loadSessionsCmd(service)

// Update on message
case SessionsLoadedMsg:
    m.UpdateSessions(msg.Sessions)
```

---

## Session Detail View

### Sections

1. **Metadata**:
   - Session ID (full + short)
   - Time range (first → last event)
   - Event count
   - Token count

2. **Analyses**:
   - List of existing analyses
   - Prompt names
   - Analysis types
   - Timestamps

3. **Actions**:
   - Analyze session
   - View analysis
   - View logs
   - Re-analyze

### Key Bindings

- `a` - Analyze
- `v` - View analysis
- `l` - View logs
- `r` - Re-analyze
- `Esc` - Back to list

---

## Analysis Viewer

### Features

- **Markdown rendering**: Glamour library
- **Syntax highlighting**: Code blocks
- **Scrolling**: Viewport component
- **Export**: Save to markdown file

### Rendering

```go
renderer, _ := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(width),
)
rendered, _ := renderer.Render(analysis.AnalysisResult)
```

### Key Bindings

- `↑/↓` - Scroll
- `s` - Save to file
- `Esc` - Back

---

## Log Viewer

### Features

- **Event list**: Session events
- **Syntax highlighting**: JSON payloads
- **Scrolling**: Full event history
- **Filtering**: By event type (future)

### Display Format

```
[1] 2024-10-21 14:30:45 | ToolInvoked
Session: abc123...
Payload: {"tool": "Read", "file": "main.go"}
---
```

### Key Bindings

- `↑/↓` - Scroll
- `Esc` - Back

---

## Styling

**Lipgloss** for consistent styling:
- Colors: Status indicators, headers
- Borders: Section separation
- Padding/margins: Layout spacing
- Alignment: Text positioning

**Example**:
```go
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("86"))
```

---

## Error Handling

### Display Errors

```go
case ErrorMsg:
    return m, tea.Println(
        errorStyle.Render("Error: " + msg.Error.Error()),
    )
```

### Recoverable Errors

- Show error message
- Allow retry
- Don't crash app

### Fatal Errors

- Display error
- Provide exit option
- Log to stderr

---

## Async Operations

All I/O operations are async:

1. **Load sessions**: Background fetch
2. **Analyze session**: Long-running LLM call
3. **Save markdown**: File write
4. **Refresh data**: Database query

**Pattern**:
```go
// Start operation
return m, analyzeSessionCmd(sessionID, service)

// Handle result
case AnalysisCompleteMsg:
    if msg.Error != nil {
        // Show error
    } else {
        // Update view
    }
```

---

## Testing Strategy

**Model testing**:
- Test state transitions
- Test message handling
- Mock services
- Verify view states

**View testing**:
- Snapshot tests (visual regression)
- Check rendering logic
- Verify layout

**Integration testing**:
- Full flow testing
- Keyboard navigation
- Error scenarios

---

## Performance Considerations

- **Lazy loading**: Load sessions on demand
- **Pagination**: Limit displayed items
- **Caching**: Cache rendered markdown
- **Debouncing**: Filter input debouncing
- **Background ops**: Don't block UI thread

---

## Files

- `app.go` - AppModel (root model)
- `sessionlist.go` - SessionListModel
- `sessiondetail.go` - SessionDetailModel
- `analysisviewer.go` - AnalysisViewerModel
- `logviewer.go` - LogViewerModel
- `types.go` - Shared types and messages
- `*_test.go` - Component tests

---

*Generated by `go-arch-lint -format=package internal/app/tui`*

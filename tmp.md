Document: TM-doc-1763125287883967505
  Title:       Iteration 28: Presentation Layer MVP Architecture
  Type:        plan
  Status:      draft
  Created:     2025-11-14 13:01:27
  Updated:     2025-11-14 14:13:23
  Attachment:  Iteration 28

Content:
# Iteration 28: Presentation Layer MVP Architecture

**Scope**: Phases 4-6 (ViewModels, Queries, Presenters)
**Duration**: 12-16 days
**Strategy**: Parallel development - old `tui` unchanged, new `tui-new` command for refactored UI

---

## Directory Structure

```
pkg/plugins/task_manager/presentation/
├── cli/
│   ├── tui_models.go              # OLD UI - DO NOT MODIFY
│   ├── command_tui.go             # OLD command: 'dw task-manager tui'
│   └── [other CLI files...]       # Existing CLI adapters
│
└── tui/                           # NEW DIRECTORY - Create this
    ├── viewmodels/                # Phase 4 - Pure data (zero imports)
    │   ├── roadmap_list_vm.go     # 6 structs: RoadmapListViewModel, IterationCardViewModel, TrackCardViewModel, BacklogTaskViewModel, TaskSummaryViewModel
    │   ├── track_detail_vm.go     # 5 structs: TrackDetailViewModel, TrackDependencyViewModel, TaskRowViewModel, ADRSummaryViewModel, ACSummaryViewModel
    │   ├── task_detail_vm.go      # 4 structs: TaskDetailViewModel, TrackInfoViewModel, IterationMembershipViewModel, ACDetailViewModel
    │   ├── iteration_list_vm.go   # 1 struct: IterationListViewModel
    │   ├── iteration_detail_vm.go # 4 structs: IterationDetailViewModel, ProgressViewModel, GroupedTasksViewModel, IterationACViewModel
    │   ├── adr_list_vm.go         # 2 structs: ADRListViewModel, ADRRowViewModel
    │   ├── ac_list_vm.go          # 2 structs: ACListViewModel, ACRowViewModel
    │   ├── ac_detail_vm.go        # 2 structs: ACDetailFullViewModel, TaskInfoViewModel
    │   ├── ac_fail_input_vm.go    # 1 struct: ACFailInputViewModel
    │   ├── error_vm.go            # 1 struct: ErrorViewModel
    │   └── loading_vm.go          # 1 struct: LoadingViewModel
    │
    ├── transformers/              # Phase 4 - Entity→ViewModel (pure functions)
    │   ├── formatting_helpers.go  # RenderStatusBadge(), RenderProgressBar(), FormatDate(), WrapText(), ComputeTaskSummary()
    │   ├── roadmap_transformer.go # TransformRoadmapToListViewModel(roadmap, iterations, tracks, backlog, ...)
    │   ├── track_transformer.go   # TransformTrackToCardViewModel(), TransformTrackToDetailViewModel()
    │   ├── task_transformer.go    # TransformTaskToRowViewModel(), TransformTaskToDetailViewModel()
    │   ├── iteration_transformer.go # TransformIterationToCardViewModel(), TransformIterationToDetailViewModel()
    │   ├── ac_transformer.go      # TransformACToDetailViewModel(), TransformACToRowViewModel()
    │   └── adr_transformer.go     # TransformADRToRowViewModel()
    │
    ├── queries/                   # Phase 5 - View data composition
    │   ├── roadmap_queries.go     # GetRoadmapViewData() → RoadmapListViewModel
    │   ├── track_queries.go       # GetTrackDetailViewData(trackID) → TrackDetailViewModel
    │   ├── iteration_queries.go   # GetIterationDetailViewData(iterNum) → IterationDetailViewModel
    │   └── ac_queries.go          # GetACListViewData(trackID) → ACListViewModel
    │
    ├── presenters/                # Phase 6 - MVP presenters (business logic)
    │   ├── presenter.go           # Base interface: Init(), Update(tea.Msg), View(), SetSize(w, h)
    │   ├── roadmap_list.go        # RoadmapListPresenter
    │   ├── track_detail.go        # TrackDetailPresenter
    │   ├── task_detail.go         # TaskDetailPresenter
    │   ├── iteration_list.go      # IterationListPresenter
    │   ├── iteration_detail.go    # IterationDetailPresenter (dual focus: tasks/ACs)
    │   ├── adr_list.go            # ADRListPresenter
    │   ├── ac_list.go             # ACListPresenter
    │   ├── ac_detail.go           # ACDetailPresenter
    │   ├── ac_fail_input.go       # ACFailInputPresenter
    │   ├── error.go               # ErrorPresenter
    │   └── loading.go             # LoadingPresenter
    │
    └── command_tui_new.go         # NEW command: 'dw task-manager tui-new'
```

**Total Files**: 11 ViewModels + 7 Transformers + 4 Queries + 12 Presenters + 1 Command = **35 new files**

---

## Phase 4: ViewModels & Transformers

### ViewModels (11 files, 33 structs)

**Location**: `pkg/plugins/task_manager/presentation/tui/viewmodels/`

**Rules**:
- ✅ Pure data structures (NO imports except stdlib)
- ✅ All strings pre-formatted (StatusBadge, PriorityIcon, DisplayText)
- ✅ All aggregations pre-computed (counts, percentages)
- ✅ All foreign keys resolved (TrackTitle in BacklogTaskViewModel)
- ❌ NO entity references
- ❌ NO lazy evaluation
- ❌ NO business logic

**Core ViewModels**:

```go
// roadmap_list_vm.go (6 structs)
type RoadmapListViewModel struct {
    Vision              string
    SuccessCriteria     string
    ActiveIterations    []IterationCardViewModel
    CompletedIterations []IterationCardViewModel
    ActiveTracks        []TrackCardViewModel
    CompletedTracks     []TrackCardViewModel
    BacklogTasks        []BacklogTaskViewModel
    ShowFullRoadmap     bool
}

type IterationCardViewModel struct {
    Number          int
    Name            string
    StatusBadge     string    // "● Current", "○ Planned", "✓ Complete"
    TaskCount       int
    TaskSummary     string    // "3 todo, 2 in progress, 5 done"
    ProgressPercent float64
    Tasks           []TaskSummaryViewModel
}

type TrackCardViewModel struct {
    ID           string
    Title        string
    StatusBadge  string    // "○ Not Started", "⊙ In Progress", "☑ Complete"
    PriorityIcon string    // "🔴 High", "🟡 Medium", "🟢 Low"
    TaskCount    int
    TaskSummary  string
    ACProgress   string    // "15/20 verified"
    Tasks        []TaskSummaryViewModel
}

type BacklogTaskViewModel struct {
    ID          string
    Title       string
    TrackTitle  string    // Resolved from TrackID
    TrackID     string
    StatusIcon  string
}

type TaskSummaryViewModel struct {
    ID           string
    Title        string
    StatusIcon   string
    PriorityIcon string
}

// track_detail_vm.go (5 structs)
type TrackDetailViewModel struct {
    ID           string
    Title        string
    Description  string    // Pre-wrapped to terminal width
    StatusBadge  string
    PriorityIcon string
    Dependencies []TrackDependencyViewModel
    Tasks        []TaskRowViewModel
    ADRSummary   ADRSummaryViewModel
    ACSummary    ACSummaryViewModel
}

type TrackDependencyViewModel struct {
    ID    string
    Title string
}

type TaskRowViewModel struct {
    ID           string
    Title        string
    StatusIcon   string
    PriorityIcon string
    Branch       string
}

type ADRSummaryViewModel struct {
    Total         int
    AcceptedCount int
    ProposedCount int
    DisplayText   string    // "3 ADRs (2 accepted, 1 proposed)"
}

type ACSummaryViewModel struct {
    Total              int
    VerifiedCount      int
    PendingReviewCount int
    FailedCount        int
    DisplayText        string    // "20 ACs (15 verified, 3 pending, 2 failed)"
}

// task_detail_vm.go (4 structs)
type TaskDetailViewModel struct {
    ID                  string
    Title               string
    Description         string
    StatusBadge         string
    PriorityIcon        string
    Branch              string
    TrackInfo           TrackInfoViewModel
    IterationMembership []IterationMembershipViewModel
    AcceptanceCriteria  []ACDetailViewModel
}

type TrackInfoViewModel struct {
    ID    string
    Title string
}

type IterationMembershipViewModel struct {
    Number int
    Name   string
    Status string
}

type ACDetailViewModel struct {
    ID                      string
    Description             string
    StatusIcon              string
    VerificationType        string
    TestingInstructions     string
    TestingInstructionsIcon string    // "📋" if has instructions
    Notes                   string    // Failure feedback
}

// iteration_list_vm.go (1 struct)
type IterationListViewModel struct {
    Iterations []IterationCardViewModel
}

// iteration_detail_vm.go (4 structs)
type IterationDetailViewModel struct {
    Number         int
    Name           string
    Goal           string
    StatusBadge    string
    Deliverable    string
    StartedAt      string
    CompletedAt    string
    Progress       ProgressViewModel
    TasksGrouped   GroupedTasksViewModel
    ACs            []IterationACViewModel
    FocusedSection string    // "tasks" or "acs"
}

type ProgressViewModel struct {
    TotalCount  int
    DoneCount   int
    Percent     float64
    DisplayText string    // "5/10 (50%)"
    BarRendered string    // Pre-rendered ASCII progress bar
}

type GroupedTasksViewModel struct {
    Todo       []TaskRowViewModel
    InProgress []TaskRowViewModel
    Done       []TaskRowViewModel
}

type IterationACViewModel struct {
    ID                      string
    TaskID                  string
    TaskTitle               string
    Description             string
    StatusIcon              string
    VerificationType        string
    TestingInstructions     string
    TestingInstructionsIcon string
    ExpandedInstructions    bool
    Notes                   string
    NotesRendered           string    // Pre-rendered with indentation
}

// adr_list_vm.go (2 structs)
type ADRListViewModel struct {
    TrackID    string
    TrackTitle string
    ADRs       []ADRRowViewModel
}

type ADRRowViewModel struct {
    ID          string
    Title       string
    StatusBadge string    // "✓ Accepted", "? Proposed", "⚠ Deprecated"
}

// ac_list_vm.go (2 structs)
type ACListViewModel struct {
    TrackID    string
    TrackTitle string
    ACs        []ACRowViewModel
}

type ACRowViewModel struct {
    ID               string
    Description      string    // Truncated to 80 chars
    StatusIcon       string
    VerificationType string
    TaskTitle        string
}

// ac_detail_vm.go (2 structs)
type ACDetailFullViewModel struct {
    ID                  string
    Description         string
    StatusIcon          string
    VerificationType    string
    TestingInstructions string
    Notes               string
    TaskInfo            TaskInfoViewModel
}

// ac_fail_input_vm.go (1 struct)
type ACFailInputViewModel struct {
    ACID         string
    Description  string
    TaskTitle    string
    CurrentNotes string
}

// error_vm.go (1 struct)
type ErrorViewModel struct {
    Message string
    Context string
}

// loading_vm.go (1 struct)
type LoadingViewModel struct {
    Message string
}
```

### Transformers (7 files)

**Location**: `pkg/plugins/task_manager/presentation/tui/transformers/`

**Rules**:
- ✅ Pure functions (Entity → ViewModel)
- ✅ Import: `domain/entities/`, `presentation/tui/viewmodels/`
- ✅ All formatting logic here (status badges, icons, progress bars)
- ❌ NO repository calls
- ❌ NO side effects

**Core Functions**:

```go
// formatting_helpers.go
RenderStatusBadge(status, entityType string) string
RenderPriorityIcon(rank int) string
RenderTaskStatusIcon(status string) string
RenderProgressBar(done, total, width int) string
FormatDate(t time.Time) string
WrapText(text string, width int) string
ComputeTaskSummary(tasks []*TaskEntity) string

// roadmap_transformer.go
TransformRoadmapToListViewModel(
    roadmap *RoadmapEntity,
    iterations []*IterationEntity,
    tracks []*TrackEntity,
    backlog []*TaskEntity,
    iterationTasks map[int][]*TaskEntity,
    trackTasks map[string][]*TaskEntity,
    trackInfo map[string]*TrackEntity,
) *RoadmapListViewModel

// track_transformer.go
TransformTrackToCardViewModel(track *TrackEntity, tasks []*TaskEntity) *TrackCardViewModel
TransformTrackToDetailViewModel(
    track *TrackEntity,
    tasks []*TaskEntity,
    dependencies []*TrackEntity,
    adrSummary ADRSummaryData,
    acSummary ACSummaryData,
    terminalWidth int,
) *TrackDetailViewModel

// task_transformer.go
TransformTaskToRowViewModel(task *TaskEntity) *TaskRowViewModel
TransformTaskToDetailViewModel(
    task *TaskEntity,
    track *TrackEntity,
    iterations []*IterationEntity,
    acs []*AcceptanceCriteriaEntity,
    terminalWidth int,
) *TaskDetailViewModel

// iteration_transformer.go
TransformIterationToCardViewModel(iteration *IterationEntity, tasks []*TaskEntity) *IterationCardViewModel
TransformIterationToDetailViewModel(
    iteration *IterationEntity,
    tasks []*TaskEntity,
    acs []*AcceptanceCriteriaEntity,
    terminalWidth int,
) *IterationDetailViewModel

// ac_transformer.go
TransformACToDetailViewModel(ac *AcceptanceCriteriaEntity) *ACDetailViewModel
TransformACToRowViewModel(ac *AcceptanceCriteriaEntity, task *TaskEntity) *ACRowViewModel

// adr_transformer.go
TransformADRToRowViewModel(adr *ADREntity) *ADRRowViewModel
```

**Coverage Target**: 90%+ (pure functions, easy to test)

---

## Phase 5: Query Services

### Purpose

Query services **compose entities from repository** and **transform to ViewModels**. They encapsulate the "fetch + transform" pattern, keeping presenters clean.

### Query Services (4 files)

**Location**: `pkg/plugins/task_manager/presentation/tui/queries/`

**Rules**:
- ✅ Use **existing** repository methods (GetRoadmap, ListIterations, etc.)
- ✅ Compose entities in memory (multiple repo calls per view is fine)
- ✅ Transform entities to ViewModels using transformers
- ✅ Handle errors
- ❌ NO new repository methods
- ❌ NO SQL optimization
- ❌ NO business logic

**Core Functions**:

```go
// roadmap_queries.go
type RoadmapQueries struct {
    repo        RoadmapRepository
    transformer *RoadmapTransformer
}

func NewRoadmapQueries(repo RoadmapRepository) *RoadmapQueries {
    return &RoadmapQueries{
        repo:        repo,
        transformer: NewRoadmapTransformer(),
    }
}

func (q *RoadmapQueries) GetRoadmapViewData(ctx context.Context) (*RoadmapListViewModel, error) {
    // Fetch entities using existing repo methods
    roadmap, err := q.repo.GetRoadmap(ctx)
    if err != nil {
        return nil, err
    }

    iterations, err := q.repo.ListIterations(ctx, nil)
    if err != nil {
        return nil, err
    }

    tracks, err := q.repo.ListTracks(ctx, nil)
    if err != nil {
        return nil, err
    }

    backlog, err := q.repo.GetBacklogTasks(ctx)
    if err != nil {
        return nil, err
    }

    // Build trackInfo map for backlog tasks
    trackInfo := make(map[string]*TrackEntity)
    for _, task := range backlog {
        if _, exists := trackInfo[task.TrackID]; !exists {
            track, _ := q.repo.GetTrack(ctx, task.TrackID)
            if track != nil {
                trackInfo[task.TrackID] = track
            }
        }
    }

    // Transform to ViewModel
    return q.transformer.TransformRoadmapToListViewModel(
        roadmap, iterations, tracks, backlog, nil, nil, trackInfo,
    ), nil
}

// track_queries.go
type TrackQueries struct {
    repo        RoadmapRepository
    transformer *TrackTransformer
}

func (q *TrackQueries) GetTrackDetailViewData(ctx context.Context, trackID string, terminalWidth int) (*TrackDetailViewModel, error) {
    // Fetch track
    track, err := q.repo.GetTrack(ctx, trackID)
    if err != nil {
        return nil, err
    }

    // Fetch tasks
    tasks, err := q.repo.ListTasks(ctx, &TaskFilter{TrackID: &trackID})
    if err != nil {
        return nil, err
    }

    // Fetch dependencies
    dependencies := make([]*TrackEntity, 0, len(track.DependsOn))
    for _, depID := range track.DependsOn {
        dep, _ := q.repo.GetTrack(ctx, depID)
        if dep != nil {
            dependencies = append(dependencies, dep)
        }
    }

    // Fetch ADR summary
    adrs, _ := q.repo.ListADR(ctx, trackID)
    adrSummary := computeADRSummary(adrs)

    // Fetch AC summary
    acs := make([]*AcceptanceCriteriaEntity, 0)
    for _, task := range tasks {
        taskACs, _ := q.repo.ListAC(ctx, task.ID)
        acs = append(acs, taskACs...)
    }
    acSummary := computeACSummary(acs)

    // Transform
    return q.transformer.TransformTrackToDetailViewModel(
        track, tasks, dependencies, adrSummary, acSummary, terminalWidth,
    ), nil
}

// iteration_queries.go
type IterationQueries struct {
    repo        RoadmapRepository
    transformer *IterationTransformer
}

func (q *IterationQueries) GetIterationDetailViewData(ctx context.Context, iterNum int, terminalWidth int) (*IterationDetailViewModel, error) {
    // Fetch iteration
    iteration, err := q.repo.GetIteration(ctx, iterNum)
    if err != nil {
        return nil, err
    }

    // Fetch tasks
    tasks, err := q.repo.GetIterationTasks(ctx, iterNum)
    if err != nil {
        return nil, err
    }

    // Fetch ACs for all tasks
    acs := make([]*AcceptanceCriteriaEntity, 0)
    for _, task := range tasks {
        taskACs, _ := q.repo.ListAC(ctx, task.ID)
        acs = append(acs, taskACs...)
    }

    // Transform
    return q.transformer.TransformIterationToDetailViewModel(
        iteration, tasks, acs, terminalWidth,
    ), nil
}

// ac_queries.go
type ACQueries struct {
    repo        RoadmapRepository
    transformer *ACTransformer
}

func (q *ACQueries) GetACListViewData(ctx context.Context, trackID string) (*ACListViewModel, error) {
    // Fetch track
    track, err := q.repo.GetTrack(ctx, trackID)
    if err != nil {
        return nil, err
    }

    // Fetch all tasks for track
    filter := &TaskFilter{TrackID: &trackID}
    tasks, err := q.repo.ListTasks(ctx, filter)
    if err != nil {
        return nil, err
    }

    // Fetch all ACs
    allACs := make([]*AcceptanceCriteriaEntity, 0)
    taskMap := make(map[string]*TaskEntity)
    for _, task := range tasks {
        taskACs, _ := q.repo.ListAC(ctx, task.ID)
        allACs = append(allACs, taskACs...)
        taskMap[task.ID] = task
    }

    // Transform
    acRows := make([]ACRowViewModel, len(allACs))
    for i, ac := range allACs {
        task := taskMap[ac.TaskID]
        acRows[i] = q.transformer.TransformACToRowViewModel(ac, task)
    }

    return &ACListViewModel{
        TrackID:    trackID,
        TrackTitle: track.Title,
        ACs:        acRows,
    }, nil
}
```

**Coverage Target**: 80%+ (with mocked repositories)

---

## Phase 6: Presenters (MVP Pattern)

### Base Interface

**Location**: `pkg/plugins/task_manager/presentation/tui/presenters/presenter.go`

```go
type Presenter interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Presenter, tea.Cmd)
    View() string
    SetSize(width, height int)
}
```

### Presenter Structure (Example: RoadmapListPresenter)

**Location**: `pkg/plugins/task_manager/presentation/tui/presenters/roadmap_list.go`

```go
type RoadmapListPresenter struct {
    // Dependencies (injected)
    ctx     context.Context
    queries *RoadmapQueries
    logger  pluginsdk.Logger

    // View state (ViewModel only!)
    viewModel *RoadmapListViewModel

    // UI state (not domain)
    width   int
    height  int

    // Bubbles components
    viewport viewport.Model

    // Selection state
    selectedSection string  // "iterations", "tracks", "backlog"
    selectedIdx     int
}

func NewRoadmapListPresenter(ctx context.Context, queries *RoadmapQueries, logger pluginsdk.Logger) *RoadmapListPresenter

func (p *RoadmapListPresenter) Init() tea.Cmd {
    return func() tea.Msg {
        vm, err := p.queries.GetRoadmapViewData(p.ctx)
        return RoadmapDataLoadedMsg{ViewModel: vm, Error: err}
    }
}

func (p *RoadmapListPresenter) Update(msg tea.Msg) (Presenter, tea.Cmd) {
    switch msg := msg.(type) {
    case RoadmapDataLoadedMsg:
        p.viewModel = msg.ViewModel
        return p, nil
    case tea.KeyMsg:
        return p.handleKeys(msg)
    }
    return p, nil
}

func (p *RoadmapListPresenter) View() string {
    if p.viewModel == nil {
        return "Loading roadmap..."
    }
    return p.renderRoadmap(p.viewModel)
}

func (p *RoadmapListPresenter) SetSize(w, h int) {
    p.width = w
    p.height = h
    p.viewport.Width = w
    p.viewport.Height = h - 5
}

func (p *RoadmapListPresenter) handleKeys(msg tea.KeyMsg) (Presenter, tea.Cmd) {
    // Navigation logic
}

func (p *RoadmapListPresenter) renderRoadmap(vm *RoadmapListViewModel) string {
    // Rendering using ViewModel
}
```

### All Presenters (12 files)

1. **presenter.go** - Base interface
2. **roadmap_list.go** - Main dashboard (3-section navigation)
3. **track_detail.go** - Track with tasks, dependencies, ADR/AC summaries
4. **task_detail.go** - Task with ACs, iteration membership
5. **iteration_list.go** - All iterations list
6. **iteration_detail.go** - Iteration with grouped tasks + ACs (dual focus)
7. **adr_list.go** - ADR list for track
8. **ac_list.go** - AC list for track
9. **ac_detail.go** - AC detail with scrollable testing instructions
10. **ac_fail_input.go** - AC failure feedback form
11. **error.go** - Error display
12. **loading.go** - Loading spinner

**Rules**:
- ✅ Call queries layer
- ✅ Store ViewModels (NOT entities)
- ✅ Testable without Bubble Tea
- ✅ Handle navigation events
- ❌ NO entity references in presenter state
- ❌ NO inline transformations (use queries)

**Coverage Target**: 70%+

---

## New Command: `tui-new`

**Location**: `pkg/plugins/task_manager/presentation/cli/command_tui_new.go`

```go
func CommandTUINew(ctx context.Context, repo RoadmapRepository, logger pluginsdk.Logger) error {
    // Initialize queries
    roadmapQueries := queries.NewRoadmapQueries(repo)

    // Initialize presenter
    presenter := presenters.NewRoadmapListPresenter(ctx, roadmapQueries, logger)

    // Launch Bubble Tea program
    program := tea.NewProgram(newAppModel(presenter))
    _, err := program.Run()
    return err
}
```

**CLI Registration** (in `plugin.go`):

```go
&cli.Command{
    Name:  "tui-new",
    Usage: "Launch TUI (refactored MVP architecture)",
    Action: func(c *cli.Context) error {
        return CommandTUINew(ctx, repo, logger)
    },
}
```

---

## Dependency Graph

```
ViewModels (pure data)
    ↑
Transformers (Entity → ViewModel)
    ↑
Queries (fetch from repo + transform)
    ↑
Presenters (queries → UI logic)
    ↑
Views/Rendering (ViewModels → String)
```

**Import Rules**:
- `viewmodels/` → ZERO imports (pure data)
- `transformers/` → `domain/entities/`, `viewmodels/`
- `queries/` → `domain/repositories/`, `domain/entities/`, `viewmodels/`, `transformers/`
- `presenters/` → `viewmodels/`, `queries/`
- `presenters/` → ❌ NO `domain/entities/`

---

## Testing Strategy

### Phase 4 Tests (90%+ coverage)

**Transformer Tests**:
```go
func TestTransformTrackToCardViewModel(t *testing.T) {
    tests := []struct {
        name     string
        track    *TrackEntity
        tasks    []*TaskEntity
        expected *TrackCardViewModel
    }{
        {
            name: "track with mixed tasks",
            track: &TrackEntity{ID: "TM-track-1", Title: "Test", Status: "in-progress", Rank: 100},
            tasks: []*TaskEntity{
                {Status: "todo"},
                {Status: "in-progress"},
                {Status: "done"},
            },
            expected: &TrackCardViewModel{
                ID:           "TM-track-1",
                Title:        "Test",
                StatusBadge:  "⊙ In Progress",
                TaskCount:    3,
                TaskSummary:  "1 todo, 1 in progress, 1 done",
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := TransformTrackToCardViewModel(tt.track, tt.tasks)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Phase 5 Tests (80%+ coverage)

**Query Tests with Mocked Repository**:
```go
func TestGetRoadmapViewData(t *testing.T) {
    mockRepo := &MockRepository{
        roadmap:    &RoadmapEntity{...},
        iterations: []*IterationEntity{...},
        tracks:     []*TrackEntity{...},
        backlog:    []*TaskEntity{...},
    }
    queries := NewRoadmapQueries(mockRepo)

    vm, err := queries.GetRoadmapViewData(context.Background())

    assert.NoError(t, err)
    assert.NotNil(t, vm)
    assert.Equal(t, 3, len(vm.ActiveIterations))
    assert.Equal(t, "⊙ In Progress", vm.ActiveTracks[0].StatusBadge)
}
```

### Phase 6 Tests (70%+ coverage)

**Presenter Tests without Bubble Tea**:
```go
func TestRoadmapListPresenter_Update(t *testing.T) {
    mockQueries := &MockRoadmapQueries{
        viewModel: &RoadmapListViewModel{...},
    }
    presenter := NewRoadmapListPresenter(ctx, mockQueries, logger)

    // Test data loading
    msg := RoadmapDataLoadedMsg{ViewModel: mockQueries.viewModel}
    updated, cmd := presenter.Update(msg)

    assert.NotNil(t, updated.(*RoadmapListPresenter).viewModel)
    assert.Nil(t, cmd)
}
```

---

## Acceptance Criteria

### Phase 4 (TM-task-129)

- [ ] 11 ViewModel files created (33 structs total)
- [ ] 7 Transformer files created
- [ ] All ViewModels have ZERO imports (pure data)
- [ ] All formatters in formatting_helpers.go (RenderStatusBadge, etc.)
- [ ] 90%+ test coverage on transformers
- [ ] go-arch-lint passes (viewmodels/ has no external imports)

### Phase 5 (TM-task-130)

- [ ] 4 Query service files created
- [ ] Queries use **existing** repository methods (no new methods added)
- [ ] Queries compose entities in memory
- [ ] Queries use transformers to create ViewModels
- [ ] 80%+ test coverage on queries (with mocked repo)

### Phase 6 (TM-task-131)

- [ ] Base Presenter interface defined
- [ ] 11 Presenter files created
- [ ] All presenters call queries layer
- [ ] All presenters store ViewModels (NOT entities)
- [ ] 70%+ test coverage on presenters (without Bubble Tea)
- [ ] command_tui_new.go created and registered
- [ ] `dw task-manager tui-new` launches refactored UI
- [ ] `dw task-manager tui` still launches old UI (unchanged)

---

## References

- **IMPLEMENTATION-INDEX.md** - Step-by-step checklists
- **domain-ui-integration.md** - ViewModel specs (§3.2), transformations (§4)
- **ui-architecture.md** - Presenter API (§3.4), component designs (§4)
- **proposed-structure.md** - Full directory layout, go-arch-lint config


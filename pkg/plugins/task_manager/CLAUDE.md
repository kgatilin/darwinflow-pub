# Package: task_manager

**Path**: `pkg/plugins/task_manager`

**Role**: Hierarchical roadmap management plugin (Roadmap → Track → Task → Iteration)

---

## Overview

The task-manager plugin provides comprehensive project/product roadmap management with:
- **Multi-project support** - Separate isolated roadmaps (e.g., "production" vs "test")
- **Hierarchical structure** - Roadmap → Track → Task → Iteration
- **SQLite database storage** - Per-project databases with full schema management
- **Full CLI commands** - 33 commands across all entities (28 entity + 5 project commands)
- **Event bus integration** - Cross-plugin communication with 17 event types
- **Interactive TUI** - Visualization and management with project context
- **Event sourcing** - Complete audit trail for all changes

---

## Architecture

### Domain Model

**Roadmap (Root Aggregate)**
- Single active roadmap per project
- Contains vision and success criteria
- Parent to all tracks

**Track (Major Work Area)**
- Represents work streams (e.g., "Framework Core", "Plugin System")
- Has status (not-started, in-progress, complete, blocked, waiting)
- Has priority (critical, high, medium, low)
- Can depend on other tracks (with circular dependency prevention)
- Contains multiple tasks

**Task (Concrete Work Item)**
- Belongs to a track
- Has status (todo, in-progress, done)
- Can have git branch association
- Atomic unit of work
- Can be grouped into iterations

**Iteration (Time-Boxed Grouping)**
- Groups tasks from multiple tracks
- Has status (planned, current, complete)
- Only one can be "current" at a time
- Auto-incrementing iteration numbers
- Deliverable-oriented (goal, deliverable description)

### Multi-Project Architecture

**Project Isolation:**
- Each project has its own SQLite database in `.darwinflow/projects/<project-name>/roadmap.db`
- Active project tracked in `.darwinflow/active-project.txt`
- Complete data isolation between projects
- Auto-migration from legacy single-database structure

**Use Cases:**
- Separate "production" and "test" roadmaps
- Multiple product roadmaps in one workspace
- Experimentation without affecting real data

**Commands:**
- All 28 entity commands support `--project <name>` flag to override active project
- 5 dedicated project management commands (create, list, switch, show, delete)

### Package Structure

**Entity Files (4):**
- `roadmap_entity.go` - RoadmapEntity with IExtensible interface
- `track_entity.go` - TrackEntity with IExtensible and ITrackable interfaces
- `task_entity.go` - TaskEntity with IExtensible and ITrackable interfaces
- `iteration_entity.go` - IterationEntity with IExtensible interface

**Repository Files (3):**
- `repository.go` - RoadmapRepository interface (34 methods)
- `sqlite_repository.go` - SQLite implementation with full CRUD and queries
- `event_emitting_repository.go` - Decorator pattern for event emission

**Command Files (6):**
- `command_project.go` - Project commands (create, list, switch, show, delete)
- `command_roadmap.go` - Roadmap commands (init, show, update)
- `command_track.go` - Track commands (create, list, show, update, delete, dependencies)
- `command_task.go` - Task commands (create, list, show, update, delete, move)
- `command_iteration.go` - Iteration commands (create, list, show, current, start, complete, etc.)
- `command_tui.go` - TUI launch command

**TUI Files (1):**
- `tui_models.go` - Bubble Tea models and views for all entity types

**Event Files (1):**
- `events.go` - Event type constants (17 event types)

**Schema Files (1):**
- `schema.go` - Database schema and migrations (6 tables)

**Test Files (9):**
- `roadmap_entity_test.go` - Roadmap entity tests
- `track_entity_test.go` - Track entity tests (6 test cases)
- `iteration_entity_test.go` - Iteration entity tests (3 test cases)
- `sqlite_repository_test.go` - Repository integration tests (8 test cases)
- `command_roadmap_test.go` - Roadmap command tests (7 test cases)
- `command_track_test.go` - Track command tests (5 test cases)
- `command_task_test.go` - Task command tests (9 test cases)
- `command_iteration_test.go` - Iteration command tests (11 test cases)
- `tui_models_test.go` - TUI tests (11 test cases)
- `plugin_test.go` - Plugin integration tests (2 test cases)

**Other Files (3):**
- `plugin.go` - TaskManagerPlugin implementation (IEntityProvider, ICommandProvider)
- `schema.go` - Database schema definitions and migrations
- `watcher.go` - File system watcher for legacy functionality

### Database Schema

**6 Tables:**
- `roadmaps` - Roadmap entities (id, vision, success_criteria)
- `tracks` - Track entities (id, roadmap_id, title, description, status, priority)
- `track_dependencies` - Track dependency relationships (track_id, depends_on_id)
- `tasks` - Task entities (id, track_id, title, description, status, priority, branch)
- `iterations` - Iteration entities (id, roadmap_id, number, name, goal, status, deliverable)
- `iteration_tasks` - Iteration-task relationships (iteration_id, task_id)

All tables have:
- Primary keys and foreign keys
- Proper indexes on frequently queried columns
- Created_at and updated_at timestamps
- Referential integrity constraints

---

## Commands Overview

### Project Commands

**Create Command**
```bash
dw task-manager project create <name>
```
- Creates new isolated project with dedicated database
- Project names: alphanumeric, hyphens, underscores only
- Auto-initializes project directory structure

**List Command**
```bash
dw task-manager project list
```
- Lists all projects with active project marked (*)
- Shows project names and database paths

**Switch Command**
```bash
dw task-manager project switch <name>
```
- Changes active project for all subsequent commands
- Updates `.darwinflow/active-project.txt`

**Show Command**
```bash
dw task-manager project show
```
- Displays current active project name

**Delete Command**
```bash
dw task-manager project delete <name> [--force]
```
- Deletes project and all its data
- Cannot delete currently active project (switch first)
- Requires --force flag for safety

### Roadmap Commands

**Init Command**
```bash
dw task-manager roadmap init --vision "..." --success-criteria "..."
```
- Creates initial roadmap entity
- Initializes database schema
- Returns roadmap details

**Show Command**
```bash
dw task-manager roadmap show
```
- Displays current roadmap vision and success criteria
- Shows summary of tracks and task counts

**Update Command**
```bash
dw task-manager roadmap update --vision "..." --success-criteria "..."
```
- Updates roadmap vision and/or success criteria
- Emits roadmap.updated event

### Track Commands

**Create Command**
```bash
dw task-manager track create --id <id> --title <title> --description <desc> --priority <priority>
```
- Creates new track
- Validates track ID format (track-*)
- Emits track.created event

**List Command**
```bash
dw task-manager track list [--status <status>] [--priority <priority>]
```
- Lists all tracks with optional filtering
- Shows track ID, title, status, priority, task count, dependencies
- Supports filtering by status and priority

**Show Command**
```bash
dw task-manager track show <track-id>
```
- Displays track details including all nested tasks
- Shows dependency information

**Update Command**
```bash
dw task-manager track update <track-id> [--title] [--description] [--status] [--priority]
```
- Updates track fields
- Emits track.updated event

**Dependency Commands**
```bash
dw task-manager track add-dependency <track-id> <depends-on>
dw task-manager track remove-dependency <track-id> <depends-on>
```
- Manages track dependencies
- Prevents circular dependencies
- Validates dependency relationships

**Delete Command**
```bash
dw task-manager track delete <track-id> [--force]
```
- Deletes track and all child tasks
- Requires --force flag for safety
- Emits track.deleted event

### Task Commands

**Create Command**
```bash
dw task-manager task create --track <track-id> --title <title> [--description] [--priority]
```
- Creates new task in specified track
- Auto-generates task ID
- Emits task.created event

**List Command**
```bash
dw task-manager task list [--track <track-id>] [--status <status>]
```
- Lists all tasks with optional filtering
- Shows task ID, title, track, status, priority
- Supports filtering by track and status

**Show Command**
```bash
dw task-manager task show <task-id>
```
- Displays task details including track, status, branch
- Shows iteration membership if applicable

**Update Command**
```bash
dw task-manager task update <task-id> [--title] [--description] [--status] [--priority] [--branch]
```
- Updates task fields
- Supports branch association for git integration
- Emits task.updated event

**Move Command**
```bash
dw task-manager task move <task-id> --track <new-track-id>
```
- Moves task to different track
- Updates parent track reference
- Emits task.moved event

**Delete Command**
```bash
dw task-manager task delete <task-id> [--force]
```
- Deletes task
- Removes from any iterations
- Requires --force flag
- Emits task.deleted event

### Iteration Commands

**Create Command**
```bash
dw task-manager iteration create --name <name> --goal <goal> --deliverable <deliverable>
```
- Creates new iteration (auto-numbered)
- Sets status to "planned"
- Emits iteration.created event

**List Command**
```bash
dw task-manager iteration list
```
- Lists all iterations
- Shows number, name, status, task count

**Show Command**
```bash
dw task-manager iteration show <iteration-number>
```
- Displays iteration details with all tasks
- Shows progress metrics

**Current Command**
```bash
dw task-manager iteration current
```
- Shows the current active iteration (status = "current")
- Displays detailed task breakdown

**Update Command**
```bash
dw task-manager iteration update <number> [--name] [--goal] [--deliverable]
```
- Updates iteration fields
- Emits iteration.updated event

**Add/Remove Task Commands**
```bash
dw task-manager iteration add-task <iteration> <task-id> [<task-id>...]
dw task-manager iteration remove-task <iteration> <task-id> [<task-id>...]
```
- Manages tasks in iterations
- Handles multiple tasks at once
- Validates task and iteration existence

**Start Command**
```bash
dw task-manager iteration start <iteration-number>
```
- Sets iteration status to "current"
- Only one iteration can be current
- Emits iteration.started event

**Complete Command**
```bash
dw task-manager iteration complete <iteration-number>
```
- Sets iteration status to "complete"
- Marks iteration as finished
- Emits iteration.completed event

**Delete Command**
```bash
dw task-manager iteration delete <iteration-number> [--force]
```
- Deletes iteration
- Removes iteration-task relationships
- Requires --force flag
- Emits iteration.deleted event

### TUI Command

**TUI Launch**
```bash
dw task-manager tui
```
- Launches interactive terminal user interface
- Uses Bubble Tea framework
- Provides multiple views (roadmap, tracks, iterations)

---

## Event Bus Integration

The plugin emits events for all CRUD operations:

**Roadmap Events:**
- `task-manager.roadmap.created` - Roadmap initialized
- `task-manager.roadmap.updated` - Vision/criteria changed

**Track Events:**
- `task-manager.track.created` - New track created
- `task-manager.track.updated` - Track fields updated
- `task-manager.track.status_changed` - Status changed (not-started → in-progress, etc.)
- `task-manager.track.completed` - Track marked complete
- `task-manager.track.blocked` - Track marked blocked

**Task Events:**
- `task-manager.task.created` - New task created
- `task-manager.task.updated` - Task fields updated
- `task-manager.task.status_changed` - Status changed (todo → in-progress, etc.)
- `task-manager.task.completed` - Task marked done
- `task-manager.task.moved` - Moved to different track

**Iteration Events:**
- `task-manager.iteration.created` - New iteration created
- `task-manager.iteration.updated` - Iteration fields updated
- `task-manager.iteration.started` - Iteration marked as current
- `task-manager.iteration.completed` - Iteration marked complete

Other plugins can subscribe to these events for notifications, automation, etc.

---

## Testing

**Test Coverage:** 58.1% (156 tests)

**Test Organization:**
- Entity tests: 13 tests (roadmap, track, iteration)
- Repository tests: 8 tests (SQLite integration)
- Command tests: 32 tests (all command types)
- TUI tests: 11 tests (views and navigation)
- Plugin tests: 2 tests (plugin lifecycle)

**Running Tests:**

```bash
# Run all tests
go test ./pkg/plugins/task_manager -v

# Run with coverage
go test -cover ./pkg/plugins/task_manager

# Generate coverage report
go test -coverprofile=coverage.out ./pkg/plugins/task_manager
go tool cover -html=coverage.out

# Run specific test suites
go test ./pkg/plugins/task_manager -run TestRoadmap
go test ./pkg/plugins/task_manager -run TestTrack
go test ./pkg/plugins/task_manager -run TestTask
go test ./pkg/plugins/task_manager -run TestIteration
go test ./pkg/plugins/task_manager -run TestTUI
```

---

## Plugin Architecture

### Plugin Interface Implementation

**TaskManagerPlugin** implements:
- `pluginsdk.Plugin` - Base plugin interface
- `pluginsdk.IEntityProvider` - Query roadmaps, tracks, tasks, iterations
- `pluginsdk.ICommandProvider` - All CLI commands

**Key Methods:**
- `GetInfo()` - Plugin metadata (name, version, description)
- `GetCapabilities()` - Lists implemented capabilities
- `GetEntityTypes()` - Returns "roadmap", "track", "task", "iteration" entity types
- `Query(ctx, query)` - Query entities with filters and pagination
- `GetEntity(ctx, id)` - Get entity by ID
- `UpdateEntity(ctx, id, fields)` - Update entity fields
- `GetCommands()` - Returns all CLI commands

### Repository Pattern

**RoadmapRepository Interface:**
- 34 methods for complete CRUD and querying
- Methods organized by entity type:
  - Roadmap: Create, Get, Update, Delete
  - Track: Create, Get, Update, Delete, AddDependency, RemoveDependency, GetDependencies
  - Task: Create, Get, Update, Delete, GetByTrack
  - Iteration: Create, Get, Update, Delete, GetCurrent, SetCurrent
  - Query: List with filtering and pagination

**SQLiteRepository Implementation:**
- Implements all 34 methods
- Manages all 6 database tables
- Handles migrations and schema creation
- Provides transaction support for complex operations
- Full error handling and validation

### Event Emission

**EventEmittingRepository Decorator:**
- Wraps underlying repository
- Emits events after successful operations
- Publishes to event bus
- Non-blocking event emission (errors don't fail operations)

---

## Usage Examples

### Complete Workflow

```bash
# 1. Initialize roadmap
dw task-manager roadmap init \
  --vision "Build plugin ecosystem" \
  --success-criteria "5 plugins, 80% coverage"

# 2. Create tracks
dw task-manager track create \
  --id track-core \
  --title "Core Framework" \
  --priority critical

dw task-manager track create \
  --id track-plugins \
  --title "Plugin System" \
  --priority high

# 3. Add dependency
dw task-manager track add-dependency track-plugins track-core

# 4. Create tasks
dw task-manager task create \
  --track track-core \
  --title "Implement event bus" \
  --priority high

dw task-manager task create \
  --track track-plugins \
  --title "Create plugin SDK" \
  --priority high

# 5. Create iteration
dw task-manager iteration create \
  --name "Sprint 1" \
  --goal "Foundation" \
  --deliverable "Event bus and SDK"

# 6. Add tasks to iteration
dw task-manager iteration add-task 1 task-001 task-002

# 7. Start iteration
dw task-manager iteration start 1

# 8. Track progress (in TUI)
dw task-manager tui
```

---

## Performance Characteristics

### Storage

- **Roadmap size**: ~50 bytes
- **Track entry**: ~200 bytes (plus dependencies)
- **Task entry**: ~250 bytes
- **Iteration entry**: ~150 bytes

### Operations

- **Create track**: O(1) - Direct insert
- **List tracks**: O(n) - Full table scan
- **Add dependency**: O(1) - Direct insert with validation
- **Get task by ID**: O(1) - Index lookup
- **Query tasks by track**: O(n) - Index scan
- **Start iteration**: O(1) - Direct update

### Database

- All primary tables have indexes on ID and common query fields
- Track dependencies indexed for fast lookups
- Iteration-task relationships indexed both directions

---

## Key Design Decisions

1. **Hierarchical Structure**: Roadmap → Track → Task → Iteration follows domain hierarchy naturally
2. **Event Sourcing**: All changes emit events for audit trail and cross-plugin notifications
3. **SQLite Persistence**: Reliable local storage without external dependencies
4. **TUI Integration**: Bubble Tea framework for rich terminal user experience
5. **Track Dependencies**: Enables workflow management and blocking detection
6. **Iteration Grouping**: Time-boxed work organizing across tracks

---

## Future Enhancements

Possible extensions:

1. **Batch Operations**: Bulk update/delete commands
2. **Export/Import**: Export roadmaps to markdown or CSV
3. **Recurring Tasks**: Auto-create tasks based on templates
4. **Progress Analytics**: Generate reports and metrics
5. **Git Integration**: Auto-create branches from tasks
6. **Notifications**: Alert on status changes
7. **Collaboration**: Multi-user roadmap management
8. **Estimation**: Story points and burn-down charts

---

## Troubleshooting

### Database Locked

**Cause**: Multiple processes accessing the database simultaneously.

**Solution**: Ensure only one instance of `dw task-manager` is running.

### Circular Dependencies

**Cause**: Attempting to create a circular dependency between tracks.

**Solution**: The system prevents this automatically. Check your dependency graph.

### Iteration Not Updating

**Cause**: Iteration in "complete" status cannot be edited.

**Solution**: Only planned and current iterations can be updated.

---

## Files

**Total: 28 files**

**Entities:** 4 files
**Repository:** 3 files
**Commands:** 5 files
**TUI:** 1 file
**Tests:** 10 files
**Other:** 3 files (plugin.go, events.go, schema.go, watcher.go)

---

## References

- **Plugin Development**: `pkg/pluginsdk/CLAUDE.md` - SDK documentation
- **Architecture**: `/workspace/CLAUDE.md` - DarwinFlow architecture guide
- **Commands**: See README.md for command examples and usage

---

*Updated: 2025-10-31 (Phase 10 - Final documentation, testing, and polish)*

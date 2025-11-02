package task_manager

import (
	"context"
)

// DefaultSystemPrompt contains the built-in system prompt explaining the task manager
// to LLMs working with the task manager. This prompt explains the entity hierarchy,
// workflows, and best practices.
const DefaultSystemPrompt = `# Task Manager System Prompt

## Overview

DarwinFlow Task Manager is a hierarchical project management system designed for coding agents and humans to collaboratively plan, track, and deliver software projects. It organizes work through a four-level hierarchy: Roadmap → Track → Task → Iteration.

Use this system prompt to understand how to effectively use the task manager to plan and track your work.

---

## Entity Hierarchy

### Roadmap (Required - Root Level)
- **Purpose**: Top-level container for a project's direction
- **Scope**: One active roadmap per project
- **Contains**: Vision statement and success criteria
- **Example**: "Build extensible plugin framework for DarwinFlow"
- **Initialization**: dw task-manager roadmap init --vision "..." --success-criteria "..."

### Track (Required - Major Work Areas)
- **Purpose**: Represents major work streams or epics
- **Examples**: "Core Framework", "Plugin System", "Documentation"
- **Status**: not-started, in-progress, complete, blocked, waiting
- **Priority**: critical, high, medium, low
- **Dependencies**: Can depend on other tracks (prevents blocking on unknown work)
- **Creates**: Use dw task-manager track create --id <id> --title <title> --description <desc> --priority <priority>

### Task (Required - Concrete Work Items)
- **Purpose**: Individual, atomic units of work within a track
- **Examples**: "Implement event bus", "Write plugin SDK documentation"
- **Status**: todo, in-progress, done
- **Parent**: Belongs to exactly one track
- **Optional**: Git branch association for version control integration
- **Creates**: Use dw task-manager task create --track <track-id> --title <title>

### Iteration (Optional - Time-Boxed Grouping)
- **Purpose**: Groups tasks for sprint planning and time-boxed delivery
- **Examples**: "Sprint 1", "Q4 2024 - Core Features"
- **Status**: planned, current, complete
- **Constraint**: Only ONE iteration can be "current" at a time
- **Numbering**: Auto-incrementing (Iteration 1, 2, 3...)
- **Creates**: Use dw task-manager iteration create --name <name> --goal <goal> --deliverable <deliverable>

---

## Required vs Optional Entities

### Absolutely Required
1. **Roadmap** - Every project needs one. Initialize it first.
2. **Tracks** - Define your major work streams before creating tasks
3. **Tasks** - The actual work items, always belong to a track

### Optional
- **Iterations** - Use for sprint planning; not required for unscheduled work
- **Track Dependencies** - Use when you need to prevent work before prerequisites complete
- **Task Branches** - Use when integrating with git-based workflows

---

## Standard Workflows

### Workflow 1: Initial Project Setup

Follow this sequence when starting a new project:

	# 1. Create project (if using multi-project setup)
	dw task-manager project create my-project --code PROJ

	# 2. Initialize roadmap with vision and success criteria
	dw task-manager roadmap init \
	  --vision "Build plugin ecosystem for DarwinFlow" \
	  --success-criteria "Support 5+ plugins, 80% test coverage"

	# 3. Create major work tracks
	dw task-manager track create \
	  --id track-core \
	  --title "Core Framework" \
	  --description "Event bus, plugin SDK, and base infrastructure" \
	  --priority critical

	dw task-manager track create \
	  --id track-plugins \
	  --title "Reference Plugins" \
	  --description "Build example plugins to validate SDK" \
	  --priority high

	# 4. Define track dependencies if needed
	dw task-manager track add-dependency track-plugins track-core

### Workflow 2: Iteration Planning (Sprint Planning)

Use iterations to organize time-boxed work:

	# 1. Create iteration
	dw task-manager iteration create \
	  --name "Sprint 1 - Foundation" \
	  --goal "Complete core framework" \
	  --deliverable "Event bus with full test coverage"

	# 2. Create tasks in relevant tracks
	dw task-manager task create --track track-core --title "Implement event bus"
	dw task-manager task create --track track-core --title "Write event bus tests"

	# 3. Add tasks to iteration
	dw task-manager iteration add-task 1 task-001 task-002 task-003

	# 4. Start iteration when ready
	dw task-manager iteration start 1

	# 5. Track progress
	dw task-manager iteration current

	# 6. Mark tasks as done as you complete them
	dw task-manager task update task-001 --status done

	# 7. Complete iteration when all tasks are done
	dw task-manager iteration complete 1

### Workflow 3: Continuous Task Management (No Iterations)

For ongoing work without formal sprints:

	# 1. Create tasks in tracks
	dw task-manager task create \
	  --track track-core \
	  --title "Add support for X" \
	  --priority high

	# 2. View all tasks
	dw task-manager task list

	# 3. Update task status as you work
	dw task-manager task update task-001 --status in-progress
	dw task-manager task update task-001 --status done

	# 4. View track progress
	dw task-manager track show track-core

### Workflow 4: Checking Track Dependencies Before Starting

	# 1. View a track to see if it has dependencies
	dw task-manager track show track-plugins

	# 2. Check if dependent tracks are complete
	dw task-manager track show track-core

	# 3. Only start work if dependencies are met
	dw task-manager task create --track track-plugins --title "New task"

---

## Command Reference by Workflow Stage

### Initialization Commands

| Command | Purpose |
|---------|---------|
| dw task-manager project create <name> | Create new isolated project |
| dw task-manager roadmap init --vision "..." --success-criteria "..." | Initialize project roadmap |

### Track Management

| Command | Purpose |
|---------|---------|
| dw task-manager track create --id <id> --title <title> --priority <priority> | Create work stream |
| dw task-manager track list | View all tracks and progress |
| dw task-manager track show <id> | View track details and tasks |
| dw task-manager track update <id> --status <status> | Update track status |
| dw task-manager track add-dependency <track> <depends-on> | Mark dependencies |
| dw task-manager track delete <id> --force | Delete track and tasks |

### Task Management

| Command | Purpose |
|---------|---------|
| dw task-manager task create --track <id> --title <title> | Create work item |
| dw task-manager task list | View all tasks |
| dw task-manager task show <id> | View task details |
| dw task-manager task update <id> --status <status> | Update task progress |
| dw task-manager task move <id> --track <new-track> | Move to different track |
| dw task-manager task delete <id> --force | Delete task |

### Iteration Management

| Command | Purpose |
|---------|---------|
| dw task-manager iteration create --name <name> --goal <goal> | Create sprint |
| dw task-manager iteration list | View all iterations |
| dw task-manager iteration current | View active iteration |
| dw task-manager iteration add-task <num> <task-id> | Add task to sprint |
| dw task-manager iteration start <num> | Begin sprint |
| dw task-manager iteration complete <num> | Mark sprint done |

### Visualization

| Command | Purpose |
|---------|---------|
| dw task-manager tui | Launch interactive terminal UI |

---

## Best Practices

### 1. Define Tracks Before Tasks
Always create your major work streams (tracks) first. This ensures:
- Clear work organization
- Dependencies are visible
- Tasks have a clear parent
- Status reporting is accurate

Example: Define "Core", "Plugins", "Documentation" BEFORE creating tasks.

### 2. Use Iterations for Structured Planning
Use iterations when you have time-boxed work (sprints, milestones):
- Groups related work from multiple tracks
- Time-boxes delivery (week, month, quarter)
- One active iteration helps focus work
- Clear finish lines for celebrations

Example: "Sprint 1" includes 3 core tasks and 5 plugin tasks.

### 3. Track Priorities to Guide Work Order
Set track and task priorities to communicate urgency:
- Critical: Blocks everything else
- High: Important for roadmap completion
- Medium: Important but not blocking
- Low: Nice to have

Example: "Core Framework" is critical, "Documentation" is medium.

### 4. Use Status Consistently
Keep status up-to-date as work progresses:
- not-started → in-progress → complete (tracks)
- todo → in-progress → done (tasks)
- planned → current → complete (iterations)

Example: Mark a task "in-progress" when you start work, "done" when complete.

### 5. Check Dependencies Before Starting Tracks
Before committing to a track, verify its dependencies:
- View the track: dw task-manager track show <id>
- Check dependencies: dw task-manager track list
- Verify dependent track is complete before starting
- Add explicit dependencies: dw task-manager track add-dependency

Example: "Plugin System" depends on "Core Framework". Wait for Core to finish.

### 6. Use Git Branches with Tasks
Associate tasks with git branches for integration:
- dw task-manager task create --track <id> --title <title> --branch feature/X
- Helps review work in progress
- Automates branch → task relationships

Example: Task "Implement event bus" on branch feature/event-bus.

### 7. Verify Acceptance Criteria (AC) When Defined
If your tasks have acceptance criteria (AC system):
- Review AC before marking task "done"
- AC must all pass before completion
- Update AC in task description

Example: Task AC: "[ ] 80% coverage [ ] All tests pass [ ] Code reviewed"

### 8. Create ADRs for Major Tracks
If using the ADR (Architecture Decision Record) system:
- Create ADR for each major track
- Documents decisions and rationale
- Helps future developers understand choices

Example: "Core Framework" ADR explains architecture choices.

### 9. Use Interactive TUI for Overview
Use the TUI for at-a-glance progress:
- dw task-manager tui
- View all roadmap elements
- Navigate hierarchy
- See progress quickly

### 10. Regular Status Updates
Keep status current:
- Daily: Update task status as work progresses
- Weekly: Review track status, adjust priorities
- Per iteration: Complete iteration when done

---

## Entity Relationships Diagram

Roadmap (1 per project)
  ├─ Track-1 (Major work stream)
  │  ├─ Task-1
  │  ├─ Task-2
  │  └─ Task-3
  │
  ├─ Track-2 (Major work stream)
  │  ├─ Task-4
  │  └─ Task-5
  │
  └─ Iteration-1 (Time-boxed grouping)
     ├─ Task-1 (from Track-1)
     ├─ Task-2 (from Track-1)
     ├─ Task-4 (from Track-2)
     └─ Task-5 (from Track-2)

---

## Status Transitions

### Track Status Flow
- **not-started** → in-progress (start work)
- **in-progress** → complete (finish all tasks)
- **in-progress** → blocked (dependency fails)
- **blocked** → in-progress (dependency resolves)
- **blocked** → waiting (waiting for external input)

### Task Status Flow
- **todo** → in-progress (start work)
- **in-progress** → done (complete and verify)
- **in-progress** → todo (return to backlog)

### Iteration Status Flow
- **planned** → current (dw task-manager iteration start <num>)
- **current** → complete (dw task-manager iteration complete <num>)

---

## Common Pitfalls to Avoid

### 1. Creating Tasks Without Tracks
- Wrong: Creating tasks first
- Right: Create tracks first, then tasks

### 2. Ignoring Track Dependencies
- Wrong: Starting work on dependent track too early
- Right: Check dependencies, verify prerequisites complete

### 3. Tasks in Multiple Iterations
- Wrong: Trying to add same task to multiple sprints
- Right: Each task belongs to one iteration (or none)

### 4. Abandoned Iterations
- Wrong: Creating sprints and never marking complete
- Right: Complete iterations when done or cancel explicitly

### 5. Unclear Task Titles
- Wrong: "Work on stuff" or "Fix things"
- Right: "Implement event bus" or "Add error handling to CLI"

### 6. No Priority Guidance
- Wrong: All tasks marked critical
- Right: Use priorities to guide work order (critical < high < medium < low)

### 7. Blocking Without Dependencies
- Wrong: Blocking tasks without explicit dependencies
- Right: Use track dependencies to show blocking relationships

---

## Integration with Other Systems

### With Acceptance Criteria (AC) System
When AC system is active:
- Define AC for important tasks
- Verify AC before marking task "done"
- Update AC tracking as you work

### With ADR (Architecture Decision Records) System
When ADR system is active:
- Create ADR for major tracks
- Link track to ADR in description
- Document architecture decisions
- Reference ADR in related tasks

### With Git Integration
When using git branches:
- Associate tasks with feature branches
- Branch names should reference task IDs
- Update task status when PR merges

---

## Examples by Project Type

### Software Project Example

	# Initialize
	dw task-manager roadmap init \
	  --vision "Build scalable SaaS platform" \
	  --success-criteria "Support 100k users, 99.9% uptime"

	# Create tracks
	dw task-manager track create --id track-backend --title "Backend API" --priority critical
	dw task-manager track create --id track-frontend --title "Frontend UI" --priority high
	dw task-manager track create --id track-infra --title "Infrastructure" --priority critical

	# Set dependencies
	dw task-manager track add-dependency track-backend track-infra
	dw task-manager track add-dependency track-frontend track-backend

	# Create iteration
	dw task-manager iteration create --name "MVP Sprint" --goal "Launch MVP" --deliverable "Core features live"

	# Create tasks
	dw task-manager task create --track track-backend --title "Implement user API" --priority high
	dw task-manager task create --track track-frontend --title "Build login page" --priority high

	# Add to iteration and start
	dw task-manager iteration add-task 1 task-001 task-002
	dw task-manager iteration start 1

### Plugin Development Example

	# Create plugin roadmap
	dw task-manager roadmap init \
	  --vision "Build email integration plugin" \
	  --success-criteria "Full OAuth2, 100% test coverage"

	# Create tracks for plugin components
	dw task-manager track create --id track-auth --title "Authentication" --priority critical
	dw task-manager track create --id track-sync --title "Email Sync" --priority high
	dw task-manager track create --id track-tests --title "Testing" --priority high

	# Create sprint
	dw task-manager iteration create --name "Week 1" --goal "Auth + sync setup" --deliverable "Core working"

	# Create and track tasks
	dw task-manager task create --track track-auth --title "Implement OAuth2"
	dw task-manager task create --track track-sync --title "Set up email polling"
	dw task-manager task create --track track-tests --title "Write integration tests"

---

## Quick Start Checklist

- [ ] Initialize roadmap with vision and success criteria
- [ ] Create 3-5 major work tracks
- [ ] Add dependencies between tracks if applicable
- [ ] Create tasks within tracks
- [ ] (Optional) Create iteration for sprint planning
- [ ] (Optional) Add tasks to iteration
- [ ] Update task status as you work
- [ ] Use TUI to visualize progress
- [ ] Complete iteration when done
- [ ] Move to next iteration

---

## Getting Help

- View all commands: dw help task-manager
- View specific command help: dw task-manager <command> --help
- View interactive UI: dw task-manager tui
- List all tasks: dw task-manager task list
- View track details: dw task-manager track show <id>

---

## Key Principles

1. **Hierarchical**: Roadmap → Track → Task → Iteration
2. **Flexible**: Use iterations for sprints, skip for continuous work
3. **Traceable**: Events logged for all changes
4. **Observable**: TUI and commands show real-time status
5. **Extensible**: Integrates with AC, ADR, and git systems
6. **Auditable**: Full history of changes via event sourcing
`

// GetSystemPrompt returns the system prompt for the task manager.
// It currently returns the default prompt but can be extended to support
// configuration-based prompts in the future.
func GetSystemPrompt(ctx context.Context) string {
	return DefaultSystemPrompt
}

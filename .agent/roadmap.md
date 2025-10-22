# DarwinFlow Roadmap - Active Items

> **Purpose**: Track active work items with links to detailed implementation docs.
> **When switching tasks**: Leave notes here, link to detailed docs with checklists.
> **On completion**: Move item to `roadmap_done.md` with completion date.

---

## Template for New Items

```markdown
## [Priority] Item Name - [Status]

**Started**: YYYY-MM-DD
**Status**: In Progress | Blocked | On Hold
**Branch**: `branch-name` (if applicable)

**Description**: One-liner summary of what this is.

**Details**: `@.agent/details/item-name.md` (link to detailed doc)

**Next Steps**:
- [ ] Next concrete action
- [ ] Another action

**Blockers**: (if any)
```

---

## Active Items

## [HIGH] Task Tool & TUI Improvements - Planned

**Status**: Planned
**Description**: Improve task tool functionality, entity view in TUI, and test how plugins work with terminal UI.

**Key Areas**:
- Enhanced task tool functionality
- Better entity view in TUI
- Plugin integration with TUI testing

---

## [HIGH] Slack Plugin - Planned

**Status**: Planned
**Description**: Connect to Slack to check messages, requires continuous background process support.

**Requirements**:
- Slack API integration
- Background process/daemon support

---

## [MEDIUM] Scheduling Support (Crontab) - Planned

**Status**: Planned
**Description**: Add crontab-like scheduling, either in plugin itself or in core framework via Schedulable interface.

**Options**:
- Plugin-level crontab support
- Framework-level Schedulable interface

---

## [MEDIUM] Task Manager Enhancements - Planned

**Status**: Planned
**Description**: Add comments to tasks, view descriptions, complete tasks with status updates.

**Features**:
- Add comments to tasks
- View task description
- Complete task (update status)
- Consider incorporating into TUI interface

---

## [MEDIUM] Project Registry - Planned

**Status**: Planned
**Description**: Register projects on `dw claude init`, support multiple projects in single UI with tabs and cross-project views.

**Goals**:
- Project registration on init
- Multi-project UI with tabs
- Cross-project entity views (e.g., all tasks)
- Single dashboard for all projects

---

## [HIGH] Database Relocation - Planned

**Status**: Planned
**Description**: Move database from local project folder to `~/.darwinflow/<project>` for cleaner repo structure.

**Changes**:
- Per-project databases in global folder
- Keep out of repository
- Store project registry in global folder
- Enable single dashboard architecture

---

## [MEDIUM] Go Plugin Template - Planned

**Status**: Planned
**Description**: Create basic product template for creating plugins in Go.

**Features**:
- Boilerplate plugin structure
- Example implementations
- Best practices and patterns
- Easy scaffolding for new plugins

---

## [MEDIUM] SDK Future Phases (5-7) - In Progress

**Started**: 2024-10-20
**Status**: In Progress (Phase 5 complete, ready for Phase 6)
**Branch**: N/A

**Description**: JSON-RPC protocol, plugin discovery, and external SDKs.

**Details**: `@.agent/details/sdk-future-phases.md`

**Next Steps**:
- [ ] Phase 6: Plugin configuration & auto-loading from .darwinflow/plugins.yaml
- [ ] Phase 7: Python/TypeScript SDKs

**Blockers**: None - ready to proceed

**Completed**:
- ✅ Phase 4 - Multi-plugin event streams (2025-10-22)
- ✅ Phase 5 - JSON-RPC for external plugins (2025-10-22)

---

## Backlog Ideas

### Tool Feedback Enhancement for Analysis Workflow

**Created**: 2024-10-18
**Priority**: Low
**Status**: Idea / Research

**Description**: Enhance pattern analysis to support tool feedback collection for cross-project learning.

**Key Features**:
- Additional prompt injection in analysis
- Structured feedback collection about tool usage
- Cross-project learning and reflection workflow
- Agent-driven tool improvement

**Reference**: `.agent/backlog/tool-feedback-enhancement.md`

**Next**: Needs further design work and prioritization

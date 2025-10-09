# Storage Architecture

**Author**: Bob Martin
**Date**: 2025-10-09
**Status**: Target Architecture - Storage Layer
**Version**: 1.0

---

## Purpose

Storage is completely abstracted from core framework. The system must work with files, SQLite, Postgres, or anything else without changing a line of core code.

**Key principle**: Core framework never knows about storage implementation.

---

## Storage Responsibilities

### Separate Concerns

Three distinct storage concerns:

| Storage Type | Purpose | Characteristics |
|--------------|---------|-----------------|
| **Workflows** | Store workflow definitions | Small, versioned, human-readable |
| **Logs** | Store execution details | Large, append-only, queryable |
| **Metrics** | Store aggregated data | Medium, updated frequently, analyzable |

**Why separate?** Different access patterns, different scaling needs, different lifecycles.

---

## Storage Interfaces

Defined in `01_core_interfaces.md`, critical operations highlighted here:

### WorkflowRepository

```
Interface: WorkflowRepository
  - Store(workflow) -> error
  - Get(id, version) -> workflow, error
  - GetLatest(id) -> workflow, error
  - List(id) -> []workflow, error
  - ListAll() -> []workflow, error
  - Delete(id, version) -> error
```

**Key characteristics**:
- Explicit versioning (no implicit "latest")
- Version immutability (can't modify v1 after creation)
- Lineage tracking (v2 knows it came from v1)

### LogRepository

```
Interface: LogRepository
  - StoreLog(log) -> error
  - GetLog(runID) -> log, error
  - QueryLogs(filter) -> []log, error
  - GetRecent(workflowID, version, limit) -> []log, error
  - Delete(runID) -> error
```

**Key characteristics**:
- Append-only (logs never modified after creation)
- Rich querying (filter by workflow, status, time range)
- Comprehensive detail (everything needed for reflection)

### MetricsRepository

```
Interface: MetricsRepository
  - RecordExecution(metric) -> error
  - GetSummary(workflowID, version) -> summary, error
  - GetPatterns(workflowID) -> []pattern, error
  - RecordPattern(pattern) -> error
```

**Key characteristics**:
- Aggregated data (not raw logs)
- Updated incrementally (after each execution)
- Pattern storage (for reflection)

---

## Storage Implementation Strategies

### File-Based Storage (MVP)

**Why files first?**

✓ Zero dependencies - no database setup required
✓ Human-readable - can inspect with cat/grep
✓ Version-controllable - git-friendly
✓ Simple - no schema migrations
✓ Debuggable - easy to understand what's happening

**When appropriate**:
- MVP (< 100 executions)
- Development environments
- Single-user scenarios
- Workflow definitions (always - YAML format)

**Limitations**:
- Slow queries (scan all files)
- No transactions
- Concurrent access challenges
- Doesn't scale beyond ~1000 logs

### Database Storage (Growth)

**Why databases?**
- Fast queries (indexed)
- Transactions (ACID guarantees)
- Concurrent access (locks, isolation)
- Better scalability

**SQLite** (Phase 2):
- Single-file database
- Zero-config
- ACID transactions
- Good for < 10K executions
- No network overhead

**PostgreSQL** (Phase 3):
- Multi-user access
- Better concurrency
- Advanced querying
- Scalable to millions of rows
- Rich ecosystem

**Concerns to address**:
- Data consistency (transactions, constraints)
- Query performance (indexes, query planning)
- Scalability (partitioning, archiving)
- Migration path (file → SQLite → Postgres)

---

## Storage Strategy by Phase

### Phase 1 (MVP)

**Use**: File-based storage for everything

**Rationale**:
- Simple to implement
- Zero dependencies
- Easy to debug
- Human-readable

**Trade-offs accepted**:
- Slow queries (acceptable for < 100 logs)
- No concurrent writes (single user)
- Manual data exploration

### Phase 2 (Growth)

**Use**: Database for logs/metrics, files for workflows

**Rationale**:
- Faster queries for analysis
- Better concurrent access
- Keep workflows human-readable

**Why keep workflows as files?**
- Git-friendly version control
- Easy manual editing
- LLM-friendly format
- No database coupling

### Phase 3 (Scale)

**Use**: Database for logs/metrics, files for workflows (or all in database)

**Options**:
- Keep workflows as files (recommended)
- Move workflows to database (if needed)
- Hybrid approach based on use case

---

## Storage Configuration

Make storage pluggable via configuration:

```yaml
storage:
  workflows:
    type: file  # or: sqlite, postgres
    path: darwinflow-data/workflows

  logs:
    type: sqlite  # or: file, postgres
    path: darwinflow.db

  metrics:
    type: sqlite  # or: file, postgres
    path: darwinflow.db
```

**Loading pattern**:
1. Read config at startup
2. Create appropriate repository implementations
3. Inject into core framework
4. Core framework doesn't know which implementation is used

**This enables**:
- Mix and match backends (files for workflows, DB for logs)
- Easy testing (in-memory for tests, files for dev, DB for prod)
- Migration without code changes (just config update)

---

## Data Concerns by Storage Type

### Workflows

**Concerns**:
- Version integrity (can't lose workflow definitions)
- Lineage tracking (know which version came from which)
- Human readability (developers need to read/edit)
- Diffing (compare versions easily)

**Best practice**: Always use files (YAML format)

### Logs

**Concerns**:
- Queryability (find logs by workflow, status, time)
- Completeness (capture everything for reflection)
- Performance (don't slow down execution)
- Retention (old logs should be archived/deleted)

**Evolution path**: Files → SQLite → Postgres as scale grows

### Metrics

**Concerns**:
- Aggregation accuracy (averages, counts must be correct)
- Update performance (updated after each execution)
- Pattern storage (structured data for reflection)
- Historical tracking (trends over time)

**Evolution path**: Files → SQLite → Postgres as scale grows

---

## Migration Strategy

### File → SQLite

**Trigger**: Query performance becomes painful (> 1000 logs)

**Process**:
1. Implement SQLite repositories
2. Write migration script (read files, write to DB)
3. Update config to use SQLite
4. Restart system
5. Verify data migrated correctly
6. Archive old files

**Key**: Same interfaces, different implementation. Core code unchanged.

### SQLite → PostgreSQL

**Trigger**: Multi-user access needed or > 10K logs

**Process**:
1. Implement Postgres repositories
2. Write migration script (read SQLite, write to Postgres)
3. Update config to use Postgres
4. Restart system
5. Verify data migrated correctly
6. Archive old SQLite file

**Key**: Again, same interfaces. Core code still unchanged.

---

## Storage Best Practices

### DO:

✓ **Separate concerns** - workflows, logs, metrics are different
✓ **Start simple** - files for MVP
✓ **Maintain same interface** - swapping storage should be one line (config change)
✓ **Test with multiple backends** - ensures abstraction works
✓ **Version data formats** - allow future migrations
✓ **Handle corruption gracefully** - skip bad files/records, don't crash

### DON'T:

✗ **Leak storage details** - core shouldn't know about SQL, file paths, etc.
✗ **Optimize prematurely** - files are fine for MVP
✗ **Assume one storage** - allow hybrid (files + DB)
✗ **Ignore data retention** - old logs should be archived/deleted
✗ **Skip backups** - especially for metrics and workflows

---

## Data Retention & Cleanup

Logs grow forever - need cleanup strategy.

**Retention concerns**:
- Disk space (logs can consume GBs)
- Query performance (more logs = slower queries)
- Compliance (may need to delete old data)
- Analysis value (old logs may have historical patterns)

**Strategies**:
- Age-based retention (delete logs older than N days)
- Count-based retention (keep only N most recent logs)
- Archival (move old logs to cold storage instead of deleting)
- Selective retention (keep successful runs short, failed runs long)

**Implementation**: Background job that runs periodically to clean up old data.

---

## Testing Storage

Each storage implementation should have comprehensive tests:

**Test coverage**:
- Store and retrieve (round-trip)
- Query operations (filters work correctly)
- Error handling (missing data, corruption)
- Concurrent access (if applicable)
- Migration (data integrity preserved)

**Testing strategy**:
- Use in-memory or temp files for tests
- Test each repository interface independently
- Verify substitutability (can swap implementations)
- Integration tests with real storage

---

## Storage Architecture Boundaries

### Core ↔ Storage Boundary

**Core defines**: Storage interfaces (WorkflowRepository, LogRepository, MetricsRepository)

**Storage implements**: Concrete repositories (FileLogRepository, SQLiteLogRepository, etc.)

**Rule**: Core never imports storage packages. Storage imports core.

**Dependency flow**:
```
Core (defines interfaces)
  ↑ implements
Storage (files, SQLite, Postgres)
```

---

## Summary

Storage architecture enables:

1. **Three separate repositories** - workflows, logs, metrics
2. **Interface-based** - core never knows implementation
3. **Progressive evolution** - files → SQLite → Postgres
4. **Configurable** - choose storage per repository
5. **Testable** - easy to mock for tests
6. **Flexible** - mix and match backends as needed

This design allows starting simple and scaling without rewrites.

---

**Next**: Read [04_evolution_roadmap.md](04_evolution_roadmap.md) for how the system evolves over time.

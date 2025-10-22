# Plugin SDK Future Phases (4-7) - Detailed Implementation

**Created**: 2024-10-20
**Status**: Planned (Phase 3 complete, ready to proceed)
**Branch**: N/A

---

## Overview

Extend the plugin SDK to support multi-plugin events, external plugins via JSON-RPC, plugin discovery, and language-specific SDKs.

**Source**: `.agent/workingfolder/sdk-future-phases-checklist.md`

---

## Current State

**Completed**:
- ✅ Phase 1: SDK Foundation (capability-driven, zero dependencies)
- ✅ Phase 2: Minimal entity capabilities (IExtensible only)
- ✅ Phase 3: Event integration (hooks emit via EmitEvent)
- ✅ Phase 4: Multi-plugin event streams (EventDispatcher, task-manager plugin, TUI real-time updates)

**Ready for**: Phase 5 - JSON-RPC for External Plugins

---

## Phase 4: Multi-Plugin Events (Async Event Streams) - ✅ COMPLETED

**Goal**: Support real-time event streams from multiple plugins simultaneously

### Task 4.1: Define IEventEmitter Capability - ✅ DONE
- ✅ Added IEventEmitter interface to SDK (`pkg/pluginsdk/capability.go:39-67`)
- ✅ StartEventStream(ctx, chan<- Event) error
- ✅ StopEventStream() error
- ✅ Updated capability registry

### Task 4.2: Implement Event Dispatcher - ✅ DONE
- ✅ Created EventDispatcher in `internal/app/event_dispatcher.go`
- ✅ Buffered event channel (size: 100)
- ✅ RegisterEmitter(), Start(), Stop(), GetMetrics()
- ✅ Background goroutine to consume events
- ✅ Stores events via PluginContext.EmitEvent()

### Task 4.3: Create Example Event Emitter Plugin - ✅ DONE
- ✅ Implemented task-manager plugin (`pkg/plugins/task_manager/`)
- ✅ File watcher using fsnotify (cross-platform)
- ✅ Emits task.created, task.updated, task.deleted events
- ✅ CLI commands: init, create, list, update
- ✅ 59.1% test coverage (13 tests)

### Task 4.4: Update TUI for Real-Time Events - ✅ DONE
- ✅ Added event subscription to TUI Model (`internal/app/tui/app.go`)
- ✅ Created EventArrivedMsg Bubble Tea message (`internal/app/tui/types.go`)
- ✅ Auto-refresh views on event arrival
- ✅ Event counter badge in session list title (+N new)
- ✅ Clear badge when user navigates to events view

### Task 4.5: Validation - ✅ DONE
- ✅ 3 plugins emitting events simultaneously (MultiplePlugins test)
- ✅ TUI updates in real-time (integration tested)
- ✅ Event throughput >30,000/sec (far exceeds 1,000/sec requirement)
- ✅ 0 architecture violations
- ✅ 6 comprehensive tests (all passing)

**Completion Date**: 2025-10-22
**Total Implementation**: ~2100 lines of code

---

## Phase 5: JSON-RPC Protocol (External Plugins) - ✅ COMPLETED

**Goal**: Enable non-Go plugins via subprocess communication

### Task 5.1: Define JSON-RPC Protocol Types - ✅ DONE
- ✅ RPCRequest, RPCResponse, RPCEvent types (`pkg/pluginsdk/rpc.go`)
- ✅ Method definitions: init, get_info, get_capabilities, query_entities, get_entity, update_entity, execute_command, start/stop_event_stream
- ✅ Protocol specification in comments and README
- ✅ Newline-delimited JSON over stdin/stdout

### Task 5.2: Implement RPC Client (Main → Plugin) - ✅ DONE
- ✅ RPCClient struct with subprocess management (`internal/infra/rpc_client.go`)
- ✅ Start(), Stop(), Call(method, params), SetEventChannel()
- ✅ Request/response correlation by atomic counter ID
- ✅ Timeout handling (default: 30s, configurable via context)
- ✅ Subprocess crash detection in monitorProcess()
- ✅ Comprehensive tests (7 test cases in `rpc_client_test.go`)

### Task 5.3: Implement Subprocess Plugin Wrapper - ✅ DONE
- ✅ SubprocessPlugin adapter (`internal/infra/subprocess_plugin.go`)
- ✅ Implements all SDK interfaces: Plugin, IEntityProvider, IEntityUpdater, ICommandProvider, IEventEmitter
- ✅ Event streaming from subprocess stdout via RPCEvent
- ✅ Graceful shutdown with timeout and force-kill fallback
- ✅ Comprehensive tests (5 test cases in `subprocess_plugin_test.go`)

### Task 5.4: Create Example Go External Plugin - ✅ DONE (Changed from Python)
- ✅ Go RPC server implementation (`examples/external_plugin/cmd/notes-plugin/main.go`)
- ✅ Notes plugin with IEntityProvider, IEntityUpdater, IEventEmitter
- ✅ Event emission for entity updates
- ✅ Comprehensive README with protocol documentation
- ✅ Working binary built to `bin/notes-plugin`

### Task 5.5: Validation - ✅ DONE
- ✅ External Go plugin working (notes-plugin)
- ✅ Protocol validated manually with echo test
- ✅ Subprocess stability verified
- ✅ 0 architecture violations (verified: SDK has no internal imports, proper layer separation)
- ✅ Clean build with no errors

**Completion Date**: 2025-10-22
**Total Implementation**: ~1200 lines of code (protocol types, RPC client, subprocess adapter, example plugin, tests, documentation)

---

## Phase 6: Plugin Configuration & Auto-Loading

**Goal**: Load external plugins from `.darwinflow/plugins.yaml` configuration

### Task 6.1: Define plugins.yaml Schema (MCP-like format)
- [ ] YAML schema with plugin entries
- [ ] Fields: name, command (executable path), args, env
- [ ] Optional fields: enabled, timeout, restart_on_crash
- [ ] Example plugins.yaml file
- [ ] Similar to MCP server configuration format

Example:
```yaml
plugins:
  notes:
    command: ./bin/notes-plugin
    args: []
    env:
      DEBUG: "false"
    enabled: true

  my-python-plugin:
    command: python3
    args: ["-m", "my_plugin"]
    env:
      PYTHONPATH: "/path/to/plugins"
    enabled: true
```

### Task 6.2: Implement Plugin Loader
- [ ] PluginLoader struct in `internal/app/plugin_loader.go`
- [ ] LoadFromConfig(configPath) method
- [ ] Parse plugins.yaml
- [ ] Create SubprocessPlugin for each entry
- [ ] Validate executable exists
- [ ] Error handling (skip invalid, log warnings)

### Task 6.3: Integrate with Bootstrap
- [ ] Update `cmd/dw/main.go` bootstrap
- [ ] Load external plugins from `.darwinflow/plugins.yaml`
- [ ] Register alongside core plugins
- [ ] Graceful degradation if plugins.yaml doesn't exist
- [ ] Log loaded plugin info

### Task 6.4: Add Plugin Management Commands
- [ ] `dw plugin list` - List all plugins (core + external)
- [ ] `dw plugin enable <name>` - Enable plugin
- [ ] `dw plugin disable <name>` - Disable plugin
- [ ] `dw plugin reload` - Reload plugins.yaml

### Task 6.5: Validation
- [ ] Multiple external plugins loaded from config
- [ ] Executable path resolution works
- [ ] Environment variables passed correctly
- [ ] Disabled plugins not loaded
- [ ] Invalid plugins don't crash system
- [ ] 0 architecture violations

**Estimated Time**: 4-6 hours

---

## Phase 7: External SDKs (Python & TypeScript)

**Goal**: Provide official SDKs for plugin development in other languages

### Task 7.1: Create Python SDK
**Repository**: darwinflow-pluginsdk-python

- [ ] Plugin base class
- [ ] EntityProvider, CommandProvider, EventEmitter classes
- [ ] RPC server implementation (stdin/stdout)
- [ ] Event emitter helper
- [ ] Example plugin
- [ ] Unit tests
- [ ] Publish to PyPI

### Task 7.2: Create TypeScript SDK
**Repository**: darwinflow-pluginsdk-typescript

- [ ] TypeScript interfaces matching SDK
- [ ] RPC server implementation (readline)
- [ ] Type definitions
- [ ] Example plugin
- [ ] Unit tests
- [ ] Publish to npm

### Task 7.3: Documentation & Plugin Development Guide
- [ ] Plugin development guide (complete)
- [ ] API reference documentation
- [ ] Best practices guide
- [ ] 5+ example plugins
- [ ] Testing guide

### Task 7.4: Validation
- [ ] Python SDK on PyPI
- [ ] TypeScript SDK on npm
- [ ] 5+ example plugins working
- [ ] Documentation comprehensive
- [ ] External developers can create plugins

**Estimated Time**: 12-16 hours

---

## Success Metrics Summary

**Phase 4**:
- ✅ 2+ plugins emitting events simultaneously
- ✅ Real-time TUI updates
- ✅ Event throughput >1000/sec

**Phase 5**:
- ✅ External plugin (Python) working
- ✅ RPC overhead <5ms
- ✅ Subprocess stability

**Phase 6**:
- ✅ Auto-discovery working
- ✅ 5+ plugins loaded
- ✅ Error recovery robust

**Phase 7**:
- ✅ Python SDK on PyPI
- ✅ TypeScript SDK on npm
- ✅ 5+ example plugins
- ✅ Complete documentation

---

## Risk Assessment

**High Risk**:
- Subprocess management complexity (Phase 5)
- RPC protocol stability (Phase 5)
- Event throughput performance (Phase 4)

**Medium Risk**:
- Plugin discovery edge cases (Phase 6)
- SDK API design (Phase 7)

**Low Risk**:
- Event dispatcher implementation (Phase 4)
- Manifest schema design (Phase 6)

**Mitigation**:
- Comprehensive testing at each phase
- Early performance benchmarks
- User feedback on SDK design

---

## Next Steps

1. **Decide**: Address interface deduplication first OR proceed with Phase 4?
2. **If proceeding**: Start with Task 4.1 (Define IEventEmitter)
3. **Commit strategy**: Commit after each task for granular rollback

---

## References

- Full checklist: `.agent/workingfolder/sdk-future-phases-checklist.md`
- Completed phases: `.agent/workingfolder/sdk-implementation-checklist.md`
- Architecture: `.agent/backlog/plugin-sdk-architecture.md`

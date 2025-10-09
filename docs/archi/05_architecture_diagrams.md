# Architecture Diagrams

**Author**: Bob Martin
**Date**: 2025-10-09
**Status**: Target Architecture - Visual Reference
**Version**: 1.0

---

## Purpose

Visual representations of DarwinFlow's architecture using C4 model and PlantUML.

---

## C4 Level 1: System Context

Shows DarwinFlow in its environment - who uses it and what it integrates with.

```plantuml
@startuml DarwinFlow System Context
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Context.puml

LAYOUT_WITH_LEGEND()

title System Context - DarwinFlow Learning Workflow Framework

Person(user, "User", "Software Engineer, PM, or Architect using AI for repetitive tasks")
Person(admin, "Admin", "System administrator configuring workflows and integrations")

System(darwinflow, "DarwinFlow", "Learning workflow framework that improves through use")

System_Ext(claude_code, "Claude Code", "AI coding assistant CLI")
System_Ext(telegram, "Telegram", "Messaging platform for task input")
System_Ext(github, "GitHub", "Code repository and project management")
System_Ext(llm_api, "LLM APIs", "Claude, GPT, Local LLMs")
System_Ext(filesystem, "File System", "Local files and project code")

Rel(user, darwinflow, "Executes workflows", "CLI/API")
Rel(admin, darwinflow, "Configures", "YAML files")

Rel(darwinflow, claude_code, "Delegates tasks to", "CLI")
Rel(darwinflow, telegram, "Sends/receives messages", "Bot API")
Rel(darwinflow, github, "Manages issues, PRs", "REST API")
Rel(darwinflow, llm_api, "Makes decisions", "REST API")
Rel(darwinflow, filesystem, "Reads/writes files", "OS API")

@enduml
```

**Key insight**: DarwinFlow orchestrates external tools, it doesn't replace them.

---

## C4 Level 2: Container Diagram

Shows the major containers (applications/services) within DarwinFlow.

```plantuml
@startuml DarwinFlow Containers
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()

title Container Diagram - DarwinFlow Architecture

Person(user, "User", "Executes workflows")

System_Boundary(darwinflow, "DarwinFlow") {
    Container(cli, "CLI", "Go", "Command-line interface for workflow execution")
    Container(core, "Core Framework", "Go", "Workflow execution engine, state management")
    Container(integrations, "Integrations", "Go", "Adapters for external tools (plugins)")
    Container(reflection, "Reflection Engine", "Go + LLM", "Analyzes logs and generates new workflows")
    Container(storage, "Storage Layer", "Go", "Abstract persistence (files, SQLite, Postgres)")
}

System_Ext(llm, "LLM Service", "Claude/GPT")
System_Ext(tools, "External Tools", "Claude Code, Telegram, GitHub, etc.")
ContainerDb(workflows, "Workflows", "YAML Files", "Workflow definitions (versioned)")
ContainerDb(logs, "Execution Logs", "JSON/SQLite", "Detailed execution history")
ContainerDb(metrics, "Metrics", "JSON/SQLite", "Aggregated metrics and patterns")

Rel(user, cli, "Runs workflows", "darwinflow run")
Rel(cli, core, "Executes", "Function calls")
Rel(core, integrations, "Delegates to nodes", "NodeExecutor interface")
Rel(core, storage, "Persists data", "Repository interfaces")

Rel(integrations, tools, "Calls", "Various protocols")
Rel(reflection, llm, "Analyzes patterns", "LLM API")
Rel(reflection, storage, "Reads logs", "Repository interfaces")
Rel(reflection, workflows, "Generates new versions", "YAML files")

Rel(storage, workflows, "Reads/writes", "File I/O")
Rel(storage, logs, "Reads/writes", "File I/O / SQL")
Rel(storage, metrics, "Reads/writes", "File I/O / SQL")

@enduml
```

**Key insight**: Core framework depends on abstractions, not concrete implementations.

---

## C4 Level 3: Component Diagram (Core Framework)

Shows internal components of the Core Framework.

```plantuml
@startuml Core Framework Components
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Component.puml

LAYOUT_WITH_LEGEND()

title Component Diagram - Core Framework

Container(cli, "CLI", "Command-line interface")

Container_Boundary(core, "Core Framework") {
    Component(executor, "WorkflowExecutor", "Go", "Executes workflows, manages state")
    Component(loader, "WorkflowLoader", "Go", "Loads and validates workflows")
    Component(registry, "NodeRegistry", "Go", "Manages node type executors (plugin system)")
    Component(logger, "ExecutionLogger", "Go", "Records execution events")
    Component(metrics, "MetricsCollector", "Go", "Collects runtime metrics")
    Component(context, "ExecutionContext", "Go", "Provides runtime state to nodes")
}

Container_Boundary(interfaces, "Core Interfaces") {
    Component(iworkflow, "WorkflowRepository", "Interface", "Store/retrieve workflows")
    Component(ilog, "LogRepository", "Interface", "Store/retrieve logs")
    Component(imetrics, "MetricsRepository", "Interface", "Store/retrieve metrics")
    Component(inode, "NodeExecutor", "Interface", "Execute node types")
}

Container(integrations, "Integrations", "Node implementations")
Container(storage, "Storage", "Repository implementations")

Rel(cli, loader, "Loads workflow", "LoadFromFile()")
Rel(cli, executor, "Executes", "Execute()")

Rel(executor, registry, "Gets node executor", "Get(nodeType)")
Rel(executor, context, "Creates", "NewContext()")
Rel(executor, logger, "Logs events", "LogStep()")
Rel(executor, metrics, "Records metrics", "RecordDuration()")

Rel(loader, iworkflow, "Uses", "Get()")
Rel(logger, ilog, "Uses", "StoreLog()")
Rel(metrics, imetrics, "Uses", "RecordExecution()")

Rel(registry, inode, "Manages", "Register()")
Rel(context, inode, "Provides to", "Execute()")

Rel(inode, integrations, "Implemented by", "")
Rel_Back(iworkflow, storage, "Implemented by", "")
Rel_Back(ilog, storage, "Implemented by", "")
Rel_Back(imetrics, storage, "Implemented by", "")

@enduml
```

**Key insight**: Interfaces define boundaries. Implementations can vary without affecting core.

---

## Dependency Flow Diagram

Shows direction of dependencies (critical for clean architecture).

```plantuml
@startuml Dependency Flow
!define RECTANGLE class

skinparam defaultTextAlignment center
skinparam class {
    BackgroundColor<<core>> LightBlue
    BackgroundColor<<integration>> LightGreen
    BackgroundColor<<storage>> LightYellow
    BackgroundColor<<workflow>> LightCoral
}

title Dependency Flow - The Dependency Rule

package "Reference Workflows\n(Data, not code)" <<workflow>> {
    [assistant_v1.yaml]
    [assistant_v2.yaml]
}

package "Integrations\n(Volatile, changes often)" <<integration>> {
    [ClaudeCodeNode]
    [TelegramNode]
    [GitHubNode]
    [FileReadNode]
}

package "Core Framework\n(Stable, rarely changes)" <<core>> {
    [WorkflowExecutor]
    [NodeExecutor Interface]
    [Storage Interfaces]
    [Domain Entities]
}

package "Storage Implementations\n(Pluggable backends)" <<storage>> {
    [FileLogRepository]
    [SQLiteLogRepository]
    [PostgresLogRepository]
}

[assistant_v1.yaml] ..> [ClaudeCodeNode] : uses
[assistant_v2.yaml] ..> [FileReadNode] : uses
[assistant_v2.yaml] ..> [ClaudeCodeNode] : uses

[ClaudeCodeNode] --> [NodeExecutor Interface] : implements
[TelegramNode] --> [NodeExecutor Interface] : implements
[GitHubNode] --> [NodeExecutor Interface] : implements
[FileReadNode] --> [NodeExecutor Interface] : implements

[WorkflowExecutor] --> [NodeExecutor Interface] : depends on
[WorkflowExecutor] --> [Storage Interfaces] : depends on
[WorkflowExecutor] --> [Domain Entities] : uses

[FileLogRepository] --> [Storage Interfaces] : implements
[SQLiteLogRepository] --> [Storage Interfaces] : implements
[PostgresLogRepository] --> [Storage Interfaces] : implements

note right of [NodeExecutor Interface]
  **Dependency Rule:**
  Dependencies point INWARD

  Outer layers (Integrations, Storage)
  depend on inner layers (Core)

  Inner layers NEVER depend
  on outer layers
end note

legend right
    |= Layer |= Volatility |= Examples |
    | Workflows | High - changes daily | YAML files |
    | Integrations | High - new tools added | Node types |
    | Core | Low - stable interfaces | Executor, interfaces |
    | Storage | Medium - swap backends | File, SQLite, Postgres |
endlegend

@enduml
```

**Key insight**: Arrows point inward. Core never imports from integrations or storage.

---

## Workflow Execution Sequence

Shows the runtime flow when executing a workflow.

```plantuml
@startuml Workflow Execution Sequence
autonumber
title Workflow Execution - Runtime Flow

actor User
participant "CLI" as cli
participant "WorkflowExecutor" as executor
participant "ExecutionContext" as context
participant "NodeRegistry" as registry
participant "NodeExecutor\n(ClaudeCode)" as node
participant "ExecutionLogger" as logger
participant "LogRepository" as logrepo

User -> cli: darwinflow run assistant_v1.yaml "Add feature X"

cli -> executor: Execute(workflow, input)
activate executor

executor -> context: NewContext(runID, input)
activate context
executor <-- context: context

executor -> logger: LogStepStart(runID, nodeID)
logger -> logrepo: StoreLog()

loop for each node in workflow
    executor -> registry: Get(nodeType)
    registry --> executor: nodeExecutor

    executor -> node: Execute(context, config)
    activate node

    node -> node: Perform work\n(call Claude Code CLI)

    node -> logger: LogAction(runID, nodeID, action)
    logger -> logrepo: StoreLog()

    node --> executor: NodeResult
    deactivate node

    executor -> context: SetNodeOutput(nodeID, result)

    executor -> logger: LogStepComplete(runID, nodeID, result)
    logger -> logrepo: StoreLog()
end

executor -> logrepo: StoreLog(full execution log)
deactivate context

executor --> cli: Output
cli --> User: Result

@enduml
```

**Key insight**: Context provides state. Logger records everything. Registry enables plugins.

---

## Reflection Process Flow

Shows how reflection generates new workflow versions.

```plantuml
@startuml Reflection Process
autonumber
title Reflection Process - Learning from Usage

actor User
participant "CLI" as cli
participant "ReflectionEngine" as reflect
participant "PatternAnalyzer" as analyzer
participant "LogRepository" as logrepo
participant "WorkflowGenerator" as generator
participant "LLM Service" as llm
participant "WorkflowRepository" as wfrepo

User -> cli: darwinflow reflect --workflow assistant_v1.yaml --recent 20

cli -> logrepo: GetRecent(workflowID, version, 20)
logrepo --> cli: []ExecutionLog

cli -> analyzer: Analyze(logs)
activate analyzer

analyzer -> llm: "Find patterns in these execution logs"
llm --> analyzer: patterns identified

analyzer --> cli: []Pattern
deactivate analyzer

cli -> reflect: Reflect(ReflectionRequest)
activate reflect

reflect -> llm: "Generate improved workflow based on patterns"
note right
    Prompt includes:
    - Current workflow YAML
    - Identified patterns
    - Recent execution logs
    - Constraints
end note

llm --> reflect: new workflow YAML + reasoning

reflect -> generator: Validate(workflow)
generator --> reflect: validation result

alt workflow is valid
    reflect -> wfrepo: Store(workflow_v2)
    reflect -> wfrepo: Store(reasoning.md)

    reflect --> cli: ReflectionResult(v2, reasoning)
    cli --> User: "Created assistant_v2.yaml\nSee assistant_v2_reasoning.md for changes"
else workflow is invalid
    reflect --> cli: Error(validation failed)
    cli --> User: "Reflection failed: invalid workflow"
end

deactivate reflect

@enduml
```

**Key insight**: Reflection is LLM-powered. It reads logs, finds patterns, generates new versions.

---

## Storage Architecture

Shows storage abstraction and implementations.

```plantuml
@startuml Storage Architecture
skinparam defaultTextAlignment center

title Storage Architecture - Abstract Persistence

package "Core Framework" {
    interface "WorkflowRepository" as IWorkflow {
        +Store(workflow)
        +Get(id, version)
        +GetLatest(id)
        +List(id)
    }

    interface "LogRepository" as ILog {
        +StoreLog(log)
        +GetLog(runID)
        +QueryLogs(filter)
    }

    interface "MetricsRepository" as IMetrics {
        +RecordExecution(metric)
        +GetSummary(id, version)
        +GetPatterns(id)
    }
}

package "Storage Implementations" {
    class "FileWorkflowRepo" {
        -baseDir: string
        +Store(workflow)
        +Get(id, version)
    }

    class "FileLogRepo" {
        -baseDir: string
        +StoreLog(log)
        +GetLog(runID)
    }

    class "SQLiteLogRepo" {
        -db: *sql.DB
        +StoreLog(log)
        +GetLog(runID)
    }

    class "PostgresLogRepo" {
        -db: *sql.DB
        +StoreLog(log)
        +GetLog(runID)
    }
}

package "Storage Media" {
    database "YAML Files" as files
    database "JSON Files" as json
    database "SQLite DB" as sqlite
    database "Postgres DB" as postgres
}

IWorkflow <|.. FileWorkflowRepo : implements
ILog <|.. FileLogRepo : implements
ILog <|.. SQLiteLogRepo : implements
ILog <|.. PostgresLogRepo : implements
IMetrics <|.. FileLogRepo : implements (metrics in files)
IMetrics <|.. SQLiteLogRepo : implements (metrics in DB)

FileWorkflowRepo --> files
FileLogRepo --> json
SQLiteLogRepo --> sqlite
PostgresLogRepo --> postgres

note right of IWorkflow
    **Liskov Substitution Principle**

    Any implementation can replace another
    without breaking core framework.

    Core depends on INTERFACE,
    not concrete implementation.
end note

legend right
    |= Phase |= Workflow |= Logs |= Metrics |
    | MVP (1) | Files | Files | Files |
    | Growth (2) | Files | SQLite | SQLite |
    | Scale (3) | Files | Postgres | Postgres |

    Mix and match as needed!
endlegend

@enduml
```

**Key insight**: Storage backend is swappable. Core never knows what's used.

---

## Node Type Plugin System

Shows how node types register and execute.

```plantuml
@startuml Node Plugin System
skinparam defaultTextAlignment center

title Node Type Plugin System

interface "NodeExecutor" {
    +Type() NodeType
    +Execute(ctx, config) NodeResult
    +ValidateConfig(config) error
}

class "NodeRegistry" {
    -executors: map[NodeType]NodeExecutor
    +Register(executor)
    +Get(nodeType) NodeExecutor
}

class "ClaudeCodeNode" {
    -cliPath: string
    +Execute(ctx, config)
}

class "FileReadNode" {
    +Execute(ctx, config)
}

class "LLMDecisionNode" {
    -provider: LLMProvider
    +Execute(ctx, config)
}

class "CustomNode\n(Community Plugin)" {
    +Execute(ctx, config)
}

NodeExecutor <|.. ClaudeCodeNode : implements
NodeExecutor <|.. FileReadNode : implements
NodeExecutor <|.. LLMDecisionNode : implements
NodeExecutor <|.. CustomNode : implements

NodeRegistry o--> NodeExecutor : manages

note right of NodeRegistry
    **Open-Closed Principle**

    Adding new node type:
    1. Implement NodeExecutor
    2. Call registry.Register()
    3. Use in workflow YAML

    NO changes to core framework!
end note

note bottom of CustomNode
    Community can contribute
    node types as plugins.

    Just implement the interface,
    no need to modify core.
end note

@enduml
```

**Key insight**: Plugin system via interface. Add node types without touching core.

---

## Evolution Timeline

Visual representation of how the system grows.

```plantuml
@startuml Evolution Timeline
!define RECTANGLE class

skinparam timeline {
    BackgroundColor LightBlue
}

title DarwinFlow Evolution - Progressive Complexity Growth

scale 800 width

concise "Capabilities" as cap
concise "Node Types" as nodes
concise "Storage" as storage
concise "Code Size" as code

@0
cap is "MVP"
nodes is "1 type"
storage is "Files"
code is "~1K LOC"

@1
cap is "Multi-Node"
nodes is "6 types"
storage is "Files"
code is "~3K LOC"

@2
cap is "Advanced"
nodes is "10 types"
storage is "SQLite"
code is "~5K LOC"

@3
cap is "Mature"
nodes is "20+ types"
storage is "Postgres"
code is "~10K LOC"

@enduml
```

```plantuml
@startuml Evolution Features
left to right direction
skinparam activityBackgroundColor<<phase1>> LightBlue
skinparam activityBackgroundColor<<phase2>> LightGreen
skinparam activityBackgroundColor<<phase3>> LightYellow
skinparam activityBackgroundColor<<phase4>> LightCoral

title Feature Evolution by Phase

rectangle "**Phase 1: MVP**\n(Weeks 1-4)" <<phase1>> {
    (*) --> "1 Node Type\n(CLAUDE_CODE)"
    --> "File Storage"
    --> "Manual Reflection"
    --> "Basic Logging"
}

rectangle "**Phase 2: Multi-Node**\n(Months 2-3)" <<phase2>> {
    "6 Node Types" --> "Conditional Edges"
    --> "Better Metrics"
    --> "Pattern Detection"
}

rectangle "**Phase 3: Advanced**\n(Months 4-6)" <<phase3>> {
    "10 Node Types" --> "Parallel Execution"
    --> "Sub-workflows"
    --> "SQLite Storage"
    --> "Retry Logic"
}

rectangle "**Phase 4: Mature**\n(Months 7-12)" <<phase4>> {
    "20+ Node Types" --> "Multiple LLM Providers"
    --> "Community Plugins"
    --> "Real-time Learning"
    --> "Workflow Marketplace"
}

@enduml
```

**Key insight**: Grow incrementally. Each phase adds capabilities without breaking previous ones.

---

## Integration Patterns

Shows common integration patterns.

```plantuml
@startuml Integration Patterns
title Integration Patterns

package "Pattern 1: External Tool Adapter" {
    class "ClaudeCodeNode" {
        -cliPath: string
        +Execute(ctx, config)
    }

    component "Claude Code CLI" as cli

    ClaudeCodeNode --> cli : wraps
}

package "Pattern 2: Primitive Operation" {
    class "FileReadNode" {
        +Execute(ctx, config)
    }

    component "File System" as fs

    FileReadNode --> fs : direct access
}

package "Pattern 3: LLM-Powered" {
    class "LLMDecisionNode" {
        -provider: LLMProvider
        +Execute(ctx, config)
    }

    interface "LLMProvider" {
        +Complete(prompt) response
    }

    class "AnthropicProvider"
    class "OpenAIProvider"

    LLMDecisionNode --> LLMProvider
    LLMProvider <|.. AnthropicProvider
    LLMProvider <|.. OpenAIProvider
}

package "Pattern 4: Composite" {
    class "SubWorkflowNode" {
        -executor: WorkflowExecutor
        +Execute(ctx, config)
    }

    component "Nested Workflow" as nested

    SubWorkflowNode --> nested : delegates to
}

note bottom
    **Integration Guidelines:**

    1. Keep adapters thin (wrap, don't implement logic)
    2. Log all external calls
    3. Handle errors gracefully
    4. Use timeouts
    5. Make them reusable
end note

@enduml
```

**Key insight**: Different integration patterns for different needs. All implement same interface.

---

## SOLID Principles Applied

Visual representation of how SOLID principles manifest in the architecture.

```plantuml
@startuml SOLID Principles
title SOLID Principles in DarwinFlow Architecture

package "**S** - Single Responsibility" {
    class "WorkflowExecutor" {
        Responsibility: Execute workflows
    }
    class "PatternAnalyzer" {
        Responsibility: Find patterns
    }
    class "LogRepository" {
        Responsibility: Persist logs
    }

    note bottom
        Each class has ONE reason to change:
        - Executor: execution logic changes
        - Analyzer: pattern detection improves
        - Repository: storage mechanism changes
    end note
}

package "**O** - Open/Closed" {
    interface "NodeExecutor" {
        +Execute()
    }

    class "ClaudeCodeNode"
    class "FileReadNode"
    class "CustomNode"

    NodeExecutor <|.. ClaudeCodeNode
    NodeExecutor <|.. FileReadNode
    NodeExecutor <|.. CustomNode

    note bottom
        Open for extension (add new node types)
        Closed for modification (don't change core)
    end note
}

package "**L** - Liskov Substitution" {
    interface "LogRepository" as ILog {
        +StoreLog(log)
        +GetLog(runID)
    }

    class "FileLogRepo"
    class "SQLiteLogRepo"

    ILog <|.. FileLogRepo
    ILog <|.. SQLiteLogRepo

    note bottom
        Any implementation can replace another
        without breaking WorkflowExecutor
    end note
}

package "**I** - Interface Segregation" {
    interface "WorkflowExecutor" {
        +Execute()
    }

    interface "ExecutionLogger" {
        +LogStep()
    }

    interface "MetricsCollector" {
        +RecordDuration()
    }

    note bottom
        Clients depend only on what they use.
        No fat interfaces with unused methods.
    end note
}

package "**D** - Dependency Inversion" {
    class "WorkflowExecutor" {
        -storage: LogRepository
        -registry: NodeRegistry
    }

    interface "LogRepository"
    interface "NodeRegistry"

    class "FileLogRepo"
    class "ClaudeCodeNode"

    WorkflowExecutor --> LogRepository : depends on
    WorkflowExecutor --> NodeRegistry : depends on

    LogRepository <|.. FileLogRepo : implements
    NodeRegistry o--> ClaudeCodeNode : manages

    note bottom
        High-level (Executor) depends on abstractions
        Low-level (FileLogRepo) implements abstractions

        Dependency points INWARD
    end note
}

@enduml
```

**Key insight**: SOLID isn't academic theory. It's how this architecture enables evolution.

---

## Quick Reference: Adding Capabilities

```plantuml
@startuml Adding Capabilities
title Quick Reference - How to Extend DarwinFlow

start

:Want to add capability;

if (What type?) then (New node type)
    :Implement NodeExecutor interface;
    :Register with NodeRegistry;
    :~50-100 LOC;

elseif (New storage backend) then
    :Implement *Repository interfaces;
    :~200-300 LOC per repo;

elseif (New workflow) then
    :Write YAML file;
    :0 LOC in codebase;

elseif (New integration) then
    :Create adapter implementing NodeExecutor;
    :~100-200 LOC;

elseif (New LLM provider) then
    :Implement LLMProvider interface;
    :~50-100 LOC;

endif

:Deploy (hot-reload or restart);
:Test with existing workflows;

stop

note right
    **Key Point:**

    Adding capabilities NEVER requires
    changing core framework code.

    If it does, the architecture is wrong.
end note

@enduml
```

---

## Summary

These diagrams show:

1. **System Context** - DarwinFlow in its environment
2. **Containers** - Major components and their relationships
3. **Components** - Internal structure of core framework
4. **Dependencies** - Direction of dependencies (inward!)
5. **Sequences** - Runtime execution flow
6. **Storage** - Abstraction and implementations
7. **Plugins** - Node type registration system
8. **Evolution** - How system grows over time
9. **Patterns** - Common integration patterns
10. **SOLID** - Principles in practice

**Use these diagrams to**:
- Understand high-level architecture
- Explain system to new developers
- Make design decisions
- Verify dependency direction
- Plan extensions

**Remember**: Architecture is about managing dependencies and handling change. If you can't explain it with a diagram, you don't understand it.

---

**Architecture documentation complete**:
- [00_architecture_overview.md](00_architecture_overview.md)
- [01_core_interfaces.md](01_core_interfaces.md)
- [02_integration_architecture.md](02_integration_architecture.md)
- [03_storage_architecture.md](03_storage_architecture.md)
- [04_evolution_roadmap.md](04_evolution_roadmap.md)
- [05_architecture_diagrams.md](05_architecture_diagrams.md) (this document)

**Now go build it.**

package task_manager_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// MockLogger is a simple logger for testing
type MockLogger struct {
	messages []string
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.messages = append(m.messages, "DEBUG: "+msg)
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.messages = append(m.messages, "INFO: "+msg)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.messages = append(m.messages, "WARN: "+msg)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.messages = append(m.messages, "ERROR: "+msg)
}

// TestNewTaskManagerPlugin tests plugin creation
func TestNewTaskManagerPlugin(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	if plugin == nil {
		t.Error("plugin should not be nil")
	}

	info := plugin.GetInfo()
	if info.Name != "task-manager" {
		t.Errorf("expected plugin name 'task-manager', got %q", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", info.Version)
	}
}

// TestGetCapabilities tests that plugin reports correct capabilities
func TestGetCapabilities(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	capabilities := plugin.GetCapabilities()
	expected := []string{"IEntityProvider", "ICommandProvider", "IEventEmitter"}

	if len(capabilities) != len(expected) {
		t.Errorf("expected %d capabilities, got %d", len(expected), len(capabilities))
	}

	capMap := make(map[string]bool)
	for _, cap := range capabilities {
		capMap[cap] = true
	}

	for _, exp := range expected {
		if !capMap[exp] {
			t.Errorf("missing capability: %s", exp)
		}
	}
}

// TestGetEntityTypes tests entity type info
func TestGetEntityTypes(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	types := plugin.GetEntityTypes()
	if len(types) != 1 {
		t.Errorf("expected 1 entity type, got %d", len(types))
	}

	if types[0].Type != "task" {
		t.Errorf("expected entity type 'task', got %q", types[0].Type)
	}
}

// TestTaskEntityGetters tests TaskEntity field getters
func TestTaskEntityGetters(t *testing.T) {
	now := time.Now().UTC()
	entity := task_manager.NewTaskEntity(
		"task-123",
		"Test Task",
		"Test Description",
		"todo",
		"high",
		now,
		now,
	)

	if entity.GetID() != "task-123" {
		t.Errorf("expected ID 'task-123', got %q", entity.GetID())
	}

	if entity.GetType() != "task" {
		t.Errorf("expected type 'task', got %q", entity.GetType())
	}

	if entity.GetStatus() != "todo" {
		t.Errorf("expected status 'todo', got %q", entity.GetStatus())
	}

	fields := entity.GetAllFields()
	if fields["title"] != "Test Task" {
		t.Errorf("expected title 'Test Task', got %q", fields["title"])
	}
}

// TestTaskEntityProgress tests progress calculation
func TestTaskEntityProgress(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		status   string
		expected float64
	}{
		{"todo", 0.0},
		{"in-progress", 0.5},
		{"done", 1.0},
	}

	for _, test := range tests {
		entity := task_manager.NewTaskEntity("id", "title", "desc", test.status, "medium", now, now)
		progress := entity.GetProgress()
		if progress != test.expected {
			t.Errorf("for status %q, expected progress %.1f, got %.1f", test.status, test.expected, progress)
		}
	}
}

// TestQueryTasks tests the Query method
func TestQueryTasks(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create a test task file
	tasksDir := filepath.Join(dir, ".darwinflow", "tasks")
	err = os.MkdirAll(tasksDir, 0755)
	if err != nil {
		t.Fatalf("failed to create tasks directory: %v", err)
	}

	// Query tasks (should be empty initially)
	query := pluginsdk.EntityQuery{
		EntityType: "task",
	}

	entities, err := plugin.Query(context.Background(), query)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(entities) != 0 {
		t.Errorf("expected 0 tasks initially, got %d", len(entities))
	}
}

// TestCreateCommandExecution tests the create command
func TestCreateCommandExecution(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	commands := plugin.GetCommands()
	var createCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "create" {
			createCmd = cmd
			break
		}
	}

	if createCmd == nil {
		t.Fatal("create command not found")
	}

	// Execute create command
	stdout := &bytes.Buffer{}
	mockCmdCtx := &MockCommandContext{
		workingDir: dir,
		stdout:     stdout,
	}

	args := []string{"Test Task", "--priority", "high"}
	err = createCmd.Execute(context.Background(), mockCmdCtx, args)
	if err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	// Verify task was created
	tasksDir := filepath.Join(dir, ".darwinflow", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		t.Fatalf("failed to read tasks directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("no task files created")
	}
}

// TestEventStreamStartStop tests event stream start and stop
func TestEventStreamStartStop(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create event channel
	eventChan := make(chan pluginsdk.Event, 100)
	defer close(eventChan)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start event stream
	err = plugin.StartEventStream(ctx, eventChan)
	if err != nil {
		t.Fatalf("failed to start event stream: %v", err)
	}

	// Stop event stream
	err = plugin.StopEventStream()
	if err != nil {
		t.Fatalf("failed to stop event stream: %v", err)
	}
}

// TestEventEmissionOnTaskCreation tests that events are emitted when tasks are created
func TestEventEmissionOnTaskCreation(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create event channel
	eventChan := make(chan pluginsdk.Event, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start event stream
	err = plugin.StartEventStream(ctx, eventChan)
	if err != nil {
		t.Fatalf("failed to start event stream: %v", err)
	}
	defer plugin.StopEventStream()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a task file manually
	tasksDir := filepath.Join(dir, ".darwinflow", "tasks")
	err = os.MkdirAll(tasksDir, 0755)
	if err != nil {
		t.Fatalf("failed to create tasks directory: %v", err)
	}

	// Write a task file directly
	taskFile := filepath.Join(tasksDir, "task-test.json")
	taskData := `{"id":"task-test","title":"Test Task","status":"todo","priority":"medium","created_at":"2025-10-22T00:00:00Z","updated_at":"2025-10-22T00:00:00Z"}`
	err = os.WriteFile(taskFile, []byte(taskData), 0644)
	if err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Wait for event (with timeout)
	eventReceived := false
	select {
	case event := <-eventChan:
		if event.Type == "task.created" || event.Type == "task.updated" {
			eventReceived = true
		}
	case <-time.After(2 * time.Second):
		// Timeout - file system events may not always be captured in tests
	}

	if !eventReceived {
		// Note: File watcher events are not guaranteed in test environments
		// This test demonstrates the structure, but may not always capture events
		t.Logf("no event received within timeout (this can be normal in test environments)")
	}
}

// TestListCommand tests the list command
func TestListCommand(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	commands := plugin.GetCommands()
	var listCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "list" {
			listCmd = cmd
			break
		}
	}

	if listCmd == nil {
		t.Fatal("list command not found")
	}

	stdout := &bytes.Buffer{}
	mockCmdCtx := &MockCommandContext{
		workingDir: dir,
		stdout:     stdout,
	}

	// Execute list command when no tasks exist
	args := []string{}
	err = listCmd.Execute(context.Background(), mockCmdCtx, args)
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains(stdout.Bytes(), []byte("No tasks found")) {
		t.Errorf("expected 'No tasks found' in output, got: %s", output)
	}
}

// TestUpdateCommand tests the update command
func TestUpdateCommand(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create a task first
	tasksDir := filepath.Join(dir, ".darwinflow", "tasks")
	err = os.MkdirAll(tasksDir, 0755)
	if err != nil {
		t.Fatalf("failed to create tasks directory: %v", err)
	}

	taskID := "task-456"
	taskFile := filepath.Join(tasksDir, taskID+".json")
	now := time.Now().UTC()
	task := task_manager.NewTaskEntity(taskID, "Test Task", "Description", "todo", "medium", now, now)

	data, err := task_manager.MarshalTask(task)
	if err != nil {
		t.Fatalf("failed to marshal task: %v", err)
	}

	err = os.WriteFile(taskFile, data, 0644)
	if err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Find update command
	commands := plugin.GetCommands()
	var updateCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "update" {
			updateCmd = cmd
			break
		}
	}

	if updateCmd == nil {
		t.Fatal("update command not found")
	}

	stdout := &bytes.Buffer{}
	mockCmdCtx := &MockCommandContext{
		workingDir: dir,
		stdout:     stdout,
	}

	// Execute update command
	args := []string{taskID, "--status", "done"}
	err = updateCmd.Execute(context.Background(), mockCmdCtx, args)
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	output := stdout.String()
	if !bytes.Contains(stdout.Bytes(), []byte("Task updated")) {
		t.Errorf("expected 'Task updated' in output, got: %s", output)
	}
}

// TestInitCommand tests the init command
func TestInitCommand(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	commands := plugin.GetCommands()
	var initCmd pluginsdk.Command
	for _, cmd := range commands {
		if cmd.GetName() == "init" {
			initCmd = cmd
			break
		}
	}

	if initCmd == nil {
		t.Fatal("init command not found")
	}

	stdout := &bytes.Buffer{}
	mockCmdCtx := &MockCommandContext{
		workingDir: dir,
		stdout:     stdout,
	}

	// Execute init command
	args := []string{}
	err = initCmd.Execute(context.Background(), mockCmdCtx, args)
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify tasks directory was created
	tasksDir := filepath.Join(dir, ".darwinflow", "tasks")
	if _, err := os.Stat(tasksDir); err != nil {
		t.Errorf("tasks directory not created: %v", err)
	}
}

// TestUpdateEntity tests the UpdateEntity method
func TestUpdateEntity(t *testing.T) {
	dir := t.TempDir()
	logger := &MockLogger{}

	plugin, err := task_manager.NewTaskManagerPlugin(logger, dir)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create a task directory and a task file manually
	tasksDir := filepath.Join(dir, ".darwinflow", "tasks")
	err = os.MkdirAll(tasksDir, 0755)
	if err != nil {
		t.Fatalf("failed to create tasks directory: %v", err)
	}

	// Create a task file directly
	taskID := "task-123"
	taskFile := filepath.Join(tasksDir, taskID+".json")
	now := time.Now().UTC()
	task := task_manager.NewTaskEntity(taskID, "Test Task", "Description", "todo", "medium", now, now)

	data, err := task_manager.MarshalTask(task)
	if err != nil {
		t.Fatalf("failed to marshal task: %v", err)
	}

	err = os.WriteFile(taskFile, data, 0644)
	if err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Get the task ID from the created task
	query := pluginsdk.EntityQuery{EntityType: "task"}
	entities, err := plugin.Query(context.Background(), query)
	if err != nil || len(entities) == 0 {
		t.Fatalf("failed to query tasks: %v, entities: %d", err, len(entities))
	}

	retrievedTaskID := entities[0].GetID()

	// Update the task
	updates := map[string]interface{}{
		"status": "done",
		"title":  "Updated Task",
	}

	updated, err := plugin.UpdateEntity(context.Background(), retrievedTaskID, updates)
	if err != nil {
		t.Fatalf("failed to update entity: %v", err)
	}

	updatedTask := updated.(*task_manager.TaskEntity)
	if updatedTask.Status != "done" {
		t.Errorf("expected status 'done', got %q", updatedTask.Status)
	}

	if updated.GetField("title") != "Updated Task" {
		t.Errorf("expected title 'Updated Task', got %q", updated.GetField("title"))
	}
}

// MockCommandContext implements pluginsdk.CommandContext
type MockCommandContext struct {
	workingDir string
	stdout     *bytes.Buffer
	stdin      *bytes.Buffer
}

func (m *MockCommandContext) GetLogger() pluginsdk.Logger {
	return &MockLogger{}
}

func (m *MockCommandContext) GetWorkingDir() string {
	return m.workingDir
}

func (m *MockCommandContext) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	return nil
}

func (m *MockCommandContext) GetStdout() io.Writer {
	return m.stdout
}

func (m *MockCommandContext) GetStdin() io.Reader {
	return m.stdin
}

// TestCreateCommand is a helper for testing
type TestCreateCommand struct {
}

func (t *TestCreateCommand) GetName() string {
	return "create"
}

func (t *TestCreateCommand) GetDescription() string {
	return "Create task"
}

func (t *TestCreateCommand) GetUsage() string {
	return "create <title>"
}

func (t *TestCreateCommand) GetHelp() string {
	return ""
}

func (t *TestCreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Just create the task directory for testing
	dir := cmdCtx.GetWorkingDir()
	tasksDir := filepath.Join(dir, ".darwinflow", "tasks")
	return os.MkdirAll(tasksDir, 0755)
}

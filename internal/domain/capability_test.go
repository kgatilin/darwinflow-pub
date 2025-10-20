package domain_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

func TestEntityContext(t *testing.T) {
	// Test creating and using EntityContext
	ctx := &domain.EntityContext{
		RelatedEntities: map[string][]string{
			"task":    {"task-1", "task-2"},
			"session": {"session-1"},
		},
		LinkedFiles: []string{"/path/to/file1.go", "/path/to/file2.go"},
		RecentActivity: []domain.ActivityRecord{
			{
				Timestamp:   time.Now(),
				Type:        "created",
				Description: "test",
				Actor:       "user",
			},
		},
		Metadata: map[string]interface{}{
			"priority": "high",
			"tags":     []string{"urgent", "feature"},
		},
	}

	// Verify RelatedEntities
	if len(ctx.RelatedEntities) != 2 {
		t.Errorf("Expected 2 related entity types, got %d", len(ctx.RelatedEntities))
	}
	if len(ctx.RelatedEntities["task"]) != 2 {
		t.Errorf("Expected 2 related tasks, got %d", len(ctx.RelatedEntities["task"]))
	}

	// Verify LinkedFiles
	if len(ctx.LinkedFiles) != 2 {
		t.Errorf("Expected 2 linked files, got %d", len(ctx.LinkedFiles))
	}
	if ctx.LinkedFiles[0] != "/path/to/file1.go" {
		t.Error("LinkedFiles[0] mismatch")
	}

	// Verify RecentActivity
	if len(ctx.RecentActivity) != 1 {
		t.Errorf("Expected 1 activity record, got %d", len(ctx.RecentActivity))
	}
	if ctx.RecentActivity[0].Type != "created" {
		t.Error("Activity action mismatch")
	}

	// Verify Metadata
	if len(ctx.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(ctx.Metadata))
	}
	if ctx.Metadata["priority"] != "high" {
		t.Error("Metadata priority mismatch")
	}
}

func TestActivityRecord(t *testing.T) {
	now := time.Now()
	record := domain.ActivityRecord{
		Timestamp:   now,
		Type:        "updated",
		Description: "status changed from pending to completed",
		Actor:       "user",
	}

	// Verify fields
	if record.Timestamp != now {
		t.Error("Timestamp mismatch")
	}
	if record.Type != "updated" {
		t.Error("Type mismatch")
	}
	if record.Description == "" {
		t.Error("Description should not be empty")
	}
	if record.Actor != "user" {
		t.Error("Actor mismatch")
	}
}

func TestEntityContext_EmptyFields(t *testing.T) {
	// Test creating EntityContext with nil/empty fields
	ctx := &domain.EntityContext{}

	if ctx.RelatedEntities != nil {
		t.Error("Expected nil RelatedEntities, got non-nil")
	}
	if ctx.LinkedFiles != nil {
		t.Error("Expected nil LinkedFiles, got non-nil")
	}
	if ctx.RecentActivity != nil {
		t.Error("Expected nil RecentActivity, got non-nil")
	}
	if ctx.Metadata != nil {
		t.Error("Expected nil Metadata, got non-nil")
	}
}

func TestActivityRecord_EmptyDetails(t *testing.T) {
	// Test creating ActivityRecord with empty description
	record := domain.ActivityRecord{
		Timestamp:   time.Now(),
		Type:        "action",
		Description: "",
		Actor:       "",
	}

	if record.Type != "action" {
		t.Error("Type should be preserved")
	}
	if record.Description != "" {
		t.Error("Description should be empty string")
	}
}

func TestEntityContext_AddRelatedEntity(t *testing.T) {
	// Test adding related entities dynamically
	ctx := &domain.EntityContext{
		RelatedEntities: make(map[string][]string),
	}

	// Add first entity type
	ctx.RelatedEntities["task"] = []string{"task-1"}

	// Add another entity of same type
	ctx.RelatedEntities["task"] = append(ctx.RelatedEntities["task"], "task-2")

	// Add different entity type
	ctx.RelatedEntities["session"] = []string{"session-1"}

	if len(ctx.RelatedEntities) != 2 {
		t.Errorf("Expected 2 entity types, got %d", len(ctx.RelatedEntities))
	}
	if len(ctx.RelatedEntities["task"]) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(ctx.RelatedEntities["task"]))
	}
}

func TestEntityContext_AddLinkedFile(t *testing.T) {
	// Test adding linked files dynamically
	ctx := &domain.EntityContext{
		LinkedFiles: []string{},
	}

	ctx.LinkedFiles = append(ctx.LinkedFiles, "/file1.go")
	ctx.LinkedFiles = append(ctx.LinkedFiles, "/file2.go")

	if len(ctx.LinkedFiles) != 2 {
		t.Errorf("Expected 2 files, got %d", len(ctx.LinkedFiles))
	}
	if ctx.LinkedFiles[1] != "/file2.go" {
		t.Error("LinkedFiles[1] mismatch")
	}
}

func TestEntityContext_AddActivity(t *testing.T) {
	// Test adding activity records dynamically
	ctx := &domain.EntityContext{
		RecentActivity: []domain.ActivityRecord{},
	}

	ctx.RecentActivity = append(ctx.RecentActivity, domain.ActivityRecord{
		Timestamp:   time.Now(),
		Type:        "created",
		Description: "entity created",
		Actor:       "system",
	})

	ctx.RecentActivity = append(ctx.RecentActivity, domain.ActivityRecord{
		Timestamp:   time.Now(),
		Type:        "updated",
		Description: "status field updated",
		Actor:       "user",
	})

	if len(ctx.RecentActivity) != 2 {
		t.Errorf("Expected 2 activity records, got %d", len(ctx.RecentActivity))
	}
	if ctx.RecentActivity[0].Type != "created" {
		t.Error("First activity action mismatch")
	}
	if ctx.RecentActivity[1].Type != "updated" {
		t.Error("Second activity action mismatch")
	}
}

func TestEntityContext_Metadata(t *testing.T) {
	// Test metadata can store various types
	ctx := &domain.EntityContext{
		Metadata: map[string]interface{}{
			"string":  "value",
			"int":     42,
			"bool":    true,
			"float":   3.14,
			"array":   []string{"a", "b", "c"},
			"map":     map[string]int{"x": 1, "y": 2},
		},
	}

	if ctx.Metadata["string"] != "value" {
		t.Error("String metadata mismatch")
	}
	if ctx.Metadata["int"] != 42 {
		t.Error("Int metadata mismatch")
	}
	if ctx.Metadata["bool"] != true {
		t.Error("Bool metadata mismatch")
	}
	if ctx.Metadata["float"] != 3.14 {
		t.Error("Float metadata mismatch")
	}

	// Verify array type
	arr, ok := ctx.Metadata["array"].([]string)
	if !ok || len(arr) != 3 {
		t.Error("Array metadata type mismatch")
	}

	// Verify map type
	m, ok := ctx.Metadata["map"].(map[string]int)
	if !ok || len(m) != 2 {
		t.Error("Map metadata type mismatch")
	}
}

// Mock implementation of IExtensible for testing
type mockExtensible struct {
	id           string
	entityType   string
	capabilities []string
	fields       map[string]interface{}
}

func (m *mockExtensible) GetID() string {
	return m.id
}

func (m *mockExtensible) GetType() string {
	return m.entityType
}

func (m *mockExtensible) GetCapabilities() []string {
	return m.capabilities
}

func (m *mockExtensible) GetField(name string) interface{} {
	return m.fields[name]
}

func (m *mockExtensible) GetAllFields() map[string]interface{} {
	return m.fields
}

func TestIExtensible_MockImplementation(t *testing.T) {
	// Test that the interface can be implemented
	mock := &mockExtensible{
		id:           "test-id",
		entityType:   "test-type",
		capabilities: []string{"IExtensible"},
		fields:       map[string]interface{}{"name": "test", "value": 42},
	}

	var entity domain.IExtensible = mock

	if entity.GetID() != "test-id" {
		t.Error("GetID mismatch")
	}
	if entity.GetType() != "test-type" {
		t.Error("GetType mismatch")
	}
	if len(entity.GetCapabilities()) != 1 {
		t.Error("GetCapabilities mismatch")
	}
	if entity.GetField("name") != "test" {
		t.Error("GetField mismatch")
	}
	if len(entity.GetAllFields()) != 2 {
		t.Error("GetAllFields mismatch")
	}
}

// Mock implementation of IHasContext for testing
type mockHasContext struct {
	mockExtensible
	context *domain.EntityContext
}

func (m *mockHasContext) GetContext() *domain.EntityContext {
	return m.context
}

func TestIHasContext_MockImplementation(t *testing.T) {
	// Test that the interface can be implemented
	mock := &mockHasContext{
		mockExtensible: mockExtensible{
			id:           "ctx-id",
			entityType:   "ctx-type",
			capabilities: []string{"IExtensible", "IHasContext"},
			fields:       map[string]interface{}{},
		},
		context: &domain.EntityContext{
			LinkedFiles: []string{"/file.go"},
		},
	}

	var entity domain.IHasContext = mock

	if entity.GetID() != "ctx-id" {
		t.Error("GetID mismatch")
	}
	if entity.GetContext() == nil {
		t.Fatal("GetContext returned nil")
	}
	if len(entity.GetContext().LinkedFiles) != 1 {
		t.Error("Context LinkedFiles mismatch")
	}
}

// Mock implementation of ITrackable for testing
type mockTrackable struct {
	mockExtensible
	status      string
	progress    float64
	blocked     bool
	blockReason string
}

func (m *mockTrackable) GetStatus() string {
	return m.status
}

func (m *mockTrackable) GetProgress() float64 {
	return m.progress
}

func (m *mockTrackable) IsBlocked() bool {
	return m.blocked
}

func (m *mockTrackable) GetBlockReason() string {
	return m.blockReason
}

func TestITrackable_MockImplementation(t *testing.T) {
	mock := &mockTrackable{
		mockExtensible: mockExtensible{
			id:           "track-id",
			entityType:   "task",
			capabilities: []string{"IExtensible", "ITrackable"},
		},
		status:      "in_progress",
		progress:    0.75,
		blocked:     true,
		blockReason: "waiting for review",
	}

	var entity domain.ITrackable = mock

	if entity.GetStatus() != "in_progress" {
		t.Error("GetStatus mismatch")
	}
	if entity.GetProgress() != 0.75 {
		t.Error("GetProgress mismatch")
	}
	if !entity.IsBlocked() {
		t.Error("IsBlocked should be true")
	}
	if entity.GetBlockReason() != "waiting for review" {
		t.Error("GetBlockReason mismatch")
	}
}

// Mock implementation of ISchedulable for testing
type mockSchedulable struct {
	mockExtensible
	startDate *time.Time
	dueDate   *time.Time
}

func (m *mockSchedulable) GetStartDate() *time.Time {
	return m.startDate
}

func (m *mockSchedulable) GetDueDate() *time.Time {
	return m.dueDate
}

func (m *mockSchedulable) IsOverdue() bool {
	if m.dueDate == nil {
		return false
	}
	return time.Now().After(*m.dueDate)
}

func TestISchedulable_MockImplementation(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	// Test not overdue
	mock := &mockSchedulable{
		mockExtensible: mockExtensible{
			id:           "sched-id",
			entityType:   "task",
			capabilities: []string{"IExtensible", "ISchedulable"},
		},
		startDate: &now,
		dueDate:   &future,
	}

	var entity domain.ISchedulable = mock

	if entity.GetStartDate() == nil {
		t.Error("GetStartDate should not be nil")
	}
	if entity.GetDueDate() == nil {
		t.Error("GetDueDate should not be nil")
	}
	if entity.IsOverdue() {
		t.Error("Should not be overdue")
	}

	// Test overdue
	mock.dueDate = &past
	if !entity.IsOverdue() {
		t.Error("Should be overdue")
	}
}

// Mock implementation of IRelatable for testing
type mockRelatable struct {
	mockExtensible
	relations map[string][]string
}

func (m *mockRelatable) GetRelated(entityType string) []string {
	return m.relations[entityType]
}

func (m *mockRelatable) GetAllRelations() map[string][]string {
	return m.relations
}

func TestIRelatable_MockImplementation(t *testing.T) {
	mock := &mockRelatable{
		mockExtensible: mockExtensible{
			id:           "rel-id",
			entityType:   "project",
			capabilities: []string{"IExtensible", "IRelatable"},
		},
		relations: map[string][]string{
			"task":     {"task-1", "task-2"},
			"document": {"doc-1"},
		},
	}

	var entity domain.IRelatable = mock

	tasks := entity.GetRelated("task")
	if len(tasks) != 2 {
		t.Errorf("Expected 2 related tasks, got %d", len(tasks))
	}

	docs := entity.GetRelated("document")
	if len(docs) != 1 {
		t.Errorf("Expected 1 related document, got %d", len(docs))
	}

	all := entity.GetAllRelations()
	if len(all) != 2 {
		t.Errorf("Expected 2 relation types, got %d", len(all))
	}
}

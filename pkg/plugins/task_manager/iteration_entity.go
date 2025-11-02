package task_manager

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// IterationEntity represents a time-boxed grouping of tasks and implements SDK capability interfaces.
// It implements the IExtensible interface.
type IterationEntity struct {
	Number      int        `json:"number"`
	Name        string     `json:"name"`
	Goal        string     `json:"goal"`
	TaskIDs     []string   `json:"task_ids"`
	Status      string     `json:"status"` // planned, current, complete
	Rank        int        `json:"rank"` // 1-1000 (lower = higher priority)
	Deliverable string     `json:"deliverable"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Valid status values for iterations
var validIterationStatuses = map[string]bool{
	"planned":  true,
	"current":  true,
	"complete": true,
}

// NewIterationEntity creates a new iteration entity with validation
func NewIterationEntity(number int, name, goal, deliverable string, taskIDs []string, status string, rank int, startedAt, completedAt, createdAt, updatedAt time.Time) (*IterationEntity, error) {
	// Validate number is positive
	if number <= 0 {
		return nil, fmt.Errorf("%w: iteration number must be positive", pluginsdk.ErrInvalidArgument)
	}

	// Validate status
	if !validIterationStatuses[status] {
		return nil, fmt.Errorf("%w: invalid iteration status: must be one of planned, current, complete", pluginsdk.ErrInvalidArgument)
	}

	// Validate rank
	if rank < 1 || rank > 1000 {
		return nil, fmt.Errorf("%w: invalid iteration rank: must be between 1 and 1000", pluginsdk.ErrInvalidArgument)
	}

	if taskIDs == nil {
		taskIDs = []string{}
	}

	var startedAtPtr, completedAtPtr *time.Time
	if !startedAt.IsZero() {
		startedAtPtr = &startedAt
	}
	if !completedAt.IsZero() {
		completedAtPtr = &completedAt
	}

	return &IterationEntity{
		Number:      number,
		Name:        name,
		Goal:        goal,
		TaskIDs:     taskIDs,
		Status:      status,
		Rank:        rank,
		Deliverable: deliverable,
		StartedAt:   startedAtPtr,
		CompletedAt: completedAtPtr,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// IExtensible implementation

// GetID returns the unique identifier for this entity
// For iterations, the ID is the number converted to string
func (i *IterationEntity) GetID() string {
	return strconv.Itoa(i.Number)
}

// GetType returns the entity type
func (i *IterationEntity) GetType() string {
	return "iteration"
}

// GetCapabilities returns list of capability names this entity supports
func (i *IterationEntity) GetCapabilities() []string {
	return []string{"IExtensible"}
}

// GetField retrieves a named field value
func (i *IterationEntity) GetField(name string) interface{} {
	fields := i.GetAllFields()
	return fields[name]
}

// GetAllFields returns all fields as a map
func (i *IterationEntity) GetAllFields() map[string]interface{} {
	return map[string]interface{}{
		"number":       i.Number,
		"name":         i.Name,
		"goal":         i.Goal,
		"task_ids":     i.TaskIDs,
		"status":       i.Status,
		"rank":         i.Rank,
		"deliverable":  i.Deliverable,
		"started_at":   i.StartedAt,
		"completed_at": i.CompletedAt,
		"created_at":   i.CreatedAt,
		"updated_at":   i.UpdatedAt,
	}
}

// AddTask adds a task ID to this iteration
func (i *IterationEntity) AddTask(taskID string) error {
	// Check if task already exists
	for _, id := range i.TaskIDs {
		if id == taskID {
			return fmt.Errorf("%w: task already in iteration", pluginsdk.ErrAlreadyExists)
		}
	}
	i.TaskIDs = append(i.TaskIDs, taskID)
	return nil
}

// RemoveTask removes a task ID from this iteration
func (i *IterationEntity) RemoveTask(taskID string) error {
	newTaskIDs := []string{}
	found := false
	for _, id := range i.TaskIDs {
		if id != taskID {
			newTaskIDs = append(newTaskIDs, id)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("%w: task not in iteration", pluginsdk.ErrNotFound)
	}

	i.TaskIDs = newTaskIDs
	return nil
}

// HasTask checks if this iteration contains a task
func (i *IterationEntity) HasTask(taskID string) bool {
	for _, id := range i.TaskIDs {
		if id == taskID {
			return true
		}
	}
	return false
}

// GetTaskCount returns the number of tasks in this iteration
func (i *IterationEntity) GetTaskCount() int {
	return len(i.TaskIDs)
}

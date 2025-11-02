package task_manager

import (
	"context"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// EventEmittingRepository is a decorator that wraps a RoadmapRepository
// and emits events to the event bus on all CRUD operations and status changes.
// This enables cross-plugin integration through the event bus.
type EventEmittingRepository struct {
	repo     RoadmapRepository
	eventBus pluginsdk.EventBus
	logger   pluginsdk.Logger
}

// NewEventEmittingRepository creates a new event-emitting repository decorator.
// If eventBus is nil, events are not emitted but operations continue normally.
func NewEventEmittingRepository(
	repo RoadmapRepository,
	eventBus pluginsdk.EventBus,
	logger pluginsdk.Logger,
) *EventEmittingRepository {
	return &EventEmittingRepository{
		repo:     repo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// ============================================================================
// Roadmap Operations
// ============================================================================

// SaveRoadmap persists a new roadmap to storage and emits EventRoadmapCreated.
func (e *EventEmittingRepository) SaveRoadmap(ctx context.Context, roadmap *RoadmapEntity) error {
	if err := e.repo.SaveRoadmap(ctx, roadmap); err != nil {
		return err
	}

	e.emitRoadmapCreatedEvent(ctx, roadmap)
	return nil
}

// GetRoadmap retrieves a roadmap by its ID (read-only, no event).
func (e *EventEmittingRepository) GetRoadmap(ctx context.Context, id string) (*RoadmapEntity, error) {
	return e.repo.GetRoadmap(ctx, id)
}

// GetActiveRoadmap retrieves the most recently created roadmap (read-only, no event).
func (e *EventEmittingRepository) GetActiveRoadmap(ctx context.Context) (*RoadmapEntity, error) {
	return e.repo.GetActiveRoadmap(ctx)
}

// UpdateRoadmap updates an existing roadmap and emits EventRoadmapUpdated.
func (e *EventEmittingRepository) UpdateRoadmap(ctx context.Context, roadmap *RoadmapEntity) error {
	if err := e.repo.UpdateRoadmap(ctx, roadmap); err != nil {
		return err
	}

	e.emitRoadmapUpdatedEvent(ctx, roadmap)
	return nil
}

// ============================================================================
// Track Operations
// ============================================================================

// SaveTrack persists a new track and emits EventTrackCreated.
func (e *EventEmittingRepository) SaveTrack(ctx context.Context, track *TrackEntity) error {
	if err := e.repo.SaveTrack(ctx, track); err != nil {
		return err
	}

	e.emitTrackCreatedEvent(ctx, track)
	return nil
}

// GetTrack retrieves a track by ID (read-only, no event).
func (e *EventEmittingRepository) GetTrack(ctx context.Context, id string) (*TrackEntity, error) {
	return e.repo.GetTrack(ctx, id)
}

// ListTracks returns all tracks for a roadmap, optionally filtered (read-only, no event).
func (e *EventEmittingRepository) ListTracks(
	ctx context.Context,
	roadmapID string,
	filters TrackFilters,
) ([]*TrackEntity, error) {
	return e.repo.ListTracks(ctx, roadmapID, filters)
}

// UpdateTrack updates an existing track and emits appropriate events.
// Emits EventTrackUpdated and EventTrackStatusChanged if status changed.
// Also emits EventTrackCompleted or EventTrackBlocked for specific status changes.
func (e *EventEmittingRepository) UpdateTrack(ctx context.Context, track *TrackEntity) error {
	// Get old track to compare status
	oldTrack, err := e.repo.GetTrack(ctx, track.ID)
	if err != nil {
		return err
	}

	if err := e.repo.UpdateTrack(ctx, track); err != nil {
		return err
	}

	// Emit base update event
	e.emitTrackUpdatedEvent(ctx, track)

	// Emit status-specific events if status changed
	if oldTrack.Status != track.Status {
		e.emitTrackStatusChangedEvent(ctx, track.ID, oldTrack.Status, track.Status)

		// Emit specific status events
		if track.Status == "complete" {
			e.emitTrackCompletedEvent(ctx, track)
		} else if track.Status == "blocked" {
			e.emitTrackBlockedEvent(ctx, track)
		}
	}

	return nil
}

// DeleteTrack removes a track and emits an event.
func (e *EventEmittingRepository) DeleteTrack(ctx context.Context, id string) error {
	return e.repo.DeleteTrack(ctx, id)
}

// AddTrackDependency adds a dependency from trackID to dependsOnID.
func (e *EventEmittingRepository) AddTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	return e.repo.AddTrackDependency(ctx, trackID, dependsOnID)
}

// RemoveTrackDependency removes a dependency from trackID to dependsOnID.
func (e *EventEmittingRepository) RemoveTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	return e.repo.RemoveTrackDependency(ctx, trackID, dependsOnID)
}

// GetTrackDependencies returns the IDs of all tracks that trackID depends on.
func (e *EventEmittingRepository) GetTrackDependencies(ctx context.Context, trackID string) ([]string, error) {
	return e.repo.GetTrackDependencies(ctx, trackID)
}

// ValidateNoCycles checks if adding/updating the track would create a circular dependency.
func (e *EventEmittingRepository) ValidateNoCycles(ctx context.Context, trackID string) error {
	return e.repo.ValidateNoCycles(ctx, trackID)
}

// ============================================================================
// Task Operations
// ============================================================================

// SaveTask persists a new task and emits EventTaskCreated.
func (e *EventEmittingRepository) SaveTask(ctx context.Context, task *TaskEntity) error {
	if err := e.repo.SaveTask(ctx, task); err != nil {
		return err
	}

	e.emitTaskCreatedEvent(ctx, task)
	return nil
}

// GetTask retrieves a task by ID (read-only, no event).
func (e *EventEmittingRepository) GetTask(ctx context.Context, id string) (*TaskEntity, error) {
	return e.repo.GetTask(ctx, id)
}

// ListTasks returns all tasks matching the filters (read-only, no event).
func (e *EventEmittingRepository) ListTasks(ctx context.Context, filters TaskFilters) ([]*TaskEntity, error) {
	return e.repo.ListTasks(ctx, filters)
}

// UpdateTask updates an existing task and emits appropriate events.
// Emits EventTaskUpdated and EventTaskStatusChanged if status changed.
// Also emits EventTaskCompleted for completion status changes.
func (e *EventEmittingRepository) UpdateTask(ctx context.Context, task *TaskEntity) error {
	// Get old task to compare status
	oldTask, err := e.repo.GetTask(ctx, task.ID)
	if err != nil {
		return err
	}

	if err := e.repo.UpdateTask(ctx, task); err != nil {
		return err
	}

	// Emit base update event
	e.emitTaskUpdatedEvent(ctx, task)

	// Emit status-specific events if status changed
	if oldTask.Status != task.Status {
		e.emitTaskStatusChangedEvent(ctx, task.ID, oldTask.Status, task.Status)

		// Emit specific status events
		if task.Status == "done" {
			e.emitTaskCompletedEvent(ctx, task)
		}
	}

	return nil
}

// DeleteTask removes a task from storage.
func (e *EventEmittingRepository) DeleteTask(ctx context.Context, id string) error {
	return e.repo.DeleteTask(ctx, id)
}

// MoveTaskToTrack moves a task from its current track to a new track.
func (e *EventEmittingRepository) MoveTaskToTrack(ctx context.Context, taskID, newTrackID string) error {
	return e.repo.MoveTaskToTrack(ctx, taskID, newTrackID)
}

// ============================================================================
// Iteration Operations
// ============================================================================

// SaveIteration persists a new iteration and emits EventIterationCreated.
func (e *EventEmittingRepository) SaveIteration(ctx context.Context, iteration *IterationEntity) error {
	if err := e.repo.SaveIteration(ctx, iteration); err != nil {
		return err
	}

	e.emitIterationCreatedEvent(ctx, iteration)
	return nil
}

// GetIteration retrieves an iteration by its number (read-only, no event).
func (e *EventEmittingRepository) GetIteration(ctx context.Context, number int) (*IterationEntity, error) {
	return e.repo.GetIteration(ctx, number)
}

// GetCurrentIteration returns the iteration with status "current" (read-only, no event).
func (e *EventEmittingRepository) GetCurrentIteration(ctx context.Context) (*IterationEntity, error) {
	return e.repo.GetCurrentIteration(ctx)
}

// ListIterations returns all iterations, ordered by number (read-only, no event).
func (e *EventEmittingRepository) ListIterations(ctx context.Context) ([]*IterationEntity, error) {
	return e.repo.ListIterations(ctx)
}

// UpdateIteration updates an existing iteration and emits EventIterationUpdated.
func (e *EventEmittingRepository) UpdateIteration(ctx context.Context, iteration *IterationEntity) error {
	if err := e.repo.UpdateIteration(ctx, iteration); err != nil {
		return err
	}

	e.emitIterationUpdatedEvent(ctx, iteration)
	return nil
}

// DeleteIteration removes an iteration from storage.
func (e *EventEmittingRepository) DeleteIteration(ctx context.Context, number int) error {
	return e.repo.DeleteIteration(ctx, number)
}

// AddTaskToIteration adds a task to an iteration.
func (e *EventEmittingRepository) AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error {
	return e.repo.AddTaskToIteration(ctx, iterationNum, taskID)
}

// RemoveTaskFromIteration removes a task from an iteration.
func (e *EventEmittingRepository) RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error {
	return e.repo.RemoveTaskFromIteration(ctx, iterationNum, taskID)
}

// GetIterationTasks returns all tasks in an iteration (read-only, no event).
func (e *EventEmittingRepository) GetIterationTasks(ctx context.Context, iterationNum int) ([]*TaskEntity, error) {
	return e.repo.GetIterationTasks(ctx, iterationNum)
}

// StartIteration marks an iteration as current and emits EventIterationStarted.
func (e *EventEmittingRepository) StartIteration(ctx context.Context, iterationNum int) error {
	if err := e.repo.StartIteration(ctx, iterationNum); err != nil {
		return err
	}

	// Get the iteration to include in the event
	iteration, err := e.repo.GetIteration(ctx, iterationNum)
	if err != nil {
		e.logger.Warn("failed to get iteration for event emission", "number", iterationNum, "error", err)
		return nil
	}

	e.emitIterationStartedEvent(ctx, iteration)
	return nil
}

// CompleteIteration marks an iteration as complete and emits EventIterationCompleted.
func (e *EventEmittingRepository) CompleteIteration(ctx context.Context, iterationNum int) error {
	if err := e.repo.CompleteIteration(ctx, iterationNum); err != nil {
		return err
	}

	// Get the iteration to include in the event
	iteration, err := e.repo.GetIteration(ctx, iterationNum)
	if err != nil {
		e.logger.Warn("failed to get iteration for event emission", "number", iterationNum, "error", err)
		return nil
	}

	e.emitIterationCompletedEvent(ctx, iteration)
	return nil
}

// ============================================================================
// Aggregate Queries
// ============================================================================

// GetRoadmapWithTracks retrieves a roadmap with all its tracks (read-only, no event).
func (e *EventEmittingRepository) GetRoadmapWithTracks(ctx context.Context, roadmapID string) (*RoadmapEntity, error) {
	return e.repo.GetRoadmapWithTracks(ctx, roadmapID)
}

// GetTrackWithTasks retrieves a track with all its tasks (read-only, no event).
func (e *EventEmittingRepository) GetTrackWithTasks(ctx context.Context, trackID string) (*TrackEntity, error) {
	return e.repo.GetTrackWithTasks(ctx, trackID)
}

// ============================================================================
// Project Metadata Operations
// ============================================================================

// GetProjectMetadata retrieves a metadata value by key (read-only, no event).
func (e *EventEmittingRepository) GetProjectMetadata(ctx context.Context, key string) (string, error) {
	return e.repo.GetProjectMetadata(ctx, key)
}

// SetProjectMetadata sets a metadata value by key (write-only, no event).
func (e *EventEmittingRepository) SetProjectMetadata(ctx context.Context, key, value string) error {
	return e.repo.SetProjectMetadata(ctx, key, value)
}

// GetProjectCode retrieves the project code (read-only, no event).
func (e *EventEmittingRepository) GetProjectCode(ctx context.Context) string {
	return e.repo.GetProjectCode(ctx)
}

// GetNextSequenceNumber retrieves the next sequence number for an entity type (read-only, no event).
func (e *EventEmittingRepository) GetNextSequenceNumber(ctx context.Context, entityType string) (int, error) {
	return e.repo.GetNextSequenceNumber(ctx, entityType)
}

// ============================================================================
// Event Emission Helpers
// ============================================================================

// emitRoadmapCreatedEvent emits EventRoadmapCreated to the event bus.
func (e *EventEmittingRepository) emitRoadmapCreatedEvent(ctx context.Context, roadmap *RoadmapEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"roadmap_id":       roadmap.ID,
		"vision":           roadmap.Vision,
		"success_criteria": roadmap.SuccessCriteria,
		"created_at":       roadmap.CreatedAt,
	}

	e.publishEvent(ctx, EventRoadmapCreated, payload)
}

// emitRoadmapUpdatedEvent emits EventRoadmapUpdated to the event bus.
func (e *EventEmittingRepository) emitRoadmapUpdatedEvent(ctx context.Context, roadmap *RoadmapEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"roadmap_id":       roadmap.ID,
		"vision":           roadmap.Vision,
		"success_criteria": roadmap.SuccessCriteria,
		"updated_at":       roadmap.UpdatedAt,
	}

	e.publishEvent(ctx, EventRoadmapUpdated, payload)
}

// emitTrackCreatedEvent emits EventTrackCreated to the event bus.
func (e *EventEmittingRepository) emitTrackCreatedEvent(ctx context.Context, track *TrackEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"track_id":   track.ID,
		"roadmap_id": track.RoadmapID,
		"title":      track.Title,
		"description": track.Description,
		"status":     track.Status,
		"rank":       track.Rank,
		"created_at": track.CreatedAt,
	}

	e.publishEvent(ctx, EventTrackCreated, payload)
}

// emitTrackUpdatedEvent emits EventTrackUpdated to the event bus.
func (e *EventEmittingRepository) emitTrackUpdatedEvent(ctx context.Context, track *TrackEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"track_id":   track.ID,
		"roadmap_id": track.RoadmapID,
		"title":      track.Title,
		"description": track.Description,
		"status":     track.Status,
		"rank":       track.Rank,
		"updated_at": track.UpdatedAt,
	}

	e.publishEvent(ctx, EventTrackUpdated, payload)
}

// emitTrackStatusChangedEvent emits EventTrackStatusChanged when track status changes.
func (e *EventEmittingRepository) emitTrackStatusChangedEvent(
	ctx context.Context,
	trackID, oldStatus, newStatus string,
) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"track_id":   trackID,
		"old_status": oldStatus,
		"new_status": newStatus,
	}

	e.publishEvent(ctx, EventTrackStatusChanged, payload)
}

// emitTrackCompletedEvent emits EventTrackCompleted when track reaches "complete" status.
func (e *EventEmittingRepository) emitTrackCompletedEvent(ctx context.Context, track *TrackEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"track_id": track.ID,
		"title":    track.Title,
		"completed_at": time.Now(),
	}

	e.publishEvent(ctx, EventTrackCompleted, payload)
}

// emitTrackBlockedEvent emits EventTrackBlocked when track reaches "blocked" status.
func (e *EventEmittingRepository) emitTrackBlockedEvent(ctx context.Context, track *TrackEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"track_id": track.ID,
		"title":    track.Title,
		"blocked_at": time.Now(),
	}

	e.publishEvent(ctx, EventTrackBlocked, payload)
}

// emitTaskCreatedEvent emits EventTaskCreated to the event bus.
func (e *EventEmittingRepository) emitTaskCreatedEvent(ctx context.Context, task *TaskEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"task_id":   task.ID,
		"track_id":  task.TrackID,
		"title":     task.Title,
		"description": task.Description,
		"status":    task.Status,
		"rank":      task.Rank,
		"branch":    task.Branch,
		"created_at": task.CreatedAt,
	}

	e.publishEvent(ctx, EventTaskCreated, payload)
}

// emitTaskUpdatedEvent emits EventTaskUpdated to the event bus.
func (e *EventEmittingRepository) emitTaskUpdatedEvent(ctx context.Context, task *TaskEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"task_id":   task.ID,
		"track_id":  task.TrackID,
		"title":     task.Title,
		"description": task.Description,
		"status":    task.Status,
		"rank":      task.Rank,
		"branch":    task.Branch,
		"updated_at": task.UpdatedAt,
	}

	e.publishEvent(ctx, EventTaskUpdated, payload)
}

// emitTaskStatusChangedEvent emits EventTaskStatusChanged when task status changes.
func (e *EventEmittingRepository) emitTaskStatusChangedEvent(
	ctx context.Context,
	taskID, oldStatus, newStatus string,
) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"task_id":   taskID,
		"old_status": oldStatus,
		"new_status": newStatus,
	}

	e.publishEvent(ctx, EventTaskStatusChanged, payload)
}

// emitTaskCompletedEvent emits EventTaskCompleted when task reaches "done" status.
func (e *EventEmittingRepository) emitTaskCompletedEvent(ctx context.Context, task *TaskEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"task_id": task.ID,
		"title":    task.Title,
		"completed_at": time.Now(),
	}

	e.publishEvent(ctx, EventTaskCompleted, payload)
}

// emitIterationCreatedEvent emits EventIterationCreated to the event bus.
func (e *EventEmittingRepository) emitIterationCreatedEvent(ctx context.Context, iteration *IterationEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"iteration_number": iteration.Number,
		"name":             iteration.Name,
		"goal":             iteration.Goal,
		"status":           iteration.Status,
		"created_at":       iteration.CreatedAt,
	}

	e.publishEvent(ctx, EventIterationCreated, payload)
}

// emitIterationUpdatedEvent emits EventIterationUpdated to the event bus.
func (e *EventEmittingRepository) emitIterationUpdatedEvent(ctx context.Context, iteration *IterationEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"iteration_number": iteration.Number,
		"name":             iteration.Name,
		"goal":             iteration.Goal,
		"status":           iteration.Status,
		"updated_at":       iteration.UpdatedAt,
	}

	e.publishEvent(ctx, EventIterationUpdated, payload)
}

// emitIterationStartedEvent emits EventIterationStarted when iteration is started.
func (e *EventEmittingRepository) emitIterationStartedEvent(ctx context.Context, iteration *IterationEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"iteration_number": iteration.Number,
		"started_at":       iteration.StartedAt,
	}

	e.publishEvent(ctx, EventIterationStarted, payload)
}

// emitIterationCompletedEvent emits EventIterationCompleted when iteration is completed.
func (e *EventEmittingRepository) emitIterationCompletedEvent(ctx context.Context, iteration *IterationEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"iteration_number": iteration.Number,
		"completed_at":     iteration.CompletedAt,
	}

	e.publishEvent(ctx, EventIterationCompleted, payload)
}

// ============================================================================
// Acceptance Criteria Operations
// ============================================================================

// SaveAC persists a new acceptance criterion and emits EventACCreated.
func (e *EventEmittingRepository) SaveAC(ctx context.Context, ac *AcceptanceCriteriaEntity) error {
	if err := e.repo.SaveAC(ctx, ac); err != nil {
		return err
	}

	e.emitACCreatedEvent(ctx, ac)
	return nil
}

// GetAC retrieves an acceptance criterion by its ID (read-only, no event).
func (e *EventEmittingRepository) GetAC(ctx context.Context, id string) (*AcceptanceCriteriaEntity, error) {
	return e.repo.GetAC(ctx, id)
}

// ListAC returns all acceptance criteria for a task (read-only, no event).
func (e *EventEmittingRepository) ListAC(ctx context.Context, taskID string) ([]*AcceptanceCriteriaEntity, error) {
	return e.repo.ListAC(ctx, taskID)
}

// UpdateAC updates an existing acceptance criterion and emits EventACUpdated.
func (e *EventEmittingRepository) UpdateAC(ctx context.Context, ac *AcceptanceCriteriaEntity) error {
	if err := e.repo.UpdateAC(ctx, ac); err != nil {
		return err
	}

	e.emitACUpdatedEvent(ctx, ac)
	return nil
}

// DeleteAC removes an acceptance criterion and emits EventACDeleted.
func (e *EventEmittingRepository) DeleteAC(ctx context.Context, id string) error {
	if err := e.repo.DeleteAC(ctx, id); err != nil {
		return err
	}

	e.emitACDeletedEvent(ctx, id)
	return nil
}

// ListACByTrack returns all acceptance criteria for all tasks in a track (read-only, no event).
func (e *EventEmittingRepository) ListACByTrack(ctx context.Context, trackID string) ([]*AcceptanceCriteriaEntity, error) {
	return e.repo.ListACByTrack(ctx, trackID)
}

// ListACByIteration returns all acceptance criteria for all tasks in an iteration (read-only, no event).
func (e *EventEmittingRepository) ListACByIteration(ctx context.Context, iterationNum int) ([]*AcceptanceCriteriaEntity, error) {
	return e.repo.ListACByIteration(ctx, iterationNum)
}

// emitACCreatedEvent emits EventACCreated to the event bus.
func (e *EventEmittingRepository) emitACCreatedEvent(ctx context.Context, ac *AcceptanceCriteriaEntity) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"id":                 ac.ID,
		"task_id":            ac.TaskID,
		"description":        ac.Description,
		"verification_type":  string(ac.VerificationType),
		"status":             string(ac.Status),
		"created_at":         ac.CreatedAt,
	}

	e.publishEvent(ctx, EventACCreated, payload)
}

// emitACUpdatedEvent emits EventACUpdated to the event bus.
func (e *EventEmittingRepository) emitACUpdatedEvent(ctx context.Context, ac *AcceptanceCriteriaEntity) {
	if e.eventBus == nil {
		return
	}

	// Emit status-specific events based on the current status
	var statusEvent string
	switch ac.Status {
	case ACStatusVerified:
		statusEvent = EventACVerified
	case ACStatusAutomaticallyVerified:
		statusEvent = EventACAutomaticallyVerified
	case ACStatusPendingHumanReview:
		statusEvent = EventACPendingReview
	case ACStatusFailed:
		statusEvent = EventACFailed
	default:
		statusEvent = EventACUpdated
	}

	payload := map[string]interface{}{
		"id":                 ac.ID,
		"task_id":            ac.TaskID,
		"description":        ac.Description,
		"verification_type":  string(ac.VerificationType),
		"status":             string(ac.Status),
		"notes":              ac.Notes,
		"updated_at":         ac.UpdatedAt,
	}

	e.publishEvent(ctx, statusEvent, payload)
}

// emitACDeletedEvent emits EventACDeleted to the event bus.
func (e *EventEmittingRepository) emitACDeletedEvent(ctx context.Context, acID string) {
	if e.eventBus == nil {
		return
	}

	payload := map[string]interface{}{
		"id": acID,
	}

	e.publishEvent(ctx, EventACDeleted, payload)
}

// publishEvent publishes an event to the event bus with error handling.
func (e *EventEmittingRepository) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if e.eventBus == nil {
		return
	}

	// Create bus event with JSON-encoded payload
	event, err := pluginsdk.NewBusEvent(eventType, PluginSourceName, payload)
	if err != nil {
		e.logger.Error("failed to create bus event", "type", eventType, "error", err)
		return
	}

	// Add labels for filtering
	event.Labels["event_type"] = eventType
	event.Labels["plugin"] = PluginSourceName

	// Publish asynchronously with a short timeout to avoid blocking
	publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.eventBus.Publish(publishCtx, event); err != nil {
		e.logger.Error("failed to publish event", "type", eventType, "error", err)
	} else {
		e.logger.Debug("published event", "type", eventType, "payload", payload)
	}
}

// ============================================================================
// ADR Operations
// ============================================================================

// SaveADR persists a new ADR and emits EventADRCreated.
func (e *EventEmittingRepository) SaveADR(ctx context.Context, adr *ADREntity) error {
	if err := e.repo.SaveADR(ctx, adr); err != nil {
		return err
	}

	e.publishEvent(ctx, EventADRCreated, map[string]interface{}{
		"id":       adr.ID,
		"track_id": adr.TrackID,
		"title":    adr.Title,
		"status":   adr.Status,
	})
	return nil
}

// GetADR retrieves an ADR by its ID (read-only, no event).
func (e *EventEmittingRepository) GetADR(ctx context.Context, id string) (*ADREntity, error) {
	return e.repo.GetADR(ctx, id)
}

// ListADRs returns all ADRs, optionally filtered by track (read-only, no event).
func (e *EventEmittingRepository) ListADRs(ctx context.Context, trackID *string) ([]*ADREntity, error) {
	return e.repo.ListADRs(ctx, trackID)
}

// UpdateADR updates an existing ADR and emits EventADRUpdated.
func (e *EventEmittingRepository) UpdateADR(ctx context.Context, adr *ADREntity) error {
	if err := e.repo.UpdateADR(ctx, adr); err != nil {
		return err
	}

	e.publishEvent(ctx, EventADRUpdated, map[string]interface{}{
		"id": adr.ID,
	})
	return nil
}

// SupersedeADR marks an ADR as superseded and emits EventADRSuperseded.
func (e *EventEmittingRepository) SupersedeADR(ctx context.Context, adrID, supersededByID string) error {
	if err := e.repo.SupersedeADR(ctx, adrID, supersededByID); err != nil {
		return err
	}

	e.publishEvent(ctx, EventADRSuperseded, map[string]interface{}{
		"id":             adrID,
		"superseded_by":  supersededByID,
	})
	return nil
}

// DeprecateADR marks an ADR as deprecated and emits EventADRDeprecated.
func (e *EventEmittingRepository) DeprecateADR(ctx context.Context, adrID string) error {
	if err := e.repo.DeprecateADR(ctx, adrID); err != nil {
		return err
	}

	e.publishEvent(ctx, EventADRDeprecated, map[string]interface{}{
		"id": adrID,
	})
	return nil
}

// GetADRsByTrack returns all ADRs for a specific track (read-only, no event).
func (e *EventEmittingRepository) GetADRsByTrack(ctx context.Context, trackID string) ([]*ADREntity, error) {
	return e.repo.GetADRsByTrack(ctx, trackID)
}

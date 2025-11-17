package repositories_test

import (
	"context"
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/repositories"
)

// TestRepositoryInterfacesCompile verifies that all repository interfaces are well-formed
// and can be implemented by checking that they satisfy the interface contract.
// This is a compile-time verification test.
func TestRepositoryInterfacesCompile(t *testing.T) {
	// This test verifies that all repository interfaces are correctly defined
	// by ensuring they can be assigned to interface{} and have expected methods.
	// The actual compilation ensures the interfaces are valid.

	var (
		_ repositories.RoadmapRepository           = (*mockRoadmapRepository)(nil)
		_ repositories.TrackRepository             = (*mockTrackRepository)(nil)
		_ repositories.TaskRepository              = (*mockTaskRepository)(nil)
		_ repositories.IterationRepository         = (*mockIterationRepository)(nil)
		_ repositories.ADRRepository               = (*mockADRRepository)(nil)
		_ repositories.AcceptanceCriteriaRepository = (*mockACRepository)(nil)
		_ repositories.DocumentRepository          = (*mockDocumentRepository)(nil)
		_ repositories.AggregateRepository         = (*mockAggregateRepository)(nil)
	)
}

// Mock implementations for interface verification

type mockRoadmapRepository struct{}

func (m *mockRoadmapRepository) SaveRoadmap(ctx context.Context, roadmap *entities.RoadmapEntity) error {
	return nil
}

func (m *mockRoadmapRepository) GetRoadmap(ctx context.Context, id string) (*entities.RoadmapEntity, error) {
	return nil, nil
}

func (m *mockRoadmapRepository) GetActiveRoadmap(ctx context.Context) (*entities.RoadmapEntity, error) {
	return nil, nil
}

func (m *mockRoadmapRepository) UpdateRoadmap(ctx context.Context, roadmap *entities.RoadmapEntity) error {
	return nil
}

type mockTrackRepository struct{}

func (m *mockTrackRepository) SaveTrack(ctx context.Context, track *entities.TrackEntity) error {
	return nil
}

func (m *mockTrackRepository) GetTrack(ctx context.Context, id string) (*entities.TrackEntity, error) {
	return nil, nil
}

func (m *mockTrackRepository) ListTracks(ctx context.Context, roadmapID string, filters entities.TrackFilters) ([]*entities.TrackEntity, error) {
	return nil, nil
}

func (m *mockTrackRepository) UpdateTrack(ctx context.Context, track *entities.TrackEntity) error {
	return nil
}

func (m *mockTrackRepository) DeleteTrack(ctx context.Context, id string) error {
	return nil
}

func (m *mockTrackRepository) AddTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	return nil
}

func (m *mockTrackRepository) RemoveTrackDependency(ctx context.Context, trackID, dependsOnID string) error {
	return nil
}

func (m *mockTrackRepository) GetTrackDependencies(ctx context.Context, trackID string) ([]string, error) {
	return nil, nil
}

func (m *mockTrackRepository) ValidateNoCycles(ctx context.Context, trackID string) error {
	return nil
}

func (m *mockTrackRepository) GetTrackWithTasks(ctx context.Context, trackID string) (*entities.TrackEntity, error) {
	return nil, nil
}

type mockTaskRepository struct{}

func (m *mockTaskRepository) SaveTask(ctx context.Context, task *entities.TaskEntity) error {
	return nil
}

func (m *mockTaskRepository) GetTask(ctx context.Context, id string) (*entities.TaskEntity, error) {
	return nil, nil
}

func (m *mockTaskRepository) ListTasks(ctx context.Context, filters entities.TaskFilters) ([]*entities.TaskEntity, error) {
	return nil, nil
}

func (m *mockTaskRepository) UpdateTask(ctx context.Context, task *entities.TaskEntity) error {
	return nil
}

func (m *mockTaskRepository) DeleteTask(ctx context.Context, id string) error {
	return nil
}

func (m *mockTaskRepository) MoveTaskToTrack(ctx context.Context, taskID, newTrackID string) error {
	return nil
}

func (m *mockTaskRepository) GetBacklogTasks(ctx context.Context) ([]*entities.TaskEntity, error) {
	return nil, nil
}

func (m *mockTaskRepository) GetIterationsForTask(ctx context.Context, taskID string) ([]*entities.IterationEntity, error) {
	return nil, nil
}

type mockIterationRepository struct{}

func (m *mockIterationRepository) SaveIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	return nil
}

func (m *mockIterationRepository) GetIteration(ctx context.Context, number int) (*entities.IterationEntity, error) {
	return nil, nil
}

func (m *mockIterationRepository) GetCurrentIteration(ctx context.Context) (*entities.IterationEntity, error) {
	return nil, nil
}

func (m *mockIterationRepository) ListIterations(ctx context.Context) ([]*entities.IterationEntity, error) {
	return nil, nil
}

func (m *mockIterationRepository) UpdateIteration(ctx context.Context, iteration *entities.IterationEntity) error {
	return nil
}

func (m *mockIterationRepository) DeleteIteration(ctx context.Context, number int) error {
	return nil
}

func (m *mockIterationRepository) AddTaskToIteration(ctx context.Context, iterationNum int, taskID string) error {
	return nil
}

func (m *mockIterationRepository) RemoveTaskFromIteration(ctx context.Context, iterationNum int, taskID string) error {
	return nil
}

func (m *mockIterationRepository) GetIterationTasks(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, error) {
	return nil, nil
}

func (m *mockIterationRepository) GetIterationTasksWithWarnings(ctx context.Context, iterationNum int) ([]*entities.TaskEntity, []string, error) {
	return nil, nil, nil
}

func (m *mockIterationRepository) StartIteration(ctx context.Context, iterationNumber int) error {
	return nil
}

func (m *mockIterationRepository) CompleteIteration(ctx context.Context, iterationNumber int) error {
	return nil
}

func (m *mockIterationRepository) GetIterationByNumber(ctx context.Context, iterationNumber int) (*entities.IterationEntity, error) {
	return nil, nil
}

func (m *mockIterationRepository) GetNextPlannedIteration(ctx context.Context) (*entities.IterationEntity, error) {
	return nil, nil
}

type mockADRRepository struct{}

func (m *mockADRRepository) SaveADR(ctx context.Context, adr *entities.ADREntity) error {
	return nil
}

func (m *mockADRRepository) GetADR(ctx context.Context, id string) (*entities.ADREntity, error) {
	return nil, nil
}

func (m *mockADRRepository) ListADRs(ctx context.Context, trackID *string) ([]*entities.ADREntity, error) {
	return nil, nil
}

func (m *mockADRRepository) UpdateADR(ctx context.Context, adr *entities.ADREntity) error {
	return nil
}

func (m *mockADRRepository) SupersedeADR(ctx context.Context, adrID, supersededByID string) error {
	return nil
}

func (m *mockADRRepository) DeprecateADR(ctx context.Context, adrID string) error {
	return nil
}

func (m *mockADRRepository) GetADRsByTrack(ctx context.Context, trackID string) ([]*entities.ADREntity, error) {
	return nil, nil
}

type mockACRepository struct{}

func (m *mockACRepository) SaveAC(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
	return nil
}

func (m *mockACRepository) GetAC(ctx context.Context, id string) (*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *mockACRepository) ListAC(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *mockACRepository) UpdateAC(ctx context.Context, ac *entities.AcceptanceCriteriaEntity) error {
	return nil
}

func (m *mockACRepository) DeleteAC(ctx context.Context, id string) error {
	return nil
}

func (m *mockACRepository) ListACByTask(ctx context.Context, taskID string) ([]*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *mockACRepository) ListACByIteration(ctx context.Context, iterationNum int) ([]*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

func (m *mockACRepository) ListFailedAC(ctx context.Context, filters entities.ACFilters) ([]*entities.AcceptanceCriteriaEntity, error) {
	return nil, nil
}

type mockDocumentRepository struct{}

func (m *mockDocumentRepository) SaveDocument(ctx context.Context, doc *entities.DocumentEntity) error {
	return nil
}

func (m *mockDocumentRepository) FindDocumentByID(ctx context.Context, id string) (*entities.DocumentEntity, error) {
	return nil, nil
}

func (m *mockDocumentRepository) FindAllDocuments(ctx context.Context) ([]*entities.DocumentEntity, error) {
	return nil, nil
}

func (m *mockDocumentRepository) FindDocumentsByTrack(ctx context.Context, trackID string) ([]*entities.DocumentEntity, error) {
	return nil, nil
}

func (m *mockDocumentRepository) FindDocumentsByIteration(ctx context.Context, iterationNumber int) ([]*entities.DocumentEntity, error) {
	return nil, nil
}

func (m *mockDocumentRepository) FindDocumentsByType(ctx context.Context, docType entities.DocumentType) ([]*entities.DocumentEntity, error) {
	return nil, nil
}

func (m *mockDocumentRepository) UpdateDocument(ctx context.Context, doc *entities.DocumentEntity) error {
	return nil
}

func (m *mockDocumentRepository) DeleteDocument(ctx context.Context, id string) error {
	return nil
}

type mockAggregateRepository struct{}

func (m *mockAggregateRepository) GetRoadmapWithTracks(ctx context.Context, roadmapID string) (*entities.RoadmapEntity, error) {
	return nil, nil
}

func (m *mockAggregateRepository) GetProjectMetadata(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *mockAggregateRepository) SetProjectMetadata(ctx context.Context, key, value string) error {
	return nil
}

func (m *mockAggregateRepository) GetProjectCode(ctx context.Context) string {
	return ""
}

func (m *mockAggregateRepository) GetNextSequenceNumber(ctx context.Context, entityType string) (int, error) {
	return 0, nil
}

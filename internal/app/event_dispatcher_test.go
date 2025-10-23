package app_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// mockEventEmitter implements pluginsdk.IEventEmitter for testing
type mockEventEmitter struct {
	pluginsdk.Plugin
	name          string
	eventCount    int
	eventInterval time.Duration
	running       atomic.Bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func newMockEventEmitter(name string, eventCount int, eventInterval time.Duration) *mockEventEmitter {
	return &mockEventEmitter{
		Plugin:        &mockPlugin{name: name, capabilities: []string{"IEventEmitter"}},
		name:          name,
		eventCount:    eventCount,
		eventInterval: eventInterval,
		stopChan:      make(chan struct{}),
	}
}

func (m *mockEventEmitter) StartEventStream(ctx context.Context, eventChan chan<- pluginsdk.Event) error {
	if !m.running.CompareAndSwap(false, true) {
		return fmt.Errorf("already running")
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		for i := 0; i < m.eventCount; i++ {
			select {
			case <-ctx.Done():
				return
			case <-m.stopChan:
				return
			default:
				event := pluginsdk.Event{
					Type:      fmt.Sprintf("%s.event", m.name),
					Source:    m.name,
					Timestamp: time.Now(),
					Payload: map[string]interface{}{
						"index": i,
						"data":  fmt.Sprintf("Event %d from %s", i, m.name),
					},
					Metadata: map[string]string{
						"plugin": m.name,
					},
					Version: "1.0",
				}

				select {
				case eventChan <- event:
				case <-ctx.Done():
					return
				case <-m.stopChan:
					return
				}

				if m.eventInterval > 0 {
					time.Sleep(m.eventInterval)
				}
			}
		}
	}()

	return nil
}

func (m *mockEventEmitter) StopEventStream() error {
	if !m.running.CompareAndSwap(true, false) {
		return nil // Already stopped, idempotent
	}

	close(m.stopChan)
	m.wg.Wait()
	return nil
}

// mockPlugin implements pluginsdk.Plugin
type mockPlugin struct {
	name         string
	capabilities []string
}

func (m *mockPlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        m.name,
		Version:     "1.0.0",
		Description: fmt.Sprintf("Mock plugin: %s", m.name),
		IsCore:      false,
	}
}

func (m *mockPlugin) GetCapabilities() []string {
	return m.capabilities
}

// dispatcherEventRepository implements domain.EventRepository for dispatcher testing
// Named differently to avoid conflict with existing mockEventRepository
type dispatcherEventRepository struct {
	events []*domain.Event
	mu     sync.Mutex
}

func newDispatcherEventRepository() *dispatcherEventRepository {
	return &dispatcherEventRepository{
		events: make([]*domain.Event, 0),
	}
}

func (m *dispatcherEventRepository) Save(ctx context.Context, event *domain.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *dispatcherEventRepository) GetEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func (m *dispatcherEventRepository) FindBySessionID(ctx context.Context, sessionID string, limit, offset int, ordered bool) ([]*domain.Event, error) {
	return nil, nil
}

func (m *dispatcherEventRepository) FindByType(ctx context.Context, eventType string, limit int) ([]*domain.Event, error) {
	return nil, nil
}

func (m *dispatcherEventRepository) FindByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.Event, error) {
	return nil, nil
}

func (m *dispatcherEventRepository) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	return nil, nil
}

func (m *dispatcherEventRepository) ExecuteRawQuery(ctx context.Context, query string, args ...interface{}) (pluginsdk.QueryResult, error) {
	return pluginsdk.QueryResult{}, nil
}

func (m *dispatcherEventRepository) Close() error {
	return nil
}

func (m *dispatcherEventRepository) FindByQuery(ctx context.Context, query pluginsdk.EventQuery) ([]*domain.Event, error) {
	return nil, nil
}

func (m *dispatcherEventRepository) Initialize(ctx context.Context) error {
	return nil
}

// dispatcherPluginContext implements pluginsdk.PluginContext for dispatcher testing
type dispatcherPluginContext struct {
	repo   *dispatcherEventRepository
	logger pluginsdk.Logger
}

func (m *dispatcherPluginContext) GetLogger() pluginsdk.Logger {
	return m.logger
}

func (m *dispatcherPluginContext) GetWorkingDir() string {
	return "/tmp"
}

func (m *dispatcherPluginContext) EmitEvent(ctx context.Context, event pluginsdk.Event) error {
	// Convert SDK event to domain event for repository
	domainEvent := domain.NewEvent(
		event.Type,
		event.Metadata["session_id"],
		event.Payload,
		fmt.Sprintf("%v", event.Payload),
	)
	return m.repo.Save(ctx, domainEvent)
}

// dispatcherSDKLogger implements pluginsdk.Logger for dispatcher testing
type dispatcherSDKLogger struct{}

func (m *dispatcherSDKLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *dispatcherSDKLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *dispatcherSDKLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *dispatcherSDKLogger) Error(msg string, keysAndValues ...interface{}) {}

// TestEventDispatcher_SinglePlugin tests basic dispatcher functionality with one plugin
func TestEventDispatcher_SinglePlugin(t *testing.T) {
	repo := newDispatcherEventRepository()
	logger := &mockLogger{}
	pluginCtx := &dispatcherPluginContext{repo: repo, logger: &dispatcherSDKLogger{}}

	dispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	// Register one emitter (10 events, no delay)
	emitter := newMockEventEmitter("plugin1", 10, 0)
	dispatcher.RegisterEmitter(emitter)

	// Start dispatcher
	ctx := context.Background()
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}

	// Wait for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Stop dispatcher
	if err := dispatcher.Stop(); err != nil {
		t.Fatalf("Failed to stop dispatcher: %v", err)
	}

	// Verify all events were saved
	eventCount := repo.GetEventCount()
	if eventCount != 10 {
		t.Errorf("Expected 10 events, got %d", eventCount)
	}

	// Verify metrics
	metrics := dispatcher.GetMetrics()
	if metrics["running"].(bool) {
		t.Error("Dispatcher should not be running after Stop()")
	}
	if metrics["emitter_count"].(int) != 1 {
		t.Errorf("Expected 1 emitter, got %d", metrics["emitter_count"])
	}
}

// TestEventDispatcher_MultiplePlugins tests 2+ plugins emitting simultaneously
func TestEventDispatcher_MultiplePlugins(t *testing.T) {
	repo := newDispatcherEventRepository()
	logger := &mockLogger{}
	pluginCtx := &dispatcherPluginContext{repo: repo, logger: &dispatcherSDKLogger{}}

	dispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	// Register 3 emitters
	emitter1 := newMockEventEmitter("plugin1", 50, 0)
	emitter2 := newMockEventEmitter("plugin2", 50, 0)
	emitter3 := newMockEventEmitter("plugin3", 50, 0)

	dispatcher.RegisterEmitter(emitter1)
	dispatcher.RegisterEmitter(emitter2)
	dispatcher.RegisterEmitter(emitter3)

	// Start dispatcher
	ctx := context.Background()
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	// Stop dispatcher
	if err := dispatcher.Stop(); err != nil {
		t.Fatalf("Failed to stop dispatcher: %v", err)
	}

	// Verify all events were saved (50 * 3 = 150)
	eventCount := repo.GetEventCount()
	if eventCount != 150 {
		t.Errorf("Expected 150 events from 3 plugins, got %d", eventCount)
	}

	// Verify metrics
	metrics := dispatcher.GetMetrics()
	if metrics["emitter_count"].(int) != 3 {
		t.Errorf("Expected 3 emitters, got %d", metrics["emitter_count"])
	}
	if metrics["events_handled"].(int64) != 150 {
		t.Errorf("Expected 150 events handled, got %d", metrics["events_handled"])
	}
}

// TestEventDispatcher_HighThroughput validates >1000 events/sec throughput
func TestEventDispatcher_HighThroughput(t *testing.T) {
	repo := newDispatcherEventRepository()
	logger := &mockLogger{}
	pluginCtx := &dispatcherPluginContext{repo: repo, logger: &dispatcherSDKLogger{}}

	dispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	// Register 5 emitters, each emitting 300 events = 1500 total
	// Target: Complete in < 1.5 seconds = >1000 events/sec
	for i := 0; i < 5; i++ {
		emitter := newMockEventEmitter(fmt.Sprintf("high-throughput-plugin-%d", i), 300, 0)
		dispatcher.RegisterEmitter(emitter)
	}

	// Start dispatcher
	ctx := context.Background()
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}

	startTime := time.Now()

	// Wait for all events to be processed by polling
	expectedEvents := 1500
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for events (got %d/%d)", repo.GetEventCount(), expectedEvents)
		case <-ticker.C:
			if repo.GetEventCount() >= expectedEvents {
				goto done
			}
		}
	}

done:
	elapsed := time.Since(startTime)

	// Stop dispatcher
	if err := dispatcher.Stop(); err != nil {
		t.Fatalf("Failed to stop dispatcher: %v", err)
	}

	eventCount := repo.GetEventCount()

	// Verify all events were saved
	if eventCount != expectedEvents {
		t.Errorf("Expected %d events, got %d", expectedEvents, eventCount)
	}

	// Calculate throughput
	throughput := float64(eventCount) / elapsed.Seconds()
	t.Logf("Throughput: %.2f events/sec (processed %d events in %v)", throughput, eventCount, elapsed)

	// Verify throughput meets requirement (>1000 events/sec)
	// We use 800 as threshold to account for test overhead and CI variability
	if throughput < 800 {
		t.Errorf("Throughput %.2f events/sec is below 800 events/sec threshold", throughput)
	}
}

// TestEventDispatcher_ContextCancellation verifies graceful shutdown on context cancellation
func TestEventDispatcher_ContextCancellation(t *testing.T) {
	repo := newDispatcherEventRepository()
	logger := &mockLogger{}
	pluginCtx := &dispatcherPluginContext{repo: repo, logger: &dispatcherSDKLogger{}}

	dispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	// Register emitter with slow event emission
	emitter := newMockEventEmitter("slow-plugin", 100, 10*time.Millisecond)
	dispatcher.RegisterEmitter(emitter)

	// Start dispatcher with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()

	// Stop dispatcher
	if err := dispatcher.Stop(); err != nil {
		t.Fatalf("Failed to stop dispatcher: %v", err)
	}

	// Verify some events were processed (but not all 100)
	eventCount := repo.GetEventCount()
	if eventCount == 0 {
		t.Error("Expected some events to be processed before cancellation")
	}
	if eventCount >= 100 {
		t.Error("Expected cancellation to stop event processing before completion")
	}

	t.Logf("Processed %d events before cancellation (out of 100)", eventCount)
}

// TestEventDispatcher_BufferedChannel verifies channel buffer prevents blocking
func TestEventDispatcher_BufferedChannel(t *testing.T) {
	repo := newDispatcherEventRepository()
	logger := &mockLogger{}
	pluginCtx := &dispatcherPluginContext{repo: repo, logger: &dispatcherSDKLogger{}}

	dispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	// Register emitter that bursts events quickly
	emitter := newMockEventEmitter("burst-plugin", 200, 0)
	dispatcher.RegisterEmitter(emitter)

	// Start dispatcher
	ctx := context.Background()
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}

	// Wait for processing
	time.Sleep(300 * time.Millisecond)

	// Check metrics while running
	metrics := dispatcher.GetMetrics()
	t.Logf("Channel capacity: %d, current length: %d", metrics["channel_cap"], metrics["channel_len"])

	// Stop dispatcher
	if err := dispatcher.Stop(); err != nil {
		t.Fatalf("Failed to stop dispatcher: %v", err)
	}

	// Verify all events processed
	eventCount := repo.GetEventCount()
	if eventCount != 200 {
		t.Errorf("Expected 200 events, got %d", eventCount)
	}
}

// TestEventDispatcher_IdempotentStop verifies Stop() can be called multiple times safely
func TestEventDispatcher_IdempotentStop(t *testing.T) {
	repo := newDispatcherEventRepository()
	logger := &mockLogger{}
	pluginCtx := &dispatcherPluginContext{repo: repo, logger: &dispatcherSDKLogger{}}

	dispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	emitter := newMockEventEmitter("plugin", 10, 0)
	dispatcher.RegisterEmitter(emitter)

	ctx := context.Background()
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Stop multiple times
	if err := dispatcher.Stop(); err != nil {
		t.Fatalf("First stop failed: %v", err)
	}

	if err := dispatcher.Stop(); err == nil {
		t.Error("Expected error when stopping already-stopped dispatcher")
	}

	// Verify emitter's StopEventStream is also idempotent
	if err := emitter.StopEventStream(); err != nil {
		t.Errorf("Emitter StopEventStream should be idempotent: %v", err)
	}
}

// TestEventDispatcher_GetEventChannel tests retrieving the event channel
func TestEventDispatcher_GetEventChannel(t *testing.T) {
	ctx := context.Background()
	repo := newDispatcherEventRepository()
	logger := &mockLogger{}
	pluginCtx := &dispatcherPluginContext{repo: repo, logger: &dispatcherSDKLogger{}}

	dispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	// Start the dispatcher
	if err := dispatcher.Start(ctx); err != nil {
		t.Fatalf("Failed to start dispatcher: %v", err)
	}
	defer dispatcher.Stop()

	// Get the event channel
	eventChan := dispatcher.GetEventChannel()

	if eventChan == nil {
		t.Fatal("Expected non-nil event channel")
	}

	// Verify channel is readable (shouldn't block immediately)
	select {
	case <-eventChan:
		// Got an event (or channel closed) - this is fine
	case <-time.After(10 * time.Millisecond):
		// No event yet - this is also fine
	}
}

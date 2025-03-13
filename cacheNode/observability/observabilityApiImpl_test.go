package observability

import (
 "context"
 "sync"
 "testing"
 "time"
 "log/slog"
)

// Test NewObservabilityHandlerImpl
func TestNewObservabilityHandlerImpl(t *testing.T) {
 ctx := context.Background()
 handler := NewObservabilityHandlerImpl(ctx, 8080)

 if handler == nil {
  t.Errorf("Expected ObservabilityHandlerImpl instance, got nil")
 }

 if handler.promPort != 8080 {
  t.Errorf("Expected promPort 8080, got %d", handler.promPort)
 }
}

// Test event recording functions
func TestRecordEventFunctions(t *testing.T) {
 ctx, cancel := context.WithCancel(context.Background())
 defer cancel()

 handler := NewObservabilityHandlerImpl(ctx, 8080)

 go func() {
  handler.RecordEventFrontend(FrontendEvent{})
  handler.RecordEventBackend(BackendEvent{})
  handler.RecordEventStorage(StorageEvent{})
  handler.RecordEventStorageDiskMetrics(StorageDiskMetricsEvent{})
 }()

 // Allow time for events to be processed
 time.Sleep(100 * time.Millisecond)
}

// Test readEventFromChannel
func TestReadEventFromChannel(t *testing.T) {
 ctx, cancel := context.WithCancel(context.Background())
 defer cancel()

 var wg sync.WaitGroup
 handler := NewObservabilityHandlerImpl(ctx, 8080)

 wg.Add(1)
 go handler.readEventFromChannel(ctx, &wg)

 // Send different event types
 handler.events <- FrontendEvent{}
 handler.events <- BackendEvent{}
 handler.events <- StorageEvent{}
 handler.events <- StorageDiskMetricsEvent{}
 //handler.events <- MockEvent{} // Unknown event

 time.Sleep(100 * time.Millisecond)

 // Cancel to stop execution
 cancel()
 close(handler.events)
 wg.Wait()

 slog.Info("TestReadEventFromChannel completed")
}

// Test runPromAgg
func TestRunPromAgg(t *testing.T) {
 ctx, cancel := context.WithCancel(context.Background())
 defer cancel()

 var wg sync.WaitGroup
 handler := NewObservabilityHandlerImpl(ctx, 9090)

 wg.Add(1)
 go handler.runPromAgg(ctx, &wg)

 // Wait to ensure the server starts
 time.Sleep(100 * time.Millisecond)

 // Cancel to stop execution
 cancel()
 wg.Wait()

 slog.Info("TestRunPromAgg completed")
}

// Test runPromAgg with invalid port
func TestRunPromAggInvalidPort(t *testing.T) {
 ctx, cancel := context.WithCancel(context.Background())
 defer cancel()

 var wg sync.WaitGroup
 handler := NewObservabilityHandlerImpl(ctx, 0) // Invalid port

 wg.Add(1)
 go handler.runPromAgg(ctx, &wg)

 time.Sleep(100 * time.Millisecond)
 cancel()
 wg.Wait()

 slog.Info("TestRunPromAggInvalidPort completed")
}

func TestReadEventFromChannelWithContextCancel(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
   
    var wg sync.WaitGroup
    handler := NewObservabilityHandlerImpl(ctx, 8080)
   
    wg.Add(1)
    go handler.readEventFromChannel(ctx, &wg)
   
    time.Sleep(50 * time.Millisecond)
    cancel() // Ensure exit condition is tested
   
    wg.Wait()
}
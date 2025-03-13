package observability

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ObservabilityHandlerImpl struct {
	ctx      context.Context
	events   chan interface{}
	promPort uint
}

func NewObservabilityHandlerImpl(ctx context.Context, port uint) *ObservabilityHandlerImpl {
	return &ObservabilityHandlerImpl{
		events:   make(chan interface{}, 10),
		promPort: port,
		ctx:      ctx,
	}
}

func (o *ObservabilityHandlerImpl) RecordEventFrontend(event FrontendEvent) {
	slog.Info("frontend event received at observability API")
	select {
	case o.events <- event:
	case <-o.ctx.Done():
	}
}
func (o *ObservabilityHandlerImpl) RecordEventBackend(event BackendEvent) {
	slog.Info("backend event received at observability API")
	select {
	case o.events <- event:
	case <-o.ctx.Done():
	}
}
func (o *ObservabilityHandlerImpl) RecordEventStorage(event StorageEvent) {
	slog.Info("storage event received at observability API")
	select {
	case o.events <- event:
	case <-o.ctx.Done():
	}
}

func (o *ObservabilityHandlerImpl) RecordEventStorageDiskMetrics(event StorageDiskMetricsEvent) {
	slog.Info("storage disk metrics event received at observability API")
	select {
	case o.events <- event:
	case <-o.ctx.Done():
	}
}

func (o *ObservabilityHandlerImpl) runPromAgg(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure the wait group counter is decremented when the function finishes.
	if o.promPort <= 0 {
		slog.Info("Not starting prometheus server...", "port", o.promPort)
		return
	}
	defer func() {
		slog.Info("prometheus server exiting")
	}()
	slog.Info("starting prometheus server...", "port", o.promPort)
	m := mux.NewRouter()
	m.Path("/metrics").Handler(promhttp.Handler())
	// Start the HTTP server
	server := http.Server{
		Addr:    fmt.Sprintf(":%v", o.promPort),
		Handler: m,
	}
	wg.Add(1)
	go func() {
		wg.Done()
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ran into an issue: %v\n", err.Error())
		}
	}()
	<-ctx.Done()
	server.Shutdown(context.Background())
}

func (o *ObservabilityHandlerImpl) readEventFromChannel(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure the wait group counter is decremented when the function finishes.
	defer func() {
		slog.Info("observabilityhandler exiting")
	}()
	slog.Info("starting observabilityhandler...")
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-o.events:
			switch e := event.(type) {
			case FrontendEvent:
				processsFrontendEvent(e)
				logFrontendEvent(e)
			case BackendEvent:
				processsBackendEvent(e)
				logBackendEvent(e)
			case StorageEvent:
				processsStorageEvent(e)
				logStorageEvent(e)
			case StorageDiskMetricsEvent:
				processsStorageDiskMetricsEvent(e)
				logStorageDiskMetricsEvent(e)
			default:
				slog.Warn("Unknown event type", "event", event)
			}
		}
	}
}

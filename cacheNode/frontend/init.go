package frontend

import (
	"context"
	"log/slog"
	"sync"

	"github.com/hcl/cdn/cacheNode/common"
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/cacheNode/observability"
)

// Initialize frontend and pass Stub Backend as a parameter
func Init(ctx context.Context, wg *sync.WaitGroup, cnfg *config.RunConfig, backend common.CachedRequestHandler, storage common.RequestHandler, observabilityHandler observability.ObservabilityHandler) error {
	
	slog.SetLogLoggerLevel(slog.LevelInfo)
	slog.Info("FE init.go : Init() - Start")

	// Initialize the observability handler if provided
	if observabilityHandler != nil {
		slog.Info("FE init.go : Init() - Observability Handler initialized.")
	}

	// Set up Frontend Configuration

	slog.Info("FE init.go : Init() - Initializing frontend configuration with validator...")
	// Initialize the Validator 
	validator := Validator{
		BackendRequest: backend, 
		ExpiredCnt:     0,
		UsableCnt:      0,
	}

	slog.Info("FE init.go : Init() - Initializing frontend configuration with finder...")

	// Initialize the Finder
	finder := Finder{
		StoragePath: storage,
		BackendPath: backend,
		ValidatePath: validator,
		TotalCnt:     0,
		GotFromStorageCnt: 0,
		GotFromBackendCnt: 0,
		FailedCnt:    0,
	}

	slog.Info("FE init.go : Init() - Initializing frontend configuration with collapser...")
	// Initialize the Collapser
	collapser := NewCollapser(&finder)

	slog.Info("FE init.go : Init() - Initializing frontend listener...")

	// Initialize the Listener
	listener := Listener{
		NextStep:   collapser,
		cfg:      cnfg,
		feObs: observabilityHandler,
	}

	wg.Add(1)

	slog.Info("FE init.go : Init() - Listener Go routine about to start...")
	// Start the listener in a goroutine
	go listener.ListenAndServe(ctx, wg, cnfg)

	slog.Info("FE init.go : Init() - End")
	return nil
}


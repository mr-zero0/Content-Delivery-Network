package observability

import (
	"context"
	"sync"

	"github.com/hcl/cdn/cacheNode/config"
	//"github.com/hcl/cdn/cacheNode/observability/prometheus_agg"
)

const DEFAULT_PROM_PORT uint = 9090

func Init(ctx context.Context, wg *sync.WaitGroup, cfg *config.RunConfig, port uint) (ObservabilityHandler, error) {
	impl := NewObservabilityHandlerImpl(ctx, port)
	//register all the prometheus objects with default prometheus server
	RegisterPromMetrics()
	// start the EventProcessor event data
	wg.Add(1)
	go impl.readEventFromChannel(ctx, wg)
	// start the Prometheus Server
	wg.Add(1)
	go impl.runPromAgg(ctx, wg)
	return impl, nil
}

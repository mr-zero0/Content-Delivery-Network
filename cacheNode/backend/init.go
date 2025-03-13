package backend

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hcl/cdn/cacheNode/common"
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/cacheNode/observability"
)

var transport = http.Transport{
	IdleConnTimeout:       10 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
}

func Init(ctx context.Context, wg *sync.WaitGroup, cfg *config.RunConfig, storage common.RequestHandler, observabilityHanlder observability.ObservabilityHandler) (common.CachedRequestHandler, error) {
	if cfg.ParentIp() != "" {
		parentUrl := fmt.Sprintf("http://%s:%d", cfg.ParentIp(), cfg.ParentPort())
		slog.Info("BE ParentMapper: Assigned Parent url", "url", parentUrl)

		url, err := url.Parse(parentUrl)
		if err != nil {

		}
		transport.Proxy = http.ProxyURL(url)
	}
	return &Backend{ctx: ctx, wg: wg, cfg: cfg, store: storage, observabilityHanlder: observabilityHanlder}, nil
}

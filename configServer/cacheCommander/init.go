package cacheCommander

import (
	"context"
	"log/slog"
	"sync"

	"github.com/hcl/cdn/configServer/configSaver"
	"github.com/hcl/cdn/configServer/inMemoryConfig"
)

var inMemConfig *inMemoryConfig.InMemoryConfig
var bgContext context.Context
var bgWg *sync.WaitGroup

func Init(ctx context.Context, wg *sync.WaitGroup, c *inMemoryConfig.InMemoryConfig) {
	inMemConfig = c
	bgContext = ctx
	bgWg = wg
	pendingInvalidateRequests = make(map[string]string)
	slog.Info("Loading delivery services from file...")
	if err := configSaver.LoadDSFromFile(); err != nil {
		slog.Error("Error loading delivery services", "error", err)
	} else {
		slog.Info("Delivery services loaded successfully.")
	}

	slog.Info("Loading cache nodes from file...")
	if err := configSaver.LoadCNFromFile(); err != nil {
		slog.Error("Error loading cache nodes", "error", err)
	} else {
		slog.Info("Cache nodes loaded successfully.")
	}
}

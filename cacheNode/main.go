package main

import (
	"context"
	"flag"
	"log/slog"
	"path"
	"sync"

	"github.com/hcl/cdn/cacheNode/backend"
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/cacheNode/frontend"
	"github.com/hcl/cdn/cacheNode/mgmtApi"
	"github.com/hcl/cdn/cacheNode/observability"
	"github.com/hcl/cdn/cacheNode/storage"
	"github.com/hcl/cdn/common/helper"
	cMgmtApi "github.com/hcl/cdn/common/mgmtApi"
)

const CONFIG_FILENAME string = "config.json"
const CONFIG_DIR string = "."

var CDN_DIR string = path.Join(".", "cdn")
var mgmtPort *uint
var promPort *uint
var cdnDir *string
var configDir *string

func parseFlags() {
	promPort = flag.Uint("promPort", observability.DEFAULT_PROM_PORT, "Prometheus Server port")
	mgmtPort = flag.Uint("mgmtport", cMgmtApi.DEFAULT_MGMT_PORT, "MGMT Server port")
	configDir = flag.String("dir", CONFIG_DIR, "Directory to persist config files")
	cdnDir = flag.String("cdndir", CDN_DIR, "Directory to persist CDN cache content")
	flag.Parse()
	slog.Info("Flags:", "mgmtPort", *mgmtPort)
	slog.Info("Flags:", "promPort", *promPort)
	slog.Info("Flags:", "cdnDir", *cdnDir)
	slog.Info("Flags:", "configDir", *configDir)
}

func main() {
	//Parent cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Wait group to synchronize exit of all routines
	wg := sync.WaitGroup{}
	defer func() {
		slog.Info("Waiting for all to terminate")
		wg.Wait()
		slog.Info("Cache Node Exit.")
	}()

	helper.Init_userInterrupt(ctx, &wg, cancel)

	parseFlags()

	// Read config from file
	slog.Info("Reading config from file" + CONFIG_FILENAME)
	cfg, err := config.NewConfigFromFile(path.Join(*configDir, CONFIG_FILENAME))
	if err != nil {
		slog.Error("Error reading config file", "error", err)
	}

	cr := config.NewConfigReconciler(ctx, &wg, cfg)

	slog.Info("Starting MgmtApi Server")
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err = mgmtApi.Init(ctx, wg, cr, nil, *mgmtPort) //storage is set to nil
		if err != nil {
			slog.Error("Error reading config file", "error", err)
			cancel()
		}
	}(&wg)

	// Wait till valid config
	err = cfg.WaitForValidConfig(ctx)
	if err != nil {
		slog.Error("Error waiting for valid config", "error", err)
		return
	}

	//Init all packages
	observabilityHandler, err := observability.Init(ctx, &wg, cfg, *promPort)
	if err != nil {
		slog.Error("Error initing observablity", "error", err)
		return
	}

	storageHandler, err := storage.Init(ctx, &wg, *cdnDir, cfg, observabilityHandler)
	if err != nil {
		slog.Error("Error initing storage", "error", err)
		return
	}

	mgmtApi.SetStorage(storageHandler)

	backendHandler, err := backend.Init(ctx, &wg, cfg, storageHandler, observabilityHandler)
	if err != nil {
		slog.Error("Error initing backend", "error", err)
		return
	}

	err = frontend.Init(ctx, &wg, cfg, backendHandler, storageHandler, observabilityHandler)
	if err != nil {
		slog.Error("Error initing frontend", "error", err)
		return
	}
	slog.Info("All Init Complete.")
}

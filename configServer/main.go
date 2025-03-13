package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/hcl/cdn/common/helper"
	"github.com/hcl/cdn/configServer/apiServer"
	"github.com/hcl/cdn/configServer/cacheCommander"
	"github.com/hcl/cdn/configServer/configPusher"
	"github.com/hcl/cdn/configServer/configSaver"
	"github.com/hcl/cdn/configServer/inMemoryConfig"
	"github.com/hcl/cdn/configServer/uiServer"
)

const SAVEDIR = "."

var apiServerPort string
var saveDir string

func parseFlags() {
	apiPort := flag.Uint("apiport", 8080, "API Server port")
	dir := flag.String("dir", SAVEDIR, "Directory to persist config files")
	flag.Parse()
	apiServerPort = fmt.Sprintf("%d", *apiPort)
	slog.Info("Flags:", "apiPort", *apiPort, " apiServerPort", apiServerPort)
	slog.Info("Flags:", "Directory", *dir)
	saveDir = *dir
}

func main() {
	parseFlags()

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

	r := mux.NewRouter()
	inMemConfig := &inMemoryConfig.InMemoryConfig{}

	apiServer.Init(ctx, &wg, inMemConfig)
	configSaver.Init(ctx, &wg, inMemConfig, saveDir)
	configPusher.Init(ctx, &wg, inMemConfig)
	cacheCommander.Init(ctx, &wg, inMemConfig)

	apiServer.RegisterHandlers(r)
	uiServer.RegisterHandlers(r)

	srv := http.Server{
		Addr:    ":" + apiServerPort,
		Handler: r,
	}

	// Start the server
	slog.Info("Server is running", "port", apiServerPort)
	wg.Add(1)
	go func() {
		wg.Done()
		log.Fatal(srv.ListenAndServe())
	}()
	<-ctx.Done()
	srv.Shutdown(context.Background())
}

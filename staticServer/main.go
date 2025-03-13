package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/hcl/cdn/common/helper"
)

var PORTS = []string{"32001", "32002"}

//go:embed static/*
var uiFiles embed.FS

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		slog.Info("Exiting...")
	}()

	var wg sync.WaitGroup
	defer wg.Wait()

	helper.Init_userInterrupt(ctx, &wg, cancel)

	handler := http.FileServer(http.FS(uiFiles))
	r := mux.Router{}
	r.PathPrefix("/static").Handler(handler).Methods("GET")
	r.PathPrefix("/dyna").Handler(NewDynamicHeaderHandler(nil)).Methods("GET")
	l := NewLogHandler(&r)

	for _, port := range PORTS {
		wg.Add(1)
		srv := http.Server{
			Addr:    ":" + port,
			Handler: l,
		}
		slog.Info(fmt.Sprintf("Listening Static Server on %v", port))
		go func(srv *http.Server) {
			wg.Done()
			srv.ListenAndServe()
		}(&srv)
	}
	slog.Info("press ctrl+c to stop")
}

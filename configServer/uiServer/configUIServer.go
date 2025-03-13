package uiServer

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed ui/*
var uiFiles embed.FS

type MyFs struct {
	fs http.Handler
}

func (m MyFs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("ui Serve", "url", r.URL.String())
	m.fs.ServeHTTP(w, r)
}

func RegisterHandlers(r *mux.Router) {
	// Serve static files
	fs := MyFs{
		fs: http.FileServerFS(uiFiles),
	}
	slog.Info("Added /ui")
	r.PathPrefix("/ui").Handler(fs).Methods("GET")
}

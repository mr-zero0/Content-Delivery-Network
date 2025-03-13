package main

import (
	"log/slog"
	"net/http"
)

type LogHandler struct {
	h http.Handler
}

func NewLogHandler(h http.Handler) *LogHandler {
	return &LogHandler{h: h}
}

func (l LogHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	slog.Info("Request", "host", req.Host, "url", req.URL.String())
	if l.h == nil {
		w.WriteHeader(http.StatusOK)
	}
	l.h.ServeHTTP(w, req)
}

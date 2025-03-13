package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type DynamicHeaderHandler struct {
	h http.Handler
}

func NewDynamicHeaderHandler(h http.Handler) *DynamicHeaderHandler {
	return &DynamicHeaderHandler{h: h}
}

func (l DynamicHeaderHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path, "/")
	slog.Info("Dyna Parts " + strings.Join(parts, ","))
	if l.h == nil {
		var status int
		var length int
		status = http.StatusOK
		defer func(l *int) {
			for i := 0; i < length; i++ {
				w.Write([]byte("="))
			}
		}(&length)
		defer func(v *int) {
			if *v == 0 {
				*v = http.StatusOK
			}
			w.WriteHeader(*v)
		}(&status)
		pos := 2
		if len(parts) <= pos {
			return
		}
		status, _ = strconv.Atoi(parts[pos])
		pos++
		if pos >= len(parts) {
			return
		}
		length, _ = strconv.Atoi(parts[pos])
		defer w.Header().Set("Content-Type", "text/plain")
		pos++
		if pos >= len(parts) {
			return
		}
		w.Header().Set("Cache-control", fmt.Sprintf("max-age=%s", parts[pos]))
		pos++
		if pos >= len(parts) {
			return
		}
		w.Header().Set("Age", parts[pos])
	}
}

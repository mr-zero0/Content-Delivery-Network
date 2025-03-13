package backend

import (
	"context"
	"io"
	"log/slog"
	"net/http"
)

func getObject(ctx context.Context, req *http.Request) (response *http.Response, err error) {
	slog.Info("BE Fetcher:", "req", req)

	client := http.Client{Transport: &transport}
	newReq := req.WithContext(ctx)
	response, err = client.Do(newReq)
	if err != nil {
		var body string
		if response != nil && response.Body != nil {
			bodybytes, _ := io.ReadAll(response.Body)
			defer response.Body.Close()
			body = string(bodybytes)
		}
		slog.Error("BE Fetcher: Error during HTTP request", "error", err, "body", body)
		return
	}
	slog.Info("BE Fetcher: Received HTTP response with status", "status", response.Status)
	return
}

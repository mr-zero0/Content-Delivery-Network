package backend

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/hcl/cdn/cacheNode/common"
	"github.com/hcl/cdn/common/helper"
)

func save(ctx context.Context, resp *http.Response, req *http.Request, store common.RequestHandler) {
	if store == nil {
		slog.Info("BE SAVER Nil Store ignoring save", "url", helper.GetString(req))
		return
	}
	slog.Info("BE SAVER  SaveOject Request", "url", helper.GetString(req))
	postreq, err := http.NewRequest("POST", helper.GetString(req), resp.Body)
	if err != nil {
		slog.Error("BE SAVER Error creating post request to store", "error", err)
		return
	}
	postreq.Header = resp.Header
	postreq = postreq.WithContext(ctx)
	storeResp, err := store.Do(postreq)
	if err != nil {
		slog.Error("BE SAVER Error saving to store", "error", err)
		return
	}
	slog.Info(" BE Saver Success", "url", helper.GetString(req), "status", storeResp.StatusCode)
}

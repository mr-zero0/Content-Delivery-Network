package storage

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/hcl/cdn/cacheNode/observability"
)

type StorageHandler struct {
	observabilityObj observability.ObservabilityHandler
	cacheManagerObj *cacheManager
}

func (storageObj *StorageHandler) Do(request *http.Request) (response *http.Response, err error) {
	response = &http.Response{}
	if request == nil {
		slog.Error("Storage:Do:http.request is nil")
		response.StatusCode = http.StatusBadRequest
		err = errors.New("http.request is nil")
		return
	}

	switch request.Method {
	case http.MethodGet, http.MethodHead:
		response, err = reader(request, storageObj)
	case http.MethodPost:
		response, err = writer(request, storageObj)
	case http.MethodDelete:
		response, err = invalidator(request, storageObj)
	default:
		slog.Error("Storage:Do:Method not supported", "method", request.Method)
		err = errors.New(request.Method + " method not supported")
		response.StatusCode = http.StatusMethodNotAllowed
	}

	return
}

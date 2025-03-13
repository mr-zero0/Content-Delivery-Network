package frontend

import (
	"net/http"
	"log/slog"
	"strings"
	"errors"
	"github.com/hcl/cdn/cacheNode/common"
	"io"
)

type Finder struct{
	StoragePath common.RequestHandler
	BackendPath	common.CachedRequestHandler
	ValidatePath Validator
	TotalCnt, GotFromStorageCnt, GotFromBackendCnt, FailedCnt uint32
}

func (f Finder) Do(req *http.Request) (*http.Response, error) {
	slog.Info("FE finder.go : Do() - Start")
	f.TotalCnt++

	errRsp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status: 	"500 Internal Server Error",
		Header:     req.Header,
		Body:       http.NoBody,
		ContentLength: 0,
		Request:    req,
	}

	var errStr error
	// 1. Execute StoragePath.Do
	stRsp, err := f.StoragePath.Do(req)
	if err != nil {
		slog.Error("FE finder.go : Storage Do() - Error - ", "error", err)
		errRsp.ContentLength = int64(len(err.Error()))
		errRsp.Body = io.NopCloser(strings.NewReader(err.Error()))
		return errRsp, nil
	} else {

		switch stRsp.StatusCode {
		case http.StatusOK:
			respV, err := f.ValidatePath.Do(req, stRsp)
			if err != nil {
				slog.Error("FE finder.go : Do() - Validator.Do() Failure ", "error", err)
				errRsp.ContentLength = int64(len(err.Error()))
				errRsp.Body = io.NopCloser(strings.NewReader(err.Error()))
				return errRsp, nil
			} else {
				slog.Info("FE finder.go : Do() - Finder Validator Response Obtained")
				return respV, nil
			}

		case http.StatusInternalServerError:
			slog.Info("FE finder.go : Do() - Storage.Do() Internal Server Error")
			return stRsp, nil

		case http.StatusNotFound:
			slog.Info("FE finder.go : Resource not found in storage, Calling Backend.Do()")
			respB, err := f.BackendPath.Do(req)
			if err != nil {
				slog.Error("FE finder.go : Do() - Backend.Do() Failure ", "error", err)
				errRsp.ContentLength = int64(len(err.Error()))
				errRsp.Body = io.NopCloser(strings.NewReader(err.Error()))
				return errRsp, nil
			} else {
				slog.Info("FE finder.go : Do() - Finder Backend Response Obtained")
				return respB, nil
			}

		default:
			errStr = errors.New("Internal Server Error")
			errRsp.ContentLength = int64(len(errStr.Error()))
			errRsp.Body = io.NopCloser(strings.NewReader(errStr.Error()))
			slog.Error("FE finder.go : Storage Do() - Unsupported Status Code : ", "statusCode", stRsp.StatusCode)
		}
	}

	slog.Info("FE finder.go : Do() - End (ERROR)")
	return errRsp, nil
}
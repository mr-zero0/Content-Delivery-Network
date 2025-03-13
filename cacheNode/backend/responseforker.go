package backend

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/hcl/cdn/common/helper"
)

func forker(resp *http.Response) (origresp *http.Response, copyresp *http.Response, err error) {
	slog.Info(" BE forker receives the response", "url", helper.GetString(resp.Request))
	dupReader := helper.NewDupReadCloser(resp.Body)
	//copying values
	respNew := http.Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Proto:            resp.Proto,
		ProtoMajor:       resp.ProtoMajor,
		ProtoMinor:       resp.ProtoMinor,
		Header:           resp.Header,
		Body:             nil,
		ContentLength:    resp.ContentLength,
		TransferEncoding: resp.TransferEncoding,
		Close:            resp.Close,
		Uncompressed:     resp.Uncompressed,
		Trailer:          resp.Trailer,
		Request:          resp.Request,
		TLS:              resp.TLS,
	}
	respNew.Body = io.NopCloser(dupReader.DupRdr())
	resp.Body = io.NopCloser(dupReader)
	slog.Info(" BE forker responseBody is updated. ")
	return resp, &respNew, nil
}

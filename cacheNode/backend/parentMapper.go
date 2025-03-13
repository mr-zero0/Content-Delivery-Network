package backend

import (

	//"fmt"

	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hcl/cdn/cacheNode/config"
	coCfg "github.com/hcl/cdn/common/config"
	"github.com/hcl/cdn/common/helper"
)

func ParentMapper(ctx context.Context, req *http.Request, cfg *config.RunConfig) (mappedReq *http.Request, response *http.Response, ds *coCfg.DeliveryService, err error) {
	mappedReq = req.Clone(ctx)
	mappedReq.RequestURI = ""
	mappedReq.Host = ""

	//takes the http request as input
	slog.Info("BE ParentMapper: Received a request", "url", helper.GetString(req))
	//route info from configteam
	ds, err = cfg.DSLookup(req) //look for Delivery serive
	if err != nil {
		response = &http.Response{}
		response.StatusCode = http.StatusNotFound
		response.Status = strconv.Itoa(http.StatusNotFound) + " Content Not Found"
		return
	}
	if cfg.ParentIp() == "" {
		orign_url, err1 := url.Parse(ds.OriginURL)
		if err1 != nil {
			response = &http.Response{}
			response.StatusCode = http.StatusNotFound
			response.Status = strconv.Itoa(http.StatusNotFound) + " Content Not Found"
			return
		}
		mappedReq.URL.Scheme = orign_url.Scheme
		mappedReq.URL.Host = orign_url.Host
		slog.Info("BE ParentMapper: Assigned Origin url", "url", helper.GetString(mappedReq))
		return
	}
	return
}

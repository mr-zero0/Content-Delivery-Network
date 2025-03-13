package backend

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/hcl/cdn/cacheNode/common"
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/cacheNode/observability"
)

type Backend struct {
	ctx                  context.Context
	wg                   *sync.WaitGroup
	cfg                  *config.RunConfig
	store                common.RequestHandler
	observabilityHanlder observability.ObservabilityHandler
}

func (b *Backend) Do(req *http.Request) (response *http.Response, err error) {
	modReq, response, ds, err := ParentMapper(b.ctx, req, b.cfg)
	if err != nil {
		return
	}
	fmt.Println(modReq)
	response, err = getObject(b.ctx, modReq)
	if err != nil {
		return
	}
	response, err = headerRewriter(req, response, ds)
	if err != nil {
		return
	}
	response, copyresp, err := forker(response)
	if err != nil {
		return
	}
	b.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		save(b.ctx, copyresp, req, b.store)
	}(b.wg)
	return
}

func (b *Backend) ReDo(req *http.Request, oldResp *http.Response) (response *http.Response, err error) {
	modReq, response, ds, err := ParentMapper(b.ctx, req, b.cfg)
	if err != nil {
		return
	}
	fmt.Println(modReq)
	//Add the Request Headers based on resp
	//IF oldResp had ETAG Header is available - include If-Match in modReq
	//if oldResp had Last-Modified-Since is available - include If-Unmodified-Since in modReq
	response, err = getObject(b.ctx, modReq)
	if err != nil {
		return
	}
	//304 Not Modified return
	// If recache required? should storage be updated?
	response, err = headerRewriter(req, response, ds)
	if err != nil {
		return
	}
	response, copyresp, err := forker(response)
	if err != nil {
		return
	}
	b.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		save(b.ctx, copyresp, req, b.store)
	}(b.wg)
	return
}

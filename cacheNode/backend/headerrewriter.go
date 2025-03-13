package backend

import (
	"fmt"
	"log/slog"
	"net/http"

	coCfg "github.com/hcl/cdn/common/config"
	"github.com/hcl/cdn/common/helper"
)

func printHeaders(headers http.Header) string {
	headersString := ""
	for key, values := range headers {
		for _, value := range values {
			headersString += fmt.Sprintf("%s: %s\n", key, value)
		}
	}
	return headersString
}

func headerRewriter(req *http.Request, resp *http.Response, ds *coCfg.DeliveryService) (updatedResponse *http.Response, err error) {
	slog.Info(" BE HeaderRewriter got the request:", "url", helper.GetString(req))
	updatedResponse = resp
	for _, rule := range ds.RewriteRules {
		switch rule.Operation {
		case coCfg.HdrReWriteOpAdd:
			updatedResponse.Header.Set(rule.HeaderName, rule.Value)
		case coCfg.HdrReWriteOpOverwrite:
			if _, exists := resp.Header[rule.HeaderName]; exists {
				updatedResponse.Header.Set(rule.HeaderName, rule.Value)
			}
		case coCfg.HdrReWriteOpDelete:
			updatedResponse.Header.Del(rule.HeaderName)
		default:
			slog.Error("BE HeaderRewriter invalid Action")
		}
	}

	headersString := printHeaders(resp.Header)
	slog.Info("BE  updated HeaderRewriter ", "header", headersString)

	return updatedResponse, nil
}

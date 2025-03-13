package common

import (
	"net/http"
)

type RequestHandler interface {
	Do(r *http.Request) (*http.Response, error)
}

type ReDoRequestHandler interface {
	ReDo(r *http.Request, oresp *http.Response) (*http.Response, error)
}
type CachedRequestHandler interface {
	RequestHandler
	ReDoRequestHandler
}

/*
type MyRequestHandler struct{}

func (h *MyRequestHandler) Do(r *http.Request) (http.Response, error) {
    client := &http.Client{}
    resp, err := client.Do(r)
    if err != nil {
        return http.Response{}, err
    }
    return *resp, nil
}
*/

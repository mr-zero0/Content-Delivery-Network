package mock

import (
	"fmt"
	"net/http"
	"strings"
	"io"
)

type MockBackend struct {
	doCnt, reDoCnt int
}

// type MockBackendIF interface{
// 	Do(r *http.Request) (http.Response, error)
// 	ReDo(r *http.Request, oresp *http.Response) (http.Response, error)
// }

// Do method is stubbed for backend requests.
func (sb *MockBackend) Do(r *http.Request) (*http.Response, error) {
	sb.doCnt++
	body := "Request DO - Hello From Backend"
	isGet := true
	t := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(strings.NewReader(body)), // Fix here
		ContentLength: int64(len(body)),
		Request:       r,
	}

	t1 := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          http.NoBody,
		ContentLength: 0,
		Request:       r,
	}

	host := r.Host

	// Simulate backend behavior based on request method and host
	switch r.Method {
	case http.MethodGet:
		if len(host) <= 6 {
			t.Header.Set("Age", "60")
			t.Header.Set("Cache-Control", "max-age=120, public")
		} else if len(host) == 7 {
			// Expired, host=xyz.com
			t.Header.Set("Age", "60")
			t.Header.Set("Cache-Control", "max-age=30, public")
		} else if len(host) == 8 {
			// Not Found, host=1234.com
			t.Status = "404 Not Found"
			t.StatusCode = 404
			t.Body = http.NoBody
			t.ContentLength = int64(0)
			t.Header.Set("Age", "0")
			t.Header.Set("Cache-Control", "max-age=0, public")
		}

	case http.MethodHead:
		isGet = false
		if len(host) <= 6 {
			t1.Header.Set("Age", "60")
			t1.Header.Set("Cache-Control", "max-age=120, public")
		} else if len(host) == 7 {
			// Expired, host=xyz.com
			t1.Header.Set("Age", "60")
			t1.Header.Set("Cache-Control", "max-age=30, public")
		} else if len(host) == 8 {
			// Not Found, host=1234.com
			t1.StatusCode = 404
			t1.Status = "404 Not Found"
			t1.Header.Set("Age", "0")
			t1.Header.Set("Cache-Control", "max-age=0, public")
		}

	default:
		return nil, fmt.Errorf("Method not supported")
	}

	if isGet {
		return t, nil
	}
	return t1, nil
}

// ReDo method for retrying requests
func (sbrd *StubBackend) ReDo(req *http.Request, oldResp *http.Response) (*http.Response, error) {
	isGet := true
	sbrd.reDoCnt++
	body := "ReRequest DO - Hello From Backend"
	t := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(strings.NewReader(body)), // Fix here
		ContentLength: int64(len(body)),
		Request:       req,
	}

	t1 := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          http.NoBody,
		ContentLength: 0,
		Request:       req,
	}

	host := req.Host

	// Simulate backend behavior based on request method and host
	switch req.Method {
	case http.MethodGet:
		if len(host) <= 6 {
			t.Header.Set("Age", "60")
			t.Header.Set("Cache-Control", "max-age=120, public")
		} else if len(host) == 7 {
			// Expired, host=xyz.com
			t.Header.Set("Age", "60")
			t.Header.Set("Cache-Control", "max-age=30, public")
		} else if len(host) == 8 {
			// Not Found, host=1234.com
			t.Status = "404 Not Found"
			t.StatusCode = 404
			t.Body = http.NoBody
			t.ContentLength = int64(0)
			t.Header.Set("Age", "0")
			t.Header.Set("Cache-Control", "max-age=0, public")
		}

	case http.MethodHead:
		isGet = false
		if len(host) <= 6 {
			t1.Header.Set("Age", "60")
			t1.Header.Set("Cache-Control", "max-age=120, public")
		} else if len(host) == 7 {
			// Expired, host=xyz.com
			t1.Header.Set("Age", "60")
			t1.Header.Set("Cache-Control", "max-age=30, public")
		} else if len(host) == 8 {
			// Not Found, host=1234.com
			t1.StatusCode = 404
			t1.Status = "404 Not Found"
			t1.Header.Set("Age", "0")
			t1.Header.Set("Cache-Control", "max-age=0, public")
		}

	default:
		return nil, fmt.Errorf("Method not supported")
	}

	if isGet {
		return t, nil
	}
	return t1, nil
}
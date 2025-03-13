package mock

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MockStorage struct {
	cnt int32
}

// type MockStorageIF interface{
// 	Do(r *http.Request) (http.Response, error)
// }

func (ms *MockStorage) Do(req *http.Request) (*http.Response, error) {
	fmt.Println("Storage Stub - Processing request : ", req.URL)
	isGet := true

	body := "Request Do - Hello From Storage"

	// Create the response for a GET request
	s := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(strings.NewReader(body)), // Use NopCloser to make it a ReadCloser
		ContentLength: int64(len(body)),
		Request:       req,
	}

	// Create the response for a HEAD request (no body)
	s1 := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          http.NoBody, // HEAD request should have no body
		ContentLength: 0,
		Request:       req,
	}

	host := req.Host

	// Simulate backend behavior based on request method and host
	switch req.Method {
	case http.MethodGet:
		// active, host = ab.com
		if len(host) <= 6 {
			s.Header.Set("Age", "60")
			s.Header.Set("Cache-Control", "max-age=120, public")
		} else if len(host) == 7 {
			// expired, host = xyz.com
			s.Header.Set("Age", "60")
			s.Header.Set("Cache-Control", "max-age=30, public")
		} else if len(host) == 8 {
			// Not Found, host = 1234.com
			s.Status = "404 Not Found"
			s.StatusCode = 404
			s.Body = http.NoBody
			s.Header.Set("Age", "0")
			s.Header.Set("Cache-Control", "max-age=0, public")
			s.ContentLength = 0
		}

	case http.MethodHead:
		isGet = false
		if len(host) <= 6 {
			// active, host=ab.com
			s1.Header.Set("Age", "60")
			s1.Header.Set("Cache-Control", "max-age=120, public")
		} else if len(host) == 7 {
			// Expired, host=xyz.com
			s1.Header.Set("Age", "60")
			s1.Header.Set("Cache-Control", "max-age=30, public")
			s1.ContentLength = 0
		} else if len(host) == 8 {
			// Not Found, host = 1234.com
			s1.Status = "404 Not Found"
			s1.StatusCode = 404
			s1.Header.Set("Age", "0")
			s1.Header.Set("Cache-Control", "max-age=0, public")
		}

	default:
		return nil, fmt.Errorf("storageStub::Do: Method not supported")
	}

	if isGet {
		return s, nil
	}
	return s1, nil
}

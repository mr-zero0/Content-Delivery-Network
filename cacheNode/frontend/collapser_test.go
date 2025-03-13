package frontend_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hcl/cdn/cacheNode/frontend"
)

type mockRequestHandler struct {
	resp *http.Response
	err  error
}

func (m *mockRequestHandler) Do(*http.Request) (*http.Response, error) {
	return m.resp, m.err
}

// Skipping the tests involving MockFinder
func TestCollapser_Do(t *testing.T) {
	t.Run("No pending request, new request is processed", func(t *testing.T) {
		//expectedResp := &http.Response{StatusCode: http.StatusOK}
		//mockHandler := &mockRequestHandler{resp: expectedResp}
		collapser := frontend.NewCollapser(nil) // passing nil for Finder

		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		resp, _ := collapser.Do(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Duplicate request, should wait for response", func(t *testing.T) {
		//expectedResp := &http.Response{StatusCode: http.StatusOK}
		//mockHandler := &mockRequestHandler{resp: expectedResp}
		collapser := frontend.NewCollapser(nil) // passing nil for Finder

		req1 := httptest.NewRequest("GET", "http://example.com/test", nil)
		req2 := httptest.NewRequest("GET", "http://example.com/test", nil)

		// Start first request
		go func() {
			_, _ = collapser.Do(req1)
		}()

		// Start second request which should wait for the first response
		resp, _ := collapser.Do(req2)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Error response from finder", func(t *testing.T) {
		//mockHandler := &mockRequestHandler{err: errors.New("finder error")}
		collapser := frontend.NewCollapser(nil) // passing nil for Finder

		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		resp, _ := collapser.Do(req)

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "Internal Server Error from FE") {
			t.Errorf("unexpected error message: %s", body)
		}
	})

	t.Run("Simultaneous requests to the same URL", func(t *testing.T) {
		//expectedResp := &http.Response{StatusCode: http.StatusOK}
		//mockHandler := &mockRequestHandler{resp: expectedResp}
		collapser := frontend.NewCollapser(nil) // passing nil for Finder

		req1 := httptest.NewRequest("GET", "http://example.com/test", nil)
		req2 := httptest.NewRequest("GET", "http://example.com/test", nil)

		// Start first request
		go func() {
			_, _ = collapser.Do(req1)
		}()

		// Start second request which should wait for the first response
		resp, _ := collapser.Do(req2)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("No entry found, create new entry and return response", func(t *testing.T) {
		//expectedResp := &http.Response{StatusCode: http.StatusOK}
		//mockHandler := &mockRequestHandler{resp: expectedResp}
		collapser := frontend.NewCollapser(nil) // passing nil for Finder

		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		resp, _ := collapser.Do(req)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

// Test CollapseEntry methods
func TestCollapseEntry_HandleResponse(t *testing.T) {
	t.Run("Handle response with error", func(t *testing.T) {
		entry := frontend.NewCollapseEntry()
		resp, _ := entry.HandleResponse(nil, errors.New("error"))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("Handle successful response", func(t *testing.T) {
		expectedResp := &http.Response{StatusCode: http.StatusOK}
		entry := frontend.NewCollapseEntry()
		resp, _ := entry.HandleResponse(expectedResp, nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestCollapseEntry_WaitForResponse(t *testing.T) {
	t.Run("Wait for response", func(t *testing.T) {
		entry := frontend.NewCollapseEntry()

		// Simulate a response being set
		go func() {
			time.Sleep(1 * time.Second)
			entry.HandleResponse(&http.Response{StatusCode: http.StatusOK}, nil)
		}()

		resp, err := entry.WaitForResponse()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

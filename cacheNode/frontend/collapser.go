package frontend

import (
	"io"
	"time"
	"net/http"
	"sync"
	"log/slog" 
	"strings"
)

// HttpResponse wraps an HTTP response and any associated error.
type HttpResponse struct {
	resp *http.Response
	err  error
}

// CollapseEntry represents a request entry used to collapse duplicate requests.
// It uses a channel to signal when a response is ready and tracks pending requests.
type CollapseEntry struct {
	responseReady chan HttpResponse // Channel to signal response readiness.
	pendingCount  int               // Number of pending requests waiting for this response.
}

// NewCollapseEntry initializes and returns a new CollapseEntry instance.
func NewCollapseEntry() *CollapseEntry {
	entry := &CollapseEntry{
		responseReady: make(chan HttpResponse),
		pendingCount:  0,
	}
	slog.Info("FE collapser.go : New CollapseEntry Instance Added")
	return entry
}

// handleResponse processes the response for a collapsed request.
// It ensures all waiting requests receive the same response.
func (e *CollapseEntry) HandleResponse(resp *http.Response, err error) (*http.Response, error) {
	slog.Info("FE collapser.go : handleResponse() - Start")
	
	respErr := &http.Response{
		Status:     "500 Internal Server Error",
		StatusCode: http.StatusInternalServerError,
		Header:     make(http.Header),
	}

	var errStr string
	errStr = "Internal Server Error from FE"

	respErr.Header.Set("Accept", "*/*")
	currentTime := time.Now().UTC().Format(http.TimeFormat)
	respErr.Header.Set("Date", currentTime)
	respErr.Header.Set("User-Agent", "Go Server Agent/1.23.4")
	respErr.Body = io.NopCloser(strings.NewReader(errStr))
	respErr.ContentLength = int64(len(errStr))

	if err != nil {
		for i := e.pendingCount; i > 0; i-- {
			e.responseReady <- HttpResponse{resp: respErr, err: nil}
		}
		slog.Info("FE collapser.go : handleResponse() - End with error")
		return respErr, nil
	}

	respRdr := resp.Body
	for i := e.pendingCount; i > 0; i-- {
		rdr, wtr := io.Pipe()
		go func() {
			defer wtr.Close()
			// Not handling the error for io.Copy() below
			io.Copy(wtr, respRdr)
		}()
		e.responseReady <- HttpResponse{
			resp: &http.Response{
				Status:     resp.Status,
				StatusCode: resp.StatusCode,
				Header:     resp.Header,
				Body:       rdr,
			},
			err: nil,
		}
		respRdr = io.NopCloser(rdr)
	}
	resp.Body = respRdr

	slog.Info("FE collapser.go : handleResponse() - End")
	return resp, nil
}

// waitForResponse waits for the response to be ready and returns it.
// This method increments the pending count for this entry.
func (e *CollapseEntry) WaitForResponse() (*http.Response, error) {
	slog.Info("FE collapser.go : waitForResponse() - Start")
	e.pendingCount++
	// This will get blocked here until the response is receied via the responseReady channel
	respObj := <-e.responseReady
	slog.Info("FE collapser.go : Got the response & waitForResponse() - End")
	return respObj.resp, respObj.err
}

// Collapser prevents duplicate requests to the same URL.
// It tracks pending requests and shares their responses with other requests to the same URL.
type Collapser struct {
	Next            *Finder    					// Downstream handler for executing requests.
	pendingMapMutex sync.RWMutex    			// Mutex to synchronize access to the pendingReq map.           
	pendingReq      map[string]*CollapseEntry 	// Map of URLs to their CollapseEntry.
}

// NewCollapser initializes and returns a new Collapser instance.
func NewCollapser(next *Finder) *Collapser {
	//slog.Info("FE collapser.go : NewCollapser() - Start")
	c := &Collapser{
		Next:  next,
		pendingReq: make(map[string]*CollapseEntry),
	}
	slog.Info("FE collapser.go : NewCollapser() - Done")
	return c
}

// Do processes an HTTP request, collapsing duplicate requests to the same URL.
func (c *Collapser) Do(r *http.Request) (*http.Response, error) {
	slog.Info("FE collapser.go : Do() - Start")
	var entry *CollapseEntry
	var reqUrlStr string
	reqUrlStr = "http://" + r.Host + r.URL.Path
	c.pendingMapMutex.RLock()
	entry, ok := c.pendingReq[reqUrlStr]
	c.pendingMapMutex.RUnlock()

	if !ok {
		slog.Info("FE collapser.go : Do() - No existing entry found, creating new one")
		c.pendingMapMutex.Lock()
		entry, ok = c.pendingReq[reqUrlStr]
		if !ok {
			entry = NewCollapseEntry()
			c.pendingReq[reqUrlStr] = entry
		}
		c.pendingMapMutex.Unlock()

		slog.Info("FE collapser.go : Do() - Invoking finder.Do() ")
		//slog.Info("FE collapser.go : Sleeping for 5 seconds...")
		time.Sleep(0*time.Second)
		resp, err := c.Next.Do(r)

		c.pendingMapMutex.Lock()
		delete(c.pendingReq, reqUrlStr)
		c.pendingMapMutex.Unlock()

		slog.Info("FE collapser.go : Do() - End with response")
		return entry.HandleResponse(resp, err)
	}

	slog.Info("FE collapser.go : Do() - Waiting for response")
	return entry.WaitForResponse()
}
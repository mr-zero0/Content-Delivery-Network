package frontend

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"strings"
	"time"
	
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/common/helper"
	"github.com/hcl/cdn/cacheNode/observability"
)

type Listener struct {
	NextStep	 *Collapser
	cfg      *config.RunConfig
	feObs observability.ObservabilityHandler
}

// SetConfigFile sets the configuration for the listener
func (l *Listener) SetConfigFile(cfgIp *config.RunConfig) {
	l.cfg = cfgIp
	slog.Info("FE listener.go : SetConfigFile() - Success")
}

// SendResponseToClient sends the response to the client
func (l *Listener) SendResponseToClient(respW http.ResponseWriter, rsp *http.Response, req *http.Request) error {
	slog.Info(fmt.Sprintf("FE listener.go : SendResponseToClient() for URL: %s", req.URL.String()))

	// Add headers to the ResponseWriter
	for key, values := range rsp.Header {
		for _, value := range values {
			respW.Header().Add(key, value)
		}
	}

	slog.Info("FE listener.go : Headers added to ResponseWriter:")
	for key, values := range respW.Header() {
		slog.Info(fmt.Sprintf("FE listener.go : %s: %v", key, values))
	}

	// Set the response status code
	respW.WriteHeader(rsp.StatusCode)

	if rsp.Body != nil {
		defer rsp.Body.Close()
		slog.Info("Writing response Body")
		_, err := io.Copy(respW, rsp.Body)
		if err != nil {
			slog.Info("FE listener.go : Error copying from Listener's Response to Client's Response:", "error", err.Error())
			slog.Info("FE listener.go : Quiting...")
			// Should we send a response as http>internalServerError ?
			return err	
		}
	} else {
		slog.Info("FE listener.go : Response Body is nil")
	}
	slog.Info("FE listener.go : Response sent successfully")
	return nil
}


// IsMatchingClientUrlAvailable checks if the URL matches a client URL in the config
func (l *Listener) IsMatchingClientUrlAvailable(req *http.Request) (bool, error) {
	slog.Info(fmt.Sprintf("FE listener.go : Trying to find a match for URL: %s", helper.GetString(req)))

	configDS, err := l.cfg.DSLookup(req)
	if err != nil {
		slog.Error("FE listener.go : No DS Config found for given URL", "error", err.Error())
		return false, errors.New("no DS Config found for given URL")
	}
	slog.Info(fmt.Sprintf("FE listener.go : URL match found in the DS with name : - %v", configDS.Name))
	return true, nil
}

func (l *Listener) reformat(req *http.Request) {
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}
}

// Records the diagnostics associated with the completion of the client request
func RecordFrontendMetrics(clientIp string, urlStr string, 
	userAgent string, responseTime int, bytes int, statusCode int, 
	cacheHit bool, observabilityObj observability.ObservabilityHandler) {
	
	frontendEvent:= observability.FrontendEvent{
		Timestamp : time.Now(),         //time of the event 
		ClientIP :  clientIp,  			//client IP
		URL : urlStr,   				//URL of the content
		UserAgent : userAgent,  			// client user agent 
		ResponseTime : responseTime,  			//response time in miliseconds
		TTFB :   		0,			//time to first byte in milliseconds
		Bytes : bytes,  					//number of bytes of content
		StatusCode : statusCode,  			//http response code to client 
		CacheHit : cacheHit,  				// true or false

	}
	if observabilityObj != nil {
		observabilityObj.RecordEventFrontend(frontendEvent)
	}else{
		slog.Error("FE listener.go : No observability Handler registered")
	}
}

// HandleClientRequest handles incoming client requests
func (l *Listener) HandleClientRequest(respW http.ResponseWriter, req *http.Request) {
	slog.Info("FE listener.go : HandleClientRequest() - Start")
	startTime:=time.Now()
	clientIP := req.RemoteAddr
	var errStr error
	urlHostStr := req.Host

	var urlStr string
	urlStr = "http://" + req.Host + "/" + req.URL.Path
	slog.Info(urlStr)

	var sCode int
	var bodyLen int
	cHit := false
	userAgent := req.Header.Get("User-Agent")
	
	defer func(){
		endTime := time.Now()
		timeTaken := endTime.Sub(startTime)
		responseTime:= int(timeTaken.Milliseconds())
		statusCode := sCode
		RecordFrontendMetrics(clientIP, urlStr, userAgent, responseTime, bodyLen, statusCode, cHit, l.feObs)
		slog.Info(fmt.Sprintf("FE listener.go : Time taken to complete the requested url: %s in %d milliseconds" , urlStr, int(timeTaken.Milliseconds())))
	}()

	respErr := &http.Response{
		Status:     "400 Bad Request",
		StatusCode: http.StatusBadRequest,
		Header:     make(http.Header),
	}

	sCode = respErr.StatusCode
	respErr.Header.Set("Accept", "*/*")
	// Set the Date header
	currentTime := time.Now().UTC().Format(http.TimeFormat)
	// Format the date in RFC1123 format
	respErr.Header.Set("Date", currentTime)
	// Set the User-Agent header
	respErr.Header.Set("User-Agent", "Go Server Agent/1.23.4")

	//slog.Info(fmt.Sprintf("FE listener.go : Length of urlHostStr: %d", len(urlHostStr)))
	slog.Info(fmt.Sprintf("FE listener.go : Listener received request from client: with IP: %s, Method: %s, Scheme: %s, Host: %s, UrlString: %s", clientIP,  req.Method, req.URL.Scheme, req.Host, req.URL.String()))


	slog.Info(fmt.Sprintf("LookupUrl: %s", helper.GetString(req)))
	l.reformat(req)
	// Check if there is a matching URL in the DS config list
	dsValid, err := l.IsMatchingClientUrlAvailable(req)
	if err != nil {
		errStr = errors.New("404 : Resource Not Found")
		respErr.ContentLength = int64(len(errStr.Error()))
		respErr.Body = io.NopCloser(strings.NewReader(errStr.Error()))
		
		respErr.StatusCode = http.StatusNotFound
		respErr.Status = "404 Not Found"
		sCode = respErr.StatusCode
		bodyLen = int(respErr.ContentLength)
		_ = l.SendResponseToClient(respW, respErr, req)
		slog.Info("FE listener.go : No matching URL found")
		return
	}

	slog.Info(fmt.Sprintf("FE listener.go : Client URL found %v - NextStep.Do() method called", dsValid))

	slog.Info("FE listener.go : Attempting NextStep.Do() - Calling the Collapser")
	
	// Send request to Collapser
	resp, err := l.NextStep.Do(req)

	if err != nil {
		errStr = errors.New("500: Internal Server Error")
		respErr.ContentLength = int64(len(errStr.Error()))
		respErr.Body = io.NopCloser(strings.NewReader(errStr.Error()))

		slog.Error("FE listener.go : Error received for URL", "url", urlHostStr, "error", err.Error())
		respErr.StatusCode = http.StatusInternalServerError
		respErr.Status = "500 Internal Server Error"
		sCode = respErr.StatusCode
		bodyLen = int(respErr.ContentLength)
		_ = l.SendResponseToClient(respW, respErr, req)
		return
	}

	// Find if the response is from the storage cache or backend
	if resp.Header.Get("X-Is-Cached") != "" {
		cHit = true
		slog.Info("FE listener.go : cacheHit : True")

	} else {
		cHit = false
		slog.Info("FE listener.go : cacheHit : False")
	}

	sCode = resp.StatusCode
	bodyLen = int(resp.ContentLength)
	_ = l.SendResponseToClient(respW, resp, req)
	slog.Info("FE listener.go : HandleClientRequest() - End")
}

// WaitForCancellation waits for cancellation signals and shuts down the server
func (l Listener) WaitForCancellation(ctxt context.Context, wg *sync.WaitGroup, srvr *http.Server) error {
	defer wg.Done()
	// Wait for cancellation signal

	<-ctxt.Done()
	slog.Info("FE listener.go : Cancellation request received. Shutting down the HTTP server.")

	// Gracefully shut down the server
	errSrvr := srvr.Shutdown(ctxt)
	if errSrvr != nil {
		slog.Error("FE listener.go : Failed to shut down server", "error", errSrvr)
		return errSrvr
	}
	slog.Info("FE listener.go : Server shut down gracefully")
	return nil
}

// ListenAndServe starts the HTTP server
func (l *Listener) ListenAndServe(ctx context.Context, wg *sync.WaitGroup, cfgIp *config.RunConfig) {
	slog.Info("FE listener.go : ListenAndServe() - Start")
	defer func() {
		slog.Info("FE: ListenAndServe() exiting.")
	}()

	defer wg.Done()
	Port := cfgIp.BindingPort()
	IP := cfgIp.BindingIp()
	l.SetConfigFile(cfgIp)

	slog.Info(fmt.Sprintf("FE listener.go : Listening for client request on %s : %d", IP, Port))

	mux := http.NewServeMux()
	mux.HandleFunc("/", l.HandleClientRequest)
	listenerAddr := fmt.Sprintf("%s:%d", IP, Port)

	srvr := http.Server{
		Addr:    listenerAddr,
		Handler: mux,
	}

	slog.Info(fmt.Sprintf("FE listener.go : Listener will start listening on %v", srvr.Addr))

	wg.Add(1)
	// Run in the background and wait for cancellation signals
	go l.WaitForCancellation(ctx, wg, &srvr)

	slog.Info("FE listener.go : Listener about to serve the client request")

	err := srvr.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("FE listener.go : Server ran into an issue", "error", err.Error())
	}
	slog.Info("FE listener.go : ListenAndServe() - End")
}

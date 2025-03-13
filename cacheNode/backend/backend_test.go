package backend_test

import (
	//"bytes"

	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/hcl/cdn/cacheNode/backend"
	"github.com/hcl/cdn/cacheNode/backend/backendTestMock"
	"github.com/hcl/cdn/cacheNode/config"
	commonConfig "github.com/hcl/cdn/common/config"

	"github.com/hcl/cdn/cacheNode/observability"
	// "github.com/hcl/cdn/cacheNode/storage"
)

func TestBackend_Do(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request : %v", r)

		w.Header().Set("Connection", "close")
		w.Write([]byte(`{"message": "Hello, World!"}`))
	}))

	defer ts.Close()
	tsURL, _ := url.Parse(ts.URL)
	fmt.Println(tsURL)
	IPadd := tsURL.Hostname()
	Port, _ := strconv.Atoi(tsURL.Port())
	fmt.Println(IPadd, Port)

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	Rules := []commonConfig.RewriteRule{
		{
			HeaderName: "Backendheader",
			Operation:  commonConfig.HdrReWriteOpAdd, // Example operation
			Value:      "some value",                 // Example value
		},

		{
			HeaderName: "Backendheader",
			Operation:  commonConfig.HdrReWriteOpOverwrite, // Example operation
			Value:      "new value",                        // Example value
		},

		/*{
			HeaderName: "Backendheader",
			Operation:  commonConfig.HdrReWriteOpDelete, // Example operation
			Value:      "some value",                    // Example value
		},*/
		// Add more rules as needed
	}

	ds_tmp := commonConfig.DeliveryService{Name: "DS1", ClientURL: "http://example.com", OriginURL: "http://originurl.com", RewriteRules: Rules}
	dss := commonConfig.DeliveryServices{Version: 1, ServiceList: []commonConfig.DeliveryService{ds_tmp}}

	cfg := config.RunConfig{
		Valid:    true,
		Filename: "test",
		Node: &commonConfig.CacheNode{
			IP:         "192.168.1.1",
			Port:       8080,
			Type:       commonConfig.CacheNodeEdge,
			ParentIP:   IPadd,
			ParentPort: Port},
		ServiceList: &dss,
	}

	observabilityHandler, err := observability.Init(ctx, wg, &cfg, 0)
	if err != nil {
		fmt.Printf("Err:%v\n", err.Error())
		return
	}

	storageHandler := &backendTestMock.RequestHandlerMock{
		HttpStatuscode:      http.StatusOK,
		Header:              map[string][]string{},
		ResponseBodyContent: "FROM_STORE",
	}

	backhandler, err := backend.Init(ctx, wg, &cfg, storageHandler, observabilityHandler)
	if err != nil {
		fmt.Println("Backend Initialization failed")
		return
	}

	// Mock the http.Request
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, _ := backhandler.Do(req)

	body, err := io.ReadAll(resp.Body)
	fmt.Println("\nfetcher body", string(body))
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != `{"message": "Hello, World!"}` {
		t.Fatalf("Resp body content is not equal")
	}

	defer resp.Body.Close()
}

/*func TestBackendport_Do(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request: %v", r)
		w.Header().Set("Connection", "close")
		w.Write([]byte(`{"message": "Hello, World!"}`))
	})
	listener, err := net.Listen("tcp", "127.0.0.1:8080") // Change to your desired port
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	ts := httptest.NewUnstartedServer(handler)
	ts.Listener.Close()
	ts.Listener = listener
	ts.Start()
	fmt.Println("Test server running at:", ts.URL)

	defer ts.Close()
	tsURL, _ := url.Parse(ts.URL)
	fmt.Println(tsURL)
	IPadd := tsURL.Hostname()
	Port, _ := strconv.Atoi(tsURL.Port())
	fmt.Println(IPadd, Port)

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	ds_tmp := commonConfig.DeliveryService{Name: "DS1", ClientURL: "http://example.com", OriginURL: "http://originurl.com", RewriteRules: []commonConfig.RewriteRule{}}
	dss := commonConfig.DeliveryServices{Version: 1, ServiceList: []commonConfig.DeliveryService{ds_tmp}}

	cfg := config.RunConfig{
		Valid:    true,
		Filename: "test",
		Node: &commonConfig.CacheNode{
			IP:         "192.168.1.1",
			Port:       8080,
			Type:       commonConfig.CacheNodeEdge,
			ParentIP:   IPadd,
			ParentPort: Port,
		},
		ServiceList: &dss,
	}

	observabilityHandler, err := observability.Init(ctx, wg, &cfg, 0)
	if err != nil {
		fmt.Printf("Err:%v\n", err.Error())
		return
	}

	storageHandler := &backendTestMock.RequestHandlerMock{
		HttpStatuscode:      http.StatusOK,
		Header:              map[string][]string{},
		ResponseBodyContent: "FROM_STORE",
	}

	backhandler, err := backend.Init(ctx, wg, &cfg, storageHandler, observabilityHandler)
	if err != nil {
		fmt.Println("Backend Initialization failed")
		return
	}

	// Mock the http.Request
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, _ := backhandler.Do(req)

	body, err := io.ReadAll(resp.Body)
	fmt.Println("\nfetcher body", string(body))
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	if string(body) != `{"message": "Hello, World!"}` {
		t.Fatalf("Resp body content is not equal")
	}

	defer resp.Body.Close()

}
*/

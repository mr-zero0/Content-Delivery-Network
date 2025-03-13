package apiServer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"log"
	"context"
	"os"
	"sync"
	//"edge"
	"github.com/gorilla/mux"
	"github.com/hcl/cdn/common/config"
	"github.com/hcl/cdn/configServer/inMemoryConfig"
	"github.com/hcl/cdn/configServer/configPusher"
	"github.com/hcl/cdn/configServer/configSaver"
)
var configDir = "C:\\Users\\prathmesh.meena\\OneDrive - HCL TECHNOLOGIES LIMITED\\Desktop\\POC\\cdn\\testData"

func setup() {
	bgContext = context.Background()
    bgWg = &sync.WaitGroup{}
	// Ensure the config directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.Mkdir(configDir, os.ModePerm)
	}

	// Remove existing files to reset the version
	os.Remove(configDir + "/deliveryServices.json")
	os.Remove(configDir + "/cacheNodes.json")

	log.Println("Initializing in-memory configuration")
	inMemConfig = &inMemoryConfig.InMemoryConfig{}
	inMemConfig.AddDs(&config.DeliveryService{
		Name:      "service1",
		ClientURL: "http://client1.com",
		OriginURL: "http://origin1.com",
	})
	inMemConfig.AddCn(&config.CacheNode{
		IP:         "127.0.0.1",
		Port:       8080,
		Type:       config.CacheNodeEdge,
		ParentIP:   "127.0.0.1",
		ParentPort: 8081,
	})
	log.Println("In-memory configuration initialized with one delivery service")

	log.Println("Initializing configPusher")
	configPusher.Init(bgContext, bgWg, inMemConfig)
	log.Println("configPusher initialized")
	log.Println("Initializing configSaver")
	configSaver.Init(bgContext, bgWg, inMemConfig, configDir)
	log.Println("configSaver initialized")

	// Save the initial configuration to files
	log.Println("Saving initial configuration to files")
	configSaver.SaveDSToFile()
	configSaver.SaveCNToFile()
	log.Println("Initial configuration saved to files")


}
func TestHandleDeliveryServicesGet(t *testing.T) {
	setup()
	req, err := http.NewRequest("GET", "/ds", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDeliveryServicesGet)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := struct {
		Version     int      `json:"version"`
		ServiceList []string `json:"serviceList"`
	}{
		Version:     1,
		ServiceList: []string{"service1"},
	}
	var response struct {
		Version     int      `json:"version"`
		ServiceList []string `json:"serviceList"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Version != expected.Version || len(response.ServiceList) != len(expected.ServiceList) || response.ServiceList[0] != expected.ServiceList[0] {
		t.Errorf("handler returned unexpected body: got %v want %v", response, expected)
	}

}
func TestHandleDeliveryServicesPost(t *testing.T) {
	setup()
	log.Println("Starting TestHandleDeliveryServicesPost")

	newService := config.DeliveryService{
		Name:      "service2",
		ClientURL: "http://client2.com",
		OriginURL: "http://origin2.com",
		RewriteRules: []config.RewriteRule{
			{HeaderName: "Header1", Operation: 1, Value: "Value1"},
		},
	}
	body, _ := json.Marshal(newService)
	req, err := http.NewRequest("POST", "/ds", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Created new POST request for /ds")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDeliveryServicesPost)
	handler.ServeHTTP(rr, req)
	log.Println("Handled POST request for /ds")

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Delivery service added"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Verify the version and the number of delivery services
	version, serviceList := inMemConfig.GetDsNames()
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
	if len(serviceList) != 2 {
		t.Errorf("expected 2 delivery services, got %d", len(serviceList))
	}
	if serviceList[0] != "service1" || serviceList[1] != "service2" {
		t.Errorf("expected service1 and service2, got %v", serviceList)
	}

	log.Println("TestHandleDeliveryServicesPost completed successfully")
}

func TestHandleDeliveryServiceByNameGet(t *testing.T) {
	setup()
	log.Println("Starting TestHandleDeliveryServiceByNameGet")

	req, err := http.NewRequest("GET", "/ds/service1", nil)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Created new GET request for /ds/service1")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/ds/{name}", handleDeliveryServiceByNameGet).Methods("GET")
	router.ServeHTTP(rr, req)
	log.Println("Handled GET request for /ds/service1")

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response config.DeliveryService
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	expected := config.DeliveryService{Name: "service1", ClientURL: "http://client1.com", OriginURL: "http://origin1.com"}
	if response.Name != expected.Name || response.ClientURL != expected.ClientURL || response.OriginURL != expected.OriginURL {
		t.Errorf("handler returned unexpected body: got %v want %v", response, expected)
	}

	log.Println("TestHandleDeliveryServiceByNameGet completed successfully")
}

func TestHandleDeliveryServiceByNamePut(t *testing.T) {
	setup()
	log.Println("Starting TestHandleDeliveryServiceByNamePut")

	updatedService := config.DeliveryService{
		Name:      "service1",
		ClientURL: "http://client1-updated.com",
		OriginURL: "http://origin1-updated.com",
	}
	body, _ := json.Marshal(updatedService)
	req, err := http.NewRequest("PUT", "/ds/service1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Created new PUT request for /ds/service1")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/ds/{name}", handleDeliveryServiceByNamePut).Methods("PUT")
	router.ServeHTTP(rr, req)
	log.Println("Handled PUT request for /ds/service1")

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Delivery service updated"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Verify the version
	version, _ := inMemConfig.GetDsNames()
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}

	// Verify the updated service details
	service, err := inMemConfig.GetDsDetail("service1")
	if err != nil {
		t.Fatal(err)
	}
	if service.ClientURL != "http://client1-updated.com" || service.OriginURL != "http://origin1-updated.com" {
		t.Errorf("service details not updated correctly: got %v", service)
	}

	log.Println("TestHandleDeliveryServiceByNamePut completed successfully")
}

func TestHandleDeliveryServiceByNameDelete(t *testing.T) {
	setup()
	log.Println("Starting TestHandleDeliveryServiceByNameDelete")

	// Add a second service to ensure we have two services
	newService := config.DeliveryService{
		Name:      "service2",
		ClientURL: "http://client2.com",
		OriginURL: "http://origin2.com",
	}
	inMemConfig.AddDs(&newService)

	// Verify initial state
	version, serviceList := inMemConfig.GetDsNames()
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
	if len(serviceList) != 2 || serviceList[0] != "service1" || serviceList[1] != "service2" {
		t.Errorf("expected service1 and service2, got %v", serviceList)
	}

	// Perform DELETE operation
	req, err := http.NewRequest("DELETE", "/ds/service1", nil)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Created new DELETE request for /ds/service1")

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/ds/{name}", handleDeliveryServiceByNameDelete).Methods("DELETE")
	router.ServeHTTP(rr, req)
	log.Println("Handled DELETE request for /ds/service1")

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Delivery service deleted"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Verify the version and the remaining delivery services
	version, serviceList = inMemConfig.GetDsNames()
	if version != 3 {
		t.Errorf("expected version 3, got %d", version)
	}
	if len(serviceList) != 1 || serviceList[0] != "service2" {
		t.Errorf("expected service2 to be the only remaining service, got %v", serviceList)
	}

	log.Println("TestHandleDeliveryServiceByNameDelete completed successfully")
}

func TestHandleCacheNodesGet(t *testing.T) {
	setup()
	req, err := http.NewRequest("GET", "/cn", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCacheNodesGet)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := struct {
		Version      int      `json:"version"`
		CacheNodeIps []string `json:"cacheNodeIps"`
	}{
		Version:      1,
		CacheNodeIps: []string{"127.0.0.1"},
	}
	var response struct {
		Version      int      `json:"version"`
		CacheNodeIps []string `json:"cacheNodeIps"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Version != expected.Version || len(response.CacheNodeIps) != len(expected.CacheNodeIps) || response.CacheNodeIps[0] != expected.CacheNodeIps[0] {
		t.Errorf("handler returned unexpected body: got %v want %v", response, expected)
	}
}



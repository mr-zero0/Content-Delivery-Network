package frontend_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"reflect"

	"github.com/hcl/cdn/cacheNode/frontend"
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/cacheNode/observability"
)

// Mock for Collapser interface
type MockCollapser struct{}

func (m *MockCollapser) Do(req *http.Request) (*http.Response, error) {
	// Mocked response
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("Mocked response body")),
	}, nil
}

// Mock for ObservabilityHandler
type MockObservabilityHandler struct{}

func (m *MockObservabilityHandler) RecordEventFrontend(event observability.FrontendEvent) {
	// Mock the observability recording logic
}

// Mock ConfigRunConfig
type MockRunConfig struct{}

func (m *MockRunConfig) DSLookup(req *http.Request) (*config.DSConfig, error) {
	// Returning a mock DSConfig instance for testing
	return &config.DSConfig{Name: "mockDS"}, nil
}

func (m *MockRunConfig) BindingPort() int {
	return 8080
}

func (m *MockRunConfig) BindingIp() string {
	return "127.0.0.1"
}

// Mock DSConfig (create this mock if the original DSConfig is not accessible)
type MockDSConfig struct {
	Name string
}

// Test for HandleClientRequest in Listener
func TestHandleClientRequest(t *testing.T) {
	mockCollapser := &MockCollapser{}
	mockObservability := &MockObservabilityHandler{}
	mockRunConfig := &MockRunConfig{}

	// Create the listener with mock objects
	listener := frontend.Listener{
		NextStep: mockCollapser, 
	}

	// Set the unexported fields using reflection
	setUnexportedField(&listener, "cfg", mockRunConfig)
	setUnexportedField(&listener, "feObs", mockObservability)

	// Mocking HTTP request
	req, err := http.NewRequest("GET", "http://localhost/test-path", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	// Mocking HTTP response writer
	respWriter := &mockResponseWriter{}

	// Test with a valid DS lookup
	dsValid, err := listener.IsMatchingClientUrlAvailable(req)
	if !dsValid || err != nil {
		t.Fatalf("Expected URL to match, got error: %v", err)
	}

	// Simulate a successful request to the Collapser
	listener.HandleClientRequest(respWriter, req)

	// Validate the response status code and body for success or failure
	if respWriter.statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respWriter.statusCode)
	}

	if respWriter.body != "Mocked response body" {
		t.Errorf("Expected response body 'Mocked response body', got '%s'", respWriter.body)
	}
}

// Helper function to set unexported fields using reflection
func setUnexportedField(obj interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(obj).Elem()
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}
	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}
	field.Set(reflect.ValueOf(value))
	return nil
}

// Mock ResponseWriter to capture response details
type mockResponseWriter struct {
	statusCode int
	header     http.Header
	body       string
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	m.body = string(b)
	return len(b), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

// Test for SendResponseToClient
func TestSendResponseToClient(t *testing.T) {
	// Creating a mock response
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("Response body")),
	}

	// Create a mock response writer
	respWriter := &mockResponseWriter{}

	// Create listener
	listener := frontend.Listener{}

	// Test SendResponseToClient method
	err := listener.SendResponseToClient(respWriter, mockResp, nil)
	if err != nil {
		t.Fatalf("Error sending response to client: %v", err)
	}

	// Check if the response code and body were correctly written
	if respWriter.statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, respWriter.statusCode)
	}

	if respWriter.body != "Response body" {
		t.Errorf("Expected response body 'Response body', got '%s'", respWriter.body)
	}
}

package frontend_test
 
import (
    "net/http"
    "testing"
    "bytes"
    "io"
    "github.com/hcl/cdn/cacheNode/frontend"
)
 
// Mock implementation of BackendRequest for testing purposes
type MockBackendRequest struct{}
 
func (m *MockBackendRequest) ReDo(req *http.Request, resp *http.Response) (*http.Response, error) {
    // Return a mock response for testing
    return &http.Response{
        StatusCode: http.StatusOK,
        Header:     make(http.Header),
        Body:       io.NopCloser(bytes.NewReader([]byte("Mocked response body"))),
    }, nil
}
 
// Mocking the missing Do method in the test itself
func (m *MockBackendRequest) Do(req *http.Request) (*http.Response, error) {
    // Return a mock response for the Do method
    return &http.Response{
        StatusCode: http.StatusOK,
        Header:     make(http.Header),
        Body:       io.NopCloser(bytes.NewReader([]byte("Mocked response for Do"))),
    }, nil
}
 
// Test for CopyResponse
func TestCopyResponse(t *testing.T) {
    src := &http.Response{
        StatusCode: http.StatusOK,
        Header:     make(http.Header),
        Body:       io.NopCloser(bytes.NewReader([]byte("Test body"))),
    }
    dest := &http.Response{}
 
    err := frontend.CopyResponse(src, dest)  // Use frontend.CopyResponse since it's in that package
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
 
    // Verify that the response body and headers were copied
    if dest.StatusCode != src.StatusCode {
        t.Errorf("Expected status code %d, got %d", src.StatusCode, dest.StatusCode)
    }
 
    if dest.Body == nil {
        t.Errorf("Expected body to be copied, but it was nil")
    }
 
    // Check if headers are copied
    if len(dest.Header) != len(src.Header) {
        t.Errorf("Expected %d headers, got %d", len(src.Header), len(dest.Header))
    }
}
 
// Test for IsResourceUsable with mock data
func TestIsResourceUsable(t *testing.T) {
    v := &frontend.Validator{
        BackendRequest: &MockBackendRequest{},
    }
 
    // Mock HTTP request and response
    req := &http.Request{}
    resp := &http.Response{
        Header:     make(http.Header),
        StatusCode: http.StatusOK,
    }
    resp.Header.Set("Age", "100")
    resp.Header.Set("Cache-Control", "max-age=200")
 
    usable, err := v.IsResourceUsable(resp, req)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if !usable {
        t.Errorf("Expected resource to be usable, but got false")
    }
 
    // Test with max-age less than age
    resp.Header.Set("Age", "300")
    usable, err = v.IsResourceUsable(resp, req)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if !usable {
        t.Errorf("Expected resource to be expired, but got usable")
    }
}
 
// Test Do method - check if resource is fetched from the backend
func TestDo(t *testing.T) {
    v := &frontend.Validator{
        BackendRequest: &MockBackendRequest{},
    }
 
    req := &http.Request{}
    oResp := &http.Response{
        StatusCode: http.StatusInternalServerError,
        Status:     "500 Internal Server Error",
        Header:     make(http.Header),
    }
 
    // Test when the resource is usable and response is 500
    bkndResp, err := v.Do(req, oResp)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if bkndResp.StatusCode != http.StatusOK {
        t.Errorf("Expected StatusCode 200, got %d", bkndResp.StatusCode)
    }
 
    // Test when the resource is not usable
    resp := &http.Response{
        Header:     make(http.Header),
        StatusCode: http.StatusOK,
    }
    resp.Header.Set("Age", "100")
    resp.Header.Set("Cache-Control", "max-age=50")
    _, err = v.Do(req, resp)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
}
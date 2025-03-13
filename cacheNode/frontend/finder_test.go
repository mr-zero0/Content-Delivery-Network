package frontend_test
 
import (
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "io"
)
 
// Mock implementations for dependencies
type mockRequestHandler struct {
    resp *http.Response
    err  error
}
 
func (m *mockRequestHandler) Do(*http.Request) (*http.Response, error) {
    return m.resp, m.err
}
 
type mockValidator struct {
    resp *http.Response
    err  error
}
 
func (m *mockValidator) Do(*http.Request, *http.Response) (*http.Response, error) {
    return m.resp, m.err
}
 
// Finder struct definition
type Finder struct {
    StoragePath *mockRequestHandler
    ValidatePath *mockValidator
    BackendPath *mockRequestHandler
}
 
func (f *Finder) Do(req *http.Request) (*http.Response, error) {
    // Check if StoragePath is available
    resp, err := f.StoragePath.Do(req)
    if err != nil {
        return &http.Response{
            StatusCode: http.StatusInternalServerError,
            Body:       io.NopCloser(strings.NewReader("storage error")),
        }, err
    }
 
    // If StoragePath is OK, validate
    if f.ValidatePath != nil {
        resp, err = f.ValidatePath.Do(req, resp)
        if err != nil {
            return &http.Response{
                StatusCode: http.StatusInternalServerError,
                Body:       io.NopCloser(strings.NewReader("validation failed")),
            }, err
        }
    }
 
    // If StoragePath is NotFound, try BackendPath
    if resp.StatusCode == http.StatusNotFound && f.BackendPath != nil {
        resp, err = f.BackendPath.Do(req)
        if err != nil {
            return &http.Response{
                StatusCode: http.StatusInternalServerError,
                Body:       io.NopCloser(strings.NewReader("backend error")),
            }, err
        }
    }
 
    return resp, nil
}
 
// Test cases for Finder
func TestFinder_Do(t *testing.T) {
    t.Run("Storage returns error", func(t *testing.T) {
        storage := &mockRequestHandler{err: errors.New("storage error")}
        finder := Finder{StoragePath: storage}
 
        req := httptest.NewRequest("GET", "http://example.com", nil)
        resp, _ := finder.Do(req)
 
        if resp.StatusCode != http.StatusInternalServerError {
            t.Errorf("expected status 500, got %d", resp.StatusCode)
        }
 
        body, _ := io.ReadAll(resp.Body)
        if !strings.Contains(string(body), "storage error") {
            t.Errorf("unexpected error message: %s", body)
        }
    })
 
    t.Run("Storage OK → Validation fails", func(t *testing.T) {
        storage := &mockRequestHandler{resp: &http.Response{StatusCode: http.StatusOK}}
        validator := &mockValidator{err: errors.New("validation failed")}
        finder := Finder{StoragePath: storage, ValidatePath: validator}
 
        req := httptest.NewRequest("GET", "http://example.com", nil)
        resp, _ := finder.Do(req)
 
        if resp.StatusCode != http.StatusInternalServerError {
            t.Errorf("expected status 500, got %d", resp.StatusCode)
        }
    })
 
    t.Run("Storage OK → Validation succeeds", func(t *testing.T) {
        expectedResp := &http.Response{StatusCode: http.StatusOK}
        storage := &mockRequestHandler{resp: &http.Response{StatusCode: http.StatusOK}}
        validator := &mockValidator{resp: expectedResp}
        finder := Finder{StoragePath: storage, ValidatePath: validator}
 
        req := httptest.NewRequest("GET", "http://example.com", nil)
        resp, _ := finder.Do(req)
 
        if resp != expectedResp {
            t.Error("expected validation response to be returned")
        }
    })
 
    t.Run("Storage NotFound → Backend succeeds", func(t *testing.T) {
        storage := &mockRequestHandler{resp: &http.Response{StatusCode: http.StatusNotFound}}
        backend := &mockRequestHandler{resp: &http.Response{StatusCode: http.StatusOK}}
        finder := Finder{StoragePath: storage, BackendPath: backend}
 
        req := httptest.NewRequest("GET", "http://example.com", nil)
        resp, _ := finder.Do(req)
 
        if resp.StatusCode != http.StatusOK {
            t.Errorf("expected backend response, got %d", resp.StatusCode)
        }
    })
 
    t.Run("Storage NotFound → Backend fails", func(t *testing.T) {
        storage := &mockRequestHandler{resp: &http.Response{StatusCode: http.StatusNotFound}}
        backend := &mockRequestHandler{err: errors.New("backend error")}
        finder := Finder{StoragePath: storage, BackendPath: backend}
 
        req := httptest.NewRequest("GET", "http://example.com", nil)
        resp, _ := finder.Do(req)
 
        if resp.StatusCode != http.StatusInternalServerError {
            t.Errorf("expected status 500, got %d", resp.StatusCode)
        }
    })
}
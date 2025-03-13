package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/hcl/cdn/cacheNode/common"
)

/*
 * Test Functions
 *		TestReader
 */

func getRequest(t *testing.T, storageHandler common.RequestHandler, caseId int, caseName string) {
	var URL string
	var expectedStatusCode int = http.StatusOK

	if caseId == 9 {
		URL = "http://abc.com/sample1"
	} else if caseId == 10 {
		URL = "http://abc.com/sample1"
		_ = os.Remove(CDNDatastore + "/abc.com/sample1/sample1.bin")
		expectedStatusCode = http.StatusNotFound
	} else if caseId == 11 {
		URL = "http://def.com/sample2"
		_ = os.Remove(CDNDatastore + "/def.com/sample2/sample2_metadata.json")
		expectedStatusCode = http.StatusNotFound
	} else if caseId == 12 {
		URL = "http://DS1/sample1"
		expectedStatusCode = http.StatusNotFound
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		t.Errorf("TestReader:Failed to create GET request: %v", err)
		return
	}
	response, err := storageHandler.Do(req)

	/*
	 * Perform response validation immediately before parsing response
	 */
	if expectedStatusCode != response.StatusCode {
		t.Errorf("TestReader:getRequest:%s:Expected status code: %d, but got %d", caseName, expectedStatusCode, response.StatusCode)
	} else {
		t.Logf("TestReader:getRequest:%s: PASS, StatusCode: %d", caseName, response.StatusCode)
	}

	if err == nil && response.StatusCode != http.StatusNotFound {
		body, err := io.ReadAll(response.Body)
		if err == nil {
			response.Body.Close()
			t.Logf("Status code: %d", response.StatusCode)
			t.Logf("Response Body: %s", string(body))
			t.Logf("Max-Age: %s | Age: %s | Last-Modified: %s", response.Header.Get("Cache-Control"), response.Header.Get("Age"), response.Header.Get("Last-Modified"))
		}
	}
}

func headRequest(t *testing.T, storageHandler common.RequestHandler, caseId int, caseName string) {
	var URL string
	var expectedStatusCode int = http.StatusOK

	if caseId == 7 {
		URL = "http://localhost:8080/sample1"
		expectedStatusCode = http.StatusOK
	} else if caseId == 8 {
		URL = "http://DS1/sample1"
		expectedStatusCode = http.StatusNotFound
	}

	req, err := http.NewRequest("HEAD", URL, nil)
	if err != nil {
		t.Logf("TestReader:Failed to create HEAD request: %v", err)
		return
	}
	if caseId == 7 {
		req.Host = "abc.com"
	}
	response, err := storageHandler.Do(req)
	/*
	 * Perform response validation immediately before parsing response
	 */
	if expectedStatusCode != response.StatusCode {
		t.Errorf("TestReader:headRequest:%s:Expected status code: %d, but got %d", caseName, expectedStatusCode, response.StatusCode)
	} else {
		t.Logf("TestReader:headRequest:%s: PASS, StatusCode: %d", caseName, response.StatusCode)
	}

	if err == nil && response.StatusCode != http.StatusNotFound {
		t.Logf("Status code: %d", response.StatusCode)
		t.Logf("Max-Age %s | Age: %s | Last-Modified: %s", response.Header.Get("Cache-Control"), response.Header.Get("Age"), response.Header.Get("Last-Modified"))
	}
}

func testReader(t *testing.T, storageHandler common.RequestHandler) {
	fmt.Println("Test Reader: HEAD Request")
	fmt.Println("========================")
	fmt.Println("Case 7: Metadata file is present")
	headRequest(t, storageHandler, 7, "Case 7")
	fmt.Println()
	fmt.Println("Case 8: Metadata file is not present")
	headRequest(t, storageHandler, 8, "Case 8")
	fmt.Println()
	fmt.Println("Test Reader: GET Request")
	fmt.Println("========================")
	fmt.Println("Case 9: Successfully retrieve metadata & content")
	getRequest(t, storageHandler, 9, "Case 9")
	fmt.Println()
	fmt.Println("Case 10: Metadata file is present, Content file is missing")
	getRequest(t, storageHandler, 10, "Case 10")
	fmt.Println()
	fmt.Println("Case 11: Metadata file is not present, Content file is present")
	getRequest(t, storageHandler, 11, "Case 11")
	fmt.Println()
	fmt.Println("Case 12: Both Metadata & Content files are missing")
	getRequest(t, storageHandler, 12, "Case 12")
	fmt.Println()
}

/*
 * Main test case function to test complete Reader functionalities
 */
func TestReader(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup

	/*
	 * Clear Root directory before running Writer test cases
	 */
	_ = os.RemoveAll(CDNDatastore)
	os.MkdirAll(CDNDatastore, os.ModePerm)

	cdnDir := GetDefaultCDNDirectory()
	CDNDatastore = cdnDir
	storageHandler, err := Init(ctx, &wg, cdnDir, nil, nil)
	if err != nil {
		t.Fatalf("TestReaderWriter:Failed to init storage")
	}

	testWriter(t, storageHandler)
	testReader(t, storageHandler)
	//wg.Done()
}

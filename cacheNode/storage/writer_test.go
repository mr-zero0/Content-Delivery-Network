package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hcl/cdn/cacheNode/common"
)

/*
 * Test Functions
 *		TestWriter
 */

func postRequest(t *testing.T, storageHandler common.RequestHandler, caseId int, caseName string) {
	URL := "http://abc.com/sample1"
	data := "Hi, Welcome to CDN project"
	var payload io.Reader
	var expectedStatusCode int = http.StatusOK

	if caseId == 1 {
		//No payload case
		payload = nil
		expectedStatusCode = http.StatusBadRequest
	} else if caseId == 2 {
		//Create 1st new content
		payload = strings.NewReader(data)
		expectedStatusCode = http.StatusCreated
	} else if caseId == 3 {
		//Create 2nd new content
		URL = "http://def.com/sample2"
		payload = strings.NewReader(data)
		expectedStatusCode = http.StatusCreated
	} else if caseId == 4 {
		URL = "http://ghi.com/sample3"
		payload = strings.NewReader(data)
		expectedStatusCode = http.StatusCreated
	} else if caseId == 5 {
		URL = "http://jkl.com/sample4"
		payload = strings.NewReader(data)
		expectedStatusCode = http.StatusCreated
	} else if caseId == 6 {
		URL = "http://mno.com/sample5"
		payload = strings.NewReader(data)
		expectedStatusCode = http.StatusCreated
	}

	request, err := http.NewRequest("POST", URL, payload)
	if err != nil {
		t.Logf("TestWriter:Failed to create POST request: %v", err)
		return
	}
	request.Header.Set("Content-Type", "application/octet")
	request.Header.Set("Content-Length", "10")

	if caseId != 4 {
		request.Header.Set("Cache-Control", "max-age=3600")
	}
	if caseId != 5 {
		request.Header.Set("Age", "60")
	}
	if caseId != 6 {
		nowTime := time.Now().UTC()
		request.Header.Set("Last-Modified", nowTime.Format(http.TimeFormat))
	}

	response, _ := storageHandler.Do(request)
	/*
	 * Perform response validation immediately before parsing response
	 */
	if expectedStatusCode != response.StatusCode {
		t.Errorf("TestReader:postRequest:%s:Expected status code: %d, but got %d", caseName, expectedStatusCode, response.StatusCode)
	} else {
		t.Logf("TestReader:postRequest:%s: PASS, StatusCode: %d", caseName, response.StatusCode)
	}
}

func putRequest(t *testing.T, storageHandler common.RequestHandler, caseId int, caseName string) {
	var expectedStatusCode int = http.StatusOK
	var request *http.Request = nil
	var err error

	if caseId == 13 {
		expectedStatusCode = http.StatusMethodNotAllowed
		request, err = http.NewRequest("PUT", "http://abc.com/sample1", nil)
	} else if caseId == 14 {
		expectedStatusCode = http.StatusBadRequest
	}

	if err != nil {
		t.Logf("TestUnsupportedMethod:Failed to create HEAD request: %v", err)
		return
	}
	response, _ := storageHandler.Do(request)

	/*
	 * Perform response validation immediately before parsing response
	 */
	if expectedStatusCode != response.StatusCode {
		t.Errorf("TestUnsupportedMethod:putRequest:%s:Expected status code: %d, but got %d", caseName, expectedStatusCode, response.StatusCode)
	} else {
		t.Logf("TestUnsupportedMethod:putRequest:%s: PASS, StatusCode: %d", caseName, response.StatusCode)
	}
}

func testWriter(t *testing.T, storageHandler common.RequestHandler) {
	fmt.Println("Test Writer: POST Request")
	fmt.Println("========================")
	fmt.Println("Case 1: No Payload Data")
	postRequest(t, storageHandler, 1, "Case 1")
	fmt.Println()
	fmt.Println("Case 2: Successfully store content & metadata")
	postRequest(t, storageHandler, 2, "Case 2")
	postRequest(t, storageHandler, 3, "Case 3")
	fmt.Println("Case 4: Store content & metadata without sending max-age")
	postRequest(t, storageHandler, 4, "Case 4")
	fmt.Println("Case 5: Store content & metadata without sending Age")
	postRequest(t, storageHandler, 5, "Case 5")
	fmt.Println("Case 6: Store content & metadata without sending Last-Modified")
	postRequest(t, storageHandler, 6, "Case 6")
	fmt.Println()
}

func testUnsupportedMethod(t *testing.T, storageHandler common.RequestHandler) {
	fmt.Println("Test Unsupported Method: PUT Request")
	fmt.Println("==================================")
	fmt.Println("Case 13: Unsuported HTTP method")
	putRequest(t, storageHandler, 13, "Case 13")
	fmt.Println()
}

func testNilHttpRequest(t *testing.T, storageHandler common.RequestHandler) {
	fmt.Println("Test Nil HTTP Request:")
	fmt.Println("==================================")
	fmt.Println("Case 14: Pass http.request as nil to storage.Do()")
	putRequest(t, storageHandler, 14, "Case 14")
	fmt.Println()
}

/*
 * Main test case function to test complete Writer functionalities
 */
func TestWriter(t *testing.T) {
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
	testUnsupportedMethod(t, storageHandler)
	testNilHttpRequest(t, storageHandler)
	//wg.Done()
}

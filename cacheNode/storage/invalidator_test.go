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
	"log/slog"

	"github.com/hcl/cdn/cacheNode/common"
)

/*
 * Test Functions
 *		TestCacheEvictorInvalidator
 */

func createDummyFilesForInvalidator(t *testing.T, storageHandler common.RequestHandler, id int) {
	data := "Hi, Welcome to CDN project"
	var payload io.Reader

	var i int = 0
	for i = 0; i < 2; i++ {
		URL := fmt.Sprintf("http://abc.com/DS_Set%d/sample%d", id, i)
		//Create 1st new content
		payload = strings.NewReader(data)
		request, err := http.NewRequest("POST", URL, payload)
		if err != nil {
			fmt.Println("Failed to create Set3 POST request:", err)
			return
		}
		request.Header.Set("Content-Type", "application/octet")
		request.Header.Set("Content-Length", "10")

		request.Header.Set("Cache-Control", "max-age=36000")
		request.Header.Set("Age", "1000")
		nowTime := time.Now().UTC()
		request.Header.Set("Last-Modified", nowTime.Format(http.TimeFormat))

		response, err := storageHandler.Do(request)
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("Status code: ", response.StatusCode)
			return
		}
		t.Logf("Status code: %d", response.StatusCode)
	}
	t.Logf("Created %d metadata & content files\n", i)
}

/*
 * Validate the presence of following directory in CDNDATASTORE after CacheEvictor invalidator has deleted stale contents.
 * 		http://abc.com/DS_Set1
 *		http://abc.com/DS_Set2
 *		http://abc.com/DS_Set3
 */
func validateInvalidator(t *testing.T, deletedDirectories *[]string) {
	for _, deletedDirectory := range *deletedDirectories {
		_, err := os.Stat(CDNDatastore + "/" + deletedDirectory)
		if err == nil || os.IsExist(err) {
			t.Fatalf("TestCacheEvictorInvalidator:CacheEvictor failed to delete %s", deletedDirectory)
		}
	}
}

func invalidateRequest(t *testing.T, storageHandler common.RequestHandler, caseId int, caseName string) {
	var URL string
	var expectedStatusCode int = http.StatusOK
	var deletedDirectories []string

	if caseId == 1 {
		/*
		 * Complete string based deletion. It will delete the directory 'DS_Set1' under 'abc.com'
		 */
		URL = "http://abc.com/DS_Set1"
		expectedStatusCode = http.StatusOK
		deletedDirectories = append(deletedDirectories, "abc.com/DS_Set1")
	} else if caseId == 2 {
		/*
		 * Pattern based deletion. It will delete all directories under 'abc.com'
		 * if the directory name contains the pattern 'sample'. here pattern wil be 'DS_Set'
		 */
		URL = "http://abc.com/DS_Set"
		expectedStatusCode = http.StatusOK
		deletedDirectories = append(deletedDirectories, "abc.com/DS_Set1", "abc.com/DS_Set2", "abc.com/DS_Set3")
	}

	req, err := http.NewRequest("DELETE", URL, nil)
	if err != nil {
		t.Logf("TestCacheEvictorInvalidator:Failed to create invalidate request: %v", err)
		return
	}

	response, _ := storageHandler.Do(req)

	/*
	 * Perform response validation immediately before parsing response
	 */
	if expectedStatusCode != response.StatusCode {
		t.Fatalf("TestCacheEvictorInvalidator:%s:Expected status code: %d, but got %d", caseName, expectedStatusCode, response.StatusCode)
	}

	validateInvalidator(t, &deletedDirectories)
}

func TestCacheEvictorInvalidator(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup
	cdnDir := GetDefaultCDNDirectory()


	/*
	 * Clear CDNDATASTORE root directory before running CacheEvictor test cases
	 */
	_ = os.RemoveAll(cdnDir)
	os.MkdirAll(cdnDir, os.ModePerm)
	setCacheEvictorInterval(2)

	storageHandler, err := Init(ctx, &wg, cdnDir, nil, nil)
	if err != nil {
		t.Fatalf("TestCacheEvictorInvalidator:Failed to init storage")
	}
	slog.SetLogLoggerLevel(slog.LevelDebug)
	//Run refresh stale content map for every 5 seconds
	setRefreshStaleContentMapInterval(5)
	fmt.Println("TestCacheEvictorInvalidator")
	fmt.Println("========================")
	fmt.Println("Case 1: Simple directory path deletion")
	createDummyFilesForInvalidator(t, storageHandler, 1)
	time.Sleep(10 * time.Second)
	invalidateRequest(t, storageHandler, 1, "Case 1")
	//Assertion
	storageObj, _ := storageHandler.(*StorageHandler)
	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	time.Sleep(10 * time.Second)
	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	//Validate whether CacheEvictor cleaned all entries from stale content map when the files are deleted from disk.
	if storageObj.cacheManagerObj.getTotalContents() > 0 {
		t.Errorf("TestCacheEvictorInvalidator:CacheEvictor failed to clean stale content map %d", storageObj.cacheManagerObj.getTotalContents())
	}
	fmt.Println()

	slog.SetLogLoggerLevel(slog.LevelInfo)
	fmt.Println("Case 2: Pattern based deletion")
	createDummyFilesForInvalidator(t, storageHandler, 1)
	createDummyFilesForInvalidator(t, storageHandler, 2)
	createDummyFilesForInvalidator(t, storageHandler, 3)
	time.Sleep(10 * time.Second)
	invalidateRequest(t, storageHandler, 2, "Case 2")
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()

	fmt.Println("Wait for 15 seconds...")
	//Give some time for CacheEvictor to sync stale content map data from disk
	time.Sleep(20 * time.Second)
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()

	//Validate whether CacheEvictor cleaned all entries from stale content map when the files are deleted from disk.
	if storageObj.cacheManagerObj.getTotalContents() > 0 {
		t.Errorf("TestCacheEvictorInvalidator:CacheEvictor failed to clean stale content map %d", storageObj.cacheManagerObj.getTotalContents())
	}
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	fmt.Println()
}

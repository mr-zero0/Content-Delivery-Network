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
 *		TestCacheEvictorMonitoringService
 */

func createDummyFilesForCacheEvictor(t *testing.T, storageObj common.RequestHandler, id int) {
	data := "Hi, Welcome to CDN project"
	var payload io.Reader
	var URL string
	var request *http.Request = nil
	var err error = nil

	var i int = 0
	for i = 0; i < 10; i++ {
		switch i {
		case 0, 1, 2:
			URL = fmt.Sprintf("http://abc.com/DS_Set%d_%d/sample1", id, i)
			payload = strings.NewReader(data)
			request, err = http.NewRequest("POST", URL, payload)
			if err != nil {
				fmt.Println("Failed to create set1 POST request:", err)
				return
			}
			/*
			 * Create 3 stale contents by setting max age as 0
			 */
			request.Header.Set("Cache-Control", "max-age=0")
			request.Header.Set("Age", "60")
			nowTime := time.Now().UTC()
			request.Header.Set("Last-Modified", nowTime.Format(http.TimeFormat))
		case 3, 4, 5:
			URL = fmt.Sprintf("http://def.com/DS_Set%d_%d/sample1", id, i)
			payload = strings.NewReader(data)
			request, err = http.NewRequest("POST", URL, payload)
			if err != nil {
				fmt.Println("Failed to create set1 POST request:", err)
				return
			}
			/*
			 * Create 3 contents by setting max age as 2hrs duration
			 * These 3 contents should be next candidates for eviction if disk usage crosses 80%
			 */
			request.Header.Set("Cache-Control", "max-age=7200")
			request.Header.Set("Age", "1000")
		case 6, 7, 8, 9:
			URL = fmt.Sprintf("http://ghi.com/DS_Set%d_%d/sample1", id, i)
			payload = strings.NewReader(data)
			request, err = http.NewRequest("POST", URL, payload)
			if err != nil {
				fmt.Println("Failed to create set1 POST request:", err)
				return
			}
			/*
			 * Create 4 contents by setting max age as 10hrs duration
			 */
			request.Header.Set("Cache-Control", "max-age=36000")
			request.Header.Set("Age", "1000")
		}
		request.Header.Set("Content-Type", "application/octet")
		request.Header.Set("Content-Length", "10")

		response, err := storageObj.Do(request)
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
 * Validate the directories in CDNDATASTORE after CacheEvictor monitoring service has deleted the stale contents.
 */
func validateCacheEvictorSet(t *testing.T, wg *sync.WaitGroup, mandatoryDeletedDirectories []string, possibleDeletedDirectories []string) {
	for _, deletedDirectory := range mandatoryDeletedDirectories {
		_, err := os.Stat(CDNDatastore + deletedDirectory)
		if err == nil || os.IsExist(err) {
			t.Fatalf("TestCacheEvictorMonitoringService:CacheEvictor failed to delete %s", deletedDirectory)
		}
	}

	var isDeleted bool = false
	for _, deletedDirectory := range possibleDeletedDirectories {
		_, err := os.Stat(CDNDatastore + deletedDirectory)
		if err != nil && os.IsNotExist(err) {
			isDeleted = true
			break
		}
	}
	if !isDeleted {
		t.Fatalf("TestCacheEvictorMonitoringService:CacheEvictor failed to delete %s", possibleDeletedDirectories)
	}

	/*
	 * Exit cacheEvictor Go-routine after successful validation
	 */
	if wg != nil {
		wg.Done()
	}
}

func TestCacheEvictorMonitoringService(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup
	cacheManagerHandler := &cacheManager{
		staleContentMap:             make(map[int][]metadataStruct),
		totalContents: 0,
	}

	storageObj := &StorageHandler{
		observabilityObj: nil,
		cacheManagerObj: cacheManagerHandler,
	}

	var mandatoryDeletedDirectories = make([]string, 0)
	var possibleDeletedDirectories = make([]string, 0)
	cdnDir := GetDefaultCDNDirectory()
	CDNDatastore = cdnDir
	slog.SetLogLoggerLevel(slog.LevelInfo)

	t.Log("TestCacheEvictorMonitoringService")
	t.Log("=================================")
	slog.SetLogLoggerLevel(slog.LevelInfo)
	/*
	 * Clear CDNDATASTORE Root directory before running CacheEvictor test cases
	 */
	_ = os.RemoveAll(CDNDatastore)
	os.MkdirAll(CDNDatastore, os.ModePerm)

	setCacheEvictorInterval(2)
	//Add some contents to validate loadCachedContentInMap()
	createDummyFilesForCacheEvictor(t, storageObj, 1)

	/*
	 * Instead of calling Init(), directly call loadCachedContentInMap() and cacheEvictor go-routine
	 * to validate the functionality of loading content metadata from disk and updating in-memory map
	 */
	err := loadCachedContentInMap(storageObj)
	if err != nil {
		t.Fatalf("TestCacheEvictorMonitoringService:Failed to load cache content in map, error=%s", err)
	}
	wg.Add(1)
	go cacheEvictor(ctx, &wg, storageObj)

	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	setFileDeletionLimit(4)
	//Run refresh stale content map for every 5 seconds
	setRefreshStaleContentMapInterval(5)
	//Simulate disk usage > disk threshold
	simulateHighDiskUsage()
	/*
	 * Give some time for CacheEvictor to delete the stale directories and reduce disk usage.
	 */
	t.Logf("TestCacheEvictorMonitoringService:CacheEvictor: Wait for 10 seconds...")
	time.Sleep(10 * time.Second)
	mandatoryDeletedDirectories = append(mandatoryDeletedDirectories, "/abc.com/DS_Set1_0", "/abc.com/DS_Set1_1", "/abc.com/DS_Set1_2")
	possibleDeletedDirectories = append(possibleDeletedDirectories, "/def.com/DS_Set1_3", "/def.com/DS_Set1_4", "/def.com/DS_Set1_5")

	validateCacheEvictorSet(t, nil, mandatoryDeletedDirectories, possibleDeletedDirectories)

	/*
	 * Validate CacheEvictor for any runtime content updation & disk usage > 80%
	 */
	createDummyFilesForCacheEvictor(t, storageObj, 2)
	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	//Simulate disk usage > disk threshold
	simulateHighDiskUsage()
	t.Logf("TestCacheEvictorMonitoringService:CacheEvictor: Wait for 15 seconds...")
	time.Sleep(15 * time.Second)
	mandatoryDeletedDirectories = append(mandatoryDeletedDirectories, "/abc.com/DS_Set2_0", "/abc.com/DS_Set2_1", "/abc.com/DS_Set2_2")
	possibleDeletedDirectories = append(possibleDeletedDirectories, "/def.com/DS_Set1_4", "/def.com/DS_Set1_5")
	validateCacheEvictorSet(t, nil, mandatoryDeletedDirectories, possibleDeletedDirectories)

	//Validate Set3
	createDummyFilesForCacheEvictor(t, storageObj, 3)
	time.Sleep(5 * time.Second)
	createDummyFilesForCacheEvictor(t, storageObj, 4)
	time.Sleep(5 * time.Second)
	createDummyFilesForCacheEvictor(t, storageObj, 5)
	time.Sleep(5 * time.Second)
	storageObj.cacheManagerObj.logStaleContentMap()
	//Simulate disk usage > disk threshold
	simulateHighDiskUsage()
	t.Logf("TestCacheEvictorMonitoringService:CacheEvictor: Wait for 15 seconds...")
	time.Sleep(15 * time.Second)
	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	mandatoryDeletedDirectories = append(mandatoryDeletedDirectories, "/abc.com/DS_Set3_0", "/abc.com/DS_Set3_1", "/abc.com/DS_Set3_2")
	possibleDeletedDirectories = append(possibleDeletedDirectories, "/def.com/DS_Set4_0", "/def.com/DS_Set4_1", "/def.com/DS_Set4_2")
	validateCacheEvictorSet(t, nil, mandatoryDeletedDirectories, possibleDeletedDirectories)

	//Validiate Set4
	//Simulate disk usage > disk threshold
	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	if true {
		//Unit test code to validate deletion of all files by CacheEvictor by simulating high disk usage after every 5 seconds
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(10 * time.Second)
		simulateHighDiskUsage()
		time.Sleep(15 * time.Second)
	} else {
		simulateHighDiskUsage()
		t.Logf("TestCacheEvictorMonitoringService:CacheEvictor: Wait for 15 seconds...")	
		time.Sleep(15 * time.Second)
	}
	mandatoryDeletedDirectories = append(mandatoryDeletedDirectories, "/abc.com/DS_Set4_1", "/abc.com/DS_Set4_2")
	possibleDeletedDirectories = append(possibleDeletedDirectories, "/abc.com/DS_Set5_0", "/def.com/DS_Set5_1", "/def.com/DS_Set5_2")
	validateCacheEvictorSet(t, &wg, mandatoryDeletedDirectories, possibleDeletedDirectories)
	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	t.Logf("TestCacheEvictorMonitoringService:CacheEvictor: PASS")

	wg.Wait()
}

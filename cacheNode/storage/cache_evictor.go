package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/*
 * Functions Defined
 * 		cacheEvictor() - Goroutine function
 *		loadCachedContentInMap()
 *		deleteStaleContent()
 *		deleteEmptyParentDirs()
 *		refreshStaleContentMapAtRuntime()
 */

func setCacheEvictorInterval(i int) {
	cacheEvictorInterval = i
}

func setRefreshStaleContentMapInterval(i int) {
	refreshStaleContentMapInterval = i
}

func setFileDeletionLimit(i int) {
	fileDeletionLimit = i
}

/*
 * Function to delete empty directories once all its child content directories are deleted from disk
 */
func deleteEmptyParentDirs(dir string) {
	entries, err := os.ReadDir(dir)
	if len(entries) == 0 && err == nil {
		err = os.Remove(dir)
		if err != nil {
			slog.Error("Storage:CacheEvictor:Failed to delete empty parent directory", "dir", dir, "error", err)
		} else {
			parentDir := filepath.Dir(dir)
			if parentDir != CDNDatastore {
				deleteEmptyParentDirs(parentDir)
			}
		}
	}
}

/*
 * Function to load all cached content from storage into in-memory map
 */
func loadCachedContentInMap(storageObj *StorageHandler) (err error) {
	count := 0
	err = filepath.Walk(CDNDatastore, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			slog.Error("Storage:loadCachedContentInMap:Failed to walk CDNDATASTORE", "error", err)
			return err
		}
		if !info.IsDir() && strings.Contains(path, "_metadata.json") {
			fileContent, err := os.ReadFile(path)
			if err != nil {
				slog.Error("Storage:loadCachedContentInMap:Failed to read metdata", "file", path, "error", err)
				return err
			}
			metadata := make(map[string][]string)
			err = json.Unmarshal(fileContent, &metadata)
			if err != nil {
				slog.Error("Storage:loadCachedContentInMap:Failed to load metadata", "file", path, "error", err)
				return err
			}

			_, ok1 := metadata["Last-Modified"]
			_, ok2 := metadata["Age"]
			_, ok3 := metadata["Cache-Control"]
			if !ok1 || !ok2 || !ok3 {
				slog.Error("Storage:CacheEvictor:CDNDATASTORE directory got corrupted. CacheEvictor will behave abnormally...")
				return errors.New("CDNDATASTORE directory got corrupted")
			}

			maxAge, age, lastModified, contentLength := validateCacheHeaders(metadata)
			storageObj.cacheManagerObj.updateCacheContentInMap(maxAge, age, lastModified, contentLength, filepath.Dir(path))
			count += 1
		}
		return nil
	})
	if err != nil {
		slog.Error("Storage:loadCachedContentInMap:Failed to open CDNDATASTORE", "error", err)
		return err
	}	
	slog.Info("Storage:loadCachedContentInMap:Successfully loaded content metadata", "count", count)
	return nil
}

/*
 * Function to get list of contents with least remaining cache duration and delete from disk
 */
func deleteStaleContent(limit int, storageObj *StorageHandler) (delCount int) {
	delCount = 0
	var bytesDeleted int = 0
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		timeTaken := endTime.Sub(startTime)
		recordStorageMetrics(nil, "", "delete", int(timeTaken.Milliseconds()), bytesDeleted, storageObj.observabilityObj)
		slog.Info("Storage:CacheEvictor:Time taken to delete ", "timeTaken(milliseconds)", timeTaken.Milliseconds(), "deletedCount", delCount, "bytesDeleted", bytesDeleted)
	}()

	//TBD - Below func will take some time based on no. of entries. Has to be optimized in future
	remainingCacheDurationSlice, remainingCacheDurationSliceLen := storageObj.cacheManagerObj.getAllKeysFromMap()
	for remainingCacheDurationSliceLen > 0 && delCount < limit {
		/*
		 * Always consider first entry from RemainingCacheDurationSlice for any stale deletion.
		 * Once deleted all contents mapped to first entry, then delete the first entry from both slice & map
		 * In this way, next stale content will be moved to first position of slice & map
		 */
		duration := remainingCacheDurationSlice[0] 
		for delCount < limit {
			metadata := storageObj.cacheManagerObj.popFirstEntryFromMap(duration)
			if metadata == nil {
				break
			}
			/*
			 * Check for the directory before deletion. If the invalidator has already deleted the directory,
			 * skip the deletion of that directory and remove it's entry from in-memory map & slice
			 */
			_, statErr := os.Stat(metadata.path)
			if statErr == nil {
				os.RemoveAll(metadata.path)
				//Remove the directory if there are no subdirectories or files inside it.
				deleteEmptyParentDirs(filepath.Dir(metadata.path))
				bytesDeleted += metadata.contentLength
				delCount += 1
				slog.Info("Storage:CacheEvictor:deleteStaleContent:Deleted directory", "dir", metadata.path)
			} else {
				slog.Debug("Storage:CacheEvictor:deleteStaleContent:Directory already deleted by Invalidator", "dir", metadata.path)
			}
		}
		// Remove deleted content entry from RemainingCacheDurationSlice if there are no entries in map
		remainingCacheDurationSlice = remainingCacheDurationSlice[1:]
		remainingCacheDurationSliceLen = len(remainingCacheDurationSlice)
	}
	return
}

/*
 * Function to scan entire StaleContentMap and update remaining cache duration for each content in in-memory map.
 */
 func refreshStaleContentMapAtRuntime(storageObj *StorageHandler) {
	numOfContents := storageObj.cacheManagerObj.getTotalContents()
	/*
	 * The following logic ensures that RemainingCacheDurationSlice is aligned with StaleContentMap.
	 */
	slog.Debug("Storage:", "TotalContents", numOfContents)
	//TBD - Below func will take some time based on no. of entries. Has to be optimized in future
	remainingCacheDurationSlice, remainingCacheDurationSliceLen := storageObj.cacheManagerObj.getAllKeysFromMap()
	for i := 0; i < remainingCacheDurationSliceLen; i++ {
		oldRemainingCacheDuration := remainingCacheDurationSlice[i]
		for {
			metadata := storageObj.cacheManagerObj.popFirstEntryFromMap(oldRemainingCacheDuration)
			if metadata == nil {
				break
			}
			// Update in-memory stale content map only if content file is present. Otherwise delete the entry from stale content map.
			_, err := os.Stat(metadata.path)
			if err == nil {
				lastModifiedTimestamp, _ := http.ParseTime(metadata.lastModified)
				age := metadata.age + int(time.Now().Unix()-lastModifiedTimestamp.Unix())
				storageObj.cacheManagerObj.updateCacheContentInMap(metadata.maxAge, age, metadata.lastModified, metadata.contentLength, metadata.path)
			}
		}
	}
	storageObj.cacheManagerObj.logStaleContentMap()
	storageObj.cacheManagerObj.logRemainingCacheDurationSlice()
	slog.Debug("Storage:CacheEvictor:rearrangeStaleContentMap:Completes one full scan of remaining cache duration")
}

/*
 * Main go-routine to monitor disk usage of /CDNDatastore root directory. If the disk usage exceeds
 * configured disk threshold value (80%), then CacheEvictor monitor routine will start deleting stale
 * contents from storage untill the disk usage is brought back to below 80%
 */
func cacheEvictor(ctx context.Context, wg *sync.WaitGroup, storageObj *StorageHandler) {
	defer wg.Done()
	then := time.Now()
	slog.Info("Storage:CacheEvictor:Starting...")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Storage:CacheEvictor:Exiting...")
			return
		case <-time.After(time.Duration(cacheEvictorInterval) * time.Second):
		}

		diskUsage := getDiskUsage(CDNDatastore)
		if diskUsage > DiskThreshold {
			slog.Info("Storage:CacheEvictor", "DiskUsage", diskUsage, "DiskThreshold", DiskThreshold)
			deleteStaleContent(fileDeletionLimit, storageObj)
		} else {
			/*
			 * In the ideal state, calculate the remaining cache duration for all entries in StaleContentMap
			 * and rearrange the map accordingly. As of now, perform below functionality only at interval of every 2 min.
			 * TBD ===> Scan the slice in chunks to improve concurrency
			 */
			now := time.Now()
			if now.Sub(then) > (time.Duration(refreshStaleContentMapInterval) * time.Second) {
				slog.Debug("Storage:CacheEvictor:Starting stale content map refresh at runtime...")
				refreshStaleContentMapAtRuntime(storageObj)
				then = now
			}
		}
		numOfContents := storageObj.cacheManagerObj.getTotalContents()
		recordStorageDiskUsageMetrics(diskUsage, numOfContents, storageObj.observabilityObj)	
	}
}

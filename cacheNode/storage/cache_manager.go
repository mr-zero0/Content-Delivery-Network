package storage

import (
	"log/slog"
	"sort"
	"sync"
)

/*
 * Functions Defined
 * 		pushEntryInMap()
 *		popFirstEntryFromMap()
 *		getAllKeysFromMap()
 *		incrementContentCount()
 *		decrementContentCount()
 *		getTotalContents()
 *		logStaleContentMap()
 *		logRemainingCacheDurationSlice()
 *		updateCacheContentInMap()
 */

type metadataStruct struct {
	path          string
	maxAge        int
	age           int
	lastModified  string
	contentLength int
}

type cacheManager struct {
	//Read/Write mutex lock for maintaining concurrent access to in-memory map & slice
	mutex                       sync.RWMutex
	staleContentMap             map[int][]metadataStruct
	totalContents               int
}
 
func(cacheManagerObj *cacheManager) pushEntryInMap(remainingCacheDuration int, newEntry metadataStruct) {
	cacheManagerObj.mutex.Lock()
	defer cacheManagerObj.mutex.Unlock()
	for i, value := range cacheManagerObj.staleContentMap[remainingCacheDuration] {
		if value.path == newEntry.path {
			//Delete old entry from StaleContentMap inner slice
			cacheManagerObj.staleContentMap[remainingCacheDuration] = append(cacheManagerObj.staleContentMap[remainingCacheDuration][:i], cacheManagerObj.staleContentMap[remainingCacheDuration][i+1:]...)
			cacheManagerObj.decrementContentCount()
		}
	}
	//Insert new entry
	cacheManagerObj.staleContentMap[remainingCacheDuration] = append(cacheManagerObj.staleContentMap[remainingCacheDuration], newEntry)
	cacheManagerObj.incrementContentCount()
}

func(cacheManagerObj *cacheManager) popFirstEntryFromMap(duration int) (value *metadataStruct) {
	cacheManagerObj.mutex.Lock()
	defer cacheManagerObj.mutex.Unlock()
	if len(cacheManagerObj.staleContentMap[duration]) <= 0 {
		value = nil
		return
	} 
	value = &cacheManagerObj.staleContentMap[duration][0]
	cacheManagerObj.staleContentMap[duration] = cacheManagerObj.staleContentMap[duration][1:]
	//Delete the map entry if inner slice is empty
	if len(cacheManagerObj.staleContentMap[duration]) <= 0 {
		delete(cacheManagerObj.staleContentMap, duration)
	}
	cacheManagerObj.decrementContentCount()
	return
}

func(cacheManagerObj *cacheManager) getAllKeysFromMap() (remainingCacheDurationSlice []int, remainingCacheDurationSliceLen int){
	cacheManagerObj.mutex.RLock()
	defer cacheManagerObj.mutex.RUnlock()
	remainingCacheDurationSlice = []int{}
	for key := range cacheManagerObj.staleContentMap {
		remainingCacheDurationSlice = append(remainingCacheDurationSlice, key)
	}
	// move stale contents to top
	sort.Ints(remainingCacheDurationSlice)
	remainingCacheDurationSliceLen = len(remainingCacheDurationSlice)	
	return
}

func(cacheManagerObj *cacheManager) incrementContentCount() {
	cacheManagerObj.totalContents += 1
}

func(cacheManagerObj *cacheManager) decrementContentCount() {
	cacheManagerObj.totalContents -= 1
}

func(cacheManagerObj *cacheManager) getTotalContents() (numOfContents int) {
	cacheManagerObj.mutex.RLock()
	defer cacheManagerObj.mutex.RUnlock()
	numOfContents = cacheManagerObj.totalContents
	return
}

func(cacheManagerObj *cacheManager) logStaleContentMap() {
	cacheManagerObj.mutex.RLock()
	defer cacheManagerObj.mutex.RUnlock()
	slog.Debug("storage:DEBUG", "StaleContentMap", cacheManagerObj.staleContentMap)
}

func(cacheManagerObj *cacheManager) logRemainingCacheDurationSlice() {
	remainingCacheDurationSlice, remainingCacheDurationSliceLen := cacheManagerObj.getAllKeysFromMap()
	slog.Debug("storage:DEBUG", "RemainingCacheDurationSliceLength", remainingCacheDurationSliceLen, "RemainingCacheDurationSlice", remainingCacheDurationSlice)
}

func (cacheManagerObj *cacheManager) updateCacheContentInMap(maxAge int, age int, lastModified string, contentLength int, path string) {
	cachedDuration := age
	remainingCacheDuration := maxAge - cachedDuration
	newEntry := metadataStruct{path, maxAge, age, lastModified, contentLength}

	cacheManagerObj.pushEntryInMap(remainingCacheDuration, newEntry)
	cacheManagerObj.logStaleContentMap()
}


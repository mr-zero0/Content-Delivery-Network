package storage

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hcl/cdn/cacheNode/common"
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/cacheNode/observability"
)

/*
 * TBD - Move below configs to Config module
 */
const (
	DiskThreshold = 80
)

var (
	fileDeletionLimit              	int = 4 //No. of files to be deleted by CacheEvictor
	CDNDatastore                   	string
	cacheEvictorInterval 			int = 30 // Unit in seconds
	refreshStaleContentMapInterval 	int = 30 // Unit in seconds
	storageLogLevel                 = slog.LevelInfo
)

// Get metadata file name and bin file name by merging CDN base directory and parsed url.
func fileNames(parsedUrl *url.URL, host string) (contentDir string, contentMetaDataFile string, contentFile string) {
	if host == "" {
		host = parsedUrl.Hostname()
	}
	contentDir = filepath.Join(CDNDatastore, host, parsedUrl.Path)
	contentMetaDataFile = filepath.Join(contentDir, filepath.Base(parsedUrl.Path)+"_metadata.json")
	contentFile = filepath.Join(contentDir, filepath.Base(parsedUrl.Path)+".bin")
	return
}

func recordStorageMetrics(parsedUrl *url.URL, host string, operation string, timeTaken int, bytes int, observabilityObj observability.ObservabilityHandler) {
	var requestUrl string
	if parsedUrl != nil {
		requestUrl = "http://" + host + "/" + parsedUrl.Path
	} else {
		requestUrl = "CacheEvictor"
	}
	storageEvent := observability.StorageEvent{
		Timestamp:      time.Now(),
		URL:            requestUrl,
		Operation:      operation,
		ResponseTime:   int(timeTaken),
		Bytes:          bytes,
	}
	if observabilityObj != nil {
		observabilityObj.RecordEventStorage(storageEvent)
	}
}


func recordStorageDiskUsageMetrics(diskUsage int, totalContents int, observabilityObj observability.ObservabilityHandler) {
	storageDiskMetricsEvent := observability.StorageDiskMetricsEvent{
		Timestamp:      time.Now(),
		DiskUsage:   diskUsage,
		TotalContents:   totalContents,
	}
	if observabilityObj != nil {
		observabilityObj.RecordEventStorageDiskMetrics(storageDiskMetricsEvent)
	}
}
func Init(ctx context.Context, wg *sync.WaitGroup, cdnDir string, cfg *config.RunConfig, observabilityHandler observability.ObservabilityHandler) (common.RequestHandler, error) {
	cacheManagerHandler := &cacheManager{
		staleContentMap:             make(map[int][]metadataStruct),
		totalContents: 0,
	}

	storageObj := &StorageHandler{
		observabilityObj: observabilityHandler,
		cacheManagerObj: cacheManagerHandler,
	}
	
	CDNDatastore = cdnDir
	slog.Info("Storage:Initializing storage module...", "cdnDir", cdnDir)

	slog.SetLogLoggerLevel(storageLogLevel)
	err := os.MkdirAll(CDNDatastore, 0755)
	if err != nil {
		slog.Error("Storage:Init:Failed to create CDNDATASTORE base directory", "error", err)
		return nil, err
	}

	err = loadCachedContentInMap(storageObj)
	if err != nil {
		slog.Error("Storage:Init:Failed to load cache content in map", "error", err)
		return nil, err
	}
	wg.Add(1)
	go cacheEvictor(ctx, wg, storageObj)

	//Observability Unit Testing
	if false {
		storageEvent := observability.StorageEvent{
			Timestamp:      time.Now(),
			URL:            "http://abc.com/Silver",
			Operation:      "write",
			ResponseTime:   12,
			Bytes:          200,
		}
		observabilityHandler.RecordEventStorage(storageEvent)

		storageEvent = observability.StorageEvent{
			Timestamp:      time.Now(),
			URL:            "http://abc.com/Silver",
			Operation:      "read",
			ResponseTime:   15,
			Bytes:          200,
		}
		observabilityHandler.RecordEventStorage(storageEvent)
			
		storageEvent = observability.StorageEvent{
			Timestamp:      time.Now(),
			URL:            "http://abc.com/Silver",
			Operation:      "delete",
			ResponseTime:   10,
			Bytes:          200,
		}
		observabilityHandler.RecordEventStorage(storageEvent)

		storageDiskMetricsEvent := observability.StorageDiskMetricsEvent{
			Timestamp:      time.Now(),
			DiskUsage:   	50,
			TotalContents:  200,
		}
		observabilityHandler.RecordEventStorageDiskMetrics(storageDiskMetricsEvent)
	}
	return storageObj, nil
}

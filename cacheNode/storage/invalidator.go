package storage

import (
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
 * TBD - Optimize the code to get bytes deleted
 */
func getDirectorySize(contentDir string, cacheManagerObj *cacheManager) (totalSize int) {
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} 
		if !info.IsDir() {
			totalSize += int(info.Size())
		}
		return nil
	})
 
	if err != nil {
		slog.Error("Storage:Invalidator:Failed to walk directory", "error", err)
	}
	return
}

func deletedPatternMatchDirectories(contentDir string, cacheManagerObj *cacheManager) (bytesDeleted int, err error) {
	pattern := filepath.Base(contentDir)
	contentDir = filepath.Dir(contentDir)
	err = filepath.Walk(contentDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			/*
			 * Check for directory's existence before attempting deletion.
			 * Skip directory walk if the directory has been already removed.
			 */
			_, statErr := os.Stat(path)
			if statErr != nil {
				if os.IsNotExist(statErr) {
					return nil
				}
				return statErr
			}

			slog.Error("Storage:Invalidator:Failed to walk directory", "error", err)
			return err
		}
		if info.IsDir() && strings.Contains(path, pattern) {
			bytesDeleted += getDirectorySize(path, cacheManagerObj)
			err = os.RemoveAll(path)
			if err != nil {
				slog.Error("Storage:Invalidator:Failed to delete directory", "error", err)
				return err
			}
		}
		return nil
	})
	return
}

/*
 * Function to delete a particular stale delivery service and its associated contents from storage.
 * The DELETE request will come from Config Mgmt API module.
 */
func invalidator(request *http.Request, storageObj *StorageHandler) (response *http.Response, err error) {
	response = &http.Response{}
	var parsedUrl *url.URL = nil
	var bytesDeleted int = 0
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		timeTaken := endTime.Sub(startTime)
		recordStorageMetrics(parsedUrl, request.Host, "delete", int(timeTaken.Milliseconds()), bytesDeleted, storageObj.observabilityObj)
		slog.Info("Storage:Invalidator:Time taken to delete ", "url", request.URL.String(), "host", request.Host, "timeTaken(milliseconds)", timeTaken.Milliseconds())
	}()
	
	parsedUrl, err = url.Parse(request.URL.String())
	if err != nil {
		slog.Error("Storage:Invalidator:Failed to parse GET request URL", "url", request.URL.String(), "error", err)
		response.StatusCode = http.StatusBadRequest
		response.Status = strconv.Itoa(http.StatusBadRequest) + " Bad Request URL"
		return
	}

	slog.Info("Storage:Invalidator:Received request ", "method", request.Method, "url", request.URL.String())
	contentDir, _, _ := fileNames(parsedUrl, request.Host)

	/*
	 * The Invalidator module will treat the URL as a simple directory path and try to locate it.
	 * If the directory is present, it will be deleted. Otherwise, it will perform a pattern match and
	 * delete all matching directories
	 */
	_, err = os.Stat(contentDir)
	if err == nil {
		bytesDeleted = getDirectorySize(contentDir, storageObj.cacheManagerObj)
		slog.Info("Storage:Invalidator:Deletion of directories based on exact match")
		err = os.RemoveAll(contentDir)
		if err != nil {
			slog.Error("Storage:Invalidator:Failed to delete directory:", "dir", contentDir, "error", err)
			response.StatusCode = http.StatusInternalServerError
			response.Status = strconv.Itoa(http.StatusInternalServerError) + " Content directory deletion failed"
			return
		}
	} else {
		slog.Info("Storage:Invalidator:Deletion of directories based on pattern match")
		bytesDeleted, err = deletedPatternMatchDirectories(contentDir, storageObj.cacheManagerObj)
		if err != nil {
			slog.Error("Storage:Invalidator:Failed to walk directory", "error", err)
			response.StatusCode = http.StatusInternalServerError
			response.Status = strconv.Itoa(http.StatusInternalServerError) + " Directory deletion failed"
			return
		}
	}

	slog.Info("Storage:Invalidator:Successfully deleted", "dir", contentDir, "bytesDeleted", bytesDeleted)
	response.StatusCode = http.StatusOK
	return
}

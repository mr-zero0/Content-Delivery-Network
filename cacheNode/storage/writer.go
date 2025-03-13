package storage

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func validateCacheHeaders(requestHeader http.Header) (maxAge int, age int, lastModified string, contentLength int) {
	var err error
	v := requestHeader.Get("Cache-Control")
	if v == "" {
		requestHeader.Set("Cache-Control", "max-age=0")
		maxAge = 0
	} else {
		maxAge, err = strconv.Atoi(strings.TrimPrefix(v, "max-age="))
		if err != nil {
			slog.Error("Storage:Writer:validateCacheHeaders:Failed to convert max-age to int", "error", err)
			maxAge = 0
		}
	}

	v = requestHeader.Get("Last-Modified")
	if v == "" {
		lastModified = time.Now().UTC().Format(http.TimeFormat)
		requestHeader.Set("Last-Modified", lastModified)
	} else {
		lastModified = v
	}

	v = requestHeader.Get("Age")
	if v == "" {
		requestHeader.Set("Age", "0")
		age = 0
	} else {
		age, err = strconv.Atoi(v)
		if err != nil {
			slog.Error("Storage:Writer:validateCacheHeaders:Failed to convert age to int", "error", err)
			age = 0
		}
	}
	lastModifiedTimestamp, _ := http.ParseTime(lastModified)
	age += int(time.Now().Unix() - lastModifiedTimestamp.Unix())

	v = requestHeader.Get("Content-Length")
	if v == "" {
		contentLength = 0
	} else {
		contentLength, err = strconv.Atoi(v)
		if err != nil {
			slog.Error("Storage:Writer:validateCacheHeaders:Failed to convert contentLength to int", "error", err)
			contentLength = 0
		}
	}
	return
}

func createContentDirectory(contentDir string) (response *http.Response, err error) {
	response = &http.Response{}
	/*
	 * Delete the content directory if already present. Treat it as new content
	 */
	_ = os.RemoveAll(contentDir)

	err = os.MkdirAll(contentDir, os.ModePerm)
	if err != nil {
		diskErr, ok := err.(*os.PathError)
		if ok && diskErr.Err == syscall.ENOSPC {
			slog.Error("Storage:Writer:Insufficient Storage to create directory", "dir", contentDir)
			response.StatusCode = http.StatusInsufficientStorage
		} else {
			slog.Error("Storage:Writer:Failed to create content directory", "dir", contentDir, "error", err)
			response.StatusCode = http.StatusInternalServerError
			response.Status = strconv.Itoa(http.StatusInternalServerError) + " Directory Creation Failed"
		}
		return
	}
	return
}

func createFile(fileName string, payload io.ReadCloser, metadata []byte) (numBytes int64, response *http.Response, err error) {
	response = &http.Response{}
	contentFileHandler, err := os.Create(fileName)
	if err != nil {
		diskErr, ok := err.(*os.PathError)
		if ok && diskErr.Err == syscall.ENOSPC {
			slog.Error("Storage:Writer:Insufficient Storage to create content file", "file", fileName)
			response.StatusCode = http.StatusInsufficientStorage
		} else {
			slog.Error("Storage:Writer:Failed to create content file", "file", fileName, "error", err)
			response.StatusCode = http.StatusInternalServerError
			response.Status = strconv.Itoa(http.StatusInternalServerError) + " Content File Creation Failed"
		}
		return
	}
	defer contentFileHandler.Close()

	if payload == nil {
		err = os.WriteFile(fileName, metadata, 0755)
	} else {
		numBytes, err = io.Copy(contentFileHandler, payload)
	}
	if err != nil {
		diskErr, ok := err.(*os.PathError)
		if ok && diskErr.Err == syscall.ENOSPC {
			slog.Error("Storage:Writer:Insufficient Storage to write content file", "file", fileName)
			response.StatusCode = http.StatusInsufficientStorage
		} else {
			slog.Error("Storage:Writer:Failed to create content file", "file", fileName, "error", err)
			response.StatusCode = http.StatusInternalServerError
			response.Status = strconv.Itoa(http.StatusInternalServerError) + " Content File Writing Failed"
		}
	}
	return
}

func writer(request *http.Request, storageObj *StorageHandler) (response *http.Response, err error) {
	response = &http.Response{}
	var parsedUrl *url.URL = nil
	var bytesStored int64 = 0
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		timeTaken := endTime.Sub(startTime)
		recordStorageMetrics(parsedUrl, request.Host, "write", int(timeTaken.Milliseconds()), int(bytesStored), storageObj.observabilityObj)
		slog.Info("Storage:Writer:Time taken to write", "url", request.URL.String(), "host", request.Host, "timeTaken(milliseconds)", timeTaken.Milliseconds())
	}()

	parsedUrl, err = url.Parse(request.URL.String())
	if err != nil {
		slog.Error("Storage:Writer:Failed to parse GET request URL", "url", request.URL.String(), "host", request.Host, "error", err)
		response.StatusCode = http.StatusBadRequest
		response.Status = strconv.Itoa(http.StatusBadRequest) + " Bad Request URL"
		return
	}

	slog.Info("Storage:Writer:Received request ", "method", request.Method, "url", request.URL.String(), "host", request.Host)

	/* Check if content is there */
	if request.Body == nil {
		slog.Error("Storage:Writer:No payload data", "url", request.URL.String(), "host", request.Host)
		response.StatusCode = http.StatusBadRequest
		response.Status = strconv.Itoa(http.StatusBadRequest) + " No Payload"
		return
	}

	contentDir, contentMetaDataFile, contentFile := fileNames(parsedUrl, request.Host)

	/*
	 * Validate Cache headers Max-age, Age & Last-Modified
	 * If below fields are missing in POST request header, then
	 *      Reset Max-age & Age to 0
	 *      Set Last-Modified as current timestamp
	 */
	maxAge, age, lastModified, contentLength := validateCacheHeaders(request.Header)

	// Create the base directory
	response, err = createContentDirectory(contentDir)
	if err != nil {
		return
	}

	// Prepare meta content
	contentMetadata, err := json.Marshal(request.Header)
	if err != nil {
		slog.Error("Storage:Writer:Failed to convert header fields to JSON", "url", request.URL.String(), "host", request.Host, "error", err)
		response.StatusCode = http.StatusBadRequest
		response.Status = strconv.Itoa(http.StatusBadRequest) + " Bad Request"
		return
	}

	/*
	 * Create content metadata file and write metadata to metadata file
	 */
	_, response, err = createFile(contentMetaDataFile, nil, contentMetadata)
	if err != nil {
		return
	}

	/*
	 * Create content file and write payload data to content file
	 */
	bytesStored, response, err = createFile(contentFile, request.Body, nil)
	if err != nil {
		return
	}
	slog.Info("Storage:Writer:Successfully stored", "url", request.URL.String(), "host", request.Host, "bytesStored", bytesStored)
	/*
	 * Update remaining cache duration for newly added content in the in-memory map/slice
	 * So that CacheEvictor will handle deleting contents with shorter remaining cache duration
	 * when the disk threshold exceeds.
	 */
	response.StatusCode = http.StatusCreated
	storageObj.cacheManagerObj.updateCacheContentInMap(maxAge, age, lastModified, contentLength, contentDir)
	return
}

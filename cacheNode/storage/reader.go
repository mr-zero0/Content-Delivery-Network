package storage

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func readContentMetaDataFile(contentMetaDataFile string) (response *http.Response, err error) {
	response = &http.Response{}
	contentMetadata, err := os.ReadFile(contentMetaDataFile)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Storage:Reader:Content metadata Not Found", "path", contentMetaDataFile)
			response.StatusCode = http.StatusNotFound
			response.Status = strconv.Itoa(http.StatusNotFound) + " Content Not Found"
			err = nil
		} else {
			slog.Error("Storage:Reader:Failed to open metadata file", "path", contentMetaDataFile, "error", err)
			response.StatusCode = http.StatusInternalServerError
			response.Status = strconv.Itoa(http.StatusInternalServerError) + " File Read Error"
		}
		return
	}

	err = json.Unmarshal(contentMetadata, &response.Header)
	if err != nil {
		slog.Error("Storage:Reader:Failed to parse metadata", "file", contentMetaDataFile, "error", err)
		response.StatusCode = http.StatusInternalServerError
		response.Status = strconv.Itoa(http.StatusInternalServerError) + " File Parse Error"
		return
	}
	return
}

func updateCacheHeaders(responseHeader *http.Header) {
	if responseHeader.Get("Last-Modified") != "" &&
		responseHeader.Get("Age") != "" &&
		responseHeader.Get("Last-Modified") != "" {
		lastModifiedTimestamp, _ := http.ParseTime(responseHeader.Get("Last-Modified"))
		age, _ := strconv.Atoi(responseHeader.Get("Age"))
		age += int(time.Now().Unix() - lastModifiedTimestamp.Unix())
		responseHeader.Set("Age", strconv.Itoa(age))
	} else {
		slog.Error("Storage:Reader:CDNDATASTORE directory got corrupted. CacheEvictor will behave abnormally...")
	}
}

func getContentFileHandle(contentFile string, responseHeader http.Header) (bytesRead int, response *http.Response, err error) {
	response = &http.Response{}
	contentFileHandle, err := os.Open(contentFile)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Error("Storage:Reader:Content file Not Found", "file", contentFile)
			response.StatusCode = http.StatusNotFound
			response.Status = strconv.Itoa(http.StatusNotFound) + " Content Not Found"
			err = nil
		} else {
			slog.Error("Storage:Reader:Failed to open content file", "file", contentFile, "error", err)
			response.StatusCode = http.StatusInternalServerError
			response.Status = strconv.Itoa(http.StatusInternalServerError) + " File Read Error"
		}
		return
	}
	response.Body = contentFileHandle
	if responseHeader.Get("Content-Length") != "" {
		bytesRead, _ = strconv.Atoi(responseHeader.Get("Content-Length"))
	} else {
		slog.Error("Storage:Reader:Failed to read metadata key Content-Length", "file", contentFile)

	}
	response.Header = responseHeader
	return
}

func reader(request *http.Request, storageObj *StorageHandler) (response *http.Response, err error) {
	response = &http.Response{}
	var parsedUrl *url.URL = nil
	var bytesRead int = 0
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		timeTaken := endTime.Sub(startTime)
		recordStorageMetrics(parsedUrl, request.Host, "read", int(timeTaken.Milliseconds()), bytesRead, storageObj.observabilityObj)
		slog.Info("Storage:Reader:Time taken to read ", "url", request.URL.String(), "host", request.Host, "timeTaken(milliseconds)", timeTaken.Milliseconds())
	}()

	parsedUrl, err = url.Parse(request.URL.String())
	if err != nil {
		slog.Error("Storage:Reader:Failed to parse GET request URL", "url", request.URL.String(), "host", request.Host, "error", err)
		response.StatusCode = http.StatusBadRequest
		response.Status = strconv.Itoa(http.StatusBadRequest) + " Bad Request URL"
		return response, err
	}

	slog.Info("Storage:Reader:Received request ", "method", request.Method, "url", request.URL.String(), "host", request.Host)

	_, contentMetaDataFile, contentFile := fileNames(parsedUrl, request.Host)

	response, err = readContentMetaDataFile(contentMetaDataFile)
	if err != nil || response.StatusCode == http.StatusNotFound {
		return
	}

	/*
	 * Update Age of the content for every GET request in GET request header
	 */
	updateCacheHeaders(&response.Header)
	if request.Method == http.MethodHead {
		response.StatusCode = http.StatusOK
		return
	}
	// EXPECT it to the GET
	bytesRead, response, err = getContentFileHandle(contentFile, response.Header)
	if err != nil || response.StatusCode == http.StatusNotFound {
		return
	}
	slog.Info("Storage:Reader:Successfully read", "url", request.URL.String(), "host", request.Host, "bytesRead", bytesRead)
	response.StatusCode = http.StatusOK
	return
}

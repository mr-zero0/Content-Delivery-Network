package observability

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"
)

// logEventToFile handles writing log messages to the appropriate file
func logEventToFile(eventType string, logMessage string) error {

	fileMapping := map[string]string{
		"FrontendEvent": "frontend.log",
		"BackendEvent":  "backend.log",
		"StorageEvent":  "storage.log",
		"StorageDiskMetricsEvent": "storagedisk.log",
	}
	fileName, exists := fileMapping[eventType]
	if !exists {
		slog.Error("Invalid event type", "eventType", eventType)
		return fmt.Errorf("invalid event type: %s", eventType)
	}
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Invalid event type", "eventType", eventType)
		return fmt.Errorf("invalid event type: %s", eventType)
	}
	defer file.Close()
	logger := log.New(file, eventType+": ", log.Ldate|log.Ltime)
	logger.Println(logMessage)
	return nil
}

// logFrontendEvent processes and logs FrontendEvent
func logFrontendEvent(e FrontendEvent) {
	logMessage := fmt.Sprintf(
		"Timestamp: %s, ClientIP: %s, URL: %s, UserAgent: %s, ResponseTime: %dms, TTFB: %dms, Bytes: %d, StatusCode: %d, CacheHit: %t",
		e.Timestamp.Format(time.RFC3339), e.ClientIP, e.URL, e.UserAgent, e.ResponseTime, e.TTFB, e.Bytes, e.StatusCode, e.CacheHit,
	)
	if err := logEventToFile("FrontendEvent", logMessage); err != nil {
		slog.Info("Error logging frontend event", "error", err)
	}
}

// logBackendEvent processes and logs BackendEvent
func logBackendEvent(e BackendEvent) {
	logMessage := fmt.Sprintf(
		"Timestamp: %s, OriginServerIP: %s, URL: %s, ResponseTime: %dms, TTFB: %dms, Bytes: %d, StatusCode: %d",
		e.Timestamp.Format(time.RFC3339), e.OriginServerIP, e.URL, e.ResponseTime, e.TTFB, e.Bytes, e.StatusCode,
	)
	if err := logEventToFile("BackendEvent", logMessage); err != nil {
		slog.Info("Error logging backend event", "error", err)
	}
}

// logStorageEvent processes and logs StorageEvent
func logStorageEvent(e StorageEvent) {
    logMessage := fmt.Sprintf(
        "Timestamp: %s, URL: %s, ResponseTime: %dms, Bytes: %d, Operation: %s",
        e.Timestamp.Format(time.RFC3339), e.URL, e.ResponseTime, e.Bytes, e.Operation,
    )
    if err := logEventToFile("StorageEvent", logMessage); err != nil {
        slog.Info("Error logging storage event: ", "error", err)
    }
}

// logStorageDiskMetricsEvent processes and logs StorageDiskMetricsEvent
func logStorageDiskMetricsEvent(e StorageDiskMetricsEvent) {
    logMessage := fmt.Sprintf(
        "Timestamp: %s, DiskUsage: %d, TotalContents: %d",
        e.Timestamp.Format(time.RFC3339), e.DiskUsage, e.TotalContents,
    )
    if err := logEventToFile("StorageDiskMetricsEvent", logMessage); err != nil {
        slog.Info("Error logging storage disk metrics event: ", "error", err)
    }
}
package observability

import (
	"strconv"
)

func processsFrontendEvent(e FrontendEvent) {
	statusCodeStr := strconv.Itoa(e.StatusCode) // Convert StatusCode to string
	cacheHitStr := strconv.FormatBool(e.CacheHit)

	fe_total_req_count.WithLabelValues(e.ClientIP, e.URL, e.UserAgent, statusCodeStr, cacheHitStr).Inc()
	fe_total_bytes_transferred.WithLabelValues(e.ClientIP, e.URL, e.UserAgent, statusCodeStr, cacheHitStr).Add(float64(e.Bytes))
	fe_req_time_to_serve_msec.WithLabelValues(e.ClientIP, e.URL, e.UserAgent, statusCodeStr, cacheHitStr).Observe(float64(e.ResponseTime))
	fe_req_ttfb_msec.WithLabelValues(e.ClientIP, e.URL, e.UserAgent, statusCodeStr, cacheHitStr).Observe(float64(e.ResponseTime))
}

func processsBackendEvent(e BackendEvent) {
	statusCodeStr := strconv.Itoa(e.StatusCode) // Convert StatusCode to string

	be_total_req_count.WithLabelValues(e.OriginServerIP, e.URL, statusCodeStr).Inc()
	be_total_bytes_transferred.WithLabelValues(e.OriginServerIP, e.URL, statusCodeStr).Add(float64(e.Bytes))
	be_req_response_time_msec.WithLabelValues(e.OriginServerIP, e.URL, statusCodeStr).Observe(float64(e.ResponseTime))
	be_req_ttfb_msec.WithLabelValues(e.OriginServerIP, e.URL, statusCodeStr).Observe(float64(e.ResponseTime))
}

func processsStorageEvent(e StorageEvent) {
	storage_event_count.WithLabelValues(e.URL, e.Operation).Inc()
	storage_total_bytes_served.WithLabelValues(e.URL, e.Operation).Add(float64(e.Bytes))
	storage_req_response_time_msec.WithLabelValues(e.URL, e.Operation).Observe(float64(e.ResponseTime))
}

func processsStorageDiskMetricsEvent(e StorageDiskMetricsEvent){
    storage_total_contents.Set(float64(e.TotalContents))
    storage_disk_usage_percentage.Set(float64(e.DiskUsage))
}

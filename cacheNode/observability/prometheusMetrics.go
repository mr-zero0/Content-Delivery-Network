package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	// Frontend Metrics
	fe_total_req_count = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:      "fe_total_req_count",
			Help:      "Total request count for frontend",
		},
		[]string{"clientip", "url", "user_agent", "status_code", "cache_hit"},
	)
	fe_total_bytes_transferred = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:      "fe_total_bytes_transferred",
			Help:      "Total bytes transferred for frontend",
		},
		[]string{"clientip", "url", "user_agent", "status_code", "cache_hit"},
	)
	fe_req_time_to_serve_msec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "namespace_mycdn",
			Name:      "fe_time_to_serve",
			Help:      "Time taken to serve frontend requests",
			Buckets:   []float64{10, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"clientip", "url", "user_agent", "status_code", "cache_hit"},
	)
	fe_req_ttfb_msec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "namespace_mycdn",
			Name:      "fe_time_to_first_byte",
			Help:      "Time to first byte for frontend",
			Buckets:   []float64{10, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"clientip", "url", "user_agent", "status_code", "cache_hit"},
	)

	// Backend Metrics
	be_total_req_count = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:      "be_total_req_count",
			Help:      "Total request count for backend",
		},
		[]string{"origin_ip", "url", "status_code"},
	)
	be_total_bytes_transferred = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:      "be_total_bytes_transferred",
			Help:      "Total bytes transferred for backend",
		},
		[]string{"origin_ip", "url", "status_code"},
	)
	be_req_response_time_msec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "namespace_mycdn",
			Name:      "be_response_time",
			Help:      "Response time for backend",
			Buckets:   []float64{10, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"origin_ip", "url", "status_code"},
	)
	be_req_ttfb_msec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "namespace_mycdn",
			Name:      "be_time_to_serve_firstbyte",
			Help:      "Time to serve first byte for backend",
			Buckets:   []float64{10, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"origin_ip", "url", "status_code"},
	)

	// Storage Metrics
	storage_event_count = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:      "storage_event_count",
			Help:      "Storage event count",
		},
		[]string{"url", "operation"},
	)
	storage_total_bytes_served = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:      "storage_total_bytes_served",
			Help:      "Total bytes served from storage",
		},
		[]string{"url", "operation"},
	)
	storage_req_response_time_msec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "namespace_mycdn",
			Name:      "storage_response_time",
			Help:      "Storage response time",
			Buckets:   []float64{10, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"url", "operation"},
	)

	// Storage Disk Metrics
	storage_disk_metrics_event_count = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:      "storage_disk_metrics_event_count",
			Help:      "Storage disk metrics event count",
		},
		[]string{"diskUsage", "totalContents"},
	)
	storage_disk_usage_percentage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name: "storage_disk_usage_percentage",
			Help: "Storage - disk usage in percentage",
		},
	)
	storage_total_contents = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "namespace_mycdn",
			Name:    "storage_total_contents",
			Help:    "Storage - Total mumber of content on disk",
		},

	)
	
)

// RegisterMetrics registers all the metrics with Prometheus
func RegisterPromMetrics() {
	//r := prometheus.NewRegistry()
	prometheus.MustRegister(
		fe_total_req_count,
		fe_total_bytes_transferred,
		fe_req_time_to_serve_msec,
		fe_req_ttfb_msec,
		be_total_req_count,
		be_total_bytes_transferred,
		be_req_response_time_msec,
		be_req_ttfb_msec,
		storage_event_count,
		storage_total_bytes_served,
		storage_req_response_time_msec,
		storage_disk_metrics_event_count,
		storage_disk_usage_percentage,
		storage_total_contents,		
	)

	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	//handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

}

package observability

import (
	"time"
)

type ObservabilityHandler interface {
	RecordEventFrontend(FrontendEvent) 
	RecordEventBackend(BackendEvent)
	RecordEventStorage(StorageEvent)
	RecordEventStorageDiskMetrics(StorageDiskMetricsEvent)
}

type FrontendEvent struct{
	Timestamp time.Time         //time of the event 
	ClientIP string    			//client IP
	URL string   				//URL of the content
	UserAgent string  			// client user agent 
	ResponseTime int  			//response time in miliseconds
 	TTFB int  					//time to first byte in milliseconds
	Bytes int  					//number of bytes of content
	StatusCode int  			//http response code to client 
 	CacheHit bool  				// true or false
 }

type BackendEvent struct{
	Timestamp time.Time         //time of the event 
	OriginServerIP string		//origin server IP
	URL string					//URL of the content
	ResponseTime int			//response time in miliseconds
 	TTFB int					//time to first byte in milliseconds
	Bytes int					//number of bytes of content
	StatusCode int				//http response code to client
}

type StorageEvent struct{
	Timestamp time.Time			//time of the event 
	URL string					//URL of the content
	Operation string			//Read or Write
	ResponseTime int			//response time in miliseconds
	Bytes int					//number of bytes of content
}

type StorageDiskMetricsEvent struct {
    Timestamp time.Time         //time of the event
    DiskUsage int               //CDNDATASTORE disk usage
    TotalContents int            //No. of contents in CDNDATASTORE
} 
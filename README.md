# SIMPLE CDN

CDN Topology Layout for Demo
----------------------------
Origin <--> Mid <--> Edge1
Origin <--> Mid <--> Edge2
 

CDN End-to-End Demo Scenarios:		
--------------------------------
1. Updation of Cache node configurations
	1. Delete cacheNodes.json & deliveryServices.json files in configServer directory.
	2. Open web browser and navigate to http://localhost:8080/ui
	3. Add new cache node entry for Mid 
		- Node Name:    Mid
		- Node IP:      127.0.0.1
		- Node Port:    9000
		- Node Type:    Mid
		- MgmtPort:     10000
		- PromPort:     20000
	4. Add new cache node for Edge1	
		- Node Name:    Edge1
		- Node IP:      127.0.0.1
		- Node Port:    9001
		- Node Type:    Edge
		- Parent IP:    127.0.0.1
		- Parent Port:  9000
		- MgmtPort:     10001
		- PromPort:     20001
	5. Add new cache node for Edge2
		- Node Name:    Edge2
		- Node IP:      127.0.0.1
		- Node Port:    9002
		- Node Type:    Edge
		- Parent IP:    127.0.0.1
		- Parent Port:  9000
		- MgmtPort:     10002
		- PromPort:     20002

2. Updation of Delivery services
	1. Open web browser and navigate to http://localhost:8080/ui
	2. Add new Delivery Service for host google.com
		- Delivery Service Name: ds1
		- Client URL:            http://google.com
		- Origin URL:            http://localhost:32001
		
	Stop the configServer
	3. Edit the file deliveryServices.json include the re-write rules, manually - 
		"rewriteRules": [
        {
          "headerName": "TestHeader",
          "operation": 0,
          "value": "ABCD"
        }
      ]

	Start the configServer
	
 	4. Add new Delivery Service for host msn.com
		- Delivery Service Name: ds2
		- Client URL:            http://msn.com
		- Origin URL:            http://localhost:32002
 
3. Content not found at Edge1, Mid & Origin (Unknown host name: amazon.com)

	1. Get fresh content "a.html" with an unknown host name "amazon.com".
	2. Execute the following invoke-request in client terminal.
		Invoke-WebRequest -Uri "http://127.0.0.1:9001/static/a.html" -v -Method Get -Headers @{"Host" = "amazon.com"}
	3. Verify Edge1 log to ensure request get failed within Edge1 due to unknown host name.
	4. Verify Origin server to ensure the request didn't reaches the origin server.

4. Content not found at Edge1 & Mid, found at origin

	1. Get fresh content "a.html" with configured host name "google.com" and max-age = 0.
	2. Execute the following invoke-request in client terminal.
		Invoke-WebRequest -Uri "http://127.0.0.1:9001/static/a.html" -v -Method Get -Headers @{"Host" = "google.com"}
	3. Verify content retrieval in Edge1 & Mid logs
	4. Verify content directory creation in mid/cdn & edge1/cdn directories
	5. Verify Origin server log to ensure the request reaches the origin server. 
	6. Open observability metrics url in web browser and refresh it to display frontend, backend & storage metrics

5. Content not found at Edge2 & Mid, found at origin

	1. Get fresh content "Europe" with configured host name "msn.com" and max-age = 300.
	2. Execute the following invoke-request in client terminal.
		Invoke-WebRequest -Uri "http://127.0.0.1:9002/dyna/200/256/300/0/Europe" -v -Method Get -Headers @{"Host" = "msn.com"}
	3. Verify content retrieval in Edge2 & Mid logs
	4. Verify content directory creation in mid/cdn & edge1/cdn directories
	5. Verify Origin server log to ensure the request reaches the origin server. 
	6. Open observability metrics url in web browser and refresh it to display frontend, backend & storage metrics
 
6. Content not found at Edge1, found at Mid

	1. Get cached content "Europe" from Mid with configured host name "msn.com".
	2. Execute the following invoke-request in client terminal.
		Invoke-WebRequest -Uri "http://127.0.0.1:9001/dyna/200/256/300/0/Europe" -v -Method Get -Headers @{"Host" = "msn.com"}
	3. Verify content retrieval in Edge1 & Mid logs
	4. Verify content directory creation in edge1/cdn directories
	5. Verify Origin server log to ensure the request didn't reaches the origin server. 
	6. Open observability metrics url in web browser and refresh it to display frontend, backend & storage metrics
 
7. Content expired at Edge1 & Mid, Redo request from Edge1 to Mid
	1. Before starting the invoke request, show the expired content "a.html" is already available in Edge1
	2. Get expired content "a.html" from Edge1 with configured host name "google.com".
		Invoke-WebRequest -Uri "http://127.0.0.1:9001/static/a.html" -v -Method Get -Headers @{"Host" = "google.com"}
	3. Verify content retrieval in Edge1 & Mid logs using redo request.
	4. Verify Origin server to ensure the request reaches the origin server. 
	5. Open observability metrics url in web browser and refresh it to display frontend, backend & storage metrics

8. Invalidate expired contents from Edge1 & Mid(Checked vi UI, close all files first)
	1. Ensure that contents are available in Mid, Edge1 under directories run/mid/cdn & run/edge1/cdn
	2. Open Config UI to trigger invalidate request to all cache nodes
	3. Execute the following invoke-request in client terminal.
		Invoke-WebRequest -Uri http://127.0.0.1:8080/Invalidate/google.com -Method GET
	4. Verify the content deletion in Edge1 & Mid logs
	5. Verify content directory deletion in edge1/cdn & mid/cdn directories
	6. Open observability metrics url in web browser and refresh it to display storage metrics
	
	Checking the status of invalidator - http://localhost:8080/invalidateStatus/1738743162787524500-2431374740746301533(copy id)

9. Trigger CacheEvictor to delete expired contents from Edge1 when its disk usage is high
	1. Ensure that contents are available in Edge1 under directory run/edge2/cdn
	2. If no contents are available, execute the folowing invoke-request
		Invoke-WebRequest -Uri "http://127.0.0.1:9001/dyna/200/256/300/0/Europe" -v -Method Get -Headers @{"Host" = "msn.com"}
		Invoke-WebRequest -Uri "http://127.0.0.1:9001/dyna/200/256/600/0/Asia" -v -Method Get -Headers @{"Host" = "msn.com"}
	3. Simulate disk full by manually creating empty file "SimulateDiskFull" under run/edge1/cdn
	4. Wait for 5 seconds to allow CacheEvictor to trigger cleanup.
	5. Verify the content deletion in Edge1 log
	6. Verify content directory deletion in edge1/cdn directories
	7. Open observability metrics url in web browser and refresh it to display storage disk metrics



Observability metrics names 
 
(
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

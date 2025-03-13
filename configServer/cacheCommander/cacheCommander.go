package cacheCommander

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hcl/cdn/common/mgmtApi" // Import the gRPC protobuf package
)

var (
	pendingInvalidateRequestsMux sync.RWMutex
	pendingInvalidateRequests    map[string]string
)

func updateStatus(uid string, status string) {
	pendingInvalidateRequestsMux.Lock()
	defer pendingInvalidateRequestsMux.Unlock()
	slog.Info("Updating status", "uid", uid, "status", status)
	
	pendingInvalidateRequests[uid] = status
}

func ExecuteInvalidateRequest(uid string, pattern string) {
	if inMemConfig == nil {
		slog.Error("Cache store not initialized")
		return
	}

	bgWg.Add(1)
	go func() {
		defer bgWg.Done()
		slog.Info("Invalidation Request started to process", "uid", uid)

		dialCtx, cancel := context.WithTimeout(bgContext, 30*time.Second)
		defer cancel()
		//dialCtx := bgContext
		updateStatus(uid, "InProgress")
		defer func() {
			slog.Info("Invalidation Request completed on all Nodes", "uid", uid)
			updateStatus(uid, "Completed")
		}()

		opWg := sync.WaitGroup{}
		defer opWg.Wait()

		_, names := inMemConfig.GetCnNames()
		for _, name := range names {
			opWg.Add(1)
			go func(name string) {
				defer opWg.Done()
				slog.Info("Processing cache node for invalidation", "name", name, "uid", uid)
				cn, err := inMemConfig.GetCnDetailByName(name)
				if err != nil {
					slog.Error("Cache node not found", "name", name, "uid", uid, "error", err)
					return
				}
				conn, err := mgmtApi.GetConn(cn.IP, cn.MgmtPort)
				if err != nil {
					slog.Error("Failed to connect to cache node", "name", name, "uid", uid, "error", err)
					return
				}
				defer conn.Close()
				client := mgmtApi.NewMgmtApiClient(conn)

				req := &mgmtApi.InvalidateCacheRequest{
					Pattern:        pattern,
					InvalidationID: uid,
				}

				resp, err := client.InvalidateCache(dialCtx, req)
				if err != nil {
					slog.Error("Failed to invalidate cache on node", "name", name, "uid", uid, "error", err)
					return
				}

				if !resp.Success {
					slog.Error("Cache invalidation failed on node", "name", name, "uid", uid, "error", resp.Message)
					return
				}
				slog.Info("Cache invalidation request successful on node", "name", name, "uid", uid, "response", resp.Message)
			}(name)
		}
	}()
}

func GetInvalidateRequestStatus(uid string) string {
	pendingInvalidateRequestsMux.RLock()
	defer pendingInvalidateRequestsMux.RUnlock()
	status, ok := pendingInvalidateRequests[uid]
	if !ok {
		return "Not Present"
	}
	slog.Info("Checking status", "uid", uid, "status", status)
	return status
}

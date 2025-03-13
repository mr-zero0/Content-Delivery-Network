package mgmtApi

import (
	"context"
	"fmt"
	"net/http"
	"log/slog"

	endec "github.com/hcl/cdn/common/mgmtApi" // Import the endec package with an alias
	pb "github.com/hcl/cdn/common/mgmtApi"
)

// MgmtApiServer implements the API logic for managing cache.
type MgmtApiServer struct {
	pb.UnimplementedMgmtApiServer
}

func (s *MgmtApiServer) UpdateConfigNode(ctx context.Context, req *pb.UpdateConfigNodeRequest) (*pb.UpdateConfigNodeResponse, error) {
	fmt.Printf("UpdateConfigNode called with request: %+v\n", req)

	// CORRECTION: Change the condition
	configNode := endec.ProtoToConfigCN(req.Node)
	if configNode == nil {
		return nil, fmt.Errorf("failed to convert protobuf to Config")
	}

	configReconciler.UpdateCacheNode(configNode)

	return &pb.UpdateConfigNodeResponse{
		Success: true,
		Message: "CacheNode details Pushed successfully",
	}, nil
}

func (s *MgmtApiServer) UpdateDsList(ctx context.Context, req *pb.UpdateDsListRequest) (*pb.UpdateDsListResponse, error) {
	// CORRECTION: Change the condition
	dsList := endec.ProtoToConfigDS(req.GetServiceList())
	if dsList == nil {
		return nil, fmt.Errorf("failed to convert protobuf to Config")
	}

	configReconciler.UpdateDeliveryServices(dsList)

	return &pb.UpdateDsListResponse{
		Success: true,
		Message: "DS list Pushed successfully",
	}, nil
}

// InvalidateCache handles the cache invalidation request.
func (s *MgmtApiServer) InvalidateCache(ctx context.Context, req *pb.InvalidateCacheRequest) (*pb.InvalidateCacheResponse, error) {
	// Get the cache pattern from the request
	slog.Info("InvalidateCache called with request", "request", req)
	cachePattern := req.Pattern

	// Create a new HTTP request
	r, err := http.NewRequestWithContext(ctx, http.MethodDelete, cachePattern, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Use the Do method of the RequestHandler to invalidate the cache
	if store == nil {
		return &pb.InvalidateCacheResponse{
			Success: false,
			Message: "Store not initialized",
		}, nil
	}
	_, err = store.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to invalidate cache: %v", err)
	}

	slog.Info("Cache invalidated successfully", "pattern", cachePattern)
	// Return a successful response

	return &pb.InvalidateCacheResponse{
		Success: true,
		Message: "Cache invalidated successfully",
	}, nil
}

func (s *MgmtApiServer) InvalidateCacheStatus(ctx context.Context, req *pb.InvalidateCacheStatusRequest) (*pb.InvalidateCacheStatusResponse, error) {
	// Implement the logic to invalidate the cache status

	// No NEED for this API

	// Return a successful response
	return &pb.InvalidateCacheStatusResponse{
		Success: true,
		Status: "Cache status invalidated successfully",
	}, nil
}

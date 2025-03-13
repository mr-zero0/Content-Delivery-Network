package mgmtApi

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/hcl/cdn/cacheNode/common"
	"github.com/hcl/cdn/cacheNode/config"
	"github.com/hcl/cdn/common/mgmtApi"
	"google.golang.org/grpc"
)

var configReconciler *config.ConfigReconciler
var store common.RequestHandler

func Init(ctx context.Context, wg *sync.WaitGroup, cr *config.ConfigReconciler, st common.RequestHandler, mgmtPort uint) error {
	configReconciler = cr
	store = st
	configReconciler.SetStorage(st)

	// Listen on all network interfaces, not just localhost
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", mgmtPort))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	srv := grpc.NewServer()
	mgmtApi.RegisterMgmtApiServer(srv, &MgmtApiServer{})

	slog.Info("Server listening on all interfaces", "port", mgmtPort)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer srv.GracefulStop()

		serveErrCh := make(chan error, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			serveErrCh <- srv.Serve(lis)
		}()

		select {
		case <-ctx.Done():
			slog.Info("Shutting down gRPC server...")
		case err := <-serveErrCh:
			if err != nil {
				slog.Error("gRPC server stopped serving", "error", err)
			}
		}
	}()

	return nil
}

func SetStorage(st common.RequestHandler) {
	if configReconciler != nil {
		configReconciler.SetStorage(st)
	}
	store =st
}

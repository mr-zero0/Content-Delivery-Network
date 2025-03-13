package mgmtApi

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetConn(ipStr string, port int) (*grpc.ClientConn, error) {
	if port == 0 {
		// default to MGMT_PORT if not configured
		port = int(DEFAULT_MGMT_PORT)
	}
	ipPortStr := fmt.Sprintf("%s:%d", ipStr, port)
	slog.Info("Network Diagnostics for IP", "ip", ipPortStr)

	conn, err := grpc.NewClient(ipPortStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed creating grpc client", "ip", ipPortStr, "error", err)
		return conn, nil
	}
	return conn, nil
}

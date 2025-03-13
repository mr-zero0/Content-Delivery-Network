package configPusher

import (
	"context"
	"log/slog"
	"sync"
	//"time"

	"github.com/hcl/cdn/common/mgmtApi"
	//"github.com/hcl/cdn/common/config"

)

func PushDsUpdate(ctx context.Context) error {
    protoDsList := inMemConfig.ConfigToProtoDS()
    if len(protoDsList) == 0 {
        slog.Info("No Delivery Services to push")
        return nil
    }

    req := &mgmtApi.UpdateDsListRequest{
        ServiceList: protoDsList,
    }

    _, cnNames := inMemConfig.GetCnNames()
    slog.Info("CN Names retrieved", "cnNames", cnNames)
    bgWg.Add(1)
    go func() {
        defer bgWg.Done()
        var wg sync.WaitGroup
        for _, name := range cnNames {
            wg.Add(1)
            go func(name string) {
                defer wg.Done()
                cn, err := inMemConfig.GetCnDetailByName(name)
                if err != nil {
                    slog.Error("Cache node not found", "name", name, "error", err)
                    return
                }
                if err := pushToSingleNode(ctx, cn.IP, cn.MgmtPort, req); err != nil {
                    slog.Error("Failed to push to node", "name", name, "error", err)
                }
            }(name)
        }
        wg.Wait()
        slog.Info("Completed pushing DS updates to all nodes")
    }()
    return nil
}
func pushToSingleNode(ctx context.Context, ip string, port int, req *mgmtApi.UpdateDsListRequest) error {
	slog.Info("Pushing to single node", "ip", ip, "port", port)
	conn, err := mgmtApi.GetConn(ip, port)
	if err != nil {
		return err
	}
	defer conn.Close()

	//dialCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	//defer cancel()
    dialCtx := bgContext
	client := mgmtApi.NewMgmtApiClient(conn)
	_, err = client.UpdateDsList(dialCtx, req)
	if err != nil {
		slog.Error("UpdateDsList failed for IP", "ip", ip, "error", err)
		return err
	}
	slog.Info("Successfully pushed DS update to node", "ip", ip)
	return nil
}

func PushCnUpdate(nodeName string) error {
    // Convert DeliveryServices to protobuf DeliveryServices
    mgmtPort, protoCn, err := inMemConfig.ConfigToProtoCN(nodeName)
    if err != nil {
        slog.Error("Node not found", "name", nodeName) // Debug info
        return err
    }
    req := &mgmtApi.UpdateConfigNodeRequest{
        Node: protoCn,
    }

    //dialCtx, cancel := context.WithTimeout(bgContext, 120*time.Second)
    //defer cancel()
    dialCtx := bgContext
    conn, err := mgmtApi.GetConn(protoCn.Ip, mgmtPort) // Use MgmtPort for connection
    if err != nil {
        return err
    }
    defer conn.Close()
    client := mgmtApi.NewMgmtApiClient(conn)
    resp, err := client.UpdateConfigNode(dialCtx, req)
    if err != nil {
        slog.Error("Error executing update on node", "name", nodeName, "error", err)
    } else {
        slog.Info("Response from node", "name", nodeName, "resp", resp) // Debug info
    }
    return nil
}
//go:build withgrpc
// +build withgrpc

package grpc

import (
    "context"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    botcfg "budget-bot/internal/pkg/config"
    "go.uber.org/zap"
)

// We will wire actual pb clients to our adapters

func WireClients(log *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient) {
    _, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    cfg, _ := botcfg.Load()
    addr := "127.0.0.1:8080"
    insecure := true
    if cfg != nil {
        if cfg.GRPC.Address != "" { addr = cfg.GRPC.Address }
        insecure = cfg.GRPC.Insecure
    }
    conn, err := Dial(DialOptions{Address: addr, Insecure: insecure})
    if err != nil {
        log.Warn("grpc dial failed, falling back to fakes", zap.Error(err))
        return nil, nil, nil, nil
    }
    cat := NewGRPCCategoryClient(pb.NewCategoryServiceClient(conn))
    rep := NewGRPCReportClient(pb.NewReportServiceClient(conn))
    ten := NewGRPCTenantClient(pb.NewTenantServiceClient(conn))
    tx := NewGRPCTransactionClient(pb.NewTransactionServiceClient(conn))
    return cat, rep, ten, tx
}



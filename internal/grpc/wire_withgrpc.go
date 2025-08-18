//go:build withgrpc
// +build withgrpc

package grpc

import (
    "context"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// We will wire actual pb clients to our adapters

func WireClients(log *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    conn, err := grpc.DialContext(ctx, "127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
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



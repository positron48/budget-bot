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

// placeholder types for real pb clients
type realCategoryClient struct{}
type realReportClient struct{}
type realTenantClient struct{}
type realTransactionClient struct{}

func WireClients(log *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    conn, err := grpc.DialContext(ctx, "127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
    if err != nil {
        log.Warn("grpc dial failed, falling back to fakes", zap.Error(err))
        return &realCategoryClient{}, &realReportClient{}, &realTenantClient{}, &realTransactionClient{}
    }
    _ = conn // TODO: wrap pb clients to our interfaces
    // Examples:
    _ = pb.NewCategoryServiceClient(conn)
    _ = pb.NewReportServiceClient(conn)
    _ = pb.NewTenantServiceClient(conn)
    _ = pb.NewTransactionServiceClient(conn)
    return &realCategoryClient{}, &realReportClient{}, &realTenantClient{}, &realTransactionClient{}
}



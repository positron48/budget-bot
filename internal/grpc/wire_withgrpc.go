//go:build withgrpc
// +build withgrpc

package grpc

import (
    "context"
    "time"

    "go.uber.org/zap"
    "google.golang.org/grpc"
)

// placeholder types for real pb clients
type realCategoryClient struct{}
type realReportClient struct{}
type realTenantClient struct{}
type realTransactionClient struct{}

func WireClients(log *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient) {
    // Here you would dial your backend and initialize real generated clients
    // conn, err := grpc.DialContext(context.Background(), addr, grpc.WithTransportCredentials(...))
    _, _ = grpc.DialContext(context.Background(), "127.0.0.1:8080", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
    return &realCategoryClient{}, &realReportClient{}, &realTenantClient{}, &realTransactionClient{}
}



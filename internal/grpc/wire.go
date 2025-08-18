//go:build !withgrpc
// +build !withgrpc

package grpc

import "go.uber.org/zap"

// WireClients (default build) returns nil clients so the app uses fakes.
// To enable real clients, build with -tags withgrpc and ensure proto is generated.
func WireClients(_ *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient) {
    return nil, nil, nil, nil
}



//go:build withgrpc
// +build withgrpc

package grpc

import (
    "context"
    "crypto/tls"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    botcfg "budget-bot/internal/pkg/config"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/credentials/insecure"
)

// We will wire actual pb clients to our adapters

func WireClients(log *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient, OAuthClient) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    cfg, err := botcfg.Load()
    if err != nil {
        log.Fatal("failed to load config", zap.Error(err))
        return nil, nil, nil, nil, nil
    }
    
    var creds credentials.TransportCredentials
    if cfg.GRPC.Insecure {
        creds = insecure.NewCredentials()
    } else {
        creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
    }
    
    log.Info("attempting to connect to gRPC server", zap.String("address", cfg.GRPC.Address), zap.Bool("insecure", cfg.GRPC.Insecure))
    conn, err := grpc.DialContext(ctx, cfg.GRPC.Address, grpc.WithTransportCredentials(creds))
    if err != nil {
        log.Warn("grpc dial failed, falling back to fakes", zap.Error(err))
        return nil, nil, nil, nil, nil
    }
    log.Info("successfully connected to gRPC server", zap.String("address", cfg.GRPC.Address))
    
    cat := NewGRPCCategoryClient(pb.NewCategoryServiceClient(conn), log)
    rep := NewGRPCReportClient(pb.NewReportServiceClient(conn), log)
    ten := NewGRPCTenantClient(pb.NewTenantServiceClient(conn), log)
    tx := NewGRPCTransactionClient(pb.NewTransactionServiceClient(conn), log)
    oauth := NewOAuthClient(pb.NewAuthServiceClient(conn), log)
    return cat, rep, ten, tx, oauth
}

// WireFxClient dials gRPC and returns a real FxClient when available; otherwise a fake.
func WireFxClient(log *zap.Logger) FxClient {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    cfg, err := botcfg.Load()
    if err != nil {
        log.Fatal("failed to load config for fx client", zap.Error(err))
        return &FakeFxClient{}
    }
    
    var creds credentials.TransportCredentials
    if cfg.GRPC.Insecure {
        creds = insecure.NewCredentials()
    } else {
        creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
    }
    
    conn, err := grpc.DialContext(ctx, cfg.GRPC.Address, grpc.WithTransportCredentials(creds))
    if err != nil {
        log.Warn("grpc dial failed for fx, using fake", zap.Error(err))
        return &FakeFxClient{}
    }
    return NewGRPCFxClient(pb.NewFxServiceClient(conn))
}



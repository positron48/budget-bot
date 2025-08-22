//go:build withgrpc
// +build withgrpc

package main

import (
    botpkg "budget-bot/internal/bot"
    grpcwire "budget-bot/internal/grpc"
    pb "budget-bot/internal/pb/budget/v1"
    "budget-bot/internal/pkg/config"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func makeAuthClient(log *zap.Logger, cfg *config.Config) botpkg.AuthClient {
    addr := cfg.GRPC.Address
    if addr == "" { addr = "127.0.0.1:8081" }
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Warn("auth grpc dial failed, using fake", zap.Error(err))
        return &fakeAuthClient{}
    }
    return grpcwire.NewAuthClient(pb.NewAuthServiceClient(conn), log)
}



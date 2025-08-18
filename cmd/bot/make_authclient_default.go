//go:build !withgrpc
// +build !withgrpc

package main

import (
    botpkg "budget-bot/internal/bot"
    "budget-bot/internal/pkg/config"
    "go.uber.org/zap"
)

func makeAuthClient(_ *zap.Logger, _ *config.Config) botpkg.AuthClient {
    return &fakeAuthClient{}
}



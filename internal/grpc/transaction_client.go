package grpc

import (
    "context"
    "time"
)

type CreateTransactionRequest struct {
    TenantID    string
    Type        string
    AmountMinor int64
    Currency    string
    Description string
    CategoryID  string
    OccurredAt  time.Time
}

type TransactionClient interface {
    CreateTransaction(ctx context.Context, req *CreateTransactionRequest, accessToken string) (string, error)
}

// FakeTransactionClient is a temporary stub.
type FakeTransactionClient struct{}

func (f *FakeTransactionClient) CreateTransaction(ctx context.Context, req *CreateTransactionRequest, accessToken string) (string, error) {
    return "tx-" + req.Description, nil
}



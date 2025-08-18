package grpc

import (
    "context"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc/metadata"
    "google.golang.org/protobuf/types/known/timestamppb"
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

type GRPCTransactionClient struct{ client pb.TransactionServiceClient }

func NewGRPCTransactionClient(c pb.TransactionServiceClient) *GRPCTransactionClient { return &GRPCTransactionClient{client: c} }

func (g *GRPCTransactionClient) CreateTransaction(ctx context.Context, req *CreateTransactionRequest, accessToken string) (string, error) {
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    pbReq := &pb.CreateTransactionRequest{
        Type:      mapType(req.Type),
        CategoryId: req.CategoryID,
        Amount:    &pb.Money{CurrencyCode: req.Currency, MinorUnits: req.AmountMinor},
        OccurredAt: timestamppb.New(req.OccurredAt),
        Comment:   req.Description,
    }
    res, err := g.client.CreateTransaction(ctx, pbReq)
    if err != nil { return "", err }
    return res.Transaction.Id, nil
}

func mapType(t string) pb.TransactionType {
    switch t {
    case "income": return pb.TransactionType_TRANSACTION_TYPE_INCOME
    case "expense": return pb.TransactionType_TRANSACTION_TYPE_EXPENSE
    default: return pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
    }
}



package grpc

import (
    "context"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "google.golang.org/grpc/metadata"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// CreateTransactionRequest is an app-level request to create a transaction.
type CreateTransactionRequest struct {
    TenantID    string
    Type        string
    AmountMinor int64
    Currency    string
    Description string
    CategoryID  string
    OccurredAt  time.Time
}

// TransactionClient exposes transaction operations.
type TransactionClient interface {
    CreateTransaction(ctx context.Context, req *CreateTransactionRequest, accessToken string) (string, error)
    ListRecent(ctx context.Context, tenantID string, limit int, accessToken string) ([]*pb.Transaction, error)
    ListForExport(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*pb.Transaction, error)
}

// FakeTransactionClient is a temporary stub.
// FakeTransactionClient is a stubbed client for tests/local runs.
type FakeTransactionClient struct{}

// CreateTransaction returns a fake transaction id.
func (f *FakeTransactionClient) CreateTransaction(_ context.Context, req *CreateTransactionRequest, _ string) (string, error) {
    return "tx-" + req.Description, nil
}

// ListRecent returns an empty list in the fake client.
func (f *FakeTransactionClient) ListRecent(_ context.Context, _ string, _ int, _ string) ([]*pb.Transaction, error) {
    return []*pb.Transaction{}, nil
}

// ListForExport returns an empty list in the fake client.
func (f *FakeTransactionClient) ListForExport(_ context.Context, _ string, _ , _ time.Time, _ int, _ string) ([]*pb.Transaction, error) {
    return []*pb.Transaction{}, nil
}

// TransactionGRPCClient calls Transaction service via gRPC.
type TransactionGRPCClient struct{ client pb.TransactionServiceClient }

// NewGRPCTransactionClient constructs a TransactionGRPCClient.
func NewGRPCTransactionClient(c pb.TransactionServiceClient) *TransactionGRPCClient { return &TransactionGRPCClient{client: c} }

// CreateTransaction creates a transaction and returns its id.
func (g *TransactionGRPCClient) CreateTransaction(ctx context.Context, req *CreateTransactionRequest, accessToken string) (string, error) {
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

// ListRecent returns recent transactions.
func (g *TransactionGRPCClient) ListRecent(ctx context.Context, tenantID string, limit int, accessToken string) ([]*pb.Transaction, error) {
    _ = tenantID
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    if limit <= 0 { limit = 10 }
    res, err := g.client.ListTransactions(ctx, &pb.ListTransactionsRequest{
        Page: &pb.PageRequest{Page: 1, PageSize: int32(limit), Sort: "occurred_at desc"},
    })
    if err != nil { return nil, err }
    return res.GetTransactions(), nil
}

// ListForExport returns transactions for a period.
func (g *TransactionGRPCClient) ListForExport(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*pb.Transaction, error) {
    _ = tenantID
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    pr := &pb.PageRequest{Page: 1, PageSize: int32(limit)}
    if limit <= 0 { pr.PageSize = 100 }
    dr := &pb.DateRange{From: timestamppb.New(from), To: timestamppb.New(to)}
    res, err := g.client.ListTransactions(ctx, &pb.ListTransactionsRequest{Page: pr, DateRange: dr})
    if err != nil { return nil, err }
    return res.GetTransactions(), nil
}

func mapType(t string) pb.TransactionType {
    switch t {
    case "income": return pb.TransactionType_TRANSACTION_TYPE_INCOME
    case "expense": return pb.TransactionType_TRANSACTION_TYPE_EXPENSE
    default: return pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
    }
}



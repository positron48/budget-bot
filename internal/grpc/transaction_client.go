package grpc

import (
	"context"
	"fmt"
	"math"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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
	UpdateTransactionCategory(ctx context.Context, txID, categoryID, accessToken string) error
	ListRecent(ctx context.Context, tenantID string, limit int, accessToken string) ([]*pb.Transaction, error)
	ListForExport(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*pb.Transaction, error)
}

// FakeTransactionClient is a temporary stub.
// FakeTransactionClient is a stubbed client for tests/local runs.
type FakeTransactionClient struct{}

// CreateTransaction returns a fake transaction id.
func (f *FakeTransactionClient) CreateTransaction(_ context.Context, req *CreateTransactionRequest, _ string) (string, error) {
	if req.Description == "FAIL" {
		return "", fmt.Errorf("forced failure")
	}
	return "tx-" + req.Description, nil
}

func (f *FakeTransactionClient) UpdateTransactionCategory(_ context.Context, _, _, _ string) error {
	return nil
}

// ListRecent returns an empty list in the fake client.
func (f *FakeTransactionClient) ListRecent(_ context.Context, _ string, _ int, _ string) ([]*pb.Transaction, error) {
	return []*pb.Transaction{}, nil
}

// ListForExport returns an empty list in the fake client.
func (f *FakeTransactionClient) ListForExport(_ context.Context, _ string, _, _ time.Time, _ int, _ string) ([]*pb.Transaction, error) {
	return []*pb.Transaction{}, nil
}

// TransactionGRPCClient calls Transaction service via gRPC.
type TransactionGRPCClient struct {
	client pb.TransactionServiceClient
	logger *zap.Logger
}

// NewGRPCTransactionClient constructs a TransactionGRPCClient.
func NewGRPCTransactionClient(c pb.TransactionServiceClient, logger *zap.Logger) *TransactionGRPCClient {
	return &TransactionGRPCClient{client: c, logger: logger}
}

// CreateTransaction creates a transaction and returns its id.
func (g *TransactionGRPCClient) CreateTransaction(ctx context.Context, req *CreateTransactionRequest, accessToken string) (string, error) {
	g.logger.Debug("CreateTransaction request",
		zap.String("tenantID", req.TenantID),
		zap.String("type", req.Type),
		zap.Int64("amountMinor", req.AmountMinor),
		zap.String("currency", req.Currency),
		zap.String("description", req.Description),
		zap.String("categoryID", req.CategoryID),
		zap.Time("occurredAt", req.OccurredAt),
		zap.String("accessToken", accessToken[:int(math.Min(float64(len(accessToken)), 10))]+"..."))

	if accessToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
	}
	pbReq := &pb.CreateTransactionRequest{
		Type:       mapType(req.Type),
		CategoryId: req.CategoryID,
		Amount:     &pb.Money{CurrencyCode: req.Currency, MinorUnits: req.AmountMinor},
		OccurredAt: timestamppb.New(req.OccurredAt),
		Comment:    req.Description,
	}

	g.logger.Debug("CreateTransaction gRPC request",
		zap.String("type", pbReq.Type.String()),
		zap.String("categoryId", pbReq.CategoryId),
		zap.String("currency", pbReq.Amount.CurrencyCode),
		zap.Int64("amountMinor", pbReq.Amount.MinorUnits),
		zap.String("comment", pbReq.Comment))

	res, err := g.client.CreateTransaction(ctx, pbReq)
	if err != nil {
		g.logger.Error("CreateTransaction gRPC call failed", zap.Error(err))
		return "", err
	}

	g.logger.Debug("CreateTransaction gRPC response",
		zap.String("transactionId", res.Transaction.Id))

	return res.Transaction.Id, nil
}

// UpdateTransactionCategory updates only category for an existing transaction.
func (g *TransactionGRPCClient) UpdateTransactionCategory(ctx context.Context, txID, categoryID, accessToken string) error {
	if accessToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
	}
	_, err := g.client.UpdateTransaction(ctx, &pb.UpdateTransactionRequest{
		Id: txID,
		Transaction: &pb.Transaction{
			CategoryId: categoryID,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"category_id"}},
	})
	return err
}

// ListRecent returns recent transactions.
func (g *TransactionGRPCClient) ListRecent(ctx context.Context, tenantID string, limit int, accessToken string) ([]*pb.Transaction, error) {
	g.logger.Debug("ListRecent request",
		zap.String("tenantID", tenantID),
		zap.Int("limit", limit),
		zap.String("accessToken", accessToken[:int(math.Min(float64(len(accessToken)), 10))]+"..."))

	_ = tenantID
	if accessToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
	}
	if limit <= 0 {
		limit = 10
	}

	req := &pb.ListTransactionsRequest{
		Page: &pb.PageRequest{Page: 1, PageSize: int32(limit), Sort: "occurred_at desc"},
	}

	g.logger.Debug("ListRecent gRPC request",
		zap.Int32("page", req.Page.Page),
		zap.Int32("pageSize", req.Page.PageSize),
		zap.String("sort", req.Page.Sort))

	res, err := g.client.ListTransactions(ctx, req)
	if err != nil {
		g.logger.Error("ListRecent gRPC call failed", zap.Error(err))
		return nil, err
	}

	transactions := res.GetTransactions()
	g.logger.Debug("ListRecent gRPC response",
		zap.Int("transactionsCount", len(transactions)))

	return transactions, nil
}

// ListForExport returns transactions for a period.
func (g *TransactionGRPCClient) ListForExport(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*pb.Transaction, error) {
	g.logger.Debug("ListForExport request",
		zap.String("tenantID", tenantID),
		zap.Time("from", from),
		zap.Time("to", to),
		zap.Int("limit", limit),
		zap.String("accessToken", accessToken[:int(math.Min(float64(len(accessToken)), 10))]+"..."))

	_ = tenantID
	if accessToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
	}
	pr := &pb.PageRequest{Page: 1, PageSize: int32(limit)}
	if limit <= 0 {
		pr.PageSize = 100
	}
	dr := &pb.DateRange{From: timestamppb.New(from), To: timestamppb.New(to)}

	req := &pb.ListTransactionsRequest{Page: pr, DateRange: dr}

	g.logger.Debug("ListForExport gRPC request",
		zap.Int32("page", req.Page.Page),
		zap.Int32("pageSize", req.Page.PageSize),
		zap.Time("dateFrom", from),
		zap.Time("dateTo", to))

	res, err := g.client.ListTransactions(ctx, req)
	if err != nil {
		g.logger.Error("ListForExport gRPC call failed", zap.Error(err))
		return nil, err
	}

	transactions := res.GetTransactions()
	g.logger.Debug("ListForExport gRPC response",
		zap.Int("transactionsCount", len(transactions)))

	return transactions, nil
}

func mapType(t string) pb.TransactionType {
	switch t {
	case "income":
		return pb.TransactionType_TRANSACTION_TYPE_INCOME
	case "expense":
		return pb.TransactionType_TRANSACTION_TYPE_EXPENSE
	default:
		return pb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
}

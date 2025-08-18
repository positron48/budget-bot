package grpc

import (
    "context"
    "time"

    pb "budget-bot/internal/pb/budget/v1"
    "budget-bot/internal/domain"
    "google.golang.org/grpc/metadata"
)

type ReportClient interface {
    GetStats(ctx context.Context, tenantID string, from, to time.Time, accessToken string) (*domain.Stats, error)
    TopCategories(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*domain.CategoryTotal, error)
    Recent(ctx context.Context, tenantID string, limit int, accessToken string) ([]string, error)
}

type FakeReportClient struct{}

func (f *FakeReportClient) GetStats(ctx context.Context, tenantID string, from, to time.Time, _ string) (*domain.Stats, error) {
    return &domain.Stats{Period: from.Format("2006-01-02") + ".." + to.Format("2006-01-02"), TotalIncome: 2500000, TotalExpense: 1750000, Currency: "RUB"}, nil
}

func (f *FakeReportClient) TopCategories(ctx context.Context, tenantID string, from, to time.Time, limit int, _ string) ([]*domain.CategoryTotal, error) {
    return []*domain.CategoryTotal{{CategoryID: "cat-food", Name: "Питание", SumMinor: 500000, Currency: "RUB"}, {CategoryID: "cat-transport", Name: "Транспорт", SumMinor: 300000, Currency: "RUB"}}, nil
}

func (f *FakeReportClient) Recent(ctx context.Context, tenantID string, limit int, _ string) ([]string, error) {
    return []string{"-1000 продукты", "-300 такси", "+50000 зарплата"}, nil
}

type GRPCReportClient struct{ client pb.ReportServiceClient }

func NewGRPCReportClient(c pb.ReportServiceClient) *GRPCReportClient { return &GRPCReportClient{client: c} }

func (g *GRPCReportClient) GetStats(ctx context.Context, tenantID string, from, to time.Time, accessToken string) (*domain.Stats, error) {
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    res, err := g.client.GetMonthlySummary(ctx, &pb.GetMonthlySummaryRequest{Year: int32(from.Year()), Month: int32(from.Month())})
    if err != nil { return nil, err }
    return &domain.Stats{Period: from.Format("2006-01") , TotalIncome: res.TotalIncome.MinorUnits, TotalExpense: res.TotalExpense.MinorUnits, Currency: res.TotalIncome.CurrencyCode}, nil
}

func (g *GRPCReportClient) TopCategories(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*domain.CategoryTotal, error) {
    // Not defined in proto; return empty for now (token attached for future use)
    if accessToken != "" { _ = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    return []*domain.CategoryTotal{}, nil
}

func (g *GRPCReportClient) Recent(ctx context.Context, tenantID string, limit int, accessToken string) ([]string, error) {
    // Not defined in proto; return empty for now (token attached for future use)
    if accessToken != "" { _ = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    return []string{}, nil
}



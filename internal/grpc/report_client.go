// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
    "context"
    "math"
    "time"
    "sort"

    pb "budget-bot/internal/pb/budget/v1"
    "budget-bot/internal/domain"
    "google.golang.org/grpc/metadata"
    "go.uber.org/zap"
)

// ReportClient exposes read-only reporting operations.
type ReportClient interface {
    GetStats(ctx context.Context, tenantID string, from, to time.Time, accessToken string) (*domain.Stats, error)
    TopCategories(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*domain.CategoryTotal, error)
    Recent(ctx context.Context, tenantID string, limit int, accessToken string) ([]string, error)
}

// FakeReportClient is a stubbed implementation for tests and local runs.
type FakeReportClient struct{}

// GetStats returns a fake stats response.
func (f *FakeReportClient) GetStats(_ context.Context, tenantID string, from, to time.Time, _ string) (*domain.Stats, error) {
    _ = tenantID
    return &domain.Stats{Period: from.Format("2006-01-02") + ".." + to.Format("2006-01-02"), TotalIncome: 2500000, TotalExpense: 1750000, Currency: "RUB"}, nil
}

// TopCategories returns fake top categories.
func (f *FakeReportClient) TopCategories(_ context.Context, tenantID string, from, to time.Time, limit int, _ string) ([]*domain.CategoryTotal, error) {
    _ = tenantID; _ = from; _ = to; _ = limit
    return []*domain.CategoryTotal{{CategoryID: "cat-food", Name: "Питание", SumMinor: 500000, Currency: "RUB"}, {CategoryID: "cat-transport", Name: "Транспорт", SumMinor: 300000, Currency: "RUB"}}, nil
}

// Recent returns fake recent lines.
func (f *FakeReportClient) Recent(_ context.Context, tenantID string, limit int, _ string) ([]string, error) {
    _ = tenantID; _ = limit
    return []string{"-1000 продукты", "-300 такси", "+50000 зарплата"}, nil
}

// ReportGRPCClient calls remote Report service via gRPC.
type ReportGRPCClient struct{ 
    client pb.ReportServiceClient 
    logger *zap.Logger
}

// NewGRPCReportClient constructs a GRPCReportClient.
func NewGRPCReportClient(c pb.ReportServiceClient, logger *zap.Logger) *ReportGRPCClient { 
    return &ReportGRPCClient{client: c, logger: logger} 
}

// GetStats fetches monthly stats for a period.
func (g *ReportGRPCClient) GetStats(ctx context.Context, tenantID string, from, _ time.Time, accessToken string) (*domain.Stats, error) {
    g.logger.Debug("GetStats request", 
        zap.String("tenantID", tenantID),
        zap.Time("from", from),
        zap.String("accessToken", accessToken[:int(math.Min(float64(len(accessToken)), 10))] + "..."))
    
    if accessToken != "" { 
        ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) 
    }
    
    // Calculate timezone offset in minutes (JS-style: minutes to add to local time to get UTC)
    _, offset := from.Zone()
    timezoneOffsetMinutes := int32(-offset / 60) // Negative because JS getTimezoneOffset() is opposite of Go
    
    req := &pb.GetMonthlySummaryRequest{
        Year: int32(from.Year()), 
        Month: int32(from.Month()),
        TimezoneOffsetMinutes: timezoneOffsetMinutes,
    }
    g.logger.Debug("GetStats gRPC request", 
        zap.Int32("year", req.Year),
        zap.Int32("month", req.Month),
        zap.Int32("timezoneOffsetMinutes", req.TimezoneOffsetMinutes))
    
    res, err := g.client.GetMonthlySummary(ctx, req)
    if err != nil { 
        g.logger.Error("GetStats gRPC call failed", zap.Error(err))
        return nil, err 
    }
    
    g.logger.Debug("GetStats gRPC response", 
        zap.Int64("totalIncome", res.TotalIncome.MinorUnits),
        zap.Int64("totalExpense", res.TotalExpense.MinorUnits),
        zap.String("currency", res.TotalIncome.CurrencyCode))
    
    return &domain.Stats{
        Period: from.Format("2006-01"), 
        TotalIncome: res.TotalIncome.MinorUnits, 
        TotalExpense: res.TotalExpense.MinorUnits, 
        Currency: res.TotalIncome.CurrencyCode,
    }, nil
}

// TopCategories returns top expense categories for the period.
func (g *ReportGRPCClient) TopCategories(ctx context.Context, tenantID string, from, _ time.Time, limit int, accessToken string) ([]*domain.CategoryTotal, error) {
    g.logger.Debug("TopCategories request", 
        zap.String("tenantID", tenantID),
        zap.Time("from", from),
        zap.Int("limit", limit),
        zap.String("accessToken", accessToken[:int(math.Min(float64(len(accessToken)), 10))] + "..."))
    
    if accessToken != "" { 
        ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) 
    }
    
    // Calculate timezone offset in minutes (JS-style: minutes to add to local time to get UTC)
    _, offset := from.Zone()
    timezoneOffsetMinutes := int32(-offset / 60) // Negative because JS getTimezoneOffset() is opposite of Go
    
    req := &pb.GetMonthlySummaryRequest{
        Year: int32(from.Year()), 
        Month: int32(from.Month()),
        TimezoneOffsetMinutes: timezoneOffsetMinutes,
    }
    g.logger.Debug("TopCategories gRPC request", 
        zap.Int32("year", req.Year),
        zap.Int32("month", req.Month),
        zap.Int32("timezoneOffsetMinutes", req.TimezoneOffsetMinutes))
    
    res, err := g.client.GetMonthlySummary(ctx, req)
    if err != nil { 
        g.logger.Error("TopCategories gRPC call failed", zap.Error(err))
        return nil, err 
    }
    
    g.logger.Debug("TopCategories gRPC response", 
        zap.Int("itemsCount", len(res.GetItems())))
    
    // Collect only expense categories and sort by total desc
    items := res.GetItems()
    out := make([]*domain.CategoryTotal, 0, len(items))
    for _, it := range items {
        if it.GetType() != pb.TransactionType_TRANSACTION_TYPE_EXPENSE { continue }
        total := it.GetTotal()
        out = append(out, &domain.CategoryTotal{CategoryID: it.GetCategoryId(), Name: it.GetCategoryName(), SumMinor: total.GetMinorUnits(), Currency: total.GetCurrencyCode()})
    }
    // sort desc by SumMinor
    sort.Slice(out, func(i, j int) bool { return out[i].SumMinor > out[j].SumMinor })
    if limit > 0 && len(out) > limit { out = out[:limit] }
    
    g.logger.Debug("TopCategories processed", 
        zap.Int("expenseItems", len(out)),
        zap.Int("limit", limit))
    
    return out, nil
}

// Recent returns recent transactions as strings (not implemented by backend).
func (g *ReportGRPCClient) Recent(ctx context.Context, _ string, _ int, accessToken string) ([]string, error) {
    g.logger.Debug("Recent request", 
        zap.String("accessToken", accessToken[:10] + "..."))
    
    // Not defined in proto; return empty for now (token attached for future use)
    if accessToken != "" { 
        _ = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) 
    }
    return []string{}, nil
}



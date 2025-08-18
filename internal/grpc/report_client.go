package grpc

import (
    "context"
    "time"

    "budget-bot/internal/domain"
)

type ReportClient interface {
    GetStats(ctx context.Context, tenantID string, from, to time.Time) (*domain.Stats, error)
    TopCategories(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.CategoryTotal, error)
    Recent(ctx context.Context, tenantID string, limit int) ([]string, error)
}

type FakeReportClient struct{}

func (f *FakeReportClient) GetStats(ctx context.Context, tenantID string, from, to time.Time) (*domain.Stats, error) {
    return &domain.Stats{Period: from.Format("2006-01-02") + ".." + to.Format("2006-01-02"), TotalIncome: 2500000, TotalExpense: 1750000, Currency: "RUB"}, nil
}

func (f *FakeReportClient) TopCategories(ctx context.Context, tenantID string, from, to time.Time, limit int) ([]*domain.CategoryTotal, error) {
    return []*domain.CategoryTotal{{CategoryID: "cat-food", Name: "Питание", SumMinor: 500000, Currency: "RUB"}, {CategoryID: "cat-transport", Name: "Транспорт", SumMinor: 300000, Currency: "RUB"}}, nil
}

func (f *FakeReportClient) Recent(ctx context.Context, tenantID string, limit int) ([]string, error) {
    return []string{"-1000 продукты", "-300 такси", "+50000 зарплата"}, nil
}



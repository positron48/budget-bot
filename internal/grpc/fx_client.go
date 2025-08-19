package grpc

import (
	"context"
	"strconv"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FxClient provides access to foreign exchange rates.
type FxClient interface {
	// GetRate returns a decimal exchange rate from -> to at a given date.
	// If accessToken is non-empty, it will be attached as authorization metadata.
	GetRate(ctx context.Context, fromCurrency, toCurrency string, asOf time.Time, accessToken string) (float64, error)
}

// FakeFxClient is a stubbed client returning a 1.0 rate for identical currencies
// and 1.0 otherwise as a placeholder.
type FakeFxClient struct{}

func (f *FakeFxClient) GetRate(ctx context.Context, fromCurrency, toCurrency string, asOf time.Time, accessToken string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}
	// Default placeholder rate
	return 1.0, nil
}

type GRPCFxClient struct{ client pb.FxServiceClient }

func NewGRPCFxClient(c pb.FxServiceClient) *GRPCFxClient { return &GRPCFxClient{client: c} }

func (g *GRPCFxClient) GetRate(ctx context.Context, fromCurrency, toCurrency string, asOf time.Time, accessToken string) (float64, error) {
	if accessToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
	}
	res, err := g.client.GetRate(ctx, &pb.GetRateRequest{
		FromCurrencyCode: fromCurrency,
		ToCurrencyCode:   toCurrency,
		AsOf:             timestamppb.New(asOf),
	})
	if err != nil {
		return 0, err
	}
	rateStr := "1"
	if res.GetRate() != nil && res.GetRate().GetRateDecimal() != "" {
		rateStr = res.GetRate().GetRateDecimal()
	}
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, err
	}
	return rate, nil
}



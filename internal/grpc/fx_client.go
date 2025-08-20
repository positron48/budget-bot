// Package grpc contains gRPC client facades used by the bot.
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

// FakeFxClient is a stubbed client returning a 1.0 rate.
type FakeFxClient struct{}

// GetRate returns a decimal exchange rate; stubbed to 1.0 for now.
func (f *FakeFxClient) GetRate(_ context.Context, fromCurrency, toCurrency string, _ time.Time, _ string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}
	// Default placeholder rate
	return 1.0, nil
}

// FxGRPCClient calls Fx service via gRPC.
type FxGRPCClient struct{ client pb.FxServiceClient }

// NewGRPCFxClient constructs a FxGRPCClient.
func NewGRPCFxClient(c pb.FxServiceClient) *FxGRPCClient { return &FxGRPCClient{client: c} }

// GetRate returns a decimal exchange rate for given currencies/date.
func (g *FxGRPCClient) GetRate(ctx context.Context, fromCurrency, toCurrency string, asOf time.Time, accessToken string) (float64, error) {
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



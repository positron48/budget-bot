package bot

import (
	"context"
	"testing"
	"time"

	grpcclient "budget-bot/internal/grpc"
	"go.uber.org/zap"
)

func TestCurrencyConverter_GetRateAndConvert_Cache(t *testing.T) {
	fx := &grpcclient.FakeFxClient{}
	cc := NewCurrencyConverter(fx, zap.NewNop())
	rate, err := cc.GetExchangeRate(context.Background(), "USD", "USD", time.Now(), "")
	if err != nil || rate != 1.0 { t.Fatalf("rate: %v %v", rate, err) }
	amount, err := cc.ConvertToBaseCurrency(context.Background(), 100, "USD", "USD", time.Now(), "")
	if err != nil || amount != 100 { t.Fatalf("convert: %v %v", amount, err) }
	// Different currency uses Fake 1.0, will fill cache
	rate2, err := cc.GetExchangeRate(context.Background(), "USD", "EUR", time.Now(), "")
	if err != nil || rate2 != 1.0 { t.Fatalf("rate2: %v %v", rate2, err) }
	// Second call should hit cache, still 1.0
	rate3, err := cc.GetExchangeRate(context.Background(), "USD", "EUR", time.Now(), "")
	if err != nil || rate3 != 1.0 { t.Fatalf("rate3: %v %v", rate3, err) }
}

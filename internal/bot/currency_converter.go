// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"
	"sync"
	"time"

	grpcclient "budget-bot/internal/grpc"
	"go.uber.org/zap"
)

// CurrencyConverter converts between currencies using FxClient with simple in-memory caching.
type CurrencyConverter struct {
	fxClient grpcclient.FxClient
	logger  *zap.Logger
	cache   *fxCache
}

type fxCache struct {
	mu    sync.RWMutex
	data  map[string]cachedRate // key: from|to|YYYY-MM-DD
}

type cachedRate struct {
	rate     float64
	storedAt time.Time
}

// NewCurrencyConverter constructs CurrencyConverter.
func NewCurrencyConverter(fx grpcclient.FxClient, logger *zap.Logger) *CurrencyConverter {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CurrencyConverter{fxClient: fx, logger: logger, cache: &fxCache{data: make(map[string]cachedRate)}}
}

// GetExchangeRate fetches or caches FX rate.
func (cc *CurrencyConverter) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time, accessToken string) (float64, error) {
	if fromCurrency == toCurrency {
		return 1.0, nil
	}
	key := cc.cacheKey(fromCurrency, toCurrency, date)
	if r, ok := cc.cacheGet(key); ok {
		return r, nil
	}
	rate, err := cc.fxClient.GetRate(ctx, fromCurrency, toCurrency, date, accessToken)
	if err != nil {
		cc.logger.Warn("fx get rate failed", zap.Error(err), zap.String("from", fromCurrency), zap.String("to", toCurrency))
		return 0, err
	}
	cc.cacheSet(key, rate)
	return rate, nil
}

// ConvertToBaseCurrency converts amount from fromCurrency to toCurrency for a given date.
func (cc *CurrencyConverter) ConvertToBaseCurrency(ctx context.Context, amountMinor int64, fromCurrency, toCurrency string, date time.Time, accessToken string) (int64, error) {
	rate, err := cc.GetExchangeRate(ctx, fromCurrency, toCurrency, date, accessToken)
	if err != nil {
		return 0, err
	}
	// multiply and round to nearest minor unit; use banker's rounding if needed later
	converted := float64(amountMinor) * rate
	return int64(converted + 0.5), nil
}

func (cc *CurrencyConverter) cacheKey(from, to string, date time.Time) string {
	return from + "|" + to + "|" + date.Format("2006-01-02")
}

func (cc *CurrencyConverter) cacheGet(key string) (float64, bool) {
	cc.cache.mu.RLock()
	defer cc.cache.mu.RUnlock()
	val, ok := cc.cache.data[key]
	if !ok {
		return 0, false
	}
	// cache TTL 24h per date key, but since key contains date, we can keep infinitely; still guard staleness
	if time.Since(val.storedAt) > 24*time.Hour {
		return 0, false
	}
	return val.rate, true
}

func (cc *CurrencyConverter) cacheSet(key string, rate float64) {
	cc.cache.mu.Lock()
	cc.cache.data[key] = cachedRate{rate: rate, storedAt: time.Now()}
	cc.cache.mu.Unlock()
}



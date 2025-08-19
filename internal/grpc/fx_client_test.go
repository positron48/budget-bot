package grpc

import (
	"context"
	"testing"
	"time"
)

func TestFakeFxClient_GetRate(t *testing.T) {
	fx := &FakeFxClient{}
	r, err := fx.GetRate(context.Background(), "RUB", "RUB", time.Now(), "")
	if err != nil || r != 1.0 { t.Fatalf("same cur rate: %v %v", r, err) }
	r2, err := fx.GetRate(context.Background(), "RUB", "USD", time.Now(), "")
	if err != nil || r2 != 1.0 { t.Fatalf("diff cur placeholder: %v %v", r2, err) }
}

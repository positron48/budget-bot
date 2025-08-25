package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFakeTransactionClient_ListRecent(t *testing.T) {
	client := &FakeTransactionClient{}
	ctx := context.Background()
	
	transactions, err := client.ListRecent(ctx, "tenant_123", 10, "access_token")
	
	assert.NoError(t, err)
	assert.Empty(t, transactions)
}

func TestFakeTransactionClient_ListForExport(t *testing.T) {
	client := &FakeTransactionClient{}
	ctx := context.Background()
	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()
	
	transactions, err := client.ListForExport(ctx, "tenant_123", from, to, 100, "access_token")
	
	assert.NoError(t, err)
	assert.Empty(t, transactions)
}

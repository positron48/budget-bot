package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFakeTransactionClient_CreateTransaction(t *testing.T) {
	client := &FakeTransactionClient{}
	ctx := context.Background()
	
	req := &CreateTransactionRequest{
		TenantID:    "tenant_123",
		Type:        "expense",
		AmountMinor: 10000,
		Currency:    "RUB",
		Description: "такси",
		CategoryID:  "cat_123",
		OccurredAt:  time.Now(),
	}
	
	transactionID, err := client.CreateTransaction(ctx, req, "access_token")
	
	assert.NoError(t, err)
	assert.Equal(t, "tx-такси", transactionID)
}

func TestFakeTransactionClient_CreateTransaction_EmptyDescription(t *testing.T) {
	client := &FakeTransactionClient{}
	ctx := context.Background()
	
	req := &CreateTransactionRequest{
		TenantID:    "tenant_123",
		Type:        "income",
		AmountMinor: 5000,
		Currency:    "USD",
		Description: "",
		CategoryID:  "cat_456",
		OccurredAt:  time.Now(),
	}
	
	transactionID, err := client.CreateTransaction(ctx, req, "access_token")
	
	assert.NoError(t, err)
	assert.Equal(t, "tx-", transactionID)
}

func TestFakeTransactionClient_CreateTransaction_SpecialCharacters(t *testing.T) {
	client := &FakeTransactionClient{}
	ctx := context.Background()
	
	req := &CreateTransactionRequest{
		TenantID:    "tenant_123",
		Type:        "expense",
		AmountMinor: 1500,
		Currency:    "EUR",
		Description: "café & restaurant",
		CategoryID:  "cat_789",
		OccurredAt:  time.Now(),
	}
	
	transactionID, err := client.CreateTransaction(ctx, req, "access_token")
	
	assert.NoError(t, err)
	assert.Equal(t, "tx-café & restaurant", transactionID)
}

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

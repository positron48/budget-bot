package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticCategoryClient_CreateCategory(t *testing.T) {
	client := &StaticCategoryClient{}
	ctx := context.Background()
	
	_, err := client.CreateCategory(ctx, "tenant_123", "FOOD", "Питание", "access_token")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category creation not supported without gRPC")
}

func TestStaticCategoryClient_UpdateCategoryName(t *testing.T) {
	client := &StaticCategoryClient{}
	ctx := context.Background()
	
	_, err := client.UpdateCategoryName(ctx, "tenant_123", "cat_123", "Новое название", "access_token")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category update not supported without gRPC")
}

func TestStaticCategoryClient_DeleteCategory(t *testing.T) {
	client := &StaticCategoryClient{}
	ctx := context.Background()
	
	err := client.DeleteCategory(ctx, "tenant_123", "access_token")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category delete not supported without gRPC")
}

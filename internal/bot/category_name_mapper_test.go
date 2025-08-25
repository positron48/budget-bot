package bot

import (
	"context"
	"testing"

	"budget-bot/internal/domain"
)

// MockCategoryClient for testing
type MockCategoryClient struct {
	categories []*domain.Category
}

func (m *MockCategoryClient) ListCategories(_ context.Context, _ string, _ string, _ domain.TransactionType, _ ...string) ([]*domain.Category, error) {
	return m.categories, nil
}

func (m *MockCategoryClient) CreateCategory(_ context.Context, _ string, _ string, _ string, _ string) (*domain.Category, error) {
	return nil, nil
}

func (m *MockCategoryClient) UpdateCategoryName(_ context.Context, _ string, _ string, _ string, _ string) (*domain.Category, error) {
	return nil, nil
}

func (m *MockCategoryClient) DeleteCategory(_ context.Context, _ string, _ string) error {
	return nil
}

func TestCategoryNameMapper_GetCategoryIDByName(t *testing.T) {
	mockClient := &MockCategoryClient{
		categories: []*domain.Category{
			{ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
			{ID: "cat-transport", Name: "Транспорт", Emoji: "🚗"},
			{ID: "cat-home", Name: "Дом", Emoji: "🏠"},
		},
	}

	mapper := NewCategoryNameMapper(mockClient)
	ctx := context.Background()

	// Test exact match
	id, err := mapper.GetCategoryIDByName(ctx, "tenant1", "token", "Питание", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != "cat-food" {
		t.Fatalf("Expected cat-food, got %s", id)
	}

	// Test case-insensitive match
	id, err = mapper.GetCategoryIDByName(ctx, "tenant1", "token", "питание", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != "cat-food" {
		t.Fatalf("Expected cat-food, got %s", id)
	}

	// Test with spaces
	id, err = mapper.GetCategoryIDByName(ctx, "tenant1", "token", "  Питание  ", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != "cat-food" {
		t.Fatalf("Expected cat-food, got %s", id)
	}

	// Test not found
	id, err = mapper.GetCategoryIDByName(ctx, "tenant1", "token", "Несуществующая", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id != "" {
		t.Fatalf("Expected empty string, got %s", id)
	}
}

func TestCategoryNameMapper_GetCategoryNameByID(t *testing.T) {
	mockClient := &MockCategoryClient{
		categories: []*domain.Category{
			{ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
			{ID: "cat-transport", Name: "Транспорт", Emoji: "🚗"},
		},
	}

	mapper := NewCategoryNameMapper(mockClient)
	ctx := context.Background()

	// Test existing ID
	name, err := mapper.GetCategoryNameByID(ctx, "tenant1", "token", "cat-food", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if name != "Питание" {
		t.Fatalf("Expected Питание, got %s", name)
	}

	// Test not found
	name, err = mapper.GetCategoryNameByID(ctx, "tenant1", "token", "non-existent", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if name != "" {
		t.Fatalf("Expected empty string, got %s", name)
	}
}

func TestCategoryNameMapper_GetCategoryByID(t *testing.T) {
	mockClient := &MockCategoryClient{
		categories: []*domain.Category{
			{ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
		},
	}

	mapper := NewCategoryNameMapper(mockClient)
	ctx := context.Background()

	// Test existing ID
	category, err := mapper.GetCategoryByID(ctx, "tenant1", "token", "cat-food", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if category == nil {
		t.Fatalf("Expected category, got nil")
	}
	if category.ID != "cat-food" || category.Name != "Питание" || category.Emoji != "🍽️" {
		t.Fatalf("Expected cat-food/Питание/🍽️, got %s/%s/%s", category.ID, category.Name, category.Emoji)
	}

	// Test not found
	category, err = mapper.GetCategoryByID(ctx, "tenant1", "token", "non-existent", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if category != nil {
		t.Fatalf("Expected nil, got %v", category)
	}
}

func TestCategoryNameMapper_DisplayNameInMessages(t *testing.T) {
	mockClient := &MockCategoryClient{
		categories: []*domain.Category{
			{ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
			{ID: "cat-transport", Name: "Транспорт", Emoji: "🚗"},
		},
	}

	mapper := NewCategoryNameMapper(mockClient)
	ctx := context.Background()

	// Test that we can get display name for confirmation messages
	displayName, err := mapper.GetCategoryNameByID(ctx, "tenant1", "token", "cat-food", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if displayName != "Питание" {
		t.Fatalf("Expected Питание, got %s", displayName)
	}

	// Test fallback to ID when name not found
	displayName, err = mapper.GetCategoryNameByID(ctx, "tenant1", "token", "non-existent", domain.TransactionExpense, "ru")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if displayName != "" {
		t.Fatalf("Expected empty string, got %s", displayName)
	}
}

// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"
	"strings"

	"budget-bot/internal/domain"
	grpcclient "budget-bot/internal/grpc"
)

// CategoryNameMapper maps category names to IDs and vice versa.
type CategoryNameMapper struct {
	categoryClient grpcclient.CategoryClient
}

// NewCategoryNameMapper constructs a CategoryNameMapper.
func NewCategoryNameMapper(categoryClient grpcclient.CategoryClient) *CategoryNameMapper {
	return &CategoryNameMapper{categoryClient: categoryClient}
}

// GetCategoryIDByName finds category ID by name (case-insensitive).
func (cnm *CategoryNameMapper) GetCategoryIDByName(ctx context.Context, tenantID, accessToken, name string, transactionType domain.TransactionType, locale string) (string, error) {
	categories, err := cnm.categoryClient.ListCategories(ctx, tenantID, accessToken, transactionType, locale)
	if err != nil {
		return "", err
	}

	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, category := range categories {
		if strings.ToLower(strings.TrimSpace(category.Name)) == normalizedName {
			return category.ID, nil
		}
	}

	return "", nil // Not found
}

// GetCategoryNameByID finds category name by ID.
func (cnm *CategoryNameMapper) GetCategoryNameByID(ctx context.Context, tenantID, accessToken, categoryID string, transactionType domain.TransactionType, locale string) (string, error) {
	categories, err := cnm.categoryClient.ListCategories(ctx, tenantID, accessToken, transactionType, locale)
	if err != nil {
		return "", err
	}

	for _, category := range categories {
		if category.ID == categoryID {
			return category.Name, nil
		}
	}

	return "", nil // Not found
}

// GetCategoryByID finds full category by ID.
func (cnm *CategoryNameMapper) GetCategoryByID(ctx context.Context, tenantID, accessToken, categoryID string, transactionType domain.TransactionType, locale string) (*domain.Category, error) {
	categories, err := cnm.categoryClient.ListCategories(ctx, tenantID, accessToken, transactionType, locale)
	if err != nil {
		return nil, err
	}

	for _, category := range categories {
		if category.ID == categoryID {
			return category, nil
		}
	}

	return nil, nil // Not found
}

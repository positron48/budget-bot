package grpc

import (
    "context"

    "budget-bot/internal/domain"
)

type CategoryClient interface {
    ListCategories(ctx context.Context, tenantID string) ([]*domain.Category, error)
}

// StaticCategoryClient is a temporary implementation returning fixed categories.
type StaticCategoryClient struct{}

func (s *StaticCategoryClient) ListCategories(_ context.Context, _ string) ([]*domain.Category, error) {
    return []*domain.Category{
        {ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
        {ID: "cat-transport", Name: "Транспорт", Emoji: "🚗"},
        {ID: "cat-home", Name: "Дом", Emoji: "🏠"},
        {ID: "cat-other", Name: "Другое", Emoji: "🎯"},
    }, nil
}



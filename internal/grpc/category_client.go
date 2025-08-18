package grpc

import (
    "context"

    pb "budget-bot/internal/pb/budget/v1"
    "budget-bot/internal/domain"
    "google.golang.org/grpc/metadata"
)

type CategoryClient interface {
    ListCategories(ctx context.Context, tenantID string, accessToken string) ([]*domain.Category, error)
}

// StaticCategoryClient is a temporary implementation returning fixed categories.
type StaticCategoryClient struct{}

func (s *StaticCategoryClient) ListCategories(_ context.Context, _ string, _ string) ([]*domain.Category, error) {
    return []*domain.Category{
        {ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
        {ID: "cat-transport", Name: "Транспорт", Emoji: "🚗"},
        {ID: "cat-home", Name: "Дом", Emoji: "🏠"},
        {ID: "cat-other", Name: "Другое", Emoji: "🎯"},
    }, nil
}

type GRPCCategoryClient struct{
    client pb.CategoryServiceClient
}

func NewGRPCCategoryClient(c pb.CategoryServiceClient) *GRPCCategoryClient { return &GRPCCategoryClient{client: c} }

func (g *GRPCCategoryClient) ListCategories(ctx context.Context, _ string, accessToken string) ([]*domain.Category, error) {
    if accessToken != "" {
        ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
    }
    res, err := g.client.ListCategories(ctx, &pb.ListCategoriesRequest{IncludeInactive: false})
    if err != nil {
        return nil, err
    }
    var out []*domain.Category
    for _, c := range res.Categories {
        name := c.Code
        if len(c.Translations) > 0 && c.Translations[0].Name != "" {
            name = c.Translations[0].Name
        }
        out = append(out, &domain.Category{ID: c.Id, Name: name})
    }
    return out, nil
}



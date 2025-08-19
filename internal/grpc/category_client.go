package grpc

import (
	"context"
	"fmt"

	pb "budget-bot/internal/pb/budget/v1"
	"budget-bot/internal/domain"
	"google.golang.org/grpc/metadata"
)

type CategoryClient interface {
    ListCategories(ctx context.Context, tenantID string, accessToken string, locale ...string) ([]*domain.Category, error)
    CreateCategory(ctx context.Context, accessToken string, code string, name string, locale string) (*domain.Category, error)
    UpdateCategoryName(ctx context.Context, accessToken string, id string, name string, locale string) (*domain.Category, error)
    DeleteCategory(ctx context.Context, accessToken string, id string) error
}

// StaticCategoryClient is a temporary implementation returning fixed categories.
type StaticCategoryClient struct{}

func (s *StaticCategoryClient) ListCategories(_ context.Context, _ string, _ string, _ ...string) ([]*domain.Category, error) {
    return []*domain.Category{
        {ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
        {ID: "cat-transport", Name: "Транспорт", Emoji: "🚗"},
        {ID: "cat-home", Name: "Дом", Emoji: "🏠"},
        {ID: "cat-other", Name: "Другое", Emoji: "🎯"},
    }, nil
}

func (s *StaticCategoryClient) CreateCategory(ctx context.Context, accessToken string, code string, name string, locale string) (*domain.Category, error) {
    return nil, fmt.Errorf("category creation not supported without gRPC")
}

func (s *StaticCategoryClient) UpdateCategoryName(ctx context.Context, accessToken string, id string, name string, locale string) (*domain.Category, error) {
    return nil, fmt.Errorf("category update not supported without gRPC")
}

func (s *StaticCategoryClient) DeleteCategory(ctx context.Context, accessToken string, id string) error {
    return fmt.Errorf("category delete not supported without gRPC")
}

type GRPCCategoryClient struct{
    client pb.CategoryServiceClient
}

func NewGRPCCategoryClient(c pb.CategoryServiceClient) *GRPCCategoryClient { return &GRPCCategoryClient{client: c} }

func (g *GRPCCategoryClient) ListCategories(ctx context.Context, _ string, accessToken string, locale ...string) ([]*domain.Category, error) {
    if accessToken != "" {
        ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
    }
    req := &pb.ListCategoriesRequest{IncludeInactive: false}
    if len(locale) > 0 && locale[0] != "" { req.Locale = locale[0] }
    res, err := g.client.ListCategories(ctx, req)
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

func (g *GRPCCategoryClient) CreateCategory(ctx context.Context, accessToken string, code string, name string, locale string) (*domain.Category, error) {
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    if locale == "" { locale = "ru" }
    req := &pb.CreateCategoryRequest{
        Kind:        pb.CategoryKind_CATEGORY_KIND_EXPENSE,
        Code:        code,
        IsActive:    true,
        Translations: []*pb.CategoryTranslation{{Locale: locale, Name: name}},
    }
    res, err := g.client.CreateCategory(ctx, req)
    if err != nil { return nil, err }
    cat := res.GetCategory()
    if cat == nil { return nil, fmt.Errorf("empty response") }
    out := &domain.Category{ID: cat.GetId(), Name: name}
    return out, nil
}

func (g *GRPCCategoryClient) UpdateCategoryName(ctx context.Context, accessToken string, id string, name string, locale string) (*domain.Category, error) {
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    if locale == "" { locale = "ru" }
    req := &pb.UpdateCategoryRequest{
        Id:           id,
        Translations: []*pb.CategoryTranslation{{Locale: locale, Name: name}},
    }
    res, err := g.client.UpdateCategory(ctx, req)
    if err != nil { return nil, err }
    cat := res.GetCategory()
    if cat == nil { return nil, fmt.Errorf("empty response") }
    out := &domain.Category{ID: cat.GetId(), Name: name}
    return out, nil
}

func (g *GRPCCategoryClient) DeleteCategory(ctx context.Context, accessToken string, id string) error {
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    _, err := g.client.DeleteCategory(ctx, &pb.DeleteCategoryRequest{Id: id})
    return err
}



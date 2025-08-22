// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
	"context"
	"fmt"

	pb "budget-bot/internal/pb/budget/v1"
	"budget-bot/internal/domain"
	"google.golang.org/grpc/metadata"
	"go.uber.org/zap"
)

// CategoryClient exposes category operations.
type CategoryClient interface {
    ListCategories(ctx context.Context, tenantID string, accessToken string, transactionType domain.TransactionType, locale ...string) ([]*domain.Category, error)
    CreateCategory(ctx context.Context, accessToken string, code string, name string, locale string) (*domain.Category, error)
    UpdateCategoryName(ctx context.Context, accessToken string, id string, name string, locale string) (*domain.Category, error)
    DeleteCategory(ctx context.Context, accessToken string, id string) error
}

// StaticCategoryClient is a temporary implementation returning fixed categories.
// StaticCategoryClient is a temporary implementation returning fixed categories.
type StaticCategoryClient struct{}

// ListCategories returns a static list of categories.
func (s *StaticCategoryClient) ListCategories(_ context.Context, _ string, _ string, transactionType domain.TransactionType, _ ...string) ([]*domain.Category, error) {
    if transactionType == domain.TransactionIncome {
        return []*domain.Category{
            {ID: "cat-salary", Name: "Зарплата", Emoji: "💰"},
            {ID: "cat-bonus", Name: "Премия", Emoji: "🎁"},
            {ID: "cat-investment", Name: "Инвестиции", Emoji: "📈"},
            {ID: "cat-other-income", Name: "Другое", Emoji: "💵"},
        }, nil
    }
    // Default to expense categories
    return []*domain.Category{
        {ID: "cat-food", Name: "Питание", Emoji: "🍽️"},
        {ID: "cat-transport", Name: "Транспорт", Emoji: "🚗"},
        {ID: "cat-home", Name: "Дом", Emoji: "🏠"},
        {ID: "cat-other", Name: "Другое", Emoji: "🎯"},
    }, nil
}

// CreateCategory is unsupported in static client.
func (s *StaticCategoryClient) CreateCategory(_ context.Context, _ string, _ string, _ string, _ string) (*domain.Category, error) {
    return nil, fmt.Errorf("category creation not supported without gRPC")
}

// UpdateCategoryName is unsupported in static client.
func (s *StaticCategoryClient) UpdateCategoryName(_ context.Context, _ string, _ string, _ string, _ string) (*domain.Category, error) {
    return nil, fmt.Errorf("category update not supported without gRPC")
}

// DeleteCategory is unsupported in static client.
func (s *StaticCategoryClient) DeleteCategory(_ context.Context, _ string, _ string) error {
    return fmt.Errorf("category delete not supported without gRPC")
}

// CategoryGRPCClient calls Category service via gRPC.
type CategoryGRPCClient struct{
    client pb.CategoryServiceClient
    logger *zap.Logger
}

// NewGRPCCategoryClient constructs a CategoryGRPCClient.
func NewGRPCCategoryClient(c pb.CategoryServiceClient, logger *zap.Logger) *CategoryGRPCClient { 
    return &CategoryGRPCClient{client: c, logger: logger} 
}

// ListCategories returns categories with optional locale translation.
func (g *CategoryGRPCClient) ListCategories(ctx context.Context, _ string, accessToken string, transactionType domain.TransactionType, locale ...string) ([]*domain.Category, error) {
    g.logger.Debug("ListCategories request", 
        zap.String("transactionType", string(transactionType)),
        zap.String("accessToken", accessToken[:10] + "..."),
        zap.Strings("locale", locale))
    
    if accessToken != "" {
        ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
    }
    
    // Determine category kind based on transaction type
    var kind pb.CategoryKind
    if transactionType == domain.TransactionIncome {
        kind = pb.CategoryKind_CATEGORY_KIND_INCOME
    } else {
        kind = pb.CategoryKind_CATEGORY_KIND_EXPENSE
    }
    
    req := &pb.ListCategoriesRequest{
        Kind: kind,
        IncludeInactive: false,
    }
    if len(locale) > 0 && locale[0] != "" { req.Locale = locale[0] }
    
    g.logger.Debug("ListCategories gRPC request", 
        zap.String("kind", kind.String()),
        zap.Bool("includeInactive", req.IncludeInactive),
        zap.String("locale", req.Locale))
    
    res, err := g.client.ListCategories(ctx, req)
    if err != nil {
        g.logger.Error("ListCategories gRPC call failed", zap.Error(err))
        return nil, err
    }
    
    g.logger.Debug("ListCategories gRPC response", 
        zap.Int("categoriesCount", len(res.Categories)))
    
    var out []*domain.Category
    for _, c := range res.Categories {
        name := c.Code
        if len(c.Translations) > 0 && c.Translations[0].Name != "" {
            name = c.Translations[0].Name
        }
        out = append(out, &domain.Category{ID: c.Id, Name: name})
    }
    
    g.logger.Debug("ListCategories processed", 
        zap.Int("categoriesReturned", len(out)))
    
    return out, nil
}

// CreateCategory creates a new category.
func (g *CategoryGRPCClient) CreateCategory(ctx context.Context, accessToken string, code string, name string, locale string) (*domain.Category, error) {
    g.logger.Debug("CreateCategory request", 
        zap.String("code", code),
        zap.String("name", name),
        zap.String("locale", locale),
        zap.String("accessToken", accessToken[:10] + "..."))
    
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    if locale == "" { locale = "ru" }
    req := &pb.CreateCategoryRequest{
        Kind:        pb.CategoryKind_CATEGORY_KIND_EXPENSE,
        Code:        code,
        IsActive:    true,
        Translations: []*pb.CategoryTranslation{{Locale: locale, Name: name}},
    }
    
    g.logger.Debug("CreateCategory gRPC request", 
        zap.String("kind", req.Kind.String()),
        zap.String("code", req.Code),
        zap.Bool("isActive", req.IsActive))
    
    res, err := g.client.CreateCategory(ctx, req)
    if err != nil { 
        g.logger.Error("CreateCategory gRPC call failed", zap.Error(err))
        return nil, err 
    }
    
    cat := res.GetCategory()
    if cat == nil { 
        g.logger.Error("CreateCategory empty response")
        return nil, fmt.Errorf("empty response") 
    }
    
    g.logger.Debug("CreateCategory gRPC response", 
        zap.String("categoryId", cat.GetId()))
    
    out := &domain.Category{ID: cat.GetId(), Name: name}
    return out, nil
}

// UpdateCategoryName updates category translation name.
func (g *CategoryGRPCClient) UpdateCategoryName(ctx context.Context, accessToken string, id string, name string, locale string) (*domain.Category, error) {
    g.logger.Debug("UpdateCategoryName request", 
        zap.String("id", id),
        zap.String("name", name),
        zap.String("locale", locale),
        zap.String("accessToken", accessToken[:10] + "..."))
    
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    if locale == "" { locale = "ru" }
    req := &pb.UpdateCategoryRequest{
        Id:           id,
        Translations: []*pb.CategoryTranslation{{Locale: locale, Name: name}},
    }
    
    g.logger.Debug("UpdateCategoryName gRPC request", 
        zap.String("id", req.Id))
    
    res, err := g.client.UpdateCategory(ctx, req)
    if err != nil { 
        g.logger.Error("UpdateCategoryName gRPC call failed", zap.Error(err))
        return nil, err 
    }
    
    cat := res.GetCategory()
    if cat == nil { 
        g.logger.Error("UpdateCategoryName empty response")
        return nil, fmt.Errorf("empty response") 
    }
    
    g.logger.Debug("UpdateCategoryName gRPC response", 
        zap.String("categoryId", cat.GetId()))
    
    out := &domain.Category{ID: cat.GetId(), Name: name}
    return out, nil
}

// DeleteCategory deletes a category by id.
func (g *CategoryGRPCClient) DeleteCategory(ctx context.Context, accessToken string, id string) error {
    g.logger.Debug("DeleteCategory request", 
        zap.String("id", id),
        zap.String("accessToken", accessToken[:10] + "..."))
    
    if accessToken != "" { ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken) }
    
    req := &pb.DeleteCategoryRequest{Id: id}
    g.logger.Debug("DeleteCategory gRPC request", 
        zap.String("id", req.Id))
    
    _, err := g.client.DeleteCategory(ctx, req)
    if err != nil {
        g.logger.Error("DeleteCategory gRPC call failed", zap.Error(err))
    } else {
        g.logger.Debug("DeleteCategory gRPC call successful")
    }
    return err
}



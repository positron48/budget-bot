package bot

import (
    "context"
    "strings"

    "budget-bot/internal/repository"
)

type CategoryMatcher struct {
    mappingRepo repository.CategoryMappingRepository
}

func NewCategoryMatcher(repo repository.CategoryMappingRepository) *CategoryMatcher {
    return &CategoryMatcher{mappingRepo: repo}
}

// FindCategory tries to find a category by exact or partial keyword match.
func (cm *CategoryMatcher) FindCategory(ctx context.Context, tenantID string, description string) (*repository.CategoryMapping, error) {
    // exact match first
    words := strings.Fields(strings.ToLower(description))
    for _, w := range words {
        if m, err := cm.mappingRepo.FindMapping(ctx, tenantID, w); err == nil && m != nil {
            return m, nil
        }
    }
    // partial match over all mappings ordered by priority
    all, err := cm.mappingRepo.ListMappings(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    low := strings.ToLower(description)
    var best *repository.CategoryMapping
    for _, m := range all {
        if strings.Contains(low, strings.ToLower(m.Keyword)) {
            if best == nil || m.Priority > best.Priority {
                best = m
            }
        }
    }
    return best, nil
}



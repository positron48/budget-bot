package llm

import "context"

// CategoryOption is a candidate category passed to LLM.
type CategoryOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SuggestCategoryRequest is the input for category suggestion.
type SuggestCategoryRequest struct {
	Description     string
	TransactionType string
	Locale          string
	Categories      []CategoryOption
}

// SuggestCategoryResponse is a validated response from LLM.
type SuggestCategoryResponse struct {
	CategoryID  string
	Probability float64
	Reason      string
}

// CategorySuggester suggests category from a bounded list.
type CategorySuggester interface {
	SuggestCategory(ctx context.Context, req SuggestCategoryRequest) (*SuggestCategoryResponse, error)
}

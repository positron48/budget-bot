package llm

import (
	"encoding/json"
	"fmt"
)

// BuildPrompt returns system and user prompts for OpenRouter.
func BuildPrompt(req SuggestCategoryRequest) (string, string, error) {
	payload := struct {
		Description     string           `json:"description"`
		TransactionType string           `json:"transaction_type"`
		Locale          string           `json:"locale"`
		Categories      []CategoryOption `json:"categories"`
	}{
		Description:     req.Description,
		TransactionType: req.TransactionType,
		Locale:          req.Locale,
		Categories:      req.Categories,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}
	system := "You are a transaction category classifier. Return strict JSON only: {\"category_id\":\"...\",\"probability\":0..1,\"reason\":\"...\"}. If uncertain, keep probability low."
	user := fmt.Sprintf("Select best category from provided list. Input: %s", string(b))
	return system, user, nil
}

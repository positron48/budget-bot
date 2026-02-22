package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenRouterClientSuggestCategory(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"category_id\":\"cat-1\",\"probability\":0.8,\"reason\":\"coffee\"}"}}]}`))
	}))
	defer ts.Close()

	c := NewOpenRouterClient(ts.URL, "k", "m", 0)
	resp, err := c.SuggestCategory(context.Background(), SuggestCategoryRequest{
		Description:     "кофе",
		TransactionType: "expense",
		Locale:          "ru",
		Categories:      []CategoryOption{{ID: "cat-1", Name: "Еда"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.CategoryID != "cat-1" {
		t.Fatalf("unexpected category: %s", resp.CategoryID)
	}
}

func TestOpenRouterClientSuggestCategoryInvalid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"category_id\":\"cat-2\",\"probability\":1.5}"}}]}`))
	}))
	defer ts.Close()

	c := NewOpenRouterClient(ts.URL, "k", "m", 0)
	_, err := c.SuggestCategory(context.Background(), SuggestCategoryRequest{
		Description:     "кофе",
		TransactionType: "expense",
		Locale:          "ru",
		Categories:      []CategoryOption{{ID: "cat-1", Name: "Еда"}},
	})
	if err == nil {
		t.Fatalf("expected err")
	}
}

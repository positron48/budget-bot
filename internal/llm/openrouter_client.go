package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// OpenRouterClient implements CategorySuggester via OpenRouter chat/completions API.
type OpenRouterClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
	initErr error
}

// NewOpenRouterClient creates a configured client.
func NewOpenRouterClient(baseURL, apiKey, model string, timeout time.Duration) *OpenRouterClient {
	return NewOpenRouterClientWithSocks5Proxy(baseURL, apiKey, model, timeout, "")
}

// NewOpenRouterClientWithSocks5Proxy creates a configured client and optionally routes
// OpenRouter requests through a SOCKS5 proxy. Expected proxy formats:
// - "ip:port"
// - "socks5://ip:port"
// - "socks5://user:pass@ip:port"
func NewOpenRouterClientWithSocks5Proxy(baseURL, apiKey, model string, timeout time.Duration, socks5Proxy string) *OpenRouterClient {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	httpClient := &http.Client{Timeout: timeout}
	if strings.TrimSpace(socks5Proxy) != "" {
		if transport, err := newSocks5Transport(socks5Proxy); err == nil {
			httpClient.Transport = transport
		} else {
			return &OpenRouterClient{
				baseURL: baseURL,
				apiKey:  apiKey,
				model:   model,
				http:    httpClient,
				initErr: fmt.Errorf("openrouter socks5 proxy is invalid: %w", err),
			}
		}
	}
	return &OpenRouterClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		http:    httpClient,
	}
}

func newSocks5Transport(raw string) (*http.Transport, error) {
	hostPort, auth, err := normalizeSocks5Proxy(raw)
	if err != nil {
		return nil, err
	}
	dialer, err := proxy.SOCKS5("tcp", hostPort, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          10,
	}, nil
}

func normalizeSocks5Proxy(raw string) (hostPort string, auth *proxy.Auth, err error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil, fmt.Errorf("empty socks5 proxy")
	}
	if !strings.Contains(value, "://") {
		return value, nil, nil
	}
	u, err := url.Parse(value)
	if err != nil {
		return "", nil, fmt.Errorf("parse socks5 proxy: %w", err)
	}
	if u.Scheme != "socks5" && u.Scheme != "socks5h" {
		return "", nil, fmt.Errorf("unsupported socks5 proxy scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return "", nil, fmt.Errorf("socks5 proxy host is empty")
	}
	if u.User != nil {
		password, _ := u.User.Password()
		auth = &proxy.Auth{User: u.User.Username(), Password: password}
	}
	return u.Host, auth, nil
}

func (c *OpenRouterClient) SuggestCategory(ctx context.Context, req SuggestCategoryRequest) (*SuggestCategoryResponse, error) {
	if c.initErr != nil {
		return nil, c.initErr
	}
	if c.apiKey == "" || c.model == "" {
		return nil, fmt.Errorf("openrouter client is not configured")
	}
	system, user, err := BuildPrompt(req)
	if err != nil {
		return nil, err
	}
	orReq := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	}
	b, err := json.Marshal(orReq)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openrouter request failed (model=%s, base_url=%s): %w", c.model, c.baseURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openrouter read body failed (status=%d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode >= 300 {
		preview := string(body)
		if len(preview) > 300 {
			preview = preview[:300]
		}
		return nil, fmt.Errorf("openrouter bad status %d, body=%q", resp.StatusCode, preview)
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("openrouter response decode failed: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("empty llm response")
	}

	content := extractJSONContent(parsed.Choices[0].Message.Content)
	var out struct {
		CategoryID  string  `json:"category_id"`
		Probability float64 `json:"probability"`
		Reason      string  `json:"reason"`
	}
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("openrouter content json decode failed: %w", err)
	}
	if out.Probability < 0 || out.Probability > 1 {
		return nil, fmt.Errorf("invalid probability")
	}
	allowed := map[string]struct{}{}
	for _, c := range req.Categories {
		allowed[c.ID] = struct{}{}
	}
	if _, ok := allowed[out.CategoryID]; !ok {
		return nil, fmt.Errorf("category is out of allowed list")
	}
	return &SuggestCategoryResponse{CategoryID: out.CategoryID, Probability: out.Probability, Reason: out.Reason}, nil
}

// extractJSONContent normalizes LLM output and tries to isolate JSON object payload.
func extractJSONContent(content string) string {
	s := strings.TrimSpace(content)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if strings.HasSuffix(s, "```") {
			s = strings.TrimSuffix(s, "```")
			s = strings.TrimSpace(s)
		}
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end >= start {
		return strings.TrimSpace(s[start : end+1])
	}
	return s
}

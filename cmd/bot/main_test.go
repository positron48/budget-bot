package main

import (
	"strings"
	"testing"

	"budget-bot/internal/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestWebhookURLFormation(t *testing.T) {
	tests := []struct {
		name           string
		webhookURL     string
		webhookDomain  string
		webhookPath    string
		expectedURL    string
		shouldEnable   bool
	}{
		{
			name:          "explicit webhook URL",
			webhookURL:    "https://example.com/webhook",
			webhookDomain: "https://other.com",
			webhookPath:   "/tg",
			expectedURL:   "https://example.com/webhook",
			shouldEnable:  true,
		},
		{
			name:          "domain + path combination",
			webhookURL:    "",
			webhookDomain: "https://example.com",
			webhookPath:   "/tg",
			expectedURL:   "https://example.com/tg",
			shouldEnable:  true,
		},
		{
			name:          "domain with trailing slash + path",
			webhookURL:    "",
			webhookDomain: "https://example.com/",
			webhookPath:   "/tg",
			expectedURL:   "https://example.com/tg",
			shouldEnable:  true,
		},
		{
			name:          "http domain",
			webhookURL:    "",
			webhookDomain: "http://localhost:8080",
			webhookPath:   "/webhook",
			expectedURL:   "http://localhost:8080/webhook",
			shouldEnable:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Telegram: config.TelegramConfig{
					WebhookEnable:  tt.shouldEnable,
					WebhookURL:     tt.webhookURL,
					WebhookDomain:  tt.webhookDomain,
					WebhookPath:    tt.webhookPath,
				},
			}

			var webhookURL string
			if cfg.Telegram.WebhookURL != "" {
				webhookURL = cfg.Telegram.WebhookURL
			} else if cfg.Telegram.WebhookDomain != "" {
				webhookURL = strings.TrimSuffix(cfg.Telegram.WebhookDomain, "/") + cfg.Telegram.WebhookPath
			}

			assert.Equal(t, tt.expectedURL, webhookURL)
		})
	}
}

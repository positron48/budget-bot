package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

type failingTelegramTransport struct{}

func (failingTelegramTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, context.DeadlineExceeded
}
func TestTelegramErrorRedactsTokenAndPreservesCause(t *testing.T) {
	token := "12345:SYNTHETIC_test_token"
	client := telegramHTTPClient{client: &http.Client{Transport: failingTelegramTransport{}}, token: token}
	req, err := http.NewRequest(http.MethodPost, "https://api.telegram.org/bot"+token+"/getUpdates", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(req)
	if err == nil || strings.Contains(err.Error(), token) || !strings.Contains(err.Error(), "[redacted]") || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected sanitized error: %v", err)
	}
}
func TestTelegramTimeoutExceedsLongPoll(t *testing.T) {
	for _, seconds := range []int{0, 30, 60} {
		timeout := telegramClientTimeout(seconds)
		tr, err := buildTelegramSocks5Transport("127.0.0.1:1080", zap.NewNop(), seconds)
		if err != nil {
			t.Fatal(err)
		}
		if timeout <= time.Duration(seconds)*time.Second || tr.ResponseHeaderTimeout <= time.Duration(seconds)*time.Second || tr.ResponseHeaderTimeout >= timeout {
			t.Fatalf("poll=%d client=%v headers=%v", seconds, timeout, tr.ResponseHeaderTimeout)
		}
	}
}

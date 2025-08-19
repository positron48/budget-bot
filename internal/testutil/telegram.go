package testutil

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// NewTestBot spins up a tiny Telegram API emulator and returns a BotAPI wired to it.
func NewTestBot(t testing.TB) *tgbotapi.BotAPI {
    t.Helper()
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        p := r.URL.Path
        switch {
        case strings.HasSuffix(p, "/getMe"):
            _, _ = w.Write([]byte(`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test","username":"testbot"}}`))
        case strings.HasSuffix(p, "/sendMessage"):
            _, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
        case strings.HasSuffix(p, "/answerCallbackQuery"):
            _, _ = w.Write([]byte(`{"ok":true,"result":true}`))
        default:
            _, _ = w.Write([]byte(`{"ok":true,"result":true}`))
        }
    }))
    t.Cleanup(ts.Close)
    endpoint := ts.URL + "/bot%s/%s"
    bot, err := tgbotapi.NewBotAPIWithAPIEndpoint("TEST:TOKEN", endpoint)
    if err != nil { t.Fatalf("new bot: %v", err) }
    return bot
}



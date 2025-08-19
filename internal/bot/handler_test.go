package bot

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"budget-bot/internal/bot/ui"
	grpcclient "budget-bot/internal/grpc"
	"budget-bot/internal/repository"
	_ "github.com/mattn/go-sqlite3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// fakeAuthClient implements AuthClient for handler tests
type fakeAuthClient struct{}

func (f *fakeAuthClient) Register(ctx context.Context, email, password, name string) (string, string, string, string, time.Time, time.Time, error) {
	return "user-1", "tenant-1", "acc", "ref", time.Now().Add(time.Hour), time.Now().Add(24*time.Hour), nil
}
func (f *fakeAuthClient) Login(ctx context.Context, email, password string) (string, string, string, string, time.Time, time.Time, error) {
	return "user-1", "tenant-1", "acc", "ref", time.Now().Add(time.Hour), time.Now().Add(24*time.Hour), nil
}
func (f *fakeAuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	return "acc2", "ref2", time.Now().Add(time.Hour), time.Now().Add(24*time.Hour), nil
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil { t.Fatalf("open sqlite: %v", err) }
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS user_sessions (telegram_id INTEGER PRIMARY KEY, user_id TEXT NOT NULL, tenant_id TEXT NOT NULL, access_token TEXT NOT NULL, refresh_token TEXT NOT NULL, access_token_expires_at TIMESTAMP NOT NULL, refresh_token_expires_at TIMESTAMP NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS category_mappings (id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, keyword TEXT NOT NULL, category_id TEXT NOT NULL, priority INTEGER DEFAULT 0, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, UNIQUE(tenant_id, keyword));`,
		`CREATE TABLE IF NOT EXISTS dialog_states (telegram_id INTEGER PRIMARY KEY, state TEXT NOT NULL, draft_id TEXT, context TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS user_preferences (telegram_id INTEGER PRIMARY KEY, language TEXT DEFAULT 'ru', default_currency TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS transaction_drafts (id TEXT PRIMARY KEY, telegram_id INTEGER NOT NULL, type TEXT, amount_minor INTEGER, currency TEXT, description TEXT, category_id TEXT, occurred_at TIMESTAMP, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil { t.Fatalf("migrate: %v", err) }
	}
	return db
}

func newTestBot(t *testing.T) *tgbotapi.BotAPI {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simplistic Telegram API mock: reply ok for any method
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		if strings.HasSuffix(path, "/sendMessage") {
			_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
			return
		}
		if strings.HasSuffix(path, "/answerCallbackQuery") {
			_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
	}))
	t.Cleanup(ts.Close)
	endpoint := ts.URL + "/bot%s/%s"
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint("TEST:TOKEN", endpoint)
	if err != nil { t.Fatalf("new bot: %v", err) }
	return bot
}

func TestHandler_Start_Login_TransactionFlow(t *testing.T) {
	log := zap.NewNop()
	db := newTestDB(t)
	defer db.Close()
	states := repository.NewSQLiteDialogStateRepository(db)
	sessions := repository.NewSQLiteSessionRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	drafts := repository.NewSQLiteDraftRepository(db)
	auth := NewAuthManager(&fakeAuthClient{}, sessions, log)
	bot := newTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs).
		WithDrafts(drafts).
		WithReportClient(&grpcclient.FakeReportClient{}).
		WithTransactionClient(&grpcclient.FakeTransactionClient{}).
		WithTenantClient(&grpcclient.FakeTenantClient{})
	// ensure formatter is present
	h.fmt = ui.NewMessageFormatter()

	ctx := context.Background()
	chatID := int64(1000)
	userID := int64(42)

	// /start
	upd := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: "/start"}}
	upd.Message.Entities = &[]tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}[0:1]
	h.HandleUpdate(ctx, upd)

	// /login -> email -> password
	upd2 := upd; upd2.UpdateID = 2; upd2.Message.Text = "/login"; upd2.Message.Entities = &[]tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}[0:1]
	h.HandleUpdate(ctx, upd2)
	updEmail := upd; updEmail.UpdateID = 3; updEmail.Message.Text = "user@example.com"; updEmail.Message.Entities = nil
	h.HandleUpdate(ctx, updEmail)
	updPass := upd; updPass.UpdateID = 4; updPass.Message.Text = "secret"; updPass.Message.Entities = nil
	h.HandleUpdate(ctx, updPass)

	// Now send a transaction text, should ask to choose category
	updTx := upd; updTx.UpdateID = 5; updTx.Message.Text = "100 кофе"; updTx.Message.Entities = nil
	h.HandleUpdate(ctx, updTx)

	// Choose category callback -> confirmation
	updCat := tgbotapi.Update{UpdateID: 6, CallbackQuery: &tgbotapi.CallbackQuery{ID: "cb1", From: &tgbotapi.User{ID: userID}, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}, Data: "cat:cat-food"}}
	h.HandleUpdate(ctx, updCat)

	// Confirm yes -> creates transaction
	updYes := tgbotapi.Update{UpdateID: 7, CallbackQuery: &tgbotapi.CallbackQuery{ID: "cb2", From: &tgbotapi.User{ID: userID}, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}, Data: "confirm:yes"}}
	h.HandleUpdate(ctx, updYes)
}

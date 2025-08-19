package bot

import (
	"context"
	"testing"
	"time"

	"budget-bot/internal/bot/ui"
	grpcclient "budget-bot/internal/grpc"
	"budget-bot/internal/repository"
	"budget-bot/internal/testutil"
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

func TestHandler_Start_Login_TransactionFlow(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	states := repository.NewSQLiteDialogStateRepository(db)
	sessions := repository.NewSQLiteSessionRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	drafts := repository.NewSQLiteDraftRepository(db)
	auth := NewAuthManager(&fakeAuthClient{}, sessions, log)
	bot := testutil.NewTestBot(t)

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
	upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
	h.HandleUpdate(ctx, upd)

	// /login -> email -> password
	upd2 := upd; upd2.UpdateID = 2; upd2.Message.Text = "/login"; upd2.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
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

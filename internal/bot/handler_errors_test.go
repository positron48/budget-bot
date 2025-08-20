package bot

import (
    "context"
    "errors"
    "testing"
    "time"

    "budget-bot/internal/bot/ui"
    "budget-bot/internal/domain"
    grpcclient "budget-bot/internal/grpc"
    pb "budget-bot/internal/pb/budget/v1"
    "budget-bot/internal/repository"
    "budget-bot/internal/testutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go.uber.org/zap"
)

// failing clients to simulate backend errors
type failReport struct{ grpcclient.ReportClient }
func (f *failReport) GetStats(ctx context.Context, tenantID string, from, to time.Time, accessToken string) (*domain.Stats, error) { return nil, errors.New("boom") }
func (f *failReport) TopCategories(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*domain.CategoryTotal, error) { return nil, errors.New("boom") }

type failTx struct{ grpcclient.TransactionClient }
func (f *failTx) ListRecent(ctx context.Context, tenantID string, limit int, accessToken string) ([]*pb.Transaction, error) { return nil, errors.New("boom") }
func (f *failTx) ListForExport(ctx context.Context, tenantID string, from, to time.Time, limit int, accessToken string) ([]*pb.Transaction, error) { return nil, errors.New("boom") }

func TestHandler_ErrorBranches_NoPanic(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    auth := NewAuthManager(&fakeAuthClient{}, sessions, log)
    bot := testutil.NewTestBot(t)

    h := NewHandler(bot, states, auth, mappings, nil, log).
        WithPreferences(prefs).
        WithDrafts(drafts).
        WithReportClient(&failReport{}).
        WithTransactionClient(&failTx{}).
        WithTenantClient(&grpcclient.FakeTenantClient{})
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(7000)
    userID := int64(11)
    // login
    updStart := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: "/login"}}
    updStart.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updStart)
    updEmail := updStart; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updStart; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)

    // /stats with failing report
    updStats := updStart; updStats.UpdateID = 4; updStats.Message.Text = "/stats"; updStats.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updStats)
    // /top_categories with failing report
    updTop := updStart; updTop.UpdateID = 5; updTop.Message.Text = "/top_categories"; updTop.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 14}}
    h.HandleUpdate(ctx, updTop)
    // /recent with failing tx
    updRecent := updStart; updRecent.UpdateID = 6; updRecent.Message.Text = "/recent"; updRecent.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 7}}
    h.HandleUpdate(ctx, updRecent)
    // /export with failing tx
    updExport := updStart; updExport.UpdateID = 7; updExport.Message.Text = "/export"; updExport.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 7}}
    h.HandleUpdate(ctx, updExport)
}



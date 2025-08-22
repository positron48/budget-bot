package bot

import (
    "context"
    "errors"
    "testing"
    "time"

    "budget-bot/internal/bot/ui"
    "budget-bot/internal/domain"
    grpcclient "budget-bot/internal/grpc"
    "budget-bot/internal/repository"
    "budget-bot/internal/testutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go.uber.org/zap"
)

type fakeCatClient struct{
    list []*domain.Category
    listErr error
}
func (f *fakeCatClient) ListCategories(_ context.Context, _ string, _ string, _ domain.TransactionType, _ ...string) ([]*domain.Category, error) {
    if f.listErr != nil { return nil, f.listErr }
    return f.list, nil
}
func (f *fakeCatClient) CreateCategory(_ context.Context, _ string, _ string, name string, _ string) (*domain.Category, error) { return &domain.Category{ID:"id1", Name:name}, nil }
func (f *fakeCatClient) UpdateCategoryName(_ context.Context, _ string, id string, name string, _ string) (*domain.Category, error) { return &domain.Category{ID:id, Name:name}, nil }
func (f *fakeCatClient) DeleteCategory(_ context.Context, _ string, _ string) error { return nil }

func TestHandler_Categories_Success_And_Error(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)

    okClient := &fakeCatClient{list: []*domain.Category{{ID:"c1", Name:"Food"}}}
    errClient := &fakeCatClient{listErr: errors.New("boom")}

    h := NewHandler(bot, states, auth, mappings, okClient, log).
        WithPreferences(prefs).
        WithDrafts(drafts).
        WithReportClient(&grpcclient.FakeReportClient{}).
        WithTransactionClient(&grpcclient.FakeTransactionClient{}).
        WithTenantClient(&grpcclient.FakeTenantClient{})
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8000)
    userID := int64(44)

    // login
    updLogin := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}
    updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updLogin; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)

    // success
    upd := updLogin; upd.UpdateID = 4; upd.Message.Text = "/categories"; upd.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:11}}
    h.HandleUpdate(ctx, upd)

    // error client
    h.WithCategoryClient(errClient)
    upd2 := updLogin; upd2.UpdateID = 5; upd2.Message.Text = "/categories"; upd2.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:11}}
    h.HandleUpdate(ctx, upd2)
    _ = time.Second
}



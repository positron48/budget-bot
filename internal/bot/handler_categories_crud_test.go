package bot

import (
    "context"
    "errors"
    "testing"
    "time"

    "budget-bot/internal/bot/ui"
    "budget-bot/internal/domain"
    "budget-bot/internal/repository"
    "budget-bot/internal/testutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go.uber.org/zap"
)

type recCatClient struct{
    lastLocale string
    shouldErr  bool
}

func (r *recCatClient) ListCategories(_ context.Context, _ string, _ string, _ domain.TransactionType, locale ...string) ([]*domain.Category, error) {
    if len(locale) > 0 { r.lastLocale = locale[0] }
    return []*domain.Category{{ID:"c1", Name:"Food"}}, nil
}
func (r *recCatClient) CreateCategory(_ context.Context, _ string, _ string, name string, _ string) (*domain.Category, error) {
    if r.shouldErr { return nil, errors.New("boom") }
    return &domain.Category{ID:"c2", Name:name}, nil
}
func (r *recCatClient) UpdateCategoryName(_ context.Context, _ string, id string, name string, _ string) (*domain.Category, error) {
    if r.shouldErr { return nil, errors.New("boom") }
    return &domain.Category{ID:id, Name:name}, nil
}
func (r *recCatClient) DeleteCategory(_ context.Context, _ string, _ string) error {
    if r.shouldErr { return errors.New("boom") }
    return nil
}

func TestHandler_Create_Rename_Delete_Category(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)

    cat := &recCatClient{}
    h := NewHandler(bot, states, auth, mappings, cat, log).
        WithPreferences(prefs).
        WithDrafts(drafts)
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8200)
    userID := int64(33)

    // login with OAuth
    updLogin := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}
    updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "user@example.com"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updCode := updLogin; updCode.UpdateID = 3; updCode.Message.Text = "123456"; updCode.Message.Entities = nil
    h.HandleUpdate(ctx, updCode)

    // /create_category code name
    updCreate := updLogin; updCreate.UpdateID = 4; updCreate.Message.Text = "/create_category food Питание"; updCreate.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updCreate)

    // /rename_category id newname
    updRename := updLogin; updRename.UpdateID = 5; updRename.Message.Text = "/rename_category c2 Еда"; updRename.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updRename)

    // /delete_category id
    updDelete := updLogin; updDelete.UpdateID = 6; updDelete.Message.Text = "/delete_category c2"; updDelete.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updDelete)

    // error paths
    cat.shouldErr = true
    updCreate2 := updLogin; updCreate2.UpdateID = 7; updCreate2.Message.Text = "/create_category code Name"; updCreate2.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updCreate2)
    updRename2 := updLogin; updRename2.UpdateID = 8; updRename2.Message.Text = "/rename_category id Name"; updRename2.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updRename2)
    updDelete2 := updLogin; updDelete2.UpdateID = 9; updDelete2.Message.Text = "/delete_category id"; updDelete2.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updDelete2)
    _ = time.Second
}

func TestHandler_Categories_UsesLocaleFromPreferences(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefsRepo := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)

    cat := &recCatClient{}
    h := NewHandler(bot, states, auth, mappings, cat, log).
        WithPreferences(prefsRepo).
        WithDrafts(drafts)
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8300)
    userID := int64(34)

    // login with OAuth
    updLogin := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}
    updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "user@example.com"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updCode := updLogin; updCode.UpdateID = 3; updCode.Message.Text = "123456"; updCode.Message.Entities = nil
    h.HandleUpdate(ctx, updCode)

    // set preferences language to en
    if err := prefsRepo.SavePreferences(ctx, &repository.UserPreferences{TelegramID:userID, Language:"en", DefaultCurrency:"USD"}); err != nil { t.Fatalf("save prefs: %v", err) }

    // call /categories
    upd := updLogin; upd.UpdateID = 4; upd.Message.Text = "/categories"; upd.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:11}}
    h.HandleUpdate(ctx, upd)

    if cat.lastLocale != "en" { t.Fatalf("expected locale en, got %s", cat.lastLocale) }
}



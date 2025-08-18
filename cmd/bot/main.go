package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	botpkg "budget-bot/internal/bot"
	"budget-bot/internal/pkg/config"
	"budget-bot/internal/pkg/db"
	botlogger "budget-bot/internal/pkg/logger"

	"budget-bot/internal/repository"
	"budget-bot/internal/metrics"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"net/url"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("config error: %w", err))
	}

	// Logger
	log, err := botlogger.New(cfg.Logging.Level)
	if err != nil {
		panic(fmt.Errorf("logger init error: %w", err))
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting bot")

	// Telegram bot init with optional BaseURL for emulator
	var bot *tgbotapi.BotAPI
	if cfg.Telegram.APIBaseURL != "" {
		endpoint := normalizeAPIEndpoint(cfg.Telegram.APIBaseURL)
		bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(cfg.Telegram.Token, endpoint)
	} else {
		bot, err = tgbotapi.NewBotAPI(cfg.Telegram.Token)
	}
	if err != nil {
		log.Fatal("failed to init bot", zap.Error(err))
	}
	bot.Debug = cfg.Telegram.Debug

	log.Info("authorized on account", zap.String("username", bot.Self.UserName))

	// Ensure data dir exists and DB is migrated
	if err := os.MkdirAll("./data", 0o755); err != nil {
		log.Fatal("failed to create data dir", zap.Error(err))
	}
	dbConn, err := db.OpenAndMigrate(cfg.Database.DSN, "./migrations", log)
	if err != nil {
		log.Fatal("database init failed", zap.Error(err))
	}

	// Health endpoint and metrics (optional)
	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("OK")) })
		if cfg.Metrics.Enabled {
			http.Handle("/metrics", metrics.Handler())
		}
		_ = http.ListenAndServe(":8088", nil)
	}()

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Updates config
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Telegram.UpdatesTimeout
	updates := bot.GetUpdatesChan(u)

	// Handler wiring
	stateRepo := repository.NewSQLiteDialogStateRepository(dbConn)
	sessionRepo := repository.NewSQLiteSessionRepository(dbConn)
	mappingRepo := repository.NewSQLiteCategoryMappingRepository(dbConn)
	prefsRepo := repository.NewSQLitePreferencesRepository(dbConn)
	fakeAuth := &fakeAuthClient{}
	authManager := botpkg.NewAuthManager(fakeAuth, sessionRepo, log)
	h := botpkg.NewHandler(bot, stateRepo, authManager, mappingRepo, nil, log).WithPreferences(prefsRepo)

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down")
			return
		case update := <-updates:
			h.HandleUpdate(ctx, update)
		}
	}

	// add tiny sleep to appease lints about empty select default (not used here)
	_ = time.Second

}

// normalizeAPIEndpoint ensures endpoint string is a valid format expected by tgbotapi: it must contain exactly two %s placeholders for token and method.
func normalizeAPIEndpoint(base string) string {
	s := strings.TrimSpace(base)
	// Fix encoded placeholders
	s = strings.ReplaceAll(s, "%25s", "%s")
	// If it already has exactly two placeholders, keep as-is
	if strings.Count(s, "%s") == 2 {
		return s
	}
	// If the placeholder count is wrong or missing, rebuild using parsed URL
	if u, err := url.Parse(s); err == nil && u.Scheme != "" && u.Host != "" {
		path := strings.TrimSuffix(u.Path, "/")
		return u.Scheme + "://" + u.Host + path + "/bot%s/%s"
	}
	// Fallback: just append the correct suffix
	if strings.HasSuffix(s, "/") {
		return s + "bot%s/%s"
	}
	return s + "/bot%s/%s"
}

// fakeAuthClient implements bot.AuthClient for local testing before real gRPC client is wired.
type fakeAuthClient struct{}

func (f *fakeAuthClient) Register(ctx context.Context, email, password, name string) (string, string, string, string, time.Time, time.Time, error) {
	return "user-123", "tenant-123", "access-token", "refresh-token", time.Now().Add(1*time.Hour), time.Now().Add(24*time.Hour), nil
}

func (f *fakeAuthClient) Login(ctx context.Context, email, password string) (string, string, string, string, time.Time, time.Time, error) {
	return "user-123", "tenant-123", "access-token", "refresh-token", time.Now().Add(1*time.Hour), time.Now().Add(24*time.Hour), nil
}

func (f *fakeAuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	return "access-token2", "refresh-token2", time.Now().Add(1*time.Hour), time.Now().Add(24*time.Hour), nil
}



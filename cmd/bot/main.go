package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"budget-bot/internal/pkg/config"
	"budget-bot/internal/pkg/db"
	botlogger "budget-bot/internal/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
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
		bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(cfg.Telegram.Token, cfg.Telegram.APIBaseURL)
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
	if _, err := db.OpenAndMigrate(cfg.Database.DSN, "./migrations", log); err != nil {
		log.Fatal("database init failed", zap.Error(err))
	}

	// Health endpoint (optional)
	go func() {
		_ = http.ListenAndServe(":8088", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("OK"))
		}))
	}()

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Updates config
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Telegram.UpdatesTimeout
	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down")
			return
		case update := <-updates:
			if update.Message == nil { // ignore non-message updates for now
				continue
			}

			if update.Message.IsCommand() {
				response := tgbotapi.NewMessage(update.Message.Chat.ID, "Command handling will be implemented soon.")
				if _, err := bot.Send(response); err != nil {
					log.Warn("failed to send cmd reply", zap.Error(err))
				}
				continue
			}

			// Echo placeholder to verify loop works
			text := fmt.Sprintf("echo: %s", update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			if _, err := bot.Send(msg); err != nil {
				log.Warn("failed to send reply", zap.Error(err))
			}
		}
	}

	// add tiny sleep to appease lints about empty select default (not used here)
	_ = time.Second
}



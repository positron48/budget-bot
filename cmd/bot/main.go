// Command budget-bot runs the Telegram bot.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	botpkg "budget-bot/internal/bot"
	"budget-bot/internal/llm"
	"budget-bot/internal/pkg/config"
	"budget-bot/internal/pkg/db"
	botlogger "budget-bot/internal/pkg/logger"

	grpcwire "budget-bot/internal/grpc"
	"budget-bot/internal/metrics"
	"budget-bot/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
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
	if cfg.Telegram.Socks5Proxy != "" {
		tr, err := buildTelegramSocks5Transport(cfg.Telegram.Socks5Proxy)
		if err != nil {
			log.Fatal("failed to build socks5 transport", zap.Error(err))
		}
		bot.Client = &http.Client{Transport: tr}
		log.Info("telegram socks5 proxy enabled", zap.String("socks5_proxy", cfg.Telegram.Socks5Proxy))
	}

	log.Info("authorized on account", zap.String("username", bot.Self.UserName))

	// Ensure data dir exists and DB is migrated
	if err := os.MkdirAll("./data", 0o755); err != nil {
		log.Fatal("failed to create data dir", zap.Error(err))
	}
	dbConn, err := db.OpenAndMigrate(cfg.Database.DSN, "./migrations", log)
	if err != nil {
		log.Fatal("database init failed", zap.Error(err))
	}

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Handler wiring
	stateRepo := repository.NewSQLiteDialogStateRepository(dbConn)
	sessionRepo := repository.NewSQLiteSessionRepository(dbConn)
	mappingRepo := repository.NewSQLiteCategoryMappingRepository(dbConn)
	prefsRepo := repository.NewSQLitePreferencesRepository(dbConn)
	draftRepo := repository.NewSQLiteDraftRepository(dbConn)
	opCtxRepo := repository.NewSQLiteOperationContextRepository(dbConn)

	// Wire OAuth clients
	catClient, reportClient, tenantClient, txClient, oauthClient, authClient := grpcwire.WireClients(log)

	// Create OAuth manager
	oauthManager := botpkg.NewOAuthManagerWithAuthClient(oauthClient, authClient, sessionRepo, log, cfg.OAuth.WebBaseURL)

	h := botpkg.NewHandler(bot, stateRepo, oauthManager, mappingRepo, catClient, log).
		WithPreferences(prefsRepo).
		WithDrafts(draftRepo).
		WithOperationContexts(opCtxRepo).
		WithCategoryClient(catClient).
		WithReportClient(reportClient).
		WithTransactionClient(txClient).
		WithTenantClient(tenantClient)
	if cfg.OpenRouter.Enable {
		if cfg.OpenRouter.APIKey == "" || cfg.OpenRouter.Model == "" {
			log.Warn("openrouter enabled but API key/model is not configured; llm fallback disabled")
		} else {
			h.WithLLM(llm.NewOpenRouterClient(cfg.OpenRouter.BaseURL, cfg.OpenRouter.APIKey, cfg.OpenRouter.Model, cfg.OpenRouter.Timeout), true)
		}
	}

	// Webhook mode vs long polling
	if cfg.Telegram.WebhookEnable {
		// Determine webhook URL
		var webhookURL string
		if cfg.Telegram.WebhookURL != "" {
			// Use explicit webhook URL if provided
			webhookURL = cfg.Telegram.WebhookURL
		} else if cfg.Telegram.WebhookDomain != "" {
			// Build webhook URL from domain and path
			webhookURL = strings.TrimSuffix(cfg.Telegram.WebhookDomain, "/") + cfg.Telegram.WebhookPath
		} else {
			log.Fatal("webhook enabled but neither webhook_url nor webhook_domain is configured")
		}

		log.Info("setting webhook", zap.String("url", webhookURL))

		// Set webhook using the configured API base URL
		whCfg, _ := tgbotapi.NewWebhook(webhookURL)
		if _, err := bot.Request(whCfg); err != nil {
			log.Fatal("failed to set webhook", zap.Error(err))
		}

		log.Info("webhook set successfully")

		// Serve webhook on configured path
		http.HandleFunc(cfg.Telegram.WebhookPath, func(w http.ResponseWriter, r *http.Request) {
			update, err := bot.HandleUpdate(r)
			if err != nil {
				log.Warn("webhook handle error", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if update != nil {
				go h.HandleUpdate(context.Background(), *update)
			}
			w.WriteHeader(http.StatusOK)
		})
		// Health and metrics on same server
		http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("OK")) })
		if cfg.Metrics.Enabled {
			http.Handle("/metrics", metrics.Handler())
		}

		log.Info("starting HTTP server for webhook", zap.String("address", cfg.Server.Address))
		go func() { _ = http.ListenAndServe(cfg.Server.Address, nil) }()

		// Wait for shutdown signal
		<-ctx.Done()

		// Clean up webhook on shutdown using the configured API base URL
		log.Info("cleaning up webhook")
		if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
			log.Warn("failed to delete webhook", zap.Error(err))
		} else {
			log.Info("webhook deleted successfully")
		}

		log.Info("shutting down")
		return
	}

	// Health endpoint and metrics (optional) for long polling
	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("OK")) })
		if cfg.Metrics.Enabled {
			http.Handle("/metrics", metrics.Handler())
		}
		_ = http.ListenAndServe(cfg.Server.Address, nil)
	}()

	// Long polling loop
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.Telegram.UpdatesTimeout
	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down")
			return
		case update := <-updates:
			h.HandleUpdate(ctx, update)
		}
	}

	// end of main
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

func buildTelegramSocks5Transport(raw string) (*http.Transport, error) {
	hostPort, auth, err := normalizeSocks5Proxy(raw)
	if err != nil {
		return nil, err
	}

	// Some SOCKS providers are slow to establish TCP/TLS handshakes.
	// We keep dial timeout relatively small for the SOCKS handshake itself,
	// but give TLS handshake more time to avoid failing startup.
	baseDialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	dialer, err := proxy.SOCKS5("tcp", hostPort, auth, baseDialer)
	if err != nil {
		return nil, fmt.Errorf("socks5 dialer init: %w", err)
	}

	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("unexpected http.DefaultTransport type %T", http.DefaultTransport)
	}
	tr := baseTransport.Clone()
	// Important: do not use HTTP proxy env vars for Telegram; SOCKS5 transport is explicit.
	tr.Proxy = nil
	tr.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}
	// Allow slower SOCKS5/TLS handshake paths; default Go value is small enough
	// to cause CrashLoopBackOff during startup for some proxies.
	tr.TLSHandshakeTimeout = 30 * time.Second
	return tr, nil
}

func normalizeSocks5Proxy(raw string) (hostPort string, auth *proxy.Auth, err error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil, fmt.Errorf("empty socks5 proxy")
	}
	// Accept "ip:port" shorthand.
	if !strings.Contains(s, "://") {
		s = "socks5://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", nil, fmt.Errorf("parse socks5 proxy: %w", err)
	}
	if u.Host == "" {
		return "", nil, fmt.Errorf("parse socks5 proxy: missing host:port in %q", raw)
	}
	var a *proxy.Auth
	if u.User != nil {
		pass, _ := u.User.Password()
		a = &proxy.Auth{User: u.User.Username(), Password: pass}
	}
	return u.Host, a, nil
}

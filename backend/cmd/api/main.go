package main

import (
	"context"
	"os"

	"realtime-session-coordination/backend/internal/analytics"
	"realtime-session-coordination/backend/internal/api"
	"realtime-session-coordination/backend/internal/auth"
	"realtime-session-coordination/backend/internal/config"
	"realtime-session-coordination/backend/internal/logging"
	"realtime-session-coordination/backend/internal/mailer"
	"realtime-session-coordination/backend/internal/otp"
	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/sessionlog"
	"realtime-session-coordination/backend/internal/user"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// initStores creates the appropriate stores based on DB_DRIVER env var.
func initStores(cfg config.Config) (session.Store, programitem.Store, sessionlog.Store, user.Store, otp.Store, analytics.IngestionStore, analytics.ProcessorStore, error) {
	switch cfg.DBDriver {
	case "memory":
		sessionStore := session.NewMemoryStore()
		return sessionStore, programitem.NewMemoryStore(sessionStore.SessionExists), sessionlog.NewMemoryStore(), user.NewMemoryStore(), otp.NewMemoryStore(), nil, nil, nil
	case "sqlite":
		sessionStore, err := session.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		programItemStore, err := programitem.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		sessionLogStore, err := sessionlog.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		userStore, err := user.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		otpStore, err := otp.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		analyticsStore, err := analytics.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}

		return sessionStore, programItemStore, sessionLogStore, userStore, otpStore, analyticsStore, analyticsStore, nil
	default:
		return nil, nil, nil, nil, nil, nil, nil, nil
	}
}

func main() {
	bootstrapLogger := logging.Default()

	if err := godotenv.Load(); err != nil {
		bootstrapLogger.Warn("env_file_not_loaded", "error", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		bootstrapLogger.Error("invalid_configuration", "error", err)
		os.Exit(1)
	}

	logger, err := logging.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		bootstrapLogger.Error("invalid_logging_configuration", "error", err)
		os.Exit(1)
	}
	appLogger := logger.With("component", "api_server")

	otpMailer, err := mailer.New(cfg.MailerMode, logger)
	if err != nil {
		appLogger.Error("mailer_initialization_failed", "error", err)
		os.Exit(1)
	}
	appLogger.Info("mailer_initialized", "mode", cfg.MailerMode)

	store, programItemStore, sessionLogStore, userStore, otpStore, analyticsIngestionStore, analyticsProcessorStore, err := initStores(cfg)
	if err != nil {
		appLogger.Error("store_initialization_failed", "error", err)
		os.Exit(1)
	}

	authService, err := auth.NewService(userStore, cfg.JWTSecret, cfg.JWTExpiry, cfg.JWTIssuer)
	if err != nil {
		appLogger.Error("auth_service_initialization_failed", "error", err)
		os.Exit(1)
	}

	otpService := otp.NewService(otpStore, userStore, authService, otpMailer, logger, otp.ServiceConfig{
		ExpiryMinutes:  cfg.OTPExpiryMinutes,
		MaxAttempts:    cfg.OTPMaxAttempts,
		ResendCooldown: cfg.OTPResendCooldown,
	})

	manager := session.NewManager(store)
	programItemManager := programitem.NewManager(programItemStore)
	sessionLogManager := sessionlog.NewManager(sessionLogStore)
	analyticsManager := analytics.NewManager()
	analyticsEmitter := analytics.NewEmitter(analyticsIngestionStore)
	if analyticsProcessorStore != nil {
		processorCfg := analytics.ProcessorConfig{
			CleanupInterval:          cfg.AnalyticsCleanupInterval,
			ProcessedOutboxRetention: cfg.AnalyticsProcessedOutboxRetention,
			DeadLetterRetention:      cfg.AnalyticsDeadLetterRetention,
			EventRetention:           cfg.AnalyticsEventRetention,
		}
		if projectionStore, ok := analyticsProcessorStore.(analytics.ProjectionStore); ok {
			processorCfg.ProjectionBuilder = analytics.NewProjectionBuilder(projectionStore, analyticsManager)
			processorCfg.GetSessionSnapshot = manager.GetSnapshot
			processorCfg.ListProgramItemSnapshots = programItemManager.ListSnapshots
			processorCfg.ListSessionSnapshots = manager.ListSnapshots
		}

		processor := analytics.NewProcessor(analyticsProcessorStore, logger, processorCfg)
		go processor.Start(context.Background())
	}
	hub := ws.NewHub(logger)
	handler := api.NewHandler(manager, programItemManager, sessionLogManager, analyticsManager, analyticsEmitter, analyticsProcessorStore, hub, authService, otpService, logger)

	router := gin.New()
	router.Use(gin.Recovery(), api.CORSMiddleware(), api.RequestLoggingMiddleware(logger))
	handler.RegisterRoutes(router)

	appLogger.Info("backend_starting", "port", cfg.Port, "db_driver", cfg.DBDriver)
	if err := router.Run(":" + cfg.Port); err != nil {
		appLogger.Error("server_failed", "error", err)
		os.Exit(1)
	}
}

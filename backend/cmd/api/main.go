package main

import (
	"os"

	"realtime-session-coordination/backend/internal/analytics"
	"realtime-session-coordination/backend/internal/api"
	"realtime-session-coordination/backend/internal/auth"
	"realtime-session-coordination/backend/internal/config"
	"realtime-session-coordination/backend/internal/logging"
	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/sessionlog"
	"realtime-session-coordination/backend/internal/user"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// initStores creates the appropriate stores based on DB_DRIVER env var.
func initStores(cfg config.Config) (session.Store, programitem.Store, sessionlog.Store, user.Store, analytics.IngestionStore, error) {
	switch cfg.DBDriver {
	case "memory":
		sessionStore := session.NewMemoryStore()
		return sessionStore, programitem.NewMemoryStore(sessionStore.SessionExists), sessionlog.NewMemoryStore(), user.NewMemoryStore(), nil, nil
	case "sqlite":
		sessionStore, err := session.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		programItemStore, err := programitem.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		sessionLogStore, err := sessionlog.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		userStore, err := user.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		analyticsStore, err := analytics.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		return sessionStore, programItemStore, sessionLogStore, userStore, analyticsStore, nil
	default:
		return nil, nil, nil, nil, nil, nil
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

	store, programItemStore, sessionLogStore, userStore, analyticsIngestionStore, err := initStores(cfg)
	if err != nil {
		appLogger.Error("store_initialization_failed", "error", err)
		os.Exit(1)
	}

	authService, err := auth.NewService(userStore, cfg.JWTSecret, cfg.JWTExpiry, cfg.JWTIssuer)
	if err != nil {
		appLogger.Error("auth_service_initialization_failed", "error", err)
		os.Exit(1)
	}

	manager := session.NewManager(store)
	programItemManager := programitem.NewManager(programItemStore)
	sessionLogManager := sessionlog.NewManager(sessionLogStore)
	analyticsManager := analytics.NewManager()
	analyticsEmitter := analytics.NewEmitter(analyticsIngestionStore)
	hub := ws.NewHub(logger)
	handler := api.NewHandler(manager, programItemManager, sessionLogManager, analyticsManager, analyticsEmitter, hub, authService, logger)

	router := gin.New()
	router.Use(gin.Recovery(), api.CORSMiddleware(), api.RequestLoggingMiddleware(logger))
	handler.RegisterRoutes(router)

	appLogger.Info("backend_starting", "port", cfg.Port, "db_driver", cfg.DBDriver)
	if err := router.Run(":" + cfg.Port); err != nil {
		appLogger.Error("server_failed", "error", err)
		os.Exit(1)
	}
}

package main

import (
	"os"

	"realtime-session-coordination/backend/internal/api"
	"realtime-session-coordination/backend/internal/auth"
	"realtime-session-coordination/backend/internal/config"
	"realtime-session-coordination/backend/internal/logging"
	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/user"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// initStores creates the appropriate session and user stores based on DB_DRIVER env var.
func initStores(cfg config.Config) (session.Store, user.Store, error) {
	switch cfg.DBDriver {
	case "memory":
		return session.NewMemoryStore(), user.NewMemoryStore(), nil
	case "sqlite":
		sessionStore, err := session.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, err
		}

		userStore, err := user.NewSqliteStore(cfg.SqliteDBPath)
		if err != nil {
			return nil, nil, err
		}

		return sessionStore, userStore, nil
	default:
		return nil, nil, nil
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

	store, userStore, err := initStores(cfg)
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
	hub := ws.NewHub(logger)
	handler := api.NewHandler(manager, hub, authService, logger)

	router := gin.New()
	router.Use(gin.Recovery(), api.CORSMiddleware(), api.RequestLoggingMiddleware(logger))
	handler.RegisterRoutes(router)

	appLogger.Info("backend_starting", "port", cfg.Port, "db_driver", cfg.DBDriver)
	if err := router.Run(":" + cfg.Port); err != nil {
		appLogger.Error("server_failed", "error", err)
		os.Exit(1)
	}
}

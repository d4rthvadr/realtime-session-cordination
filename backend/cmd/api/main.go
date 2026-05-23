package main

import (
	"log"

	"realtime-session-coordination/backend/internal/api"
	"realtime-session-coordination/backend/internal/auth"
	"realtime-session-coordination/backend/internal/config"
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
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	store, userStore, err := initStores(cfg)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	authService, err := auth.NewService(userStore, cfg.JWTSecret, cfg.JWTExpiry, cfg.JWTIssuer)
	if err != nil {
		log.Fatalf("failed to initialize auth service: %v", err)
	}

	manager := session.NewManager(store)
	hub := ws.NewHub()
	handler := api.NewHandler(manager, hub, authService)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), api.CORSMiddleware())
	handler.RegisterRoutes(router)

	log.Printf("backend listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

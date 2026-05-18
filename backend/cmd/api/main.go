package main

import (
	"log"
	"os"

	"realtime-session-coordination/backend/internal/api"
	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// initStore creates the appropriate Store implementation based on DB_DRIVER env var
func initStore() (session.Store, error) {
	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	switch dbDriver {
	case "memory":
		return session.NewMemoryStore(), nil
	case "sqlite":
		dbPath := os.Getenv("SQLITE_DB_PATH")
		if dbPath == "" {
			dbPath = "./sessions.db"
		}
		return session.NewSqliteStore(dbPath)
	default:
		log.Fatalf("unknown DB_DRIVER: %s (must be 'memory' or 'sqlite')", dbDriver)
		return nil, nil
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store, err := initStore()
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	manager := session.NewManager(store)
	hub := ws.NewHub()
	handler := api.NewHandler(manager, hub)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), api.CORSMiddleware())
	handler.RegisterRoutes(router)

	log.Printf("backend listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

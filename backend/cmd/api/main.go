package main

import (
	"log"
	"os"

	"realtime-session-coordination/backend/internal/api"
	"realtime-session-coordination/backend/internal/session"
	"realtime-session-coordination/backend/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	manager := session.NewManager(session.NewMemoryStore())
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

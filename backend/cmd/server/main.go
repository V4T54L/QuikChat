package main

import (
	"fmt"
	"log"
	"net/http"

	"chat-app/internal/adapter/postgres"
	"chat-app/internal/service"
	"chat-app/pkg/config"

	delivery "chat-app/internal/delivery/http"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dbPool := postgres.NewDB(cfg.DatabaseURL)
	defer dbPool.Close()

	// Initialize repositories
	userRepo := postgres.NewPostgresUserRepository(dbPool)
	sessionRepo := postgres.NewPostgresSessionRepository(dbPool)

	// Initialize use cases/services
	authUsecase := service.NewAuthService(userRepo, sessionRepo, cfg)

	// Initialize router
	router := delivery.NewRouter(authUsecase)

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}


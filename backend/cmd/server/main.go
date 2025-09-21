package main

import (
	"fmt"
	"log"
	nethttp "net/http"

	"chat-app/internal/adapter/localfile"
	"chat-app/internal/adapter/postgres"
	delivery "chat-app/internal/delivery/http"
	"chat-app/internal/service"
	"chat-app/pkg/config"
)

func main() {
	log.Println("Starting QuikChat server...")

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dbPool, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established.")

	// Initialize repositories
	userRepo := postgres.NewPostgresUserRepository(dbPool)
	sessionRepo := postgres.NewPostgresSessionRepository(dbPool)
	fileRepo, err := localfile.NewLocalFileRepository(cfg.UploadDir)
	if err != nil {
		log.Fatalf("could not create file repository: %v", err)
	}

	// Initialize use cases/services
	authUsecase := service.NewAuthService(userRepo, sessionRepo, cfg)
	userUsecase := service.NewUserService(userRepo, fileRepo)

	// Initialize router
	router := delivery.NewRouter(cfg, authUsecase, userUsecase)

	// Start server
	server := &nethttp.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Printf("Server listening on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
		log.Fatalf("could not listen on %s: %v\n", cfg.Port, err)
	}
}


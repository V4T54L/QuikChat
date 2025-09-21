package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"chat-app/internal/adapter/localfile"
	"chat-app/internal/adapter/postgres"
	"chat-app/internal/adapter/redis"
	"chat-app/internal/delivery/http/router"
	"chat-app/internal/delivery/websocket"
	"chat-app/internal/service"
	"chat-app/pkg/config"

	postgresRepo "chat-app/internal/adapter/postgres"
	redisRepo "chat-app/internal/adapter/redis"
)

func main() {
	log.Println("Starting QuikChat server...")

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db := postgres.NewDB(cfg.DatabaseURL)
	defer db.Close()
	log.Println("Database connection established.")

	// Initialize Redis client
	redisClient := redis.NewClient(cfg.RedisURL)

	// Initialize repositories
	userRepo := postgresRepo.NewPostgresUserRepository(db)
	sessionRepo := postgresRepo.NewPostgresSessionRepository(db)
	fileRepo, err := localfile.NewLocalFileRepository(cfg.UploadDir)
	if err != nil {
		log.Fatalf("failed to create file repository: %v", err)
	}
	redisEventRepo := redisRepo.NewRedisEventRepository(redisClient)
	pgEventRepo := postgresRepo.NewPostgresEventRepository(db)
	friendRepo := postgresRepo.NewPostgresFriendRepository(db)

	// Initialize use cases/services
	authUsecase := service.NewAuthService(userRepo, sessionRepo, cfg)
	userUsecase := service.NewUserService(userRepo, fileRepo)
	eventUsecase := service.NewEventService(redisEventRepo, pgEventRepo, userRepo)
	friendUsecase := service.NewFriendService(friendRepo, userRepo, eventUsecase)

	// Initialize WebSocket Hub
	hub := websocket.NewHub(eventUsecase)
	go hub.Run()

	// Start background worker for event persistence
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Run every minute
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Running background worker to persist events...")
				if err := eventUsecase.PersistBufferedEvents(context.Background()); err != nil {
					log.Printf("Error persisting buffered events: %v", err)
				}
			}
		}
	}()

	// Initialize router
	r := router.NewRouter(cfg, authUsecase, userUsecase, friendUsecase, hub)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}


package main

import (
	"context"
	"log"
	"time"

	"chat-app/internal/adapter/localfile"
	"chat-app/internal/adapter/postgres"
	"chat-app/internal/adapter/redis"
	"chat-app/internal/delivery/http" // Changed from router
	"chat-app/internal/delivery/websocket"
	"chat-app/internal/service"
	"chat-app/pkg/config"

	nethttp "net/http" // Alias for standard http package
)

func main() {
	log.Println("Starting QuikChat server...")

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dbPool := postgres.NewDB(cfg.DatabaseURL) // Changed variable name to dbPool
	defer dbPool.Close()
	log.Println("Database connection established.")

	// Initialize Redis client
	redisClient := redis.NewClient(cfg.RedisURL)

	// Initialize repositories
	userRepo := postgres.NewPostgresUserRepository(dbPool)
	sessionRepo := postgres.NewPostgresSessionRepository(dbPool)
	fileRepo, err := localfile.NewLocalFileRepository(cfg.UploadDir)
	if err != nil {
		log.Fatalf("failed to create file repository: %v", err)
	}
	redisEventRepo := redis.NewRedisEventRepository(redisClient)
	pgEventRepo := postgres.NewPostgresEventRepository(dbPool)
	friendRepo := postgres.NewPostgresFriendRepository(dbPool)
	groupRepo := postgres.NewPostgresGroupRepository(dbPool) // Added groupRepo

	// Initialize use cases/services
	authUsecase := service.NewAuthService(userRepo, sessionRepo, cfg)
	userUsecase := service.NewUserService(userRepo, fileRepo)
	eventUsecase := service.NewEventService(redisEventRepo, pgEventRepo, userRepo)
	friendUsecase := service.NewFriendService(friendRepo, userRepo, eventUsecase)
	groupUsecase := service.NewGroupService(groupRepo, userRepo, friendRepo, fileRepo, eventUsecase) // Added groupUsecase

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
	router := http.NewRouter(cfg, authUsecase, userUsecase, friendUsecase, groupUsecase, hub) // Updated router package and added groupUsecase

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := nethttp.ListenAndServe(":"+cfg.Port, router); err != nil && err != nethttp.ErrServerClosed { // Used nethttp alias
		log.Fatalf("Failed to start server: %v", err)
	}
}


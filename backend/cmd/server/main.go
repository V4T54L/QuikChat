package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"chat-app/backend/internal/adapter/localfile"
	"chat-app/backend/internal/adapter/postgres"
	"chat-app/backend/internal/adapter/redis"
	"chat-app/backend/internal/delivery/websocket"
	"chat-app/backend/internal/repository" // Added from attempted
	"chat-app/backend/internal/service"
	"chat-app/backend/internal/usecase" // Added from attempted
	"chat-app/backend/pkg/config"

	httpRouter "chat-app/backend/internal/delivery/http" // Alias for custom http router
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dbPool, err := postgres.NewDB(cfg.DatabaseURL) // Changed to include error handling
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established.")

	// Initialize Redis client
	redisClient := redis.NewClient(cfg.RedisURL)

	// Initialize repositories
	userRepo := postgres.NewPostgresUserRepository(dbPool)
	sessionRepo := postgres.NewPostgresSessionRepository(dbPool)
	fileRepo, err := localfile.NewLocalFileRepository(cfg.UploadDir)
	if err != nil {
		log.Fatalf("Could not create file repository: %v", err) // Updated error message
	}
	redisEventRepo := redis.NewRedisEventRepository(redisClient)
	pgEventRepo := postgres.NewPostgresEventRepository(dbPool)
	friendRepo := postgres.NewPostgresFriendRepository(dbPool)
	groupRepo := postgres.NewPostgresGroupRepository(dbPool)
	messageRepo := postgres.NewPostgresMessageRepository(dbPool) // Added messageRepo

	// Initialize use cases/services
	authUsecase := service.NewAuthService(userRepo, sessionRepo, cfg)
	userUsecase := service.NewUserService(userRepo, fileRepo)
	eventUsecase := service.NewEventService(redisEventRepo, pgEventRepo, userRepo)
	friendUsecase := service.NewFriendService(friendRepo, userRepo, eventUsecase)
	groupUsecase := service.NewGroupService(groupRepo, userRepo, friendRepo, fileRepo, eventUsecase)
	messageUsecase := service.NewMessageService(messageRepo, userRepo, groupRepo, eventUsecase) // Added messageUsecase

	// Initialize WebSocket Hub
	hub := websocket.NewHub(eventUsecase, messageUsecase) // Added messageUsecase to hub
	go hub.Run()

	// Start background worker for event persistence
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("Running background job: Persisting buffered events...") // Updated log message
				if err := eventUsecase.PersistBufferedEvents(context.Background()); err != nil {
					log.Printf("Error persisting buffered events: %v", err)
				}
			}
		}
	}()

	// Initialize router
	router := httpRouter.NewRouter(cfg, authUsecase, userUsecase, friendUsecase, groupUsecase, messageUsecase, hub) // Updated router package and added messageUsecase

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router); err != nil { // Used fmt.Sprintf and direct http.ListenAndServe
		log.Fatalf("Could not start server: %v", err) // Updated error message
	}
}

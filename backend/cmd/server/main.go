package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"chat-app/internal/adapter/localfile"
	"chat-app/internal/adapter/postgres"
	"chat-app/internal/adapter/redis"
	"chat-app/internal/delivery/websocket" // Added from attempted
	"chat-app/internal/service"            // Added from attempted
	"chat-app/pkg/config"

	httpRouter "chat-app/internal/delivery/http" // Alias for custom http router
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
		for event := range ticker.C {
			log.Println("Running background job: Persisting buffered events... :", event) // Updated log message
			eventUsecase.PersistBufferedEvents(context.Background())
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

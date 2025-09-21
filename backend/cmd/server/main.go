package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"chat-app/internal/adapter/localfile"
	"chat-app/internal/adapter/postgres"
	"chat-app/internal/adapter/redis"
	"chat-app/internal/delivery/http/handler"
	"chat-app/internal/delivery/http/router"
	"chat-app/internal/delivery/websocket"
	"chat-app/internal/service"
	"chat-app/pkg/config"
)

func main() {
	log.Println("Starting QuikChat server...")

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db := postgres.NewDB(cfg.DatabaseURL)
	// The attempted content removed explicit error handling here, assuming NewDB handles it internally or panics.
	defer db.Close()
	log.Println("Database connection established.")

	// Initialize Redis client
	redisClient := redis.NewClient(cfg.RedisURL)

	// Initialize repositories
	userRepo := postgres.NewPostgresUserRepository(db)
	sessionRepo := postgres.NewPostgresSessionRepository(db)
	fileRepo, err := localfile.NewLocalFileRepository(cfg.UploadDir)
	if err != nil {
		log.Fatalf("Failed to create file repository: %v", err)
	}
	redisEventRepo := redis.NewRedisEventRepository(redisClient)
	pgEventRepo := postgres.NewPostgresEventRepository(db)

	// Initialize use cases/services
	authUsecase := service.NewAuthService(userRepo, sessionRepo, cfg)
	userUsecase := service.NewUserService(userRepo, fileRepo)
	eventUsecase := service.NewEventService(redisEventRepo, pgEventRepo, userRepo)

	// Initialize WebSocket Hub
	hub := websocket.NewHub(eventUsecase)
	go hub.Run()

	// Start background worker for event persistence
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Run every minute
		defer ticker.Stop()
		for {
			<-ticker.C
			log.Println("Running background worker to persist events...")
			eventUsecase.PersistBufferedEvents(context.Background())
		}
	}()

	// Initialize router
	r := router.NewRouter(cfg, authUsecase, userUsecase, hub)

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}


package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"chat-app/adapter/filesystem"
	http_handler "chat-app/adapter/handler/http"
	"chat-app/adapter/handler/ws"
	"chat-app/adapter/middleware"
	"chat-app/adapter/postgres"
	"chat-app/adapter/redis"
	"chat-app/adapter/util"
	"chat-app/config"
	"chat-app/repository"
	"chat-app/usecase"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	rdb := redis.NewRedisClient(cfg)
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	// Repositories
	userRepo := postgres.NewPostgresUserRepository(db)
	sessionRepo := postgres.NewPostgresSessionRepository(db)
	fileRepo := filesystem.NewLocalStorage(cfg.ProfilePicDir, cfg.ProfilePicRoute)
	friendRepo := postgres.NewPostgresFriendshipRepository(db)
	groupRepo := postgres.NewPostgresGroupRepository(db)
	redisEventRepo := redis.NewRedisEventRepository(rdb)
	postgresEventRepo := postgres.NewPostgresEventRepository(db)

	// Utilities
	tokenGen := util.NewTokenGenerator(cfg.JWTSecret, cfg.AccessTokenExp, cfg.RefreshTokenExp)

	// Usecases
	eventUsecase := usecase.NewEventUsecase(redisEventRepo, postgresEventRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, tokenGen)
	userUsecase := usecase.NewUserUsecase(userRepo, fileRepo)
	friendUsecase := usecase.NewFriendUsecase(userRepo, friendRepo, eventUsecase)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo, friendRepo, fileRepo, eventUsecase)

	// Handlers
	authHandler := http_handler.NewAuthHandler(authUsecase)
	userHandler := http_handler.NewUserHandler(userUsecase)
	friendHandler := http_handler.NewFriendHandler(friendUsecase)
	groupHandler := http_handler.NewGroupHandler(groupUsecase)

	// WebSocket Hub
	hub := ws.NewHub(eventUsecase, groupUsecase)
	go hub.Run()

	// Background worker for Redis -> Postgres event persistence
	go func() {
		ticker := time.NewTicker(10 * time.Second) // Run every 10 seconds
		defer ticker.Stop()
		for range ticker.C {
			ctx := context.Background()
			events, err := redisEventRepo.GetBufferedEvents(ctx, 100)
			if err != nil || len(events) == 0 {
				continue
			}

			if err := postgresEventRepo.StoreBatch(ctx, events); err != nil {
				log.Printf("worker: failed to store events in postgres: %v", err)
				continue // Retry later
			}

			if err := redisEventRepo.DeleteBufferedEvents(ctx, events); err != nil {
				log.Printf("worker: failed to delete events from redis: %v", err)
			}
		}
	}()

	// Router
	r := chi.NewRouter()

	r.Use(chi_middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Post("/api/auth/register", userHandler.Register)
		r.Post("/api/auth/login", authHandler.Login)
		r.Post("/api/auth/refresh", authHandler.Refresh)
		r.Post("/api/auth/logout", authHandler.Logout)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Validate)

		// User routes
		r.Get("/api/users/{username}", userHandler.GetUserByUsername)
		r.Put("/api/users/profile", userHandler.UpdateProfile)

		// Friend routes
		r.Post("/api/friends/requests", friendHandler.SendRequest)
		r.Put("/api/friends/requests/{requesterID}", friendHandler.RespondToRequest)
		r.Delete("/api/friends/{friendID}", friendHandler.Unfriend)
		r.Get("/api/friends", friendHandler.ListFriends)
		r.Get("/api/friends/requests/pending", friendHandler.ListPendingRequests)

		// Group routes
		r.Post("/api/groups", groupHandler.CreateGroup)
		r.Post("/api/groups/join", groupHandler.JoinGroup)
		r.Post("/api/groups/{groupID}/leave", groupHandler.LeaveGroup)
		r.Post("/api/groups/{groupID}/members", groupHandler.AddMember)
		r.Delete("/api/groups/{groupID}/members/{memberID}", groupHandler.RemoveMember)
		r.Get("/api/groups/search", groupHandler.SearchGroups)

		// WebSocket
		r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWs(hub, w, r)
		})
	})

	// Serve static files for profile pictures
	fileServer(r, cfg.ProfilePicRoute, http.Dir(cfg.ProfilePicDir))

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}


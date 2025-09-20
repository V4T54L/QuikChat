package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"chat-app/adapter/filesystem"
	httpHandler "chat-app/adapter/handler/http"
	"chat-app/adapter/handler/ws"
	"chat-app/adapter/middleware"
	"chat-app/adapter/postgres"
	"chat-app/adapter/redis"
	"chat-app/adapter/util"
	"chat-app/config"
	"chat-app/usecase"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
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
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	rdb := redis.NewRedisClient(cfg)
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// Repositories
	userRepo := postgres.NewPostgresUserRepository(db)
	sessionRepo := postgres.NewPostgresSessionRepository(db)
	friendshipRepo := postgres.NewPostgresFriendshipRepository(db)
	groupRepo := postgres.NewPostgresGroupRepository(db)
	fileRepo := filesystem.NewLocalStorage(cfg.ProfilePicDir, cfg.ProfilePicRoute)
	redisEventRepo := redis.NewRedisEventRepository(rdb)
	dbEventRepo := postgres.NewPostgresEventRepository(db)

	// Utilities
	tokenGenerator := util.NewTokenGenerator(cfg.JWTSecret, cfg.AccessTokenExp, cfg.RefreshTokenExp)

	// Usecases
	eventUsecase := usecase.NewEventUsecase(redisEventRepo, dbEventRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, tokenGenerator)
	userUsecase := usecase.NewUserUsecase(userRepo, fileRepo)
	friendUsecase := usecase.NewFriendUsecase(userRepo, friendshipRepo, eventUsecase)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo, friendshipRepo, fileRepo, eventUsecase)

	// Handlers
	authHandler := httpHandler.NewAuthHandler(authUsecase)
	userHandler := httpHandler.NewUserHandler(userUsecase)
	friendHandler := httpHandler.NewFriendHandler(friendUsecase)
	groupHandler := httpHandler.NewGroupHandler(groupUsecase)
	webHandler := httpHandler.NewWebHandler("web/templates")

	// WebSocket Hub
	hub := ws.NewHub(eventUsecase, groupUsecase)
	go hub.Run()

	// Background worker for event persistence
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				events, err := redisEventRepo.GetBufferedEvents(ctx, 100)
				if err != nil || len(events) == 0 {
					cancel()
					continue
				}

				if err := dbEventRepo.StoreBatch(ctx, events); err != nil {
					log.Printf("Error storing event batch to DB: %v", err)
					cancel()
					continue
				}

				if err := redisEventRepo.DeleteBufferedEvents(ctx, events); err != nil {
					log.Printf("Error deleting buffered events from Redis: %v", err)
				}
				cancel()
			}
		}
	}()

	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.Logging)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	// Protected API routes
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMiddleware.Validate)

		// User routes
		r.Get("/users/{username}", userHandler.GetUserByUsername)
		r.Put("/users/me/profile", userHandler.UpdateProfile)

		// Friend routes
		r.Get("/friends", friendHandler.ListFriends)
		r.Get("/friends/pending", friendHandler.ListPendingRequests)
		r.Post("/friends/request", friendHandler.SendRequest)
		r.Put("/friends/request/{requesterID}", friendHandler.RespondToRequest)
		r.Delete("/friends/{friendID}", friendHandler.Unfriend)

		// Group routes
		r.Post("/groups", groupHandler.CreateGroup)
		r.Post("/groups/join", groupHandler.JoinGroup)
		r.Post("/groups/{groupID}/leave", groupHandler.LeaveGroup)
		r.Post("/groups/{groupID}/members", groupHandler.AddMember)
		r.Delete("/groups/{groupID}/members/{memberID}", groupHandler.RemoveMember)
		r.Get("/groups/search", groupHandler.SearchGroups)

		// WebSocket route
		r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWs(hub, w, r)
		})
	})

	// Serve static files
	fileServer(r, "/static", http.Dir("web/static"))
	fileServer(r, cfg.ProfilePicRoute, http.Dir(cfg.ProfilePicDir))

	// Serve Web App
	r.Get("/", webHandler.ServeApp)

	// Server setup
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

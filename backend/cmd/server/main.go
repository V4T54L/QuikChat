package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"chat-app/backend/adapter/filesystem"
	httpHandler "chat-app/backend/adapter/handler/http"
	"chat-app/backend/adapter/handler/ws"
	"chat-app/backend/adapter/middleware"
	"chat-app/backend/adapter/postgres"
	"chat-app/backend/adapter/redis"
	"chat-app/backend/adapter/util"
	"chat-app/backend/config"
	"chat-app/backend/usecase"

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
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
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
	friendRepo := postgres.NewPostgresFriendshipRepository(db)
	groupRepo := postgres.NewPostgresGroupRepository(db)
	fileRepo := filesystem.NewLocalStorage(cfg.ProfilePicDir, cfg.ProfilePicRoute)
	redisEventRepo := redis.NewRedisEventRepository(rdb)
	dbEventRepo := postgres.NewPostgresEventRepository(db)

	// Utilities
	tokenGen := util.NewTokenGenerator(cfg.JWTSecret, cfg.AccessTokenExp, cfg.RefreshTokenExp)

	// Usecases
	eventUsecase := usecase.NewEventUsecase(redisEventRepo, dbEventRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, tokenGen)
	userUsecase := usecase.NewUserUsecase(userRepo, fileRepo)
	friendUsecase := usecase.NewFriendUsecase(userRepo, friendRepo, eventUsecase)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo, friendRepo, fileRepo, eventUsecase)

	// Handlers
	authHandler := httpHandler.NewAuthHandler(authUsecase)
	userHandler := httpHandler.NewUserHandler(userUsecase)
	friendHandler := httpHandler.NewFriendHandler(friendUsecase)
	groupHandler := httpHandler.NewGroupHandler(groupUsecase)
	webHandler := httpHandler.NewWebHandler("./web/templates")

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
				if err != nil {
					log.Printf("error getting buffered events: %v", err)
					cancel()
					continue
				}
				if len(events) > 0 {
					if err := dbEventRepo.StoreBatch(ctx, events); err != nil {
						log.Printf("error storing event batch to db: %v", err)
					} else {
						if err := redisEventRepo.DeleteBufferedEvents(ctx, events); err != nil {
							log.Printf("error deleting buffered events from redis: %v", err)
						}
					}
				}
				cancel()
			}
		}
	}()

	router := chi.NewRouter()
	router.Use(chiMiddleware.Recoverer)
	router.Use(middleware.Logging)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health and Metrics endpoints
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := struct {
			ConnectedClients int `json:"connected_clients"`
		}{
			ConnectedClients: hub.GetClientCount(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(metrics)
	})

	// Public API routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.RateLimit)
		r.Post("/api/v1/register", userHandler.Register)
		r.Post("/api/v1/login", authHandler.Login)
		r.Post("/api/v1/refresh", authHandler.Refresh)
		r.Post("/api/v1/logout", authHandler.Logout)
	})

	// Protected API routes
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Validate)
		r.Use(middleware.RateLimit)

		// User routes
		r.Get("/api/v1/users/{username}", userHandler.GetUserByUsername)
		r.Put("/api/v1/me", userHandler.UpdateProfile)

		// Friend routes
		r.Post("/api/v1/friends/requests", friendHandler.SendRequest)
		r.Put("/api/v1/friends/requests/{requesterID}", friendHandler.RespondToRequest)
		r.Delete("/api/v1/friends/{friendID}", friendHandler.Unfriend)
		r.Get("/api/v1/friends", friendHandler.ListFriends)
		r.Get("/api/v1/friends/requests/pending", friendHandler.ListPendingRequests)

		// Group routes
		r.Post("/api/v1/groups", groupHandler.CreateGroup)
		r.Post("/api/v1/groups/join", groupHandler.JoinGroup)
		r.Post("/api/v1/groups/{groupID}/leave", groupHandler.LeaveGroup)
		r.Post("/api/v1/groups/{groupID}/members", groupHandler.AddMember)
		r.Delete("/api/v1/groups/{groupID}/members/{memberID}", groupHandler.RemoveMember)
		r.Get("/api/v1/groups/search", groupHandler.SearchGroups)

		// WebSocket route
		r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWs(hub, w, r)
		})
	})

	// Serve static files
	fileServer(router, "/static", http.Dir("web/static"))
	fileServer(router, cfg.ProfilePicRoute, http.Dir(cfg.ProfilePicDir))

	// Serve Web App
	router.Get("/*", webHandler.ServeApp)

	// Server setup
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
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

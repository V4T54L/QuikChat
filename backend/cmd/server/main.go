package main

import (
	"chat-app/backend/adapter/filesystem"
	"chat-app/backend/adapter/handler/http"
	"chat-app/backend/adapter/middleware"
	"chat-app/backend/adapter/postgres"
	"chat-app/backend/adapter/util"
	"chat-app/backend/config"
	"chat-app/backend/usecase"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

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

	// Repositories
	userRepo := postgres.NewPostgresUserRepository(db)
	sessionRepo := postgres.NewPostgresSessionRepository(db)
	fileRepo := filesystem.NewLocalStorage(cfg.ProfilePicDir, cfg.ProfilePicRoute)

	// Usecases
	tokenGenerator := util.NewTokenGenerator(cfg.JWTSecret, cfg.AccessTokenExp, cfg.RefreshTokenExp)
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, tokenGenerator)
	userUsecase := usecase.NewUserUsecase(userRepo, fileRepo)

	// Handlers
	authHandler := http.NewAuthHandler(authUsecase)
	userHandler := http.NewUserHandler(userUsecase)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

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

	// Serve static files (profile pictures)
	fileServer(r, cfg.ProfilePicRoute, http.Dir(cfg.ProfilePicDir))

	// Public routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)

		r.Get("/users/{username}", userHandler.GetUserByUsername)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Validate)
			r.Put("/profile", userHandler.UpdateProfile)
		})
	})

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}


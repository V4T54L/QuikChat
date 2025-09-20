package main

import (
	"log"
	"net/http"
	"strings"

	"chat-app/adapter/filesystem"
	http_handler "chat-app/adapter/handler/http"
	"chat-app/adapter/middleware"
	"chat-app/adapter/postgres"
	"chat-app/adapter/util"
	"chat-app/config"
	"chat-app/usecase"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewPostgresUserRepository(db)
	sessionRepo := postgres.NewPostgresSessionRepository(db)
	fileRepo := filesystem.NewLocalStorage(cfg.ProfilePicDir, cfg.ProfilePicRoute)
	friendRepo := postgres.NewPostgresFriendshipRepository(db)
	groupRepo := postgres.NewPostgresGroupRepository(db)

	// Initialize utilities
	tokenGen := util.NewTokenGenerator(cfg.JWTSecret, cfg.AccessTokenExp, cfg.RefreshTokenExp)

	// Initialize use cases
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, tokenGen)
	userUsecase := usecase.NewUserUsecase(userRepo, fileRepo)
	friendUsecase := usecase.NewFriendUsecase(userRepo, friendRepo)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo, friendRepo, fileRepo)

	// Initialize handlers
	authHandler := http_handler.NewAuthHandler(authUsecase)
	userHandler := http_handler.NewUserHandler(userUsecase)
	friendHandler := http_handler.NewFriendHandler(friendUsecase)
	groupHandler := http_handler.NewGroupHandler(groupUsecase)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// Setup router
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

	// Serve static files for profile pictures
	fileServer(r, cfg.ProfilePicRoute, http.Dir(cfg.ProfilePicDir))

	// Public routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
		r.Get("/users/{username}", userHandler.GetUserByUsername)
		r.Get("/groups/search", groupHandler.SearchGroups)
	})

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authMiddleware.Validate)

		// User profile
		r.Put("/profile", userHandler.UpdateProfile)

		// Friends
		r.Route("/friends", func(r chi.Router) {
			r.Get("/", friendHandler.ListFriends)
			r.Delete("/{friendID}", friendHandler.Unfriend)

			r.Route("/requests", func(r chi.Router) {
				r.Get("/", friendHandler.ListPendingRequests)
				r.Post("/", friendHandler.SendRequest)
				r.Put("/{requesterID}", friendHandler.RespondToRequest)
			})
		})

		// Groups
		r.Route("/groups", func(r chi.Router) {
			r.Post("/", groupHandler.CreateGroup)
			r.Post("/join", groupHandler.JoinGroup)

			r.Route("/{groupID}", func(r chi.Router) {
				// r.Get("/", groupHandler.GetDetails)
				// r.Put("/", groupHandler.UpdateGroup)
				r.Post("/leave", groupHandler.LeaveGroup)

				r.Route("/members", func(r chi.Router) {
					// r.Get("/", groupHandler.ListMembers)
					r.Post("/", groupHandler.AddMember)
					r.Delete("/{memberID}", groupHandler.RemoveMember)
				})
			})
		})
	})

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

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


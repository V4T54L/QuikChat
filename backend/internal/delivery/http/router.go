package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chat-app/internal/delivery/http/handler"
	"chat-app/internal/delivery/http/middleware"
	"chat-app/internal/delivery/websocket"
	"chat-app/internal/usecase"
	"chat-app/pkg/config"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, authUsecase usecase.AuthUsecase, userUsecase usecase.UserUsecase, friendUsecase usecase.FriendUsecase, hub *websocket.Hub) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Heartbeat("/healthz"))

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	wsHandler := handler.NewWebSocketHandler(hub)
	friendHandler := handler.NewFriendHandler(friendUsecase)

	// Public routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Get("/ws", wsHandler.ServeWS)

		r.Route("/api/v1/users", func(r chi.Router) {
			r.Get("/me", userHandler.GetMyProfile)
			r.Put("/me", userHandler.UpdateMyProfile)
			r.Get("/{username}", userHandler.GetUserProfile)
		})

		r.Route("/api/v1/friends", func(r chi.Router) {
			r.Post("/requests", friendHandler.SendRequest)
			r.Get("/requests/pending", friendHandler.GetPendingRequests)
			r.Put("/requests/{requestID}/accept", friendHandler.AcceptRequest)
			r.Put("/requests/{requestID}/reject", friendHandler.RejectRequest)
			r.Delete("/{userID}", friendHandler.Unfriend)
			r.Get("/", friendHandler.ListFriends)
		})

		r.Post("/api/v1/auth/logout", authHandler.Logout)
	})

	// Serve frontend files
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	FileServer(r, "/", filesDir)

	return r
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
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
		// Check if the file exists
		_, err := root.Open(r.URL.Path)
		if os.IsNotExist(err) {
			// If not, serve chat.html for SPA routing on any sub-path
			http.ServeFile(w, r, filepath.Join("web", "chat.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}


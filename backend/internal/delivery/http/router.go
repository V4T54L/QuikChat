package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chat-app/internal/delivery/http/handler"
	"chat-app/internal/delivery/http/middleware"
	ws "chat-app/internal/delivery/websocket"
	"chat-app/internal/usecase"
	"chat-app/pkg/config"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, authUsecase usecase.AuthUsecase, userUsecase usecase.UserUsecase, hub *ws.Hub) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Heartbeat("/healthz"))

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	wsHandler := handler.NewWebSocketHandler(hub)
	authMiddleware := middleware.AuthMiddleware(cfg)

	// Public routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/ws", wsHandler.ServeWS)

		r.Route("/api/v1/users", func(r chi.Router) {
			r.Get("/me", userHandler.GetMyProfile)
			r.Put("/me", userHandler.UpdateMyProfile)
			r.Get("/{username}", userHandler.GetUserProfile)
		})
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
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		// Check if the file exists
		_, err := root.Open(r.URL.Path)
		if os.IsNotExist(err) {
			// If not, serve chat.html as the SPA fallback
			http.ServeFile(w, r, filepath.Join(root.(http.Dir).String(), "templates/chat.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}


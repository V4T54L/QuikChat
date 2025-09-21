package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chat-app/internal/delivery/http/handler"
	"chat-app/internal/delivery/http/middleware"
	"chat-app/internal/usecase"
	"chat-app/pkg/config"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, authUsecase usecase.AuthUsecase, userUsecase usecase.UserUsecase) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Heartbeat("/healthz"))

	authHandler := handler.NewAuthHandler(authUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	authMiddleware := middleware.AuthMiddleware(cfg)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", authHandler.SignUp)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		r.With(authMiddleware).Route("/users", func(r chi.Router) {
			r.Get("/me", userHandler.GetMyProfile)
			r.Put("/me", userHandler.UpdateMyProfile)
		})
		r.Get("/users/{username}", userHandler.GetUserProfile)
	})

	workDir, _ := os.Getwd()
	staticPath := http.Dir(filepath.Join(workDir, "web/static"))
	uploadsPath := http.Dir(filepath.Join(workDir, cfg.UploadDir))

	FileServer(r, "/static", staticPath)
	FileServer(r, "/uploads", uploadsPath)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "web/templates/index.html"))
	})
	r.Get("/chat", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(workDir, "web/templates/chat.html"))
	})

	return r
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
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


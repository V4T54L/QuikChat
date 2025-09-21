package http

import (
	"encoding/json"
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
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(
	cfg *config.Config,
	authUsecase usecase.AuthUsecase,
	userUsecase usecase.UserUsecase,
	friendUsecase usecase.FriendUsecase,
	groupUsecase usecase.GroupUsecase,
	messageUsecase usecase.MessageUsecase,
	hub *websocket.Hub,
) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	// r.Use(middleware.NewIPRateLimiter(1, 5).Limit)

	// Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	wsHandler := handler.NewWebSocketHandler(hub)
	friendHandler := handler.NewFriendHandler(friendUsecase)
	groupHandler := handler.NewGroupHandler(groupUsecase)
	messageHandler := handler.NewMessageHandler(messageUsecase)

	// Public routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		// WebSocket
		r.Get("/ws", wsHandler.ServeWS)

		// Auth
		r.Post("/api/v1/auth/logout", authHandler.Logout) // Moved back to protected

		// User routes
		r.Route("/api/v1/users", func(r chi.Router) {
			r.Get("/me", userHandler.GetMyProfile)
			r.Put("/me", userHandler.UpdateMyProfile)
			r.Get("/{username}", userHandler.GetUserProfile)
		})

		// Friend routes
		r.Route("/api/v1/friends", func(r chi.Router) {
			r.Get("/", friendHandler.ListFriends)
			r.Post("/requests", friendHandler.SendRequest)
			r.Get("/requests/pending", friendHandler.GetPendingRequests)
			r.Put("/requests/{request_id}/accept", friendHandler.AcceptRequest)
			r.Put("/requests/{request_id}/reject", friendHandler.RejectRequest)
			r.Delete("/{user_id}", friendHandler.Unfriend)
		})

		// Group routes
		r.Route("/api/v1/groups", func(r chi.Router) {
			r.Post("/", groupHandler.CreateGroup)
			r.Get("/search", groupHandler.SearchGroups)
			r.Get("/me", groupHandler.ListMyGroups)
			r.Post("/{handle}/join", groupHandler.JoinGroup)
			r.Get("/{group_id}", groupHandler.GetGroupDetails)
			r.Put("/{group_id}", groupHandler.UpdateGroup)
			r.Put("/{group_id}/transfer-ownership", groupHandler.TransferOwnership)
			r.Post("/{group_id}/members", groupHandler.AddMember)
			r.Delete("/{group_id}/members/{member_id}", groupHandler.RemoveMember)
			r.Post("/{group_id}/leave", groupHandler.LeaveGroup)
		})

		// Messages routes
		r.Route("/api/v1/conversations", func(r chi.Router) {
			r.Get("/{conversation_id}/messages", messageHandler.GetHistory)
		})
	})

	// Monitoring endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	r.Handle("/metrics", promhttp.Handler())

	// Serve frontend files
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "web"))
	FileServer(r, "/", filesDir)

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
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
		f, err := root.Open(r.URL.Path)
		if os.IsNotExist(err) {
			// If not found, serve the main chat.html for SPA routing
			http.ServeFile(w, r, filepath.Join(string(root.(http.Dir)), "templates/chat.html"))
			return
		}
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		f.Close()

		// Otherwise, serve the file
		fs.ServeHTTP(w, r)
	})
}

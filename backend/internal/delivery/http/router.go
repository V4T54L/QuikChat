package http

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

func NewRouter(
	cfg *config.Config,
	authUsecase usecase.AuthUsecase,
	userUsecase usecase.UserUsecase,
	friendUsecase usecase.FriendUsecase,
	groupUsecase usecase.GroupUsecase,
	hub *websocket.Hub,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Heartbeat("/healthz"))

	authHandler := handler.NewAuthHandler(authUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	wsHandler := handler.NewWebSocketHandler(hub)
	friendHandler := handler.NewFriendHandler(friendUsecase)
	groupHandler := handler.NewGroupHandler(groupUsecase)

	// Public API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/signup", authHandler.SignUp)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)
	})

	// Protected API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Get("/ws", wsHandler.ServeWS)
		r.Post("/auth/logout", authHandler.Logout)

		// User routes
		r.Get("/users/me", userHandler.GetMyProfile)
		r.Put("/users/me", userHandler.UpdateMyProfile)
		r.Get("/users/{username}", userHandler.GetUserProfile)

		// Friend routes
		r.Post("/friends/requests", friendHandler.SendRequest)
		r.Get("/friends/requests/pending", friendHandler.GetPendingRequests)
		r.Put("/friends/requests/{request_id}/accept", friendHandler.AcceptRequest)
		r.Put("/friends/requests/{request_id}/reject", friendHandler.RejectRequest)
		r.Delete("/friends/{user_id}", friendHandler.Unfriend)
		r.Get("/friends", friendHandler.ListFriends)

		// Group routes
		r.Post("/groups", groupHandler.CreateGroup)
		r.Get("/groups/search", groupHandler.SearchGroups)
		r.Post("/groups/{handle}/join", groupHandler.JoinGroup)
		r.Get("/groups/{group_id}", groupHandler.GetGroupDetails)
		r.Put("/groups/{group_id}", groupHandler.UpdateGroup)
		r.Put("/groups/{group_id}/transfer-ownership", groupHandler.TransferOwnership)
		r.Post("/groups/{group_id}/members", groupHandler.AddMember)
		r.Delete("/groups/{group_id}/members/{member_id}", groupHandler.RemoveMember)
		r.Post("/groups/{group_id}/leave", groupHandler.LeaveGroup)
		r.Get("/groups/me", groupHandler.ListMyGroups)
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
		f, err := root.Open(r.URL.Path)
		if os.IsNotExist(err) {
			// If not, serve the main chat.html for SPA routing
			http.ServeFile(w, r, filepath.Join("web", "templates", "chat.html"))
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


package handler

import (
	"chat-app/internal/delivery/http/middleware"
	"chat-app/internal/service"
	"chat-app/internal/usecase"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userUsecase usecase.UserUsecase
}

func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: uc}
}

func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := h.userUsecase.GetUserByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max memory
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var updateInput usecase.UpdateUserInput
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != "" {
		updateInput.Username = &username
	}
	if password != "" {
		updateInput.Password = &password
	}

	if updateInput.Username != nil || updateInput.Password != nil {
		if _, err := h.userUsecase.UpdateProfile(r.Context(), userID, updateInput); err != nil {
			switch err {
			case service.ErrUserExists, service.ErrInvalidUsername, service.ErrInvalidPassword:
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			}
			return
		}
	}

	file, header, err := r.FormFile("profile_pic")
	if err == nil {
		defer file.Close()
		if _, err := h.userUsecase.UpdateProfilePicture(r.Context(), userID, file, header); err != nil {
			switch err {
			case service.ErrFileSizeExceeded, service.ErrInvalidFileType:
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
			}
			return
		}
	} else if err != http.ErrMissingFile {
		http.Error(w, "Could not process file: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedUser, err := h.userUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}


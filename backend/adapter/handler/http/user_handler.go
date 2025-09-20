package http

import (
	"chat-app/backend/adapter/middleware"
	"chat-app/backend/adapter/util"
	"chat-app/backend/models"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := h.userUsecase.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrUsernameTaken) || err.Error() == "invalid username format" || err.Error() == "password is too short" {
			util.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Username is required")
		return
	}

	user, err := h.userUsecase.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			util.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		util.RespondWithError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	var username, password *string
	if val := r.FormValue("username"); val != "" {
		username = &val
	}
	if val := r.FormValue("password"); val != "" {
		password = &val
	}

	file, header, err := r.FormFile("profilePic")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	if file != nil {
		defer file.Close()
	}

	user, err := h.userUsecase.UpdateProfile(r.Context(), userID, username, password, file, header)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			util.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, models.ErrUsernameTaken) {
			util.RespondWithError(w, http.StatusConflict, err.Error())
			return
		}
		// Check for validation errors
		if err.Error() == "invalid username format" || err.Error() == "password is too short" || err.Error() == "file size exceeds 200KB" || err.Error() == "invalid file type" {
			util.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, user)
}


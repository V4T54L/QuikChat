package http

import (
	"chat-app/backend/adapter/util"
	"chat-app/backend/models"
	"encoding/json"
	"errors"
	"net/http"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	accessToken, refreshToken, err := h.authUsecase.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			util.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type refreshResponse struct {
	AccessToken string `json:"accessToken"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	newAccessToken, err := h.authUsecase.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) || errors.Is(err, models.ErrInvalidToken) {
			util.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, refreshResponse{AccessToken: newAccessToken})
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.authUsecase.Logout(r.Context(), req.RefreshToken); err != nil {
		// We can choose to not return an error to the client for logout failures
		// for security reasons, but for simplicity we will.
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}


package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"chat-app/internal/service"
	"chat-app/internal/usecase"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var input usecase.SignUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authUsecase.SignUp(r.Context(), input)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	// Don't return password hash
	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input usecase.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokens, err := h.authUsecase.Login(r.Context(), input)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokens, err := h.authUsecase.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authUsecase.Logout(r.Context(), body.RefreshToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserExists),
		errors.Is(err, service.ErrInvalidUsername),
		errors.Is(err, service.ErrInvalidPassword):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrSessionNotFound),
		errors.Is(err, service.ErrSessionExpired):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}


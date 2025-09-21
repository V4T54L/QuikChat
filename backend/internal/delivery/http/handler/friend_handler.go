package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"chat-app/internal/delivery/http/middleware"
	"chat-app/internal/service"
	"chat-app/internal/usecase"

	"github.com/go-chi/chi/v5"
)

type FriendHandler struct {
	friendUsecase usecase.FriendUsecase
}

func NewFriendHandler(uc usecase.FriendUsecase) *FriendHandler {
	return &FriendHandler{friendUsecase: uc}
}

func (h *FriendHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	friendRequest, err := h.friendUsecase.SendFriendRequest(r.Context(), userID, req.Username)
	if err != nil {
		handleFriendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(friendRequest)
}

func (h *FriendHandler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := h.friendUsecase.GetPendingRequests(r.Context(), userID)
	if err != nil {
		handleFriendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func (h *FriendHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requestID := chi.URLParam(r, "requestID")

	if err := h.friendUsecase.AcceptFriendRequest(r.Context(), userID, requestID); err != nil {
		handleFriendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	requestID := chi.URLParam(r, "requestID")

	if err := h.friendUsecase.RejectFriendRequest(r.Context(), userID, requestID); err != nil {
		handleFriendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendHandler) Unfriend(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	friendID := chi.URLParam(r, "userID")

	if err := h.friendUsecase.Unfriend(r.Context(), userID, friendID); err != nil {
		handleFriendError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	friends, err := h.friendUsecase.ListFriends(r.Context(), userID)
	if err != nil {
		handleFriendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(friends)
}

func handleFriendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrFriendRequestInvalid),
		errors.Is(err, service.ErrFriendRequestYourself):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, service.ErrFriendRequestExists),
		errors.Is(err, service.ErrAlreadyFriends):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, service.ErrFriendRequestNotReceiver):
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}


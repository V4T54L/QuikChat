package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"chat-app/adapter/middleware"
	"chat-app/adapter/util"
	"chat-app/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *FriendHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.friendUsecase.SendRequest(r.Context(), userID, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserNotFound):
			util.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, models.ErrAlreadyFriends), errors.Is(err, models.ErrFriendRequestExists), errors.Is(err, models.ErrCannotFriendSelf):
			util.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not send friend request")
		}
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Friend request sent"})
}

func (h *FriendHandler) RespondToRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	requesterIDStr := chi.URLParam(r, "requesterID")
	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid requester ID")
		return
	}

	var req struct {
		Action string `json:"action"` // "accept" or "reject"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var usecaseErr error
	message := ""
	if req.Action == "accept" {
		usecaseErr = h.friendUsecase.AcceptRequest(r.Context(), userID, requesterID)
		message = "Friend request accepted"
	} else if req.Action == "reject" {
		usecaseErr = h.friendUsecase.RejectRequest(r.Context(), userID, requesterID)
		message = "Friend request rejected"
	} else {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid action")
		return
	}

	if usecaseErr != nil {
		switch {
		case errors.Is(usecaseErr, models.ErrFriendRequestNotFound):
			util.RespondWithError(w, http.StatusNotFound, usecaseErr.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not respond to friend request")
		}
		return
	}

	util.RespondWithJSON(w, http.StatusOK, map[string]string{"message": message})
}

func (h *FriendHandler) Unfriend(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	friendIDStr := chi.URLParam(r, "friendID")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid friend ID")
		return
	}

	err = h.friendUsecase.Unfriend(r.Context(), userID, friendID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFriends), errors.Is(err, models.ErrFriendRequestNotFound):
			util.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not unfriend user")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FriendHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	friends, err := h.friendUsecase.ListFriends(r.Context(), userID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Could not list friends")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, friends)
}

func (h *FriendHandler) ListPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	requests, err := h.friendUsecase.ListPendingRequests(r.Context(), userID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Could not list pending requests")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, requests)
}


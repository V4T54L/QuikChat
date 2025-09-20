package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"chat-app/backend/adapter/middleware"
	"chat-app/backend/adapter/util"
	"chat-app/backend/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxGroupPhotoSize = 200 * 1024 // 200 KB

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	if err := r.ParseMultipartForm(maxGroupPhotoSize); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Could not parse form")
		return
	}

	handle := r.FormValue("handle")
	name := r.FormValue("name")
	if handle == "" || name == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Handle and name are required")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		util.RespondWithError(w, http.StatusBadRequest, "Could not get photo")
		return
	}
	if file != nil {
		defer file.Close()
	}

	group, err := h.groupUsecase.CreateGroup(r.Context(), userID, handle, name, file, header)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrGroupHandleTaken):
			util.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not create group")
		}
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, group)
}

func (h *GroupHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	var req struct {
		Handle string `json:"handle"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.groupUsecase.JoinGroup(r.Context(), userID, req.Handle)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrGroupNotFound):
			util.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, models.ErrAlreadyGroupMember):
			util.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not join group")
		}
		return
	}

	util.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Successfully joined group"})
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	groupIDStr := chi.URLParam(r, "groupID")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	err = h.groupUsecase.LeaveGroup(r.Context(), userID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrGroupNotFound), errors.Is(err, models.ErrNotGroupMember):
			util.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not leave group")
		}
		return
	}

	util.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Successfully left group"})
}

func (h *GroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	adderID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	groupIDStr := chi.URLParam(r, "groupID")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.groupUsecase.AddMember(r.Context(), adderID, req.Username, groupID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotGroupMember), errors.Is(err, models.ErrUserNotFound), errors.Is(err, models.ErrNotFriends):
			util.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, models.ErrAlreadyGroupMember):
			util.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not add member")
		}
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Member added successfully"})
}

func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	groupIDStr := chi.URLParam(r, "groupID")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	memberIDStr := chi.URLParam(r, "memberID")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	err = h.groupUsecase.RemoveMember(r.Context(), ownerID, memberID, groupID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotGroupOwner), errors.Is(err, models.ErrCannotRemoveOwner):
			util.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, models.ErrGroupNotFound), errors.Is(err, models.ErrNotGroupMember):
			util.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			util.RespondWithError(w, http.StatusInternalServerError, "Could not remove member")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) SearchGroups(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	groups, err := h.groupUsecase.SearchGroups(r.Context(), query)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Could not search for groups")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, groups)
}

package handler

import (
	"chat-app/internal/delivery/http/middleware"
	"chat-app/internal/service"
	"chat-app/internal/usecase"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type GroupHandler struct {
	uc usecase.GroupUsecase
}

func NewGroupHandler(uc usecase.GroupUsecase) *GroupHandler {
	return &GroupHandler{uc: uc}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	var input usecase.CreateGroupInput
	input.Handle = r.FormValue("handle")
	input.Name = r.FormValue("name")

	file, header, err := r.FormFile("photo")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		http.Error(w, "Failed to get photo from form", http.StatusBadRequest)
		return
	}
	if file != nil {
		defer file.Close()
	}

	group, err := h.uc.CreateGroup(r.Context(), userID, input, file, header)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, group)
}

func (h *GroupHandler) SearchGroups(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	groups, err := h.uc.SearchGroups(r.Context(), query)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, groups)
}

func (h *GroupHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	handle := chi.URLParam(r, "handle")

	group, err := h.uc.JoinGroup(r.Context(), userID, "#"+handle)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, group)
}

func (h *GroupHandler) GetGroupDetails(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "group_id")

	details, err := h.uc.GetGroupDetails(r.Context(), groupID)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, details)
}

func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	groupID := chi.URLParam(r, "group_id")

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	var input usecase.UpdateGroupInput
	if name := r.FormValue("name"); name != "" {
		input.Name = &name
	}

	file, header, err := r.FormFile("photo")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		http.Error(w, "Failed to get photo from form", http.StatusBadRequest)
		return
	}
	if file != nil {
		defer file.Close()
	}

	group, err := h.uc.UpdateGroup(r.Context(), userID, groupID, input, file, header)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, group)
}

func (h *GroupHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	groupID := chi.URLParam(r, "group_id")

	var input struct {
		NewOwnerID string `json:"new_owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.uc.TransferOwnership(r.Context(), userID, groupID, input.NewOwnerID)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	groupID := chi.URLParam(r, "group_id")

	var input struct {
		FriendID string `json:"friend_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.uc.AddMember(r.Context(), userID, groupID, input.FriendID)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	groupID := chi.URLParam(r, "group_id")
	memberID := chi.URLParam(r, "member_id")

	err := h.uc.RemoveMember(r.Context(), userID, groupID, memberID)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	groupID := chi.URLParam(r, "group_id")

	err := h.uc.LeaveGroup(r.Context(), userID, groupID)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)

	groups, err := h.uc.ListUserGroups(r.Context(), userID)
	if err != nil {
		handleGroupError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, groups)
}

func handleGroupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidGroupHandle), errors.Is(err, service.ErrAddNotFriend), errors.Is(err, service.ErrTransferToNonMember), errors.Is(err, service.ErrTransferToSelf), errors.Is(err, service.ErrRemoveSelf):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrNotGroupOwner), errors.Is(err, service.ErrNotGroupMember):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, repository.ErrNotFound):
		http.Error(w, "Group not found", http.StatusNotFound)
	case errors.Is(err, repository.ErrGroupHandleExists), errors.Is(err, repository.ErrGroupMemberExists):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "An internal error occurred", http.StatusInternalServerError)
	}
}


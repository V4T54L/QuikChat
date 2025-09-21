package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat-app/backend/internal/delivery/http/middleware"
	"chat-app/backend/internal/usecase"

	"github.com/go-chi/chi/v5"
)

type MessageHandler struct {
	messageUsecase usecase.MessageUsecase
}

func NewMessageHandler(uc usecase.MessageUsecase) *MessageHandler {
	return &MessageHandler{messageUsecase: uc}
}

func (h *MessageHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversationID := chi.URLParam(r, "conversation_id")
	if conversationID == "" {
		http.Error(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	beforeStr := r.URL.Query().Get("before")
	before := time.Now().UTC()
	if beforeStr != "" {
		var err error
		before, err = time.Parse(time.RFC3339Nano, beforeStr)
		if err != nil {
			http.Error(w, "Invalid 'before' timestamp format", http.StatusBadRequest)
			return
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			limit = 50 // Reset to default if invalid
		}
	}

	messages, err := h.messageUsecase.GetMessageHistory(r.Context(), userID, conversationID, before, limit)
	if err != nil {
		// Here you might want to map specific service errors to HTTP status codes
		http.Error(w, "Failed to fetch message history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}


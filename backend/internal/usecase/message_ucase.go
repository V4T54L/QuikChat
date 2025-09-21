package usecase

import (
	"context"
	"time"

	"chat-app/backend/internal/domain"
)

type SendMessageInput struct {
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
}

type MessageUsecase interface {
	SendMessage(ctx context.Context, senderID string, input SendMessageInput) (*domain.Message, error)
	GetMessageHistory(ctx context.Context, userID, conversationID string, before time.Time, limit int) ([]*domain.Message, error)
}


package repository

import (
	"context"
	"time"

	"chat-app/backend/internal/domain"
)

type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) error
	GetByConversationID(ctx context.Context, conversationID string, before time.Time, limit int) ([]*domain.Message, error)
}


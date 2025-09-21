package usecase

import (
	"chat-app/internal/domain"
	"context"
)

type EventUsecase interface {
	CreateAndBufferEvent(ctx context.Context, eventType string, payload interface{}, recipientID string) error
	PersistBufferedEvents(ctx context.Context)
}


package usecase

import (
	"context"
)

type EventUsecase interface {
	CreateAndBufferEvent(ctx context.Context, eventType string, payload interface{}, recipientID string) error
	PersistBufferedEvents(ctx context.Context)
}

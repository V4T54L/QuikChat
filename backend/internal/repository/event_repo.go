package repository

import (
	"context"
	"chat-app/internal/domain"
)

type EventRepository interface {
	// Redis operations
	BufferEvent(ctx context.Context, event *domain.Event) error
	GetBufferedEventsForUser(ctx context.Context, userID string) ([]*domain.Event, error)
	ClearUserBuffer(ctx context.Context, userID string) error

	// Postgres operations
	StoreEvents(ctx context.Context, events []*domain.Event) error
	GetUndeliveredEvents(ctx context.Context, userID string) ([]*domain.Event, error)
	MarkEventsAsDelivered(ctx context.Context, eventIDs []string) error
}


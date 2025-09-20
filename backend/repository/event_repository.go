package repository

import (
	"context"
	"chat-app/models"
	"github.com/google/uuid"
	"time"
)

type EventRepository interface {
	// For Redis (buffering)
	BufferEvent(ctx context.Context, event *models.Event) error
	GetBufferedEvents(ctx context.Context, count int) ([]*models.Event, error)
	DeleteBufferedEvents(ctx context.Context, events []*models.Event) error

	// For Postgres (durable storage)
	Store(ctx context.Context, event *models.Event) error
	FetchUndelivered(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]*models.Event, error)
	Delete(ctx context.Context, eventID uuid.UUID) error
	StoreBatch(ctx context.Context, events []*models.Event) error
}


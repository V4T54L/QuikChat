package usecase

import (
	"context"
	"time"

	"chat-app/models"
	"chat-app/repository"

	"github.com/google/uuid"
)

type EventUsecase interface {
	StoreEvent(ctx context.Context, event *models.Event) error
	GetUndeliveredEvents(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]*models.Event, error)
	MarkEventAsDelivered(ctx context.Context, eventID uuid.UUID) error
}

type eventUsecase struct {
	redisRepo repository.EventRepository
	dbRepo    repository.EventRepository
}

func NewEventUsecase(redisRepo, dbRepo repository.EventRepository) EventUsecase {
	return &eventUsecase{
		redisRepo: redisRepo,
		dbRepo:    dbRepo,
	}
}

func (u *eventUsecase) StoreEvent(ctx context.Context, event *models.Event) error {
	// All events are buffered in Redis first
	return u.redisRepo.BufferEvent(ctx, event)
}

func (u *eventUsecase) GetUndeliveredEvents(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]*models.Event, error) {
	// Fetch from durable storage (Postgres)
	return u.dbRepo.FetchUndelivered(ctx, userID, cursor, limit)
}

func (u *eventUsecase) MarkEventAsDelivered(ctx context.Context, eventID uuid.UUID) error {
	// Once delivered, remove from durable storage
	return u.dbRepo.Delete(ctx, eventID)
}


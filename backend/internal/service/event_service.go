package service

import (
	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"chat-app/internal/usecase"
	"chat-app/pkg/util"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type eventService struct {
	redisEventRepo repository.EventRepository
	pgEventRepo    repository.EventRepository
	userRepo       repository.UserRepository
}

func NewEventService(redisEventRepo, pgEventRepo repository.EventRepository, userRepo repository.UserRepository) usecase.EventUsecase {
	return &eventService{
		redisEventRepo: redisEventRepo,
		pgEventRepo:    pgEventRepo,
		userRepo:       userRepo,
	}
}

func (s *eventService) CreateAndBufferEvent(ctx context.Context, eventType string, payload interface{}, recipientID string) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	event := &domain.Event{
		ID:          util.NewUUID(),
		Type:        eventType,
		Payload:     payloadJSON,
		RecipientID: recipientID,
		CreatedAt:   time.Now(),
	}

	return s.redisEventRepo.BufferEvent(ctx, event)
}

// PersistBufferedEvents is intended to be run as a background job.
// It's a simplified implementation. A real-world scenario would need more robust handling
// of user lists, locking, and error recovery.
func (s *eventService) PersistBufferedEvents(ctx context.Context) {
	// In a real app, you'd get a list of active users with buffered events.
	// For this project, we'll assume we can iterate through all users, which is not scalable.
	// This is a placeholder for a more complex logic.
	log.Println("Background worker: Persisting buffered events from Redis to Postgres is not fully implemented for scalability. This is a conceptual placeholder.")
	// A proper implementation would:
	// 1. Get a list of all user IDs that have buffered events (e.g., from a Redis SET).
	// 2. For each user:
	//    a. Get all buffered events.
	//    b. Store them in Postgres.
	//    c. Clear the Redis buffer for that user.
	// This needs careful implementation to avoid race conditions and data loss.
}


package redis

import (
	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	userEventBufferKeyPrefix = "events:"
	bufferTTL                = 48 * time.Hour
)

type redisEventRepository struct {
	client *redis.Client
}

func NewRedisEventRepository(client *redis.Client) repository.EventRepository {
	return &redisEventRepository{client: client}
}

func (r *redisEventRepository) userBufferKey(userID string) string {
	return fmt.Sprintf("%s%s", userEventBufferKeyPrefix, userID)
}

func (r *redisEventRepository) BufferEvent(ctx context.Context, event *domain.Event) error {
	key := r.userBufferKey(event.RecipientID)

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, eventJSON)
	pipe.Expire(ctx, key, bufferTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute redis pipeline for buffering event: %w", err)
	}
	return nil
}

func (r *redisEventRepository) GetBufferedEventsForUser(ctx context.Context, userID string) ([]*domain.Event, error) {
	key := r.userBufferKey(userID)
	eventStrings, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get buffered events from redis: %w", err)
	}

	events := make([]*domain.Event, 0, len(eventStrings))
	for _, eventStr := range eventStrings {
		var event domain.Event
		if err := json.Unmarshal([]byte(eventStr), &event); err != nil {
			// Log error but continue processing other events
			fmt.Printf("Error unmarshalling event from redis: %v\n", err)
			continue
		}
		events = append(events, &event)
	}
	return events, nil
}

func (r *redisEventRepository) ClearUserBuffer(ctx context.Context, userID string) error {
	key := r.userBufferKey(userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear user event buffer in redis: %w", err)
	}
	return nil
}

// These methods are for the Postgres implementation, so they are no-ops here.
func (r *redisEventRepository) StoreEvents(ctx context.Context, events []*domain.Event) error {
	return nil
}
func (r *redisEventRepository) GetUndeliveredEvents(ctx context.Context, userID string) ([]*domain.Event, error) {
	return nil, nil
}
func (r *redisEventRepository) MarkEventsAsDelivered(ctx context.Context, eventIDs []string) error {
	return nil
}


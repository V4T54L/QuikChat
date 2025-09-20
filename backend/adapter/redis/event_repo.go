package redis

import (
	"context"
	"encoding/json"
	"time"

	"chat-app/backend/models"
	"chat-app/backend/repository"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	eventBufferKey = "event_buffer"
)

type redisEventRepository struct {
	rdb *redis.Client
}

func NewRedisEventRepository(rdb *redis.Client) repository.EventRepository {
	return &redisEventRepository{rdb: rdb}
}

// These methods are for the Redis part of the EventRepository interface
func (r *redisEventRepository) BufferEvent(ctx context.Context, event *models.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	// Use a sorted set, score is timestamp. This allows easy retrieval in order.
	return r.rdb.ZAdd(ctx, eventBufferKey, &redis.Z{
		Score:  float64(event.CreatedAt.UnixNano()),
		Member: data,
	}).Err()
}

func (r *redisEventRepository) GetBufferedEvents(ctx context.Context, count int) ([]*models.Event, error) {
	results, err := r.rdb.ZRange(ctx, eventBufferKey, 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}

	var events []*models.Event
	for _, res := range results {
		var event models.Event
		if err := json.Unmarshal([]byte(res), &event); err == nil {
			events = append(events, &event)
		}
	}
	return events, nil
}

func (r *redisEventRepository) DeleteBufferedEvents(ctx context.Context, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}
	members := make([]interface{}, len(events))
	for i, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			// Log this error but continue
			continue
		}
		members[i] = data
	}
	return r.rdb.ZRem(ctx, eventBufferKey, members...).Err()
}

// These methods are for the Postgres part of the interface, so they are no-ops here.
func (r *redisEventRepository) Store(ctx context.Context, event *models.Event) error {
	return nil // No-op
}
func (r *redisEventRepository) FetchUndelivered(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]*models.Event, error) {
	return nil, nil // No-op
}
func (r *redisEventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	return nil // No-op
}
func (r *redisEventRepository) StoreBatch(ctx context.Context, events []*models.Event) error {
	return nil // No-op
}

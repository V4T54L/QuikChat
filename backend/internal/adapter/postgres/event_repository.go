package postgres

import (
	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresEventRepository struct {
	db *pgxpool.Pool
}

func NewPostgresEventRepository(db *pgxpool.Pool) repository.EventRepository {
	return &postgresEventRepository{db: db}
}

func (r *postgresEventRepository) StoreEvents(ctx context.Context, events []*domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO events (id, type, payload, recipient_id, created_at) VALUES ($1, $2, $3, $4, $5)`
	for _, event := range events {
		_, err := tx.Exec(ctx, query, event.ID, event.Type, event.Payload, event.RecipientID, event.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresEventRepository) GetUndeliveredEvents(ctx context.Context, userID string) ([]*domain.Event, error) {
	query := `SELECT id, type, payload, recipient_id, is_delivered, created_at FROM events WHERE recipient_id = $1 AND is_delivered = FALSE ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query undelivered events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.ID, &event.Type, &event.Payload, &event.RecipientID, &event.IsDelivered, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}
		events = append(events, &event)
	}
	return events, nil
}

func (r *postgresEventRepository) MarkEventsAsDelivered(ctx context.Context, eventIDs []string) error {
	if len(eventIDs) == 0 {
		return nil
	}
	query := `UPDATE events SET is_delivered = TRUE WHERE id = ANY($1::uuid[])`
	_, err := r.db.Exec(ctx, query, eventIDs)
	if err != nil {
		return fmt.Errorf("failed to mark events as delivered: %w", err)
	}
	return nil
}

// These methods are for the Redis implementation, so they are no-ops here.
func (r *postgresEventRepository) BufferEvent(ctx context.Context, event *domain.Event) error {
	return nil
}
func (r *postgresEventRepository) GetBufferedEventsForUser(ctx context.Context, userID string) ([]*domain.Event, error) {
	return nil, nil
}
func (r *postgresEventRepository) ClearUserBuffer(ctx context.Context, userID string) error {
	return nil
}


package postgres

import (
	"context"
	"database/sql"
	"time"

	"chat-app/backend/models"
	"chat-app/backend/repository"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type postgresEventRepository struct {
	db *sql.DB
}

func NewPostgresEventRepository(db *sql.DB) repository.EventRepository {
	return &postgresEventRepository{db: db}
}

// These methods are for the Postgres part of the EventRepository interface
func (r *postgresEventRepository) Store(ctx context.Context, event *models.Event) error {
	query := `INSERT INTO events (id, type, payload, recipient_id, sender_id, created_at)
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.Type, event.Payload, event.RecipientID, event.SenderID, event.CreatedAt)
	return err
}

func (r *postgresEventRepository) StoreBatch(ctx context.Context, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("events", "id", "type", "payload", "recipient_id", "sender_id", "created_at"))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		_, err = stmt.ExecContext(ctx, event.ID, event.Type, event.Payload, event.RecipientID, event.SenderID, event.CreatedAt)
		if err != nil {
			return err
		}
	}

	if _, err = stmt.ExecContext(ctx); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresEventRepository) FetchUndelivered(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]*models.Event, error) {
	query := `SELECT id, type, payload, recipient_id, sender_id, created_at
              FROM events
              WHERE recipient_id = $1 AND created_at > $2
              ORDER BY created_at ASC
              LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, userID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Type, &event.Payload, &event.RecipientID, &event.SenderID, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}

func (r *postgresEventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, eventID)
	return err
}

// These methods are for the Redis part of the interface, so they are no-ops here.
func (r *postgresEventRepository) BufferEvent(ctx context.Context, event *models.Event) error {
	return nil // No-op
}
func (r *postgresEventRepository) GetBufferedEvents(ctx context.Context, count int) ([]*models.Event, error) {
	return nil, nil // No-op
}
func (r *postgresEventRepository) DeleteBufferedEvents(ctx context.Context, events []*models.Event) error {
	return nil // No-op
}

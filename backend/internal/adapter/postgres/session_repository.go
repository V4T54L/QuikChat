package postgres

import (
	"context"
	"errors"

	"chat-app/internal/domain"
	"chat-app/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresSessionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSessionRepository(db *pgxpool.Pool) repository.SessionRepository {
	return &postgresSessionRepository{db: db}
}

func (r *postgresSessionRepository) Store(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, session.ID, session.UserID, session.ExpiresAt)
	return err
}

func (r *postgresSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	query := `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1`

	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}
	return session, nil
}

func (r *postgresSessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *postgresSessionRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

